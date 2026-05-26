package i18n

import (
	"context"
	"fmt"
	"testing"

	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
)

func TestResolveLang_ExtendLangPriority(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	result := ResolveLang(context.Background(), "en", "", cfg)
	if result != "en" {
		t.Fatalf("期望 extendLang 优先，实际 '%s'", result)
	}
}

func TestResolveLang_ReqLangFallback(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	result := ResolveLang(context.Background(), "", "ja", cfg)
	if result != "ja" {
		t.Fatalf("期望 reqLang 兜底，实际 '%s'", result)
	}
}

func TestResolveLang_ConfigFallback(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("system_cfg", "default_lang_code", "fr")

	result := ResolveLang(context.Background(), "", "", cfg)
	if result != "fr" {
		t.Fatalf("期望 configsdk 降级，实际 '%s'", result)
	}
}

func TestResolveLang_DefaultFallback(t *testing.T) {
	result := ResolveLang(context.Background(), "", "", nil)
	if result != DefaultLang {
		t.Fatalf("期望默认值 '%s'，实际 '%s'", DefaultLang, result)
	}
}

func TestResolveLang_TrimSpace(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	result := ResolveLang(context.Background(), "  zh-CN  ", "", cfg)
	if result != "zh-CN" {
		t.Fatalf("期望去除空格，实际 '%s'", result)
	}
}

func TestResolveLang_PriorityOrder(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("system_cfg", "default_lang_code", "de")

	result := ResolveLang(context.Background(), "extend", "req", cfg)
	if result != "extend" {
		t.Fatalf("期望 extendLang 最高优先级，实际 '%s'", result)
	}

	result2 := ResolveLang(context.Background(), "", "req", cfg)
	if result2 != "req" {
		t.Fatalf("期望 reqLang 第二优先级，实际 '%s'", result2)
	}
}

func TestMustResolveLang_Alias(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	result := MustResolveLang(context.Background(), "en", "", cfg)
	if result != "en" {
		t.Fatalf("MustResolveLang 应与 ResolveLang 行为一致")
	}
}

var _ = fmt.Sprintf
