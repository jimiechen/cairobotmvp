package domain

import "testing"

func TestTemplateType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		tt       TemplateType
		expected bool
	}{
		{"plain 合法", TemplatePlain, true},
		{"named 合法", TemplateNamed, true},
		{"icu 合法", TemplateIcu, true},
		{"空字符串不合法", TemplateType(""), false},
		{"非法值不合法", TemplateType("markdown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tt.IsValid(); got != tt.expected {
				t.Errorf("TemplateType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateTemplate_Plain(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"纯文本合法", "确定", false},
		{"纯英文合法", "OK", false},
		{"含占位符非法", "欢迎{name}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.value, TemplatePlain, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTemplate_Named(t *testing.T) {
	params := []LangParam{
		{Name: "name", Type: "string", Required: true},
		{Name: "count", Type: "int", Required: true},
	}

	tests := []struct {
		name    string
		value   string
		params  []LangParam
		wantErr bool
	}{
		{"合法 named 模板", "欢迎 {name}，你有 {count} 条消息", params, false},
		{"未定义的占位符", "欢迎 {name}，你有 {unknown} 条消息", params, true},
		{"缺少必填参数", "欢迎 {name}", params, true},
		{"无占位符但定义了参数", "欢迎", params, true},
		{"空参数列表", "欢迎 {name}", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.value, TemplateNamed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseParamsSchema(t *testing.T) {
	jsonStr := `[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`

	params, err := ParseParamsSchema(jsonStr)
	if err != nil {
		t.Fatalf("ParseParamsSchema() error = %v", err)
	}

	if len(params) != 2 {
		t.Fatalf("ParseParamsSchema() returned %d params, want 2", len(params))
	}

	if params[0].Name != "name" || params[1].Name != "count" {
		t.Error("ParseParamsSchema() returned wrong param names")
	}
}

func TestExtractPlaceholders(t *testing.T) {
	value := "欢迎 {name}，你有 {count} 条消息"
	phs := ExtractPlaceholders(value)

	if len(phs) != 2 {
		t.Fatalf("ExtractPlaceholders() returned %d, want 2", len(phs))
	}

	if phs[0] != "name" || phs[1] != "count" {
		t.Error("ExtractPlaceholders() returned wrong placeholders")
	}
}
