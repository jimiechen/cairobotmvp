package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
	"go-admin/app/admin/config_admin/models"
)

type SchemaApi struct {
	svc configAdmin.ConfigSchemaService
}

func NewSchemaApi(svc configAdmin.ConfigSchemaService) SchemaApi {
	return SchemaApi{svc: svc}
}

func (e *SchemaApi) requireSvc(c *gin.Context) bool {
	if e.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "配置服务未初始化"})
		return false
	}
	return true
}

func (e SchemaApi) GetSchemaList(c *gin.Context) {
	if !e.requireSvc(c) { return }
	var req models.ListSchemaReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	items, err := e.svc.ListSchemas(c.Request.Context(), req.ModuleKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "msg": "查询成功"})
}

func (e SchemaApi) CreateSchema(c *gin.Context) {
	if !e.requireSvc(c) { return }
	var req models.CreateSchemaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	result, err := e.svc.CreateSchema(c.Request.Context(), configAdmin.CreateSchemaRequest{
		ModuleKey:    req.ModuleKey,
		FieldKey:     req.FieldKey,
		FieldType:    configAdmin.ToFieldType(req.FieldType),
		DefaultValue: req.DefaultValue,
		Validator:    req.Validator,
		IsRequired:   req.IsRequired,
		IsSecret:     req.IsSecret,
		Description:  req.Description,
		ClientScope:  req.ClientScope,
		SortOrder:    req.SortOrder,
		Operator:     operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "创建成功"})
}

func (e SchemaApi) UpdateSchema(c *gin.Context) {
	if !e.requireSvc(c) { return }
	var req models.UpdateSchemaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	result, err := e.svc.UpdateSchema(c.Request.Context(), configAdmin.UpdateSchemaRequest{
		ID:           req.ID,
		FieldType:    configAdmin.ToFieldType(req.FieldType),
		DefaultValue: req.DefaultValue,
		Validator:    req.Validator,
		IsRequired:   req.IsRequired,
		IsSecret:     req.IsSecret,
		Description:  req.Description,
		ClientScope:  req.ClientScope,
		SortOrder:    req.SortOrder,
		Operator:     operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "更新成功"})
}

func (e SchemaApi) DeleteSchema(c *gin.Context) {
	if !e.requireSvc(c) { return }
	idStr := c.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 id 参数"})
		return
	}
	operator := c.GetString("x-user-name")
	if delErr := e.svc.DeleteSchema(c.Request.Context(), id, operator); delErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败", "error": delErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
