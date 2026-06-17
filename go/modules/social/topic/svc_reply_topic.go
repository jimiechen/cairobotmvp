package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcReplyTopic 回复主题服务（minType=3043 AddTopicReply）
type SvcReplyTopic struct {
	repo Repository
}

// NewSvcReplyTopic 创建服务实例
func NewSvcReplyTopic(repo Repository) *SvcReplyTopic {
	return &SvcReplyTopic{repo: repo}
}

// Handle 处理回复主题请求
func (s *SvcReplyTopic) Handle(ctx context.Context, req *pb.AddTopicReplyRequest) (*pb.AddTopicReplyResponse, error) {
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	userID := getUserIDFromContext(ctx)
	now := time.Now().UnixMilli()
	reply := &TopicReply{
		ID:           generateReplyID(),
		TopicID:      req.TopicId,
		Content:      req.Content,
		AuthorID:     userID,
		ParentReplyID: req.ParentReplyId,
		Status:       ReplyStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateReply(ctx, reply); err != nil {
		return nil, fmt.Errorf("创建回复失败: %w", err)
	}

	return &pb.AddTopicReplyResponse{
		Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "回复成功"},
		ReplyInfo: convertToProtoReplyInfo(reply),
		ReplyId:  reply.ID,
	}, nil
}

// validateRequest 校验回复请求必填字段
func (s *SvcReplyTopic) validateRequest(req *pb.AddTopicReplyRequest) (*pb.AddTopicReplyResponse, error) {
	if req.TopicId == "" {
		return &pb.AddTopicReplyResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}
	if req.Content == "" {
		return &pb.AddTopicReplyResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "content 不能为空"},
		}, nil
	}
	return nil, nil
}
