package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcFavoriteTopic_正常收藏 当is_favorite=true且topic_id合法时_应返回成功且is_favorited=true
func TestSvcFavoriteTopic_正常收藏(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Status: TopicStatusInactive}
	svc := NewSvcFavoriteTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.FavoriteTopicRequest{TopicId: "topic-001", IsFavorite: true}

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
	svc := NewSvcFavoriteTopic(mockRepo, nil)

	req := &pb.FavoriteTopicRequest{TopicId: "", IsFavorite: true}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}

// TestSvcFavoriteTopic_已收藏用户取消收藏 当is_favorite=false时_应返回成功且is_favorited=false
func TestSvcFavoriteTopic_已收藏_取消成功(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.favorites["user-001:topic-001"] = &TopicFavorite{TopicID: "topic-001", UserID: "user-001"}
	svc := NewSvcFavoriteTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.FavoriteTopicRequest{TopicId: "topic-001", IsFavorite: false}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.IsFavorited {
		t.Error("期望 IsFavorited = false")
	}
}

// TestSvcFavoriteTopic_未收藏用户取消收藏 应幂等返回成功
func TestSvcFavoriteTopic_未收藏_幂等成功(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcFavoriteTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.FavoriteTopicRequest{TopicId: "topic-001", IsFavorite: false}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.IsFavorited {
		t.Error("期望 IsFavorited = false（未收藏状态）")
	}
}
