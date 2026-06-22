package member

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/internal/dispatch"
)

// Handler 协议分发器，按 minType 路由到对应的 svc
// 使用 dispatch.ProtoRouter 消除重复的 Unmarshal→Handle→Marshal 样板代码
type Handler struct {
	// 用户 Repository（延迟注入 JWTManager 时需要重建 svc）
	repo Repository
	// 领域事件发布器（可为 nil，nil 时不发布事件）
	publisher event.Publisher
	// 协议路由器
	router *dispatch.ProtoRouter
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

// NewHandler 创建 Handler 实例，注入所有 svc 依赖并注册协议路由
func NewHandler(repo Repository, publisher event.Publisher, opts ...HandlerOption) *Handler {
	h := &Handler{
		repo:              repo,
		publisher:         publisher,
		router:            dispatch.NewProtoRouter(),
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
	h.registerRoutes()
	return h
}

// registerRoutes 注册所有 Member 域协议路由到 ProtoRouter
func (h *Handler) registerRoutes() {
	// 用户注册（minType=1021）
	dispatch.Register(h.router, "1021", h.registerSvc)
	// 用户登录（minType=1023）
	dispatch.Register(h.router, "1023", h.loginSvc)
	// 用户登出（minType=1025）
	dispatch.Register(h.router, "1025", h.logoutSvc)
	// 令牌刷新（minType=1027）
	dispatch.Register(h.router, "1027", h.refreshSvc)
	// 查询用户信息（minType=1029）
	dispatch.Register(h.router, "1029", h.getUserInfoSvc)
	// 更新用户信息（minType=1031）
	dispatch.Register(h.router, "1031", h.updateUserInfoSvc)
	// 拉黑用户（minType=1039）
	dispatch.Register(h.router, "1039", h.blockSvc)
	// 取消拉黑（minType=1041）
	dispatch.Register(h.router, "1041", h.unblockSvc)
	// 查询拉黑列表（minType=1043）
	dispatch.Register(h.router, "1043", h.getBlockListSvc)
	// 更新用户状态（minType=1033）
	dispatch.Register(h.router, "1033", h.updateMemberStatusSvc)
	// 获取用户统计（minType=1045）
	dispatch.Register(h.router, "1045", h.getUserStatsSvc)
}

// InjectJWTManager 延迟注入 JWT 管理器并重建依赖它的 svc
// 用于解决 Module 创建时 JWT 依赖尚未就绪的循环依赖问题
func (h *Handler) InjectJWTManager(m *JWTManager) {
	h.jwtManager = m
	// 重建依赖 JWTManager 的三个 svc（login/logout/refresh）
	h.loginSvc = NewSvcLogin(h.repo, h.jwtManager)
	h.logoutSvc = NewSvcLogout(h.tokenStore, h.jwtManager)
	h.refreshSvc = NewSvcRefresh(h.tokenStore, h.jwtManager, h.repo)
	// 重新注册路由（svc 引用已更新）
	h.registerRoutes()
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
	// 重新注册路由（svc 引用已更新）
	h.registerRoutes()
}

// Dispatch 根据 minType 分发到对应的 svc 处理（委托给 ProtoRouter）
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	return h.router.Dispatch(ctx, minType, reqBytes)
}
