package admin

import (
	"context"
	"encoding/json"
	"time"
)

// i18nInvalidateEvent 构造 i18n InvalidateEvent
func i18nInvalidateEvent(langCodes []string) map[string]interface{} {
	return map[string]interface{}{
		"tenant_id":  "default",
		"scope":      "i18n",
		"env":        "dev",
		"lang_codes": langCodes,
		"version":    time.Now().UnixMilli(),
		"timestamp":  time.Now().Unix(),
	}
}

// marshalEvent 序列化事件为 JSON 字符串
func marshalEvent(evt interface{}) (string, error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Publisher 变更广播接口（与 config/admin 共享同一接口定义语义）
type Publisher interface {
	Publish(ctx context.Context, channel, message string) error
}

// AuditEntry 审计日志条目（与 config/admin 共享结构）
type AuditEntry struct {
	TenantID   string
	Action     string
	TargetType string
	TargetID   string
	Operator   string
	Detail     string
	RequestIP  string
	OccurredAt time.Time
}

// AuditWriter 审计写入器接口
type AuditWriter interface {
	Write(ctx context.Context, entry AuditEntry) error
}

// NoopAuditWriter 空操作审计写入器
type NoopAuditWriter struct{}

func (NoopAuditWriter) Write(_ context.Context, _ AuditEntry) error { return nil }
