package group

import (
	"context"
	"testing"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcMuteMember_正常禁言 当目标成员存在且非群主时_应更新禁言状态
func TestSvcMuteMember_正常禁言(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcMuteMember(mockRepo, nil)
	member := &GroupMember{
		ID:      "member-001",
		GroupID: "group-001",
		UserID:  "user-target",
		Role:    GroupMemberRoleMember, // 普通成员
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-target"] = member

	req := &pb.MuteMemberRequest{
		GroupId:      "group-001",
		UserId:       "user-target",
		MuteDuration: pb.MuteDuration_MUTE_DURATION_1_HOUR,
		MuteReason:   "违规发言",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	// 验证禁言状态已更新
	updatedMember := mockRepo.memberIdx["group-001:user-target"]
	if updatedMember.Status != GroupMemberStatusMuted { // 已禁言
		t.Errorf("期望状态为已禁言(4)，实际得到 %d", updatedMember.Status)
	}
	if updatedMember.MutedUntil == 0 {
		t.Error("期望 MutedUntil 被设置")
	}
}

// TestSvcMuteMember_目标非成员 当用户不是成员时_应返回非成员错误
func TestSvcMuteMember_目标非成员(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcMuteMember(mockRepo, nil)
	req := &pb.MuteMemberRequest{
		GroupId:      "group-001",
		UserId:       "user-nonexistent",
		MuteDuration: pb.MuteDuration_MUTE_DURATION_1_HOUR,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER) {
		t.Errorf("期望非成员错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER, resp.Result.Code)
	}
}

// TestSvcMuteMember_禁言群主 当目标是群主时_应返回禁止操作错误
func TestSvcMuteMember_禁言群主(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcMuteMember(mockRepo, nil)
	member := &GroupMember{
		ID:      "member-owner",
		GroupID: "group-001",
		UserID:  "user-owner",
		Role:    GroupMemberRoleOwner, // 群主
		Status:  GroupMemberStatusActive,
	}
	mockRepo.members[member.ID] = member
	mockRepo.memberIdx["group-001:user-owner"] = member

	req := &pb.MuteMemberRequest{
		GroupId:      "group-001",
		UserId:       "user-owner",
		MuteDuration: pb.MuteDuration_MUTE_DURATION_1_HOUR,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER) {
		t.Errorf("期望不能禁言群主错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER, resp.Result.Code)
	}
}

// TestSvcMuteMember_calcMutedUntil 验证不同时长枚举的到期时间计算正确性
func TestSvcMuteMember_calcMutedUntil(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	hourMs := int64(time.Hour / time.Millisecond)

	tests := []struct {
		duration pb.MuteDuration
		expected int64
	}{
		{pb.MuteDuration_MUTE_DURATION_1_HOUR, nowMs + hourMs},
		{pb.MuteDuration_MUTE_DURATION_7_DAYS, nowMs + 7*24*hourMs},
	}

	for _, tt := range tests {
		result := calcMutedUntil(nowMs, tt.duration)
		if result != tt.expected {
			t.Errorf("calcMutedUntil(%v): 期望 %d，实际得到 %d", tt.duration, tt.expected, result)
		}
	}
}
