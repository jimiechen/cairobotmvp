package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcFavoriteTopic_正常收藏 当topic_id合法时_应返回成功且is_favorited=true
func TestSvcFavoriteTopic_正常收藏(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Status: TopicStatusInactive}
	svc := NewSvcFavoriteTopic(mockRepo)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.FavoriteTopicRequest{TopicId: "topic-001"}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if !resp.IsFavorited {
		t.Error("期望 IsFavorited = true")
	}
}

// TestSvcFavoriteTopic_缺少topic_id 当topic_id为空时_应返回参数校验错误
func TestSvcFavoriteTopic_缺少topic_id(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcFavoriteTopic(mockRepo)

	req := &pb.FavoriteTopicRequest{TopicId: ""}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
