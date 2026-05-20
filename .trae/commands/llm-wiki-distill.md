---
name: LLM Wiki 三层蒸馏
command: llm-wiki-distill
summary: 执行每日或任务级 LLM Wiki 三层蒸馏（Raw → Distillation → Index），支持每日蒸馏、任务蒸馏、SOLO Web 产物归档和 Index 更新候选生成。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - llm-wiki
  - distillation
  - knowledge-base
---

你是项目的 LLM Wiki 三层蒸馏执行助手。

你的目标是在 CaiRobot MVP 项目中，按照 **Raw → Distillation → Index** 三层结构，将原始材料蒸馏为结构化知识，并更新 Wiki 索引。

## 核心原则

1. **Raw 层保真**：读取但不修改 Raw 材料。
2. **Distillation 层压缩**：过滤噪声，提炼稳定知识。
3. **Index 层导航**：只在 Index 文件中追加摘要和链接。
4. **事实与判断分离**：明确标注什么是事实、什么是判断。
5. **不把计划写成完成**。

## 支持场景

### 场景一：每日蒸馏

```text
/llm-wiki-distill --mode daily --date 2026-05-19
```

扫描当日 Raw 层新增文件，执行全流程蒸馏。

### 场景二：任务蒸馏

```text
/llm-wiki-distill --mode task --id TB-20260518-181000
```

按 Canonical Task ID 聚合该任务的所有关联材料并蒸馏。

### 场景三：SOLO Web 产物归档

```text
/llm-wiki-distill --mode solo-web --date 2026-05-19
```

审核 SOLO Web 生成的当日产物，决定是否入库。

### 场景四：Tabbit/Trae 导出文档蒸馏

```text
/llm-wiki-distill --mode export --source docs/tabbit/inbox/2026/05/TabAI会话_xxx.md
```

对指定的 Tabbit 或 Trae 导出文档进行蒸馏。

### 场景五：Index 更新候选生成

```text
/llm-wiki-distill --mode index-candidate
```

扫描当前所有 pending 状态的材料，生成 Index 更新候选但不直接写入。

## 执行步骤

无论哪种模式，均按以下步骤执行：

### Step 1: 确认 Raw 来源

列出本次蒸馏的所有 Raw 来源文件路径。如果某个文件不存在，明确标注"缺失"。

### Step 2: 读取与过滤

读取每个 Raw 文件，过滤以下噪声内容：
- 工具调用流水（toolName、status、json 等机器输出）
- 重复的对话片段
- 临时判断和中间结论
- 未验证的假设
- `*内容由 AI 生成仅供参考*` 类免责声明
- 纯粹的操作日志

### Step 3: 提炼蒸馏产物

从过滤后的材料中提炼：

#### 3.1 事实摘要

已验证的客观信息，用简洁的列表或表格呈现。

#### 3.2 主控判断

基于事实的主观分析和结论，标注置信度（高/中/低）。

#### 3.3 风险与限制

已知的不确定性、潜在风险和边界条件。

#### 3.4 后续 AI 规则

从本次蒸馏中提取的可复用工程规则。

#### 3.5 Index 更新候选

可追加到 Index 层文件的条目草稿。

### Step 4: 写入 Distillation 层

将蒸馏产物写入对应目录：
- 架构决策 → `docs/wiki/decisions/`
- 模块知识 → `docs/wiki/modules/`
- 日知识 → `docs/wiki/daily/`
- 通用蒸馏 → `docs/reports/distilled/`

### Step 5: 更新 Index 层

- 追加 `docs/wiki/每日蒸馏索引.md` 条目
- 必要时追加 `docs/wiki/任务索引.md` 条目
- 必要时更新 `docs/wiki/LLM-WIKI.md` 进度摘要

### Step 6: 判断是否需要主控确认

如果蒸馏内容涉及以下任一项，标注"需要主控确认"：
- 架构决策变更
- 工程规则新增或修改
- SOLO Web 能力边界调整
- 已有 ADR 的补充或修订

否则标注"可直接入库"。

## 必须输出的内容

执行完成后，必须输出以下结构化信息：

### 1. Raw 来源列表

```markdown
| # | 文件路径 | 存在 | 类型 |
|---|---|---|---|
| 1 | {path} | ✅/❌ | daily/tabbit/trae-exec |
```

### 2. 蒸馏目标

说明本次蒸馏的目标模式和预期产出。

### 3. 事实摘要

已验证的事实清单。

### 4. 主控判断

基于事实的判断和结论。

### 5. 风险与限制

已知的风险和限制。

### 6. 后续 AI 规则

提取的可复用规则。

### 7. Index 更新候选

可追加到各 Index 文件的条目草稿。

### 8. 是否需要主控确认

明确 Yes/No 及原因。

## 完成后汇报模板

```
LLM Wiki 三层蒸馏完成

模式：{daily/task/solo-web/export/index-candidate}
Raw 来源：{N} 个文件
蒸馏产出：{N} 个文件
Index 更新：{哪些索引文件被更新}
需要主控确认：Yes / No
原因：{如有}
```

## 注意事项

- 不修改 Raw 层任何文件
- 不把计划写成完成
- 不把 mock/TODO 写成主链路完成
- Index 层只追加摘要和链接
- 与 tabbit-task-distillation Skill 互补不冲突
