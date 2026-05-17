---
name: cairobot-active-gap-filling
description: 当用户布置任务时，主动对照 AGENTS.md 和 .trae/rules/ 检查工程闭环是否完整。如有缺口必须补齐或显式说明不适用原因，禁止机械执行。
trigger:
  - 任何来自用户的开发任务
  - 任何 Issue 启动
  - 任何 PR 创建
  - "实现"
  - "新增功能"
  - "修复 Bug"
  - "补测试"
priority: highest
blocking: true
---

# CaiRobot MVP 主动补齐工程闭环 Skill

## 1. Skill 职责

本 Skill 强制 Trae 在执行任何开发任务前，先做"工程闭环缺口扫描"。
用户任务描述不一定完整，Trae 必须主动识别缺口并补齐，而不是机械按字面执行。

**负责**：
- 工程闭环缺口扫描
- 缺口补齐或显式说明
- 重大缺口上报

**不负责**：
- 产品需求定义（由 PRD 处理）
- 协议编号分配（由 cairobot-proto-registry-guard 处理）
- 测试执行（由 cairobot-tdd-loop 处理）

详细规则参见：[.trae/rules/reporting.md](../../.trae/rules/reporting.md)

## 2. 强制执行步骤

收到任何任务后，**第一步**必须执行以下扫描：

### 2.1 工程闭环 10 项检查清单

逐项确认，每项必须输出"已具备 / 需补齐 / 不适用（说明原因）"：

| 序号 | 检查项 | 检查内容 |
|---|---|---|
| 1 | PRD 存在性 | 对应 PRD 文件是否在 docs/prd/ 下存在？路径？ |
| 2 | ADR 存在性 | 对应 ADR 文件是否在 docs/adr/ 下存在或需要新建？路径？ |
| 3 | Protobuf 协议 | proto/ 下协议是否已定义？编号是否已登记到 docs/api/协议编号注册表.md？ |
| 4 | 失败测试 | 是否已有失败测试文件？路径？ |
| 5 | 目录规范 | 实现代码目录是否符合 docs.md 规定？ |
| 6 | 分支规范 | 分支名是否符合 git.md 规范？ |
| 7 | CI 覆盖 | CI Workflow 是否覆盖本次变更？ |
| 8 | 测试报告 | docs/reports/testing/ 是否准备好？ |
| 9 | 日报模板 | 日报模板是否准备好？ |
| 10 | LLM Wiki | docs/wiki/LLM-WIKI.md 是否需要更新索引？ |

### 2.2 缺口补齐策略

对于"需补齐"项：

| 缺口类型 | 处理方式 |
|---|---|
| 简单缺口 | 直接补齐（目录、模板、占位文件） |
| 中等缺口 | 先草拟、列出待确认事项、再继续执行 |
| 重大缺口 | 必须停止当前任务，上报项目主控 A/B |

### 2.3 不适用项说明要求

对于"不适用"项，必须显式说明：
- 为什么不适用
- 哪条规则允许跳过
- 如果未来变得适用应在何时重新评估

## 3. 完成前硬校验清单

执行任何任务前，必须先输出以下内容，**否则视为违规**：

- [ ] 工程闭环 10 项扫描结果（逐条说明）
- [ ] 已补齐项清单
- [ ] 待补齐项清单及补齐计划
- [ ] 不适用项清单及理由
- [ ] 需项目主控 A/B 决策的事项

## 4. 联动 Skill

扫描完成后，根据缺口类型激活对应 Skill：

| 缺口类型 | 激活 Skill |
|---|---|
| 协议缺口 | cairobot-proto-registry-guard |
| 测试缺口 | cairobot-tdd-loop |
| 文档缺口 | cairobot-doc-placement |
| CI 缺口 | cairobot-ci-gatekeeper |
| 汇报缺口 | cairobot-daily-report |
| Git 规范 | cairobot-git-discipline |
| 编码规范 | cairobot-coding-standard |

## 5. 违规阻断

以下行为视为违规，必须**立即停止**并上报：

- 跳过扫描直接开始写代码
- 把"用户没说"作为不补齐的理由
- 把缺口隐藏在最终汇报中而不显式列出
- 自行决定重大缺口的处理方案而不上报主控
- 未经确认自行修改 PRD/ADR 范围
- 未经确认自行引入新依赖或新架构
