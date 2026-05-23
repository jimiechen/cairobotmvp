package service

import (
	"testing"
)

func TestApplyCompatFilter(t *testing.T) {
	entries := []LangStringEntry{
		{Key: "key1", Value: "确定", TemplateType: "plain"},
		{Key: "key2", Value: "欢迎 {name}", TemplateType: "named"},
		{Key: "key3", Value: "ICU message", TemplateType: "icu"},
	}

	tests := []struct {
		name          string
		clientVersion string
		expectedCount int
	}{
		{"老版本只返回 plain", "1.0.0", 1},
		{"新版本返回全部", "2.0.0", 3},
		{"更新版本返回全部", "3.0.0", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := ApplyCompatFilter(entries, tt.clientVersion)
			if len(filtered) != tt.expectedCount {
				t.Errorf("ApplyCompatFilter() returned %d entries, want %d", len(filtered), tt.expectedCount)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"v1 < v2", "1.0.0", "2.0.0", -1},
		{"v1 > v2", "2.0.0", "1.0.0", 1},
		{"v1 == v2", "1.0.0", "1.0.0", 0},
		{"v1 == v2 (不同格式)", "1.0", "1.0.0", 0},
		{"v1 < v2 (patch)", "1.0.0", "1.0.1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareVersions(tt.v1, tt.v2); got != tt.expected {
				t.Errorf("compareVersions() = %v, want %v", got, tt.expected)
			}
		})
	}
}
