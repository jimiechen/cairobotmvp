# TRAE 任务执行结果导出收件箱

## 目录定位

本目录是 **TRAE 任务执行完成后导出的会话/评审文件的收件箱（inbox）**。

## 核心规则：所有文件必须被自动重命名

**不要保留原始文件名。** 所有放入本目录的文件都必须通过 `/tabbit-task` 自动理解内容后，按 Canonical Task ID 重命名。

**不需要上游提供 Task ID。** `/tabbit-task` 会根据文档内容自动生成。

## 目录结构

```text
docs/trae-export/
├── README.md                    # 本说明文件（入库）
└── inbox/                       # 原始导出文件（不入库）
    └── {YYYY}/{MM}/             # 按年月组织
        └── TB-{YYYYMMDD}-{HHMMSS}-{slug}.trae.exec.md
```

## `/tabbit-task` 输入模式

| 模式 | 用法 | 说明 |
|---|---|---|
| 显式 ID | `Task ID: TB-xxx` + 文件路径 | 沿用已有 ID |
| 自动 ID | 只提供文件路径 | TRAE 读内容自动生成 ID（最常用） |
| 批量迁移 | 提供目录路径 | 扫描所有未重命名文件逐个处理 |

## 文件处理后缀规范

| 来源 | 后缀 | 示例 |
|---|---|---|
| TRAE 执行导出 | `.trae.exec.md` | `TB-20260518-193000-tabbage-review.trae.exec.md` |

## 处理流程

```text
TRAE 执行完任务并导出 → 放入 docs/trae-export/inbox/
    ↓
触发 /tabbit-task（只需提供文件路径，无需 Task ID）
    ↓
TRAE 读取内容 → 自动生成 Canonical Task ID
    ↓
重命名为 TB-{ID}.trae.exec.md → 移入 inbox/{YYYY}/{MM}/
    ↓
生成 archive.md + manifest.md (distillation_status: pending)
    ↓
夜间 tabbit-task-distillation Skill 蒸馏
```

## 当前 inbox 文件（待批量迁移）

| 文件名 | 说明 | 已重命名 |
|---|---|---|
| `评审TabAI会话导出方案.md` | TRAE 评审输出 | ❌ 待回溯绑定 |

## 相关文档

- 决策依据：[ADR-0009 v3](../adr/ADR-0009-tabbit-task-archive-flow.md)
- 归档命令：[`.trae/commands/tabbit-task.md`](../../.trae/commands/tabbit-task.md)
- 蒸馏 Skill：[`.trae/skills/tabbit-task-distillation/SKILL.md`](../../.trae/skills/tabbit-task-distillation/SKILL.md)
- 归档目标：[docs/wiki/tasks/](../wiki/tasks/)
