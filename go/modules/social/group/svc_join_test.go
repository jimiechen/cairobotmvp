package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	memberpkg "github.com/jimiechen/mineplanet/go/modules/social/member"
)

// TestSvcJoin_正常加入 当圈子存在且用户未加入时_应返回成功响应
func TestSvcJoin_正常加入(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcJoin(mockRepo, nil)

	// 预先创建一个圈子
	group := &Group{ID: "group-001", Slug: "test-group"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

ctx := context.WithValue(context.Background(), memberpkg.CtxKeyUserID, "user-001")
	req := &pb.JoinGroupRequest{
		GroupId:    "group-001",
		JoinReason: "想学习技术",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.MemberId == "" {
		t.Error("期望 MemberId 不为空")
	}
	if resp.Status != pb.JoinStatus_JOIN_STATUS_JOINED {
		t.Errorf("期望状态 JOINED，实际得到 %v", resp.Status)
	}
}

// TestSvcJoin_圈子不存在 当groupId不存在时_应返回圈子不存在错误
func TestSvcJoin_圈子不存在(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcJoin(mockRepo, nil)

	ctx := context.WithValue(context.Background(), memberpkg.CtxKeyUserID, "user-001")
	req := &pb.JoinGroupRequest{
		GroupId: "nonexistent-group",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND) {
		t.Errorf("期望圈子不存在错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_FOUND, resp.Result.Code)
	}
}

// TestSvcJoin_已成员重复加入 当用户已是成员时_应返回已成员错误
func TestSvcJoin_已成员重复加入(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcJoin(mockRepo, nil)

	// 预先创建圈子和成员关系
	group := &Group{ID: "group-002", Slug: "test-group-2"}
	mockRepo.groups[group.ID] = group
	mockRepo.groups[group.Slug] = group

	member := &GroupMember{ID: "member-001", GroupID: "group-002", UserID: "user-002"}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx[member.GroupID+":"+member.UserID] = member

	ctx := context.WithValue(context.Background(), memberpkg.CtxKeyUserID, "user-002")
	req := &pb.JoinGroupRequest{
		GroupId: "group-002",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_ALREADY_MEMBER) {
		t.Errorf("期望已成员错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_ALREADY_MEMBER, resp.Result.Code)
	}
}
