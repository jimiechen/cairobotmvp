---
name: 定时知识蒸馏
slug: scheduled-knowledge-distillation
summary: 定义定时知识蒸馏的完整规则体系，SOLO Web 定时任务和 Trae IDE 手动补跑的统一规范。含 SRC-ID/SOURCE-MAP 模型和候选态机制。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - scheduled-distillation
  - solo-web
  - src-id
  - source-map
trigger:
  - "定时蒸馏"
  - "scheduled distillation"
  - "每日知识蒸馏"
  - "daily distill"
  - "scheduled knowledge distillation"
priority: high
blocking: true
---

# CaiRobot MVP 定时任务知识蒸馏 Skill

## 1. Skill 职责

本 Skill 定义**定时知识蒸馏**的完整规则体系，是 SOLO Web 每日自动化任务和 Trae IDE 手动补跑的统一规范。

**负责**：
- 定义 Raw → Distillation → Index 三层蒸馏流程规则
- 规定 Source Record ID 和 Source Map 在蒸馏中的使用方式
- 约束 SOLO Web 自动化任务的边界（禁止 auto commit/push、禁止直接覆盖 Index）
- 区分"候选"与"已确认"的知识状态

**不负责**：
- 单次 Trae 任务的 Raw 归档（由 `cairobot-task-raw-archive` 负责）
- Tabbit 单任务蒸馏（由 `tabbit-task-distillation` 负责）
- 业务代码修改
- 日报提交（由 `cairobot-daily-report` 负责）

详细规则参见：
- [LLM Wiki 三层架构决策](../../docs/wiki/decisions/llm-wiki-three-layer-architecture.md)
- [LLM Wiki 架构模块](../../docs/wiki/modules/llm-wiki-architecture.md)
- [Source Map](../../docs/wiki/source-map.md)
- [SOLO Web 自动化边界](../../docs/wiki/modules/solo-web-automation.md)
- [.trae/rules/docs.md](../../.trae/rules/docs.md)
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md)

## 2. 与其他 Skill 的关系

| Skill | 关系 | 说明 |
|---|---|---|
| `cairobot-task-raw-archive` | **上游依赖** | 蒸馏前优先读取 Raw Source Record，本 Skill 不生成 Raw |
| `cairobot-llm-wiki-distillation` | **包含/扩展** | 本 Skill 是其"定时/自动化"子集，增加 SOLO Web 约束和 SRC-ID/SOURCE-MAP 规则 |
| `tabbit-task-distillation` | **并行互补** | 该 Skill 处理 Tabbit 单任务蒸馏，本 Skill 处理全局定时蒸馏 |
| `cairobot-daily-report` | **联动** | 蒸馏完成后可能触发日报更新 |
| `cairobot-doc-placement` | **依赖** | 新建文件时需遵守目录规范 |

## 3. 三层蒸馏核心规则

### 3.1 Raw 层：只读事实来源

**蒸馏必须从 Raw 层读取材料，不得跳过 Raw 直接编造知识。**

Raw 层来源包括：

| 来源 | 路径 | 身份标识 |
|---|---|---|
| Tabbit 导出 | `docs/tabbit/inbox/` | 原始文件名（保留原名） |
| TRAE 执行记录 | `docs/trae-export/inbox/` | 原始文件名或 SRC-ID |
| TRAE 结构化归档 | `docs/trae-export/inbox/tasks/YYYY/MM/` | **Source Record ID (SRC-ID)** |
| 每日日报 | `docs/reports/daily/` | 日期 + 标题 |
| 测试报告 | `docs/reports/testing/` | 报告名称 |
| 审计报告 | `docs/reports/audit/` | 审计名称 |

**Raw 层硬约束**：
1. ✅ Raw 文件**只读不改**
2. ✅ 蒸馏必须引用 Raw 文件的 **Source Record ID 或原始路径**作为来源追溯
3. ✅ 如 Raw 记录有 Relation Group (RG)，蒸馏产物应继承同一 RG
4. ❌ 不得仅凭 Task ID 判断跨来源关联
5. ❌ 不得为了对齐 Task ID 改写 Raw 文件名

### 3.2 Distillation 层：知识加工（候选态）

蒸馏产出的知识处于**候选态**，不是最终确定的知识。

**Distillation 输出位置**：

| 类型 | 目录 | 命名 |
|---|---|---|
| 每日蒸馏 | `docs/wiki/daily/{YYYY-MM-DD}-{slug}.md` | 日期+标题 |
| 决策沉淀 | `docs/wiki/decisions/{title}.md` | 英文标题 |
| 模块知识页 | `docs/wiki/modules/{module}.md` | 模块名 |
| 任务归档 | `docs/wiki/tasks/{TASK_ID}.md` | Task ID（可选） |

**每篇蒸馏文档必须区分**：

| 标记 | 含义 | 写入规则 |
|---|---|---|
| ✅ **事实** | 已验证的客观信息 | 必须有 Raw 来源引用 |
| ⚠️ **判断** | 基于事实的主观分析 | 必须标注置信度 |
| 🔴 **风险** | 潜在问题或不确定事项 | 必须说明影响范围 |
| 📋 **规则** | 经确认应遵守的约束 | 必须说明适用场景 |
| ➡️ **后续行动** | 待办事项 | **不能写成已完成** |

**禁止行为**：
- ❌ 不把计划写成完成
- ❌ 不把设计写成实现
- ❌ 不把 mock、TODO、空实现写成主链路完成
- ❌ 不把未确认内容写成长期确定规则
- ❌ 不把 candidate 关联写成 confirmed
- ❌ 不把模型推测写成事实

### 3.3 Index 层：导航服务（确认态）

Index 层是经过主控确认后的稳定导航入口。

**Index 文件列表**：

| 文件 | 职责 | 更新方式 |
|---|---|---|
| `LLM-WIKI.md` | 主入口索引 | **候选追加，不直接覆盖** |
| `每日蒸馏索引.md` | 每日蒸馏产物索引 | **候选追加** |
| `任务索引.md` | 任务级索引 | **候选追加**（使用 SRC-ID / RG） |
| `source-map.md` | 跨来源关联索引 | **候选追加** |

**Index 更新硬规则**：

1. ✅ 蒸馏任务**只生成 Index 更新候选**，不直接写入正式 Index
2. ✅ 候选格式明确标记为"待主控确认"
3. ✅ 正式 Index 更新由主控或本地 Trae IDE 确认后执行
4. ❌ **SOLO Web 不得直接修改 LLM-WIKI.md**
5. ❌ **SOLO Web 不得自动 commit / push 任何文件**
6. ❌ **不得把 candidate 关联写入 source-map.md 的 confirmed 状态**

## 4. Source Record 与 Source Map 在蒸馏中的使用

### 4.1 蒸馏输入：按 Source Record 扫描

定时蒸馏启动时，扫描 Raw 层当日新增的所有 Source Record：

```
Step 1: 扫描 Raw 层
  ├── docs/tabbit/inbox/{date}/*.md          → 记录原始路径作为来源
  ├── docs/trae-export/inbox/{date}/*.md      → 记录原始路径或 SRC-ID
  ├── docs/trae-export/inbox/tasks/{date}/*   → 读取 SRC-ID（主键）
  └── docs/reports/daily/{date}*.md           → 记录日报路径

Step 2: 为每条 Raw 记录分配/复用 SRC-ID
  ├── 已有 SRC-ID → 直接复用
  └── 无 SRC-ID → 生成 SRC-{SOURCE}-{timestamp}-{hash8}

Step 3: 查询 source-map.md 中是否有已有 RG
  ├── 有匹配 RG → 继承该 RG
  └── 无匹配 RG → 标记为 none（可选创建新 RG 候选）
```

### 4.2 蒸馏输出：携带来源元数据

每条蒸馏产物必须携带：

```markdown
## 来源元数据

- **Source Records**:
  - SRC-TABBIT-20260520-114812-a13f9c02 (`path/to/raw`)
  - SRC-TRAE-20260520-115001-b82ad119 (`path/to/raw`)
- **Relation Group**: RG-20260520-001（如有关联）
- **蒸馏日期**: YYYY-MM-DD
- **蒸馏执行者**: SOLO Web / Trae IDE / 手动
- **置信度**: high / medium / low
- **状态**: candidate（默认）/ confirmed（仅主控可设）
```

### 4.3 Index 候选输出

蒸馏完成后，生成以下 Index 更新候选文件：

```markdown
# Index 更新候选 — {YYYY-MM-DD}

## 每日蒸馏索引候选

| 日期 | Raw 来源 | 蒸馏产物 | 状态 |
|---|---|---|---|
| {date} | {SRC-ID or path} | {distilled path} | candidate |

## Source Map 候选

| RG-ID | SRC-ID | 关联方式 | 置信度 | 状态 |
|---|---|---|---|---|
| {RG} | {SRC-ID} | semantic/candidate/none | med/low | candidate |

## 任务索引候选

| 记录ID | 任务名称 | Raw 来源 | 蒸馏文件 | 状态 |
|---|---|---|---|---|
| {SRC-ID or TB-ID} | {title} | {path} | candidate |
```

> **关键**：以上全部为"候选"（candidate），需经主控确认后才可合并到正式 Index。

## 5. SOLO Web 定时任务约束

### 5.1 SOLO Web Prompt 必须包含的内容

SOLO Web 的定时任务 Prompt **必须引用本 Skill 的以下规则**：

1. **三层结构**：Raw（只读）→ Distillation（候选）→ Index（确认态）
2. **Source Record ID**：Raw 层主键，不强制 Task ID
3. **Source Map**：跨来源关联通过 RG 维护，禁止假关联
4. **候选 vs 确认**：蒸馏产出候选态知识，Index 更新为候选追加
5. **禁止 auto commit/push**：所有文件只写不提交
6. **禁止直接覆盖 LLM-WIKI.md**：只能生成候选追加内容
7. **禁止 candidate→confirmed**：AI 不得自行提升关联置信度

### 5.2 SOLO Web 可做与不可做

| 可做 | 不可做 |
|---|---|
| ✅ 扫描 Raw 层当日新增 | ❌ 修改或删除任何 Raw 文件 |
| ✅ 生成 Distillation 层文档 | ❌ 直接修改 LLM-WIKI.md |
| ✅ 生成 Index 更新候选文件 | ❌ git add / commit / push |
| ✅ 引用 Source Record ID 作为来源 | ❌ 仅凭 Task ID 强制对齐跨来源文件 |
| ✅ 标记 candidate 关联到 source-map 候选 | ❌ 把 candidate 写成 confirmed |
| ✅ 输出蒸馏报告和产物清单 | ❌ 自行决定架构性变更 |

### 5.3 Trae IDE 手动补跑

`/daily-distill` Command 支持手动补跑任意日期的蒸馏：

- 支持指定日期（默认当天）
- 支持指定 Raw 来源范围（默认全量）
- 输出与 SOLO Web 相同格式的候选产物
- 手动模式下可由操作者决定是否将候选提升为 confirmed

## 6. 每日定时蒸馏标准流程

```
┌─────────────────────────────────────────┐
│  Phase 1: Raw 层扫描                    │
│  ├─ 扫描 docs/tabbit/inbox/{date}/      │
│  ├─ 扫描 docs/trae-export/inbox/{date}/ │
│  ├─ 扫描 docs/trae-export/inbox/tasks/ │
│  ├─ 扫描 docs/reports/daily/{date}      │
│  └─ 为每条记录分配/复用 SRC-ID          │
└──────────────┬──────────────────────────┘
               ▼
┌─────────────────────────────────────────┐
│  Phase 2: 去噪与过滤                    │
│  ├─ 删除工具调用流水                      │
│  ├─ 删除重复片段                        │
│  ├─ 删除 AI 免责声明                     │
│  └─ 保留核心结论和决策点                  │
└──────────────┬──────────────────────────┘
               ▼
┌─────────────────────────────────────────┐
│  Phase 3: 知识提取（Distillation 层）    │
│  ├─ 稳定决策 → wiki/decisions/         │
│  ├─ 可复用流程 → wiki/modules/         │
│  ├─ 日知识摘要 → wiki/daily/            │
│  └─ 每条产物携带 SRC-ID + RG 元数据     │
└──────────────┬──────────────────────────┘
               ▼
┌─────────────────────────────────────────┐
│  Phase 4: Index 候选生成                │
│  ├─ 每日蒸馏索引候选                    │
│  ├─ Source Map 候选（新增关联）          │
│  ├─ 任务索引候选                        │
│  └─ 全部标记为 candidate 态             │
└──────────────┬──────────────────────────┘
               ▼
┌─────────────────────────────────────────┐
│  Phase 5: 输出报告                      │
│  ├─ 蒸馏产物清单（含路径和 SRC-ID）       │
│  ├─ Index 候选清单                      │
│  ├─ 待确认项                            │
│  └─ 是否允许进入正式 Index               │
└─────────────────────────────────────────┘
```

## 7. 完成前硬校验清单

每次蒸馏任务完成后确认：

### 7.1 Raw 层校验

- [ ] 未修改、未删除、未覆盖任何 Raw 文件
- [ ] 所有蒸馏产物均引用了 Raw Source Record ID 或原始路径
- [ ] 未制造假关联（无仅凭 Task ID 对齐的跨来源声明）

### 7.2 Distillation 层校验

- [ ] 蒸馏文档写入了正确目录（`docs/wiki/` 下对应子目录）
- [ ] 蒸馏内容不含工具调用噪声
- [ ] 蒸馏内容区分了事实/判断/风险/规则/行动
- [ ] 计划/设计/mock/TODO 未写成已完成
- [ ] 每条蒸馏产物携带来源元数据（SRC-ID / RG / 日期 / 执行者）

### 7.3 Index 层校验

- [ ] **未直接修改 LLM-WIKI.md**
- [ ] **未直接修改正式 Index 文件**
- [ ] Index 更新以"候选"形式输出到单独文件
- [ ] candidate 关联未被写成 confirmed
- [ ] Source Map 候选中无 confirmed 状态的新增关联

### 7.4 Git 校验

- [ ] **未执行 git add**
- [ ] **未执行 git commit**
- [ ] **未执行 git push**

## 8. 违规阻断

以下行为视为违规，必须立即停止并上报：

- 修改或删除 Raw 层文件
- 跳过 Raw 层直接编造知识
- 仅凭 Task ID 强制对齐不同来源的文件
- 直接覆盖 LLM-WIKI.md 或其他正式 Index 文件
- 将 candidate 关联提升为 confirmed
- 执行 git add / commit / push
- 把模型推测的跨来源关联写入长期规则
- 在 SOLO Web Prompt 中省略本 Skill 的核心规则
