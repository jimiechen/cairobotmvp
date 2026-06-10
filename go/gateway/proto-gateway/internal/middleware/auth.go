// Package middleware 提供 Gateway HTTP 层中间件
// 本文件实现 Auth 中间件，在 ServeHTTP 的路由匹配后、Invoker.Invoke 前执行 Token 校验
package middleware

import (
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/tars/auth"
)

// AuthMiddleware Gateway 层鉴权中间件
// 在路由匹配之后、业务 Handler 执行之前拦截请求，校验 JWT Token
type AuthMiddleware struct {
	authService *auth.AuthService
}

// NewAuthMiddleware 创建 AuthMiddleware 实例
func NewAuthMiddleware(authService *auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// InterceptResult 中间件拦截结果
//
// 当 Intercept 返回非 nil 时表示需要中断请求并直接返回错误响应。
// 返回 nil 表示放行，继续执行下游 Invoker.Invoke。
type InterceptResult struct {
	ResponsePacket *adapter.MessagePacket
}

// Intercept 在 Invoker.Invoke 前执行 Token 校验
//
// 处理逻辑：
//  1. 检查 route.AuthRequired，false 则直接放行
//  2. 从 packet.Extend["token"] 提取 JWT
//  3. Token 为空 → 返回 40101 错误
//  4. JWT 校验失败 → 返回 40102 错误
//  5. 校验通过 → 将 userId/userRole 注入 extend，返回 nil 放行
//
// 参数：
//   - packet: 已反序列化的 MessagePacket（可能被修改以注入 user_id/user_role）
//   - authRequired: 当前路由的 auth_required 配置值
//
// 输出：*InterceptResult 为 nil 表示放行，非 nil 表示需要中断并返回该响应
func (m *AuthMiddleware) Intercept(packet *adapter.MessagePacket, authRequired bool) *InterceptResult {
	// 免鉴权路由直接放行
	if !authRequired {
		return nil
	}

	// 提取 token
	tokenStr := packet.Extend["token"]
	if tokenStr == "" {
		return &InterceptResult{
			ResponsePacket: buildAuthErrorResponse(packet, auth.ErrCodeMissingToken, "missing token"),
		}
	}

	// 校验 JWT
	claims, err := m.authService.ValidateToken(tokenStr)
	if err != nil {
		return &InterceptResult{
			ResponsePacket: buildAuthErrorResponse(packet, auth.ErrCodeInvalidToken, "invalid token: "+err.Error()),
		}
	}

	// 注入用户身份到 extend，供下游 handler 使用
	if packet.Extend == nil {
		packet.Extend = make(map[string]string)
	}
	packet.Extend["user_id"] = claims.UserID
	packet.Extend["user_role"] = claims.Role

	return nil // 校验通过，放行
}

// buildAuthErrorResponse 构造鉴权错误响应 MessagePacket
func buildAuthErrorResponse(req *adapter.MessagePacket, code, message string) *adapter.MessagePacket {
	resp := adapter.BuildErrorPacket(req, atoi32(code), message)
	// 保持原始 maxType/minType 用于客户端路由匹配
	if req != nil {
		resp.MaxType = req.MaxType
		resp.MinType = req.MinType
	}
	return resp
}

// atoi32 将字符串错误码转为 int32（如 "40101" → 40101）
func atoi32(s string) int32 {
	var n int32
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int32(c-'0')
		}
	}
	return n
}
