package group

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// SvcCreate 创建圈子服务（minType=2005 CreateGroup）
// 负责创建群组记录并将创建者加入为群主
// 不负责权限校验（创建为公开操作），不负责付费配置校验（MVP-P0 简化）
type SvcCreate struct {
	repo      Repository
	publisher event.Publisher
}

// NewSvcCreate 创建服务实例
func NewSvcCreate(repo Repository, publisher event.Publisher) *SvcCreate {
	return &SvcCreate{repo: repo, publisher: publisher}
}

// Handle 处理创建圈子请求，遵循 DevGuide §7 五步模式
func (s *SvcCreate) Handle(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	// Step 1: 参数校验 — 必填字段非空
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 创建圈子为公开操作，无需权限检查

	// Step 3: 1级数据读写 — 检查 slug 唯一性 → 创建群组记录
	existing, _ := s.repo.GetGroupBySlug(ctx, req.Slug)
	if existing != nil {
		return &pb.CreateGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NAME_ALREADY_EXISTS),
				Message: "圈子标识符已被占用",
			},
		}, nil
	}

	now := time.Now().UnixMilli()
	ownerID := getOwnerIDFromContext(ctx)
	group := &Group{
		ID:          generateGroupID(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Category:    req.Category,
		OwnerID:     ownerID,
		Status:      GroupStatusActive,      // 正常
		Visibility:  GroupVisibilityPublic, // 公开
		JoinMode:    int8(req.JoinMode),
		MaxMembers:  500,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, fmt.Errorf("创建圈子失败: %w", err)
	}

	// 将创建者自动加入为群主（role=1=owner），确保权限链路正确
	ownerMember := &GroupMember{
		ID:        generateMemberID(),
		GroupID:   group.ID,
		UserID:    ownerID,
		Role:   GroupMemberRoleOwner,    // 群主
		Status: GroupMemberStatusActive, // 正常
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateMember(ctx, ownerMember); err != nil {
		return nil, fmt.Errorf("创建群主成员记录失败: %w", err)
	}

	// Step 4: 领域事件 — 发布 GroupCreated 事件
	if s.publisher != nil {
		groupEvt, err := event.NewDomainEvent(event.NewEventOptions{
			Type:          event.EventGroupCreated,
			AggregateType: event.AggregateGroup,
			AggregateID:   group.ID,
			ActorID:       ownerID,
			Payload: event.GroupCreatedPayload{
				GroupID:    group.ID,
				OwnerID:    ownerID,
				Type:       "free",
				Visibility: group.Visibility,
				JoinMode:   group.JoinMode,
			},
		})
		if err != nil {
			fmt.Printf("[EVENT] 构造 GroupCreated 事件失败: %v\n", err)
		} else if pubErr := s.publisher.Publish(ctx, groupEvt); pubErr != nil {
			fmt.Printf("[EVENT] 发布 GroupCreated 事件失败: %v\n", pubErr)
		}
	}

	// Step 5: 返回响应
	return &pb.CreateGroupResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "创建成功",
		},
		GroupId:   group.ID,
		GroupInfo: convertToProtoGroupInfo(group),
	}, nil
}

// validateRequest 校验创建圈子请求必填字段
func (s *SvcCreate) validateRequest(req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	if req.Name == "" {
		return &pb.CreateGroupResponse{
			Result: &base.Result{
				Code:    int32(base.GroupErrorCode_GROUP_ERROR_NAME_EMPTY),
				Message: "圈子名称不能为空",
			},
		}, nil
	}
	if req.Slug == "" {
		return &pb.CreateGroupResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "圈子标识符不能为空",
			},
		}, nil
	}
	return nil, nil
}
