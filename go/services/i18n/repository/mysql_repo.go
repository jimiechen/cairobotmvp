package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// 确保 mysql 驱动已注册
// 实际驱动应在 main 包中导入: _ "github.com/go-sql-driver/mysql"
func init() {
	// 驱动注册由 main 包负责
}

// MySQLRepo I18nRepository 的 MySQL 实现
type MySQLRepo struct {
	db *sql.DB
}

// NewMySQLRepo 创建 MySQL 语言包仓库实例
func NewMySQLRepo(cfg *config.MySQLConfig) (*MySQLRepo, error) {
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

	return &MySQLRepo{db: db}, nil
}

// GetPackByLangCode 根据语言代码查询语言包
func (r *MySQLRepo) GetPackByLangCode(langCode string, env string) (*domain.LangPack, error) {
	query := `
		SELECT id, lang_code, lang_name, env, version, created_at, updated_at
		FROM sys_lang_pack
		WHERE lang_code = ? AND env = ?
		LIMIT 1
	`
	row := r.db.QueryRow(query, langCode, env)
	return scanLangPack(row)
}

// GetStringsByPackID 根据语言包 ID 查询所有字符串
func (r *MySQLRepo) GetStringsByPackID(packID int64) ([]domain.LangString, error) {
	query := `
		SELECT id, pack_id, string_key, string_value, param_schema, operation_type, version, create_by, update_by, created_at, updated_at
		FROM sys_lang_string
		WHERE pack_id = ? AND operation_type != 'DEL'
		ORDER BY string_key
	`
	rows, err := r.db.Query(query, packID)
	if err != nil {
		return nil, fmt.Errorf("查询语言字符串失败: %w", err)
	}
	defer rows.Close()

	var strings []domain.LangString
	for rows.Next() {
		s, err := scanLangString(rows)
		if err != nil {
			return nil, err
		}
		strings = append(strings, *s)
	}
	return strings, rows.Err()
}

// GetDiffSince 查询指定版本之后的增量变更
func (r *MySQLRepo) GetDiffSince(packID int64, sinceVersion int) ([]domain.LangString, error) {
	query := `
		SELECT id, pack_id, string_key, string_value, param_schema, operation_type, version, create_by, update_by, created_at, updated_at
		FROM sys_lang_string
		WHERE pack_id = ? AND version > ? AND operation_type != 'DEL'
		ORDER BY version ASC, string_key
	`
	rows, err := r.db.Query(query, packID, sinceVersion)
	if err != nil {
		return nil, fmt.Errorf("查询增量变更失败: %w", err)
	}
	defer rows.Close()

	var strings []domain.LangString
	for rows.Next() {
		s, err := scanLangString(rows)
		if err != nil {
			return nil, err
		}
		strings = append(strings, *s)
	}
	return strings, rows.Err()
}

// ListPacks 列出所有已发布的语言包
func (r *MySQLRepo) ListPacks(env string) ([]domain.LangPack, error) {
	query := `
		SELECT id, lang_code, lang_name, env, version, created_at, updated_at
		FROM sys_lang_pack
		WHERE env = ?
		ORDER BY lang_code
	`
	rows, err := r.db.Query(query, env)
	if err != nil {
		return nil, fmt.Errorf("查询语言包列表失败: %w", err)
	}
	defer rows.Close()

	var packs []domain.LangPack
	for rows.Next() {
		pack, err := scanLangPackRows(rows)
		if err != nil {
			return nil, err
		}
		packs = append(packs, *pack)
	}
	return packs, rows.Err()
}

// scanLangPack 从 sql.Row 扫描 LangPack
func scanLangPack(row *sql.Row) (*domain.LangPack, error) {
	var pack domain.LangPack
	var langName string
	err := row.Scan(
		&pack.ID,
		&pack.LangCode,
		&langName,
		&pack.Env,
		&pack.Version,
		&pack.CreatedAt,
		&pack.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("扫描语言包失败: %w", err)
	}
	pack.PackName = langName
	return &pack, nil
}

// scanLangPackRows 从 sql.Rows 扫描 LangPack
func scanLangPackRows(rows *sql.Rows) (*domain.LangPack, error) {
	var pack domain.LangPack
	var langName string
	err := rows.Scan(
		&pack.ID,
		&pack.LangCode,
		&langName,
		&pack.Env,
		&pack.Version,
		&pack.CreatedAt,
		&pack.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("扫描语言包失败: %w", err)
	}
	pack.PackName = langName
	return &pack, nil
}

// scanLangString 从 sql.Rows 扫描 LangString
func scanLangString(rows *sql.Rows) (*domain.LangString, error) {
	var s domain.LangString
	var paramSchema sql.NullString

	err := rows.Scan(
		&s.ID,
		&s.PackID,
		&s.StringKey,
		&s.StringValue,
		&paramSchema,
		&s.OperationType,
		&s.Version,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("扫描语言字符串失败: %w", err)
	}

	if paramSchema.Valid {
		s.ParamsSchema = paramSchema.String
	}

	return &s, nil
}
