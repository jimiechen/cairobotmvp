package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func TestGetModule_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"app_name": domain.NewTypedValue(domain.FieldTypeString, "test_app"),
			"version":  domain.NewTypedValue(domain.FieldTypeInt, int64(1)),
		},
	})
	snapshot, err := client.GetModule(context.Background(), "base_cfg")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshot.ModuleKey != "base_cfg" {
		t.Fatalf("expected module_key base_cfg, got %s", snapshot.ModuleKey)
	}
	if len(snapshot.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(snapshot.Fields))
	}
}

func TestGetModule_NotFound(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{})
	snapshot, err := client.GetModule(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for empty module, got %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(snapshot.Fields) != 0 {
		t.Fatalf("expected 0 fields, got %d", len(snapshot.Fields))
	}
}

type testBindStruct struct {
	AppName string `config:"app_name"`
	Version int64  `config:"version"`
	Enabled bool   `config:"enabled"`
}

func TestBind_Success(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"app_name": domain.NewTypedValue(domain.FieldTypeString, "bound_app"),
			"version":  domain.NewTypedValue(domain.FieldTypeInt, int64(42)),
			"enabled":  domain.NewTypedValue(domain.FieldTypeBool, true),
		},
	})
	var out testBindStruct
	err := client.Bind(context.Background(), "base_cfg", &out)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.AppName != "bound_app" {
		t.Fatalf("expected AppName bound_app, got %s", out.AppName)
	}
	if out.Version != 42 {
		t.Fatalf("expected Version 42, got %d", out.Version)
	}
	if !out.Enabled {
		t.Fatal("expected Enabled true")
	}
}

type testAutoBindStruct struct {
	AppName string
	Version int64
}

func TestBind_AutoSnakeCase(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{
		"base_cfg": {
			"app_name": domain.NewTypedValue(domain.FieldTypeString, "auto_app"),
			"version":  domain.NewTypedValue(domain.FieldTypeInt, int64(99)),
		},
	})
	var out testAutoBindStruct
	err := client.Bind(context.Background(), "base_cfg", &out)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.AppName != "auto_app" {
		t.Fatalf("expected AppName auto_app, got %s", out.AppName)
	}
	if out.Version != 99 {
		t.Fatalf("expected Version 99, got %d", out.Version)
	}
}

func TestBind_NonPointer(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{})
	var out testBindStruct
	err := client.Bind(context.Background(), "base_cfg", out)
	if !errors.Is(err, ErrBindFailed) {
		t.Fatalf("expected ErrBindFailed for non-pointer, got %v", err)
	}
}

func TestBind_NilPointer(t *testing.T) {
	client, _ := newTestClient(map[string]map[string]*domain.TypedValue{})
	err := client.Bind(context.Background(), "base_cfg", nil)
	if !errors.Is(err, ErrBindFailed) {
		t.Fatalf("expected ErrBindFailed for nil pointer, got %v", err)
	}
}

func TestModuleSnapshot_GetField(t *testing.T) {
	snapshot := &ModuleSnapshot{
		ModuleKey: "test",
		Fields: map[string]*domain.TypedValue{
			"key1": domain.NewTypedValue(domain.FieldTypeString, "val1"),
		},
	}
	tv := snapshot.GetField("key1")
	if tv == nil {
		t.Fatal("expected non-nil TypedValue")
	}
	if tv.String() != "val1" {
		t.Fatalf("expected val1, got %s", tv.String())
	}
	tv = snapshot.GetField("nonexistent")
	if tv != nil {
		t.Fatal("expected nil for nonexistent field")
	}
}
