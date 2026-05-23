package service

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// ValidateTemplate 校验模板一致性
// 质量门：保存前必须通过校验
//
// 校验规则：
// 1. value 中 {xxx} 占位符必须出现在 paramsSchema 中
// 2. paramsSchema 中 required=true 的参数必须出现在 value 中
// 3. 多语言 value 的占位符集合必须一致（由调用方保证）
// 4. preview_sample 必须能成功渲染（由客户端完成，此处不做）
//
// 本方法委托给 domain.ValidateTemplate 执行具体校验
func (s *I18nServiceImpl) ValidateTemplate(value string, templateType domain.TemplateType, params []domain.LangParam) error {
	return domain.ValidateTemplate(value, templateType, params)
}

// ValidateCrossLanguagePlaceholders 校验多语言占位符一致性
// 确保同一 key 的不同语言版本的占位符集合完全一致
//
// Args:
//   - entries: 同一 key 的多语言条目列表
//
// Returns:
//   - error: 如果占位符不一致则返回错误
func ValidateCrossLanguagePlaceholders(entries []LangStringEntry) error {
	if len(entries) <= 1 {
		return nil
	}

	referencePlaceholders := extractEntryPlaceholders(entries[0].Value)
	for i := 1; i < len(entries); i++ {
		currentPlaceholders := extractEntryPlaceholders(entries[i].Value)
		if !isEqualPlaceholderSet(referencePlaceholders, currentPlaceholders) {
			return newPlaceholderMismatchError(entries[0].Key, entries[i-1].Value, entries[i].Value)
		}
	}

	return nil
}

// extractEntryPlaceholders 从 LangStringEntry.Value 提取占位符
func extractEntryPlaceholders(value string) []string {
	return domain.ExtractPlaceholders(value)
}

// isEqualPlaceholderSet 判断两个占位符集合是否一致（忽略顺序）
func isEqualPlaceholderSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}
	for _, item := range b {
		if !set[item] {
			return false
		}
	}
	return true
}

// PlaceholderMismatchError 占位符不一致错误
type PlaceholderMismatchError struct {
	Key     string
	Value1  string
	Value2  string
}

func (e *PlaceholderMismatchError) Error() string {
	return "placeholder mismatch for key " + e.Key + ": " + e.Value1 + " vs " + e.Value2
}

func newPlaceholderMismatchError(key, value1, value2 string) error {
	return &PlaceholderMismatchError{
		Key:    key,
		Value1: value1,
		Value2: value2,
	}
}
