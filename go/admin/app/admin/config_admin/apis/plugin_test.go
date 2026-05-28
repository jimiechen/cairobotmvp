package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
)

func init() { gin.SetMode(gin.TestMode) }

func setupTestRouter(svc configAdmin.ConfigAdminService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	schemaApi := NewSchemaApi(svc)
	valueApi := NewValueApi(svc)

	apiV1 := r.Group("/api/admin/v1/config")
	schemaGroup := apiV1.Group("/schema")
	{
		schemaGroup.GET("", schemaApi.GetSchemaList)
		schemaGroup.POST("", schemaApi.CreateSchema)
		schemaGroup.PUT("", schemaApi.UpdateSchema)
		schemaGroup.DELETE("", schemaApi.DeleteSchema)
	}
	valueGroup := apiV1.Group("/value")
	{
		valueGroup.POST("/publish", valueApi.PublishValue)
		valueGroup.GET("/versions", valueApi.GetValueVersions)
	}
	return r
}

type fakeConfigSvc struct {
	createResult *configAdmin.SchemaItem
	createErr    error
	publishErr   error
}

func (f *fakeConfigSvc) ListSchemas(_ context.Context, _ string) ([]*configAdmin.SchemaItem, error) {
	return []*configAdmin.SchemaItem{{ID: 1, FieldKey: "test"}}, nil
}
func (f *fakeConfigSvc) CreateSchema(_ context.Context, _ configAdmin.CreateSchemaRequest) (*configAdmin.SchemaItem, error) {
	return f.createResult, f.createErr
}
func (f *fakeConfigSvc) UpdateSchema(_ context.Context, _ configAdmin.UpdateSchemaRequest) (*configAdmin.SchemaItem, error) {
	return &configAdmin.SchemaItem{ID: 99}, nil
}
func (f *fakeConfigSvc) DeleteSchema(_ context.Context, _ int64, _ string) error { return nil }
func (f *fakeConfigSvc) PublishValue(_ context.Context, _ configAdmin.PublishValueRequest) (*configAdmin.ValueVersion, error) {
	return nil, f.publishErr
}

// ---- Schema API Tests ----

func TestGetSchemaList_正常请求(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/config/schema?module_key=test_mod", nil))
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String()) }
}

func TestGetSchemaList_缺少ModuleKey(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/config/schema", nil))
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400，实际=%d", w.Code) }
}

func TestCreateSchema_正常创建(t *testing.T) {
	fake := &fakeConfigSvc{createResult: &configAdmin.SchemaItem{ID: 1, FieldKey: "timeout"}}
	r := setupTestRouter(fake)
	body := `{"module_key":"mod","field_key":"timeout","field_type":"int","default_value":"30"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/config/schema", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String()) }
}

func TestCreateSchema_空Body(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/config/schema", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400，实际=%d", w.Code) }
}

func TestUpdateSchema_正常更新(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	body := `{"id":1,"field_type":"string","default_value":"new_val"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/config/schema", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String()) }
}

func TestDeleteSchema_无效ID(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/admin/v1/config/schema?id=abc", nil))
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400，实际=%d", w.Code) }
}

func TestDeleteSchema_正常删除(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/admin/v1/config/schema?id=42", nil))
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d", w.Code) }
}

// ---- Value API Tests ----

func TestPublishValue_正常发布(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	body := `{"module_key":"mod","env":"dev","fields":[{"field_key":"port","value":8080}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/config/value/publish", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String()) }
}

func TestPublishValue_校验错误返回10400(t *testing.T) {
	fake := &fakeConfigSvc{publishErr: &configAdmin.ValidationError{}}
	r := setupTestRouter(fake)
	body := `{"module_key":"mod","env":"dev","fields":[{"field_key":"port","value":99999}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/config/value/publish", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400（10400），实际=%d body=%s", w.Code, w.Body.String()) }
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if int(resp["code"].(float64)) != 10400 { t.Errorf("响应 code 应为 10400，实际=%v", resp["code"]) }
}

func TestPublishValue_空Fields(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	body := `{"module_key":"mod","env":"dev","fields":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/config/value/publish", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400，实际=%d", w.Code) }
}

func TestGetValueVersions_缺ModuleKey(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/config/value/versions?env=dev", nil))
	if w.Code != http.StatusBadRequest { t.Errorf("期望 400，实际=%d", w.Code) }
}

func TestGetValueVersions_正常查询(t *testing.T) {
	r := setupTestRouter(&fakeConfigSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/config/value/versions?module_key=mod&env=dev", nil))
	if w.Code != http.StatusOK { t.Errorf("期望 200，实际=%d", w.Code) }
	if !strings.Contains(w.Body.String(), `"versions"`) { t.Errorf("应含 versions 字段，body=%s", w.Body.String()) }
}
