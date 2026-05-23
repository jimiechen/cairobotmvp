# 配置服务模块 (config-service)

## 模块概述

**职责**: 全局应用配置元模型的管理与查询服务，提供版本化、Schema 驱动的强类型配置能力。

**不负责**:
- HTTP 路由处理（由 tars/provider-admin 负责）
- UI 渲染（由 web/provider-admin 负责）
- 客户端缓存策略（由 configsdk 负责）

**关键输入**:
- `env`: 环境标识（dev/staging/prod）
- `clientScope`: 客户端范围（all/app/miniprogram）
- `clientVersion`: 客户端版本号
- `moduleKey`: 模块键名

**关键输出**:
- `AppConfigResponse`: 包含 StaticModules + DynamicModules 的完整配置视图
- `VersionInfoResponse`: 版本轮询响应，包含变更标记

## 目录结构与职责

```
go/services/config/
├── domain/              # 领域实体与值对象
│   ├── module_key.go    # 预定义 8 个强类型模块键常量
│   ├── schema.go        # FieldSchema / ModuleSchema 字段元数据
│   ├── value.go         # TypedValue 类型安全值包装器
│   └── version.go       # ConfigVersion 版本实体
├── repository/          # 数据访问层
│   ├── interface.go     # ConfigRepository / SchemaRepository 接口定义
│   ├── mysql_repo.go    # MySQL 持久化实现
│   ├── sqlite_repo.go   # SQLite 实现（测试用）
│   ├── schema_repo.go   # Schema 专用 Repository
│   └── utils.go         # DDL 建表工具函数
├── cache/               # 缓存层
│   ├── interface.go     # Cache 接口定义
│   └── mock_cache.go    # Mock 实现（测试用）
├── service/             # 业务逻辑层
│   ├── interface.go     # ConfigService 接口 + AppConfigService 实现
│   ├── compose.go       # 静态/动态模块分流组装算法
│   ├── fetch.go         # 多级加载逻辑（Cache → DB）
│   ├── parse.go         # JSON → TypedMap 解析
│   ├── validate.go      # Schema 校验逻辑
│   └── schema_service.go # Schema 查询服务
└── sdk/                 # 客户端 SDK
    ├── client.go        # Client 接口 + configClient 实现
    ├── types.go         # ModuleSnapshot 等类型定义
    ├── errors.go        # 错误常量定义
    ├── get.go           # GetString/GetInt/GetBool/GetFloat/GetJSON
    ├── module.go        # GetModule/Bind
    ├── watch.go         # Watch 变更订阅
    ├── cache_lru.go     # L1 LRU 缓存实现
    ├── pubsub.go        # Redis Pub/Sub 订阅
    └── remote.go        # L3 远程调用（TarsGo 占位）
```

**行数约束**: 单文件不超过 200 行，单方法不超过 30 行。

## 核心数据流

### 写路径 (Admin → MySQL)

```
Provider Admin (React)
    ↓ HTTP POST /api/v1/configs/:module_key
Tars Provider Admin (Gin Handler)
    ↓ 调用 ConfigRepository.Save()
MySQL (sys_config_version 表)
    ↓ 写入 config_json + version +1
```

### 读路径 (Client → Gateway → Tars → Service → DB)

```
Client App
    ↓ Proto 6001 AppConfigsReq (env, client_scope, requested_modules)
API Gateway (TarsGo)
    ↓ LocalInvoker 调用 ConfigService.GetAppConfigs()
ConfigService (AppConfigService)
    ↓ 1. 查 Cache (L2 Redis)
    ↓ 2. miss 则查 Repository (MySQL)
    ↓ 3. ParseConfigJSON() 解析为 TypedMap
    ↓ 4. ComposeFullResponse() 组装响应
AppConfigsRsp (Proto)
    ↓ static_modules (8 个强类型 map)
    ↓ dynamic_modules (其余模块列表)
Client App 收到完整配置
```

## 核心实体

### ConfigVersion

对应 `sys_config_version` 表，承载某个 module_key 在特定 env 下的版本化配置快照。

```go
type ConfigVersion struct {
    ID          int64
    ModuleKey   string      // 模块键名
    Env         string      // 环境
    Version     int64       // 版本号（单调递增）
    ConfigJSON  string      // JSON 格式的配置内容
    IsPublished bool        // 是否已发布
    PublishedAt *time.Time  // 发布时间
}
```

**关键方法**: `IsReleased()` 判断是否可对外提供服务。

### FieldSchema

对应 `sys_config_schema` 表，描述单个字段的结构约束。

```go
type FieldSchema struct {
    ID           int64
    ModuleKey    string
    FieldKey     string
    FieldType    FieldType   // string/int/bool/float/enum/json/list
    DefaultValue string
    Validator    string      // 校验规则表达式
    IsRequired   bool
    IsSecret     bool        // 是否敏感字段（脱敏）
    ClientScope  string      // 客户端可见范围
    MinAppVer    string      // 最低支持版本
}
```

**关键方法**: `MatchClientScope()` 判断字段对指定客户端是否可见。

### TypedValue

类型安全的配置值包装器，防止下游直接使用 interface{} 导致的类型 panic。

```go
type TypedValue struct {
    Type  FieldType
    Value any
}

// 提供类型安全访问方法
func (tv *TypedValue) String() string
func (tv *TypedValue) Int() int64
func (tv *TypedValue) Bool() bool
func (tv *TypedValue) Float() float64
func (tv *TypedValue) JSON() json.RawMessage
```

### DynamicModuleView

动态模块的业务视图，从 ConfigVersion + ModuleSchema 组装而来。

```go
type DynamicModuleView struct {
    ModuleKey   string
    Version     int64
    Fields      map[string]*domain.TypedValue
    Descriptors []*FieldDescriptorView  // 来自 sys_config_schema
}
```

## Compose 核心算法

`ComposeFullResponse()` 是配置服务的核心组装函数，负责将原始数据分为静态和动态两组。

### 分流逻辑

```go
func ComposeFullResponse(
    env, clientScope string,
    versions []*domain.ConfigVersion,
    schemaRepo repository.SchemaRepository,
    requestedModules []string,
) *AppConfigResponse {
    for _, ver := range versions {
        typedMap, _ := ParseConfigJSON(ver.ConfigJSON, ver.ModuleKey, schemaRepo)

        if domain.IsStaticModule(ver.ModuleKey) {
            // 8 个预定义模块 → static_modules
            resp.StaticModules[ver.ModuleKey] = typedMap
        } else {
            // 其余模块 → dynamic_modules
            dm := BuildDynamicModule(ver, typedMap, schemaRepo, clientScope)
            resp.DynamicModules = append(resp.DynamicModules, dm)
        }
    }
    return resp
}
```

### 判断标准

`IsStaticModule()` 检查 module_key 是否在以下 8 个预定义列表中:

| 序号 | ModuleKey | 说明 |
|---|---|---|
| 1 | base_cfg | 基础配置 |
| 2 | wap_cfg | WAP 配置 |
| 3 | regex_cfg | 正则配置 |
| 4 | pay_cfg | 支付配置 |
| 5 | oss_cfg | OSS 配置 |
| 6 | lang_cfg | 语言配置 |
| 7 | mute_cfg | 免打扰配置 |
| 8 | group_cfg | 群组配置 |

不在列表中的 module_key 自动归入 `dynamic_modules`。

## Schema 驱动机制

`sys_config_schema` 表决定 dynamic_modules 的字段列表和校验规则。

### 工作流程

1. **Admin 创建模块**: 在 sys_config_schema 中注册字段定义
2. **Service 加载 Schema**: `schemaRepo.ListByModule(moduleKey)` 加载字段列表
3. **Compose 组装 Descriptors**: 将 Schema 转换为 `FieldDescriptorView[]`
4. **下发给客户端**: 客户端根据 Descriptors 动态渲染 UI

### Schema 示例

```sql
INSERT INTO sys_config_schema (module_key, field_key, field_type, is_required)
VALUES ('custom_module', 'timeout_ms', 'int', true);
```

生成的 Descriptor:

```json
{
  "field_key": "timeout_ms",
  "field_type": "int",
  "is_required": true,
  "default_val": "3000"
}
```

## API 接口

### ConfigService 接口

```go
type ConfigService interface {
    // 获取应用配置（主查询接口）
    GetAppConfigs(req *AppConfigRequest) (*AppConfigResponse, error)

    // 版本轮询（客户端定期调用检测变更）
    GetVersionInfo(env string, knownVersions map[string]int64) (*VersionInfoResponse, error)
}
```

### AppConfigRequest 参数

```go
type AppConfigRequest struct {
    Env            string              // 环境标识
    ClientScope    string              // 客户端范围
    ClientVersion  string              // 客户端版本号
    RequestedModules []string          // 请求的模块列表（空=全部）
}
```

### AppConfigResponse 响应

```go
type AppConfigResponse struct {
    StaticModules   map[string]map[string]*domain.TypedValue  // 8 个强类型模块
    DynamicModules []*DynamicModuleView                        // 动态模块列表
}
```

## 使用方式

### 通过 LocalInvoker 注册到 Gateway

```go
// 在 Tars Gateway 初始化时注册
invoker.Register("ConfigObj", configService)
```

客户端通过 Proto 6001 调用:

```protobuf
message AppConfigsReq {
    string env = 1;
    string client_scope = 2;
    repeated string requested_modules = 3;
}
```

### 通过 configsdk 引用

```go
import "github.com/jimiechen/mineplanet/go/services/config/sdk"

client, _ := sdk.Default(
    sdk.WithMode(sdk.ModeInProcess),
    sdk.WithService(configService),
    sdk.WithEnv("prod"),
)

value, err := client.GetString(ctx, "base_cfg", "app_name")
```

## 配置项概念说明

| 概念 | 说明 | 示例 |
|---|---|---|
| env | 环境标识，隔离不同环境的配置 | dev/staging/prod |
| clientScope | 客户端可见范围，控制字段粒度 | all/app/miniprogram |
| moduleKey | 模块键名，标识一组相关配置 | base_cfg/wap_cfg |
| version | 版本号，单调递增，用于增量同步 | 1/2/3... |
| fieldType | 字段类型，决定值的解析方式 | string/int/bool/float/json |

## 相关文档

- PRD: [PRD-01-服务商后台系统.md](../prd/PRD-01-服务商后台系统.md)
- ADR: [ADR-0012-polyglot-monorepo-directory-layout.md](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- Proto: [app_config.proto](../../../proto/app_config.proto)
- 协议编号: 6001-6005（见 [协议编号注册表.md](../api/协议编号注册表.md)）
