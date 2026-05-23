package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

func setupConfigValueRouter() (*gin.Engine, *repository.MemConfigRepo) {
	gin.SetMode(gin.TestMode)

	memRepo := repository.NewMemConfigRepo()
	handler := NewConfigValueHandler(memRepo)

	router := gin.New()
	configGroup := router.Group("/api/v1/config")
	{
		configGroup.GET("/value/:env", handler.GetPublishedValues)
		configGroup.PUT("/value", handler.UpdateConfigValue)
		configGroup.POST("/value/publish", handler.PublishConfig)
	}

	return router, memRepo
}

func TestConfigValueHandler_GetPublishedValues(t *testing.T) {
	router, memRepo := setupConfigValueRouter()

	now := time.Now()
	memRepo.Save(&domain.ConfigVersion{
		ID:          1,
		ModuleKey:   "base_cfg",
		Env:         "dev",
		Version:     1,
		ConfigJSON:  `{"app_name":"CaiRobot"}`,
		IsPublished: true,
		PublishedAt: &now,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/value/dev", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
		t.Logf("响应体: %s", w.Body.String())
	}
}

func TestConfigValueHandler_UpdateConfigValue(t *testing.T) {
	router, _ := setupConfigValueRouter()

	reqBody := map[string]interface{}{
		"module_key":  "base_cfg",
		"env":        "dev",
		"config_json": `{"app_name":"CaiRobot"}`,
		"create_by":   "admin",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/config/value", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
		t.Logf("响应体: %s", w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	if data["id"] == nil || data["id"].(float64) == 0 {
		t.Errorf("期望返回有效的 id，实际 %v", data["id"])
	}
}

func TestConfigValueHandler_UpdateConfigValue_MissingFields(t *testing.T) {
	router, _ := setupConfigValueRouter()

	reqBody := map[string]interface{}{
		"module_key": "base_cfg",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/config/value", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d（参数验证失败），实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestConfigValueHandler_PublishConfig(t *testing.T) {
	router, memRepo := setupConfigValueRouter()

	memRepo.Save(&domain.ConfigVersion{
		ID:         1,
		ModuleKey:  "base_cfg",
		Env:        "dev",
		Version:    1,
		ConfigJSON: `{"app_name":"CaiRobot"}`,
	})

	reqBody := map[string]interface{}{
		"module_key": "base_cfg",
		"env":       "dev",
		"version":   float64(1),
		"update_by": "admin",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/value/publish", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
		t.Logf("响应体: %s", w.Body.String())
	}
}

func TestConfigValueHandler_PublishConfig_NotFound(t *testing.T) {
	router, _ := setupConfigValueRouter()

	reqBody := map[string]interface{}{
		"module_key": "base_cfg",
		"env":       "dev",
		"version":   float64(999),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/value/publish", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d（未找到），实际 %d", http.StatusNotFound, w.Code)
	}
}
