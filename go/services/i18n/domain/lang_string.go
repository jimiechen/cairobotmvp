package domain

import "time"

// LangString 语言字符串实体
// 对应 sys_lang_string 表，表示一个语言包中的字符串条目
//
// 职责：
// - 封装字符串的基本属性和模板信息
// - 提供模板类型判断方法
// - 提供 ParamsSchema 解析方法
//
// 不负责：
// - 数据库操作（由 Repository 负责）
// - 模板校验（由 TemplateValidator 负责）
type LangString struct {
	ID            int64
	PackID        int64
	StringKey     StringKey
	StringValue   string
	GroupName     string
	Version       int
	OperationType OperationType
	PrevValue     *string
	TemplateType  TemplateType
	ParamsSchema  string
	PreviewSample string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// StringKey 字符串键类型
// 使用自定义类型防止与普通字符串混淆
type StringKey string

// IsPlain 判断是否为纯文本模板
func (s *LangString) IsPlain() bool {
	return s.TemplateType == TemplatePlain
}

// IsNamed 判断是否为命名参数模板
func (s *LangString) IsNamed() bool {
	return s.TemplateType == TemplateNamed
}

// IsIcu 判断是否为 ICU 模板
func (s *LangString) IsIcu() bool {
	return s.TemplateType == TemplateIcu
}

// GetParams 解析并返回参数列表
func (s *LangString) GetParams() ([]LangParam, error) {
	return ParseParamsSchema(s.ParamsSchema)
}
