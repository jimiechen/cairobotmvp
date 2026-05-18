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

## 3. MessagePacket 入口协议

MessagePacket 是 CaiRobot MVP 对外单网关唯一业务入口报文。

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
| 6000-6999 | 开放平台 API |
| 7000-7999 | AI 服务系统 |
| 8000-8999 | 设备通信与设备网关 |
| 9000-9999 | App、Web、前端交互协议 |

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

## 6. TarsCloud/TarsGo 内部服务治理

- TarsCloud App 统一为 `CaiRobot`
- Server 使用 `XxxServer`
- Servant 使用 `XxxObj`
- IDL module 使用 `CaiRobotXxxApp`
- interface 使用 `XxxObj`
- 每个 servant 都暴露 Health / HealthCheck
- Tars 方法统一 bytes 签名

## 7. Tars 标准 bytes 接口

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

- `request`：Protobuf Request message 序列化 bytes
- `extend`：上下文透传 map
- `response`：Protobuf Response message 序列化 bytes
- `return int`：项目统一状态码

## 8. Tars App / Server / Servant / module / interface 命名规范

| 层级 | 命名规则 | 示例 |
|---|---|---|
| App | CaiRobot | CaiRobot |
| Server | XxxServer | UserCenterServer |
| Servant | XxxObj | UserCenterObj |
| module | CaiRobotXxxApp | CaiRobotUserCenterApp |
| interface | XxxObj | UserCenterObj |

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

## 10. Tars return 与 Result.code 分层规则

- Tars return 表示内部 Tars 方法调用的处理状态
- Protobuf Response message 内部的 Result.code 表示业务响应状态
- Gateway 对外返回 MessagePacket 时，优先使用业务 Response.Result.code
- Tars 框架异常由 Gateway 映射为项目统一错误码

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

## 12. Health / HealthCheck 基础接口

每个 Tars interface 必须包含：

```tars
int Health(vector<byte> request, map<string,string> extend, out vector<byte> response);
int HealthCheck(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

## 13. 目录结构说明

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

gateway/
  proto-gateway/
    README.md
    configs/
      routes.yaml
    internal/
      router/
        README.md
      adapter/
        README.md
      tarsclient/
        README.md

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
  go/
    system/
    auth/
    provider-admin/
    user-center/
    open-platform/
    ai-bridge/
    device-gateway/
    audit/

deploy/
  tarscloud/
    README.md
    configs/
    templates/
```

## 14. 依赖关系

- Protobuf 定义协议身份和业务字段
- MessagePacket 定义单网关入口
- routes.yaml 定义 maxType/minType 到 Tars 目标的映射
- Tars IDL 定义内部服务方法
- TarsCloud 负责服务治理
- TarsGo 服务负责业务逻辑

## 15. 开发与运行方式

当前为 S0 阶段，重点是文档、规范、目录骨架、配置示例和架构决策，不实现复杂业务逻辑。

## 16. 测试与校验要求

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

## 17. 相关文档索引

- [protobuf规范.md](../api/protobuf规范.md)
- [tars规范.md](../api/tars规范.md)
- [http-gateway规范.md](../api/http-gateway规范.md)
- [openapi-protobuf映射规范.md](../api/openapi-protobuf映射规范.md)
- [OpenAPI规范.md](../api/OpenAPI规范.md)
- [协议编号注册表.md](../api/协议编号注册表.md)
- [ADR-0001-总体系统架构.md](../adr/ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)

## 18. 变更日志

| 日期 | 变更内容 |
|---|---|
| 2026-05-18 | 根据 ADR-0008，内部核心服务主链路从 gRPC 调整为 TarsCloud/TarsGo；外部入口收敛为单网关 POST /api/hello；MessagePacket 成为唯一入口报文 |
