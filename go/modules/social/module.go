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

// NewModule 创建社交域模块实例，注入各域 Repository 和事件发布器依赖
func NewModule(memberRepo member.Repository, groupRepo group.Repository, topicRepo topic.Repository, publisher event.Publisher) *Module {
	return &Module{
		MemberServant: member.NewServant(memberRepo, publisher),
		GroupServant:  group.NewServant(groupRepo, publisher),
		TopicServant:   topic.NewServant(topicRepo, publisher),
	}
}
