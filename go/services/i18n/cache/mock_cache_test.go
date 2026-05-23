package cache

import (
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

func TestMockCache_GetSetPack(t *testing.T) {
	cache := NewMockCache()

	pack := &domain.LangPack{
		ID:          1,
		PackName:    "webp",
		LangCode:    "zh-CN",
		Version:     1,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	cache.SetPack("zh-CN", "dev", pack)

	got, exists := cache.GetPack("zh-CN", "dev")
	if !exists {
		t.Fatal("GetPack() should return true for existing key")
	}

	if got.ID != 1 || got.LangCode != "zh-CN" {
		t.Error("GetPack() returned wrong data")
	}
}

func TestMockCache_GetPack_NotFound(t *testing.T) {
	cache := NewMockCache()

	_, exists := cache.GetPack("ja-JP", "dev")
	if exists {
		t.Error("GetPack() should return false for non-existing key")
	}
}

func TestMockCache_GetSetStrings(t *testing.T) {
	cache := NewMockCache()

	strings := []domain.LangString{
		{
			ID:          1,
			StringKey:   "svc_common_ok",
			StringValue: "确定",
			TemplateType: domain.TemplatePlain,
		},
	}

	cache.SetStrings(1, strings)

	got, exists := cache.GetStrings(1)
	if !exists {
		t.Fatal("GetStrings() should return true for existing key")
	}

	if len(got) != 1 || got[0].StringKey != "svc_common_ok" {
		t.Error("GetStrings() returned wrong data")
	}
}

func TestMockCache_Invalidate(t *testing.T) {
	cache := NewMockCache()

	pack := &domain.LangPack{ID: 1, LangCode: "zh-CN"}
	cache.SetPack("zh-CN", "dev", pack)

	cache.Invalidate("zh-CN", "dev")

	_, exists := cache.GetPack("zh-CN", "dev")
	if exists {
		t.Error("Invalidate() should remove the cached pack")
	}
}
