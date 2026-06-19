package group

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetGroupStats 获取圈子统计信息（minType=2039 GetGroupStats）
// 返回指定圈子的统计数据（成员数、帖子数等 2级数据）
// 不负责权限校验（由上层确保调用者有查看圈子信息的权限）
type SvcGetGroupStats struct {
	repo Repository
}

// NewSvcGetGroupStats 创建获取圈子统计服务实例
func NewSvcGetGroupStats(repo Repository) *SvcGetGroupStats {
	return &SvcGetGroupStats{repo: repo}
}

// Handle 处理获取圈子统计请求
func (s *SvcGetGroupStats) Handle(ctx context.Context, req *pb.GetGroupStatsRequest) (*pb.GetGroupStatsResponse, error) {
	// 参数校验 — group_id 不能为空
	if req.GroupId == "" {
		return &pb.GetGroupStatsResponse{
			Result:  &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "group_id 不能为空"},
			GroupId: req.GroupId,
		}, nil
	}

	// 获取或初始化统计记录（2级数据允许懒初始化）
	stats, err := s.repo.GetOrCreateStats(ctx, req.GroupId)
	if err != nil {
		return nil, fmt.Errorf("获取圈子统计失败: %w", err)
	}

	pbStats := &pb.GroupStats{
		MembersCount: int32(stats.MembersCount),
		TopicsCount:  int32(stats.TopicsCount),
	}

	return &pb.GetGroupStatsResponse{
		Result:  &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "success"},
		Stats:   pbStats,
		GroupId: req.GroupId,
	}, nil
}
