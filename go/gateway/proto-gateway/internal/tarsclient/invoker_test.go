package tarsclient

import (
	"context"
	"testing"
)

// mockHandler 用于测试
type mockHandler struct {
	returnCode int
	response   []byte
	err        error
}

func (m *mockHandler) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	return m.returnCode, m.response, m.err
}

func TestLocalInvoker(t *testing.T) {
	li := NewLocalInvoker()
	target := Target{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}
	key := ToTargetKey(target)

	t.Run("handler not found", func(t *testing.T) {
		code, resp, err := li.Invoke(context.Background(), target, []byte("test"), nil)
		if code != 10404 {
			t.Fatalf("expected code 10404, got %d", code)
		}
		if err == nil {
			t.Fatal("expected error")
		}
		if resp != nil {
			t.Fatal("expected nil response")
		}
	})

	t.Run("handler returns 10200", func(t *testing.T) {
		handler := &mockHandler{returnCode: 10200, response: []byte("ok")}
		li.Register(key, handler)

		code, resp, err := li.Invoke(context.Background(), target, []byte("test"), nil)
		if code != 10200 {
			t.Fatalf("expected code 10200, got %d", code)
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(resp) != "ok" {
			t.Fatalf("expected response 'ok', got %q", string(resp))
		}
	})

	t.Run("handler returns 10401", func(t *testing.T) {
		handler := &mockHandler{returnCode: 10401, response: []byte("unauthorized")}
		li.Register(key, handler)

		code, resp, err := li.Invoke(context.Background(), target, []byte("test"), nil)
		if code != 10401 {
			t.Fatalf("expected code 10401, got %d", code)
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(resp) != "unauthorized" {
			t.Fatalf("expected response 'unauthorized', got %q", string(resp))
		}
	})

	t.Run("handler returns 10500", func(t *testing.T) {
		handler := &mockHandler{returnCode: 10500, response: nil, err: nil}
		li.Register(key, handler)

		code, resp, err := li.Invoke(context.Background(), target, []byte("test"), nil)
		if code != 10500 {
			t.Fatalf("expected code 10500, got %d", code)
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp != nil {
			t.Fatal("expected nil response")
		}
	})
}

func TestRegisterSystemHandlers(t *testing.T) {
	invoker := NewLocalInvoker()
	RegisterSystemHandlers(invoker)

	tests := []struct {
		name   string
		target Target
	}{
		{
			name: "HealthCheck",
			target: Target{
				App:     "CaiRobot",
				Server:  "SystemServer",
				Servant: "SystemObj",
				Method:  "HealthCheck",
			},
		},
		{
			name: "HelloWorld",
			target: Target{
				App:     "CaiRobot",
				Server:  "SystemServer",
				Servant: "SystemObj",
				Method:  "HelloWorld",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, resp, err := invoker.Invoke(context.Background(), tt.target, []byte{}, map[string]string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != 10200 {
				t.Fatalf("expected code 10200, got %d", code)
			}
			if len(resp) == 0 {
				t.Fatal("expected non-empty response")
			}
		})
	}

	t.Run("unregistered handler returns 10404", func(t *testing.T) {
		unknownTarget := Target{
			App:     "CaiRobot",
			Server:  "SystemServer",
			Servant: "SystemObj",
			Method:  "UnknownMethod",
		}
		code, resp, err := invoker.Invoke(context.Background(), unknownTarget, []byte{}, nil)
		if code != 10404 {
			t.Fatalf("expected code 10404, got %d", code)
		}
		if err == nil {
			t.Fatal("expected error")
		}
		if resp != nil {
			t.Fatal("expected nil response")
		}
	})
}

func TestTarsGoInvoker(t *testing.T) {
	invoker := NewTarsGoInvoker()
	target := Target{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}

	code, resp, err := invoker.Invoke(context.Background(), target, []byte("test"), nil)
	if code != 10500 {
		t.Fatalf("expected code 10500, got %d", code)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
}
