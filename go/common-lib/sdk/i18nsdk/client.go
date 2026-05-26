package i18nsdk

import (
	"context"
)

// Client 国际化 SDK 客户端接口
// 提供多语言模板渲染、参数化消息、降级回退能力
// 所有用户可见文案必须通过此接口渲染，禁止硬编码中英文文案
type Client interface {
	// T 渲染多语言文本（Template 渲染模式）
	// lang: 语言代码，如 zh-CN、en
	// key: 文案键名，如 svc_hello_greeting
	// params: 模板参数，如 {"name": "张三", "server_name": "CaiRobot"}
	// 返回: 渲染后的文本和错误（失败时返回 fallback 文案，不抛异常）
	T(ctx context.Context, lang string, key string, params map[string]any) (string, error)

	// Raw 获取原始模板文本（由客户端按 template_type 自行渲染）
	// 用于客户端可渲染的场景
	Raw(ctx context.Context, lang string, key string) (string, string, error)

	// Ping 健康检查，验证国际化服务可用性
	Ping(ctx context.Context) error
}

// TemplateType 模板类型枚举
type TemplateType int

const (
	// TemplateTypeNamed 命名参数模板，如 "你好，{name}！欢迎使用 {server_name}。"
	TemplateTypeNamed TemplateType = iota + 1
	// TemplateTypeICU ICU MessageFormat 模板，支持 plural/select/gender 等
	TemplateTypeICU
)
