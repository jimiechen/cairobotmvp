package config_admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	configAdmin "github.com/jimiechen/mineplanet/go/services/config/admin"
	"go-admin/app/admin/config_admin/apis"
)

var adminSvc configAdmin.ConfigAdminService

// InitAdminService 注入 AdminConfigService（在 main.go 启动时调用）
func InitAdminService(svc configAdmin.ConfigAdminService) {
	adminSvc = svc
}

// ConfigAdminRouter 返回路由注册函数，供 init_router.go 调用
func ConfigAdminRouter(r *gin.Engine, jwtAuth gin.HandlerFunc, casbinHandler gin.HandlerFunc) {
	apiV1 := r.Group("/api/admin/v1/config")
	authed := apiV1.Use(jwtAuth, casbinHandler)
	registerConfigAdminRoutes(authed.(*gin.RouterGroup))
}

func registerConfigAdminRoutes(r *gin.RouterGroup) {
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

func serviceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"code": 503,
		"msg":  "配置服务未初始化，请检查后端服务配置",
	})
}
