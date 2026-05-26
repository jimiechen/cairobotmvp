package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// SQLiteConfigRepo ConfigRepository 的 SQLite 实现
// 用于单测和本地开发环境，从环境变量 CONFIG_DB_PATH 读取数据库路径
// 默认路径为项目根目录下的 data/config.db
type SQLiteConfigRepo struct {
	db *sql.DB
}

// DB 返回底层 *sql.DB 连接，供 SchemaRepo 复用
func (r *SQLiteConfigRepo) DB() *sql.DB {
	return r.db
}

// NewSQLiteConfigRepo 创建 SQLite 配置仓库实例
// 自动初始化表结构（如果不存在），适合测试和开发阶段使用
func NewSQLiteConfigRepo(dbPath string) (*SQLiteConfigRepo, error) {
	if dbPath == "" {
		dbPath = os.Getenv("CONFIG_DB_PATH")
	}
	if dbPath == "" {
		dbPath = filepath.Join(".", "data", "config.db")
	}
	dir := filepath.Dir(dbPath)
	os.MkdirAll(dir, 0o755)

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 失败: %w", err)
	}
	repo := &SQLiteConfigRepo{db: db}
	if err = repo.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化表失败: %w", err)
	}
	return repo, nil
}

// NewSQLiteConfigRepoFromDB 基于已有 *sql.DB 连接创建配置仓库实例
// 用于测试场景，调用方自行管理数据库连接和生命周期
func NewSQLiteConfigRepoFromDB(db *sql.DB) (*SQLiteConfigRepo, error) {
	repo := &SQLiteConfigRepo{db: db}
	if err := repo.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表失败: %w", err)
	}
	return repo, nil
}

func (r *SQLiteConfigRepo) initTables() error {
	schemaSQL := `
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
	CREATE INDEX IF NOT EXISTS idx_cfg_ver_module_env_version
		ON sys_config_version (module_key, env, version);
	CREATE INDEX IF NOT EXISTS idx_cfg_ver_latest
		ON sys_config_version (module_key, env, is_published, version DESC);

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
		created_at    TEXT DEFAULT (datetime('now')),
		updated_at    TEXT DEFAULT (datetime('now')),
		UNIQUE (module_key, field_key)
	);
	`
	_, err := r.db.Exec(schemaSQL)
	return err
}

// Close 关闭数据库连接，应在不再使用时调用以释放资源
func (r *SQLiteConfigRepo) Close() error {
	return r.db.Close()
}
