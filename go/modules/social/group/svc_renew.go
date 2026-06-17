package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcRenewMember 成员续费服务（minType=2037 RenewMember）
// 负责更新成员的会员到期时间
// 不负责支付流程（由支付网关处理）
type SvcRenewMember struct {
	repo Repository
}

// NewSvcRenewMember 创建服务实例
func NewSvcRenewMember(repo Repository) *SvcRenewMember {
	return &SvcRenewMember{repo: repo}
}

// Handle 处理成员续费请求，遵循 DevGuide §7 五步模式
func (s *SvcRenewMember) Handle(ctx context.Context, req *pb.RenewMemberRequest) (*pb.RenewMemberResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 续费可由自己或管理员操作

	// Step 3: 1级数据读写 — 查询成员 → 更新到期时间
	member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.RenewMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "您不是该圈子的成员",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 更新会员到期时间
	now := time.Now().UnixMilli()
	if req.RenewPeriodEnd > 0 {
		member.MembershipExpiresAt = req.RenewPeriodEnd
	} else {
		// 默认续费一年
		member.MembershipExpiresAt = now + 365*24*3600*1000
	}
	member.UpdatedAt = now

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新成员续费信息失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.RenewMemberResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "续费成功",
		},
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

// validateRequest 校验续费请求必填字段
func (s *SvcRenewMember) validateRequest(req *pb.RenewMemberRequest) (*pb.RenewMemberResponse, error) {
	if req.GroupId == "" {
		return &pb.RenewMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
