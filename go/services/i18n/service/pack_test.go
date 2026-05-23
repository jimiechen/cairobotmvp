package service

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/cache"
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
)

func setupTestService(t *testing.T) *I18nServiceImpl {
	t.Helper()

	repo := repository.SetupMockRepoWithSeedData()
	c := cache.NewMockCache()
	svc := NewI18nService(repo, c, "dev")

	return svc
}

func TestGetLangPack(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.GetLangPack("zh-CN", "2.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangPack() error = %v", err)
	}

	if resp.PackVersion != 1 {
		t.Errorf("PackVersion = %d, want %d", resp.PackVersion, 1)
	}

	if len(resp.Strings) != 3 {
		t.Errorf("Strings count = %d, want %d", len(resp.Strings), 3)
	}
}

func TestGetLangPack_NotFound(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.GetLangPack("ja-JP", "1.0.0", "dev")
	if err != nil {
		t.Fatalf("GetLangPack() error = %v", err)
	}

	if len(resp.Strings) != 0 {
		t.Errorf("Strings count should be 0 for non-existing lang code, got %d", len(resp.Strings))
	}
}

func TestConvertToEntries(t *testing.T) {
	strings := []domain.LangString{
		{
			ID:            1,
			StringKey:     domain.StringKey("svc_common_ok"),
			StringValue:   "确定",
			TemplateType:  domain.TemplatePlain,
			OperationType: domain.OperationAdd,
		},
		{
			ID:            2,
			StringKey:     domain.StringKey("svc_msg_welcome"),
			StringValue:   "欢迎 {name}，你有 {count} 条消息",
			TemplateType:  domain.TemplateNamed,
			ParamsSchema:  `[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`,
			OperationType: domain.OperationAdd,
		},
	}

	entries := convertToEntries(strings)

	if len(entries) != 2 {
		t.Fatalf("convertToEntries() returned %d entries, want 2", len(entries))
	}

	if entries[0].Key != "svc_common_ok" || entries[0].Value != "确定" {
		t.Error("convertToEntries() returned wrong data for plain entry")
	}

	if entries[1].TemplateType != "named" || len(entries[1].Params) != 2 {
		t.Error("convertToEntries() returned wrong data for named entry")
	}
}
