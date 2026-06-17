package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcCreateTopic 创建主题服务（minType=3001 CreateTopic）
// 负责新帖子创建流程：参数校验 → 构建模型 → 持久化 → 返回结果
type SvcCreateTopic struct {
	repo Repository
}

// NewSvcCreateTopic 创建服务实例
func NewSvcCreateTopic(repo Repository) *SvcCreateTopic {
	return &SvcCreateTopic{repo: repo}
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

	// Step 4: 领域事件 — MVP-P0 可延迟

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
