package apis

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
)

func init() { gin.SetMode(gin.TestMode) }

func setupI18nTestRouter(svc i18nAdmin.I18nAdminService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	stringApi := NewStringApi(svc)
	packApi := NewPackApi(svc)
	importApi := NewImportExportApi(svc)

	apiV1 := r.Group("/api/admin/v1/i18n")
	strGroup := apiV1.Group("/string")
	{
		strGroup.GET("", stringApi.ListStrings)
		strGroup.POST("", stringApi.CreateString)
		strGroup.PUT("", stringApi.UpdateString)
		strGroup.DELETE("", stringApi.DeleteString)
	}
	packGroup := apiV1.Group("/pack")
	{
		packGroup.POST("/publish", packApi.PublishPack)
		packGroup.POST("/rollback", packApi.RollbackPack)
	}
	importGroup := apiV1.Group("")
	{
		importGroup.POST("/import/csv", importApi.ImportStringsFromCSV)
		importGroup.GET("/export/csv", importApi.ExportStringsToCSV)
	}
	return r
}

// ---- Fake 实现 I18nAdminService 接口（测试用）----

type fakeI18nSvc struct {
	createResult *i18nAdmin.StringItem
	createErr    error
	updateResult *i18nAdmin.StringItem
	updateErr    error
	deleteErr    error
	listResult   []*i18nAdmin.StringItem
	listErr      error
	publishResult *i18nAdmin.PackVersion
	publishErr   error
	rollbackErr  error
	importResult *i18nAdmin.ImportResult
	importErr    error
	exportData   []byte
	exportErr    error
}

func (f *fakeI18nSvc) CreateString(_ context.Context, _ i18nAdmin.CreateStringRequest) (*i18nAdmin.StringItem, error) {
	return f.createResult, f.createErr
}

func (f *fakeI18nSvc) UpdateString(_ context.Context, _ i18nAdmin.UpdateStringRequest) (*i18nAdmin.StringItem, error) {
	return f.updateResult, f.updateErr
}

func (f *fakeI18nSvc) DeleteString(_ context.Context, _ int64, _ string) error {
	return f.deleteErr
}

func (f *fakeI18nSvc) ListStrings(_ int64) ([]*i18nAdmin.StringItem, error) {
	return f.listResult, f.listErr
}

func (f *fakeI18nSvc) PublishPack(_ context.Context, _ i18nAdmin.PublishPackRequest) (*i18nAdmin.PackVersion, error) {
	return f.publishResult, f.publishErr
}

func (f *fakeI18nSvc) RollbackPack(_ context.Context, _ int64, _ int, _ string) error {
	return f.rollbackErr
}

func (f *fakeI18nSvc) ImportStringsFromCSV(_ context.Context, _ interface{}, _ int64, _ string) (*i18nAdmin.ImportResult, error) {
	return f.importResult, f.importErr
}

func (f *fakeI18nSvc) ExportStringsToCSV(_ context.Context, _ i18nAdmin.ExportCSVRequest) ([]byte, error) {
	return f.exportData, f.exportErr
}

// ==================== String API Tests ====================

func TestCreateString_正常创建(t *testing.T) {
	fake := &fakeI18nSvc{createResult: &i18nAdmin.StringItem{ID: 1, StringKey: "greeting.hello"}}
	r := setupI18nTestRouter(fake)
	body := `{"pack_id":1,"string_key":"greeting.hello","string_value":"你好","template_type":"plain"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/string", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"StringKey":"greeting.hello"`) {
		t.Errorf("应包含创建的 string_key，body=%s", w.Body.String())
	}
}

func TestCreateString_缺少必填字段(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/string", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestCreateString_模板校验错误返回10400(t *testing.T) {
	fake := &fakeI18nSvc{createErr: fmt.Errorf("模板校验失败: 占位符不匹配")}
	r := setupI18nTestRouter(fake)
	body := `{"pack_id":1,"string_key":"greeting","string_value":"Hello {name}","template_type":"named"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/string", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400（10400），实际=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	jsonUnmarshal(t, w.Body.Bytes(), &resp)
	if int(resp["code"].(float64)) != 10400 {
		t.Errorf("响应 code 应为 10400，实际=%v", resp["code"])
	}
}

func TestUpdateString_正常更新(t *testing.T) {
	fake := &fakeI18nSvc{updateResult: &i18nAdmin.StringItem{ID: 42}}
	r := setupI18nTestRouter(fake)
	body := `{"id":42,"string_value":"更新后的值"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/i18n/string", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateString_空Body(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/i18n/string", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestDeleteString_无效ID(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/admin/v1/i18n/string?id=abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestDeleteString_正常删除(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/admin/v1/i18n/string?id=99", nil))
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d", w.Code)
	}
}

func TestListStrings_正常查询(t *testing.T) {
	fake := &fakeI18nSvc{listResult: []*i18nAdmin.StringItem{
		{ID: 1, StringKey: "hello"},
		{ID: 2, StringKey: "bye"},
	}}
	r := setupI18nTestRouter(fake)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/i18n/string?pack_id=1", nil))
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"StringKey":"hello"`) {
		t.Errorf("应包含字符串列表数据，body=%s", w.Body.String())
	}
}

func TestListStrings_缺少PackID(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/i18n/string", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

// ==================== Pack API Tests ====================

func TestPublishPack_正常发布(t *testing.T) {
	fake := &fakeI18nSvc{publishResult: &i18nAdmin.PackVersion{PackID: 1, LangCode: "zh-CN", Version: 5}}
	r := setupI18nTestRouter(fake)
	body := `{"pack_id":1,"lang_code":"zh-CN","env":"prod"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/pack/publish", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPublishPack_空Body(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/pack/publish", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestRollbackPack_正常回滚(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	body := `{"pack_id":1,"target_version":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/pack/rollback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRollbackPack_空Body(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/pack/rollback", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

// ==================== Import/Export API Tests ====================

func TestImportCSV_正常导入(t *testing.T) {
	csvContent := buildCSV([][]string{{"greeting.hello", "你好", "common", "plain"}})
	fake := &fakeI18nSvc{importResult: &i18nAdmin.ImportResult{
		TotalRows:    1,
		SuccessCount: 1,
		FailCount:    0,
	}}
	r := setupI18nTestRouter(fake)
	req := createMultipartImportReq(csvContent, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	jsonUnmarshal(t, w.Body.Bytes(), &resp)
	if sc, ok := resp["success_count"]; ok {
		if int(sc.(float64)) != 1 {
			t.Errorf("success_count 应为 1，实际=%v", sc)
		}
	}
}

func TestImportCSV_部分失败返回10400(t *testing.T) {
	csvContent := buildCSV([][]string{{"k1", "v1"}, {"k2", "v2"}})
	fake := &fakeI18nSvc{importResult: &i18nAdmin.ImportResult{
		TotalRows:    2,
		SuccessCount: 1,
		FailCount:    1,
		Errors: []i18nAdmin.ImportError{{RowNum: 2, Reason: "模板格式错误"}},
	}}
	r := setupI18nTestRouter(fake)
	req := createMultipartImportReq(csvContent, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	jsonUnmarshal(t, w.Body.Bytes(), &resp)
	if int(resp["code"].(float64)) != 10400 {
		t.Errorf("响应 code 应为 10400（部分失败），实际=%v", resp["code"])
	}
}

func TestImportCSV_缺少文件(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/import/csv?pack_id=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestImportCSV_无效PackID(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/admin/v1/i18n/import/csv?pack_id=abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

func TestExportCSV_正常导出(t *testing.T) {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll([][]string{
		{"string_key", "string_value", "group_name", "template_type"},
		{"greeting.hello", "你好", "common", "plain"},
	})
	w.Flush()

	fake := &fakeI18nSvc{exportData: buf.Bytes()}
	r := setupI18nTestRouter(fake)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/admin/v1/i18n/export/csv?pack_id=1", nil))
	if w2.Code != http.StatusOK {
		t.Errorf("期望 200，实际=%d body=%s", w2.Code, w2.Body.String())
	}
	contentType := w2.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("Content-Type 应含 text/csv，实际=%s", contentType)
	}
	disposition := w2.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, ".csv") {
		t.Errorf("Content-Disposition 应含 .csv，实际=%s", disposition)
	}
}

func TestExportCSV_无效PackID(t *testing.T) {
	r := setupI18nTestRouter(&fakeI18nSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/admin/v1/i18n/export/csv?pack_id=abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400，实际=%d", w.Code)
	}
}

// ==================== 辅助函数 ====================

func jsonUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("JSON 解析失败: %v", err)
	}
}

func buildCSV(rows [][]string) io.Reader {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.WriteAll(rows)
	w.Flush()
	return buf
}

func createMultipartImportReq(csvBody io.Reader, packID int64) *http.Request {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	io.Copy(part, csvBody)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost,
		"/api/admin/v1/i18n/import/csv?pack_id="+intToStr(packID), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func intToStr(n int64) string {
	if n == 0 { return "0" }
	s := ""
	for n > 0 {
		s = string('0'+byte(n%10)) + s
		n /= 10
	}
	return s
}
