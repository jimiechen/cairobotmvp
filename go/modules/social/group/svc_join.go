package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcJoin 加入圈子服务（minType=2013 JoinGroup）
// 负责将用户加入群组为成员
// 不负责邀请码验证（MVP-P0 简化）
type SvcJoin struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcJoin 创建服务实例
func NewSvcJoin(repo Repository, publisher event.Publisher) *SvcJoin {
	return &SvcJoin{repo: repo, publisher: publisher}
}

// Handle 处理加入圈子请求，遵循 DevGuide §7 五步模式
func (s *SvcJoin) Handle(ctx context.Context, req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 加入为公开操作，无需权限检查

	// 从上下文获取当前用户 ID（proto JoinGroupRequest 不含 user_id 字段）
	userID := getUserIDFromContext(ctx)

	// Step 3: 1级数据读写 — 检查群组存在性 + 检查是否已成员 → 创建成员记录
	group, err := s.repo.GetGroupByID(ctx, req.GroupId)
	if err != nil {
		return nil, fmt.Errorf("查询圈子失败: %w", err)
	}
	if group == nil {
		return &pb.JoinGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NOT_FOUND),
				Message: "圈子不存在",
			},
			GroupId: req.GroupId,
		}, nil
	}

	// 检查是否已是成员（幂等检查）
	isMember, err := s.repo.IsUserMember(ctx, req.GroupId, userID)
	if err != nil {
		return nil, fmt.Errorf("检查成员状态失败: %w", err)
	}
	if isMember {
		return &pb.JoinGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ALREADY_MEMBER),
				Message: "您已经是该圈子的成员",
			},
			GroupId: req.GroupId,
		}, nil
	}

	now := time.Now().UnixMilli()
	member := &GroupMember{
		ID:          generateMemberID(),
		GroupID:     req.GroupId,
		UserID:      userID,
		Role:        GroupMemberRoleMember,   // 普通成员
		Status:      GroupMemberStatusActive, // 正常
		JoinReason:  req.JoinReason,
		JoinedAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("创建成员记录失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 GroupJoined 事件
	if s.publisher != nil {
		joinEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventGroupJoined,
			AggregateType: event.AggregateGroup,
			AggregateID:   req.GroupId,
			ActorID:       userID,
			Payload: event.GroupJoinedPayload{
				GroupID:   req.GroupId,
				UserID:    userID,
				MemberID:  member.ID,
				Status:    GroupMemberStatusActive,
				JoinSource: "direct",
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 GroupJoined 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, joinEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 GroupJoined 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.JoinGroupResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "加入成功",
		},
		MemberId: member.ID,
		Status:   pb.JoinStatus_JOIN_STATUS_JOINED,
		GroupId:  req.GroupId,
	}, nil
}

// validateRequest 校验加入圈子请求必填字段
func (s *SvcJoin) validateRequest(req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	if req.GroupId == "" {
		return &pb.JoinGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_ID_EMPTY),
				Message: "圈子ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
