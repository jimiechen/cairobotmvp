# Webhook 规范

## 1. 概述

开放平台支持 Webhook 回调，用于推送事件通知。

## 2. 事件类型

- `device.online`：设备上线
- `device.offline`：设备离线
- `device.alert`：设备告警
- `learning.session.start`：学习会话开始
- `learning.session.end`：学习会话结束
- 等等

## 3. 回调格式

```http
POST /your/webhook/url HTTP/1.1
Content-Type: application/json
X-Webhook-Signature: sha256=xxx
X-Webhook-Event: device.online
X-Webhook-Timestamp: 1716000000
X-Request-ID: xxx-xxx-xxx

{
  "event": "device.online",
  "timestamp": 1716000000,
  "data": {
    "device_id": "xxx",
    "device_name": "xxx"
  }
}
```

## 4. 签名验证

- 使用 HMAC-SHA256 签名
- 签名包含：请求体 + timestamp
- 签名放在 `X-Webhook-Signature` Header 中
- 格式：`sha256=base64(hmac_sha256(app_secret, request_body + timestamp))`

## 5. 重试机制

- 回调失败自动重试
- 重试间隔：1s, 5s, 30s, 5min, 30min
- 最多重试 5 次

## 6. 相关文档

- [OpenAPI规范.md](OpenAPI规范.md)
- [ADR-0006-开放平台API边界.md](../adr/ADR-0006-开放平台API边界.md)
