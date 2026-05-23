package cache

import (
	"fmt"
	"sync"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// MockCache 内存 Map 缓存实现
// 用于开发和测试环境，使用内存 Map 存储缓存数据
//
// 职责：
// - 实现 Cache 接口的内存版本
// - 提供线程安全的缓存读写操作
//
// 不负责：
// - 持久化（重启后数据丢失）
// - 过期策略（需要手动 Invalidate）
type MockCache struct {
	packs   map[string]*domain.LangPack
	strings map[int64][]domain.LangString
	mu      sync.RWMutex
}

// NewMockCache 创建 Mock 缓存实例
func NewMockCache() *MockCache {
	return &MockCache{
		packs:   make(map[string]*domain.LangPack),
		strings: make(map[int64][]domain.LangString),
	}
}

// packKey 生成语言包的缓存键
func packKey(langCode string, env string) string {
	return fmt.Sprintf("%s:%s", langCode, env)
}

// GetPack 从缓存获取语言包
func (c *MockCache) GetPack(langCode string, env string) (*domain.LangPack, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := packKey(langCode, env)
	pack, exists := c.packs[key]
	return pack, exists
}

// SetPack 将语言包写入缓存
func (c *MockCache) SetPack(langCode string, env string, pack *domain.LangPack) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := packKey(langCode, env)
	c.packs[key] = pack
}

// GetStrings 从缓存获取字符串列表
func (c *MockCache) GetStrings(packID int64) ([]domain.LangString, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	strings, exists := c.strings[packID]
	return strings, exists
}

// SetStrings 将字符串列表写入缓存
func (c *MockCache) SetStrings(packID int64, strings []domain.LangString) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.strings[packID] = strings
}

// Invalidate 使缓存失效
func (c *MockCache) Invalidate(langCode string, env string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := packKey(langCode, env)
	delete(c.packs, key)

	for pid := range c.strings {
		delete(c.strings, pid)
	}
}
