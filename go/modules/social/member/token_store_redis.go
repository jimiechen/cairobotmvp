// token_store_redis.go — TokenStore 的 Redis 实现
// 使用 Redis SETEX 存储黑名单，TTL 自动过期释放
// 适用场景：生产环境、集成测试（替代 MemoryTokenStore）
package member

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore 基于 Redis 的令牌黑名单存储
type RedisTokenStore struct {
	client    *redis.Client
	keyPrefix string // key 前缀，避免与其他业务冲突
}

// NewRedisTokenStore 创建 Redis 令牌黑名单存储
func NewRedisTokenStore(client *redis.Client, keyPrefix string) *RedisTokenStore {
	if keyPrefix == "" {
		keyPrefix = "social:tl:"
	}
	return &RedisTokenStore{client: client, keyPrefix: keyPrefix}
}

// Blacklist 将令牌加入 Redis 黑名单，设置 TTL 自动过期
func (s *RedisTokenStore) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	key := s.keyPrefix + token
	return s.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查令牌是否在 Redis 黑名单中
func (s *RedisTokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := s.keyPrefix + token
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return result > 0, nil
}
