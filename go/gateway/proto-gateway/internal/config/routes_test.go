package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRoutes(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid routes",
			content: `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: ServiceHealthCheck
    request_proto: com.mineplanet.pojo.health.ServiceHealthCheckRequest
    response_max: 2100
    response_min: 2098
    response_proto: com.mineplanet.pojo.health.ServiceHealthCheckResponse
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: HealthCheck
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false`,
			wantErr: false,
		},
		{
			name: "duplicate request_max/request_min",
			content: `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: A
    request_proto: proto.A
    response_max: 2100
    response_min: 2098
    response_proto: proto.B
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: A
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: B
    request_proto: proto.C
    response_max: 2100
    response_min: 2099
    response_proto: proto.D
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: B
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false`,
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name: "route_key mismatch",
			content: `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2098"
    command_name: A
    request_proto: proto.A
    response_max: 2100
    response_min: 2098
    response_proto: proto.B
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: A
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false`,
			wantErr: true,
			errMsg:  "route_key",
		},
		{
			name: "invalid tars_request_type",
			content: `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: A
    request_proto: proto.A
    response_max: 2100
    response_min: 2098
    response_proto: proto.B
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: A
    tars_request_type: string
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false`,
			wantErr: true,
			errMsg:  "vector<byte>",
		},
		{
			name: "timeout_ms <= 0",
			content: `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: A
    request_proto: proto.A
    response_max: 2100
    response_min: 2098
    response_proto: proto.B
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: A
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 0
    auth_required: false
    audit_required: false`,
			wantErr: true,
			errMsg:  "timeout_ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "routes.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("write test file failed: %v", err)
			}

			cfg, err := LoadRoutes(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})
	}
}

func TestLoadRoutesWithEnv(t *testing.T) {
	validContent := `routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    command_name: ServiceHealthCheck
    request_proto: com.mineplanet.pojo.health.ServiceHealthCheckRequest
    response_max: 2100
    response_min: 2098
    response_proto: com.mineplanet.pojo.health.ServiceHealthCheckResponse
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: HealthCheck
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
    timeout_ms: 3000
    auth_required: false
    audit_required: false`

	t.Run("env overrides default", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "env_routes.yaml")
		if err := os.WriteFile(envPath, []byte(validContent), 0644); err != nil {
			t.Fatalf("write test file failed: %v", err)
		}
		t.Setenv("GATEWAY_ROUTES_PATH", envPath)

		cfg, err := LoadRoutesWithEnv("/nonexistent/default.yaml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil || len(cfg.Routes) != 1 {
			t.Fatalf("expected 1 route, got %v", cfg)
		}
	})

	t.Run("fallback to default when env empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		defaultPath := filepath.Join(tmpDir, "default_routes.yaml")
		if err := os.WriteFile(defaultPath, []byte(validContent), 0644); err != nil {
			t.Fatalf("write test file failed: %v", err)
		}
		t.Setenv("GATEWAY_ROUTES_PATH", "")

		cfg, err := LoadRoutesWithEnv(defaultPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil || len(cfg.Routes) != 1 {
			t.Fatalf("expected 1 route, got %v", cfg)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
