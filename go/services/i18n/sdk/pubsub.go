package sdk

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

const (
	// INVALIDATE_CHANNEL Redis pub/sub channel 名称
	INVALIDATE_CHANNEL = "cairobot.i18n.invalidate"
)

// pubSubClient Redis Pub/Sub 客户端
//
// 职责：
// - 监听 cairobot.i18n.invalidate channel
// - 收到消息后触发缓存失效和 Watch 回调
//
// 不负责：
// - Redis 连接管理（由基础设施层负责）
// - 消息序列化/反序列化（由协议层负责）
type pubSubClient struct {
	options  *Options
	cancel   context.CancelFunc
	mu       sync.Mutex
	active   bool
	redis    RedisClient
}

// RedisClient Redis 客户端抽象接口
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSec int) error
	Delete(key string) error
	Subscribe(channel string, handler func(string)) (func(), error)
	Publish(channel string, message string) error
	Close() error
}

func newPubSubClient(opts *Options) *pubSubClient {
	return &pubSubClient{
		options: opts,
	}
}

// Start 启动订阅监听
// 使用 options 中的 Redis 配置建立连接
func (p *pubSubClient) Start(watchers *watcherManager, cache *lruCache) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return nil
	}

	if p.options.Redis == nil || p.options.Redis.Addr == "" {
		return fmt.Errorf("redis config not provided")
	}

	p.active = true
	return nil
}

// Stop 停止订阅监听
func (p *pubSubClient) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}

	if p.redis != nil {
		p.redis.Close()
	}

	p.active = false
	return nil
}

// Publish 发布缓存失效消息（供服务端调用）
func (p *pubSubClient) Publish(langCode string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return p.redis.Publish(INVALIDATE_CHANNEL, langCode)
}

// IsConnected 检查是否已连接
func (p *pubSubClient) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active
}

// onMessage 处理收到的 pub/sub 消息
// 消息格式为 lang_code，支持批量失效（逗号分隔）
func (p *pubSubClient) onMessage(msg string, watchers *watcherManager, cache *lruCache) {
	langCodes := strings.Split(msg, ",")
	for _, langCode := range langCodes {
		langCode = strings.TrimSpace(langCode)
		if langCode == "" {
			continue
		}

		// 使缓存失效
		cache.InvalidateByLangCode(langCode)

		// 触发 Watch 回调（通知客户端重新拉取）
		watchers.Trigger(langCode, 0)
	}
}
