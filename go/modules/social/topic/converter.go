package topic

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

type contextKey string

const ctxKeyUserID contextKey = "user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func getUserIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return uid
	}
	return ""
}

func generateTopicID() string {
	return fmt.Sprintf("topic-%d", time.Now().UnixNano())
}

func generateReplyID() string {
	return fmt.Sprintf("reply-%d", time.Now().UnixNano())
}

func convertToProtoTopicInfo(t *Topic) *pb.TopicInfo {
	if t == nil { return nil }
	return &pb.TopicInfo{
		TopicId: t.ID, Title: t.Title, Content: t.Content,
		AuthorId: t.AuthorID, AuthorName: t.AuthorName, GroupId: t.GroupID,
		TopicType: pb.TopicType(t.TopicType), CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt, Summary: t.Summary,
		AuthorAvatar: t.AuthorAvatar, Visibility: pb.Visibility(t.Visibility),
		ContentFormat: pb.ContentFormat(t.ContentFormat), CoverImage: t.CoverImage,
	}
}

// convertToProtoReplyInfo 将内部 TopicReply 模型转换为 proto ReplyInfo
func convertToProtoReplyInfo(r *TopicReply) *pb.ReplyInfo {
	if r == nil { return nil }
	return &pb.ReplyInfo{
		ReplyId: r.ID, TopicId: r.TopicID, Content: r.Content,
		AuthorId: r.AuthorID, AuthorName: r.AuthorName,
		ParentReplyId: r.ParentReplyID, Status: pb.ReplyStatus(r.Status),
		LikeCount: int32(r.LikeCount), CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
