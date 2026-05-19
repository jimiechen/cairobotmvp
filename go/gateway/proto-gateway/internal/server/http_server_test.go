package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/adapter"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/config"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/router"
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/internal/tarsclient"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// mockInvoker 用于 HTTP 测试
type mockInvoker struct {
	returnCode int
	response   []byte
	err        error
}

func (m *mockInvoker) Invoke(ctx context.Context, target tarsclient.Target, request []byte, extend map[string]string) (int, []byte, error) {
	return m.returnCode, m.response, m.err
}

func TestGatewayServer_ServeHTTP(t *testing.T) {
	cfg := &config.RoutesConfig{
		Routes: []config.Route{
			{
				RequestMax:    2100,
				RequestMin:    2097,
				RouteKey:      "2100:2097",
				CommandName:   "ServiceHealthCheck",
				ResponseMax:   2100,
				ResponseMin:   2098,
				TarsApp:       "CaiRobot",
				TarsServer:    "SystemServer",
				TarsServant:   "SystemObj",
				TarsModule:    "CaiRobotSystemApp",
				TarsInterface: "SystemObj",
				TarsMethod:    "HealthCheck",
			},
		},
	}
	r := router.NewRouter(cfg)

	t.Run("GET returns 405", func(t *testing.T) {
		invoker := &mockInvoker{returnCode: 10200, response: []byte("ok")}
		gs := NewGatewayServer(r, invoker, "local")
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})

	t.Run("wrong Content-Type", func(t *testing.T) {
		invoker := &mockInvoker{returnCode: 10200, response: []byte("ok")}
		gs := NewGatewayServer(r, invoker, "local")
		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader([]byte("test")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("expected 415, got %d", rec.Code)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		invoker := &mockInvoker{returnCode: 10200, response: []byte("ok")}
		gs := NewGatewayServer(r, invoker, "local")
		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("invalid message packet", func(t *testing.T) {
		invoker := &mockInvoker{returnCode: 10200, response: []byte("ok")}
		gs := NewGatewayServer(r, invoker, "local")
		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, 10400, "test error")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("expected Content-Type application/octet-stream, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestWritePacket(t *testing.T) {
	rec := httptest.NewRecorder()
	packet := &adapter.MessagePacket{
		MaxType: 2100,
		MinType: 2098,
		Extend:  map[string]string{"code": "10200"},
		Data:    []byte("ok"),
	}
	writePacket(rec, packet)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("expected Content-Type application/octet-stream, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestGatewayServer_E2E_HealthCheck(t *testing.T) {
	cfg := &config.RoutesConfig{
		Routes: []config.Route{
			{
				RequestMax:    2100,
				RequestMin:    2097,
				RouteKey:      "2100:2097",
				CommandName:   "ServiceHealthCheck",
				ResponseMax:   2100,
				ResponseMin:   2098,
				TarsApp:       "CaiRobot",
				TarsServer:    "SystemServer",
				TarsServant:   "SystemObj",
				TarsModule:    "CaiRobotSystemApp",
				TarsInterface: "SystemObj",
				TarsMethod:    "HealthCheck",
			},
		},
	}
	r := router.NewRouter(cfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterSystemHandlers(invoker)

	gs := NewGatewayServer(r, invoker, "local")

	reqPacket := &adapter.MessagePacket{
		MaxType:  2100,
		MinType:  2097,
		Platform: 1,
		Extend:   map[string]string{"traceId": "e2e-abc"},
		Data:     []byte{},
	}
	body, err := adapter.SerializeMessagePacket(reqPacket)
	if err != nil {
		t.Fatalf("serialize request failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/octet-stream")
	rec := httptest.NewRecorder()
	gs.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	respBody := rec.Body.Bytes()
	if len(respBody) == 0 {
		t.Fatal("expected non-empty response body")
	}

	respPacket, err := adapter.DeserializeMessagePacket(respBody)
	if err != nil {
		t.Fatalf("deserialize response failed: %v", err)
	}
	if respPacket.MaxType != 2100 {
		t.Fatalf("expected response MaxType 2100, got %d", respPacket.MaxType)
	}
	if respPacket.MinType != 2098 {
		t.Fatalf("expected response MinType 2098, got %d", respPacket.MinType)
	}
	if respPacket.Extend["code"] != "10200" {
		t.Fatalf("expected code 10200, got %s", respPacket.Extend["code"])
	}

	var healthResp pb.ServiceHealthCheckResponse
	if err := proto.Unmarshal(respPacket.Data, &healthResp); err != nil {
		t.Fatalf("unmarshal health response failed: %v", err)
	}
	if healthResp.Status != "OK" {
		t.Fatalf("expected status 'OK', got %q", healthResp.Status)
	}
	if healthResp.Result == nil || healthResp.Result.Code != 10200 {
		t.Fatalf("expected Result.Code 10200, got %v", healthResp.Result)
	}
}
