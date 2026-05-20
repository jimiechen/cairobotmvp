---
name: TDD 红绿重构循环
slug: tdd-loop
summary: 工程级 TDD 闭环，强制红-绿-重构流程。当用户要求实现功能、修复 Bug、补充测试，或涉及代码目录变更时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - tdd
  - testing
  - code-development
trigger:
  - "实现"
  - "新增功能"
  - "修复 Bug"
  - "补测试"
  - "test:"
  - "feat:"
  - "fix:"
  - 涉及 services/、ai/、web/、proto/ 目录
priority: high
blocking: true
---

# CaiRobot MVP 工程级 TDD 闭环 Skill

## 1. Skill 职责

本 Skill 强制 Trae 在任何代码变更前后走完完整 TDD 闭环。

**负责**：
- 红-绿-重构流程执行
- 测试编写与执行
- 报告生成

**不负责**：
- 产品需求定义（由 PRD 处理）
- 协议编号分配（由 cairobot-proto-registry-guard 处理）

详细规则参见：
- [.trae/rules/tdd.md](../../.trae/rules/tdd.md)
- [.trae/rules/testing.md](../../.trae/rules/testing.md)

## 2. 强制执行步骤

执行任何代码任务前，必须按以下顺序完成：

### 2.1 需求确认（红）
- [ ] 阅读相关 PRD，确认验收标准
- [ ] 确认非目标
- [ ] 确认涉及模块

### 2.2 协议确认（红）
- [ ] 定义 Protobuf / OpenAPI 契约（如需要）
- [ ] 注册 max + min 编号到 docs/api/协议编号注册表.md
- [ ] 更新 OpenAPI 映射

### 2.3 测试编写（红）
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] 编写契约测试（如涉及协议）
- [ ] 编写边界测试
- [ ] 运行测试，确认失败原因符合预期

### 2.4 最小实现（绿）
- [ ] 只写让当前测试通过的最小代码
- [ ] 不得顺手实现未来功能
- [ ] 不得引入不必要抽象

### 2.5 重构（如有必要）
- [ ] 外部行为不变
- [ ] 测试全部通过
- [ ] 命名更清晰
- [ ] 结构更简单

### 2.6 报告生成
- [ ] 运行所有测试
- [ ] 生成测试报告到 docs/reports/testing/
- [ ] 输出 Standalone HTML 报告到 docs/reports/html/
- [ ] 截图和视频证据

## 3. 完成前硬校验清单

任务汇报"已完成"前，必须能回答以下问题：

- [ ] 对应 PRD 文件路径：__________
- [ ] 对应 ADR 文件路径：__________
- [ ] 对应 Protobuf 文件路径：__________
- [ ] 协议编号 max+min：__________（已登记？是/否）
- [ ] 失败测试文件路径：__________
- [ ] 测试通过证据（命令 + 输出摘要）：__________
- [ ] 测试报告路径：__________
- [ ] 日报路径：__________
- [ ] CI 是否通过：__________

**新增硬校验**：
- [ ] 测试通过证据必须粘贴终端原始输出，不得仅填"通过"或"已运行"。

## 4. 禁止行为

以下行为**绝对禁止**：

- 先写实现再补测试
- 删除或修改失败测试以掩盖 Bug
- 没有协议先写实现（涉及 proto/ 时）
- 没有测试通过证据就宣称完成
- 静默跳过 CI 检查项

## 5. 联动 Skill

- 协议变更时必须激活 cairobot-proto-registry-guard
- 提交时必须激活 cairobot-git-discipline
- 任务完成时必须激活 cairobot-daily-report
- 合并前必须激活 cairobot-ci-gatekeeper
