package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetGroupMemberUserIds_正常查询_返回用户ID列表 提供有效group_id_应返回成员用户ID列表
func TestSvcGetGroupMemberUserIds_正常查询_返回用户ID列表(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetGroupMemberUserIds(mockRepo)

	// 预先创建成员
	member1 := &GroupMember{
		ID:     "member-001",
		GroupID: "group-001",
		UserID: "user-001",
		Role:   int8(pb.GroupMemberRole_GROUP_ROLE_MEMBER),
		Status: 1,
	}
	member2 := &GroupMember{
		ID:     "member-002",
		GroupID: "group-001",
		UserID: "user-002",
		Role:   int8(pb.GroupMemberRole_GROUP_ROLE_MEMBER),
		Status: 1,
	}
	mockRepo.members[member1.ID] = member1
	mockRepo.memberIdx[member1.GroupID+":"+member1.UserID] = member1
	mockRepo.members[member2.ID] = member2
	mockRepo.memberIdx[member2.GroupID+":"+member2.UserID] = member2

	req := &pb.GetGroupMemberUserIdsRequest{
		GroupId:   "group-001",
		Page:      1,
		PageSize:  20,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if len(resp.UserIds) != 2 {
		t.Errorf("期望返回 2 个用户 ID，实际得到 %d", len(resp.UserIds))
	}
	if resp.Total != 2 {
		t.Errorf("期望 Total 为 2，实际得到 %d", resp.Total)
	}
}

// TestSvcGetGroupMemberUserIds_空group_id_返回参数错误 group_id为空_应返回参数校验错误
func TestSvcGetGroupMemberUserIds_空group_id_返回参数错误(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetGroupMemberUserIds(mockRepo)

	req := &pb.GetGroupMemberUserIdsRequest{
		GroupId: "", // 空 group_id
		Page:    1,
		PageSize: 20,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code, resp.Result.Message)
	}
}
