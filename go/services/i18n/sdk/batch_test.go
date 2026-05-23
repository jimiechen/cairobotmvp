package sdk

import (
	"context"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

func TestBatchT_AllPlainTemplates(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "a", Value: "值A", TemplateType: "plain"},
				{Key: "b", Value: "值B", TemplateType: "plain"},
				{Key: "c", Value: "值C", TemplateType: "plain"},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	results, err := client.BatchT(context.Background(), "zh-CN", []string{"a", "b", "c"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results["a"] != "值A" || results["b"] != "值B" || results["c"] != "值C" {
		t.Errorf("results mismatch: %+v", results)
	}
}

func TestBatchT_MixedTemplateTypes(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "plain_key", Value: "纯文本", TemplateType: "plain"},
				{Key: "named_key", Value: "你好 {name}", TemplateType: "named",
					Params: []service.LangParamEntry{{Name: "name", Type: "string", Required: true}}},
				{Key: "icu_key", Value: "{count, plural, one{# item} other{# items}}", TemplateType: "icu"},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	params := map[string]any{"name": "世界"}
	results, err := client.BatchT(context.Background(), "zh-CN", []string{"plain_key", "named_key", "icu_key"}, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["plain_key"] != "纯文本" {
		t.Errorf("expected plain_key='纯文本', got '%s'", results["plain_key"])
	}
	if results["named_key"] != "你好 世界" {
		t.Errorf("expected named_key='你好 世界', got '%s'", results["named_key"])
	}
	if results["icu_key"] != "icu_key" {
		t.Errorf("expected icu_key fallback to key, got '%s'", results["icu_key"])
	}
}

func TestBatchT_SomeKeysNotFound(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "exists", Value: "存在", TemplateType: "plain"},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	results, err := client.BatchT(context.Background(), "zh-CN", []string{"exists", "missing"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["exists"] != "存在" {
		t.Errorf("expected exists='存在', got '%s'", results["exists"])
	}
	if results["missing"] != "missing" {
		t.Errorf("expected missing fallback to 'missing', got '%s'", results["missing"])
	}
}

func TestBatchT_EmptyKeys(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings:     []service.LangStringEntry{},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	results, err := client.BatchT(context.Background(), "zh-CN", []string{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty result for empty keys, got %d entries", len(results))
	}
}

func TestBatchT_NamedTemplate_MissingRequiredParam(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "req_key", Value: "需要 {param}", TemplateType: "named",
					Params: []service.LangParamEntry{{Name: "param", Type: "string", Required: true}}},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	results, err := client.BatchT(context.Background(), "zh-CN", []string{"req_key"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["req_key"] != "req_key" {
		t.Errorf("expected fallback to key when required param missing, got '%s'", results["req_key"])
	}
}

func TestBatchT_UsesSamePackForAllKeys(t *testing.T) {
	callCount := 0
	mockSvc := &mockI18nServiceWithCounter{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "k1", Value: "v1", TemplateType: "plain"},
				{Key: "k2", Value: "v2", TemplateType: "plain"},
			},
		},
		counter: &callCount,
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	_, _ = client.BatchT(context.Background(), "zh-CN", []string{"k1", "k2"}, nil)

	if callCount != 1 {
		t.Errorf("expected 1 service call for batch, got %d", callCount)
	}
}

type mockI18nServiceWithCounter struct {
	pack    *service.LangPackResponse
	counter *int
}

func (m *mockI18nServiceWithCounter) GetLanguages(clientVersion string) ([]service.LanguageMeta, error) {
	return nil, nil
}

func (m *mockI18nServiceWithCounter) GetLangPack(langCode string, clientVersion string, env string) (*service.LangPackResponse, error) {
	*m.counter++
	return m.pack, nil
}

func (m *mockI18nServiceWithCounter) GetLangDifference(langCode string, sinceVersion int64, clientVersion string, env string) (*service.LangDiffResponse, error) {
	return nil, nil
}

func (m *mockI18nServiceWithCounter) ValidateTemplate(value string, templateType domain.TemplateType, params []domain.LangParam) error {
	return nil
}
