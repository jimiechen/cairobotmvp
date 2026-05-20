# LLM Wiki 架构

## 1. 概述

本文是 LLM Wiki 的模块级知识页，说明 Raw → Distillation → Index 三层结构的详细设计。

## 2. 架构图

```
┌─────────────────────────────────────────────────────┐
│                  Index 目录层                        │
│  LLM-WIKI.md │ CODE-WIKI.md │ 每日蒸馏索引 │ 任务索引  │
│  职责：导航 / 摘要 / 索引 / 状态入口                  │
└──────────────────┬──────────────────────────────────┘
                   │ 引用
┌──────────────────▼──────────────────────────────────┐
│               Distillation 蒸馏层                    │
│  daily/ │ tasks/ │ decisions/ │ modules/             │
│  职责：提炼稳定知识 / 过滤噪声 / 区分事实与判断        │
└──────────────────┬──────────────────────────────────┘
                   │ 来源于
┌──────────────────▼──────────────────────────────────┐
│                  Raw 原始层                          │
│  tabbit/inbox/ │ trae-export/inbox/ │ reports/       │
│  职责：保真 / 不删 / 不覆盖 / 事实依据                │
└─────────────────────────────────────────────────────┘
```

## 3. 数据流向

```
Raw 层（事实输入）
  │
  ├── Tabbit/TabAI 会话导出 → docs/tabbit/inbox/ （保留原始文件名，不强求重命名）
  ├── TRAE 执行记录 → docs/trae-export/inbox/
  │   ├── 手动导出文件 → docs/trae-export/inbox/ （保留原始文件名）
  │   └── 结构化任务 Raw 归档 → docs/trae-export/inbox/tasks/YYYY/MM/TRAE-{timestamp}-{slug}.raw.md
  │       或 {source_record_id}.raw.md
  ├── 每日日报 → docs/reports/daily/
  ├── 测试报告 → docs/reports/testing/
  └── 审计/覆盖率报告 → docs/reports/audit/, coverage/
  │
  ▼
Distillation 层（知识加工）
  │
  ├── tabbit-task-distillation Skill → docs/reports/distilled/
  ├── llm-wiki-distillation Skill → docs/wiki/daily/, decisions/, modules/
  └── 手动审核确认
  │
  ▼
Index 层（导航服务 + 关联索引）
  │
  ├── LLM-WIKI.md（主入口）
  ├── 每日蒸馏索引.md
  ├── 任务索引.md
  ├── source-map.md（跨来源关联索引）
  └── CODE-WIKI.md
```

## 4. 文件命名规范

| 层级 | 命名模式 | 示例 |
|---|---|---|
| Raw - 日报 | `{YYYY-MM-DD}-{标题}.md` | `2026-05-17-HelloWorld验收日报.md` |
| Raw - Tabbit 导出 | **保留原始文件名**（不强求重命名） | 原始导出文件名不变 |
| Raw - TRAE 执行导出（SOLO Web） | SOLO Web 自行决定 | SOLO 命名体系 |
| Raw - TRAE 任务归档（手动） | `TRAE-{YYYYMMDD-HHMMSS}-{slug}.raw.md` 或 `{SRC-ID}.raw.md` | `TRAE-20260520-115001-task-impl.raw.md` |
| Distillation - 蒸馏 | `{TASK_ID}.distilled.md`（Task ID 可选） | `TB-20260518-181000.distilled.md` 或按 SRC-ID |
| Distillation - 归档 | `{TASK_ID}.archive.md` | 同上 |
| Distillation - Manifest | `{TASK_ID}.manifest.md` | 同上 |
| Index | 固定名称 | `LLM-WIKI.md`, `每日蒸馏索引.md`, `source-map.md` |

> **身份模型**：
> - **Source Record ID (SRC-ID)**：Raw 层主键，格式 `SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}`
> - **Task ID (TB-ID)**：可选字段，仅在主控指定或已有明确任务链时使用
> - 已有 TB-ID 记录保留兼容，标记为 legacy
>
> **来源保真原则**：
> - Tabbit / Trae 原始导出文件的文件名**不得为了对齐 Task ID 而改名**
> - 结构化 Trae 任务记录推荐使用 `TRAE-*` 前缀或 SRC-ID
>
> **路径约束**：所有 Trae 手动归档的 `.raw.md` 文件必须放在 `docs/trae-export/inbox/tasks/YYYY/MM/`，禁止放入 `docs/wiki/`。
>
> **跨来源关联**：通过 `docs/wiki/source-map.md` 的 Relation Group (RG) 机制维护，详见 source-map.md。

## 5. 与 ADR-0009 的关系

本三层架构与 ADR-0009（Tabbit 任务归档流程）的关系：

| ADR-0009 概念 | 三层架构对应 |
|---|---|
| 原始导出文件 (.raw.md) | Raw 层 |
| 归档文档 (.archive.md) | Distillation 层（tasks/） |
| Manifest (.manifest.md) | Distillation 层（tasks/） |
| 蒸馏产物 (.distilled.md) | Distillation 层（daily/ 或 reports/distilled/） |
| LLM-WIKI.md 索引条目 | Index 层 |

两者互补，不冲突。ADR-0009 管理单个任务的生命周期，三层架构管理整体知识库的组织方式。

## 6. 定时蒸馏数据流

### 6.1 标准蒸馏流程（5 阶段）

```
Phase 1: Raw 层扫描
  ├─ 扫描 docs/tabbit/inbox/{date}/
  ├─ 扫描 docs/trae-export/inbox/{date}/
  ├─ 扫描 docs/trae-export/inbox/tasks/{date}/
  ├─ 扫描 docs/reports/daily/{date}
  └─ 为每条记录分配/复用 Source Record ID (SRC-ID)
      │
      ▼
Phase 2: 去噪与过滤
  ├─ 删除工具调用流水
  ├─ 删除重复片段和免责声明
  └─ 保留核心结论、决策点、产物清单
      │
      ▼
Phase 3: 知识提取（Distillation 层，候选态）
  ├─ 稳定决策 → wiki/decisions/
  ├─ 可复用流程 → wiki/modules/
  ├─ 日知识摘要 → wiki/daily/
  └─ 每条产物携带 SRC-ID + RG 元数据 + candidate 态标记
      │
      ▼
Phase 4: Index 候选生成（不直接覆盖正式 Index）
  ├─ 每日蒸馏索引候选
  ├─ Source Map 候选（新增关联，全部 candidate）
  ├─ 任务索引候选（使用 SRC-ID / RG）
  └─ 输出到独立候选文件
      │
      ▼
Phase 5: 输出报告
  ├─ 蒸馏产物清单（含路径和 SRC-ID）
  ├─ Index 候选清单
  ├─ 待确认项
  └─ 是否允许进入正式 Index
```

### 6.2 触发方式

| 方式 | 入口 | 说明 |
|---|---|---|
| SOLO Web 定时任务 | SOLO Web 自动触发 | Prompt 必须引用 `cairobot-scheduled-knowledge-distillation` Skill |
| Trae IDE 手动补跑 | `/daily-distill` Command | 支持指定日期 `--date` 和来源范围 `--sources` |

### 6.3 关键约束

- ✅ Raw 文件只读不改
- ✅ 蒸馏产物为 **candidate** 态
- ✅ Index 更新为**候选追加**，不直接覆盖
- ✅ 使用 **Source Record ID (SRC-ID)** 追溯 Raw 来源
- ❌ SOLO Web 不得 auto commit / push
- ❌ 不得直接修改 LLM-WIKI.md
- ❌ 不得把 candidate 关联写成 confirmed

详见 [cairobot-scheduled-knowledge-distillation Skill](../../.trae/skills/cairobot-scheduled-knowledge-distillation/SKILL.md)。

## 7. 与 ADR-0009 Task ID 的关系
