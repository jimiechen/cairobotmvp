package repository

import (
	"database/sql"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

const timeLayout = "2006-01-02 15:04:05"

// GetPackByLangCode 根据语言代码和环境查询已发布的语言包
// 返回 nil, nil 表示未找到
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

// ListPacks 查询指定环境下所有已发布语言包，按 lang_code 排序
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

// GetLangPackByID 根据 ID 查询语言包（不限制 is_published）
// 返回 nil, nil 表示未找到
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
