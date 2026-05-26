package repository

import (
	"database/sql"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// scanConfigVersion 从 sql.Row 扫描单条 ConfigVersion 记录
// 用于 QueryRow 返回的单行结果集
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
	fillTimeFields(v, publishedAtStr, createdAtStr, updatedAtStr)
	return v, nil
}

// scanConfigVersionFromRows 从 sql.Rows 扫描单条 ConfigVersion 记录
// 用于 Query 返回的多行结果集中的逐行遍历
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
	fillTimeFields(v, publishedAtStr, createdAtStr, updatedAtStr)
	return v, nil
}

// fillTimeFields 将字符串时间字段解析后填充到 ConfigVersion 结构体
// 解析失败时静默跳过，保持零值
func fillTimeFields(v *domain.ConfigVersion, publishedAtStr, createdAtStr, updatedAtStr *string) {
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
}
