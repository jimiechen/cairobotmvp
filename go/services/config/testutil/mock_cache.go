package testutil

import (
	"context"
	"sync"
	"time"
)

// MockCache 统一的 redisx.Client Mock 实现
// 用于测试中模拟 Redis 客户端行为，支持调用记录和状态跟踪
type MockCache struct {
	mu           sync.Mutex
	data         map[string]any
	invalidated  []string
	deleted      []string
	setCalls     []SetCall
	getError     error
	setError     error
	deleteError  error
	invalidateError error
}

// SetCall 记录 Set 调用的参数
type SetCall struct {
	Key   string
	Value any
	TTL   time.Duration
}

// NewMockCache 创建 Mock 缓存实例
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]any),
	}
}

// Get 从缓存中读取值
func (m *MockCache) Get(_ context.Context, key string) (string, error) {
	if m.getError != nil {
		return "", m.getError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.data[key]; ok {
		return val.(string), nil
	}
	return "", nil
}

// Set 写入键值对到缓存
func (m *MockCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	if m.setError != nil {
		return m.setError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.setCalls = append(m.setCalls, SetCall{Key: key, Value: value, TTL: ttl})
	return nil
}

// Delete 删除键
func (m *MockCache) Delete(_ context.Context, keys ...string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.data, key)
		m.deleted = append(m.deleted, key)
	}
	return nil
}

// Scan 扫描匹配的键
func (m *MockCache) Scan(_ context.Context, pattern string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []string
	for key := range m.data {
		result = append(result, key)
	}
	return result, nil
}

// Invalidate 按前缀批量失效
func (m *MockCache) Invalidate(_ context.Context, pattern string) error {
	if m.invalidateError != nil {
		return m.invalidateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invalidated = append(m.invalidated, pattern)
	return nil
}

// Ping 检查连接状态
func (m *MockCache) Ping(_ context.Context) error {
	return nil
}

// Close 关闭连接
func (m *MockCache) Close() error {
	return nil
}

// GetData 返回当前缓存数据（用于测试断言）
func (m *MockCache) GetData() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]any)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// GetInvalidated 返回被失效的键列表（用于测试断言）
func (m *MockCache) GetInvalidated() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.invalidated
}

// GetDeleted 返回被删除的键列表（用于测试断言）
func (m *MockCache) GetDeleted() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleted
}

// GetSetCalls 返回所有 Set 调用记录（用于测试断言）
func (m *MockCache) GetSetCalls() []SetCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.setCalls
}

// SetData 直接设置缓存数据（用于测试准备）
func (m *MockCache) SetData(data map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = data
}

// WithGetError 设置 Get 方法返回的错误（用于测试错误场景）
func (m *MockCache) WithGetError(err error) *MockCache {
	m.getError = err
	return m
}

// WithSetError 设置 Set 方法返回的错误（用于测试错误场景）
func (m *MockCache) WithSetError(err error) *MockCache {
	m.setError = err
	return m
}

// WithDeleteError 设置 Delete 方法返回的错误（用于测试错误场景）
func (m *MockCache) WithDeleteError(err error) *MockCache {
	m.deleteError = err
	return m
}

// WithInvalidateError 设置 Invalidate 方法返回的错误（用于测试错误场景）
func (m *MockCache) WithInvalidateError(err error) *MockCache {
	m.invalidateError = err
	return m
}
