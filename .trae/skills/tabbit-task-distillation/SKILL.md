---
name: Tabbit 任务蒸馏
slug: tabbit-task-distillation
summary: 按 manifest pending 状态扫描 Tabbit 归档任务，按 Canonical Task ID 聚合材料蒸馏为 LLM Wiki 稳定知识。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - tabbit
  - task-distillation
  - canonical-task-id
trigger:
  - "蒸馏"
  - "distill"
  - "知识蒸馏"
  - "tabbit distill"
  - manifest pending
priority: medium
blocking: false
---

# Tabbit Task Distillation Skill (v3)

## 1. Skill 职责

本 Skill 将已通过 `/tabbit-task` 归档的 Tabbit / TRAE 任务资产，按 **Canonical Task ID** 聚合并蒸馏为可进入 LLM Wiki 的稳定知识。

**负责**：
- 扫描 pending manifest（按 `distillation_status: pending` 过滤）
- 按 **Canonical Task ID** 聚合材料
- 过滤噪声内容
- 提炼稳定知识
- 生成 distilled 文档
- 更新 LLM Wiki 索引
- 更新 manifest 的 `distillation_status` 为 `distilled`

**不负责**：
- 接收原始任务（由 `/tabbit-task` 负责）
- 重命名文件（由 `/tabbit-task` 负责）
- 生成归档文档（由 `/tabbit-task` 负责）
- 替代日报机制

详细规则参见：
- [ADR-0009 v3](../../docs/adr/ADR-0009-tabbit-task-archive-flow.md)
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md)

## 2. 输入

默认扫描路径：

```
docs/wiki/tasks/**/*.manifest.md
```

只处理 `distillation_status = pending` 的 manifest。

也可以显式指定单个 Canonical Task ID 进行蒸馏。

## 3. 处理流程

### 3.1 读取 manifest

从 manifest 中获取：

- **Canonical Task ID**（主键，必填）
- **External Task ID**（可选追溯字段）
- 关联文件路径列表
- `needs_human_review` 标记
- 蒸馏目标和优先级

> 如果 `needs_human_review: true`，蒸馏时应更谨慎，并在产出中标注置信度。

### 3.2 读取关联材料

按 manifest 中的关联文件路径依次读取：

1. `.tabbit.raw.md` — Tabbit / TabAI 原始导出
2. `.manual.raw.md` — 手动创建文件
3. `.trae.exec.md` — TRAE 执行记录
4. `.archive.md` — 正式归档文档

如果某类文件不存在则跳过。

### 3.3 过滤噪声

过滤以下内容，不进入蒸馏产物：

- 重复的对话片段
- 工具调用流水（toolName/status/json 等机器输出）
- 临时判断和中间结论
- 未验证的假设
- 与长期知识无关的过程性描述
- 纯粹的操作日志
- `*内容由 AI 生成仅供参考*` 类免责声明

### 3.4 提炼稳定知识

从过滤后的材料中提炼：

1. **稳定决策**：经过验证的架构决策、取舍理由
2. **可复用流程**：可以重复执行的步骤序列
3. **命名规范**：文件名、目录名、Canonical Task ID 的约定
4. **模板沉淀**：Command、Skill、Markdown 模板
5. **工程约束**：必须遵守的规则和边界条件
6. **自动化机会**：可以被脚本化的操作

### 3.5 生成蒸馏文件

输出到：

```
docs/reports/distilled/{YYYY}/{MM}/{CANONICAL_TASK_ID}.distilled.md
```

### 3.6 更新索引和状态

1. 更新或追加 `docs/wiki/LLM-WIKI.md` 中的索引条目
2. 将 manifest 中的 `distillation_status` 从 `pending` 改为 `distilled`
3. 将 manifest 中的 `updated_at` 更新为当前时间

## 4. 蒸馏文档模板

# {知识标题}

## 来源任务

- Canonical Task ID: {TASK_ID}
- External Task ID: {external_or_none}
- 归档文件: {archive_path}
- 原始材料:
  - {tabbit_raw_path} （如有）
  - {manual_raw_path} （如有）
  - {trae_exec_path} （如有）

## 稳定结论

记录可以长期保留的工程决策。

## 可复用流程

记录可以重复执行的步骤。

## 命名规范

记录文件、目录、Canonical Task ID 的规范。

## 模板

如果本任务产生了 Command、Skill、Markdown 模板，在此完整沉淀。

## 不进入 Wiki 的内容

列出被过滤掉的临时内容类型及其原因。

## Wiki 更新建议

生成可追加到 `docs/wiki/LLM-WIKI.md` 的条目。

## 5. 完成前校验清单

蒸馏完成后确认：

- [ ] distilled 文件已写入正确路径
- [ ] LLM-WIKI.md 索引已更新
- [ ] manifest `distillation_status` 已改为 `distilled`
- [ ] manifest `updated_at` 已更新
- [ ] 蒸馏内容不含工具调用噪声
- [ ] 蒸馏内容不含未验证假设
- [ ] 如 `needs_human_review: true`，蒸馏产出已标注置信度

## 6. 联动 Skill

- 蒸馏完成后如发现归档质量问题，激活 `cairobot-doc-placement` 修正
- 如需提交日报，激活 `cairobot-daily-report`

## 7. 夜间定时集成

本 Skill 可挂载到每日工作流：

```text
21:30 停止开发
  ↓
生成日报
  ↓
处理 Tabbit pending manifests（本 Skill）
  ↓
蒸馏 Markdown
  ↓
同步更新 LLM Wiki
```

调度入口：扫描 `docs/wiki/tasks/**/*.manifest.md` 中 `distillation_status = pending` 的条目。
