package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/models"
)

// StringApi 语言字符串管理 HTTP Handler
type StringApi struct {
	svc i18nAdmin.I18nAdminService
}

// NewStringApi 创建 String API 实例
func NewStringApi(svc i18nAdmin.I18nAdminService) StringApi {
	return StringApi{svc: svc}
}

// CreateString 新增语言字符串
// POST /api/admin/v1/i18n/string
// 权限：i18n:string:write
func (e StringApi) CreateString(c *gin.Context) {
	var req models.CreateStringReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	result, err := e.svc.CreateString(c.Request.Context(), i18nAdmin.CreateStringRequest{
		PackID:       req.PackID,
		StringKey:    i18nAdmin.ToStringKey(req.StringKey),
		StringValue:  req.StringValue,
		GroupName:    req.GroupName,
		TemplateType:  i18nAdmin.ToTemplateType(req.TemplateType),
		ParamsSchema: req.ParamsSchema,
		PreviewSample: req.PreviewSample,
		Operator:     operator,
	})
	if err != nil {
		if isTemplateError(err) {
			c.JSON(http.StatusBadRequest, buildTemplateErrorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "创建成功"})
}

// UpdateString 更新语言字符串
// PUT /api/admin/v1/i18n/string
// 权限：i18n:string:write
func (e StringApi) UpdateString(c *gin.Context) {
	var req models.UpdateStringReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	result, err := e.svc.UpdateString(c.Request.Context(), i18nAdmin.UpdateStringRequest{
		ID:           req.ID,
		StringValue:  req.StringValue,
		GroupName:    req.GroupName,
		ParamsSchema: req.ParamsSchema,
		PreviewSample: req.PreviewSample,
		Operator:     operator,
	})
	if err != nil {
		if isTemplateError(err) {
			c.JSON(http.StatusBadRequest, buildTemplateErrorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "更新成功"})
}

// DeleteString 删除语言字符串（标记 DEL）
// DELETE /api/admin/v1/i18n/string?id=xxx
// 权限：i18n:string:delete
func (e StringApi) DeleteString(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 id 参数"})
		return
	}
	operator := c.GetString("x-user-name")
	if delErr := e.svc.DeleteString(c.Request.Context(), id, operator); delErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败", "error": delErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// ListStrings 查询指定语言包下所有字符串
// GET /api/admin/v1/i18n/string?pack_id=xxx
// 权限：i18n:string:read
func (e StringApi) ListStrings(c *gin.Context) {
	var req models.ListStringsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	items, err := e.svc.ListStrings(req.PackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "msg": "查询成功"})
}

// isTemplateError 判断错误是否为模板校验错误
func isTemplateError(err error) bool {
	return err != nil && (containsError(err, "模板") || containsError(err, "template"))
}

// buildTemplateErrorResponse 构建 10400 模板校验错误响应
func buildTemplateErrorResponse(err error) gin.H {
	return gin.H{
		"code":    10400,
		"message": "模板校验失败",
		"errors": []models.ImportErrorItem{{RowNum: 0, Reason: err.Error()}},
	}
}

// containsError 简单判断 error message 是否包含关键字
func containsError(err error, keyword string) bool {
	if err == nil { return false }
	return len(err.Error()) > 0 && containsStr(err.Error(), keyword)
}

// containsStr 字符串包含检查
func containsStr(s, sub string) bool { return len(s) >= len(sub) && findSubstr(s, sub) >= 0 }

// findSubstr 简单子串查找
func findSubstr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub { return i }
	}
	return -1
}
