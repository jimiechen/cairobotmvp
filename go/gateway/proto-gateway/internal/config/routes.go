package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Route 定义单条路由配置
type Route struct {
	RequestMax       int32  `yaml:"request_max"`
	RequestMin       int32  `yaml:"request_min"`
	RouteKey         string `yaml:"route_key"`
	CommandName      string `yaml:"command_name"`
	Description      string `yaml:"description"`
	RequestProto     string `yaml:"request_proto"`
	ResponseMax      int32  `yaml:"response_max"`
	ResponseMin      int32  `yaml:"response_min"`
	ResponseProto    string `yaml:"response_proto"`
	TarsApp          string `yaml:"tars_app"`
	TarsServer       string `yaml:"tars_server"`
	TarsServant      string `yaml:"tars_servant"`
	TarsModule       string `yaml:"tars_module"`
	TarsInterface    string `yaml:"tars_interface"`
	TarsMethod       string `yaml:"tars_method"`
	TarsRequestType  string `yaml:"tars_request_type"`
	TarsResponseType string `yaml:"tars_response_type"`
	TimeoutMs        int32  `yaml:"timeout_ms"`
	AuthRequired     bool   `yaml:"auth_required"`
	AuditRequired    bool   `yaml:"audit_required"`
}

// RoutesConfig 定义 routes.yaml 根结构
type RoutesConfig struct {
	Routes []Route `yaml:"routes"`
}

// LoadRoutes 从指定路径加载路由配置
// 支持环境变量 GATEWAY_ROUTES_PATH 覆盖默认路径
func LoadRoutes(path string) (*RoutesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read routes.yaml failed: %w", err)
	}

	var cfg RoutesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal routes.yaml failed: %w", err)
	}

	if err := validateRoutes(&cfg); err != nil {
		return nil, fmt.Errorf("validate routes.yaml failed: %w", err)
	}

	return &cfg, nil
}

// LoadRoutesWithEnv 从环境变量或默认路径加载路由配置
// 优先使用 GATEWAY_ROUTES_PATH 环境变量，否则使用传入的默认路径
func LoadRoutesWithEnv(defaultPath string) (*RoutesConfig, error) {
	path := os.Getenv("GATEWAY_ROUTES_PATH")
	if path == "" {
		path = defaultPath
	}
	return LoadRoutes(path)
}

// validateRoutes 校验路由配置
func validateRoutes(cfg *RoutesConfig) error {
	seen := make(map[string]bool)
	for i, r := range cfg.Routes {
		key := fmt.Sprintf("%d:%d", r.RequestMax, r.RequestMin)

		if seen[key] {
			return fmt.Errorf("route %d: duplicate request_max/request_min: %s", i, key)
		}
		seen[key] = true

		if r.RouteKey != key {
			return fmt.Errorf("route %d: route_key %s must equal %s", i, r.RouteKey, key)
		}

		if r.ResponseMax == 0 || r.ResponseMin == 0 {
			return fmt.Errorf("route %d: response_max/response_min required", i)
		}

		if r.RequestProto == "" || r.ResponseProto == "" {
			return fmt.Errorf("route %d: request_proto/response_proto required", i)
		}

		if r.TarsApp == "" || r.TarsServer == "" || r.TarsServant == "" ||
			r.TarsModule == "" || r.TarsInterface == "" || r.TarsMethod == "" {
			return fmt.Errorf("route %d: tars target fields required", i)
		}

		if r.TarsRequestType != "vector<byte>" || r.TarsResponseType != "vector<byte>" {
			return fmt.Errorf("route %d: tars_request_type/tars_response_type must be vector<byte>", i)
		}

		if r.TimeoutMs <= 0 {
			return fmt.Errorf("route %d: timeout_ms must > 0", i)
		}
	}

	return nil
}
