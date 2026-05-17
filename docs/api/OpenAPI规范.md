# OpenAPI 规范

## 1. 概述

开放平台对外提供 HTTPS JSON 接口，使用 OpenAPI 3.0 规范文档描述。

## 2. 认证方式

- API Key：在 HTTP Header 中传递 `X-API-Key`
- OAuth 2.0：Bearer Token
- 签名验证：HMAC-SHA256

## 3. 请求头

所有请求必须包含：
```http
X-Request-ID: xxx-xxx-xxx
X-API-Version: v1
Content-Type: application/json
```

## 4. 响应格式

- 成功响应：返回对应的数据结构
- 错误响应：返回标准错误格式

## 5. 版本管理

- API 版本在路径中：`/api/v1/xxx`
- 或在 Header 中：`X-API-Version: v1`

## 6. 限流

- API 调用有频率限制
- 超过限制返回 429 Too Many Requests
- Header 中包含限流信息：
  ```http
  X-RateLimit-Limit: 1000
  X-RateLimit-Remaining: 999
  X-RateLimit-Reset: 1716000000
  ```

## 7. 相关文档

- [HTTP-gateway规范.md](HTTP-gateway规范.md)
- [Webhook规范.md](Webhook规范.md)
- [ADR-0006-开放平台API边界.md](../adr/ADR-0006-开放平台API边界.md)
