package sdk

import (
	"context"
	"errors"
	"testing"
	
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

type mockConfigService struct {
	configs map[string]map[string]*domain.TypedValue
	err     error
}

func (m *mockConfigService) GetAppConfigs(req *service.AppConfigRequest) (*service.AppConfigResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	resp := &service.AppConfigResponse{
		StaticModules:   make(map[string]map[string]*domain.TypedValue),
		DynamicModules: []*service.DynamicModuleView{},
	}
	for moduleKey, fields := range m.configs {
		if domain.IsStaticModule(moduleKey) {
			resp.StaticModules[moduleKey] = fields
		} else {
			resp.DynamicModules = append(resp.DynamicModules, &service.DynamicModuleView{
				ModuleKey: moduleKey,
				Version:   1,
				Fields:    fields,
			})
		}
	}
	return resp, nil
}

func (m *mockConfigService) GetVersionInfo(env string, knownVersions map[string]int64) (*service.VersionInfoResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &service.VersionInfoResponse{
		ConfigVersions: make(map[string]int64),
		HasChanges:     false,
	}, nil
}

func TestDefault_InProcess(t *testing.T) {
	svc := &mockConfigService{
		configs: map[string]map[string]*domain.TypedValue{
			"base_cfg": {
				"app_name": domain.NewTypedValue(domain.FieldTypeString, "test_app"),
			},
		},
	}
	client, err := Default(WithService(svc), WithMode(ModeInProcess))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestDefault_MissingService(t *testing.T) {
	_, err := Default(WithMode(ModeInProcess))
	if !errors.Is(err, ErrServiceRequired) {
		t.Fatalf("expected ErrServiceRequired, got %v", err)
	}
}

func TestDefault_RemoteMode(t *testing.T) {
	client, err := Default(WithMode(ModeRemote))
	if err != nil {
		t.Fatalf("expected no error for remote mode without service, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for remote mode")
	}
}

func TestDefault_WithOptions(t *testing.T) {
	svc := &mockConfigService{}
	client, err := Default(
		WithService(svc),
		WithEnv("prod"),
		WithClientScope("ios"),
		WithCacheSize(512),
		WithCacheTTL(60),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	c := client.(*configClient)
	if c.options.Env != "prod" {
		t.Fatalf("expected env prod, got %s", c.options.Env)
	}
	if c.options.ClientScope != "ios" {
		t.Fatalf("expected scope ios, got %s", c.options.ClientScope)
	}
	if c.options.CacheSize != 512 {
		t.Fatalf("expected cache size 512, got %d", c.options.CacheSize)
	}
	if c.options.CacheTTLSec != 60 {
		t.Fatalf("expected cache TTL 60, got %d", c.options.CacheTTLSec)
	}
}

func TestPing_Success(t *testing.T) {
	svc := &mockConfigService{}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	err := client.Ping(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPing_ServiceError(t *testing.T) {
	svc := &mockConfigService{err: errors.New("connection failed")}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	err := client.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error when service fails")
	}
}

func TestPing_NoService(t *testing.T) {
	_, err := Default(WithMode(ModeInProcess), WithService(nil))
	if !errors.Is(err, ErrServiceRequired) {
		t.Fatalf("expected ErrServiceRequired when creating client with nil service, got %v", err)
	}
}
