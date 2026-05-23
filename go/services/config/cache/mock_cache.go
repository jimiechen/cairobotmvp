package cache

import (
	"sync"
)

// MockCache 基于 sync.Map 的内存缓存实现
// 仅用于单测场景，不适合生产环境（无 TTL、无容量限制、无淘汰策略）
// 生产环境应替换为 Redis 或其他分布式缓存实现
type MockCache struct {
	data sync.Map
}

// NewMockCache 创建空内存缓存实例
func NewMockCache() *MockCache {
	return &MockCache{}
}

// Get 从缓存中读取值
// 返回 (value, true) 表示命中，(nil, false) 表示未命中
func (c *MockCache) Get(key string) (any, bool) {
	return c.data.Load(key)
}

// Set 写入键值对到缓存
// 无 TTL，数据永久驻留直到显式 Delete 或 Invalidate
func (c *MockCache) Set(key string, value any) {
	c.data.Store(key, value)
}

// Delete 删除单个键
func (c *MockCache) Delete(key string) {
	c.data.Delete(key)
}

// Invalidate 按前缀批量失效
// 遍历全量 key 做前缀匹配，仅适用于单测等小规模场景
func (c *MockCache) Invalidate(prefix string) {
	c.data.Range(func(key, _ any) bool {
		if k, ok := key.(string); ok {
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
				c.data.Delete(k)
			}
		}
		return true
	})
}

// Size 返回当前缓存条目数（用于测试断言）
func (c *MockCache) Size() int {
	count := 0
	c.data.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
