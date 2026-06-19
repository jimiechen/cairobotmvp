package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcFavoriteTopic 收藏/取消收藏主题服务（minType=3063 FavoriteTopic）
// base proto 定义：3063=FavoriteTopicRequest(奇数), 3064=FavoriteTopicResponse(偶数)
// 通过 req.IsFavorite 字段区分操作：true=收藏, false=取消收藏
type SvcFavoriteTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcFavoriteTopic 创建服务实例
func NewSvcFavoriteTopic(repo Repository, publisher event.Publisher) *SvcFavoriteTopic {
	return &SvcFavoriteTopic{repo: repo, publisher: publisher}
}

// Handle 处理收藏/取消收藏请求
// req.IsFavorite=true 表示收藏，req.IsFavorite=false 表示取消收藏
func (s *SvcFavoriteTopic) Handle(ctx context.Context, req *pb.FavoriteTopicRequest) (*pb.FavoriteTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)

	// 取消收藏分支（is_favorite=false）
	if !req.GetIsFavorite() {
		return s.handleUnfavorite(ctx, req, userID)
	}

	// 收藏分支（is_favorite=true）：幂等检查
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
	s.publishReactedEvent(ctx, req.TopicId, userID, event.ReactionTypeFavorite)

	count, _ := s.repo.CountFavoritesByTopicID(ctx, req.TopicId)
	return &pb.FavoriteTopicResponse{
		Result:        &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "收藏成功"},
		IsFavorited:   true,
		FavoriteCount: int32(count),
		TopicId:      req.TopicId,
	}, nil
}

// handleUnfavorite 处理取消收藏逻辑：幂等检查 → 删除收藏记录
func (s *SvcFavoriteTopic) handleUnfavorite(ctx context.Context, req *pb.FavoriteTopicRequest, userID string) (*pb.FavoriteTopicResponse, error) {
	// 幂等：未收藏则直接返回成功
	alreadyFav, _ := s.repo.IsTopicFavorited(ctx, req.TopicId, userID)
	if !alreadyFav {
		count, _ := s.repo.CountFavoritesByTopicID(ctx, req.TopicId)
		return &pb.FavoriteTopicResponse{
			Result:        &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "未收藏"},
			IsFavorited:   false,
			FavoriteCount: int32(count),
			TopicId:      req.TopicId,
		}, nil
	}
	if err := s.repo.DeleteFavorite(ctx, req.TopicId, userID); err != nil {
		return nil, fmt.Errorf("取消收藏失败: %w", err)
	}
	count, _ := s.repo.CountFavoritesByTopicID(ctx, req.TopicId)
	return &pb.FavoriteTopicResponse{
		Result:        &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "取消收藏成功"},
		IsFavorited:   false,
		FavoriteCount: int32(count),
		TopicId:      req.TopicId,
	}, nil
}

// publishReactedEvent 发布 TopicReacted 领域事件
func (s *SvcFavoriteTopic) publishReactedEvent(ctx context.Context, topicID, userID string, reactionType string) {
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

// validateRequest 校验收藏请求必填字段
func (s *SvcFavoriteTopic) validateRequest(req *pb.FavoriteTopicRequest) (*pb.FavoriteTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.FavoriteTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
