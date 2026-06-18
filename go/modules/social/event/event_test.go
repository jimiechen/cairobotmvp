package event

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== DomainEvent 构造与序列化测试 =====

func TestNewDomainEvent_基本构造(t *testing.T) {
	payload := MemberRegisteredPayload{
		UserID:   "user-001",
		Username: "testuser",
		Nickname: "Test User",
	}

	evt, err := NewDomainEvent(NewEventOptions{
		Type:          EventMemberRegistered,
		AggregateType: AggregateMember,
		AggregateID:   "user-001",
		ActorID:       "user-001",
		Payload:       payload,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, evt.ID)
	assert.Equal(t, EventMemberRegistered, evt.Type)
	assert.Equal(t, EventVersionCurrent, evt.Version)
	assert.Equal(t, AggregateMember, evt.AggregateType)
	assert.Equal(t, "user-001", evt.AggregateID)
	assert.Equal(t, "user-001", evt.ActorID)
	assert.Greater(t, evt.OccurredAt, int64(0))
	assert.NotEmpty(t, evt.Payload)
}

func TestNewDomainEvent_Payload序列化为JSON(t *testing.T) {
	payload := GroupCreatedPayload{
		GroupID:    "group-001",
		OwnerID:    "owner-001",
		Type:       "free",
		Visibility: 1,
		JoinMode:   1,
	}

	evt, err := NewDomainEvent(NewEventOptions{
		Type:        EventGroupCreated,
		Payload:     payload,
		AggregateID: "group-001",
	})

	require.NoError(t, err)

	// 验证 Payload 是合法 JSON
	var decoded map[string]any
	err = json.Unmarshal(evt.Payload, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "group-001", decoded["group_id"])
	assert.Equal(t, "owner-001", decoded["owner_id"])
}

func TestMustNewDomainEvent_成功构造(t *testing.T) {
	evt := MustNewDomainEvent(NewEventOptions{
		Type:        EventTopicCreated,
		AggregateID: "topic-001",
	})
	assert.Equal(t, EventTopicCreated, evt.Type)
	assert.NotEmpty(t, evt.ID)
}

func TestNewDomainEvent_NilPayload不报错(t *testing.T) {
	evt, err := NewDomainEvent(NewEventOptions{
		Type:        EventGroupLeft,
		AggregateID: "group-001",
		Payload:     nil,
	})
	require.NoError(t, err)
	assert.Nil(t, evt.Payload)
}

// ===== NoopPublisher 测试 =====

func TestNoopPublisher_Publish返回nil(t *testing.T) {
	pub := NoopPublisher{}
	err := pub.Publish(context.Background(), DomainEvent{Type: "test"})
	assert.NoError(t, err)
}

// ===== MemoryBus 基础测试 =====

func TestMemoryBus_SubscribeAndPublish(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	var receivedEvt DomainEvent
	handler := HandlerFunc(func(_ context.Context, evt DomainEvent) error {
		receivedEvt = evt
		return nil
	})

	err := bus.Subscribe(ctx, EventMemberRegistered, handler)
	require.NoError(t, err)

	evt := MustNewDomainEvent(NewEventOptions{
		Type:        EventMemberRegistered,
		AggregateID: "user-001",
	})

	err = bus.Publish(ctx, evt)
	assert.NoError(t, err)
	assert.Equal(t, "user-001", receivedEvt.AggregateID)
	assert.Equal(t, EventMemberRegistered, receivedEvt.Type)
}

func TestMemoryBus_Publish无注册Handler不报错(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	evt := MustNewDomainEvent(NewEventOptions{
		Type: EventGroupCreated,
	})

	err := bus.Publish(ctx, evt)
	assert.NoError(t, err)
}

func TestMemoryBus_多个Handler按顺序执行(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	var order []int
	h1 := HandlerFunc(func(_ context.Context, _ DomainEvent) error { order = append(order, 1); return nil })
	h2 := HandlerFunc(func(context.Context, DomainEvent) error { order = append(order, 2); return nil })
	h3 := HandlerFunc(func(context.Context, DomainEvent) error { order = append(order, 3); return nil })

	_ = bus.Subscribe(ctx, EventGroupJoined, h1)
	_ = bus.Subscribe(ctx, EventGroupJoined, h2)
	_ = bus.Subscribe(ctx, EventGroupJoined, h3)

	_ = bus.Publish(ctx, DomainEvent{Type: EventGroupJoined})
	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestMemoryBus_BestEffort不阻断其他Handler(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	var secondCalled bool
	hFail := HandlerFunc(func(context.Context, DomainEvent) error { return errors.New("handler failed") })
	hOk := HandlerFunc(func(context.Context, DomainEvent) error { secondCalled = true; return nil })

	_ = bus.Subscribe(ctx, EventTopicCreated, hFail)
	_ = bus.Subscribe(ctx, EventTopicCreated, hOk)

	err := bus.Publish(ctx, DomainEvent{Type: EventTopicCreated})
	assert.Error(t, err) // 返回聚合错误
	assert.True(t, secondCalled) // 第二个 handler 仍然执行了
}

func TestMemoryBus_HandlerPanic被Recover(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	var secondCalled bool
	hPanic := HandlerFunc(func(context.Context, DomainEvent) error { panic("模拟 panic") })
	hOk := HandlerFunc(func(context.Context, DomainEvent) error { secondCalled = true; return nil })

	_ = bus.Subscribe(ctx, EventTopicDeleted, hPanic)
	_ = bus.Subscribe(ctx, EventTopicDeleted, hOk)

	err := bus.Publish(ctx, DomainEvent{Type: EventTopicDeleted})
	assert.Error(t, err) // 返回包含 panic 错误的聚合错误
	assert.True(t, secondCalled)
}

// ===== MemoryBus 并发安全测试 =====

func TestMemoryBus_ConcurrentSubscribePublish(t *testing.T) {
	bus := NewMemoryBus()
	ctx := context.Background()

	// 并发 Subscribe
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			_ = bus.Subscribe(ctx, EventTopicReacted, HandlerFunc(
				func(_ context.Context, _ DomainEvent) error { return nil },
			))
			done <- true
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	// 并发 Publish 不应 panic
	for i := 0; i < 10; i++ {
		go func() {
			_ = bus.Publish(ctx, DomainEvent{Type: EventTopicReacted})
		}()
	}
	// 等待并发 publish 完成
	time.Sleep(100 * time.Millisecond)
}

// ===== FakePublisher 测试 =====

func TestFakePublisher_捕获事件(t *testing.T) {
	pub := &FakePublisher{}
	ctx := context.Background()

	evt1 := DomainEvent{Type: EventMemberRegistered, AggregateID: "u1"}
	evt2 := DomainEvent{Type: EventGroupCreated, AggregateID: "g1"}

	_ = pub.Publish(ctx, evt1)
	_ = pub.Publish(ctx, evt2)

	assert.Equal(t, 2, pub.EventCount())
	assert.Equal(t, EventMemberRegistered, pub.Events()[0].Type)
	assert.Equal(t, EventGroupCreated, pub.Events()[1].Type)
	assert.Equal(t, evt2, pub.LastEvent())
}

func TestFakePublisher_SetError(t *testing.T) {
	pub := &FakePublisher{}
	ctx := context.Background()

	pub.SetError(errors.New("publish failed"))
	err := pub.Publish(ctx, DomainEvent{Type: "test"})

	assert.ErrorContains(t, err, "publish failed")
}

func TestFakePublisher_Reset清空状态(t *testing.T) {
	pub := &FakePublisher{}
	ctx := context.Background()

	_ = pub.Publish(ctx, DomainEvent{Type: "test"})
	assert.Equal(t, 1, pub.EventCount())

	pub.Reset()
	assert.Equal(t, 0, pub.EventCount())
	assert.Empty(t, pub.LastEvent().Type)
}

// ===== 常量完整性检查 =====

func Test事件常量不为空(t *testing.T) {
	assert.NotEmpty(t, EventMemberRegistered)
	assert.NotEmpty(t, EventUserStatusChanged)
	assert.NotEmpty(t, EventGroupCreated)
	assert.NotEmpty(t, EventGroupJoined)
	assert.NotEmpty(t, EventGroupLeft)
	assert.NotEmpty(t, EventGroupMemberRemoved)
	assert.NotEmpty(t, EventGroupMemberBanned)
	assert.NotEmpty(t, EventGroupMemberMuted)
	assert.NotEmpty(t, EventGroupMemberRecovered)
	assert.NotEmpty(t, EventGroupPlanCreated)
	assert.NotEmpty(t, EventGroupOrderPaid)
	assert.NotEmpty(t, EventGroupMemberActivated)
	assert.NotEmpty(t, EventTopicCreated)
	assert.NotEmpty(t, EventTopicDeleted)
	assert.NotEmpty(t, EventTopicCommentCreated)
	assert.NotEmpty(t, EventTopicReacted)
	assert.NotEmpty(t, EventTopicApproved)
	assert.NotEmpty(t, EventTopicRejected)
	assert.NotEmpty(t, EventTopicBanned)
}

func Test聚合根常量不为空(t *testing.T) {
	assert.Equal(t, "member", AggregateMember)
	assert.Equal(t, "group", AggregateGroup)
	assert.Equal(t, "topic", AggregateTopic)
	assert.Equal(t, "group_order", AggregateOrder)
	assert.Equal(t, "topic_comment", AggregateComment)
}

func Test动作和互动类型常量不为空(t *testing.T) {
	assert.Equal(t, "ban", ActionBan)
	assert.Equal(t, "remove", ActionRemove)
	assert.Equal(t, "mute", ActionMute)
	assert.Equal(t, "recover", ActionRecover)
	assert.Equal(t, "like", ReactionTypeLike)
	assert.Equal(t, "favorite", ReactionTypeFavorite)
	assert.Equal(t, "share", ReactionTypeShare)
}
