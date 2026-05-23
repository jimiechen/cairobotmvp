package repository

import (
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func newTestRepo(t *testing.T) *SQLiteConfigRepo {
	t.Helper()
	repo, err := NewSQLiteConfigRepo(":memory:")
	if err != nil {
		t.Fatalf("创建测试仓库失败: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return repo
}

func TestSQLiteConfigRepo_SaveAndGetLatestVersion(t *testing.T) {
	repo := newTestRepo(t)
	now := time.Now()

	input := &domain.ConfigVersion{
		ModuleKey:   "base_cfg",
		Env:         "prod",
		Version:     3,
		ConfigJSON:  `{"domain_root":"api.example.com"}`,
		IsPublished: true,
		PublishedAt: &now,
		CreateBy:    "test",
	}

	err := repo.Save(input)
	if err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	if input.ID == 0 {
		t.Error("Save 后应设置自增 ID")
	}

	got, err := repo.GetLatestVersion("base_cfg", "prod")
	if err != nil {
		t.Fatalf("GetLatestVersion 失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回已发布的版本")
	}
	if got.ModuleKey != "base_cfg" || got.Version != 3 {
		t.Errorf("数据不匹配: module=%s version=%d", got.ModuleKey, got.Version)
	}
	if !got.IsReleased() {
		t.Error("已发布版本 IsReleased 应返回 true")
	}
}

func TestSQLiteConfigRepo_GetLatestVersion_无数据返回nil(t *testing.T) {
	repo := newTestRepo(t)
	got, err := repo.GetLatestVersion("nonexistent", "dev")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if got != nil {
		t.Error("无数据时应返回 nil")
	}
}

func TestSQLiteConfigRepo_GetByModuleAndVersion(t *testing.T) {
	repo := newTestRepo(t)
	now := time.Now()

	err := repo.Save(&domain.ConfigVersion{
		ModuleKey: "wap_cfg", Env: "dev", Version: 5,
		ConfigJSON: `{"url":"http://example.com"}`,
		IsPublished: true, PublishedAt: &now, CreateBy: "test",
	})
	if err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	got, err := repo.GetByModuleAndVersion("wap_cfg", "dev", 5)
	if err != nil {
		t.Fatalf("GetByModuleAndVersion 失败: %v", err)
	}
	if got == nil || got.Version != 5 {
		t.Error("未找到预期版本")
	}
}

func TestSQLiteConfigRepo_ListPublishedVersions(t *testing.T) {
	repo := newTestRepo(t)
	now := time.Now()

	for _, mod := range []string{"base_cfg", "wap_cfg"} {
		repo.Save(&domain.ConfigVersion{
			ModuleKey: mod, Env: "dev", Version: 1,
			ConfigJSON: `{}`, IsPublished: true,
			PublishedAt: &now, CreateBy: "test",
		})
	}

	versions, err := repo.ListPublishedVersions("dev")
	if err != nil {
		t.Fatalf("ListPublishedVersions 失败: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("期望 2 条记录, 实际 %d", len(versions))
	}
}

func TestSQLiteConfigRepo_未发布版本不应被Latest查到(t *testing.T) {
	repo := newTestRepo(t)

	repo.Save(&domain.ConfigVersion{
		ModuleKey: "mute_cfg", Env: "dev", Version: 1,
		ConfigJSON: `[]`, IsPublished: false, CreateBy: "test",
	})

	got, _ := repo.GetLatestVersion("mute_cfg", "dev")
	if got != nil {
		t.Error("未发布版本不应被 GetLatestVersion 返回")
	}
}
