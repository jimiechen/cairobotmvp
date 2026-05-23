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
	cacheInvalidate CacheInvalidator
}

// CacheInvalidator 缓存失效接口
// 用于在写操作后通知 SDK 层失效本地缓存
//
// TODO(M-05): 实现 Redis pub/sub 失效广播
// 未来应通过 Redis PUBLISH 命令发送以下消息：
//   - 主题: "cairobot.config.invalidate"
//   - 消息体: {"module_key": "xxx", "action": "create|update|delete", "timestamp": ...}
//
// 实现步骤：
//   1. 在 provider-admin 初始化时创建 Redis 连接
//   2. 实现此接口的 Redis 版本：RedisCacheInvalidator
//   3. 在每个写操作成功后调用 InvalidateConfigCache()
//   4. SDK 层订阅 "cairobot.config.invalidate" 主题，收到消息后清空 LRU 缓存
//
// 当前行为：no-op（空实现），SDK 缓存依赖 TTL 过期（30s-10min）
// 目标行为：热更新延迟 <100ms（通过 pub/sub 即时推送）
//
// 参见评审报告: docs/reviews/review-config-i18n-implementation.md#M-05
type CacheInvalidator interface {
	InvalidateConfigCache(moduleKey string, action string)
}

// NoopCacheInvalidator 空操作的缓存失效器（MVP 默认实现）
type NoopCacheInvalidator struct{}

func (n *NoopCacheInvalidator) InvalidateConfigCache(moduleKey string, action string) {

}

// NewConfigHandler 创建配置 Schema 处理器实例
func NewConfigHandler(schemaService *service.SchemaService) *ConfigHandler {
	return &ConfigHandler{
		schemaService:    schemaService,
		cacheInvalidate: &NoopCacheInvalidator{},
	}
}

// NewConfigHandlerWithCache 创建带缓存失效能力的处理器（未来 Redis 集成时使用）
func NewConfigHandlerWithCache(schemaService *service.SchemaService, cacheInvalidate CacheInvalidator) *ConfigHandler {
	return &ConfigHandler{
		schemaService:    schemaService,
		cacheInvalidate: cacheInvalidate,
	}
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

	h.cacheInvalidate.InvalidateConfigCache(req.ModuleKey, "create")

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
		"data": schemas,
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

	h.cacheInvalidate.InvalidateConfigCache(req.ModuleKey, "update")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": nil,
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

	h.cacheInvalidate.InvalidateConfigCache("", "delete")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": nil,
	})
}
