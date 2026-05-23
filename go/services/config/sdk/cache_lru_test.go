package sdk

import (
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func TestNewLRUCache(t *testing.T) {
	cache := newLRUCache(10)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.capacity != 10 {
		t.Fatalf("expected capacity 10, got %d", cache.capacity)
	}
}

func TestLRUCache_GetSet(t *testing.T) {
	cache := newLRUCache(10)
	snapshot := &ModuleSnapshot{
		ModuleKey: "test_module",
		Fields:    make(map[string]*domain.TypedValue),
	}
	cache.set("test:key", snapshot)
	got, ok := cache.get("test:key")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.ModuleKey != "test_module" {
		t.Fatalf("expected module_key test_module, got %s", got.ModuleKey)
	}
}

func TestLRUCache_Miss(t *testing.T) {
	cache := newLRUCache(10)
	_, ok := cache.get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestLRUCache_Overwrite(t *testing.T) {
	cache := newLRUCache(10)
	snapshot1 := &ModuleSnapshot{ModuleKey: "v1", Fields: make(map[string]*domain.TypedValue)}
	snapshot2 := &ModuleSnapshot{ModuleKey: "v2", Fields: make(map[string]*domain.TypedValue)}
	cache.set("key", snapshot1)
	cache.set("key", snapshot2)
	got, _ := cache.get("key")
	if got.ModuleKey != "v2" {
		t.Fatalf("expected v2, got %s", got.ModuleKey)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := newLRUCache(3)
	for i := 0; i < 4; i++ {
		snapshot := &ModuleSnapshot{
			ModuleKey: string(rune('a' + i)),
			Fields:    make(map[string]*domain.TypedValue),
		}
		cache.set(string(rune('a'+i)), snapshot)
	}
	if cache.size() != 3 {
		t.Fatalf("expected size 3, got %d", cache.size())
	}
	_, ok := cache.get("a")
	if ok {
		t.Fatal("expected 'a' to be evicted")
	}
	_, ok = cache.get("d")
	if !ok {
		t.Fatal("expected 'd' to exist")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := newLRUCache(10, 1)
	snapshot := &ModuleSnapshot{ModuleKey: "ttl_test", Fields: make(map[string]*domain.TypedValue)}
	cache.set("ttl:key", snapshot)
	got, ok := cache.get("ttl:key")
	if !ok {
		t.Fatal("expected cache hit before TTL expiry")
	}
	if got.ModuleKey != "ttl_test" {
		t.Fatalf("expected ttl_test, got %s", got.ModuleKey)
	}
	time.Sleep(2 * time.Second)
	_, ok = cache.get("ttl:key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := newLRUCache(10)
	snapshot := &ModuleSnapshot{ModuleKey: "delete_test", Fields: make(map[string]*domain.TypedValue)}
	cache.set("del:key", snapshot)
	cache.delete("del:key")
	_, ok := cache.get("del:key")
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := newLRUCache(10)
	for i := 0; i < 5; i++ {
		snapshot := &ModuleSnapshot{
			ModuleKey: string(rune('a' + i)),
			Fields:    make(map[string]*domain.TypedValue),
		}
		cache.set(string(rune('a'+i)), snapshot)
	}
	cache.clear()
	if cache.size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", cache.size())
	}
}

func TestLRUCache_AccessOrder(t *testing.T) {
	cache := newLRUCache(3)
	for i := 0; i < 3; i++ {
		snapshot := &ModuleSnapshot{
			ModuleKey: string(rune('a' + i)),
			Fields:    make(map[string]*domain.TypedValue),
		}
		cache.set(string(rune('a'+i)), snapshot)
	}
	cache.get("a")
	snapshotD := &ModuleSnapshot{ModuleKey: "d", Fields: make(map[string]*domain.TypedValue)}
	cache.set("d", snapshotD)
	_, ok := cache.get("b")
	if ok {
		t.Fatal("expected 'b' to be evicted (not recently accessed)")
	}
	_, ok = cache.get("a")
	if !ok {
		t.Fatal("expected 'a' to exist (recently accessed)")
	}
}
