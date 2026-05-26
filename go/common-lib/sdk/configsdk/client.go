package configsdk

import (
	"context"
)

// Client 配置 SDK 客户端接口
// 提供强类型配置读取、热更新订阅、健康检查能力
// 所有业务模块必须通过此接口读取配置，禁止直接访问配置存储
type Client interface {
	// GetString 读取字符串类型配置值
	// moduleKey: 模块配置键名，如 hello_cfg、device_cfg
	// fieldKey: 字段键名，如 server_name、max_name_length
	// 返回: 配置值和错误（读不到时降级到 default，仅记录 warn 日志）
	GetString(ctx context.Context, moduleKey string, fieldKey string) (string, error)

	// GetInt 读取整数类型配置值
	GetInt(ctx context.Context, moduleKey string, fieldKey string) (int64, error)

	// GetBool 读取布尔类型配置值
	GetBool(ctx context.Context, moduleKey string, fieldKey string) (bool, error)

	// Watch 订阅配置变更通知
	// callback: 配置变更时的回调函数
	Watch(ctx context.Context, moduleKey string, callback func(fieldKey string, oldValue, newValue interface{})) error

	// Ping 健康检查，验证配置服务可用性
	Ping(ctx context.Context) error
}
