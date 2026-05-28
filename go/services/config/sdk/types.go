package sdk

import (
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// InvalidateEvent pub/sub 失效广播事件结构（JSON payload）
// 用于 admin 写入后通知 SDK 刷新缓存
// 向后兼容：SDK 消费端同时支持 JSON 格式和旧逗号分隔格式
type InvalidateEvent struct {
	TenantID   string   `json:"tenant_id"`
	Scope      string   `json:"scope"`                 // "config" | "i18n"
	Env        string   `json:"env"`
	ModuleKeys []string `json:"module_keys,omitempty"` // config 用
	LangCodes  []string `json:"lang_codes,omitempty"`   // i18n 用
	Version    int64    `json:"version"`
	Timestamp  int64    `json:"timestamp"`
	TraceID    string   `json:"trace_id,omitempty"`
}

// ModuleSnapshot 模块配置快照
// 包含某个 module_key 在特定时间点的所有字段值
// 用于 GetModule 返回和 Watch 回调参数
type ModuleSnapshot struct {
	ModuleKey string
	Version   int64
	Fields    map[string]*domain.TypedValue
}

// GetField 从快照中获取指定字段的 TypedValue
// 找不到返回 nil，调用方需自行判空
func (s *ModuleSnapshot) GetField(fieldKey string) *domain.TypedValue {
	if s.Fields == nil {
		return nil
	}
	return s.Fields[fieldKey]
}
