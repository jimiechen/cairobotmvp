package sdk

import (
	"context"
	"strings"
	"sync"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

const (
	pubsubChannel = "cairobot.config.invalidate"
)

// pubsubManager Redis pub/sub 订阅管理器
// 监听 cairobot.config.invalidate channel，收到消息后：
// 1. 失效对应 module 的 LRU 缓存
// 2. 触发 Watch 回调（重新拉取最新配置后）
type pubsubManager struct {
	redis   RedisClient
	cache   *lruCache
	watcher *moduleWatcher
	cancel  func()
	mu      sync.Mutex
	active  bool
}

// newPubsubManager 创建 pub/sub 管理器并启动订阅
func newPubsubManager(redis RedisClient, cache *lruCache, watcher *moduleWatcher) *pubsubManager {
	p := &pubsubManager{
		redis:   redis,
		cache:   cache,
		watcher: watcher,
	}
	p.start()
	return p
}

// start 启动 Redis pub/sub 订阅
// 订阅 cairobot.config.invalidate channel
func (p *pubsubManager) start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active {
		return
	}
	cancel, err := p.redis.Subscribe(pubsubChannel, p.onMessage)
	if err != nil {
		return
	}
	p.cancel = cancel
	p.active = true
}

// stop 停止订阅
func (p *pubsubManager) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.active {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	p.active = false
}

// onMessage 处理收到的 pub/sub 消息
// 消息格式为 module_key，支持批量失效（逗号分隔）
func (p *pubsubManager) onMessage(msg string) {
	moduleKeys := strings.Split(msg, ",")
	for _, key := range moduleKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		cacheKey := buildCacheKey(key)
		p.cache.delete(cacheKey)
		p.watcher.notify(key, &ModuleSnapshot{
			ModuleKey: key,
			Fields:    make(map[string]*domain.TypedValue),
		})
	}
}

// Ping 检查配置服务是否可用
// InProcess 模式下调用 GetVersionInfo 验证连接
func (c *configClient) Ping(ctx context.Context) error {
	return c.pingService(ctx)
}
