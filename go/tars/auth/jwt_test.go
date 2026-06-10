package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAuthService_默认值 验证构造函数的默认参数
func TestNewAuthService_默认值(t *testing.T) {
	svc := NewAuthService(nil, "", 0)
	assert.NotNil(t, svc)
	assert.Equal(t, "cairobot", svc.issuer)
	assert.Equal(t, 24*time.Hour, svc.expiration)
}

// TestGenerateToken_正常签发 验证 Token 签发和基本字段
func TestGenerateToken_正常签发(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	token, err := svc.GenerateToken("user-001", "parent")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Token 应以 eyJ 开头（Base64URL 编码的 JWT）
	assert.Contains(t, token, "eyJ")
}

// TestGenerateToken_空UserID 验证空 userID 返回错误
func TestGenerateToken_空UserID(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	_, err := svc.GenerateToken("", "parent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userID 不能为空")
}

// TestValidateToken_有效Token 验证正确签发的 Token 可以通过校验
func TestValidateToken_有效Token(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	token, _ := svc.GenerateToken("user-001", "child")

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-001", claims.UserID)
	assert.Equal(t, "child", claims.Role)
	assert.Equal(t, "cairobot", claims.Issuer)
}

// TestValidateToken_过期Token 验证已过期的 Token 被拒绝
func TestValidateToken_过期Token(t *testing.T) {
	// 创建一个立即过期的 service
	svc := NewAuthService([]byte("test-secret"), "cairobot", 1*time.Millisecond)

	token, _ := svc.GenerateToken("user-001", "parent")

	// 等待 Token 过期
	time.Sleep(10 * time.Millisecond)

	_, err := svc.ValidateToken(token)
	assert.Error(t, err)
}

// TestValidateToken_错误签名 验证被篡改的 Token 被拒绝
func TestValidateToken_错误签名(t *testing.T) {
	svc1 := NewAuthService([]byte("secret-A"), "cairobot", time.Hour)
	svc2 := NewAuthService([]byte("secret-B"), "cairobot", time.Hour)

	// 用 secret-A 签发
	token, _ := svc1.GenerateToken("user-001", "parent")

	// 用 secret-B 校验 → 应该失败
	_, err := svc2.ValidateToken(token)
	assert.Error(t, err)
}

// TestValidateToken_篡改Payload 验证手动修改 Payload 后签名不匹配
func TestValidateToken_篡改Payload(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	token, _ := svc.GenerateToken("user-001", "admin")

	// 手动解析 JWT 并修改 payload（这会导致签名验证失败）
	parsed, _, err := jwt.NewParser().ParseUnverified(token, &TokenClaims{})
	require.NoError(t, err)

	claims, ok := parsed.Claims.(*TokenClaims)
	require.True(t, ok)
	claims.UserID = "hacker" // 篡改用户 ID

	// 尝试用篡改后的 claims 重新编码（没有合法签名）
	forgedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	require.NoError(t, err)

	// 原始 service 校验伪造 token → 失败
	_, err = svc.ValidateToken(forgedToken)
	assert.Error(t, err)
}

// TestValidateToken_空字符串 验证空字符串返回错误
func TestValidateToken_空字符串(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	_, err := svc.ValidateToken("")
	assert.Error(t, err)
}

// TestValidateToken_非法格式 验证非 JWT 格式字符串返回错误
func TestValidateToken_非法格式(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	_, err := svc.ValidateToken("not-a-jwt-token")
	assert.Error(t, err)
}

// TestGenerateToken_默认角色 验证空 role 默认为 parent
func TestGenerateToken_默认角色(t *testing.T) {
	svc := NewAuthService([]byte("test-secret"), "cairobot", time.Hour)

	token, _ := svc.GenerateToken("user-001", "")

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "parent", claims.Role)
}
