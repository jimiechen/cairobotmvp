package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jimiechen/mineplanet/go/tars/provider-admin/internal/server"
)

func main() {
	port := os.Getenv("ADMIN_PORT")
	if port == "" {
		port = "8090"
	}

	dbPath := os.Getenv("ADMIN_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/admin.db"
	}

	gin.SetMode(gin.ReleaseMode)

	httpServer, err := server.NewHTTPServer(dbPath)
	if err != nil {
		log.Fatalf("初始化 HTTP 服务器失败: %v", err)
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("ProviderAdminServer 启动，监听地址: %s", addr)

	if err := httpServer.Run(addr); err != nil {
		log.Fatalf("HTTP 服务器启动失败: %v", err)
	}
}
