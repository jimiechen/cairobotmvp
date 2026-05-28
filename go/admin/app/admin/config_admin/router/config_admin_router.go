package config_admin

import (
	"github.com/gin-gonic/gin"

	"go-admin/app/admin/config_admin/apis"
	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
)

// AdminServiceHolder 持有 AdminConfigService 实例
var adminSvc configAdmin.ConfigAdminService

// InitAdminService 注入 AdminConfigService（在 main.go 启动时调用）
func InitAdminService(svc configAdmin.ConfigAdminService) {
	adminSvc = svc
}

// registerConfigAdminRoutes 注册配置管理路由
func registerConfigAdminRoutes(r *gin.RouterGroup) {
	if adminSvc == nil {
		return
	}
	schemaApi := apis.NewSchemaApi(adminSvc)
	valueApi := apis.NewValueApi(adminSvc)

	schemaGroup := r.Group("/schema")
	{
		schemaGroup.GET("", schemaApi.GetSchemaList)
		schemaGroup.POST("", schemaApi.CreateSchema)
		schemaGroup.PUT("", schemaApi.UpdateSchema)
		schemaGroup.DELETE("", schemaApi.DeleteSchema)
	}

	valueGroup := r.Group("/value")
	{
		valueGroup.POST("/publish", valueApi.PublishValue)
		valueGroup.GET("/versions", valueApi.GetValueVersions)
	}
}

// ConfigAdminRouter 返回路由注册函数，供 init_router.go 调用
func ConfigAdminRouter(r *gin.Engine, jwtAuth gin.HandlerFunc, casbinHandler gin.HandlerFunc) {
	apiV1 := r.Group("/api/admin/v1/config")
	authed := apiV1.Use(jwtAuth, casbinHandler)
	registerConfigAdminRoutes(authed.(*gin.RouterGroup))
}
