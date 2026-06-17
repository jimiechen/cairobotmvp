package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcBanMember_正常封禁 当目标成员存在且非群主时_应更新封禁状态
func TestSvcBanMember_正常封禁(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcBanMember(mockRepo)

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-target",
		Role:    GroupMemberRoleMember, // 普通成员
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-target"] = member

	req := &pb.BanMemberRequest{
		GroupId: "group-001",
		UserId:  "user-target",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证封禁状态已更新
	updatedMember := mockRepo.memberIdx["group-001:user-target"]
	if updatedMember.Status != 3 { // 已移除/封禁
		t.Errorf("期望状态为已移除(3)，实际得到 %d", updatedMember.Status)
	}
	if updatedMember.BannedAt == 0 {
		t.Error("期望 BannedAt 被设置")
	}
}

// TestSvcBanMember_目标非成员 当用户不是成员时_应返回非成员错误
func TestSvcBanMember_目标非成员(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcBanMember(mockRepo)

	req := &pb.BanMemberRequest{
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

// TestSvcBanMember_封禁群主 当目标是群主时_应返回禁止操作错误
func TestSvcBanMember_封禁群主(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcBanMember(mockRepo)

	member := &GroupMember{
		ID:      "member-owner",
		GroupID: "group-001",
		UserID:  "user-owner",
		Role:    GroupMemberRoleOwner, // 群主
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-owner"] = member

	req := &pb.BanMemberRequest{
		GroupId: "group-001",
		UserId:  "user-owner",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER) {
		t.Errorf("期望不能封禁群主错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER, resp.Result.Code)
	}
}
