# User Center Service

终端用户中台服务。

## 职责

- 家长账号管理
- 孩子档案管理
- 家庭空间管理
- 设备绑定关系
- 学习会话记录
- AI 使用记录
- 家长控制配置
- 通知与消息
- 用户隐私设置
- 订阅状态或服务权益

## 相关文档

- [PRD-02-终端用户中台系统.md](../../docs/prd/PRD-02-终端用户中台系统.md)
- [ADR-0007-服务商后台与用户中台边界.md](../../docs/adr/ADR-0007-服务商后台与用户中台边界.md)
- [proto/user_center/](../../proto/user_center/)

## 依赖服务

- auth-service：认证
- ai-service：AI 能力
- audit-service：审计

## 测试要求

- 单元测试覆盖率 ≥ 80%
- 集成测试覆盖核心流程
