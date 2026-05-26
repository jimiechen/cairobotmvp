package sdk

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// 渲染相关错误定义
var (
	// ErrICUNotSupported ICU 模板暂不支持
	ErrICUNotSupported = errors.New("icu template not supported in MVP")
	// ErrMissingRequiredParam 缺少必需参数
	ErrMissingRequiredParam = errors.New("missing required parameter")
)

// namedPlaceholderRegex 匹配 {paramName} 格式的命名占位符
var namedPlaceholderRegex = regexp.MustCompile(`\{(\w+)\}`)

// convertPackToTemplates 将 LangPackResponse 转换为 map[string]*Template
func convertPackToTemplates(resp *service.LangPackResponse) map[string]*Template {
	templates := make(map[string]*Template, len(resp.Strings))
	for _, entry := range resp.Strings {
		params := make([]ParamInfo, 0, len(entry.Params))
		for _, p := range entry.Params {
			params = append(params, ParamInfo{
				Name:     p.Name,
				Type:     p.Type,
				Required: p.Required,
			})
		}
		templates[entry.Key] = &Template{
			Key:          entry.Key,
			Value:        entry.Value,
			TemplateType: entry.TemplateType,
			Params:       params,
		}
	}
	return templates
}

// renderNamedTemplate 渲染 named 类型模板
//
// 算法说明：
// 1. 解析 value 中的 {paramName} 占位符
// 2. 检查 params 中是否包含所有 required 参数
// 3. 用 params map 中对应的值替换占位符
// 4. 多余参数忽略，未声明的占位符保留原样
//
// 示例：
//
//	输入: value="欢迎 {name}，你有 {count} 条新消息", params={"name": "张三", "count": 42}
//	输出: "欢迎 张三，你有 42 条新消息"
func renderNamedTemplate(value string, params map[string]any, paramInfos []ParamInfo) (string, error) {
	requiredParams := make(map[string]bool)
	for _, info := range paramInfos {
		if info.Required {
			requiredParams[info.Name] = true
		}
	}

	for paramName := range requiredParams {
		if _, exists := params[paramName]; !exists {
			return "", fmt.Errorf("%w: %s", ErrMissingRequiredParam, paramName)
		}
	}

	result := namedPlaceholderRegex.ReplaceAllStringFunc(value, func(match string) string {
		submatches := namedPlaceholderRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		paramName := submatches[1]
		if val, exists := params[paramName]; exists {
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	return result, nil
}

// ExtractPlaceholders 从模板值中提取占位符名称（供测试使用）
func ExtractPlaceholders(value string) []string {
	matches := namedPlaceholderRegex.FindAllStringSubmatch(value, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 2 {
			result = append(result, m[1])
		}
	}
	return result
}

// HasPlaceholder 判断字符串是否包含占位符
func HasPlaceholder(value string) bool {
	return strings.Contains(value, "{")
}
