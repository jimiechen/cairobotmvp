# TRAE 评审 TabAI 会话导出方案

## 元信息

- 来源：TRAE 导出
- 原始文件：`docs/trae-export/inbox/评审TabAI会话导出方案.md`
- 创建时间：2026-05-18 18:20:00
- 归档路径：`docs/wiki/tasks/2026/05/2026-05-18_182000_trae-tabbit-archive-review.md`
- 任务状态：已完成
- 标签：#trae #wiki-task #review #tabbit

## 原始任务

用户要求对 `docs/tabbit/TabAI会话_1779097841153.md` 中的"Tabbit 任务到 TRAE Markdown 归档流程设计"方案进行工程规范评审，输出结构化评审意见。

## 任务理解

本任务是对原始归档方案（提议新建 `docs/llm-wiki/` 完整体系）进行工程合规性评审。

核心目标：
1. 判断方案是否与项目现有工程规范冲突。
2. 识别必须修改项和建议修改项。
3. 给出可执行的修订建议。

关键约束：
- 必须对照 AGENTS.md、`.trae/rules/docs.md`、`cairobot-doc-placement` Skill 等现有规则。
- 评审结论需包含"符合/必须修改/建议修改/测试缺口/文档缺口/风险提示"完整维度。

## 执行过程

### 步骤一：读取原始方案

读取 `docs/tabbit/TabAI会话_1779097841153.md`，理解原始方案全貌：
- 提议新建 `docs/llm-wiki/` 目录体系（inbox/tasks/distilled/decisions/prompts/assets/index.md）
- 创建 `/tabbit-task` 和 `/distill-wiki` 两个 Slash Command
- 设计完整的 Markdown 模板和文件命名规范

### 步骤二：项目现状扫描

检查项目实际结构，发现：
- 已有 [docs/wiki/LLM-WIKI.md](../../wiki/LLM-WIKI.md)（243 行）
- 已有 `docs/reports/daily/`、`docs/reports/distilled/` 报告机制
- 已有 `cairobot-html-distillation` Skill 负责蒸馏
- `.trae/commands/` 目录尚不存在（无任何 Command 文件）
- `docs/tabbit/` 下存在 4 个不可读文件名

### 步骤三：逐维评审

按 review.md 规范输出七维评审：

**一、结论**：建议修改

**二、符合要求部分**（6 项）：痛点识别准确、Slash Command 先行策略合理、命名规范合理、模板结构完整、强制约束必要、职责分离正确。

**三、必须修改项**（4 项）：
1. ❌ `docs/llm-wiki/` 与 `docs/wiki/` 冲突（R0）
2. ❌ 与现有日报/蒸馏/Skill 机制完全未对齐（R0）
3. ❌ 缺少 ADR 支撑（R1）
4. ❌ 未考虑 CI 集成（R2）

**四、建议修改项**（4 项）：目录定位不明确、Command 过长、与 Skill 关系未说明、缺少 .gitignore

**五、测试缺口**：非代码变更可不写测试，但建议手动端到端验证

**六、文档缺口**（3 项）：ADR、docs.md 更新、LLM-WIKI.md 更新

**七、风险提示**（3 项）：维护混乱 R1、过度工程化 R2、inbox 垃圾场 R2

### 步骤四：输出修订建议

给出最小可行路径（8 步），核心为"复用 docs/wiki/ 不另起炉灶"。

## 最终产出

### 评审结论

| 维度 | 结论 |
|---|---|
| 总体 | 建议修改 |
| 方向 | 正确，需对齐现有规范 |
| 最大风险 | 新建 docs/llm-wiki/ 导致两套 Wiki 并存 |

### 核心修订建议（8 条）

```
1. 不新建 docs/llm-wiki/，改为复用 docs/wiki/ 扩展
2. 将 docs/tabbit/ 定位为 inbox，在 docs.md 中补登记
3. /tabbit-task Command 产出的任务文档 → docs/wiki/tasks/{YYYY}/{MM}/
4. /distill-wiki 合并到现有 cairobot-html-distillation Skill 流程中
5. 补充 ADR-0009 记录此决策
6. 先用 1.md 这类真实文件做一次端到端验证
7. 明确 docs/trae-export/ 也纳入归档范围
8. 后续如需检查 wiki/tasks/，单独设计 wiki-task-check
```

## 文件变更建议

| 类型 | 路径 | 说明 |
|---|---|---|
| 新建 | `docs/adr/ADR-0009-tabbit-task-archive-flow.md` | 决策记录 |
| 新建 | `.trae/commands/tabbit-task.md` | 归档命令（瘦身版） |
| 新建 | `docs/tabbit/README.md` | 收件箱说明 |
| 新建 | `docs/trae-export/README.md` | 收件箱说明 |
| 新建 | `docs/wiki/tasks/2026/05/*.md` | 归档目录 |
| 更新 | `docs/wiki/LLM-WIKI.md` | 追加 tasks 入口段 |
| 更新 | `.trae/rules/docs.md` | 补充目录登记 |
| 新建 | `.gitignore` | inbox 忽略规则 |

## Wiki 索引候选

- [TRAE 评审 TabAI 会话导出方案](./tasks/2026/05/2026-05-18_182000_trae-tabbit-archive-review.md)
  - 时间：2026-05-18 18:20:00
  - 来源：TRAE
  - 标签：#trae #wiki-task #review #tabbit
  - 摘要：对原始 Tabbit 归档方案进行七维工程评审，发现 4 个必须修改项和 4 个建议修改项，核心结论是不应新建 docs/llm-wiki/

## 后续处理

1. ✅ 本评审意见已被架构师采纳（见 `TabAI会话_1779098249005.md`）
2. ✅ 基于评审意见的修订方案已落地实施（见首个归档文件）
3. `docs/tabbit/inbox/` 下仍有 3 个待归档文件（`1.md`、`TabAI会话_1779079297836.md`、`TabAI会话_1779093720973.md`）
