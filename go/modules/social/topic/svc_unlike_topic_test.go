package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcUnlikeTopic_正常取消点赞 当已有点赞记录时_应返回成功且is_liked=false
func TestSvcUnlikeTopic_正常取消点赞(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.likes["user-001:topic-001"] = &TopicLike{TopicID: "topic-001", UserID: "user-001"}
	svc := NewSvcUnlikeTopic(mockRepo)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.LikeTopicRequest{TopicId: "topic-001"}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.IsLiked {
		t.Error("期望 IsLiked = false")
	}
}

// TestSvcUnlikeTopic_缺少topic_id 当topic_id为空时_应返回参数校验错误
func TestSvcUnlikeTopic_缺少topic_id(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcUnlikeTopic(mockRepo)

	req := &pb.LikeTopicRequest{TopicId: ""}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
