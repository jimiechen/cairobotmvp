package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

// ConfigHandler 配置 Schema API 处理器
// 负责配置字段定义的 CRUD 操作
type ConfigHandler struct {
	schemaService *service.SchemaService
}

// NewConfigHandler 创建配置 Schema 处理器实例
func NewConfigHandler(schemaService *service.SchemaService) *ConfigHandler {
	return &ConfigHandler{schemaService: schemaService}
}

// CreateSchema 新增字段定义
// POST /api/v1/config/schema
func (h *ConfigHandler) CreateSchema(c *gin.Context) {
	var req domain.FieldSchema
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	err := h.schemaService.CreateFieldSchema(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"id": req.ID,
		},
	})
}

// ListSchemas 查询模块下所有字段 Schema
// GET /api/v1/config/schema?module=xxx
func (h *ConfigHandler) ListSchemas(c *gin.Context) {
	moduleKey := c.Query("module")
	if moduleKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "module 参数必填",
			"data":    nil,
		})
		return
	}

	schemas, err := h.schemaService.ListFieldSchemas(moduleKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    schemas,
	})
}

// UpdateSchema 修改字段定义
// PUT /api/v1/config/schema/:id
func (h *ConfigHandler) UpdateSchema(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 ID",
			"data":    nil,
		})
		return
	}

	var req domain.FieldSchema
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	req.ID = id

	err = h.schemaService.UpdateFieldSchema(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    nil,
	})
}

// DeleteSchema 软删除字段（标记禁用）
// DELETE /api/v1/config/schema/:id
func (h *ConfigHandler) DeleteSchema(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 ID",
			"data":    nil,
		})
		return
	}

	err = h.schemaService.DeleteFieldSchema(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    nil,
	})
}
