---
name: 任务 Raw 归档
slug: task-raw-archive
summary: 将每次 Trae 执行任务的提示词、执行结果、产物清单和待确认项归档到 Raw 层，作为后续蒸馏的事实依据。Trae 任务完成时必须激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - raw-archive
  - trae-task
  - src-id
  - source-map
trigger:
  - "任务 Raw 归档"
  - "task raw archive"
  - "Raw 归档"
  - "任务完成汇报持久化"
  - "trae task archive"
priority: high
blocking: true
---

# CaiRobot MVP Trae 任务 Raw 归档 Skill

## 1. Skill 职责

本 Skill 强制要求每次由 Trae IDE 执行的任务，都必须将**任务提示词、执行结果、产物清单**结构化落盘到 Raw 层，保证后续 Distillation 和 Index 有可追溯的事实依据。

**负责**：
- 定义 Trae 任务必须归档 Raw 的硬规则
- 规定 Raw 记录的必填内容、格式和存放路径
- 明确与 `reporting.md` / `project.md` 任务完成汇报的关系
- 指导 `/task-raw-archive` Command 的正确使用

**不负责**：
- 任务执行本身（由具体业务 Skill 负责）
- 蒸馏过程（由 `cairobot-llm-wiki-distillation` 负责）
- 日报提交（由 `cairobot-daily-report` 负责）
- 业务代码修改

详细规则参见：
- [LLM Wiki 三层架构决策](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md)
- [LLM Wiki 架构模块](../../docs/wiki/modules/llm-wiki-architecture.md)
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md)
- [.trae/rules/project.md](../../.trae/rules/project.md)

## 2. 与其他 Skill 的关系

| Skill | 关系 | 说明 |
|---|---|---|
| `cairobot-daily-report` | **联动** | Raw 归档是任务完成汇报的结构化持久化版本，不是额外重复流程 |
| `cairobot-llm-wiki-distillation` | **下游依赖** | 蒸馏时优先从 Raw 文件读取 Trae 执行上下文 |
| `cairobot-active-gap-filling` | **上游触发** | 缺口扫描中检查"本次任务是否已生成 Trae Raw 归档" |
| `cairobot-doc-placement` | **约束** | Raw 文件路径必须符合目录规范 |
| `tabbit-task-distillation` | **并行互补** | 该 Skill 处理 Tabbit 任务级蒸馏，本 Skill 处理 Trae 任务级 Raw 归档 |

## 3. 核心硬规则

### 3.1 归档强制规则

1. **每次 Trae 执行任务，都必须生成 Raw 归档文件。无例外。**
2. Raw 归档必须在任务完成后、日报前生成。
3. 未生成 Raw 归档的任务不得宣称"已完成"。

### 3.2 身份模型：Source Record ID（非 Task ID）

**Raw 层不强制使用 Canonical Task ID 作为主键。**

Raw 层使用 **Source Record ID (SRC-ID)** 作为每条记录的唯一标识：

```
SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}
```

| 字段 | 说明 | 示例 |
|---|---|---|
| `SRC` | 固定前缀 | `SRC` |
| `{SOURCE}` | 来源类型：`TABBIT` / `TRAE` / `SOLO` / `REPORT` / `MANUAL` | `TRAE` |
| `{YYYYMMDD-HHMMSS}` | 记录生成时间戳 | `20260520-115001` |
| `{HASH8}` | 内容哈希前 8 位（基于文件路径+时间戳） | `b82ad119` |

**完整示例**：
```
SRC-TRAE-20260520-115001-b82ad119
SRC-TABBIT-20260520-114812-a13f9c02
SRC-SOLO-20260519-223000-c91eaa77
```

**Task ID 处理规则**：
- Task ID（`TB-*` 格式）作为**可选字段**保留
- 仅在主控明确指定或已有明确任务链时使用
- 不强制要求每条 Raw 记录都有 Task ID
- 已有的 TB-ID 记录保留兼容，标记为 legacy

### 3.3 路径强制规则

```
docs/trae-export/inbox/tasks/YYYY/MM/
```

该目录下存放所有结构化的 Trae 任务 Raw 归档文件。

**文件命名优先级**：

| 优先级 | 命名格式 | 适用场景 |
|---|---|---|
| 1（推荐） | `TRAE-{YYYYMMDD-HHMMSS}-{slug}.raw.md` | Trae IDE 手动归档的结构化任务记录 |
| 2 | `{source_record_id}.raw.md` | 无 slug 或自动生成场景 |
| 3（legacy） | `{TASK_ID}.trae.raw.md` | 早期已生成的记录，不再强制用于新记录 |

**禁止事项**：
- ❌ 不允许把 Raw 记录放入 `docs/wiki/`（wiki/ 只承载 Distillation 层和 Index 层）
- ❌ 不允许放入 `docs/reports/raw-tasks/`（除非主控另行裁决）
- ❌ 不允许强制使用 `TB-{YYYYMMDD}-{HHMMSS}-{slug}.trae.raw.md` 格式
- ❌ 不允许为了对齐 Task ID 而改写 Tabbit 或 Trae 原始导出文件的文件名

### 3.4 内容必填规则

每条 Raw 记录**必须**包含以下 9 个部分：

| # | 部分 | 必填内容 | 可否省略 |
|---|---|---|---|
| 1 | 基本信息 | **Source Record ID**（必填）、Task ID（可选）、日期时间、执行者、任务类型、状态、Original Filename、Content Hash | SRC-ID 不可省略 |
| 2 | 任务提示词 | 用户原始要求、主控追加要求、约束条件、禁止事项 | 提示词不可省略 |
| 3 | 输入材料 | 读取的文件、参考文档、使用的目录和命令 | 无则写"无" |
| 4 | 执行结果 | 已完成项、未完成项、失败项、不确定项 | 无则写"无" |
| 5 | 产物清单 | 新增/修改/删除文件列表、未提交变更说明 | 必须列出，即使为空 |
| 6 | 测试或检查结果 | 运行的命令、测试结果摘要 | 无测试则说明原因 |
| 7 | 待确认项 | 需主控确认的内容、是否允许提交 | 无则写"无待确认项" |
| 8 | 主控结论 | 已确认内容 | 初始归档时可留空等待确认 |
| 9 | 后续蒸馏建议 + 关联信息 | 可进入 Distillation 的内容、不应进入长期知识库的内容、Relation Group（可选） | 蒸馏建议必须，RG 可选 |

### 3.4 保真规则

1. **Raw 层保真**：原始材料不可篡改、不可覆盖、不可删除
2. **提示词完整记录**：用户原始提示词必须原样保存，不得省略或概括
3. **结果如实记录**：成功就是成功，失败就是失败，不允许美化
4. **产物如实记录**：没有产物就写"无新增产物"，不允许伪造
5. **未确认标注清楚**：不确定的事项必须明确标注，不能混入已确认内容

### 3.5 与任务完成汇报的关系

**Raw 归档 = 任务完成汇报的结构化持久化版本**

| 维度 | 任务完成汇报（reporting.md） | Raw 归档（本 Skill） |
|---|---|---|
| 输出时机 | 任务完成后输出到终端/聊天窗口 | 任务完成后写入文件系统 |
| 格式 | 自由文本 + 表格 | 结构化 Markdown 模板 |
| 生命周期 | 会话结束后可能丢失 | 持久化存储，可被后续蒸馏引用 |
| 受众 | 项目主控（即时阅读） | AI 动手手 + 蒸馏程序（机器可读） |

**关键约束**：
- Raw 归档**不是**要求 Trae 重新写一遍完整汇报
- 如果任务完成汇报已包含完整信息，`/task-raw-archive` Command 可以直接从汇报内容生成 Raw 文件
- 后续日报、蒸馏、任务索引应**优先引用**该 Raw 文件，而不是依赖聊天窗口上下文

### 3.6 轻量任务简化规则

对于小任务（如只改一个 typo、只更新一个文档），允许使用轻量模板，但以下内容**仍不可省略**：
- ✅ 任务提示词（至少摘要）
- ✅ 执行结果（至少一行）
- ✅ 产物状态（新增/修改了哪些文件）
- ✅ Source Record ID 和路径
- ✅ Original Filename（原始文件名，如有）

可以简化的部分：
- 输入材料可以合并为一行
- 测试结果可以写"本次变更无需测试（文档类）"

## 4. 文件命名规范

### 4.1 Source Record ID 格式（Raw 层主键）

```
SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}
```

**核心原则**：每条 Raw 记录有且只有一个 SRC-ID。SRC-ID 由记录生成时自动创建，不依赖外部 Task ID。

### 4.2 Trae 结构化归档文件命名

**推荐格式**：
```
TRAE-{YYYYMMDD-HHMMSS}-{slug}.raw.md
```

示例：
```
TRAE-20260520-115001-task-raw-archive-impl.raw.md
```

**备选格式**（无 slug 场景）：
```
{source_record_id}.raw.md
```

示例：
```
SRC-TRAE-20260520-115001-b82ad119.raw.md
```

### 4.3 来源文件保留原貌

以下文件的原始文件名**不得为了对齐 Task ID 而改名**：

| 来源 | 原始存放位置 | 命名规则 |
|---|---|---|
| Tabbit 导出 | `docs/tabbit/inbox/` | **保留原始文件名**，不强制重命名 |
| Trae 导出（手动） | `docs/trae-export/inbox/` | **保留原始文件名**或使用 TRAE-* 前缀 |
| SOLO Web 自动生成 | `docs/trae-export/inbox/` | SOLO Web 自行决定 |

### 4.4 后缀区分

| 后缀 | 来源 | 说明 |
|---|---|---|
| 原始文件名（不变） | Tabbit / TabAI 导出 | 保留导出时的原始名称 |
| `.trae.exec.md` | SOLO Web 自动生成的执行记录 | SOLO Web 命名体系 |
| `.raw.md` / `*.raw.md` | Trae IDE 手动归档的结构化任务记录 | 使用 TRAE-* 或 SRC-ID 前缀 |

## 5. Raw 记录模板

```markdown
# Trae 任务 Raw 记录

## 1. 基本信息
- **Source Record ID**：SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}
- **Task ID**：（可选，如有）TB-{YYYYMMDD}-{HHMMSS}-{slug}
- **Legacy Task ID**：（仅早期记录）TB-{...}
- **日期时间**：YYYY-MM-DD HH:MM:SS
- **执行者**：Trae IDE
- **任务类型**：daily / task / refactor / audit / distillation / docs
- **状态**：已执行 / 部分完成 / 待确认 / 失败
- **Original Filename**：（原始文件名，如有）
- **Content Hash**：（内容哈希摘要）
- **Relation Group**：（可选）RG-{YYYYMMDD}-{NNN}

## 2. 任务提示词
### 用户原始要求
{粘贴用户原始输入}

### 主控追加要求
{如有}

### 约束条件
{如有}

### 禁止事项
{如有}

## 3. 输入材料
- **读取的文件**：
  - {path} — {用途}
- **参考的文档**：
  - {path} — {用途}
- **使用的目录**：
  - {dir}
- **使用的命令**：
  - `{command}` — {目的}

## 4. 执行结果
### 已完成
- {item}

### 未完成
- {item} — {原因}

### 失败
- {item} — {原因}

### 不确定
- {item} — {疑点}

## 5. 产物清单
### 新增文件
| 文件路径 | 说明 |
|---|---|

### 修改文件
| 文件路径 | 变更说明 |
|---|---|

### 删除文件
| 文件路径 | 原因 |
|---|---|

### 未提交变更说明
{git diff --stat 摘要}

## 6. 测试或检查结果
### 运行的命令
```bash
{command}
```

### 结果摘要
{pass/fail/skip + 关键输出}

## 7. 待确认项
- {需要主控确认的内容}
- **是否允许提交**：是 / 否 / 待确认

## 8. 主控结论
{初始归档时留空，等待主控确认后补充}

## 9. 后续蒸馏建议
### 可进入 Distillation 层的内容
- {稳定知识条目}

### 不应进入长期知识库的内容
- {临时性、上下文相关、未确认的内容}

### Index 更新候选
- 是否需要更新 `任务索引.md`：是 / 否
- 是否需要更新 `LLM-WIKI.md`：是 / 否
- 是否需要更新 `source-map.md`：是 / 否
```

## 6. Source Map 与 Relation Group

### 6.1 为什么需要 Source Map

Raw 层中 Tabbit、Trae、SOLO Web、日报、蒸馏文件来自不同来源，命名体系和粒度差异很大：
- Tabbit 导出通常是时间戳命名
- Trae 导出通常带中文标题
- 二者无法可靠地通过 Task ID 强制对齐
- 强制对齐会制造**假关联**，破坏 Raw 层保真原则

因此引入 **Source Map（`docs/wiki/source-map.md`）** 作为跨来源关联索引。

### 6.2 Relation Group (RG)

```
RG-{YYYYMMDD}-{NNN}
```

| 字段 | 说明 | 示例 |
|---|---|---|
| `RG` | 固定前缀 | `RG` |
| `{YYYYMMDD}` | 关联组创建日期 | `20260520` |
| `{NNN}` | 当日序号（3 位，从 001 开始） | `001` |

每个 Relation Group 包含一条或多条关联记录。

### 6.3 关联记录格式

每条 Source Map 记录包含：

| # | 字段 | 必填 | 说明 |
|---|---|---|---|
| 1 | RG-ID | ✅ | Relation Group ID |
| 2 | SRC-ID | ✅ | 本记录的 Source Record ID |
| 3 | 来源文件路径 | ✅ | 完整相对路径 |
| 4 | 来源类型 | ✅ | `tabbit` / `trae` / `solo` / `report` / `wiki` |
| 5 | 标题或摘要 | ✅ | 文件内容的一行摘要 |
| 6 | 关联方式 | ✅ | 见下方定义 |
| 7 | 置信度 | ✅ | `high` / `medium` / `low` |
| 8 | 状态 | ✅ | `confirmed` / `pending` / `rejected` |
| 9 | 说明 | — | 关联理由或备注 |

### 6.4 关联方式定义

| 关联方式 | 含义 | 使用场景 |
|---|---|---|
| `exact` | 同一物理文件的不同版本或重命名 | 文件被移动/重命名后的追踪 |
| `manual` | 主控手动确认属于同一任务 | 主控明确指定关联 |
| `semantic` | 内容语义高度相似 | AI 判断为同一主题的不同来源记录 |
| `candidate` | 可能有关联但未确认 | 待人工审核的候选关联 |
| `none` | 无明确关联 | 独立记录，不与其他来源绑定 |

### 6.5 关联禁止事项

以下行为**严格禁止**：
1. ❌ 不得仅凭 Task ID 判断 Tabbit 与 Trae 属于同一任务
2. ❌ 不得为了对齐而改写 Raw 原始文件名
3. ❌ 不得把语义相似写成 confirmed（只能写 semantic 或 candidate）
4. ❌ 不得把模型推测关联写成事实
5. ❌ 不得把 candidate 关联进入长期规则
6. ❌ 不得要求 Trae 执行时识别 Tabbit 导出的时间戳
7. ❌ 不得要求 Tabbit 导出文件反向匹配 Trae 的中文摘要标题

## 7. 自动化触发机制（软约束）

### 7.1 在工程工作流中的嵌入位置

以下 Skill 在执行时应检查 Raw 归档状态：

| 检查位置 | 检查方式 |
|---|---|
| `cairobot-active-gap-filling` 缺口扫描 | 第 11 项："本次任务是否已生成 Trae Raw 归档？" |
| `cairobot-engineering-workflow` 闭环校验 | 完成前确认 Raw 文件已写入 |
| `cairobot-llm-wiki-distillation` 蒸馏前 | 优先查找对应 `*.raw.md` / `*.trae.raw.md` 作为事实来源 |
| `cairobot-daily-report` 日报生成 | 引用 Raw 文件路径作为当日依据 |

### 7.2 CI 集成计划（未来）

MVP 阶段 CI report-check **暂不强制定**，后续稳定后加入抽查：
- 抽查当日 `docs/trae-export/inbox/tasks/` 下是否有新增 Raw 文件
- 抽查 Raw 文件是否包含必填的 9 个部分
- 抽查是否包含 Source Record ID

## 8. 未来演进说明

### MVP 阶段（当前）
- 单 Command：`/task-raw-archive`，任务完成时一次性归档
- 软约束：靠 Skill 规则驱动，CI 不硬卡
- Source Map 手动维护

### S1/S2 阶段（可选演进）
如任务频率升高，可拆分为两个 Command：
- `/task-start`：任务开始时建档（记录提示词、约束、生成 SRC-ID）
- `/task-close`：任务完成时补全（执行结果、产物清单、测试结果、待确认项）

拆分前提：
- 存在跨调用状态维护机制
- SRC-ID 在 start 阶段即可确定
- 主控确认拆分收益大于复杂度成本

## 9. 完成前硬校验清单

每次执行 `/task-raw-archive` 后确认：

- [ ] Raw 文件已写入 `docs/trae-export/inbox/tasks/YYYY/MM/` 目录
- [ ] 文件名**不强制**使用 TB-* 格式，优先使用 TRAE-* 或 SRC-ID 格式
- [ ] 包含 **Source Record ID**（SRC-{SOURCE}-{timestamp}-{hash8}）
- [ ] 包含全部 9 个必填部分（基本信息 → 后续蒸馏建议+关联信息）
- [ ] 任务提示词已原样记录（未省略、未概括）
- [ ] 产物清单已列出（包括"无新增产物"的显式声明）
- [ ] 待确认项已明确写出（或声明"无待确认项"）
- [ ] 文件**不在** `docs/wiki/` 下
- [ ] 未确认内容未被写成已确定的长期规则
- [ ] 最终汇报中输出了 Raw 文件路径和 **Source Record ID**
- [ ] 未制造假关联（未强行将不同来源的文件绑定到同一 Task ID）

## 10. 违规阻断

以下行为视为违规，必须立即停止：

- 宣称任务完成但未生成 Raw 归档文件
- 将 Raw 文件放入 `docs/wiki/` 目录
- 省略任务提示词或产物清单
- 把未确认内容提升为长期规则
- 用"无实质变化"跳过归档（即使是纯分析任务也必须记录）
- **仅凭 Task ID 判断 Tabbit 与 Trae 属于同一任务**
- **为了对齐 Task ID 改写 Tabbit 或 Trae 原始导出文件名**
- **把语义相似关联标记为 confirmed（只能用 semantic 或 candidate）**
- **把模型推测的跨来源关联写入长期规则**
