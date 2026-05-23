package service

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// GetLangDifference 获取增量语言包
// 流程：查 pack → 查 since_version 之后的 ADD/MOD → 筛选 DEL → 返回 additions + deletions
func (s *I18nServiceImpl) GetLangDifference(langCode string, sinceVersion int64, clientVersion string, env string) (*LangDiffResponse, error) {
	pack, err := s.getOrLoadPack(langCode, env)
	if err != nil {
		return nil, err
	}
	if pack == nil {
		return &LangDiffResponse{
			CurrentVersion: 0,
			Additions:      []LangStringEntry{},
			Deletions:      []string{},
		}, nil
	}

	diffStrings, err := s.repo.GetDiffSince(pack.ID, int(sinceVersion))
	if err != nil {
		return nil, err
	}

	additions := make([]LangStringEntry, 0)
	deletions := []string{}

	for _, ds := range diffStrings {
		if ds.OperationType == domain.OperationDel {
			deletions = append(deletions, string(ds.StringKey))
			continue
		}

		params, _ := ds.GetParams()
		entry := LangStringEntry{
			Key:          string(ds.StringKey),
			Value:        ds.StringValue,
			TemplateType: string(ds.TemplateType),
			OperationType: string(ds.OperationType),
		}
		for _, p := range params {
			entry.Params = append(entry.Params, LangParamEntry{
				Name:     p.Name,
				Type:     p.Type,
				Required: p.Required,
				DefaultV: p.DefaultV,
			})
		}
		additions = append(additions, entry)
	}

	filteredAdditions := ApplyCompatFilter(additions, clientVersion)

	return &LangDiffResponse{
		CurrentVersion: int64(pack.Version),
		Additions:      filteredAdditions,
		Deletions:      deletions,
	}, nil
}
