package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
)

// context key 类型，避免冲突（仅 group 内部使用的 key）
type contextKey string

// ctxKeyOwnerID 上下文中存储群主用户 ID 的 key（创建圈子时使用）
const ctxKeyOwnerID contextKey = "owner_id"

// WithOwnerID 向上下文注入群主用户 ID（创建圈子时使用）
func WithOwnerID(ctx context.Context, ownerID string) context.Context {
	return context.WithValue(ctx, ctxKeyOwnerID, ownerID)
}

// getUserIDFromContext 从上下文获取当前用户 ID（使用 member 包统一常量）
func getUserIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(member.CtxKeyUserID).(string); ok {
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

// getOperatorIDFromContext 从上下文获取操作者用户 ID
// 用于高权限操作（封禁/移除/禁言）的事件 payload，语义上区分于普通用户
func getOperatorIDFromContext(ctx context.Context) string {
	return getUserIDFromContext(ctx)
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
