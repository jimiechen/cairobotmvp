package service

import (
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

// SchemaService 配置 Schema 管理（运营后台接口）
// 负责 sys_config_schema 的 CRUD 操作，供管理后台调用
type SchemaService struct {
	schemaRepo repository.SchemaRepository
}

// NewSchemaService 创建 Schema 管理服务实例
func NewSchemaService(schemaRepo repository.SchemaRepository) *SchemaService {
	return &SchemaService{schemaRepo: schemaRepo}
}

// ListFieldSchemas 查询指定模块下所有字段 Schema
func (s *SchemaService) ListFieldSchemas(moduleKey string) ([]*domain.FieldSchema, error) {
	return s.schemaRepo.ListByModule(moduleKey)
}

// CreateFieldSchema 新增字段定义
func (s *SchemaService) CreateFieldSchema(schema *domain.FieldSchema) error {
	if schema.ModuleKey == "" || schema.FieldKey == "" {
		return ErrInvalidSchemaInput
	}
	return s.schemaRepo.Create(schema)
}

// UpdateFieldSchema 更新字段定义
func (s *SchemaService) UpdateFieldSchema(schema *domain.FieldSchema) error {
	if schema.ID <= 0 {
		return ErrInvalidSchemaInput
	}
	return s.schemaRepo.Update(schema)
}

// DeleteFieldSchema 软删除字段（标记禁用）
func (s *SchemaService) DeleteFieldSchema(id int64) error {
	if id <= 0 {
		return ErrInvalidSchemaInput
	}
	return s.schemaRepo.DeleteSoft(id)
}

// GetModuleSchema 聚合某模块的完整 Schema 视图
// 返回 ModuleSchema（含 FindField / EnabledFields 等便捷方法）
func (s *SchemaService) GetModuleSchema(moduleKey string) (*domain.ModuleSchema, error) {
	fields, err := s.schemaRepo.ListByModule(moduleKey)
	if err != nil {
		return nil, err
	}
	return &domain.ModuleSchema{ModuleKey: moduleKey, Fields: fields}, nil
}

// ErrInvalidSchemaInput Schema 输入参数非法
var ErrInvalidSchemaInput = InvalidInputError("schema input validation failed")

type InvalidInputError string

func (e InvalidInputError) Error() string { return string(e) }
