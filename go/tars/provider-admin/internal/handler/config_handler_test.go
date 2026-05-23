package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	configService "github.com/jimiechen/mineplanet/go/services/config/service"
)

func setupConfigRouter() (*gin.Engine, *repository.MemSchemaRepo) {
	gin.SetMode(gin.TestMode)

	memRepo := repository.NewMemSchemaRepo()
	schemaSvc := configService.NewSchemaService(memRepo)
	handler := NewConfigHandler(schemaSvc)

	router := gin.New()
	configGroup := router.Group("/api/v1/config")
	{
		configGroup.POST("/schema", handler.CreateSchema)
		configGroup.GET("/schema", handler.ListSchemas)
		configGroup.PUT("/schema/:id", handler.UpdateSchema)
		configGroup.DELETE("/schema/:id", handler.DeleteSchema)
	}

	return router, memRepo
}

func TestConfigHandler_CreateSchema(t *testing.T) {
	router, _ := setupConfigRouter()

	reqBody := domain.FieldSchema{
		ModuleKey:    "base_cfg",
		FieldKey:     "app_name",
		FieldType:    domain.FieldTypeString,
		DefaultValue: "CaiRobot",
		Description:  "应用名称",
		ClientScope:  "all",
		IsRequired:   true,
		SortOrder:    10,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/schema", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
		t.Logf("响应体: %s", w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["code"].(float64) != 0 {
		t.Errorf("期望 code=0，实际 %v", response["code"])
	}
}

func TestConfigHandler_CreateSchema_InvalidInput(t *testing.T) {
	router, _ := setupConfigRouter()

	reqBody := domain.FieldSchema{
		FieldKey: "app_name",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/schema", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（验证失败），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestConfigHandler_ListSchemas(t *testing.T) {
	router, memRepo := setupConfigRouter()

	memRepo.Create(&domain.FieldSchema{
		ID:          1,
		ModuleKey:   "base_cfg",
		FieldKey:    "app_name",
		FieldType:   domain.FieldTypeString,
		Description: "应用名称",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/schema?module=base_cfg", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}
}

func TestConfigHandler_ListSchemas_MissingModule(t *testing.T) {
	router, _ := setupConfigRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/schema", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestConfigHandler_UpdateSchema(t *testing.T) {
	router, memRepo := setupConfigRouter()

	memRepo.Create(&domain.FieldSchema{
		ID:        1,
		ModuleKey: "base_cfg",
		FieldKey:  "app_name",
	})

	reqBody := domain.FieldSchema{
		Description: "更新后的描述",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/config/schema/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
		t.Logf("响应体: %s", w.Body.String())
	}
}

func TestConfigHandler_DeleteSchema(t *testing.T) {
	router, memRepo := setupConfigRouter()

	memRepo.Create(&domain.FieldSchema{
		ID:        1,
		ModuleKey: "base_cfg",
		FieldKey:  "app_name",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/config/schema/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}
}
