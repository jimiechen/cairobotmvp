package topic

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcListTopic 主题列表服务（minType=3005 GetTopicList）
type SvcListTopic struct {
	repo Repository
}

// NewSvcListTopic 创建服务实例
func NewSvcListTopic(repo Repository) *SvcListTopic {
	return &SvcListTopic{repo: repo}
}

// Handle 处理主题列表查询请求
func (s *SvcListTopic) Handle(ctx context.Context, req *pb.GetTopicListRequest) (*pb.GetTopicListResponse, error) {
	page := 1
	size := 20
	if req.Page > 0 {
		page = int(req.Page)
	}
	if req.PageSize > 0 {
		size = int(req.PageSize)
	}

	filters := make(map[string]interface{})
	if req.GroupId != "" {
		filters["group_id"] = req.GroupId
	}

	topics, total, err := s.repo.ListTopics(ctx, page, size, filters, "")
	if err != nil {
		return nil, fmt.Errorf("查询帖子列表失败: %w", err)
	}

	var pbTopics []*pb.TopicInfo
	for _, t := range topics {
		pbTopics = append(pbTopics, convertToProtoTopicInfo(t))
	}

	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	return &pb.GetTopicListResponse{
		Result:     &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "查询成功"},
		Topics:     pbTopics,
		Total:      total,
		Page:       int32(page),
		PageSize:   int32(size),
		TotalPages: int32(totalPages),
		GroupId:    req.GroupId,
	}, nil
}
