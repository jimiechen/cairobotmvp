package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
	"go-admin/app/admin/config_admin/models"
)

// SchemaApi Schema 管理 HTTP Handler
type SchemaApi struct {
	svc configAdmin.ConfigSchemaService
}

// NewSchemaApi 创建 Schema API 实例
func NewSchemaApi(svc configAdmin.ConfigSchemaService) SchemaApi {
	return SchemaApi{svc: svc}
}

// GetSchemaList 获取指定模块下所有字段 Schema
// GET /api/admin/v1/config/schema?module_key=xxx
// 权限：config:schema:read
func (e SchemaApi) GetSchemaList(c *gin.Context) {
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

// CreateSchema 新增字段定义
// POST /api/admin/v1/config/schema
// 权限：config:schema:write
func (e SchemaApi) CreateSchema(c *gin.Context) {
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

// UpdateSchema 更新字段定义
// PUT /api/admin/v1/config/schema
// 权限：config:schema:write
func (e SchemaApi) UpdateSchema(c *gin.Context) {
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

// DeleteSchema 软删除字段定义
// DELETE /api/admin/v1/config/schema?id=xxx
// 权限：config:schema:delete
func (e SchemaApi) DeleteSchema(c *gin.Context) {
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
