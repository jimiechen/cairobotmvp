package member

import (
	"context"
	"encoding/json"
	"fmt"
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

// ========== ParseJTI 测试 ==========

func TestJWTManager_ParseJTI_从AccessToken提取(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	token, _, _ := mgr.GenerateAccessToken("user-jti-001")

	jti := mgr.ParseJTI(token)

	assert.NotEmpty(t, jti)
	assert.Contains(t, jti, "acc-") // access token 的 jti 前缀
}

func TestJWTManager_ParseJTI_从RefreshToken提取(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	token, _, _ := mgr.GenerateRefreshToken("user-jti-002")

	jti := mgr.ParseJTI(token)

	assert.NotEmpty(t, jti)
	assert.Contains(t, jti, "ref-") // refresh token 的 jti 前缀
}

func TestJWTManager_ParseJTI_不同令牌不同JTI(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	tok1, _, _ := mgr.GenerateAccessToken("user-001")
	tok2, _, _ := mgr.GenerateAccessToken("user-001") // 同一用户，不同时间签发

	jti1 := mgr.ParseJTI(tok1)
	jti2 := mgr.ParseJTI(tok2)

	assert.NotEmpty(t, jti1)
	assert.NotEmpty(t, jti2)
	assert.NotEqual(t, jti1, jti2) // nanos 精度保证唯一性
}

func TestJWTManager_ParseJTI_无效令牌返回空(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	jti := mgr.ParseJTI("not-a-valid-jwt")
	assert.Empty(t, jti)

	jti = mgr.ParseJTI("")
	assert.Empty(t, jti)
}

func TestJWTManager_ParseJTI_篡改Payload仍可提取JTI(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	validToken, _, _ := mgr.GenerateAccessToken("user-003")

	// 篾改 payload（签名无效但 payload 结构完整）
	parts := strings.Split(validToken, ".")
	if len(parts) == 3 {
		tampered := parts[0] + "." + "tampered_payload" + "." + parts[2]
		jti := mgr.ParseJTI(tampered)
		// 篡改后 base64 解码失败，返回空
		assert.Empty(t, jti)
	}
}

// ========== TokenStore 新接口 (jti-based) 测试 ==========

func TestMemoryTokenStore_StoreExistsDelete_完整生命周期(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	// 初始状态
	exists, _ := store.Exists(ctx, "jti-test-001")
	assert.False(t, exists)

	// Store
	err := store.Store(ctx, "jti-test-001", 3600)
	assert.NoError(t, err)

	// Exists → true
	exists, err = store.Exists(ctx, "jti-test-001")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = store.Delete(ctx, "jti-test-001")
	assert.NoError(t, err)

	// Exists → false（已删除）
	exists, err = store.Exists(ctx, "jti-test-001")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryTokenStore_StoreTTL过期自动清理(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	store.Store(ctx, "jti-short", 0) // TTL=0 表示立即过期（边界情况）

	// 立即检查：TTL=0 的条目应被视为已过期（time.Now().Add(0) = now，After(now)=true）
	exists, _ := store.Exists(ctx, "jti-short")
	assert.False(t, exists, "TTL=0 的条目应视为已过期")

	// 正常 TTL 条目不受影响
	store.Store(ctx, "jti-normal", 3600)
	exists2, _ := store.Exists(ctx, "jti-normal")
	assert.True(t, exists2)
}

func TestMemoryTokenStore_Delete不存在不报错(t *testing.T) {
	store := NewMemoryTokenStore()
	ctx := context.Background()

	err := store.Delete(ctx, "jti-nonexistent")
	assert.NoError(t, err) // 删除不存在的 key 不应报错
}

// ========== JWT 边界测试 ==========

func TestJWTManager_并发签发唯一性(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())

	// 并发签发 100 个 access_token
	type result struct {
		token string
		err   error
	}
	ch := make(chan result, 100)
	for i := 0; i < 100; i++ {
		go func(id int) {
			tok, _, e := mgr.GenerateAccessToken(fmt.Sprintf("concurrent-user-%d", id))
			ch <- result{token: tok, err: e}
		}(i)
	}

	tokens := make(map[string]int)
	for i := 0; i < 100; i++ {
		r := <-ch
		require.NoError(t, r.err)
		tokens[r.token]++
	}

	// 所有 100 个 token 必须唯一
	assert.Len(t, tokens, 100, "并发签发的 token 必须全部唯一")
}

// ========== CheckTokenBlacklist 三域公共黑名单检查测试 ==========

func TestCheckTokenBlacklist_正常token_应放行(t *testing.T) {
	mgr, err := NewJWTManager(newTestJWTConfig())
	require.NoError(t, err)
	store := NewMemoryTokenStore()
	ctx := context.Background()

	// 生成一个有效 token
	tokenStr, _, err := mgr.GenerateAccessToken("user-001")
	require.NoError(t, err)

	extend := map[string]string{"token": tokenStr}
	code, _, err := CheckTokenBlacklist(ctx, extend, mgr, store)

	assert.Equal(t, 0, code, "正常 token 应返回 code=0（放行）")
	assert.NoError(t, err)
}

func TestCheckTokenBlacklist_已撤销token_应拒绝(t *testing.T) {
	mgr, err := NewJWTManager(newTestJWTConfig())
	require.NoError(t, err)
	store := NewMemoryTokenStore()
	ctx := context.Background()

	// 生成 token 并加入黑名单
	tokenStr, _, err := mgr.GenerateAccessToken("user-001")
	require.NoError(t, err)
	jti := mgr.ParseJTI(tokenStr)
	require.NotEmpty(t, jti)
	_ = store.Store(ctx, jti, 3600)

	extend := map[string]string{"token": tokenStr}
	code, respBytes, err := CheckTokenBlacklist(ctx, extend, mgr, store)

	assert.Equal(t, 200, code, "已撤销 token 应返回 200+error body")
	assert.NoError(t, err)
	assert.NotEmpty(t, respBytes, "响应体不应为空")

	// 验证响应体包含 10401 错误码
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(respBytes, &resp))
	result := resp["result"].(map[string]interface{})
	// JSON number 默认为 float64
	assert.Equal(t, float64(10401), result["code"])
}

func TestCheckTokenBlacklist_nil依赖_应降级放行(t *testing.T) {
	ctx := context.Background()
	extend := map[string]string{"token": "some-token"}

	// jwtMgr 为 nil
	code, _, err := CheckTokenBlacklist(ctx, extend, nil, NewMemoryTokenStore())
	assert.Equal(t, 0, code, "nil jwtMgr 应降级放行")

	// tokenStore 为 nil
	mgr2, _ := NewJWTManager(newTestJWTConfig())
	code, _, err = CheckTokenBlacklist(ctx, extend, mgr2, nil)
	assert.Equal(t, 0, code, "nil tokenStore 应降级放行")

	// 两者都为 nil
	code, _, err = CheckTokenBlacklist(ctx, extend, nil, nil)
	assert.Equal(t, 0, code, "两者都 nil 应降级放行")
	assert.NoError(t, err)
}

func TestCheckTokenBlacklist_空token_应放行(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	store := NewMemoryTokenStore()

	code, _, err := CheckTokenBlacklist(context.Background(), map[string]string{}, mgr, store)
	assert.Equal(t, 0, code, "空 extend 应放行")

	code, _, err = CheckTokenBlacklist(context.Background(), map[string]string{"token": ""}, mgr, store)
	assert.Equal(t, 0, code, "空 token 值应放行")
	assert.NoError(t, err)
}

func TestCheckTokenBlacklist_无效token格式_应放行(t *testing.T) {
	mgr, _ := NewJWTManager(newTestJWTConfig())
	store := NewMemoryTokenStore()

	extend := map[string]string{"token": "not-a-valid-jwt"}
	code, _, err := CheckTokenBlacklist(context.Background(), extend, mgr, store)

	assert.Equal(t, 0, code, "无效 token 格式应放行（ParseJTI 返回空）")
	assert.NoError(t, err)
}
