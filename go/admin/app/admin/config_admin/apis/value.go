package apis

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"go-admin/app/admin/config_admin/models"
)

type ValueApi struct {
	svc configAdmin.ConfigAdminService
}

func NewValueApi(svc configAdmin.ConfigAdminService) ValueApi {
	return ValueApi{svc: svc}
}

func (e *ValueApi) requireSvc(c *gin.Context) bool {
	if e.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "配置服务未初始化"})
		return false
	}
	return true
}

func (e ValueApi) PublishValue(c *gin.Context) {
	if !e.requireSvc(c) { return }
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

func convertPublishFields(items []models.PublishFieldItem) map[string]*domain.TypedValue {
	result := make(map[string]*domain.TypedValue, len(items))
	for _, item := range items {
		result[item.FieldKey] = domain.NewTypedValue(domain.FieldTypeString, item.Value)
	}
	return result
}

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
