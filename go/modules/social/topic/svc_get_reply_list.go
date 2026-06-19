package topic

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetReplyList 获取帖子评论列表（minType=3065 GetReplyList）
// 分页返回指定帖子的评论/回复列表，支持按父评论 ID 过滤获取子回复
// 不负责权限校验（由上层根据帖子可见性和评论状态过滤结果）
type SvcGetReplyList struct {
	repo Repository
}

// NewSvcGetReplyList 创建获取评论列表服务实例
func NewSvcGetReplyList(repo Repository) *SvcGetReplyList {
	return &SvcGetReplyList{repo: repo}
}

// Handle 处理获取评论列表请求
func (s *SvcGetReplyList) Handle(ctx context.Context, req *pb.GetReplyListRequest) (*pb.GetReplyListResponse, error) {
	// 参数校验 — topic_id 不能为空
	if req.TopicId == "" {
		return &pb.GetReplyListResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "topic_id 不能为空"},
		}, nil
	}

	// 分页参数标准化
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	// 父评论 ID 过滤（非空时仅返回该父评论下的子回复）
	var parentReplyID *string
	if req.ParentReplyId != "" {
		parentReplyID = &req.ParentReplyId
	}

	// 查询评论列表
	replies, total, err := s.repo.ListReplies(ctx, req.TopicId, page, pageSize, parentReplyID)
	if err != nil {
		return nil, err
	}

	// 转换为 proto 评论列表
	var pbReplies []*pb.ReplyInfo
	for _, r := range replies {
		pbReplies = append(pbReplies, &pb.ReplyInfo{
			ReplyId:         r.ID,
			TopicId:         r.TopicID,
			Content:         r.Content,
			AuthorId:        r.AuthorID,
			AuthorName:      r.AuthorName,
			ParentReplyId:   r.ParentReplyID,
			Status:          pb.ReplyStatus(r.Status),
			LikeCount:       int32(r.LikeCount),
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
			RepliesCount:    int32(r.RepliesCount),
			IsPinned:        r.IsPinned,
			Level:           int32(r.Level),
			ReplyToUserId:   r.ReplyToUserID,
			ReplyToUserName: r.ReplyToUserName,
		})
	}

	return &pb.GetReplyListResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "success"},
		Replies:   pbReplies,
		Total:     total,
		Page:      int32(page),
		PageSize: int32(pageSize),
		TopicId:  req.TopicId,
	}, nil
}
