package member

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
	// 用户注册/登录
	registerSvc *SvcRegister
	loginSvc    *SvcLogin
	// 用户登出/令牌
	logoutSvc   *SvcLogout
	refreshSvc  *SvcRefresh
	// 用户信息查询/更新
	getUserInfoSvc    *SvcGetUserInfo
	updateUserInfoSvc *SvcUpdateUserInfo
	// 用户屏蔽管理
	blockSvc      *SvcBlock
	unblockSvc    *SvcUnblock
	getBlockListSvc *SvcGetBlockList
}

// NewHandler 创建 Handler 实例，注入所有 svc 依赖
func NewHandler(repo Repository, publisher event.Publisher) *Handler {
	return &Handler{
		publisher:         publisher,
		registerSvc:      NewSvcRegister(repo, publisher),
		loginSvc:         NewSvcLogin(repo),
		logoutSvc:        NewSvcLogout(),
		refreshSvc:       NewSvcRefresh(),
		getUserInfoSvc:   NewSvcGetUserInfo(repo),
		updateUserInfoSvc: NewSvcUpdateUserInfo(repo),
		blockSvc:         NewSvcBlock(repo),
		unblockSvc:       NewSvcUnblock(repo),
		getBlockListSvc:  NewSvcGetBlockList(repo),
	}
}

// Dispatch 根据 minType 分发到对应的 svc 处理
// 每个 case 统一：Unmarshal → svc.Handle → Marshal
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	switch minType {
	// 用户注册（minType=1021）
	case "1021":
		var req pb.UserRegisterRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UserRegisterRequest failed: %w", err)
		}
		rsp, err := h.registerSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 用户登录（minType=1023）
	case "1023":
		var req pb.UserLoginRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UserLoginRequest failed: %w", err)
		}
		rsp, err := h.loginSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 用户登出（minType=1025）
	case "1025":
		var req pb.UserLogoutRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UserLogoutRequest failed: %w", err)
		}
		rsp, err := h.logoutSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 令牌刷新（minType=1027）
	case "1027":
		var req pb.RefreshTokenRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal RefreshTokenRequest failed: %w", err)
		}
		rsp, err := h.refreshSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 查询用户信息（minType=1029）
	case "1029":
		var req pb.GetUserInfoRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetUserInfoRequest failed: %w", err)
		}
		rsp, err := h.getUserInfoSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 更新用户信息（minType=1031）
	case "1031":
		var req pb.UpdateUserInfoRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UpdateUserInfoRequest failed: %w", err)
		}
		rsp, err := h.updateUserInfoSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 拉黑用户（minType=1039）
	case "1039":
		var req pb.BlockUserRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal BlockUserRequest failed: %w", err)
		}
		rsp, err := h.blockSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 取消拉黑（minType=1041）
	case "1041":
		var req pb.UnblockUserRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UnblockUserRequest failed: %w", err)
		}
		rsp, err := h.unblockSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 查询拉黑列表（minType=1043）
	case "1043":
		var req pb.GetBlockListRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetBlockListRequest failed: %w", err)
		}
		rsp, err := h.getBlockListSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}
