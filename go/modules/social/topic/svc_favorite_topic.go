package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcFavoriteTopic 收藏主题服务（minType=3061 FavoriteTopic）
type SvcFavoriteTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcFavoriteTopic 创建服务实例
func NewSvcFavoriteTopic(repo Repository, publisher event.Publisher) *SvcFavoriteTopic {
	return &SvcFavoriteTopic{repo: repo, publisher: publisher}
}

// Handle 处理收藏请求，支持幂等（已收藏则直接返回成功）
func (s *SvcFavoriteTopic) Handle(ctx context.Context, req *pb.FavoriteTopicRequest) (*pb.FavoriteTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)

	// 幂等检查：已收藏则直接返回
	alreadyFav, _ := s.repo.IsTopicFavorited(ctx, req.TopicId, userID)
	if alreadyFav {
		count, _ := s.repo.CountFavoritesByTopicID(ctx, req.TopicId)
		return &pb.FavoriteTopicResponse{
			Result:        &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "已收藏"},
			IsFavorited:   true,
			FavoriteCount: int32(count),
			TopicId:      req.TopicId,
		}, nil
	}

	fav := &TopicFavorite{
		ID:        fmt.Sprintf("fav-%d", time.Now().UnixNano()),
		TopicID:   req.TopicId,
		UserID:    userID,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.repo.CreateFavorite(ctx, fav); err != nil {
		return nil, fmt.Errorf("收藏失败: %w", err)
	}

	// 领域事件 — 发布 TopicReacted 事件（收藏）
	if s.publisher != nil {
		evt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventTopicReacted,
			AggregateType: event.AggregateTopic,
			AggregateID:   req.TopicId,
			ActorID:       userID,
			Payload: event.TopicReactedPayload{
				TopicID:      req.TopicId,
				UserID:       userID,
				ReactionType: event.ReactionTypeFavorite,
				Status:       1, // active
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 TopicReacted(favorite) 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, evt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 TopicReacted(favorite) 事件失败: %v\n", pubErr)
		}
	}

	count, _ := s.repo.CountFavoritesByTopicID(ctx, req.TopicId)
	return &pb.FavoriteTopicResponse{
		Result:        &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "收藏成功"},
		IsFavorited:   true,
		FavoriteCount: int32(count),
		TopicId:      req.TopicId,
	}, nil
}

// validateRequest 校验收藏请求必填字段
func (s *SvcFavoriteTopic) validateRequest(req *pb.FavoriteTopicRequest) (*pb.FavoriteTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.FavoriteTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
