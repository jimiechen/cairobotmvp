package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jimiechen/mineplanet/go/services/config/domain"

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

// GetLatestVersion 查询指定模块在指定环境下最新已发布的版本
// 优先返回 is_published=true 且 version 最大的记录
func (r *SQLiteConfigRepo) GetLatestVersion(moduleKey, env string) (*domain.ConfigVersion, error) {
	query := `
	SELECT id, module_key, env, version, config_json, is_published, published_at,
	       created_at, updated_at, create_by, update_by
	FROM sys_config_version
	WHERE module_key = ? AND env = ? AND is_published = 1
	ORDER BY version DESC LIMIT 1
	`
	row := r.db.QueryRow(query, moduleKey, env)
	return scanConfigVersion(row)
}

// GetByModuleAndVersion 精确查询某模块在某环境下的特定版本
// 用于客户端增量拉取（已知 version 时直接定位）
func (r *SQLiteConfigRepo) GetByModuleAndVersion(moduleKey, env string, version int64) (*domain.ConfigVersion, error) {
	query := `
	SELECT id, module_key, env, version, config_json, is_published, published_at,
	       created_at, updated_at, create_by, update_by
	FROM sys_config_version
	WHERE module_key = ? AND env = ? AND version = ?
	LIMIT 1
	`
	row := r.db.QueryRow(query, moduleKey, env, version)
	return scanConfigVersion(row)
}

// ListPublishedVersions 列出指定环境下所有已发布的配置版本
// 用于 compose.go 批量组装 AppConfigsRsp
func (r *SQLiteConfigRepo) ListPublishedVersions(env string) ([]*domain.ConfigVersion, error) {
	query := `
	SELECT id, module_key, env, version, config_json, is_published, published_at,
	       created_at, updated_at, create_by, update_by
	FROM sys_config_version
	WHERE env = ? AND is_published = 1
	ORDER BY module_key, version DESC
	`
	rows, err := r.db.Query(query, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.ConfigVersion
	for rows.Next() {
		v, err := scanConfigVersionFromRows(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// Save 新增或更新配置版本记录
// 根据 module_key + env + version 做 upsert 语义
func (r *SQLiteConfigRepo) Save(version *domain.ConfigVersion) error {
	query := `
	INSERT INTO sys_config_version
	  (module_key, env, version, config_json, is_published, published_at, create_by, update_by)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		version.ModuleKey, version.Env, version.Version,
		version.ConfigJSON, boolToInt(version.IsPublished),
		timePtrToStr(version.PublishedAt),
		version.CreateBy, version.UpdateBy,
	)
	if err != nil {
		return fmt.Errorf("保存 config_version 失败: %w", err)
	}
	id, _ := result.LastInsertId()
	version.ID = id
	return nil
}

func scanConfigVersion(row *sql.Row) (*domain.ConfigVersion, error) {
	v := &domain.ConfigVersion{}
	var publishedAtStr, createdAtStr, updatedAtStr *string
	err := row.Scan(
		&v.ID, &v.ModuleKey, &v.Env, &v.Version, &v.ConfigJSON,
		&v.IsPublished, &publishedAtStr,
		&createdAtStr, &updatedAtStr, &v.CreateBy, &v.UpdateBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if publishedAtStr != nil {
		t, parseErr := parseTime(*publishedAtStr)
		if parseErr == nil {
			v.PublishedAt = &t
		}
	}
	if createdAtStr != nil {
		v.CreatedAt, _ = parseTime(*createdAtStr)
	}
	if updatedAtStr != nil {
		v.UpdatedAt, _ = parseTime(*updatedAtStr)
	}
	return v, nil
}

func scanConfigVersionFromRows(rows *sql.Rows) (*domain.ConfigVersion, error) {
	v := &domain.ConfigVersion{}
	var publishedAtStr, createdAtStr, updatedAtStr *string
	err := rows.Scan(
		&v.ID, &v.ModuleKey, &v.Env, &v.Version, &v.ConfigJSON,
		&v.IsPublished, &publishedAtStr,
		&createdAtStr, &updatedAtStr, &v.CreateBy, &v.UpdateBy,
	)
	if err != nil {
		return nil, err
	}
	if publishedAtStr != nil {
		t, parseErr := parseTime(*publishedAtStr)
		if parseErr == nil {
			v.PublishedAt = &t
		}
	}
	if createdAtStr != nil {
		v.CreatedAt, _ = parseTime(*createdAtStr)
	}
	if updatedAtStr != nil {
		v.UpdatedAt, _ = parseTime(*updatedAtStr)
	}
	return v, nil
}
