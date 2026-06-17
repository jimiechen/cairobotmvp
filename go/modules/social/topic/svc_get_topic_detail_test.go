package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcGetTopicDetail_正常查询 当topic存在时_应返回详情信息
func TestSvcGetTopicDetail_正常查询(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{
		ID: "topic-001", Title: "详情帖", Content: "详细内容",
		AuthorID: "user-001", GroupID: "group-001", Status: TopicStatusInactive,
	}
	svc := NewSvcGetTopicDetail(mockRepo)

	req := &pb.BatchGetTopicInfoRequest{TopicIds: []string{"topic-001"}}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if len(resp.Topics) != 1 {
		t.Errorf("期望 1 条详情，实际得到 %d", len(resp.Topics))
	}
	if resp.Topics[0].TopicId != "topic-001" {
		t.Errorf("期望 topic_id=topic-001，实际得到 %s", resp.Topics[0].TopicId)
	}
}

// TestSvcGetTopicDetail_批量查询 当传入多个id时_应返回所有存在的详情
func TestSvcGetTopicDetail_批量查询(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Title: "帖子1", Status: TopicStatusInactive}
	mockRepo.topics["topic-002"] = &Topic{ID: "topic-002", Title: "帖子2", Status: TopicStatusInactive}
	svc := NewSvcGetTopicDetail(mockRepo)

	req := &pb.BatchGetTopicInfoRequest{TopicIds: []string{"topic-001", "topic-002", "topic-noexist"}}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if len(resp.Topics) != 2 {
		t.Errorf("期望 2 条详情（1条不存在），实际得到 %d", len(resp.Topics))
	}
}
