package group

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/internal/dispatch"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
)

// Handler 协议分发器，按 minType 路由到对应的 svc
// 使用 dispatch.ProtoRouter 消除重复的 Unmarshal→Handle→Marshal 样板代码
type Handler struct {
	// 领域事件发布器（可为 nil，nil 时不发布事件）
	publisher event.Publisher
	// 协议路由器
	router *dispatch.ProtoRouter
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

// NewHandler 创建 Handler 实例，注入 Repository 并初始化所有 svc 和协议路由
func NewHandler(repo Repository, publisher event.Publisher) *Handler {
	h := &Handler{
		publisher:      publisher,
		router:         dispatch.NewProtoRouter(),
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
	h.registerRoutes()
	return h
}

// registerRoutes 注册所有 Group 域协议路由到 ProtoRouter
func (h *Handler) registerRoutes() {
	// 创建圈子（minType=2005）
	dispatch.Register(h.router, "2005", h.createSvc)
	// 加入圈子（minType=2013）
	dispatch.Register(h.router, "2013", h.joinSvc)
	// 用户进入圈子（minType=2087）
	dispatch.Register(h.router, "2087", h.enterSvc)
	// 退出圈子（minType=2015）
	dispatch.Register(h.router, "2015", h.leaveSvc)
	// 成员续费（minType=2037）
	dispatch.Register(h.router, "2037", h.renewSvc)
	// 计算应付金额（minType=2073）
	dispatch.Register(h.router, "2073", h.calcPayableSvc)
	// 禁言成员（minType=2019）
	dispatch.Register(h.router, "2019", h.muteSvc)
	// 封禁成员（minType=2023）
	dispatch.Register(h.router, "2023", h.banSvc)
	// 踢出成员（minType=2027）
	dispatch.Register(h.router, "2027", h.removeSvc)
	// 修改成员角色（minType=2029）
	dispatch.Register(h.router, "2029", h.updateRoleSvc)
	// 获取圈子统计（minType=2039）
	dispatch.Register(h.router, "2039", h.getGroupStatsSvc)
	// 批量获取圈子信息（minType=2047）
	dispatch.Register(h.router, "2047", h.batchGetGroupsSvc)
	// 分页查询成员 UserID 列表（minType=2077）
	dispatch.Register(h.router, "2077", h.getMemberUserIdsSvc)
}

// Dispatch 根据 minType 分发到对应的 svc 处理（委托给 ProtoRouter）
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	return h.router.Dispatch(ctx, minType, reqBytes)
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
