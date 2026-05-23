package domain

import "testing"

func TestPackVersion_IsNewerThan(t *testing.T) {
	tests := []struct {
		name     string
		version  PackVersion
		other    int
		expected bool
	}{
		{"版本 5 > 版本 3", PackVersion{LangCode: "zh-CN", Version: 5}, 3, true},
		{"版本 3 不大于版本 5", PackVersion{LangCode: "zh-CN", Version: 3}, 5, false},
		{"相同版本不大于", PackVersion{LangCode: "zh-CN", Version: 3}, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.IsNewerThan(tt.other); got != tt.expected {
				t.Errorf("PackVersion.IsNewerThan() = %v, want %v", got, tt.expected)
			}
		})
	}
}
