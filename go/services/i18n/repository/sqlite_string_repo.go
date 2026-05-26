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
