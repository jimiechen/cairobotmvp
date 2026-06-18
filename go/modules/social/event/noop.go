package event

import "context"

// NoopPublisher 空操作发布器实现
// 用于尚未注入真实 publisher 的场景，避免因事件系统未配置导致业务不可用
// 所有 Publish 调用立即返回 nil，不执行任何操作
type NoopPublisher struct{}

// Publish 空实现，始终返回成功
func (p NoopPublisher) Publish(_ context.Context, _ DomainEvent) error {
	return nil
}
