package group

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
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

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}
