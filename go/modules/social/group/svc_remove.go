package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcRemoveMember 踢出成员服务（minType=2027 RemoveMember）
// 负责将成员从群组中移除（管理员主动踢出）
// 不负责权限校验（由上层确保操作者有管理权限）
type SvcRemoveMember struct {
	repo Repository
}

// NewSvcRemoveMember 创建服务实例
func NewSvcRemoveMember(repo Repository) *SvcRemoveMember {
	return &SvcRemoveMember{repo: repo}
}

// Handle 处理踢出成员请求，遵循 DevGuide §7 五步模式
func (s *SvcRemoveMember) Handle(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证操作者为管理员/群主

	// Step 3: 1级数据读写 — 查询目标成员 → 更新为已移除状态
	member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.RemoveMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "目标用户不是该圈子的成员",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 不能移除群主
	if member.Role == GroupMemberRoleOwner {
		return &pb.RemoveMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_REMOVE_OWNER),
				Message: "不能移除圈子主",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	now := time.Now().UnixMilli()
	member.Status = 3 // 已移除
	member.BanReason = req.Reason
	member.UpdatedAt = now

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新成员状态失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.RemoveMemberResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "踢出成功",
		},
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

// validateRequest 校验踢出请求必填字段
func (s *SvcRemoveMember) validateRequest(req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	if req.GroupId == "" {
		return &pb.RemoveMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	if req.UserId == "" {
		return &pb.RemoveMemberResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "目标用户ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
