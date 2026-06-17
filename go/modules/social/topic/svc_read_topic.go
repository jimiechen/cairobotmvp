package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcReadTopic 已读主题服务（minType=3099 ReadTopic）
type SvcReadTopic struct {
	repo Repository
}

// NewSvcReadTopic 创建服务实例
func NewSvcReadTopic(repo Repository) *SvcReadTopic {
	return &SvcReadTopic{repo: repo}
}

// Handle 处理已读记录请求
func (s *SvcReadTopic) Handle(ctx context.Context, req *pb.CheckTopicActionsRequest) (*pb.CheckTopicActionsResponse, error) {
	if req.TopicId == "" {
		return &pb.CheckTopicActionsResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}

	userID := getUserIDFromContext(ctx)
	now := time.Now().UnixMilli()
	read := &TopicRead{
		ID:      fmt.Sprintf("read-%d", time.Now().UnixNano()),
		TopicID: req.TopicId,
		UserID:  userID,
		ReadAt:  now,
	}
	if err := s.repo.UpsertReadRecord(ctx, read); err != nil {
		return nil, fmt.Errorf("记录已读失败: %w", err)
	}

	return &pb.CheckTopicActionsResponse{
		Result:          &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "已读记录成功"},
		TopicId:         req.TopicId,
		AvailableActions: []string{},
	}, nil
}
