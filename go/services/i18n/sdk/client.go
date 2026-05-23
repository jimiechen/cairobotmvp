package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// Mode 表示 SDK 运行模式
type Mode string

const (
	// ModeInProcess 进程内模式：直接调用 service 层
	ModeInProcess Mode = "in_process"
	// ModeRemote 远程模式：通过 TarsGo 调用远程服务（MVP 占位）
	ModeRemote Mode = "remote"
)

// Client 国际化 SDK 客户端接口
// 提供翻译、模板查询、批量操作和变更订阅能力
//
// 职责：
// - 封装语言包查找和渲染逻辑
// - 提供多级缓存（LRU → Service → Remote）
// - 支持变更订阅通知
//
// 不负责：
// - 语言包存储和管理（由 service 层负责）
// - HTTP 路由处理（由调用方负责）
type Client interface {
	// T 翻译指定 key 的文本并渲染参数
	T(ctx context.Context, langCode, key string, params map[string]any) (string, error)

	// Raw 获取原始模板信息（不渲染）
	Raw(ctx context.Context, langCode, key string) (*Template, error)

	// BatchT 批量翻译多个 key
	BatchT(ctx context.Context, langCode string, keys []string, params map[string]any) (map[string]string, error)

	// Watch 订阅语言包版本变更
	Watch(langCode string, handler func(packVersion int64)) (cancel func())

	// Ping 检查服务可用性
	Ping(ctx context.Context) error
}

// Template 模板结构体
// 包含原始模板值、类型和参数描述
type Template struct {
	Key          string
	Value        string
	TemplateType string // plain/named/icu
	Params       []ParamInfo
}

// ParamInfo 参数描述信息
type ParamInfo struct {
	Name     string
	Type     string
	Required bool
}

// Options SDK 配置选项
type Options struct {
	// Mode 运行模式：进程内或远程
	Mode Mode
	// Service I18nService 实例（InProcess 模式必填）
	Service service.I18nService
	// Env 环境标识（dev/staging/prod）
	Env string
	// DefaultLangCode 默认语言代码
	DefaultLangCode string
	// TarsServant 远程服务地址（Remote 模式使用）
	TarsServant string
	// Redis Redis 连接配置（pub/sub 使用）
	Redis *RedisConfig
	// RemoteClient 远程客户端实例（Remote 模式使用）
	RemoteClient RemoteClient
	// Timeout 远程调用超时
	Timeout time.Duration
	// RetryCount 重试次数
	RetryCount int
	// RetryInterval 重试间隔
	RetryInterval time.Duration
	// ClientVersion 客户端版本号
	ClientVersion string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DefaultOptions 返回默认配置
func DefaultOptions() *Options {
	return &Options{
		Mode:           ModeInProcess,
		Env:            "dev",
		DefaultLangCode: "zh-CN",
	}
}

// clientImpl Client 接口的默认实现
type clientImpl struct {
	options    *Options
	cache      *lruCache
	watchers   *watcherManager
	remote     *remoteClient
	pubsub     *pubSubClient
	initOnce   sync.Once
	initErr    error
}

// Default 创建默认的 SDK 客户端实例
// 使用 InProcess 模式，必须提供 I18nService 实例
func Default(opts ...func(*Options)) (Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if options.Mode == ModeInProcess && options.Service == nil {
		return nil, fmt.Errorf("in_process mode requires Service")
	}

	c := &clientImpl{
		options:  options,
		cache:    newLRUCache(128),
		watchers: newWatcherManager(),
	}

	if options.Mode == ModeRemote {
		c.remote = newRemoteClient(options)
		c.pubsub = newPubSubClient(options)
	}

	return c, nil
}
