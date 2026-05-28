package admin

import (
	"context"
	"time"
)

// AuditWriter 审计日志写入接口
// 解耦具体存储实现，便于测试时替换
type AuditWriter interface {
	Write(ctx context.Context, entry AuditEntry) error
}

// AuditEntry 审计日志条目
type AuditEntry struct {
	TenantID    string
	Action      string // "create_schema" | "update_schema" | "delete_schema" | "publish_value"
	TargetType  string // "schema" | "value"
	TargetID    string
	Operator    string
	Detail      string
	RequestIP   string
	OccurredAt  time.Time
}

// NoopAuditWriter 空操作审计写入器（测试/开发用）
type NoopAuditWriter struct{}

func (NoopAuditWriter) Write(_ context.Context, _ AuditEntry) error { return nil }
