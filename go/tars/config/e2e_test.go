package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	"github.com/jimiechen/mineplanet/go/services/config/cache"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

const seedSchemaSQL = `
INSERT OR IGNORE INTO sys_config_schema (module_key, field_key, field_type, default_value, description, sort_order) VALUES
('base_cfg', 'domain_root', 'string', '', 'API 根域名', 1),
('base_cfg', 'domain_wap', 'string', '', 'WAP 页面域名', 2),
('base_cfg', 'sign_rand', 'string', '', '签名随机盐值', 3),
('base_cfg', 'construct_email', 'string', '', '反馈联系邮箱', 4),
('wap_cfg', 'user_agreement_url', 'string', '', '用户协议 URL', 1),
('wap_cfg', 'privacy_policy_url', 'string', '', '隐私政策 URL', 2),
('regex_cfg', 'regex_email', 'string', '', '邮箱正则表达式', 1),
('regex_cfg', 'regex_password', 'string', '', '密码正则表达式', 2),
('regex_cfg', 'regex_phone', 'string', '', '手机号正则表达式', 3),
('regex_cfg', 'regex_nick', 'string', '', '昵称正则表达式', 4),
('regex_cfg', 'regex_circle_name', 'string', '', '圈子名称正则表达式', 5),
('oss_cfg', 'oss_host', 'string', '', 'OSS 主机地址', 1),
('oss_cfg', 'oss_domain', 'string', '', 'OSS 域名', 2),
('oss_cfg', 'cdn_domain', 'string', '', 'CDN 域名', 3),
('lang_cfg', 'lang_code', 'string', 'zh-CN', '默认语言代码', 1),
('mute_cfg', 'durations', 'json', '[]', '静音时长选项列表', 1),
('group_cfg', 'group_config_pay_notice', 'string', '', '群组支付公告文案', 1);
`

func setupFullStack(t *testing.T) (*tarsclient.LocalInvoker, *repository.SQLiteConfigRepo, *repository.SQLiteSchemaRepo) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-config.db")

	configRepo, err := repository.NewSQLiteConfigRepo(dbPath)
	if err != nil {
		t.Fatalf("创建 configRepo 失败: %v", err)
	}
	t.Cleanup(func() { configRepo.Close() })

	db := configRepo.DB()

	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	lruCache := cache.NewMockCache()
	configSvc := service.NewAppConfigService(configRepo, schemaRepo, lruCache)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterConfigI18nHandlers(invoker, configSvc, nil)

	return invoker, configRepo, schemaRepo
}

func seedBaseData(t *testing.T, repo *repository.SQLiteConfigRepo) {
	t.Helper()
	db := repo.DB()
	if _, err := db.Exec(seedSchemaSQL); err != nil {
		t.Fatalf("插入种子 Schema 数据失败: %v", err)
	}
}

func insertPublishedVersion(t *testing.T, repo *repository.SQLiteConfigRepo, moduleKey, configJSON string) {
	t.Helper()
	db := repo.DB()
	_, err := db.Exec(
		`INSERT INTO sys_config_version (module_key, env, version, config_json, is_published, published_at, create_by, update_by) VALUES (?, 'dev', 1, ?, 1, '2026-01-01 00:00:00', 'e2e-test', '')`,
		moduleKey, configJSON,
	)
	if err != nil {
		t.Fatalf("插入版本数据 %s 失败: %v", moduleKey, err)
	}
}

func TestGetAppConfigs_E2E(t *testing.T) {
	invoker, repo, _ := setupFullStack(t)
	seedBaseData(t, repo)

	insertPublishedVersion(t, repo, "base_cfg",
		`{"domain_root":"https://api.cairobot.dev","domain_wap":"https://w.cairobot.dev","sign_rand":"s@lt123","construct_email":"feedback@cairobot.dev"}`)
	insertPublishedVersion(t, repo, "lang_cfg", `{"lang_code":"zh-CN"}`)

	reqBody, _ := json.Marshal(service.AppConfigRequest{
		Env:           "dev",
		ClientScope:   "all",
		ClientVersion: "1.0.0",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "ConfigServer",
		Servant: "ConfigObj",
		Method:  "GetAppConfigs",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var appResp service.AppConfigResponse
	if err := json.Unmarshal(resp, &appResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	if _, ok := appResp.StaticModules["base_cfg"]; !ok {
		t.Error("StaticModules 应包含 base_cfg")
	}
	if _, ok := appResp.StaticModules["lang_cfg"]; !ok {
		t.Error("StaticModules 应包含 lang_cfg")
	}

	if len(appResp.DynamicModules) != 0 {
		t.Errorf("DynamicModules 应为空（种子数据只有静态模块），实际 %d 条", len(appResp.DynamicModules))
	}

	baseCfg := appResp.StaticModules["base_cfg"]
	if domainVal, ok := baseCfg["domain_root"]; ok {
		if domainVal.String() != "https://api.cairobot.dev" {
			t.Errorf("domain_root 期望 https://api.cairobot.dev，实际 %s", domainVal.String())
		}
	} else {
		t.Error("base_cfg 应包含 domain_root 字段")
	}
}

func TestAppConfigVersion_E2E(t *testing.T) {
	invoker, repo, _ := setupFullStack(t)

	for _, mod := range []string{"base_cfg", "lang_cfg"} {
		insertPublishedVersion(t, repo, mod, `{}`)
	}

	reqBody, _ := json.Marshal(map[string]any{
		"env":           "dev",
		"known_versions": map[string]int64{},
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "ConfigServer",
		Servant: "ConfigObj",
		Method:  "AppConfigVersion",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var verResp service.VersionInfoResponse
	if err := json.Unmarshal(resp, &verResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	if len(verResp.ConfigVersions) < 2 {
		t.Errorf("ConfigVersions 应至少包含 2 个模块，实际 %d", len(verResp.ConfigVersions))
	}
	if ver, ok := verResp.ConfigVersions["base_cfg"]; !ok || ver != 1 {
		t.Error("ConfigVersions 应包含 base_cfg -> version 1")
	}
	if ver, ok := verResp.ConfigVersions["lang_cfg"]; !ok || ver != 1 {
		t.Error("ConfigVersions 应包含 lang_cfg -> version 1")
	}
	if !verResp.HasChanges {
		t.Error("已知版本为空时 HasChanges 应为 true")
	}
}

func TestDynamicModule_NewField_NoCodeChange_E2E(t *testing.T) {
	invoker, repo, schemaRepo := setupFullStack(t)
	seedBaseData(t, repo)

	insertPublishedVersion(t, repo, "base_cfg", `{"domain_root":"https://api.cairobot.dev"}`)

	for _, fs := range []struct {
		key   string
		typ   string
		def   string
		desc  string
		order int
	}{
		{"push_enabled", "bool", "true", "推送开关", 1},
		{"push_title", "string", "", "推送标题模板", 2},
		{"quiet_hours_start", "string", "22:00", "免打扰开始时间", 3},
	} {
		err := schemaRepo.Create(&domain.FieldSchema{
			ModuleKey:    "notification_cfg",
			FieldKey:     fs.key,
			FieldType:    domain.FieldType(fs.typ),
			DefaultValue: fs.def,
			Description:  fs.desc,
			SortOrder:    fs.order,
			IsEnabled:    true,
			ClientScope:  "all",
		})
		if err != nil {
			t.Fatalf("创建 schema %s 失败: %v", fs.key, err)
		}
	}

	insertPublishedVersion(t, repo, "notification_cfg",
		`{"push_enabled":true,"push_title":"CaiRobot通知","quiet_hours_start":"22:00"}`)

	reqBody, _ := json.Marshal(service.AppConfigRequest{
		Env:           "dev",
		ClientScope:   "all",
		ClientVersion: "1.0.0",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "ConfigServer",
		Servant: "ConfigObj",
		Method:  "GetAppConfigs",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var appResp service.AppConfigResponse
	if err := json.Unmarshal(resp, &appResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	if len(appResp.DynamicModules) == 0 {
		t.Fatal("DynamicModules 应包含 notification_cfg 模块")
	}

	found := false
	for _, dm := range appResp.DynamicModules {
		if dm.ModuleKey == "notification_cfg" {
			found = true
			if dm.Version != 1 {
				t.Errorf("notification_cfg 版本期望 1，实际 %d", dm.Version)
			}

			if pushVal, ok := dm.Fields["push_enabled"]; !ok {
				t.Error("notification_cfg.Fields 应包含 push_enabled")
			} else if pushVal.Bool() != true {
				t.Errorf("push_enabled 期望 true，实际 %v", pushVal.Bool())
			}

			if titleVal, ok := dm.Fields["push_title"]; !ok {
				t.Error("notification_cfg.Fields 应包含 push_title")
			} else if titleVal.String() != "CaiRobot通知" {
				t.Errorf("push_title 期望 CaiRobot通知，实际 %s", titleVal.String())
			}

			descKeys := make(map[string]bool)
			for _, desc := range dm.Descriptors {
				descKeys[desc.FieldKey] = true
			}
			for _, expectedKey := range []string{"push_enabled", "push_title", "quiet_hours_start"} {
				if !descKeys[expectedKey] {
					t.Errorf("Descriptors 应包含 %s", expectedKey)
				}
			}
			break
		}
	}
	if !found {
		t.Error("DynamicModules 中未找到 notification_cfg")
	}

	if _, ok := appResp.StaticModules["base_cfg"]; !ok {
		t.Error("StaticModules 仍应包含 base_cfg（静态模块不受动态模块影响）")
	}
}
