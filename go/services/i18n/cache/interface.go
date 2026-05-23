package cache

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// Cache 缓存接口
// 定义语言包缓存的抽象契约，支持多种缓存实现
//
// 职责：
// - 提供语言包和字符串的缓存读写能力
// - 隔离业务逻辑与缓存实现细节
//
// 实现类：
// - MockCache: 内存 Map 实现（开发/测试用）
// - RedisCache: Redis 实现（生产环境，TODO）
type Cache interface {
	// GetPack 从缓存获取语言包
	GetPack(langCode string, env string) (*domain.LangPack, bool)

	// SetPack 将语言包写入缓存
	SetPack(langCode string, env string, pack *domain.LangPack)

	// GetStrings 从缓存获取字符串列表
	GetStrings(packID int64) ([]domain.LangString, bool)

	// SetStrings 将字符串列表写入缓存
	SetStrings(packID int64, strings []domain.LangString)

	// Invalidate 使缓存失效
	Invalidate(langCode string, env string)
}
