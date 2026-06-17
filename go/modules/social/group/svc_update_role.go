package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcUpdateMemberRole 修改成员角色服务（minType=2029 UpdateMemberRole）
// 负责修改群组成员的角色（如提升为管理员、降级为普通成员）
// 不负责权限校验（由上层确保操作者为群主）
type SvcUpdateMemberRole struct {
	repo Repository
}

// NewSvcUpdateMemberRole 创建服务实例
func NewSvcUpdateMemberRole(repo Repository) *SvcUpdateMemberRole {
	return &SvcUpdateMemberRole{repo: repo}
}

// Handle 处理修改成员角色请求，遵循 DevGuide §7 五步模式
func (s *SvcUpdateMemberRole) Handle(ctx context.Context, req *pb.UpdateMemberRoleRequest) (*pb.UpdateMemberRoleResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证操作者为群主

	// Step 3: 1级数据读写 — 查询目标成员 → 更新角色
	member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.UpdateMemberRoleResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "目标用户不是该圈子的成员",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 角色合法性检查：不允许设为群主（群主角色只能通过转让）
	if req.Role == pb.GroupMemberRole_GROUP_ROLE_OWNER {
		return &pb.UpdateMemberRoleResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_INVALID_ROLE),
				Message: "无法通过此接口设置群主角色",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	now := time.Now().UnixMilli()
	member.Role = int8(req.Role)
	member.UpdatedAt = now

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新成员角色失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.UpdateMemberRoleResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "角色修改成功",
		},
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

// validateRequest 校验修改角色请求必填字段
func (s *SvcUpdateMemberRole) validateRequest(req *pb.UpdateMemberRoleRequest) (*pb.UpdateMemberRoleResponse, error) {
	if req.GroupId == "" {
		return &pb.UpdateMemberRoleResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	if req.UserId == "" {
		return &pb.UpdateMemberRoleResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "目标用户ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
