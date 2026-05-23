package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

func TestFetchModule_InProcess(t *testing.T) {
	svc := &mockConfigService{
		configs: map[string]map[string]*domain.TypedValue{
			"base_cfg": {
				"key1": domain.NewTypedValue(domain.FieldTypeString, "val1"),
			},
		},
	}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	snapshot, err := client.(*configClient).fetchModule(context.Background(), "base_cfg")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshot.ModuleKey != "base_cfg" {
		t.Fatalf("expected base_cfg, got %s", snapshot.ModuleKey)
	}
}

func TestFetchModule_InProcessError(t *testing.T) {
	svc := &mockConfigService{err: errors.New("service error")}
	client, _ := Default(WithService(svc), WithMode(ModeInProcess))
	_, err := client.(*configClient).fetchModule(context.Background(), "base_cfg")
	if err == nil {
		t.Fatal("expected error when service fails")
	}
}

func TestFetchModule_RemoteNoClient(t *testing.T) {
	client, _ := Default(WithMode(ModeRemote))
	_, err := client.(*configClient).fetchModule(context.Background(), "base_cfg")
	if err == nil {
		t.Fatal("expected error for remote mode without remote client")
	}
}

// mockRemoteClient 用于测试的模拟远程客户端
type mockRemoteClient struct {
	response []byte
	err      error
}

func (m *mockRemoteClient) Invoke(ctx context.Context, method string, request []byte) ([]byte, error) {
	return m.response, m.err
}

func TestFetchModule_RemoteWithMock(t *testing.T) {
	mockClient := &mockRemoteClient{err: errors.New("network error")}
	client, _ := Default(WithMode(ModeRemote), WithRemoteClient(mockClient))
	_, err := client.(*configClient).fetchModule(context.Background(), "base_cfg")
	if err == nil {
		t.Fatal("expected error from mock remote client")
	}
}

func TestFetchModule_NoService(t *testing.T) {
	client, err := Default(WithMode(ModeInProcess), WithService(nil))
	if err != nil {
		t.Logf("expected error when creating client with nil service: %v", err)
		return
	}
	_, err = client.(*configClient).fetchModule(context.Background(), "base_cfg")
	if !errors.Is(err, ErrServiceRequired) {
		t.Fatalf("expected ErrServiceRequired, got %v", err)
	}
}

func TestExtractModule_StaticModule(t *testing.T) {
	resp := &mockAppConfigResponse{
		staticModules: map[string]map[string]*domain.TypedValue{
			"base_cfg": {
				"name": domain.NewTypedValue(domain.FieldTypeString, "static"),
			},
		},
	}
	snapshot, _ := extractModule(resp.toReal(), "base_cfg")
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	tv := snapshot.GetField("name")
	if tv == nil || tv.String() != "static" {
		t.Fatal("expected field name=static")
	}
}

func TestExtractModule_DynamicModule(t *testing.T) {
	resp := &mockAppConfigResponse{
		dynamicModules: []mockDynamicModule{
			{moduleKey: "custom_mod", fields: map[string]*domain.TypedValue{
				"key": domain.NewTypedValue(domain.FieldTypeString, "dynamic"),
			}},
		},
	}
	snapshot, _ := extractModule(resp.toReal(), "custom_mod")
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot for dynamic module")
	}
	tv := snapshot.GetField("key")
	if tv == nil || tv.String() != "dynamic" {
		t.Fatal("expected field key=dynamic")
	}
}

type mockAppConfigResponse struct {
	staticModules  map[string]map[string]*domain.TypedValue
	dynamicModules []mockDynamicModule
}

type mockDynamicModule struct {
	moduleKey string
	fields    map[string]*domain.TypedValue
}

func (m *mockAppConfigResponse) toReal() *service.AppConfigResponse {
	resp := &service.AppConfigResponse{
		StaticModules:   m.staticModules,
		DynamicModules: []*service.DynamicModuleView{},
	}
	for _, dm := range m.dynamicModules {
		resp.DynamicModules = append(resp.DynamicModules, &service.DynamicModuleView{
			ModuleKey: dm.moduleKey,
			Fields:    dm.fields,
		})
	}
	return resp
}
