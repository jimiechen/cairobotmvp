package domain

// FieldType 字段类型枚举，与 DDL sys_config_schema.field_type 一致
// 新增类型需同步更新 TypedValue 的类型转换方法
type FieldType string

const (
 FieldTypeString FieldType = "string"
 FieldTypeInt    FieldType = "int"
 FieldTypeBool   FieldType = "bool"
 FieldTypeFloat  FieldType = "float"
 FieldTypeEnum   FieldType = "enum"
 FieldTypeJSON   FieldType = "json"
 FieldTypeList   FieldType = "list"
)

// FieldSchema 单个字段的元数据描述，对应 sys_config_schema 表一行
// 负责描述一个配置字段的结构约束（类型、必填、脱敏、校验规则）
// 不负责值的存储和校验执行，校验由 service/validate.go 完成
type FieldSchema struct {
	ID           int64
	ModuleKey    string
	FieldKey     string
	FieldType    FieldType
	DefaultValue string
	Validator    string
	IsRequired   bool
	IsSecret     bool
	Description  string
	ClientScope  string
	MinAppVer    string
	SortOrder    int
	IsEnabled    bool
}

// MatchClientScope 判断当前字段是否对指定客户端范围可见
// clientScope 为 "all" 时匹配所有客户端；否则精确匹配
func (f *FieldSchema) MatchClientScope(clientScope string) bool {
	if f.ClientScope == "all" || f.ClientScope == clientScope {
		return true
	}
	return false
}

// ModuleSchema 一个模块下所有字段 Schema 的聚合视图
// 用于 service/compose.go 组装 DynamicConfigModule.descriptors
type ModuleSchema struct {
	ModuleKey string
	Fields    []*FieldSchema
}

// FindField 按 field_key 查找字段 Schema
// 找不到返回 nil，调用方需自行判空
func (m *ModuleSchema) FindField(fieldKey string) *FieldSchema {
	for _, f := range m.Fields {
		if f.FieldKey == fieldKey {
			return f
		}
	}
	return nil
}

// EnabledFields 返回所有启用状态的字段列表
// 过滤掉 is_enabled=false 的字段（运营下线场景）
func (m *ModuleSchema) EnabledFields() []*FieldSchema {
	var result []*FieldSchema
	for _, f := range m.Fields {
		if f.IsEnabled {
			result = append(result, f)
		}
	}
	return result
}
