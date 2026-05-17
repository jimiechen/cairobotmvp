# Git 工作流规则

## 1. 分支规则

推荐分支：

- `main`：稳定分支，只允许合并通过评审的代码。
- `dev`：集成分支，日常开发合并目标。
- `feature/*`：功能分支。
- `fix/*`：修复分支。
- `docs/*`：文档分支。
- `test/*`：测试相关分支。
- `refactor/*`：重构分支。
- `chore/*`：杂务分支（依赖更新、配置调整等）。
- `ci/*`：CI/CD 配置分支。
- `hardware/*`：硬件相关分支。

所有功能开发应从 `dev` 创建分支。

## 2. 分支命名规范

### 2.1 命名格式

```text
type/module-short-description
```

### 2.2 示例

```text
feature/ai-intent-classifier
feature/app-learning-mode
fix/firmware-lock-timeout
docs/prd-00-mvp-overview
test/ai-forbidden-intents
hardware/mvp-baseboard-v1
refactor/app-device-state
```

### 2.3 约束

- 一个分支只做一类事情
- 文档不要和功能混在同一个分支
- 测试补充不要和硬件设计混在同一个分支
- 超过 3 天没合并的分支要重新同步 dev

## 3. 提交规则

### 3.1 提交粒度

每次提交应该只表达一个清晰明确的动作：

- 加一个测试
- 修一个 bug
- 重构一个类
- 补一份文档

不要一个提交里同时：

- 改 PRD + 改固件 + 改 App + 改测试 + 改目录结构

### 3.2 提交信息格式

```text
type(scope): 中文说明
```

示例：

```text
docs(prd): 添加MVP总纲
test(ai): 添加意图分类失败测试
feat(ai): 实现基础意图分类
fix(app): 修复设备断连状态判断
refactor(firmware): 简化锁仓状态机
```

允许的 type（全部小写）：

- `docs`：文档变更
- `test`：测试变更
- `feat`：功能实现
- `fix`：缺陷修复
- `refactor`：代码重构
- `chore`：杂务（依赖更新、配置等）
- `ci`：CI/CD 配置变更
- `build`：构建系统变更
- `hardware`：硬件相关变更

> **注意**：Commit message 格式必须严格遵守 `type(scope): 中文说明`，不得随意编写。

## 4. PR 规则

### 4.1 基本要求

每个 PR 必须：

- 关联 Issue。
- 说明变更内容。
- 说明测试结果。
- 说明文档是否更新。
- 控制变更范围。
- 不混入无关格式化。

### 4.2 PR 粒度

一个 PR 最好只做：

- 一个 Issue
- 或一组强相关的小 Issue

PR 里不要混：

- 功能 + 大重构 + 格式化 + 文档整理

### 4.3 PR 改动规模建议

- 单个 PR 改动文件数：推荐 ≤ 10 个
- 单个 PR 代码行数：推荐 ≤ 300 行

超过以上限制需在 PR 中说明原因。

## 5. 合并规则

以下情况不得合并：

- 测试未通过。
- 没有关联 Issue。
- 没有说明测试结果。
- 涉及行为变化但未更新文档。
- PR 范围过大且无合理说明。
- 存在未解释的新增依赖。
