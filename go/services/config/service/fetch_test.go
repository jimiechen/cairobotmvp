package service

import (
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/cache"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func setupFetchTest(t *testing.T) (*AppConfigService, repository.ConfigRepository, repository.SchemaRepository) {
	t.Helper()
	configRepo, _ := repository.NewSQLiteConfigRepo(":memory:")
	schemaRepo := repository.NewSQLiteSchemaRepo(configRepo.DB())
	mockCache := cache.NewMockCache()

	now := time.Now()

	testVersions := []versionInput{
		{"base_cfg", "dev", 1, `{"domain_root":"api.dev.com","sign_rand":"salt123"}`, true},
		{"custom_mod", "dev", 2, `{"feature_flag":true,"timeout":30}`, true},
		{"base_cfg", "dev", 2, `{"domain_root":"api.v2.dev.com"}`, true},
	}

	for _, v := range testVersions {
		cv := v.toDomain()
		cv.PublishedAt = &now
		configRepo.Save(cv)
	}

	svc := NewAppConfigService(configRepo, schemaRepo, mockCache)
	return svc, configRepo, schemaRepo
}

func TestGetAppConfigs_静态模块和动态模块分流(t *testing.T) {
	svc, _, _ := setupFetchTest(t)

	req := &AppConfigRequest{Env: "dev"}
	resp, err := svc.GetAppConfigs(req)
	if err != nil {
		t.Fatalf("GetAppConfigs 失败: %v", err)
	}

	if _, ok := resp.StaticModules["base_cfg"]; !ok {
		t.Error("base_cfg 应在 StaticModules 中")
	}
	if len(resp.DynamicModules) != 1 {
		t.Fatalf("期望 1 个动态模块, 实际 %d", len(resp.DynamicModules))
	}
	if resp.DynamicModules[0].ModuleKey != "custom_mod" {
		t.Errorf("动态模块 key 错误: %s", resp.DynamicModules[0].ModuleKey)
	}
}

func TestGetAppConfigs_RequestedModules过滤(t *testing.T) {
	svc, _, _ := setupFetchTest(t)

	req := &AppConfigRequest{
		Env:              "dev",
		RequestedModules: []string{"base_cfg"},
	}
	resp, err := svc.GetAppConfigs(req)
	if err != nil {
		t.Fatalf("GetAppConfigs 失败: %v", err)
	}

	if len(resp.DynamicModules) > 0 {
		t.Error("请求仅 base_cfg 时不应包含动态模块")
	}
}

func TestGetAppConfigs_空环境默认dev(t *testing.T) {
	svc, _, _ := setupFetchTest(t)

	req := &AppConfigRequest{Env: ""}
	resp, err := svc.GetAppConfigs(req)
	if err != nil {
		t.Fatalf("GetAppConfigs 失败: %v", err)
	}
	if resp == nil {
		t.Error("空 env 应默认使用 dev")
	}
}
