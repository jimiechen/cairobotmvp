package health

import (
	"context"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
	"github.com/jimiechen/mineplanet/go/third_party/mysqlx"
	"github.com/jimiechen/mineplanet/go/third_party/redisx"
)

const checkerTimeout = 1 * time.Second

// ConfigChecker 配置服务健康检查器
// 通过 configsdk.Ping 验证配置服务可用性
type ConfigChecker struct {
	client configsdk.Client
}

// NewConfigChecker 创建配置检查器
func NewConfigChecker(client configsdk.Client) *ConfigChecker {
	return &ConfigChecker{client: client}
}

func (c *ConfigChecker) Name() string { return "config" }

func (c *ConfigChecker) Check(ctx context.Context) health.ComponentStatus {
	start := time.Now()

	err := c.client.Ping(ctx)
	if err != nil {
		return health.NewUnhealthyComponentStatus("config", time.Since(start).Milliseconds(), err)
	}

	return health.NewComponentStatus("config", time.Since(start).Milliseconds())
}

// I18nChecker 国际化服务健康检查器
// 通过 i18nsdk.Ping 验证国际化服务可用性
type I18nChecker struct {
	client i18nsdk.Client
}

// NewI18nChecker 创建国际化检查器
func NewI18nChecker(client i18nsdk.Client) *I18nChecker {
	return &I18nChecker{client: client}
}

func (i *I18nChecker) Name() string { return "i18n" }

func (i *I18nChecker) Check(ctx context.Context) health.ComponentStatus {
	start := time.Now()

	if i.client == nil {
		return health.NewComponentStatus("i18n", time.Since(start).Milliseconds())
	}

	err := i.client.Ping(ctx)
	if err != nil {
		return health.NewUnhealthyComponentStatus("i18n", time.Since(start).Milliseconds(), err)
	}

	return health.NewComponentStatus("i18n", time.Since(start).Milliseconds())
}

// MySQLChecker MySQL 数据库健康检查器
// 通过 mysqlx.DB.Ping 验证数据库连接可用性
type MySQLChecker struct {
	db mysqlx.DB
}

// NewMySQLChecker 创建 MySQL 检查器
// db: MySQL 数据库连接（nil 时返回不健康状态）
func NewMySQLChecker(db mysqlx.DB) *MySQLChecker {
	return &MySQLChecker{db: db}
}

func (m *MySQLChecker) Name() string { return "mysql" }

func (m *MySQLChecker) Check(ctx context.Context) health.ComponentStatus {
	start := time.Now()

	if m.db == nil {
		return health.NewUnhealthyComponentStatus("mysql", 0, ErrDBNil)
	}

	ctx, cancel := context.WithTimeout(ctx, checkerTimeout)
	defer cancel()

	err := m.db.Ping(ctx)
	if err != nil {
		return health.NewUnhealthyComponentStatus("mysql", time.Since(start).Milliseconds(), err)
	}

	return health.NewComponentStatus("mysql", time.Since(start).Milliseconds())
}

// RedisChecker Redis 缓存健康检查器
// 通过 redisx.Client.Ping 验证缓存连接可用性
type RedisChecker struct {
	client redisx.Client
}

// NewRedisChecker 创建 Redis 检查器
// client: Redis 客户端（nil 时返回不健康状态）
func NewRedisChecker(client redisx.Client) *RedisChecker {
	return &RedisChecker{client: client}
}

func (r *RedisChecker) Name() string { return "redis" }

func (r *RedisChecker) Check(ctx context.Context) health.ComponentStatus {
	start := time.Now()

	if r.client == nil {
		return health.NewUnhealthyComponentStatus("redis", 0, ErrClientNil)
	}

	ctx, cancel := context.WithTimeout(ctx, checkerTimeout)
	defer cancel()

	err := r.client.Ping(ctx)
	if err != nil {
		return health.NewUnhealthyComponentStatus("redis", time.Since(start).Milliseconds(), err)
	}

	return health.NewComponentStatus("redis", time.Since(start).Milliseconds())
}

var (
	ErrDBNil     = fmt.Errorf("mysql db is nil")
	ErrClientNil = fmt.Errorf("redis client is nil")
)

var _ = fmt.Sprintf
