package sdk

import (
	"sync"
)

// moduleWatcher 模块变更订阅管理器
// 维护 moduleKey → handler 列表的映射，支持注册/取消/触发
// InProcess 模式下由外部调用 Notify 触发，Remote 模式下由 pubsub.go 触发
type moduleWatcher struct {
	handlers map[string][]*watchHandler
	mu       sync.RWMutex
}

// watchHandler 单个订阅句柄，包含回调函数和是否活跃的标记
type watchHandler struct {
	id      uint64
	handler func(*ModuleSnapshot)
	active  bool
}

// newModuleWatcher 创建模块变更订阅管理器
func newModuleWatcher() *moduleWatcher {
	return &moduleWatcher{
		handlers: make(map[string][]*watchHandler),
	}
}

// register 注册模块变更监听
// 返回取消函数，调用后移除该监听
func (w *moduleWatcher) register(moduleKey string, handler func(*ModuleSnapshot)) func() {
	w.mu.Lock()
	defer w.mu.Unlock()
	h := &watchHandler{
		id:      nextHandlerID(),
		handler: handler,
		active:  true,
	}
	w.handlers[moduleKey] = append(w.handlers[moduleKey], h)
	return func() {
		w.unregister(moduleKey, h.id)
	}
}

// unregister 移除指定 ID 的监听
func (w *moduleWatcher) unregister(moduleKey string, id uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	handlers := w.handlers[moduleKey]
	for i, h := range handlers {
		if h.id == id {
			h.active = false
			w.handlers[moduleKey] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	if len(w.handlers[moduleKey]) == 0 {
		delete(w.handlers, moduleKey)
	}
}

// notify 触发指定模块的所有活跃监听
// 由外部（pubsub 或测试）调用，传入最新的 ModuleSnapshot
func (w *moduleWatcher) notify(moduleKey string, snapshot *ModuleSnapshot) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, h := range w.handlers[moduleKey] {
		if h.active {
			go h.handler(snapshot)
		}
	}
}

// Watch 注册模块变更订阅
// 当模块版本变更时触发 handler 回调，返回取消函数
func (c *configClient) Watch(moduleKey string, handler func(*ModuleSnapshot)) func() {
	return c.watcher.register(moduleKey, handler)
}

var (
	handlerCounter uint64
	handlerMu      sync.Mutex
)

// nextHandlerID 生成全局唯一的 handler ID
func nextHandlerID() uint64 {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	handlerCounter++
	return handlerCounter
}
