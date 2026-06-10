package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/jimiechen/mineplanet/go/common-lib/config"
	gatewaycfg "github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/middleware"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/server"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	"github.com/jimiechen/mineplanet/go/tars/auth"
)

func main() {
	mode := os.Getenv("GATEWAY_INVOKER_MODE")
	if mode == "" {
		mode = "local"
	}

	configPath := os.Getenv("TARS_CONFIG")
	if configPath == "" {
		configPath = "configs/gateway/gateway.local.conf"
	}
	tars.ServerConfigPath = configPath

	cfg, err := gatewaycfg.LoadRoutesWithEnv("configs/gateway/routes.yaml")
	if err != nil {
		tars.TLOG.Error("load routes failed: " + err.Error())
		os.Exit(1)
	}

	rt := router.NewRouteTable(cfg)
	tars.TLOG.Info("route table loaded, count=" + fmt.Sprintf("%d", len(cfg.Routes)))

	var invoker tarsclient.TarsInvoker
	var authMw *middleware.AuthMiddleware
	if mode == "local" {
		invoker = tarsclient.NewLocalInvoker()
		tarsclient.RegisterAllLocalHandlers(invoker.(*tarsclient.LocalInvoker))
		tars.TLOG.Info("invoker mode=local (monolith TarsGo adapter) | handlers=System+Config+I18n (noop)")
	} else if mode == "mysql" {
		mysqlCfg := loadMySQLConfigFromEnv()
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "dev"
		}
		svc, err := tarsclient.BuildRealServices(mysqlCfg, env)
		if err != nil {
			tars.TLOG.Error("BuildRealServices failed: " + err.Error())
			os.Exit(1)
		}
		// S1: 初始化 JWT Auth 中间件
		jwtSecret := []byte(getEnv("JWT_SECRET", "cairobot-s1-default-secret-change-in-production"))
		authSvc := auth.NewAuthService(jwtSecret, "cairobot", 24*time.Hour)
		authMw = middleware.NewAuthMiddleware(authSvc)

		invoker = tarsclient.NewLocalInvoker()
		tarsclient.RegisterRealHandlers(invoker.(*tarsclient.LocalInvoker), svc)
		tars.TLOG.Info("invoker mode=mysql (real services from go_biz) | handlers=System+Config+I18n (MySQL) | auth=JWT enabled")
	} else if mode == "tars" {
		tars.TLOG.Error("tars microservice invoker is not implemented yet")
		os.Exit(1)
	} else {
		tars.TLOG.Error("unknown invoker mode=" + mode)
		os.Exit(1)
	}

	mux := &tars.TarsHttpMux{}
	gs := server.NewGatewayServer(rt, invoker, mode, authMw)
	mux.Handle("/api/hello", gs)

	svrCfg := tars.GetServerConfig()
	objName := svrCfg.App + "." + svrCfg.Server + ".GatewayHttpObj"
	tars.AddHttpServant(mux, objName)

	tars.TLOG.Info("proto-gateway started | mode=" + mode + " | tarsObj=" + objName + " | adapter=GatewayHttpObjAdapter")

	tars.Run()
}

// loadMySQLConfigFromEnv 从环境变量加载 MySQL 连接配置
// 环境变量：MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DATABASE
func loadMySQLConfigFromEnv() *config.MySQLConfig {
	port := 3306
	if v := os.Getenv("MYSQL_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")
	db := getEnv("MYSQL_DATABASE", "go_biz")

	return &config.MySQLConfig{
		Host:            host,
		Port:            port,
		Username:        user,
		Password:        pass,
		Database:        db,
		Charset:         "utf8mb4",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: "1h",
		ConnMaxIdleTime: "10m",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
