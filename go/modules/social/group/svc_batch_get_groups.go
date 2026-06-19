package group

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcBatchGetGroups 批量获取圈子信息（minType=2047 BatchGetGroups）
// 根据一组 group_id 列表返回对应的圈子基本信息
// 不负责权限校验（由上层根据圈子可见性过滤结果）
type SvcBatchGetGroups struct {
	repo Repository
}

// NewSvcBatchGetGroups 创建批量获取圈子服务实例
func NewSvcBatchGetGroups(repo Repository) *SvcBatchGetGroups {
	return &SvcBatchGetGroups{repo: repo}
}

// Handle 处理批量获取圈子请求
func (s *SvcBatchGetGroups) Handle(ctx context.Context, req *pb.BatchGetGroupsRequest) (*pb.BatchGetGroupsResponse, error) {
	// 参数校验 — group_ids 不能为空
	if len(req.GroupIds) == 0 {
		return &pb.BatchGetGroupsResponse{
			Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "group_ids 不能为空"},
			GroupIds: req.GroupIds,
		}, nil
	}

	// 查询圈子列表（使用足够大的分页参数以覆盖所有请求的 ID）
	groups, _, err := s.repo.ListGroups(ctx, 1, len(req.GroupIds), nil)
	if err != nil {
		return nil, err
	}

	// 构建 ID → Group 映射，按请求顺序返回结果
	groupMap := make(map[string]*Group)
	for _, g := range groups {
		groupMap[g.ID] = g
	}

	var pbGroups []*pb.GroupInfo
	for _, gid := range req.GroupIds {
		if g, ok := groupMap[gid]; ok {
			pbGroups = append(pbGroups, convertToProtoGroupInfo(g))
		}
	}

	return &pb.BatchGetGroupsResponse{
		Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "success"},
		Groups:   pbGroups,
		GroupIds: req.GroupIds,
	}, nil
}
