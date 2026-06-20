// token_store_redis_test.go — RedisTokenStore 集成测试
// 直连真实 Redis（192.168.1.6:6379 DB=2），验证黑名单 CRUD 和 TTL 过期
// 运行条件：需设置 REDIS_HOST 环境变量

package member

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const redisTestPrefix = "test_tl_" // 测试用 key 前缀，避免污染生产数据

func setupRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		t.Skip("跳过 Redis 集成测试：未设置 REDIS_HOST（需 source .env.local）")
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		fmt.Sscanf(v, "%d", &db)
	}

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
		DB:   db,
	})

	require.NoError(t, client.Ping(context.Background()).Err())
	return client
}

func cleanupRedisKeys(t *testing.T, rdb *redis.Client) {
	t.Helper()
	ctx := context.Background()
	iter := rdb.Scan(ctx, 0, redisTestPrefix+"*", 10).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		rdb.Del(ctx, key)
	}
}

// ========== Blacklist / IsBlacklisted 核心路径 ==========

func TestRedisTokenStore_BlacklistAndCheck(t *testing.T) {
	rdb := setupRedisClient(t)
	defer cleanupRedisKeys(t, rdb)
	store := NewRedisTokenStore(rdb, redisTestPrefix)
	ctx := context.Background()

	token := "test_token_blacklist_001"
	ttl := 5 * time.Minute

	// 初始状态：不在黑名单
	blacklisted, err := store.IsBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.False(t, blacklisted)

	// 加入黑名单
	err = store.Blacklist(ctx, token, ttl)
	require.NoError(t, err)

	// 应在黑名单中
	blacklisted, err = store.IsBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestRedisTokenStore_NotBlacklistedForUnknownToken(t *testing.T) {
	rdb := setupRedisClient(t)
	defer cleanupRedisKeys(t, rdb)
	store := NewRedisTokenStore(rdb, redisTestPrefix)
	ctx := context.Background()

	blacklisted, err := store.IsBlacklisted(ctx, "totally_unknown_token_xyz")
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestRedisTokenStore_MultipleTokensIndependent(t *testing.T) {
	rdb := setupRedisClient(t)
	defer cleanupRedisKeys(t, rdb)
	store := NewRedisTokenStore(rdb, redisTestPrefix)
	ctx := context.Background()

	tokens := []string{"multi_a", "multi_b", "multi_c"}
	for _, tok := range tokens {
		err := store.Blacklist(ctx, tok, 10*time.Minute)
		require.NoError(t, err)
	}

	// 每个都应在黑名单中
	for _, tok := range tokens {
		blacklisted, err := store.IsBlacklisted(ctx, tok)
		require.NoError(t, err)
		assert.True(t, blacklisted, "token %s should be blacklisted", tok)
	}

	// 未加入的不在黑名单中
	blacklisted, err := store.IsBlacklisted(ctx, "multi_not_added")
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestRedisTokenStore_OverwriteBlacklist(t *testing.T) {
	rdb := setupRedisClient(t)
	defer cleanupRedisKeys(t, rdb)
	store := NewRedisTokenStore(rdb, redisTestPrefix)
	ctx := context.Background()

	token := "test_overwrite"

	// 第一次加入，TTL=1分钟
	err := store.Blacklist(ctx, token, 1*time.Minute)
	require.NoError(t, err)

	// 覆盖写入，TTL=10分钟（TTL 应被更新）
	err = store.Blacklist(ctx, token, 10*time.Minute)
	require.NoError(t, err)

	blacklisted, err := store.IsBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.True(t, blacklisted)
}

// ========== TTL 过期行为 ==========

func TestRedisTokenStore_TTLExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过 TTL 过期测试（需要等待过期）")
	}

	rdb := setupRedisClient(t)
	defer cleanupRedisKeys(t, rdb)
	store := NewRedisTokenStore(rdb, redisTestPrefix)
	ctx := context.Background()

	token := "test_ttl_expiry_001"
	shortTTL := 2 * time.Second // 极短 TTL 用于测试

	err := store.Blacklist(ctx, token, shortTTL)
	require.NoError(t, err)

	// 立即检查：应在黑名单中
	blacklisted, err := store.IsBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// 等待 TTL 过期
	time.Sleep(3 * time.Second)

	// 过期后应不在黑名单中
	blacklisted, err = store.IsBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.False(t, blacklisted, "token should be expired after TTL")
}

// ========== 接口编译检查 ==========

func TestRedisTokenStoreImplementsInterface(t *testing.T) {
	var _ TokenStore = (*RedisTokenStore)(nil)
}
