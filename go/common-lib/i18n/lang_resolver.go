package i18n

import (
	"context"
	"strings"

	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
)

// DefaultLang 默认语言代码
const DefaultLang = "zh-CN"

// ResolveLang 按优先级解析语言代码
// 优先级 1（最高）: extendLang（MessagePacket.extend.langCode）
// 优先级 2（兜底） : reqLang（协议体 lang_code 字段）
// 优先级 3（降级）: configsdk 读取 system_cfg.default_lang_code
// 优先级 4（最终降级）: 硬编码默认值 "zh-CN"
//
// 禁止在 usecase 中直接读取 req.LangCode，必须通过此函数解析
func ResolveLang(ctx context.Context, extendLang string, reqLang string, cfg configsdk.Client) string {
	if extendLang != "" {
		return strings.TrimSpace(extendLang)
	}

	if reqLang != "" {
		return strings.TrimSpace(reqLang)
	}

	if cfg != nil {
		defaultLang, err := cfg.GetString(ctx, "system_cfg", "default_lang_code")
		if err == nil && defaultLang != "" {
			return strings.TrimSpace(defaultLang)
		}
	}

	return DefaultLang
}

// MustResolveLang 必须成功返回语言代码的版本
// 如果所有来源都为空，直接返回默认值而不报错
func MustResolveLang(ctx context.Context, extendLang string, reqLang string, cfg configsdk.Client) string {
	return ResolveLang(ctx, extendLang, reqLang, cfg)
}
