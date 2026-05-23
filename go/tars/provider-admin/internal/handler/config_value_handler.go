package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

// ConfigValueHandler 配置值 API 处理器
// 负责配置版本和值的查询、更新、发布操作
type ConfigValueHandler struct {
	configRepo repository.ConfigRepository
}

// NewConfigValueHandler 创建配置值处理器实例
func NewConfigValueHandler(configRepo repository.ConfigRepository) *ConfigValueHandler {
	return &ConfigValueHandler{configRepo: configRepo}
}

// GetPublishedValues 按 env 获取所有已发布模块的配置值
// GET /api/v1/config/value/:env
func (h *ConfigValueHandler) GetPublishedValues(c *gin.Context) {
	env := c.Param("env")
	if env == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "env 参数必填",
			"data":    nil,
		})
		return
	}

	versions, err := h.configRepo.ListPublishedVersions(env)
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
		"data":    versions,
	})
}

// UpdateConfigValue 更新某模块的配置值
// PUT /api/v1/config/value
func (h *ConfigValueHandler) UpdateConfigValue(c *gin.Context) {
	var req struct {
		ModuleKey  string `json:"module_key" binding:"required"`
		Env        string `json:"env" binding:"required"`
		ConfigJSON string `json:"config_json" binding:"required"`
		CreateBy   string `json:"create_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	version := &domain.ConfigVersion{
		ModuleKey:  req.ModuleKey,
		Env:        req.Env,
		ConfigJSON: req.ConfigJSON,
		CreateBy:   req.CreateBy,
	}

	err := h.configRepo.Save(version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"id":      version.ID,
			"version": version.Version,
		},
	})
}

// PublishConfig 发布配置版本（标记 is_published=1）
// POST /api/v1/config/value/publish
func (h *ConfigValueHandler) PublishConfig(c *gin.Context) {
	var req struct {
		ModuleKey string `json:"module_key" binding:"required"`
		Env       string `json:"env" binding:"required"`
		Version   int64  `json:"version" binding:"required"`
		UpdateBy  string `json:"update_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}

	version, err := h.configRepo.GetByModuleAndVersion(req.ModuleKey, req.Env, req.Version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询版本失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	if version == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "版本不存在",
			"data":    nil,
		})
		return
	}

	version.IsPublished = true
	version.UpdateBy = req.UpdateBy

	err = h.configRepo.Save(version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "发布失败: " + err.Error(),
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
