# OpenAPI 规范

## 1. 概述

OpenAPI 在当前架构中仅描述单网关入口 `/api/hello` 的外层 MessagePacket 结构和调试方式，不再作为每个业务命令的独立 REST path 契约。具体业务请求和响应结构由 `maxType/minType` 对应的 Protobuf message 决定。

历史规划中曾使用 OpenAPI 描述多 REST path 业务接口（`/api/v1/xxx`），该方案已根据 ADR-0008 调整为当前单网关架构。

## 2. 单网关入口定义

OpenAPI 只描述以下入口：

```http
POST /api/hello
Content-Type: application/octet-stream
```

请求体为 `MessagePacket` 序列化后的二进制内容。OpenAPI 文档中可使用逻辑等价的 JSON Schema 描述 MessagePacket 结构，但必须明确标注真实协议为二进制 Protobuf bytes。

### 2.1 MessagePacket 结构（文档展示用）

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
  "data": "业务 Protobuf Request 序列化后的 bytes（Base64 编码）"
}
```

> 注意：以上 JSON 仅用于 OpenAPI 文档展示和调试说明，真实协议入口是 `application/octet-stream` 的 MessagePacket bytes。

## 3. 认证方式

- API Key：在 HTTP Header 中传递 `X-API-Key`，或在 `MessagePacket.extend.token` 中传递
- OAuth 2.0：Bearer Token
- 签名验证：HMAC-SHA256

## 4. 请求头

所有请求必须包含：

```http
X-Request-ID: xxx-xxx-xxx
Content-Type: application/octet-stream
```

> 历史版本中的 `X-API-Version: v1` 和 `Content-Type: application/json` 已调整为当前单网关模式。

## 5. 响应格式

- 成功响应：返回 `MessagePacket` 序列化后的二进制内容
- 错误响应：仍返回 `MessagePacket`，其中 `extend.code` 和 `extend.message` 携带错误信息

逻辑 JSON 示例（仅用于文档展示）：

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
  "data": "业务 Protobuf Response 序列化后的 bytes（Base64 编码）"
}
```

## 6. 版本管理

- API 版本通过 `MessagePacket.extend.version` 或 Header `X-API-Version` 传递
- 不再在 URL path 中体现版本（如 `/api/v1/xxx`）
- 业务语义由 `maxType/minType` 决定，版本信息作为辅助字段

## 7. 限流

- API 调用有频率限制
- 超过限制返回 429 Too Many Requests
- Header 中包含限流信息：

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1716000000
```

## 8. OpenAPI 的用途

OpenAPI 可用于：
- 文档展示和调试说明
- SDK 生成时的外层包装说明
- 第三方接入时的入口协议文档

真实业务契约仍以 Protobuf + 协议编号注册表为准。`MessagePacket.data` 的具体业务结构由 `maxType/minType` 对应的 Protobuf message 决定。

## 9. 相关文档

- [HTTP-gateway规范.md](HTTP-gateway规范.md)
- [protobuf规范.md](protobuf规范.md)
- [openapi-protobuf映射规范.md](openapi-protobuf映射规范.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)
