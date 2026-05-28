package sdk

import (
	"context"
	"encoding/json"
	"log"
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
// 三级降级策略：
// 1. 完整 JSON（含 tenant_id）→ 结构化处理
// 2. 部分 JSON（可解析但无 tenant_id，有 module_keys）→ 提取 module_keys 后走结构化
// 3. 纯逗号分隔 / 无法识别格式 → 旧版兼容处理
func (p *pubsubManager) onMessage(msg string) {
	var evt InvalidateEvent
	if err := json.Unmarshal([]byte(msg), &evt); err == nil {
		if evt.TenantID != "" {
			p.handleStructured(evt)
			return
		}
		if len(evt.ModuleKeys) > 0 {
			log.Printf("[WARN] config-sdk: received JSON without tenant_id, extracting module_keys. msg=%s", truncateMsg(msg, 80))
			p.handleStructured(evt)
			return
		}
	}
	log.Printf("[WARN] config-sdk: received legacy pub/sub format, migrating to JSON. msg=%s", truncateMsg(msg, 80))
	moduleKeys := strings.Split(msg, ",")
	p.handleLegacy(moduleKeys)
}

// handleStructured 处理结构化 InvalidateEvent 消息
func (p *pubsubManager) handleStructured(evt InvalidateEvent) {
	for _, key := range evt.ModuleKeys {
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

// handleLegacy 处理旧版逗号分隔格式的消息
func (p *pubsubManager) handleLegacy(moduleKeys []string) {
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

func truncateMsg(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Ping 检查配置服务是否可用
// InProcess 模式下调用 GetVersionInfo 验证连接
func (c *configClient) Ping(ctx context.Context) error {
	return c.pingService(ctx)
}
