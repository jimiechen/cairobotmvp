// Package eventhandler 社交域事件消费者
// 负责消费领域事件并驱动 2级数据更新、缓存失效、通知预留
//
// 本包依赖 member/group/topic repository，不能放在 event 包内（避免循环依赖）
// 消费者按事件类型注册到 MemoryBus 或 RedisSubscriber
package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jimiechen/mineplanet/go/modules/social/cache"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// StatsHandler 统计更新事件消费者
// 监听成员/群组/帖子相关事件，驱动 2级统计数据更新（member_stats/group_stats/topic_stats）
// MVP-P0 首版：优先保证事件分发正确性，统计更新可先保留 TODO
type StatsHandler struct {
	// TODO(eventhandler): 注入 member/group/topic repository 用于统计更新
	// 当前骨架阶段仅记录日志，后续补充真实 DB 更新逻辑
}

// NewStatsHandler 创建统计 handler 实例
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

// Handle 处理领域事件，根据事件类型路由到对应统计更新逻辑
func (h *StatsHandler) Handle(ctx context.Context, evt event.DomainEvent) error {
	switch evt.Type {
	case event.EventMemberRegistered:
		return h.onMemberRegistered(ctx, evt)
	case event.EventGroupCreated:
		return h.onGroupCreated(ctx, evt)
	case event.EventGroupJoined:
		return h.onGroupJoined(ctx, evt)
	case event.EventGroupLeft:
		return h.onGroupLeft(ctx, evt)
	case event.EventGroupOrderPaid:
		return h.onGroupOrderPaid(ctx, evt)
	case event.EventTopicCreated:
		return h.onTopicCreated(ctx, evt)
	case event.EventTopicDeleted:
		return h.onTopicDeleted(ctx, evt)
	case event.EventTopicCommentCreated:
		return h.onTopicCommentCreated(ctx, evt)
	case event.EventTopicReacted:
		return h.onTopicReacted(ctx, evt)
	default:
		// 不关心的事件类型，跳过
		return nil
	}
}

// ===== 各事件处理方法 =====
// MVP-P0 骨架：返回 nil 不报错，后续补充真实统计更新逻辑

func (h *StatsHandler) onMemberRegistered(_ context.Context, _ event.DomainEvent) error {
	// TODO: 初始化 member_stats 记录
	// TODO: 发送欢迎通知（通知系统就绪后）
	return nil
}

func (h *StatsHandler) onGroupCreated(_ context.Context, _ event.DomainEvent) error {
	// TODO: 初始化 group_stats 记录（members_count=1 含 owner）
	return nil
}

func (h *StatsHandler) onGroupJoined(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 group_stats.members_count += 1（从 1级数据重算更安全）
	return nil
}

func (h *StatsHandler) onGroupLeft(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 group_stats.members_count -= 1（从 1级数据重算更安全）
	return nil
}

func (h *StatsHandler) onGroupOrderPaid(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 group_stats.paid_members_count
	return nil
}

func (h *StatsHandler) onTopicCreated(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 group_stats.topics_count += 1
	// TODO: 更新 member.stats.topics_count += 1
	return nil
}

func (h *StatsHandler) onTopicDeleted(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 group_stats.topics_count -= 1
	// TODO: 更新 member.stats.topics_count -= 1
	return nil
}

func (h *StatsHandler) onTopicCommentCreated(_ context.Context, _ event.DomainEvent) error {
	// TODO: 更新 topic_stats.comments_count += 1
	return nil
}

func (h *StatsHandler) onTopicReacted(_ context.Context, evt event.DomainEvent) error {
	// TODO: 根据 reaction_type 更新 topic_stats 对应计数
	_ = evt
	return nil
}

// CacheHandler 缓存失效事件消费者
// 监听领域事件，主动失效相关缓存 key
// MVP-P0 首版使用 NoopCacheInvalidator，不执行真实 Redis 操作
type CacheHandler struct {
	invalidator cache.CacheInvalidator
}

// NewCacheHandler 创建缓存失效 handler 实例
func NewCacheHandler(invalidator cache.CacheInvalidator) *CacheHandler {
	if invalidator == nil {
		invalidator = cache.NoopCacheInvalidator{}
	}
	return &CacheHandler{invalidator: invalidator}
}

// Handle 处理领域事件，根据事件类型路由到对应缓存失效逻辑
func (h *CacheHandler) Handle(ctx context.Context, evt event.DomainEvent) error {
	switch evt.Type {
	case event.EventMemberRegistered:
		// 新用户注册，可预失效 profile 缓存（如有旧缓存）
		return nil
	case event.EventUserStatusChanged:
		// 用户状态变更，失效 profile 和权限相关缓存
		var payload event.UserStatusChangedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx,
			cache.MemberProfileKey(payload.UserID),
		)
	case event.EventGroupCreated:
		// 群组创建，失效推荐候选池
		return h.invalidator.DeletePattern(ctx, cache.GroupRecommendKey("*"))
	case event.EventGroupJoined:
		var payload event.GroupJoinedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx,
			cache.GroupMemberKey(payload.GroupID, payload.UserID),
			cache.GroupMembersKey(payload.GroupID, 1), // 第一页成员列表
			cache.GroupStatsKey(payload.GroupID),
		)
	case event.EventGroupLeft:
		var payload event.GroupLeftPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx,
			cache.GroupMemberKey(payload.GroupID, payload.UserID),
			cache.GroupMembersKey(payload.GroupID, 1),
			cache.GroupStatsKey(payload.GroupID),
		)
	case event.EventGroupMemberBanned, event.EventGroupMemberRemoved:
		var payload event.GroupMemberChangedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx,
			cache.GroupMemberKey(payload.GroupID, payload.TargetUserID),
			cache.GroupMembersKey(payload.GroupID, 1),
			cache.GroupStatsKey(payload.GroupID),
		)
	case event.EventGroupPlanCreated:
		var payload event.GroupPlanCreatedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx, cache.GroupPlansKey(payload.GroupID))
	case event.EventGroupOrderPaid:
		var payload event.GroupOrderPaidPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx,
			cache.GroupMemberKey(payload.GroupID, payload.UserID),
			cache.GroupStatsKey(payload.GroupID),
		)
	case event.EventTopicCreated:
		var payload event.TopicCreatedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.DeletePattern(ctx, cache.GroupTopicsKey(payload.GroupID, 0))
	case event.EventTopicDeleted:
		var payload event.TopicDeletedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		if err := h.invalidator.Delete(ctx,
			cache.TopicDetailKey(payload.TopicID),
		); err != nil {
			return err
		}
		return h.invalidator.DeletePattern(ctx, cache.GroupTopicsKey(payload.GroupID, 0))
	case event.EventTopicCommentCreated:
		var payload event.TopicCommentCreatedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx, cache.TopicStatsKey(payload.TopicID))
	case event.EventTopicReacted:
		var payload event.TopicReactedPayload
		if err := json.Unmarshal(evt.Payload, &payload); err != nil {
			return err
		}
		return h.invalidator.Delete(ctx, cache.TopicStatsKey(payload.TopicID))
	default:
		return nil
	}
}

// NotifyHandler 通知预留事件消费者（MVP-P0 骨架）
// 监听需要通知用户的高权限操作事件
// MVP-P0 仅做日志记录，不写入 notifications 表
type NotifyHandler struct{}

// NewNotifyHandler 创建通知 handler 实例
func NewNotifyHandler() *NotifyHandler {
	return &NotifyHandler{}
}

// Handle 处理需要通知的事件
func (h *NotifyHandler) Handle(_ context.Context, evt event.DomainEvent) error {
	switch evt.Type {
	case event.EventGroupMemberBanned, event.EventGroupMemberRemoved, event.EventGroupMemberMuted:
		// TODO: 通知被操作用户
		_ = fmt.Sprintf("[NOTIFY] 用户收到成员管理通知: event=%s id=%s", evt.Type, evt.ID)
	case event.EventGroupOrderPaid:
		// TODO: 通知用户权益已开通
		_ = fmt.Sprintf("[NOTIFY] 用户权益开通通知: event=%s id=%s", evt.Type, evt.ID)
	case event.EventTopicApproved, event.EventTopicRejected, event.EventTopicBanned:
		// TODO: 通知帖子作者审核结果
		_ = fmt.Sprintf("[NOTIFY] 帖子审核结果通知: event=%s id=%s", evt.Type, evt.ID)
	}
	return nil
}
