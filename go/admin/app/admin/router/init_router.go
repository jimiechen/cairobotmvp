package router

import (
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	common "go-admin/common/middleware"

	configAdminRouter "go-admin/app/admin/config_admin/router"
	i18nAdminRouter "go-admin/app/admin/i18n_admin/router"
)

// InitRouter 路由初始化，不要怀疑，这里用到了
func InitRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		log.Fatal("not found engine...")
		os.Exit(-1)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		os.Exit(-1)
	}

	// the jwt middleware
	authMiddleware, err := common.AuthInit()
	if err != nil {
		log.Fatalf("JWT Init Error, %s", err.Error())
	}
	casbinHandler := common.AuthCheckRole()

	// 注册系统路由
	InitSysRouter(r, authMiddleware)

	// 注册业务路由
	InitExamplesRouter(r, authMiddleware)

	// 注册配置中心路由（/api/admin/v1/config/*）
	configAdminRouter.ConfigAdminRouter(r, authMiddleware.MiddlewareFunc(), casbinHandler)

	// 注册多语言管理路由（/api/admin/v1/i18n/*）
	i18nAdminRouter.I18nAdminRouter(r, authMiddleware.MiddlewareFunc(), casbinHandler)
}
