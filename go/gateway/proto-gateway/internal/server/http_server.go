package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/tarsclient"
)

// GatewayServer TarsGo HTTP Servant，作为 proto-gateway 的唯一 HTTP 入口
// 通过 tars.AddHttpServant 注册到 TarsGo 框架
type GatewayServer struct {
	routeTable *router.RouteTable
	invoker    tarsclient.TarsInvoker
	mode       string
}

// NewGatewayServer 创建 GatewayServer
func NewGatewayServer(rt *router.RouteTable, invoker tarsclient.TarsInvoker, mode string) *GatewayServer {
	return &GatewayServer{
		routeTable: rt,
		invoker:    invoker,
		mode:       mode,
	}
}

// ServeHTTP 实现 http.Handler，处理 POST /api/hello 请求
func (gs *GatewayServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tars.TLOG.Warn("method not allowed: " + r.Method + " path=" + r.URL.Path)
		writeError(w, http.StatusMethodNotAllowed, commonlib.CodeBadRequest, "method not allowed")
		return
	}

	if r.Header.Get("Content-Type") != "application/octet-stream" {
		tars.TLOG.Debug("unsupported Content-Type: " + r.Header.Get("Content-Type"))
		writeError(w, http.StatusUnsupportedMediaType, commonlib.CodeBadRequest, "unsupported media type")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		tars.TLOG.Error("read body failed: " + err.Error())
		writeError(w, http.StatusBadRequest, commonlib.CodeBadRequest, "read body failed")
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		tars.TLOG.Debug("empty body received from " + r.RemoteAddr)
		writeError(w, http.StatusBadRequest, commonlib.CodeBadRequest, "empty body")
		return
	}

	packet, err := adapter.DeserializeMessagePacket(body)
	if err != nil {
		tars.TLOG.Error("invalid message packet: " + err.Error() + " bodySize=" + fmt.Sprintf("%d", len(body)))
		writeError(w, http.StatusBadRequest, commonlib.CodeBadRequest, "invalid message packet: "+err.Error())
		return
	}

	if packet.MaxType <= 0 || packet.MinType <= 0 {
		tars.TLOG.Debug("invalid maxType/minType: maxType=" + fmt.Sprintf("%d", packet.MaxType) + " minType=" + fmt.Sprintf("%d", packet.MinType))
		writeError(w, http.StatusBadRequest, commonlib.CodeBadRequest, "maxType/minType must > 0")
		return
	}

	if len(packet.Data) == 0 {
		tars.TLOG.Debug("empty data field in packet routeKey=" + fmt.Sprintf("%d:%d", packet.MaxType, packet.MinType))
		writeError(w, http.StatusBadRequest, commonlib.CodeBadRequest, "data is empty")
		return
	}

	routeKey := fmt.Sprintf("%d:%d", packet.MaxType, packet.MinType)
	tars.TLOG.Debug("incoming request: routeKey=" + routeKey + " dataSize=" + fmt.Sprintf("%d", len(packet.Data)) + " method=" + packet.Extend["method"])

	route, ok := gs.routeTable.FindRoute(packet.MaxType, packet.MinType)
	if !ok {
		tars.TLOG.Warn("route not found: " + routeKey)
		resp := adapter.BuildErrorPacket(packet, commonlib.CodeNotFound, "route not found")
		writePacket(w, resp)
		return
	}

	tars.TLOG.Info("route matched: " + routeKey + " -> " + route.CommandName)

	extend := adapter.BuildTarsExtend(packet, route.RouteKey, route.RequestProto, route.ResponseProto, route.AuthRequired, route.AuditRequired)

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
		tars.TLOG.Error("invoke failed ["+route.CommandName+"]: " + err.Error())
		resp := adapter.BuildErrorPacket(packet, commonlib.CodeInternalError, "invoke failed: "+err.Error())
		writePacket(w, resp)
		return
	}

	resp := adapter.BuildResponsePacket(packet, route.ResponseMax, route.ResponseMin, responseBytes, int32(returnCode))
	tars.TLOG.Info("invoke success ["+route.CommandName+"] returnCode=" + fmt.Sprintf("%d", returnCode))
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
