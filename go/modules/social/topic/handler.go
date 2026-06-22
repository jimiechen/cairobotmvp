package topic

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/internal/dispatch"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
)

// Handler 协议分发器，按 minType 路由到对应的 svc
// 使用 dispatch.ProtoRouter 消除重复的 Unmarshal→Handle→Marshal 样板代码
type Handler struct {
	publisher        event.Publisher
	router           *dispatch.ProtoRouter
	createTopicSvc   *SvcCreateTopic
	listTopicSvc     *SvcListTopic
	deleteTopicSvc    *SvcDeleteTopic
	replyTopicSvc     *SvcReplyTopic
	likeTopicSvc      *SvcLikeTopic
	favoriteTopicSvc  *SvcFavoriteTopic
	getTopicDetailSvc *SvcGetTopicDetail
	updateTopicSvc    *SvcUpdateTopic
	createReportSvc   *SvcCreateReport
	readTopicSvc      *SvcReadTopic
	getReplyListSvc   *SvcGetReplyList
}

// NewHandler 创建 Handler 实例，注入所有 svc 依赖并注册协议路由
func NewHandler(repo Repository, publisher event.Publisher) *Handler {
	h := &Handler{
		publisher:        publisher,
		router:           dispatch.NewProtoRouter(),
		createTopicSvc:   NewSvcCreateTopic(repo, publisher),
		listTopicSvc:     NewSvcListTopic(repo),
		deleteTopicSvc:   NewSvcDeleteTopic(repo),
		replyTopicSvc:    NewSvcReplyTopic(repo, publisher),
		likeTopicSvc:     NewSvcLikeTopic(repo, publisher),
		favoriteTopicSvc: NewSvcFavoriteTopic(repo, publisher),
		getTopicDetailSvc: NewSvcGetTopicDetail(repo),
		updateTopicSvc:   NewSvcUpdateTopic(repo),
		createReportSvc:  NewSvcCreateReport(repo),
		readTopicSvc:     NewSvcReadTopic(repo),
		getReplyListSvc:  NewSvcGetReplyList(repo),
	}
	h.registerRoutes()
	return h
}

// registerRoutes 注册所有 Topic 域协议路由到 ProtoRouter
func (h *Handler) registerRoutes() {
	// 创建主题（minType=3001）
	dispatch.Register(h.router, "3001", h.createTopicSvc)
	// 主题列表（minType=3005）
	dispatch.Register(h.router, "3005", h.listTopicSvc)
	// 删除主题（minType=3009）
	dispatch.Register(h.router, "3009", h.deleteTopicSvc)
	// 回复主题（minType=3043）
	dispatch.Register(h.router, "3043", h.replyTopicSvc)
	// 点赞/取消点赞主题（minType=3061, 通过 is_like 区分）
	dispatch.Register(h.router, "3061", h.likeTopicSvc)
	// 收藏/取消收藏主题（minType=3063, 通过 is_favorite 区分）
	dispatch.Register(h.router, "3063", h.favoriteTopicSvc)
	// 主题详情/批量查询（minType=3057）
	dispatch.Register(h.router, "3057", h.getTopicDetailSvc)
	// 提交举报（minType=3095）
	dispatch.Register(h.router, "3095", h.createReportSvc)
	// 已读主题（minType=3099）
	dispatch.Register(h.router, "3099", h.readTopicSvc)
	// 获取评论列表（minType=3065）
	dispatch.Register(h.router, "3065", h.getReplyListSvc)
}

// Dispatch 根据 minType 分发到对应的 svc 处理（委托给 ProtoRouter）
func (h *Handler) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	return h.router.Dispatch(ctx, minType, reqBytes)
}

// InjectJWTManager 延迟注入 JWT 管理器（预留接口，Topic 域当前不直接使用 JWT）
func (h *Handler) InjectJWTManager(m *member.JWTManager) {
	// Topic 域 Handler 当前不需要直接持有 jwtMgr
}

// InjectTokenStore 延迟注入令牌黑名单存储（预留接口）
func (h *Handler) InjectTokenStore(ts member.TokenStore) {
	// Topic 域 Handler 当前不需要直接持有 tokenStore
}
