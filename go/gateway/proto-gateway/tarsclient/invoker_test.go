package tarsclient

import (
	"context"
	"encoding/json"
	"testing"

	configservice "github.com/jimiechen/mineplanet/go/services/config/service"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
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

// mustMarshalProto 序列化 Protobuf 消息，失败时 panic（仅用于测试数据构造）
func mustMarshalProto(m proto.Message) []byte {
	data, err := proto.Marshal(m)
	if err != nil {
		panic("proto.Marshal failed: " + err.Error())
	}
	return data
}

// TestRegisterAllLocalHandlers 验证 System + Config + I18n 全量 handler 注册
// 修复 BUG-E2E-001: 确保 local 模式下 6000 段协议不再返回 10404
func TestRegisterAllLocalHandlers(t *testing.T) {
	invoker := NewLocalInvoker()
	RegisterAllLocalHandlers(invoker)

	// 定义全部已注册的 handler 目标键（2 System + 5 Config/I18n）
	expectedTargets := []struct {
		name   string
		target Target
		// System handler 用 Protobuf bytes，Config/I18n 用 JSON
		input []byte
	}{
		{"HealthCheck (2100:2097)", Target{App: "CaiRobot", Server: "SystemServer", Servant: "SystemObj", Method: "HealthCheck"}, mustMarshalProto(&pb.ServiceHealthCheckRequest{})},
		{"HelloWorld (2100:2101)", Target{App: "CaiRobot", Server: "SystemServer", Servant: "SystemObj", Method: "HelloWorld"}, mustMarshalProto(&pb.HelloWorldRequest{Name: "test"})},
		{"GetAppConfigs (6000:6001)", Target{App: "CaiRobot", Server: "ConfigServer", Servant: "ConfigObj", Method: "GetAppConfigs"}, []byte(`{}`)},
		{"AppConfigVersion (6000:6009)", Target{App: "CaiRobot", Server: "ConfigServer", Servant: "ConfigObj", Method: "AppConfigVersion"}, []byte(`{}`)},
		{"GetAppLanguage (6000:6003)", Target{App: "CaiRobot", Server: "I18nServer", Servant: "I18nObj", Method: "GetAppLanguage"}, []byte(`{}`)},
		{"GetLangPack (6000:6005)", Target{App: "CaiRobot", Server: "I18nServer", Servant: "I18nObj", Method: "GetLangPack"}, []byte(`{}`)},
		{"GetLangDifference (6000:6007)", Target{App: "CaiRobot", Server: "I18nServer", Servant: "I18nObj", Method: "GetLangDifference"}, []byte(`{}`)},
	}

	// 验证每个目标都有对应 handler，且调用后不再返回 10404
	for _, tt := range expectedTargets {
		t.Run(tt.name+" 已注册且不返回 10404", func(t *testing.T) {
			code, resp, err := invoker.Invoke(context.Background(), tt.target, tt.input, nil)
			if err != nil {
				t.Fatalf("invoke error (handler may be missing): %v", err)
			}
			// 核心断言：绝不能返回 10404（handler not found）
			if code == 10404 {
				t.Fatalf("handler not found — 返回 10404，目标: %s", ToTargetKey(tt.target).String())
			}
			// 响应体必须非空（noop stub 返回有效数据）
			if len(resp) == 0 {
				t.Fatal("expected non-empty response from stub")
			}
			// Config/I18n handler 返回 JSON，额外校验格式
			if tt.target.Server == "ConfigServer" || tt.target.Server == "I18nServer" {
				var raw json.RawMessage
				if json.Unmarshal(resp, &raw) != nil {
					t.Fatalf("response is not valid JSON: %s", string(resp))
				}
			}
			t.Logf("✅ %s → code=%d, resp_size=%d", tt.name, code, len(resp))
		})
	}

	// 验证未注册的 target 仍然返回 10404
	t.Run("未注册的 target 仍返回 10404", func(t *testing.T) {
		unknownTarget := Target{
			App:     "CaiRobot",
			Server:  "UnknownServer",
			Servant: "UnknownObj",
			Method:  "UnknownMethod",
		}
		code, resp, err := invoker.Invoke(context.Background(), unknownTarget, []byte("{}"), nil)
		if code != 10404 {
			t.Fatalf("expected 10404 for unregistered target, got %d", code)
		}
		if err == nil {
			t.Fatal("expected error for unregistered target")
		}
		if resp != nil {
			t.Fatal("expected nil response for unregistered target")
		}
	})
}

// TestNoopConfigServiceStub 验证 noop Config stub 返回值结构正确
func TestNoopConfigServiceStub(t *testing.T) {
	svc := &noopConfigService{}

	resp, err := svc.GetAppConfigs(&configservice.AppConfigRequest{Env: "dev"})
	if err != nil {
		t.Fatalf("GetAppConfigs error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.StaticModules == nil {
		t.Fatal("StaticModules should not be nil")
	}
	if resp.DynamicModules == nil {
		t.Fatal("DynamicModules should not be nil")
	}

	verResp, err := svc.GetVersionInfo("dev", nil)
	if err != nil {
		t.Fatalf("GetVersionInfo error: %v", err)
	}
	if verResp.HasChanges {
		t.Fatal("noop should report no changes")
	}
}

// TestNoopI18nServiceStub 验证 noop I18n stub 返回默认语言列表
func TestNoopI18nServiceStub(t *testing.T) {
	svc := &noopI18nService{}

	languages, err := svc.GetLanguages("1.0.0")
	if err != nil {
		t.Fatalf("GetLanguages error: %v", err)
	}
	if len(languages) == 0 {
		t.Fatal("expected at least 1 default language")
	}
	foundZhCN := false
	for _, lang := range languages {
		if lang.Code == "zh-CN" {
			foundZhCN = true
			break
		}
	}
	if !foundZhCN {
		t.Fatal("expected zh-CN in default languages")
	}

	pack, err := svc.GetLangPack("zh-CN", "1.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangPack error: %v", err)
	}
	if pack == nil {
		t.Fatal("expected non-nil pack response")
	}

	diff, err := svc.GetLangDifference("zh-CN", 1, "1.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangDifference error: %v", err)
	}
	if diff == nil {
		t.Fatal("expected non-nil diff response")
	}
}
