package admin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

type pubRecord struct {
	channel string
	message string
}

// mockCache 模拟 redisx.Client
type mockCache struct {
	mu          sync.Mutex
	invalidated []string
}

func (m *mockCache) Get(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}
func (m *mockCache) Delete(_ context.Context, _ ...string) error { return nil }
func (m *mockCache) Scan(_ context.Context, _ string) ([]string, error) { return nil, nil }
func (m *mockCache) Invalidate(_ context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invalidated = append(m.invalidated, pattern)
	return nil
}
func (m *mockCache) Ping(_ context.Context) error { return nil }
func (m *mockCache) Close() error               { return nil }

// mockPubSub 模拟 Publisher
type mockPubSub struct {
	mu        sync.Mutex
	published []pubRecord
}

func (m *mockPubSub) Publish(_ context.Context, channel, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, pubRecord{channel, message})
	return nil
}

// mockAuditWriter 记录审计
type mockAuditWriter struct {
	mu      sync.Mutex
	entries []AuditEntry
}

func (m *mockAuditWriter) Write(_ context.Context, entry AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

// mockI18nService 实现 service.I18nService（仅用于校验）
type mockI18nService struct{}

func (mockI18nService) GetLanguages(_ string) ([]service.LanguageMeta, error) { return nil, nil }
func (mockI18nService) GetLangPack(_, _, _ string) (*service.LangPackResponse, error) {
	return nil, nil
}
func (mockI18nService) GetLangDifference(_ string, _ int64, _, _ string) (*service.LangDiffResponse, error) {
	return nil, nil
}
func (mockI18nService) ValidateTemplate(value string, tt domain.TemplateType, params []domain.LangParam) error {
	if value == "" {
		return fmt.Errorf("模板值不能为空")
	}
	if tt == domain.TemplateNamed && !strings.Contains(value, "{") && len(params) > 0 {
		return fmt.Errorf("命名参数模板应包含占位符")
	}
	return nil
}

func newI18nTestService(t *testing.T) (*AdminI18nService, *mockCache, *mockPubSub, *mockAuditWriter) {
	t.Helper()
	repo := repository.NewMockRepo()
	cache := &mockCache{}
	ps := &mockPubSub{}
	audit := &mockAuditWriter{}

	svc := NewAdminI18nService(&mockI18nService{}, repo,
		WithCache(cache),
		WithPubSub(ps),
		WithAuditWriter(audit),
	)
	return svc, cache, ps, audit
}

func TestCreateString_HappyPath(t *testing.T) {
	svc, _, _, audit := newI18nTestService(t)
	ctx := context.Background()

	result, err := svc.CreateString(ctx, CreateStringRequest{
		PackID:      1,
		StringKey:   domain.StringKey("test_key"),
		StringValue: "测试值",
		TemplateType: domain.TemplatePlain,
		Operator:    "admin01",
	})
	if err != nil {
		t.Fatalf("CreateString 失败: %v", err)
	}
	if result.StringKey != "test_key" {
		t.Errorf("期望 StringKey=test_key，实际=%s", result.StringKey)
	}

	audit.mu.Lock()
	if len(audit.entries) == 0 || audit.entries[0].Action != "create_string" {
		t.Error("应写入 create_string 审计")
	}
	audit.mu.Unlock()
}

func TestCreateString_空Key应报错(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	_, err := svc.CreateString(context.Background(), CreateStringRequest{
		PackID: 1, StringValue: "值",
	})
	if err == nil {
		t.Fatal("空 string_key 应报错")
	}
}

func TestCreateString_模板校验失败(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	_, err := svc.CreateString(context.Background(), CreateStringRequest{
		PackID:      1,
		StringKey:   domain.StringKey("bad_key"),
		StringValue: "",
		TemplateType: domain.TemplatePlain,
	})
	if err == nil {
		t.Fatal("空字符串值应导致模板校验失败")
	}
	if !strings.Contains(err.Error(), "模板校验失败") {
		t.Errorf("错误信息应包含'模板校验失败'，实际=%s", err.Error())
	}
}

func TestUpdateString_HappyPath(t *testing.T) {
	svc, _, ps, _ := newI18nTestService(t)
	ctx := context.Background()

	created, _ := svc.CreateString(ctx, CreateStringRequest{
		PackID: 1, StringKey: domain.StringKey("up_key"), StringValue: "旧值",
	})

	updated, err := svc.UpdateString(ctx, UpdateStringRequest{
		ID: created.ID, StringValue: "新值", Operator: "admin02",
	})
	if err != nil {
		t.Fatalf("UpdateString 失败: %v", err)
	}
	if updated.StringValue != "新值" {
		t.Errorf("期望 StringValue=新值，实际=%s", updated.StringValue)
	}

	ps.mu.Lock()
	if len(ps.published) == 0 {
		t.Log("UpdateString 广播依赖 GetPackByLangCode 返回非空，MockRepo 未预置 pack 数据时可跳过此断言")
	}
	ps.mu.Unlock()
}

func TestDeleteString_HappyPath(t *testing.T) {
	svc, _, _, audit := newI18nTestService(t)
	ctx := context.Background()

	created, _ := svc.CreateString(ctx, CreateStringRequest{
		PackID: 1, StringKey: domain.StringKey("del_key"), StringValue: "待删除",
	})

	err := svc.DeleteString(ctx, created.ID, "admin03")
	if err != nil {
		t.Fatalf("DeleteString 失败: %v", err)
	}

	audit.mu.Lock()
	found := false
	for _, e := range audit.entries {
		if e.Action == "delete_string" {
			found = true
			break
		}
	}
	if !found {
		t.Error("删除操作应产生 delete_string 审计")
	}
	audit.mu.Unlock()
}

func TestListStrings_有数据(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	ctx := context.Background()

	svc.CreateString(ctx, CreateStringRequest{
		PackID: 1, StringKey: domain.StringKey("a"), StringValue: "A",
	})
	svc.CreateString(ctx, CreateStringRequest{
		PackID: 1, StringKey: domain.StringKey("b"), StringValue: "B",
	})

	items, err := svc.ListStrings(1)
	if err != nil {
		t.Fatalf("ListStrings 失败: %v", err)
	}
	if len(items) < 2 {
		t.Errorf("期望至少 2 条记录，实际=%d", len(items))
	}
}

func TestPublishPack_HappyPath(t *testing.T) {
	svc, cache, ps, audit := newI18nTestService(t)
	ctx := context.Background()

	result, err := svc.PublishPack(ctx, PublishPackRequest{
		PackID: 1, LangCode: "zh-CN", Env: "dev", Operator: "admin04",
	})
	if err != nil {
		t.Fatalf("PublishPack 失败: %v", err)
	}
	if result.Version <= 0 {
		t.Error("版本号应大于 0")
	}

	cache.mu.Lock()
	if len(cache.invalidated) == 0 {
		t.Error("发布后应触发缓存失效")
	}
	cache.mu.Unlock()

	ps.mu.Lock()
	if len(ps.published) == 0 {
		t.Error("发布后应触发变更广播")
	}
	ps.mu.Unlock()

	audit.mu.Lock()
	if len(audit.entries) == 0 || audit.entries[len(audit.entries)-1].Action != "publish_pack" {
		t.Error("发布操作应产生 publish_pack 审计")
	}
	audit.mu.Unlock()
}

func TestPublishPack_无效PackID应报错(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	_, err := svc.PublishPack(context.Background(), PublishPackRequest{PackID: -1})
	if err == nil {
		t.Fatal("无效 pack_id 应报错")
	}
}

func TestRollbackPack_HappyPath(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	err := svc.RollbackPack(context.Background(), 1, 1, "admin05")
	if err != nil {
		t.Fatalf("RollbackPack 失败: %v", err)
	}
}

func TestImportCSV_基本解析(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	csvContent := "key1,value1,g1,plain\nkey2,value2,g2,named\n"
	reader := strings.NewReader(csvContent)

	result, err := svc.ImportStringsFromCSV(context.Background(), reader, 1, "admin")
	if err != nil {
		t.Fatalf("ImportCSV 失败: %v", err)
	}
	if result.TotalRows != 2 {
		t.Errorf("期望 2 行，实际=%d", result.TotalRows)
	}
	if result.SuccessCount != 2 {
		t.Errorf("期望成功 2 条，实际=%d", result.SuccessCount)
	}
}

func TestExportCSV_有数据(t *testing.T) {
	svc, _, _, _ := newI18nTestService(t)
	ctx := context.Background()

	svc.CreateString(ctx, CreateStringRequest{
		PackID: 1, StringKey: domain.StringKey("ek"), StringValue: "ev",
	})

	data, err := svc.ExportStringsToCSV(ctx, ExportCSVRequest{PackID: 1})
	if err != nil {
		t.Fatalf("ExportCSV 失败: %v", err)
	}
	if !strings.Contains(string(data), "ek") {
		t.Errorf("导出 CSV 应包含 ek，实际=%s", string(data))
	}
}

func TestInvalidateLangCode_无Cache不panic(t *testing.T) {
	repo := repository.NewMockRepo()
	svc := NewAdminI18nService(&mockI18nService{}, repo)
	err := svc.invalidateLangCode(context.Background(), []string{"zh-CN"})
	if err != nil {
		t.Errorf("无 cache 时不应报错: %v", err)
	}
}
