package repository

import (
	"database/sql"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// SchemaRepository 字段元数据（sys_config_schema）的数据访问接口
// 负责配置字段的 CRUD 操作，供运营后台管理 schema 使用
// 与 ConfigRepository 分离：前者管版本数据，后者管字段定义
type SchemaRepository interface {
	ListByModule(moduleKey string) ([]*domain.FieldSchema, error)
	Create(schema *domain.FieldSchema) error
	Update(schema *domain.FieldSchema) error
	DeleteSoft(id int64) error
}

// SQLiteSchemaRepo SchemaRepository 的 SQLite 实现
// 复用 SQLiteConfigRepo 的 db 连接，操作 sys_config_schema 表
type SQLiteSchemaRepo struct {
	db *sql.DB
}

// NewSQLiteSchemaRepo 基于已有 DB 连接创建 Schema 仓库
// 要求传入的 db 已执行过 initTables()（包含 sys_config_schema 建）
func NewSQLiteSchemaRepo(db *sql.DB) *SQLiteSchemaRepo {
	return &SQLiteSchemaRepo{db: db}
}

// ListByModule 查询指定模块下所有字段 Schema（含已禁用的）
// 调用方根据需要自行过滤 IsEnabled
func (r *SQLiteSchemaRepo) ListByModule(moduleKey string) ([]*domain.FieldSchema, error) {
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
func (r *SQLiteSchemaRepo) Create(schema *domain.FieldSchema) error {
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
// 仅更新可变字段，不修改 module_key 和 field_key
func (r *SQLiteSchemaRepo) Update(schema *domain.FieldSchema) error {
	query := `
	UPDATE sys_config_schema
	SET field_type = ?, default_value = ?, validator = ?,
	    is_required = ?, is_secret = ?, description = ?,
	    client_scope = ?, min_app_ver = ?, sort_order = ?,
	    is_enabled = ?, updated_at = datetime('now')
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
// 保留历史引用完整性，已下发的客户端不受影响
func (r *SQLiteSchemaRepo) DeleteSoft(id int64) error {
	query := `UPDATE sys_config_schema SET is_enabled = 0, updated_at = datetime('now') WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("软删除 field_schema 失败: %w", err)
	}
	return nil
}
