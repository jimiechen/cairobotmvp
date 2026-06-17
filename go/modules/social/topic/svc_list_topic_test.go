package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcListTopic_正常查询 当有帖子数据时_应返回帖子列表和总数
func TestSvcListTopic_正常查询(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Title: "帖子1", GroupID: "group-001", Status: TopicStatusInactive}
	mockRepo.topics["topic-002"] = &Topic{ID: "topic-002", Title: "帖子2", GroupID: "group-001", Status: TopicStatusInactive}

	svc := NewSvcListTopic(mockRepo)

	req := &pb.GetTopicListRequest{
		GroupId:  "group-001",
		Page:     1,
		PageSize: 10,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d", resp.Result.Code)
	}
	if resp.Total < 2 {
		t.Errorf("期望至少 2 条帖子，实际得到 %d", resp.Total)
	}
}

// TestSvcListTopic_空列表 当没有帖子时_应返回空列表
func TestSvcListTopic_空列表(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcListTopic(mockRepo)

	req := &pb.GetTopicListRequest{
		Page:     1,
		PageSize: 10,
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d", resp.Result.Code)
	}
	if len(resp.Topics) != 0 {
		t.Errorf("期望空列表，实际得到 %d 条", len(resp.Topics))
	}
}
