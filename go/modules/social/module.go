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
	publisher event.Publisher
}

// WithPublisher 注入事件发布器
// 不注入时默认使用 NoopPublisher（事件不阻塞业务）
func WithPublisher(p event.Publisher) ModuleOption {
	return func(c *moduleConfig) {
		c.publisher = p
	}
}

// NewModule 创建社交域模块实例，注入各域 Repository 和可选的事件发布器
// 不传 WithPublisher 时默认使用 NoopPublisher
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

	return &Module{
		MemberServant: member.NewServant(memberRepo, cfg.publisher),
		GroupServant:  group.NewServant(groupRepo, cfg.publisher),
		TopicServant:   topic.NewServant(topicRepo, cfg.publisher),
	}
}
