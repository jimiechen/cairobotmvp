package main

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/server"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
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

	cfg, err := config.LoadRoutesWithEnv("configs/gateway/routes.yaml")
	if err != nil {
		tars.TLOG.Error("load routes failed: " + err.Error())
		os.Exit(1)
	}

	rt := router.NewRouteTable(cfg)
	tars.TLOG.Info("route table loaded, count=" + fmt.Sprintf("%d", len(cfg.Routes)))

	var invoker tarsclient.TarsInvoker
	if mode == "local" {
		invoker = tarsclient.NewLocalInvoker()
		tarsclient.RegisterSystemHandlers(invoker.(*tarsclient.LocalInvoker))
		tars.TLOG.Info("invoker mode=local (monolith TarsGo adapter)")
	} else if mode == "tars" {
		tars.TLOG.Error("tars microservice invoker is not implemented yet")
		os.Exit(1)
	} else {
		tars.TLOG.Error("unknown invoker mode=" + mode)
		os.Exit(1)
	}

	mux := &tars.TarsHttpMux{}
	gs := server.NewGatewayServer(rt, invoker, mode)
	mux.Handle("/api/hello", gs)

	svrCfg := tars.GetServerConfig()
	objName := svrCfg.App + "." + svrCfg.Server + ".GatewayHttpObj"
	tars.AddHttpServant(mux, objName)

	tars.TLOG.Info("proto-gateway started | mode=" + mode + " | tarsObj=" + objName + " | adapter=GatewayHttpObjAdapter")

	tars.Run()
}
