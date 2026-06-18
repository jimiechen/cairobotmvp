package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcMuteMember 禁言成员服务（minType=2019 MuteMember）
// 负责对群组成员执行禁言操作
// 不负责权限校验（由上层确保操作者有管理权限），不负责通知被禁言者
type SvcMuteMember struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcMuteMember 创建服务实例
func NewSvcMuteMember(repo Repository, publisher event.Publisher) *SvcMuteMember {
	return &SvcMuteMember{repo: repo, publisher: publisher}
}

// Handle 处理禁言成员请求，遵循 DevGuide §7 五步模式
func (s *SvcMuteMember) Handle(ctx context.Context, req *pb.MuteMemberRequest) (*pb.MuteMemberResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 由上层保证操作者为管理员/群主

	// Step 3: 1级数据读写 — 查询目标成员 → 更新禁言状态
	member, err := s.repo.GetMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("查询成员信息失败: %w", err)
	}
	if member == nil {
		return &pb.MuteMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_MEMBER),
				Message: "目标用户不是该圈子的成员",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 不能禁言群主
	if member.Role == GroupMemberRoleOwner {
		return &pb.MuteMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_CANNOT_BAN_OWNER),
				Message: "不能禁言圈子主",
			},
			GroupId: req.GroupId,
			UserId:  req.UserId,
		}, nil
	}

	// 计算禁言到期时间
	now := time.Now().UnixMilli()
	mutedUntil := calcMutedUntil(now, req.MuteDuration)

	member.MutedUntil = mutedUntil
	member.Status = GroupMemberStatusMuted // 已禁言
	member.UpdatedAt = now

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("更新禁言状态失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 GroupMemberMuted 事件
	operatorID := getOperatorIDFromContext(ctx)
	if s.publisher != nil {
		muteEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventGroupMemberMuted,
			AggregateType: event.AggregateGroup,
			AggregateID:   req.GroupId,
			ActorID:       operatorID,
			Payload: event.GroupMemberChangedPayload{
				GroupID:      req.GroupId,
				OperatorID:   operatorID,
				TargetUserID: req.UserId,
				Action:       event.ActionMute,
				MutedUntil:   mutedUntil,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 GroupMemberMuted 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, muteEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 GroupMemberMuted 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.MuteMemberResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "禁言成功",
		},
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

// validateRequest 校验禁言请求必填字段
func (s *SvcMuteMember) validateRequest(req *pb.MuteMemberRequest) (*pb.MuteMemberResponse, error) {
	if req.GroupId == "" {
		return &pb.MuteMemberResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	if req.UserId == "" {
		return &pb.MuteMemberResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "目标用户ID不能为空",
			},
		}, nil
	}
	return nil, nil
}

// calcMutedUntil 根据禁言时长枚举计算到期时间戳（毫秒）
func calcMutedUntil(nowMs int64, duration pb.MuteDuration) int64 {
	switch duration {
	case pb.MuteDuration_MUTE_DURATION_1_HOUR:
		return nowMs + int64(time.Hour/time.Millisecond)
	case pb.MuteDuration_MUTE_DURATION_4_HOURS:
		return nowMs + 4*int64(time.Hour/time.Millisecond)
	case pb.MuteDuration_MUTE_DURATION_12_HOURS:
		return nowMs + 12*int64(time.Hour/time.Millisecond)
	case pb.MuteDuration_MUTE_DURATION_1_DAY:
		return nowMs + 24*int64(time.Hour/time.Millisecond)
	case pb.MuteDuration_MUTE_DURATION_7_DAYS:
		return nowMs + 7*24*int64(time.Hour/time.Millisecond)
	default:
		return nowMs + int64(time.Hour/time.Millisecond) // 默认1小时
	}
}
