package redisx

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// CancelFunc 取消订阅的函数类型
type CancelFunc func()

// MessageHandler pub/sub 消息处理函数类型
type MessageHandler func(channel string, message string)

// PubSubClient 发布订阅客户端封装
// 用于 SDK 缓存失效通知和 admin-server 变更广播
type PubSubClient struct {
	client    *redis.Client
	handlers  map[string]MessageHandler
	mu        sync.RWMutex
	cancelCtx context.CancelFunc
	active    bool
}

// NewPubSubClient 创建 Pub/Sub 客户端
func NewPubSubClient(cfg *Config) (*PubSubClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redisx: pubsub config is nil")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     2,
		MinIdleConns: 1,
		DialTimeout:  cfg.DialTimeout,
	})

	ctx, cancel := context.WithCancel(context.Background())

	psc := &PubSubClient{
		client:    rdb,
		handlers:  make(map[string]MessageHandler),
		cancelCtx: cancel,
		active:    true,
	}

	go psc.listenLoop(ctx)

	return psc, nil
}

// Subscribe 订阅指定 channel 的消息
func (p *PubSubClient) Subscribe(channel string, handler MessageHandler) (CancelFunc, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return nil, fmt.Errorf("redisx: pubsub client is closed")
	}

	p.handlers[channel] = handler

	return func() {
		p.mu.Lock()
		delete(p.handlers, channel)
		p.mu.Unlock()
	}, nil
}

// Publish 向指定 channel 发布消息
func (p *PubSubClient) Publish(ctx context.Context, channel string, message string) error {
	return p.client.Publish(ctx, channel, message).Err()
}

// listenLoop 监听循环，处理所有已订阅的 channel 消息
func (p *PubSubClient) listenLoop(ctx context.Context) {
	sub := p.client.Subscribe(ctx)
	channels := []string{}

	p.mu.RLock()
	for ch := range p.handlers {
		channels = append(channels, ch)
	}
	p.mu.RUnlock()

	if len(channels) > 0 {
		err := sub.Subscribe(ctx, channels...)
		if err != nil {
			return
		}
	}

	ch := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			sub.Close()
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			
			p.mu.RLock()
			handler, exists := p.handlers[msg.Channel]
			p.mu.RUnlock()
			
			if exists {
				handler(msg.Channel, msg.Payload)
			}
		}
	}
}

// Close 关闭 Pub/Sub 客户端
func (p *PubSubClient) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.active = false
	p.cancelCtx()
	p.handlers = make(map[string]MessageHandler)
	return p.client.Close()
}
