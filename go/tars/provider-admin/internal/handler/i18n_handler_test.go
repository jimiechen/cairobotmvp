package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

func setupI18nRouter() (*gin.Engine, *AdminSQLiteRepo) {
	gin.SetMode(gin.TestMode)

	repo := NewAdminSQLiteRepo(nil)
	handler := NewI18nHandler(repo, nil)

	router := gin.New()
	i18nGroup := router.Group("/api/v1/i18n")
	{
		i18nGroup.POST("/pack", handler.CreatePack)
		i18nGroup.GET("/pack/:lang_code", handler.GetPack)
		i18nGroup.POST("/string", handler.CreateString)
		i18nGroup.PUT("/string/:id", handler.UpdateString)
		i18nGroup.DELETE("/string/:id", handler.DeleteString)
		i18nGroup.GET("/diff", handler.GetDiff)
		i18nGroup.POST("/publish", handler.PublishLangPack)
	}

	return router, repo
}

func TestI18nHandler_CreatePack(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"pack_name":  "中文语言包",
		"env":        "dev",
		"version":    float64(1),
		"lang_code":  "zh-CN",
		"description": "简体中文语言包",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/i18n/pack", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_CreatePack_MissingFields(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"pack_name": "中文语言包",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/i18n/pack", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化或参数缺失），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_GetPack(t *testing.T) {
	router, _ := setupI18nRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/i18n/pack/zh-CN?env=dev", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_CreateString(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"pack_id":       float64(1),
		"string_key":    "svc_msg_welcome",
		"string_value":  "欢迎 {name}，你有 {count} 条新消息",
		"group_name":    "app",
		"template_type": "named",
		"params_schema": `[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/i18n/string", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_CreateString_InvalidTemplate(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"pack_id":       float64(1),
		"string_key":    "test_key",
		"string_value":  "欢迎 {name}",
		"template_type": "plain",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/i18n/string", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d（模板校验失败），实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestI18nHandler_UpdateString(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"string_value": "更新后的消息内容",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/i18n/string/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_DeleteString(t *testing.T) {
	router, _ := setupI18nRouter()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/i18n/string/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestI18nHandler_GetDiff_MissingLang(t *testing.T) {
	router, _ := setupI18nRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/i18n/diff", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d（缺少 lang 参数），实际 %d", http.StatusBadRequest, w.Code)
	}
}

func TestI18nHandler_PublishLangPack(t *testing.T) {
	router, _ := setupI18nRouter()

	reqBody := map[string]interface{}{
		"pack_id":      float64(1),
		"published_by": float64(1),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/i18n/publish", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d（db 未初始化），实际 %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAdminI18nRepo_CreatePack(t *testing.T) {
	repo := &AdminSQLiteRepo{}

	pack := &domain.LangPack{
		PackName:    "测试包",
		Env:         "dev",
		Version:     1,
		LangCode:    "zh-CN",
		Description: "测试描述",
		IsPublished: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreatePack(pack)
	if err == nil {
		t.Error("期望返回错误（db 为 nil）")
	}
}

func TestAdminI18nRepo_CreateString(t *testing.T) {
	repo := &AdminSQLiteRepo{}

	s := &domain.LangString{
		PackID:      1,
		StringKey:   domain.StringKey("test_key"),
		StringValue: "测试值",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateString(s)
	if err == nil {
		t.Error("期望返回错误（db 为 nil）")
	}
}
