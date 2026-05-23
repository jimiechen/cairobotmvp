package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "modernc.org/sqlite"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	"github.com/jimiechen/mineplanet/go/services/i18n/cache"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	fmt.Println("🌍 CaiRobot I18n Server starting...")

	dbPath := os.Getenv("I18N_DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(".", "data", "i18n.db")
	}
	dir := filepath.Dir(dbPath)
	os.MkdirAll(dir, 0o755)

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("❌ 打开 SQLite 失败: %v", err)
	}
	defer db.Close()

	initI18nTables(db)

	i18nRepo := repository.NewSQLiteRepo(db)
	lruCache := cache.NewMockCache()
	i18nSvc := service.NewI18nService(i18nRepo, lruCache, "dev")

	// 注册 I18n 模块的本地 TarsGo servant handler 到 LocalInvoker
	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterConfigI18nHandlers(invoker, nil, i18nSvc)

	fmt.Println("✅ I18n Server started successfully")
	fmt.Printf("   Database: %s\n", dbPath)
	fmt.Println("   Methods: GetAppLanguage, GetLangPack, GetLangDifference")
	fmt.Println("   Invoker: LocalInvoker ready (monolith mode)")
	fmt.Println("\n📡 Press Ctrl+C to stop")

	<-ctx.Done()
	fmt.Println("\n👋 I18n Server stopped")
}

func initI18nTables(db *sql.DB) {
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
		published_by  TEXT,
		created_at    TEXT DEFAULT (datetime('now')),
		updated_at    TEXT DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_lang_pack_lang_code_env ON sys_lang_pack (lang_code, env);

	CREATE TABLE IF NOT EXISTS sys_lang_string (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		pack_id         INTEGER NOT NULL,
		string_key      TEXT NOT NULL,
		string_value    TEXT NOT NULL,
		group_name      TEXT DEFAULT 'common',
		version         INTEGER NOT NULL DEFAULT 1,
		operation_type  TEXT NOT NULL DEFAULT 'ADD',
		prev_value      TEXT,
		template_type   TEXT NOT NULL DEFAULT 'plain',
		params_schema   TEXT,
		preview_sample  TEXT,
		created_at      TEXT DEFAULT (datetime('now')),
		updated_at      TEXT DEFAULT (datetime('now'))
	);
	CREATE INDEX IF NOT EXISTS idx_lang_string_pack_id ON sys_lang_string (pack_id);
	`
	if _, err := db.Exec(schemaSQL); err != nil {
		log.Fatalf("❌ 初始化 i18n 表失败: %v", err)
	}
}
