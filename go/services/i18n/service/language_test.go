package service

import (
	"testing"
)

func TestGetLanguages(t *testing.T) {
	svc := setupTestService(t)

	languages, err := svc.GetLanguages("1.0.0")
	if err != nil {
		t.Fatalf("GetLanguages() error = %v", err)
	}

	if len(languages) != 2 {
		t.Errorf("GetLanguages() returned %d languages, want 2", len(languages))
	}

	foundZhCN := false
	foundEn := false
	for _, lang := range languages {
		if lang.Code == "zh-CN" {
			foundZhCN = true
			if !lang.IsDefault {
				t.Error("zh-CN should be default language")
			}
		}
		if lang.Code == "en" {
			foundEn = true
		}
	}

	if !foundZhCN || !foundEn {
		t.Error("GetLanguages() should return zh-CN and en")
	}
}

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		langCode string
		expected string
	}{
		{"zh-CN", "Chinese (Simplified)"},
		{"en", "English"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.langCode, func(t *testing.T) {
			if got := getLanguageName(tt.langCode); got != tt.expected {
				t.Errorf("getLanguageName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetNativeLanguageName(t *testing.T) {
	tests := []struct {
		langCode string
		expected string
	}{
		{"zh-CN", "简体中文"},
		{"en", "English"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.langCode, func(t *testing.T) {
			if got := getNativeLanguageName(tt.langCode); got != tt.expected {
				t.Errorf("getNativeLanguageName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
