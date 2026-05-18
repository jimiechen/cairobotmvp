# Tabbit 任务归档流程设计与评审

## 元信息

- 来源：Tabbit + TRAE
- 原始文件：`docs/tabbit/TabAI会话_1779097841153.md`、`docs/tabbit/TabAI会话_1779098249005.md`
- 创建时间：2026-05-18 18:10:00
- 归档路径：`docs/wiki/tasks/2026/05/2026-05-18_181000_tabbit-task-archive-flow-design.md`
- 任务状态：已完成
- 标签：#tabbit #wiki-task #adr #slash-command

## 原始任务

用户希望解决 TRAE 每次执行 Tabbit 任务后不会自动生成 Markdown 文档的问题。目前需要手动复制结果、新建 `1.md` 保存任务结果，导致：
1. 任务上下文丢失，无法知道文档来自哪个 Tabbit 任务。
2. 文件名不可读（如 `TabAI会话_1779079297836.md`），无法表达主题。
3. 无法进入后续 LLM Wiki 知识蒸馏流程。

同时，TRAE 自身导出的文件（存放在 `docs/trae-export/`）也面临同样问题。

## 任务理解

本任务的核心目标是建立一条从"原始 AI 会话导出"到"结构化 Wiki 任务归档"的工程化流水线。

关键约束：
1. **不能另起炉灶** — 必须复用现有 `docs/wiki/`、`docs/reports/`、Skill 体系。
2. **两个收件箱** — `docs/tabbit/`（Tabbit/TabAI 导出）和 `docs/trae-export/`（TRAE 导出）都需要覆盖。
3. **最小可行** — 先跑通 `/tabbit-task` 一个命令，暂不新增 `/distill-wiki`。

## 执行过程

### 步骤一：原始方案评审

读取 `docs/tabbit/TabAI会话_1779097841153.md` 中的原始方案，该方案提议：
- 新建 `docs/llm-wiki/` 完整目录体系
- 创建 `/tabbit-task` 和 `/distill-wiki` 两个命令
- 建立独立的蒸馏和 Wiki 索引机制

### 步骤二：TRAE 评审

对原始方案进行工程规范评审，发现 4 个必须修改项：
1. ❌ `docs/llm-wiki/` 与现有 `docs/wiki/` 冲突（R0）
2. ❌ 与现有日报/蒸馏/Skill 机制完全未对齐（R0）
3. ❌ 缺少 ADR 支撑（R1）
4. ❌ 未考虑 CI 集成（R2）

结论为"建议修改"。

### 步骤三：架构师修订

读取 `docs/tabbit/TabAI会话_1779098249005.md`（架构师修订意见），采纳核心判断：
- 不新建 `docs/llm-wiki/`，改为复用 `docs/wiki/` 扩展
- `docs/tabbit/` 定位为 inbox，最终归档进入 `docs/wiki/tasks/`
- 删除 `/distill-wiki`，蒸馏继续复用现有 Skill
- 补充轻量 ADR-0009
- 同时覆盖 `docs/trae-export/` 的 TRAE 导出文件

### 步骤四：落地实施

按修订方案依次创建：
1. ADR-0009 决策记录
2. `.trae/commands/tabbit-task.md`（瘦身版）
3. `docs/tabbit/README.md`
4. `docs/trae-export/README.md`
5. `docs/wiki/tasks/2026/05/` 首次归档（本文件）
6. 更新 `docs/wiki/LLM-WIKI.md`
7. 更新 `.trae/rules/docs.md`

## 最终产出

### 新增文件

| 类型 | 路径 | 说明 |
|---|---|---|
| 新建 | `docs/adr/ADR-0009-tabbit-task-archive-flow.md` | 决策记录 |
| 新建 | `.trae/commands/tabbit-task.md` | 归档命令 |
| 新建 | `docs/tabbit/README.md` | 收件箱说明 |
| 新建 | `docs/trae-export/README.md` | 收件箱说明 |
| 新建 | `docs/wiki/tasks/2026/05/2026-05-18_181000_tabbit-task-archive-flow-design.md` | 本次归档（本文件） |

### 修改文件

| 类型 | 路径 | 说明 |
|---|---|---|
| 更新 | `docs/wiki/LLM-WIKI.md` | 增加 tasks 归档入口 |
| 更新 | `.trae/rules/docs.md` | 补充 tabbit / trae-export 目录登记 |

### 最终目录结构

```text
docs/
├── tabbit/                          # Tabbit/TabAI 导出收件箱
│   ├── README.md                    # ← 新增
│   ├── 1.md                         # 待归档
│   ├── TabAI会话_1779079297836.md    # 待归档
│   ├── TabAI会话_1779093720973.md    # 待归档
│   └── TabAI会话_1779097841153.md    # 已作为本次输入
├── trae-export/                     # TRAE 导出收件箱
│   ├── README.md                    # ← 新增
│   └── 评审TabAI会话导出方案.md      # 待归档
└── wiki/
    ├── LLM-WIKI.md                  # ← 更新
    └── tasks/                       # ← 新增目录
        └── 2026/05/
            └── 2026-05-18_181000_tabbit-task-archive-flow-design.md  # ← 本文件
```

## 文件变更建议

| 类型 | 路径 | 说明 |
|---|---|---|
| 新建 | `docs/adr/ADR-0009-tabbit-task-archive-flow.md` | 决策记录 |
| 新建 | `.trae/commands/tabbit-task.md` | Slash Command |
| 新建 | `docs/tabbit/README.md` | 收件箱定位说明 |
| 新建 | `docs/trae-export/README.md` | 收件箱定位说明 |
| 新建 | `docs/wiki/tasks/2026/05/*.md` | 归档目录 |
| 更新 | `docs/wiki/LLM-WIKI.md` | 追加 tasks 入口段 |
| 更新 | `.trae/rules/docs.md` | 补充目录登记 |

## Wiki 索引候选

- [Tabbit 任务归档流程设计与评审](./tasks/2026/05/2026-05-18_181000_tabbit-task-archive-flow-design.md)
  - 时间：2026-05-18 18:10:00
  - 来源：Tabbit + TRAE
  - 标签：#tabbit #wiki-task #adr #slash-command
  - 摘要：设计并落地了从 Tabbit/TRAE 原始导出到 docs/wiki/tasks 的归档流程，含 ADR-0009 和 /tabbit-task Command

## 后续处理

1. 用 `/tabbit-task` 对 `docs/tabbit/1.md`、`TabAI会话_1779079297836.md`、`TabAI会话_1779093720973.md` 执行批量归档。
2. 用 `/tabbit-task` 对 `docs/trae-export/评审TabAI会话导出方案.md` 执行归档。
3. 后续每次 Tabbit 任务或 TRAE 导出后，立即触发 `/tabbit-task`。
