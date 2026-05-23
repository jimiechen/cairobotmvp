package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	configService "github.com/jimiechen/mineplanet/go/services/config/service"
	"github.com/jimiechen/mineplanet/go/tars/provider-admin/internal/handler"
	"github.com/jimiechen/mineplanet/go/tars/provider-admin/internal/middleware"
)

// HTTPServer HTTP 服务器封装
// 负责初始化 Gin 引擎、注册路由、注入依赖
type HTTPServer struct {
	engine     *gin.Engine
	configRepo *repository.SQLiteConfigRepo
}

// NewHTTPServer 创建 HTTP 服务器实例
// 初始化数据库连接、创建表结构、注册所有路由
func NewHTTPServer(dbPath string) (*HTTPServer, error) {
	configRepo, err := repository.NewSQLiteConfigRepo(dbPath)
	if err != nil {
		return nil, fmt.Errorf("配置仓库初始化失败: %w", err)
	}

	db := configRepo.DB()

	schemaRepo := repository.NewSQLiteSchemaRepo(db)
	schemaSvc := configService.NewSchemaService(schemaRepo)

	adminI18nRepo := handler.NewAdminSQLiteRepo(db)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(requestLogger())

	healthHandler := handler.NewHealthHandler()
	configHandler := handler.NewConfigHandler(schemaSvc)
	configValueHandler := handler.NewConfigValueHandler(configRepo)
	i18nHandler := handler.NewI18nHandler(adminI18nRepo, nil)

	registerRoutes(engine, healthHandler, configHandler, configValueHandler, i18nHandler)

	return &HTTPServer{
		engine:     engine,
		configRepo: configRepo,
	}, nil
}

// Run 启动 HTTP 服务器（阻塞）
func (s *HTTPServer) Run(addr string) error {
	return s.engine.Run(addr)
}

// Engine 获取 Gin 引擎实例（用于测试）
func (s *HTTPServer) Engine() *gin.Engine {
	return s.engine
}

// Close 关闭数据库连接
func (s *HTTPServer) Close() error {
	return s.configRepo.Close()
}

// registerRoutes 注册所有 API 路由
func registerRoutes(
	r *gin.Engine,
	health *handler.HealthHandler,
	config *handler.ConfigHandler,
	configValue *handler.ConfigValueHandler,
	i18n *handler.I18nHandler,
) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", health.HealthCheck)

		configGroup := v1.Group("/config")
		{
			configGroup.POST("/schema", config.CreateSchema)
			configGroup.GET("/schema", config.ListSchemas)
			configGroup.PUT("/schema/:id", config.UpdateSchema)
			configGroup.DELETE("/schema/:id", config.DeleteSchema)

			configGroup.GET("/value/:env", configValue.GetPublishedValues)
			configGroup.PUT("/value", configValue.UpdateConfigValue)
			configGroup.POST("/value/publish", configValue.PublishConfig)
		}

		i18nGroup := v1.Group("/i18n")
		{
			i18nGroup.POST("/pack", i18n.CreatePack)
			i18nGroup.GET("/pack/:lang_code", i18n.GetPack)
			i18nGroup.POST("/string", i18n.CreateString)
			i18nGroup.PUT("/string/:id", i18n.UpdateString)
			i18nGroup.DELETE("/string/:id", i18n.DeleteString)
			i18nGroup.GET("/diff", i18n.GetDiff)
			i18nGroup.POST("/publish", i18n.PublishLangPack)
			i18nGroup.POST("/preview", i18n.Preview)
		}
	}
}

// requestLogger 请求日志中间件
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[%s] %s %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	}
}
