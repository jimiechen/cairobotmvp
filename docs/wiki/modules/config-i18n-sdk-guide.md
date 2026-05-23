# Config + I18n SDK 使用指南

## SDK 概述

CaiRobot MVP 提供双 SDK 架构，分别封装配置服务和多语言服务的客户端访问能力：

| SDK | 包路径 | 职责 |
|---|---|---|
| configsdk | `go/services/config/sdk` | 类型安全的配置读取、缓存、变更订阅 |
| i18nsdk | `go/services/i18n/sdk` | 多语言模板渲染、批量翻译、变更订阅 |

**设计原则**:
- 屏蔽 InProcess/Remote 模式差异
- 提供类型安全的 API（避免 interface{}）
- 内置多级缓存（L1 LRU → L2 Redis → L3 Service）
- 支持变更订阅（Watch 机制）

## 快速开始

### InProcess 模式初始化

适用于单进程部署（如 TarsGo Gateway 进程内调用）。

```go
package main

import (
    "context"
    "fmt"
    "log"

    configsdk "github.com/jimiechen/mineplanet/go/services/config/sdk"
    i18nsdk "github.com/jimiechen/mineplanet/go/services/i18n/sdk"
)

func main() {
    ctx := context.Background()

    // 1. 初始化 ConfigService（假设已有实现）
    configService := NewAppConfigService(configRepo, schemaRepo, cache)

    // 2. 创建 configsdk 客户端
    configClient, err := configsdk.Default(
        configsdk.WithMode(configsdk.ModeInProcess),
        configsdk.WithService(configService),
        configsdk.WithEnv("prod"),
        configsdk.WithClientScope("app"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 3. 初始化 I18nService
    i18nService := NewI18nServiceImpl(i18nRepo, cache)

    // 4. 创建 i18nsdk 客户端
    i18nClient, err := i18nsdk.Default(
        func(opts *i18nsdk.Options) {
            opts.Mode = i18nsdk.ModeInProcess
            opts.Service = i18nService
            opts.Env = "prod"
            opts.DefaultLangCode = "zh-CN"
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // 5. 使用 SDK
    appName, _ := configClient.GetString(ctx, "base_cfg", "app_name")
    greeting, _ := i18nClient.T(ctx, "zh-CN", "greeting.welcome",
        map[string]any{"userName": "张三", "taskCount": 5})

    fmt.Printf("App: %s\n", appName)
    fmt.Printf("Greeting: %s\n", greeting)
}
```

## configsdk API 详解

### 核心接口

```go
type Client interface {
    // 获取字符串类型配置值
    GetString(ctx context.Context, moduleKey, fieldKey string) (string, error)

    // 获取整数类型配置值（int64）
    GetInt(ctx context.Context, moduleKey, fieldKey string) (int64, error)

    // 获取布尔类型配置值
    GetBool(ctx context.Context, moduleKey, fieldKey string) (bool, error)

    // 获取浮点数类型配置值（float64）
    GetFloat(ctx context.Context, moduleKey, fieldKey string) (float64, error)

    // 获取 JSON 类型配置值并反序列化到 out
    GetJSON(ctx context.Context, moduleKey, fieldKey string, out any) error

    // 获取整个模块的配置快照
    GetModule(ctx context.Context, moduleKey string) (*ModuleSnapshot, error)

    // 将模块配置绑定到结构体（自动映射字段）
    Bind(ctx context.Context, moduleKey string, out any) error

    // 订阅模块配置变更
    Watch(moduleKey string, handler func(*ModuleSnapshot)) (cancel func())

    // 检查服务可用性
    Ping(ctx context.Context) error
}
```

### GetString / GetInt / GetBool / GetFloat

类型安全的单个字段读取。

```go
// 读取字符串
appName, err := configClient.GetString(ctx, "base_cfg", "app_name")

// 读取整数
timeout, err := configClient.GetInt(ctx, "base_cfg", "request_timeout_ms")

// 读取布尔
isDebug, err := configClient.GetBool(ctx, "base_cfg", "debug_mode")

// 读取浮点数
rate, err := configClient.GetFloat(ctx, "pay_cfg", "exchange_rate")
```

**查询路径**: L1 LRU → L2 Redis → L3 远程服务

### GetJSON

读取 JSON 类型字段并反序列化。

```go
type OSSConfig struct {
    Endpoint  string `json:"endpoint"`
    Bucket    string `json:"bucket"`
    AccessKey string `json:"access_key"`
}

var ossConfig OSSConfig
err := configClient.GetJSON(ctx, "oss_cfg", "credentials", &ossConfig)
```

**适用场景**: FieldTypeJSON / FieldTypeList

### GetModule

获取整个模块的配置快照。

```go
snapshot, err := configClient.GetModule(ctx, "base_cfg")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Module: %s, Version: %d\n", snapshot.ModuleKey, snapshot.Version)
for fieldKey, tv := range snapshot.Fields {
    fmt.Printf("  %s = %v (%s)\n", fieldKey, tv.Value, tv.Type)
}
```

**返回结构**:

```go
type ModuleSnapshot struct {
    ModuleKey string
    Version   int64
    Fields    map[string]*domain.TypedValue
}
```

### Bind

将模块配置自动绑定到结构体（基于字段名映射）。

```go
type BaseConfig struct {
    AppName         string  `config:"app_name"`
    RequestTimeout  int64   `config:"request_timeout_ms"`
    DebugMode       bool    `config:"debug_mode"`
}

var cfg BaseConfig
err := configClient.Bind(ctx, "base_cfg", &cfg)
```

### Watch

订阅模块配置变更。

```go
cancel := configClient.Watch("base_cfg", func(snapshot *ModuleSnapshot) {
    log.Printf("Config changed: version=%d", snapshot.Version)
    // 重新加载配置...
})

// 取消订阅
// cancel()
```

**返回**: cancel 函数，调用后取消订阅。

## i18nsdk API 详解

### 核心接口

```go
type Client interface {
    // 翻译指定 key 的文本并渲染参数
    T(ctx context.Context, langCode, key string, params map[string]any) (string, error)

    // 获取原始模板信息（不渲染）
    Raw(ctx context.Context, langCode, key string) (*Template, error)

    // 批量翻译多个 key
    BatchT(ctx context.Context, langCode string, keys []string, params map[string]any) (map[string]string, error)

    // 订阅语言包版本变更
    Watch(langCode string, handler func(packVersion int64)) (cancel func())

    // 检查服务可用性
    Ping(ctx context.Context) error
}
```

### T() 渲染

核心翻译方法，支持参数化模板渲染。

```go
// plain 类型：直接返回
title, _ := i18nClient.T(ctx, "zh-CN", "app.name", nil)
// → "CaiRobot 智能学习助手"

// named 类型：参数替换
greeting, _ := i18nClient.T(ctx, "zh-CN", "greeting.welcome",
    map[string]any{
        "userName":  "张三",
        "taskCount": 42,
    })
// → "欢迎 张三，今天有 42 个任务待完成"

// key 不存在：返回 key 本身（fallback）
unknown, _ := i18nClient.T(ctx, "zh-CN", "nonexistent.key", nil)
// → "nonexistent.key"（不报错）
```

**icu 类型**: 返回 `ErrICUNotSupported` 错误（MVP 未实现）。

### Raw() 原始

获取原始模板信息，不执行渲染。

```go
tmpl, err := i18nClient.Raw(ctx, "zh-CN", "greeting.welcome")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Key: %s\n", tmpl.Key)
fmt.Printf("Value: %s\n", tmpl.Value)
fmt.Printf("Type: %s\n", tmpl.TemplateType)  // named
for _, param := range tmpl.Params {
    fmt.Printf("  Param: %s (%s, required=%v)\n",
        param.Name, param.Type, param.Required)
}
```

**返回结构**:

```go
type Template struct {
    Key          string
    Value        string
    TemplateType string  // plain/named/icu
    Params       []ParamInfo
}

type ParamInfo struct {
    Name     string
    Type     string
    Required bool
}
```

### BatchT 批量

批量翻译多个 key，减少网络开销。

```go
keys := []string{"app.name", "greeting.welcome", "common.confirm"}
results, err := i18nClient.BatchT(ctx, "zh-CN", keys,
    map[string]any{"userName": "张三", "taskCount": 42})

for key, value := range results {
    fmt.Printf("%s = %s\n", key, value)
}
```

**返回**: `map[string]string`，key 对应翻译结果。

### Watch 订阅

订阅语言包版本变更。

```go
cancel := i18nClient.Watch("zh-CN", func(packVersion int64) {
    log.Printf("Language pack updated: version=%d", packVersion)
    // 清除本地缓存，下次 T() 时重新拉取
})

// 取消订阅
// cancel()
```

## named 模板渲染规则

### {paramName} 替换算法

```go
func renderNamedTemplate(value string, params map[string]any, paramInfos []ParamInfo) (string, error) {
    // 1. 收集所有 required 参数
    requiredParams := make(map[string]bool)
    for _, info := range paramInfos {
        if info.Required {
            requiredParams[info.Name] = true
        }
    }

    // 2. 检查必需参数是否齐全
    for paramName := range requiredParams {
        if _, exists := params[paramName]; !exists {
            return "", ErrMissingRequiredParam
        }
    }

    // 3. 正则替换占位符
    result := namedPlaceholderRegex.ReplaceAllStringFunc(value, func(match string) string {
        paramName := extractParamName(match)
        if val, exists := params[paramName]; exists {
            return fmt.Sprintf("%v", val)
        }
        return match  // 未声明的占位符保留原样
    })

    return result, nil
}
```

### required 参数缺失行为

当 required 参数未提供时，返回错误：

```go
_, err := i18nClient.T(ctx, "zh-CN", "greeting.welcome",
    map[string]any{"userName": "张三"})
// → ErrMissingRequiredParam: taskCount
```

### fallback 策略

| 场景 | 行为 |
|---|---|
| key 不存在 | 返回 key 本身（不报错） |
| 多余参数 | 忽略 |
| 未声明占位符 | 保留原样 `{xxx}` |
| icu 类型 | 返回 `ErrICUNotSupported` |

## 双模式切换

### ModeInProcess（进程内模式）

直接调用 Service 层，无网络开销。适用于：
- TarsGo Gateway 进程内调用
- 单元测试
- 本地开发

```go
client, _ := configsdk.Default(
    configsdk.WithMode(configsdk.ModeInProcess),
    configsdk.WithService(configService),
)
```

### ModeRemote（远程模式）

通过 TarsGo 调用远程服务。适用于：
- 微服务架构
- 多实例部署

```go
client, _ := configsdk.Default(
    configsdk.WithMode(configsdk.ModeRemote),
    configsdk.WithTarsServant("ConfigServer.ConfigObj"),
    configsdk.WithRedis(redisClient),
)
```

**注意**: Remote 模式的完整实现在 MVP 阶段为占位状态。

## 三级缓存架构

```
┌─────────────────────────────────────────────┐
│ Client App                                   │
│  ┌──────────┐                                │
│  │ L1 LRU   │ ← 内存缓存，容量可配（默认 256）│
│  │ Cache    │                                │
│  └────┬─────┘                                │
│       │ miss                                  │
│  ┌────▼─────┐                                │
│  │ L2 Redis │ ← 分布式缓存，支持 Pub/Sub      │
│  │ Cache    │                                │
│  └────┬─────┘                                │
│       │ miss                                  │
│  ┌────▼─────┐                                │
│  │ L3 Service│ ← 远程服务（TarsGo / HTTP）    │
│  └──────────┘                                │
└─────────────────────────────────────────────┘
```

### L1 LRU 缓存

- **位置**: 进程内存
- **数据结构**: LRU（最近最少使用）
- **容量**: 可配置（默认 256）
- **TTL**: 可配置（默认 30 秒）
- **失效策略**: TTL 过期 + Watch 主动失效

### L2 Redis 缓存

- **位置**: Redis 集群
- **用途**: 跨实例共享 + Pub/Sub 变更通知
- **TTL**: 与 L1 一致或更长
- **特性**: 支持 Subscribe 订阅变更频道

### L3 Service 层

- **位置**: ConfigService / I18nService
- **实现**: InProcess 直接调用 / Remote TarsGo RPC
- **降级**: 返回默认值或错误

## Watch 机制

### 配置变更订阅

```go
cancel := configClient.Watch("base_cfg", func(snapshot *ModuleSnapshot) {
    log.Printf("Config v%d changed:", snapshot.Version)
    for k, v := range snapshot.Fields {
        log.Printf("  %s = %v", k, v.Value)
    }
})
```

**工作原理**:
1. L2 Redis Pub/Sub 订阅 `config:change:{moduleKey}` 频道
2. 收到消息后回调 handler
3. 同时清除 L1 LRU 缓存中对应 module

### 语言包变更订阅

```go
cancel := i18nClient.Watch("zh-CN", func(packVersion int64) {
    log.Printf("Lang pack updated to v%d", packVersion)
})
```

### 取消订阅

所有 Watch 方法返回 cancel 函数：

```go
cancel()  // 取消订阅，释放资源
```

**最佳实践**: 在 goroutine 生命周期结束时调用 cancel()。

## 错误处理

### configsdk 错误常量

```go
var (
    ErrServiceRequired = errors.New("config service is required in InProcess mode")
    ErrUnsupportedMode = errors.New("unsupported SDK mode")
    ErrModuleNotFound  = errors.New("module not found")
    ErrFieldNotFound   = errors.New("field not found")
    ErrTypeMismatch    = errors.New("field type mismatch")
    ErrBindFailed      = errors.New("bind to struct failed")
)
```

### i18nsdk 错误常量

```go
var (
    ErrICUNotSupported     = errors.New("icu template not supported in MVP")
    ErrMissingRequiredParam = errors.New("missing required parameter")
)
```

### 错误处理示例

```go
value, err := configClient.GetString(ctx, "base_cfg", "app_name")
if err != nil {
    switch err {
    case configsdk.ErrModuleNotFound:
        log.Printf("Module not found, using default")
        value = "Default App Name"
    case configsdk.ErrFieldNotFound:
        log.Printf("Field not found, using default")
        value = "Default App Name"
    default:
        log.Fatalf("Fatal error: %v", err)
    }
}
```

## 最佳实践

### 单例初始化

SDK Client 应全局单例，避免重复创建：

```go
var (
    configClient configsdk.Client
    i18nClient   i18nsdk.Client
    initOnce     sync.Once
    initErr      error
)

func InitSDKs() error {
    initOnce.Do(func() {
        configClient, initErr = configsdk.Default(...)
        i18nClient, initErr = i18nsdk.Default(...)
    })
    return initErr
}

func GetConfigClient() configsdk.Client {
    return configClient
}

func GetI18nClient() i18nsdk.Client {
    return i18nClient
}
```

### 避免频繁创建 Client

**错误做法**:

```go
// 每次请求都创建新 Client（性能差）
func handleRequest(req Request) {
    client, _ := configsdk.Default(...)
    value, _ := client.GetString(ctx, ...)
}
```

**正确做法**:

```go
// 全局复用 Client
var globalClient configsdk.Client

func init() {
    globalClient, _ = configsdk.Default(...)
}

func handleRequest(req Request) {
    value, _ := globalClient.GetString(ctx, ...)
}
```

### 合理设置缓存容量

根据业务场景调整 LRU 容量：

| 场景 | 建议容量 | 说明 |
|---|---|---|
| 小型应用 | 128-256 | 默认值，适合模块较少的场景 |
| 中型应用 | 512-1024 | 模块较多时使用 |
| 大型应用 | 2048+ | 微服务架构，模块数量大 |

```go
client, _ := configsdk.Default(
    configsdk.WithCacheSize(512),
    configsdk.WithCacheTTL(60),  // 60 秒 TTL
)
```

### Watch 资源管理

确保在组件销毁时取消订阅：

```go
type MyComponent struct {
    cancelFunc func()
}

func (c *MyComponent) Start() {
    c.cancelFunc = configClient.Watch("my_module", c.onConfigChanged)
}

func (c *MyComponent) Stop() {
    if c.cancelFunc != nil {
        c.cancelFunc()
    }
}
```

## 相关文档

- [config-service.md](./config-service.md): 配置服务模块文档
- [i18n-service.md](./i18n-service.md): 多语言服务模块文档
- Proto: [app_config.proto](../../../proto/app_config.proto), [i18n.proto](../../../proto/i18n.proto)
- 协议编号: 6001-6010（见 [协议编号注册表.md](../api/协议编号注册表.md)）
