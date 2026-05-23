package sdk

import (
	"sync"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func TestWatch_RegisterAndNotify(t *testing.T) {
	watcher := newModuleWatcher()
	var received *ModuleSnapshot
	var mu sync.Mutex
	cancel := watcher.register("test_module", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		received = snapshot
	})
	snapshot := &ModuleSnapshot{
		ModuleKey: "test_module",
		Version:   42,
	}
	watcher.notify("test_module", snapshot)
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if received == nil {
		t.Fatal("expected to receive notification")
	}
	if received.Version != 42 {
		t.Fatalf("expected version 42, got %d", received.Version)
	}
	cancel()
}

func TestWatch_Cancel(t *testing.T) {
	watcher := newModuleWatcher()
	callCount := 0
	var mu sync.Mutex
	cancel := watcher.register("test_module", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
	})
	watcher.notify("test_module", &ModuleSnapshot{ModuleKey: "test_module"})
	time.Sleep(50 * time.Millisecond)
	cancel()
	watcher.notify("test_module", &ModuleSnapshot{ModuleKey: "test_module"})
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Fatalf("expected 1 call after cancel, got %d", callCount)
	}
}

func TestWatch_MultipleHandlers(t *testing.T) {
	watcher := newModuleWatcher()
	var count1, count2 int
	var mu sync.Mutex
	watcher.register("test_module", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		count1++
	})
	watcher.register("test_module", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		count2++
	})
	watcher.notify("test_module", &ModuleSnapshot{ModuleKey: "test_module"})
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if count1 != 1 || count2 != 1 {
		t.Fatalf("expected both handlers called (count1=%d, count2=%d)", count1, count2)
	}
}

func TestWatch_DifferentModules(t *testing.T) {
	watcher := newModuleWatcher()
	var received string
	var mu sync.Mutex
	watcher.register("module_a", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		received = "a"
	})
	watcher.register("module_b", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		received = "b"
	})
	watcher.notify("module_a", &ModuleSnapshot{ModuleKey: "module_a"})
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if received != "a" {
		t.Fatalf("expected module_a handler called, got %s", received)
	}
}

func TestClient_Watch(t *testing.T) {
	svc := &mockConfigService{
		configs: map[string]map[string]*domain.TypedValue{},
	}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	var received *ModuleSnapshot
	var mu sync.Mutex
	cancel := client.Watch("watch_test", func(snapshot *ModuleSnapshot) {
		mu.Lock()
		defer mu.Unlock()
		received = snapshot
	})
	client.(*configClient).watcher.notify("watch_test", &ModuleSnapshot{
		ModuleKey: "watch_test",
		Version:   100,
	})
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if received == nil {
		t.Fatal("expected to receive notification via Client.Watch")
	}
	if received.Version != 100 {
		t.Fatalf("expected version 100, got %d", received.Version)
	}
	cancel()
}
