---
name: LLM Wiki 知识蒸馏
slug: llm-wiki-distillation
summary: 从 Raw 材料提炼知识、更新 Wiki 索引和执行每日蒸馏。需要从 Raw 层提炼知识、更新 LLM Wiki 或执行蒸馏时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - llm-wiki
  - distillation
  - knowledge-base
trigger:
  - "LLM Wiki 蒸馏"
  - "wiki distill"
  - "三层蒸馏"
  - "知识库更新"
  - "更新 LLM-WIKI"
priority: medium
blocking: false
---

# CaiRobot MVP LLM Wiki 三层蒸馏 Skill

## 1. Skill 职责

本 Skill 指导 AI 助手按照 **Raw → Distillation → Index** 三层结构执行知识蒸馏。

**负责**：
- 从 Raw 层材料中提取可复用的结构化知识
- 区分事实、判断、风险、规则、后续行动
- 生成 Distillation 层文档（daily/、decisions/、modules/）
- 更新 Index 层索引文件（LLM-WIKI.md、每日蒸馏索引.md、任务索引.md）

**不负责**：
- 接收原始任务归档（由 `/tabbit-task` 负责）
- Tabbit 任务的重命名和 manifest 生成（由 `tabbit-task-distillation` 负责）
- 替代日报机制（由 `cairobot-daily-report` 负责）
- 业务代码修改

详细规则参见：
- [LLM Wiki 三层架构决策](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md)
- [.trae/rules/docs.md](../../.trae/rules/docs.md)
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md)

## 2. 与其他 Skill 的关系

| Skill | 关系 | 说明 |
|---|---|---|
| `tabbit-task-distillation` | **互补** | 该 Skill 处理 Tabbit 任务级蒸馏（按 Canonical Task ID），本 Skill 处理全局 Wiki 结构维护和每日蒸馏 |
| `cairobot-scheduled-knowledge-distillation` | **被包含** | 本 Skill 是定时蒸馏规则的子集和扩展；SOLO Web 定时任务应优先引用 scheduled-knowledge-distillation Skill |
| `cairobot-task-raw-archive` | **上游依赖** | 蒸馏前优先读取 Raw Source Record（SRC-ID），通过 source-map.md 追溯跨来源关联 |
| `cairobot-daily-report` | **联动** | 蒸馏完成后可能需要更新日报 |
| `cairobot-doc-placement` | **依赖** | 新建文件时需遵守目录规范 |

**职责边界**：
- `tabbit-task-distillation`：单个任务的 raw → distilled 转换，输出到 `docs/reports/distilled/`
- `cairobot-scheduled-knowledge-distillation`：定义定时/自动化蒸馏的完整规则体系（含 SOLO Web 约束、SRC-ID/SOURCE-MAP 模型）
- `cairobot-llm-wiki-distillation`（本 Skill）：全局 Wiki 层的知识整理，输出到 `docs/wiki/daily/`、`docs/wiki/decisions/`、`docs/wiki/modules/`

## 3. 三层结构规则

### 3.1 Raw 层处理规则

- 原始材料**只读**，不可修改、不可删除、不可覆盖
- Raw 来源包括：
  - `docs/tabbit/inbox/` — Tabbit / TabAI 导出
  - `docs/trae-export/inbox/` — TRAE 执行记录
  - `docs/reports/daily/` — 每日日报
  - `docs/reports/testing/` — 测试报告
  - `docs/reports/audit/` — 审计报告
- 从 Raw 读取后，提取关键信息写入 Distillation 层

### 3.2 Distillation 层处理规则

- 每篇蒸馏文档必须区分：
  - ✅ **事实**：已验证的客观信息
  - ⚠️ **判断**：基于事实的主观分析（需标注置信度）
  - 🔴 **风险**：潜在问题或不确定事项
  - 📋 **规则**：经确认应遵守的约束
  - ➡️ **后续行动**：待办事项（不能写成已完成）

- **禁止行为**：
  - 不把计划写成完成
  - 不把设计写成实现
  - 不把 mock、TODO、空实现写成主链路完成
  - 不把未确认内容写成长期确定规则

### 3.3 Index 层处理规则

- Index 文件**只保留摘要和链接**
- 不承载完整日报、完整审计、完整对话或完整蒸馏正文
- LLM-WIKI.md 是导航入口，不是知识正文池
- 更新 Index 时同步更新相关引用

## 4. 每日自动化任务规则

### 4.1 触发条件

以下任一条件触发本 Skill：

- 每日 22:00 后的 Markdown 蒸馏流程
- SOLO Web 生成每日产物后的入库审核
- 手动触发的知识库整理需求

### 4.2 每日蒸馏流程

```
1. 扫描 Raw 层当日新增文件
   ├── docs/reports/daily/{今日日期}*.md
   └── docs/tabbit/inbox/{今日日期}/*.md

2. 读取并过滤噪声
   ├── 删除工具调用流水（toolName/status/json）
   ├── 删除重复对话片段
   ├── 删除 "*内容由 AI 生成仅供参考*" 免责声明
   └── 保留核心结论和决策点

3. 提炼稳定知识
   ├── 稳定决策 → docs/wiki/decisions/
   ├── 可复用流程 → docs/wiki/modules/
   ├── 日知识摘要 → docs/wiki/daily/
   └── 工程约束 → 对应 modules 页面

4. 更新 Index 层
   ├── 追加 每日蒸馏索引.md 条目
   ├── 如有新任务，追加 任务索引.md 条目
   └── 必要时更新 LLM-WIKI.md 进度摘要

5. 输出蒸馏报告
```

## 5. 输出格式

### 5.1 Distillation 文档模板

```markdown
# {知识标题}

## 来源

- Raw 来源：{路径}
- 蒸馏日期：{YYYY-MM-DD}
- 蒸馏执行者：{SOLO Web / Trae IDE / 手动}

## 事实摘要

{已验证的客观信息，用列表或表格}

## 判断与决策

{基于事实的分析，标注置信度}

## 规则沉淀

{从本次蒸馏中提取的可复用规则}

## 风险与限制

{已知的不确定性和限制}

## 后续行动

{待办事项，不写已完成}
```

### 5.2 Index 更新候选格式

当蒸馏完成后，生成可追加到 Index 文件的候选条目：

```markdown
## 每日蒸馏索引追加候选

| 日期 | Raw | 日报 | 蒸馏 | 主控汇报 | 状态 |
|---|---|---|---|---|---|
| {date} | {path} | {path} | {path} | {path} | 待主控确认 |
```

## 6. 主控确认机制

以下情况**必须等待主控确认**后方可写入长期 Index：

- 涉及架构决策变更的内容
- 涉及工程规则新增或修改的内容
- 涉及 SOLO Web 能力边界调整的内容
- 涉及已有 ADR 补充或修订的内容

以下情况**可直接写入**：

- 每日例行蒸馏产物追加到 每日蒸馏索引.md
- 已有任务的蒸馏状态更新（pending → distilled）
- LLM-WIKI.md 中进度摘要的文字更新

## 7. 完成前校验清单

蒸馏完成后确认：

- [ ] Raw 层文件未被修改或删除
- [ ] Distillation 文件写入了正确目录（docs/wiki/ 下）
- [ ] 蒸馏内容不含工具调用噪声
- [ ] 蒸馏内容不含未验证假设
- [ ] 计划/设计/mock/TODO 未写成已完成
- [ ] Index 层只追加了摘要和链接，未塞入完整正文
- [ ] 如涉及架构决策，已标注"待主控确认"

## 8. 联动 Skill

- 蒸馏后发现归档质量问题 → 激活 `cairobot-doc-placement`
- 需提交日报 → 激活 `cairobot-daily-report`
- 有 Tabbit pending manifest → 可联动 `tabbit-task-distillation`
