package group

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcGetGroupMemberUserIds 分页查询成员 UserID 列表（minType=2077 GetGroupMemberUserIds）
// 返回指定圈子的成员用户 ID 列表，用于批量查询用户信息等场景
// 不负责权限校验（由上层确保调用者有查看成员列表的权限）
type SvcGetGroupMemberUserIds struct {
	repo Repository
}

// NewSvcGetGroupMemberUserIds 创建分页查询成员 UserID 服务实例
func NewSvcGetGroupMemberUserIds(repo Repository) *SvcGetGroupMemberUserIds {
	return &SvcGetGroupMemberUserIds{repo: repo}
}

// Handle 处理分页查询成员 UserID 请求
func (s *SvcGetGroupMemberUserIds) Handle(ctx context.Context, req *pb.GetGroupMemberUserIdsRequest) (*pb.GetGroupMemberUserIdsResponse, error) {
	// 参数校验 — group_id 不能为空
	if req.GroupId == "" {
		return &pb.GetGroupMemberUserIdsResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "group_id 不能为空"},
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
	if pageSize > 100 {
		pageSize = 100
	}

	// 查询成员列表
	members, total, err := s.repo.ListMembers(ctx, req.GroupId, page, pageSize, nil, nil)
	if err != nil {
		return nil, err
	}

	// 提取 UserID 列表
	var userIDs []string
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	return &pb.GetGroupMemberUserIdsResponse{
		Result:   &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "success"},
		UserIds:  userIDs,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
		GroupId:  req.GroupId,
	}, nil
}
