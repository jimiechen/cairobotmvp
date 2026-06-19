package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetReplyList_正常查询_返回评论列表 提供有效topic_id_应返回评论回复列表
func TestSvcGetReplyList_正常查询_返回评论列表(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetReplyList(mockRepo)

	// 预先创建评论
	reply1 := &TopicReply{
		ID:         "reply-001",
		TopicID:    "topic-001",
		Content:    "第一条评论",
		AuthorID:   "user-001",
		AuthorName: "用户A",
		Status:     1,
		LikeCount:  5,
		Level:      1,
	}
	reply2 := &TopicReply{
		ID:         "reply-002",
		TopicID:    "topic-001",
		Content:    "第二条评论",
		AuthorID:   "user-002",
		AuthorName: "用户B",
		Status:     1,
		LikeCount:  3,
		Level:      1,
	}
	mockRepo.replies[reply1.ID] = reply1
	mockRepo.replies[reply2.ID] = reply2

	req := &pb.GetReplyListRequest{
		TopicId:  "topic-001",
		Page:     1,
		PageSize: 20,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.Total != 2 {
		t.Errorf("期望 Total 为 2，实际得到 %d", resp.Total)
	}
	if resp.TopicId != "topic-001" {
		t.Errorf("期望 TopicId 为 topic-001，实际得到 %s", resp.TopicId)
	}
	// 验证返回评论数量（map 返回顺序不确定，不校验顺序）
	if len(resp.Replies) != 2 {
		t.Errorf("期望返回 2 条评论，实际得到 %d", len(resp.Replies))
	}
}

// TestSvcGetReplyList_空topic_id_返回参数错误 topic_id为空_应返回参数校验错误
func TestSvcGetReplyList_空topic_id_返回参数错误(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcGetReplyList(mockRepo)

	req := &pb.GetReplyListRequest{
		TopicId:  "", // 空 topic_id
		Page:     1,
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
