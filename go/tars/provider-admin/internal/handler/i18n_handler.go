package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
	i18nsdk "github.com/jimiechen/mineplanet/go/services/i18n/sdk"
)

// AdminI18nRepository 管理后台专用的 i18n 仓库接口
// 扩展基础 I18nRepository，增加写操作能力
type AdminI18nRepository interface {
	repository.I18nRepository
	CreatePack(pack *domain.LangPack) error
	CreateString(s *domain.LangString) error
	UpdateString(s *domain.LangString) error
	DeleteString(id int64) error
	PublishPack(packID int64, publishedBy int64) error
}

// I18nHandler 多语言 API 处理器
// 负责语言包和字符串的 CRUD 操作、发布和预览
type I18nHandler struct {
	repo             AdminI18nRepository
	client           i18nsdk.Client
	cacheInvalidate I18nCacheInvalidator
}

// I18nCacheInvalidator I18n 缓存失效接口
// 用于在写操作后通知 SDK 层失效语言包缓存
//
// TODO(M-05): 实现 Redis pub/sub 失效广播
// 未来应通过 Redis PUBLISH 命令发送以下消息：
//   - 主题: "cairobot.i18n.invalidate"
//   - 消息体: {"lang_code": "xxx", "action": "create|update|delete|publish", "timestamp": ...}
//
// 参见评审报告: docs/reviews/review-config-i18n-implementation.md#M-05
type I18nCacheInvalidator interface {
	InvalidateI18nCache(langCode string, action string)
}

// NoopI18nCacheInvalidator 空操作的 I18n 缓存失效器（MVP 默认实现）
type NoopI18nCacheInvalidator struct{}

func (n *NoopI18nCacheInvalidator) InvalidateI18nCache(langCode string, action string) {

}

// NewI18nHandler 创建多语言处理器实例
func NewI18nHandler(repo AdminI18nRepository, client i18nsdk.Client) *I18nHandler {
	return &I18nHandler{
		repo:             repo,
		client:           client,
		cacheInvalidate: &NoopI18nCacheInvalidator{},
	}
}

// NewI18nHandlerWithCache 创建带缓存失效能力的处理器（未来 Redis 集成时使用）
func NewI18nHandlerWithCache(repo AdminI18nRepository, client i18nsdk.Client, cacheInvalidate I18nCacheInvalidator) *I18nHandler {
	return &I18nHandler{
		repo:             repo,
		client:           client,
		cacheInvalidate: cacheInvalidate,
	}
}

// CreatePack 创建/更新语言包
// POST /api/v1/i18n/pack
func (h *I18nHandler) CreatePack(c *gin.Context) {
	var req domain.LangPack
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	err := h.repo.CreatePack(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.cacheInvalidate.InvalidateI18nCache(req.LangCode, "create")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"id": req.ID,
		},
	})
}

// GetPack 获取语言包信息
// GET /api/v1/i18n/pack/:lang_code
func (h *I18nHandler) GetPack(c *gin.Context) {
	langCode := c.Param("lang_code")
	env := c.DefaultQuery("env", "dev")

	pack, err := h.repo.GetPackByLangCode(langCode, env)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	if pack == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "语言包不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    pack,
	})
}

// CreateString 新增多语言 key
// POST /api/v1/i18n/string
func (h *I18nHandler) CreateString(c *gin.Context) {
	var req struct {
		PackID       int64  `json:"pack_id" binding:"required"`
		StringKey    string `json:"string_key" binding:"required"`
		StringValue  string `json:"string_value" binding:"required"`
		GroupName    string `json:"group_name"`
		TemplateType string `json:"template_type"`
		ParamsSchema string `json:"params_schema"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	templateType := domain.TemplateType(req.TemplateType)
	if templateType != "" && !templateType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板类型",
			"data":    nil,
		})
		return
	}

	params, _ := domain.ParseParamsSchema(req.ParamsSchema)
	if templateType != "" && req.StringValue != "" {
		if err := domain.ValidateTemplate(req.StringValue, templateType, params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "模板校验失败: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	s := &domain.LangString{
		PackID:       req.PackID,
		StringKey:    domain.StringKey(req.StringKey),
		StringValue:  req.StringValue,
		GroupName:    req.GroupName,
		TemplateType: templateType,
		ParamsSchema: req.ParamsSchema,
	}

	err := h.repo.CreateString(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.cacheInvalidate.InvalidateI18nCache("", "create_string")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"id": s.ID,
		},
	})
}

// UpdateString 修改多语言 key
// PUT /api/v1/i18n/string/:id
func (h *I18nHandler) UpdateString(c *gin.Context) {
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

	var req struct {
		StringValue  string `json:"string_value" binding:"required"`
		GroupName    string `json:"group_name"`
		TemplateType string `json:"template_type"`
		ParamsSchema string `json:"params_schema"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	templateType := domain.TemplateType(req.TemplateType)
	if templateType != "" && !templateType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的模板类型",
			"data":    nil,
		})
		return
	}

	params, _ := domain.ParseParamsSchema(req.ParamsSchema)
	if templateType != "" && req.StringValue != "" {
		if err := domain.ValidateTemplate(req.StringValue, templateType, params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "模板校验失败: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	s := &domain.LangString{
		ID:           id,
		StringValue:  req.StringValue,
		GroupName:    req.GroupName,
		TemplateType: templateType,
		ParamsSchema: req.ParamsSchema,
	}

	err = h.repo.UpdateString(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.cacheInvalidate.InvalidateI18nCache("", "update_string")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    nil,
	})
}

// DeleteString 删除多语言 key
// DELETE /api/v1/i18n/string/:id
func (h *I18nHandler) DeleteString(c *gin.Context) {
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

	err = h.repo.DeleteString(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.cacheInvalidate.InvalidateI18nCache("", "delete_string")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    nil,
	})
}

// GetDiff 获取增量变更
// GET /api/v1/i18n/diff?lang=xx&since=N
func (h *I18nHandler) GetDiff(c *gin.Context) {
	langCode := c.Query("lang")
	if langCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "lang 参数必填",
			"data":    nil,
		})
		return
	}

	sinceStr := c.DefaultQuery("since", "0")
	since, err := strconv.Atoi(sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的 since 参数",
			"data":    nil,
		})
		return
	}

	env := c.DefaultQuery("env", "dev")

	pack, err := h.repo.GetPackByLangCode(langCode, env)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询语言包失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	if pack == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "语言包不存在",
			"data":    nil,
		})
		return
	}

	strings, err := h.repo.GetDiffSince(pack.ID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询增量失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"current_version": pack.Version,
			"strings":         strings,
		},
	})
}

// PublishLangPack 发布语言包版本
// POST /api/v1/i18n/publish
func (h *I18nHandler) PublishLangPack(c *gin.Context) {
	var req struct {
		PackID      int64 `json:"pack_id" binding:"required"`
		PublishedBy int64 `json:"published_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	err := h.repo.PublishPack(req.PackID, req.PublishedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "发布失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	h.cacheInvalidate.InvalidateI18nCache("", "publish")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    nil,
	})
}

// Preview 预览模板渲染效果
// POST /api/v1/i18n/preview
func (h *I18nHandler) Preview(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "i18n SDK 未初始化",
			"data":    nil,
		})
		return
	}

	var req struct {
		LangCode string                 `json:"lang_code" binding:"required"`
		Key      string                 `json:"key" binding:"required"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	rendered, err := h.client.T(context.Background(), req.LangCode, req.Key, req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "渲染失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"rendered": rendered,
		},
	})
}
