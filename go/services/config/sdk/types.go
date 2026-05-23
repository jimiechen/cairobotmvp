package sdk

import (
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

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
