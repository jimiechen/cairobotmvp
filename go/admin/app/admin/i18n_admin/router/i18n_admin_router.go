package i18n_admin

import (
	"github.com/gin-gonic/gin"

	"github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/apis"
)

var adminSvc admin.I18nAdminService

// InitAdminService 注入 AdminI18nService（在 main.go 启动时调用）
func InitAdminService(svc admin.I18nAdminService) {
	adminSvc = svc
}

// I18nAdminRouter 返回路由注册函数，供 init_router.go 调用
func I18nAdminRouter(r *gin.Engine, jwtAuth gin.HandlerFunc, casbinHandler gin.HandlerFunc) {
	apiV1 := r.Group("/api/admin/v1/i18n")
	authed := apiV1.Use(jwtAuth, casbinHandler)
	registerI18nRoutes(authed.(*gin.RouterGroup))
}

func registerI18nRoutes(r *gin.RouterGroup) {
	if adminSvc == nil {
		return
	}
	stringApi := apis.NewStringApi(adminSvc)
	packApi := apis.NewPackApi(adminSvc)
	importApi := apis.NewImportExportApi(adminSvc)

	strGroup := r.Group("/string")
	{
		strGroup.GET("", stringApi.ListStrings)
		strGroup.POST("", stringApi.CreateString)
		strGroup.PUT("", stringApi.UpdateString)
		strGroup.DELETE("", stringApi.DeleteString)
	}

	packGroup := r.Group("/pack")
	{
		packGroup.POST("/publish", packApi.PublishPack)
		packGroup.POST("/rollback", packApi.RollbackPack)
	}

	importGroup := r.Group("")
	{
		importGroup.POST("/import/csv", importApi.ImportStringsFromCSV)
		importGroup.GET("/export/csv", importApi.ExportStringsToCSV)
	}
}
