package sdk

import (
	"container/list"
	"sync"
	"time"
)

// lruCache L1 进程内 LRU 缓存实现
// 使用 container/list 实现双向链表 + map 索引的 O(1) 查找/淘汰
// 缓存 key 格式: {env}:{module_key}
// 缓存值: *ModuleSnapshot
// 不负责数据加载，由 get.go / remote.go 负责回源
type lruCache struct {
	capacity int
	ttlSec   int
	items    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
}

// cacheEntry 缓存条目，包含值和过期时间
type cacheEntry struct {
	key        string
	value      *ModuleSnapshot
	expireAt   time.Time
}

// newLRUCache 创建 LRU 缓存实例
// capacity: 最大缓存条目数，ttlSec: 过期时间（秒）
func newLRUCache(capacity int, ttlSec ...int) *lruCache {
	ttl := 30
	if len(ttlSec) > 0 && ttlSec[0] > 0 {
		ttl = ttlSec[0]
	}
	return &lruCache{
		capacity: capacity,
		ttlSec:   ttl,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// get 从缓存中获取模块快照
// 命中且未过期返回值 + true，否则返回 nil + false
func (c *lruCache) get(key string) (*ModuleSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}
	entry := elem.Value.(*cacheEntry)
	if time.Now().After(entry.expireAt) {
		return nil, false
	}
	c.order.MoveToFront(elem)
	return entry.value, true
}

// set 向缓存中写入模块快照
// 如果 key 已存在则更新值并移到队头；如果容量已满则淘汰队尾条目
func (c *lruCache) set(key string, value *ModuleSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expireAt = time.Now().Add(time.Duration(c.ttlSec) * time.Second)
		c.order.MoveToFront(elem)
		return
	}
	entry := &cacheEntry{
		key:      key,
		value:    value,
		expireAt: time.Now().Add(time.Duration(c.ttlSec) * time.Second),
	}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
	if c.order.Len() > c.capacity {
		c.evict()
	}
}

// delete 从缓存中删除指定 key
func (c *lruCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, exists := c.items[key]; exists {
		c.order.Remove(elem)
		delete(c.items, key)
	}
}

// evict 淘汰最久未访问的缓存条目（队尾）
// 前置条件：必须持有写锁
func (c *lruCache) evict() {
	elem := c.order.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
	c.order.Remove(elem)
}

// clear 清空所有缓存条目（用于测试或关闭时）
func (c *lruCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// size 返回当前缓存条目数
func (c *lruCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
