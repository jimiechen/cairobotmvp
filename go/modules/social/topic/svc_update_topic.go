package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcUpdateTopic 更新主题服务（minType=3095 UpdateTopic）
type SvcUpdateTopic struct {
	repo Repository
}

// NewSvcUpdateTopic 创建服务实例
func NewSvcUpdateTopic(repo Repository) *SvcUpdateTopic {
	return &SvcUpdateTopic{repo: repo}
}

// Handle 处理更新主题请求
func (s *SvcUpdateTopic) Handle(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// 查询已存在的帖子
	existing, err := s.repo.GetTopicByID(ctx, req.TopicId)
	if err != nil {
		return nil, fmt.Errorf("查询帖子失败: %w", err)
	}
	if existing == nil {
		return &pb.CreateTopicResponse{
			Result: &base.Result{Code: 30301, Message: "帖子不存在"},
		}, nil
	}

	// 按需更新字段（仅更新非空字段）
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Content != "" {
		existing.Content = req.Content
	}
	if req.CoverImage != "" {
		existing.CoverImage = req.CoverImage
	}
	existing.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateTopic(ctx, existing); err != nil {
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}

	return &pb.CreateTopicResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "更新成功"},
		TopicInfo: convertToProtoTopicInfo(existing),
		TopicId:   existing.ID,
	}, nil
}

// validateRequest 校验更新请求必填字段
func (s *SvcUpdateTopic) validateRequest(req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.CreateTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
