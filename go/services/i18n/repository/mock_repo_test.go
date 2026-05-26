package repository

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
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

// TestMockRepo_GetDiffSince 增量查询
func TestMockRepo_GetDiffSince(t *testing.T) {
	repo := NewMockRepo()
	repo.Strings[100] = []domain.LangString{
		{ID: 1, PackID: 100, StringKey: "old", StringValue: "旧值",
			Version: 1, OperationType: domain.OperationAdd},
		{ID: 2, PackID: 100, StringKey: "new", StringValue: "新值",
			Version: 3, OperationType: domain.OperationAdd},
		{ID: 3, PackID: 100, StringKey: "mod", StringValue: "修改后",
			Version: 3, OperationType: domain.OperationMod},
		{ID: 4, PackID: 100, StringKey: "del", StringValue: "已删除",
			Version: 2, OperationType: domain.OperationDel},
	}

	diff, err := repo.GetDiffSince(100, 2)
	if err != nil {
		t.Fatalf("GetDiffSince() error = %v", err)
	}
	if len(diff) != 2 {
		t.Errorf("期望 2 条 ADD/MOD 记录(version>2), 实际 %d", len(diff))
	}
}
