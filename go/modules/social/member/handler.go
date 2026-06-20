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
	// 用户 Repository（延迟注入 JWTManager 时需要重建 svc）
	repo Repository
	// 领域事件发布器（可为 nil，nil 时不发布事件）
	publisher event.Publisher
	// JWT 令牌管理器
	jwtManager *JWTManager
	// 令牌黑名单存储
	tokenStore TokenStore
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
	// 用户状态/统计查询（P1-E 补齐）
	updateMemberStatusSvc *SvcUpdateMemberStatus
	getUserStatsSvc      *SvcGetUserStats
}

// HandlerOption Handler 可选配置函数
type HandlerOption func(*Handler)

// WithJWTManager 注入 JWT 管理器
func WithJWTManager(m *JWTManager) HandlerOption {
	return func(h *Handler) {
		h.jwtManager = m
	}
}

// WithTokenStore 注入令牌黑名单存储
func WithTokenStore(ts TokenStore) HandlerOption {
	return func(h *Handler) {
		h.tokenStore = ts
	}
}

// NewHandler 创建 Handler 实例，注入所有 svc 依赖
func NewHandler(repo Repository, publisher event.Publisher, opts ...HandlerOption) *Handler {
	h := &Handler{
		repo:              repo,
		publisher:         publisher,
		registerSvc:      NewSvcRegister(repo, publisher),
		loginSvc:         NewSvcLogin(repo, nil), // 延迟：通过 opts 注入 jwtManager
		logoutSvc:        NewSvcLogout(nil, nil), // 延迟：通过 opts 注入 tokenStore + jwtManager
		refreshSvc:       NewSvcRefresh(nil, nil, repo), // 延迟：通过 opts 注入 tokenStore + jwtManager
		getUserInfoSvc:   NewSvcGetUserInfo(repo),
		updateUserInfoSvc: NewSvcUpdateUserInfo(repo),
		blockSvc:         NewSvcBlock(repo),
		unblockSvc:       NewSvcUnblock(repo),
		getBlockListSvc:  NewSvcGetBlockList(repo),
		updateMemberStatusSvc: NewSvcUpdateMemberStatus(repo, publisher),
		getUserStatsSvc:      NewSvcGetUserStats(repo),
	}
	for _, opt := range opts {
		opt(h)
	}
	// 重新创建依赖 JWTManager/TokenStore 的 svc
	if h.jwtManager != nil {
		h.loginSvc = NewSvcLogin(repo, h.jwtManager)
		h.logoutSvc = NewSvcLogout(h.tokenStore, h.jwtManager)
		h.refreshSvc = NewSvcRefresh(h.tokenStore, h.jwtManager, repo)
	}
	return h
}

// InjectJWTManager 延迟注入 JWT 管理器并重建依赖它的 svc
// 用于解决 Module 创建时 JWT 依赖尚未就绪的循环依赖问题
func (h *Handler) InjectJWTManager(m *JWTManager) {
	h.jwtManager = m
	// 重建依赖 JWTManager 的三个 svc（login/logout/refresh）
	h.loginSvc = NewSvcLogin(h.repo, h.jwtManager)
	h.logoutSvc = NewSvcLogout(h.tokenStore, h.jwtManager)
	h.refreshSvc = NewSvcRefresh(h.tokenStore, h.jwtManager, h.repo)
}

// InjectTokenStore 延迟注入令牌黑名单存储并重建依赖它的 svc
// 用于解决 Module 创建时 Redis 依赖尚未就绪的循环依赖问题
func (h *Handler) InjectTokenStore(ts TokenStore) {
	h.tokenStore = ts
	// 重建依赖 TokenStore 的两个 svc（logout/refresh）
	if h.jwtManager != nil {
		h.logoutSvc = NewSvcLogout(h.tokenStore, h.jwtManager)
		h.refreshSvc = NewSvcRefresh(h.tokenStore, h.jwtManager, h.repo)
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

	// 更新用户状态（minType=1033）
	case "1033":
		var req pb.UpdateMemberStatusRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal UpdateMemberStatusRequest failed: %w", err)
		}
		rsp, err := h.updateMemberStatusSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	// 获取用户统计（minType=1045）
	case "1045":
		var req pb.GetUserStatsRequest
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal GetUserStatsRequest failed: %w", err)
		}
		rsp, err := h.getUserStatsSvc.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(rsp)

	default:
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
}
