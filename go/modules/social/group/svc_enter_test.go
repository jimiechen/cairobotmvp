package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcEnter_正常进入 当圈子存在时_应返回圈子详情和成员信息
func TestSvcEnter_正常进入(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcEnter(mockRepo)

	group := &Group{
		ID:        "group-001",
		Name:      "测试圈子",
		Slug:      "test-enter",
		OwnerID:   "owner-001",
		Status:    GroupStatusActive,
		Visibility: GroupVisibilityPublic,
	}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	// 预设成员
	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-001",
		Role:    GroupMemberRoleMember,
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-001"] = member

	ctx := WithUserID(context.Background(), "user-001")
	req := &pb.GroupUserEnterRequest{
		GroupId: "group-001",
		UserId:  "user-001",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.GroupInfo == nil || resp.GroupInfo.Name != "测试圈子" {
		t.Error("期望返回正确的圈子信息")
	}
	if resp.UserMemberInfo == nil || resp.UserMemberInfo.UserId != "user-001" {
		t.Error("期望返回当前用户成员信息")
	}
}

// TestSvcEnter_圈子不存在 当groupId无效时_应返回圈子不存在错误
func TestSvcEnter_圈子不存在(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcEnter(mockRepo)

	ctx := context.Background()
	req := &pb.GroupUserEnterRequest{
		GroupId: "nonexistent",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND) {
		t.Errorf("期望圈子不存在错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_FOUND, resp.Result.Code)
	}
}

// TestSvcEnter_游客进入 未登录用户进入时_应返回圈子信息但无成员信息
func TestSvcEnter_游客进入(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcEnter(mockRepo)

	group := &Group{ID: "group-003", Name: "公开圈子", Slug: "public"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	ctx := context.Background() // 无 user_id
	req := &pb.GroupUserEnterRequest{
		GroupId: "group-003",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code)
	}
	if resp.UserMemberInfo != nil {
		t.Error("游客进入时 UserMemberInfo 应为 nil")
	}
}
