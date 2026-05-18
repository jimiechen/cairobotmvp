# CaiRobot MVP Code Wiki

## 1. 项目概述

CaiRobot MVP 是一个多系统、多技术栈的微服务架构项目，旨在为教育机器人设备提供完整的软硬件解决方案。系统包含服务商后台、终端用户中台、开放平台、AI服务、设备通信等多个子系统，通过统一的Protobuf协议进行服务间通信。

### 1.1 项目阶段

当前处于 **S0：规则与目录骨架** 阶段，主要完成以下工作：

- 目录骨架搭建
- 工程规范制定
- PRD/ADR文档建设
- Protobuf协议定义框架
- CI/CD流程配置

### 1.2 核心约束

| 约束项 | 说明 |
|-------|------|
| 开发流程 | PRD驱动 + Issue拆分 + TDD测试驱动 |
| 代码规范 | 遵循coding.md命名和规模限制 |
| 文档同步 | 代码变更必须同步更新相关文档 |
| PR要求 | 必须关联Issue，必须有测试结果 |

## 2. 技术栈概览

### 2.1 后端技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 核心语言 | Golang | 高性能、编译型、并发友好 |
| 服务治理框架 | TarsCloud | 服务注册、发现、负载均衡、熔断、监控 |
| 内部通信协议 | TarsGo | Tars 协议，高性能服务间通信 |
| 对外接口契约 | Protobuf | 统一接口定义 |
| 对外通信 | gRPC + HTTPS JSON | gRPC 直接接入或通过 grpc-gateway 转换 |
| Web框架 | TarsGo 内置 / Gin/Echo（待定） | HTTP网关和HTTP服务 |
| ORM | 待定 | 数据库访问层 |
| 依赖管理 | Go Modules | go.mod管理依赖 |

### 2.2 AI服务技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 核心语言 | Python | 机器学习生态最完善 |
| Web框架 | FastAPI/Flask（待定） | HTTP/gRPC服务 |
| AI框架 | 待定 | 模型推理封装 |

### 2.3 前端技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 核心框架 | ReactJS | 组件化、生态成熟 |
| UI组件库 | 待定（Ant Design/Material-UI） | 企业级组件 |
| 状态管理 | 待定（Redux/MobX/Context） | 状态管理方案 |
| 路由 | React Router | 前端路由 |
| 测试 | Jest + React Testing Library | 单元/组件测试 |

### 2.4 协议与基础设施

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 接口定义 | Protobuf | 统一契约来源 |
| 对外API | HTTPS JSON + OpenAPI | grpc-gateway转换 |
| 版本控制 | Git | 分支策略见git.md |

## 3. 系统架构

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                          外部系统                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ 第三方系统 │  │ 渠道系统  │  │ 终端App  │  │ 服务商后台 │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
└───────┼─────────────┼─────────────┼─────────────┼─────────────┘
        │             │             │             │
        │ HTTPS JSON / gRPC / Protobuf
        └─────────────┴──────┬──────┴─────────────┘
                             │
                    ┌────────▼────────┐
                    │   API Gateway   │
                    │ (grpc-gateway)  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Proto Gateway  │
                    │ (协议解析与路由) │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ TarsCloud Routing│
                    │ (服务治理与转发) │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐   ┌────────▼────────┐   ┌───────▼───────┐
│Provider Admin │   │  User Center   │   │ Open Platform │
│   TarsServer  │   │   TarsServer   │   │   TarsServer  │
└───────┬───────┘   └────────┬────────┘   └───────┬───────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐   ┌────────▼────────┐   ┌───────▼───────┐
│   Auth Server │   │  AiBridgeServer │   │Device Gateway │
│   TarsServer  │   │   (Python桥接)  │   │   TarsServer  │
└───────────────┘   └────────┬────────┘   └───────┬───────┘
                             │                    │
                             │            ┌───────▼───────┐
                             │            │   AI Service   │
                             │            │   (Python)     │
                             │            └───────┬───────┘
                             │                    │
                             │            ┌───────▼───────┐
                             │            │    设备        │
                             │            │  (Robot)      │
                             │            └───────────────┘
                             │
                    ┌────────▼────────┐
                    │  Audit Server   │
                    │   TarsServer    │
                    └─────────────────┘
```

### 3.2 架构说明

CaiRobot MVP 在现有 Protobuf 统一协议基础上，引入 TarsCloud 作为内部服务治理与路由转发框架：

1. **对外接口层**：仍然以 Protobuf 作为统一契约来源，可通过 gRPC 直接接入，或通过 grpc-gateway 暴露 HTTPS JSON API。

2. **API Gateway**：负责基础的接入、限流、鉴权，将 HTTPS JSON 转换为 Protobuf。

3. **Proto Gateway**：负责解析 Protobuf 请求，根据 package/service/method 查找路由配置。

4. **TarsCloud 路由层**：负责服务发现、负载均衡、路由转发，将请求转发到内部 TarsGo servant。

5. **TarsGo 内部服务**：核心业务逻辑采用 TarsGo 实现，通过 TarsCloud 获得完整的服务治理能力。

6. **外部服务桥接**：如 AI Service（Python）、设备等，通过 AiBridgeServer 或 DeviceGatewayServer 接入 TarsCloud 生态。

这种架构的优势：
- 职责清晰：Gateway 只做协议转换和路由，不承载业务逻辑。
- 治理完备：内部服务通过 TarsCloud 获得完整的服务治理能力。
- 易于扩展：新增服务只需在路由配置中添加，无需修改 Gateway。
- 灰度友好：可以方便地实现灰度发布，降低发布风险。

### 3.2 服务职责矩阵

| 服务 | 目录 | 职责 | 技术栈 | 协议端口 |
|------|------|------|--------|----------|
| API Gateway | services/api-gateway/ | 统一入口、路由、限流、鉴权 | Golang | gRPC/HTTP |
| Provider Admin | services/provider-admin/ | 服务商账号、设备批次、工单管理 | Golang | gRPC |
| User Center | services/user-center/ | 用户账号、家庭、设备绑定、学习记录 | Golang | gRPC |
| Open Platform | services/open-platform/ | 开放API认证、配额、日志 | Golang | gRPC |
| Device Gateway | services/device-gateway/ | 设备连接、通信、控制 | Golang | gRPC |
| Auth Service | services/auth-service/ | 统一认证、授权、Token管理 | Golang | gRPC |
| Audit Service | services/audit-service/ | 操作日志、审计追踪 | Golang | gRPC |
| AI Service | ai/service/ | 意图分类、提示词改写、回答审核 | Python | gRPC/HTTP |

### 3.3 服务间通信关系

```
App/第三方
    │
    ▼
API Gateway ──────────────────┐
    │                         │
    ├──► Provider Admin       │
    │                         │
    ├──► User Center ────────►│──► Auth Service
    │                         │         │
    ├──► Open Platform ──────►│──► AI Service
    │                         │         │
    └──► Device Gateway ─────►│──► Audit Service
                                  │
                                  ▼
                               设备
```

## 4. 目录结构

### 4.1 顶层目录

```
cairobotmvp/
├── .github/                # GitHub配置（Issue模板、PR模板、CI/CD）
├── .trae/                  # 工程规范（rules、skills）
├── docs/                   # 文档目录
│   ├── adr/               # 架构决策记录
│   ├── api/                # API与协议文档
│   ├── prd/               # 产品需求文档
│   ├── reports/            # 报告目录
│   ├── testing/           # 测试策略文档
│   └── wiki/              # Wiki索引
├── proto/                  # Protocol Buffers协议定义
├── gateway/                # Gateway层
│   └── proto-gateway/      # Protobuf协议转换与路由网关
├── tars/                   # TarsGo服务目录
│   └── go/                # TarsGo服务实现
│       ├── user-center/    # 用户中心服务
│       ├── provider-admin/ # 服务商后台服务
│       ├── open-platform/  # 开放平台服务
│       ├── device-gateway/ # 设备网关服务
│       ├── auth/          # 认证服务
│       ├── audit/         # 审计服务
│       └── ai-bridge/     # AI服务桥接
├── services/               # 原Golang后端服务（逐步迁移）
├── ai/                     # AI服务（Python）
├── web/                    # ReactJS前端项目
├── app/                    # 终端App（可选）
├── hardware/               # 硬件文档（可选）
├── firmware/              # 固件代码（可选）
├── deploy/                 # 部署相关
│   └── tarscloud/         # TarsCloud部署配置
├── scripts/                # 工具脚本
├── tests/                  # 测试目录（可选）
└── AGENTS.md               # 工程协作规范
```

### 4.2 服务目录结构（Golang服务）

每个Golang服务遵循DDD分层架构：

```
services/[service-name]/
├── README.md              # 服务说明文档
├── cmd/                   # 入口程序
│   └── server/           # 服务启动入口
│       └── main.go
├── internal/              # 内部代码
│   ├── domain/           # 领域层
│   │   ├── model/        # 领域模型
│   │   ├── repository/   # 仓储接口
│   │   └── service/     # 领域服务接口
│   ├── application/     # 应用层
│   │   ├── usecase/     # 用例编排
│   │   └── dto/         # 数据传输对象
│   ├── infrastructure/   # 基础设施层
│   │   ├── persistence/ # 持久化实现
│   │   ├── cache/       # 缓存实现
│   │   └── external/    # 外部服务调用
│   └── interfaces/      # 接口层
│       ├── grpc/        # gRPC处理器
│       └── http/        # HTTP处理器
├── api/                   # 生成的API代码
├── configs/              # 配置文件
├── tests/                 # 测试文件
├── go.mod                 # Go模块定义
└── go.sum                 # 依赖锁定
```

### 4.3 AI服务目录结构（Python）

```
ai/
├── README.md
└── service/              # AI服务主目录
    ├── README.md
    ├── app/              # 应用代码
    │   ├── __init__.py
    │   ├── api/          # API层
    │   │   ├── __init__.py
    │   │   ├── grpc_server.py
    │   │   └── http_server.py
    │   ├── core/         # 核心业务逻辑
    │   │   ├── __init__.py
    │   │   ├── classifier.py      # 意图分类器
    │   │   ├── prompt_rewriter.py # 提示词改写
    │   │   └── answer_reviewer.py # 回答审核
    │   ├── prompts/      # 提示词模板
    │   │   └── templates/
    │   ├── safety/       # 安全策略
    │   │   ├── __init__.py
    │   │   └── policies/
    │   ├── inference/    # 推理封装
    │   │   ├── __init__.py
    │   │   └── model_gateway.py
    │   └── schemas/      # 数据模型
    │       ├── __init__.py
    │       └── requests.py
    ├── tests/            # 测试目录
    │   ├── unit/
    │   ├── safety/
    │   └── contract/
    ├── fixtures/         # 测试数据
    ├── requirements.txt  # Python依赖
    └── pyproject.toml    # 项目配置
```

### 4.4 前端目录结构（ReactJS）

```
web/
├── README.md
├── [project-name]/       # 各前端项目
│   ├── README.md
│   ├── public/          # 静态资源
│   ├── src/             # 源代码
│   │   ├── components/ # 组件
│   │   ├── pages/      # 页面
│   │   ├── services/   # API服务
│   │   ├── stores/     # 状态管理
│   │   ├── hooks/      # 自定义Hook
│   │   ├── utils/      # 工具函数
│   │   └── App.tsx     # 根组件
│   ├── package.json
│   └── tsconfig.json
└── ...
```

### 4.5 Proto Gateway 目录结构

```
gateway/proto-gateway/
├── README.md                    # Gateway说明文档
├── cmd/
│   └── server/                  # 服务启动入口
│       └── main.go
├── internal/
│   ├── router/                  # 路由模块
│   │   ├── router.go            # 路由查找逻辑
│   │   └── config.go            # 路由配置加载
│   ├── adapter/                 # 协议适配模块
│   │   ├── proto_adapter.go     # Protobuf与Tars协议转换
│   │   └── result_adapter.go    # 结果转换
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # 鉴权中间件
│   │   ├── audit.go             # 审计中间件
│   │   └── tracing.go           # 链路追踪中间件
│   ├── tarsclient/              # Tars客户端封装
│   │   ├── client.go            # Tars客户端
│   │   └── pool.go              # 连接池
│   └── server/                  # 服务端
│       ├── grpc.go              # gRPC服务
│       └── http.go              # HTTP服务（可选）
├── configs/
│   └── routes.yaml              # 路由配置文件
├── api/                         # 生成的API代码
└── go.mod                       # Go模块定义
```

### 4.6 TarsGo 服务目录结构

每个 TarsGo 服务遵循以下目录结构：

```
tars/go/[service-name]/
├── README.md                    # 服务说明文档
├── tars/                       # Tars定义文件
│   └── [service].tars          # Tars接口定义
├── go/                        # Tars生成代码
│   └── [servant]/
│       ├── [servant].go        # Servant接口
│       └── ...
├── cmd/
│   └── server/                 # 服务启动入口
│       └── main.go
├── internal/
│   ├── domain/                 # 领域层
│   │   ├── model/              # 领域模型
│   │   ├── repository/         # 仓储接口
│   │   └── service/           # 领域服务
│   ├── application/            # 应用层
│   │   ├── usecase/           # 用例编排
│   │   └── dto/               # 数据传输对象
│   ├── infrastructure/         # 基础设施层
│   │   ├── persistence/       # 持久化实现
│   │   ├── cache/             # 缓存实现
│   │   └── external/          # 外部服务调用
│   └── interfaces/            # 接口层
│       └── servant/           # Tars Servant实现
├── configs/                   # 配置文件
│   └── config.conf            # Tars服务配置
├── tests/                     # 测试文件
└── go.mod                     # Go模块定义
```

### 4.7 协议目录结构

```
proto/
├── README.md             # 协议总体说明
├── base/                 # 基础协议
│   ├── health.proto     # 健康检查
│   ├── message.proto    # 网关统一入口
│   └── result.proto     # 通用返回
├── common/              # 通用定义
│   ├── error.proto     # 错误码定义
│   └── pagination.proto # 分页定义
├── provider_admin/      # 服务商后台协议
│   └── v1/
│       └── *.proto
├── user_center/         # 用户中台协议
│   └── v1/
│       └── *.proto
├── open_platform/       # 开放平台协议
│   └── v1/
│       └── *.proto
├── ai_service/          # AI服务协议
│   └── v1/
│       └── *.proto
└── device/             # 设备通信协议
    └── v1/
        └── *.proto
```

## 5. TarsCloud 路由层设计

### 5.1 定位与职责

TarsCloud 路由层是 Protobuf 请求到内部 TarsGo 服务的统一路由转发层，它位于 Proto Gateway 与内部 TarsGo 服务之间。

**核心职责**：
1. **协议解析**：解析 Protobuf 请求，提取 package、service、method。
2. **上下文透传**：透传 trace-id、auth-token 等上下文信息。
3. **协议适配**：将 Protobuf message 与内部 TarsGo 请求结构进行适配转换。
4. **Tars servant 定位**：根据路由配置找到目标 TarsGo servant。
5. **服务发现**：通过 TarsCloud 发现可用的服务实例。
6. **负载均衡**：选择合适的服务实例进行调用。
7. **超时控制**：设置合理的超时时间，防止请求堆积。
8. **错误码转换**：将内部 Tars 错误转换为统一的 Protobuf Result 响应。
9. **链路追踪**：记录和传播 trace-id，便于问题排查。
10. **统一日志与指标上报**：记录请求日志，上报监控指标。

**非职责**：
- 不承载复杂业务逻辑。
- 不直接访问业务数据库。
- 不绕过 AuthServer / AuditServer。
- 不修改 Protobuf 契约语义。

### 5.2 TarsGo 内部服务规范

#### 5.2.1 命名规范

| 层级 | 规范 | 示例 |
|------|------|------|
| App 名 | CaiRobot（固定） | CaiRobot |
| Server 名 | UpperCamelCase，以 Server 结尾 | UserCenterServer |
| Servant 名 | UpperCamelCase，以 Obj 结尾 | UserCenterObj |
| RPC 方法名 | UpperCamelCase | BindDevice |

#### 5.2.2 服务清单

| Server 名 | Servant 名 | 职责说明 |
|-----------|------------|----------|
| UserCenterServer | UserCenterObj | 家庭、用户、设备绑定、学习记录管理 |
| ProviderAdminServer | ProviderAdminObj | 服务商账号、设备批次、工单管理 |
| OpenPlatformServer | OpenPlatformObj | 第三方 API、认证、配额、Webhook |
| DeviceGatewayServer | DeviceGatewayObj | 设备连接、设备状态、指令下发 |
| AuthServer | AuthObj | 登录、鉴权、Token 校验、权限上下文 |
| AuditServer | AuditObj | 审计日志、操作追踪 |
| AiBridgeServer | AiBridgeObj | 连接 Python AI 服务，封装意图分类、提示词改写、回答审核等能力 |

### 5.3 Protobuf 到 TarsGo 路由规则

路由通过 routes.yaml 配置文件定义，支持以下路由策略：

1. **按 proto package 路由**：将同一 package 的所有请求路由到特定服务。
2. **按 proto service 路由**：将同一 service 的所有请求路由到特定服务。
3. **按 proto method 路由**：将特定 method 的请求路由到特定服务。
4. **按 protocol code 路由**：按协议编号（max+min）路由（兼容现有协议编号机制）。

路由配置示例：

```yaml
routes:
  - proto_package: cairobot.user.v1
    proto_service: UserCenterService
    proto_method: BindDevice
    tars_app: CaiRobot
    tars_server: UserCenterServer
    tars_servant: UserCenterObj
    tars_method: BindDevice
    timeout_ms: 3000
    auth_required: true
    audit_required: true

  - proto_package: cairobot.device.v1
    proto_service: DeviceGatewayService
    proto_method: SendCommand
    tars_app: CaiRobot
    tars_server: DeviceGatewayServer
    tars_servant: DeviceGatewayObj
    tars_method: SendCommand
    timeout_ms: 5000
    auth_required: true
    audit_required: true

  - proto_package: cairobot.open.v1
    proto_service: OpenPlatformService
    proto_method: CreateWebhook
    tars_app: CaiRobot
    tars_server: OpenPlatformServer
    tars_servant: OpenPlatformObj
    tars_method: CreateWebhook
    timeout_ms: 3000
    auth_required: true
    audit_required: true
```

路由映射表字段说明：

| 字段 | 说明 | 示例 |
|------|------|------|
| proto_package | Protobuf 包名 | cairobot.user.v1 |
| proto_service | Protobuf service 名 | UserCenterService |
| proto_method | Protobuf method 名 | BindDevice |
| tars_app | Tars App 名 | CaiRobot |
| tars_server | Tars Server 名 | UserCenterServer |
| tars_servant | Tars Servant 名 | UserCenterObj |
| tars_method | Tars RPC 方法名 | BindDevice |
| timeout_ms | 超时时间（毫秒） | 3000 |
| auth_required | 是否需要鉴权 | true |
| audit_required | 是否需要审计 | true |

### 5.4 调用链路示例

以"绑定设备"为例，完整调用链路如下：

1. **App 发起请求**：用户在 App 上点击"绑定设备"，App 发起 HTTPS JSON 请求到 API Gateway。
2. **API Gateway 处理**：API Gateway 进行基础鉴权、限流，通过 grpc-gateway 将 HTTPS JSON 转换为 Protobuf 请求。
3. **Proto Gateway 解析**：Proto Gateway 解析 Protobuf 请求，读取 package（cairobot.user.v1）、service（UserCenterService）、method（BindDevice）。
4. **路由匹配**：Proto Gateway 查找 routes.yaml 配置，匹配到目标 TarsGo servant：CaiRobot.UserCenterServer.UserCenterObj。
5. **TarsCloud 路由**：TarsCloud 路由层通过服务发现找到可用的 UserCenterServer 实例，进行负载均衡选择。
6. **TarsGo 调用**：TarsGo client 调用 CaiRobot.UserCenterServer.UserCenterObj.BindDevice。
7. **业务逻辑执行**：UserCenterServer 执行设备绑定业务逻辑，可能调用 AuthServer 鉴权，调用 AuditServer 记录审计日志。
8. **返回响应**：UserCenterServer 返回 Tars 响应，TarsCloud 路由层将其转换为 Protobuf Result。
9. **Gateway 转换响应**：API Gateway 将 Protobuf 响应转换为 HTTPS JSON 返回给 App。
10. **结果展示**：App 展示设备绑定成功的结果。

### 5.5 错误码转换规范

内部 Tars 错误必须转换为统一的 Protobuf Result 响应，不允许直接暴露底层 Tars 错误细节给外部调用方。

错误码映射表：

| Protobuf Result.code | 说明 | 对应 Tars 错误 |
|---------------------|------|---------------|
| 10200 | 成功 | TARS-SUCCESS |
| 10400 | 请求参数错误 | TARS-PARAMERR |
| 10401 | 未授权 | TARS-AUTHERR |
| 10403 | 无权限 | TARS-PERMERR |
| 10404 | 资源不存在 | TARS-NOTFOUND |
| 10409 | 冲突 | TARS-CONFLICT |
| 10429 | 请求过多（限流） | TARS-OVERLOAD |
| 10500 | 内部错误 | TARS-SERVERERR |
| 10503 | 服务不可用 | TARS-SERVNOTFOUND |
| 10504 | 超时 | TARS-TIMEOUT |

错误响应格式（基于 proto/base/result.proto）：

```protobuf
message Result {
  int32 code = 1;      // 响应码，成功建议使用 10200
  string message = 2;  // 响应消息
}

message ErrorDetail {
  string field = 1;    // 字段名
  string message = 2;  // 错误消息
  string code = 3;     // 错误码
}

message ValidationResult {
  repeated ErrorDetail errors = 1;  // 错误列表
}
```

错误处理原则：
1. **外部友好**：外部返回的错误信息应该是用户友好的，不暴露技术细节。
2. **内部详细**：内部日志应该记录详细的错误信息，便于排查问题。
3. **统一格式**：所有错误响应都应该遵循统一的 Result 格式。
4. **可追溯**：错误响应应该包含 trace-id，便于问题追踪。

## 6. 核心模块详解

### 6.1 协议层（proto/）

#### 6.1.1 基础协议

| 协议文件 | Message | 说明 |
|---------|---------|------|
| base/message.proto | MessagePacket | 网关统一入口报文 |
| base/result.proto | Result, PageInfo, ErrorDetail | 通用返回结构 |
| base/health.proto | ServiceHealthCheckRequest/Response | 健康检查 |

#### 6.1.2 MessagePacket结构

```protobuf
message MessagePacket {
  int32 maxType = 1;              // 协议大类
  int32 minType = 2;              // 协议小类
  map<string, string> extend = 3;  // 透传上下文
  Platform platform = 4;           // 平台类型
  bytes data = 5;                  // 业务协议包
}
```

路由规则：
1. 解析maxType和minType
2. 根据max+min查找对应业务协议
3. 解析data为具体业务Message

#### 6.1.3 Result通用返回

```protobuf
message Result {
  int32 code = 1;      // 响应码，成功10200
  string message = 2;   // 响应消息
}
```

### 6.2 服务层（services/ & tars/go/）

#### 6.2.1 API Gateway

| 类/函数 | 职责 | 说明 |
|---------|------|------|
| Router | 路由分发 | 根据maxType路由到对应服务 |
| AuthMiddleware | 鉴权中间件 | Token/API Key验证 |
| RateLimiter | 限流器 | 请求频率限制 |

#### 5.2.2 Provider Admin Service

| 模块 | 职责 |
|------|------|
| account/ | 服务商账号管理 |
| tenant/ | 租户/渠道管理 |
| device/ | 设备批次管理 |
| ticket/ | 售后工单管理 |
| analytics/ | 运营数据看板 |

#### 5.2.3 User Center Service

| 模块 | 职责 |
|------|------|
| user/ | 用户账号管理 |
| family/ | 家庭空间管理 |
| child/ | 孩子档案管理 |
| binding/ | 设备绑定管理 |
| learning/ | 学习会话记录 |

#### 5.2.4 Device Gateway Service

| 模块 | 职责 |
|------|------|
| connection/ | 设备连接管理 |
| control/ | 设备指令下发 |
| status/ | 设备状态上报 |
| event/ | 设备事件推送 |

#### 6.2.5 Auth Service

| 模块 | 职责 |
|------|------|
| token/ | Token生成与验证 |
| oauth/ | OAuth认证 |
| session/ | 会话管理 |
| permission/ | 权限管理 |

#### 6.2.6 Audit Service

| 模块 | 职责 |
|------|------|
| logger/ | 操作日志记录 |
| tracker/ | 审计追踪 |
| reporter/ | 审计报告生成 |

### 6.3 AI服务层（ai/service/）

#### 6.3.1 核心模块

| 模块 | 职责 | 关键类/函数 |
|------|------|------------|
| classifier | 意图分类 | IntentClassifier.classify() |
| prompt_rewriter | 提示词改写 | PromptRewriter.rewrite() |
| answer_reviewer | 回答审核 | AnswerReviewer.review() |
| model_gateway | 模型网关 | ModelGateway.infer() |
| safety | 安全策略 | SafetyPolicy.check() |

#### 6.3.2 意图分类器

```python
class IntentClassifier:
    def classify(self, user_input: str, context: dict) -> IntentType:
        """
        输入: 用户输入文本和学习上下文
        输出: 意图类型枚举
        支持的意图:
        - HOMEWORK_EXPLANATION: 作业讲解
        - VOCABULARY_HELP: 词汇帮助
        - READING_SUPPORT: 阅读支持
        - OPEN_GAME: 打开游戏
        - PARENT_CONTROL: 家长控制
        """
```

#### 6.3.3 提示词改写器

```python
class PromptRewriter:
    def rewrite(self, user_input: str, intent: IntentType, 
                safety_policy: SafetyPolicy) -> str:
        """
        输入: 用户输入、识别意图、安全策略
        输出: 改写后的安全提示词
        """
```

#### 6.3.4 回答审核器

```python
class AnswerReviewer:
    def review(self, answer: str, context: dict) -> ReviewResult:
        """
        输入: AI回答和学习上下文
        输出: 审核结果(通过/过滤/人工复核)
        """
```

## 7. API与通信协议

### 7.1 协议编号规范

| max范围 | 模块 | 说明 |
|--------|------|------|
| 1000-1999 | 通用基础协议 | 通用定义 |
| 2000-2999 | 系统基础能力 | 健康检查、网关基础 |
| 3000-3999 | 认证与权限 | 认证服务 |
| 4000-4999 | 服务商后台 | Provider Admin |
| 5000-5999 | 终端用户中台 | User Center |
| 6000-6999 | 开放平台 | Open Platform |
| 7000-7999 | AI服务 | AI Service |
| 8000-8999 | 设备通信 | Device Gateway |
| 9000-9999 | App/Web前端 | 前端交互协议 |

### 7.2 已注册协议

| max | min | 报文类型 | Message | 说明 |
|-----|-----|---------|---------|------|
| 2100 | 2097 | Request | ServiceHealthCheckRequest | 健康检查请求 |
| 2100 | 2098 | Response | ServiceHealthCheckResponse | 健康检查响应 |

### 7.3 服务间通信

```protobuf
// 服务定义示例
service AIService {
  rpc ClassifyIntent(IntentRequest) returns (IntentResponse);
  rpc RewritePrompt(PromptRewriteRequest) returns (PromptRewriteResponse);
  rpc ReviewAnswer(AnswerReviewRequest) returns (AnswerReviewResponse);
}
```

### 7.4 开放平台对外API

开放平台通过grpc-gateway转换为HTTPS JSON对外暴露：

```
POST /api/v1/intent/classify
POST /api/v1/prompt/rewrite
POST /api/v1/answer/review
```

## 8. 工程规范

### 8.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 变量/参数 | lowerCamelCase | userId, requestBody |
| 常量 | UPPER_SNAKE_CASE | MAX_RETRY_COUNT |
| 类/接口 | UpperCamelCase | UserService, DeviceController |
| 布尔变量 | is/has/can/should开头 | isActive, hasPermission |
| 方法 | lowerCamelCase | getUserById, createOrder |

### 8.2 文件规模限制

| 项目 | 推荐上限 | 绝对上限 |
|------|---------|----------|
| 单文件 | 300行 | 500行 |
| 单类 | 200行 | - |
| 单方法 | 30行 | 50行 |
| 单PR | 10个文件 | - |
| 单PR代码行 | 300行 | - |

### 8.3 分支策略

| 分支 | 说明 |
|------|------|
| main | 稳定分支，只合并评审通过的代码 |
| dev | 集成分支，日常开发合并目标 |
| feature/* | 功能开发分支 |
| fix/* | 缺陷修复分支 |
| docs/* | 文档变更分支 |
| test/* | 测试相关分支 |

### 8.4 提交规范

```bash
# 格式
type(scope): 中文说明

# 示例
docs(prd): 添加MVP总纲
feat(ai): 实现基础意图分类
fix(app): 修复设备断连状态处理
test(services): 添加认证服务单元测试
```

### 8.5 TDD流程

```
需求红 → 协议红 → 测试红 → 实现绿 → 报告绿 → CI绿 → 重构 → 沉淀
```

## 9. 开发与运行

### 9.1 前置要求

| 组件 | 版本要求 |
|------|---------|
| Go | 1.21+ |
| Python | 3.10+ |
| Node.js | 18+ |
| Protocol Buffers | 3.x |
| Docker | 20+ (可选) |

### 8.2 本地开发

#### 8.2.1 Golang服务开发

```bash
# 进入服务目录
cd services/[service-name]

# 安装依赖
go mod download

# 生成Protobuf代码
./scripts/gen_proto.sh

# 运行服务
go run cmd/server/main.go

# 运行测试
go test ./...

# 本地验证
python3 scripts/ci/check_required_docs.py
```

#### 9.2.2 AI服务开发

```bash
# 进入AI服务目录
cd ai/service

# 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/Mac
# 或 venv\Scripts\activate  # Windows

# 安装依赖
pip install -r requirements.txt

# 运行服务
python -m app.api.grpc_server

# 运行测试
pytest tests/
```

#### 9.2.3 前端开发

```bash
# 进入前端项目
cd web/[project-name]

# 安装依赖
npm install

# 启动开发服务器
npm start

# 运行测试
npm test

# 构建生产版本
npm run build
```

### 9.3 CI/CD

GitHub Actions自动执行以下检查：

| Job | 检查内容 |
|-----|---------|
| docs-check | 关键文档存在性 |
| proto-check | 协议编号唯一性 |
| go-test | Golang单元测试 |
| python-test | Python单元测试 |
| web-test | ReactJS测试 |
| admin-web-test | AdminWeb测试 |
| report-check | 报告存在性 |

### 9.4 常用脚本

```bash
# 检查文档完整性
python3 scripts/ci/check_required_docs.py

# 检查协议编号
python3 scripts/ci/check_proto_registry.py

# 检查报告
python3 scripts/ci/check_reports.py
```

## 10. 依赖关系

### 10.1 服务依赖图

```
API Gateway
├── Auth Service
├── Provider Admin Service
│   └── Auth Service
├── User Center Service
│   ├── Auth Service
│   └── AI Service
├── Open Platform Service
│   ├── Auth Service
│   ├── User Center Service
│   └── AI Service
└── Device Gateway Service
    ├── Auth Service
    ├── Audit Service
    └── AI Service

AI Service (独立部署)
└── 外部模型API

Audit Service (被多个服务调用)
```

### 10.2 外部依赖

| 依赖 | 用途 | 说明 |
|------|------|------|
| Protocol Buffers | 协议定义 | 核心依赖 |
| grpc-go | gRPC通信 | Golang服务 |
| grpcio | gRPC通信 | Python服务 |
| React | 前端框架 | Web项目 |
| TensorFlow/PyTorch | AI模型 | AI服务(可选) |

## 11. 相关文档索引

### 11.1 核心规范

| 文档 | 路径 | 说明 |
|------|------|------|
| AGENTS.md | AGENTS.md | 工程协作总纲 |
| coding.md | .trae/rules/coding.md | 编码规范 |
| git.md | .trae/rules/git.md | Git工作流 |
| tdd.md | .trae/rules/tdd.md | TDD规范 |
| testing.md | .trae/rules/testing.md | 测试规范 |
| reporting.md | .trae/rules/reporting.md | 汇报规范 |
| review.md | .trae/rules/review.md | 评审规范 |
| docs.md | .trae/rules/docs.md | 文档规范 |

### 11.2 产品文档

| 文档 | 路径 |
|------|------|
| PRD-00-MVP总纲 | docs/prd/PRD-00-MVP总纲.md |
| PRD-01-服务商后台系统 | docs/prd/PRD-01-服务商后台系统.md |
| PRD-02-终端用户中台系统 | docs/prd/PRD-02-终端用户中台系统.md |
| PRD-03-开放平台API | docs/prd/PRD-03-开放平台API.md |
| PRD-04-AI服务系统 | docs/prd/PRD-04-AI服务系统.md |
| PRD-05-App前端系统 | docs/prd/PRD-05-App前端系统.md |
| PRD-06-设备通信与协议 | docs/prd/PRD-06-设备通信与协议.md |

### 11.3 架构文档

| 文档 | 路径 |
|------|------|
| ADR-0001-总体系统架构 | docs/adr/ADR-0001-总体系统架构.md |
| ADR-0002-后端使用Golang | docs/adr/ADR-0002-后端使用Golang.md |
| ADR-0003-服务协议使用Protobuf | docs/adr/ADR-0003-服务协议使用Protobuf.md |
| ADR-0004-AI服务使用Python | docs/adr/ADR-0004-AI服务使用Python.md |
| ADR-0005-App前端使用ReactJS | docs/adr/ADR-0005-App前端使用ReactJS.md |
| ADR-0006-开放平台API边界 | docs/adr/ADR-0006-开放平台API边界.md |
| ADR-0007-服务商后台与用户中台边界 | docs/adr/ADR-0007-服务商后台与用户中台边界.md |
| ADR-0008-使用TarsCloud作为Protobuf到TarsGo的内部路由转发层 | docs/adr/ADR-0008-use-tarscloud-routing-layer.md |

### 11.4 协议文档

| 文档 | 路径 |
|------|------|
| protobuf规范 | docs/api/protobuf规范.md |
| gRPC接口规范 | docs/api/gRPC接口规范.md |
| 协议编号注册表 | docs/api/协议编号注册表.md |
| OpenAPI规范 | docs/api/OpenAPI规范.md |
| openapi-protobuf映射规范 | docs/api/openapi-protobuf映射规范.md |

## 12. 变更日志

| 日期 | 版本 | 变更内容 | 作者 |
|------|------|---------|------|
| 2026-05-18 | 1.1.0 | 新增TarsCloud路由层设计，目录结构调整 | Trae |
| 2026-05-18 | 1.0.0 | 初始版本 | Trae |
