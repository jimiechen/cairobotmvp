package sdk

import (
	"sync"
	"testing"
	"time"
)

func TestWatcherManager_Trigger_SingleWatcher(t *testing.T) {
	wm := newWatcherManager()

	received := make(chan int64, 1)
	cancel := wm.registerTestWatcher("zh-CN", func(v int64) { received <- v })
	defer cancel()

	wm.Trigger("zh-CN", 42)

	select {
	case v := <-received:
		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for callback")
	}
}

func TestWatcherManager_Trigger_MultipleWatchers(t *testing.T) {
	wm := newWatcherManager()

	var wg sync.WaitGroup
	results := make([]int64, 0, 3)
	var mu sync.Mutex

	wg.Add(3)
	for i := 0; i < 3; i++ {
		handler := func(v int64) {
			mu.Lock()
			results = append(results, v)
			mu.Unlock()
			wg.Done()
		}
	 wm.registerTestWatcher("en-US", handler)
	}

	wm.Trigger("en-US", 99)

	wg.Wait()

	if len(results) != 3 {
		t.Fatalf("expected 3 callbacks, got %d", len(results))
	}
	for _, v := range results {
		if v != 99 {
			t.Errorf("expected all callbacks to receive 99, got %d", v)
		}
	}
}

func TestWatcherManager_Remove(t *testing.T) {
	wm := newWatcherManager()

	callCount := 0
	cancel1 := wm.registerTestWatcher("zh-CN", func(int64) { callCount++ })
	_ = wm.registerTestWatcher("zh-CN", func(int64) { callCount++ })

	cancel1()

	wm.Trigger("zh-CN", 1)

	if callCount != 1 {
		t.Errorf("expected 1 callback after remove, got %d", callCount)
	}
}

func TestWatcherManager_RemoveLast_CleansUp(t *testing.T) {
	wm := newWatcherManager()

	cancel := wm.registerTestWatcher("zh-CN", func(int64) {})

	cancel()

	wm.mu.RLock()
	_, exists := wm.watchers["zh-CN"]
	wm.mu.RUnlock()
	if exists {
		t.Error("expected language key to be cleaned up after last watcher removed")
	}
}

func TestWatcherManager_Trigger_NonexistentLangCode(t *testing.T) {
	wm := newWatcherManager()

	notCalled := true
	_ = wm.registerTestWatcher("zh-CN", func(int64) { notCalled = false })

	wm.Trigger("en-US", 1)

	if !notCalled {
		t.Error("triggering nonexistent lang code should not affect other watchers")
	}
}

func TestWatcherManager_ConcurrentAccess(t *testing.T) {
	wm := newWatcherManager()

	var wg sync.WaitGroup
	cancels := make([]func(), 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				wm.Trigger("zh-CN", int64(id))
			} else if id < len(cancels) && cancels[id] != nil {
				cancels[id]()
			}
		}(i)
		if cancels[i] == nil {
			cancels[i] = wm.registerTestWatcher("zh-CN", func(int64) {})
		}
	}

	wg.Wait()
}

// registerTestWatcher 辅助方法：直接注册观察者（绕过 Client 接口）
func (w *watcherManager) registerTestWatcher(langCode string, handler func(packVersion int64)) func() {
	entry := &watcherEntry{
		id:      w.nextID,
		handler: handler,
	}
	w.nextID++

	w.mu.Lock()
	w.watchers[langCode] = append(w.watchers[langCode], entry)
	w.mu.Unlock()

	cancelFunc := func() {
		w.remove(langCode, entry.id)
	}

	return cancelFunc
}
