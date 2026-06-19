package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetGroupStats_正常查询_返回统计信息 提供有效group_id_应返回圈子统计信息
func TestSvcGetGroupStats_正常查询_返回统计信息(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetGroupStats(mockRepo)

	// 预先设置统计数据
	stats := &GroupStats{
		GroupID:      "group-001",
		MembersCount: 100,
		TopicsCount:  50,
	}
	mockRepo.stats["group-001"] = stats

	req := &pb.GetGroupStatsRequest{
		GroupId: "group-001",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.Stats == nil {
		t.Fatal("期望 Stats 不为空")
	}
	if resp.Stats.MembersCount != 100 {
		t.Errorf("期望 MembersCount 为 100，实际得到 %d", resp.Stats.MembersCount)
	}
	if resp.Stats.TopicsCount != 50 {
		t.Errorf("期望 TopicsCount 为 50，实际得到 %d", resp.Stats.TopicsCount)
	}
	if resp.GroupId != "group-001" {
		t.Errorf("期望 GroupId 为 group-001，实际得到 %s", resp.GroupId)
	}
}

// TestSvcGetGroupStats_空group_id_返回参数错误 group_id为空_应返回参数校验错误
func TestSvcGetGroupStats_空group_id_返回参数错误(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetGroupStats(mockRepo)

	req := &pb.GetGroupStatsRequest{
		GroupId: "", // 空 group_id
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code, resp.Result.Message)
	}
}
