# Tabbit 归档方案升级为 Task ID 驱动架构

## 元信息

- Task ID: TB-20260518-190000-tabbit-task-id-upgrade
- 来源：Tabbit + 架构师 + TRAE
- 原始文件：
  - `docs/tabbit/inbox/TabAI会话_1779100587049.md`（架构师 Task ID 升级修订）
- 创建时间：2026-05-18 19:00:00
- 归档路径：`docs/wiki/tasks/2026/05/TB-20260518-190000-tabbit-task-id-upgrade.archive.md`
- 任务状态：已完成
- 标签：#tabbit #wiki-task #task-id #manifest #distillation

## 原始任务

用户要求对 `docs/tabbit/inbox/TabAI会话_1779100587049.md` 中的架构师修订意见（Task ID 驱动方案）进行实施。核心变更：

1. 从"语义化文件名 + 保留原始文件名"升级为 **Task ID 驱动的任务资产链**。
2. 新增 Task ID 规范：`TB-{YYYYMMDD}-{HHMMSS}-{topic-slug}`。
3. 新增 manifest 机制作为蒸馏调度入口。
4. 新增 `tabbit-task-distillation` Skill。
5. inbox 目录增加 `{YYYY}/{MM}/` 分层。

## 关联文件

| 类型 | 路径 | 状态 |
|---|---|---|
| Tabbit 原始导出 | `docs/tabbit/inbox/2026/05/TB-20260518-190000-tabbit-task-id-upgrade.tabbit.raw.md` | 待回溯绑定 |
| 正式归档 | `docs/wiki/tasks/2026/05/TB-20260518-190000-tabbit-task-id-upgrade.archive.md` | 本文件 |

## 任务理解

本任务是 ADR-0009 的 v2 升级实施。v1 的核心缺陷是"保留原始文件名"，导致无法建立跨文件的 task 追踪链路。v2 通过统一的 Task ID 将所有相关资产编织在一起。

## 执行过程

### 步骤一：工程闭环扫描

10 项检查结果：3 项需补齐（ADR 更新、目录登记、LLM Wiki），其余不适用或已具备。无阻断性问题。

### 步骤二：更新核心文件

1. **ADR-0009** → v2 重写：新增 Task ID 规范、文件后缀规范、manifest 机制、蒸馏 Skill 边界、Tabbit Task ID 块规范。
2. **`.trae/commands/tabbit-task.md`** → v2 重写：从"接收描述→归档"改为"接收 Task ID→检索→重命名→归档→manifest"。
3. **新建 `.trae/skills/tabbit-task-distillation/SKILL.md`**：独立蒸馏 Skill，负责扫描 pending manifest → 聚合材料 → 过滤噪声 → 生成 distilled → 更新索引和状态。
4. **两个 README** → 反映 Task ID 重命名规则和 `{YYYY}/{MM}` 分层结构。

### 步骤三：回溯重命名现有归档

将已有的两个归档文件从 v1 命名格式迁移到 v2 Task ID 格式：

```
2026-05-18_181000_tabbit-task-archive-flow-design.md
  → TB-20260518-181000-tabbit-task-archive-flow-design.archive.md

2026-05-18_182000_trae-tabbit-archive-review.md
  → TB-20260518-182000-trae-tabbage-archive-review.archive.md
```

并为两者生成了对应的 `.manifest.md`。

### 步骤四：生成本次任务的第三个归档和 manifest

本文件即为第三个正式归档，Task ID 为 `TB-20260518-190000-tabbit-task-id-upgrade`。

## 最终产出

### 新增/修改文件

| 类型 | 路径 | 说明 |
|---|---|---|
| 更新 | `docs/adr/ADR-0009-tabbit-task-archive-flow.md` | v2：Task ID 驱动版 |
| 重写 | `.trae/commands/tabbit-task.md` | v2：Task ID 驱动版 |
| 新建 | `.trae/skills/tabbit-task-distillation/SKILL.md` | 蒸馏 Skill |
| 更新 | `docs/tabbit/README.md` | Task ID 规则 + {YYYY}/{MM} 分层 |
| 更新 | `docs/trae-export/README.md` | Task ID 规则 + {YYYY}/{MM} 分层 |
| 重命名 | `docs/wiki/tasks/...archive.md × 2` | Task ID 格式 |
| 新建 | `docs/wiki/tasks/...manifest.md × 2` | 已有任务的关联清单 |
| 新建 | `docs/wiki/tasks/...archive.md` | 本文件（第 3 个归档） |
| 新建 | `docs/wiki/tasks/...manifest.md` | 本任务的关联清单 |

### 最终目录结构

```text
docs/
├── tabbit/
│   ├── README.md
│   └── inbox/
│       ├── .gitkeep
│       └── 2026/05/
├── trae-export/
│   ├── README.md
│   └── inbox/
│       ├── .gitkeep
│       └── 2026/05/
└── wiki/
    └── tasks/
        └── 2026/05/
            ├── TB-20260518-181000-*.archive.md    ← 归档 #1（已重命名）
            ├── TB-20260518-181000-*.manifest.md   ← 清单 #1（新建）
            ├── TB-20260518-182000-*.archive.md    ← 归档 #2（已重命名）
            ├── TB-20260518-182000-*.manifest.md   ← 清单 #2（新建）
            ├── TB-20260518-190000-*.archive.md    ← 归档 #3（本文件）
            └── TB-20260518-190000-*.manifest.md   ← 清单 #3（待生成）
```

## 后续处理

1. 对 `docs/tabbit/inbox/` 下 6 个历史文件执行 `/tabbit-task` 回溯绑定。
2. 对 `docs/trae-export/inbox/` 下 1 个历史文件执行 `/tabbit-task` 回溯绑定。
3. 本任务生成 manifest 后，可触发 `tabbit-task-distillation` Skill 验证完整链路。
