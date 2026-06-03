package testutil

import (
	"sync"
)

// MessageHandler 消息处理函数类型（与 SDK RedisClient 接口一致）
type MessageHandler func(msg string)

// CancelFunc 取消函数类型
type CancelFunc func()

// MockRedisClient 统一的 Redis 客户端 Mock 实现
// 用于测试中模拟 Redis 客户端行为，支持 Pub/Sub 和基本 CRUD 操作
type MockRedisClient struct {
	mu          sync.Mutex
	data        map[string]string
	handlers    map[string]MessageHandler
	cancelFuncs map[string]CancelFunc
	subscribed  []string
	getCalls    []string
	setCalls    []SetCallInfo
	deleteCalls []string
}

// SetCallInfo 记录 Set 调用信息
type SetCallInfo struct {
	Key    string
	Value  string
	TTLSec int
}

// NewMockRedisClient 创建 Mock Redis 客户端实例
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:        make(map[string]string),
		handlers:    make(map[string]MessageHandler),
		cancelFuncs: make(map[string]CancelFunc),
	}
}

// Get 从 Redis 中读取值
func (m *MockRedisClient) Get(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls = append(m.getCalls, key)
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil
}

// Set 写入键值对到 Redis
func (m *MockRedisClient) Set(key string, value string, ttlSec int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.setCalls = append(m.setCalls, SetCallInfo{Key: key, Value: value, TTLSec: ttlSec})
	return nil
}

// Delete 删除键
func (m *MockRedisClient) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	m.deleteCalls = append(m.deleteCalls, key)
	return nil
}

// Subscribe 订阅频道
func (m *MockRedisClient) Subscribe(channel string, handler MessageHandler) (CancelFunc, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = handler
	m.subscribed = append(m.subscribed, channel)
	
	cancel := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.handlers, channel)
		delete(m.cancelFuncs, channel)
	}
	m.cancelFuncs[channel] = cancel
	
	return cancel, nil
}

// PublishMessage 模拟发布消息（用于测试触发订阅回调）
func (m *MockRedisClient) PublishMessage(channel string, message string) {
	m.mu.Lock()
	handler, ok := m.handlers[channel]
	m.mu.Unlock()
	if ok && handler != nil {
		handler(message)
	}
}

// GetData 返回当前数据（用于测试断言）
func (m *MockRedisClient) GetData() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]string)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// SetData 直接设置数据（用于测试准备）
func (m *MockRedisClient) SetData(data map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = data
}

// GetSubscribed 获取已订阅的频道列表（用于测试断言）
func (m *MockRedisClient) GetSubscribed() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.subscribed))
	copy(result, m.subscribed)
	return result
}

// GetGetCalls 获取所有 Get 调用记录（用于测试断言）
func (m *MockRedisClient) GetGetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.getCalls))
	copy(result, m.getCalls)
	return result
}

// GetSetCalls 获取所有 Set 调用记录（用于测试断言）
func (m *MockRedisClient) GetSetCalls() []SetCallInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]SetCallInfo, len(m.setCalls))
	copy(result, m.setCalls)
	return result
}

// GetDeleteCalls 获取所有 Delete 调用记录（用于测试断言）
func (m *MockRedisClient) GetDeleteCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.deleteCalls))
	copy(result, m.deleteCalls)
	return result
}

// Clear 清空所有数据（用于测试重置）
func (m *MockRedisClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]string)
	m.handlers = make(map[string]MessageHandler)
	m.cancelFuncs = make(map[string]CancelFunc)
	m.subscribed = nil
	m.getCalls = nil
	m.setCalls = nil
	m.deleteCalls = nil
}
