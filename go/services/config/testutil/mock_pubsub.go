package testutil

import (
	"context"
	"sync"

	"github.com/jimiechen/mineplanet/go/third_party/redisx"
)

// MockPubSub 统一的 Pub/Sub 客户端 Mock 实现
// 用于测试中模拟 Pub/Sub 行为，支持调用记录
type MockPubSub struct {
	mu        sync.Mutex
	published []PubRecord
}

// PubRecord 记录 Publish 调用的参数
type PubRecord struct {
	Channel string
	Message string
}

// NewMockPubSub 创建 Mock Pub/Sub 实例
func NewMockPubSub() *MockPubSub {
	return &MockPubSub{
		published: make([]PubRecord, 0),
	}
}

// Publish 发布消息
func (m *MockPubSub) Publish(_ context.Context, channel string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, PubRecord{Channel: channel, Message: message})
	return nil
}

// Subscribe 订阅频道
func (m *MockPubSub) Subscribe(_ string, _ redisx.MessageHandler) (redisx.CancelFunc, error) {
	return func() {}, nil
}

// Close 关闭连接
func (m *MockPubSub) Close() error {
	return nil
}

// GetPublished 获取所有发布记录（用于测试断言）
func (m *MockPubSub) GetPublished() []PubRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]PubRecord, len(m.published))
	copy(result, m.published)
	return result
}

// Clear 清空记录（用于测试重置）
func (m *MockPubSub) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = nil
}

// MockAuditWriter 统一的审计日志写入器 Mock 实现
// 使用 interface{} 接受任意类型的审计条目，避免循环依赖
type MockAuditWriter struct {
	mu      sync.Mutex
	entries []any
}

// NewMockAuditWriter 创建 Mock 审计写入器实例
func NewMockAuditWriter() *MockAuditWriter {
	return &MockAuditWriter{
		entries: make([]any, 0),
	}
}

// Write 写入审计日志
func (m *MockAuditWriter) Write(_ context.Context, entry any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

// GetEntries 获取所有审计条目（用于测试断言，返回[]any）
func (m *MockAuditWriter) GetEntries() []any {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]any, len(m.entries))
	copy(result, m.entries)
	return result
}

// GetLastEntry 获取最后一条审计条目（用于测试断言，返回any）
func (m *MockAuditWriter) GetLastEntry() any {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return nil
	}
	return m.entries[len(m.entries)-1]
}

// Clear 清空记录（用于测试重置）
func (m *MockAuditWriter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
}
