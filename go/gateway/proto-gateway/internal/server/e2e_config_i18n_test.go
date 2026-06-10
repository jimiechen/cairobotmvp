package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	configcache "github.com/jimiechen/mineplanet/go/services/config/cache"
	configdomain "github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	configservice "github.com/jimiechen/mineplanet/go/services/config/service"
	i18ncache "github.com/jimiechen/mineplanet/go/services/i18n/cache"
	i18nrepo "github.com/jimiechen/mineplanet/go/services/i18n/repository"
	i18nservice "github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// TestE2E_GetAppConfigs_FullChain 验证 Config 模块全链路：
// HTTP POST → Gateway Server → MessagePacket 解码 → LocalInvoker → Config Service → SQLite DB → JSON 响应
func TestE2E_GetAppConfigs_FullChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo, _ := repository.NewSQLiteConfigRepoFromDB(db)
	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	mockCache := configcache.NewMockCache()
	configSvc := configservice.NewAppConfigService(configRepo, schemaRepo, mockCache)

	i18nRepo := i18nrepo.NewSQLiteRepo(db)
	mockI18nCache := i18ncache.NewMockCache()
	i18nSvc := i18nservice.NewI18nService(i18nRepo, mockI18nCache, "dev")

	seedTestData(t, db, configRepo)

	routesCfg := loadConfigI18nRoutes()
	rt := router.NewRouteTable(routesCfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)
	tarsclient.RegisterConfigI18nHandlers(invoker, configSvc, i18nSvc)

	gs := NewGatewayServer(rt, invoker, "local", nil)

	t.Run("GetAppConfigs 返回动态模块配置", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6001,
			Data:   []byte(`{"env":"dev","client_scope":"all"}`),
		}

		body, err := adapter.SerializeMessagePacket(reqPacket)
		if err != nil {
			t.Fatalf("serialize packet failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("deserialize response failed: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("expected code '10200', got '%s'", respPacket.Extend["code"])
		}
		if respPacket.MaxType != 6000 || respPacket.MinType != 6002 {
			t.Fatalf("expected MaxType:6000 MinType:6002, got %d:%d", respPacket.MaxType, respPacket.MinType)
		}

		var configResp configservice.AppConfigResponse
		if err := json.Unmarshal(respPacket.Data, &configResp); err != nil {
			t.Fatalf("unmarshal config response failed: %v", err)
		}

		if len(configResp.DynamicModules) == 0 {
			t.Fatal("expected non-empty dynamic_modules")
		}

		t.Logf("✅ GetAppConfigs 成功返回 %d 个动态模块", len(configResp.DynamicModules))
		for _, mod := range configResp.DynamicModules {
			t.Logf("   - ModuleKey=%s Version=%d Fields=%d", mod.ModuleKey, mod.Version, len(mod.Fields))
		}
	})
}

// TestE2E_GetLangPack_FullChain 验证 I18n 模块全链路：
// HTTP POST → Gateway Server → MessagePacket 解码 → LocalInvoker → I18n Service → SQLite DB → JSON 响应
func TestE2E_GetLangPack_FullChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo, _ := repository.NewSQLiteConfigRepo(":memory:")
	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	mockCache := configcache.NewMockCache()
	configSvc := configservice.NewAppConfigService(configRepo, schemaRepo, mockCache)

	i18nRepo := i18nrepo.NewSQLiteRepo(db)
	mockI18nCache := i18ncache.NewMockCache()
	i18nSvc := i18nservice.NewI18nService(i18nRepo, mockI18nCache, "dev")

	seedI18nData(t, db)

	routesCfg := loadConfigI18nRoutes()
	rt := router.NewRouteTable(routesCfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)
	tarsclient.RegisterConfigI18nHandlers(invoker, configSvc, i18nSvc)

	gs := NewGatewayServer(rt, invoker, "local", nil)

	t.Run("GetLangPack 返回语言包数据", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 6000,
			MinType: 6005,
			Data:   []byte(`{"lang_code":"zh-CN","client_version":"1.0.0","env":"dev"}`),
		}

		body, err := adapter.SerializeMessagePacket(reqPacket)
		if err != nil {
			t.Fatalf("serialize packet failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("deserialize response failed: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("expected code '10200', got '%s', data: %s", respPacket.Extend["code"], string(respPacket.Data))
		}
		if respPacket.MaxType != 6000 || respPacket.MinType != 6006 {
			t.Fatalf("expected MaxType:6000 MinType:6006, got %d:%d", respPacket.MaxType, respPacket.MinType)
		}

		var packResp i18nservice.LangPackResponse
		if err := json.Unmarshal(respPacket.Data, &packResp); err != nil {
			t.Fatalf("unmarshal lang pack response failed: %v", err)
		}

		t.Logf("✅ GetLangPack 成功返回 PackVersion=%d Strings=%d", packResp.PackVersion, len(packResp.Strings))
	})
}

// TestE2E_NewFieldWithoutCodeChange 验证新增字段不改代码
// 核心能力：dynamic_modules 自描述机制允许服务端新增 module_key 而无需修改 proto 和客户端代码
func TestE2E_NewFieldWithoutCodeChange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	configRepo, _ := repository.NewSQLiteConfigRepoFromDB(db)
	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	mockCache := configcache.NewMockCache()
	configSvc := configservice.NewAppConfigService(configRepo, schemaRepo, mockCache)

	i18nRepo := i18nrepo.NewSQLiteRepo(db)
	mockI18nCache := i18ncache.NewMockCache()
	i18nSvc := i18nservice.NewI18nService(i18nRepo, mockI18nCache, "dev")

	seedTestData(t, db, configRepo)

	routesCfg := loadConfigI18nRoutes()
	rt := router.NewRouteTable(routesCfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)
	tarsclient.RegisterConfigI18nHandlers(invoker, configSvc, i18nSvc)

	gs := NewGatewayServer(rt, invoker, "local", nil)

	newModuleKey := fmt.Sprintf("test_dynamic_module_%d", time.Now().UnixNano())

	t.Run("首次调用获取初始配置", func(t *testing.T) {
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

		respPacket, _ := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		var configResp configservice.AppConfigResponse
		json.Unmarshal(respPacket.Data, &configResp)

		initialModuleCount := len(configResp.DynamicModules)
		t.Logf("📊 初始模块数量: %d", initialModuleCount)

		if initialModuleCount == 0 {
			t.Fatal("expected at least 1 initial module")
		}
	})

	t.Run("新增模块后再次调用验证动态扩展", func(t *testing.T) {
		insertNewModule(t, db, configRepo, newModuleKey)

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

		respPacket, _ := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		var configResp configservice.AppConfigResponse
		json.Unmarshal(respPacket.Data, &configResp)

		found := false
		for _, mod := range configResp.DynamicModules {
			if mod.ModuleKey == newModuleKey {
				found = true
				t.Logf("✅ 发现新模块: ModuleKey=%s Version=%d", mod.ModuleKey, mod.Version)
				break
			}
		}

		if !found {
			t.Fatalf("expected dynamic_modules to contain new module '%s', got %d modules", newModuleKey, len(configResp.DynamicModules))
		}
	})
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	initAllTables(t, db)
	return db
}

func initAllTables(t *testing.T, db *sql.DB) {
	t.Helper()

	configSQL := `
	CREATE TABLE IF NOT EXISTS sys_config_version (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		module_key    TEXT NOT NULL,
		env           TEXT NOT NULL DEFAULT 'dev',
		version       INTEGER NOT NULL DEFAULT 1,
		config_json   TEXT NOT NULL,
		is_published  INTEGER NOT NULL DEFAULT 0,
		published_at  TEXT,
		created_at    TEXT DEFAULT (datetime('now')),
		updated_at    TEXT DEFAULT (datetime('now')),
		create_by     TEXT,
		update_by     TEXT
	);
	CREATE TABLE IF NOT EXISTS sys_config_schema (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		module_key    TEXT NOT NULL,
		field_key     TEXT NOT NULL,
		field_type    TEXT NOT NULL DEFAULT 'string',
		default_value TEXT,
		validator     TEXT,
		is_required   INTEGER DEFAULT 0,
		is_secret     INTEGER DEFAULT 0,
		description   TEXT,
		client_scope  TEXT DEFAULT 'all',
		min_app_ver   TEXT,
		sort_order    INTEGER DEFAULT 0,
		is_enabled    INTEGER DEFAULT 1,
		UNIQUE (module_key, field_key)
	);
	`

	i18nSQL := `
	CREATE TABLE IF NOT EXISTS sys_lang_pack (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		pack_name     TEXT NOT NULL,
		env           TEXT NOT NULL DEFAULT 'dev',
		version       INTEGER NOT NULL DEFAULT 1,
		lang_code     TEXT NOT NULL,
		description   TEXT,
		is_published  INTEGER NOT NULL DEFAULT 0,
		published_at  TEXT,
		published_by  INTEGER,
		created_at    TEXT DEFAULT (datetime('now')),
		updated_at    TEXT DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS sys_lang_string (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		pack_id         INTEGER NOT NULL,
		string_key      TEXT NOT NULL,
		string_value    TEXT NOT NULL,
		group_name      TEXT,
		version         INTEGER NOT NULL DEFAULT 1,
		operation_type  TEXT NOT NULL DEFAULT 'ADD',
		prev_value      TEXT,
		template_type   TEXT NOT NULL DEFAULT 'plain',
		params_schema   TEXT,
		preview_sample  TEXT,
		created_at      TEXT DEFAULT (datetime('now')),
		updated_at      TEXT DEFAULT (datetime('now'))
	);
	`

	for _, sql := range []string{configSQL, i18nSQL} {
		if _, err := db.Exec(sql); err != nil {
			t.Fatalf("init table failed: %v", err)
		}
	}
}

func seedTestData(t *testing.T, db *sql.DB, configRepo *repository.SQLiteConfigRepo) {
	t.Helper()

	schemaRepo := repository.NewSQLiteSchemaRepo(db)

	moduleKey := "base_config"

	err := schemaRepo.Create(&configdomain.FieldSchema{
		ModuleKey:  moduleKey,
		FieldKey:   "app_name",
		FieldType:  configdomain.FieldTypeString,
		IsRequired: false,
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("create schema failed: %v", err)
	}

	now := time.Now()
	configRepo.Save(&configdomain.ConfigVersion{
		ModuleKey:   moduleKey,
		Env:         "dev",
		Version:     1,
		ConfigJSON:  `{"app_name":"CaiRobot MVP","version":"1.0.0"}`,
		IsPublished: true,
		PublishedAt: &now,
	})

	t.Log("✅ 种子数据已插入: base_config module")
}

func seedI18nData(t *testing.T, db *sql.DB) {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO sys_lang_pack (pack_name, env, version, lang_code, is_published)
		VALUES ('zh-CN-pack', 'dev', 1, 'zh-CN', 1)
	`)
	if err != nil {
		t.Fatalf("insert lang pack failed: %v", err)
	}
	packID, _ := result.LastInsertId()

	_, err = db.Exec(`
		INSERT INTO sys_lang_string (pack_id, string_key, string_value, version, operation_type, template_type)
		VALUES (?, 'hello.world', '你好，世界！', 1, 'ADD', 'plain')
	`, packID)
	if err != nil {
		t.Fatalf("insert lang string failed: %v", err)
	}

	t.Log("✅ i18n 种子数据已插入: zh-CN language pack")
}

func insertNewModule(t *testing.T, db *sql.DB, configRepo *repository.SQLiteConfigRepo, moduleKey string) {
	t.Helper()

	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	schemaRepo.Create(&configdomain.FieldSchema{
		ModuleKey:  moduleKey,
		FieldKey:   "test_field",
		FieldType:  configdomain.FieldTypeString,
		IsRequired: false,
		IsEnabled:  true,
	})

	now := time.Now()
	configRepo.Save(&configdomain.ConfigVersion{
		ModuleKey:   moduleKey,
		Env:         "dev",
		Version:     1,
		ConfigJSON:  `{"test_field":"dynamic_value"}`,
		IsPublished: true,
		PublishedAt: &now,
	})

	t.Logf("✅ 新模块已插入: %s", moduleKey)
}

func loadConfigI18nRoutes() *config.RoutesConfig {
	return &config.RoutesConfig{
		Routes: []config.Route{
			{
				RequestMax: 6000, RequestMin: 6001, RouteKey: "6000:6001",
				CommandName: "GetAppConfigs",
				ResponseMax: 6000, ResponseMin: 6002,
				TarsApp: "CaiRobot", TarsServer: "ConfigServer", TarsServant: "ConfigObj",
				TarsModule: "CaiRobotConfigApp", TarsInterface: "ConfigObj", TarsMethod: "GetAppConfigs",
				RequestProto: "com.mineplanet.pojo.AppConfigsReq", ResponseProto: "com.mineplanet.pojo.AppConfigsRsp",
				TarsRequestType: "vector<byte>", TarsResponseType: "vector<byte>", TimeoutMs: 5000,
				AuthRequired: true, AuditRequired: false,
			},
			{
				RequestMax: 6000, RequestMin: 6009, RouteKey: "6000:6009",
				CommandName: "AppConfigVersion",
				ResponseMax: 6000, ResponseMin: 6010,
				TarsApp: "CaiRobot", TarsServer: "ConfigServer", TarsServant: "ConfigObj",
				TarsModule: "CaiRobotConfigApp", TarsInterface: "ConfigObj", TarsMethod: "AppConfigVersion",
				RequestProto: "com.mineplanet.pojo.AppConfigVersionReq", ResponseProto: "com.mineplanet.pojo.AppConfigVersionRsp",
				TarsRequestType: "vector<byte>", TarsResponseType: "vector<byte>", TimeoutMs: 3000,
				AuthRequired: true, AuditRequired: false,
			},
			{
				RequestMax: 6000, RequestMin: 6003, RouteKey: "6000:6003",
				CommandName: "GetAppLanguage",
				ResponseMax: 6000, ResponseMin: 6004,
				TarsApp: "CaiRobot", TarsServer: "I18nServer", TarsServant: "I18nObj",
				TarsModule: "CaiRobotI18nApp", TarsInterface: "I18nObj", TarsMethod: "GetAppLanguage",
				RequestProto: "com.mineplanet.pojo.AppFetchLanguageReq", ResponseProto: "com.mineplanet.pojo.AppFetchLanguageRsp",
				TarsRequestType: "vector<byte>", TarsResponseType: "vector<byte>", TimeoutMs: 3000,
				AuthRequired: false, AuditRequired: false,
			},
			{
				RequestMax: 6000, RequestMin: 6005, RouteKey: "6000:6005",
				CommandName: "GetLangPack",
				ResponseMax: 6000, ResponseMin: 6006,
				TarsApp: "CaiRobot", TarsServer: "I18nServer", TarsServant: "I18nObj",
				TarsModule: "CaiRobotI18nApp", TarsInterface: "I18nObj", TarsMethod: "GetLangPack",
				RequestProto: "com.mineplanet.pojo.AppFetchLangPackReq", ResponseProto: "com.mineplanet.pojo.AppFetchLangPackRsp",
				TarsRequestType: "vector<byte>", TarsResponseType: "vector<byte>", TimeoutMs: 5000,
				AuthRequired: true, AuditRequired: false,
			},
			{
				RequestMax: 6000, RequestMin: 6007, RouteKey: "6000:6007",
				CommandName: "GetLangDifference",
				ResponseMax: 6000, ResponseMin: 6008,
				TarsApp: "CaiRobot", TarsServer: "I18nServer", TarsServant: "I18nObj",
				TarsModule: "CaiRobotI18nApp", TarsInterface: "I18nObj", TarsMethod: "GetLangDifference",
				RequestProto: "com.mineplanet.pojo.AppFetchLangDiffReq", ResponseProto: "com.mineplanet.pojo.AppFetchLangDiffRsp",
				TarsRequestType: "vector<byte>", TarsResponseType: "vector<byte>", TimeoutMs: 5000,
				AuthRequired: true, AuditRequired: false,
			},
		},
	}
}
