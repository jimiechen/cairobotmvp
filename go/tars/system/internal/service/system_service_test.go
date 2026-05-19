package service

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

func TestSystemService_HealthCheck(t *testing.T) {
	svc := NewSystemService()
	respBytes, err := svc.HealthCheck(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(respBytes) == 0 {
		t.Fatal("expected non-empty response bytes")
	}

	var resp pb.ServiceHealthCheckResponse
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if resp.Status != "OK" {
		t.Fatalf("expected status 'OK', got %q", resp.Status)
	}
	if resp.Result == nil || resp.Result.Code != 10200 {
		t.Fatalf("expected Result.Code 10200, got %v", resp.Result)
	}
	if resp.Timestamp <= 0 {
		t.Fatalf("expected timestamp > 0, got %d", resp.Timestamp)
	}
}

func TestSystemService_HelloWorld(t *testing.T) {
	svc := NewSystemService()

	t.Run("with name", func(t *testing.T) {
		respBytes, err := svc.HelloWorld(context.Background(), "CaiRobot")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var resp pb.HelloWorldResponse
		if err := proto.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if resp.Message != "Hello, CaiRobot!" {
			t.Fatalf("expected message 'Hello, CaiRobot!', got %q", resp.Message)
		}
		if resp.Result == nil || resp.Result.Code != 10200 {
			t.Fatalf("expected Result.Code 10200, got %v", resp.Result)
		}
		if resp.Timestamp <= 0 {
			t.Fatalf("expected timestamp > 0, got %d", resp.Timestamp)
		}
	})

	t.Run("without name", func(t *testing.T) {
		respBytes, err := svc.HelloWorld(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var resp pb.HelloWorldResponse
		if err := proto.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if resp.Message != "Hello, World!" {
			t.Fatalf("expected message 'Hello, World!', got %q", resp.Message)
		}
	})
}
