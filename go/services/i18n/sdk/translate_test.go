package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

func TestRenderNamedTemplate_Basic(t *testing.T) {
	value := "欢迎 {name}，你有 {count} 条新消息"
	params := map[string]any{
		"name":  "张三",
		"count": 42,
	}
	paramInfos := []ParamInfo{
		{Name: "name", Type: "string", Required: true},
		{Name: "count", Type: "int", Required: true},
	}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "欢迎 张三，你有 42 条新消息"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRenderNamedTemplate_MissingRequiredParam(t *testing.T) {
	value := "欢迎 {name}，你有 {count} 条新消息"
	params := map[string]any{
		"name": "张三",
	}
	paramInfos := []ParamInfo{
		{Name: "name", Type: "string", Required: true},
		{Name: "count", Type: "int", Required: true},
	}

	_, err := renderNamedTemplate(value, params, paramInfos)
	if err == nil {
		t.Error("expected error for missing required param")
	}
	if !errors.Is(err, ErrMissingRequiredParam) {
		t.Errorf("expected ErrMissingRequiredParam, got %v", err)
	}
}

func TestRenderNamedTemplate_ExtraParams_Ignored(t *testing.T) {
	value := "你好 {name}"
	params := map[string]any{
		"name":    "用户",
		"extra":   "多余参数",
		"another": 123,
	}
	paramInfos := []ParamInfo{
		{Name: "name", Type: "string", Required: true},
	}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "你好 用户"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRenderNamedTemplate_UndeclaredPlaceholder_Kept(t *testing.T) {
	value := "测试 {unknown} 占位符"
	params := map[string]any{}
	paramInfos := []ParamInfo{}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("expected undeclared placeholder to be kept, got '%s'", result)
	}
}

func TestRenderNamedTemplate_NoPlaceholders(t *testing.T) {
	value := "纯文本，没有占位符"
	params := map[string]any{}
	paramInfos := []ParamInfo{}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != value {
		t.Errorf("expected plain text unchanged, got '%s'", result)
	}
}

func TestRenderNamedTemplate_SinglePlaceholder(t *testing.T) {
	value := "值: {value}"
	params := map[string]any{"value": 3.14}
	paramInfos := []ParamInfo{{Name: "value", Type: "float", Required: true}}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "值: 3.14" {
		t.Errorf("expected '值: 3.14', got '%s'", result)
	}
}

func TestRenderNamedTemplate_RepeatedPlaceholder(t *testing.T) {
	value := "{greeting}，{name}！{greeting}，欢迎回来！"
	params := map[string]any{
		"greeting": "你好",
		"name":     "世界",
	}
	paramInfos := []ParamInfo{
		{Name: "greeting", Type: "string", Required: true},
		{Name: "name", Type: "string", Required: true},
	}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "你好，世界！你好，欢迎回来！"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRenderNamedTemplate_OptionalParam_Missing(t *testing.T) {
	value := "可选参数测试：{optional}"
	params := map[string]any{}
	paramInfos := []ParamInfo{
		{Name: "optional", Type: "string", Required: false},
	}

	result, err := renderNamedTemplate(value, params, paramInfos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "可选参数测试：{optional}" {
		t.Errorf("expected optional param to be kept when missing, got '%s'", result)
	}
}

func TestT_ICU_Template_ReturnsError(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "icu_key", Value: "{gender, select, male{他} female{她} other{它}}", TemplateType: "icu"},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	_, err := client.T(context.Background(), "zh-CN", "icu_key", nil)
	if err == nil {
		t.Error("expected error for ICU template")
	}
	if !errors.Is(err, ErrICUNotSupported) {
		t.Errorf("expected ErrICUNotSupported, got %v", err)
	}
}

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"无占位符", nil},
		{"{single}", []string{"single"}},
		{"{first} 和 {second}", []string{"first", "second"}},
		{"重复 {x} 和 {x}", []string{"x", "x"}},
		{"{param_with_underscore}", []string{"param_with_underscore"}},
		{"{param123}", []string{"param123"}},
	}

	for _, tt := range tests {
		result := ExtractPlaceholders(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("for input '%s': expected %d placeholders, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("for input '%s'[%d]: expected '%s', got '%s'", tt.input, i, tt.expected[i], v)
			}
		}
	}
}

func TestHasPlaceholder(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"纯文本", false},
		{"包含 {brace}", true},
		{"{start", true},
		{"{end}", true},
	}

	for _, tt := range tests {
		result := HasPlaceholder(tt.input)
		if result != tt.expected {
			t.Errorf("HasPlaceholder('%s'): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestConvertPackToTemplates(t *testing.T) {
	resp := &service.LangPackResponse{
		PackVersion: 42,
		Strings: []service.LangStringEntry{
			{
				Key:          "key1",
				Value:        "值1",
				TemplateType: "plain",
			},
			{
				Key:          "key2",
				Value:        "值 {param}",
				TemplateType: "named",
				Params: []service.LangParamEntry{
					{Name: "param", Type: "string", Required: true, DefaultV: "默认"},
				},
			},
		},
	}

	templates := convertPackToTemplates(resp)

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	tmpl1 := templates["key1"]
	if tmpl1.Value != "值1" || tmpl1.TemplateType != "plain" {
		t.Errorf("template key1 mismatch: %+v", tmpl1)
	}

	tmpl2 := templates["key2"]
	if tmpl2.TemplateType != "named" || len(tmpl2.Params) != 1 {
		t.Errorf("template key2 mismatch: %+v", tmpl2)
	}
	if tmpl2.Params[0].Name != "param" || !tmpl2.Params[0].Required {
		t.Errorf("template key2 param mismatch: %+v", tmpl2.Params[0])
	}
}

func TestT_ServiceError_ReturnsFallbackKey(t *testing.T) {
	mockSvc := &mockI18nService{
		err: errors.New("service connection failed"),
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	result, err := client.T(context.Background(), "zh-CN", "any_key", nil)
	if err != nil {
		t.Errorf("expected no error (fallback to key), got %v", err)
	}
	if result != "any_key" {
		t.Errorf("expected fallback to 'any_key', got '%s'", result)
	}
}

func TestRaw_KeyNotFound(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings:     []service.LangStringEntry{},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	_, err := client.Raw(context.Background(), "zh-CN", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key in Raw()")
	}
}
