package domain

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// TemplateType 表示模板类型
// 定义字符串值的渲染方式
type TemplateType string

const (
	// TemplatePlain 纯文本，无占位符
	TemplatePlain TemplateType = "plain"
	// TemplateNamed 命名参数模板，使用 {paramName} 占位符
	TemplateNamed TemplateType = "named"
	// TemplateIcu ICU MessageFormat 复杂模板
	TemplateIcu TemplateType = "icu"
)

// IsValid 判断模板类型是否合法
func (tt TemplateType) IsValid() bool {
	switch tt {
	case TemplatePlain, TemplateNamed, TemplateIcu:
		return true
	default:
		return false
	}
}

// LangParam 参数描述
// 定义单个占位符的元信息
type LangParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	DefaultV string `json:"default_v,omitempty"`
}

// ValidateTemplate 校验模板一致性
// 质量门：保存前必须通过校验
//
// 校验规则：
// 1. value 中 {xxx} 占位符必须出现在 paramsSchema 中
// 2. paramsSchema 中 required=true 的参数必须出现在 value 中
// 3. 多语言 value 的占位符集合必须一致（本函数不处理跨语言，由上层保证）
// 4. preview_sample 必须能成功渲染（本函数不做渲染校验，由客户端完成）
func ValidateTemplate(value string, templateType TemplateType, params []LangParam) error {
	switch templateType {
	case TemplatePlain:
		return validatePlainTemplate(value)
	case TemplateNamed:
		return validateNamedTemplate(value, params)
	case TemplateIcu:
		return validateIcuTemplate(value, params)
	default:
		return errors.New("invalid template type")
	}
}

// validatePlainTemplate 校验纯文本模板
// plain 类型不应包含任何 {xxx} 占位符
func validatePlainTemplate(value string) error {
	re := regexp.MustCompile(`\{[^}]+\}`)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return errors.New("plain template should not contain placeholders")
	}
	return nil
}

// validateNamedTemplate 校验命名参数模板
// 规则1：value 中的占位符必须在 params 中定义
// 规则2：params 中 required 的参数必须出现在 value 中
func validateNamedTemplate(value string, params []LangParam) error {
	valuePlaceholders := ExtractPlaceholders(value)
	paramMap := make(map[string]LangParam)
	requiredParams := make(map[string]bool)

	for _, p := range params {
		paramMap[p.Name] = p
		if p.Required {
			requiredParams[p.Name] = true
		}
	}

	for _, ph := range valuePlaceholders {
		if _, exists := paramMap[ph]; !exists {
			return errors.New("placeholder " + ph + " not defined in params schema")
		}
	}

	for reqParam := range requiredParams {
		found := false
		for _, ph := range valuePlaceholders {
			if ph == reqParam {
				found = true
				break
			}
		}
		if !found {
			return errors.New("required param " + reqParam + " missing in value")
		}
	}

	return nil
}

// validateIcuTemplate 校验 ICU 模板
// ICU 模板语法复杂，此处只做基础校验
func validateIcuTemplate(value string, params []LangParam) error {
	if len(params) == 0 && strings.Contains(value, "{") {
		return errors.New("icu template with placeholders must define params")
	}
	return nil
}

// ExtractPlaceholders 从 value 中提取所有 {xxx} 占位符
func ExtractPlaceholders(value string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(value, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 2 {
			result = append(result, m[1])
		}
	}
	return result
}

// ParseParamsSchema 从 JSON 字符串解析参数列表
func ParseParamsSchema(jsonStr string) ([]LangParam, error) {
	if jsonStr == "" || jsonStr == "null" {
		return nil, nil
	}
	var params []LangParam
	err := json.Unmarshal([]byte(jsonStr), &params)
	if err != nil {
		return nil, err
	}
	return params, nil
}
