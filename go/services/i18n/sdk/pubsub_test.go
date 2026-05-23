package sdk

import (
	"sync"
	"testing"
)

// mockRedisClient 用于测试的模拟 Redis 客户端
type mockRedisClient struct {
	mu        sync.Mutex
	data      map[string]string
	handlers  map[string]func(string)
	published map[string][]string
	closed    bool
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data:      make(map[string]string),
		handlers:  make(map[string]func(string)),
		published: make(map[string][]string),
	}
}

func (m *mockRedisClient) Get(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key], nil
}

func (m *mockRedisClient) Set(key string, value string, ttlSec int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockRedisClient) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *mockRedisClient) Subscribe(channel string, handler func(string)) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = handler
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.handlers, channel)
	}, nil
}

func (m *mockRedisClient) Publish(channel string, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published[channel] = append(m.published[channel], message)
	return nil
}

func (m *mockRedisClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockRedisClient) trigger(channel string, msg string) {
	m.mu.Lock()
	handler, exists := m.handlers[channel]
	m.mu.Unlock()
	if exists {
		handler(msg)
	}
}

func TestPubSubClient_New(t *testing.T) {
	opts := &Options{
		Mode: ModeRemote,
		Redis: &RedisConfig{
			Addr: "localhost:6379",
			DB:   0,
		},
	}
	ps := newPubSubClient(opts)

	if ps.options == nil {
		t.Error("options should not be nil")
	}
	if ps.options.Redis.Addr != "localhost:6379" {
		t.Errorf("expected redis addr 'localhost:6379', got '%s'", ps.options.Redis.Addr)
	}
}

func TestPubSubClient_StartStop(t *testing.T) {
	mockRedis := newMockRedisClient()
	ps := newPubSubClient(&Options{
		Redis: &RedisConfig{Addr: "localhost:6379"},
	})

	// 手动注入 mock redis
	ps.redis = mockRedis
	ps.active = true

	if !ps.IsConnected() {
		t.Error("expected pubsub to be connected")
	}

	err := ps.Stop()
	if err != nil {
		t.Errorf("unexpected error on Stop: %v", err)
	}

	if ps.IsConnected() {
		t.Error("expected pubsub to be disconnected after Stop")
	}

	if !mockRedis.closed {
		t.Error("expected mock redis to be closed")
	}
}

func TestPubSubClient_Start_NoRedisConfig(t *testing.T) {
	ps := newPubSubClient(&Options{})

	err := ps.Start(nil, nil)
	if err == nil {
		t.Error("expected error when redis config is missing")
	}
}

func TestPubSubClient_Publish(t *testing.T) {
	mockRedis := newMockRedisClient()
	ps := newPubSubClient(&Options{
		Redis: &RedisConfig{Addr: "localhost:6379"},
	})
	ps.redis = mockRedis

	err := ps.Publish("zh-CN")
	if err != nil {
		t.Errorf("unexpected error on Publish: %v", err)
	}

	msgs := mockRedis.published[INVALIDATE_CHANNEL]
	if len(msgs) != 1 || msgs[0] != "zh-CN" {
		t.Errorf("expected published message 'zh-CN', got %v", msgs)
	}
}

func TestPubSubClient_Publish_NotInitialized(t *testing.T) {
	ps := newPubSubClient(&Options{
		Redis: &RedisConfig{Addr: "localhost:6379"},
	})

	err := ps.Publish("zh-CN")
	if err == nil {
		t.Error("expected error when redis not initialized")
	}
}

func TestPubSubClient_OnMessage(t *testing.T) {
	mockRedis := newMockRedisClient()
	cache := newLRUCache(10)
	watchers := newWatcherManager()

	// 预填充缓存
	cache.Set(cacheKey{env: "dev", langCode: "zh-CN"}, map[string]*Template{
		"hello": {Key: "hello", Value: "你好"},
	})

	var notified bool
	var mu sync.Mutex

	// 注册 watcher
	entry := &watcherEntry{
		id: 1,
		handler: func(packVersion int64) {
			mu.Lock()
			defer mu.Unlock()
			notified = true
		},
	}
	watchers.watchers["zh-CN"] = []*watcherEntry{entry}

	ps := newPubSubClient(&Options{
		Redis: &RedisConfig{Addr: "localhost:6379"},
	})
	ps.redis = mockRedis
	ps.active = true

	// 模拟收到消息
	ps.onMessage("zh-CN", watchers, cache)

	// 验证缓存已失效
	if cache.Len() != 0 {
		t.Errorf("expected cache to be empty, got %d", cache.Len())
	}

	// 验证 watcher 被触发
	mu.Lock()
	if !notified {
		t.Error("expected watcher to be notified")
	}
	mu.Unlock()
}

func TestPubSubClient_OnMessage_Batch(t *testing.T) {
	mockRedis := newMockRedisClient()
	cache := newLRUCache(10)
	watchers := newWatcherManager()

	// 预填充缓存
	cache.Set(cacheKey{env: "dev", langCode: "zh-CN"}, map[string]*Template{
		"hello": {Key: "hello", Value: "你好"},
	})
	cache.Set(cacheKey{env: "dev", langCode: "en-US"}, map[string]*Template{
		"hello": {Key: "hello", Value: "Hello"},
	})

	ps := newPubSubClient(&Options{
		Redis: &RedisConfig{Addr: "localhost:6379"},
	})
	ps.redis = mockRedis
	ps.active = true

	// 模拟批量失效消息
	ps.onMessage("zh-CN, en-US", watchers, cache)

	if cache.Len() != 0 {
		t.Errorf("expected all cache to be invalidated, got %d", cache.Len())
	}
}

func TestPubSubClient_IsConnected_False(t *testing.T) {
	ps := newPubSubClient(&Options{})

	if ps.IsConnected() {
		t.Error("expected IsConnected to be false initially")
	}
}

func TestInvalidateChannel_Constant(t *testing.T) {
	expected := "cairobot.i18n.invalidate"
	if INVALIDATE_CHANNEL != expected {
		t.Errorf("expected INVALIDATE_CHANNEL '%s', got '%s'", expected, INVALIDATE_CHANNEL)
	}
}

func TestRedisConfig_DefaultValues(t *testing.T) {
	cfg := &RedisConfig{}

	if cfg.Addr != "" {
		t.Errorf("expected empty addr by default, got '%s'", cfg.Addr)
	}
	if cfg.Password != "" {
		t.Errorf("expected empty password by default, got '%s'", cfg.Password)
	}
	if cfg.DB != 0 {
		t.Errorf("expected DB 0 by default, got %d", cfg.DB)
	}
}

func TestPubSubClient_Stop_NilRedis(t *testing.T) {
	ps := newPubSubClient(&Options{})

	err := ps.Stop()
	if err != nil {
		t.Errorf("unexpected error on Stop with nil redis: %v", err)
	}
}

func TestGoRedisClient_Interface(t *testing.T) {
	// 验证 mockRedisClient 实现了 RedisClient 接口
	var _ RedisClient = (*mockRedisClient)(nil)
}
