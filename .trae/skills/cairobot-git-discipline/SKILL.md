---
name: Git 提交纪律
slug: git-discipline
summary: Git 分支命名、提交信息格式和 PR 粒度规范。创建分支、提交代码或创建 PR 时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - git
  - commit
  - pr
trigger:
  - "创建分支"
  - "git commit"
  - "创建 PR"
  - "push"
  - "合并"
priority: medium
blocking: false
---

# CaiRobot MVP Git 规范 Skill

## 1. Skill 职责

本 Skill 强制执行 CaiRobot MVP 项目的 Git 协作规范。

**负责**：
- 分支命名规范
- 提交信息格式
- PR 粒度控制

详细规则参见：
- [.trae/rules/git.md](../../.trae/rules/git.md)

## 2. 分支命名规范

### 2.1 分支类型

| 类型 | 用途 | 示例 |
|---|---|---|
| `main` | 稳定分支 | - |
| `dev` | 集成分支 | - |
| `feature/*` | 功能开发 | `feature/ai-intent-classifier` |
| `fix/*` | 缺陷修复 | `fix/firmware-lock-timeout` |
| `docs/*` | 文档变更 | `docs/prd-00-mvp-overview` |
| `test/*` | 测试变更 | `test/ai-forbidden-intents` |
| `refactor/*` | 代码重构 | `refactor/app-device-state` |
| `chore/*` | 杂务 | `chore/update-deps` |
| `ci/*` | CI 配置 | `ci/add-workflow` |
| `hardware/*` | 硬件变更 | `hardware/mvp-baseboard-v1` |

### 2.2 命名格式

```
type/module-short-description
```

示例：
```
feature/ai-intent-classifier
fix/device-state-race
docs/api-proto-mapping
```

## 3. 提交信息格式

### 3.1 格式

```
type(scope): 中文说明
```

### 3.2 允许的 type

| type | 说明 |
|---|---|
| `docs` | 文档变更 |
| `test` | 测试变更 |
| `feat` | 功能实现 |
| `fix` | 缺陷修复 |
| `refactor` | 代码重构 |
| `chore` | 杂务（依赖更新等） |
| `ci` | CI/CD 配置变更 |
| `build` | 构建系统变更 |
| `hardware` | 硬件相关变更 |

### 3.3 示例

```
docs(prd): 添加MVP总纲
test(ai): 添加意图分类失败测试
feat(ai): 实现基础意图分类
fix(app): 修复设备断连状态判断
refactor(firmware): 简化锁仓状态机
```

## 4. PR 粒度

### 4.1 推荐规模

| 项目 | 推荐上限 |
|---|---|
| 改动文件数 | 10 个 |
| 代码行数 | 300 行 |

### 4.2 PR 内容要求

每个 PR 必须：
- 关联 Issue
- 说明变更内容
- 说明测试结果
- 说明文档更新
- 控制变更范围

## 5. 完成前校验清单

提交前必须确认：

- [ ] 分支名符合规范
- [ ] 提交信息格式正确
- [ ] 不在单个提交中混合多个不相关变更
- [ ] PR 关联了 Issue

## 6. 违规警告

以下行为需警告：

- 分支名不符合规范
- 提交信息格式不正确
- PR 改动过大未说明原因
- 提交信息与实际变更不符

## 7. 联动 Skill

- 分支创建时激活 cairobot-active-gap-filling
- PR 创建前激活 cairobot-ci-gatekeeper
- 任务完成时激活 cairobot-daily-report
