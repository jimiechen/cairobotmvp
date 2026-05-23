package sdk

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

const (
	// DEFAULT_CACHE_CAPACITY 默认缓存容量（语言包数量）
	DEFAULT_CACHE_CAPACITY = 128
	// DEFAULT_CACHE_TTL 默认缓存过期时间（秒）
	DEFAULT_CACHE_TTL = 60
)

// cacheKey 缓存键格式: {env}:{lang_code}:{pack_version}
type cacheKey struct {
	env         string
	langCode    string
	packVersion int64
}

func (k cacheKey) String() string {
	if k.packVersion > 0 {
		return fmt.Sprintf("%s:%s:%d", k.env, k.langCode, k.packVersion)
	}
	return fmt.Sprintf("%s:%s:latest", k.env, k.langCode)
}

// cacheEntry 缓存条目
type cacheEntry struct {
	key        cacheKey
	value      map[string]*Template // key → Template 映射
	expireAt   time.Time
	listElem   *list.Element
}

// lruCache 线程安全 LRU 缓存
// 使用 container/list 实现 O(1) 的访问和淘汰
//
// 职责：
// - 缓存已加载的语言包（key → Template 映射）
// - 提供 TTL 过期和 LRU 淘汰机制
// - 保证并发安全
//
// 不负责：
// - 语言包加载逻辑（由 translate 层负责）
// - 远程缓存同步（由 pubsub 层负责）
type lruCache struct {
	capacity int
	ttl      time.Duration
	items    map[cacheKey]*cacheEntry
	evictList *list.List
	mu       sync.RWMutex
}

// newLRUCache 创建 LRU 缓存实例
func newLRUCache(capacity int) *lruCache {
	if capacity <= 0 {
		capacity = DEFAULT_CACHE_CAPACITY
	}
	return &lruCache{
		capacity: capacity,
		ttl:      DEFAULT_CACHE_TTL * time.Second,
		items:    make(map[cacheKey]*cacheEntry),
		evictList: list.New(),
	}
}

// Get 获取缓存中的模板映射
// 返回 (value, found, expired)
func (c *lruCache) Get(key cacheKey) (map[string]*Template, bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false, false
	}

	if time.Now().After(entry.expireAt) {
		return entry.value, true, true
	}

	c.evictList.MoveToFront(entry.listElem)
	return entry.value, true, false
}

// Set 写入缓存
// 如果容量已满，淘汰最久未使用的条目
func (c *lruCache) Set(key cacheKey, value map[string]*Template) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, exists := c.items[key]; exists {
		existing.value = value
		existing.expireAt = time.Now().Add(c.ttl)
		c.evictList.MoveToFront(existing.listElem)
		return
	}

	for len(c.items) >= c.capacity {
		c.evictOldest()
	}

	entry := &cacheEntry{
		key:      key,
		value:    value,
		expireAt: time.Now().Add(c.ttl),
		listElem: c.evictList.PushFront(key),
	}
	c.items[key] = entry
}

// Delete 删除指定缓存条目
func (c *lruCache) Delete(key cacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.items[key]; exists {
		c.evictList.Remove(entry.listElem)
		delete(c.items, key)
	}
}

// InvalidateByLangCode 使指定语言代码的所有版本缓存失效
func (c *lruCache) InvalidateByLangCode(langCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.items {
		if key.langCode == langCode {
			c.evictList.Remove(entry.listElem)
			delete(c.items, key)
		}
	}
}

// Clear 清空所有缓存
func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[cacheKey]*cacheEntry)
	c.evictList.Init()
}

// Len 返回当前缓存条目数
func (c *lruCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// evictOldest 淘汰最久未使用的条目（必须在持有写锁时调用）
func (c *lruCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		key := elem.Value.(cacheKey)
		c.evictList.Remove(elem)
		delete(c.items, key)
	}
}
