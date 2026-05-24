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
	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestGateway_E2E_Modules_HealthCheck 验证完整链路：
// HTTP POST → Gateway Server → MessagePacket 解码 → LocalInvoker → modules/health.Check() → Protobuf 响应
func TestGateway_E2E_Modules_HealthCheck(t *testing.T) {
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
	r := router.NewRouteTable(cfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker) // 🎯 使用新模块化注册

	gs := NewGatewayServer(r, invoker, "local")

	t.Run("正常 HealthCheck 请求返回有效 Protobuf 响应", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType:  2100,
			MinType:  2097,
			Platform: pb.Platform_WEB,
			Extend:   map[string]string{"traceId": "e2e-modules-health-001"},
		}

		healthReq := &pb.ServiceHealthCheckRequest{
			ServiceName: "modules-e2e-test",
		}
		reqData, err := proto.Marshal(healthReq)
		if err != nil {
			t.Fatalf("marshal request failed: %v", err)
		}
		reqPacket.Data = reqData

		body, err := adapter.SerializeMessagePacket(reqPacket)
		if err != nil {
			t.Fatalf("serialize packet failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", rec.Code)
		}

		respBody := rec.Body.Bytes()
		if len(respBody) == 0 {
			t.Fatal("expected non-empty response body")
		}

		respPacket, err := adapter.DeserializeMessagePacket(respBody)
		if err != nil {
			t.Fatalf("deserialize response failed: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("expected code '10200', got '%s'", respPacket.Extend["code"])
		}
		if respPacket.MaxType != 2100 || respPacket.MinType != 2098 {
			t.Fatalf("expected MaxType:2100 MinType:2098, got %d:%d", respPacket.MaxType, respPacket.MinType)
		}

		var healthResp pb.ServiceHealthCheckResponse
		if err := proto.Unmarshal(respPacket.Data, &healthResp); err != nil {
			t.Fatalf("unmarshal health response data failed: %v", err)
		}
		if healthResp.Status != "OK" {
			t.Fatalf("expected status 'OK', got '%s'", healthResp.Status)
		}
		if healthResp.Result == nil || healthResp.Result.Code != 10200 {
			t.Fatalf("expected Result.Code 10200, got %v", healthResp.Result)
		}
	})

	t.Run("空 ServiceName 也返回 OK", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 2100,
			MinType: 2097,
			Data:   []byte{0x08, 0x00}, // 最小有效 Protobuf bytes（非空）
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", rec.Code)
		}

		respPacket, _ := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		var healthResp pb.ServiceHealthCheckResponse
		proto.Unmarshal(respPacket.Data, &healthResp)
		if healthResp.Status != "OK" {
			t.Fatalf("expected status 'OK' for empty name, got '%s'", healthResp.Status)
		}
	})
}

// TestGateway_E2E_Modules_HelloWorld 验证完整链路：
// HTTP POST → Gateway Server → MessagePacket 解码 → LocalInvoker → modules/hello.SayHello() → Protobuf 响应
func TestGateway_E2E_Modules_HelloWorld(t *testing.T) {
	cfg := &config.RoutesConfig{
		Routes: []config.Route{
			{
				RequestMax:    2100,
				RequestMin:    2101,
				RouteKey:      "2100:2101",
				CommandName:   "HelloWorld",
				ResponseMax:   2100,
				ResponseMin:   2102,
				TarsApp:       "CaiRobot",
				TarsServer:    "SystemServer",
				TarsServant:   "SystemObj",
				TarsModule:    "CaiRobotSystemApp",
				TarsInterface: "SystemObj",
				TarsMethod:    "HelloWorld",
			},
		},
	}
	r := router.NewRouteTable(cfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker) // 🎯 使用新模块化注册

	gs := NewGatewayServer(r, invoker, "local")

	t.Run("带名称的 Hello 请求返回正确问候语", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType:  2100,
			MinType:  2101,
			Platform: pb.Platform_WEB,
			Extend:   map[string]string{"traceId": "e2e-modules-hello-001"},
		}

		helloReq := &pb.HelloWorldRequest{
			Name: "ModulesRefactor",
		}
		reqData, err := proto.Marshal(helloReq)
		if err != nil {
			t.Fatalf("marshal request failed: %v", err)
		}
		reqPacket.Data = reqData

		body, err := adapter.SerializeMessagePacket(reqPacket)
		if err != nil {
			t.Fatalf("serialize packet failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", rec.Code)
		}

		respBody := rec.Body.Bytes()
		respPacket, err := adapter.DeserializeMessagePacket(respBody)
		if err != nil {
			t.Fatalf("deserialize response failed: %v", err)
		}

		if respPacket.Extend["code"] != "10200" {
			t.Fatalf("expected code '10200', got '%s'", respPacket.Extend["code"])
		}
		if respPacket.MaxType != 2100 || respPacket.MinType != 2102 {
			t.Fatalf("expected MaxType:2100 MinType:2102, got %d:%d", respPacket.MaxType, respPacket.MinType)
		}

		var helloResp pb.HelloWorldResponse
		if err := proto.Unmarshal(respPacket.Data, &helloResp); err != nil {
			t.Fatalf("unmarshal hello response data failed: %v", err)
		}
		if helloResp.Message != "Hello, ModulesRefactor!" {
			t.Fatalf("expected message 'Hello, ModulesRefactor!', got '%s'", helloResp.Message)
		}
		if helloResp.Result == nil || helloResp.Result.Code != 10200 {
			t.Fatalf("expected Result.Code 10200, got %v", helloResp.Result)
		}
	})

	t.Run("空名称返回默认 Hello World", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 2100,
			MinType: 2101,
			Data:   []byte{0x08, 0x00}, // 最小有效 Protobuf bytes（非空）
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d", rec.Code)
		}

		respPacket, _ := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		var helloResp pb.HelloWorldResponse
		proto.Unmarshal(respPacket.Data, &helloResp)
		if helloResp.Message != "Hello, World!" {
			t.Fatalf("expected default 'Hello, World!', got '%s'", helloResp.Message)
		}
	})
}

// TestGateway_E2E_Modules_ErrorHandling 验证错误处理链路：
// 无效 Protobuf 数据 → 模块层错误 → 10500 错误码 → HTTP 500 + 错误响应
func TestGateway_E2E_Modules_ErrorHandling(t *testing.T) {
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
	r := router.NewRouteTable(cfg)

	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)

	gs := NewGatewayServer(r, invoker, "local")

	t.Run("无效 Protobuf Data 返回业务错误", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 2100,
			MinType: 2097,
			Data:   []byte("this-is-not-valid-protobuf-bytes"),
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 (Gateway wraps errors in packet), got %d", rec.Code)
		}

		respPacket, err := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if err != nil {
			t.Fatalf("should still return valid error packet: %v", err)
		}
		if respPacket.Extend["code"] != "10500" {
			t.Fatalf("expected business error code '10500', got '%s'", respPacket.Extend["code"])
		}
	})

	t.Run("路由未匹配返回业务错误 10404", func(t *testing.T) {
		reqPacket := &adapter.MessagePacket{
			MaxType: 9999,
			MinType: 9999,
			Data:   []byte{0x08, 0x00},
		}
		body, _ := adapter.SerializeMessagePacket(reqPacket)

		req := httptest.NewRequest(http.MethodPost, "/api/hello", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/octet-stream")
		rec := httptest.NewRecorder()
		gs.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 (Gateway wraps errors in packet), got %d", rec.Code)
		}

		respPacket, _ := adapter.DeserializeMessagePacket(rec.Body.Bytes())
		if respPacket.Extend["code"] != "10404" {
			t.Fatalf("expected business error code '10404', got '%s'", respPacket.Extend["code"])
		}
	})
}

// TestGateway_E2E_DirectInvokerTest 直接调用 Invoker 层测试（绕过 HTTP）
// 用于隔离验证 LocalInvoker → modules 链路
func TestGateway_E2E_DirectInvokerTest(t *testing.T) {
	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterModuleHandlers(invoker)

	t.Run("直接调用 HealthCheck 模块", func(t *testing.T) {
		target := tarsclient.Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HealthCheck",
		}

		req := &pb.ServiceHealthCheckRequest{ServiceName: "direct-invoker-test"}
		reqData, _ := proto.Marshal(req)

		code, resp, err := invoker.Invoke(context.Background(), target, reqData, nil)
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}

		var healthResp pb.ServiceHealthCheckResponse
		proto.Unmarshal(resp, &healthResp)
		if healthResp.Status != "OK" {
			t.Fatalf("expected status OK, got %s", healthResp.Status)
		}
	})

	t.Run("直接调用 HelloWorld 模块", func(t *testing.T) {
		target := tarsclient.Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HelloWorld",
		}

		req := &pb.HelloWorldRequest{Name: "DirectInvoker"}
		reqData, _ := proto.Marshal(req)

		code, resp, err := invoker.Invoke(context.Background(), target, reqData, nil)
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}

		var helloResp pb.HelloWorldResponse
		proto.Unmarshal(resp, &helloResp)
		if helloResp.Message != "Hello, DirectInvoker!" {
			t.Fatalf("expected 'Hello, DirectInvoker!', got %s", helloResp.Message)
		}
	})
}
