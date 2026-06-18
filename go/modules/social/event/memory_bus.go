package event

import (
	"context"
	"fmt"
	"sync"
)

// MemoryBus 内存事件总线（并发安全）
// MVP-P0 用于单测和本地开发环境，支持同步或异步模式
// 生产环境应使用 Redis Pub/Sub 实现
//
// 并发安全说明：
//   - Subscribe 使用写锁保护 handlers map
//   - Publish 使用读锁复制 handler 切片后释放锁再执行，避免 handler 执行时持有锁导致死锁
//   - 同一个 eventType 的多个 handler 按注册顺序执行
type MemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewMemoryBus 创建新的内存事件总线实例
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe 注册指定类型的事件处理 handler
// 同一 eventType 可注册多个 handler，按注册顺序执行
func (b *MemoryBus) Subscribe(_ context.Context, eventType string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Publish 发布事件到已注册的 handler
// 采用 best-effort 策略：
//   - 一个 handler 失败不阻止其他 handler 执行
//   - 收集所有错误并返回 aggregated error
//   - handler panic 会被 recover，不影响其他 handler 和主业务
func (b *MemoryBus) Publish(ctx context.Context, evt DomainEvent) error {
	b.mu.RLock()
	handlers := b.handlers[evt.Type]
	// 复制切片引用，释放读锁后再执行 handler
	if len(handlers) == 0 {
		b.mu.RUnlock()
		return nil
	}
	// 深拷贝切片避免并发修改问题
	handlerCopy := make([]Handler, len(handlers))
	copy(handlerCopy, handlers)
	b.mu.RUnlock()

	var errs []error
	for _, h := range handlerCopy {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// handler panic 不影响其他 handler 执行
					errs = append(errs, &HandlerPanicError{
						EventType: evt.Type,
						EventID:   evt.ID,
						Recovered: r,
					})
				}
			}()
			if err := h.Handle(ctx, evt); err != nil {
				errs = append(errs, err)
			}
		}()
	}

	if len(errs) > 0 {
		return &AggregatedError{Errors: errs}
	}
	return nil
}

// HandlerPanicError handler 执行时发生 panic 的错误
type HandlerPanicError struct {
	EventType string
	EventID   string
	Recovered any
}

func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("event handler panic: type=%s id=%v recovered=%v", e.EventType, e.EventID, e.Recovered)
}

// AggregatedError 多个 handler 错误的聚合
type AggregatedError struct {
	Errors []error
}

func (e *AggregatedError) Error() string {
	return fmt.Sprintf("event handlers errors: %d failures", len(e.Errors))
}
