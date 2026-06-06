package tarsclient

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/hello"
)

func TestRegisterModuleHandlers(t *testing.T) {
	invoker := NewLocalInvoker()
	RegisterModuleHandlers(invoker)

	t.Run("HealthCheck 模块化调用返回有效 Protobuf 响应", func(t *testing.T) {
		target := Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HealthCheck",
		}

		req := &pb.ServiceHealthCheckRequest{
			ServiceName: "TestService",
		}
		reqData, err := proto.Marshal(req)
		if err != nil {
			t.Fatalf("marshal request failed: %v", err)
		}

		code, resp, err := invoker.Invoke(context.Background(), target, reqData, nil)
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var healthResp pb.ServiceHealthCheckResponse
		if err := proto.Unmarshal(resp, &healthResp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if healthResp.Result.Code != 10200 {
			t.Fatalf("expected result code 10200, got %d", healthResp.Result.Code)
		}
		if healthResp.Status != "OK" {
			t.Fatalf("expected status OK, got %s", healthResp.Status)
		}
	})

	t.Run("HelloWorld 模块化调用返回有效 Protobuf 响应", func(t *testing.T) {
		target := Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HelloWorld",
		}

		req := &pb.HelloWorldRequest{
			Name: "ModularRefactor",
		}
		reqData, err := proto.Marshal(req)
		if err != nil {
			t.Fatalf("marshal request failed: %v", err)
		}

		code, resp, err := invoker.Invoke(context.Background(), target, reqData, nil)
		if err != nil {
			t.Fatalf("invoke failed: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var helloResp pb.HelloWorldResponse
		if err := proto.Unmarshal(resp, &helloResp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if helloResp.Result.Code != 10200 {
			t.Fatalf("expected result code 10200, got %d", helloResp.Result.Code)
		}
		if helloResp.Message != "Hello, ModularRefactor!" {
			t.Fatalf("expected message 'Hello, ModularRefactor!', got %s", helloResp.Message)
		}
	})

	t.Run("HelloWorld 默认名称为 World", func(t *testing.T) {
		target := Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HelloWorld",
		}

		req := &pb.HelloWorldRequest{}
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
		if helloResp.Message != "Hello, World!" {
			t.Fatalf("expected 'Hello, World!', got %s", helloResp.Message)
		}
	})

	t.Run("无效 Protobuf 输入返回错误", func(t *testing.T) {
		target := Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "HealthCheck",
		}

		code, resp, err := invoker.Invoke(context.Background(), target, []byte("invalid-protobuf"), nil)
		if err == nil {
			t.Fatal("expected error for invalid protobuf")
		}
		if code != 10500 {
			t.Fatalf("expected code 10500 for module error, got %d", code)
		}
		if resp != nil {
			t.Fatal("expected nil response on error")
		}
	})
}

func TestModuleHandlerAdapter(t *testing.T) {
	t.Run("模块服务成功时返回 10200", func(t *testing.T) {
		svc := hello.New(buildMinimalDeps())
		handler := NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
			return svc.SayHello(ctx, req)
		})

		req, _ := proto.Marshal(&pb.HelloWorldRequest{Name: "AdapterTest"})
		code, resp, err := handler.Invoke(context.Background(), req, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}
	})

	t.Run("模块服务失败时返回 10500", func(t *testing.T) {
		handler := NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
			return nil, fmt.Errorf("simulated module failure")
		})

		code, resp, err := handler.Invoke(context.Background(), []byte("test"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if code != 10500 {
			t.Fatalf("expected 10500, got %d", code)
		}
		if resp != nil {
			t.Fatal("expected nil response")
		}
	})
}
