package service

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

func TestValidateTemplate(t *testing.T) {
	svc := &I18nServiceImpl{}

	tests := []struct {
		name        string
		value       string
		templateType domain.TemplateType
		params      []domain.LangParam
		wantErr     bool
	}{
		{
			name:         "plain 模板合法",
			value:        "确定",
			templateType: domain.TemplatePlain,
			params:       nil,
			wantErr:      false,
		},
		{
			name:         "named 模板合法",
			value:        "欢迎 {name}，你有 {count} 条消息",
			templateType: domain.TemplateNamed,
			params: []domain.LangParam{
				{Name: "name", Type: "string", Required: true},
				{Name: "count", Type: "int", Required: true},
			},
			wantErr: false,
		},
		{
			name:         "named 模板含未定义占位符",
			value:        "欢迎 {name}，你有 {unknown} 条消息",
			templateType: domain.TemplateNamed,
			params: []domain.LangParam{
				{Name: "name", Type: "string", Required: true},
				{Name: "count", Type: "int", Required: true},
			},
			wantErr: true,
		},
		{
			name:         "named 模板缺少必填参数",
			value:        "欢迎 {name}",
			templateType: domain.TemplateNamed,
			params: []domain.LangParam{
				{Name: "name", Type: "string", Required: true},
				{Name: "count", Type: "int", Required: true},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateTemplate(tt.value, tt.templateType, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCrossLanguagePlaceholders(t *testing.T) {
	tests := []struct {
		name    string
		entries []LangStringEntry
		wantErr bool
	}{
		{
			name: "单语言不报错",
			entries: []LangStringEntry{
				{Key: "key1", Value: "欢迎 {name}"},
			},
			wantErr: false,
		},
		{
			name: "多语言占位符一致",
			entries: []LangStringEntry{
				{Key: "key1", Value: "欢迎 {name}，你有 {count} 条消息"},
				{Key: "key1", Value: "Welcome {name}, you have {count} messages"},
			},
			wantErr: false,
		},
		{
			name: "多语言占位符不一致",
			entries: []LangStringEntry{
				{Key: "key1", Value: "欢迎 {name}，你有 {count} 条消息"},
				{Key: "key1", Value: "Welcome {name}"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCrossLanguagePlaceholders(tt.entries)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCrossLanguagePlaceholders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsEqualPlaceholderSet(t *testing.T) {
	tests := []struct {
		name   string
		a      []string
		b      []string
		expect bool
	}{
		{"相同集合", []string{"name", "count"}, []string{"count", "name"}, true},
		{"不同长度", []string{"name", "count"}, []string{"name"}, false},
		{"不同元素", []string{"name", "count"}, []string{"name", "age"}, false},
		{"空集合", []string{}, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEqualPlaceholderSet(tt.a, tt.b); got != tt.expect {
				t.Errorf("isEqualPlaceholderSet() = %v, want %v", got, tt.expect)
			}
		})
	}
}
