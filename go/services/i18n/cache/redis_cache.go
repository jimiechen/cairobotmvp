package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// RedisClient 抽象接口，解耦与具体 Redis 实现的依赖
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
	Close() error
}

// RedisCache Cache 接口的 Redis 实现
type RedisCache struct {
	client RedisClient
	ttl    time.Duration
}

// NewRedisCache 创建 Redis 缓存实例
// 注意：实际 RedisClient 实现应在 main 包或初始化代码中提供
func NewRedisCache(cfg *config.RedisConfig, ttlSeconds int, client RedisClient) (*RedisCache, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("redis disabled")
	}

	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	return &RedisCache{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}, nil
}

// GetPack 从缓存获取语言包
func (c *RedisCache) GetPack(langCode string, env string) (*domain.LangPack, bool) {
	key := fmt.Sprintf("i18n:pack:%s:%s", env, langCode)
	val, err := c.client.Get(key)
	if err != nil {
		return nil, false
	}

	var pack domain.LangPack
	if err := json.Unmarshal([]byte(val), &pack); err != nil {
		return nil, false
	}
	return &pack, true
}

// SetPack 将语言包写入缓存
func (c *RedisCache) SetPack(langCode string, env string, pack *domain.LangPack) {
	key := fmt.Sprintf("i18n:pack:%s:%s", env, langCode)
	bytes, err := json.Marshal(pack)
	if err != nil {
		return
	}
	c.client.Set(key, string(bytes), c.ttl)
}

// GetStrings 从缓存获取字符串列表
func (c *RedisCache) GetStrings(packID int64) ([]domain.LangString, bool) {
	key := fmt.Sprintf("i18n:strings:%d", packID)
	val, err := c.client.Get(key)
	if err != nil {
		return nil, false
	}

	var strings []domain.LangString
	if err := json.Unmarshal([]byte(val), &strings); err != nil {
		return nil, false
	}
	return strings, true
}

// SetStrings 将字符串列表写入缓存
func (c *RedisCache) SetStrings(packID int64, strings []domain.LangString) {
	key := fmt.Sprintf("i18n:strings:%d", packID)
	bytes, err := json.Marshal(strings)
	if err != nil {
		return
	}
	c.client.Set(key, string(bytes), c.ttl)
}

// Invalidate 使缓存失效
func (c *RedisCache) Invalidate(langCode string, env string) {
	key := fmt.Sprintf("i18n:pack:%s:%s", env, langCode)
	c.client.Delete(key)
}

// Close 关闭 Redis 连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}
