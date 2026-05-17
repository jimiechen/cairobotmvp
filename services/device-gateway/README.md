# Device Gateway Service

设备网关服务。

## 职责

- 设备连接管理
- 设备认证
- 设备指令下发
- 设备状态上报
- 设备事件推送

## 相关文档

- [PRD-06-设备通信与协议.md](../../docs/prd/PRD-06-设备通信与协议.md)
- [proto/device/](../../proto/device/)

## 依赖服务

- auth-service：认证
- audit-service：审计

## 测试要求

- 单元测试覆盖率 ≥ 80%
