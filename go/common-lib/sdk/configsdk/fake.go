package configsdk

import (
	"context"
	"sync"
)

// FakeClient 配置 SDK 的 Fake 实现
// 用于单元测试，不依赖真实配置服务
type FakeClient struct {
	mu      sync.RWMutex
	store   map[string]map[string]interface{}
	watches map[string][]func(string, interface{}, interface{})
}

// NewFakeClient 创建 Fake 配置客户端
func NewFakeClient() *FakeClient {
	return &FakeClient{
		store:   make(map[string]map[string]interface{}),
		watches: make(map[string][]func(string, interface{}, interface{})),
	}
}

// Set 设置配置值（测试辅助方法）
func (f *FakeClient) Set(moduleKey, fieldKey string, value interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.store[moduleKey] == nil {
		f.store[moduleKey] = make(map[string]interface{})
	}
	f.store[moduleKey][fieldKey] = value
}

// GetString 读取字符串配置
func (f *FakeClient) GetString(ctx context.Context, moduleKey string, fieldKey string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if module, ok := f.store[moduleKey]; ok {
		if val, ok := module[fieldKey]; ok {
			if str, ok := val.(string); ok {
				return str, nil
			}
		}
	}
	return "", nil
}

// GetInt 读取整数配置
func (f *FakeClient) GetInt(ctx context.Context, moduleKey string, fieldKey string) (int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if module, ok := f.store[moduleKey]; ok {
		if val, ok := module[fieldKey]; ok {
			switch v := val.(type) {
			case int:
				return int64(v), nil
			case int64:
				return v, nil
			case int32:
				return int64(v), nil
			case float64:
				return int64(v), nil
			}
		}
	}
	return 0, nil
}

// GetBool 读取布尔配置
func (f *FakeClient) GetBool(ctx context.Context, moduleKey string, fieldKey string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if module, ok := f.store[moduleKey]; ok {
		if val, ok := module[fieldKey]; ok {
			if b, ok := val.(bool); ok {
				return b, nil
			}
		}
	}
	return false, nil
}

// Watch 订阅配置变更（Fake 实现仅记录回调，不触发）
func (f *FakeClient) Watch(ctx context.Context, moduleKey string, callback func(fieldKey string, oldValue, newValue interface{})) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.watches[moduleKey] = append(f.watches[moduleKey], callback)
	return nil
}

// Ping 健康检查（Fake 实现永远返回 nil）
func (f *FakeClient) Ping(ctx context.Context) error {
	return nil
}
