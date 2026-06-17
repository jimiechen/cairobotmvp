package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcRemoveMember_正常踢出 当目标成员存在且非群主时_应更新移除状态
func TestSvcRemoveMember_正常踢出(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRemoveMember(mockRepo)

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-target",
		Role:    GroupMemberRoleMember, // 普通成员
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-target"] = member

	req := &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "user-target",
		Reason:  "多次违规",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证移除状态已更新
	updatedMember := mockRepo.memberIdx["group-001:user-target"]
	if updatedMember.Status != GroupMemberStatusBanned { // 已移除
		t.Errorf("期望状态为已移除(3)，实际得到 %d", updatedMember.Status)
	}
	if updatedMember.BanReason != "多次违规" {
		t.Errorf("期望踢出原因为 '多次违规'，实际得到 '%s'", updatedMember.BanReason)
	}
}

// TestSvcRemoveMember_目标非成员 当用户不是成员时_应返回非成员错误
func TestSvcRemoveMember_目标非成员(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRemoveMember(mockRepo)

	req := &pb.RemoveMemberRequest{
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

// TestSvcRemoveMember_踢出群主 当目标是群主时_应返回禁止操作错误
func TestSvcRemoveMember_踢出群主(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcRemoveMember(mockRepo)

	member := &GroupMember{
		ID:      "member-owner",
		GroupID: "group-001",
		UserID:  "user-owner",
		Role:    GroupMemberRoleOwner, // 群主
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-owner"] = member

	req := &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "user-owner",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_REMOVE_OWNER) {
		t.Errorf("期望不能移除群主错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_CANNOT_REMOVE_OWNER, resp.Result.Code)
	}
}
