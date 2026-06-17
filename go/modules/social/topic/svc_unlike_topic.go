package topic

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcUnlikeTopic 取消点赞主题服务（minType=3063 UnlikeTopic）
type SvcUnlikeTopic struct {
	repo Repository
}

// NewSvcUnlikeTopic 创建服务实例
func NewSvcUnlikeTopic(repo Repository) *SvcUnlikeTopic {
	return &SvcUnlikeTopic{repo: repo}
}

// Handle 处理取消点赞请求
func (s *SvcUnlikeTopic) Handle(ctx context.Context, req *pb.LikeTopicRequest) (*pb.LikeTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)

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

// validateRequest 校验取消点赞请求必填字段
func (s *SvcUnlikeTopic) validateRequest(req *pb.LikeTopicRequest) (*pb.LikeTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.LikeTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
