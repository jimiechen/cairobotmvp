package repository

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// newTestSQLiteRepo 创建内存 SQLite 测试仓库，自动初始化表结构
// 每个测试用例获得独立的数据库实例，避免测试间数据干扰
func newTestSQLiteRepo(t *testing.T) *SQLiteRepo {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}

	if err := initI18nTables(db); err != nil {
		db.Close()
		t.Fatalf("初始化表结构失败: %v", err)
	}

	repo := NewSQLiteRepo(db)
	t.Cleanup(func() { db.Close() })
	return repo
}

// initI18nTables 初始化 i18n 相关的 SQLite 表结构
// 用于测试环境，确保表结构与生产查询字段一致
func initI18nTables(db *sql.DB) error {
	schemaSQL := `
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
	CREATE INDEX IF NOT EXISTS idx_lang_pack_code_env 
		ON sys_lang_pack (lang_code, env, is_published);

	CREATE TABLE IF NOT EXISTS sys_lang_string (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		pack_id         INTEGER NOT NULL,
		string_key      TEXT NOT NULL,
		string_value    TEXT NOT NULL,
		group_name      TEXT,
		version         INTEGER NOT NULL DEFAULT 1,
		operation_type  TEXT NOT NULL DEFAULT 'ADD',
		prev_value      TEXT,
		template_type   TEXT NOT NULL DEFAULT 'PLAIN',
		params_schema   TEXT,
		preview_sample  TEXT,
		created_at      TEXT DEFAULT (datetime('now')),
		updated_at      TEXT DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_lang_string_pack_id 
		ON sys_lang_string (pack_id, operation_type);
	CREATE INDEX IF NOT EXISTS idx_lang_string_version 
		ON sys_lang_string (pack_id, version);
	`
	_, err := db.Exec(schemaSQL)
	return err
}

// insertTestPack 插入测试语言包数据，返回插入的 ID
func insertTestPack(t *testing.T, db *sql.DB, pack domain.LangPack) int64 {
	t.Helper()
	result, err := db.Exec(
		`INSERT INTO sys_lang_pack (pack_name, env, version, lang_code, description, is_published, published_at, published_by) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		pack.PackName, pack.Env, pack.Version, pack.LangCode,
		pack.Description, boolToInt(pack.IsPublished), formatTime(pack.PublishedAt), pack.PublishedBy,
	)
	if err != nil {
		t.Fatalf("插入语言包失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

// insertTestString 插入测试字符串数据
func insertTestString(t *testing.T, db *sql.DB, s domain.LangString) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, prev_value, template_type, params_schema, preview_sample) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.PackID, s.StringKey, s.StringValue, s.GroupName, s.Version,
		string(s.OperationType), s.PrevValue, string(s.TemplateType),
		s.ParamsSchema, s.PreviewSample,
	)
	if err != nil {
		t.Fatalf("插入字符串失败: %v", err)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// TestGetPackByLangCode 命中返回场景
func TestGetPackByLangCode_命中返回(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	now := time.Now()
	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName:    "中文包",
		Env:         "prod",
		Version:     2,
		LangCode:    "zh-CN",
		Description: "简体中文",
		IsPublished: true,
		PublishedAt: &now,
		PublishedBy: 1001,
	})

	got, err := repo.GetPackByLangCode("zh-CN", "prod")
	if err != nil {
		t.Fatalf("GetPackByLangCode 失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回已发布的语言包")
	}
	if got.LangCode != "zh-CN" || got.Env != "prod" {
		t.Errorf("数据不匹配: lang=%s env=%s", got.LangCode, got.Env)
	}
	if got.PackName != "中文包" {
		t.Errorf("PackName 不匹配: %s", got.PackName)
	}
	if !got.IsPublished {
		t.Error("IsPublished 应为 true")
	}
	if got.PublishedBy != 1001 {
		t.Errorf("PublishedBy 不匹配: %d", got.PublishedBy)
	}
}

// TestGetPackByLangCode 未命中返回nil
func TestGetPackByLangCode_未命中返回nil(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	got, err := repo.GetPackByLangCode("ko-KR", "dev")
	if err != nil {
		t.Fatalf("GetPackByLangCode 失败: %v", err)
	}
	if got != nil {
		t.Error("未找到时应返回 nil")
	}
}

// TestGetPackByLangCode 未发布不应命中
func TestGetPackByLangCode_未发布不应命中(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName:    "草稿包",
		Env:         "dev",
		Version:     1,
		LangCode:    "ja-JP",
		IsPublished: false,
	})

	got, err := repo.GetPackByLangCode("ja-JP", "dev")
	if err != nil {
		t.Fatalf("GetPackByLangCode 失败: %v", err)
	}
	if got != nil {
		t.Error("未发布的语言包不应被返回")
	}
}

// TestListPacks 返回多行数据
func TestListPacks_返回多行(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	for _, p := range []domain.LangPack{
		{PackName: "中文", Env: "dev", Version: 1, LangCode: "zh-CN", IsPublished: true},
		{PackName: "英文", Env: "dev", Version: 1, LangCode: "en-US", IsPublished: true},
		{PackName: "日文", Env: "dev", Version: 1, LangCode: "ja-JP", IsPublished: true},
	} {
		insertTestPack(t, repo.DB(), p)
	}

	packs, err := repo.ListPacks("dev")
	if err != nil {
		t.Fatalf("ListPacks 失败: %v", err)
	}
	if len(packs) != 3 {
		t.Fatalf("期望 3 条记录, 实际 %d", len(packs))
	}
}

// TestListPacks 空表返回空切片
func TestListPacks_空表返回空(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	packs, err := repo.ListPacks("staging")
	if err != nil {
		t.Fatalf("ListPacks 失败: %v", err)
	}
	if len(packs) != 0 {
		t.Errorf("期望 0 条记录, 实际 %d", len(packs))
	}
}

// TestListPacks 只返回指定环境的包
func TestListPacks_只返回指定环境(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "Dev中文", Env: "dev", Version: 1, LangCode: "zh-CN", IsPublished: true,
	})
	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "Prod中文", Env: "prod", Version: 2, LangCode: "zh-CN", IsPublished: true,
	})

	devPacks, _ := repo.ListPacks("dev")
	if len(devPacks) != 1 {
		t.Errorf("dev 环境应只有 1 条, 实际 %d", len(devPacks))
	}
}

// TestGetLangPackByID 正常返回
func TestGetLangPackByID_正常返回(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	now := time.Now()

	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName:    "测试包",
		Env:         "test",
		Version:     5,
		LangCode:    "fr-FR",
		Description: "法语测试包",
		IsPublished: true,
		PublishedAt: &now,
		PublishedBy: 2001,
	})

	got, err := repo.GetLangPackByID(packID)
	if err != nil {
		t.Fatalf("GetLangPackByID 失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回语言包")
	}
	if got.ID != packID {
		t.Errorf("ID 不匹配: 期望 %d, 实际 %d", packID, got.ID)
	}
	if got.LangCode != "fr-FR" {
		t.Errorf("LangCode 不匹配: %s", got.LangCode)
	}
	if got.Description != "法语测试包" {
		t.Errorf("Description 不匹配: %s", got.Description)
	}
}

// TestGetLangPackByID 不存在返回nil
func TestGetLangPackByID_不存在返回nil(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	got, err := repo.GetLangPackByID(99999)
	if err != nil {
		t.Fatalf("GetLangPackByID 失败: %v", err)
	}
	if got != nil {
		t.Error("不存在的 ID 应返回 nil")
	}
}

// TestGetLangPackByID 包含未发布记录
func TestGetLangPackByID_包含未发布记录(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "草稿", Env: "dev", Version: 1, LangCode: "de-DE", IsPublished: false,
	})

	got, err := repo.GetLangPackByID(packID)
	if err != nil {
		t.Fatalf("GetLangPackByID 失败: %v", err)
	}
	if got == nil {
		t.Fatal("GetLangPackByID 应返回未发布的记录")
	}
	if got.IsPublished {
		t.Error("该记录应为未发布状态")
	}
}

// TestGetLangPackByID 验证所有字段映射
func TestGetLangPackByID_验证所有字段(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	now := time.Now()

	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName:    "全字段包",
		Env:         "staging",
		Version:     10,
		LangCode:    "pt-BR",
		Description: "葡萄牙语测试",
		IsPublished: true,
		PublishedAt: &now,
		PublishedBy: 9999,
	})

	got, _ := repo.GetLangPackByID(packID)
	if got.PackName != "全字段包" {
		t.Errorf("PackName 不匹配: %s", got.PackName)
	}
	if got.Env != "staging" {
		t.Errorf("Env 不匹配: %s", got.Env)
	}
	if got.Version != 10 {
		t.Errorf("Version 不匹配: %d", got.Version)
	}
	if got.PublishedBy != 9999 {
		t.Errorf("PublishedBy 不匹配: %d", got.PublishedBy)
	}
	if got.PublishedAt == nil {
		t.Error("PublishedAt 不应为 nil")
	}
}

// TestListPacks 包含未发布的包不应返回
func TestListPacks_过滤未发布包(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "已发布", Env: "dev", Version: 1, LangCode: "zh-CN", IsPublished: true,
	})
	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "未发布", Env: "dev", Version: 1, LangCode: "en-US", IsPublished: false,
	})

	packs, _ := repo.ListPacks("dev")
	for _, p := range packs {
		if p.LangCode == "en-US" {
			t.Error("未发布的包不应出现在 ListPacks 结果中")
		}
	}
}

// TestNewSQLiteRepo_DB方法返回底层连接
func TestNewSQLiteRepo_DB方法(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	if repo.DB() == nil {
		t.Error("DB() 不应返回 nil")
	}
}
