package eventhandler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== StatsHandler 测试 =====

func TestStatsHandler_注册事件路由正确(t *testing.T) {
	h := NewStatsHandler()
	ctx := context.Background()

	testCases := []struct {
		name      string
		eventType string
		payload   any
	}{
		{"MemberRegistered", event.EventMemberRegistered, event.MemberRegisteredPayload{UserID: "u1"}},
		{"GroupCreated", event.EventGroupCreated, event.GroupCreatedPayload{GroupID: "g1"}},
		{"GroupJoined", event.EventGroupJoined, event.GroupJoinedPayload{GroupID: "g1", UserID: "u1"}},
		{"GroupLeft", event.EventGroupLeft, event.GroupLeftPayload{GroupID: "g1"}},
		{"GroupOrderPaid", event.EventGroupOrderPaid, event.GroupOrderPaidPayload{GroupID: "g1"}},
		{"TopicCreated", event.EventTopicCreated, event.TopicCreatedPayload{TopicID: "t1"}},
		{"TopicDeleted", event.EventTopicDeleted, event.TopicDeletedPayload{TopicID: "t1"}},
		{"TopicCommentCreated", event.EventTopicCommentCreated, event.TopicCommentCreatedPayload{TopicID: "t1"}},
		{"TopicReacted", event.EventTopicReacted, event.TopicReactedPayload{TopicID: "t1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payloadJSON, _ := json.Marshal(tc.payload)
			evt := event.DomainEvent{
				Type:    tc.eventType,
				Payload: payloadJSON,
			}
			err := h.Handle(ctx, evt)
			assert.NoError(t, err, "%s 事件处理不应返回错误", tc.name)
		})
	}
}

func TestStatsHandler_不关心的事件跳过(t *testing.T) {
	h := NewStatsHandler()
	ctx := context.Background()

	// 传入 StatsHandler 不处理的事件类型
	evt := event.DomainEvent{Type: "UnknownEvent"}
	err := h.Handle(ctx, evt)
	assert.NoError(t, err)
}

// ===== CacheHandler 测试 =====

func TestCacheHandler_缓存失效Key构造正确(t *testing.T) {
	h := NewCacheHandler(nil)
	ctx := context.Background()

	// GroupJoined → 应失效 member + members + stats key
	joinPayload, _ := json.Marshal(event.GroupJoinedPayload{
		GroupID: "group-001",
		UserID:  "user-001",
	})
	err := h.Handle(ctx, event.DomainEvent{
		Type:    event.EventGroupJoined,
		Payload: joinPayload,
	})
	assert.NoError(t, err)

	// GroupLeft → 应失效 member + members + stats key
	leftPayload, _ := json.Marshal(event.GroupLeftPayload{
		GroupID: "group-001",
		UserID:  "user-001",
	})
	err = h.Handle(ctx, event.DomainEvent{
		Type:    event.EventGroupLeft,
		Payload: leftPayload,
	})
	assert.NoError(t, err)

	// TopicReacted → 应失效 topic stats
	reactPayload, _ := json.Marshal(event.TopicReactedPayload{
		TopicID: "topic-001",
	})
	err = h.Handle(ctx, event.DomainEvent{
		Type:    event.EventTopicReacted,
		Payload: reactPayload,
	})
	assert.NoError(t, err)
}

func TestCacheHandler_NoopInvalidator不报错(t *testing.T) {
	h := NewCacheHandler(nil)
	ctx := context.Background()

	// 每种事件类型提供最小 payload，确保 Noop 模式下不报错
	testCases := []struct {
		eventType string
		payload   any
	}{
		{event.EventMemberRegistered, nil}, // MemberRegistered 不解析 payload，跳过
		{event.EventGroupCreated, nil},     // GroupCreated 不解析 payload，跳过
		{event.EventGroupJoined, event.GroupJoinedPayload{GroupID: "g1", UserID: "u1"}},
		{event.EventGroupLeft, event.GroupLeftPayload{GroupID: "g1", UserID: "u1"}},
		{event.EventGroupMemberBanned, event.GroupMemberChangedPayload{GroupID: "g1", TargetUserID: "u1", Action: event.ActionBan}},
		{event.EventGroupMemberRemoved, event.GroupMemberChangedPayload{GroupID: "g1", TargetUserID: "u1", Action: event.ActionRemove}},
		{event.EventGroupPlanCreated, event.GroupPlanCreatedPayload{GroupID: "g1"}},
		{event.EventGroupOrderPaid, event.GroupOrderPaidPayload{GroupID: "g1", UserID: "u1"}},
		{event.EventTopicCreated, event.TopicCreatedPayload{TopicID: "t1", GroupID: "g1"}},
		{event.EventTopicDeleted, event.TopicDeletedPayload{TopicID: "t1", GroupID: "g1"}},
		{event.EventTopicCommentCreated, event.TopicCommentCreatedPayload{TopicID: "t1"}},
		{event.EventTopicReacted, event.TopicReactedPayload{TopicID: "t1"}},
	}
	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			var payloadJSON []byte
			if tc.payload != nil {
				payloadJSON, _ = json.Marshal(tc.payload)
			}
			err := h.Handle(ctx, event.DomainEvent{Type: tc.eventType, Payload: payloadJSON})
			require.NoError(t, err)
		})
	}
}

// ===== NotifyHandler 测试 =====

func TestNotifyHandler_通知事件不报错(t *testing.T) {
	h := NewNotifyHandler()
	ctx := context.Background()

	notifyEvents := []string{
		event.EventGroupMemberBanned,
		event.EventGroupMemberRemoved,
		event.EventGroupMemberMuted,
		event.EventGroupOrderPaid,
		event.EventTopicApproved,
		event.EventTopicRejected,
		event.EventTopicBanned,
	}
	for _, et := range notifyEvents {
		t.Run(et, func(t *testing.T) {
			err := h.Handle(ctx, event.DomainEvent{Type: et, ID: "test-id"})
			assert.NoError(t, err)
		})
	}
}
