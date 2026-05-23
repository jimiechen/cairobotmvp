package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	"github.com/jimiechen/mineplanet/go/services/i18n/cache"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"

	_ "modernc.org/sqlite"
)

const i18nDDL = `
DROP TABLE IF EXISTS sys_lang_string;
DROP TABLE IF EXISTS sys_lang_pack;

CREATE TABLE sys_lang_pack (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	pack_name     TEXT NOT NULL DEFAULT '',
	env           TEXT NOT NULL DEFAULT 'dev',
	version       INTEGER NOT NULL DEFAULT 1,
	lang_code     TEXT NOT NULL,
	description   TEXT,
	is_published  INTEGER NOT NULL DEFAULT 0,
	published_at  TEXT,
	published_by  INTEGER,
	created_at    TEXT,
	updated_at    TEXT,
	UNIQUE (pack_name, env, lang_code)
);
CREATE INDEX idx_lang_pack_published ON sys_lang_pack (is_published);

CREATE TABLE sys_lang_string (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	pack_id        INTEGER NOT NULL,
	string_key     TEXT NOT NULL,
	string_value   TEXT NOT NULL,
	group_name     TEXT NOT NULL DEFAULT 'common',
	version        INTEGER NOT NULL DEFAULT 1,
	operation_type TEXT NOT NULL DEFAULT 'ADD',
	prev_value     TEXT,
	template_type  TEXT DEFAULT 'plain',
	params_schema  TEXT DEFAULT NULL,
	preview_sample TEXT DEFAULT NULL,
	created_at     TEXT,
	updated_at     TEXT,
	UNIQUE (pack_id, string_key)
);
CREATE INDEX idx_lang_str_pack_id ON sys_lang_string (pack_id);
`

func openI18nDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, t.Name()+".db")

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		t.Fatalf("打开 SQLite 失败: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(i18nDDL); err != nil {
		t.Fatalf("执行 DDL 失败: %v", err)
	}
	return db
}

func setupI18nStack(t *testing.T, db *sql.DB) *tarsclient.LocalInvoker {
	t.Helper()

	i18nRepo := repository.NewSQLiteRepo(db)
	lruCache := cache.NewMockCache()
	i18nSvc := service.NewI18nService(i18nRepo, lruCache, "dev")

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterConfigI18nHandlers(invoker, nil, i18nSvc)

	return invoker
}

func seedLangPacks(t *testing.T, db *sql.DB) map[string]int64 {
	t.Helper()
	packIDs := make(map[string]int64)
	nowStr := time.Now().UTC().Format("2006-01-02 15:04:05")

	for _, p := range []struct {
		packName string
		langCode string
		desc     string
	}{
		{"webp", "zh-CN", "简体中文"},
		{"webp", "en", "English"},
	} {
		result, err := db.Exec(
			`INSERT INTO sys_lang_pack (pack_name, env, lang_code, version, description, is_published, published_by, created_at, updated_at) VALUES (?, 'dev', ?, 1, ?, 1, 0, ?, ?)`,
			p.packName, p.langCode, p.desc, nowStr, nowStr,
		)
		if err != nil {
			t.Fatalf("插入语言包 %s 失败: %v", p.langCode, err)
		}
		id, _ := result.LastInsertId()
		packIDs[p.langCode] = id
	}
	return packIDs
}

func seedStrings(t *testing.T, db *sql.DB, packID int64, strings []struct {
	key, value, group, tmplType, paramsSchema string
}) {
	t.Helper()
	for _, s := range strings {
		_, err := db.Exec(
			`INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema) 
			 VALUES (?, ?, ?, ?, 1, 'ADD', ?, ?)`,
			packID, s.key, s.value, s.group, s.tmplType, s.paramsSchema,
		)
		if err != nil {
			t.Fatalf("插入字符串 %s 失败: %v", s.key, err)
		}
	}
}

func TestGetAppLanguage_E2E(t *testing.T) {
	db := openI18nDB(t)
	seedLangPacks(t, db)
	invoker := setupI18nStack(t, db)

	reqBody, _ := json.Marshal(map[string]string{
		"client_version": "1.0.0",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetAppLanguage",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var languages []service.LanguageMeta
	if err := json.Unmarshal(resp, &languages); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	if len(languages) < 2 {
		t.Fatalf("期望至少 2 种语言，实际 %d", len(languages))
	}

	foundZhCN := false
	foundEn := false
	for _, lang := range languages {
		if lang.Code == "zh-CN" {
			foundZhCN = true
		}
		if lang.Code == "en" {
			foundEn = true
		}
	}
	if !foundZhCN {
		t.Error("语言列表应包含 zh-CN")
	}
	if !foundEn {
		t.Error("语言列表应包含 en")
	}
}

func TestGetLangPack_E2E(t *testing.T) {
	db := openI18nDB(t)
	packIDs := seedLangPacks(t, db)

	zhCNID := packIDs["zh-CN"]
	enID := packIDs["en"]

	seedStrings(t, db, zhCNID, []struct {
		key, value, group, tmplType, paramsSchema string
	}{
		{"svc_common_ok", "确定", "common", "plain", ""},
		{"svc_common_cancel", "取消", "common", "plain", ""},
		{"svc_msg_welcome", "欢迎 {name}，你有 {count} 条新消息", "app", "named",
			`[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`},
	})
	seedStrings(t, db, enID, []struct {
		key, value, group, tmplType, paramsSchema string
	}{
		{"svc_common_ok", "OK", "common", "plain", ""},
		{"svc_msg_welcome", "Welcome {name}, you have {count} new messages", "app", "named",
			`[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`},
	})

	invoker := setupI18nStack(t, db)

	reqBody, _ := json.Marshal(map[string]string{
		"lang_code":      "zh-CN",
		"client_version": "2.0.0",
		"env":            "dev",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetLangPack",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var packResp service.LangPackResponse
	if err := json.Unmarshal(resp, &packResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	if packResp.PackVersion <= 0 {
		t.Errorf("PackVersion 应 > 0，实际 %d", packResp.PackVersion)
	}

	keyMap := make(map[string]service.LangStringEntry)
	for _, entry := range packResp.Strings {
		keyMap[entry.Key] = entry
	}

	if _, ok := keyMap["svc_common_ok"]; !ok {
		t.Error("Strings 应包含 svc_common_ok")
	}
	if _, ok := keyMap["svc_msg_welcome"]; !ok {
		t.Error("Strings 应包含 svc_msg_welcome")
	}

	welcome, ok := keyMap["svc_msg_welcome"]
	if !ok {
		t.Fatal("缺少 svc_msg_welcome 条目")
	}
	if welcome.TemplateType != "named" {
		t.Errorf("svc_msg_welcome 的 template_type 期望 named，实际 %s", welcome.TemplateType)
	}
	if len(welcome.Params) < 2 {
		t.Errorf("svc_msg_welcome 的 Params 应至少有 2 个参数，实际 %d", len(welcome.Params))
	}

	paramNames := make(map[string]bool)
	for _, p := range welcome.Params {
		paramNames[p.Name] = true
	}
	for _, expected := range []string{"name", "count"} {
		if !paramNames[expected] {
			t.Errorf("Params 应包含 %s 参数", expected)
		}
	}
}

func TestGetLangDifference_E2E(t *testing.T) {
	db := openI18nDB(t)
	packIDs := seedLangPacks(t, db)

	zhCNID := packIDs["zh-CN"]

	seedStrings(t, db, zhCNID, []struct {
		key, value, group, tmplType, paramsSchema string
	}{
		{"svc_common_ok", "确定", "common", "plain", ""},
		{"svc_msg_hello", "你好", "greeting", "plain", ""},
	})

	_, err := db.Exec(
		`INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema) 
		 VALUES (?, 'svc_new_feature', '发现新功能', 'app', 2, 'ADD', 'plain', '')`,
		zhCNID,
	)
	if err != nil {
		t.Fatalf("插入增量字符串失败: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema) 
		 VALUES (?, 'svc_updated_tip', '提示已更新', 'app', 2, 'MOD', 'plain', '')`,
		zhCNID,
	)
	if err != nil {
		t.Fatalf("插入修改字符串失败: %v", err)
	}

	invoker := setupI18nStack(t, db)

	reqBody, _ := json.Marshal(map[string]any{
		"lang_code":      "zh-CN",
		"since_version":  int64(1),
		"client_version": "2.0.0",
		"env":            "dev",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetLangDifference",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var diffResp service.LangDiffResponse
	if err := json.Unmarshal(resp, &diffResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	addKeys := make(map[string]bool)
	for _, a := range diffResp.Additions {
		addKeys[a.Key] = true
	}

	if !addKeys["svc_new_feature"] {
		t.Error("Additions 应包含 svc_new_feature")
	}
	if !addKeys["svc_updated_tip"] {
		t.Error("Additions 应包含 svc_updated_tip（MOD 类型也属于 additions）")
	}
}

func TestNamedTemplate_NewKey_E2E(t *testing.T) {
	db := openI18nDB(t)
	packIDs := seedLangPacks(t, db)

	zhCNID := packIDs["zh-CN"]

	seedStrings(t, db, zhCNID, []struct {
		key, value, group, tmplType, paramsSchema string
	}{
		{"svc_common_ok", "确定", "common", "plain", ""},
		{"svc_msg_welcome", "欢迎 {name}，你有 {count} 条新消息", "app", "named",
			`[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`},
	})

	_, err := db.Exec(
		`INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema) 
		 VALUES (?, 'svc_new_greeting', '你好 {user}，欢迎回来！', 'greeting', 2, 'ADD', 'named',
		 '[{"name":"user","type":"string","required":true,"description":"用户名"}]')`,
		zhCNID,
	)
	if err != nil {
		t.Fatalf("插入 named 模板字符串失败: %v", err)
	}

	_, err = db.Exec(`UPDATE sys_lang_pack SET version = 2 WHERE id = ?`, zhCNID)
	if err != nil {
		t.Fatalf("更新语言包版本失败: %v", err)
	}

	invoker := setupI18nStack(t, db)

	reqBody, _ := json.Marshal(map[string]string{
		"lang_code":      "zh-CN",
		"client_version": "2.0.0",
		"env":            "dev",
	})

	target := tarsclient.Target{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetLangPack",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, reqBody, nil)
	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if code != commonlib.CodeSuccess {
		t.Fatalf("期望返回码 %d，实际 %d", commonlib.CodeSuccess, code)
	}

	var packResp service.LangPackResponse
	if err := json.Unmarshal(resp, &packResp); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	keyMap := make(map[string]service.LangStringEntry)
	for _, entry := range packResp.Strings {
		keyMap[entry.Key] = entry
	}

	newGreeting, ok := keyMap["svc_new_greeting"]
	if !ok {
		t.Fatal("Strings 应包含新增的 svc_new_greeting 条目")
	}
	if newGreeting.TemplateType != "named" {
		t.Errorf("svc_new_greeting 的 template_type 期望 named，实际 %s", newGreeting.TemplateType)
	}
	if newGreeting.Value != "你好 {user}，欢迎回来！" {
		t.Errorf("svc_new_greeting 的 Value 不匹配，实际 %s", newGreeting.Value)
	}

	if len(newGreeting.Params) == 0 {
		t.Fatal("svc_new_greeting 的 Params 不应为空")
	}
	foundUserParam := false
	for _, p := range newGreeting.Params {
		if p.Name == "user" {
			foundUserParam = true
			if p.Type != "string" {
				t.Errorf("user 参数的 type 期望 string，实际 %s", p.Type)
			}
			if !p.Required {
				t.Error("user 参数的 required 期望 true")
			}
		}
	}
	if !foundUserParam {
		t.Error("Params 应包含 user 参数描述")
	}
}
