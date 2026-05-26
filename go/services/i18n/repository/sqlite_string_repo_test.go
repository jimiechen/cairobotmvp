package repository

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// TestGetStringsByPackID 正常返回多行
func TestGetStringsByPackID_正常返回多行(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "测试包", Env: "dev", Version: 1, LangCode: "zh-CN", IsPublished: true,
	})

	prevValue := "旧值"
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "greeting", StringValue: "你好",
		Version: 1, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "farewell", StringValue: "再见",
		GroupName: "common", Version: 2, OperationType: domain.OperationMod,
		PrevValue: &prevValue,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "confirm", StringValue: "确认",
		GroupName: "action", Version: 1, OperationType: domain.OperationDel,
	})

	strings, err := repo.GetStringsByPackID(packID)
	if err != nil {
		t.Fatalf("GetStringsByPackID 失败: %v", err)
	}
	if len(strings) != 2 {
		t.Fatalf("期望 2 条非删除记录, 实际 %d", len(strings))
	}
}

// TestGetStringsByPackID 空包返回空切片
func TestGetStringsByPackID_空包返回空(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "空包", Env: "dev", Version: 1, LangCode: "en-US", IsPublished: true,
	})

	strings, err := repo.GetStringsByPackID(packID)
	if err != nil {
		t.Fatalf("GetStringsByPackID 失败: %v", err)
	}
	if len(strings) != 0 {
		t.Errorf("期望 0 条记录, 实际 %d", len(strings))
	}
}

// TestGetStringsByPackID 不存在包返回空
func TestGetStringsByPackID_不存在包返回空(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	strings, err := repo.GetStringsByPackID(99999)
	if err != nil {
		t.Fatalf("GetStringsByPackID 失败: %v", err)
	}
	if len(strings) != 0 {
		t.Errorf("不存在的包应返回空, 实际 %d 条", len(strings))
	}
}

// TestGetStringsByPackID 验证字段映射
func TestGetStringsByPackID_验证字段映射(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "字段测试", Env: "dev", Version: 1, LangCode: "zh-CN", IsPublished: true,
	})

	paramsSchema := `{"params":["name"]}`
	previewSample := "你好，{{name}}！"
	insertTestString(t, repo.DB(), domain.LangString{
		PackID:        packID,
		StringKey:     "welcome",
		StringValue:   "欢迎，{{name}}！",
		GroupName:     "greeting",
		Version:       3,
		OperationType: domain.OperationMod,
		PrevValue:     strPtr("欢迎"),
		TemplateType:  domain.TemplateNamed,
		ParamsSchema:  paramsSchema,
		PreviewSample: previewSample,
	})

	strings, _ := repo.GetStringsByPackID(packID)
	if len(strings) != 1 {
		t.Fatalf("期望 1 条记录, 实际 %d", len(strings))
	}

	s := strings[0]
	if s.StringKey != "welcome" {
		t.Errorf("StringKey 不匹配: %s", s.StringKey)
	}
	if s.StringValue != "欢迎，{{name}}！" {
		t.Errorf("StringValue 不匹配: %s", s.StringValue)
	}
	if s.GroupName != "greeting" {
		t.Errorf("GroupName 不匹配: %s", s.GroupName)
	}
	if s.Version != 3 {
		t.Errorf("Version 不匹配: %d", s.Version)
	}
	if s.OperationType != domain.OperationMod {
		t.Errorf("OperationType 不匹配: %s", s.OperationType)
	}
	if s.PrevValue == nil || *s.PrevValue != "欢迎" {
		t.Errorf("PrevValue 不匹配: %v", s.PrevValue)
	}
	if s.TemplateType != domain.TemplateNamed {
		t.Errorf("TemplateType 不匹配: %v", s.TemplateType)
	}
	if s.ParamsSchema != paramsSchema {
		t.Errorf("ParamsSchema 不匹配: %s", s.ParamsSchema)
	}
	if s.PreviewSample != previewSample {
		t.Errorf("PreviewSample 不匹配: %s", s.PreviewSample)
	}
}

// TestGetDiffSince 增量查询返回差异
func TestGetDiffSince_增量查询返回差异(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "版本测试", Env: "dev", Version: 3, LangCode: "zh-CN", IsPublished: true,
	})

	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "old_key", StringValue: "旧值",
		Version: 1, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "new_key", StringValue: "新值",
		Version: 3, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "mod_key", StringValue: "修改后",
		Version: 3, OperationType: domain.OperationMod,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "del_key", StringValue: "已删除",
		Version: 2, OperationType: domain.OperationDel,
	})

	diff, err := repo.GetDiffSince(packID, 2)
	if err != nil {
		t.Fatalf("GetDiffSince 失败: %v", err)
	}
	if len(diff) != 2 {
		t.Fatalf("版本 > 2 的 ADD/MOD 记录应有 2 条, 实际 %d", len(diff))
	}
}

// TestGetDiffSince 无差异返回空
func TestGetDiffSince_无差异返回空(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "无差异", Env: "dev", Version: 5, LangCode: "en-US", IsPublished: true,
	})

	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "stable", StringValue: "稳定值",
		Version: 1, OperationType: domain.OperationAdd,
	})

	diff, err := repo.GetDiffSince(packID, 5)
	if err != nil {
		t.Fatalf("GetDiffSince 失败: %v", err)
	}
	if len(diff) != 0 {
		t.Errorf("无增量变更时应返回空, 实际 %d 条", len(diff))
	}
}

// TestGetDiffSince 只返回 ADD 和 MOD
func TestGetDiffSince_只返回ADD和MOD(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "操作类型", Env: "dev", Version: 2, LangCode: "ja-JP", IsPublished: true,
	})

	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "add_item", StringValue: "新增",
		Version: 2, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "mod_item", StringValue: "修改",
		Version: 2, OperationType: domain.OperationMod,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "del_item", StringValue: "删除",
		Version: 2, OperationType: domain.OperationDel,
	})

	diff, _ := repo.GetDiffSince(packID, 1)
	for _, s := range diff {
		if s.OperationType == domain.OperationDel {
			t.Errorf("DEL 类型不应出现在增量结果中, key=%s", s.StringKey)
		}
	}
}

func strPtr(s string) *string {
	return &s
}
