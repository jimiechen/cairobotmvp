package sdk

import (
	"context"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/service"
)

// Client 配置 SDK 客户端接口
// 对外暴露类型安全的配置读取能力，屏蔽 InProcess/Remote 模式差异
// 不负责配置的持久化和版本管理，由 ConfigService 处理
type Client interface {
	GetString(ctx context.Context, moduleKey, fieldKey string) (string, error)
	GetInt(ctx context.Context, moduleKey, fieldKey string) (int64, error)
	GetBool(ctx context.Context, moduleKey, fieldKey string) (bool, error)
	GetFloat(ctx context.Context, moduleKey, fieldKey string) (float64, error)
	GetJSON(ctx context.Context, moduleKey, fieldKey string, out any) error
	GetModule(ctx context.Context, moduleKey string) (*ModuleSnapshot, error)
	Bind(ctx context.Context, moduleKey string, out any) error
	Watch(moduleKey string, handler func(*ModuleSnapshot)) (cancel func())
	Ping(ctx context.Context) error
}

// Mode SDK 运行模式枚举
// 决定配置数据来源：进程内服务调用 或 远程 TarsGo 调用
type Mode string

const (
	ModeInProcess Mode = "in_process" // 进程内模式，直接调用 ConfigService
	ModeRemote    Mode = "remote"     // 远程模式，通过 TarsGo 调用远程 ConfigServer
)

// Options SDK 客户端构造选项
// 聚合所有可配置项，支持函数式选项模式扩展
type Options struct {
	Mode          Mode
	Service       service.ConfigService
	Env           string
	ClientScope   string
	TarsServant   string
	Redis         RedisClient
	RemoteClient  RemoteClient
	Timeout       time.Duration
	RetryCount    int
	RetryInterval time.Duration
	CacheSize     int
	CacheTTLSec   int
}

// Option 函数式选项接口
type Option func(*Options)

// WithMode 设置运行模式
func WithMode(mode Mode) Option {
	return func(o *Options) { o.Mode = mode }
}

//WithService 设置进程内 ConfigService 实例
func WithService(svc service.ConfigService) Option {
	return func(o *Options) { o.Service = svc }
}

// WithEnv 设置环境标识
func WithEnv(env string) Option {
	return func(o *Options) { o.Env = env }
}

// WithClientScope 设置客户端范围
func WithClientScope(scope string) Option {
	return func(o *Options) { o.ClientScope = scope }
}

// WithTarsServant 设置远程 Tars 服务地址
func WithTarsServant(servant string) Option {
	return func(o *Options) { o.TarsServant = servant }
}

// WithRedis 设置 Redis 客户端（用于 L2 缓存 + pub/sub）
func WithRedis(redis RedisClient) Option {
	return func(o *Options) { o.Redis = redis }
}

// WithCacheSize 设置 LRU 缓存容量
func WithCacheSize(size int) Option {
	return func(o *Options) { o.CacheSize = size }
}

// WithCacheTTL 设置 LRU 缓存 TTL（秒）
func WithCacheTTL(ttl int) Option {
	return func(o *Options) { o.CacheTTLSec = ttl }
}

// defaultOptions 返回默认配置选项
// MVP 阶段默认使用 InProcess 模式，缓存容量 256，TTL 30 秒
func defaultOptions() Options {
	return Options{
		Mode:        ModeInProcess,
		Env:         "dev",
		ClientScope: "all",
		CacheSize:   256,
		CacheTTLSec: 30,
	}
}

// MessageHandler pub/sub 消息处理函数类型
type MessageHandler func(msg string)

// CancelFunc 取消函数类型
type CancelFunc func()

// RedisClient Redis 客户端抽象接口
// 解耦 SDK 与具体 Redis 实现（go-redis / mock / fake）
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSec int) error
	Delete(key string) error
	Subscribe(channel string, handler MessageHandler) (CancelFunc, error)
}

// configClient Client 接口的默认实现
// 组合 L1 LRU 缓存 + L2 Redis + L3 远程服务的三级缓存架构
type configClient struct {
	options  Options
	lruCache *lruCache
	watcher  *moduleWatcher
	pubsub   *pubsubManager
}

// Default 创建 SDK 客户端实例（工厂方法）
// 使用默认选项 + 可选的 Option 覆盖
// 前置条件：InProcess 模式下 Service 不能为 nil
func Default(opts ...Option) (Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if options.Mode == ModeInProcess && options.Service == nil {
		return nil, ErrServiceRequired
	}
	client := &configClient{
		options:  options,
		lruCache: newLRUCache(options.CacheSize),
		watcher:  newModuleWatcher(),
	}
	if options.Redis != nil {
		client.pubsub = newPubsubManager(options.Redis, client.lruCache, client.watcher)
	}
	return client, nil
}
