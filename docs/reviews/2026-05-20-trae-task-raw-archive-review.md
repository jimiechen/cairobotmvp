# 评审意见：Trae 任务 Raw 归档规范设计

## 一、结论

**建议修改** — 方案方向正确、问题定义准确，但存在 **3 个必须修改项** 和 **4 个建议修改项**，修改后可进入实施。

---

## 二、评审对象

| 项目 | 内容 |
|---|---|
| 方案来源 | `docs/tabbit/inbox/2026/05/TabAI会话_1779206504027.md` |
| 方案名称 | Trae 任务 Raw 归档规范设计 |
| 评审日期 | 2026-05-20 |
| 评审范围 | Raw 层扩展方案、Skill 设计、Command 设计、与现有三层架构的兼容性 |

---

## 三、符合要求的部分

### 3.1 问题定义准确

方案核心论点——**"没有 Raw 层任务上下文，Distillation 层知识会失真"**——是成立的。现有三层架构中，Raw 层缺少 Trae 每次执行任务的原始记录（提示词、执行过程、产物清单），导致后续蒸馏只能看到最终日报/产物，看不到任务是如何被要求的。这是一个真实的工程缺口。

### 3.2 Raw 记录模板设计合理

方案提出的 7 段式模板（基本信息 → 提示词 → 输入材料 → 执行结果 → 产物清单 → 待确认项 → 蒸馏建议）覆盖了完整的任务生命周期，结构清晰。

### 3.3 Skill + Command 分离的设计模式正确

Skill 定义规则 + Command 执行动作的分离方式，与本项目现有的 Skill 体系（如 `cairobot-llm-wiki-distillation` + `llm-wiki-distill`）保持一致。

### 3.4 与 ADR-0009 Task ID 体系对齐的意识正确

方案提到使用 Canonical Task ID 命名文件，与现有 ADR-0009 体系一致。

### 3.5 "Raw 保真、Distillation 压缩、Index 导航" 三原则重申到位

方案明确强调 Raw 层不可省略核心信息，这与 [llm-wiki-three-layer-architecture.md](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md) 的核心约束一致。

---

## 四、必须修改项

### 4.1 🔴 Raw 目录放置位置违反现有三层架构约定

**问题描述**：

方案建议将 Trae 任务 Raw 记录放在：
```
docs/wiki/raw/tasks/YYYY/MM/
```

但根据现有 [llm-wiki-three-layer-architecture.md](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md) §3.1 的定义，**`docs/wiki/` 目录只承载 Distillation 层和 Index 层**，不包含 Raw 层。当前 Raw 层的实际目录为：

| Raw 内容 | 现有路径 |
|---|---|
| Tabbit 导出 | `docs/tabbit/inbox/` |
| TRAE 执行导出 | `docs/trae-export/inbox/` |
| 每日日报 | `docs/reports/daily/` |
| 测试报告 | `docs/reports/testing/` |

将 `raw/tasks/` 放在 `docs/wiki/` 下会破坏这一约定，导致 Raw 层和 Distillation 层边界模糊。

**建议修改为以下两种方案之一**：

**方案 A（推荐）**：复用已有的 `docs/trae-export/inbox/` 作为 Trae 任务 Raw 入口
```
docs/trae-export/inbox/tasks/YYYY/MM/TB-{timestamp}-{slug}.md
```
理由：`docs/trae-export/inbox/` 已在三层架构决策中被定义为"TRAE 执行记录"的 Raw 层位置，无需新增目录。

**方案 B**：在 `docs/reports/` 下新增子目录
```
docs/reports/raw-tasks/YYYY/MM/TB-{timestamp}-{slug}.md
```
理由：与其他 reports 子目录（daily、testing、audit、coverage）保持平级。

**无论选择哪种方案，都不应放在 `docs/wiki/` 下。**

### 4.2 🔴 与 `docs/trae-export/inbox/` 的关系未澄清

**问题描述**：

现有三层架构已经定义了 `docs/trae-export/inbox/` 作为"TRAE 执行记录"的 Raw 层入口（见 [llm-wiki-three-layer-architecture.md](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md) §3.1 和 [llm-wiki-architecture.md](../../docs/wiki/modules/llm-wiki-architecture.md) §4 的命名规范 `*.trae.exec.md`）。新方案未说明与该已有目录的关系。

**可能的关系**：
- **替代关系**：新的 task-raw-archive 完全取代旧的 trae-export/inbox 用途
- **互补关系**：trae-export/inbox 保留给 SOLO Web 自动导出，task-raw-archive 用于 Trae IDE 手动归档
- **合并关系**：统一为一个入口

方案必须明确这一点，否则会出现两套 Raw 记录体系的混乱。

### 4.3 🔴 与现有 reporting.md 规则的重叠未处理

**问题描述**：

[reporting.md](../../.trae/rules/reporting.md) §10 "任务完成汇报模板" 已经要求输出：
- 完成内容、修改文件、测试命令、测试结果、Bug、事故、文档同步、风险遗留

[project.md](../../.trae/rules/project.md) §5 也要求输出：
- 本次完成了什么、修改了哪些文件、新增了哪些测试、运行了哪些命令

新方案的 Raw 记录模板 §4"执行结果"和 §5"产物清单"与上述内容高度重叠。如果每次任务同时执行"任务完成汇报"和"Raw 归档"，会产生大量重复工作。

**建议**：
- 明确 Raw 归档是任务完成汇报的**结构化持久化版本**，不是额外负担
- 或者将 `/task-raw-archive` Command 设计为从任务完成汇报自动生成 Raw 文件
- 在 Skill 中说明与 `cairobot-daily-report` Skill 的联动关系

---

## 五、建议修改项

### 5.1 🟡 Command 拆分策略在 MVP 阶段偏重

**问题描述**：

方案提出"一个 Skill + 两个 Command"（task-start + task-close），或最小版本合并为一个 Command。

在 S0 阶段，建议先实现**单一合并 Command**（`task-raw-archive.md`），原因：
- 当前任务量不大，拆分 start/close 的收益不明显
- 两阶段 Command 需要在两次调用间维护状态（Task ID、开始时间等），增加复杂度
- 等 S1/S2 阶段任务频率上升后再考虑拆分

**建议**：MVP 版本先做单 Command，在 Skill 中标注"未来可拆分为 task-start / task-close"。

### 5.2 🟡 任务索引表列设计需与现有格式兼容

**问题描述**：

方案建议在 `任务索引.md` 中增加一列变为：
```text
| Task ID | 任务名称 | Raw 记录 | Distillation | 状态 | 主控结论 |
```

现有 `任务索引.md` 格式为 7 列：
```text
| Task ID | 任务名称 | 执行者 | Raw 来源 | 蒸馏文件 | 当前状态 | 主控结论 |
```

新方案的建议列数不同且列名不完全一致（"Raw 记录" vs "Raw 来源"，"Distillation" vs "蒸馏文件"，少了"执行者"列）。

**建议**：
- 直接复用现有 7 列格式
- 新增的 Trae 任务 Raw 记录路径填入"Raw 来源"列即可
- 无需改动表格结构

### 5.3 🟡 文件命名应严格遵循 Canonical Task ID 规范

**问题描述**：

方案给出的示例文件名：
```
TB-20260519-235953-llm-wiki-raw-task-archive.md
```

以及时间戳备选：
```
2026-05-19-235953-llm-wiki-raw-task-archive.md
```

根据 [llm-wiki-architecture.md](../../docs/wiki/modules/llm-wiki-architecture.md) §4 的命名规范，TRAE 导出文件的格式应为 `{TASK_ID}.trae.exec.md`。新方案应遵循此规范，而非引入新的命名模式。

**建议**：文件名格式统一为：
```
TB-{YYYYMMDD}-{HHMMSS}-{slug}.trae.raw.md
```

后缀 `.trae.raw.md` 区别于 SOLO Web 自动生成的 `.trae.exec.md`。

### 5.4 🟡 缺少自动化触发机制设计

**问题描述**：

方案要求"每个 Trae 任务都必须先建 Raw 记录再开始执行"，但未说明如何确保这一点被强制执行。目前依赖 AI 助手自觉遵守，缺乏硬性保障。

**建议**：
- 在 `cairobot-active-gap-filling` Skill 中增加检查项："本次任务是否已建立 Raw 记录"
- 或者在 `cairobot-engineering-workflow` 启动时强制触发 Raw 建档步骤
- CI report-check 可增加对当日 Raw 记录完整性的抽查

---

## 六、测试缺口

本方案属于工程规范/流程类变更，不涉及代码实现，因此不需要单元/集成测试。但建议补充以下验证：

| 验证项 | 方法 | 通过标准 |
|---|---|---|
| Raw 目录路径是否符合三层架构约定 | 对照 `llm-wiki-three-layer-architecture.md` §3.1 | 不在 `docs/wiki/` 下 |
| 文件命名是否遵循 Canonical Task ID 规范 | 对照 `llm-wiki-architecture.md` §4 | 符合 `{TASK_ID}.*.md` 格式 |
| Skill 是否与现有 12 个 Skill 无冲突 | 逐个检查职责重叠 | 有明确的职责边界声明 |
| Command 输入是否可从现有汇报模板推导 | 对照 `reporting.md` §10 | 无矛盾、可联动 |
| 任务索引更新是否兼容现有格式 | 对照 `任务索引.md` 当前表头 | 不需要改表结构 |

---

## 七、文档缺口

| 缺失文档 | 说明 | 建议 |
|---|---|---|
| 更新后的 `llm-wiki-three-layer-architecture.md` | 方案要求补充 Trae 任务归档条目到 Raw 层定义 | 必须同步更新 §3.1 Raw 目录列表 |
| 更新后的 `llm-wiki-architecture.md` | 方案要求补充数据流向图 | 必须在 §3 数据流图中增加 "Trae 任务执行 → Raw" 箭头 |
| 更新后的 `任务索引.md` | 方案要求增加 Raw 记录索引能力 | 建议在维护规则中补充 Trae 任务的录入方式 |
| Skill 文件 `cairobot-task-raw-archive/SKILL.md` | 方案设计了 Skill 但未产出实际文件 | 实施时创建 |
| Command 文件 `task-raw-archive.md` | 方案设计了 Command 但未产出实际文件 | 实施时创建 |

---

## 八、风险提示

| 风险ID | 等级 | 描述 | 影响 | 缓解措施 |
|---|---|---|---|---|
| R-RAW-001 | R1 | Raw 记录变成形式主义，AI 助手复制粘贴填充模板 | Raw 层失去保真价值 | Skill 中明确禁止无实质内容的归档；CI 抽查内容质量 |
| R-RAW-002 | R2 | 每次任务都归档导致 Raw 层膨胀过快 | 存储成本上升、检索效率下降 | 设置保留策略（如按月压缩归档）；Slug 命名要语义化便于检索 |
| R-RAW-003 | R2 | 两套 Raw 体系（trae-export/inbox vs 新 tasks/）并行导致混乱 | 后续蒸馏不知道读哪个来源 | 必须先解决 §4.2 的关系定义问题 |
| R-RAW-004 | R3 | 小任务（如只改一个 typo）也走完整归档流程显得过重 | 开发体验下降 | Skill 中设置"轻量任务简化模板"，允许 5 行以内的精简记录 |

---

## 九、总结评分

| 维度 | 评分 | 说明 |
|---|---|---|
| 问题准确性 | ★★★★★ | 击中了真实痛点——知识失真 |
| 方案完整性 | ★★★★☆ | 覆盖面广，但缺少数落地的细节 |
| 与现有体系兼容性 | ★★★☆☆ | 目录放置、命名规范、索引格式存在冲突需修正 |
| 可操作性 | ★★★★☆ | Skill+Command 模式成熟，但 MVP 阶段应先简后繁 |
| 风险控制 | ★★★☆☆ | 未充分考虑膨胀风险和与现有体系的关系 |

**综合评定：方向正确，修正 3 个必须修改项后可进入实施。**
