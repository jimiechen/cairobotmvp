package service

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/cache"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func TestNewAppConfigService_应正确组合依赖(t *testing.T) {
	mockCache := cache.NewMockCache()
	svc := NewAppConfigService(nil, nil, mockCache)
	if svc == nil {
		t.Fatal("NewAppConfigService 不应返回 nil")
	}
	if svc.cache == nil {
		t.Error("cache 应被注入")
	}
}

func TestContains_存在应返回true(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("contains 应找到 b")
	}
}

func TestContains_不存在应返回false(t *testing.T) {
	if contains([]string{"a", "b"}, "x") {
		t.Error("contains 不应找到 x")
	}
}

func TestContains_空列表(t *testing.T) {
	if contains([]string{}, "a") {
		t.Error("空列表 contains 应返回 false")
	}
}

func mockRepoWithVersions(t *testing.T, versions []*versionInput) repository.ConfigRepository {
	t.Helper()
	repo, err := repository.NewSQLiteConfigRepo(":memory:")
	if err != nil {
		t.Fatalf("创建仓库失败: %v", err)
	}
	for _, v := range versions {
		repo.Save(v.toDomain())
	}
	return repo
}

type versionInput struct {
	moduleKey string
	env       string
	version   int64
	configJSON string
	published bool
}

func (vi *versionInput) toDomain() *domain.ConfigVersion {
	return &domain.ConfigVersion{
		ModuleKey:  vi.moduleKey,
		Env:        vi.env,
		Version:    vi.version,
		ConfigJSON: vi.configJSON,
		IsPublished: vi.published,
		CreateBy:   "test",
	}
}
