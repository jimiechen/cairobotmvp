package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/middleware"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	"github.com/jimiechen/mineplanet/go/tars/auth"
)

// TestE2E_Auth_Middleware 验证 S1 阶段 JWT Auth 中间件全链路：
//
// 前置条件：
//   - go_biz 数据库已初始化（002 + 003 迁移脚本已执行）
//   - MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PASSWORD 环境变量可用
//
// 链路：HTTP POST → Gateway → MessagePacket → AuthMiddleware.Intercept → Invoker.Invoke → JSON 响应
//
// 跳过条件：环境变量 MYSQL_HOST 未设置时跳过（CI 无 MySQL）
func TestE2E_Auth_Middleware(t *testing.T) {
	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost == "" {
		t.Skip("MYSQL_HOST 未设置，跳过 Auth E2E 测试")
	}

	mysqlCfg := &config.MySQLConfig{
		Host:            mysqlHost,
		Port:            getEnvInt("MYSQL_PORT", 3306),
		Username:        getEnv("MYSQL_USER", "root"),
		Password:        getEnv("MYSQL_PASSWORD", ""),
		Database:        getEnv("MYSQL_DATABASE", "go_biz"),
		Charset:         "utf8mb4",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: "10m",
		ConnMaxIdleTime: "5m",
	}

	svc, err := tarsclient.BuildRealServices(mysqlCfg, "dev")
	if err != nil {
		t.Fatalf("BuildRealServices 失败: %v", err)
	}

	// 创建带 Auth 中间件的 GatewayServer
	jwtSecret := []byte("cairobot-e2e-test-secret")
	authSvc := auth.NewAuthService(jwtSecret, "cairobot", 24*0)
	authMw := middleware.NewAuthMiddleware(authSvc)

	routesCfg := loadConfigI18nRoutes()
	rt := router.NewRouteTable(routesCfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)
	tarsclient.RegisterConfigI18nHandlers(invoker, svc.ConfigSvc, svc.I18nSvc)

	gs := NewGatewayServer(rt, invoker, "mysql", authMw)

	// ---- Case 1: GetAppConfigs (6001, auth_required=true) 缺失 Token → 40101 ----
	t.Run("GetAppConfigs_缺失Token返回40101", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6001,
			Data:   []byte(`{"env":"dev","client_scope":"all"}`),
			Extend: map[string]string{"traceId": "auth-e2e-001"},
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("HTTP %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if respPacket.Extend["code"] != auth.ErrCodeMissingToken {
			t.Fatalf("期望 code=%s, got '%s'", auth.ErrCodeMissingToken, respPacket.Extend["code"])
		}
		t.Logf("✅ 缺失 Token 正确返回 code=%s msg=%s", respPacket.Extend["code"], respPacket.Extend["message"])
	})

	// ---- Case 2: GetAppConfigs (6001) 错误 Token → 40102 ----
	t.Run("GetAppConfigs_错误Token返回40102", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6001,
			Data:   []byte(`{"env":"dev","client_scope":"all"}`),
			Extend: map[string]string{
				"token":   "this.is.not.a.valid.jwt.token",
				"traceId": "auth-e2e-002",
			},
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("HTTP %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if respPacket.Extend["code"] != auth.ErrCodeInvalidToken {
			t.Fatalf("期望 code=%s, got '%s'", auth.ErrCodeInvalidToken, respPacket.Extend["code"])
		}
		t.Logf("✅ 错误 Token 正确返回 code=%s msg=%s", respPacket.Extend["code"], respPacket.Extend["message"])
	})

	// ---- Case 3: GetAppLanguage (6003, auth_required=false) 无 Token 正常访问 ----
	t.Run("GetAppLanguage_免鉴权无Token正常", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6003,
			Data:   []byte(`{"env":"dev","client_version":"1.0.0"}`),
			Extend: map[string]string{"traceId": "auth-e2e-003"},
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("HTTP %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("期望 code=10200, got '%s'", respPacket.Extend["code"])
		}
		t.Logf("✅ 免鉴权路由无 Token 正常返回 code=10200")
	})

	// ---- Case 4: GetAppConfigs (6001) 有效 Token → 通过鉴权并返回真实数据 ----
	t.Run("GetAppConfigs_有效Token返回真实数据", func(t *testing.T) {
		// 用同一个 secret 签发 token
		token, err := authSvc.GenerateToken("test-user-001", "parent")
		if err != nil {
			t.Fatalf("GenerateToken 失败: %v", err)
		}

		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6001,
			Data:   []byte(`{"env":"dev","client_scope":"all"}`),
			Extend: map[string]string{
				"token":   token,
				"traceId": "auth-e2e-004",
			},
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("HTTP %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("期望 code=10200, got '%s'", respPacket.Extend["code"])
		}

		var configResp map[string]interface{}
		if err := json.Unmarshal(respPacket.Data, &configResp); err != nil {
			t.Fatalf("解析响应数据失败: %v", err)
		}

		t.Logf("✅ 有效 Token 通过鉴权，返回真实配置数据")
	})
}
