# Health 模块

## 1. 模块职责

演示 i18nsdk ICU plural 模板 + Checker 抽象 + 多依赖健康汇总的参考实现，提供服务间心跳检测和依赖健康状态监控。

## 2. 协议清单

| 协议 | maxType | minType | 说明 |
|------|---------|---------|------|
| ServiceHealthCheckRequest | 2100 | 2097 | 健康检查请求 |
| ServiceHealthCheckResponse | 2100 | 2098 | 健康检查响应 |
| ComponentStatus | - | - | 组件状态（内嵌 message） |

### 请求字段 (ServiceHealthCheckRequest)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| service_name | string | 否 | 请求方服务名称（保留兼容） |
| lang_code | string | 否 | 语言代码，如 zh-CN、en，默认 zh-CN |
| depth | int32 | 否 | 检查深度：0=仅自身存活，1=依赖 ping，2=依赖+实际查询 |

### 响应字段 (ServiceHealthCheckResponse)

| 字段 | 类型 | 说明 |
|------|------|------|
| result | Result | 统一返回结果码 |
| status | string | 状态："OK" 或 "Unhealthy" |
| timestamp | int64 | 服务端时间戳（Unix 秒） |
| version | string | **新增** 构建版本号（configsdk 读取） |
| message | string | **新增** i18n 渲染的状态摘要（ICU plural） |
| components[] | ComponentStatus | **新增** 各依赖组件健康状态列表 |

### 内嵌字段 (ComponentStatus)

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 组件名称：mysql、redis、config、i18n |
| healthy | bool | 是否健康 |
| latency_ms | int64 | 检查耗时（毫秒） |
| error | string | 错误信息（healthy=true 时为空） |

## 3. 配置 Schema 清单

**module_key:** `system_cfg` / `health_cfg`

| module_key | field_key | 类型 | 默认值 | 校验规则 | 说明 |
|------------|-----------|------|--------|----------|------|
| system_cfg | build_version | string | "0.0.0-dev" | - | 构建版本号，CI/CD 注入 |
| health_cfg | max_depth | int | 2 | min=0, max=3 | 最大检查深度限制 |

## 4. 多语言 Key 清单

**key:** `svc_health_status_summary`
**template_type:** icu（ICU MessageFormat）
**params_schema:** `[{healthy, int, required}, {total, int, required}]`

| 语言代码 | 模板文本 |
|----------|----------|
| zh-CN | `{healthy, plural, =0 {所有依赖均不可用（{total} 项）} other {# / {total} 项依赖正常}}` |
| en | `{healthy, plural, =0 {All {total} dependencies are down} other {# of {total} dependencies healthy}}` |

## 5. 依赖关系

| 依赖 | 类型 | 用途 |
|------|------|------|
| configsdk.Client | SDK（必选） | 读取 build_version、max_depth 配置 |
| i18nsdk.Client | SDK（可选） | 渲染 ICU plural 状态摘要，nil 时降级到英文 fallback |
| health.Checker[] | 接口数组（可选） | 各依赖健康检查器，depth=0 时不调用 |
| log.Logger | 基础设施（必选） | 日志记录 |

### 内置 Checker 实现

| Checker | 名称 | 实现状态 |
|---------|------|----------|
| ConfigChecker | config | ✅ 已实现（通过 configsdk.Ping） |
| I18nChecker | i18n | ✅ 已实现（通过 i18nsdk.Ping） |
| MySQLChecker | mysql | 🔧 预留接口（后续接入 mysqlx.DB） |
| RedisChecker | redis | 🔧 预留接口（后续接入 redisx.Client） |

### 禁止事项

- ❌ 禁止直接 import services/config 或 services/i18n 内部包
- ❌ 禁止直接 sql.Open / redis.NewClient
- ❌ 禁止硬编码用户可见中英文文案

## 6. 健康检查

Health 模块自身实现 Checker 接口注册机制：
- 每个 Checker 有独立 1s 超时
- 并发执行所有 Checker，互不影响
- 失败状态写入 ComponentStatus.Error 字段

### 接入示例

```go
// main.go 中装配
deps := module.Deps{
    Config: configsdk.Default(),
    I18n:   i18nsdk.Default(),
    Logger: log.Default(),
}

checkers := []health.Checker{
    health.NewConfigChecker(deps.Config),
    health.NewI18nChecker(deps.I18n),
}

healthSvc := health.New(deps, checkers)
```

## Depth 参数说明

| Depth | 行为 | 适用场景 |
|-------|------|----------|
| 0 | 仅返回自身存活状态（永远 OK） | 快速探活、负载均衡检查 |
| 1 | Ping 各依赖服务 | 标准健康检查 |
| 2 | Ping + 一次轻量查询（如 SELECT 1） | 深度巡检、故障诊断 |

## 文件结构

```
go/modules/health/
├── service.go           # 依赖装配入口（≤80 行）
├── handler.go           # Protobuf 编解码层（≤120 行）
├── usecase.go           # 核心业务逻辑（≤120 行）
├── checker.go           # 四个 Checker 实现（≤150 行）
├── usecase_test.go      # 单元测试（使用 Fake SDK）
├── checker_test.go      # Checker 测试
└── README.md            # 本文档
```

## Seed 数据

模块首启时需执行 [health_seed.sql](../../../migrations/seed/health_seed.sql) 注入配置 Schema 和 ICU 多语言文案。

## SDK 引用清单（Checklist #11）

### configsdk 调用点

| module_key | field_key | 调用方法 | 读时机 | 降级默认值 |
|------------|-----------|----------|--------|------------|
| system_cfg | build_version | GetString | 每次请求时 | "0.0.0-dev" |
| health_cfg | max_depth | GetInt | 每次请求时 | 2 |
| system_cfg | default_lang_code | GetString | ResolveLang 内部 | "zh-CN" |

### i18nsdk 调用点

| key | template_type | 参数 schema | 调用时刻 | fallback |
|-----|---------------|-------------|----------|----------|
| svc_health_status_summary | icu | `{healthy: int, total: int}` | 响应渲染时 | "{healthy} of {total} dependencies healthy" |

### health.Checker 注册表

| Checker 名称 | 依赖 | 超时 | 实现位置 |
|-------------|------|------|----------|
| ConfigChecker | configsdk.Client.Ping() | 1s | [checker.go](checker.go) |
| I18nChecker | i18nsdk.Client.Ping() | 1s | [checker.go](checker.go) |
| MySQLChecker | mysqlx.DB.Ping() | 1s | [checker.go](checker.go) |
| RedisChecker | redisx.Client.Ping() | 1s | [checker.go](checker.go) |

> MySQLChecker / RedisChecker 在 Deps.DB 或 Deps.Cache 为 nil 时自动跳过，返回 unhealthy 状态。

### TruncateError 使用说明

所有 Checker 的 `ComponentStatus.Error` 字段经 [TruncateError()](../../common-lib/i18n/error_truncater.go) 截断，默认上限 512 字符，UTF-8 rune 级别安全截断。

### 语言解析优先级

1. `MessagePacket.extend.langCode`（最高）
2. 协议体 `lang_code` 字段
3. `configsdk.GetString(ctx, "system_cfg", "default_lang_code")`
4. 硬编码 `"zh-CN"`

实现入口：[ResolveLang()](../../common-lib/i18n/lang_resolver.go)

## 相关文档

- [Checker 接口定义](../../common-lib/health/checker.go) - 所有模块复用的健康检查抽象
- [Hello 模块](../hello/README.md) - configsdk 接入范例
