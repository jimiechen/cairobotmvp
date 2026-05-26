# Hello 模块

## 1. 模块职责

演示 configsdk + i18nsdk 接入规范的参考实现，提供多语言问候服务。

## 2. 协议清单

| 协议 | maxType | minType | 说明 |
|------|---------|---------|------|
| HelloWorldRequest | 2100 | 2101 | 问候请求 |
| HelloWorldResponse | 2100 | 2102 | 问候响应 |

### 请求字段 (HelloWorldRequest)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 问候名称，为空时默认 "World" |
| lang_code | string | 否 | 语言代码，如 zh-CN、en，默认 zh-CN |

### 响应字段 (HelloWorldResponse)

| 字段 | 类型 | 说明 |
|------|------|------|
| result | Result | 统一返回结果码 |
| message | string | 兼容字段，保留旧版格式 |
| timestamp | int64 | 服务端时间戳（Unix 秒） |
| greeting | string | **新增** i18n 渲染的多语言问候语 |
| server_name | string | **新增** 配置驱动的服务名称 |
| max_name_length | int32 | **新增** 配置驱动的名称长度限制 |

## 3. 配置 Schema 清单

**module_key:** `hello_cfg`

| field_key | 类型 | 默认值 | 校验规则 | 说明 |
|-----------|------|--------|----------|------|
| server_name | string | "CaiRobot" | - | 服务名称，用于问候语渲染 |
| max_name_length | int | 32 | min=1, max=256 | 用户名最大长度，超出返回 10400 |

## 4. 多语言 Key 清单

**key:** `svc_hello_greeting`
**template_type:** named
**params_schema:** `[{name, string, required}, {server_name, string, required}]`

| 语言代码 | 模板文本 |
|----------|----------|
| zh-CN | 你好，{name}！欢迎使用 {server_name}。 |
| en | Hello, {name}! Welcome to {server_name}. |

## 5. 依赖关系

| 依赖 | 类型 | 用途 |
|------|------|------|
| configsdk.Client | SDK（必选） | 读取 server_name、max_name_length 配置 |
| i18nsdk.Client | SDK（可选） | 渲染多语言问候语，nil 时降级到英文 fallback |
| log.Logger | 基础设施（必选） | 日志记录 |

### 禁止事项

- ❌ 禁止直接 import services/config 或 services/i18n 内部包
- ❌ 禁止直接 sql.Open / redis.NewClient
- ❌ 禁止硬编码用户可见中英文文案

## 6. 健康检查

Hello 模块无外部依赖，无需实现 Checker 接口。

## 文件结构

```
go/modules/hello/
├── service.go           # 依赖装配入口（≤80 行）
├── handler.go           # Protobuf 编解码层（≤120 行）
├── usecase.go           # 核心业务逻辑（≤150 行）
├── usecase_test.go      # 单元测试（使用 Fake SDK）
└── README.md            # 本文档
```

## 接入示例

```go
// main.go 中装配
deps := module.Deps{
    Config: configsdk.Default(),
    I18n:   i18nsdk.Default(),
    Logger: log.Default(),
}
helloSvc := hello.New(deps)
```

## Seed 数据

模块首启时需执行 [hello_seed.sql](../../../migrations/seed/hello_seed.sql) 注入配置 Schema 和多语言文案。

## SDK 引用清单（Checklist #11）

### configsdk 调用点

| module_key | field_key | 调用方法 | 读时机 | 降级默认值 |
|------------|-----------|----------|--------|------------|
| hello_cfg | server_name | GetString | 每次请求时 | "CaiRobot" |
| hello_cfg | max_name_length | GetInt | 每次请求时 | 32 |
| system_cfg | default_lang_code | GetString | ResolveLang 内部 | "zh-CN" |

### i18nsdk 调用点

| key | template_type | 参数 schema | 调用时机 | fallback |
|-----|---------------|-------------|----------|----------|
| svc_hello_greeting | named | `{name: string, server_name: string}` | 响应渲染时 | "Hello, {name}! Welcome to {server_name}." |

### health.Checker 注册表

Hello 模块无外部依赖，不注册 Checker。

### 语言解析优先级

1. `MessagePacket.extend.langCode`（最高）
2. 协议体 `lang_code` 字段
3. `configsdk.GetString(ctx, "system_cfg", "default_lang_code")`
4. 硬编码 `"zh-CN"`

实现入口：[ResolveLang()](../../common-lib/i18n/lang_resolver.go)
