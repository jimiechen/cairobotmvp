package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/models"
)

type StringApi struct {
	svc i18nAdmin.I18nAdminService
}

func NewStringApi(svc i18nAdmin.I18nAdminService) StringApi {
	return StringApi{svc: svc}
}

func (e *StringApi) requireSvc(c *gin.Context) bool {
	if e.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "多语言服务未初始化"})
		return false
	}
	return true
}

func (e StringApi) ListStrings(c *gin.Context) {
	if !e.requireSvc(c) { return }
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

func (e StringApi) CreateString(c *gin.Context) {
	if !e.requireSvc(c) { return }
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

func (e StringApi) UpdateString(c *gin.Context) {
	if !e.requireSvc(c) { return }
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

func (e StringApi) DeleteString(c *gin.Context) {
	if !e.requireSvc(c) { return }
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

func isTemplateError(err error) bool {
	return err != nil && (containsError(err, "模板") || containsError(err, "template"))
}

func buildTemplateErrorResponse(err error) gin.H {
	return gin.H{
		"code":    10400,
		"message": "模板校验失败",
		"errors": []models.ImportErrorItem{{RowNum: 0, Reason: err.Error()}},
	}
}

func containsError(err error, keyword string) bool {
	if err == nil { return false }
	return len(err.Error()) > 0 && containsStr(err.Error(), keyword)
}

func containsStr(s, sub string) bool { return len(s) >= len(sub) && findSubstr(s, sub) >= 0 }

func findSubstr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub { return i }
	}
	return -1
}
