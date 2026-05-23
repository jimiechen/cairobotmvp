package sdk

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

func newTestClient(configs map[string]map[string]*domain.TypedValue) (Client, error) {
	svc := &mockConfigService{configs: configs}
	return Default(WithService(svc), WithMode(ModeInProcess))
}

func TestGetString_Success(t *testing.T) {
	client, err := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"app_name": domain.NewTypedValue(domain.FieldTypeString, "my_app"),
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	val, err := client.GetString(context.Background(), "base_cfg", "app_name")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != "my_app" {
		t.Fatalf("expected my_app, got %s", val)
	}
}

func TestGetString_NotFound(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {},
	})
	_, err := client.GetString(context.Background(), "base_cfg", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent field")
	}
}

func TestGetInt_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"timeout": domain.NewTypedValue(domain.FieldTypeInt, int64(30)),
		},
	})
	val, err := client.GetInt(context.Background(), "base_cfg", "timeout")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != 30 {
		t.Fatalf("expected 30, got %d", val)
	}
}

func TestGetBool_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"enabled": domain.NewTypedValue(domain.FieldTypeBool, true),
		},
	})
	val, err := client.GetBool(context.Background(), "base_cfg", "enabled")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !val {
		t.Fatal("expected true")
	}
}

func TestGetFloat_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"rate": domain.NewTypedValue(domain.FieldTypeFloat, 3.14),
		},
	})
	val, err := client.GetFloat(context.Background(), "base_cfg", "rate")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != 3.14 {
		t.Fatalf("expected 3.14, got %f", val)
	}
}

type testJSONConfig struct {
	Name string `json:"name"`
	Ver  int    `json:"version"`
}

func TestGetJSON_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"meta": domain.NewTypedValue(domain.FieldTypeJSON, json.RawMessage(`{"name":"test","version":1}`)),
		},
	})
	var out testJSONConfig
	err := client.GetJSON(context.Background(), "base_cfg", "meta", &out)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Name != "test" || out.Ver != 1 {
		t.Fatalf("expected {name:test, version:1}, got %+v", out)
	}
}

func TestGetJSON_TypeMismatch(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"plain": domain.NewTypedValue(domain.FieldTypeString, "not_json"),
		},
	})
	var out map[string]interface{}
	err := client.GetJSON(context.Background(), "base_cfg", "plain", &out)
	if err == nil {
		t.Fatal("expected error for non-JSON field")
	}
}

func TestGetString_CacheHit(t *testing.T) {
	callCount := 0
	svc := &mockCountingService{
		configs: map[string]map[string]*domain.TypedValue{
			"base_cfg": {
				"cached_field": domain.NewTypedValue(domain.FieldTypeString, "cached_value"),
			},
		},
		count: &callCount,
	}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	val1, _ := client.GetString(context.Background(), "base_cfg", "cached_field")
	val2, _ := client.GetString(context.Background(), "base_cfg", "cached_field")
	if val1 != "cached_value" || val2 != "cached_value" {
		t.Fatal("expected cached_value")
	}
	if callCount != 1 {
		t.Fatalf("expected 1 service call (cache hit on second), got %d", callCount)
	}
}

type mockCountingService struct {
	configs map[string]map[string]*domain.TypedValue
	count   *int
	err     error
}

func (m *mockCountingService) GetAppConfigs(req *service.AppConfigRequest) (*service.AppConfigResponse, error) {
	*m.count++
	if m.err != nil {
		return nil, m.err
	}
	return (&mockConfigService{configs: m.configs}).GetAppConfigs(req)
}

func (m *mockCountingService) GetVersionInfo(env string, knownVersions map[string]int64) (*service.VersionInfoResponse, error) {
	return &service.VersionInfoResponse{}, nil
}
