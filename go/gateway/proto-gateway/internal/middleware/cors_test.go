package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockHandler 用于测试的简单 handler，返回 200 + "ok"
type mockHandler struct{}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestCORSMiddleware_OptionsPreflight(t *testing.T) {
	handler := &mockHandler{}
	cors := NewCORSMiddleware(handler)

	t.Run("OPTIONS 请求返回 204 NoContent 并设置 CORS 头", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/hello", nil)
		req.Header.Set("Origin", "http://127.0.0.1:3002")
		rec := httptest.NewRecorder()

		cors.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}

		// 验证 CORS 响应头
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:3002" {
			t.Errorf("expected Allow-Origin header, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		if rec.Header().Get("Access-Control-Allow-Methods") != "POST, OPTIONS" {
			t.Errorf("expected Allow-Methods POST, OPTIONS, got: %s", rec.Header().Get("Access-Control-Allow-Methods"))
		}
		if rec.Header().Get("Access-Control-Max-Age") != "86400" {
			t.Errorf("expected Max-Age 86400, got: %s", rec.Header().Get("Access-Control-Max-Age"))
		}
	})

	t.Run("OPTIONS 无 Origin 也正常返回 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/hello", nil)
		rec := httptest.NewRecorder()

		cors.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestCORSMiddleware_PostRequest_PassesThrough(t *testing.T) {
	handler := &mockHandler{}
	cors := NewCORSMiddleware(handler)

	t.Run("POST 请求透传给下游 handler 并附加 CORS 头", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/hello", nil)
		req.Header.Set("Origin", "http://127.0.0.1:3002")
		rec := httptest.NewRecorder()

		cors.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 from downstream, got %d", rec.Code)
		}
		if rec.Body.String() != "ok" {
			t.Fatalf("expected body 'ok', got '%s'", rec.Body.String())
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:3002" {
			t.Errorf("CORS header missing on POST response")
		}
	})
}

func TestCORSMiddleware_OriginWhitelist(t *testing.T) {
	handler := &mockHandler{}

	// 仅允许特定 Origin
	cors := NewCORSMiddleware(handler, "http://localhost:3000")

	t.Run("白名单内 Origin 允许访问", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/hello", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		cors.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("expected whitelisted origin in header")
		}
	})

	t.Run("白名单外 Origin 不返回 Allow-Origin（但请求仍透传）", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/hello", nil)
		req.Header.Set("Origin", "http://evil.com")
		rec := httptest.NewRecorder()

		cors.ServeHTTP(rec, req)

		// 下游仍应收到请求（CORS 只是头，不阻止服务端处理）
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 (downstream still processes), got %d", rec.Code)
		}
		// 但不返回 Allow-Origin
		if rec.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
			t.Error("should not allow non-whitelisted origin")
		}
	})
}

func TestCORSMiddleware_EmptyWhitelist_AllowsAll(t *testing.T) {
	handler := &mockHandler{}
	// 空白名单 = 允许所有 Origin（dev 模式默认行为）
	cors := NewCORSMiddleware(handler)

	req := httptest.NewRequest(http.MethodOptions, "/api/hello", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rec := httptest.NewRecorder()

	cors.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://any-origin.com" {
		t.Errorf("empty whitelist should allow any origin, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
