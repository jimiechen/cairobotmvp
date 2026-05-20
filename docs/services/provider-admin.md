# Provider Admin Service

服务商后台服务。

## 职责

- 服务商账号管理
- 租户/渠道管理
- 设备批次管理
- 设备绑定与解绑管理
- 用户服务状态查看
- 售后工单管理
- 设备异常查看
- 服务套餐或权益管理
- 运营数据看板
- 系统配置管理

## 相关文档

- [PRD-01-服务商后台系统.md](../../docs/prd/PRD-01-服务商后台系统.md)
- [ADR-0007-服务商后台与用户中台边界.md](../../docs/adr/ADR-0007-服务商后台与用户中台边界.md)
- [proto/provider_admin/](../../proto/provider_admin/)

## 依赖服务

- auth-service：认证
- audit-service：审计

## 测试要求

- 单元测试覆盖率 ≥ 80%
- 集成测试覆盖核心流程
