package member

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetUserStats 获取用户统计信息（minType=1045 GetUserStats）
// 返回当前登录用户的统计数据（帖子数、评论数、点赞数等 2级数据）
// 不负责权限校验（用户只能查看自己的统计信息）
type SvcGetUserStats struct {
	repo Repository
}

// NewSvcGetUserStats 创建获取用户统计服务实例
func NewSvcGetUserStats(repo Repository) *SvcGetUserStats {
	return &SvcGetUserStats{repo: repo}
}

// Handle 处理获取用户统计请求
func (s *SvcGetUserStats) Handle(ctx context.Context, req *pb.GetUserStatsRequest) (*pb.GetUserStatsResponse, error) {
	// 从上下文提取当前登录用户 ID
	userID := ctx.Value(CtxKeyUserID)
	if userID == nil || userID.(string) == "" {
		return &pb.GetUserStatsResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "缺少用户身份信息"},
		}, nil
	}
	uidStr := userID.(string)

	// 获取或初始化用户统计记录（2级数据允许懒初始化）
	stats, err := s.repo.GetOrCreateStats(ctx, uidStr)
	if err != nil {
		return nil, fmt.Errorf("获取用户统计失败: %w", err)
	}

	pbStats := &pb.UserStats{
		UserId:         stats.UserID,
		TopicsCount:    int32(stats.TopicsCount),
		CommentsCount:  int32(stats.RepliesCount),
		LikesReceived:  int32(stats.LikesReceived),
		GroupsCount:    int32(stats.GroupsJoined),
	}

	return &pb.GetUserStatsResponse{
		Result:    &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "success"},
		UserStats: pbStats,
		UserId:    uidStr,
	}, nil
}
