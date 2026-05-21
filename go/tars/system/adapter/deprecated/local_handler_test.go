package localhandler

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

func TestHandler_Invoke(t *testing.T) {
	handler := NewHandler()

	t.Run("HealthCheck returns Protobuf bytes", func(t *testing.T) {
		code, resp, err := handler.Invoke(context.Background(), []byte{}, map[string]string{"method": "HealthCheck"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var healthResp pb.ServiceHealthCheckResponse
		if err := proto.Unmarshal(resp, &healthResp); err != nil {
			t.Fatalf("expected Protobuf bytes, unmarshal failed: %v", err)
		}
		if healthResp.Status != "OK" {
			t.Fatalf("expected status 'OK', got %q", healthResp.Status)
		}
		if healthResp.Result == nil || healthResp.Result.Code != 10200 {
			t.Fatalf("expected Result.Code 10200, got %v", healthResp.Result)
		}
	})

	t.Run("HelloWorld returns Protobuf bytes", func(t *testing.T) {
		code, resp, err := handler.Invoke(context.Background(), []byte{}, map[string]string{"method": "HelloWorld"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var helloResp pb.HelloWorldResponse
		if err := proto.Unmarshal(resp, &helloResp); err != nil {
			t.Fatalf("expected Protobuf bytes, unmarshal failed: %v", err)
		}
		if helloResp.Message != "Hello, World!" {
			t.Fatalf("expected message 'Hello, World!', got %q", helloResp.Message)
		}
		if helloResp.Result == nil || helloResp.Result.Code != 10200 {
			t.Fatalf("expected Result.Code 10200, got %v", helloResp.Result)
		}
	})

	t.Run("unknown method", func(t *testing.T) {
		code, resp, err := handler.Invoke(context.Background(), []byte{}, map[string]string{"method": "Unknown"})
		if err == nil {
			t.Fatal("expected error")
		}
		if code != 10404 {
			t.Fatalf("expected code 10404, got %d", code)
		}
		if resp != nil {
			t.Fatal("expected nil response")
		}
	})

	t.Run("default method is HealthCheck", func(t *testing.T) {
		code, resp, err := handler.Invoke(context.Background(), []byte{}, map[string]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if len(resp) == 0 {
			t.Fatal("expected non-empty response")
		}

		var healthResp pb.ServiceHealthCheckResponse
		if err := proto.Unmarshal(resp, &healthResp); err != nil {
			t.Fatalf("expected Protobuf bytes, unmarshal failed: %v", err)
		}
		if healthResp.Status != "OK" {
			t.Fatalf("expected status 'OK', got %q", healthResp.Status)
		}
	})
}
