package sdk

import (
	"container/list"
	"testing"
	"time"
)

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := newLRUCache(10)

	key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}
	value := map[string]*Template{
		"key1": {Key: "key1", Value: "值1", TemplateType: "plain"},
	}

	cache.Set(key, value)

	result, found, expired := cache.Get(key)
	if !found {
		t.Error("expected to find cached value")
	}
	if expired {
		t.Error("expected value not expired")
	}
	if result["key1"].Value != "值1" {
		t.Errorf("expected '值1', got '%s'", result["key1"].Value)
	}
}

func TestLRUCache_Get_NotFound(t *testing.T) {
	cache := newLRUCache(10)

	key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 999}
	_, found, _ := cache.Get(key)
	if found {
		t.Error("expected not found")
	}
}

func TestLRUCache_TTL_Expiration(t *testing.T) {
	cache := &lruCache{
		capacity: 10,
		ttl:      50 * time.Millisecond,
		items:    make(map[cacheKey]*cacheEntry),
		evictList: newListForTest(),
	}

	key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}
	value := map[string]*Template{"key1": {Key: "key1", Value: "test"}}
	cache.Set(key, value)

	_, found, expired := cache.Get(key)
	if !found {
		t.Error("expected found before expiration")
	}
	if expired {
		t.Error("expected not expired immediately")
	}

	time.Sleep(60 * time.Millisecond)

	_, found, expired = cache.Get(key)
	if !found {
		t.Error("expected found after expiration (entry exists)")
	}
	if !expired {
		t.Error("expected expired after TTL")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := newLRUCache(2)

	for i := 0; i < 3; i++ {
		key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: int64(i)}
		value := map[string]*Template{string(rune('a' + i)): {}}
		cache.Set(key, value)
	}

	if cache.Len() != 2 {
		t.Errorf("expected capacity 2, got %d", cache.Len())
	}

	oldestKey := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 0}
	_, found, _ := cache.Get(oldestKey)
	if found {
		t.Error("expected oldest entry to be evicted")
	}
}

func TestLRUCache_Update_RefreshesTTL(t *testing.T) {
	cache := newLRUCache(10)

	key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}
	value := map[string]*Template{"key1": {Value: "original"}}
	cache.Set(key, value)

	newValue := map[string]*Template{"key1": {Value: "updated"}}
	cache.Set(key, newValue)

	result, found, _ := cache.Get(key)
	if !found {
		t.Error("expected found after update")
	}
	if result["key1"].Value != "updated" {
		t.Errorf("expected 'updated', got '%s'", result["key1"].Value)
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := newLRUCache(10)

	key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}
	value := map[string]*Template{"key1": {}}
	cache.Set(key, value)

	cache.Delete(key)

	_, found, _ := cache.Get(key)
	if found {
		t.Error("expected not found after delete")
	}
}

func TestLRUCache_InvalidateByLangCode(t *testing.T) {
	cache := newLRUCache(10)

	key1 := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}
	key2 := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 2}
	key3 := cacheKey{env: "dev", langCode: "en-US", packVersion: 1}

	cache.Set(key1, map[string]*Template{})
	cache.Set(key2, map[string]*Template{})
	cache.Set(key3, map[string]*Template{})

	cache.InvalidateByLangCode("zh-CN")

	if cache.Len() != 1 {
		t.Errorf("expected 1 entry after invalidation, got %d", cache.Len())
	}

	_, foundEn, _ := cache.Get(key3)
	if !foundEn {
		t.Error("expected en-US entry to remain")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := newLRUCache(10)

	for i := 0; i < 5; i++ {
		key := cacheKey{env: "dev", langCode: "zh-CN", packVersion: int64(i)}
		cache.Set(key, map[string]*Template{})
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", cache.Len())
	}
}

func TestCacheKey_String(t *testing.T) {
	tests := []struct {
		key      cacheKey
		expected string
	}{
		{cacheKey{env: "dev", langCode: "zh-CN", packVersion: 1}, "dev:zh-CN:1"},
		{cacheKey{env: "prod", langCode: "en-US", packVersion: 0}, "prod:en-US:latest"},
	}

	for _, tt := range tests {
		result := tt.key.String()
		if result != tt.expected {
			t.Errorf("expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestLRUCache_DefaultCapacity(t *testing.T) {
	cache := newLRUCache(0)
	if cache.capacity != DEFAULT_CACHE_CAPACITY {
		t.Errorf("expected default capacity %d, got %d", DEFAULT_CACHE_CAPACITY, cache.capacity)
	}
}

func TestLRUCache_LRU_Order(t *testing.T) {
	cache := newLRUCache(3)

	keys := make([]cacheKey, 3)
	for i := 0; i < 3; i++ {
		keys[i] = cacheKey{env: "dev", langCode: "zh-CN", packVersion: int64(i)}
		cache.Set(keys[i], map[string]*Template{})
	}

	cache.Get(keys[0])

	newKey := cacheKey{env: "dev", langCode: "zh-CN", packVersion: 99}
	cache.Set(newKey, map[string]*Template{})

	if cache.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", cache.Len())
	}

	_, found0, _ := cache.Get(keys[0])
	if !found0 {
		t.Error("expected keys[0] to survive (recently accessed)")
	}

	_, found1, _ := cache.Get(keys[1])
	if found1 {
		t.Error("expected keys[1] to be evicted (oldest)")
	}
}

func newListForTest() *list.List {
	l := list.New()
	l.Init()
	return l
}
