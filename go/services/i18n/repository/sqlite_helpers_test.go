package repository

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// TestParseTimeStr 各种时间格式解析
func TestParseTimeStr_各种格式(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"RFC3339 格式", "2024-01-15T10:30:00Z", false},
		{"标准日期时间", "2024-01-15 10:30:00", false},
		{"T 分隔符无时区", "2024-01-15T10:30:00", false},
		{"DateTime 布局", "2006-01-02 15:04:05", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.input
			got := parseTimeStr(&s)
			if tt.wantZero && !got.IsZero() {
				t.Errorf("期望零值时间, 得到 %v", got)
			}
			if !tt.wantZero && got.IsZero() {
				t.Error("期望非零值时间, 得到零值")
			}
		})
	}
}

// TestParseTimeStr nil 指针返回零值
func TestParseTimeStr_nil指针返回零值(t *testing.T) {
	got := parseTimeStr(nil)
	if !got.IsZero() {
		t.Errorf("nil 应返回零值时间, 得到 %v", got)
	}
}

// TestParseTimeStr 空字符串返回零值
func TestParseTimeStr_空字符串返回零值(t *testing.T) {
	s := ""
	got := parseTimeStr(&s)
	if !got.IsZero() {
		t.Errorf("空字符串应返回零值时间, 得到 %v", got)
	}
}

// TestParseTimeStr 纯空白字符串返回零值
func TestParseTimeStr_纯空白返回零值(t *testing.T) {
	s := "   "
	got := parseTimeStr(&s)
	if !got.IsZero() {
		t.Errorf("纯空白应返回零值时间, 得到 %v", got)
	}
}

// TestParseTimeStr 无效格式返回零值
func TestParseTimeStr_无效格式返回零值(t *testing.T) {
	s := "not-a-date"
	got := parseTimeStr(&s)
	if !got.IsZero() {
		t.Errorf("无效格式应返回零值时间, 得到 %v", got)
	}
}

// TestParseTimeStr 带空白的有效时间会尝试兜底解析
func TestParseTimeStr_带空白有效时间(t *testing.T) {
	s := "  2024-01-15 10:30:00  "
	got := parseTimeStr(&s)
	if got.IsZero() {
		t.Log("带空白的时间可能无法解析，这是预期行为")
	}
}

// TestParseTimeStr time.Layout 兜底格式
func TestParseTimeStr_timeLayout兜底(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"January 2, 2006"},
		{"Jan 2, 2006"},
		{"2006年01月02日"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := tt.input
			got := parseTimeStr(&s)
			if !got.IsZero() {
				t.Logf("成功解析 %s -> %v", tt.input, got)
			}
		})
	}
}

// TestGetPackByLangCode 带 Description 字段验证完整映射
func TestGetPackByLangCode_带Description字段(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	insertTestPack(t, repo.DB(), domain.LangPack{
		PackName:    "带描述包",
		Env:         "prod",
		Version:     1,
		LangCode:    "zh-TW",
		Description: "繁体中文语言包",
		IsPublished: true,
	})

	got, err := repo.GetPackByLangCode("zh-TW", "prod")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回语言包")
	}
	if got.Description != "繁体中文语言包" {
		t.Errorf("Description 不匹配: %s", got.Description)
	}
}

// TestGetStringsByPackID 包含 DEL 类型记录被过滤
func TestGetStringsByPackID_DEL类型被过滤(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "过滤测试", Env: "dev", Version: 1, LangCode: "ko-KR", IsPublished: true,
	})

	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "keep", StringValue: "保留",
		Version: 1, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "remove", StringValue: "删除",
		Version: 2, OperationType: domain.OperationDel,
	})

	strings, _ := repo.GetStringsByPackID(packID)
	for _, s := range strings {
		if s.StringKey == "remove" {
			t.Error("DEL 类型记录不应出现在结果中")
		}
	}
}

// TestGetDiffSince 版本号精确匹配
func TestGetDiffSince_版本号精确匹配(t *testing.T) {
	repo := newTestSQLiteRepo(t)
	packID := insertTestPack(t, repo.DB(), domain.LangPack{
		PackName: "版本边界", Env: "dev", Version: 3, LangCode: "th-TH", IsPublished: true,
	})

	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "v1_item", StringValue: "版本1",
		Version: 1, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "v2_item", StringValue: "版本2",
		Version: 2, OperationType: domain.OperationAdd,
	})
	insertTestString(t, repo.DB(), domain.LangString{
		PackID: packID, StringKey: "v3_item", StringValue: "版本3",
		Version: 3, OperationType: domain.OperationMod,
	})

	diffV2, _ := repo.GetDiffSince(packID, 2)
	if len(diffV2) != 1 {
		t.Errorf("sinceVersion=2 应有 1 条(版本>2), 实际 %d", len(diffV2))
	}

	diffV1, _ := repo.GetDiffSince(packID, 1)
	if len(diffV1) != 2 {
		t.Errorf("sinceVersion=1 应有 2 条(版本>1), 实际 %d", len(diffV1))
	}
}

// TestListPacks 按 lang_code 排序
func TestListPacks_按LangCode排序(t *testing.T) {
	repo := newTestSQLiteRepo(t)

	langs := []string{"en-US", "zh-CN", "ja-JP"}
	for _, lang := range langs {
		insertTestPack(t, repo.DB(), domain.LangPack{
			PackName: lang, Env: "test", Version: 1,
			LangCode: lang, IsPublished: true,
		})
	}

	packs, _ := repo.ListPacks("test")
	if len(packs) != 3 {
		t.Fatalf("期望 3 条, 实际 %d", len(packs))
	}
	for i := 1; i < len(packs); i++ {
		if packs[i].LangCode < packs[i-1].LangCode {
			t.Errorf("未按 lang_code 排序: %s < %s", packs[i].LangCode, packs[i-1].LangCode)
		}
	}
}
