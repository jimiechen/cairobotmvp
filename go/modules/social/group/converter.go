package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// context key 类型，避免冲突
type contextKey string

const (
	// ctxKeyUserID 上下文中存储当前用户 ID 的 key
	ctxKeyUserID contextKey = "user_id"
	// ctxKeyOwnerID 上下文中存储群主用户 ID 的 key（创建圈子时使用）
	ctxKeyOwnerID contextKey = "owner_id"
)

// WithUserID 向上下文注入当前用户 ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

// WithOwnerID 向上下文注入群主用户 ID
func WithOwnerID(ctx context.Context, ownerID string) context.Context {
	return context.WithValue(ctx, ctxKeyOwnerID, ownerID)
}

// getUserIDFromContext 从上下文获取当前用户 ID
func getUserIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return uid
	}
	return ""
}

// getOwnerIDFromContext 从上下文获取群主用户 ID
func getOwnerIDFromContext(ctx context.Context) string {
	if oid, ok := ctx.Value(ctxKeyOwnerID).(string); ok {
		return oid
	}
	return ""
}

// generateGroupID 生成群组内部 ID（MVP-P0 使用时间戳+随机数格式）
func generateGroupID() string {
	return fmt.Sprintf("group-%d", time.Now().UnixNano())
}

// generateMemberID 生成成员关系内部 ID
func generateMemberID() string {
	return fmt.Sprintf("member-%d", time.Now().UnixNano())
}

// convertToProtoGroupInfo 将内部 Group 模型转换为 proto GroupInfo
func convertToProtoGroupInfo(g *Group) *pb.GroupInfo {
	if g == nil {
		return nil
	}
	return &pb.GroupInfo{
		GroupId:     g.ID,
		Name:        g.Name,
		Slug:        g.Slug,
		Description: g.Description,
		Avatar:      g.Avatar,
		CoverImage:  g.CoverImage,
		Category:    g.Category,
		OwnerId:     g.OwnerID,
		Status:      pb.GroupStatus(g.Status),
		Visibility:  pb.GroupVisibility(g.Visibility),
		JoinMode:    pb.JoinMode(g.JoinMode),
		IsOfficial:  g.IsOfficial,
		IsFeatured:  g.IsFeatured,
		MaxMembers:  int32(g.MaxMembers),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// convertToProtoUserMemberInfo 将内部 GroupMember 模型转换为 proto UserMemberInfo
func convertToProtoUserMemberInfo(m *GroupMember) *pb.UserMemberInfo {
	if m == nil {
		return nil
	}
	return &pb.UserMemberInfo{
		UserId:                 m.UserID,
		Role:                   pb.GroupMemberRole(m.Role),
		Status:                 pb.MemberStatus(m.Status),
		JoinedAt:               m.JoinedAt,
		LastActivityAt:         m.LastActivityAt,
		MembershipExpiresAt:    m.MembershipExpiresAt,
		QuestionQuotaRemaining: int32(m.RemainingQuota),
	}
}
