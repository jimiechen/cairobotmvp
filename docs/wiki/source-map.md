# Source Map（跨来源关联索引）

## 1. 说明

本文是 LLM Wiki 三层结构中的 **Index 层**文件，用于维护 Tabbit、Trae、SOLO Web、日报、蒸馏文件之间的跨来源关联关系。

**核心原则**：
- 不同来源的文件命名体系和粒度差异很大，**不得强制对齐 Task ID**
- 关联必须显式声明，**禁止假关联**
- 语义相似 ≠ 同一任务，只能标记为 `semantic` 或 `candidate`

详见 `cairobot-task-raw-archive` Skill §6。

## 2. Relation Group 列表

### RG-20260520-001 — Trae 任务 Raw 归档规范修正（身份模型重构）

| # | SRC-ID | 来源类型 | 来源文件路径 | 标题或摘要 | 关联方式 | 置信度 | 状态 | 说明 |
|---|---|---|---|---|---|---|---|---|
| 1 | SRC-TABBIT-20260520-114812-a13f9c02 | tabbit | `docs/tabbit/inbox/2026/05/TabAI会话_1779206504027.md` | Trae 任务 Raw 归档规范设计方案（原始方案） | exact | high | confirmed | 本 RG 的原始需求来源 |
| 2 | SRC-TRAE-20260520-000756-b82ad119 | trae | `docs/reviews/2026-05-20-trae-task-raw-archive-review.md` | 评审意见文档（3 必须修改 + 4 建议） | manual | high | confirmed | 主控评审输出 |
| 3 | SRC-TRAE-20260520-001200-c91eaa77 | trae | `docs/trae-export/inbox/tasks/2026/05/TB-20260520-000756-trae-task-raw-archive.trae.raw.md` | 首次实施 Raw 归档（legacy TB-ID） | manual | high | confirmed | 使用旧 TB-ID 格式的首次实施 |
| 4 | SRC-WIKI-20260520-001500-d33f88a2 | wiki | `docs/wiki/decisions/llm-wiki-three-layer-architecture.md` | 三层架构决策（已更新：Raw 身份模型） | manual | high | confirmed | 决策文档同步更新 |
| 5 | SRC-WIKI-20260520-001550-e44f99b3 | wiki | `docs/wiki/modules/llm-wiki-architecture.md` | 架构模块知识页（已更新：命名规范+Source Map） | manual | high | confirmed | 模块文档同步更新 |

> **关于本 RG 的说明**：本组关联全部为 `manual` + `confirmed`，因为所有文件均由同一次主控裁决驱动生成。这是 Source Map 的首批种子数据。

## 3. Relation Group 定义

```
RG-{YYYYMMDD}-{NNN}
```

| 字段 | 说明 | 示例 |
|---|---|---|
| `RG` | 固定前缀 | `RG` |
| `{YYYYMMDD}` | 关联组创建日期 | `20260520` |
| `{NNN}` | 当日序号（3 位，从 001 开始） | `001` |

## 4. 关联方式定义

| 关联方式 | 含义 | 可升级为 | 降级为 |
|---|---|---|---|
| `exact` | 同一物理文件的不同版本或重命名 | — | `manual` |
| `manual` | 主控手动确认属于同一任务 | — | `semantic` |
| `semantic` | 内容语义高度相似（AI 判断） | `manual`（需主控确认） | `candidate` |
| `candidate` | 可能有关联但未确认 | `semantic`（AI 增强） | `none` / `rejected` |
| `none` | 无明确关联 | `candidate`（发现新线索） | — |
| `rejected` | 已明确否定关联 | — | — |

## 5. 置信度定义

| 置信度 | 含义 |
|---|---|
| `high` | 有明确证据支持（主控确认、同一裁决驱动等） |
| `medium` | 有间接证据（时间重叠、关键词匹配等） |
| `low` | 仅模型推测，无外部验证 |

## 6. 状态定义

| 状态 | 含义 |
|---|---|
| `confirmed` | 关联已被主控或可靠证据确认 |
| `pending` | 待人工审核 |
| `rejected` | 已被否决 |

## 7. 关联禁止事项（硬规则）

以下行为**严格禁止**：

1. ❌ 不得仅凭 Task ID 判断 Tabbit 与 Trae 属于同一任务
2. ❌ 不得为了对齐而改写 Raw 原始文件名
3. ❌ 不得把语义相似写成 confirmed（只能写 semantic 或 candidate）
4. ❌ 不得把模型推测关联写成事实
5. ❌ 不得把 candidate 关联进入长期规则
6. ❌ 不得要求 Trae 执行时识别 Tabbit 导出的时间戳
7. ❌ 不得要求 Tabbit 导出文件反向匹配 Trae 的中文摘要标题

违反以上任何一条视为**假关联违规**，必须在 source-map.md 中标注并上报主控。

## 8. 维护规则

- 新增关联时必须指定 RG-ID（可复用已有 RG 或创建新 RG）
- 每条记录必须包含完整的 9 个字段
- candidate 关联在 7 天内未升级为 semantic/manual 则降级为 none
- rejected 关联必须说明拒绝理由
- 不得删除历史记录（包括 rejected），只追加新状态
