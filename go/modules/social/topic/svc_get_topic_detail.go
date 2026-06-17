package topic

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetTopicDetail 获取主题详情服务（minType=3081 GetTopicDetail）
type SvcGetTopicDetail struct {
	repo Repository
}

// NewSvcGetTopicDetail 创建服务实例
func NewSvcGetTopicDetail(repo Repository) *SvcGetTopicDetail {
	return &SvcGetTopicDetail{repo: repo}
}

// Handle 处理获取主题详情请求，支持批量查询
func (s *SvcGetTopicDetail) Handle(ctx context.Context, req *pb.BatchGetTopicInfoRequest) (*pb.BatchGetTopicInfoResponse, error) {
	if len(req.TopicIds) == 0 {
		return &pb.BatchGetTopicInfoResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_ids 不能为空"},
		}, nil
	}

	var topics []*pb.TopicInfo
	for _, id := range req.TopicIds {
		t, err := s.repo.GetTopicByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("查询帖子详情失败: %w", err)
		}
		if t != nil {
			topics = append(topics, convertToProtoTopicInfo(t))
		}
	}

	return &pb.BatchGetTopicInfoResponse{
		Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "查询成功"},
		Topics:   topics,
		TopicIds: req.TopicIds,
	}, nil
}
