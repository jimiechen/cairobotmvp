# CODE-WIKI

## 1. 项目概述

CaiRobot MVP 是一个**运营管理综合平台**，采用多系统、多技术栈的微服务架构，核心特征为：

- **单网关入口**：所有外部请求统一通过 `POST /api/hello` 进入
- **Protobuf 通讯**：协议身份与业务字段均由 Protobuf 定义
- **TarsCloud/TarsGo 服务治理**：内部服务通过 Tars 框架进行路由与调用
- **多语言 Monorepo**：Go（后端）+ TypeScript（前端）+ Python（AI）统一仓库管理

核心设计原则：

```text
Protobuf 定义协议身份和业务字段。
MessagePacket 定义单网关入口。
routes.yaml 定义 maxType/minType 到 Tars 目标的映射。
Tars IDL 定义内部服务方法，方法签名统一为 bytes 接口。
TarsCloud 负责服务治理。
TarsGo 服务负责业务逻辑。
```

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

## 3. MessagePacket 入口协议

MessagePacket 是对外单网关唯一业务入口报文。

```proto
message MessagePacket {
  int32 maxType = 1;             // 必填，协议大类
  int32 minType = 2;             // 必填，协议小类
  map<string, string> extend = 3; // 非必填，通用透传上下文
  Platform platform = 4;         // 非必填，平台类型
  bytes data = 5;                // 必填，业务协议包
}
```

- `maxType/minType` 来自业务 Request message 的 `Type.max/Type.min`
- `data` 是业务 Request message 序列化后的 bytes
- 响应也使用 MessagePacket，其中 `maxType/minType` 来自 Response message 的 `Type.max/Type.min`

---

## 4. Protobuf 协议编号规范

- 协议编号 `max + min` 是接口报文的唯一身份
- 每个业务 Request/Response message 内部必须声明 enum Type
- `Type.max` 表示协议大类，`Type.min` 表示协议小类
- `max + min` 组合在全仓库内必须唯一
- Request 和 Response 应分别拥有独立编号

编号范围：

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

## 5. routes.yaml 路由设计

routes.yaml 以 `request_max/request_min` 为主路由键：

```yaml
routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    request_proto: com.mineplanet.pojo.health.ServiceHealthCheckRequest
    response_max: 2100
    response_min: 2098
    response_proto: com.mineplanet.pojo.health.ServiceHealthCheckResponse
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: HealthCheck
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
```

### 5.1 配置与多语言协议路由示例（6000 段）

```yaml
# 6001/6002: 获取全量应用配置（App 启动时调用）
- request_max: 6000
  request_min: 6001
  route_key: "6000:6001"
  command_name: GetAppConfigs
  tars_app: CaiRobot
  tars_server: ConfigServer
  tars_servant: ConfigObj
  tars_method: GetAppConfigs

# 6009/6010: 配置与语言包版本轮询
- request_max: 6000
  request_min: 6009
  route_key: "6000:6009"
  command_name: AppConfigVersion
  tars_server: ConfigServer
  tars_servant: ConfigObj
  tars_method: AppConfigVersion

# 6003/6004: 语言元数据
- request_max: 6000
  request_min: 6003
  tars_server: I18nServer
  tars_servant: I18nObj
  tars_method: GetAppLanguage

# 6005/6006: 全量语言包
- request_max: 6000
  request_min: 6005
  tars_server: I18nServer
  tars_servant: I18nObj
  tars_method: GetLangPack

# 6007/6008: 增量语言包
- request_max: 6000
  request_min: 6007
  tars_server: I18nServer
  tars_servant: I18nObj
  tars_method: GetLangDifference
```

---

## 6. TarsCloud/TarsGo 内部服务治理

- TarsCloud App 统一为 `CaiRobot`
- Server 使用 `XxxServer`
- Servant 使用 `XxxObj`
- IDL module 使用 `CaiRobotXxxApp`
- interface 使用 `XxxObj`
- 每个 servant 都暴露 Health / HealthCheck
- Tars 方法统一 bytes 签名

---

## 7. Tars 标准 bytes 接口

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

- `request`：Protobuf Request message 序列化 bytes
- `extend`：上下文透传 map
- `response`：Protobuf Response message 序列化 bytes
- `return int`：项目统一状态码

---

## 8. Tars App / Server / Servant / module / interface 命名规范

| 层级 | 命名规则 | 示例 |
|---|---|---|
| App | CaiRobot | CaiRobot |
| Server | XxxServer | UserCenterServer |
| Servant | XxxObj | UserCenterObj |
| module | CaiRobotXxxApp | CaiRobotUserCenterApp |
| interface | XxxObj | UserCenterObj |

---

## 9. extend map 上下文透传规范

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
| clientVersion | 客户端版本号 | 客户端或 Gateway（用于模板兼容性过滤）|

### 9.5 模块接入规范

> **2026-05-26 新增**：Hello / Health 模块已升级为 SDK + Schema 驱动的参考实现。

所有业务模块必须遵循统一接入规范，详见 [sample-module.md](./modules/sample-module.md)。

**核心要点**：
- 使用 `common-lib/module.Deps` 统一依赖装配（内联接口解耦）
- 通过 configsdk / i18nsdk 读取配置和渲染文案
- 实现 health.Checker 接口注册依赖健康检查（含真实 mysqlx.Ping / redisx.Ping）
- 提供 Seed 脚本注入 Schema 和 i18n 数据
- 单元测试使用 Fake SDK，覆盖率 ≥80%
- 语言解析通过 `i18n.ResolveLang()` 4 级优先级统一管理
- 错误信息通过 `i18n.TruncateError()` UTF-8 安全截断（≤512 字符）

**强制合规检查（`make module-lint`）**：
- 10+1 项自动检查（L1-L10 + SDK_USAGE 清单），任一失败 = PR 不予合入
- CI 已集成 module-lint job（required）
- 检查脚本：[module_lint.sh](../../scripts/lint/module_lint.sh)

**参考实现**：
- [Hello 模块（configsdk 范例）](./modules/hello/README.md) — 覆盖率 82.9%
- [Health 模块（i18nsdk ICU + Checker 范例）](./modules/health/README.md) — 覆盖率 77.8%

---

## 10. Tars return 与 Result.code 分层规则

- Tars return 表示内部 Tars 方法调用的处理状态
- Protobuf Response message 内部的 Result.code 表示业务响应状态
- Gateway 对外返回 MessagePacket 时，优先使用业务 Response.Result.code
- Tars 框架异常由 Gateway 映射为项目统一错误码

---

## 11. Gateway 调用 Tars 完整流程

1. 客户端请求 `POST /api/hello`
2. 请求体是 `MessagePacket` bytes
3. Gateway 反序列化 MessagePacket
4. Gateway 校验 maxType/minType/data
5. Gateway 使用 `maxType:minType` 查询 routes.yaml
6. Gateway 校验协议编号是否已登记
7. Gateway 根据 `request_proto` 反序列化 data
8. Gateway 构造 Tars extend 并调用 TarsGo servant
9. TarsGo 服务处理业务，返回 response bytes 和 return code
10. Gateway 根据 `response_proto` 反序列化 response bytes
11. Gateway 封装响应 MessagePacket 返回客户端

---

## 12. Health / HealthCheck 基础接口

每个 Tars interface 必须包含：

```tars
int Health(vector<byte> request, map<string,string> extend, out vector<byte> response);
int HealthCheck(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

---

## 13. 目录结构说明

### 13.1 文档目录

```text
docs/
  api/
    protobuf规范.md
    tars规范.md
    http-gateway规范.md
    openapi-protobuf映射规范.md
    OpenAPI规范.md
    协议编号注册表.md
  adr/
    ADR-0001-总体系统架构.md
    ADR-0003-服务协议使用Protobuf.md
    ADR-0008-use-tarscloud-routing-layer.md
  wiki/
    CODE-WIKI.md
    LLM-WIKI.md
```

### 13.2 Go 语言资产

```text
go/
  go.work                                    # Workspace 总控
  common-lib/                                # 公共库（错误码、类型定义）
    codes.go
    codes_test.go
    types.go
  modules/                                   # 业务模块（独立 go.mod）
    hello/                                    # Hello 模块
      service.go
      service_test.go
      go.mod
    health/                                   # Health 模块
      service.go
      service_test.go
      go.mod
    (users/auth/groups/topics/readonly 预留)
  gateway/
    proto-gateway/
      README.md
      go.mod                          # 引入 github.com/TarsCloud/TarsGo v1.4.6
      cmd/
        server/
          main.go                         # TarsGo 入口：TarsHttpMux + AddHttpServant + Run
        testclient/
          main.go                         # E2E 测试客户端（参考实现）
      configs/
        gateway/
          gateway.local.conf               # TarsGo 单体部署本地配置（locator 为空）
      internal/
        config/
          routes.go
          routes_test.go
        server/
          http_server.go
          http_server_test.go
          e2e_modules_test.go             # E2E 全链路测试（Gateway → Modules）
        tarsclient/
          invoker.go                     # TarsInvoker 接口 + LocalInvoker + TarsGoInvoker + ModuleHandler
          invoker_test.go
          module_handler_test.go         # 模块注册测试
        adapter/
          message_packet.go
          message_packet_test.go
  tars/
    system/
      go.mod
      cmd/
        main.go
      internal/
        service/
          system_service.go              # @deprecated 标记废弃，保留兼容
          system_service_test.go
      adapter/                            # Adapter 层（替代旧 localhandler）
        system_adapter.go                # LocalHandler 接口适配器
        system_adapter_test.go
        deprecated/                       # 废弃代码归档
          local_handler.go               # 旧 LocalHandler 实现
          local_handler_test.go
    auth/                                 # 预留
    audit/                                # 预留
    ...
  services/                                   # 业务服务层（独立 go.mod）
    config/                                    # 全局配置服务
      domain/                                  # 领域实体
      repository/                              # 数据访问（SQLite/MySQL 双实现）
      cache/                                   # 缓存抽象
      service/                                 # 业务逻辑
      sdk/                                     # configsdk（阶段 B.5 新增）
    i18n/                                      # 多语言服务
      domain/ repository/ cache/ service/ sdk/ # 同上结构
  tars/
    config/                                    # Config Tars Servant
      cmd/main.go                              # 服务入口 + LocalInvoker 注册
      adapter/config_adapter.go                # bytes → service 适配器
      e2e_test.go                              # E2E 集成测试（3 场景）
    i18n/                                      # I18n Tars Servant
      cmd/main.go adapter/i18n_adapter.go e2e_test.go # 同上
    admin/                                     # Admin 管理后台（go-admin v2.2.0）
      main.go                                   # go-admin 引擎入口
      config/settings.yml                        # 配置文件
      app/admin/apis/                            # API 处理器（我们写的插件）
          config_handler.go                     # Schema CRUD
          config_value_handler.go               # 配置值管理
          i18n_handler.go                       # 多语言全量 API
          health_handler.go                    # 健康检查
          admin_i18n_repo.go                   # 扩展写操作仓库
        middleware/cors.go                      # CORS 中间件
  shared/
    audit/
    config/
    result/
    protoadapter/
  third_party/
    TarsGo/
      README.md                           # TarsGo v1.4.6 依赖基线说明
      TarsGo-1.4.6/                       # TarsCloud/TarsGo v1.4.6 源码（replace 指向）
  gateway/proto-gateway/
    tarsclient/                                # 从 internal 移出的公共 Tars 客户端包
      invoker.go                               # LocalInvoker + RegisterConfigI18nHandlers
```

**架构分层**：

| 层级 | 目录 | 职责 |
|------|------|------|
| 公共库 | `common-lib/` | 错误码、类型定义、常量 |
| 业务模块 | `modules/*` | 独立 go.mod，可单独构建测试 |
| Gateway | `gateway/proto-gateway/` | HTTP 入口、路由、调用分发 |
| Tars 适配层 | `tars/system/adapter/` | LocalHandler → ModuleInvokeFunc 适配 |
| Tars 服务 | `tars/system/internal/service/` | 具体业务逻辑（标记 @deprecated） |

### 13.3 Python 语言资产

```text
python/
  ai/
    service/
    README.md
  tools/
    README.md
```

### 13.4 TypeScript 语言资产

```text
typescript/
  web/
    src/                                     # 业务源码（纯实现，无测试文件）
      pages/
        hello/
          HelloPage.tsx                      # React 页面组件
      utils/                                 # 工具函数库（可复用）
        proto-client.ts                      # Proto-Gateway 客户端（buildPacket/postGateway）
    tests/                                   # 测试目录（与源码隔离）
      e2e/                                   # E2E 集成测试
        gateway-modules.test.ts              # Gateway → Modules 全链路测试
      unit/                                  # 单元测试
        hello/
          HelloPage.test.tsx                 # HelloPage 组件单元测试
    package.json                             # 依赖配置（google-protobuf + 路径别名）
    pnpm-lock.yaml
    vite.config.ts                            # Vite 配置（@proto/@utils/@pages 别名）
    tsconfig.json                             # TS 编译配置（paths 映射）
    tsconfig.node.json                        # Node.js 类型声明
  admin-web/                                 # 预留
  app-h5/                                    # 预留
  packages/                                  # 预留
  README.md
```

**TS 测试分层规范**：

| 目录 | 类型 | 特征 |
|------|------|------|
| `tests/unit/` | 单元测试 | Mock 外部依赖，快速执行，不依赖真实服务 |
| `tests/e2e/` | E2E 集成测试 | 真实 HTTP 请求，验证完整链路，支持优雅降级 |

**路径别名**：

| 别名 | 指向 | 用途 |
|------|------|------|
| `@proto/*` | `../../proto/generated/ts/*` | Protobuf 生成类型 |
| `@utils/*` | `src/utils/*` | 工具函数库 |
| `@pages/*` | `src/pages/*` | 页面组件 |

### 13.5 Tars 协议目录

```text
tars/
  protocol/
    tars/
      system.tars
      auth.tars
      provider_admin.tars
      user_center.tars
      open_platform.tars
      ai_bridge.tars
      device_gateway.tars
      audit.tars
```

### 13.6 部署目录

```text
deploy/
  tarscloud/
    README.md
    configs/
    templates/
```

---

## 14. Go Workspace 管理

Go Workspace 位于 `go/go.work`，管理所有 Go 子模块：

```text
go/go.work
├── common-lib                  (github.com/jimiechen/mineplanet/go/common-lib)
├── modules/hello               (github.com/jimiechen/mineplanet/go/modules/hello)
├── modules/health              (github.com/jimiechen/mineplanet/go/modules/health)
├── gateway/proto-gateway        (github.com/jimiechen/mineplanet/go/gateway/proto-gateway)
└── tars/system                 (github.com/jimiechen/mineplanet/go/tars/system)
```

**模块化架构原则**：

| 原则 | 说明 |
|------|------|
| 独立 go.mod | 每个业务模块独立版本管理 |
| replace 路径 | 通过相对路径引用 common-lib |
| Workspace 统一 | go.work 统一构建和测试入口 |
| 单一职责 | 每个模块只承担一类明确职责 |

**当前模块清单**：

| 模块 | Path | 职责 | 测试数 |
|------|------|------|--------|
| common-lib | `go/common-lib` | 错误码、类型定义 | 7 |
| modules/hello | `go/modules/hello` | Hello 业务逻辑 | 3 |
| modules/health | `go/modules/health` | Health 业务逻辑 | 4 |
| gateway/proto-gateway | `go/gateway/proto-gateway` | HTTP 网关、路由、调用分发 | 16 |
| tars/system | `go/tars/system` | Tars 适配层、服务骨架 | 11 |
| services/config | `go/services/config` | 全局配置领域服务 | ~30 |
| services/i18n | `go/services/i18n` | 多语言参数化模板服务 | ~25 |
| admin | `go/admin` | Admin 管理后台（go-admin v2.2.0） | ~50 (框架+插件) |
| tars/config | `go/tars/config` | Config Tars Servant | ~3 |
| tars/i18n | `go/tars/i18n` | I18n Tars Servant | ~3 |

**执行命令前需 `cd go/`**

---

## 15. 依赖关系

- Protobuf 定义协议身份和业务字段
- MessagePacket 定义单网关入口
- routes.yaml 定义 maxType/minType 到 Tars 目标的映射
- Tars IDL 定义内部服务方法
- TarsCloud 负责服务治理
- TarsGo 服务负责业务逻辑

---

## 16. Gateway 运行模式

Gateway 基于 TarsCloud/TarsGo v1.4.6 技术基线，支持两种**部署拓扑**：

### 16.1 单体部署模式（默认）

```bash
GATEWAY_INVOKER_MODE=local
```

- 本地开发、测试、演示使用
- **使用 TarsGo 框架运行**（TarsHttpMux / AddHttpServant / Run）
- **不连接远程 TarsCloud 注册中心**（locator 为空），但不是不依赖 TarsGo
- 通过 **LocalInvoker**（本进程 TarsGo servant adapter）调用同部署单元内的业务 servant
- 所有 TarsGo servant 在同一进程或同一部署单元中
- 仍然严格走 routes.yaml
- 严格遵守 Tars bytes 契约：request/response 均为 Protobuf bytes

### 16.2 微服务部署模式

```bash
GATEWAY_INVOKER_MODE=tars
```

- 正式部署或集成环境使用
- **使用 TarsGo 框架运行**（与单体模式相同的技术基线）
- 连接远程 TarsCloud 注册中心，通过 **TarsGoInvoker**（远程 TarsGo client）调用独立部署的 TarsCloud servant
- GatewayServer、SystemServer 等独立部署为不同进程
- 当前 **TarsGoInvoker 远程调用尚未实现**，启动会报错（S1 阶段）

### 16.3 TarsInvoker 接口

统一调用接口：

```go
type TarsInvoker interface {
    Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}
```

实现：
- **LocalInvoker**：单体部署模式下的本进程 TarsGo servant adapter。不绕过 Tars 框架，而是在同部署单元内通过进程内调用转发到 TarsGo servant。严格遵守 Tars bytes 契约。
- **TarsGoInvoker**：微服务部署模式下的远程 TarsGo client invoker。通过 TarsGo client 远程调用独立部署的 TarsCloud servant。与 LocalInvoker 共享同一接口和 Tars bytes 契约。S1 未实现。

---

## 17. 技术栈与架构依赖

### 17.1 后端技术栈（Go）

| 类别 | 技术/框架 | 版本 | 说明 |
|---|---|---|---|
| 语言 | Go | 1.25.5 | 主开发语言 |
| RPC 框架 | TarsCloud/TarsGo | v1.4.6 | 服务治理与内部通讯 |
| HTTP 框架 | Gin | v1.10.0 | Admin 后台 HTTP 服务 |
| 协议序列化 | Protobuf | v1.36.11 | 统一协议格式 |
| ORM | GORM | v1.25.12 | 数据库访问 |
| 数据库驱动 | MySQL / SQLite / Postgres | - | 多数据库支持 |
| 缓存 | Redis | - | 缓存与消息发布订阅 |
| 管理后台框架 | go-admin | v2.2.0 | 快速搭建 Admin 后台 |
| 日志 | TarsGo/contrib/log | - | 统一日志接口 |
| 配置管理 | Viper | - | 配置文件读取 |
| 测试 | Go testing + testify | - | 单元与集成测试 |

### 17.2 前端技术栈（TypeScript）

| 类别 | 技术/框架 | 版本 | 说明 |
|---|---|---|---|
| 语言 | TypeScript | - | 主开发语言 |
| UI 库 | React | v18.2.0 | 页面组件构建 |
| 构建工具 | Vite | v5.0.0 | 前端构建与 HMR |
| 测试框架 | Vitest | v1.0.0 | 单元测试 |
| Protobuf 运行时 | google-protobuf | v3.21.2 | Protobuf 序列化/反序列化 |
| 管理后台 UI | Vue 2 + Element UI | - | admin-web 页面 |
| E2E 测试 | Playwright | - | 端到端自动化测试 |

### 17.3 AI 服务技术栈（Python）

| 类别 | 技术/框架 | 版本 | 说明 |
|---|---|---|---|
| 语言 | Python | 3.11 | AI 服务开发 |
| Web 框架 | FastAPI | 0.104.1 | API 服务 |
| 服务运行 | Uvicorn | 0.24.0 | ASGI 服务器 |
| 测试 | pytest | 7.4.3 | 单元测试 |

### 17.4 关键依赖版本说明

- **TarsGo v1.4.6**：内部服务统一使用 TarsGo 框架，源码通过 `replace` 指向 `go/third_party/TarsGo/TarsGo-1.4.6`
- **Protobuf**：统一使用 `google.golang.org/protobuf v1.36.11`，生成的代码位于 `proto/generated/{go,ts,python}`
- **go-admin v2.2.0**：管理后台基于 go-admin 框架，前端使用 Vue 2 + Element UI（go-admin-ui v2.0.9）
- **GORM**：支持 MySQL、SQLite、Postgres、SQL Server 多驱动，Admin 默认使用 SQLite（CGO 编译）

---

## 18. 开发与运行方式

当前为 **S1 阶段**，在 S0 文档、规范、目录骨架基础上，已实现：
- **全局配置服务（Config）**：Schema Registry + DynamicConfigModule 自描述容器，支持运营自助扩展配置字段
- **多语言服务（I18n）**：参数化模板架构，支持 plain/named/icu 三种模板类型
- **Admin 管理后台（admin）**：go-admin v2.2.0 框架 + 自定义插件，提供 Schema CRUD / 配置值管理 / 多语言全量 API
- **Config/I18n Tars Servant**：标准 bytes 接口适配器 + LocalInvoker 注册 + E2E 集成测试
- **SDK 层（configsdk / i18nsdk）**：三层缓存（L1 LRU → L2 Redis → L3 远程兜底），业务服务通过 SDK 引用配置和多语言能力

### 18.1 常用运行命令

```bash
# 初始化环境
make bootstrap

# 生成 Protobuf 代码
make proto

# 运行全部测试
make test

# 运行 CI 检查
make ci

# 启动 Gateway（单体模式）
make gateway-start

# 启动 Admin MVP 前后端
make admin-dev

# 停止 Admin
make admin-stop

# 运行 E2E 测试
make e2e-admin
```

### 18.2 本地开发环境要求

| 工具 | 最低版本 | 说明 |
|---|---|---|
| Go | 1.21 | 后端开发 |
| Node.js | 20 | 前端开发 |
| Python | 3.11 | AI 服务与脚本 |
| protoc | - | Protobuf 代码生成（仅本地开发需要） |
| make | - | 工程编排 |
| sqlite3 | - | Admin 本地数据库 |

---

## 19. 测试与校验要求

### 19.1 Gateway 测试

| 测试项 | 要求 |
|---|---|
| MessagePacket 解析测试 | 验证 application/octet-stream 请求体可反序列化 |
| maxType/minType 路由测试 | 验证 route_key 正确命中 |
| routes.yaml 重复路由测试 | 重复 request_max/request_min 应启动失败 |
| 协议编号注册表一致性测试 | routes.yaml 中编号必须存在于注册表 |
| request_proto/response_proto 一致性测试 | proto 类型与 Type.max/min 一致 |
| .tars 方法存在性测试 | tars_method 必须存在于 interface |
| Tars bytes 调用测试 | request bytes / response bytes 完整透传 |
| Tars return 映射测试 | 10200/10401/10500/10504 正确映射 |
| extend 透传测试 | traceId/requestId/userId/tenantId 正确传递 |

### 19.2 System 模块测试

| 测试项 | 要求 |
|---|---|
| SystemService 业务逻辑测试 | HealthCheck / HelloWorld 返回正确 |
| LocalHandler bytes 适配测试 | 请求 bytes 正确反序列化，响应正确序列化 |
| LocalHandler 错误处理测试 | 非法 bytes / 未知 maxType/minType 返回错误 |

---

## 20. Module Path 规范

所有 Go module 使用统一 path 前缀：

```text
github.com/jimiechen/mineplanet/go/...
```

当前模块：

| 模块 | Path | 状态 |
|---|---|---|
| common-lib | github.com/jimiechen/mineplanet/go/common-lib | ✅ 新增 (2026-05-21) |
| modules/hello | github.com/jimiechen/mineplanet/go/modules/hello | ✅ 新增 (2026-05-21) |
| modules/health | github.com/jimiechen/mineplanet/go/modules/health | ✅ 新增 (2026-05-21) |
| gateway/proto-gateway | github.com/jimiechen/mineplanet/go/gateway/proto-gateway | ✅ 已有 |
| tars/system | github.com/jimiechen/mineplanet/go/tars/system | ✅ 已有 |
| services/config | github.com/jimiechen/mineplanet/go/services/config | ✅ 新增 (2026-05-22) |
| services/i18n | github.com/jimiechen/mineplanet/go/services/i18n | ✅ 新增 (2026-05-22) |
| admin | github.com/jimiechen/mineplanet/go/admin | ✅ 新增 (2026-05-26, 从 provider-admin 迁移) |
| tars/config | github.com/jimiechen/mineplanet/go/tars/config | ✅ 新增 (2026-05-22) |
| tars/i18n | github.com/jimiechen/mineplanet/go/tars/i18n | ✅ 新增 (2026-05-22) |

**预留模块（尚未实现）**：

| 模块 | Path | 计划阶段 |
|---|---|---|
| modules/users | go/modules/users | MVP2 |
| modules/auth | go/modules/auth | MVP2 |
| modules/groups | go/modules/groups | MVP2 |
| modules/topics | go/modules/topics | MVP2 |
| modules/readonly | go/modules/readonly | MVP2 |

---

## 21. 关键类与函数说明

### 21.1 Gateway 核心类型

#### `GatewayServer`（`go/gateway/proto-gateway/internal/server/http_server.go`）

```go
type GatewayServer struct {
    routeTable     *router.RouteTable
    invoker        tarsclient.TarsInvoker
    mode           string
    authMiddleware *middleware.AuthMiddleware
}
```

- `NewGatewayServer(rt, invoker, mode, authMiddleware)`：创建 GatewayServer
- `ServeHTTP(w, r)`：处理 POST `/api/hello`，完成 MessagePacket 反序列化、路由查找、鉴权、Tars 调用、响应封装

#### `TarsInvoker` 接口（`go/gateway/proto-gateway/tarsclient/invoker.go`）

```go
type TarsInvoker interface {
    Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}
```

- `LocalInvoker`：单体部署模式，本进程内通过 `handlers map` 转发调用
- `TarsGoInvoker`：微服务部署模式（S1 未实现）

#### `Target` / `TargetKey`（`go/gateway/proto-gateway/tarsclient/invoker.go`）

```go
type Target struct {
    App, Server, Servant, Module, Interface, Method string
}
type TargetKey struct { App, Server, Servant, Method string }
```

- `ToTargetKey(t Target) TargetKey`：生成目标唯一键
- `LocalInvoker.Register(key TargetKey, handler LocalHandler)`：注册本地 handler

### 21.2 公共库核心类型

#### `commonlib.Code`（`go/common-lib/codes.go`）

统一业务错误码：

```go
const (
    CodeSuccess            = 10200 // 操作成功
    CodeBadRequest         = 10400 // 请求参数错误
    CodeUnauthorized       = 10401 // 未授权
    CodeNotFound           = 10404 // 资源未找到
    CodeInternalError      = 10500 // 内部错误
    CodeTarsNotImplemented = 10501 // Tars 远程调用未实现（S1 阶段）
)
```

#### `module.Deps`（`go/common-lib/module/deps.go`）

模块统一依赖装配结构：

```go
type Deps struct {
    Config ConfigReader  // 必选：配置 SDK
    I18n   I18nRenderer // 可选：国际化 SDK
    DB     interface{}   // 可选：数据库连接
    Cache  interface{}   // 可选：缓存客户端
    Logger Logger       // 必选：日志组件
}
```

- `ConfigReader` 接口：`GetString / GetInt / GetBool / Watch / Ping`
- `I18nRenderer` 接口：`T / Raw / Ping`
- `Logger` 接口：`Info / Infof / Error / Errorf / Warn / Debug`

### 21.3 Hello 模块核心类型

#### `hello.Service`（`go/modules/hello/service.go`）

```go
type Service struct { handler *Handler }
func New(deps module.Deps) *Service
func (s *Service) SayHello(ctx context.Context, request []byte) ([]byte, error)
```

#### `hello.Handler`（`go/modules/hello/handler.go`）

```go
type Handler struct { usecase *Usecase; logger module.Logger }
func NewHandler(usecase *Usecase, logger module.Logger) *Handler
func (h *Handler) HandleSayHello(ctx context.Context, request []byte) ([]byte, error)
```

职责：Protobuf bytes ↔ 业务逻辑转换，不包含业务规则。

#### `hello.Usecase`（`go/modules/hello/usecase.go`）

```go
type Usecase struct { cfg ConfigReader; i18n I18nRenderer }
func NewUsecase(cfg ConfigReader, i18n I18nRenderer) *Usecase
func (u *Usecase) Greet(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldResponse, error)
```

职责：组合 configsdk + i18nsdk，执行核心业务逻辑（配置读取、长度校验、多语言渲染）。

### 21.4 Config 服务核心类型

#### `configservice.ConfigService`（`go/services/config/service/interface.go`）

```go
type ConfigService interface {
    GetAppConfigs(req *AppConfigRequest) (*AppConfigResponse, error)
    GetVersionInfo(env string, knownVersions map[string]int64) (*VersionInfoResponse, error)
}
```

#### `configsdk.Client`（`go/services/config/sdk/client.go`）

三层缓存客户端：

```go
type Client struct {
    lru      *lruCache      // L1 本地缓存
    redis    redisx.Client  // L2 Redis 缓存
    remote   RemoteClient   // L3 远程服务兜底
}
```

- `GetString / GetInt / GetBool`：读取配置值
- `Watch`：监听配置变更
- `Invalidate`：缓存失效（支持批量 + pub/sub 广播）

### 21.5 Admin 管理后台核心类型

#### `main.go`（`go/admin/main.go`）

```go
func main() { cmd.Execute() }
```

- 基于 `go-admin` 框架，通过 `cobra` 命令行启动
- 支持 `server`、`migrate` 等子命令
- 配置文件：`go/admin/config/settings.yml`

#### Admin Plugin APIs（`go/admin/app/admin/apis/`）

| 文件 | 职责 |
|---|---|
| `config_handler.go` | Schema CRUD、配置值发布 |
| `config_value_handler.go` | 配置值管理 |
| `i18n_handler.go` | 多语言字符串/语言包全量 API |
| `health_handler.go` | 健康检查 |

---

## 22. 项目规则规范

### 22.1 编码规范（`.trae/rules/coding.md`）

| 类别 | 规范 |
|---|---|
| 变量命名 | lowerCamelCase，布尔以 is/has/can/should 开头，禁止拼音 |
| 常量命名 | UPPER_SNAKE_CASE |
| 方法命名 | lowerCamelCase，动作/判断语义，查询用 get/find/list/calculate |
| 类命名 | UpperCamelCase，职责明确，Service/Controller/StateMachine 后缀 |
| 文件命名 | 类文件用类名，模块/脚本用 lower_snake_case |
| 文件规模 | 单文件 ≤300 行（推荐），绝对 ≤500 行；单方法 ≤30 行（推荐），绝对 ≤50 行 |
| PR 规模 | 推荐 ≤10 个文件，≤300 行代码 |
| 注释 | 默认中文，解释"为什么"和"约束"，禁止无意义注释 |
| TODO/FIXME | `TODO(姓名或模块, 截止阶段): 说明` / `FIXME(姓名或模块, 问题编号): 说明` |
| 职责 | 单一职责，禁止上帝类；状态机只做状态迁移；服务类做一类业务；工具类只放纯函数 |
| 代码风格 | 提前返回，嵌套 ≤3 层，禁止魔法数字，不允许吞异常，关键错误必须记日志 |

### 22.2 Git 工作流（`.trae/rules/git.md`）

| 类别 | 规范 |
|---|---|
| 分支 | `main`（稳定）、`dev`（集成）、`feature/*`、`fix/*`、`docs/*`、`test/*`、`refactor/*`、`chore/*`、`ci/*`、`hardware/*` |
| 分支命名 | `type/module-short-description`，如 `feature/ai-intent-classifier` |
| 提交粒度 | 一个提交只做一件事（加测试、修 bug、重构类、补文档） |
| Commit 格式 | `type(scope): 中文说明`，如 `feat(ai): 实现基础意图分类` |
| PR 要求 | 必须关联 Issue、说明变更内容、测试结果、文档是否更新、控制变更范围 |
| PR 规模 | 推荐 ≤10 个文件，≤300 行代码 |
| 合并规则 | 测试未通过、无关联 Issue、未说明测试结果、未更新文档、范围过大 均不得合并 |

### 22.3 TDD 开发规则（`.trae/rules/tdd.md`）

| 阶段 | 动作 |
|---|---|
| 需求红 | 阅读 PRD，明确验收标准、非目标、涉及模块 |
| 协议红 | 定义 Protobuf/OpenAPI 契约，注册 max+min 编号，更新注册表和映射规范 |
| 测试红 | 编写单元/集成/契约/边界测试，测试必须先失败且失败原因符合预期 |
| 实现绿 | 只写让当前测试通过的最小代码，不得顺手实现未来功能 |
| 报告绿 | 运行所有测试，生成测试报告，输出 Standalone HTML，截图/视频证据 |
| CI 绿 | GitHub Actions 通过，或提供本地等价命令和结果 |
| 重构 | 测试与 CI 通过后再优化，外部行为不变 |
| 沉淀 | 提交日报、Markdown 蒸馏、更新 LLM Wiki |

**禁止行为**：
- 先写实现再补测试
- 删除测试来通过构建
- 只测正常路径不测异常路径
- 用 Mock 跳过核心逻辑
- 单元测试调用真实大模型接口
- 没有 CI 结果就宣称完成

### 22.4 测试规则（`.trae/rules/testing.md`）

| 类别 | 规范 |
|---|---|
| 测试目标 | 验证需求，回答：功能是否符合 PRD、异常是否处理、状态转换是否正确、协议是否稳定、安全边界是否破坏 |
| 测试分层 | 单元测试、参数化测试、状态机测试、协议测试、集成测试、端到端测试、回归测试 |
| 测试位置 | AI 模块 `ai/.../tests/`，固件 `firmware/.../test/`，App `app/.../test/`，E2E `tests/e2e/`，fixtures `tests/fixtures/` |
| 测试命名 | 文件：`intent_classifier_test.py`；函数：表达业务含义，如 `test_当舱门未关闭时_锁定命令应失败` |
| 外部服务 | 单元测试不得调用真实外部服务（大模型、OCR、云端、硬件），使用 Mock/Fake/fixture |
| 测试规模 | 单个测试文件 ≤300 行 |
| CI 要求 | 必须通过 GitHub Actions，跳过需明确原因 |
| 报告位置 | 测试报告 `docs/reports/testing/`，日报 `docs/reports/daily/`，蒸馏 `docs/reports/distilled/`，HTML `docs/reports/html/` |

---

## 23. 文档路径索引

### 23.1 工程规范文档（`.trae/rules/`）

| 文件 | 职责 |
|---|---|
| `coding.md` | 编码规范（命名、文件规模、注释、职责、代码风格） |
| `commenting.md` | 中文注释规范（Go/Protobuf/TS/Python 注释规则） |
| `docs.md` | 文档规则（目录约束、PRD/ADR/API/测试文档规范） |
| `git.md` | Git 工作流（分支、提交、PR、合并规则） |
| `makefile.md` | Makefile 工程入口规范（三层结构、target 列表、CI 集成） |
| `project.md` | Trae 项目执行规则（任务前必读、TDD 强制、变更边界） |
| `reporting.md` | 任务汇报规则（日报、Bug、事故、风险、阻塞、评审机制） |
| `review.md` | 评审规则（评审维度、AI 评审输出格式、不允许合并情况） |
| `tdd.md` | TDD 开发规则（红绿重构流程、完成标准、禁止行为） |
| `testing.md` | 测试规则（分层、命名、外部服务、CI、报告） |

### 23.2 架构决策文档（`docs/adr/`）

| 文件 | 决策内容 |
|---|---|
| `ADR-0001-总体系统架构.md` | 总体架构选型 |
| `ADR-0002-后端使用Golang.md` | 后端语言选型 |
| `ADR-0003-服务协议使用Protobuf.md` | Protobuf 协议决策 |
| `ADR-0004-AI服务使用Python.md` | AI 服务语言选型 |
| `ADR-0005-App前端使用ReactJS.md` | 前端框架选型 |
| `ADR-0006-开放平台API边界.md` | 开放平台边界 |
| `ADR-0007-服务商后台与用户中台边界.md` | 后台与中台边界 |
| `ADR-0008-use-tarscloud-routing-layer.md` | TarsCloud 路由层选型 |
| `ADR-0009-tabbit-task-archive-flow.md` | Tabbit 任务归档流程 |
| `ADR-0012-polyglot-monorepo-directory-layout.md` | 多语言 Monorepo 目录布局 |
| `ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md` | Makefile 工程入口 |
| `ADR-0014-message-packet-data-format-protobuf-bytes.md` | MessagePacket 数据格式 |
| `ADR-0015-hello-health-as-sample-module.md` | Hello/Health 作为参考模块 |
| `ADR-009-config-i18n-schema-template.md` | Config/I18n Schema 模板 |
| `ADR-010-admin-boundary-sdk.md` | Admin 边界与 SDK |

### 23.3 产品需求文档（`docs/prd/`）

| 文件 | 需求内容 |
|---|---|
| `PRD-00-MVP总纲.md` | MVP 总纲 |
| `PRD-01-服务商后台系统.md` | 服务商后台 |
| `PRD-01-核心用户流程.md` | 核心用户流程 |
| `PRD-02-工程交付与验证规范.md` | 工程交付规范 |
| `PRD-02-终端用户中台系统.md` | 终端用户中台 |
| `PRD-03-开放平台API.md` | 开放平台 API |
| `PRD-04-AI服务系统.md` | AI 服务系统 |
| `PRD-05-App前端系统.md` | App 前端系统 |
| `PRD-06-设备通信与协议.md` | 设备通信与协议 |
| `PRD-09-HelloWorld与HealthCheck验收规范.md` | HelloWorld 验收规范 |
| `PRD-10-Admin管理后台MVP.md` | Admin 管理后台 MVP |
| `PRD-11-proto-tester-v2.md` | Proto Tester V2 |

### 23.4 API 与协议文档（`docs/api/`）

| 文件 | 内容 |
|---|---|
| `protobuf规范.md` | Protobuf 编码规范 |
| `tars规范.md` | Tars 接口规范 |
| `http-gateway规范.md` | HTTP 网关规范 |
| `openapi-protobuf映射规范.md` | OpenAPI 与 Protobuf 映射 |
| `OpenAPI规范.md` | OpenAPI 规范 |
| `协议编号注册表.md` | max+min 协议编号注册表 |
| `Webhook规范.md` | Webhook 规范 |
| `gRPC接口规范.md` | gRPC 接口规范（历史） |

### 23.5 测试与报告文档（`docs/testing/` / `docs/reports/`）

| 文件 | 内容 |
|---|---|
| `docs/testing/TDD规范.md` | TDD 规范 |
| `docs/testing/测试报告规范.md` | 测试报告规范 |
| `docs/testing/测试用例注册表.md` | 测试用例注册表 |
| `docs/testing/测试用例规范.md` | 测试用例规范 |
| `docs/testing/截图视频证据规范.md` | 截图视频证据规范 |
| `docs/reports/daily/` | 日报 |
| `docs/reports/distilled/` | Markdown 蒸馏 |
| `docs/reports/html/` | Standalone HTML 报告 |
| `docs/reports/testing/` | 测试报告 |

---

## 24. Admin 管理后台（go-admin v2.2.0）

### 24.1 架构定位

```text
Admin HTTP → go/admin/plugins/{config,i18n}_admin/apis/
         → services/{config,i18n}/admin/（写入层）
           → 复用 services/{config,i18n}/service/（校验层）
           → repository → MySQL
           → redisx.Client.Invalidate（缓存失效）
           → redisx.PubSubClient.Publish（变更广播）
```

**铁律：admin 插件禁止直写 sys_config_* / sys_lang_* 表。**

### 24.2 Redis 访问约定

admin 子包同时持有两个 Redis 实例，职责严格分离：

| 实例 | 接口 | 用途 | 方法 |
|---|---|---|---|
| cache | `redisx.Client` | 缓存失效 | `Invalidate(ctx, pattern)` |
| bus | `redisx.PubSubClient` | 变更广播 | `Publish(ctx, channel, payloadJSON)` |

**禁止在 `redisx.Client` 接口上扩展 Publish（保持职责单一）。**

### 24.3 pub/sub 协议

**Channel 命名：**
- config 失效：`cairobot.config.invalidate`
- i18n 失效：`cairobot.i18n.invalidate`

**Payload 格式（InvalidateEvent JSON）：**
```json
{
  "tenant_id": "default",
  "scope": "config|i18n",
  "env": "prod|dev|test",
  "module_keys": ["key1","key2"],
  "lang_codes": ["zh-CN","en"],
  "version": 42,
  "timestamp": 1716739200,
  "trace_id": "uuid-v4"
}
```

**SDK 消费端升级策略（向后兼容）：**
1. 优先尝试 `json.Unmarshal` → 解析成功且 tenant_id 非空 → handleStructured
2. 否则按逗号分隔降级 → handleLegacy + WARN 日志
3. 降级分支保留至 S2 阶段

### 24.4 职责边界自检

```bash
# admin 不做字段级校验
grep -E "field_type ==|switch.*field_type|validator JSON" \
  go/services/config/admin/*.go go/services/i18n/admin/*.go
# 必须为空

# admin 必须复用 service 层
grep "ValidateSchema\|ValidateValue\|ValidateLangString" \
  go/services/config/admin/*.go go/services/i18n/admin/*.go
# 必须命中
```

---

## 25. proto-tester 协议测试客户端

### 25.1 架构定位

proto-tester 是研发/QA 团队的**内部协议测试工具**，用于 Protobuf 协议联调场景下的 MessagePacket 构造、发送、响应解析和链路追踪。

```text
开发者 / QA
    ↓ [Web UI 或 CLI]
proto-tester（:3001）
    ↓ [HTTP POST + Bearer Token]
Gateway（:8080）
    ↓ [MessagePacket 路由]
TarsGo 服务
```

### 25.2 技术选型

| 类别 | 技术 | 说明 |
|------|------|------|
| Protobuf 运行时 | google-protobuf 3.21.2 | 与 typescript/web 共享 |
| JSON 编辑器 | CodeMirror 6 | <250KB，MIT 许可 |
| UI 框架 | antd 5 | 企业级组件 |
| 状态管理 | zustand 4.x | Token 100% 内存化 |
| 本地存储 | idb 8.x | IndexedDB 封装（请求历史） |
| 构建工具 | vite 5.x | 含覆盖率阈值 |

### 25.3 核心特性

- **元数据驱动**：构建期 `gen-protocols.mjs` 从 .proto 生成 `protocols.json`，运行时零 .proto 依赖
- **Token 内存化**：不持久化到任何存储，刷新即清空
- **CSP 硬编码**：`index.html` 内嵌 Content-Security-Policy
- **CLI 支持**：send / trace / run / capture 四个子命令
- **5 个预置角色账号**：admin / operator / viewer / user / attacker

### 25.4 安全约束

- ⚠️ **禁止公网部署**
- 默认绑定 `127.0.0.1:3001`
- CSP connect-src 仅允许内网 Gateway 地址
- CLI `--env prod` 在 blocklist 中被拦截

### 25.5 相关文档

- 详细 ADR: [ADR-016](adr/ADR-016-proto-tester.md)
- 项目 README: `typescript/proto-tester/README.md`
- 测试覆盖: 100 tests, 81%+ statements coverage

### 24.5 M0'~M5' 交付清单（2026-05-27 完成）

| 批次 | 层级 | 模块 | 文件数 | 测试数 | 状态 |
|------|------|------|--------|--------|------|
| M0' | SDK 升级 | config/sdk/pubsub.go 三阶兼容 + i18n 可行性报告 | 3 报告 | 17 (pubsub) | ✅ |
| M1' | 服务层 | config/admin + i18n/admin（CRUD+审计+缓存+广播） | 12 文件 | 24 | ✅ |
| M2' | 插件层 | config_admin/apis（Schema CRUD + Value 发布） | 6 文件 | 12 | ✅ |
| M3' | 插件层 | i18n_admin/apis（字符串 CRUD + 包管理 + CSV 导入导出） | 7 文件 | 20 | ✅ |
| M4' | 前端 | admin-web 配置管理页（schema-list + value-publish） | 5 文件 | build✅ | ✅ |
| M5' | 前端 | admin-web 国际化管理页（string-list + pack-manage + import-export） | 7 文件 | build✅ | ✅ |
| **合计** | | | **40 文件** | **56+17=73** | **✅** |

### 24.6 关键接口速查

**config/admin 核心接口：**
```go
type ConfigAdminService interface {
    ConfigSchemaService   // ListSchemas/CreateSchema/UpdateSchema/DeleteSchema
    ConfigValueService    // PublishValue/GetValueVersions
}
```

**i18n/admin 核心接口：**
```go
type I18nAdminService interface {
    I18nStringService     // CreateString/UpdateString/DeleteString/ListStrings
    I18nPackService       // PublishPack/RollbackPack/ImportStringsFromCSV/ExportStringsToCSV
}
```

**HTTP 路由前缀：**
- 配置：`/api/admin/v1/config/{schema,value}/*`
- 国际化：`/api/admin/v1/i18n/{string,pack}/*` + `/import/csv` + `/export/csv`

**10400 错误码：** 校验失败统一返回 `{"code":10400,"errors":[{"field":"...","message":"..."}]}`

---

## 26. Makefile 工程入口

### 26.1 三层架构

```text
Makefile（根目录总控）
├── go/Makefile
├── typescript/Makefile
├── python/Makefile
└── scripts/
    ├── ci/          # check_*.py / check_*.sh
    ├── proto/       # generate-*.sh
    └── coverage/    # *_coverage.sh
```

### 26.2 常用命令

```bash
make help        # 显示帮助
make bootstrap   # 初始化环境
make proto       # 生成 Protobuf 代码
make lint        # Lint 检查
make test        # 全部测试
make unit        # 单元测试
make ci          # 完整 CI 检查
make rules       # 规范检查
make clean       # 清理产物
```

### 26.3 相关文档

- [ADR-0013](../adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md)
- [.trae/rules/makefile.md](../../.trae/rules/makefile.md)

---

## 27. CI / GitHub Actions 配置

### 27.1 Workflow 文件

`.github/workflows/ci.yml`

### 27.2 CI Jobs

| Job | 说明 | 触发条件 |
|---|---|---|
| `ci-full` | 完整 CI 检查（`make ci` 编排） | PR / push 到 main/dev/feature/fix/docs/test |
| `docs-check` | 检查关键文档是否存在 | PR |
| `proto-check` | 检查协议编号唯一性和注册表同步 | PR |
| `report-check` | 检查报告完整性 | PR |
| `module-lint` | 模块合规性 Lint（10+1 项强制检查） | PR |

### 27.3 运行环境

| 工具 | 版本 |
|---|---|
| Ubuntu | latest |
| Go | 1.21 |
| Node.js | 20 |
| Python | 3.11 |

---

## 28. 变更日志

| 日期 | 变更内容 |
|---|---|
| 2026-05-26 | **M0' 阶段 Admin 管理后台升级**：废弃 tars/provider-admin（Gin），迁移至 go/admin（go-admin v2.2.0 框架）；新增 typescript/admin-web（go-admin-ui v2.0.9 Vue2 前端）；新增 redisx.Invalidate 分批删除 API；pub/sub payload 升级为 JSON InvalidateEvent 格式；DSN 加密存储（AES-256-CBC）；provider-admin 源码归档至 archive/provider-admin-v0 分支（保留至 2026-11） |
| 2026-05-21 | **Go 多模块重构**：新增 common-lib、modules/hello、modules/health 独立 go.mod；实现 LocalInvoker 模块注册机制（ModuleHandler + Adapter）；迁移 localhandler → adapter/deprecated（标记废弃）；**TS E2E 集成**：集成 proto/generated/ts 官方 Protobuf 类型；新增 tests/e2e/gateway-modules.test.ts 全链路测试；抽取 src/utils/proto-client.ts 工具函数库；目录规范化为 tests/unit + tests/e2e 分层结构；Makefile 多语言工具链自动发现 Go/Python/Node.js 路径；全量测试 89/89 PASS |
| 2026-05-20 | **架构口径修正**：local 模式不是"不使用 Tars"，而是 TarsGo 单体部署模式（monolith）；LocalInvoker 是本进程 TarsGo servant adapter，非绕过 Tars 的普通 Go 调用；proto-gateway 改造为基于 TarsGo HTTP 模块（TarsHttpMux / AddHttpServant / Run）的 TarsGo HTTP Servant；引入 TarsCloud/TarsGo v1.4.6 技术基线到 go/third_party/TarsGo/ |
| 2026-05-19 | 按 ADR-0012 重构多语言 monorepo 目录布局：Go 进 go/、Python 进 python/、TypeScript 进 typescript/；删除根目录 Makefile；go.work 移至 go/ |
| 2026-05-19 | 恢复根目录 Makefile，采用三层结构（总控 + 子 Makefile + scripts）；新增 16 个 target；新增 CI 规范检查脚本；建立测试用例注册表；新增中文注释规范 |
| 2026-05-19 | 实现 Gateway 单体/微服务双模式骨架；System 模块独立存在；建立 Go Workspace；统一 TarsInvoker 接口；module path 标准化 |
| 2026-05-18 | 根据 ADR-0008，内部核心服务主链路从 gRPC 调整为 TarsCloud/TarsGo；外部入口收敛为单网关 POST /api/hello；MessagePacket 成为唯一入口报文 |

---

## 29. 相关文档索引

**ADR（架构决策）**：
- [ADR-0001-总体系统架构](../adr/ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf](../adr/ADR-0003-服务协议使用Protobuf.md)
- [ADR-0008-use-tarscloud-routing-layer](../adr/ADR-0008-use-tarscloud-routing-layer.md)
- [ADR-0012-polyglot-monorepo-directory-layout](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- [ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement](../adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md)
- [ADR-0014-message-packet-data-format-protobuf-bytes](../adr/ADR-0014-message-packet-data-format-protobuf-bytes.md)
- [ADR-009-config-i18n-schema-template](../adr/ADR-009-config-i18n-schema-template.md)
- [ADR-010-admin-boundary-sdk](../adr/ADR-010-admin-boundary-sdk.md)

**API 规范**：
- [protobuf规范](../api/protobuf规范.md)
- [tars规范](../api/tars规范.md)
- [http-gateway规范](../api/http-gateway规范.md)
- [openapi-protobuf映射规范](../api/openapi-protobuf映射规范.md)
- [OpenAPI规范](../api/OpenAPI规范.md)
- [协议编号注册表](../api/协议编号注册表.md)

**设计文档**：
- [Go Monorepo 模块化重构设计](../superpowers/specs/2026-05-21-go-monorepo-modular-refactoring-design.md)
