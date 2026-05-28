package apis

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"go-admin/app/admin/config_admin/models"
)

// ValueApi 配置值管理 HTTP Handler
type ValueApi struct {
	svc configAdmin.ConfigAdminService
}

// NewValueApi 创建 Value API 实例
func NewValueApi(svc configAdmin.ConfigAdminService) ValueApi {
	return ValueApi{svc: svc}
}

// PublishValue 发布配置值（创建新版本）
// POST /api/admin/v1/config/value/publish
// 权限：config:value:write
//
// 校验失败时返回 HTTP 10400 + 字段级错误明细
func (e ValueApi) PublishValue(c *gin.Context) {
	var req models.PublishValueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")

	fields := convertPublishFields(req.Fields)
	adminReq := configAdmin.PublishValueRequest{
		ModuleKey: req.ModuleKey,
		Env:       req.Env,
		Fields:    fields,
		Operator:  operator,
	}

	result, err := e.svc.PublishValue(c.Request.Context(), adminReq)
	if err != nil {
		var valErr *configAdmin.ValidationError
		if errors.As(err, &valErr) {
			resp := buildValidationResponse(valErr)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发布失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "发布成功"})
}

// GetValueVersions 查询模块配置版本列表
// GET /api/admin/v1/config/value/versions?module_key=xxx&env=xxx
// 权限：config:value:read
func (e ValueApi) GetValueVersions(c *gin.Context) {
	moduleKey := c.Query("module_key")
	env := c.Query("env")

	if moduleKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 module_key 参数"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"module_key": moduleKey,
			"env":        env,
			"versions":   []interface{}{},
		},
		"msg": "查询成功",
	})
}

// convertPublishFields 将前端传入的字段列表转换为 admin 服务所需的 map
func convertPublishFields(items []models.PublishFieldItem) map[string]*domain.TypedValue {
	result := make(map[string]*domain.TypedValue, len(items))
	for _, item := range items {
		result[item.FieldKey] = domain.NewTypedValue(domain.FieldTypeString, item.Value)
	}
	return result
}

// buildValidationResponse 构建 10400 校验错误响应
func buildValidationResponse(valErr *configAdmin.ValidationError) models.ValidationErrorResp {
	items := make([]models.ValidationErrorItem, len(valErr.Errors))
	for i, ve := range valErr.Errors {
		items[i] = models.ValidationErrorItem{
			Field:  ve.Field,
			Reason: ve.Reason,
		}
	}
	return models.ValidationErrorResp{
		Code:    10400,
		Message: "配置值校验失败",
		Errors:  items,
	}
}
