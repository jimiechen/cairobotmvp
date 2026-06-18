package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcLikeTopic 点赞主题服务（minType=3065 LikeTopic）
type SvcLikeTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcLikeTopic 创建服务实例
func NewSvcLikeTopic(repo Repository, publisher event.Publisher) *SvcLikeTopic {
	return &SvcLikeTopic{repo: repo, publisher: publisher}
}

// Handle 处理点赞请求，支持幂等（已点赞则直接返回成功）
func (s *SvcLikeTopic) Handle(ctx context.Context, req *pb.LikeTopicRequest) (*pb.LikeTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)

	// 幂等检查：已点赞则直接返回
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
	if s.publisher != nil {
		evt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventTopicReacted,
			AggregateType: event.AggregateTopic,
			AggregateID:   req.TopicId,
			ActorID:       userID,
			Payload: event.TopicReactedPayload{
				TopicID:      req.TopicId,
				UserID:       userID,
				ReactionType: event.ReactionTypeLike,
				Status:       1, // active
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 TopicReacted(like) 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, evt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 TopicReacted(like) 事件失败: %v\n", pubErr)
		}
	}

	count, _ := s.repo.CountLikesByTopicID(ctx, req.TopicId)
	return &pb.LikeTopicResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "点赞成功"},
		IsLiked:   true,
		LikeCount: int32(count),
		TopicId:   req.TopicId,
	}, nil
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
