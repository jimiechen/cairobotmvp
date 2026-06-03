package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/models"
)

type PackApi struct {
	svc i18nAdmin.I18nAdminService
}

func NewPackApi(svc i18nAdmin.I18nAdminService) PackApi {
	return PackApi{svc: svc}
}

func (e *PackApi) requireSvc(c *gin.Context) bool {
	if e.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "多语言服务未初始化"})
		return false
	}
	return true
}

func (e PackApi) PublishPack(c *gin.Context) {
	if !e.requireSvc(c) { return }
	var req models.PublishPackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	result, err := e.svc.PublishPack(c.Request.Context(), i18nAdmin.PublishPackRequest{
		PackID:   req.PackID,
		LangCode: req.LangCode,
		Env:      req.Env,
		Operator: operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发布失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result, "msg": "发布成功"})
}

func (e PackApi) RollbackPack(c *gin.Context) {
	if !e.requireSvc(c) { return }
	var req models.RollbackPackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败"})
		return
	}
	operator := c.GetString("x-user-name")
	if err := e.svc.RollbackPack(c.Request.Context(), req.PackID, req.TargetVersion, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "回滚失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "回滚成功"})
}
