package testutil

import (
	"sync"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// MockSchemaRepo 统一的 SchemaRepository Mock 实现
// 用于测试中模拟 Schema 仓库行为，支持灵活的数据注入和查询
type MockSchemaRepo struct {
	mu      sync.Mutex
	schemas map[string][]*domain.FieldSchema
}

// NewMockSchemaRepo 创建 Mock Schema 仓库实例
func NewMockSchemaRepo() *MockSchemaRepo {
	return &MockSchemaRepo{
		schemas: make(map[string][]*domain.FieldSchema),
	}
}

// ListByModule 查询指定模块下所有字段 Schema
func (m *MockSchemaRepo) ListByModule(moduleKey string) ([]*domain.FieldSchema, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if schemas, ok := m.schemas[moduleKey]; ok {
		result := make([]*domain.FieldSchema, len(schemas))
		copy(result, schemas)
		return result, nil
	}
	return []*domain.FieldSchema{}, nil
}

// FindSchema 按主键 ID 查询单条字段 Schema
func (m *MockSchemaRepo) FindSchema(id int64) (*domain.FieldSchema, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, schemas := range m.schemas {
		for _, schema := range schemas {
			if schema.ID == id {
				cp := *schema
				return &cp, nil
			}
		}
	}
	return nil, nil
}

// Create 新增一条字段 Schema 记录
func (m *MockSchemaRepo) Create(schema *domain.FieldSchema) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := schema.ModuleKey
	m.schemas[key] = append(m.schemas[key], schema)
	return nil
}

// Update 更新字段 Schema（按主键 ID）
func (m *MockSchemaRepo) Update(schema *domain.FieldSchema) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, schemas := range m.schemas {
		for i, s := range schemas {
			if s.ID == schema.ID {
				schemas[i] = schema
				return nil
			}
		}
	}
	return nil
}

// DeleteSoft 软删除字段 Schema（标记为禁用）
func (m *MockSchemaRepo) DeleteSoft(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, schemas := range m.schemas {
		for _, s := range schemas {
			if s.ID == id {
				s.IsEnabled = false
				return nil
			}
		}
	}
	return nil
}

// AddSchema 为指定模块添加 Schema 定义（用于测试准备）
func (m *MockSchemaRepo) AddSchema(moduleKey string, schemas ...*domain.FieldSchema) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schemas[moduleKey] = append(m.schemas[moduleKey], schemas...)
}

// SetSchemas 设置指定模块的所有 Schema（用于测试准备）
func (m *MockSchemaRepo) SetSchemas(moduleKey string, schemas []*domain.FieldSchema) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schemas[moduleKey] = schemas
}

// GetSchemas 获取指定模块的所有 Schema（用于测试断言）
func (m *MockSchemaRepo) GetSchemas(moduleKey string) []*domain.FieldSchema {
	m.mu.Lock()
	defer m.mu.Unlock()
	if schemas, ok := m.schemas[moduleKey]; ok {
		result := make([]*domain.FieldSchema, len(schemas))
		copy(result, schemas)
		return result
	}
	return nil
}

// GetAllSchemas 获取所有模块的 Schema（用于测试断言）
func (m *MockSchemaRepo) GetAllSchemas() map[string][]*domain.FieldSchema {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string][]*domain.FieldSchema)
	for k, v := range m.schemas {
		copies := make([]*domain.FieldSchema, len(v))
		copy(copies, v)
		result[k] = copies
	}
	return result
}

// Clear 清空所有数据（用于测试重置）
func (m *MockSchemaRepo) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schemas = make(map[string][]*domain.FieldSchema)
}
