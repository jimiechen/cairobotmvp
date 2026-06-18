package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcReplyTopic 回复主题服务（minType=3043 AddTopicReply）
type SvcReplyTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcReplyTopic 创建服务实例
func NewSvcReplyTopic(repo Repository, publisher event.Publisher) *SvcReplyTopic {
	return &SvcReplyTopic{repo: repo, publisher: publisher}
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

	// 领域事件 — 发布 TopicCommentCreated 事件
	if s.publisher != nil {
		// 从 topic 查询 GroupID（AddTopicReplyRequest 不携带 group_id）
		groupID := ""
		if t, err := s.repo.GetTopicByID(ctx, reply.TopicID); err == nil && t != nil {
			groupID = t.GroupID
		}
		evt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventTopicCommentCreated,
			AggregateType: event.AggregateComment,
			AggregateID:   reply.ID,
			ActorID:       userID,
			Payload: event.TopicCommentCreatedPayload{
				CommentID: reply.ID,
				TopicID:   reply.TopicID,
				GroupID:   groupID,
				UserID:    userID,
				ParentID:  reply.ParentReplyID,
				Status:    reply.Status,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 TopicCommentCreated 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, evt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 TopicCommentCreated 事件失败: %v\n", pubErr)
		}
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
