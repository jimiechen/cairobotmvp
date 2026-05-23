package service

import (
	"testing"
)

func TestGetLangDifference(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.GetLangDifference("zh-CN", 0, "1.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangDifference() error = %v", err)
	}

	if resp.CurrentVersion != 1 {
		t.Errorf("CurrentVersion = %d, want %d", resp.CurrentVersion, 1)
	}
}

func TestGetLangDifference_NotFound(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.GetLangDifference("ja-JP", 0, "1.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangDifference() error = %v", err)
	}

	if len(resp.Additions) != 0 || len(resp.Deletions) != 0 {
		t.Error("GetLangDifference() should return empty for non-existing lang code")
	}
}
