package service

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// I18nService 国际化服务接口
// 定义语言包服务的业务能力契约
//
// 职责：
// - 定义服务层的标准接口
// - 协调 Repository 和 Cache 完成业务逻辑
//
// 实现类：
// - I18nServiceImpl: 默认实现
type I18nService interface {
	// GetLanguages 获取支持的语言列表（元数据）
	GetLanguages(clientVersion string) ([]LanguageMeta, error)

	// GetLangPack 获取全量语言包
	// 用于首次加载或本地缓存失效时的全量拉取
	GetLangPack(langCode string, clientVersion string, env string) (*LangPackResponse, error)

	// GetLangDifference 获取增量语言包
	// 客户端传入当前版本号，服务端返回增量 diff
	GetLangDifference(langCode string, sinceVersion int64, clientVersion string, env string) (*LangDiffResponse, error)

	// ValidateTemplate 校验模板一致性
	// 质量门：保存前必须通过校验
	ValidateTemplate(value string, templateType domain.TemplateType, params []domain.LangParam) error
}

// LanguageMeta 语言元信息
// 对应 Proto 的 LanguageMeta message
type LanguageMeta struct {
	Code       string
	Name       string
	NativeName string
	IsDefault  bool
}

// LangPackResponse 全量语言包响应
type LangPackResponse struct {
	PackVersion int64
	Strings     []LangStringEntry
}

// LangStringEntry 语言字符串条目（含参数化模板支持）
// 对应 Proto 的 LangStringEntry message
type LangStringEntry struct {
	Key          string
	Value        string
	TemplateType string
	Params       []LangParamEntry
	OperationType string
}

// LangParamEntry 参数描述条目
type LangParamEntry struct {
	Name     string
	Type     string
	Required bool
	DefaultV string
}

// LangDiffResponse 增量语言包响应
type LangDiffResponse struct {
	CurrentVersion int64
	Additions      []LangStringEntry
	Deletions      []string
}
