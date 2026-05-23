package service

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/cache"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
)

// I18nServiceImpl 国际化服务默认实现
//
// 职责：
// - 协调 Repository 和 Cache 完成业务逻辑
// - 实现全量查询、增量查询、元数据查询等核心能力
// - 应用模板校验和兼容性过滤
//
// 不负责：
// - HTTP 接口处理（由 Controller 层负责）
// - 数据库连接管理（由基础设施层负责）
type I18nServiceImpl struct {
	repo  repository.I18nRepository
	cache cache.Cache
	env   string
}

// NewI18nService 创建国际化服务实例
func NewI18nService(repo repository.I18nRepository, c cache.Cache, env string) *I18nServiceImpl {
	return &I18nServiceImpl{
		repo:  repo,
		cache: c,
		env:   env,
	}
}

// GetLangPack 获取全量语言包
// 流程：查缓存 → 缓存未命中则查数据库 → 写入缓存 → 返回结果
func (s *I18nServiceImpl) GetLangPack(langCode string, clientVersion string, env string) (*LangPackResponse, error) {
	pack, err := s.getOrLoadPack(langCode, env)
	if err != nil {
		return nil, err
	}
	if pack == nil {
		return &LangPackResponse{
			PackVersion: 0,
			Strings:     []LangStringEntry{},
		}, nil
	}

	strings, err := s.getOrLoadStrings(pack.ID)
	if err != nil {
		return nil, err
	}

	entries := convertToEntries(strings)

	filtered := ApplyCompatFilter(entries, clientVersion)

	return &LangPackResponse{
		PackVersion: int64(pack.Version),
		Strings:     filtered,
	}, nil
}

// getOrLoadPack 获取或加载语言包（带缓存）
func (s *I18nServiceImpl) getOrLoadPack(langCode string, env string) (*domain.LangPack, error) {
	if pack, exists := s.cache.GetPack(langCode, env); exists {
		return pack, nil
	}

	pack, err := s.repo.GetPackByLangCode(langCode, env)
	if err != nil {
		return nil, err
	}

	if pack != nil {
		s.cache.SetPack(langCode, env, pack)
	}

	return pack, nil
}

// getOrLoadStrings 获取或加载字符串列表（带缓存）
func (s *I18nServiceImpl) getOrLoadStrings(packID int64) ([]domain.LangString, error) {
	if strings, exists := s.cache.GetStrings(packID); exists {
		return strings, nil
	}

	strings, err := s.repo.GetStringsByPackID(packID)
	if err != nil {
		return nil, err
	}

	s.cache.SetStrings(packID, strings)

	return strings, nil
}

// convertToEntries 将 domain.LangString 列表转换为 LangStringEntry 列表
func convertToEntries(strings []domain.LangString) []LangStringEntry {
	entries := make([]LangStringEntry, 0, len(strings))
	for _, s := range strings {
		params, _ := s.GetParams()
		entry := LangStringEntry{
			Key:          string(s.StringKey),
			Value:        s.StringValue,
			TemplateType: string(s.TemplateType),
			OperationType: string(s.OperationType),
		}
		for _, p := range params {
			entry.Params = append(entry.Params, LangParamEntry{
				Name:     p.Name,
				Type:     p.Type,
				Required: p.Required,
				DefaultV: p.DefaultV,
			})
		}
		entries = append(entries, entry)
	}
	return entries
}
