package domain

import "testing"
import "time"

func TestConfigVersion_IsReleased_已发布且有发布时间应返回true(t *testing.T) {
	now := time.Now()
	v := &ConfigVersion{
		IsPublished: true,
		PublishedAt: &now,
	}
	if !v.IsReleased() {
		t.Error("已发布且有 PublishedAt 的版本应返回 true")
	}
}

func TestConfigVersion_IsReleased_未发布应返回false(t *testing.T) {
	now := time.Now()
	v := &ConfigVersion{
		IsPublished: false,
		PublishedAt: &now,
	}
	if v.IsReleased() {
		t.Error("未发布的版本应返回 false")
	}
}

func TestConfigVersion_IsReleased_PublishedAt为nil应返回false(t *testing.T) {
	v := &ConfigVersion{
		IsPublished: true,
		PublishedAt: nil,
	}
	if v.IsReleased() {
		t.Error("PublishedAt 为 nil 时即使 IsPublished=true 也应返回 false")
	}
}
