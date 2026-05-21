package adapter

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/common-lib"
)

func TestSystemAdapter_Invoke(t *testing.T) {
	adapter := NewSystemAdapter()

	t.Run("HealthCheck 返回有效 Protobuf 响应", func(t *testing.T) {
		req := &pb.ServiceHealthCheckRequest{ServiceName: "TestService"}
		reqData, _ := proto.Marshal(req)

		code, resp, err := adapter.Invoke(context.Background(), reqData, map[string]string{"method": "HealthCheck"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != commonlib.CodeSuccess {
			t.Fatalf("expected code %d, got %d", commonlib.CodeSuccess, code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var healthResp pb.ServiceHealthCheckResponse
		if err := proto.Unmarshal(resp, &healthResp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if healthResp.Status != "OK" {
			t.Fatalf("expected status OK, got %s", healthResp.Status)
		}
	})

	t.Run("HelloWorld 返回有效 Protobuf 响应", func(t *testing.T) {
		req := &pb.HelloWorldRequest{Name: "AdapterTest"}
		reqData, _ := proto.Marshal(req)

		code, resp, err := adapter.Invoke(context.Background(), reqData, map[string]string{"method": "HelloWorld"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != commonlib.CodeSuccess {
			t.Fatalf("expected code %d, got %d", commonlib.CodeSuccess, code)
		}

		var helloResp pb.HelloWorldResponse
		proto.Unmarshal(resp, &helloResp)
		if helloResp.Message != "Hello, AdapterTest!" {
			t.Fatalf("expected 'Hello, AdapterTest!', got %s", helloResp.Message)
		}
	})

	t.Run("默认方法为 HealthCheck", func(t *testing.T) {
		reqData, _ := proto.Marshal(&pb.ServiceHealthCheckRequest{})

		code, resp, err := adapter.Invoke(context.Background(), reqData, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != commonlib.CodeSuccess {
			t.Fatalf("expected default HealthCheck to succeed, got code %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}
	})

	t.Run("未知方法返回 10404", func(t *testing.T) {
		code, resp, err := adapter.Invoke(context.Background(), []byte{}, map[string]string{"method": "UnknownMethod"})
		if err == nil {
			t.Fatal("expected error for unknown method")
		}
		if code != commonlib.CodeNotFound {
			t.Fatalf("expected code %d, got %d", commonlib.CodeNotFound, code)
		}
		if resp != nil {
			t.Fatal("expected nil response on error")
		}
	})
}
