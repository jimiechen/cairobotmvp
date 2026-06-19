package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// TestSvcCreateTopic_创建帖子成功_应发布TopicCreated事件 验证创建帖子成功后发布 TopicCreated 领域事件
func TestSvcCreateTopic_创建帖子成功_应发布TopicCreated事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	publisher := &event.FakePublisher{}
	svc := NewSvcCreateTopic(mockRepo, publisher)
	ctx := WithUserID(context.Background(), "user-event-topic-001")

	req := &pb.CreateTopicRequest{
		Title:   "事件测试帖子标题",
		Content: "事件测试帖子内容",
		GroupId: "group-event-topic-001",
	}

	// Act
	resp, err := svc.Handle(ctx, req)

	// Assert — 业务结果正确
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Fatalf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}

	// Assert — 事件已发布
	if publisher.EventCount() != 1 {
		t.Fatalf("期望发布 1 个事件，实际发布 %d 个", publisher.EventCount())
	}

	publishedEvt := publisher.Events()[0]
	if publishedEvt.Type != event.EventTopicCreated {
		t.Errorf("期望事件类型 %s，实际得到 %s", event.EventTopicCreated, publishedEvt.Type)
	}
	if publishedEvt.AggregateID != resp.TopicId {
		t.Errorf("期望 AggregateID 为 topic ID %s，实际得到 %s", resp.TopicId, publishedEvt.AggregateID)
	}
	if publishedEvt.ActorID != "user-event-topic-001" {
		t.Errorf("期望 ActorID 为 user-event-topic-001，实际得到 %s", publishedEvt.ActorID)
	}
}

// TestSvcCreateTopic_nil_publisher_不发布事件 验证 publisher 为 nil 时不 panic 且业务正常
func TestSvcCreateTopic_nil_publisher_不发布事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcCreateTopic(mockRepo, nil) // nil publisher
	ctx := WithUserID(context.Background(), "user-nil-topic-001")

	req := &pb.CreateTopicRequest{
		Title:   "NilPub帖子",
		Content: "内容",
		GroupId: "group-nil-topic-001",
	}

	// Act — 不应 panic
	resp, err := svc.Handle(ctx, req)

	// Assert — 业务正常完成
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Fatalf("期望成功码 10200，实际得到 %d", resp.Result.Code)
	}
}
