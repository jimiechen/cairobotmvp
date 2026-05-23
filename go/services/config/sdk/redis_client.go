package sdk

import (
	"fmt"
)

// goRedisClient 占位实现
// 实际的 go-redis 实现应在 main 包中提供
type goRedisClient struct {
	addr     string
	password string
	db       int
}

// NewGoRedisClient 创建 Redis 客户端实例
// 注意：此为占位实现，实际实现应在基础设施层
func NewGoRedisClient(addr, password string, db int) (RedisClient, error) {
	return &goRedisClient{
		addr:     addr,
		password: password,
		db:       db,
	}, nil
}

func (c *goRedisClient) Get(key string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *goRedisClient) Set(key string, value string, ttlSec int) error {
	return fmt.Errorf("not implemented")
}

func (c *goRedisClient) Delete(key string) error {
	return fmt.Errorf("not implemented")
}

func (c *goRedisClient) Subscribe(channel string, handler MessageHandler) (CancelFunc, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *goRedisClient) Publish(channel string, message string) error {
	return fmt.Errorf("not implemented")
}

func (c *goRedisClient) Close() error {
	return nil
}

// LoadRedisConfig 从统一配置加载 Redis 客户端
func LoadRedisConfig(addr, password string, db int) (RedisClient, error) {
	return NewGoRedisClient(addr, password, db)
}
