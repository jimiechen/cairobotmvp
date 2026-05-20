# LLM Wiki 三层架构决策

## 元信息

- 决策类型：Wiki 结构架构决策
- 决策日期：2026-05-19
- 状态：**已接受**
- 决策者：Tabbit 主控 + Trae IDE 执行
- 相关方案：`docs/tabbit/inbox/2026/05/TabAI会话_1779205708319.md`（任务指令）
- 原始 Tabbit 导出：`docs/tabbit/inbox/2026/05/TabAI会话_1779206504027.md`（含 Trae 任务 Raw 归档补充需求）

## 1. 背景

### 1.1 问题

LLM-WIKI.md 从项目初期开始作为"单文档累积型知识库"，随着项目推进持续膨胀：
- 原始 308 行，包含系统组成、TDD 规则、协议编号、已完成事项清单、Tabbit 任务索引等混合内容
- 新增知识只能追加到文件末尾，无法分类管理
- AI 助手阅读成本高，关键信息被淹没在大量细节中
- 与 CODE-WIKI.md 存在内容重复（如 Protobuf 协议规则）

### 1.2 为什么不能继续无限膨胀

1. **可读性下降**：308 行且持续增长，AI 助手上下文窗口有限
2. **职责混乱**：同一文件承担导航、正文、索引、原始记录四种角色
3. **更新风险**：追加式更新容易引入不一致
4. **自动化障碍**：SOLO Web 等自动化工具难以精准定位写入位置

## 2. 决策

采用 **Raw → Distillation → Index** 三层知识库结构重构 LLM Wiki。

## 3. 三层结构定义

### 3.1 Raw 原始层

**职责**：
- 保存原始事实、导出文档、命令输出、git 状态、测试结果、审计报告、日报原始材料
- 不做过度总结
- 不删除、不覆盖原始材料
- 作为后续所有蒸馏内容的事实依据

**目录**：
- `docs/tabbit/inbox/` — Tabbit / TabAI 原始导出
- `docs/trae-export/inbox/` — TRAE 执行导出（含手动导出文件和结构化任务 Raw 归档）
  - `docs/trae-export/inbox/tasks/YYYY/MM/` — Trae IDE 通过 `/task-raw-archive` 生成的结构化任务记录（使用 Source Record ID 作为主键，Task ID 可选）
- `docs/reports/daily/` — 每日原始日报
- `docs/reports/testing/` — 测试报告
- `docs/reports/audit/` — 审计报告
- `docs/reports/coverage/` — 覆盖率报告

> **关键约束**：Raw 层目录不得放在 `docs/wiki/` 下（wiki/ 只承载 Distillation 层和 Index 层）。Trae 每次执行任务的原始记录统一归入 `docs/trae-export/inbox/tasks/`。
>
> **身份模型**：Raw 层使用 **Source Record ID (SRC-{SOURCE}-{timestamp}-{hash8})** 作为主键，不强制使用 Canonical Task ID。详见 `cairobot-task-raw-archive` Skill §3.2 和 `docs/wiki/source-map.md`。

### 3.2 Distillation 蒸馏层

**职责**：
- 从 Raw 层材料中提取可复用的结构化知识
- 区分事实、判断、风险、规则、后续行动
- 不把计划写成完成
- 不把设计写成实现
- 不把 mock、TODO、空实现写成主链路完成

**目录**：
- `docs/wiki/daily/` — 每日蒸馏的稳定知识
- `docs/wiki/tasks/` — 归档任务（ADR-009 体系）
- `docs/wiki/decisions/` — 经确认的技术决策
- `docs/wiki/modules/` — 各模块稳定知识页

### 3.3 Index 目录层

**职责**：
- 只做导航、摘要、索引、状态入口
- 不承载完整日报、完整审计、完整对话、完整蒸馏正文
- LLM-WIKI.md 应成为项目级入口索引，而不是巨型知识正文文件

**文件**：
- `docs/wiki/LLM-WIKI.md` — 主入口索引（本文件的上层）
- `docs/wiki/CODE-WIKI.md` — 代码架构知识
- `docs/wiki/ADR索引.md` — ADR 索引
- `docs/wiki/每日蒸馏索引.md` — 每日蒸馏产物索引
- `docs/wiki/任务索引.md` — 任务级索引

## 4. 各方职责

### 4.1 SOLO Web

| 可做 | 不可做 |
|---|---|
| 定时触发 | 业务代码修复 |
| 调用大模型 | 架构性重构 |
| 生成 Markdown 文件 | 自动提交 main |
| 生成 Raw / 日报 / 蒸馏 / 主控汇报 | 未经确认直接改 LLM-WIKI.md |
| 遵守不 commit / 不 push 限制 | Skill / Command 制作 |

### 4.2 Trae IDE（本地）

| 可做 | 不可做 |
|---|---|
| 结构性重构 | 未经主控确认 git push |
| Skill / Command 制作 | 业务代码修改（本次范围外） |
| docs/wiki/ 结构调整 | 删除已有原始材料 |
| LLM-WIKI.md 重构 | 破坏 ADR-009 归档体系 |
| **每次任务归档到 Raw 层** | **跳过 Raw 归档直接宣称完成** |

**硬规则：Trae 每次执行任务，必须先在 Raw 层留痕。** 必须记录任务提示词、执行结果、产物清单。详见 `cairobot-task-raw-archive` Skill 和 `/task-raw-archive` Command。

### 4.3 Tabbit 主控

- 确认三层架构决策
- 审核 SOLO Web 产物质量
- 审批 Trae IDE 的结构性变更
- 决定是否 git push

## 5. 禁止事项

1. 不允许删除已有日报、蒸馏、审计、任务归档原始材料
2. 不允许把原始材料直接塞进 LLM-WIKI.md
3. 不允许把未确认内容写成长期确定规则
4. 不允许把 SOLO Web 试运行产物写成"已提交"或"最终完成"
5. 不允许修改业务代码（Wiki 重构任务范围内）
6. 不允许修复 Gateway/Tars 代码（Wiki 重构任务范围内）
7. 不允许重构 go/python/typescript/proto 业务目录（Wiki 重构任务范围内）
8. 不允许执行 destructive 命令
9. 未经主控确认，不允许 git push
10. 如果需要 commit，必须先输出提交范围和 diff 摘要，等待主控确认

## 6. 后续迁移策略

### 6.1 已完成

- [x] LLM-WIKI.md 重构为 Index 层入口
- [x] 建立 decisions/、modules/ 子目录
- [x] 建立每日蒸馏索引和任务索引
- [x] 新增 cairobot-llm-wiki-distillation Skill
- [x] 新增 llm-wiki-distill Command
- [x] 新增 cairobot-task-raw-archive Skill（Trae 任务 Raw 归档强制规则）
- [x] 新增 task-raw-archive Command（Trae 任务结构化归档入口）
- [x] Raw 层目录定义补充：`docs/trae-export/inbox/tasks/YYYY/MM/`

### 6.2 待后续执行

- [ ] 将 LLM-WIKI.md 中已压缩的"已完成事项 30+ 条"迁移到独立文件或保留为统计摘要
- [ ] 将 §15 Tabbit 任务索引表完全移交至 任务索引.md
- [ ] SOLO Web 正式任务 prompt 按此结构同步更新
- [ ] CI report-check 脚本适配新目录结构
- [ ] AGENTS.md §19 Skill 索引表同步新增条目
