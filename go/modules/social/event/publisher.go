package event

import "context"

// Publisher 领域事件发布接口
// svc 层只依赖此接口，不允许直接依赖 MemoryBus 或 Redis 实现
// 单测中注入 FakePublisher 可捕获并断言发布的事件
type Publisher interface {
	// Publish 发布领域事件到事件总线
	// MVP-P0：发布失败不回滚业务结果，但必须记录日志
	Publish(ctx context.Context, evt DomainEvent) error
}

// Subscriber 领域事件订阅接口
// 用于注册事件消费者 handler，支持 Redis PubSub 和 MemoryBus 两种实现
type Subscriber interface {
	// Subscribe 订阅指定类型的事件，注册处理 handler
	Subscribe(ctx context.Context, eventType string, handler Handler) error
}

// Handler 领域事件消费者接口
// 每个 handler 实现对特定事件的消费逻辑（统计更新、缓存失效、通知预留等）
type Handler interface {
	Handle(ctx context.Context, evt DomainEvent) error
}

// HandlerFunc 函数适配器，允许用普通函数实现 Handler 接口
type HandlerFunc func(ctx context.Context, evt DomainEvent) error

func (f HandlerFunc) Handle(ctx context.Context, evt DomainEvent) error {
	return f(ctx, evt)
}
