package event

import (
	"context"
	"testing"
)

// TestFakePublisher_基本功能 验证 FakePublisher 捕获事件的基本行为
func TestFakePublisher_基本功能(t *testing.T) {
	publisher := &FakePublisher{}

	evt := MustNewDomainEvent(NewEventOptions{
		Type:          EventMemberRegistered,
		AggregateType: AggregateMember,
		AggregateID:   "user-001",
	})

	err := publisher.Publish(context.Background(), evt)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}

	if publisher.EventCount() != 1 {
		t.Errorf("期望 1 个事件，实际 %d", publisher.EventCount())
	}
	if publisher.LastEvent().Type != EventMemberRegistered {
		t.Errorf("期望类型 %s，实际 %s", EventMemberRegistered, publisher.LastEvent().Type)
	}
}

// TestFakePublisher_Reset 验证 Reset 清空已捕获事件
func TestFakePublisher_Reset(t *testing.T) {
	publisher := &FakePublisher{}

	evt := MustNewDomainEvent(NewEventOptions{
		Type:          EventGroupCreated,
		AggregateType: AggregateGroup,
		AggregateID:   "group-001",
	})
	publisher.Publish(context.Background(), evt)

	if publisher.EventCount() != 1 {
		t.Fatalf("期望 1 个事件")
	}

	publisher.Reset()

	if publisher.EventCount() != 0 {
		t.Errorf("Reset 后期望 0 个事件，实际 %d", publisher.EventCount())
	}
}
