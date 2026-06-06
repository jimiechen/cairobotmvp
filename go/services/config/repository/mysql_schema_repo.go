package repository

import (
	"database/sql"
	"fmt"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// MySQLSchemaRepo SchemaRepository 的 MySQL 实现
// 操作 sys_config_schema 表，与 SQLiteSchemaRepo 接口一致
// 复用 utils.go 中的 scanFieldSchema / scanFieldSchemaRow / boolToInt
type MySQLSchemaRepo struct {
	db *sql.DB
}

// NewMySQLSchemaRepo 创建 MySQL Schema 仓库实例
// 前置条件：cfg 包含有效的 MySQL 连接信息
func NewMySQLSchemaRepo(cfg *config.MySQLConfig) (*MySQLSchemaRepo, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL 失败: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	return &MySQLSchemaRepo{db: db}, nil
}

// NewMySQLSchemaRepoWithDB 基于已有 *sql.DB 连接创建 Schema 仓库
// 用于共享连接场景（与 MySQLConfigRepo 共用同一连接）
func NewMySQLSchemaRepoWithDB(db *sql.DB) *MySQLSchemaRepo {
	return &MySQLSchemaRepo{db: db}
}

// ListByModule 查询指定模块下所有字段 Schema（含已禁用的）
func (r *MySQLSchemaRepo) ListByModule(moduleKey string) ([]*domain.FieldSchema, error) {
	query := `
	SELECT id, module_key, field_key, field_type, default_value, validator,
	       is_required, is_secret, description, client_scope, min_app_ver,
	       sort_order, is_enabled
	FROM sys_config_schema
	WHERE module_key = ?
	ORDER BY sort_order ASC
	`
	rows, err := r.db.Query(query, moduleKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []*domain.FieldSchema
	for rows.Next() {
		fs, err := scanFieldSchema(rows)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, fs)
	}
	return schemas, rows.Err()
}

// Create 新增一条字段 Schema 记录
// 违反 UNIQUE(module_key, field_key) 约束时由数据库层报错
func (r *MySQLSchemaRepo) Create(schema *domain.FieldSchema) error {
	query := `
	INSERT INTO sys_config_schema
	  (module_key, field_key, field_type, default_value, validator,
	   is_required, is_secret, description, client_scope, min_app_ver,
	   sort_order, is_enabled)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		schema.ModuleKey, schema.FieldKey, string(schema.FieldType),
		schema.DefaultValue, schema.Validator,
		boolToInt(schema.IsRequired), boolToInt(schema.IsSecret),
		schema.Description, schema.ClientScope, schema.MinAppVer,
		schema.SortOrder, boolToInt(schema.IsEnabled),
	)
	if err != nil {
		return fmt.Errorf("创建 field_schema 失败: %w", err)
	}
	id, _ := result.LastInsertId()
	schema.ID = id
	return nil
}

// Update 更新字段 Schema（按主键 ID）
func (r *MySQLSchemaRepo) Update(schema *domain.FieldSchema) error {
	query := `
	UPDATE sys_config_schema
	SET field_type = ?, default_value = ?, validator = ?,
	    is_required = ?, is_secret = ?, description = ?,
	    client_scope = ?, min_app_ver = ?, sort_order = ?,
	    is_enabled = ?, updated_at = NOW()
	WHERE id = ?
	`
	_, err := r.db.Exec(query,
		string(schema.FieldType), schema.DefaultValue, schema.Validator,
		boolToInt(schema.IsRequired), boolToInt(schema.IsSecret),
		schema.Description, schema.ClientScope, schema.MinAppVer,
		schema.SortOrder, boolToInt(schema.IsEnabled), schema.ID,
	)
	if err != nil {
		return fmt.Errorf("更新 field_schema 失败: %w", err)
	}
	return nil
}

// DeleteSoft 软删除字段 Schema（标记为禁用而非物理删除）
func (r *MySQLSchemaRepo) DeleteSoft(id int64) error {
	query := `UPDATE sys_config_schema SET is_enabled = 0, updated_at = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("软删除 field_schema 失败: %w", err)
	}
	return nil
}

// FindSchema 按主键 ID 查询单条字段 Schema
func (r *MySQLSchemaRepo) FindSchema(id int64) (*domain.FieldSchema, error) {
	query := `
	SELECT id, module_key, field_key, field_type, default_value, validator,
	       is_required, is_secret, description, client_scope, min_app_ver,
	       sort_order, is_enabled
	FROM sys_config_schema WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	fs, err := scanFieldSchemaRow(row)
	if err != nil {
		return nil, fmt.Errorf("查询 schema 失败: %w", err)
	}
	return fs, nil
}
