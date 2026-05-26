package repository

import (
	"sync"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// MemConfigRepo 基于内存 Map 的 ConfigRepository 实现
// 零外部依赖，用于单测和开发阶段
// 不持久化，进程重启后数据丢失；生产环境应替换为 MySQL 实现
type MemConfigRepo struct {
	mu       sync.RWMutex
	versions []*domain.ConfigVersion
}

// NewMemConfigRepo 创建空内存配置仓库实例
func NewMemConfigRepo() *MemConfigRepo {
	return &MemConfigRepo{versions: make([]*domain.ConfigVersion, 0)}
}

// GetLatestVersion 查询指定模块在指定环境下最新已发布的版本
func (r *MemConfigRepo) GetLatestVersion(moduleKey, env string) (*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.ConfigVersion
	for _, v := range r.versions {
		if v.ModuleKey == moduleKey && v.Env == env && v.IsPublished {
			if latest == nil || v.Version > latest.Version {
				latest = v
			}
		}
	}
	return latest, nil
}

// GetByModuleAndVersion 精确查询某模块在某环境下的特定版本
func (r *MemConfigRepo) GetByModuleAndVersion(moduleKey, env string, version int64) (*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, v := range r.versions {
		if v.ModuleKey == moduleKey && v.Env == env && v.Version == version {
			return v, nil
		}
	}
	return nil, nil
}

// ListPublishedVersions 列出指定环境下所有已发布的配置版本
func (r *MemConfigRepo) ListPublishedVersions(env string) ([]*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.ConfigVersion
	for _, v := range r.versions {
		if v.Env == env && v.IsPublished {
			result = append(result, v)
		}
	}
	return result, nil
}

// Save 新增配置版本记录到内存
func (r *MemConfigRepo) Save(version *domain.ConfigVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	version.ID = int64(len(r.versions)) + 1
	r.versions = append(r.versions, version)
	return nil
}

// Clear 清空所有数据（仅测试用）
func (r *MemConfigRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.versions = make([]*domain.ConfigVersion, 0)
}

// Close 空实现，满足接口约定（内存实现无需关闭资源）
func (r *MemConfigRepo) Close() error { return nil }
