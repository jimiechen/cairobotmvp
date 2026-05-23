package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// 确保 mysql 驱动已注册
// 实际驱动应在 main 包中导入: _ "github.com/go-sql-driver/mysql"
func init() {
	// 驱动注册由 main 包负责
}

// MySQLConfigRepo ConfigRepository 的 MySQL 实现
type MySQLConfigRepo struct {
	db *sql.DB
}

// NewMySQLConfigRepo 创建 MySQL 配置仓库实例
func NewMySQLConfigRepo(cfg *config.MySQLConfig) (*MySQLConfigRepo, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL 失败: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	if lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		db.SetConnMaxLifetime(lifetime)
	}

	if idleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
		db.SetConnMaxIdleTime(idleTime)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	return &MySQLConfigRepo{db: db}, nil
}

// GetLatestVersion 获取指定模块和环境下的最新已发布版本
func (r *MySQLConfigRepo) GetLatestVersion(moduleKey, env string) (*domain.ConfigVersion, error) {
	query := `
		SELECT id, module_key, env, version, config_json, is_published, published_at, created_at, updated_at, create_by, update_by
		FROM sys_config_version
		WHERE module_key = ? AND env = ? AND is_published = 1
		ORDER BY version DESC
		LIMIT 1
	`
	row := r.db.QueryRow(query, moduleKey, env)
	return scanMySQLConfigVersion(row)
}

// GetByModuleAndVersion 根据模块和版本号获取配置
func (r *MySQLConfigRepo) GetByModuleAndVersion(moduleKey, env string, version int64) (*domain.ConfigVersion, error) {
	query := `
		SELECT id, module_key, env, version, config_json, is_published, published_at, created_at, updated_at, create_by, update_by
		FROM sys_config_version
		WHERE module_key = ? AND env = ? AND version = ?
	`
	row := r.db.QueryRow(query, moduleKey, env, version)
	return scanMySQLConfigVersion(row)
}

// ListPublishedVersions 列出指定环境下所有已发布的最新版本
func (r *MySQLConfigRepo) ListPublishedVersions(env string) ([]*domain.ConfigVersion, error) {
	query := `
		SELECT v.id, v.module_key, v.env, v.version, v.config_json, v.is_published, v.published_at, v.created_at, v.updated_at, v.create_by, v.update_by
		FROM sys_config_version v
		INNER JOIN (
			SELECT module_key, MAX(version) as max_version
			FROM sys_config_version
			WHERE env = ? AND is_published = 1
			GROUP BY module_key
		) vm ON v.module_key = vm.module_key AND v.version = vm.max_version
		WHERE v.env = ?
	`
	rows, err := r.db.Query(query, env, env)
	if err != nil {
		return nil, fmt.Errorf("查询已发布版本失败: %w", err)
	}
	defer rows.Close()

	var versions []*domain.ConfigVersion
	for rows.Next() {
		ver, err := scanMySQLConfigVersionRows(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, ver)
	}
	return versions, rows.Err()
}

// Save 保存配置版本
func (r *MySQLConfigRepo) Save(version *domain.ConfigVersion) error {
	query := `
		INSERT INTO sys_config_version (module_key, env, version, config_json, is_published, published_at, create_by, update_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			config_json = VALUES(config_json),
			is_published = VALUES(is_published),
			published_at = VALUES(published_at),
			update_by = VALUES(update_by),
			updated_at = NOW()
	`
	_, err := r.db.Exec(query,
		version.ModuleKey,
		version.Env,
		version.Version,
		version.ConfigJSON,
		version.IsPublished,
		version.PublishedAt,
		version.CreateBy,
		version.UpdateBy,
	)
	if err != nil {
		return fmt.Errorf("保存配置版本失败: %w", err)
	}
	return nil
}

// scanMySQLConfigVersion 从 sql.Row 扫描 ConfigVersion（MySQL 实现）
func scanMySQLConfigVersion(row *sql.Row) (*domain.ConfigVersion, error) {
	var ver domain.ConfigVersion
	var publishedAt sql.NullTime
	var createBy, updateBy sql.NullString

	err := row.Scan(
		&ver.ID,
		&ver.ModuleKey,
		&ver.Env,
		&ver.Version,
		&ver.ConfigJSON,
		&ver.IsPublished,
		&publishedAt,
		&ver.CreatedAt,
		&ver.UpdatedAt,
		&createBy,
		&updateBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("扫描配置版本失败: %w", err)
	}

	if publishedAt.Valid {
		ver.PublishedAt = &publishedAt.Time
	}
	if createBy.Valid {
		ver.CreateBy = createBy.String
	}
	if updateBy.Valid {
		ver.UpdateBy = updateBy.String
	}

	return &ver, nil
}

// scanMySQLConfigVersionRows 从 sql.Rows 扫描 ConfigVersion（MySQL 实现）
func scanMySQLConfigVersionRows(rows *sql.Rows) (*domain.ConfigVersion, error) {
	var ver domain.ConfigVersion
	var publishedAt sql.NullTime
	var createBy, updateBy sql.NullString

	err := rows.Scan(
		&ver.ID,
		&ver.ModuleKey,
		&ver.Env,
		&ver.Version,
		&ver.ConfigJSON,
		&ver.IsPublished,
		&publishedAt,
		&ver.CreatedAt,
		&ver.UpdatedAt,
		&createBy,
		&updateBy,
	)
	if err != nil {
		return nil, fmt.Errorf("扫描配置版本失败: %w", err)
	}

	if publishedAt.Valid {
		ver.PublishedAt = &publishedAt.Time
	}
	if createBy.Valid {
		ver.CreateBy = createBy.String
	}
	if updateBy.Valid {
		ver.UpdateBy = updateBy.String
	}

	return &ver, nil
}
