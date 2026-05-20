---
name: 工程闭环流程
slug: engineering-workflow
summary: 工程协作总入口，协调其他 Skill 完成完整工程闭环。当用户开始新任务、Issue 或 PR 时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - workflow
  - orchestration
  - task-start
trigger:
  - "开始新任务"
  - "执行任务"
  - "开发"
  - 任何 Issue 启动
  - 任何 PR 创建
priority: high
blocking: true
---

# CaiRobot MVP 工程协作总入口 Skill

## 1. Skill 职责

本 Skill 是 CaiRobot MVP 工程协作的总入口，协调其他 Skill 完成完整工程闭环。

**负责**：
- 任务启动协调
- Skill 级联调度
- 完整闭环确认

**不负责**：
- 具体实现逻辑
- 具体测试编写
- 具体文档编写

## 2. 强制执行步骤

### 2.1 任务启动流程

任何开发任务启动时，必须按以下顺序激活 Skill：

```
1. cairobot-active-gap-filling
   ↓（缺口扫描完成后）
2. 根据缺口类型激活对应 Skill
   ↓（所有 Skill 完成后）
3. cairobot-daily-report
   ↓（日报完成后）
4. cairobot-ci-gatekeeper
   ↓（CI 通过后）
5. 输出完整任务完成汇报
```

### 2.2 Skill 调度规则

| 任务类型 | 必须激活的 Skill |
|---|---|
| 功能开发 | cairobot-active-gap-filling → cairobot-tdd-loop → cairobot-coding-standard → cairobot-daily-report → cairobot-ci-gatekeeper |
| Bug 修复 | cairobot-active-gap-filling → cairobot-tdd-loop → cairobot-daily-report → cairobot-ci-gatekeeper |
| 协议变更 | cairobot-active-gap-filling → cairobot-proto-registry-guard → cairobot-tdd-loop → cairobot-daily-report → cairobot-ci-gatekeeper |
| 文档变更 | cairobot-active-gap-filling → cairobot-doc-placement → cairobot-daily-report → cairobot-ci-gatekeeper |
| CI 配置 | cairobot-active-gap-filling → cairobot-ci-gatekeeper → cairobot-daily-report |
| 纯文档 PR | cairobot-active-gap-filling → cairobot-doc-placement → cairobot-daily-report |

## 3. 完成前硬校验清单

任务完成汇报"已完成"前，必须确认以下闭环完整：

- [ ] cairobot-active-gap-filling 已执行，10 项检查完成
- [ ] 所有相关 Skill 已激活并完成
- [ ] 测试通过或有跳过说明
- [ ] CI 通过或有本地等价命令
- [ ] 日报已提交
- [ ] LLM Wiki 已更新（如需要）
- [ ] PR 已创建（如需要）
- [ ] 项目主控评审请求已发送（如需要）

## 4. 联动 Skill

| Skill | 激活时机 |
|---|---|
| cairobot-active-gap-filling | 任务启动第一步 |
| cairobot-tdd-loop | 功能开发、Bug 修复 |
| cairobot-proto-registry-guard | 协议变更 |
| cairobot-coding-standard | 功能开发 |
| cairobot-doc-placement | 文档变更 |
| cairobot-daily-report | 任务完成 |
| cairobot-ci-gatekeeper | PR 创建前 |
| cairobot-git-discipline | 分支创建、提交时 |

## 5. 违规阻断

以下行为视为违规，必须立即停止：

- 跳过 cairobot-active-gap-filling 直接激活其他 Skill
- 未完成闭环就宣称任务完成
- 跳过 CI 直接合并
