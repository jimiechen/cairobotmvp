package repository

import (
	"database/sql"
	"strings"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

const timeLayout = "2006-01-02 15:04:05"

// SQLiteRepo I18nRepository 的 SQLite 实现
// 用于开发和测试环境，生产环境应使用 MySQL
type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) GetPackByLangCode(langCode string, env string) (*domain.LangPack, error) {
	query := `SELECT id, pack_name, env, version, lang_code, description, is_published, published_at, published_by, created_at, updated_at 
	          FROM sys_lang_pack WHERE lang_code = ? AND env = ? AND is_published = 1`

	row := r.db.QueryRow(query, langCode, env)

	var pack domain.LangPack
	var publishedAtStr, createdAtStr, updatedAtStr *string
	var publishedBy *int64
	var description *string

	err := row.Scan(
		&pack.ID,
		&pack.PackName,
		&pack.Env,
		&pack.Version,
		&pack.LangCode,
		&description,
		&pack.IsPublished,
		&publishedAtStr,
		&publishedBy,
		&createdAtStr,
		&updatedAtStr,
	)

	if description != nil {
		pack.Description = *description
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if publishedAtStr != nil && *publishedAtStr != "" {
		if t, e := time.Parse(timeLayout, *publishedAtStr); e == nil {
			pack.PublishedAt = &t
		}
	}
	if publishedBy != nil {
		pack.PublishedBy = *publishedBy
	}
	if createdAtStr != nil && *createdAtStr != "" {
		pack.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr)
	}
	if updatedAtStr != nil && *updatedAtStr != "" {
		pack.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr)
	}

	return &pack, nil
}

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

func (r *SQLiteRepo) ListPacks(env string) ([]domain.LangPack, error) {
	query := `SELECT id, pack_name, env, version, lang_code, description, is_published, published_at, published_by, created_at, updated_at 
	          FROM sys_lang_pack WHERE env = ? AND is_published = 1 ORDER BY lang_code`

	rows, err := r.db.Query(query, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packs []domain.LangPack
	for rows.Next() {
		var pack domain.LangPack
		var publishedAtStr, createdAtStr, updatedAtStr *string
		var publishedBy *int64
		var description *string

		err := rows.Scan(
			&pack.ID,
			&pack.PackName,
			&pack.Env,
			&pack.Version,
			&pack.LangCode,
			&description,
			&pack.IsPublished,
			&publishedAtStr,
			&publishedBy,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		if description != nil {
			pack.Description = *description
		}
		if publishedAtStr != nil && *publishedAtStr != "" {
			if t, e := time.Parse(timeLayout, *publishedAtStr); e == nil {
				pack.PublishedAt = &t
			}
		}
		if publishedBy != nil {
			pack.PublishedBy = *publishedBy
		}
		if createdAtStr != nil && *createdAtStr != "" {
			pack.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr)
		}
		if updatedAtStr != nil && *updatedAtStr != "" {
			pack.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr)
		}

		packs = append(packs, pack)
	}

	return packs, rows.Err()
}

func (r *SQLiteRepo) GetLangPackByID(id int64) (*domain.LangPack, error) {
	query := `SELECT id, pack_name, env, version, lang_code, description, is_published, published_at, published_by, created_at, updated_at 
	          FROM sys_lang_pack WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var pack domain.LangPack
	var publishedAtStr, createdAtStr, updatedAtStr *string
	var publishedBy *int64
	var description *string

	err := row.Scan(
		&pack.ID,
		&pack.PackName,
		&pack.Env,
		&pack.Version,
		&pack.LangCode,
		&description,
		&pack.IsPublished,
		&publishedAtStr,
		&publishedBy,
		&createdAtStr,
		&updatedAtStr,
	)

	if description != nil {
		pack.Description = *description
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if publishedAtStr != nil && *publishedAtStr != "" {
		if t, e := time.Parse(timeLayout, *publishedAtStr); e == nil {
			pack.PublishedAt = &t
		}
	}
	if publishedBy != nil {
		pack.PublishedBy = *publishedBy
	}
	if createdAtStr != nil && *createdAtStr != "" {
		pack.CreatedAt, _ = time.Parse(timeLayout, *createdAtStr)
	}
	if updatedAtStr != nil && *updatedAtStr != "" {
		pack.UpdatedAt, _ = time.Parse(timeLayout, *updatedAtStr)
	}

	return &pack, nil
}

func parseTimeStr(s *string) time.Time {
	if s == nil || *s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.DateTime,
	} {
		if t, err := time.Parse(layout, *s); err == nil {
			return t
		}
	}
	if strings.TrimSpace(*s) != "" {
		if t, err := time.Parse(time.Layout, *s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func init() {
	time.Local = time.UTC
}
