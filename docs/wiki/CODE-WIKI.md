# CODE-WIKI

## 1. 项目概述

CaiRobot MVP 是一个多系统、多技术栈的微服务架构项目，采用单网关 + MessagePacket + TarsCloud/TarsGo 的统一架构。

核心设计原则：

```text
Protobuf 定义协议身份和业务字段。
MessagePacket 定义单网关入口。
routes.yaml 定义 maxType/minType 到 Tars 目标的映射。
Tars IDL 定义内部服务方法，方法签名统一为 bytes 接口。
TarsCloud 负责服务治理。
TarsGo 服务负责业务逻辑。
```

**项目阶段**：S1（PRD + ADR 落地阶段）

**技术栈**：
- 后端：Go 1.21+（TarsGo v1.4.6 + Gin）
- 前端：TypeScript / ReactJS
- AI 服务：Python 3.11+
- 协议：Protobuf 3 + Tars IDL
- 部署：TarsCloud

---

## 2. 单网关总体架构

```text
Client / App / Web / Third Party / Admin
        ↓
POST /api/hello
        ↓
MessagePacket
        ↓
maxType / minType 路由
        ↓
routes.yaml
        ↓
TarsCloud Routing Layer
        ↓
TarsGo 标准 bytes 接口
        ↓
TarsGo Internal Services
        ↓
业务处理 / DB / Redis / MQ / AI Service / Device Gateway
        ↓
返回 Protobuf Response bytes
        ↓
Gateway 根据 response_proto 解析 response bytes
        ↓
封装 MessagePacket / 统一响应
        ↓
返回客户端
```

---

## 3. 目录结构总览

```text
/workspace/
├── Makefile                          # 根目录编排器（16 个 target）
├── go/
│   ├── go.work                       # Go Workspace 总控
│   ├── Makefile                      # Go 子模块 Makefile
│   ├── common-lib/                   # 公共库（错误码、类型定义）
│   ├── modules/                      # 业务模块（独立 go.mod）
│   │   ├── hello/                    # HelloWorld 模块
│   │   └── health/                   # HealthCheck 模块
│   ├── gateway/
│   │   └── proto-gateway/            # HTTP 网关、路由、调用分发
│   ├── services/                     # 业务服务层（独立 go.mod）
│   │   ├── config/                   # 全局配置领域服务
│   │   │   ├── domain/               # 领域实体（ModuleKey、Schema、Value、Version）
│   │   │   ├── repository/           # 数据访问（SQLite/MySQL 双实现）
│   │   │   ├── cache/                # 缓存抽象（LRU / Redis）
│   │   │   ├── service/              # 业务逻辑（ConfigService 接口 + 实现）
│   │   │   └── sdk/                  # configsdk（三层缓存架构）
│   │   └── i18n/                     # 多语言参数化模板服务
│   │       ├── domain/               # 领域实体（LangPack、LangString、Template）
│   │       ├── repository/           # 数据访问
│   │       ├── cache/                # 缓存抽象
│   │       ├── service/              # 业务逻辑（I18nService 接口 + 实现）
│   │       └── sdk/                  # i18nsdk（三层缓存架构）
│   ├── tars/                         # Tars Servant 层
│   │   ├── system/                   # System Tars Servant（HealthCheck/HelloWorld）
│   │   ├── config/                   # Config Tars Servant（6001/6009 协议）
│   │   ├── i18n/                     # I18n Tars Servant（6003/6005/6007 协议）
│   │   └── provider-admin/           # Admin 管理后台（Gin HTTP）
│   └── third_party/
│       └── TarsGo/                   # TarsCloud/TarsGo v1.4.6 源码
├── proto/
│   ├── base/                         # 基础 Protobuf 协议定义
│   │   ├── message.proto             # MessagePacket 网关入口报文
│   │   ├── result.proto              # 通用返回结果、分页、错误详情
│   │   ├── health.proto              # 健康检查请求/响应
│   │   ├── hello.proto               # HelloWorld 请求/响应
│   │   ├── app_config.proto          # 应用配置协议（6001/6009）
│   │   └── i18n.proto                # 多语言协议（6003/6005/6007）
│   └── generated/                    # 生成的多语言代码
│       ├── go/                       # Go 生成代码
│       ├── python/                   # Python 生成代码
│       └── tarsgo/                   # TarsGo 生成代码
├── python/
│   ├── Makefile                      # Python 子模块 Makefile
│   └── ai/
│       └── hello/                    # AI HelloWorld 服务
├── typescript/
│   ├── Makefile                      # TypeScript 子模块 Makefile
│   ├── web/                          # ReactJS App 前端
│   └── admin-web/                    # Admin 前端（预留）
├── tars/
│   └── protocol/
│       └── tars/                     # Tars IDL 协议定义
│           ├── system.tars
│           ├── auth.tars
│           ├── audit.tars
│           ├── provider_admin.tars
│           ├── user_center.tars
│           ├── open_platform.tars
│           ├── ai_bridge.tars
│           └── device_gateway.tars
├── configs/
│   └── gateway/
│       └── routes.yaml               # 网关路由表
├── scripts/
│   ├── ci/                           # CI 检查脚本
│   ├── proto/                        # Protobuf 生成脚本
│   └── coverage/                     # 覆盖率脚本
└── docs/
    ├── prd/                          # 产品需求文档
    ├── adr/                          # 架构决策记录
    ├── api/                          # API/协议文档
    ├── testing/                      # 测试策略
    └── wiki/                         # 知识库（本文件所在目录）
```

---

## 4. 主要模块职责

### 4.1 Gateway 层（proto-gateway）

**位置**：`go/gateway/proto-gateway/`

**职责**：
- 对外唯一 HTTP 入口（POST /api/hello）
- MessagePacket 序列化/反序列化
- maxType/minType 路由查找
- Tars 调用分发（LocalInvoker / TarsGoInvoker）
- 统一响应封装

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `GatewayServer` | `internal/server/http_server.go` | TarsGo HTTP Servant，处理所有请求 |
| `ServeHTTP` | `internal/server/http_server.go` | HTTP 请求处理主流程 |
| `RouteTable` | `internal/router/route_table.go` | 路由表查找 |
| `FindRoute` | `internal/router/route_table.go` | 根据 maxType:minType 查找路由 |
| `MessagePacket` | `internal/adapter/message_packet.go` | Protobuf 结构体别名 |
| `BuildErrorPacket` | `internal/adapter/message_packet.go` | 构造错误响应 |
| `BuildResponsePacket` | `internal/adapter/message_packet.go` | 构造成功响应 |
| `BuildTarsExtend` | `internal/adapter/message_packet.go` | 构造 Tars 调用 extend map |
| `DeserializeMessagePacket` | `internal/adapter/message_packet.go` | 反序列化 MessagePacket |
| `SerializeMessagePacket` | `internal/adapter/message_packet.go` | 序列化 MessagePacket |

### 4.2 TarsClient 层

**位置**：`go/gateway/proto-gateway/tarsclient/`

**职责**：
- 统一 Tars 调用接口（TarsInvoker）
- LocalInvoker：单体部署模式下的本进程适配器
- TarsGoInvoker：微服务部署模式下的远程调用器（S1 未实现）
- 模块 Handler 注册机制

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `TarsInvoker` | `invoker.go` | 统一调用接口（Invoke 方法） |
| `LocalInvoker` | `invoker.go` | 单体模式本进程调用适配器 |
| `TarsGoInvoker` | `invoker.go` | 微服务模式远程调用器（占位） |
| `Target` | `invoker.go` | Tars 调用目标定义 |
| `TargetKey` | `invoker.go` | 目标唯一键 |
| `LocalHandler` | `invoker.go` | 本地 handler 接口 |
| `ModuleInvokeFunc` | `invoker.go` | 模块服务调用函数签名 |
| `NewModuleHandler` | `invoker.go` | 创建模块适配 handler |
| `RegisterModuleHandlers` | `invoker.go` | 注册 hello/health 模块 handler |
| `RegisterConfigI18nHandlers` | `invoker.go` | 注册 config/i18n 模块 handler |
| `RegisterSystemHandlers` | `invoker.go` | 注册 System 模块 handler（已废弃） |

### 4.3 业务模块层（modules）

**位置**：`go/modules/`

**职责**：
- 独立 go.mod，可单独构建测试
- 只接收 Protobuf bytes，返回 Protobuf bytes
- 不依赖 MessagePacket

**模块清单**：

| 模块 | 位置 | 职责 | 关键类 |
|---|---|---|---|
| hello | `modules/hello/` | HelloWorld 业务逻辑 | `HelloService`（接口）、`Service`（实现）、`SayHello` |
| health | `modules/health/` | HealthCheck 业务逻辑 | `HealthService`（接口）、`Service`（实现）、`Check` |

### 4.4 业务服务层（services）

**位置**：`go/services/`

**职责**：
- 领域驱动设计（DDD）分层：domain / repository / service / sdk
- 支持 SQLite（开发）和 MySQL（生产）双实现
- 提供 SDK 层供其他服务引用

#### 4.4.1 Config 服务

**位置**：`go/services/config/`

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `ConfigService` | `service/interface.go` | 配置服务接口 |
| `AppConfigService` | `service/interface.go` | 默认实现，组合 Repo + Cache + Schema |
| `GetAppConfigs` | `service/interface.go` | 获取全量应用配置 |
| `GetVersionInfo` | `service/interface.go` | 获取版本轮询信息 |
| `ConfigRepository` | `repository/interface.go` | 配置数据访问接口 |
| `SchemaRepository` | `repository/interface.go` | Schema 元数据访问接口 |
| `SQLiteConfigRepo` | `repository/sqlite_repo.go` | SQLite 实现 |
| `MySQLConfigRepo` | `repository/mysql_repo.go` | MySQL 实现 |
| `ModuleKey` | `domain/module_key.go` | 模块键领域实体 |
| `Schema` | `domain/schema.go` | Schema 领域实体 |
| `TypedValue` | `domain/value.go` | 类型化值领域实体 |
| `ConfigVersion` | `domain/version.go` | 版本领域实体 |
| `Client` | `sdk/client.go` | SDK 客户端接口 |
| `configClient` | `sdk/client.go` | SDK 默认实现（L1 LRU + L2 Redis + L3 远程） |
| `Default` | `sdk/client.go` | SDK 工厂方法 |

#### 4.4.2 I18n 服务

**位置**：`go/services/i18n/`

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `I18nService` | `service/interface.go` | 国际化服务接口 |
| `GetLanguages` | `service/interface.go` | 获取支持的语言列表 |
| `GetLangPack` | `service/interface.go` | 获取全量语言包 |
| `GetLangDifference` | `service/interface.go` | 获取增量语言包 |
| `ValidateTemplate` | `service/interface.go` | 校验模板一致性 |
| `LanguageMeta` | `service/interface.go` | 语言元信息结构 |
| `LangPackResponse` | `service/interface.go` | 全量语言包响应 |
| `LangDiffResponse` | `service/interface.go` | 增量语言包响应 |
| `LangStringEntry` | `service/interface.go` | 语言字符串条目 |
| `Client` | `sdk/client.go` | I18n SDK 客户端接口 |
| `clientImpl` | `sdk/client.go` | SDK 默认实现 |
| `T` | `sdk/client.go` | 翻译并渲染参数 |
| `Raw` | `sdk/client.go` | 获取原始模板 |
| `BatchT` | `sdk/client.go` | 批量翻译 |

### 4.5 Tars Servant 层

**位置**：`go/tars/`

**职责**：
- Tars IDL 实现层
- 标准 bytes 接口适配器
- LocalInvoker 注册

**模块清单**：

| 模块 | 位置 | 职责 | 关键入口 |
|---|---|---|---|
| system | `tars/system/` | HealthCheck / HelloWorld | `cmd/main.go`、`adapter/system_adapter.go` |
| config | `tars/config/` | Config Tars Servant（6001/6009） | `cmd/main.go` |
| i18n | `tars/i18n/` | I18n Tars Servant（6003/6005/6007） | `cmd/main.go` |
| provider-admin | `tars/provider-admin/` | Admin 管理后台（Gin HTTP） | `cmd/main.go`、`internal/server/http_server.go` |

### 4.6 Admin 管理后台（provider-admin）

**位置**：`go/tars/provider-admin/`

**职责**：
- Gin HTTP 服务（非 Tars bytes 接口）
- Schema CRUD
- 配置值管理
- 多语言全量 API

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `HTTPServer` | `internal/server/http_server.go` | HTTP 服务器封装 |
| `registerRoutes` | `internal/server/http_server.go` | 注册所有 API 路由 |
| `HealthHandler` | `internal/handler/health_handler.go` | 健康检查处理器 |
| `ConfigHandler` | `internal/handler/config_handler.go` | Schema CRUD 处理器 |
| `ConfigValueHandler` | `internal/handler/config_value_handler.go` | 配置值管理处理器 |
| `I18nHandler` | `internal/handler/i18n_handler.go` | 多语言处理器 |
| `CORS` | `internal/middleware/cors.go` | 跨域中间件 |

### 4.7 公共库（common-lib）

**位置**：`go/common-lib/`

**职责**：
- 统一错误码定义
- 共享类型定义

**关键类/函数**：

| 名称 | 位置 | 职责 |
|---|---|---|
| `CodeSuccess` | `codes.go` | 成功码 10200 |
| `CodeBadRequest` | `codes.go` | 请求参数错误 10400 |
| `CodeUnauthorized` | `codes.go` | 未授权 10401 |
| `CodeNotFound` | `codes.go` | 资源未找到 10404 |
| `CodeInternalError` | `codes.go` | 内部错误 10500 |
| `CodeTarsNotImplemented` | `codes.go` | Tars 远程调用未实现 10501 |
| `ModuleInvokeFunc` | `types.go` | 模块服务调用函数签名 |
| `ModuleHandler` | `types.go` | 模块处理器接口 |

---

## 5. 关键协议定义

### 5.1 MessagePacket（网关入口报文）

**文件**：`proto/base/message.proto`

```protobuf
message MessagePacket {
  int32 maxType = 1;              // 协议大类
  int32 minType = 2;              // 协议小类
  map<string, string> extend = 3; // 通用透传上下文
  Platform platform = 4;          // 平台类型
  bytes data = 5;                 // 业务协议包
}
```

**Platform 枚举**：
- `UNKNOWN = 0`：未知平台
- `WEB = 1`：网页浏览器
- `PC = 2`：桌面客户端
- `ANDROID = 3`：安卓移动设备
- `IOS = 4`：苹果移动设备
- `OTHER = 5`：其他平台

### 5.2 Result（通用返回结果）

**文件**：`proto/base/result.proto`

```protobuf
message Result {
  int32 code = 1;     // 响应码，成功建议 10200
  string message = 2; // 人类可读消息
}

message PageInfo {
  int32 pageSize = 1;
  int32 pageToken = 2;
  int32 totalCount = 3;
  string nextPageToken = 4;
}

message ErrorDetail {
  string field = 1;
  string message = 2;
  string code = 3;
}

message ValidationResult {
  repeated ErrorDetail errors = 1;
}
```

### 5.3 HealthCheck 协议（2100 段）

**文件**：`proto/base/health.proto`

| 方向 | max | min | Message |
|---|---|---|---|
| C->S | 2100 | 2097 | `ServiceHealthCheckRequest` |
| S->C | 2100 | 2098 | `ServiceHealthCheckResponse` |

```protobuf
message ServiceHealthCheckRequest {
  string service_name = 1;
}

message ServiceHealthCheckResponse {
  com.mineplanet.pojo.pb.Result result = 1;
  string status = 2;      // "OK" 或 "Unhealthy"
  int64 timestamp = 3;    // Unix 秒
}
```

### 5.4 HelloWorld 协议（2100 段）

**文件**：`proto/base/hello.proto`

| 方向 | max | min | Message |
|---|---|---|---|
| C->S | 2100 | 2101 | `HelloWorldRequest` |
| S->C | 2100 | 2102 | `HelloWorldResponse` |

```protobuf
message HelloWorldRequest {
  string name = 1;
}

message HelloWorldResponse {
  com.mineplanet.pojo.pb.Result result = 1;
  string message = 2;
  int64 timestamp = 3;
}
```

### 5.5 应用配置协议（6000 段）

**文件**：`proto/base/app_config.proto`

| 方向 | max | min | Message | 说明 |
|---|---|---|---|---|
| C->S | 6000 | 6001 | `AppConfigsReq` | 获取全量应用配置 |
| S->C | 6000 | 6002 | `AppConfigsRsp` | 应用配置响应 |
| C->S | 6000 | 6009 | `AppConfigVersionReq` | 版本轮询请求 |
| S->C | 6000 | 6010 | `AppConfigVersionRsp` | 版本轮询响应 |

**关键消息**：

```protobuf
message AppConfigsReq {
  string env = 1;
  string client_scope = 2;
  string client_version = 3;
  repeated string requested_modules = 4;
}

message AppConfigsRsp {
  com.mineplanet.pojo.pb.Result result = 1;
  AppBaseConfigs base_cfg = 2;
  AppWapUrlConfigs wap_cfg = 3;
  AppRegexConfigs regex_cfg = 4;
  AppPayConfigs pay_cfg = 5;
  AppOssConfigs oss_cfg = 6;
  AppLanguageConfigs lang_cfg = 7;
  AppMuteConfigs mute_cfg = 8;
  AppGroupConfigs group_cfg = 9;
  repeated DynamicConfigModule dynamic_modules = 100;
}

message DynamicConfigModule {
  string module_key = 1;
  int64 version = 2;
  map<string, string> fields = 3;
  repeated FieldDescriptor descriptors = 4;
}

message FieldDescriptor {
  string field_key = 1;
  string field_type = 2;   // string/int/bool/float/enum/json/list
  bool is_required = 3;
  string default_val = 4;
}
```

### 5.6 多语言协议（6000 段）

**文件**：`proto/base/i18n.proto`

| 方向 | max | min | Message | 说明 |
|---|---|---|---|---|
| C->S | 6000 | 6003 | `AppFetchLanguageReq` | 获取语言元数据 |
| S->C | 6000 | 6004 | `AppFetchLanguageRsp` | 语言元数据响应 |
| C->S | 6000 | 6005 | `AppFetchLangPackReq` | 获取全量语言包 |
| S->C | 6000 | 6006 | `AppFetchLangPackRsp` | 全量语言包响应 |
| C->S | 6000 | 6007 | `AppFetchLangDifferenceReq` | 获取增量语言包 |
| S->C | 6000 | 6008 | `AppFetchLangDifferenceRsp` | 增量语言包响应 |

**关键消息**：

```protobuf
message LangStringEntry {
  string key = 1;
  string value = 2;
  string template_type = 3;   // plain / named / icu
  repeated LangParam params = 4;
  string operation_type = 5;  // ADD / MOD / DEL
}

message LangParam {
  string name = 1;
  string type = 2;            // string/int/float/date
  bool required = 3;
  string default_v = 4;
}
```

### 5.7 Tars IDL 标准接口

**文件**：`tars/protocol/tars/*.tars`

所有 Tars 接口统一使用 bytes 签名：

```tars
interface XxxObj {
    int Health(vector<byte> request, map<string,string> extend, out vector<byte> response);
    int HealthCheck(vector<byte> request, map<string,string> extend, out vector<byte> response);
    int XxxMethod(vector<byte> request, map<string,string> extend, out vector<byte> response);
};
```

### 5.8 协议编号范围

| max 范围 | 模块 |
|---:|---|
| 1000-1999 | 通用基础协议 |
| 2000-2999 | 系统、健康检查、网关基础能力 |
| 3000-3999 | 认证与权限 |
| 4000-4999 | 服务商后台系统 |
| 5000-5999 | 终端用户中台系统 |
| 6000-6999 | App、Web、前端交互协议（配置/i18n/业务） |
| 7000-7999 | AI 服务系统 |
| 8000-8999 | 设备通信与设备网关 |
| 9000-9999 | App、Web、前端交互协议 |

---

## 6. 依赖关系

### 6.1 模块依赖图

```text
Client/App/Web
    ↓ POST /api/hello (application/octet-stream)
GatewayServer (proto-gateway)
    ↓ MessagePacket bytes
adapter.DeserializeMessagePacket
    ↓ maxType/minType
RouteTable.FindRoute
    ↓ routes.yaml
TarsInvoker.Invoke
    ├── LocalInvoker（单体模式）→ 进程内 handler 调用
    └── TarsGoInvoker（微服务模式）→ 远程 TarsCloud 调用（S1 未实现）
        ↓
    ModuleHandler / LocalHandler
        ↓
    modules/hello.Service.SayHello
    modules/health.Service.Check
    services/config.AppConfigService.GetAppConfigs
    services/i18n.I18nService.GetLangPack
        ↓
    Protobuf Response bytes
        ↓
    Gateway 封装 MessagePacket 返回
```

### 6.2 Go Module 依赖关系

```text
go.work
├── proto/generated/go          # Protobuf 生成代码（被所有模块依赖）
├── common-lib                  # 错误码、类型定义（被所有模块依赖）
├── modules/hello               # 依赖 common-lib, proto/generated/go
├── modules/health              # 依赖 common-lib, proto/generated/go
├── gateway/proto-gateway       # 依赖 common-lib, proto/generated/go, modules/*, services/*
├── services/config             # 依赖 common-lib, proto/generated/go
├── services/i18n               # 依赖 common-lib, proto/generated/go
├── tars/system                 # 依赖 common-lib, proto/generated/go, modules/*
├── tars/config                 # 依赖 common-lib, proto/generated/go, services/config
├── tars/i18n                   # 依赖 common-lib, proto/generated/go, services/i18n
└── tars/provider-admin         # 依赖 common-lib, services/config, services/i18n
```

### 6.3 外部依赖

| 依赖 | 版本 | 用途 |
|---|---|---|
| TarsGo | v1.4.6 | TarsCloud 服务框架 |
| Protobuf | v3 | 协议定义 |
| Gin | latest | Admin 管理后台 HTTP 框架 |
| google/uuid | latest | traceId/requestId 生成 |

---

## 7. 项目运行方式

### 7.1 环境准备

```bash
# 初始化开发环境（检查工具链 + 安装依赖）
make bootstrap

# 需要的工具
- Go 1.21+
- Python 3.11+
- Node.js 20+
- protoc（Protobuf 编译器）
- make
```

### 7.2 常用命令

```bash
# 显示帮助
make help

# 生成 Protobuf 代码（Go/TS/Python/TarsGo）
make proto

# 校验 Protobuf 生成代码（CI 用，不需要 protoc）
make proto-check

# 运行所有测试（单元 + 集成）
make test

# 运行单元测试
make unit

# 运行集成测试
make integration

# 运行 Lint 检查
make lint

# 生成覆盖率报告
make coverage

# 构建可执行文件
make build

# 执行完整 CI 检查
make ci

# 执行工程规范检查
make rules

# 清理构建产物
make clean
```

### 7.3 Gateway 运行

```bash
# 编译 proto-gateway
gateway-build
# 或：make -C go gateway-build

# 启动 proto-gateway（local 模式，默认端口 8080）
gateway-start
# 或：make -C go gateway-start

# 停止 proto-gateway
gateway-stop
# 或：make -C go gateway-stop

# 运行 Gateway 测试
gateway-test
# 或：make -C go gateway-test

# 冒烟测试（编译 + 启动 + 验证 + 停止）
gateway-smoke
# 或：make -C go gateway-smoke

# 完整验证（编译 + 单元测试 + TarsGo 依赖检查）
gateway-verify
# 或：make -C go gateway-verify
```

**单体部署模式（默认）**：

```bash
GATEWAY_INVOKER_MODE=local
```

- 不连接远程 TarsCloud 注册中心
- 通过 LocalInvoker 进程内调用
- 所有 servant 在同一部署单元

**微服务部署模式**：

```bash
GATEWAY_INVOKER_MODE=tars
```

- 连接远程 TarsCloud 注册中心
- 通过 TarsGoInvoker 远程调用
- **S1 阶段未实现**

### 7.4 Go 模块测试

```bash
# 运行 Go 全量测试（common-lib + modules + tars + gateway + E2E）
make go-all
# 或：make -C go go-all

# 测试 common-lib
make common-lib-test
# 或：make -C go common-lib-test

# 测试业务模块（hello + health）
make modules-test
# 或：make -C go modules-test

# 测试 Tars 调用层
make tars-test
# 或：make -C go tars-test

# 运行 Gateway E2E 链路测试
make gateway-e2e
# 或：make -C go gateway-e2e
```

### 7.5 Config/I18n 服务运行

```bash
# 运行 Config Tars Servant
cd go/tars/config && go run cmd/main.go

# 运行 I18n Tars Servant
cd go/tars/i18n && go run cmd/main.go

# 运行 Admin 管理后台
cd go/tars/provider-admin && go run cmd/main.go
```

### 7.6 前端运行

```bash
# Web 前端（TypeScript/ReactJS）
cd typescript/web && pnpm install && pnpm dev

# Admin 前端
cd typescript/admin-web && pnpm install && pnpm dev
```

---

## 8. extend map 上下文透传规范

标准字段：

| key | 说明 | 来源 |
|---|---|---|
| traceId | 链路追踪 ID | 客户端或 Gateway |
| requestId | 请求唯一 ID | 客户端或 Gateway |
| token | Token | MessagePacket.extend 或 Header |
| caller | 调用方 | Gateway |
| clientIp | 客户端 IP | Gateway |
| userId | 用户 ID | AuthServer 校验结果 |
| tenantId | 租户 / 服务商 ID | AuthServer 校验结果 |
| clientVersion | 客户端版本号 | 客户端或 Gateway |
| maxType | 协议大类 | Gateway 自动填充 |
| minType | 协议小类 | Gateway 自动填充 |
| platform | 平台类型 | Gateway 自动填充 |
| routeKey | 路由键 | Gateway 自动填充 |
| requestProto | 请求 Proto 类型 | Gateway 自动填充 |
| responseProto | 响应 Proto 类型 | Gateway 自动填充 |
| authRequired | 是否需要鉴权 | Gateway 自动填充 |
| auditRequired | 是否需要审计 | Gateway 自动填充 |

---

## 9. Tars 命名规范

| 层级 | 命名规则 | 示例 |
|---|---|---|
| App | CaiRobot | CaiRobot |
| Server | XxxServer | UserCenterServer |
| Servant | XxxObj | UserCenterObj |
| module | CaiRobotXxxApp | CaiRobotUserCenterApp |
| interface | XxxObj | UserCenterObj |

---

## 10. 错误码规范

| 错误码 | 含义 | 使用场景 |
|---|---|---|
| 10200 | 成功 | 操作成功 |
| 10400 | 请求参数错误 | 参数校验失败 |
| 10401 | 未授权 | Token 无效或过期 |
| 10404 | 资源未找到 | 路由不存在、handler 未注册 |
| 10500 | 内部错误 | 服务内部异常 |
| 10501 | Tars 远程调用未实现 | S1 阶段占位 |

---

## 11. 相关文档索引

### ADR（架构决策）
- [ADR-0001-总体系统架构](../adr/ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf](../adr/ADR-0003-服务协议使用Protobuf.md)
- [ADR-0008-use-tarscloud-routing-layer](../adr/ADR-0008-use-tarscloud-routing-layer.md)
- [ADR-0012-polyglot-monorepo-directory-layout](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- [ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement](../adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md)
- [ADR-0014-message-packet-data-format-protobuf-bytes](../adr/ADR-0014-message-packet-data-format-protobuf-bytes.md)
- [ADR-009-config-i18n-schema-template](../adr/ADR-009-config-i18n-schema-template.md)
- [ADR-010-admin-boundary-sdk](../adr/ADR-010-admin-boundary-sdk.md)

### API 规范
- [protobuf规范](../api/protobuf规范.md)
- [tars规范](../api/tars规范.md)
- [http-gateway规范](../api/http-gateway规范.md)
- [openapi-protobuf映射规范](../api/openapi-protobuf映射规范.md)
- [OpenAPI规范](../api/OpenAPI规范.md)
- [协议编号注册表](../api/协议编号注册表.md)

### 工程规范
- [编码规范](../../.trae/rules/coding.md)
- [中文注释规范](../../.trae/rules/commenting.md)
- [TDD 规范](../../.trae/rules/tdd.md)
- [测试规范](../../.trae/rules/testing.md)
- [Git 规范](../../.trae/rules/git.md)
- [Makefile 规范](../../.trae/rules/makefile.md)

---

## 12. 变更日志

| 日期 | 变更内容 |
|---|---|
| 2026-05-24 | **更新 CODE-WIKI**：补充 services/config 和 services/i18n 完整模块说明、SDK 三层缓存架构、Admin 管理后台、协议定义详解、extend map 规范、错误码规范 |
| 2026-05-22 | **S1 阶段 Config/I18n 基础设施**：新增 services/config（全局配置领域服务）和 services/i18n（多语言参数化模板服务）业务服务层；新增 tars/config 和 tars/i18n Tars Servant；新增 tars/provider-admin Admin 管理后台（Gin HTTP）；SDK 层（configsdk / i18nsdk）三层缓存架构；6000 段协议编号范围重新定义 |
| 2026-05-21 | **Go 多模块重构**：新增 common-lib、modules/hello、modules/health 独立 go.mod；实现 LocalInvoker 模块注册机制；TS E2E 集成；全量测试 89/89 PASS |
| 2026-05-20 | **架构口径修正**：LocalInvoker 是本进程 TarsGo servant adapter；proto-gateway 改造为基于 TarsGo HTTP 模块 |
| 2026-05-19 | **恢复根目录 Makefile**：采用三层结构（总控 + 子 Makefile + scripts）；新增 16 个 target |
| 2026-05-18 | **内部核心服务主链路从 gRPC 调整为 TarsCloud/TarsGo** |
