package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcLeave_正常退出 当用户是普通成员时_应返回成功
func TestSvcLeave_正常退出(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcLeave(mockRepo, nil)
	// 预设群组和成员
	group := &Group{ID: "group-001", Slug: "test-leave"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-001",
		Role:    GroupMemberRoleMember, // 普通成员
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-001"] = member

	ctx := WithUserID(context.Background(), "user-001")
	req := &pb.LeaveGroupRequest{GroupId: "group-001"}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证成员状态已更新为已退出(2)
	updatedMember := mockRepo.memberIdx["group-001:user-001"]
	if updatedMember.Status != GroupMemberStatusLeft {
		t.Errorf("期望成员状态为已退出(2)，实际得到 %d", updatedMember.Status)
	}
}

// TestSvcLeave_非成员退出 当用户不是成员时_应返回非成员错误
func TestSvcLeave_非成员退出(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcLeave(mockRepo, nil)
	group := &Group{ID: "group-002", Slug: "test-leave-2"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	ctx := WithUserID(context.Background(), "user-nonexistent")
	req := &pb.LeaveGroupRequest{GroupId: "group-002"}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER) {
		t.Errorf("期望非成员错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER, resp.Result.Code)
	}
}

// TestSvcLeave_群主退出 当用户是群主时_应返回禁止退出错误
func TestSvcLeave_群主退出(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcLeave(mockRepo, nil)
	group := &Group{ID: "group-003", Slug: "test-leave-3"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	member := &GroupMember{
		ID:      "member-owner",
		GroupID: "group-003",
		UserID:  "user-owner",
		Role:    GroupMemberRoleOwner, // 群主
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-003:user-owner"] = member

	ctx := WithUserID(context.Background(), "user-owner")
	req := &pb.LeaveGroupRequest{GroupId: "group-003"}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10731 { // GROUP_ERROR_OWNER_CANNOT_LEAVE
		t.Errorf("期望群主禁止退出错误码 10731，实际得到 %d", resp.Result.Code)
	}
}
