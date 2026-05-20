---
name: 每日知识蒸馏
command: daily-distill
summary: 按日期读取 Raw 材料，生成日报、知识蒸馏、主控汇报、Index 候选和 Source Map 候选。用于本地 Trae 手动补跑或回放 SOLO Web 的每日定时蒸馏流程。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - daily
  - distillation
  - solo-web
  - scheduled
---

你是项目的定时知识蒸馏执行助手。

你的目标是按照 `cairobot-scheduled-knowledge-distillation` Skill 的规则，从 Raw 层材料中**提取结构化知识**，生成 Distillation 层产物和 Index 更新候选。

## 核心原则

1. **Raw → Distillation → Index 三层结构**：蒸馏从 Raw 读取，输出到 Distillation（候选态），Index 只生成候选追加
2. **Source Record ID 主键**：每条 Raw 记录用 SRC-ID 追溯，不强制 Task ID
3. **Source Map 关联审慎**：跨来源关联通过 RG 维护，禁止假关联
4. **候选 vs 确认**：所有 Index 更新为 candidate 态，不直接覆盖正式 Index
5. **只读 Raw**：不得修改或删除任何 Raw 文件
6. **禁止 auto commit/push**：只写文件，不执行 git 操作

详细规则参见 [cairobot-scheduled-knowledge-distillation Skill](../skills/cairobot-scheduled-knowledge-distillation/SKILL.md)。

## 使用方式

### 方式一：标准模式（默认当天）

```text
/daily-distill
```

扫描今天（YYYY-MM-DD）的 Raw 层新增文件，执行完整蒸馏流程。

### 方式二：指定日期补跑

```text
/daily-distill --date 2026-05-18
```

扫描指定日期的 Raw 层文件，执行蒸馏。用于补跑遗漏的日期。

### 方式三：指定来源范围

```text
/daily-distill --sources trae,tabbit --date 2026-05-19
```

仅扫描指定来源的 Raw 文件。支持值：`trae` / `tabbit` / `report` / `all`（默认）。

### 方式四：轻量模式

```text
/daily-distill --lightweight
```

适用于 Raw 层当日新增较少的场景，简化 Phase 2 去噪步骤但仍完成全部 5 个阶段。

## 执行步骤

收到 `/daily-distill` 后，按以下顺序执行：

### Step 1: Raw 层扫描

1. 确定目标日期（默认今天）
2. 扫描以下目录中该日期的新增/修改文件：
   - `docs/tabbit/inbox/{YYYY/MM/}*.md`
   - `docs/trae-export/inbox/{YYYY/MM/}*.md`
   - `docs/trae-export/inbox/tasks/{YYYY/MM/}*.md`
   - `docs/reports/daily/*{date}*.md`
   - `docs/reports/testing/*{date}*.md`
3. 为每条记录确定身份：
   - 已有 SRC-ID → 直接复用
   - 无 SRC-ID → 生成 `SRC-{SOURCE}-{timestamp}-{hash8}`
4. 查询 `source-map.md` 是否有匹配的 Relation Group
5. 输出 Raw 扫描结果清单

### Step 2: 去噪与过滤

对每条 Raw 记录：
1. 删除工具调用流水（toolName/status/json 块）
2. 删除重复对话片段
3. 删除 "*内容由 AI 生成仅供参考*" 免责声明
4. 保留核心结论、决策点、产物清单、测试结果、风险事项

### Step 3: 知识提取（Distillation 层）

基于去噪后的内容，分类提取：

| 分类 | 输出目录 | 判断标准 |
|---|---|---|
| 稳定决策 | `docs/wiki/decisions/` | 经主控确认或明显正确的技术决策 |
| 可复用流程 | `docs/wiki/modules/` | 可在后续任务中复用的工程方法 |
| 日知识摘要 | `docs/wiki/daily/` | 当日关键信息压缩 |
| 工程约束 | 对应 modules 页面 | 规则、约束、边界条件 |

每条蒸馏产物必须包含**来源元数据**：

```markdown
## 来源元数据
- **Source Records**: SRC-ID list + paths
- **Relation Group**: RG-ID or none
- **蒸馏日期**: YYYY-MM-DD
- **蒸馏执行者**: Trae IDE
- **置信度**: high / medium / low
- **状态**: candidate
```

### Step 4: Index 候选生成

生成独立的 Index 更新候选文件：

```markdown
# Index 更新候选 — {YYYY-MM-DD}

## 每日蒸馏索引候选
| 日期 | Raw 来源 (SRC-ID) | 蒸馏产物路径 | 状态 |
|---|---|---|---|

## Source Map 候选
| RG-ID | SRC-ID | 来源类型 | 关联方式 | 置信度 | 状态 |
|---|---|---|---|---|---|

## 任务索引候选
| 记录ID (SRC-ID) | 任务名称 | Raw 来源 | 蒸馏文件 | 状态 |
|---|---|---|---|---|
```

> 所有条目状态为 `candidate`。需经主控确认后合并到正式 Index。

### Step 5: 输出最终报告

## 完成后必须输出的内容

```
定时知识蒸馏完成

1. 蒸馏日期：
   YYYY-MM-DD

2. Raw 来源清单：
   - SRC-TABBIT-{...} : {path} ({title})
   - SRC-TRAE-{...} : {path} ({title})
   - 共 N 条 Source Record

3. 蒸馏产物清单：
   - 新增 N 个 Distillation 文件
     * docs/wiki/daily/{file}
     * docs/wiki/modules/{file}
     * docs/wiki/decisions/{file}
   - 修改 N 个已有文件

4. Index 候选：
   - 每日蒸馏索引：N 条候选
   - Source Map：N 条候选关联
   - 任务索引：N 条候选
   - 候选文件路径：{path}

5. 待确认项：
   -
   -

6. 是否允许进入正式 Index：
   Yes / No（说明原因）

7. Git 操作：
   ✅ 未 add / 未 commit / 未 push

8. 风险与限制：
   -
   -
```

## 注意事项

- **Raw 文件只读不改**
- **不直接修改 LLM-WIKI.md 或任何正式 Index 文件**
- **不执行 git add / commit / push**
- **candidate 关联不得写成 confirmed**
- 与 `/task-raw-archive` 的关系：本 Command 消费 Raw（读取），不生成 Raw
- 与 `/llm-wiki-distill` 的关系：本 Command 是其"定时/手动补跑"版本，增加 SRC-ID/SOURCE-MAP 规范
- SOLO Web 定时任务的 Prompt 必须引用 `cairobot-scheduled-knowledge-distillation` Skill 的核心规则
