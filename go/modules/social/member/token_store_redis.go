// token_store_redis.go — TokenStore 的 Redis 实现
// 使用 Redis SETEX 存储黑名单（key=jti），TTL 自动过期释放
// 适用场景：生产环境、集成测试（替代 MemoryTokenStore）
package member

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore 基于 Redis 的令牌黑名单存储
// key 格式: {prefix}{jti}，不存储完整 token
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

// Store 将 jti 加入 Redis 黑名单，设置 TTL 自动过期（ttl 单位：秒）
func (s *RedisTokenStore) Store(ctx context.Context, jti string, ttlSec int64) error {
	key := s.keyPrefix + jti
	return s.client.Set(ctx, key, "1", time.Duration(ttlSec)*time.Second).Err()
}

// Exists 检查 jti 是否在 Redis 黑名单中
func (s *RedisTokenStore) Exists(ctx context.Context, jti string) (bool, error) {
	key := s.keyPrefix + jti
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return result > 0, nil
}

// Delete 从 Redis 黑名单移除指定 jti
func (s *RedisTokenStore) Delete(ctx context.Context, jti string) error {
	key := s.keyPrefix + jti
	return s.client.Del(ctx, key).Err()
}

// ========== 兼容性别名（逐步废弃，供过渡期使用）==========

// Blacklist 兼容旧接口：从 token 提取 jti 后存入 Redis
func (s *RedisTokenStore) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	jti := extractJTIFromToken(token)
	if jti == "" {
		jti = token // fallback
	}
	return s.Store(ctx, jti, int64(ttl.Seconds()))
}

// IsBlacklisted 兼容旧接口
func (s *RedisTokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	jti := extractJTIFromToken(token)
	if jti == "" {
		jti = token
	}
	return s.Exists(ctx, jti)
}
