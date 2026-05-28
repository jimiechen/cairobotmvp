package repository

import (
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// GetStringsByPackID 查询指定语言包下所有非删除状态的字符串
func (r *SQLiteRepo) GetStringsByPackID(packID int64) ([]domain.LangString, error) {
	query := `SELECT id, pack_id, string_key, string_value, group_name, version, operation_type, prev_value, template_type, params_schema, preview_sample, created_at, updated_at 
	          FROM sys_lang_string WHERE pack_id = ? AND operation_type != 'DEL'`

	rows, err := r.db.Query(query, packID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strings []domain.LangString
	for rows.Next() {
		var s domain.LangString
		var prevValue *string
		var previewSample *string
		var createdAtStr, updatedAtStr *string
		var groupName *string
		var paramsSchema *string

		err := rows.Scan(
			&s.ID,
			&s.PackID,
			&s.StringKey,
			&s.StringValue,
			&groupName,
			&s.Version,
			&s.OperationType,
			&prevValue,
			&s.TemplateType,
			&paramsSchema,
			&previewSample,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		if groupName != nil {
			s.GroupName = *groupName
		}
		if prevValue != nil {
			s.PrevValue = prevValue
		}
		if paramsSchema != nil {
			s.ParamsSchema = *paramsSchema
		}
		if previewSample != nil {
			s.PreviewSample = *previewSample
		}
		if createdAtStr != nil && *createdAtStr != "" {
			s.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr)
		}
		if updatedAtStr != nil && *updatedAtStr != "" {
			s.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr)
		}

		strings = append(strings, s)
	}

	return strings, rows.Err()
}

// GetDiffSince 查询指定版本之后新增或修改的字符串（不含删除）
func (r *SQLiteRepo) GetDiffSince(packID int64, sinceVersion int) ([]domain.LangString, error) {
	query := `SELECT id, pack_id, string_key, string_value, group_name, version, operation_type, prev_value, template_type, params_schema, preview_sample, created_at, updated_at 
	          FROM sys_lang_string WHERE pack_id = ? AND version > ? AND operation_type IN ('ADD', 'MOD')`

	rows, err := r.db.Query(query, packID, sinceVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.LangString
	for rows.Next() {
		var s domain.LangString
		var prevValue *string
		var previewSample *string
		var createdAtStr, updatedAtStr *string
		var groupName *string
		var paramsSchema *string

		err := rows.Scan(
			&s.ID,
			&s.PackID,
			&s.StringKey,
			&s.StringValue,
			&groupName,
			&s.Version,
			&s.OperationType,
			&prevValue,
			&s.TemplateType,
			&paramsSchema,
			&previewSample,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		if groupName != nil {
			s.GroupName = *groupName
		}
		if prevValue != nil {
			s.PrevValue = prevValue
		}
		if paramsSchema != nil {
			s.ParamsSchema = *paramsSchema
		}
		if previewSample != nil {
			s.PreviewSample = *previewSample
		}
		if createdAtStr != nil && *createdAtStr != "" {
			s.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr)
		}
		if updatedAtStr != nil && *updatedAtStr != "" {
			s.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr)
		}

		result = append(result, s)
	}

	return result, rows.Err()
}

// SaveString 新增或更新一条语言字符串
func (r *SQLiteRepo) SaveString(s *domain.LangString) error {
	if s.ID > 0 {
		query := `UPDATE sys_lang_string SET string_key=?, string_value=?, group_name=?,
		          template_type=?, params_schema=?, preview_sample=?, operation_type='MOD',
		          prev_value=string_value, version=version+1, updated_at=datetime('now')
		          WHERE id=?`
		_, err := r.db.Exec(query, s.StringKey, s.StringValue, s.GroupName,
			s.TemplateType, s.ParamsSchema, s.PreviewSample, s.ID)
		return err
	}
	query := `INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name,
	          template_type, params_schema, preview_sample, operation_type, version, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, 'ADD', 1, datetime('now'), datetime('now'))`
	result, err := r.db.Exec(query, s.PackID, s.StringKey, s.StringValue, s.GroupName,
		s.TemplateType, s.ParamsSchema, s.PreviewSample)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	s.ID = id
	return nil
}

// DeleteString 标记删除一条语言字符串（operation_type=DEL）
func (r *SQLiteRepo) DeleteString(id int64) error {
	query := `UPDATE sys_lang_string SET operation_type='DEL', updated_at=datetime('now') WHERE id=?`
	_, err := r.db.Exec(query, id)
	return err
}

// FindStringByKey 按 string_key 查询单条字符串
func (r *SQLiteRepo) FindStringByKey(packID int64, key domain.StringKey) (*domain.LangString, error) {
	query := `SELECT id, pack_id, string_key, string_value, group_name, version,
	          operation_type, prev_value, template_type, params_schema, preview_sample,
	          created_at, updated_at FROM sys_lang_string WHERE pack_id=? AND string_key=? AND operation_type!='DEL' LIMIT 1`
	row := r.db.QueryRow(query, packID, key)
	var s domain.LangString
	var prevValue *string
	var previewSample *string
	var createdAtStr, updatedAtStr *string
	var groupName *string
	var paramsSchema *string
	err := row.Scan(&s.ID, &s.PackID, &s.StringKey, &s.StringValue, &groupName,
		&s.Version, &s.OperationType, &prevValue, &s.TemplateType, &paramsSchema,
		&previewSample, &createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, err
	}
	if groupName != nil { s.GroupName = *groupName }
	if prevValue != nil { s.PrevValue = prevValue }
	if paramsSchema != nil { s.ParamsSchema = *paramsSchema }
	if previewSample != nil { s.PreviewSample = *previewSample }
	if createdAtStr != nil && *createdAtStr != "" { s.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr) }
	if updatedAtStr != nil && *updatedAtStr != "" { s.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr) }
	return &s, nil
}
