package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcCreateTopic 创建主题服务（minType=3001 CreateTopic）
// 负责新帖子创建流程：参数校验 → 构建模型 → 持久化 → 返回结果
type SvcCreateTopic struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcCreateTopic 创建服务实例
func NewSvcCreateTopic(repo Repository, publisher event.Publisher) *SvcCreateTopic {
	return &SvcCreateTopic{repo: repo, publisher: publisher}
}

// Handle 处理创建主题请求，遵循 DevGuide §7 五步模式
func (s *SvcCreateTopic) Handle(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	// Step 1: 参数校验
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — MVP-P0 公开操作，无需额外权限

	// Step 3: 1级数据读写 — 构建模型并持久化
	userID := getUserIDFromContext(ctx)
	now := time.Now().UnixMilli()
	topic := &Topic{
		ID:            generateTopicID(),
		Title:         req.Title,
		Content:       req.Content,
		AuthorID:      userID,
		GroupID:       req.GroupId,
		TopicType:     int8(req.TopicType),
		ContentFormat: int8(req.ContentFormat),
		Status:        TopicStatusInactive, // 已发布（对应模型值 2）
		CoverImage:    req.CoverImage,
		Summary:       req.Summary,
		DraftID:       req.DraftId,
		QuestionText:  req.QuestionText,
		QAPrivate:     req.QaPrivate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateTopic(ctx, topic); err != nil {
		return nil, fmt.Errorf("创建帖子失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 TopicCreated 事件
	if s.publisher != nil {
		evt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventTopicCreated,
			AggregateType: event.AggregateTopic,
			AggregateID:   topic.ID,
			ActorID:       userID,
			Payload: event.TopicCreatedPayload{
				TopicID:    topic.ID,
				GroupID:    topic.GroupID,
				AuthorID:   topic.AuthorID,
				Status:     topic.Status,
				Visibility: topic.Visibility,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 TopicCreated 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, evt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 TopicCreated 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.CreateTopicResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "创建成功",
		},
		TopicInfo: convertToProtoTopicInfo(topic),
		TopicId:   topic.ID,
		DraftId:   req.DraftId,
	}, nil
}

// validateRequest 校验创建请求必填字段
func (s *SvcCreateTopic) validateRequest(req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	if req.Title == "" {
		return &pb.CreateTopicResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "标题不能为空"},
		}, nil
	}
	return nil, nil
}
