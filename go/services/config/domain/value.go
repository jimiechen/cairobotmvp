package domain

import (
	"encoding/json"
	"fmt"
)

// TypedValue 类型安全的配置值包装器
// 将 config_json 中解析出的原始 any 值按 FieldType 做类型守卫
// 防止下游直接使用 interface{} 导致的类型 panic
type TypedValue struct {
	Type  FieldType
	Value any
}

// NewTypedValue 创建带类型标记的 TypedValue
// 调用方应确保 value 类型与 fieldType 声明一致
func NewTypedValue(fieldType FieldType, value any) *TypedValue {
	return &TypedValue{Type: fieldType, Value: value}
}

// String 以 string 类型返回值
// 当实际类型不是 string 时返回格式化字符串而非 panic，保证降级可用
func (tv *TypedValue) String() string {
	if tv.Value == nil {
		return ""
	}
	if s, ok := tv.Value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", tv.Value)
}

// Int 以 int64 类型返回值
// 用于 FieldTypeInt / FieldTypeEnum 场景
func (tv *TypedValue) Int() int64 {
	if tv.Value == nil {
		return 0
	}
	switch v := tv.Value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	}
	return 0
}

// Bool 以 bool 类型返回值
func (tv *TypedValue) Bool() bool {
	if tv.Value == nil {
		return false
	}
	if b, ok := tv.Value.(bool); ok {
		return b
	}
	return false
}

// Float 以 float64 类型返回值
// 用于 FieldTypeFloat 场景
func (tv *TypedValue) Float() float64 {
	if tv.Value == nil {
		return 0
	}
	switch v := tv.Value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	}
	return 0
}

// JSON 以 JSON 原始字节返回值
// 用于 FieldTypeJSON / FieldTypeList 场景，保留完整结构
func (tv *TypedValue) JSON() json.RawMessage {
	if tv.Value == nil {
		return nil
	}
	switch v := tv.Value.(type) {
	case json.RawMessage:
		return v
	case []byte:
		return v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return bytes
	}
}
