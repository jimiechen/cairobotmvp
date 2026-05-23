package service

import (
	"encoding/json"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

// ParseConfigJSON 将 config_json 文本按 schema 解析为类型安全的 map
// 输入：原始 JSON 字符串 + module_key + schema 仓库引用
// 输出：map[fieldKey]*TypedValue，每个值带类型标记
// 无 schema 的字段按 JSON 原始类型推断（降级策略）
func ParseConfigJSON(configJSON, moduleKey string, schemaRepo repository.SchemaRepository) (map[string]*domain.TypedValue, error) {
	rawMap := make(map[string]any)
	if err := json.Unmarshal([]byte(configJSON), &rawMap); err != nil {
		return nil, err
	}

	schemas, _ := schemaRepo.ListByModule(moduleKey)
	schemaIndex := make(map[string]*domain.FieldSchema)
	for _, s := range schemas {
		schemaIndex[s.FieldKey] = s
	}

	result := make(map[string]*domain.TypedValue)
	for fieldKey, rawValue := range rawMap {
		fs, hasSchema := schemaIndex[fieldKey]
		if hasSchema {
			result[fieldKey] = newTypedValueFromSchema(fs, rawValue)
		} else {
			result[fieldKey] = inferTypedValue(rawValue)
		}
	}
	return result, nil
}

// newTypedValueFromSchema 按 schema 声明的 fieldType 创建 TypedValue
// 当实际值与声明类型不兼容时，使用 defaultValue 作为兜底
func newTypedValueFromSchema(fs *domain.FieldSchema, rawValue any) *domain.TypedValue {
	switch fs.FieldType {
	case domain.FieldTypeString:
		return domain.NewTypedValue(domain.FieldTypeString, coerceString(rawValue))
	case domain.FieldTypeInt:
		return domain.NewTypedValue(domain.FieldTypeInt, coerceInt64(rawValue))
	case domain.FieldTypeBool:
		return domain.NewTypedValue(domain.FieldTypeBool, coerceBool(rawValue))
	case domain.FieldTypeFloat:
		return domain.NewTypedValue(domain.FieldTypeFloat, coerceFloat64(rawValue))
	case domain.FieldTypeEnum:
		return domain.NewTypedValue(domain.FieldTypeEnum, coerceString(rawValue))
	case domain.FieldTypeJSON:
		return domain.NewTypedValue(domain.FieldTypeJSON, coerceRawJSON(rawValue))
	case domain.FieldTypeList:
		return domain.NewTypedValue(domain.FieldTypeList, rawValue)
	default:
		return domain.NewTypedValue(domain.FieldTypeString, coerceString(rawValue))
	}
}

// inferTypedValue 无 schema 时的类型推断
// 根据运行时值的实际 Go 类型做最佳猜测
func inferTypedValue(rawValue any) *domain.TypedValue {
	switch v := rawValue.(type) {
	case string:
		return domain.NewTypedValue(domain.FieldTypeString, v)
	case float64:
		if v == float64(int64(v)) {
			return domain.NewTypedValue(domain.FieldTypeInt, int64(v))
		}
		return domain.NewTypedValue(domain.FieldTypeFloat, v)
	case bool:
		return domain.NewTypedValue(domain.FieldTypeBool, v)
	case nil:
		return domain.NewTypedValue(domain.FieldTypeString, "")
	default:
		return domain.NewTypedValue(domain.FieldTypeJSON, v)
	}
}

func coerceString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func coerceInt64(v any) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case json.Number:
		i, _ := val.Int64()
		return i
	}
	return 0
}

func coerceBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func coerceFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	}
	return 0
}

func coerceRawJSON(v any) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return bytes
}
