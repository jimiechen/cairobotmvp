# SOLO Web 自动化能力边界

## 1. 定位

本文记录 SOLO Web 在 CaiRobot MVP 项目中的已验证能力、未确认能力和推荐职责边界。

## 2. SOLO Web 定时任务应遵守的 Skill

SOLO Web 的每日定时任务 Prompt **必须引用**以下 Skill：

| 优先级 | Skill | 触发条件 | 核心规则 |
|---|---|---|---|
| **P0** | `cairobot-scheduled-knowledge-distillation` | 每次定时蒸馏任务 | 三层结构、SRC-ID/SOURCE-MAP、候选态、禁止 auto commit/push |
| P1 | `cairobot-llm-wiki-distillation` | 全局 Wiki 知识整理 | 蒸馏规则、去噪、分类提取、Index 更新 |
| P1 | `cairobot-task-raw-archive` | 如涉及 Trae 任务 Raw 归档 | Raw 归档强制规则（如 SOLO 需生成 Raw 记录） |
| P2 | `tabbit-task-distillation` | 如处理 Tabbit 导出文件 | Tabbit 原始归档蒸馏 |

> **关键**：`cairobot-scheduled-knowledge-distillation` 是 SOLO Web 定时任务的**主控 Skill**。该 Skill 定义了：
> - Raw → Distillation → Index 三层结构
> - Source Record ID (SRC-ID) 作为 Raw 层主键，不强制 Task ID
> - Source Map / Relation Group (RG) 跨来源关联机制
> - 蒸馏产物为 candidate 态，不直接覆盖正式 Index
> - 禁止 auto commit/push、禁止直接修改 LLM-WIKI.md
>
> 详见 [.trae/skills/cairobot-scheduled-knowledge-distillation/SKILL.md](../../.trae/skills/cairobot-scheduled-knowledge-distillation/SKILL.md)

## 3. 已验证能力

以下能力已在试运行中验证通过：

| 能力 | 说明 | 验证状态 |
|---|---|---|
| 定时触发 | 可按定时任务自动触发 | ✅ 已验证 |
| 调用大模型 | 可调用 LLM 生成内容 | ✅ 已验证 |
| 生成 Markdown 文件 | 可生成符合格式的 .md 文件 | ✅ 已验证 |
| 遵守禁止 commit/push | 可拦截 git 操作，不自动提交 | ✅ 已验证 |
| 输出多类型产物 | 可同时生成 Raw、日报、蒸馏、主控汇报 | ✅ 已验证 |

## 3. 未确认能力

以下能力尚未经过充分验证：

| 能力 | 风险 |
|---|---|
| 自动提交到 feature 分支 | 可能产生不符合规范的 commit |
| 自动创建 PR | PR 描述可能不完整 |
| 跨分支一致性 | 多分支场景下的行为不确定 |
| 长期稳定性 | 连续运行数天后的可靠性未知 |
| 大规模文件操作 | 批量修改时的安全性未知 |

## 4. 推荐职责

基于已验证能力，SOLO Web 推荐承担以下职责：

| 职责 | 频率 | 输出位置 |
|---|---|---|
| 每日 Raw 采集 | 每天 | `docs/reports/daily/` |
| 日报生成 | 每天 | `docs/reports/daily/` |
| 蒸馏生成 | 每天 | `docs/reports/distilled/` |
| 主控汇报生成 | 每天 | `docs/reports/daily/` |
| Index 更新候选生成 | 每天 | 候选文件，待主控确认后入库 |

## 5. 不推荐职责

以下职责不建议由 SOLO Web 承担：

| 职责 | 原因 | 推荐执行者 |
|---|---|---|
| 业务代码修复 | 缺乏项目上下文和测试能力 | 本地 Trae IDE |
| 架构性重构 | 需要全局理解仓库结构 | 本地 Trae IDE |
| 自动提交 main 分支 | 高风险，不可逆 | 主控人工确认 |
| 直接修改 LLM-WIKI.md | 需要理解三层结构规范 | 本地 Trae IDE |
| Skill / Command 制作 | 需要对齐现有格式体系 | 本地 Trae IDE |
| 长期知识库结构调整 | 需要主控决策 | 主控 + Trae IDE |

## 6. 与其他执行者的协作关系

```
SOLO Web（每日自动化）
  ↓ 产出 Raw / 日报 / 蒸馏 / 汇报候选
Trae IDE（本地结构性工作）
  ↓ 审核 + 确认 + 入库 + 重构
Tabbit 主控（最终决策）
  ↓ 确认 + git push 批准
```
