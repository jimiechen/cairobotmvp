package config

import (
	"testing"
)

// TestDebug_LoadedRoutes 验证 routes.yaml 是否正确加载了 Social 路由
func TestDebug_LoadedRoutes(t *testing.T) {
	cfg, err := LoadRoutes("/Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp/configs/gateway/routes.yaml")
	if err != nil {
		t.Fatalf("LoadRoutesWithEnv failed: %v", err)
	}

	t.Logf("=== 已加载路由总数: %d ===", len(cfg.Routes))

	// 检查 Social 关键路由是否存在
	socialKeys := map[string]string{
		"1000:1021": "UserRegister",
		"1000:1025": "UserLogin",
		"2000:2005": "CreateGroup",
		"2000:2013": "JoinGroup",
		"3000:3001": "CreateTopic",
	}

	t.Logf("\n=== Social 路由检查 ===")
	for expectedKey, name := range socialKeys {
		found := false
		for _, r := range cfg.Routes {
			key := r.RouteKey // 已经是 "max:min" 格式
			if key == expectedKey {
				found = true
				t.Logf("  ✅ %s → %s (tars_method=%s)", key, name, r.TarsMethod)
				break
			}
		}
		if !found {
			t.Errorf("  ❌ %s (%s) — 未找到！", expectedKey, name)
		}
	}

	// 列出所有已加载的路由 key
	t.Logf("\n=== 全部已加载路由 ===")
	for i, r := range cfg.Routes {
		t.Logf("  [%2d] %s → %s | %s.%s.%s:%s",
			i+1, r.RouteKey, r.CommandName,
			r.TarsApp, r.TarsServer, r.TarsServant, r.TarsMethod)
	}
}
