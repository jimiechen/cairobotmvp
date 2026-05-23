package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandler()
	router := gin.New()
	router.GET("/api/v1/health", handler.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
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
