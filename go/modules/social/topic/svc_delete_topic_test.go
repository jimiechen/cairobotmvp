package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcDeleteTopic_正常删除 当topic_id合法时_应返回成功
func TestSvcDeleteTopic_正常删除(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Title: "待删帖", Status: TopicStatusInactive}

	svc := NewSvcDeleteTopic(mockRepo)

	req := &pb.DeleteTopicRequest{TopicId: "topic-001"}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.TopicId != "topic-001" {
		t.Errorf("期望透传 topic_id=topic-001，实际得到 %s", resp.TopicId)
	}
}

// TestSvcDeleteTopic_缺少topic_id 当topic_id为空时_应返回参数校验错误
func TestSvcDeleteTopic_缺少topic_id(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcDeleteTopic(mockRepo)

	req := &pb.DeleteTopicRequest{TopicId: ""}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
