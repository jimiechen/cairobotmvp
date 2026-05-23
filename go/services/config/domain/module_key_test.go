package domain

import "testing"

func TestAllStaticModuleKeys_应返回8个预定义模块(t *testing.T) {
	keys := AllStaticModuleKeys()
	if len(keys) != 8 {
		t.Fatalf("期望 8 个静态模块键，实际 %d", len(keys))
	}
	expectedSet := map[string]bool{
		ModuleKeyBase: true, ModuleKeyWap: true, ModuleKeyRegex: true,
		ModuleKeyPay: true, ModuleKeyOss: true, ModuleKeyLang: true,
		ModuleKeyMute: true, ModuleKeyGroup: true,
	}
	for _, k := range keys {
		if !expectedSet[k] {
			t.Errorf("意外的模块键: %s", k)
		}
		delete(expectedSet, k)
	}
	if len(expectedSet) > 0 {
		t.Errorf("缺少模块键: %v", expectedSet)
	}
}

func TestIsStaticModule_已知模块应返回true(t *testing.T) {
	testCases := []string{ModuleKeyBase, ModuleKeyWap, ModuleKeyRegex, ModuleKeyPay, ModuleKeyOss, ModuleKeyLang, ModuleKeyMute, ModuleKeyGroup}
	for _, tc := range testCases {
		if !IsStaticModule(tc) {
			t.Errorf("IsStaticModule(%q) 应返回 true", tc)
		}
	}
}

func TestIsStaticModule_未知模块应返回false(t *testing.T) {
	unknownKeys := []string{"custom_module", "", "dynamic_new", "notification_cfg"}
	for _, k := range unknownKeys {
		if IsStaticModule(k) {
			t.Errorf("IsStaticModule(%q) 应返回 false", k)
		}
	}
}

func Test常量值一致性(t *testing.T) {
	if ModuleKeyBase != "base_cfg" {
		t.Error("ModuleKeyBase 值不正确")
	}
	if ModuleKeyGroup != "group_cfg" {
		t.Error("ModuleKeyGroup 值不正确")
	}
}
