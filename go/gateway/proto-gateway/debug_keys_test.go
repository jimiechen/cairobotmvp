package main

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/gateway/proto-gateway/tarsclient"
)

// TestDebug_RegisteredKeys 打印 LocalInvoker 中所有已注册的 handler key
// 用于排查 Social handler 是否成功注册
func TestDebug_RegisteredKeys(t *testing.T) {
	invoker := tarsclient.NewLocalInvoker()
	tarsclient.RegisterAllLocalHandlers(invoker)

	t.Logf("=== 已注册的 LocalInvoker Handler Keys ===")
	count := 0
	for key := range invoker.HandlersForTest() {
		count++
		t.Logf("  [%d] %s", count, key)
	}

	// 检查 Social 关键 key 是否存在
	socialKeys := []string{
		"CaiRobot.SocialServer.SocialObj.HandleMember",
		"CaiRobot.SocialServer.SocialObj.HandleGroup",
		"CaiRobot.SocialServer.SocialObj.HandleTopic",
	}

	t.Logf("\n=== Social Handler 注册检查 ===")
	for _, expectedKey := range socialKeys {
		if _, ok := invoker.HandlersForTest()[expectedKey]; ok {
			t.Logf("  ✅ %s — 已注册", expectedKey)
		} else {
			t.Fatalf("  ❌ %s — 未注册！这是 P0 阻塞项", expectedKey)
		}
	}

	t.Logf("\n总计注册 %d 个 handler", count)
}
