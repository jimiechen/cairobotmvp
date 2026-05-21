# CODE-WIKI

> **本文档是 CaiRobot MVP 项目的代码架构知识库（Index 层）。**
> **记录项目整体架构、主要模块职责、关键类与函数、依赖关系及运行方式。**
> **最后更新：2026-05-21**

---

## 1. 项目概述

CaiRobot MVP 是一个多系统、多技术栈的微服务架构项目，采用**单网关 + MessagePacket + TarsCloud/TarsGo** 的统一架构。

### 1.1 核心设计原则

```text
Protobuf 定义协议身份和业务字段。
MessagePacket 定义单网关入口。
routes.yaml 定义 maxType/minType 到 Tars 目标的映射。
Tars IDL 定义内部服务方法，方法签名统一为 bytes 接口。
TarsCloud 负责服务治理。
TarsGo 服务负责业务逻辑。
```

### 1.2 系统组成

| # | 系统 | 技术栈 | 职责 |
|---|---|---|---|
| 1 | 服务商后台系统 | Golang + ReactJS | 设备、服务、工单、租户管理 |
| 2 | 终端用户中台系统 | Golang + ReactJS | 家庭空间、孩子档案、设备绑定 |
| 3 | 开放平台 API | Golang + Protobuf | 认证、授权、设备状态、Webhook |
| 4 | AI 服务系统 | Python | 意图分类、提示词改写、审核、OCR |
| 5 | Golang 后端服务 | Golang + TarsGo | 服务商后台、用户中台、设备网关 |
| 6 | ReactJS 前端系统 | ReactJS + TypeScript | 服务商前端、用户中台前端、App/H5 |
| 7 | Protobuf 协议层 | Protobuf | 服务间通信统一接口定义 |

### 1.3 当前阶段

**S0：规则与目录骨架** 阶段。

- 工程规范框架已建立
- HelloWorld 验收通过
- Gateway 双模式骨架完成
- 多语言 monorepo 布局已建立

---

## 2. 单网关总体架构

### 2.1 请求流转图

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

### 2.2 架构分层

| 层级 | 目录/模块 | 职责 |
|------|-----------|------|
| 接入层 | `gateway/proto-gateway` | HTTP 入口、路由、调用分发 |
| 路由层 | `internal/router` | maxType/minType → Route 映射 |
| 调用层 | `internal/tarsclient` | TarsInvoker 接口（Local/TarsGo） |
| 适配层 | `tars/system/adapter` | LocalHandler → ModuleInvokeFunc 适配 |
| 业务模块层 | `modules/*` | 独立 go.mod，可单独构建测试 |
| 公共库层 | `common-lib` | 错误码、类型定义、常量 |

---

## 3. MessagePacket 入口协议

MessagePacket 是 CaiRobot MVP **对外单网关唯一业务入口报文**。

```protobuf
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

### 3.1 extend map 上下文透传规范

| key | 说明 | 来源 |
|---|---|---|
| traceId | 链路追踪 ID | 客户端或 Gateway |
| requestId | 请求唯一 ID | 客户端或 Gateway |
| token | Token | MessagePacket.extend 或 Header |
| caller | 调用方 | Gateway |
| clientIp | 客户端 IP | Gateway |
| userId | 用户 ID | AuthServer 校验结果 |
| tenantId | 租户 / 服务商 ID | AuthServer 校验结果 |

---

## 4. Protobuf 协议编号规范

- 协议编号 `max + min` 是接口报文的唯一身份
- 每个业务 Request/Response message 内部必须声明 enum Type
- `Type.max` 表示协议大类，`Type.min` 表示协议小类
- `max + min` 组合在全仓库内必须唯一
- Request 和 Response 应分别拥有独立编号

### 4.1 编号范围

| max 范围 | 模块 |
|---:|---|
| 1000-1999 | 通用基础协议 |
| 2000-2999 | 系统、健康检查、网关基础能力 |
| 3000-3999 | 认证与权限 |
| 4000-4999 | 服务商后台系统 |
| 5000-5999 | 终端用户中台系统 |
| 6000-6999 | 开放平台 API |
| 7000-7999 | AI 服务系统 |
| 8000-8999 | 设备通信与设备网关 |
| 9000-9999 | App、Web、前端交互协议 |

### 4.2 已注册协议

| 协议 | max | min | 类型 | 说明 |
|---|---|---|---|---|
| ServiceHealthCheckRequest | 2100 | 2097 | Request | 健康检查请求 |
| ServiceHealthCheckResponse | 2100 | 2098 | Response | 健康检查响应 |
| HelloWorldRequest | 2100 | 2101 | Request | HelloWorld 请求 |
| HelloWorldResponse | 2100 | 2102 | Response | HelloWorld 响应 |

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

---

## 6. TarsCloud/TarsGo 内部服务治理

### 6.1 命名规范

| 层级 | 命名规则 | 示例 |
|---|---|---|
| App | CaiRobot | CaiRobot |
| Server | XxxServer | UserCenterServer |
| Servant | XxxObj | UserCenterObj |
| module | CaiRobotXxxApp | CaiRobotUserCenterApp |
| interface | XxxObj | UserCenterObj |

### 6.2 Tars 标准 bytes 接口

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

- `request`：Protobuf Request message 序列化 bytes
- `extend`：上下文透传 map
- `response`：Protobuf Response message 序列化 bytes
- `return int`：项目统一状态码

### 6.3 Tars return 与 Result.code 分层规则

- Tars return 表示内部 Tars 方法调用的处理状态
- Protobuf Response message 内部的 Result.code 表示业务响应状态
- Gateway 对外返回 MessagePacket 时，优先使用业务 Response.Result.code
- Tars 框架异常由 Gateway 映射为项目统一错误码

### 6.4 健康检查接口

每个 Tars interface 必须包含：

```tars
int Health(vector<byte> request, map<string,string> extend, out vector<byte> response);
int HealthCheck(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

---

## 7. Gateway 调用 Tars 完整流程

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

## 8. Gateway 运行模式

Gateway 基于 TarsCloud/TarsGo v1.4.6 技术基线，支持两种**部署拓扑**：

### 8.1 单体部署模式（默认）

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

### 8.2 微服务部署模式

```bash
GATEWAY_INVOKER_MODE=tars
```

- 正式部署或集成环境使用
- **使用 TarsGo 框架运行**（与单体模式相同的技术基线）
- 连接远程 TarsCloud 注册中心，通过 **TarsGoInvoker**（远程 TarsGo client）调用独立部署的 TarsCloud servant
- GatewayServer、SystemServer 等独立部署为不同进程
- 当前 **TarsGoInvoker 远程调用尚未实现**，启动会报错（S1 阶段）

### 8.3 TarsInvoker 接口

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

## 9. 关键类与函数说明

### 9.1 Gateway 层（`gateway/proto-gateway`）

#### GatewayServer

```go
// GatewayServer TarsGo HTTP Servant，作为 proto-gateway 的唯一 HTTP 入口
// 通过 tars.AddHttpServant 注册到 TarsGo 框架
type GatewayServer struct {
    routeTable *router.RouteTable
    invoker    tarsclient.TarsInvoker
    mode       string
}
```

- `ServeHTTP(w, r)`：处理 POST /api/hello 请求，完整链路：反序列化 → 路由查找 → Tars 调用 → 响应封装

#### RouteTable

```go
// RouteTable 负责根据 maxType:minType 查找路由
type RouteTable struct {
    routes map[string]config.Route
}
```

- `NewRouteTable(cfg)`：从配置创建路由表
- `FindRoute(maxType, minType)`：查找路由，返回 Route 和是否命中

#### TarsInvoker / LocalInvoker / TarsGoInvoker

```go
type TarsInvoker interface {
    Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}
```

- `LocalInvoker`：单体模式，本进程 handler 映射表
- `TarsGoInvoker`：微服务模式，远程调用（S1 未实现）
- `RegisterModuleHandlers(invoker)`：注册 HealthCheck + HelloWorld 模块

#### MessagePacket 适配器

```go
type MessagePacket = pb.MessagePacket

func DeserializeMessagePacket(data []byte) (*MessagePacket, error)
func SerializeMessagePacket(packet *MessagePacket) ([]byte, error)
func BuildErrorPacket(req *MessagePacket, code int32, message string) *MessagePacket
func BuildResponsePacket(req *MessagePacket, responseMax, responseMin int32, data []byte, code int32) *MessagePacket
func BuildTarsExtend(req *MessagePacket, routeKey, requestProto, responseProto string, authRequired, auditRequired bool) map[string]string
```

#### Route 配置

```go
type Route struct {
    RequestMax       int32  `yaml:"request_max"`
    RequestMin       int32  `yaml:"request_min"`
    RouteKey         string `yaml:"route_key"`
    CommandName      string `yaml:"command_name"`
    RequestProto     string `yaml:"request_proto"`
    ResponseMax      int32  `yaml:"response_max"`
    ResponseMin      int32  `yaml:"response_min"`
    ResponseProto    string `yaml:"response_proto"`
    TarsApp          string `yaml:"tars_app"`
    TarsServer       string `yaml:"tars_server"`
    TarsServant      string `yaml:"tars_servant"`
    TarsModule       string `yaml:"tars_module"`
    TarsInterface    string `yaml:"tars_interface"`
    TarsMethod       string `yaml:"tars_method"`
    TarsRequestType  string `yaml:"tars_request_type"`
    TarsResponseType string `yaml:"tars_response_type"`
    TimeoutMs        int32  `yaml:"timeout_ms"`
    AuthRequired     bool   `yaml:"auth_required"`
    AuditRequired    bool   `yaml:"audit_required"`
}
```

### 9.2 业务模块层（`modules/*`）

#### Hello 模块

```go
type HelloService interface {
    SayHello(ctx context.Context, request []byte) ([]byte, error)
}

type Service struct{}
func NewService() *Service
func (s *Service) SayHello(ctx context.Context, request []byte) ([]byte, error)
```

- 反序列化 `HelloWorldRequest` → 生成问候消息 → 序列化 `HelloWorldResponse`
- 默认问候名称为 "World"

#### Health 模块

```go
type HealthService interface {
    Check(ctx context.Context, request []byte) ([]byte, error)
}

type Service struct{}
func NewService() *Service
func (s *Service) Check(ctx context.Context, request []byte) ([]byte, error)
```

- 反序列化 `ServiceHealthCheckRequest` → 返回 OK 状态 → 序列化 `ServiceHealthCheckResponse`

### 9.3 Tars 适配层（`tars/system/adapter`）

#### SystemAdapter

```go
type SystemAdapter struct {
    helloSvc  hello.HelloService
    healthSvc health.HealthService
}

func NewSystemAdapter() *SystemAdapter
func (a *SystemAdapter) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error)
```

- 将 Tars servant 接口适配到模块化业务服务
- 根据 extend["method"] 分发到 HealthCheck 或 HelloWorld
- 返回码转换：成功→10200，失败→10500/10404

### 9.4 公共库（`common-lib`）

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

---

## 10. 目录结构

### 10.1 文档目录

```text
docs/
  api/
    protobuf规范.md
    tars规范.md
    http-gateway规范.md
    openapi-protobuf映射规范.md
    OpenAPI规范.md
    协议编号注册表.md
    Webhook规范.md
    gRPC接口规范.md
  adr/
    ADR-0001-总体系统架构.md
    ADR-0003-服务协议使用Protobuf.md
    ADR-0005-App前端使用ReactJS.md
    ADR-0008-use-tarscloud-routing-layer.md
    ADR-0012-多语言Monorepo目录布局.md
    ADR-0013-Makefile工程入口与规则强制执行.md
    ADR-0014-message-packet-data-format-protobuf-bytes.md
  prd/
    PRD-00-MVP总纲.md
    PRD-01-核心用户流程.md
    PRD-01-服务商后台系统.md
    PRD-02-终端用户中台系统.md
    PRD-02-工程交付与验证规范.md
    PRD-03-开放平台API.md
    PRD-04-AI服务系统.md
    PRD-05-App前端系统.md
    PRD-06-设备通信与协议.md
    PRD-09-HelloWorld与HealthCheck验收规范.md
  wiki/
    LLM-WIKI.md          # Wiki 总索引
    CODE-WIKI.md         # 代码架构知识（本文件）
    工程规范索引.md
    ADR索引.md
    PRD索引.md
    测试索引.md
    Bug索引.md
    决策记录.md
    每日蒸馏索引.md
    任务索引.md
    README.md
  services/
    api-gateway.md
    auth-service.md
    audit-service.md
    device-gateway.md
    open-platform.md
    provider-admin.md
    user-center.md
```

### 10.2 Go 语言资产

```text
go/
  go.work                                    # Workspace 总控
  common-lib/                                # 公共库（错误码、类型定义）
    codes.go
    codes_test.go
    types.go
    go.mod
  modules/                                   # 业务模块（独立 go.mod）
    hello/                                    # Hello 模块
      service.go
      service_test.go
      go.mod
    health/                                   # Health 模块
      service.go
      service_test.go
      go.mod
  gateway/
    proto-gateway/                            # 网关服务
      go.mod
      cmd/
        server/
          main.go                             # TarsGo 入口
        testclient/
          main.go                             # E2E 测试客户端
      configs/
        gateway/
          gateway.local.conf                  # 本地配置
          routes.yaml                         # 路由表
      internal/
        config/
          routes.go                           # Route 配置结构 + 加载校验
          routes_test.go
        router/
          route_table.go                      # RouteTable 路由查找
          route_table_test.go
        server/
          http_server.go                      # GatewayServer HTTP 处理
          http_server_test.go
          e2e_modules_test.go                 # E2E 全链路测试
        tarsclient/
          invoker.go                          # TarsInvoker + LocalInvoker + TarsGoInvoker
          invoker_test.go
          module_handler_test.go
        adapter/
          message_packet.go                   # MessagePacket 序列化/反序列化
          message_packet_test.go
  tars/
    system/                                   # System Tars 服务
      go.mod
      cmd/
        main.go
      internal/
        service/
          system_service.go                   # @deprecated 标记废弃
          system_service_test.go
      adapter/
        system_adapter.go                     # SystemAdapter 适配器
        system_adapter_test.go
        deprecated/
          local_handler.go                    # 旧 LocalHandler 实现
          local_handler_test.go
  services/                                   # 旧服务目录（保留兼容）
    hello_service_test.go
    main.go
    go.mod
  third_party/
    TarsGo/
      README.md
      TarsGo-1.4.6/                           # TarsCloud/TarsGo v1.4.6 源码
```

### 10.3 Python 语言资产

```text
python/
  Makefile
  ai/
    hello/
      pyproject.toml                          # FastAPI + uvicorn
      requirements.txt
      README.md
    service/
    README.md
```

### 10.4 TypeScript 语言资产

```text
typescript/
  Makefile
  web/
    package.json                              # React + Vite + Vitest
    src/
    README.md
  admin-web/
    README.md
```

### 10.5 Protobuf 协议目录

```text
proto/
  base/
    hello.proto                               # HelloWorld 协议
    health.proto                              # HealthCheck 协议
    message.proto                             # MessagePacket 协议
    result.proto                              # Result 通用结果
  generated/
    go/                                       # Go 生成代码
    ts/                                       # TypeScript 生成代码
```

### 10.6 Tars 协议目录

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

### 10.7 部署目录

```text
deploy/
  tarscloud/
    README.md
    configs/
    templates/
```

---

## 11. Go Workspace 管理

Go Workspace 位于 `go/go.work`，管理以下模块：

```text
go 1.23

use (
    ../proto/generated/go
    ./common-lib
    ./modules/hello
    ./modules/health
    ./gateway/proto-gateway
    ./tars/system
)
```

### 11.1 Module Path 规范

所有 Go module 使用统一 path 前缀：

```text
github.com/jimiechen/mineplanet/go/...
```

| 模块 | Path |
|---|---|
| gateway/proto-gateway | github.com/jimiechen/mineplanet/go/gateway/proto-gateway |
| tars/system | github.com/jimiechen/mineplanet/go/tars/system |
| common-lib | github.com/jimiechen/mineplanet/go/common-lib |
| modules/hello | github.com/jimiechen/mineplanet/go/modules/hello |
| modules/health | github.com/jimiechen/mineplanet/go/modules/health |

### 11.2 依赖关系图

```text
gateway/proto-gateway
  ├── common-lib
  ├── modules/hello
  ├── modules/health
  └── TarsGo v1.4.6 (third_party)

tars/system
  ├── common-lib
  ├── modules/hello
  ├── modules/health
  └── proto/generated/go

modules/hello
  ├── common-lib
  └── proto/generated/go

modules/health
  ├── common-lib
  └── proto/generated/go
```

---

## 12. 项目运行方式

### 12.1 环境要求

- Go 1.23+
- Python 3.11+
- Node.js 20+
- make
- protoc（仅本地生成 Protobuf 代码时需要）

### 12.2 初始化环境

```bash
make bootstrap          # 检查工具链 + 安装所有依赖
```

### 12.3 编译 Gateway

```bash
make gateway-build      # 编译 proto-gateway 二进制
```

### 12.4 启动 Gateway（单体模式）

```bash
make gateway-start      # 启动 proto-gateway，监听 :8080
```

或手动：

```bash
cd go/gateway/proto-gateway
GATEWAY_INVOKER_MODE=local go run ./cmd/server/main.go
```

### 12.5 运行测试

```bash
make test               # 运行所有测试（单元 + 集成）
make unit               # 运行单元测试
make integration        # 运行集成测试
make go-all             # Go 全量测试
make common-lib-test    # 测试 common-lib
make modules-test       # 测试业务模块
make tars-test          # 测试 Tars 层
make gateway-e2e        # Gateway E2E 链路测试
make gateway-smoke      # Gateway 冒烟测试
make gateway-verify     # Gateway 完整验证
```

### 12.6 生成 Protobuf 代码

```bash
make proto              # 生成所有语言的 Protobuf 代码
```

### 12.7 CI 检查

```bash
make ci                 # 完整 CI 检查（本地等价于 GitHub Actions）
make docs               # 检查文档完整性
make rules              # 执行工程规范检查
make proto-check        # 校验 Protobuf 生成代码
make lint               # 运行 Lint 检查
make coverage           # 生成覆盖率报告
```

### 12.8 其他常用命令

```bash
make help               # 显示帮助
make build              # 构建所有语言的可执行文件
make package            # 打包发布产物
make clean              # 清理构建产物
```

---

## 13. Makefile 工程入口

### 13.1 三层架构

```text
Makefile（根目录总控，16 个 target）
├── go/Makefile
├── typescript/Makefile
├── python/Makefile
└── scripts/
    ├── ci/          # check_*.py / check_*.sh
    ├── proto/       # generate-*.sh
    └── coverage/    # *_coverage.sh
```

### 13.2 根目录 Makefile Target 列表

| Target | 功能 |
|---|---|
| `help` | 显示帮助信息 |
| `bootstrap` | 初始化开发环境 |
| `proto` | 生成 Protobuf 代码 |
| `proto-check` | 校验 Protobuf 代码 |
| `lint` | 运行 Lint 检查 |
| `test` | 运行所有测试 |
| `unit` | 运行单元测试 |
| `integration` | 运行集成测试 |
| `coverage` | 生成覆盖率报告 |
| `build` | 构建可执行文件 |
| `package` | 打包发布产物 |
| `docs` | 检查文档完整性 |
| `rules` | 执行工程规范检查 |
| `testcase-check` | 检查测试用例注册表 |
| `comment-check` | 检查中文注释 |
| `ci` | 完整 CI 检查 |
| `clean` | 清理构建产物 |

---

## 14. 测试与校验要求

### 14.1 Gateway 测试

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

### 14.2 System 模块测试

| 测试项 | 要求 |
|---|---|
| SystemService 业务逻辑测试 | HealthCheck / HelloWorld 返回正确 |
| LocalHandler bytes 适配测试 | 请求 bytes 正确反序列化，响应正确序列化 |
| LocalHandler 错误处理测试 | 非法 bytes / 未知 maxType/minType 返回错误 |

---

## 15. GitHub Actions CI

### 15.1 Workflow 文件

`.github/workflows/ci.yml`

### 15.2 CI Jobs

| Job | 说明 |
|---|---|
| `ci-full` | 完整 CI 检查（`make ci`） |
| `docs-check` | 检查关键文档是否存在 |
| `proto-check` | 检查协议编号唯一性和注册表同步 |
| `report-check` | 检查测试报告、日报、蒸馏、LLM Wiki |

### 15.3 触发条件

- Pull Request
- Push 到 main / dev / feature/** / fix/** / docs/** / test/**

---

## 16. 相关文档索引

- [protobuf规范.md](../api/protobuf规范.md)
- [tars规范.md](../api/tars规范.md)
- [http-gateway规范.md](../api/http-gateway规范.md)
- [openapi-protobuf映射规范.md](../api/openapi-protobuf映射规范.md)
- [OpenAPI规范.md](../api/OpenAPI规范.md)
- [协议编号注册表.md](../api/协议编号注册表.md)
- [ADR-0001-总体系统架构.md](../adr/ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)
- [ADR-0012-多语言Monorepo目录布局.md](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- [ADR-0013-Makefile工程入口与规则强制执行.md](../adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md)
- [ADR-0014-message-packet-data-format-protobuf-bytes.md](../adr/ADR-0014-message-packet-data-format-protobuf-bytes.md)

---

## 17. 变更日志

| 日期 | 变更内容 |
|---|---|
| 2026-05-18 | 根据 ADR-0008，内部核心服务主链路从 gRPC 调整为 TarsCloud/TarsGo；外部入口收敛为单网关 POST /api/hello；MessagePacket 成为唯一入口报文 |
| 2026-05-19 | 实现 Gateway 单体/微服务双模式骨架；System 模块独立存在；建立 Go Workspace；统一 TarsInvoker 接口；module path 标准化 |
| 2026-05-19 | 按 ADR-0012 重构多语言 monorepo 目录布局：Go 进 go/、Python 进 python/、TypeScript 进 typescript/；删除根目录 Makefile；go.work 移至 go/ |
| 2026-05-19 | 恢复根目录 Makefile，采用三层结构（总控 + 子 Makefile + scripts）；新增 16 个 target；新增 CI 规范检查脚本；建立测试用例注册表；新增中文注释规范 |
| 2026-05-20 | **架构口径修正**：local 模式不是"不使用 Tars"，而是 TarsGo 单体部署模式（monolith）；LocalInvoker 是本进程 TarsGo servant adapter，非绕过 Tars 的普通 Go 调用；proto-gateway 改造为基于 TarsGo HTTP 模块（TarsHttpMux / AddHttpServant / Run）的 TarsGo HTTP Servant；引入 TarsCloud/TarsGo v1.4.6 技术基线到 go/third_party/TarsGo/ |
| 2026-05-21 | **更新 CODE-WIKI**：补充完整的关键类与函数说明、依赖关系图、项目运行方式、目录结构详解 |
