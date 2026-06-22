package group

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
)

// Handler 协议分发器，按 minType 路由到对应的 svc
// switch case 以外禁止业务逻辑
type Handler struct {
	// 领域事件发布器（可为 nil，nil 时不发布事件）
	publisher event.Publisher
	// 群组创建/加入/进入
	createSvc *SvcCreate
	joinSvc   *SvcJoin
	enterSvc  *SvcEnter
	// 群组成员管理
	leaveSvc       *SvcLeave
	renewSvc       *SvcRenewMember
	calcPayableSvc *SvcCalcPayableAmount
	muteSvc        *SvcMuteMember
	banSvc         *SvcBanMember
	removeSvc      *SvcRemoveMember
	updateRoleSvc  *SvcUpdateMemberRole
	// 群组查询（P1-E 补齐）
	batchGetGroupsSvc     *SvcBatchGetGroups
	getGroupStatsSvc      *SvcGetGroupStats
	getMemberUserIdsSvc   *SvcGetGroupMemberUserIds
}

// NewHandler 创建 Handler 实例，注入 Repository 并初始化所有 svc
func NewHandler(repo Repository, publisher event.Publisher) *Handler {
	return &Handler{
		publisher:      publisher,
		createSvc:      NewSvcCreate(repo, publisher),
		joinSvc:        NewSvcJoin(repo, publisher),
		enterSvc:       NewSvcEnter(repo),
		leaveSvc:       NewSvcLeave(repo, publisher),
		renewSvc:       NewSvcRenewMember(repo),
		calcPayableSvc: NewSvcCalcPayableAmount(repo),
		muteSvc:        NewSvcMuteMember(repo, publisher),
		banSvc:         NewSvcBanMember(repo, publisher),
		removeSvc:      NewSvcRemoveMember(repo, publisher),
		updateRoleSvc:  NewSvcUpdateMemberRole(repo),
		batchGetGroupsSvc:   NewSvcBatchGetGroups(repo),
		getGroupStatsSvc:    NewSvcGetGroupStats(repo),
		getMemberUserIdsSvc: NewSvcGetGroupMemberUserIds(repo),
	}
}

// Dispatch 根据 minType 分发到对应的 svc 处理
// 每个 case 统一：Unmarshal → svc.Handle → Marshal
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	switch minType {
	// 创建圈子（minType=2005）
	case "2005":
		var req pb.CreateGroupRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CreateGroupRequest failed: %w", err)
		}
		rsp, err := h.createSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 加入圈子（minType=2013）
	case "2013":
		var req pb.JoinGroupRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal JoinGroupRequest failed: %w", err)
		}
		rsp, err := h.joinSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 用户进入圈子（minType=2087）
	case "2087":
		var req pb.GroupUserEnterRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GroupUserEnterRequest failed: %w", err)
		}
		rsp, err := h.enterSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 退出圈子（minType=2015）
	case "2015":
		var req pb.LeaveGroupRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal LeaveGroupRequest failed: %w", err)
		}
		rsp, err := h.leaveSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 成员续费（minType=2037）
	case "2037":
		var req pb.RenewMemberRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal RenewMemberRequest failed: %w", err)
		}
		rsp, err := h.renewSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 计算应付金额（minType=2073）
	case "2073":
		var req pb.CalcPayableAmountRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal CalcPayableAmountRequest failed: %w", err)
		}
		rsp, err := h.calcPayableSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 禁言成员（minType=2019）
	case "2019":
		var req pb.MuteMemberRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal MuteMemberRequest failed: %w", err)
		}
		rsp, err := h.muteSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 封禁成员（minType=2023）
	case "2023":
		var req pb.BanMemberRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal BanMemberRequest failed: %w", err)
		}
		rsp, err := h.banSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 踢出成员（minType=2027）
	case "2027":
		var req pb.RemoveMemberRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal RemoveMemberRequest failed: %w", err)
		}
		rsp, err := h.removeSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 修改成员角色（minType=2029）
	case "2029":
		var req pb.UpdateMemberRoleRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UpdateMemberRoleRequest failed: %w", err)
		}
		rsp, err := h.updateRoleSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 获取圈子统计（minType=2039）
	case "2039":
		var req pb.GetGroupStatsRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetGroupStatsRequest failed: %w", err)
		}
		rsp, err := h.getGroupStatsSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 批量获取圈子信息（minType=2047）
	case "2047":
		var req pb.BatchGetGroupsRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal BatchGetGroupsRequest failed: %w", err)
		}
		rsp, err := h.batchGetGroupsSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 分页查询成员 UserID 列表（minType=2077）
	case "2077":
		var req pb.GetGroupMemberUserIdsRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetGroupMemberUserIdsRequest failed: %w", err)
		}
		rsp, err := h.getMemberUserIdsSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}

// InjectJWTManager 延迟注入 JWT 管理器（预留接口，Group 域当前不直接使用 JWT）
func (h *Handler) InjectJWTManager(m *member.JWTManager) {
	// Group 域 Handler 当前不需要直接持有 jwtMgr
	// 保留接口以保持三域 Servant 注入模式一致
}

// InjectTokenStore 延迟注入令牌黑名单存储（预留接口）
func (h *Handler) InjectTokenStore(ts member.TokenStore) {
	// Group 域 Handler 当前不需要直接持有 tokenStore
	// 黑名单检查在 Servant.Handle 层通过 social.CheckTokenBlacklist 完成
}
