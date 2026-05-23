package sdk

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

var (
	// ErrICUNotSupported ICU 模板暂不支持
	ErrICUNotSupported = errors.New("icu template not supported in MVP")
	// ErrMissingRequiredParam 缺少必需参数
	ErrMissingRequiredParam = errors.New("missing required parameter")
)

var namedPlaceholderRegex = regexp.MustCompile(`\{(\w+)\}`)

// T 翻译指定 key 的文本并渲染参数
//
// 核心逻辑：
// 1. 按 langCode + key 查找语言包条目（先 LRU → miss 则查 service）
// 2. 如果 template_type == "plain"：直接返回 value
// 3. 如果 template_type == "named"：执行 named 参数替换
// 4. 如果 template_type == "icu"：返回 ErrICUNotSupported
// 5. 如果 key 不存在：返回 key 本身作为 fallback（不报错）
func (c *clientImpl) T(ctx context.Context, langCode, key string, params map[string]any) (string, error) {
	tmpl, err := c.getRawOrLoad(ctx, langCode, key)
	if err != nil {
		return key, nil // fallback: key 不存在时返回 key 本身
	}

	switch tmpl.TemplateType {
	case "plain":
		return tmpl.Value, nil
	case "named":
		return renderNamedTemplate(tmpl.Value, params, tmpl.Params)
	case "icu":
		return "", ErrICUNotSupported
	default:
		return tmpl.Value, nil
	}
}

// Raw 获取原始模板信息（不渲染）
//
// 返回 Template 结构体包含：
// - Key: 模板键
// - Value: 原始模板值
// - TemplateType: 模板类型
// - Params: 参数描述列表
func (c *clientImpl) Raw(ctx context.Context, langCode, key string) (*Template, error) {
	return c.getRawOrLoad(ctx, langCode, key)
}

// getRawOrLoad 获取原始模板，缓存未命中时从 service 加载
func (c *clientImpl) getRawOrLoad(ctx context.Context, langCode, key string) (*Template, error) {
	cacheKey := cacheKey{
		env:      c.options.Env,
		langCode: langCode,
	}

	if templates, found, expired := c.cache.Get(cacheKey); found && !expired {
		if tmpl, exists := templates[key]; exists {
			return tmpl, nil
		}
	}

	pack, err := c.loadPack(ctx, langCode)
	if err != nil {
		return nil, err
	}

	if tmpl, exists := pack[key]; exists {
		return tmpl, nil
	}

	return nil, fmt.Errorf("key %s not found", key)
}

// loadPack 从 service 加载完整语言包并写入缓存
func (c *clientImpl) loadPack(ctx context.Context, langCode string) (map[string]*Template, error) {
	var packResp *service.LangPackResponse
	var err error

	switch c.options.Mode {
	case ModeInProcess:
		packResp, err = c.options.Service.GetLangPack(langCode, "", c.options.Env)
	case ModeRemote:
		packResp, err = c.remote.getLangPack(ctx, langCode)
	default:
		return nil, fmt.Errorf("unsupported mode: %s", c.options.Mode)
	}

	if err != nil {
		return nil, err
	}

	templates := convertPackToTemplates(packResp)

	cacheKey := cacheKey{
		env:         c.options.Env,
		langCode:    langCode,
		packVersion: packResp.PackVersion,
	}
	c.cache.Set(cacheKey, templates)

	return templates, nil
}

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
// 输入: value="欢迎 {name}，你有 {count} 条新消息", params={"name": "张三", "count": 42}
// 输出: "欢迎 张三，你有 42 条新消息"
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

// Ping 检查服务可用性
func (c *clientImpl) Ping(ctx context.Context) error {
	switch c.options.Mode {
	case ModeInProcess:
		if c.options.Service == nil {
			return errors.New("service not initialized")
		}
		_, err := c.options.Service.GetLanguages("")
		return err
	case ModeRemote:
		return c.remote.ping(ctx)
	default:
		return fmt.Errorf("unsupported mode: %s", c.options.Mode)
	}
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
