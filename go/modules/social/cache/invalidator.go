package cache

import "context"

// CacheInvalidator 缓存失效接口
// 定义统一的缓存删除能力，支持具体 key 删除和 pattern 批量失效
// MVP-P0 提供 Noop 实现，后续接入 Redis 后替换为真实实现
type CacheInvalidator interface {
	// Delete 删除一个或多个指定 key
	Delete(ctx context.Context, keys ...string) error
	// DeletePattern 按 pattern 批量删除匹配的 key（如 "group:member:groupId:*"）
	DeletePattern(ctx context.Context, pattern string) error
}

// NoopCacheInvalidator 空操作缓存失效器
// MVP-P0 使用此实现，不执行任何实际操作，避免 Redis 未配置时阻塞业务
type NoopCacheInvalidator struct{}

// Delete 空实现
func (n NoopCacheInvalidator) Delete(_ context.Context, _ ...string) error { return nil }

// DeletePattern 空实现
func (n NoopCacheInvalidator) DeletePattern(_ context.Context, _ string) error { return nil }
