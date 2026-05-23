package sdk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Mode != ModeInProcess {
		t.Errorf("expected mode %s, got %s", ModeInProcess, opts.Mode)
	}
	if opts.Env != "dev" {
		t.Errorf("expected env dev, got %s", opts.Env)
	}
	if opts.DefaultLangCode != "zh-CN" {
		t.Errorf("expected defaultLangCode zh-CN, got %s", opts.DefaultLangCode)
	}
}

func TestDefault_InProcessMode_WithoutService(t *testing.T) {
	_, err := Default()
	if err == nil {
		t.Error("expected error when service is nil in in_process mode")
	}
}

func TestDefault_InProcessMode_WithService(t *testing.T) {
	mockSvc := &mockI18nService{}
	client, err := Default(func(o *Options) {
		o.Service = mockSvc
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestDefault_RemoteMode(t *testing.T) {
	client, err := Default(func(o *Options) {
		o.Mode = ModeRemote
		o.TarsServant = "test-object"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestClient_Ping_Success(t *testing.T) {
	mockSvc := &mockI18nService{
		languages: []service.LanguageMeta{
			{Code: "zh-CN", Name: "简体中文"},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Ping_Failure(t *testing.T) {
	mockSvc := &mockI18nService{
		err: errors.New("service unavailable"),
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	err := client.Ping(context.Background())
	if err == nil {
		t.Error("expected error when service fails")
	}
}

type mockI18nService struct {
	languages []service.LanguageMeta
	pack      *service.LangPackResponse
	err       error
}

func (m *mockI18nService) GetLanguages(clientVersion string) ([]service.LanguageMeta, error) {
	return m.languages, m.err
}

func (m *mockI18nService) GetLangPack(langCode string, clientVersion string, env string) (*service.LangPackResponse, error) {
	return m.pack, m.err
}

func (m *mockI18nService) GetLangDifference(langCode string, sinceVersion int64, clientVersion string, env string) (*service.LangDiffResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockI18nService) ValidateTemplate(value string, templateType domain.TemplateType, params []domain.LangParam) error {
	return nil
}

func TestClient_T_PlainTemplate(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{
					Key:          "greeting",
					Value:        "你好世界",
					TemplateType: "plain",
				},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	result, err := client.T(context.Background(), "zh-CN", "greeting", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "你好世界" {
		t.Errorf("expected '你好世界', got '%s'", result)
	}
}

func TestClient_T_NamedTemplate(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{
					Key:          "welcome",
					Value:        "欢迎 {name}，你有 {count} 条新消息",
					TemplateType: "named",
					Params: []service.LangParamEntry{
						{Name: "name", Type: "string", Required: true},
						{Name: "count", Type: "int", Required: true},
					},
				},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	params := map[string]any{
		"name":  "张三",
		"count": 42,
	}
	result, err := client.T(context.Background(), "zh-CN", "welcome", params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "欢迎 张三，你有 42 条新消息"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestClient_T_KeyNotFound_ReturnsKey(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings:     []service.LangStringEntry{},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	result, err := client.T(context.Background(), "zh-CN", "nonexistent", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "nonexistent" {
		t.Errorf("expected fallback to key 'nonexistent', got '%s'", result)
	}
}

func TestClient_Raw(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{
					Key:          "test_key",
					Value:        "测试值 {param}",
					TemplateType: "named",
					Params: []service.LangParamEntry{
						{Name: "param", Type: "string", Required: true},
					},
				},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	tmpl, err := client.Raw(context.Background(), "zh-CN", "test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tmpl.Key != "test_key" {
		t.Errorf("expected key 'test_key', got '%s'", tmpl.Key)
	}
	if tmpl.Value != "测试值 {param}" {
		t.Errorf("expected value '测试值 {param}', got '%s'", tmpl.Value)
	}
	if tmpl.TemplateType != "named" {
		t.Errorf("expected type 'named', got '%s'", tmpl.TemplateType)
	}
	if len(tmpl.Params) != 1 || tmpl.Params[0].Name != "param" {
		t.Errorf("expected 1 param named 'param', got %v", tmpl.Params)
	}
}

func TestClient_BatchT(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings: []service.LangStringEntry{
				{Key: "key1", Value: "值1", TemplateType: "plain"},
				{Key: "key2", Value: "值2", TemplateType: "plain"},
				{Key: "key3", Value: "欢迎 {name}", TemplateType: "named",
					Params: []service.LangParamEntry{{Name: "name", Type: "string", Required: true}}},
			},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	params := map[string]any{"name": "用户"}
	results, err := client.BatchT(context.Background(), "zh-CN", []string{"key1", "key2", "key3"}, params)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if results["key1"] != "值1" {
		t.Errorf("expected key1='值1', got '%s'", results["key1"])
	}
	if results["key2"] != "值2" {
		t.Errorf("expected key2='值2', got '%s'", results["key2"])
	}
	if results["key3"] != "欢迎 用户" {
		t.Errorf("expected key3='欢迎 用户', got '%s'", results["key3"])
	}
}

func TestClient_Watch(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings:     []service.LangStringEntry{},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	received := make(chan int64, 1)
	cancel := client.Watch("zh-CN", func(version int64) {
		received <- version
	})

	ci := client.(*clientImpl)
	ci.watchers.Trigger("zh-CN", 42)

	select {
	case v := <-received:
		if v != 42 {
			t.Errorf("expected version 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for watcher callback")
	}

	cancel()
}

func TestClient_Watch_Cancel(t *testing.T) {
	mockSvc := &mockI18nService{
		pack: &service.LangPackResponse{
			PackVersion: 1,
			Strings:     []service.LangStringEntry{},
		},
	}
	client, _ := Default(func(o *Options) {
		o.Service = mockSvc
	})

	callCount := 0
	cancel := client.Watch("zh-CN", func(version int64) {
		callCount++
	})

	cancel()

	ci := client.(*clientImpl)
	ci.watchers.Trigger("zh-CN", 99)

	if callCount != 0 {
		t.Errorf("expected no callback after cancel, got %d calls", callCount)
	}
}
