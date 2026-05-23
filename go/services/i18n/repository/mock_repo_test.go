package repository

import (
	"testing"
)

func TestMockRepo_GetPackByLangCode(t *testing.T) {
	repo := SetupMockRepoWithSeedData()

	pack, err := repo.GetPackByLangCode("zh-CN", "dev")
	if err != nil {
		t.Fatalf("GetPackByLangCode() error = %v", err)
	}

	if pack == nil {
		t.Fatal("GetPackByLangCode() returned nil for existing lang code")
	}

	if pack.LangCode != "zh-CN" {
		t.Errorf("LangCode = %v, want %v", pack.LangCode, "zh-CN")
	}
}

func TestMockRepo_GetPackByLangCode_NotFound(t *testing.T) {
	repo := NewMockRepo()

	pack, err := repo.GetPackByLangCode("ja-JP", "dev")
	if err != nil {
		t.Fatalf("GetPackByLangCode() error = %v", err)
	}

	if pack != nil {
		t.Error("GetPackByLangCode() should return nil for non-existing lang code")
	}
}

func TestMockRepo_GetStringsByPackID(t *testing.T) {
	repo := SetupMockRepoWithSeedData()

	strings, err := repo.GetStringsByPackID(1)
	if err != nil {
		t.Fatalf("GetStringsByPackID() error = %v", err)
	}

	if len(strings) != 3 {
		t.Errorf("GetStringsByPackID() returned %d strings, want 3", len(strings))
	}
}

func TestMockRepo_ListPacks(t *testing.T) {
	repo := SetupMockRepoWithSeedData()

	packs, err := repo.ListPacks("dev")
	if err != nil {
		t.Fatalf("ListPacks() error = %v", err)
	}

	if len(packs) != 2 {
		t.Errorf("ListPacks() returned %d packs, want 2", len(packs))
	}
}
