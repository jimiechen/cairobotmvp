package group

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcCalcPayableAmount 计算应付金额服务（minType=2073 CalcPayableAmount）
// 负责根据折扣配置计算用户应付金额
// 不负责支付流程（由支付网关处理）
type SvcCalcPayableAmount struct {
	repo Repository
}

// NewSvcCalcPayableAmount 创建服务实例
func NewSvcCalcPayableAmount(repo Repository) *SvcCalcPayableAmount {
	return &SvcCalcPayableAmount{repo: repo}
}

// Handle 处理计算应付金额请求，遵循 DevGuide §7 五步模式
func (s *SvcCalcPayableAmount) Handle(ctx context.Context, req *pb.CalcPayableAmountRequest) (*pb.CalcPayableAmountResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 公开查询，无需权限

	// Step 3: 1级数据读写 — 查询群组付费配置 → 计算折扣后金额
	group, err := s.repo.GetGroupByID(ctx, req.GroupId)
	if err != nil {
		return nil, fmt.Errorf("查询圈子失败: %w", err)
	}
	if group == nil {
		return &pb.CalcPayableAmountResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND),
				Message: "圈子不存在",
			},
			GroupId: req.GroupId,
		}, nil
	}

	// 查询付费配置
	payConfig, _ := s.repo.GetPayConfigByGroupID(ctx, req.GroupId)

	// 计算最终金额：原始金额 * 折扣率（MVP-P0 简化：无折扣时返回原价）
	finalAmount := req.OriginalAmount
	discountRate := 1.0
	if payConfig != nil && !payConfig.IsEnabled {
		// 付费功能未启用时返回原价（或提示错误）
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.CalcPayableAmountResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "成功",
		},
		FinalAmount:   finalAmount,
		DiscountRate:  discountRate,
		GroupId:       req.GroupId,
	}, nil
}

// validateRequest 校验计算金额请求必填字段
func (s *SvcCalcPayableAmount) validateRequest(req *pb.CalcPayableAmountRequest) (*pb.CalcPayableAmountResponse, error) {
	if req.GroupId == "" {
		return &pb.CalcPayableAmountResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	if req.OriginalAmount <= 0 {
		return &pb.CalcPayableAmountResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "原始金额必须大于0",
			},
		}, nil
	}
	return nil, nil
}
