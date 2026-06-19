# NovelForge 方案评审 — 多AI模型多分支小说自动化编辑器

**基于 OpenCode 二次开发，借鉴 Claude Code 全量架构，面向全球多语言市场**

| 项目 | 内容 |
|---|---|
| 文档类型 | 方案评审 |
| 创建日期 | 2026-06-12 |
| 评估对象 | OpenCode v1.17.3 + Claude Code v2.1.88 |

---

## 执行摘要

Claude Code v2.1.88 泄露的 512K 行源码中，我们系统梳理出 **15 项核心架构特性**。前期可行性评估仅覆盖了其中 7 项，本次评审补充了剩余的 8 项关键机制，并重新评估全部 15 项的 NovelForge 映射价值。

**核心结论**：在 15 项 Claude Code 核心机制中，**11 项可在 OpenCode 基础上以中低难度实现**，4 项需要较高投入。全球小说编辑器市场存在显著痛点：中文工具功能陈旧、英文工具学习成本高、小语种市场几乎空白。NovelForge 通过"多AI模型路由 + 多分支故事探索 + Skills/Commands 指令系统 + 全球多语言原生支持"四大差异化支柱，具备成为**全球首个AI原生多语言小说自动化编辑器**的战略窗口期。

---

## 第一章：Claude Code 全量架构解析 — 15 项机制

### 前期已覆盖的 7 项机制

| 机制 | 核心能力 | NovelForge 映射 | 难度 | 优先级 |
|---|---|---|---|---|
| Hook 系统 | PreToolUse / PostToolUse / PermissionDenied | AI 续写前敏感词拦截 | 低 | P0 |
| Named Subagents | @mention 调用 + Git Worktree 隔离 | 角色审校 Agent | 低 | P0 |
| KAIROS 持久化 | Daily-Log + Cron 调度 | 角色状态追踪 | 低 | P0 |
| StreamingToolExecutor | 边收边跑的流式执行 | AI 续写边生成边显示 | 中 | P0 |
| Git Worktree 隔离 | 物理文件系统隔离 | 多章节并行 AI 审校 | 中 | P1 |
| Ultraplan 远程卸载 | 重计算任务远程执行 | 全书大纲生成远程执行 | 高 | P2 |
| Agent Teams 对等协作 | 多进程共享 Task List | 多角色并行审校 | 高 | P2 |

### 新增 8 项关键机制

#### 机制八：Skills 渐进式披露系统

Claude Code 的 Skills 系统允许用户通过自然语言描述创建可复用的 AI 能力模块。每个 Skill 包含描述元数据、知识库、工具集，并根据当前上下文动态选择相关 Skill 加载。

**NovelForge 映射**：定义"古风小说 Skill""科幻设定 Skill""悬疑推理 Skill"等垂直领域模块。当用户打开古风小说项目时，相关 Skill 自动激活，AI 的续写和审校建议都会遵循古风规范。**这是传统编辑器完全无法实现的动态领域适配能力。**

#### 机制九：Commands 三层指令体系

支持 local（本地 shell）、local-jsx（JSX 组件渲染）、prompt（结构化 AI 交互）三种命令类型。Commands 与 Skills 联动 — Skill 激活时自动注册相关 Commands。

**NovelForge 映射**：构建小说创作专用 Commands 体系：`/outline generate`、`/character create`、`/branch split`、`/review consistency`。古风 Skill 激活时自动注册 `/poetry insert`、`/idiom suggest` 等专属命令。

#### 机制十：QueryEngine 七阶段循环

七阶段循环：Context Projection → Auto-Compaction Check → API Streaming → Error Recovery → Tool Execution → Attachment Processing → Continuation Decision。

**NovelForge 映射**：Context Projection 阶段针对小说场景定制 — 解析"让主角在这里黑化"时，投影到角色档案、剧情进度、世界观设定等多个维度。**这是从"通用 AI 助手"升级为"领域专家 Agent"的关键架构。**

#### 机制十一：五级上下文压缩管道

| 级别 | 名称 | 触发条件 | 压缩策略 |
|---|---|---|---|
| L1 | Tool Result Budget | 单个工具结果超预算 | 截断或摘要工具输出 |
| L2 | History Snip | 历史消息累计超出窗口 | 移除最早非关键消息 |
| L3 | Microcompact | 上下文密度过高 | 合并相邻同类消息 |
| L4 | Context Collapse | 接近窗口上限 | 将历史对话压缩为摘要 |
| L5 | Autocompact | 窗口已满 | 全量重构上下文，保留关键状态 |

**NovelForge 映射**：L1 压缩章节摘要、L2 移除已完结支线、L3 合并润色操作、L4 卷级压缩、L5 全量重构保留角色状态和世界设定索引。

#### 机制十二：九种 Continue 语义

| 语义 | 行为 | 小说场景 |
|---|---|---|
| continue | 继续正常循环 | 常规续写 |
| yield | 暂停等待用户输入 | 关键剧情节点 |
| ask_followup | 主动追问澄清 | 指令模糊时 |
| switch_mode | 切换 Agent 模式 | 创作/审校切换 |
| delegate | 委托给子 Agent | 自动委托 Character Agent |
| abort | 终止当前任务 | 严重一致性错误 |
| retry | 重试当前步骤 | 质量不达标 |
| fallback | 降级到备用模型 | 主模型不可用 |
| compact | 触发上下文压缩 | token 超 80% 预算 |

#### 机制十三：四级错误恢复机制

1. **Retry**：同一模型重试，调整参数
2. **Fallback**：切换到备用模型
3. **Degrade**：降级功能
4. **Abort**：终止任务，报告错误

**NovelForge 映射**：多模型路由场景下，DeepSeek 不稳定时 fallback 到 Qwen，Claude 拒绝生成敏感内容时 degrade 为提示用户手动处理。

#### 机制十四：Buddy 终端伙伴

以"终端电子宠物"形式显示 Agent 状态、反馈任务进度、支持用户自定义外观和性格。

**NovelForge 映射**：演变为"写作伙伴" — 常驻编辑器角落，卡文时提供灵感，达成日更目标时给予鼓励，长期项目中成为"最懂这个故事"的虚拟伙伴。**在情感层面建立竞品无法复制的用户粘性。**

#### 机制十五：多模型智能路由

根据任务复杂度自动选择 Haiku/Sonnet/Opus，基于任务类型、历史成功率、当前负载、成本预算决策。平均成本降低 60%。

**NovelForge 映射**：大纲生成用 DeepSeek，情感描写用 Claude，创意发散用 GPT，小语种用 Qwen，一致性检查用本地模型。**动态路由 + 用户反馈优化，形成成本与质量的帕累托最优。**

### 全量 15 项机制评估矩阵

| # | 机制 | NovelForge 价值 | 难度 | 优先级 | 预计工期 |
|---|---|---|---|---|---|
| 1 | Hook 增强 | 质量关卡拦截 | 低 | P0 | 3 天 |
| 2 | Named Subagents | 专业化 Agent 分工 | 低 | P0 | 5 天 |
| 3 | Daily-Log 记忆 | 跨会话状态持久化 | 低 | P0 | 3 天 |
| 4 | Skills 渐进披露 | 动态领域适配 | 低 | P0 | 7 天 |
| 5 | Commands 三层指令 | 小说专用指令集 | 低 | P0 | 7 天 |
| 6 | StreamingToolExecutor | 边生成边显示 | 中 | P0 | 10 天 |
| 7 | 九种 Continue 语义 | 精细会话控制 | 中 | P1 | 10 天 |
| 8 | 四级错误恢复 | 多模型稳定性 | 中 | P1 | 7 天 |
| 9 | Git Worktree 隔离 | 多分支并行审校 | 中 | P1 | 10 天 |
| 10 | 五级上下文压缩 | 长篇小说上下文管理 | 中 | P1 | 14 天 |
| 11 | QueryEngine 七阶段 | 领域专家推理管道 | 中 | P1 | 14 天 |
| 12 | 多模型智能路由 | 成本优化 + 质量保障 | 中 | P1 | 10 天 |
| 13 | Buddy 写作伙伴 | 情感化用户粘性 | 中 | P2 | 14 天 |
| 14 | Ultraplan 远程卸载 | 重计算任务远程执行 | 高 | P2 | 21 天 |
| 15 | Agent Teams 对等协作 | 多 Agent 并行审校 | 高 | P2 | 21 天 |

---

## 第二章：全球主流编辑器深度评估与痛点分析

### 中文市场

| 产品 | 定位 | AI 能力 | 核心痛点 | 价格 |
|---|---|---|---|---|
| 墨者写作 | 专业网文写作 | 基础续写、敏感词检测 | 功能陈旧，AI 能力停留在 2023 年 | 免费/￥15/月 |
| 橙瓜码字 | 网文创作 + 投稿 | 简单续写、拼字 | 无角色管理，无世界观追踪 | 免费/￥18/月 |
| 壹写作 | 长篇创作 | 大纲辅助 | 移动端体验差 | ￥30/月 |
| 作家助手（番茄） | 番茄小说官方 | AI 续写、润色 | 平台锁定，无法导出 | 免费 |
| 笔灵 AI | AI 写作助手 | 续写、润色、扩写 | 非专业编辑器，无项目管理 | ￥29/月 |

**痛点总结**：所有中文工具的共同痛点是 AI 能力停留在"续写/润色"层面，缺乏角色一致性管理、世界观追踪、多分支探索、跨会话记忆等深度功能。

### 英文市场

| 产品 | 定位 | AI 能力 | 核心痛点 | 价格 |
|---|---|---|---|---|
| Scrivener | 专业长篇写作 | 无原生 AI | 学习曲线极陡，无 AI 集成 | $59 一次性 |
| LivingWriter | 云原生小说写作 | AI 辅助（有限） | 订阅昂贵，AI 功能弱 | $96/年 |
| Atticus | 写作 + 排版出版 | 无原生 AI | 排版强但写作弱，价格高 | $147-249 |
| SudoWrite | AI 小说写作 | 续写、描述、头脑风暴 | **仅支持英文，价格高昂** | $19-59/月 |
| NovelAI | AI 故事生成 | 故事生成、图像生成 | 质量不稳定，无专业编辑功能 | $10-25/月 |
| Jasper | 通用 AI 写作 | 多场景写作 | 非小说专用，模板化严重 | $49-125/月 |

**痛点总结**：传统工具 AI 能力缺失，AI 工具仅支持英文且价格高昂，缺乏专业编辑功能。

### 日文/韩文市场

| 产品 | 定位 | AI 能力 | 核心痛点 |
|---|---|---|---|
| AI のべりすと | AI 小说续写 | 续写、风格模仿 | 功能单一，模型锁定 |
| ノベルアップ+ | 小说投稿平台 | AI 辅助（有限） | 平台锁定，编辑功能弱 |
| 문피아 | 网文平台 | 无原生 AI | 纯平台，无创作工具 |
| 조아라 | 网文平台 | 无原生 AI | 纯平台，无创作工具 |

### 小语种市场 — 蓝海中的蓝海

**几乎完全空白**。当地作者使用 Google Docs、Microsoft Word、Wattpad（纯平台）等通用工具写作，没有任何专门的 AI 小说编辑工具。

| 市场 | 人口 | 网文读者 | 年增长率 | AI 工具供给 | 机会评级 |
|---|---|---|---|---|---|
| 印尼 | 2.78 亿 | 3500 万+ | 28% | 零 | ★★★★★ |
| 巴西 | 2.15 亿 | 2800 万+ | 22% | 零 | ★★★★★ |
| 泰国 | 7000 万 | 1200 万+ | 25% | 零 | ★★★★☆ |
| 土耳其 | 8500 万 | 900 万+ | 20% | 零 | ★★★★☆ |
| 沙特/UAE | 4200 万 | 600 万+ | 30% | 零 | ★★★★☆ |
| 越南 | 1 亿 | 1500 万+ | 18% | 零 | ★★★☆☆ |
| 墨西哥 | 1.28 亿 | 1100 万+ | 15% | 零 | ★★★☆☆ |
| 俄罗斯 | 1.46 亿 | 2000 万+ | 10% | 零 | ★★★☆☆ |

---

## 第三章：多语言出海市场战略分析

### 全球网文市场规模

- **2026 年市场规模**：$28B
- **2030 年预测规模**：$42B
- **年复合增长率**：12.3%
- **全球读者数**：4.2 亿
- **非英语市场占比**：73%

### 多语言 LLM 能力评估

| 模型 | 语言数 | 小语种质量 | 文学性 | 成本 | 推荐场景 |
|---|---|---|---|---|---|
| Qwen 3.5 | 201 | 优秀 | 中 | 低 | 小语种创作首选 |
| Claude 4.5 | 30+ | 优秀 | 极高 | 高 | 情感描写、文学润色 |
| GPT-5 | 50+ | 良好 | 高 | 中高 | 创意发散 |
| Gemini 3.1 | 40+ | 良好 | 中 | 中 | 多语言混合场景 |
| DeepSeek V4 | 20+ | 中等 | 中 | 极低 | 大纲、逻辑推理 |

**Qwen 3.5 支持 201 种语言，是小语种创作的首选模型。**

### 出海定价策略

| 市场层级 | 地区 | 建议定价 | 策略 |
|---|---|---|---|
| Tier 1 | 北美、西欧、日本、韩国 | $15-25/月 | 与 SudoWrite 竞争 |
| Tier 2 | 中国、东南亚（新/马） | ￥39-69/月 | 与墨者/橙瓜竞争 |
| Tier 3 | 印尼、泰国、越南、巴西、土耳其 | $3-8/月 | 低价渗透，抢占空白 |
| Tier 4 | 印度、非洲、中东（非海湾） | 免费 + 高级付费 | Freemium，培养习惯 |

---

## 第四章：NovelForge 差异化架构设计

### 四大差异化支柱

1. **多AI模型智能路由** — 75+ 模型自由选择，成本与质量的最优平衡
2. **多分支故事探索** — 基于 Git Worktree 的物理隔离，真正的平行世界管理
3. **Skills/Commands 指令系统** — 动态领域适配 + 小说专用指令集
4. **全球多语言原生出海** — UI 25+ 语言，AI 创作 200+ 语言

### 多分支故事引擎（核心差异化）

基于 Git Worktree 的物理隔离，实现真正的"平行世界"管理：

- **What-if 分支**："如果主角选择 A 而不是 B"
- **风格分支**：同一章节用古风/现代/科幻三种风格撰写
- **角色分支**：从不同角色视角重写同一段剧情
- **长度分支**：精简版/标准版/扩展版
- **AU 分支**：Alternative Universe

```
/branch create <name> --from <chapter> --type what-if
/branch list
/branch switch <name>
/branch merge <name> --into main
/branch diff <branch-a> <branch-b>
/branch compare <branch-a> <branch-b> --ai-review
```

**没有任何现有编辑器支持"多分支故事探索"。** Scrivener 的"快照"只能保存单一版本，无法并行演进。

### 角色与世界观管理系统

```yaml
# .novelforge/characters/protagonist.yaml
name: 林墨
aliases: [墨哥, 林公子]
age: 24
appearance:
  height: 180cm
  features: [剑眉, 薄唇, 左手腕有烫伤疤痕]
personality:
  mbti: INTJ
  traits: [冷静, 腹黑, 护短]
speech_pattern: 简洁，少用语气词，紧张时会摸左手腕
relationships:
  - { target: 苏婉, type: 恋人, status: 隐瞒 }
arc:
  start: 冷漠孤傲的世家弃子
  end: 放下执念的江湖侠客
```

### 伏笔与剧情追踪系统

```yaml
# .novelforge/plot/foreshadowing.yaml
foreshadowings:
  - id: fs-001
    description: 林墨左手腕的烫伤疤痕
    planted_in: 第 3 章
    resolved_in: 第 47 章
    status: resolved
    importance: major

  - id: fs-002
    description: 苏婉提到的"那个约定"
    planted_in: 第 12 章
    resolved_in: null
    status: pending
    deadline_chapter: 第 60 章
```

---

## 第五章：Skills / Commands 系统规格

### Skills 渐进式披露

```markdown
---
name: 古风武侠
match:
  - glob: "**/*.{古风,武侠}*"
  - regex: "(?i)(江湖|门派|武功|内力)"
description: 古风武侠小说的写作规范与常用套路
---

# 古风武侠 Skill

## 写作规范
- 对话使用半角标点，叙述使用全角标点
- 武功招式名称用【】标注

## 常用套路
- 开局：主角身世之谜 / 灭门惨案 / 奇遇得宝
- 中期：门派斗争 / 江湖恩怨 / 修炼突破
- 结局：大仇得报 / 归隐江湖 / 开创新派

## 禁忌清单
- 现代词汇（手机、电脑、网络等）
- 西方文化元素（圣诞节、咖啡等）
```

### Commands 三层指令（v1.0 完整列表）

| 命令 | 类型 | 功能 |
|---|---|---|
| `/outline generate` | prompt | 生成全书大纲 |
| `/outline expand <chapter>` | prompt | 扩展章节细纲 |
| `/character create` | prompt | 创建角色档案 |
| `/character check` | prompt | 检查角色一致性 |
| `/world build` | prompt | 构建世界观设定 |
| `/continue` | prompt | AI 续写 |
| `/polish` | prompt | 润色当前段落 |
| `/branch create` | local | 创建故事分支 |
| `/branch list` | local | 列出所有分支 |
| `/branch switch` | local | 切换分支 |
| `/branch compare` | prompt | AI 对比分支差异 |
| `/review consistency` | prompt | 一致性检查 |
| `/review plot` | prompt | 剧情逻辑检查 |
| `/review style` | prompt | 风格评估 |
| `/foreshadow list` | local | 列出所有伏笔 |
| `/foreshadow check` | prompt | 检查伏笔回收状态 |
| `/export epub` | local | 导出 EPUB |
| `/export pdf` | local | 导出 PDF |
| `/memory search <query>` | local | 搜索历史记忆 |
| `/model switch <model>` | local | 切换 AI 模型 |

---

## 第六章：多AI模型路由与多分支引擎

### 路由决策矩阵

| 任务类型 | 首选模型 | 备选模型 | 选择理由 | 预估成本/千字 |
|---|---|---|---|---|
| 大纲生成 | DeepSeek V4 | Qwen 3.5 | 逻辑强、成本低 | ￥0.03 |
| 章节续写 | Claude 4.5 | GPT-5 | 文学性最佳 | ￥0.15 |
| 情感描写 | Claude 4.5 | GPT-5 | 细腻、有温度 | ￥0.15 |
| 对话生成 | GPT-5 | Claude 4.5 | 自然、符合人物性格 | ￥0.12 |
| 创意发散 | GPT-5 | Claude 4.5 | 脑洞大、联想丰富 | ￥0.12 |
| 一致性检查 | 本地模型 | DeepSeek V4 | 隐私、快、便宜 | ￥0.005 |
| 小语种创作 | Qwen 3.5 | Gemini 3.1 | 语言覆盖最广 | ￥0.08 |
| 润色优化 | Claude 4.5 | Qwen 3.5 | 文学质量高 | ￥0.10 |
| 摘要生成 | DeepSeek V4 | 本地模型 | 快、便宜 | ￥0.01 |

### 五级上下文压缩（小说场景定制）

| 级别 | 名称 | 小说场景触发 | 压缩策略 |
|---|---|---|---|
| L1 | 章节摘要 | 单章内容超过 8K tokens | 将章节压缩为 500 字摘要 |
| L2 | 支线归档 | 已完结支线剧情占用上下文 | 将完结支线压缩为梗概 |
| L3 | 操作合并 | 连续的润色/修改操作 | 合并为最终版本 |
| L4 | 卷级压缩 | 已完成卷的内容 | 压缩为卷摘要 + 关键事件列表 |
| L5 | 全量重构 | 上下文完全满载 | 重构为：角色状态 + 世界观快照 + 当前卷摘要 + 最近 3 章全文 |

---

## 第七章：完整方案评审矩阵与实施路线图

### 全量方案评审矩阵

| 方案 | 来源 | 核心能力 | 难度 | 用户价值 | 竞争壁垒 | 优先级 |
|---|---|---|---|---|---|---|
| Hook 增强系统 | Claude Code | AI 输出质量关卡 | 低 | 高 | 中 | P0 |
| Named Subagents | Claude Code | 专业化 Agent 分工 | 低 | 高 | 中 | P0 |
| Daily-Log 记忆 | Claude Code | 跨会话状态持久化 | 低 | 高 | 中 | P0 |
| Skills 渐进披露 | Claude Code | 动态领域适配 | 低 | 极高 | 高 | P0 |
| Commands 三层指令 | Claude Code | 小说专用指令集 | 低 | 极高 | 高 | P0 |
| StreamingToolExecutor | Claude Code | 边生成边显示 | 中 | 高 | 中 | P0 |
| 多分支故事引擎 | NovelForge 原创 | 平行世界管理 | 中 | 极高 | 极高 | P0 |
| 角色/世界观管理 | NovelForge 原创 | 一致性保障 | 中 | 极高 | 高 | P0 |
| 多模型智能路由 | Claude Code + 原创 | 成本优化 + 质量保障 | 中 | 高 | 高 | P1 |
| 九种 Continue 语义 | Claude Code | 精细会话控制 | 中 | 中 | 中 | P1 |
| 四级错误恢复 | Claude Code | 多模型稳定性 | 中 | 中 | 中 | P1 |
| 五级上下文压缩 | Claude Code | 长篇小说上下文管理 | 中 | 高 | 高 | P1 |
| QueryEngine 七阶段 | Claude Code | 领域专家推理管道 | 中 | 中 | 高 | P1 |
| Git Worktree 隔离 | Claude Code | 并行审校 | 中 | 中 | 中 | P1 |
| Buddy 写作伙伴 | Claude Code + 原创 | 情感化用户粘性 | 中 | 中 | 高 | P2 |
| 伏笔追踪系统 | NovelForge 原创 | 质量保障 | 中 | 高 | 高 | P1 |
| Ultraplan 远程卸载 | Claude Code | 重计算任务远程执行 | 高 | 中 | 低 | P2 |
| Agent Teams 对等协作 | Claude Code | 多 Agent 并行审校 | 高 | 中 | 中 | P2 |
| 多语言 UI (25+) | OpenCode 基础 | 全球本地化 | 低 | 高 | 中 | P0 |
| 小语种 AI 创作 (200+) | Qwen 3.5 等 | 小语种市场覆盖 | 低 | 极高 | 极高 | P0 |

### 实施路线图（26 周）

**Phase 1：MVP 核心（Week 1-4）**
- Hook 增强系统 + Named Subagents + Daily-Log 记忆
- Skills 渐进披露（3 个基础 Skill）+ Commands 三层指令（20+ 命令）
- 多语言 UI（中/英/日/韩）

**Phase 2：体验升级（Week 5-8）**
- StreamingToolExecutor + 多分支故事引擎
- 角色/世界观管理系统 + 多模型路由（3 模型）
- 小语种 AI 创作（印尼/泰/越/葡/阿）

**Phase 3：深度能力（Week 9-14）**
- 五级上下文压缩 + QueryEngine 七阶段循环
- 九种 Continue 语义 + 四级错误恢复
- 伏笔追踪系统 + Git Worktree 并行审校

**Phase 4：生态与出海（Week 15-20）**
- Buddy 写作伙伴 + Skill 市场
- 出海本地化（印尼/巴西/泰国深度运营）
- Ultraplan 远程卸载（简化版）

**Phase 5：高级协作（Week 21-26）**
- Agent Teams 对等协作 + 实时协作编辑
- 高级数据分析 + 企业版功能

---

## 第八章：风险分析与缓解策略

### 技术风险

| 风险 | 概率 | 影响 | 缓解策略 |
|---|---|---|---|
| OpenCode API Breaking Changes | 高 | 高 | 锁定依赖版本；核心功能使用稳定 API；自动化回归测试 |
| LLM API 价格波动 | 中 | 中 | 多模型路由对冲；本地模型兜底；成本监控 |
| Git Worktree 冲突 | 中 | 中 | 只读审校模式；冲突时自动回退串行 |
| 上下文压缩信息丢失 | 中 | 高 | 压缩前用户确认；关键信息索引保护 |
| 多模型输出风格不一致 | 中 | 中 | 风格锚定机制；后处理统一化 |

### 市场风险

| 风险 | 概率 | 影响 | 缓解策略 |
|---|---|---|---|
| 大厂入局（腾讯/字节/Anthropic） | 中 | 高 | 快速占领小语种空白市场；建立社区壁垒 |
| 小语种 LLM 质量不达标 | 中 | 高 | Qwen 3.5 覆盖 201 语言；人工审核机制 |
| 出海合规问题 | 中 | 高 | 本地优先架构；GDPR/CCPA 合规；内容过滤 |
| 用户付费意愿低 | 中 | 中 | Freemium 模式；Tier 定价 |
| 文化差异导致水土不服 | 中 | 中 | 本地运营团队；A/B 测试；快速迭代 |

### 竞争风险

| 风险 | 概率 | 影响 | 缓解策略 |
|---|---|---|---|
| SudoWrite 增加多语言支持 | 低 | 高 | 多模型路由技术壁垒；小语种先发优势；价格优势 |
| Scrivener 集成 AI 功能 | 中 | 中 | Scrivener 技术债务重，迭代慢；NovelForge AI 原生优势 |
| 中国大厂推出竞品 | 中 | 高 | 出海差异化；开源社区壁垒；多模型灵活性 |
| OpenCode 自身转向小说领域 | 低 | 中 | MIT 许可允许 fork；垂直领域深度；社区生态 |

---

## 最终结论

**NovelForge 战略定位：全球首个 AI 原生、多模型、多分支、多语言的小说自动化编辑器。**

通过借鉴 Claude Code 的 15 项核心架构特性，在 OpenCode 开源基础上构建差异化能力：

- **多AI模型智能路由** — 75+ 模型自由选择，成本与质量的最优平衡
- **多分支故事探索** — 基于 Git Worktree 的物理隔离，真正的平行世界管理
- **Skills/Commands 指令系统** — 动态领域适配 + 小说专用指令集
- **全球多语言原生出海** — UI 25+ 语言，AI 创作 200+ 语言，抢占小语种空白市场

实施周期约 26 周（6 个月），按"高价值低难度优先"原则分 5 个阶段推进。MVP 可在 4 周内交付，具备核心差异化能力。

---

## Sources

1. Claude Code Hooks Guide - https://claudefa.st/blog/tools/hooks/hooks-guide
2. Claude Code Subagents 与 Git Worktree - https://hatohato.jp/blog/core/single.php?id=598
3. KAIROS 持久化助手模式 - https://www.markdown.engineering/learn-claude-code/46-kairos-always-on
4. Ultraplan 深度规划 - https://zhanghandong.github.io/harness-engineering-from-cc-to-ai-coding/part6/ch20c.html
5. Claude Code Agent Loop 深度解析 - https://blog.vincentqiao.com/en/posts/claude-code-agent-loop/
6. Claude Code 源码泄露分析 - https://juejin.cn/post/7628254115258286143
7. OpenCode 官方文档 - https://opencode.ai/docs/
8. OpenCode 插件系统 - https://opencode.ai/docs/plugins/
9. SudoWrite 官网 - https://www.sudowrite.com/
10. Scrivener 官网 - https://www.literatureandlatte.com/scrivener
11. Qwen 3.5 多语言能力 - https://qwenlm.github.io/blog/qwen3.5/
