# HTTP Gateway 规范

## 1. 概述

CaiRobot MVP 对外业务入口统一为单网关模式。Gateway 负责接收外部请求、解析 MessagePacket、根据 maxType/minType 路由到内部 TarsGo 服务，并将响应封装为 MessagePacket 返回。

历史规划中曾使用 grpc-gateway 将 gRPC 服务转换为多 REST path 的 HTTP JSON 接口，该方案已根据 ADR-0008 调整为当前单网关架构。

## 2. 单网关入口

CaiRobot MVP 对外业务入口统一为：

```http
POST /api/hello
Content-Type: application/octet-stream
```

请求体为 `MessagePacket` 序列化后的二进制内容。业务语义由 `MessagePacket.maxType + MessagePacket.minType` 决定，业务数据由 `MessagePacket.data` 承载。

### 2.1 调试与文档展示

如果为了调试、文档展示或开发环境需要 JSON 示例，可以写成逻辑等价形式，但必须明确真实协议入口以 `MessagePacket` 为准：

```json
{
  "maxType": 2100,
  "minType": 2097,
  "extend": {
    "traceId": "trace-20260518-000001",
    "requestId": "req-20260518-000001",
    "token": "user-token",
    "caller": "app",
    "clientIp": "127.0.0.1"
  },
  "platform": "ANDROID",
  "data": "业务 Protobuf Request 序列化后的 bytes"
}
```

> 注意：以上 JSON 仅用于调试和文档展示，真实协议入口是 `application/octet-stream` 的 MessagePacket bytes。

## 3. 不再使用的多 REST path 模式

以下历史方案已废弃，不再作为业务入口：

```text
/api/user/bind-device
/api/device/send-command
/api/auth/login
/api/open-platform/create-webhook
/api/v1/xxx
```

所有业务命令统一收敛到 `POST /api/hello`，由 `MessagePacket.maxType + MessagePacket.minType` 决定具体业务语义。

## 4. Gateway 处理流程

1. 接收客户端请求 `POST /api/hello`
2. 反序列化请求体为 `MessagePacket`
3. 校验 `maxType/minType/data` 必填字段
4. 使用 `maxType:minType` 作为 route_key 查询 `routes.yaml`
5. 校验协议编号是否已登记到协议编号注册表
6. 根据 `request_proto` 将 `MessagePacket.data` 反序列化为业务 Protobuf Request
7. 执行基础参数校验
8. 根据 `auth_required` 判断是否需要鉴权
9. 合并 `MessagePacket.extend`、Header、鉴权结果、路由元信息，构造 Tars extend map
10. 将业务 Protobuf Request 重新序列化为 `vector<byte> request`
11. 根据路由目标调用 TarsGo servant
12. 接收 Tars return code 和 response bytes
13. 根据 `response_proto` 反序列化 response bytes
14. 使用 `response_max/response_min` 封装响应 MessagePacket
15. 在响应 `MessagePacket.extend` 中写入 `code/message/traceId/requestId`
16. 返回客户端

## 5. 响应格式

响应同样使用 `MessagePacket`，其中：

- `maxType`：响应 Protobuf Response message 的 Type.max
- `minType`：响应 Protobuf Response message 的 Type.min
- `extend`：回传 traceId、requestId、code、message 等上下文信息
- `platform`：可回传请求平台或网关平台
- `data`：业务 Protobuf Response message 序列化后的 bytes

逻辑 JSON 示例（仅用于调试展示）：

```json
{
  "maxType": 2100,
  "minType": 2098,
  "extend": {
    "traceId": "trace-20260518-000001",
    "requestId": "req-20260518-000001",
    "code": "10200",
    "message": "OK"
  },
  "platform": "ANDROID",
  "data": "业务 Protobuf Response 序列化后的 bytes"
}
```

## 6. 状态码

Gateway 对外响应使用项目统一状态码：

| HTTP 状态码 | 含义 | 说明 |
|---:|---|---|
| 200 | 成功 | 请求处理完成，具体业务结果见 MessagePacket.extend.code |
| 400 | 请求参数错误 | MessagePacket 解析失败或必填字段缺失 |
| 401 | 未认证 | 鉴权失败 |
| 403 | 无权限 | 权限不足 |
| 404 | 路由不存在 | maxType/minType 未在 routes.yaml 中注册 |
| 429 | 请求过于频繁 | 限流触发 |
| 500 | 内部错误 | Gateway 或 Tars 调用异常 |

MessagePacket.extend 中的项目统一状态码：

| code | 含义 |
|---:|---|
| 10200 | 成功 |
| 10400 | 请求参数错误 |
| 10401 | 无权限 |
| 10404 | 资源不存在 / 路由不存在 |
| 10429 | 请求过于频繁 |
| 10500 | 失败 / 内部错误 |
| 10504 | 上游超时 / Tars 调用超时 |

## 7. 相关文档

- [protobuf规范.md](protobuf规范.md)
- [OpenAPI规范.md](OpenAPI规范.md)
- [tars规范.md](tars规范.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)
