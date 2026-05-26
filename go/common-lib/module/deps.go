package module

import (
	"context"
)

// ConfigReader 配置读取接口（内联定义，避免跨包导入）
// 对应 configsdk.Client 的完整方法集
type ConfigReader interface {
	GetString(ctx context.Context, moduleKey string, fieldKey string) (string, error)
	GetInt(ctx context.Context, moduleKey string, fieldKey string) (int64, error)
	GetBool(ctx context.Context, moduleKey string, fieldKey string) (bool, error)
	Watch(ctx context.Context, moduleKey string, callback func(fieldKey string, oldValue, newValue interface{})) error
	Ping(ctx context.Context) error
}

// I18nRenderer 国际化渲染接口（内联定义，避免跨包导入）
// 对应 i18nsdk.Client 的核心方法子集
type I18nRenderer interface {
	T(ctx context.Context, lang string, key string, params map[string]any) (string, error)
	Raw(ctx context.Context, lang string, key string) (string, string, error)
	Ping(ctx context.Context) error
}

// Logger 日志接口（内联定义，避免跨包导入）
// 对应 TarsCloud contrib/log.Logger 的核心方法子集
type Logger interface {
	Info(ctx context.Context, v ...interface{})
	Infof(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, v ...interface{})
	Errorf(ctx context.Context, format string, v ...interface{})
	Warn(ctx context.Context, v ...interface{})
	Debug(ctx context.Context, v ...interface{})
}

// Deps 模块统一依赖装配结构
// 所有业务模块必须通过此结构接收依赖，禁止直接 import services/* 内部包
// 禁止直接 sql.Open / redis.NewClient
//
// 各字段类型的实际对应：
//   - Config: configsdk.Client（实现 ConfigReader 接口）
//   - I18n: i18nsdk.Client（实现 I18nRenderer 接口）
//   - DB: mysqlx.DB（可选，无持久化为 nil）
//   - Cache: redisx.Client（可选，无缓存为 nil）
//   - Logger: *log.Logger（实现 Logger 接口）
type Deps struct {
	Config ConfigReader  // 必选：配置 SDK
	I18n   I18nRenderer // 可选：国际化 SDK（无文案场景为 nil）
	DB     interface{}   // 可选：数据库连接（预留 mysqlx.DB 类型）
	Cache  interface{}   // 可选：缓存客户端（预留 redisx.Client 类型）
	Logger Logger       // 必选：日志组件
}
