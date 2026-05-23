package domain

import "testing"

func TestLangString_TemplateTypeCheckers(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		isPlain      bool
		isNamed      bool
		isIcu        bool
	}{
		{"plain 类型", TemplatePlain, true, false, false},
		{"named 类型", TemplateNamed, false, true, false},
		{"icu 类型", TemplateIcu, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LangString{TemplateType: tt.templateType}
			if got := s.IsPlain(); got != tt.isPlain {
				t.Errorf("IsPlain() = %v, want %v", got, tt.isPlain)
			}
			if got := s.IsNamed(); got != tt.isNamed {
				t.Errorf("IsNamed() = %v, want %v", got, tt.isNamed)
			}
			if got := s.IsIcu(); got != tt.isIcu {
				t.Errorf("IsIcu() = %v, want %v", got, tt.isIcu)
			}
		})
	}
}

func TestLangString_GetParams(t *testing.T) {
	jsonStr := `[{"name":"name","type":"string","required":true}]`
	s := &LangString{ParamsSchema: jsonStr}

	params, err := s.GetParams()
	if err != nil {
		t.Fatalf("GetParams() error = %v", err)
	}

	if len(params) != 1 || params[0].Name != "name" {
		t.Error("GetParams() returned wrong params")
	}
}

func TestLangString_GetParams_Empty(t *testing.T) {
	s := &LangString{ParamsSchema: ""}

	params, err := s.GetParams()
	if err != nil {
		t.Fatalf("GetParams() error = %v", err)
	}

	if params != nil {
		t.Error("GetParams() should return nil for empty schema")
	}
}
