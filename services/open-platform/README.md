# Open Platform Service

开放平台服务。

## 职责

- 开放 API 认证
- AppKey / AppSecret 管理
- OAuth 或 Token 鉴权
- API 调用配额
- API 调用日志
- Webhook 回调

## 相关文档

- [PRD-03-开放平台API.md](../../docs/prd/PRD-03-开放平台API.md)
- [ADR-0006-开放平台API边界.md](../../docs/adr/ADR-0006-开放平台API边界.md)
- [proto/open_platform/](../../proto/open_platform/)

## 依赖服务

- auth-service：认证
- audit-service：审计

## 测试要求

- 单元测试覆盖率 ≥ 80%
