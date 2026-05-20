package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/tarsclient"
)

// GatewayServer HTTP 网关服务器
type GatewayServer struct {
	router  *router.Router
	invoker tarsclient.TarsInvoker
	mode    string
}

// NewGatewayServer 创建 GatewayServer
func NewGatewayServer(r *router.Router, invoker tarsclient.TarsInvoker, mode string) *GatewayServer {
	return &GatewayServer{
		router:  r,
		invoker: invoker,
		mode:    mode,
	}
}

// ServeHTTP 实现 http.Handler
func (gs *GatewayServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, 10400, "method not allowed")
		return
	}

	if r.Header.Get("Content-Type") != "application/octet-stream" {
		writeError(w, http.StatusUnsupportedMediaType, 10400, "unsupported media type")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, 10400, "read body failed")
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, 10400, "empty body")
		return
	}

	// 解析 MessagePacket
	packet, err := adapter.DeserializeMessagePacket(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, 10400, "invalid message packet: "+err.Error())
		return
	}

	if packet.MaxType <= 0 || packet.MinType <= 0 {
		writeError(w, http.StatusBadRequest, 10400, "maxType/minType must > 0")
		return
	}

	if len(packet.Data) == 0 {
		writeError(w, http.StatusBadRequest, 10400, "data is empty")
		return
	}

	// 查找路由
	route, ok := gs.router.FindRoute(packet.MaxType, packet.MinType)
	if !ok {
		resp := adapter.BuildErrorPacket(packet, 10404, "route not found")
		writePacket(w, resp)
		return
	}

	// 构造 extend
	extend := adapter.BuildTarsExtend(packet, route.RouteKey, route.RequestProto, route.ResponseProto, route.AuthRequired, route.AuditRequired)

	// 调用 Invoker
	target := tarsclient.Target{
		App:       route.TarsApp,
		Server:    route.TarsServer,
		Servant:   route.TarsServant,
		Module:    route.TarsModule,
		Interface: route.TarsInterface,
		Method:    route.TarsMethod,
	}

	returnCode, responseBytes, err := gs.invoker.Invoke(r.Context(), target, packet.Data, extend)
	if err != nil {
		resp := adapter.BuildErrorPacket(packet, 10500, "invoke failed: "+err.Error())
		writePacket(w, resp)
		return
	}

	// 构造响应 MessagePacket
	resp := adapter.BuildResponsePacket(packet, route.ResponseMax, route.ResponseMin, responseBytes, int32(returnCode))
	writePacket(w, resp)
}

func writeError(w http.ResponseWriter, httpCode int, businessCode int32, message string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(httpCode)
	resp := adapter.BuildErrorPacket(nil, businessCode, message)
	data, _ := adapter.SerializeMessagePacket(resp)
	w.Write(data)
}

func writePacket(w http.ResponseWriter, packet *adapter.MessagePacket) {
	w.Header().Set("Content-Type", "application/octet-stream")
	data, err := adapter.SerializeMessagePacket(packet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// Run 启动 Gateway 服务器
func Run(ctx context.Context) error {
	mode := os.Getenv("GATEWAY_INVOKER_MODE")
	if mode == "" {
		mode = "local"
	}

	fmt.Printf("proto-gateway started, invoker_mode=%s\n", mode)

	if mode == "tars" {
		return fmt.Errorf("tars invoker is not implemented yet")
	}

	// 加载路由配置
	cfg, err := config.LoadRoutesWithEnv("configs/gateway/routes.yaml")
	if err != nil {
		return fmt.Errorf("load routes failed: %w", err)
	}

	r := router.NewRouter(cfg)

	// 创建 Invoker
	var invoker tarsclient.TarsInvoker
	if mode == "local" {
		invoker = tarsclient.NewLocalInvoker()
		// 注册 System handler
		tarsclient.RegisterSystemHandlers(invoker.(*tarsclient.LocalInvoker))
	} else {
		return fmt.Errorf("unknown invoker mode: %s", mode)
	}

	gs := NewGatewayServer(r, invoker, mode)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: gs,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}
