package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcReplyTopic_正常回复 当topic_id和content合法时_应返回成功包含reply_id
func TestSvcReplyTopic_正常回复(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Status: TopicStatusInactive}
	svc := NewSvcReplyTopic(mockRepo)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.AddTopicReplyRequest{
		TopicId: "topic-001",
		Content: "这是评论内容",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.ReplyId == "" {
		t.Error("期望 ReplyId 不为空")
	}
}

// TestSvcReplyTopic_缺少content 当content为空时_应返回参数校验错误
func TestSvcReplyTopic_缺少content(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcReplyTopic(mockRepo)

	req := &pb.AddTopicReplyRequest{
		TopicId: "topic-001",
		Content: "",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
