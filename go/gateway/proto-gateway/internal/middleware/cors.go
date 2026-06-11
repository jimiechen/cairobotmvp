// Package middleware 提供 Gateway HTTP 层中间件
// 本文件实现 CORS 中间件，解决前端跨域访问 Gateway 的问题
//
// 问题背景：
//   - Proto Tester 前端运行在 Vite dev server (http://127.0.0.1:3002)
//   - Gateway 运行在 TarsGo HTTP server (http://localhost:8080)
//   - 浏览器同源策略会阻止跨域 POST 请求，先发送 OPTIONS 预检
//   - 原 ServeHTTP 只接受 POST，导致 OPTIONS 被 405 拒绝 → net::ERR_FAILED
package middleware

import (
	"net/http"
)

// CORSMiddleware 包装 http.Handler，添加 CORS 头和 OPTIONS 预检处理
type CORSMiddleware struct {
	handler     http.Handler
	allowedOrigins []string // 允许的 Origin 列表，空表示允许所有
}

// NewCORSMiddleware 创建 CORS 中间件
func NewCORSMiddleware(handler http.Handler, allowedOrigins ...string) *CORSMiddleware {
	return &CORSMiddleware{
		handler:         handler,
		allowedOrigins: allowedOrigins,
	}
}

// isOriginAllowed 检查请求 Origin 是否在白名单中
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	if len(m.allowedOrigins) == 0 {
		return true // 空白名单 = 允许所有（dev 模式）
	}
	for _, allowed := range m.allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// ServeHTTP 实现 http.Handler 接口
func (m *CORSMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// 设置 CORS 响应头
	if origin != "" && m.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")
	}

	// 处理 OPTIONS 预检请求（浏览器自动发送）
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 小时缓存预检结果
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 非预检请求，透传给下游 handler
	m.handler.ServeHTTP(w, r)
}
