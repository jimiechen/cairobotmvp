package domain

import "testing"

func TestLangPack_IsPublishedStatus(t *testing.T) {
	pack := &LangPack{
		ID:          1,
		PackName:    "webp",
		Env:         "dev",
		Version:     1,
		LangCode:    "zh-CN",
		Description: "简体中文",
		IsPublished: true,
	}

	if !pack.IsPublishedStatus() {
		t.Error("expected published pack to return true")
	}

	unpublishedPack := &LangPack{
		IsPublished: false,
	}

	if unpublishedPack.IsPublishedStatus() {
		t.Error("expected unpublished pack to return false")
	}
}
