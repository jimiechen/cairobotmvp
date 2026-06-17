package topic

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcDeleteTopic 删除主题服务（minType=3009 DeleteTopic）
type SvcDeleteTopic struct {
	repo Repository
}

// NewSvcDeleteTopic 创建服务实例
func NewSvcDeleteTopic(repo Repository) *SvcDeleteTopic {
	return &SvcDeleteTopic{repo: repo}
}

// Handle 处理删除主题请求
func (s *SvcDeleteTopic) Handle(ctx context.Context, req *pb.DeleteTopicRequest) (*pb.DeleteTopicResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	if err := s.repo.DeleteTopic(ctx, req.TopicId); err != nil {
		return nil, fmt.Errorf("删除帖子失败: %w", err)
	}

	return &pb.DeleteTopicResponse{
		Result:  &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "删除成功"},
		TopicId: req.TopicId,
	}, nil
}

// validateRequest 校验删除请求必填字段
func (s *SvcDeleteTopic) validateRequest(req *pb.DeleteTopicRequest) (*pb.DeleteTopicResponse, error) {
	if req.TopicId == "" {
		return &pb.DeleteTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	return nil, nil
}
