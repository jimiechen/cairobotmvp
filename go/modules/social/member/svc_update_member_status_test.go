package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// fakePublisher 实现 event.Publisher 接口的测试用 Fake，用于捕获发布的事件
type fakePublisher struct {
	publishedEvents []event.DomainEvent
}

func newFakePublisher() *fakePublisher {
	return &fakePublisher{
		publishedEvents: make([]event.DomainEvent, 0),
	}
}

func (f *fakePublisher) Publish(ctx context.Context, evt event.DomainEvent) error {
	f.publishedEvents = append(f.publishedEvents, evt)
	return nil
}

// TestSvcUpdateMemberStatus_封禁正常用户_成功 正常活跃用户封禁_状态应更新为封禁
func TestSvcUpdateMemberStatus_封禁正常用户_成功(t *testing.T) {
	mockRepo := newMockRepository()
	pub := newFakePublisher()
	svc := NewSvcUpdateMemberStatus(mockRepo, pub)

	// 预先创建一个正常活跃用户
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Nickname: "测试用户",
		Status:   UserStatusActive,
	}
	mockRepo.users[existingUser.ID] = existingUser

	ctx := context.WithValue(context.Background(), CtxKeyUserID, "admin-001")

	req := &pb.UpdateMemberStatusRequest{
		UserId: "user-001",
		Status: int32(UserStatusSuspended),
		Reason: "违反社区规范",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.Status != int32(UserStatusSuspended) {
		t.Errorf("期望状态为 %d（封禁），实际得到 %d", UserStatusSuspended, resp.Status)
	}

	// 验证数据库中的状态已更新
	updatedUser, _ := mockRepo.GetUserByID(ctx, "user-001")
	if updatedUser.Status != UserStatusSuspended {
		t.Errorf("期望数据库中状态为 %d（封禁），实际得到 %d", UserStatusSuspended, updatedUser.Status)
	}

	// 验证事件已发布
	if len(pub.publishedEvents) != 1 {
		t.Fatalf("期望发布 1 个事件，实际发布 %d 个", len(pub.publishedEvents))
	}
	if pub.publishedEvents[0].Type != event.EventUserStatusChanged {
		t.Errorf("期望事件类型为 %s，实际得到 %s", event.EventUserStatusChanged, pub.publishedEvents[0].Type)
	}
}

// TestSvcUpdateMemberStatus_幂等操作_同一状态不重复写 已封禁用户再次封禁_应直接返回成功不重复写入
func TestSvcUpdateMemberStatus_幂等操作_同一状态不重复写(t *testing.T) {
	mockRepo := newMockRepository()
	pub := newFakePublisher()
	svc := NewSvcUpdateMemberStatus(mockRepo, pub)

	// 预先创建一个已封禁用户
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Nickname: "测试用户",
		Status:   UserStatusSuspended,
		UpdatedAt: 1000, // 记录原始更新时间
	}
	mockRepo.users[existingUser.ID] = existingUser

	ctx := context.WithValue(context.Background(), CtxKeyUserID, "admin-001")

	req := &pb.UpdateMemberStatusRequest{
		UserId: "user-001",
		Status: int32(UserStatusSuspended), // 再次设置为封禁（相同状态）
		Reason: "重复封禁操作",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}

	// 验证 UpdatedAt 未被修改（幂等：未执行 DB 写入）
	unchangedUser, _ := mockRepo.GetUserByID(ctx, "user-001")
	if unchangedUser.UpdatedAt != 1000 {
		t.Errorf("期望 UpdatedAt 保持不变（幂等不写 DB），实际得到 %d", unchangedUser.UpdatedAt)
	}

	// 幂等操作不应发布事件
	if len(pub.publishedEvents) != 0 {
		t.Errorf("期望幂等操作不发布事件，实际发布 %d 个", len(pub.publishedEvents))
	}
}

// TestSvcUpdateMemberStatus_用户不存在_返回错误 用户ID不存在_应返回目标用户不存在错误
func TestSvcUpdateMemberStatus_用户不存在_返回错误(t *testing.T) {
	mockRepo := newMockRepository()
	pub := newFakePublisher()
	svc := NewSvcUpdateMemberStatus(mockRepo, pub)

	ctx := context.Background()

	req := &pb.UpdateMemberStatusRequest{
		UserId: "nonexistent-user",
		Status: int32(UserStatusSuspended),
		Reason: "封禁不存在用户",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code, resp.Result.Message)
	}
}

// TestSvcUpdateMemberStatus_空user_id_返回参数错误 user_id为空_应返回参数校验错误
func TestSvcUpdateMemberStatus_空user_id_返回参数错误(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUpdateMemberStatus(mockRepo, nil)

	req := &pb.UpdateMemberStatusRequest{
		UserId: "", // 空 user_id
		Status: int32(UserStatusSuspended),
		Reason: "测试",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code, resp.Result.Message)
	}
}

// TestSvcUpdateMemberStatus_恢复已封禁用户_成功 封禁用户恢复为正常_状态应更新为活跃
func TestSvcUpdateMemberStatus_恢复已封禁用户_成功(t *testing.T) {
	mockRepo := newMockRepository()
	pub := newFakePublisher()
	svc := NewSvcUpdateMemberStatus(mockRepo, pub)

	// 预先创建一个已封禁用户
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Nickname: "测试用户",
		Status:   UserStatusSuspended,
	}
	mockRepo.users[existingUser.ID] = existingUser

	ctx := context.WithValue(context.Background(), CtxKeyUserID, "admin-001")

	req := &pb.UpdateMemberStatusRequest{
		UserId: "user-001",
		Status: int32(UserStatusActive), // 恢复为正常活跃
		Reason: "解封恢复",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.Status != int32(UserStatusActive) {
		t.Errorf("期望状态为 %d（正常活跃），实际得到 %d", UserStatusActive, resp.Status)
	}

	// 验证数据库中的状态已更新
	updatedUser, _ := mockRepo.GetUserByID(ctx, "user-001")
	if updatedUser.Status != UserStatusActive {
		t.Errorf("期望数据库中状态为 %d（正常活跃），实际得到 %d", UserStatusActive, updatedUser.Status)
	}

	// 验证事件已发布（状态从封禁变为活跃）
	if len(pub.publishedEvents) != 1 {
		t.Fatalf("期望发布 1 个事件，实际发布 %d 个", len(pub.publishedEvents))
	}
	if pub.publishedEvents[0].Type != event.EventUserStatusChanged {
		t.Errorf("期望事件类型为 %s，实际得到 %s", event.EventUserStatusChanged, pub.publishedEvents[0].Type)
	}
}
