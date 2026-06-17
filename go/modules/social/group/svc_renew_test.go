package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcRenewMember_正常续费 当成员存在时_应更新到期时间并返回成功
func TestSvcRenewMember_正常续费(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRenewMember(mockRepo)

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-001",
		Role:    GroupMemberRoleMember,
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-001"] = member

	req := &pb.RenewMemberRequest{
		GroupId:        "group-001",
		UserId:         "user-001",
		RenewPeriodEnd: 9999999999999, // 未来时间戳
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证到期时间已更新
	updatedMember := mockRepo.memberIdx["group-001:user-001"]
	if updatedMember.MembershipExpiresAt != 9999999999999 {
		t.Errorf("期望到期时间被更新为 9999999999999，实际得到 %d", updatedMember.MembershipExpiresAt)
	}
}

// TestSvcRenewMember_非成员续费 当用户不是成员时_应返回非成员错误
func TestSvcRenewMember_非成员续费(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRenewMember(mockRepo)

	req := &pb.RenewMemberRequest{
		GroupId: "group-001",
		UserId:  "user-nonexistent",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER) {
		t.Errorf("期望非成员错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER, resp.Result.Code)
	}
}

// TestSvcRenewMember_缺少GroupId 当圈子ID为空时_应返回参数校验错误
func TestSvcRenewMember_缺少GroupId(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRenewMember(mockRepo)

	req := &pb.RenewMemberRequest{
		GroupId: "", // 空 groupId
		UserId:  "user-001",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY) {
		t.Errorf("期望圈子ID不能为空错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_ID_EMPTY, resp.Result.Code)
	}
}
