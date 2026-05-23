package repository

import (
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// ConfigRepository 配置版本数据访问接口
// 抽象了底层存储（SQLite / MySQL）的差异，Service 层只依赖此接口
// 不负责缓存，缓存由独立的 Cache 层处理
type ConfigRepository interface {
	GetLatestVersion(moduleKey, env string) (*domain.ConfigVersion, error)
	GetByModuleAndVersion(moduleKey, env string, version int64) (*domain.ConfigVersion, error)
	ListPublishedVersions(env string) ([]*domain.ConfigVersion, error)
	Save(version *domain.ConfigVersion) error
}
