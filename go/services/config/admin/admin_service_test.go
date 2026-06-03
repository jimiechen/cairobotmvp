package admin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	"github.com/jimiechen/mineplanet/go/services/config/testutil"
)

func newTestService(t *testing.T) (*AdminConfigService, *testutil.MockCache, *testutil.MockPubSub, *testutil.MockAuditWriter) {
	t.Helper()
	schemaRepo := repository.NewMemSchemaRepo()
	configRepo := repository.NewMemConfigRepo()
	innerSchema := service.NewSchemaService(schemaRepo)
	cache := testutil.NewMockCache()
	ps := testutil.NewMockPubSub()
	auditWriter := testutil.NewMockAuditWriter()

	svc := NewAdminConfigService(innerSchema, schemaRepo, configRepo,
		WithCache(cache),
		WithPubSub(ps),
		WithAuditWriter(&auditWrapper{writer: auditWriter}),
	)
	return svc, cache, ps, auditWriter
}

type auditWrapper struct {
	writer *testutil.MockAuditWriter
}

func (w *auditWrapper) Write(ctx context.Context, entry AuditEntry) error {
	return w.writer.Write(ctx, entry)
}

func TestCreateSchema_HappyPath(t *testing.T) {
	svc, cache, ps, audit := newTestService(t)
	ctx := context.Background()

	result, err := svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey:    "test_mod",
		FieldKey:     "timeout_sec",
		FieldType:    domain.FieldTypeInt,
		DefaultValue: "30",
		Validator:    "range:1,300",
		IsRequired:   true,
		Operator:     "admin01",
	})
	if err != nil {
		t.Fatalf("CreateSchema 失败: %v", err)
	}
	if result == nil {
		t.Fatal("返回结果不应为 nil")
	}
	if result.FieldKey != "timeout_sec" {
		t.Errorf("期望 field_key=timeout_sec，实际=%s", result.FieldKey)
	}
	if !result.IsEnabled {
		t.Error("新创建的 schema 应默认启用")
	}

	if len(cache.GetInvalidated()) == 0 {
		t.Error("应触发缓存失效")
	}

	if len(ps.GetPublished()) == 0 {
		t.Error("应触发变更广播")
	} else {
		msg := ps.GetPublished()[0].Message
		if !strings.Contains(msg, `"tenant_id"`) {
			t.Errorf("广播消息应为 JSON 格式且含 tenant_id，实际=%s", msg)
		}
		if !strings.Contains(msg, "test_mod") {
			t.Errorf("广播消息应包含 module_key=test_mod，实际=%s", msg)
		}
	}

	if len(audit.GetEntries()) == 0 {
		t.Error("应写入审计日志")
	} else if entry, ok := audit.GetEntries()[0].(AuditEntry); ok && entry.Action != "create_schema" {
		t.Errorf("审计 action 应为 create_schema，实际=%s", entry.Action)
	}
}

func TestCreateSchema_空ModuleKey应报错(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, err := svc.CreateSchema(context.Background(), CreateSchemaRequest{
		FieldKey:  "some_field",
		FieldType: domain.FieldTypeString,
	})
	if err == nil {
		t.Fatal("空 module_key 应报错")
	}
	if !strings.Contains(err.Error(), "module_key") {
		t.Errorf("错误信息应包含 module_key 提示，实际=%s", err.Error())
	}
}

func TestUpdateSchema_HappyPath(t *testing.T) {
	svc, _, ps, _ := newTestService(t)
	ctx := context.Background()

	created, _ := svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "up_mod", FieldKey: "old_name",
		FieldType: domain.FieldTypeString,
	})

	updated, err := svc.UpdateSchema(ctx, UpdateSchemaRequest{
		ID:           created.ID,
		FieldType:    domain.FieldTypeInt,
		DefaultValue: "100",
		Description:  "更新后的描述",
		Operator:     "admin02",
	})
	if err != nil {
		t.Fatalf("UpdateSchema 失败: %v", err)
	}
	if updated.FieldType != "int" {
		t.Errorf("期望 FieldType=int，实际=%s", updated.FieldType)
	}
	if updated.Description != "更新后的描述" {
		t.Error("描述未正确更新")
	}

	found := false
	for _, p := range ps.GetPublished() {
		if strings.Contains(p.Message, "up_mod") {
			found = true
			break
		}
	}
	if !found {
		t.Error("更新后应广播 up_mod 的失效")
	}
}

func TestUpdateSchema_无效ID应报错(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, err := svc.UpdateSchema(context.Background(), UpdateSchemaRequest{ID: -1})
	if err == nil {
		t.Fatal("无效 ID 应报错")
	}
}

func TestDeleteSchema_HappyPath(t *testing.T) {
	svc, _, ps, audit := newTestService(t)
	ctx := context.Background()

	created, _ := svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "del_mod", FieldKey: "to_delete",
		FieldType: domain.FieldTypeString,
	})

	err := svc.DeleteSchema(ctx, created.ID, "admin03")
	if err != nil {
		t.Fatalf("DeleteSchema 失败: %v", err)
	}

	listed, _ := svc.ListSchemas(ctx, "del_mod")
	for _, item := range listed {
		if item.ID == created.ID && item.IsEnabled {
			t.Error("删除后 schema 应被禁用")
		}
	}

	if len(ps.GetPublished()) == 0 {
		t.Error("删除后应广播失效")
	}

	if len(audit.GetEntries()) == 0 {
		t.Fatal("审计条目不应为空")
	}
	lastEntry := audit.GetLastEntry()
	if lastEntry == nil {
		t.Fatal("最后一条审计条目不应为nil")
	}
	entry, ok := lastEntry.(AuditEntry)
	if !ok {
		t.Fatalf("审计条目类型错误，期望 AuditEntry")
	}
	if entry.Action != "delete_schema" {
		t.Error("删除操作应产生 delete_schema 审计")
	}
}

func TestListSchemas_有数据(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	ctx := context.Background()

	svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "list_mod", FieldKey: "field_a", FieldType: domain.FieldTypeString,
	})
	svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "list_mod", FieldKey: "field_b", FieldType: domain.FieldTypeInt,
	})

	items, err := svc.ListSchemas(ctx, "list_mod")
	if err != nil {
		t.Fatalf("ListSchemas 失败: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("期望 2 条记录，实际=%d", len(items))
	}
}

func TestListSchemas_空模块返回空列表(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	items, err := svc.ListSchemas(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("ListSchemas 失败: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("不存在的模块应返回空列表，实际=%d", len(items))
	}
}

func TestPublishValue_HappyPath(t *testing.T) {
	svc, cache, ps, audit := newTestService(t)
	ctx := context.Background()

	svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "pub_mod", FieldKey: "max_conn",
		FieldType: domain.FieldTypeInt,
	})

	fields := map[string]*domain.TypedValue{
		"max_conn": domain.NewTypedValue(domain.FieldTypeInt, 100),
	}
	result, err := svc.PublishValue(ctx, PublishValueRequest{
		ModuleKey: "pub_mod",
		Env:       "dev",
		Fields:    fields,
		Operator:  "admin04",
	})
	if err != nil {
		t.Fatalf("PublishValue 失败: %v", err)
	}
	if result.ModuleKey != "pub_mod" {
		t.Errorf("期望 ModuleKey=pub_mod，实际=%s", result.ModuleKey)
	}
	if result.FieldCount != 1 {
		t.Errorf("期望 FieldCount=1，实际=%d", result.FieldCount)
	}

	if len(cache.GetInvalidated()) == 0 {
		t.Error("发布配置值后应触发缓存失效")
	}

	if len(ps.GetPublished()) == 0 {
		t.Error("发布配置值后应触发变更广播")
	}

	if len(audit.GetEntries()) == 0 {
		t.Fatal("审计条目不应为空")
	}
	lastEntry := audit.GetLastEntry()
	if lastEntry == nil {
		t.Fatal("最后一条审计条目不应为nil")
	}
	entry, ok := lastEntry.(AuditEntry)
	if !ok {
		t.Fatalf("审计条目类型错误，期望 AuditEntry")
	}
	if entry.Action != "publish_value" {
		t.Error("发布操作应产生 publish_value 审计")
	}
}

func TestPublishValue_校验失败应返回字段级错误(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	ctx := context.Background()

	svc.CreateSchema(ctx, CreateSchemaRequest{
		ModuleKey: "val_mod", FieldKey: "port",
		FieldType: domain.FieldTypeInt, Validator: "range:1,65535",
	})

	fields := map[string]*domain.TypedValue{
		"port": domain.NewTypedValue(domain.FieldTypeInt, 99999),
	}
	_, err := svc.PublishValue(ctx, PublishValueRequest{
		ModuleKey: "val_mod",
		Env:       "dev",
		Fields:    fields,
	})
	if err == nil {
		t.Fatal("超出范围的值应导致校验失败")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("错误类型应为 ValidationError，实际=%T", err)
	}
	if len(valErr.Errors) == 0 {
		t.Error("ValidationError 应包含至少一个字段错误")
	}
}

func TestPublishValue_空Fields应报错(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	_, err := svc.PublishValue(context.Background(), PublishValueRequest{
		ModuleKey: "empty_mod",
		Env:       "dev",
	})
	if err == nil {
		t.Fatal("空 fields 应报错")
	}
}

func TestInvalidateAndBroadcast_无Cache不panic(t *testing.T) {
	schemaRepo := repository.NewMemSchemaRepo()
	configRepo := repository.NewMemConfigRepo()
	innerSchema := service.NewSchemaService(schemaRepo)

	svc := NewAdminConfigService(innerSchema, schemaRepo, configRepo)
	err := svc.invalidateAndBroadcast(context.Background(), []string{"test"})
	if err != nil {
		t.Errorf("无 cache 时不应报错: %v", err)
	}
}

func TestNewAdminConfigService_选项注入(t *testing.T) {
	schemaRepo := repository.NewMemSchemaRepo()
	configRepo := repository.NewMemConfigRepo()
	innerSchema := service.NewSchemaService(schemaRepo)

	svc := NewAdminConfigService(innerSchema, schemaRepo, configRepo,
		WithCache(testutil.NewMockCache()),
		WithPubSub(testutil.NewMockPubSub()),
		WithAuditWriter(&auditWrapper{writer: testutil.NewMockAuditWriter()}),
	)
	if svc.cache == nil {
		t.Error("WithCache 未生效")
	}
	if svc.pubsub == nil {
		t.Error("WithPubSub 未生效")
	}
	if svc.auditWriter == nil {
		t.Error("WithAuditWriter 未生效")
	}
}
