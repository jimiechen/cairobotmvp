package middleware

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/tars/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 创建测试用 AuthService 辅助函数
func newTestAuthService() *auth.AuthService {
	return auth.NewAuthService([]byte("test-middleware-secret"), "cairobot", 24*0)
}

// newTestPacket 创建测试用 MessagePacket
func newTestPacket(maxType, minType int32, token string) *adapter.MessagePacket {
	ext := make(map[string]string)
	if token != "" {
		ext["token"] = token
	}
	ext["traceId"] = "test-trace-123"
	return &adapter.MessagePacket{
		MaxType: maxType,
		MinType: minType,
		Data:   []byte(`{}`),
		Extend: ext,
	}
}

// TestIntercept_免鉴权路由放行 验证 authRequired=false 时直接放行
func TestIntercept_免鉴权路由放行(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	packet := newTestPacket(6000, 6003, "") // GetAppLanguage 无需鉴权

	result := mw.Intercept(packet, false)
	assert.Nil(t, result, "免鉴权路由应放行")
}

// TestIntercept_缺失Token 验证 authRequired=true 但无 token 时返回 40101
func TestIntercept_缺失Token(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	packet := newTestPacket(6000, 6001, "") // GetAppConfigs 需要鉴权，无 token

	result := mw.Intercept(packet, true)
	require.NotNil(t, result, "缺失 Token 应拦截")

	resp := result.ResponsePacket
	assert.Equal(t, auth.ErrCodeMissingToken, resp.Extend["code"])
	assert.Contains(t, resp.Extend["message"], "missing token")
}

// TestIntercept_无效Token 验证格式错误的 token 返回 40102
func TestIntercept_无效Token(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	packet := newTestPacket(6000, 6001, "this-is-not-a-valid-jwt")

	result := mw.Intercept(packet, true)
	require.NotNil(t, result, "无效 Token 应拦截")

	resp := result.ResponsePacket
	assert.Equal(t, auth.ErrCodeInvalidToken, resp.Extend["code"])
	assert.Contains(t, resp.Extend["message"], "invalid token")
}

// TestIntercept_有效Token放行 验证有效 JWT 通过校验后放行并注入 user_id
func TestIntercept_有效Token放行(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	// 先签发一个合法 token
	token, err := svc.GenerateToken("user-abc", "admin")
	require.NoError(t, err)

	packet := newTestPacket(6000, 6001, token)

	result := mw.Intercept(packet, true)
	assert.Nil(t, result, "有效 Token 应放行")

	// 验证 extend 中注入了用户信息
	assert.Equal(t, "user-abc", packet.Extend["user_id"])
	assert.Equal(t, "admin", packet.Extend["user_role"])
}

// TestIntercept_错误签名Token 验证不同密钥签发的 token 被拒绝
func TestIntercept_错误签名Token(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	// 用另一个密钥签发
	otherSvc := auth.NewAuthService([]byte("different-secret"), "cairobot", 24*0)
	token, _ := otherSvc.GenerateToken("user-abc", "parent")

	packet := newTestPacket(6000, 6001, token)

	result := mw.Intercept(packet, true)
	require.NotNil(t, result, "错误签名 Token 应拦截")
	assert.Equal(t, auth.ErrCodeInvalidToken, result.ResponsePacket.Extend["code"])
}

// TestIntercept_PacketNilExtend 验证 extend 为 nil 时不会 panic
func TestIntercept_PacketNilExtend(t *testing.T) {
	svc := newTestAuthService()
	mw := NewAuthMiddleware(svc)

	packet := &adapter.MessagePacket{
		MaxType: 6000,
		MinType: 6001,
		Data:    []byte(`{}`),
		Extend:  nil, // nil extend
	}

	result := mw.Intercept(packet, true)
	require.NotNil(t, result, "nil extend + required auth 应拦截")
	assert.Equal(t, auth.ErrCodeMissingToken, result.ResponsePacket.Extend["code"])
}
