# CaiRobot MVP 单网关架构设计

## 元信息

- Canonical Task ID: TB-20260518-200000-cairobot-single-gateway
- External Task ID: 无
- 原始文件名:
  - `TabAI会话_1779079297836.md`
  - `TabAI会话_1779093720973.md`
- 重命名后路径:
  - `docs/tabbit/inbox/2026/05/TB-20260518-200000-cairobot-single-gateway.tabbit.raw.md`
  - `docs/tabbit/inbox/2026/05/TB-20260518-200000-cairobot-single-gateway-revision.tabbit.raw.md`
- 创建时间：2026-05-18 20:00:00
- 主题 slug: cairobot-single-gateway
- 任务状态: archived
- 蒸馏状态: pending
- 是否需要人工确认: false
- 输入模式: 批量迁移

## 原始任务

本任务包含两轮迭代：

**第一轮（`TabAI会话_1779079297836.md`）**：
设计 CaiRobot MVP 单网关统一架构。核心方案为 `POST /api/hello` → MessagePacket → maxType/minType 路由 → TarsCloud/TarsGo 内部服务。对齐项目已有的 `docs/api/protobuf规范.md` 中 MessagePacket 定义。

**第二轮（`TabAI会话_1779093720973.md`）**：
基于 TRAE 评审意见的修正版执行需求。明确 TarsCloud/TarsGo 作为内部主 RPC 与服务治理方案；gRPC 不再作为内部主链路；解决 ADR 冲突、gRPC/Tars 边界不清、OpenAPI/HTTP Gateway 旧规范冲突等问题。

## 自动理解结果

- **任务目标**: 将 CaiRobot MVP 项目收敛到"单网关 + MessagePacket + TarsCloud/TarsGo 统一架构"
- **关键上下文**: 项目存在多协议并存问题（gRPC vs Tars），需要通过评审意见统一到单一入口
- **产出类型**: 架构设计方案 + 执行需求文档
- **主题置信度**: high（文档有明确的 Protobuf 定义和路由规划）

## 执行过程

本次为批量迁移模式三处理。TRAE 扫描 `docs/tabbit/inbox/` 下未重命名文件，识别这两个文件属于同一任务的两次迭代（初版 + 评审修订版），合并为一个 Canonical Task ID 归档。

## 最终产出

两个原始文件记录了 CaiRobot 单网关架构从初稿到评审修正的完整迭代过程：

1. 初稿定义了 MessagePacket 协议和 Tars 标准接口
2. 修订稿解决了 ADR/gRPC 冲突，明确了 TarsCloud 作为唯一内部 RPC 方案

## 关联文件

| 类型 | 原始路径 | 重命名后路径 | 状态 |
|---|---|---|---|
| tabbit.raw | docs/tabbit/inbox/TabAI会话_1779079297836.md | docs/tabbit/inbox/2026/05/TB-20260518-200000-cairobot-single-gateway.tabbit.raw.md | ✅ renamed |
| tabbit.raw | docs/tabbit/inbox/TabAI会话_1779093720973.md | docs/tabbit/inbox/2026/05/TB-20260518-200000-cairobot-single-gateway-revision.tabbit.raw.md | ✅ renamed (revision) |
| archive | — | docs/wiki/tasks/2026/05/TB-20260518-200000-cairobot-single-gateway.archive.md | generated |

## Wiki 索引候选

- [CaiRobot MVP 单网关架构设计](./tasks/2026/05/TB-20260518-200000-cairobot-single-gateway.archive.md)
  - Canonical Task ID: TB-20260518-200000-cairobot-single-gateway
  - 来源：Tabbit
  - 状态：已归档，待蒸馏
  - 摘要：CaiRobot 单网关 MessagePacket + TarsCloud/TarsGo 统一架构方案的初稿与评审修订版

## 夜间蒸馏输入

本任务已生成 manifest，等待夜间 `tabbit-task-distillation` Skill 按 Canonical Task ID 聚合两轮迭代材料进行知识蒸馏。
