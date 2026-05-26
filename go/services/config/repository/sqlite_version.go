package repository

import (
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

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
