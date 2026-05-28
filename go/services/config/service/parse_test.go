package service

import (
	"encoding/json"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func newMockSchemaRepo() repository.SchemaRepository {
	return &mockSchemaRepoImpl{schemas: make(map[string][]*domain.FieldSchema)}
}

type mockSchemaRepoImpl struct {
	schemas map[string][]*domain.FieldSchema
}

func (m *mockSchemaRepoImpl) ListByModule(moduleKey string) ([]*domain.FieldSchema, error) {
	return m.schemas[moduleKey], nil
}

func (m *mockSchemaRepoImpl) Create(schema *domain.FieldSchema) error   { return nil }
func (m *mockSchemaRepoImpl) Update(schema *domain.FieldSchema) error   { return nil }
func (m *mockSchemaRepoImpl) DeleteSoft(id int64) error                { return nil }
func (m *mockSchemaRepoImpl) FindSchema(_ int64) (*domain.FieldSchema, error) {
	return nil, nil
}

func TestParseConfigJSON_带schema解析(t *testing.T) {
	repo := newMockSchemaRepo()
	repo.(*mockSchemaRepoImpl).schemas["test_mod"] = []*domain.FieldSchema{
		{FieldKey: "name", FieldType: domain.FieldTypeString},
		{FieldKey: "count", FieldType: domain.FieldTypeInt},
		{FieldKey: "active", FieldType: domain.FieldTypeBool},
	}

	result, err := ParseConfigJSON(`{"name":"hello","count":42,"active":true}`, "test_mod", repo)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if result["name"].String() != "hello" {
		t.Errorf("name 值错误: %s", result["name"].String())
	}
	if result["count"].Int() != 42 {
		t.Errorf("count 值错误: %d", result["count"].Int())
	}
	if !result["active"].Bool() {
		t.Error("active 应为 true")
	}
}

func TestParseConfigJSON_无schema推断类型(t *testing.T) {
	repo := newMockSchemaRepo()

	result, err := ParseConfigJSON(`{"str":"val","num":3.14,"flag":false}`, "unknown_mod", repo)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if result["str"].Type != domain.FieldTypeString {
		t.Error("字符串推断失败")
	}
	if result["num"].Type != domain.FieldTypeFloat {
		t.Error("浮点数推断失败")
	}
	if result["flag"].Type != domain.FieldTypeBool {
		t.Error("布尔值推断失败")
	}
}

func TestParseConfigJSON_非法JSON应报错(t *testing.T) {
	repo := newMockSchemaRepo()

	_, err := ParseConfigJSON(`{invalid json`, "mod", repo)
	if err == nil {
		t.Error("非法 JSON 应返回 error")
	}
}

func TestParseConfigJSON_JSON字段保留原始结构(t *testing.T) {
	repo := newMockSchemaRepo()
	repo.(*mockSchemaRepoImpl).schemas["json_mod"] = []*domain.FieldSchema{
		{FieldKey: "meta", FieldType: domain.FieldTypeJSON},
	}

	result, _ := ParseConfigJSON(`{"meta":{"k":"v"}}`, "json_mod", repo)
	raw := result["meta"].JSON()
	var m map[string]string
	json.Unmarshal(raw, &m)
	if m["k"] != "v" {
		t.Error("JSON 字段内容丢失")
	}
}

func TestCoerceInt64_各类型转换(t *testing.T) {
	testCases := []struct {
		input    any
		expected int64
	}{
		{float64(42), 42},
		{int64(99), 99},
		{int(7), 7},
		{json.Number("100"), 100},
		{nil, 0},
		{"abc", 0},
	}
	for _, tc := range testCases {
		got := coerceInt64(tc.input)
		if got != tc.expected {
			t.Errorf("coerceInt64(%v) = %d, 期望 %d", tc.input, got, tc.expected)
		}
	}
}
