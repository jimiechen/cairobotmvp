package cache

import (
	"testing"
)

func TestMockCache_SetAndGet_命中(t *testing.T) {
	c := NewMockCache()
	c.Set("cfg:prod:base_cfg", "value123")

	got, found := c.Get("cfg:prod:base_cfg")
	if !found {
		t.Fatal("期望命中缓存")
	}
	if got != "value123" {
		t.Errorf("期望 value123, 实际 %v", got)
	}
}

func TestMockCache_Get_未命中(t *testing.T) {
	c := NewMockCache()
	_, found := c.Get("nonexistent")
	if found {
		t.Error("不存在的键不应命中")
	}
}

func TestMockCache_Delete(t *testing.T) {
	c := NewMockCache()
	c.Set("key1", "val1")
	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("删除后不应命中")
	}
}

func TestMockCache_Invalidate_按前缀清除(t *testing.T) {
	c := NewMockCache()
	c.Set("cfg:prod:base_cfg", "v1")
	c.Set("cfg:prod:wap_cfg", "v2")
	c.Set("other:key", "v3")

	c.Invalidate("cfg:prod:")

	if _, found := c.Get("cfg:prod:base_cfg"); found {
		t.Error("base_cfg 应被清除")
	}
	if _, found := c.Get("cfg:prod:wap_cfg"); found {
		t.Error("wap_cfg 应被清除")
	}
	if _, found := c.Get("other:key"); !found {
		t.Error("无此前缀的键不应被清除")
	}
}

func TestMockCache_Size(t *testing.T) {
	c := NewMockCache()
	if c.Size() != 0 {
		t.Errorf("空缓存大小应为 0, 实际 %d", c.Size())
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Size() != 2 {
		t.Errorf("期望大小 2, 实际 %d", c.Size())
	}
}
