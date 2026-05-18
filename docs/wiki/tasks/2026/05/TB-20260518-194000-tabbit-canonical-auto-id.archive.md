# Tabbit 归档方案 v3：Canonical Task ID 自动生成

## 元信息

- Canonical Task ID: TB-20260518-194000-tabbit-canonical-auto-id
- External Task ID: 无
- 原始文件名: `docs/tabbit/inbox/TabAI会话_1779102107054.md`
- 重命名后路径: `docs/tabbit/inbox/2026/05/TB-20260518-194000-tabbit-canonical-auto-id.tabbit.raw.md`
- 创建时间: 2026-05-18 19:40:00
- 主题 slug: tabbit-canonical-auto-id
- 任务状态: archived
- 蒸馏状态: pending
- 是否需要人工确认: false
- 输入模式: 自动ID

## 原始任务

用户要求对 `docs/tabbit/inbox/TabAI会话_1779102107054.md` 中的架构师修订意见实施。核心变更：

v2 要求 Tabbit 必须生成 Task ID，但实际流程中无法保证。v3 修正为：**Canonical Task ID 由 `/tabbit-task` 兜底自动生成**，不强依赖任何上游工具。

具体变更：
1. 新增 External Task ID vs Canonical Task ID 双层概念
2. `/tabbit-task` 支持四种输入模式（显式 ID / 自动 ID / 批量迁移 / 纯文本）
3. topic-slug 从文档内容自动提取（标题→需求首句→文件名→关键词→兜底）
4. 兜底命名 `untitled-tabbit-note` + `needs_human_review: true`
5. manifest 新增 `External Task ID`、`Input Mode`、`Needs Human Review` 字段

## 自动理解结果

- **任务目标**: 将归档入口从"被动接收 ID"升级为"主动理解内容并生成 ID"
- **关键上下文**: v2 的致命缺陷是强依赖 Tabbit 输出规范，历史文件和手动文件无法纳入体系
- **产出类型**: 工程流程升级（ADR、Command、Skill、README 全面更新）
- **主题置信度**: high（文档有明确标题和多轮迭代结构）

## 执行过程

### 步骤一：工程闭环扫描

10 项检查：3 项需补齐（ADR 更新、目录登记、LLM Wiki），无阻断性问题。

### 步骤二：更新核心文件（6 个）

1. **ADR-0009** → v3 重写：新增版本演进表、ID 生成优先级、四种输入模式、topic-slug 自动提取规则、External vs Canonical 双层 ID。
2. **`.trae/commands/tabit-task.md`** → v3 重写：从"接收 ID"改为"自动建档"，新增输入模式识别、ID 生成三步法（检查→生成→去重）、自动理解结果章节、needs_human_review 标记。
3. **`.trae/skills/tabbit-task-distillation/SKILL.md`** → v3 更新：适配 Canonical Task ID 主键、External Task ID 可选字段、needs_human_review 谨慎蒸馏策略、免责声明过滤。
4. **两个 README** → 反映自动建档 + 四种模式 + "无需上游提供 ID"。

### 步骤三：生成本次任务的第四个归档和 manifest

本文件即为第四个正式归档，Canonical Task ID 为 `TB-20260518-194000-tabbit-canonical-auto-id`。

## 最终产出

### 完整链路（v3 最终版）

```text
用户把文件放入 inbox/
    ↓
触发 /tabbit-task（只需文件路径，无需 Task ID）
    ↓
TRAE 读取内容 → 检查是否有已有 ID
    ↓
有 → 沿用；无 → 自动生成 Canonical Task ID
    ↓
识别文件类型 → 重命名 → 移入 inbox/{YYYY}/{MM}/
    ↓
生成 archive.md + manifest.md (pending)
    ↓
夜间 tabbit-task-distillation 扫描 pending manifest
    ↓
聚合材料 → 过滤噪声 → distilled.md
    ↓
更新 LLM-WIKI.md + manifest 状态改为 distilled
```

### 新增/修改文件

| 类型 | 路径 | 说明 |
|---|---|---|
| 更新 | [ADR-0009](docs/adr/ADR-0009-tabbit-task-archive-flow.md) | v3：Canonical Task ID 自动生成版 |
| 重写 | [.trae/commands/tabbit-task.md](.trae/commands/tabbit-task.md) | v3：四种输入模式 + 自动 ID |
| 更新 | [.trae/skills/tabbit-task-distillation/SKILL.md](.trae/skills/tabbit-task-distillation/SKILL.md) | v3：适配 Canonical Task ID |
| 更新 | [docs/tabbit/README.md](docs/tabbit/README.md) | 反映自动建档 + 四种模式 |
| 更新 | [docs/trae-export/README.md](docs/trae-export/README.md) | 同上 |
| 新建 | [本文件](docs/wiki/tasks/2026/05/TB-20260518-194000-tabbit-canonical-auto-id.archive.md) | 归档 #4 |
| 新建 | [对应 manifest](docs/wiki/tasks/2026/05/TB-20260518-194000-tabbit-canonical-auto-id.manifest.md) | 清单 #4 |

## 关联文件

| 类型 | 原始路径 | 重命名后路径 | 状态 |
|---|---|---|---|
| tabbit.raw | `docs/tabbit/inbox/TabAI会话_1779102107054.md` | `docs/tabbit/inbox/2026/05/TB-20260518-194000-tabbit-canonical-auto-id.tabbit.raw.md` | 待回溯绑定 |
| archive | - | `docs/wiki/tasks/2026/05/TB-20260518-194000-tabbage-canonical-auto-id.archive.md` | generated |

## Wiki 索引候选

- [Tabbit 归档方案 v3：Canonical Task ID 自动生成](./tasks/2026/05/TB-20260518-194000-tabbit-canonical-auto-id.archive.md)
  - Canonical Task ID: TB-20260518-194000-tabbit-canonical-auto-id
  - 来源：Tabbit + 架构师 + TRAE
  - 状态：已归档，待蒸馏
  - 摘要：将归档入口从"被动接收 ID"升级为"主动理解内容并生成 ID"，支持四种输入模式和自动建档

## 夜间蒸馏输入

本任务已生成 manifest，等待夜间 `tabbit-task-distillation` Skill 按 Canonical Task ID 聚合材料进行知识蒸馏。
