package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
)

// RedisClient 抽象接口，解耦与具体 Redis 实现的依赖
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
	Scan(prefix string) ([]string, error)
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

// Get 从 Redis 获取缓存值
func (c *RedisCache) Get(key string) (any, bool) {
	val, err := c.client.Get(key)
	if err != nil {
		return nil, false
	}

	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}
	return result, true
}

// Set 将值存入 Redis 缓存
func (c *RedisCache) Set(key string, value any) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(key, string(bytes), c.ttl)
}

// Delete 从 Redis 删除指定 key
func (c *RedisCache) Delete(key string) {
	c.client.Delete(key)
}

// Invalidate 按前缀批量删除缓存
func (c *RedisCache) Invalidate(prefix string) {
	keys, err := c.client.Scan(prefix + ":*")
	if err != nil {
		return
	}
	for _, key := range keys {
		c.client.Delete(key)
	}
}

// Close 关闭 Redis 连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}
