package event

import (
	"context"
	"fmt"
)

// TODO(event, MVP-P1): 实现 Redis Pub/Sub 发布器
// 当前为占位实现，MVP-P0 使用 MemoryBus 或 NoopPublisher
// 实现时应基于 go-redis/redis v9 的 PubSub 功能

// RedisPublisherConfig Redis 发布器配置
type RedisPublisherConfig struct {
	Addr     string // Redis 地址，如 "localhost:6379"
	Password string // Redis 密码
	DB       int    // Redis 数据库编号
}

// RedisPublisher 基于 Redis Pub/Sub 的事件发布器（占位）
// 生产环境默认使用此实现，支持跨实例事件传播
type RedisPublisher struct {
	cfg RedisPublisherConfig
}

// NewRedisPublisher 创建 Redis 发布器实例
func NewRedisPublisher(cfg RedisPublisherConfig) *RedisPublisher {
	return &RedisPublisher{cfg: cfg}
}

// Publish 发布事件到 Redis 频道
// 格式：social:event:{eventType} 作为 channel，DomainEvent JSON 作为 message
func (p *RedisPublisher) Publish(_ context.Context, evt DomainEvent) error {
	// TODO: 接入 redis.Client.Publish(ctx, "social:event:"+evt.Type, jsonBytes)
	// MVP-P0 暂时返回 nil，不阻塞业务
	return fmt.Errorf("RedisPublisher 未实现: event type=%s", evt.Type)
}

// RedisSubscriber 基于 Redis Pub/Sub 的事件订阅器（占位）
type RedisSubscriber struct {
	cfg RedisPublisherConfig
}

// NewRedisSubscriber 创建 Redis 订阅器实例
func NewRedisSubscriber(cfg RedisPublisherConfig) *RedisSubscriber {
	return &RedisSubscriber{cfg: cfg}
}

// Subscribe 订阅指定类型事件的 Redis 频道
func (s *RedisSubscriber) Subscribe(_ context.Context, eventType string, handler Handler) error {
	// TODO: 接入 redis.Client.Subscribe + goroutine 分发到 handler
	return fmt.Errorf("RedisSubscriber 未实现: event type=%s", eventType)
}
