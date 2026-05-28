package repository

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// I18nRepository 语言包数据仓库接口
// 定义数据访问层的抽象契约，支持多种数据库实现
//
// 职责：
// - 定义数据访问的标准接口
// - 隔离业务逻辑与数据库实现细节
//
// 实现类：
// - SQLiteRepo: SQLite 实现（开发/测试用）
// - MySQLRepo: MySQL 实现（生产环境）
type I18nRepository interface {
	// GetPackByLangCode 根据语言代码查询语言包
	GetPackByLangCode(langCode string, env string) (*domain.LangPack, error)

	// GetStringsByPackID 根据语言包 ID 查询所有字符串
	GetStringsByPackID(packID int64) ([]domain.LangString, error)

	// GetDiffSince 查询指定版本之后的增量变更
	// 返回新增和修改的字符串（不含 DEL 类型）
	GetDiffSince(packID int64, sinceVersion int) ([]domain.LangString, error)

	// ListPacks 列出所有已发布的语言包
	ListPacks(env string) ([]domain.LangPack, error)

	// --- 以下为 Admin 写入操作 ---

	// SaveString 新增或更新一条语言字符串
	SaveString(s *domain.LangString) error

	// DeleteString 标记删除一条语言字符串（operation_type=DEL）
	DeleteString(id int64) error

	// PublishPack 发布语言包（标记 IsPublished + 更新 Version）
	PublishPack(packID int64, version int) error

	// FindStringByKey 按 string_key 查询单条字符串
	FindStringByKey(packID int64, key domain.StringKey) (*domain.LangString, error)
}
