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
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	configsvc "github.com/jimiechen/mineplanet/go/services/config/service"
	i18nsvc "github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// TestE2E_MySQL_RealServices 验证 S1 阶段 MySQL 真实服务全链路：
//
// 前置条件：
//   - go_biz 数据库已初始化（002 + 003 迁移脚本已执行）
//   - MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PASSWORD 环境变量可用
//
// 链路：HTTP POST → Gateway → MessagePacket → LocalInvoker → RealServices(MySQL) → JSON 响应
//
// 跳过条件：环境变量 MYSQL_HOST 未设置时跳过（CI 无 MySQL）
func TestE2E_MySQL_RealServices(t *testing.T) {
	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost == "" {
		t.Skip("MYSQL_HOST 未设置，跳过 MySQL 真实服务 E2E 测试")
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

	routesCfg := loadConfigI18nRoutes()
	rt := router.NewRouteTable(routesCfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)
	tarsclient.RegisterConfigI18nHandlers(invoker, svc.ConfigSvc, svc.I18nSvc)

	gs := NewGatewayServer(rt, invoker, "mysql", nil)

	// ---- Case 1: GetAppConfigs (6000:6001) 返回 go_biz 种子数据 ----
	t.Run("GetAppConfigs_返回真实配置数据", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6001,
			Data:   []byte(`{"env":"dev","client_scope":"all"}`),
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

		var configResp configsvc.AppConfigResponse
		if err := json.Unmarshal(respPacket.Data, &configResp); err != nil {
			t.Fatalf("解析 AppConfigResponse 失败: %v", err)
		}

		if len(configResp.DynamicModules) == 0 {
			t.Fatal("期望非空 DynamicModules（go_biz 中有 base_config 种子数据）")
		}

		foundBaseConfig := false
		for _, mod := range configResp.DynamicModules {
			if mod.ModuleKey == "base_config" {
				foundBaseConfig = true
				if mod.Version == 0 {
					t.Errorf("base_config Version 应 > 0, got %d", mod.Version)
				}
				t.Logf("✅ base_config: version=%d fields=%d descriptors=%d",
					mod.Version, len(mod.Fields), len(mod.Descriptors))
			}
		}
		if !foundBaseConfig {
			t.Error("DynamicModules 中未找到 base_config 模块")
		}
	})

	// ---- Case 2: GetAppLanguage (6000:6003) 返回语言列表 ----
	t.Run("GetAppLanguage_返回真实语言列表", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6003,
			Data:   []byte(`{"env":"dev","client_version":"1.0.0"}`),
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

		var langResp []i18nsvc.LanguageMeta
		if err := json.Unmarshal(respPacket.Data, &langResp); err != nil {
			t.Fatalf("解析 LanguageMeta 列表失败: %v", err)
		}

		if len(langResp) == 0 {
			t.Fatal("期望非空 Languages 列表（go_biz 有 zh-CN + en-US）")
		}
		t.Logf("✅ GetAppLanguage 返回 %d 种语言", len(langResp))
		for _, lang := range langResp {
			t.Logf("   - code=%s name=%s", lang.Code, lang.Name)
		}
	})

	// ---- Case 3: GetLangPack (6000:6005) 返回真实字符串 ----
	t.Run("GetLangPack_返回真实字符串数据", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6005,
			Data:   []byte(`{"lang_code":"zh-CN","client_version":"1.0.0","env":"dev"}`),
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

		var packResp i18nsvc.LangPackResponse
		if err := json.Unmarshal(respPacket.Data, &packResp); err != nil {
			t.Fatalf("解析 LangPackResponse 失败: %v", err)
		}

		if packResp.PackVersion == 0 {
			t.Fatal("PackVersion 应 > 0（go_biz 中 webp pack version=1）")
		}
		if len(packResp.Strings) == 0 {
			t.Fatal("期望非空 Strings（go-biz 有 6 条 zh-CN 字符串）")
		}
		t.Logf("✅ GetLangPack(zh-CN): version=%d strings=%d",
			packResp.PackVersion, len(packResp.Strings))

		// 验证包含已知种子 key
		foundOk := false
		for _, s := range packResp.Strings {
			if s.Key == "common.ok" {
				foundOk = true
				if s.Value != "确定" {
					t.Errorf("common.ok 期望='确定', got='%s'", s.Value)
				}
			}
		}
		if !foundOk {
			t.Error("Strings 中未找到 common.key=common.ok")
		}
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if err := json.Unmarshal([]byte(v), &n); err == nil {
		return n
	}
	return fallback
}
