# AI Service

AI 服务。

## 职责

- 意图分类
- 提示词改写
- 回答审核
- OCR 结果理解
- 模型网关封装
- 安全策略执行

## 不负责

- 用户账号管理
- 服务商权限
- 设备绑定
- 开放平台鉴权
- 订单或权益
- 设备控制最终决策

## 技术栈

- 语言：Python
- Web 框架：FastAPI / Flask（待定）
- 通信：gRPC + Protobuf / HTTP JSON
- 测试：pytest

## 相关文档

- [PRD-04-AI服务系统.md](../../docs/prd/PRD-04-AI服务系统.md)
- [ADR-0004-AI服务使用Python.md](../../docs/adr/ADR-0004-AI服务使用Python.md)
- [proto/ai_service/](../../proto/ai_service/)

## 测试要求

- 单元测试覆盖率 ≥ 80%
- 安全测试覆盖边界情况
- 契约测试与 Protobuf 保持一致
