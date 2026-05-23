package service

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func newTestSchemaService(t *testing.T) (*SchemaService, repository.ConfigRepository) {
	t.Helper()
	configRepo, _ := repository.NewSQLiteConfigRepo(":memory:")
	schemaRepo := repository.NewSQLiteSchemaRepo(configRepo.DB())
	svc := NewSchemaService(schemaRepo)
	return svc, configRepo
}

func TestSchemaService_CreateAndGet(t *testing.T) {
	svc, _ := newTestSchemaService(t)

	schema := &domain.FieldSchema{
		ModuleKey: "test_mod", FieldKey: "new_field",
		FieldType: domain.FieldTypeString, Description: "测试字段",
		IsEnabled: true,
	}
	err := svc.CreateFieldSchema(schema)
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	list, err := svc.ListFieldSchemas("test_mod")
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(list) != 1 || list[0].FieldKey != "new_field" {
		t.Error("Create 后 List 未返回新建记录")
	}
}

func TestSchemaService_Update(t *testing.T) {
	svc, _ := newTestSchemaService(t)

	schema := &domain.FieldSchema{
		ModuleKey: "mod", FieldKey: "f1",
		FieldType: domain.FieldTypeString, Description: "原描述",
		IsEnabled: true,
	}
	svc.CreateFieldSchema(schema)

	schema.Description = "更新后的描述"
	err := svc.UpdateFieldSchema(schema)
	if err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	list, _ := svc.ListFieldSchemas("mod")
	if list[0].Description != "更新后的描述" {
		t.Errorf("Update 未生效: %s", list[0].Description)
	}
}

func TestSchemaService_DeleteSoft(t *testing.T) {
	svc, _ := newTestSchemaService(t)

	schema := &domain.FieldSchema{
		ModuleKey: "mod", FieldKey: "to_del",
		FieldType: domain.FieldTypeString, IsEnabled: true,
	}
	svc.CreateFieldSchema(schema)

	err := svc.DeleteFieldSchema(schema.ID)
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	modSchema, _ := svc.GetModuleSchema("mod")
	enabled := modSchema.EnabledFields()
	if len(enabled) != 0 {
		t.Error("软删除后 EnabledFields 应为空")
	}
}

func TestSchemaService_非法输入校验(t *testing.T) {
	svc, _ := newTestSchemaService(t)

	err := svc.CreateFieldSchema(&domain.FieldSchema{})
	if err != ErrInvalidSchemaInput {
		t.Error("空 schema 应返回 ErrInvalidSchemaInput")
	}

	err = svc.UpdateFieldSchema(&domain.FieldSchema{ID: -1})
	if err != ErrInvalidSchemaInput {
		t.Error("负 ID 应返回 ErrInvalidSchemaInput")
	}

	err = svc.DeleteFieldSchema(0)
	if err != ErrInvalidSchemaInput {
		t.Error("ID=0 应返回 ErrInvalidSchemaInput")
	}
}

func TestSchemaService_GetModuleSchema(t *testing.T) {
	svc, _ := newTestSchemaService(t)

	svc.CreateFieldSchema(&domain.FieldSchema{
		ModuleKey: "info_mod", FieldKey: "title",
		FieldType: domain.FieldTypeString, IsEnabled: true,
	})
	svc.CreateFieldSchema(&domain.FieldSchema{
		ModuleKey: "info_mod", FieldKey: "subtitle",
		FieldType: domain.FieldTypeString, IsEnabled: false,
	})

	ms, err := svc.GetModuleSchema("info_mod")
	if err != nil {
		t.Fatalf("GetModuleSchema 失败: %v", err)
	}
	if ms.ModuleKey != "info_mod" {
		t.Errorf("ModuleKey 错误: %s", ms.ModuleKey)
	}
	if len(ms.Fields) != 2 {
		t.Errorf("Fields 数量错误: %d", len(ms.Fields))
	}
	if len(ms.EnabledFields()) != 1 {
		t.Errorf("EnabledFields 数量错误: %d", len(ms.EnabledFields()))
	}
}
