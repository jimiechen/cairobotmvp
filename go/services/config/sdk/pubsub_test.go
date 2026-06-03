package sdk

import (
	"sync"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/testutil"
)

type redisClientAdapter struct {
	client *testutil.MockRedisClient
}

func (a *redisClientAdapter) Get(key string) (string, error) {
	return a.client.Get(key)
}

func (a *redisClientAdapter) Set(key string, value string, ttlSec int) error {
	return a.client.Set(key, value, ttlSec)
}

func (a *redisClientAdapter) Delete(key string) error {
	return a.client.Delete(key)
}

func (a *redisClientAdapter) Subscribe(channel string, handler MessageHandler) (CancelFunc, error) {
	cancel, err := a.client.Subscribe(channel, func(msg string) { handler(msg) })
	if err != nil {
		return nil, err
	}
	return func() { cancel() }, nil
}

func (a *redisClientAdapter) PublishMessage(channel string, message string) {
	a.client.PublishMessage(channel, message)
}

func newTestRedis() *redisClientAdapter {
	return &redisClientAdapter{client: testutil.NewMockRedisClient()}
}

func TestPubsubManager_StartStop(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	pm := newPubsubManager(redis, cache, watcher)
	if !pm.active {
		t.Fatal("expected pubsub manager to be active after start")
	}
	pm.stop()
	if pm.active {
		t.Fatal("expected pubsub manager to be inactive after stop")
	}
}

func TestPubsubManager_OnMessage(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	snapshot := &ModuleSnapshot{
		ModuleKey: "test_module",
		Version:   1,
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("sdk:test_module", snapshot)
	var notified bool
	var mu sync.Mutex
	watcher.register("test_module", func(s *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		notified = true
	})
	pm := newPubsubManager(redis, cache, watcher)
	redis.PublishMessage(pubsubChannel, "test_module")
	time.Sleep(50 * time.Millisecond)
	_, ok := cache.get("sdk:test_module")
	if ok {
		t.Fatal("expected cache to be invalidated")
	}
	mu.Lock()
	defer mu.Unlock()
	if !notified {
		t.Fatal("expected watcher to be notified")
	}
	pm.stop()
}

func TestPubsubManager_BatchInvalidate(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	for _, key := range []string{"mod_a", "mod_b"} {
		snapshot := &ModuleSnapshot{
			ModuleKey: key,
			Fields:    make(map[string]*domain.TypedValue),
		}
		cache.set("sdk:"+key, snapshot)
	}
	pm := newPubsubManager(redis, cache, watcher)
	redis.PublishMessage(pubsubChannel, "mod_a, mod_b")
	time.Sleep(50 * time.Millisecond)
	for _, key := range []string{"mod_a", "mod_b"} {
		_, ok := cache.get("sdk:" + key)
		if ok {
			t.Fatalf("expected %s to be invalidated", key)
		}
	}
	pm.stop()
}

func TestBuildCacheKey(t *testing.T) {
	key := buildCacheKey("base_cfg")
	if key != "sdk:base_cfg" {
		t.Fatalf("expected sdk:base_cfg, got %s", key)
	}
}

func TestOnMessage_JsonStructured(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	snapshot := &ModuleSnapshot{
		ModuleKey: "hello_cfg",
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("sdk:hello_cfg", snapshot)

	var notified bool
	var mu sync.Mutex
	watcher.register("hello_cfg", func(s *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		notified = true
	})

	pm := newPubsubManager(redis, cache, watcher)

	jsonPayload := `{"tenant_id":"default","scope":"config","module_keys":["hello_cfg"],"version":1,"timestamp":1716739200}`
	redis.PublishMessage(pubsubChannel, jsonPayload)
	time.Sleep(50 * time.Millisecond)

	_, ok := cache.get("sdk:hello_cfg")
	if ok {
		t.Fatal("expected cache invalidated by structured event")
	}
	mu.Lock()
	defer mu.Unlock()
	if !notified {
		t.Fatal("expected watcher notified")
	}
	pm.stop()
}

func TestOnMessage_JsonMissingTenantId_Fallback(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	snapshot := &ModuleSnapshot{
		ModuleKey: "legacy_mod",
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("sdk:legacy_mod", snapshot)

	pm := newPubsubManager(redis, cache, watcher)

	jsonNoTenant := `{"scope":"config","module_keys":["legacy_mod"]}`
	redis.PublishMessage(pubsubChannel, jsonNoTenant)
	time.Sleep(50 * time.Millisecond)

	_, ok := cache.get("sdk:legacy_mod")
	if ok {
		t.Fatal("expected cache invalidated (fallback to legacy)")
	}
	pm.stop()
}

func TestOnMessage_LegacyCommaFormat(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	snapshot := &ModuleSnapshot{
		ModuleKey: "old_mod",
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("sdk:old_mod", snapshot)

	pm := newPubsubManager(redis, cache, watcher)

	redis.PublishMessage(pubsubChannel, "old_mod")
	time.Sleep(50 * time.Millisecond)

	_, ok := cache.get("sdk:old_mod")
	if ok {
		t.Fatal("expected cache invalidated by legacy format")
	}
	pm.stop()
}

func TestOnMessage_InvalidJson_Fallback(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()
	snapshot := &ModuleSnapshot{
		ModuleKey: "broken_mod",
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("sdk:broken_mod", snapshot)

	pm := newPubsubManager(redis, cache, watcher)

	redis.PublishMessage(pubsubChannel, "{invalid json")
	time.Sleep(50 * time.Millisecond)

	_, ok := cache.get("sdk:broken_mod")
	if !ok {
		t.Fatal("unrecognizable input must NOT invalidate unrelated caches")
	}
	pm.stop()
}

func TestOnMessage_Empty_NoPanic(t *testing.T) {
	redis := newTestRedis()
	cache := newLRUCache(10)
	watcher := newModuleWatcher()

	pm := newPubsubManager(redis, cache, watcher)

	redis.PublishMessage(pubsubChannel, "")
	time.Sleep(50 * time.Millisecond)
	pm.stop()
}
