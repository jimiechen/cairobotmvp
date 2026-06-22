// Package social 社交域模块入口，聚合所有域 Servant 注册
// 不包含业务逻辑，仅做模块级编排
package social

import (
	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
)

// Module 聚合社交域所有子域的 Servant
type Module struct {
	MemberServant *member.Servant
	GroupServant  *group.Servant
	TopicServant   *topic.Servant
}

// ModuleOption 模块构造选项（函数式选项模式）
type ModuleOption func(*moduleConfig)

// moduleConfig 模块内部配置，支持可选依赖注入
type moduleConfig struct {
	publisher  event.Publisher
	jwtManager *member.JWTManager
	tokenStore member.TokenStore
}

// WithPublisher 注入事件发布器
// 不注入时默认使用 NoopPublisher（事件不阻塞业务）
func WithPublisher(p event.Publisher) ModuleOption {
	return func(c *moduleConfig) {
		c.publisher = p
	}
}

// WithJWTManager 注入 JWT 管理器（三域黑名单检查需要）
// 不注入时黑名单检查静默跳过（降级放行）
func WithJWTManager(m *member.JWTManager) ModuleOption {
	return func(c *moduleConfig) {
		c.jwtManager = m
	}
}

// WithTokenStore 注入令牌黑名单存储（三域鉴权路径拦截需要）
// 不注入时黑名单检查静默跳过（降级放行）
func WithTokenStore(ts member.TokenStore) ModuleOption {
	return func(c *moduleConfig) {
		c.tokenStore = ts
	}
}

// NewModule 创建社交域模块实例，注入各域 Repository 和可选的事件发布器
// 不传 WithPublisher 时默认使用 NoopPublisher
//
// JWTManager 和 TokenStore 会同步注入到 Member/Group/Topic 三个 Servant，
// 确保三域鉴权路径的黑名单检查一致生效。
func NewModule(
	memberRepo member.Repository,
	groupRepo group.Repository,
	topicRepo topic.Repository,
	opts ...ModuleOption,
) *Module {
	cfg := &moduleConfig{
		publisher: &event.NoopPublisher{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	socialMod := &Module{
		MemberServant: member.NewServant(memberRepo, cfg.publisher),
		GroupServant:  group.NewServant(groupRepo, cfg.publisher),
		TopicServant:   topic.NewServant(topicRepo, cfg.publisher),
	}
	// 延迟注入 JWTManager 到三域 Servant（解决循环依赖）
	if cfg.jwtManager != nil {
		socialMod.MemberServant.InjectJWTManager(cfg.jwtManager)
		socialMod.GroupServant.InjectJWTManager(cfg.jwtManager)
		socialMod.TopicServant.InjectJWTManager(cfg.jwtManager)
	}
	// 延迟注入 TokenStore 到三域 Servant（解决循环依赖）
	if cfg.tokenStore != nil {
		socialMod.MemberServant.InjectTokenStore(cfg.tokenStore)
		socialMod.GroupServant.InjectTokenStore(cfg.tokenStore)
		socialMod.TopicServant.InjectTokenStore(cfg.tokenStore)
	}
	return socialMod
}
