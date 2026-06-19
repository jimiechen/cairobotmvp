package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcLikeTopic 点赞/取消点赞主题服务（minType=3061 LikeTopic）
// base proto 定义：3061=LikeTopicRequest(奇数), 3062=LikeTopicResponse(偶数)
// 通过 req.IsLike 字段区分操作：true=点赞, false=取消点赞
type SvcLikeTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcLikeTopic 创建服务实例
func NewSvcLikeTopic(repo Repository, publisher event.Publisher) *SvcLikeTopic {
	return &SvcLikeTopic{repo: repo, publisher: publisher}
}

// Handle 处理点赞/取消点赞请求
// req.IsLike=true 表示点赞，req.IsLike=false 表示取消点赞
func (s *SvcLikeTopic) Handle(ctx context.Context, req *pb.LikeTopicRequest) (*pb.LikeTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)

	// 取消点赞分支（is_like=false）
	if !req.GetIsLike() {
		return s.handleUnlike(ctx, req, userID)
	}

	// 点赞分支（is_like=true）：幂等检查
	alreadyLiked, _ := s.repo.IsTopicLiked(ctx, req.TopicId, userID)
	if alreadyLiked {
		count, _ := s.repo.CountLikesByTopicID(ctx, req.TopicId)
		return &pb.LikeTopicResponse{
			Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "已点赞"},
			IsLiked:   true,
			LikeCount: int32(count),
			TopicId:   req.TopicId,
		}, nil
	}

	like := &TopicLike{
		ID:        fmt.Sprintf("like-%d", time.Now().UnixNano()),
		TopicID:   req.TopicId,
		UserID:    userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.repo.CreateLike(ctx, like); err != nil {
		return nil, fmt.Errorf("点赞失败: %w", err)
	}

	// 领域事件 — 发布 TopicReacted 事件（点赞）
	s.publishReactedEvent(ctx, req.TopicId, userID, event.ReactionTypeLike)

	count, _ := s.repo.CountLikesByTopicID(ctx, req.TopicId)
	return &pb.LikeTopicResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "点赞成功"},
		IsLiked:   true,
		LikeCount: int32(count),
		TopicId:   req.TopicId,
	}, nil
}

// handleUnlike 处理取消点赞逻辑：删除点赞记录并返回 is_liked=false
func (s *SvcLikeTopic) handleUnlike(ctx context.Context, req *pb.LikeTopicRequest, userID string) (*pb.LikeTopicResponse, error) {
	if err := s.repo.DeleteLike(ctx, req.TopicId, userID); err != nil {
		return nil, fmt.Errorf("取消点赞失败: %w", err)
	}
	count, _ := s.repo.CountLikesByTopicID(ctx, req.TopicId)
	return &pb.LikeTopicResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "取消点赞成功"},
		IsLiked:   false,
		LikeCount: int32(count),
		TopicId:   req.TopicId,
	}, nil
}

// publishReactedEvent 发布 TopicReacted 领域事件
func (s *SvcLikeTopic) publishReactedEvent(ctx context.Context, topicID, userID string, reactionType string) {
	if s.publisher == nil {
		return
	}
	evt, err := event.NewDomainEvent(event.NewEventOptions{
		Type:          event.EventTopicReacted,
		AggregateType: event.AggregateTopic,
		AggregateID:   topicID,
		ActorID:       userID,
		Payload: event.TopicReactedPayload{
			TopicID:      topicID,
			UserID:       userID,
			ReactionType: reactionType,
			Status:       1, // active
		},
	})
	if err != nil {
		fmt.Printf("[EVENT] 构造 TopicReacted 事件失败: %v\n", err)
	} else if pubErr := s.publisher.Publish(ctx, evt); pubErr != nil {
		fmt.Printf("[EVENT] 发布 TopicReacted 事件失败: %v\n", pubErr)
	}
}

// validateRequest 校验点赞请求必填字段
func (s *SvcLikeTopic) validateRequest(req *pb.LikeTopicRequest) (*pb.LikeTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.LikeTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
