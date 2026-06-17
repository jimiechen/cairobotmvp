package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcReadTopic_正常记录已读 当topic_id合法时_应返回成功
func TestSvcReadTopic_正常记录已读(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcReadTopic(mockRepo)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.CheckTopicActionsRequest{TopicId: "topic-001"}

	resp, err := svc.Handle(ctx, req)
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

// TestSvcReadTopic_缺少topic_id 当topic_id为空时_应返回参数校验错误
func TestSvcReadTopic_缺少topic_id(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcReadTopic(mockRepo)

	req := &pb.CheckTopicActionsRequest{TopicId: ""}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
