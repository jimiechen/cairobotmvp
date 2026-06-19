package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcLikeTopic_正常点赞 当is_like=true且topic_id合法时_应返回成功且is_liked=true
func TestSvcLikeTopic_正常点赞(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Status: TopicStatusInactive}
	svc := NewSvcLikeTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.LikeTopicRequest{TopicId: "topic-001", IsLike: true}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if !resp.IsLiked {
		t.Error("期望 IsLiked = true")
	}
}

// TestSvcLikeTopic_重复点赞 当已点赞时_应返回幂等成功
func TestSvcLikeTopic_重复点赞(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.topics["topic-001"] = &Topic{ID: "topic-001", Status: TopicStatusInactive}
	mockRepo.likes["user-001:topic-001"] = &TopicLike{TopicID: "topic-001", UserID: "user-001"}
	svc := NewSvcLikeTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.LikeTopicRequest{TopicId: "topic-001", IsLike: true}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望幂等成功 10200，实际得到 %d", resp.Result.Code)
	}
}

// TestSvcLikeTopic_正常取消点赞 当is_like=false且已有点赞记录时_应返回成功且is_liked=false
func TestSvcLikeTopic_正常取消点赞(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.likes["user-001:topic-001"] = &TopicLike{TopicID: "topic-001", UserID: "user-001"}
	svc := NewSvcLikeTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.LikeTopicRequest{TopicId: "topic-001", IsLike: false}

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

// TestSvcLikeTopic_缺少topic_id 当topic_id为空时_应返回参数校验错误
func TestSvcLikeTopic_缺少topic_id(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcLikeTopic(mockRepo, nil)

	req := &pb.LikeTopicRequest{TopicId: "", IsLike: true}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
