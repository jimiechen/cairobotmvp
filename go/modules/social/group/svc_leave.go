package group

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcLeave 退出圈子服务（minType=2015 LeaveGroup）
// 负责将用户从群组中移除（主动退出）
// 不负责群主退出拦截（由业务规则层处理）
type SvcLeave struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcLeave 创建服务实例
func NewSvcLeave(repo Repository, publisher event.Publisher) *SvcLeave {
	return &SvcLeave{repo: repo, publisher: publisher}
}

// Handle 处理退出圈子请求，遵循 DevGuide §7 五步模式
func (s *SvcLeave) Handle(ctx context.Context, req *pb.LeaveGroupRequest) (*pb.LeaveGroupResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// 从上下文获取当前用户 ID（proto LeaveGroupRequest 不含 user_id 字段）
	userID := getUserIDFromContext(ctx)

	// Step 2: 权限校验 — 检查是否为成员
	member, err := s.repo.GetMember(ctx, req.GroupId, userID)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.LeaveGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "您不是该圈子的成员",
			},
			GroupId: req.GroupId,
		}, nil
	}

	// 群主不允许直接退出
	if member.Role == GroupMemberRoleOwner {
		return &pb.LeaveGroupResponse{
			Result: &base.Result{
				Code:    int32(10731), // GROUP_ERROR_OWNER_CANNOT_LEAVE
				Message: "群主不能直接退出，请先转让所有权",
			},
			GroupId: req.GroupId,
		}, nil
	}

	// Step 3: 1级数据读写 — 更新成员状态为已退出
	member.Status = GroupMemberStatusLeft // 已退出（使用常量，禁止硬编码裸数字）
	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新成员状态失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 GroupLeft 事件
	if s.publisher != nil {
		leftEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventGroupLeft,
			AggregateType: event.AggregateGroup,
			AggregateID:   req.GroupId,
			ActorID:       userID,
			Payload: event.GroupLeftPayload{
				GroupID:  req.GroupId,
				UserID:   userID,
				MemberID: member.ID,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 GroupLeft 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, leftEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 GroupLeft 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.LeaveGroupResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "退出成功",
		},
		GroupId: req.GroupId,
	}, nil
}

// validateRequest 校验退出圈子请求必填字段
func (s *SvcLeave) validateRequest(req *pb.LeaveGroupRequest) (*pb.LeaveGroupResponse, error) {
	if req.GroupId == "" {
		return &pb.LeaveGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
