package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "admin.db")

	server, err := NewHTTPServer(dbPath)
	if err != nil {
		t.Fatalf("创建 HTTP 服务器失败: %v", err)
	}
	defer server.Close()

	if server.Engine() == nil {
		t.Error("期望 Engine 不为 nil")
	}
}

func TestHTTPServer_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "admin.db")

	server, err := NewHTTPServer(dbPath)
	if err != nil {
		t.Fatalf("创建 HTTP 服务器失败: %v", err)
	}
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	server.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if response["code"].(float64) != 0 {
		t.Errorf("期望 code=0，实际 %v", response["code"])
	}

	data := response["data"].(map[string]interface{})
	if data["status"] != "UP" {
		t.Errorf("期望 status=UP，实际 %v", data["status"])
	}
}

func TestHTTPServer_ConfigSchemaEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "admin.db")

	server, err := NewHTTPServer(dbPath)
	if err != nil {
		t.Fatalf("创建 HTTP 服务器失败: %v", err)
	}
	defer server.Close()

	engine := server.Engine()

	testCases := []struct {
		name       string
		method     string
		path       string
		body       string
		expectCode int
	}{
		{
			name:       "ListSchemas 缺少 module 参数",
			method:     http.MethodGet,
			path:       "/api/v1/config/schema",
			expectCode: http.StatusBadRequest,
		},
		{
			name:   "CreateSchema 无效 JSON",
			method: http.MethodPost,
			path:   "/api/v1/config/schema",
			body:   `{invalid}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "UpdateSchema 无效 ID",
			method:     http.MethodPut,
			path:       "/api/v1/config/schema/abc",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "DeleteSchema 无效 ID",
			method:     http.MethodDelete,
			path:       "/api/v1/config/schema/abc",
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != tc.expectCode {
				t.Errorf("期望状态码 %d，实际 %d，响应: %s", tc.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestHTTPServer_I18nEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "admin.db")

	server, err := NewHTTPServer(dbPath)
	if err != nil {
		t.Fatalf("创建 HTTP 服务器失败: %v", err)
	}
	defer server.Close()

	engine := server.Engine()

	testCases := []struct {
		name       string
		method     string
		path       string
		expectCode int
	}{
		{
			name:       "GetDiff 缺少 lang 参数",
			method:     http.MethodGet,
			path:       "/api/v1/i18n/diff",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "CreatePack 无效请求体",
			method:     http.MethodPost,
			path:       "/api/v1/i18n/pack",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "CreateString 无效请求体",
			method:     http.MethodPost,
			path:       "/api/v1/i18n/string",
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != tc.expectCode {
				t.Errorf("期望状态码 %d，实际 %d", tc.expectCode, w.Code)
			}
		})
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
