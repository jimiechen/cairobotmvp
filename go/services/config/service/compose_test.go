package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func setupComposeTest(t *testing.T) (repository.ConfigRepository, repository.SchemaRepository) {
	t.Helper()
	configRepo, _ := repository.NewSQLiteConfigRepo(":memory:")
	schemaRepo := repository.NewSQLiteSchemaRepo(configRepo.DB())
	now := time.Now()

	configRepo.Save(&domain.ConfigVersion{
		ModuleKey: "base_cfg", Env: "prod", Version: 5,
		ConfigJSON: `{"domain_root":"api.prod.com","sign_rand":"s1"}`,
		IsPublished: true, PublishedAt: &now, CreateBy: "test",
	})
	configRepo.Save(&domain.ConfigVersion{
		ModuleKey: "notification_cfg", Env: "prod", Version: 2,
		ConfigJSON: `{"push_enabled":true,"interval":300}`,
		IsPublished: true, PublishedAt: &now, CreateBy: "test",
	})
	schemaRepo.Create(&domain.FieldSchema{
		ModuleKey: "notification_cfg", FieldKey: "push_enabled",
		FieldType: domain.FieldTypeBool, IsEnabled: true, ClientScope: "all",
	})
	schemaRepo.Create(&domain.FieldSchema{
		ModuleKey: "notification_cfg", FieldKey: "interval",
		FieldType: domain.FieldTypeInt, IsEnabled: true, ClientScope: "all",
	})

	return configRepo, schemaRepo
}

func TestBuildDynamicModule_非静态模块应进入动态列表(t *testing.T) {
	configRepo, schemaRepo := setupComposeTest(t)
	versions, _ := configRepo.ListPublishedVersions("prod")

	dynamicVers := []*domain.ConfigVersion{}
	for _, v := range versions {
		if !domain.IsStaticModule(v.ModuleKey) {
			dynamicVers = append(dynamicVers, v)
		}
	}
	if len(dynamicVers) != 1 {
		t.Fatalf("期望 1 个动态版本, 实际 %d", len(dynamicVers))
	}

	typedMap, _ := ParseConfigJSON(dynamicVers[0].ConfigJSON, dynamicVers[0].ModuleKey, schemaRepo)
	dm := BuildDynamicModule(dynamicVers[0], typedMap, schemaRepo, "all")

	if dm.ModuleKey != "notification_cfg" {
		t.Errorf("module_key 错误: %s", dm.ModuleKey)
	}
	if dm.Version != 2 {
		t.Errorf("version 错误: %d", dm.Version)
	}
	if len(dm.Descriptors) != 2 {
		t.Errorf("descriptors 数量错误: %d, 期望 2", len(dm.Descriptors))
	}
}

func TestClassifyModules_分流静态与动态(t *testing.T) {
	versions := []*domain.ConfigVersion{
		{ModuleKey: "base_cfg"},
		{ModuleKey: "custom_new"},
		{ModuleKey: "wap_cfg"},
		{ModuleKey: "another_dynamic"},
	}
	static, dynamic := ClassifyModules(versions)
	if len(static) != 2 {
		t.Errorf("静态模块数错误: %d", len(static))
	}
	if len(dynamic) != 2 {
		t.Errorf("动态模块数错误: %d", len(dynamic))
	}
}

func TestToJSONMap_序列化TypedValue(t *testing.T) {
	fields := map[string]*domain.TypedValue{
		"a": domain.NewTypedValue(domain.FieldTypeString, "hello"),
		"b": domain.NewTypedValue(domain.FieldTypeInt, int64(42)),
	}
	result := ToJSONMap(fields)
	if result[`a`] != `"hello"` {
		t.Errorf("a 序列化错误: %s", result[`a`])
	}
	var b int
	json.Unmarshal([]byte(result["b"]), &b)
	if b != 42 {
		t.Error("b 反序列化后值错误")
	}
}

func TestComposeFullResponse_完整组装(t *testing.T) {
	configRepo, schemaRepo := setupComposeTest(t)
	versions, _ := configRepo.ListPublishedVersions("prod")

	resp := ComposeFullResponse("prod", "all", versions, schemaRepo, nil)

	if _, ok := resp.StaticModules["base_cfg"]; !ok {
		t.Error("base_cfg 应在 StaticModules 中")
	}
	if len(resp.DynamicModules) != 1 {
		t.Fatalf("期望 1 个动态模块, 实际 %d", len(resp.DynamicModules))
	}
}

func TestComposeFullResponse_RequestedModules过滤(t *testing.T) {
	configRepo, schemaRepo := setupComposeTest(t)
	versions, _ := configRepo.ListPublishedVersions("prod")

	resp := ComposeFullResponse("prod", "all", versions, schemaRepo, []string{"base_cfg"})
	if len(resp.DynamicModules) > 0 {
		t.Error("过滤后不应包含动态模块")
	}
}
