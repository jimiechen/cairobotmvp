package admin

import (
	"context"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// CreateSchemaRequest 新增 Schema 请求 DTO
type CreateSchemaRequest struct {
	ModuleKey   string
	FieldKey    string
	FieldType   domain.FieldType
	DefaultValue string
	Validator   string
	IsRequired  bool
	IsSecret    bool
	Description string
	ClientScope string
	SortOrder   int
	Operator    string
}

// UpdateSchemaRequest 更新 Schema 请求 DTO
type UpdateSchemaRequest struct {
	ID           int64
	FieldType    domain.FieldType
	DefaultValue string
	Validator    string
	IsRequired   bool
	IsSecret     bool
	Description  string
	ClientScope  string
	SortOrder    int
	Operator     string
}

// SchemaItem Schema 列表项（返回给前端）
type SchemaItem struct {
	ID           int64
	ModuleKey    string
	FieldKey     string
	FieldType    string
	DefaultValue string
	Validator    string
	IsRequired   bool
	IsSecret     bool
	Description  string
	ClientScope  string
	SortOrder    int
	IsEnabled    bool
}

// CreateSchema 新增字段定义
// 流程：DTO 转换 → inner.CreateFieldSchema → 审计 → 失效+广播
func (s *AdminConfigService) CreateSchema(ctx context.Context, req CreateSchemaRequest) (*SchemaItem, error) {
	if req.ModuleKey == "" || req.FieldKey == "" {
		return nil, fmt.Errorf("module_key 和 field_key 不能为空")
	}
	schema := &domain.FieldSchema{
		ModuleKey:    req.ModuleKey,
		FieldKey:     req.FieldKey,
		FieldType:    req.FieldType,
		DefaultValue: req.DefaultValue,
		Validator:    req.Validator,
		IsRequired:   req.IsRequired,
		IsSecret:     req.IsSecret,
		Description:  req.Description,
		ClientScope:  req.ClientScope,
		SortOrder:    req.SortOrder,
		IsEnabled:    true,
	}
	if err := s.innerSchema.CreateFieldSchema(schema); err != nil {
		return nil, fmt.Errorf("创建 schema 失败: %w", err)
	}
	s.writeAudit(ctx, "create_schema", "schema", fmt.Sprintf("%d", schema.ID), req.Operator, map[string]string{
		"module_key": req.ModuleKey,
		"field_key":  req.FieldKey,
		"field_type": string(req.FieldType),
	})
	if err := s.invalidateAndBroadcast(ctx, []string{req.ModuleKey}); err != nil {
		return toSchemaItem(schema), nil
	}
	return toSchemaItem(schema), nil
}

// UpdateSchema 更新字段定义
func (s *AdminConfigService) UpdateSchema(ctx context.Context, req UpdateSchemaRequest) (*SchemaItem, error) {
	if req.ID <= 0 {
		return nil, fmt.Errorf("无效的 schema ID")
	}
	existing, err := s.schemaRepo.FindSchema(req.ID)
	if err != nil {
		return nil, fmt.Errorf("查询 schema 失败: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("schema 不存在: id=%d", req.ID)
	}
	existing.FieldType = req.FieldType
	existing.DefaultValue = req.DefaultValue
	existing.Validator = req.Validator
	existing.IsRequired = req.IsRequired
	existing.IsSecret = req.IsSecret
	existing.Description = req.Description
	existing.ClientScope = req.ClientScope
	existing.SortOrder = req.SortOrder
	if err := s.innerSchema.UpdateFieldSchema(existing); err != nil {
		return nil, fmt.Errorf("更新 schema 失败: %w", err)
	}
	s.writeAudit(ctx, "update_schema", "schema", fmt.Sprintf("%d", req.ID), req.Operator, map[string]string{
		"module_key": existing.ModuleKey,
		"field_key":  existing.FieldKey,
	})
	s.invalidateAndBroadcast(ctx, []string{existing.ModuleKey})
	return toSchemaItem(existing), nil
}

// DeleteSchema 软删除字段（标记禁用）
func (s *AdminConfigService) DeleteSchema(ctx context.Context, id int64, operator string) error {
	if id <= 0 {
		return fmt.Errorf("无效的 schema ID")
	}
	existing, err := s.schemaRepo.FindSchema(id)
	if err != nil {
		return fmt.Errorf("查询 schema 失败: %w", err)
	}
	if err := s.innerSchema.DeleteFieldSchema(id); err != nil {
		return fmt.Errorf("删除 schema 失败: %w", err)
	}
	s.writeAudit(ctx, "delete_schema", "schema", fmt.Sprintf("%d", id), operator, map[string]string{
		"module_key": existing.ModuleKey,
		"field_key":  existing.FieldKey,
	})
	return s.invalidateAndBroadcast(ctx, []string{existing.ModuleKey})
}

// ListSchemas 查询指定模块下所有字段 Schema
func (s *AdminConfigService) ListSchemas(ctx context.Context, moduleKey string) ([]*SchemaItem, error) {
	fields, err := s.innerSchema.ListFieldSchemas(moduleKey)
	if err != nil {
		return nil, fmt.Errorf("查询 schema 列表失败: %w", err)
	}
	items := make([]*SchemaItem, len(fields))
	for i, f := range fields {
		items[i] = toSchemaItem(f)
	}
	return items, nil
}

func toSchemaItem(f *domain.FieldSchema) *SchemaItem {
	if f == nil {
		return nil
	}
	return &SchemaItem{
		ID:           f.ID,
		ModuleKey:    f.ModuleKey,
		FieldKey:     f.FieldKey,
		FieldType:    string(f.FieldType),
		DefaultValue: f.DefaultValue,
		Validator:    f.Validator,
		IsRequired:   f.IsRequired,
		IsSecret:     f.IsSecret,
		Description:  f.Description,
		ClientScope:  f.ClientScope,
		SortOrder:    f.SortOrder,
		IsEnabled:    f.IsEnabled,
	}
}

func (s *AdminConfigService) writeAudit(ctx context.Context, action, targetType, targetID, operator string, detail map[string]string) {
	if s.auditWriter == nil {
		return
	}
	detailStr := ""
	if detail != nil {
		// 简单序列化，生产环境可用 JSON
		for k, v := range detail {
			detailStr += k + "=" + v + ";"
		}
	}
	_ = s.auditWriter.Write(ctx, AuditEntry{
		TenantID:   "default",
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Operator:   operator,
		Detail:     detailStr,
	})
}

// ToFieldType 将前端传入的字符串转换为 domain.FieldType
// 前端传 "string"/"int"/"bool" 等小写字符串
func ToFieldType(s string) domain.FieldType {
	return domain.FieldType(s)
}
