package member

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcUnblock 取消拉黑服务（minType=1041 UnblockUser）
// 负责删除用户之间的拉黑关系（单向解除）
// 不负责权限校验（由上层保证 unblocked_by 为当前登录用户）
type SvcUnblock struct {
	repo Repository
}

// NewSvcUnblock 创建取消拉黑服务实例
func NewSvcUnblock(repo Repository) *SvcUnblock {
	return &SvcUnblock{repo: repo}
}

// Handle 处理取消拉黑请求，遵循 DevGuide §7 五步模式
func (s *SvcUnblock) Handle(ctx context.Context, req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证 unblocked_by 为当前登录用户

	// Step 3: 1级数据读写 — 删除拉黑记录（幂等：不存在也成功）
	if err := s.repo.DeleteBlock(ctx, req.UnblockedBy, req.UserId); err != nil {
		return nil, fmt.Errorf("删除拉黑记录失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.UnblockUserResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "取消拉黑成功",
		},
		UserId: req.UserId,
	}, nil
}

// validateRequest 校验取消拉黑请求必填字段
func (s *SvcUnblock) validateRequest(req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
	if req.UserId == "" {
		return &pb.UnblockUserResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "被解封用户ID不能为空",
			},
		}, nil
	}
	if req.GroupId == "" {
		return &pb.UnblockUserResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
