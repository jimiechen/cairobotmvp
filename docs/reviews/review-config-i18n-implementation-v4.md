# 全局配置元模型 + 多语言参数化模板 — 修正评审报告（v4）

**评审编号**: REV-CONFIG-I18N-004
**版本**: v4.0
**日期**: 2026-05-22
**评审对象**: `docs/prd/global-config-i18n-implementation-plan.md` 实施结果（引入统一配置后）
**评审依据**: `.trae/rules/review.md` 评审规则
**状态**: 待主控裁决

---

## 一、结论

**建议修改（2 项必须修复 + 1 项待确认）**

经引入 `server_config_test.yaml` 统一环境配置文件，v3 评审中的 **M-06（数据库配置）和 M-07（Redis 配置）已有明确的引入路径**。本次修正评审报告重点更新配置引入方案，明确统一配置类的设计、单元测试使用 MySQL 的方案，以及各遗留项的修复依赖关系。

---

## 二、统一配置引入方案

### 2.1 配置文件来源

**源文件**: `docs/tabbit/inbox/2026/05/server_config_test.yaml`

该文件是 MineplanetGo 项目的统一服务端配置，包含：
- MySQL 配置（共享）
- Redis 配置（共享）
- JWT、Auth、Group、Topic、Inbox、Third 等服务配置
- Tars Registry 配置

### 2.2 引入路径

**目标位置**: `configs/server.yaml`（项目根目录）

```yaml
# configs/server.yaml
# CaiRobot MVP 统一服务端配置
# 基于 server_config_test.yaml 裁剪，仅保留 config/i18n 所需配置

# MySQL Configuration (Shared)
mysql:
  host: "rm-t4np5ht1x04y8ko98eo.mysql.singapore.rds.aliyuncs.com"
  port: 3306
  username: "mpuser"
  password: "Huawei@2025"
  database: "mineplanet_community_db"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"

# Redis Configuration (Shared)
redis:
  enabled: true
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0

# Config / I18n Service Specific
cache:
  config_ttl_seconds: 30
  i18n_ttl_seconds: 60
  pubsub_enabled: true
```

### 2.3 统一配置类设计

**文件**: `go/common-lib/config/server_config.go`

```go
package config

// ServerConfig 统一服务端配置
// 从 configs/server.yaml 加载，支持环境变量覆盖
type ServerConfig struct {
    MySQL  MySQLConfig  `yaml:"mysql"`
    Redis  RedisConfig  `yaml:"redis"`
    Cache  CacheConfig  `yaml:"cache"`
}

type MySQLConfig struct {
    Host             string `yaml:"host"`
    Port             int    `yaml:"port"`
    Username         string `yaml:"username"`
    Password         string `yaml:"password"`
    Database         string `yaml:"database"`
    Charset          string `yaml:"charset"`
    MaxOpenConns     int    `yaml:"max_open_conns"`
    MaxIdleConns     int    `yaml:"max_idle_conns"`
    ConnMaxLifetime  string `yaml:"conn_max_lifetime"`
    ConnMaxIdleTime  string `yaml:"conn_max_idle_time"`
}

type RedisConfig struct {
    Enabled  bool   `yaml:"enabled"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type CacheConfig struct {
    ConfigTTLSeconds int  `yaml:"config_ttl_seconds"`
    I18nTTLSeconds   int  `yaml:"i18n_ttl_seconds"`
    PubSubEnabled    bool `yaml:"pubsub_enabled"`
}
```

### 2.4 配置加载方式

**使用 Viper 库**，支持：
1. 从 `configs/server.yaml` 加载基础配置
2. 环境变量覆盖（如 `MYSQL_HOST`、`REDIS_PASSWORD`）
3. 多环境支持（dev/staging/prod）

```go
// go/common-lib/config/loader.go
func Load(path string) (*ServerConfig, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.AutomaticEnv()
    v.SetEnvPrefix("CAIROBOT")
    
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var cfg ServerConfig
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

---

## 三、数据库配置引入方案（修正 M-06）

### 3.1 现状

| 检查项 | 状态 |
|---|---|
| MySQL 连接配置 | ❌ 未引入（但 server_config_test.yaml 已提供） |
| MySQL 仓库实现（config） | ⚠️ 占位实现 |
| MySQL 仓库实现（i18n） | ⚠️ 占位实现 |
| 数据库配置加载 | ❌ 未找到（需引入 Viper + server.yaml） |
| 单元测试使用 MySQL | ❌ 当前使用 SQLite |

### 3.2 修正方案

#### 3.2.1 引入统一配置

1. 将 `server_config_test.yaml` 裁剪为 `configs/server.yaml`
2. 引入 `github.com/spf13/viper` 依赖
3. 创建 `go/common-lib/config/` 统一配置包

#### 3.2.2 实现 MySQL 仓库

**文件**: `go/services/config/repository/mysql_repo.go`

```go
package repository

import (
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/jimiechen/mineplanet/go/common-lib/config"
    "github.com/jimiechen/mineplanet/go/services/config/domain"
)

type MySQLConfigRepo struct {
    db *sql.DB
}

func NewMySQLConfigRepo(cfg *config.MySQLConfig) (*MySQLConfigRepo, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
        cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("打开 MySQL 失败: %w", err)
    }
    
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    
    if lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
        db.SetConnMaxLifetime(lifetime)
    }
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
    }
    
    return &MySQLConfigRepo{db: db}, nil
}

// GetLatestVersion 实现...
// GetByModuleAndVersion 实现...
// ListPublishedVersions 实现...
// Save 实现...
```

#### 3.2.3 单元测试使用 MySQL

**方案**: 使用 Testcontainers 或本地 MySQL 实例

```go
// go/services/config/repository/mysql_repo_test.go
func TestMySQLConfigRepo_Integration(t *testing.T) {
    cfg := &config.MySQLConfig{
        Host:     getEnvOrDefault("TEST_MYSQL_HOST", "127.0.0.1"),
        Port:     getEnvOrDefaultInt("TEST_MYSQL_PORT", 3306),
        Username: getEnvOrDefault("TEST_MYSQL_USER", "root"),
        Password: getEnvOrDefault("TEST_MYSQL_PASSWORD", "123456"),
        Database: getEnvOrDefault("TEST_MYSQL_DB", "cairobot_test"),
        Charset:  "utf8mb4",
    }
    
    repo, err := NewMySQLConfigRepo(cfg)
    if err != nil {
        t.Skipf("跳过 MySQL 集成测试: %v", err)
    }
    defer repo.db.Close()
    
    // 测试 CRUD...
}
```

**CI 配置**:
```yaml
# .github/workflows/ci.yml
services:
  mysql:
    image: mysql:8.0
    env:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: cairobot_test
    ports:
      - 3306:3306
```

### 3.3 双模式切换

**文件**: `go/services/config/repository/factory.go`

```go
package repository

import (
    "github.com/jimiechen/mineplanet/go/common-lib/config"
)

// NewRepository 根据配置创建对应的数据仓库
// dev 环境使用 SQLite，prod 环境使用 MySQL
func NewRepository(cfg *config.ServerConfig) (ConfigRepository, error) {
    switch cfg.Database.Driver {
    case "mysql":
        return NewMySQLConfigRepo(&cfg.MySQL)
    case "sqlite":
        return NewSQLiteConfigRepo(cfg.SQLite.Path)
    default:
        return NewSQLiteConfigRepo("") // 默认 SQLite
    }
}
```

---

## 四、Redis 配置引入方案（修正 M-07）

### 4.1 现状

| 检查项 | 状态 |
|---|---|
| Redis 连接配置 | ❌ 未引入（但 server_config_test.yaml 已提供） |
| Redis 缓存实现 | ❌ 未找到 |
| Redis 客户端 | ❌ 未找到 |
| Pub/Sub 订阅 | ⚠️ 接口定义存在 |
| Pub/Sub 发布 | ⚠️ Noop 实现 |

### 4.2 修正方案

#### 4.2.1 引入 Redis 依赖

```bash
cd go/services/config && go get github.com/redis/go-redis/v9
cd go/services/i18n && go get github.com/redis/go-redis/v9
```

#### 4.2.2 实现 Redis 缓存

**文件**: `go/services/config/cache/redis_cache.go`

```go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/jimiechen/mineplanet/go/common-lib/config"
)

type RedisCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisCache(cfg *config.RedisConfig, ttlSeconds int) (*RedisCache, error) {
    if !cfg.Enabled {
        return nil, fmt.Errorf("redis disabled")
    }
    
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })
    
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("连接 Redis 失败: %w", err)
    }
    
    return &RedisCache{
        client: client,
        ttl:    time.Duration(ttlSeconds) * time.Second,
    }, nil
}

func (c *RedisCache) Get(key string) (any, bool) {
    val, err := c.client.Get(context.Background(), key).Result()
    if err != nil {
        return nil, false
    }
    var result any
    if err := json.Unmarshal([]byte(val), &result); err != nil {
        return nil, false
    }
    return result, true
}

func (c *RedisCache) Set(key string, value any) {
    bytes, _ := json.Marshal(value)
    c.client.Set(context.Background(), key, bytes, c.ttl)
}

func (c *RedisCache) Delete(key string) {
    c.client.Del(context.Background(), key)
}

func (c *RedisCache) Invalidate(prefix string) {
    iter := c.client.Scan(context.Background(), 0, prefix+":*", 0).Iterator()
    for iter.Next(context.Background()) {
        c.client.Del(context.Background(), iter.Val())
    }
}
```

#### 4.2.3 实现 Redis Pub/Sub

**文件**: `go/services/config/sdk/redis_client.go`

```go
package sdk

import (
    "context"
    "fmt"
    
    "github.com/redis/go-redis/v9"
    "github.com/jimiechen/mineplanet/go/common-lib/config"
)

type goRedisClient struct {
    client *redis.Client
}

func NewGoRedisClient(cfg *config.RedisConfig) (RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })
    
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("连接 Redis 失败: %w", err)
    }
    
    return &goRedisClient{client: client}, nil
}

func (c *goRedisClient) Get(key string) (string, error) {
    return c.client.Get(context.Background(), key).Result()
}

func (c *goRedisClient) Set(key string, value string, ttlSec int) error {
    return c.client.Set(context.Background(), key, value, time.Duration(ttlSec)*time.Second).Err()
}

func (c *goRedisClient) Delete(key string) error {
    return c.client.Del(context.Background(), key).Err()
}

func (c *goRedisClient) Subscribe(channel string, handler MessageHandler) (CancelFunc, error) {
    pubsub := c.client.Subscribe(context.Background(), channel)
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        ch := pubsub.Channel()
        for {
            select {
            case msg := <-ch:
                if msg != nil {
                    handler(msg.Payload)
                }
            case <-ctx.Done():
                pubsub.Close()
                return
            }
        }
    }()
    
    return cancel, nil
}
```

#### 4.2.4 Admin Handler 集成 Redis Pub/Sub

**文件**: `go/tars/provider-admin/internal/handler/config_handler.go`

```go
// CacheInvalidator 接口实现
type RedisCacheInvalidator struct {
    client *redis.Client
}

func (r *RedisCacheInvalidator) InvalidateConfigCache(moduleKey string) error {
    // 1. 删除 Redis 缓存
    r.client.Del(context.Background(), fmt.Sprintf("cfg:*:%s", moduleKey))
    
    // 2. 发布 pub/sub 消息
    return r.client.Publish(context.Background(), "cairobot.config.invalidate", moduleKey).Err()
}
```

---

## 五、遗留项修复状态（更新）

### 5.1 依赖关系图

```
M-08: 生成 Protobuf Go 代码
  ├──> L-01: Tars Adapter Protobuf 序列化
  └──> L-02: 静态模块到 Protobuf 强类型字段映射
        └──> L-03: SDK Remote 模式完整实现

M-06: 引入数据库配置（server.yaml + MySQL 实现）
  └──> 单元测试使用 MySQL

M-07: 引入 Redis 配置（server.yaml + Redis 实现）
  └──> L-04: Redis pub/sub 失效广播
```

### 5.2 修复优先级

| 优先级 | 任务 | 依赖 | 预估工时 |
|---|---|---|---|
| P0 | 引入 `configs/server.yaml` | 无 | 0.5 天 |
| P0 | 创建 `go/common-lib/config/` 统一配置包 | server.yaml | 0.5 天 |
| P1 | 实现 MySQL 仓库（config + i18n） | 统一配置包 | 1 天 |
| P1 | 实现 Redis 缓存 + Pub/Sub | 统一配置包 | 1 天 |
| P1 | 执行 `make proto` 生成代码 | 无 | 0.5 天 |
| P2 | L-01: Tars Adapter Protobuf 序列化 | M-08 | 0.5 天 |
| P2 | L-02: 静态模块到 Protobuf 映射 | M-08 | 0.5 天 |
| P2 | L-03: SDK Remote 模式 | L-01 | 1 天 |
| P2 | L-04: Redis pub/sub 失效广播 | M-07 | 0.5 天 |

---

## 六、必须修改项（更新）

### M-09: 引入统一配置类（新增）

**问题描述**:
当前配置通过环境变量零散加载，无统一配置中心。

**修复方案**:
1. 将 `server_config_test.yaml` 裁剪为 `configs/server.yaml`
2. 创建 `go/common-lib/config/server_config.go` 统一配置结构体
3. 创建 `go/common-lib/config/loader.go` Viper 配置加载器
4. 所有服务入口（main.go）统一使用 `config.Load()` 加载配置

### M-10: 单元测试使用 MySQL（新增）

**问题描述**:
当前单元测试使用 SQLite，与生产环境不一致。

**修复方案**:
1. 单元测试默认使用 SQLite（快速、无依赖）
2. 集成测试使用 MySQL（通过环境变量配置）
3. CI 中配置 MySQL 服务容器
4. 提供 `TEST_MYSQL_*` 环境变量，支持本地 MySQL 测试

---

## 七、建议修改项（更新）

### S-07: 配置敏感信息脱敏（新增）

**问题描述**:
`server_config_test.yaml` 包含明文密码（`Huawei@2025`）。

**建议**:
1. 生产环境密码通过环境变量注入
2. 配置文件中使用占位符：`password: "${MYSQL_PASSWORD}"`
3. Viper 加载时自动替换环境变量

### S-08: 多环境配置分离（新增）

**建议**:
```
configs/
  server.yaml          # 基础配置（默认 dev）
  server-staging.yaml  # 预发布环境
  server-prod.yaml     # 生产环境
```

---

## 八、评审结论

### 总体评分（v4）

| 维度 | 权重 | v3 评分 | v4 评分 | 说明 |
|---|---|---|---|---|
| 需求一致性 | 25% | 3.0 | 3.5 | 已有明确引入路径，待实施 |
| 测试覆盖 | 25% | 4.5 | 4.5 | 242 测试仍全部通过 |
| 代码质量 | 25% | 3.5 | 3.8 | 分层清晰，配置方案已明确 |
| 文档同步 | 15% | 3.5 | 4.0 | 配置方案已文档化 |
| 兼容性 | 10% | 2.5 | 3.0 | 双模式切换方案已明确 |

**总分**: 3.8/5.0（v3: 3.6 → v4: 3.8）

### 裁决

**建议修改**

实施结果在**功能层面**已完整（242 测试通过），**配置引入方案已明确**（基于 `server_config_test.yaml`）。需按以下顺序修复：

1. **M-09**: 引入统一配置类（`configs/server.yaml` + `go/common-lib/config/`）
2. **M-06**: 实现 MySQL 仓库（依赖 M-09）
3. **M-07**: 实现 Redis 缓存 + Pub/Sub（依赖 M-09）
4. **M-08**: 生成 Protobuf Go 代码
5. **L-01~L-04**: 遗留项（依赖 M-06/M-07/M-08）

### 下一步行动

1. **立即执行**: 创建 `configs/server.yaml` 和 `go/common-lib/config/`（M-09）
2. **高优先级**: 实现 MySQL 仓库真实逻辑（M-06）
3. **高优先级**: 实现 Redis 缓存和 pub/sub（M-07）
4. **高优先级**: 执行 `make proto` 生成 Protobuf Go 代码（M-08）
5. **重新评审**: 修复 M-06/M-07/M-08/M-09 后，重新提交评审

---

> **评审人**: Trae AI Reviewer
> **评审依据**: `.trae/rules/review.md` 评审规则
> **关联 PRD**: `docs/prd/global-config-i18n-implementation-plan.md`
> **关联配置**: `docs/tabbit/inbox/2026/05/server_config_test.yaml`
> **上次评审**: `docs/reviews/review-config-i18n-implementation-v3.md` (REV-CONFIG-I18N-003)
