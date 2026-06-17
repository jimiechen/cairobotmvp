package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcUpdateMemberRole_正常修改角色 当目标成员存在且角色合法时_应更新角色
func TestSvcUpdateMemberRole_正常修改角色(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUpdateMemberRole(mockRepo)

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-target",
		Role:    GroupMemberRoleMember, // 普通成员
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-target"] = member

	req := &pb.UpdateMemberRoleRequest{
		GroupId: "group-001",
		UserId:  "user-target",
		Role:    pb.GroupMemberRole_GROUP_ROLE_ADMIN, // 提升为管理员
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证角色已更新
	updatedMember := mockRepo.memberIdx["group-001:user-target"]
	if updatedMember.Role != 2 { // 管理员
		t.Errorf("期望角色为管理员(2)，实际得到 %d", updatedMember.Role)
	}
}

// TestSvcUpdateMemberRole_目标非成员 当用户不是成员时_应返回非成员错误
func TestSvcUpdateMemberRole_目标非成员(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUpdateMemberRole(mockRepo)

	req := &pb.UpdateMemberRoleRequest{
		GroupId: "group-001",
		UserId:  "user-nonexistent",
		Role:    pb.GroupMemberRole_GROUP_ROLE_ADMIN,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER) {
		t.Errorf("期望非成员错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER, resp.Result.Code)
	}
}

// TestSvcUpdateMemberRole_设为群主 当尝试设为群主角色时_应返回无效角色错误
func TestSvcUpdateMemberRole_设为群主(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUpdateMemberRole(mockRepo)

	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-target",
		Role:    GroupMemberRoleMember,
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-target"] = member

	req := &pb.UpdateMemberRoleRequest{
		GroupId: "group-001",
		UserId:  "user-target",
		Role:    pb.GroupMemberRole_GROUP_ROLE_OWNER, // 尝试设为群主
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_INVALID_ROLE) {
		t.Errorf("期望无效角色错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_INVALID_ROLE, resp.Result.Code)
	}
}
