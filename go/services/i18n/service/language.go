package service

// GetLanguages 获取支持的语言列表
// 流程：查所有已发布的 pack → 去重 lang_code → 组装 LanguageMeta 列表
func (s *I18nServiceImpl) GetLanguages(clientVersion string) ([]LanguageMeta, error) {
	packs, err := s.repo.ListPacks(s.env)
	if err != nil {
		return nil, err
	}

	metaMap := make(map[string]LanguageMeta)
	for _, p := range packs {
		if _, exists := metaMap[p.LangCode]; !exists {
			metaMap[p.LangCode] = LanguageMeta{
				Code:       p.LangCode,
				Name:       getLanguageName(p.LangCode),
				NativeName: getNativeLanguageName(p.LangCode),
				IsDefault:  p.LangCode == "zh-CN",
			}
		}
	}

	result := make([]LanguageMeta, 0, len(metaMap))
	for _, meta := range metaMap {
		result = append(result, meta)
	}

	return result, nil
}

// getLanguageName 获取语言英文名
func getLanguageName(langCode string) string {
	names := map[string]string{
		"zh-CN": "Chinese (Simplified)",
		"en":     "English",
		"ja-JP":  "Japanese",
		"ko-KR":  "Korean",
	}
	if name, ok := names[langCode]; ok {
		return name
	}
	return langCode
}

// getNativeLanguageName 获取语言本地名称
func getNativeLanguageName(langCode string) string {
	names := map[string]string{
		"zh-CN": "简体中文",
		"en":     "English",
		"ja-JP":  "日本語",
		"ko-KR":  "한국어",
	}
	if name, ok := names[langCode]; ok {
		return name
	}
	return langCode
}
