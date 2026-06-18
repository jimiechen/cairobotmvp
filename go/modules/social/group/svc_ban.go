package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcBanMember 封禁成员服务（minType=2023 BanMember）
// 负责对群组成员执行封禁操作（移除并标记为封禁状态）
// 不负责权限校验（由上层确保操作者有管理权限）
type SvcBanMember struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcBanMember 创建服务实例
func NewSvcBanMember(repo Repository, publisher event.Publisher) *SvcBanMember {
	return &SvcBanMember{repo: repo, publisher: publisher}
}

// Handle 处理封禁成员请求，遵循 DevGuide §7 五步模式
func (s *SvcBanMember) Handle(ctx context.Context, req *pb.BanMemberRequest) (*pb.BanMemberResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证操作者为管理员/群主

	// Step 3: 1级数据读写 — 查询目标成员 → 更新封禁状态
	member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.BanMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "目标用户不是该圈子的成员",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 不能封禁群主
	if member.Role == GroupMemberRoleOwner {
		return &pb.BanMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER),
				Message: "不能封禁圈子主",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 管理员不能互相封禁（MVP-P0 简化：仅警告不拦截）
	now := time.Now().UnixMilli()
	member.Status = GroupMemberStatusBanned // 已移除/封禁
	member.BannedAt = now
	member.BanExpiresAt = 0 // 永久封禁
	member.UpdatedAt = now

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新封禁状态失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 GroupMemberBanned 事件
	operatorID := getOperatorIDFromContext(ctx)
	if s.publisher != nil {
		banEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventGroupMemberBanned,
			AggregateType: event.AggregateGroup,
			AggregateID:   req.GroupId,
			ActorID:       operatorID,
			Payload: event.GroupMemberChangedPayload{
				GroupID:      req.GroupId,
				OperatorID:   operatorID,
				TargetUserID: req.UserId,
				Action:       event.ActionBan,
				OldStatus:    int32(GroupMemberStatusActive),
				NewStatus:    int32(GroupMemberStatusBanned),
				Reason:       "", // BanMemberRequest 无 Reason 字段，预留
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 GroupMemberBanned 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, banEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 GroupMemberBanned 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.BanMemberResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "封禁成功",
		},
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

// validateRequest 校验封禁请求必填字段
func (s *SvcBanMember) validateRequest(req *pb.BanMemberRequest) (*pb.BanMemberResponse, error) {
	if req.GroupId == "" {
		return &pb.BanMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	if req.UserId == "" {
		return &pb.BanMemberResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "目标用户ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
