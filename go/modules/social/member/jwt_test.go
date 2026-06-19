package member

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== JWTManager 测试 ==========

func newTestJWTConfig() *JWTConfig {
	return DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-unit-tests-32b!")
}

func TestNewJWTManager_成功(t *testing.T) {
	cfg := newTestJWTConfig()
	mgr, err := NewJWTManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, mgr)
}

func TestNewJWTManager_密钥为空(t *testing.T) {
	cfg := DefaultJWTConfig()
	_, err := NewJWTManager(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SecretKey 不能为空")
}

func TestNewJWTManager_密钥过短(t *testing.T) {
	cfg := DefaultJWTConfig().SetSecretKey("short")
	_, err := NewJWTManager(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "长度不足")
}

func TestJWTManager_GenerateAccessToken_生成有效令牌(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	userID := "user-001"

	token, exp, err := mgr.GenerateAccessToken(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, exp, time.Now().UnixMilli())
	// JWT 应包含 3 段（header.payload.signature）
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestJWTManager_GenerateRefreshToken_生成有效令牌(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	token, _, err := mgr.GenerateRefreshToken("user-001")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestJWTManager_AccessToken与RefreshToken不同(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	accessTok, _, _ := mgr.GenerateAccessToken("user-001")
	refreshTok, _, _ := mgr.GenerateRefreshToken("user-001")

	// 同一用户的 access 和 refresh 令牌必须不同
	assert.NotEqual(t, accessTok, refreshTok)
}

func TestJWTManager_ParseToken_有效令牌(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	originalToken, _, _ := mgr.GenerateAccessToken("user-123")

	userID, tokenType, err := mgr.ParseToken(originalToken)

	assert.NoError(t, err)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "access", tokenType)
}

func TestJWTManager_ParseToken_Refresh类型(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	originalToken, _, _ := mgr.GenerateRefreshToken("user-456")

	userID, tokenType, err := mgr.ParseToken(originalToken)

	assert.NoError(t, err)
	assert.Equal(t, "user-456", userID)
	assert.Equal(t, "refresh", tokenType)
}

func TestJWTManager_ParseToken_无效签名(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	validToken, _, _ := mgr.GenerateAccessToken("user-001")

	// 篡改 payload 部分
	parts := strings.Split(validToken, ".")
	tampered := parts[0] + "." + "tampered_payload" + "." + parts[2]

	_, _, err := mgr.ParseToken(tampered)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestJWTManager_ParseToken_过期令牌(t *testing.T) {
	// 创建一个 TTL 为负数的配置来模拟过期
	cfg := newTestJWTConfig()
	cfg.AccessTTL = -1 * time.Second // 已过期
	mgr, _ := NewJWTManager(cfg)

	expiredToken, _, _ := mgr.GenerateAccessToken("user-001")

	_, _, err := mgr.ParseToken(expiredToken)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestJWTManager_ParseToken_错误格式(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	_, _, err := mgr.ParseToken("not-a-valid-jwt-token")
	assert.Error(t, err)
}

// ========== MemoryTokenStore 测试 ==========

func TestMemoryTokenStore_BlacklistAndCheck(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	err := store.Blacklist(ctx, "token-abc", time.Hour)
	assert.NoError(t, err)

	blacklisted, err := store.IsBlacklisted(ctx, "token-abc")
	assert.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestMemoryTokenStore_未黑名单返回false(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	blacklisted, err := store.IsBlacklisted(ctx, "nonexistent-token")
	assert.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestMemoryTokenStore_过期自动清理(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	// 加入黑名单，TTL 极短
	store.Blacklist(ctx, "short-lived", 1*time.Millisecond)

	// 等待过期
	time.Sleep(10 * time.Millisecond)

	blacklisted, err := store.IsBlacklisted(ctx, "short-lived")
	assert.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestMemoryTokenStore_多个令牌独立管理(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	store.Blacklist(ctx, "token-a", time.Hour)
	store.Blacklist(ctx, "token-b", 2*time.Hour)

	a, _ := store.IsBlacklisted(ctx, "token-a")
	b, _ := store.IsBlacklisted(ctx, "token-b")
	c, _ := store.IsBlacklisted(ctx, "token-c")

	assert.True(t, a)
	assert.True(t, b)
	assert.False(t, c)
}

func TestMemoryTokenStore_Blacklist幂等覆盖(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	store.Blacklist(ctx, "same-token", time.Hour)
	store.Blacklist(ctx, "same-token", 30*time.Minute) // 覆盖：更短 TTL

	blacklisted, _ := store.IsBlacklisted(ctx, "same-token")
	assert.True(t, blacklisted)
}
