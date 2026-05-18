# OpenAPI 与 Protobuf 映射规范

## 1. 基本原则

本项目同时使用 OpenAPI 和 Protobuf。

- Protobuf 是内部服务、网关报文和跨语言协议的事实来源。
- OpenAPI 是对外 HTTP/JSON 接口的说明文档。
- 每个 OpenAPI 接口必须映射到对应的 Protobuf Request 和 Response。
- 每个 Protobuf Request 和 Response 必须拥有唯一的 `max + min` 编号。

## 2. 单网关映射关系

当前架构中，OpenAPI 只描述单网关入口 `/api/hello` 的外层 MessagePacket 结构。具体业务请求和响应结构由 `maxType/minType` 对应的 Protobuf message 决定。

### 2.1 映射表

| OpenAPI Path | Method | 外层结构 | 业务 Request | Request max | Request min | 业务 Response | Response max | Response min |
|---|---|---|---|---:|---:|---|---:|---:|
| `/api/hello` | POST | `MessagePacket` | `ServiceHealthCheckRequest` | 2100 | 2097 | `ServiceHealthCheckResponse` | 2100 | 2098 |
| `/api/hello` | POST | `MessagePacket` | `HelloWorldRequest` | 2100 | 2101 | `HelloWorldResponse` | 2100 | 2102 |

> 注意：以上业务 Request/Response 仅为当前已登记协议示例。真实业务语义由 `MessagePacket.maxType + MessagePacket.minType` 决定，Gateway 根据 routes.yaml 路由到对应 TarsGo 服务。

### 2.2 历史映射调整

历史规划中 OpenAPI 曾映射到多 REST path（如 `/health/check`、`/api/v1/xxx`），该方案已根据 ADR-0008 调整为单网关模式。OpenAPI 不再为每个业务命令生成独立 REST path，只描述 `/api/hello` 的 MessagePacket 外层结构。

## 3. 规则

1. 不允许存在没有 Protobuf 映射的 OpenAPI 接口。
2. 不允许存在没有登记编号的 Protobuf 接口。
3. OpenAPI 的请求字段必须与 Protobuf Request 对齐。
4. OpenAPI 的响应字段必须与 Protobuf Response 对齐。
5. 如果 Protobuf 字段变更，必须同步更新 OpenAPI。
6. 如果 OpenAPI 变更，必须同步更新 Protobuf、协议注册表和测试用例。
7. OpenAPI 只描述 `/api/hello` 的 MessagePacket 外层结构，不描述独立业务 REST path。

## 4. 映射关系

### OpenAPI 接口 → Protobuf

每个 OpenAPI 接口的：
- Request Body 对应 `MessagePacket` 外层结构
- `MessagePacket.data` 对应具体业务 Protobuf Request Message
- Response Body 对应 `MessagePacket` 外层结构
- `MessagePacket.data` 对应具体业务 Protobuf Response Message

### Protobuf → OpenAPI

每个 Protobuf 接口对：
- Request Message 对应 `MessagePacket.data` 的业务 Schema
- Response Message 对应 `MessagePacket.data` 的业务 Schema
- 可通过工具从 Protobuf 生成 OpenAPI 的 MessagePacket 包装说明

## 5. 报文路由

网关通过 MessagePacket 中的 `maxType + minType` 定位到具体的业务协议：
1. 解析 MessagePacket
2. 提取 maxType 和 minType
3. 根据 `max + min` 找到对应的业务协议
4. 解析 data 字段为对应的业务 Protobuf Message
5. 根据 routes.yaml 路由到内部 TarsGo servant

## 6. 变更流程

协议变更必须按以下流程进行：
1. 更新 Protobuf 文件
2. 更新协议编号注册表
3. 更新 OpenAPI 映射
4. 更新测试用例
5. 更新相关文档
6. 提交 PR 评审

## 7. 相关文档

- [protobuf规范.md](protobuf规范.md)
- [OpenAPI规范.md](OpenAPI规范.md)
- [tars规范.md](tars规范.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)
