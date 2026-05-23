package sdk

import (
	"sync"
)

// watcherManager 变更观察者管理器
//
// 职责：
// - 管理语言包版本变更的回调注册和触发
// - 支持取消订阅
// - 保证并发安全
//
// 不负责：
// - 缓存失效（由调用方协调）
// - 消息来源（由 pubsub 或远程层负责）
type watcherManager struct {
	mu       sync.RWMutex
	watchers map[string][]*watcherEntry
	nextID   int64
}

type watcherEntry struct {
	id      int64
	handler func(packVersion int64)
}

func newWatcherManager() *watcherManager {
	return &watcherManager{
		watchers: make(map[string][]*watcherEntry),
	}
}

// Watch 订阅语言包版本变更
//
// 参数：
// - langCode: 要监听的语言代码
// - handler: 版本变更时的回调函数，参数为新的 packVersion
//
// 返回：
// - cancel: 取消订阅的函数
func (c *clientImpl) Watch(langCode string, handler func(packVersion int64)) (cancel func()) {
	entry := &watcherEntry{
		id:      c.watchers.nextID,
		handler: handler,
	}
	c.watchers.nextID++

	c.watchers.mu.Lock()
	defer c.watchers.mu.Unlock()

	c.watchers.watchers[langCode] = append(c.watchers.watchers[langCode], entry)

	cancelFunc := func() {
		c.watchers.remove(langCode, entry.id)
	}

	return cancelFunc
}

// Trigger 触发指定语言代码的所有观察者（供内部使用）
func (w *watcherManager) Trigger(langCode string, packVersion int64) {
	w.mu.RLock()
	entries, exists := w.watchers[langCode]
	w.mu.RUnlock()

	if !exists {
		return
	}

	for _, entry := range entries {
		if entry.handler != nil {
			entry.handler(packVersion)
		}
	}
}

// remove 移除指定的观察者（必须在持有写锁时调用）
func (w *watcherManager) remove(langCode string, id int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entries, exists := w.watchers[langCode]
	if !exists {
		return
	}

	filtered := make([]*watcherEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.id != id {
		 filtered = append(filtered, entry)
		}
	}

	if len(filtered) == 0 {
		delete(w.watchers, langCode)
	} else {
		w.watchers[langCode] = filtered
	}
}
