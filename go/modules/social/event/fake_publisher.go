package event

import "context"

// FakePublisher 测试用假发布器
// 捕获所有 Publish 调用的事件，用于单测中断言事件发布行为
//
// 使用方式：
//
//	publisher := &event.FakePublisher{}
//	svc := member.NewSvcRegister(repo, publisher)
//	svc.Handle(ctx, req)
//	require.Len(t, publisher.Events(), 1)
//	assert.Equal(t, event.EventMemberRegistered, publisher.Events()[0].Type)
type FakePublisher struct {
	events []DomainEvent
	err    error
}

// Publish 记录事件到内部切片，可选择性返回错误
func (p *FakePublisher) Publish(_ context.Context, evt DomainEvent) error {
	p.events = append(p.events, evt)
	return p.err
}

// Events 返回已捕获的所有事件（按调用顺序）
func (p *FakePublisher) Events() []DomainEvent {
	return p.events
}

// SetError 设置下次 Publish 调用时返回的错误
// 用于测试事件发布失败场景
func (p *FakePublisher) SetError(err error) {
	p.err = err
}

// Reset 清空已捕获的事件和错误状态
func (p *FakePublisher) Reset() {
	p.events = nil
	p.err = nil
}

// EventCount 返回已捕获的事件数量
func (p *FakePublisher) EventCount() int {
	return len(p.events)
}

// LastEvent 返回最后一个捕获的事件，无事件时返回零值
func (p *FakePublisher) LastEvent() DomainEvent {
	if len(p.events) == 0 {
		return DomainEvent{}
	}
	return p.events[len(p.events)-1]
}
