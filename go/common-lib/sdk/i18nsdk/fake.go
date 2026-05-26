package i18nsdk

import (
	"context"
)

// FakeClient 国际化 SDK 的 Fake 实现
// 用于单元测试，不依赖真实国际化服务
type FakeClient struct {
	translations map[string]map[string]string // lang -> key -> template
	errOnKey     string                       // 模拟渲染失败的 key
}

// NewFakeClient 创建 Fake 国际化客户端
func NewFakeClient() *FakeClient {
	return &FakeClient{
		translations: make(map[string]map[string]string),
	}
}

// SetTranslation 设置翻译文本（测试辅助方法）
func (f *FakeClient) SetTranslation(lang, key, template string) {
	if f.translations[lang] == nil {
		f.translations[lang] = make(map[string]string)
	}
	f.translations[lang][key] = template
}

// SetErrorOnKey 模拟指定 key 渲染失败
func (f *FakeClient) SetErrorOnKey(key string) {
	f.errOnKey = key
}

// T 渲染多语言文本（简化实现：仅做字符串替换）
func (f *FakeClient) T(ctx context.Context, lang string, key string, params map[string]any) (string, error) {
	if key == f.errOnKey {
		return "", ErrRenderFailed
	}

	template, err := f.getRawTemplate(lang, key)
	if err != nil {
		return "", err
	}

	result := template
	for k, v := range params {
		result = replacePlaceholder(result, k, v)
	}
	return result, nil
}

// Raw 获取原始模板
func (f *FakeClient) Raw(ctx context.Context, lang string, key string) (string, string, error) {
	template, err := f.getRawTemplate(lang, key)
	if err != nil {
		return "", "", err
	}
	return template, "named", nil
}

// Ping 健康检查（Fake 实现永远返回 nil）
func (f *FakeClient) Ping(ctx context.Context) error {
	return nil
}

// getRawTemplate 获取原始模板文本
func (f *FakeClient) getRawTemplate(lang, key string) (string, error) {
	if langMap, ok := f.translations[lang]; ok {
		if template, ok := langMap[key]; ok {
			return template, nil
		}
	}
	return "", ErrKeyNotFound
}

// replacePlaceholder 简单的命名参数替换
func replacePlaceholder(template string, key string, value any) string {
	result := template
	placeholder := "{" + key + "}"
	switch v := value.(type) {
	case string:
		result = replaceAll(result, placeholder, v)
	default:
		result = replaceAll(result, placeholder, formatValue(v))
	}
	return result
}

// replaceAll 替换所有匹配项
func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	result += s
	return result
}

// indexOf 查找子串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// formatValue 格式化值为字符串
func formatValue(v any) string {
	switch val := v.(type) {
	case int:
		return formatInt(val)
	case int64:
		return formatInt(int(val))
	case int32:
		return formatInt(int(val))
	case float64:
		return formatFloat(val)
	case string:
		return val
	default:
		return ""
	}
}

// formatInt 整数转字符串
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := make([]byte, 0, 20)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10	}
	reverse(digits)
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// reverse 反转字节切片
func reverse(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

// formatFloat 浮点数转字符串（简化实现）
func formatFloat(f float64) string {
	intPart := int(f)
 fracPart := int((f - float64(intPart)) * 100)
	if fracPart == 0 {
		return formatInt(intPart)
	}
	return formatInt(intPart) + "." + formatInt(fracPart)
}

// 错误定义
var (
	ErrKeyNotFound    = Error("i18n key not found")
	ErrRenderFailed   = Error("i18n render failed")
	ErrLangNotSupported = Error("language not supported")
)

// Error 自定义错误类型
type Error string

func (e Error) Error() string { return string(e) }
