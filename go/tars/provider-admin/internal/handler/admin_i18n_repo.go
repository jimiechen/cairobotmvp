package handler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	i18nRepo "github.com/jimiechen/mineplanet/go/services/i18n/repository"
)

// AdminSQLiteRepo 管理后台专用的 SQLite 仓库实现
// 扩展基础 I18nRepository，增加写操作能力
type AdminSQLiteRepo struct {
	db *sql.DB
}

// NewAdminSQLiteRepo 创建管理后台专用的 SQLite 仓库实例
func NewAdminSQLiteRepo(db *sql.DB) *AdminSQLiteRepo {
	return &AdminSQLiteRepo{db: db}
}

// GetPackByLangCode 根据语言代码查询语言包
func (r *AdminSQLiteRepo) GetPackByLangCode(langCode string, env string) (*domain.LangPack, error) {
	if r.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}
	query := `SELECT id, pack_name, env, version, lang_code, description, is_published, published_at, published_by, created_at, updated_at 
	          FROM sys_lang_pack WHERE lang_code = ? AND env = ?`

	row := r.db.QueryRow(query, langCode, env)

	var pack domain.LangPack
	var publishedAt sql.NullTime

	err := row.Scan(
		&pack.ID,
		&pack.PackName,
		&pack.Env,
		&pack.Version,
		&pack.LangCode,
		&pack.Description,
		&pack.IsPublished,
		&publishedAt,
		&pack.PublishedBy,
		&pack.CreatedAt,
		&pack.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if publishedAt.Valid {
		pack.PublishedAt = &publishedAt.Time
	}

	return &pack, nil
}

// GetStringsByPackID 根据语言包 ID 查询所有字符串
func (r *AdminSQLiteRepo) GetStringsByPackID(packID int64) ([]domain.LangString, error) {
	if r.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}
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
		var prevValue sql.NullString

		err := rows.Scan(
			&s.ID,
			&s.PackID,
			&s.StringKey,
			&s.StringValue,
			&s.GroupName,
			&s.Version,
			&s.OperationType,
			&prevValue,
			&s.TemplateType,
			&s.ParamsSchema,
			&s.PreviewSample,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if prevValue.Valid {
			s.PrevValue = &prevValue.String
		}

		strings = append(strings, s)
	}

	return strings, rows.Err()
}

// GetDiffSince 查询指定版本之后的增量变更
func (r *AdminSQLiteRepo) GetDiffSince(packID int64, sinceVersion int) ([]domain.LangString, error) {
	if r.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}
	query := `SELECT id, pack_id, string_key, string_value, group_name, version, operation_type, prev_value, template_type, params_schema, preview_sample, created_at, updated_at 
	          FROM sys_lang_string WHERE pack_id = ? AND version > ? AND operation_type IN ('ADD', 'MOD')`

	rows, err := r.db.Query(query, packID, sinceVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strings []domain.LangString
	for rows.Next() {
		var s domain.LangString
		var prevValue sql.NullString

		err := rows.Scan(
			&s.ID,
			&s.PackID,
			&s.StringKey,
			&s.StringValue,
			&s.GroupName,
			&s.Version,
			&s.OperationType,
			&prevValue,
			&s.TemplateType,
			&s.ParamsSchema,
			&s.PreviewSample,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if prevValue.Valid {
			s.PrevValue = &prevValue.String
		}

		strings = append(strings, s)
	}

	return strings, rows.Err()
}

// ListPacks 列出所有已发布的语言包
func (r *AdminSQLiteRepo) ListPacks(env string) ([]domain.LangPack, error) {
	if r.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}
	query := `SELECT id, pack_name, env, version, lang_code, description, is_published, published_at, published_by, created_at, updated_at 
	          FROM sys_lang_pack WHERE env = ? ORDER BY lang_code`

	rows, err := r.db.Query(query, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packs []domain.LangPack
	for rows.Next() {
		var pack domain.LangPack
		var publishedAt sql.NullTime

		err := rows.Scan(
			&pack.ID,
			&pack.PackName,
			&pack.Env,
			&pack.Version,
			&pack.LangCode,
			&pack.Description,
			&pack.IsPublished,
			&publishedAt,
			&pack.PublishedBy,
			&pack.CreatedAt,
			&pack.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			pack.PublishedAt = &publishedAt.Time
		}

		packs = append(packs, pack)
	}

	return packs, rows.Err()
}

// CreatePack 创建或更新语言包
func (r *AdminSQLiteRepo) CreatePack(pack *domain.LangPack) error {
	if r.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	query := `
	INSERT INTO sys_lang_pack (pack_name, env, version, lang_code, description, is_published)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(lang_code, env) DO UPDATE SET
		pack_name = excluded.pack_name,
		description = excluded.description,
		updated_at = datetime('now')
	`
	result, err := r.db.Exec(query,
		pack.PackName,
		pack.Env,
		pack.Version,
		pack.LangCode,
		pack.Description,
		boolToInt(pack.IsPublished),
	)
	if err != nil {
		return fmt.Errorf("创建语言包失败: %w", err)
	}
	id, _ := result.LastInsertId()
	pack.ID = id
	return nil
}

// CreateString 新增多语言字符串
func (r *AdminSQLiteRepo) CreateString(s *domain.LangString) error {
	if r.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	query := `
	INSERT INTO sys_lang_string (pack_id, string_key, string_value, group_name, template_type, params_schema, operation_type)
	VALUES (?, ?, ?, ?, ?, ?, 'ADD')
	`
	result, err := r.db.Exec(query,
		s.PackID,
		s.StringKey,
		s.StringValue,
		s.GroupName,
		string(s.TemplateType),
		s.ParamsSchema,
	)
	if err != nil {
		return fmt.Errorf("创建字符串失败: %w", err)
	}
	id, _ := result.LastInsertId()
	s.ID = id
	return nil
}

// UpdateString 更新多语言字符串
func (r *AdminSQLiteRepo) UpdateString(s *domain.LangString) error {
	if r.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	query := `UPDATE sys_lang_string 
	SET string_value = ?, group_name = ?, template_type = ?, params_schema = ?, 
	    operation_type = 'MOD', prev_value = string_value, updated_at = datetime('now') 
	WHERE id = ?`
	_, err := r.db.Exec(query,
		s.StringValue,
		s.GroupName,
		string(s.TemplateType),
		s.ParamsSchema,
		s.ID,
	)
	if err != nil {
		return fmt.Errorf("更新字符串失败: %w", err)
	}
	return nil
}

// DeleteString 删除多语言字符串（软删除）
func (r *AdminSQLiteRepo) DeleteString(id int64) error {
	if r.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	query := `UPDATE sys_lang_string 
	SET operation_type = 'DEL', updated_at = datetime('now') 
	WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除字符串失败: %w", err)
	}
	return nil
}

// PublishPack 发布语言包版本
func (r *AdminSQLiteRepo) PublishPack(packID int64, publishedBy int64) error {
	if r.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	now := time.Now()
	query := `UPDATE sys_lang_pack 
	SET is_published = 1, published_at = ?, published_by = ?, version = version + 1, updated_at = datetime('now') 
	WHERE id = ?`
	_, err := r.db.Exec(query, now, publishedBy, packID)
	if err != nil {
		return fmt.Errorf("发布语言包失败: %w", err)
	}

	versionQuery := `UPDATE sys_lang_string SET version = (SELECT version FROM sys_lang_pack WHERE id = ?) WHERE pack_id = ?`
	_, err = r.db.Exec(versionQuery, packID, packID)
	if err != nil {
		return fmt.Errorf("更新字符串版本失败: %w", err)
	}
	return nil
}

var _ i18nRepo.I18nRepository = (*AdminSQLiteRepo)(nil)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
