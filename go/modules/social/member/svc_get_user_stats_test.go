package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetUserStats_正常查询_返回统计信息 从上下文提取用户ID_应返回用户统计信息
func TestSvcGetUserStats_正常查询_返回统计信息(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetUserStats(mockRepo)

	// 通过上下文注入当前登录用户 ID
	ctx := context.WithValue(context.Background(), CtxKeyUserID, "user-001")

	req := &pb.GetUserStatsRequest{}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.UserStats == nil {
		t.Fatal("期望 UserStats 不为空")
	}
	if resp.UserId != "user-001" {
		t.Errorf("期望 UserId 为 user-001，实际得到 %s", resp.UserId)
	}
	// 验证默认统计值（懒初始化）
	if resp.UserStats.TopicsCount != 0 {
		t.Errorf("期望默认 TopicsCount 为 0，实际得到 %d", resp.UserStats.TopicsCount)
	}
	if resp.UserStats.CommentsCount != 0 {
		t.Errorf("期望默认 CommentsCount 为 0，实际得到 %d", resp.UserStats.CommentsCount)
	}
	if resp.UserStats.LikesReceived != 0 {
		t.Errorf("期望默认 LikesReceived 为 0，实际得到 %d", resp.UserStats.LikesReceived)
	}
	if resp.UserStats.GroupsCount != 0 {
		t.Errorf("期望默认 GroupsCount 为 0，实际得到 %d", resp.UserStats.GroupsCount)
	}
}
