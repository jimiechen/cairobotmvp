# ADR-0009 Tabbit 任务归档与 Wiki 接入流程（v3：Canonical Task ID 自动生成版）

## 状态

已采纳（v3：Canonical Task ID 由 `/tabbit-task` 兜底自动生成）

## 背景

### 历史演进

| 版本 | 核心机制 | 遗留问题 |
|---|---|---|
| v1 | 语义化文件名 + 保留原始文件名 | 无法建立跨文件 task 追踪链路 |
| v2 | Task ID 驱动（由 Tabbit 生成） | 强依赖上游 Tabbit 输出规范 |
| **v3** | **Canonical Task ID 由 `/tabbit-task` 兜底自动生成** | — |

### v2 遗留问题

v2 要求 Tabbit 每次回复必须包含 `TABBIT_TASK_ID` 块。但实际流程中：

1. Tabbit 可能不生成 ID（历史文件、非标准流程）
2. 用户可能手动创建文件（如 `1.md`）无任何 ID 标记
3. TRAE 导出文件无 Task ID 信息
4. 一旦缺少 ID，整个链路断裂

**核心修正**：Task ID 的生成责任应落在"归档入口"（`/tabbit-task`），而非"上游对话工具"（Tabbit）。

## 核心原则

**一个任务，一个 Canonical Task ID，一组关联文件。ID 可外部传入，但必须由 `/tabbit-task` 保证存在。**

```
Tabbit 有 External ID → TRAE 沿用为 Canonical Task ID
Tabbit 无 ID          → /tabbit-task 自动生成 Canonical Task ID
```

## 决策

### Canonical Task ID 规范

格式：

```
TB-{YYYYMMDD}-{HHMMSS}-{topic-slug}
```

示例：

```
TB-20260518-193000-tabbit-auto-id-archive
```

### ID 生成优先级

| 优先级 | 条件 | 动作 |
|---|---|---|
| 1 | 文件内容含 `Task ID:` / `TABBIT_TASK_ID:` / `TB-*` 格式 | **沿用**该 ID 作为 Canonical Task ID |
| 2 | 用户显式提供 Task ID | **沿用**用户提供的 ID |
| 3 | 以上均无 | **自动生成** Canonical Task ID |

### topic-slug 自动提取规则

当需要自动生成 ID 时，按以下顺序提取 topic-slug：

1. Markdown 一级标题 (`# ...`)
2. 用户原始需求首句
3. 文件名语义（去除 `TabAI会话_`、时间戳等噪声）
4. 高频关键词
5. 任务产出主题

如果无法判断主题，使用：

```
untitled-tabbit-note
```

并在 archive 中标记 `needs_human_review: true`。

### 三种输入模式

#### 模式一：显式 ID 模式

```text
/tabbit-task

Task ID: TB-20260518-183504-tabbit-archive-distillation
来源文件：docs/tabbit/inbox/1.md
```

TRAE 沿用用户提供/文档中已有的 Task ID。

#### 模式二：自动 ID 模式

```text
/tabbit-task

来源文件：docs/tabbit/inbox/TabAI会话_1779079297836.md
```

TRAE 读取文件内容，自动生成 Canonical Task ID。

#### 模式三：批量迁移模式

```text
/tabbit-task

来源目录：docs/tabbit/inbox/
```

TRAE 扫描目录下所有未重命名文件，逐个理解内容、生成 ID、重命名、归档。

### 文件类型后缀规范

同一 Canonical Task ID 下，不同类型的文件使用不同后缀：

| 类型 | 后缀 | 存放位置 | 入库 |
|---|---|---|---|
| Tabbit / TabAI 原始导出 | `.tabbit.raw.md` | `docs/tabbit/inbox/{YYYY}/{MM}/` | ❌ |
| 手动创建文件 | `.manual.raw.md` | `docs/tabbit/inbox/{YYYY}/{MM}/` | ❌ |
| TRAE 执行导出 | `.trae.exec.md` | `docs/trae-export/inbox/{YYYY}/{MM}/` | ❌ |
| 正式归档文档 | `.archive.md` | `docs/wiki/tasks/{YYYY}/{MM}/` | ✅ |
| 任务清单（关联索引） | `.manifest.md` | `docs/wiki/tasks/{YYYY}/{MM}/` | ✅ |
| 蒸馏知识产物 | `.distilled.md` | `docs/reports/distilled/{YYYY}/{MM}/` | ✅ |

### 目录结构

```text
docs/
├── tabbit/                          # Tabbit / TabAI 原始导出收件箱
│   ├── README.md                    # 目录说明（入库）
│   └── inbox/                       # 原始导出文件（不入库）
│       └── {YYYY}/{MM}/             # 按年月组织
│           └── TB-*.tabbit/raw.md
├── trae-export/                     # TRAE 任务执行结果导出收件箱
│   ├── README.md                    # 目录说明（入库）
│   └── inbox/                       # TRAE 导出文件（不入库）
│       └── {YYYY}/{MM}/             # 按年月组织
│           └── TB-*.trae.exec.md
├── wiki/
│   ├── LLM-WIKI.md                  # Wiki 主入口
│   └── tasks/                       # 结构化归档 + manifest（入库）
│       └── {YYYY}/{MM}/
│           ├── TB-*.archive.md
│           └── TB-*.manifest.md
└── reports/
    └── distilled/                   # 蒸馏产物（入库）
        └── {YYYY}/{MM}/
            └── TB-*.distilled.md
```

### `/tabbit-task` 职责（v3）

#### 负责

1. 接收输入（文件路径 / 目录路径 / Task ID + 文件路径 / 纯文本任务）
2. 读取文件内容，理解文档主题
3. 判断是否已有 Task ID（检查内容 + 用户输入）
4. 如无 ID，自动生成 Canonical Task ID
5. 识别每个文件的类型后缀
6. 将所有关联文件重命名为 `{TASK_ID}.{type}.md`
7. 生成正式归档文件 `{TASK_ID}.archive.md`
8. 生成任务清单 `{TASK_ID}.manifest.md`
9. 生成 LLM-WIKI.md 索引候选
10. 输出执行后汇报

#### 不负责

1. 新建 `docs/llm-wiki/`
2. 知识蒸馏（由 `tabbit-task-distillation` Skill 负责）
3. 替代现有日报机制

### `tabbit-task-distillation` Skill 职责

1. 扫描 `docs/wiki/tasks/**/*.manifest.md` 中 `distillation_status = pending` 的条目
2. 按 Canonical Task ID 读取 raw / manual / exec / archive 文件
3. 过滤临时对话、工具噪声、重复内容
4. 提炼稳定知识
5. 生成 `{TASK_ID}.distilled.md`
6. 更新 LLM-WIKI.md 索引
7. 将 manifest 的 `distillation_status` 更新为 `distilled`

### Git 版本控制规则

| 路径 | 是否入库 | 说明 |
|---|---|---|
| `docs/tabbit/README.md` | ✅ 入库 | 收件箱说明 |
| `docs/tabbit/inbox/**/*.md` | ❌ 忽略 | 原始导出文件不作为长期知识资产 |
| `docs/trae-export/README.md` | ✅ 入库 | 收件箱说明 |
| `docs/trae-export/inbox/**/*.md` | ❌ 忽略 | TRAE 导出文件不作为长期知识资产 |
| `docs/wiki/tasks/**/*` | ✅ 入库 | 归档文档 + manifest |
| `docs/reports/distilled/**/*` | ✅ 入库 | 蒸馏产物 |

### CI 边界

| 路径 | 是否纳入 CI | 说明 |
|---|---|---|
| `docs/wiki/tasks/` | ❌ 暂不纳入 | 后续如需检查，单独设计 wiki-task-check |

## 影响

v3 将归档入口从"被动接收 ID"升级为"主动理解内容并生成 ID"：

- 不再依赖 Tabbit 或任何上游工具输出规范格式
- 所有历史文件均可通过批量迁移模式纳入体系
- `/tabbit-task` 成为真正的"自动建档命令"
- manifest 中的 `External Task ID` 字段保留外部 ID 追溯能力

## 替代方案

曾考虑以下方案但未采纳：

| 方案 | 内容 | 未采纳原因 |
|---|---|---|
| 方案 A | 新建 `docs/llm-wiki/` 完整体系 | 与现有体系重复 |
| 方案 B | 只用 `/distill-wiki` 不做归档 | 上下文丢失 |
| 方案 C | 把原始文件直接放入 `docs/wiki/` | 污染 Wiki |
| 方案 D (v1) | 语义化文件名 + 保留原始文件名 | 无法建立追踪链路 |
| 方案 E (v2) | Task ID 由 Tabbit 生成 | 强依赖上游，历史文件无法处理 |

## 与现有机制的对接关系

| 本决策产物 | 对接现有机制 | 说明 |
|---|---|---|
| `{TASK_ID}.archive.md` | `cairobot-html-distillation` Skill | 归档文档作为蒸馏输入原料之一 |
| `{TASK_ID}.manifest.md` | `tabbit-task-distillation` Skill | manifest 是蒸馏调度入口 |
| `{TASK_ID}.distilled.md` | `docs/reports/distilled/` + LLM-WIKI.md | 蒸馏产物进入报告和 Wiki |
| Wiki 索引候选 | `docs/wiki/LLM-WIKI.md` | 追加到已有索引结构中 |

## 后续动作

1. ✅ `.trae/commands/tabbit-task.md` v3（三种输入模式 + 自动 ID）
2. ✅ `.trae/skills/tabbit-task-distillation/SKILL.md`（适配 Canonical Task ID）
3. ✅ `docs/tabit/README.md` 和 `docs/trae-export/README.md`
4. ✅ `docs/wiki/LLM-WIKI.md` §15 更新
5. ⬜ 对 inbox 下 7 个历史文件执行批量迁移（模式三）
