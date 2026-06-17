package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcUpdateTopic_正常更新 当帖子存在且有新标题时_应返回更新后的信息
func TestSvcUpdateTopic_正常更新(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Title: "原标题", Content: "原内容", Status: TopicStatusInactive}
	svc := NewSvcUpdateTopic(mockRepo)

	req := &pb.CreateTopicRequest{
		TopicId: "topic-001",
		Title:   "新标题",
		Content: "新内容",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.TopicInfo.Title != "新标题" {
		t.Errorf("期望标题为'新标题'，实际得到 '%s'", resp.TopicInfo.Title)
	}
}

// TestSvcUpdateTopic_帖子不存在 当topic_id不存在时_应返回帖子不存在错误
func TestSvcUpdateTopic_帖子不存在(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUpdateTopic(mockRepo)

	req := &pb.CreateTopicRequest{
		TopicId: "topic-nonexist",
		Title:   "新标题",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望帖子不存在错误，实际返回成功")
	}
}
