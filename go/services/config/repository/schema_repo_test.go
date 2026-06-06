package repository

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func newTestSchemaRepo(t *testing.T) *SQLiteSchemaRepo {
	t.Helper()
	configRepo, err := NewSQLiteConfigRepo(":memory:")
	if err != nil {
		t.Fatalf("创建测试仓库失败: %v", err)
	}
	t.Cleanup(func() { configRepo.Close() })
	return NewSQLiteSchemaRepo(configRepo.db)
}

func TestSQLiteSchemaRepo_Create与ListByModule(t *testing.T) {
	repo := newTestSchemaRepo(t)

	schema := &domain.FieldSchema{
		ModuleKey: "base_cfg", FieldKey: "domain_root",
		FieldType: domain.FieldTypeString, DefaultValue: "",
		Description: "API 根域名", SortOrder: 1, IsEnabled: true,
	}

	err := repo.Create(schema)
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if schema.ID == 0 {
		t.Error("Create 后应设置自增 ID")
	}

	list, err := repo.ListByModule("base_cfg")
	if err != nil {
		t.Fatalf("ListByModule 失败: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("期望 1 条记录, 实际 %d", len(list))
	}
	if list[0].FieldKey != "domain_root" {
		t.Errorf("FieldKey 不匹配: %s", list[0].FieldKey)
	}
}

func TestSQLiteSchemaRepo_Update(t *testing.T) {
	repo := newTestSchemaRepo(t)

	schema := &domain.FieldSchema{
		ModuleKey: "base_cfg", FieldKey: "f1",
		FieldType: domain.FieldTypeString, Description: "原始描述",
		IsEnabled: true,
	}
	repo.Create(schema)

	schema.Description = "更新后的描述"
	err := repo.Update(schema)
	if err != nil {
		t.Fatalf("Update 失败: %v", err)
	}

	list, _ := repo.ListByModule("base_cfg")
	if list[0].Description != "更新后的描述" {
		t.Errorf("Update 未生效: %s", list[0].Description)
	}
}

func TestSQLiteSchemaRepo_DeleteSoft(t *testing.T) {
	repo := newTestSchemaRepo(t)

	schema := &domain.FieldSchema{
		ModuleKey: "base_cfg", FieldKey: "to_delete",
		FieldType: domain.FieldTypeString, IsEnabled: true,
	}
	repo.Create(schema)

	err := repo.DeleteSoft(schema.ID)
	if err != nil {
		t.Fatalf("DeleteSoft 失败: %v", err)
	}

	list, _ := repo.ListByModule("base_cfg")
	if list[0].IsEnabled {
		t.Error("软删除后 IsEnabled 应为 false")
	}
}

func TestSQLiteSchemaRepo_ListByModule_空模块应返回空列表(t *testing.T) {
	repo := newTestSchemaRepo(t)
	list, err := repo.ListByModule("nonexistent")
	if err != nil {
		t.Fatalf("ListByModule 失败: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("空模块应返回空列表, 实际 %d", len(list))
	}
}

// TestMySQLSchemaRepo_InterfaceVerify 编译期接口满足性验证
// 如果 MySQLSchemaRepo 未实现 SchemaRepository 全部方法，编译将失败
func TestMySQLSchemaRepo_InterfaceVerify(t *testing.T) {
	var _ SchemaRepository = (*MySQLSchemaRepo)(nil)
}
