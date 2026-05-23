package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jimiechen/mineplanet/go/services/config/cache"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	fmt.Println("🚀 CaiRobot Config Server starting...")

	dbPath := os.Getenv("CONFIG_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/config.db"
	}

	configRepo, err := repository.NewSQLiteConfigRepo(dbPath)
	if err != nil {
		log.Fatalf("❌ 初始化配置仓库失败: %v", err)
	}
	defer configRepo.Close()

	schemaRepo := repository.NewSQLiteSchemaRepo(configRepo.DB())
	lruCache := cache.NewMockCache()

	configSvc := service.NewAppConfigService(configRepo, schemaRepo, lruCache)

	// 注册 Config 模块的本地 TarsGo servant handler 到 LocalInvoker
	// Gateway 通过 LocalInvoker 进程内调用此服务
	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterConfigI18nHandlers(invoker, configSvc, nil)

	fmt.Println("✅ Config Server started successfully")
	fmt.Printf("   Database: %s\n", dbPath)
	fmt.Println("   Methods: GetAppConfigs, AppConfigVersion")
	fmt.Println("   Invoker: LocalInvoker ready (monolith mode)")
	fmt.Println("\n📡 Press Ctrl+C to stop")

	<-ctx.Done()
	fmt.Println("\n👋 Config Server stopped")
}
