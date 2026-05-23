package domain

import "testing"

func TestMatchClientScope_all应匹配任意客户端(t *testing.T) {
	f := &FieldSchema{ClientScope: "all"}
	scopes := []string{"android", "ios", "web", "admin"}
	for _, s := range scopes {
		if !f.MatchClientScope(s) {
			t.Errorf("client_scope=all 应匹配 %s", s)
		}
	}
}

func TestMatchClientScope_精确匹配(t *testing.T) {
	f := &FieldSchema{ClientScope: "android"}
	if !f.MatchClientScope("android") {
		t.Error("精确匹配 android 应返回 true")
	}
	if f.MatchClientScope("ios") {
		t.Error("android scope 不应匹配 ios")
	}
}

func TestModuleSchema_FindField_找到应返回对应schema(t *testing.T) {
	ms := &ModuleSchema{
		Fields: []*FieldSchema{
			{FieldKey: "domain_root"},
			{FieldKey: "sign_rand"},
		},
	}
	found := ms.FindField("sign_rand")
	if found == nil || found.FieldKey != "sign_rand" {
		t.Error("应找到 sign_rand 字段")
	}
}

func TestModuleSchema_FindField_找不到应返回nil(t *testing.T) {
	ms := &ModuleSchema{Fields: []*FieldSchema{{FieldKey: "a"}}}
	if ms.FindField("nonexistent") != nil {
		t.Error("不存在的字段应返回 nil")
	}
}

func TestModuleSchema_EnabledFields_过滤禁用字段(t *testing.T) {
	ms := &ModuleSchema{
		Fields: []*FieldSchema{
			{FieldKey: "a", IsEnabled: true},
			{FieldKey: "b", IsEnabled: false},
			{FieldKey: "c", IsEnabled: true},
		},
	}
	enabled := ms.EnabledFields()
	if len(enabled) != 2 {
		t.Fatalf("期望 2 个启用的字段，实际 %d", len(enabled))
	}
	if enabled[0].FieldKey != "a" || enabled[1].FieldKey != "c" {
		t.Error("启用的字段内容不正确")
	}
}
