package member

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcUpdateMemberStatus 更新用户状态服务（minType=1033 UpdateMemberStatus）
// 负责管理员更新用户账号状态（正常/封禁/注销等）
// 不负责权限校验（由上层确保操作者有管理权限）
// 范围：状态持久化 + 基本校验 + 幂等 + UserStatusChanged 事件发布
type SvcUpdateMemberStatus struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcUpdateMemberStatus 创建更新用户状态服务实例
func NewSvcUpdateMemberStatus(repo Repository, publisher event.Publisher) *SvcUpdateMemberStatus {
	return &SvcUpdateMemberStatus{repo: repo, publisher: publisher}
}

// Handle 处理更新用户状态请求，遵循 DevGuide §7 五步模式
func (s *SvcUpdateMemberStatus) Handle(ctx context.Context, req *pb.UpdateMemberStatusRequest) (*pb.UpdateMemberStatusResponse, error) {
	// Step 1: 参数校验 — 必填字段 + 状态值合法性
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证操作者有管理权限

	// Step 3: 1级数据读写 — 查询目标用户 → 校验存在性 → 幂等检查 → 更新状态
	targetUser, err := s.repo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询目标用户失败: %w", err)
	}
	if targetUser == nil {
		return &pb.UpdateMemberStatusResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "目标用户不存在",
			},
			UserId: req.UserId,
		}, nil
	}

	targetStatus := int8(req.Status)
	oldStatus := targetUser.Status

	// 幂等检查：当前状态与目标状态一致时，直接返回成功（不写 DB）
	if oldStatus == targetStatus {
		return &pb.UpdateMemberStatusResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_SUCCESS), Message: "状态未变更"},
			UserId: req.UserId,
			Status: int32(targetStatus),
		}, nil
	}

	// 更新用户状态
	now := time.Now().UnixMilli()
	targetUser.Status = targetStatus
	targetUser.UpdatedAt = now

	if err := s.repo.UpdateUser(ctx, targetUser); err != nil {
		return nil, fmt.Errorf("更新用户状态失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 UserStatusChanged 事件
	operatorID := s.getOperatorIDFromContext(ctx)
	if s.publisher != nil {
		statusEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventUserStatusChanged,
			AggregateType: event.AggregateMember,
			AggregateID:   req.UserId,
			ActorID:       operatorID,
			Payload: event.UserStatusChangedPayload{
				UserID:     req.UserId,
				OldStatus:  int32(oldStatus),
				NewStatus:  int32(targetStatus),
				OperatorID: operatorID,
				Reason:     req.Reason,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 UserStatusChanged 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, statusEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 UserStatusChanged 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.UpdateMemberStatusResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "状态更新成功",
		},
		UserId: req.UserId,
		Status: int32(targetStatus),
	}, nil
}

// validateRequest 校验更新状态请求必填字段和状态值合法性
func (s *SvcUpdateMemberStatus) validateRequest(req *pb.UpdateMemberStatusRequest) (*pb.UpdateMemberStatusResponse, error) {
	if req.UserId == "" {
		return &pb.UpdateMemberStatusResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "user_id 不能为空"},
			UserId: req.UserId,
		}, nil
	}

	// 状态值必须为合法枚举值：1-正常活跃 2-未激活 3-封禁 4-已注销
	targetStatus := int8(req.Status)
	switch targetStatus {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended, UserStatusDeleted:
		// 合法状态值
	default:
		return &pb.UpdateMemberStatusResponse{
			Result: &base.Result{Code: int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST), Message: "无效的状态值"},
			UserId: req.UserId,
		}, nil
	}

	return nil, nil
}

// getOperatorIDFromContext 从上下文获取操作者用户 ID
// 用于事件 payload 中记录谁执行了状态变更操作
func (s *SvcUpdateMemberStatus) getOperatorIDFromContext(ctx context.Context) string {
	if uid, ok := ctx.Value(CtxKeyUserID).(string); ok {
		return uid
	}
	return ""
}
