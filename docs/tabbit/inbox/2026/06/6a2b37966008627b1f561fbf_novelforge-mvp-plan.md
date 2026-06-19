# NovelForge MVP 方案 — 4 周逐周实施计划

**基于 OpenCode v1.17.3，从 0 到可演示的产品原型**

| 项目 | 内容 |
|---|---|
| 文档类型 | MVP 实施计划 |
| 创建日期 | 2026-06-12 |
| 实施周期 | 4 周（20 个工作日） |
| 团队规模 | 3-4 人（前端 + 后端 + 测试） |

---

## MVP 范围与目标

### 核心特性（P0 优先级）

| 特性 | 来源 | Week | 验收标准 |
|---|---|---|---|
| Hook 增强系统 | Claude Code | Week 1 | 敏感词拦截 + 一致性预检可运行 |
| Named Subagents | Claude Code | Week 2 | 角色审校/世界观检查 Agent 可调用 |
| Daily-Log 记忆 | Claude Code | Week 2 | 跨会话角色状态持久化 |
| Skills 渐进披露 | Claude Code | Week 3 | 3 个基础 Skill 可动态加载 |
| Commands 三层指令 | Claude Code | Week 3 | 20+ 小说专用命令可用 |
| 多语言 UI | OpenCode 基础 | Week 4 | 中/英/日/韩 UI 可切换 |

### 技术栈

- **底座**：OpenCode v1.17.3（MIT 许可）
- **前端**：Electron + OpenTUI（Zig + SolidJS）
- **后端**：Node.js / Bun + TypeScript
- **SDK**：`@opencode-ai/sdk`
- **AI 模型**：DeepSeek V4（默认）+ Claude 4.5（情感描写）+ GPT-5（创意）
- **数据存储**：本地文件系统（YAML + Markdown）
- **版本控制**：Git + Git Worktree

### MVP 成功标准

1. 用户可以在 NovelForge 中创建小说项目，使用 `/outline generate` 生成大纲
2. AI 续写内容前自动通过 Hook 进行敏感词拦截和一致性预检
3. 用户可以通过 `@character-checker` 调用角色审校 Agent
4. 关闭并重新打开项目后，角色状态和写作进度自动恢复
5. 打开古风项目时，古风 Skill 自动激活，AI 续写遵循古风规范
6. UI 可在中文/英文/日文/韩文之间切换

---

## Week 1：环境搭建与 Hook 增强系统

**目标**：完成开发环境配置，实现 AI 输出质量关卡

### Day 1（周一）：Fork OpenCode 与开发环境配置

**上午 — Fork 与依赖安装**
- Fork OpenCode 官方仓库到 NovelForge 组织
- 锁定依赖版本到 ~1.17.3
- 安装 Bun runtime 和 Node.js 环境
- 运行 `opencode serve` 验证基础服务启动
- **Deliverable**: 可运行的 OpenCode 本地服务

**下午 — 项目结构初始化**
- 创建 `packages/novelforge/` 目录结构
- 初始化 TypeScript 配置和 ESLint
- 创建插件入口文件 `src/index.ts`
- 配置开发热重载（`bun --watch`）
- **Deliverable**: 插件项目骨架 + 热重载配置

### Day 2（周二）：OpenCode 插件系统深入理解

**上午 — Hook 系统源码阅读**
- 阅读 OpenCode 插件文档和 30+ 事件钩子列表
- 重点研究 `tool.execute.before` 和 `tool.execute.after`
- 阅读 `session.created` 和 `session.idle` 事件
- 理解 Plugin 上下文对象（client/project/directory/worktree/$）
- **Deliverable**: Hook 系统技术笔记

**下午 — Hello World 插件**
- 编写第一个 NovelForge 插件：日志打印 Hook
- 在 `tool.execute.before` 中打印工具名和参数
- 测试插件加载和事件触发
- 调试插件与 OpenCode 核心的事件流
- **Deliverable**: 可运行的 Hello World 插件

### Day 3（周三）：敏感词拦截 Hook

**上午 — 敏感词库设计**
- 设计敏感词数据结构（支持多语言、分类、严重级别）
- 创建 `.novelforge/config/sensitive-words.yaml`
- 实现敏感词匹配引擎（Trie 树或 Aho-Corasick）
- 支持自定义敏感词和通配符匹配
- **Deliverable**: 敏感词库 + 匹配引擎

**下午 — 拦截 Hook 实现**
- 在 `tool.execute.before` 中注入敏感词检查
- 对 `novel_continue` / `novel_polish` 工具的 content 参数扫描
- 实现阻断逻辑：发现敏感词时抛出 Error 并返回原因
- 编写单元测试覆盖各种敏感词场景
- **Deliverable**: 敏感词拦截 Hook + 测试用例

### Day 4（周四）：角色一致性预检 Hook

**上午 — 角色档案数据结构**
- 设计角色档案 YAML Schema（name/aliases/age/personality/relationships/arc）
- 创建示例角色档案 `.novelforge/characters/`
- 实现角色档案解析器
- 设计一致性检查规则（语言风格、行为模式、关系状态）
- **Deliverable**: 角色档案 Schema + 解析器

**下午 — 一致性预检 Hook**
- 在 `tool.execute.before` 中注入一致性预检
- 提取 AI 生成内容中的角色名，与档案对比
- 检测明显矛盾（如角色死亡后再次出现）
- 将一致性警告附加到工具参数中（不阻断，仅提示）
- **Deliverable**: 一致性预检 Hook + 测试用例

### Day 5（周五）：风格匹配检查 + 本周集成

**上午 — 风格匹配检查**
- 设计风格描述数据结构（tone/vocabulary/sentence_length/dialogue_style）
- 实现风格相似度计算（基于关键词和句式特征）
- 在 Hook 中注入风格匹配检查
- 风格不匹配时附加 styleHint 到工具参数
- **Deliverable**: 风格匹配检查 Hook

**下午 — Week 1 集成测试**
- 整合敏感词拦截 + 一致性预检 + 风格匹配三个 Hook
- 编写端到端测试：模拟 AI 续写全流程
- 测试各种边界情况（空内容、超长内容、特殊字符）
- 输出 Week 1 技术总结文档
- **Deliverable**: 集成测试通过 + 技术总结

### Week 1 验收标准

- `opencode serve` 启动后，NovelForge 插件自动加载
- 调用 `novel_continue` 工具时，敏感词自动拦截（测试用例：包含预设敏感词的输入被阻断）
- 角色一致性预检在检测到矛盾时附加警告信息
- 所有 Hook 单元测试通过率 100%

---

## Week 2：Named Subagents + Daily-Log 记忆

**目标**：实现专业化 Agent 分工和跨会话记忆持久化

### Day 6（周一）：Named Subagents 基础架构

**上午 — OpenCode Agent 系统研究**
- 阅读 OpenCode Agent 文档（Markdown 定义 + @mention 调用）
- 研究 `permission.task` 的 glob 模式控制
- 理解主 Agent（Primary）和子 Agent（Subagent）的通信机制
- 分析 Claude Code Named Subagents 的 YAML frontmatter 格式
- **Deliverable**: Agent 系统技术笔记

**下午 — 角色审校 Agent 定义**
- 创建 `.novelforge/agents/character-checker.md`
- 编写 Agent 描述：角色行为一致性检查器
- 配置工具白名单（read/grep/novel_character_check/write:false/bash:false）
- 定义系统提示词：职责、检查项、输出格式
- **Deliverable**: character-checker Agent 定义文件

### Day 7（周二）：角色审校 Agent 实现

**上午 — 自定义工具开发**
- 使用 Zod Schema 定义 `novel_character_check` 工具参数
- 实现工具逻辑：读取角色档案 + 读取目标章节 + 对比分析
- 设计输出格式：问题列表（描述/段落/严重程度/修复建议）
- 注册工具到插件系统
- **Deliverable**: novel_character_check 工具实现

**下午 — Agent 调用测试**
- 通过 `@character-checker 检查第三章中主角的行为` 测试调用
- 验证 Agent 只能访问白名单中的工具
- 测试 Agent 输出格式是否符合预期
- 调试 Agent 与主 Agent 的上下文传递
- **Deliverable**: character-checker Agent 可正常运行

### Day 8（周三）：世界观检查 Agent

**上午 — 世界观数据结构**
- 设计世界观设定 YAML Schema（rules/levels/locations/history）
- 创建示例世界观 `.novelforge/world/cultivation-system.yaml`
- 实现世界观解析器
- 设计世界观一致性检查规则
- **Deliverable**: 世界观 Schema + 解析器

**下午 — 世界观检查 Agent**
- 创建 `.novelforge/agents/world-checker.md`
- 实现 `novel_world_check` 自定义工具
- 测试 `@world-checker 检查第 5 章修炼体系描述`
- 验证世界观矛盾检测（如炼气期角色使用金丹期能力）
- **Deliverable**: world-checker Agent + 工具实现

### Day 9（周四）：Daily-Log 记忆系统

**上午 — 记忆数据结构设计**
- 设计 Daily-Log 文件结构 `.novelforge/memory/logs/YYYY/MM/YYYY-MM-DD.md`
- 定义记忆条目格式（type/timestamp/content/importance）
- 实现记忆提取器：从会话中提取关键变更（角色状态/世界观变更/新增伏笔）
- 设计记忆索引文件，加速查询
- **Deliverable**: Daily-Log 数据结构设计文档

**下午 — 记忆写入与读取 Hook**
- 在 `session.idle` Hook 中实现记忆自动写入
- 在 `session.created` Hook 中实现记忆自动读取
- 限制加载最近 7 天的记忆（避免日志膨胀）
- 实现记忆格式化：将历史记忆注入系统提示词
- **Deliverable**: Daily-Log 写入/读取 Hook

### Day 10（周五）：跨会话记忆验证 + 本周集成

**上午 — 跨会话记忆测试**
- 测试场景 1：创建角色 → 关闭项目 → 重新打开 → 验证角色状态恢复
- 测试场景 2：修改世界观规则 → 关闭项目 → 重新打开 → 验证规则记忆
- 测试场景 3：多轮会话后，验证历史记忆是否正确累积
- 测试日志归档：旧日志自动按周/月归档
- **Deliverable**: 跨会话记忆测试通过

**下午 — Week 2 集成测试**
- 整合 Named Subagents + Daily-Log + Week 1 Hook 系统
- 编写端到端测试：完整创作流程（大纲 → 续写 → 审校 → 记忆）
- 性能测试：Agent 调用延迟 < 3 秒，记忆加载 < 1 秒
- 输出 Week 2 技术总结文档
- **Deliverable**: 集成测试通过 + 技术总结

### Week 2 验收标准

- `@character-checker` 和 `@world-checker` 可通过 @mention 正常调用
- 角色审校 Agent 能检测出角色行为与档案的矛盾
- 世界观检查 Agent 能检测出违反世界设定的描述
- 关闭并重新打开项目后，角色状态和世界观变更自动恢复
- 所有 Agent 和记忆测试通过率 100%

---

## Week 3：Skills 渐进披露 + Commands 指令体系

**目标**：实现动态领域适配和小说专用指令集

### Day 11（周一）：Skills 系统设计

**上午 — Skill 文件格式设计**
- 设计 Skill Markdown 格式（YAML frontmatter + 知识库内容）
- 定义 match 条件：glob 文件匹配 + regex 内容匹配
- 设计 Skill 元数据（name/description/author/version）
- 确定 Skill 存储位置 `.novelforge/skills/`
- **Deliverable**: Skill 文件格式规范

**下午 — Skill 加载引擎**
- 实现 Skill 发现器：扫描 `.novelforge/skills/` 目录
- 实现 Skill 匹配器：根据项目文件和内容匹配相关 Skill
- 实现 Skill 加载器：解析 YAML frontmatter 和 Markdown 内容
- 实现 Skill 管理器：维护已加载 Skill 的内存缓存
- **Deliverable**: Skill 加载引擎实现

### Day 12（周二）：基础 Skills 开发

**上午 — 古风武侠 Skill**
- 创建 `skills/ancient-wuxia.md`
- 编写写作规范（标点/招式标注/时间表述）
- 整理常用套路（开局/中期/结局）
- 编写禁忌清单（现代词汇/西方元素）
- **Deliverable**: 古风武侠 Skill 文件

**下午 — 科幻/悬疑 Skill**
- 创建 `skills/sci-fi.md`：硬科幻设定规范、科技名词库
- 创建 `skills/mystery.md`：悬疑推理结构、伏笔技巧、红鲱鱼手法
- 为每个 Skill 编写示例 prompt 模板
- 测试 Skill 匹配：创建不同 genre 项目，验证 Skill 自动激活
- **Deliverable**: 3 个基础 Skill 文件 + 匹配测试

### Day 13（周三）：Skills 与 AI 集成

**上午 — 系统提示词注入**
- 在 `session.created` 中注入已加载 Skill 的知识库内容
- 设计提示词格式：Skill 内容作为 system prompt 的附录
- 控制注入长度：单个 Skill 不超过 2000 tokens
- 实现 Skill 优先级：多个 Skill 匹配时的排序逻辑
- **Deliverable**: Skill 提示词注入机制

**下午 — Skills 效果验证**
- 测试：古风项目续写时，AI 是否使用古风词汇和句式
- 测试：科幻项目续写时，AI 是否遵循科技设定
- 测试：悬疑项目续写时，AI 是否埋设伏笔
- 对比测试：同一 prompt 在有/无 Skill 时的输出差异
- **Deliverable**: Skills 效果对比测试报告

### Day 14（周四）：Commands 体系架构

**上午 — Commands 注册系统**
- 设计 Command 接口（name/description/type/args/execute）
- 实现 Command 注册器：支持插件动态注册命令
- 实现 Command 解析器：解析用户输入的 `/command args`
- 设计 Command 帮助系统：自动生成命令文档
- **Deliverable**: Command 注册/解析系统

**下午 — Local 命令实现**
- 实现 `/branch create/list/switch` 命令（本地 Git 操作）
- 实现 `/export epub/pdf` 命令（文件导出）
- 实现 `/foreshadow list` 命令（伏笔列表）
- 实现 `/memory search` 命令（记忆搜索）
- **Deliverable**: 8+ Local 命令实现

### Day 15（周五）：Prompt 命令 + Commands 与 Skills 联动

**上午 — Prompt 命令实现**
- 实现 `/outline generate`：调用 AI 生成全书大纲
- 实现 `/continue`：结构化续写 prompt（注入角色档案 + 前文摘要）
- 实现 `/character create`：交互式角色创建向导
- 实现 `/review consistency/plot/style`：审校命令
- **Deliverable**: 10+ Prompt 命令实现

**下午 — Commands-Skills 联动 + 集成**
- 实现 Skills 激活时自动注册专属 Commands
- 古风 Skill 激活时注册 `/poetry insert` / `/idiom suggest`
- 科幻 Skill 激活时注册 `/tech explain` / `/timeline build`
- Week 3 集成测试：完整 Commands 列表验证（20+ 命令）
- **Deliverable**: 20+ Commands + Skills 联动 + 集成测试

### Week 3 验收标准

- 创建古风项目时，古风 Skill 自动激活，AI 续写使用古风词汇
- `/outline generate` 可生成符合 genre 规范的大纲
- `/continue` 续写时自动注入角色档案和前文摘要
- `@character-checker` 可通过 `/review consistency` 快捷调用
- 全部 20+ Commands 可用，帮助文档自动生成

---

## Week 4：多语言 UI + 集成测试 + Demo

**目标**：完成多语言支持，通过全量测试，输出可演示的 MVP

### Day 16（周一）：多语言 UI 框架

**上午 — i18n 架构设计**
- 调研 OpenCode 现有 i18n 实现（25+ 语言支持）
- 设计 NovelForge 专属 i18n 架构（UI 文本 + Commands 描述 + Skill 内容）
- 创建翻译文件结构 `locales/{lang}/`
- 实现语言切换机制（运行时切换，无需重启）
- **Deliverable**: i18n 架构设计文档

**下午 — 中文/英文翻译**
- 提取所有 UI 文本到翻译文件
- 完成 `locales/zh/` 中文翻译
- 完成 `locales/en/` 英文翻译
- 验证翻译完整性：所有用户可见文本均有翻译
- **Deliverable**: 中/英文翻译文件

### Day 17（周二）：日文/韩文翻译 + RTL 准备

**上午 — 日文/韩文翻译**
- 完成 `locales/ja/` 日文翻译
- 完成 `locales/ko/` 韩文翻译
- 处理日文/韩文特有的排版问题（换行、标点）
- 验证 Commands 在日文/韩文环境下的显示
- **Deliverable**: 日/韩文翻译文件

**下午 — RTL 排版准备**
- 评估阿拉伯语/希伯来语 RTL 排版需求
- 在 CSS 中预留 RTL 支持（`dir="rtl"` 适配）
- 测试 UI 在 RTL 模式下的布局
- 记录 RTL 完整支持所需的后续工作（Phase 2）
- **Deliverable**: RTL 评估报告 + CSS 预留

### Day 18（周三）：端到端测试

**上午 — 测试用例编写**
- 编写 E2E 测试：完整创作流程（创建项目 → 生成大纲 → 续写 → 审校 → 导出）
- 编写 Hook 测试：敏感词拦截、一致性预检、风格匹配
- 编写 Agent 测试：character-checker、world-checker 调用
- 编写记忆测试：跨会话状态恢复
- **Deliverable**: 50+ 测试用例

**下午 — 测试执行与修复**
- 执行全部测试用例，记录失败项
- 修复 P0 级别 bug（阻断主流程的问题）
- 评估 P1 级别 bug（影响体验的问题），记录到 Phase 2
- 输出测试覆盖率报告
- **Deliverable**: 测试报告（覆盖率 > 80%）

### Day 19（周四）：性能优化与文档

**上午 — 性能优化**
- 优化 Skill 加载：懒加载 + 缓存，避免启动时全量加载
- 优化 Agent 调用：减少不必要的上下文传递
- 优化 Daily-Log 读取：索引加速，避免全量扫描
- 基准测试：启动时间 < 5 秒，命令响应 < 1 秒
- **Deliverable**: 性能基准测试报告

**下午 — 技术文档**
- 编写《NovelForge 插件开发指南》
- 编写《Skills 开发规范》
- 编写《Commands 开发规范》
- 编写《Agent 定义规范》
- **Deliverable**: 4 份技术规范文档

### Day 20（周五）：Demo 准备与 MVP 发布

**上午 — Demo 脚本与录制**
- 编写 Demo 脚本：5 分钟展示 NovelForge 核心能力
- 场景 1：创建古风项目 → Skill 自动激活 → AI 续写古风内容
- 场景 2：调用 `@character-checker` 审校角色一致性
- 场景 3：切换语言为日文/韩文，验证 UI 多语言
- 录制 Demo 视频（屏幕录制 + 旁白）
- **Deliverable**: 5 分钟 Demo 视频

**下午 — MVP 发布**
- 创建 GitHub Release：v0.1.0 MVP
- 编写 Release Notes：功能列表、已知问题、后续计划
- 打包安装程序（Electron 应用 + 插件包）
- 输出《MVP 总结报告》：4 周成果、技术债务、Phase 2 计划
- **Deliverable**: GitHub Release + MVP 总结报告

### Week 4 验收标准（MVP 最终验收）

- UI 可在中文/英文/日文/韩文之间切换，所有文本正确翻译
- 完整创作流程 E2E 测试通过（创建 → 大纲 → 续写 → 审校 → 导出）
- 测试覆盖率 > 80%，P0 bug 清零
- 5 分钟 Demo 视频可完整展示核心差异化能力
- GitHub Release v0.1.0 可下载安装

---

## 交付物清单

### 代码交付物

| 交付物 | 路径 | Week | 说明 |
|---|---|---|---|
| NovelForge 插件主包 | `packages/novelforge/` | W1-W4 | TypeScript 源码，包含所有核心功能 |
| Hook 增强系统 | `src/hooks/` | W1 | 敏感词拦截 + 一致性预检 + 风格匹配 |
| Named Subagents | `.novelforge/agents/` | W2 | character-checker + world-checker |
| Daily-Log 记忆 | `src/memory/` | W2 | 跨会话记忆持久化 |
| Skills 系统 | `src/skills/` | W3 | Skill 加载引擎 + 3 个基础 Skill |
| Commands 体系 | `src/commands/` | W3 | 20+ 小说专用命令 |
| 多语言 UI | `locales/` | W4 | 中/英/日/韩 4 语言翻译 |

### 文档交付物

| 交付物 | Week | 说明 |
|---|---|---|
| 《NovelForge 插件开发指南》 | W4 | 如何基于 OpenCode 开发 NovelForge 插件 |
| 《Skills 开发规范》 | W4 | Skill 文件格式、匹配规则、知识库编写 |
| 《Commands 开发规范》 | W4 | Command 接口、注册、测试规范 |
| 《Agent 定义规范》 | W4 | Agent Markdown 格式、工具白名单、系统提示词 |
| 《MVP 测试报告》 | W4 | 测试覆盖率、性能基准、已知问题 |
| 《MVP 总结报告》 | W4 | 4 周成果、技术债务、Phase 2 计划 |

### 媒体交付物

| 交付物 | Week | 说明 |
|---|---|---|
| 5 分钟 Demo 视频 | W4 | 展示 NovelForge 核心差异化能力 |
| GitHub Release v0.1.0 | W4 | 可下载的安装包 + Release Notes |

---

## 风险与应对

| 风险 | 概率 | 影响 | 应对策略 | 责任 Week |
|---|---|---|---|---|
| OpenCode API Breaking Change | 中 | 高 | 锁定 ~1.17.3；每日关注 OpenCode changelog；核心功能使用稳定 API | W1 |
| LLM API 响应慢/超时 | 高 | 中 | 设置 10 秒超时；实现 fallback 到备用模型；本地模型兜底 | W1 |
| Agent 调用上下文传递失败 | 中 | 中 | 增加上下文序列化/反序列化校验；子 Agent 失败时回退到主 Agent | W2 |
| Daily-Log 文件膨胀 | 中 | 低 | 限制单文件大小（< 1MB）；自动按周归档；启动时只加载最近 7 天 | W2 |
| Skill 匹配不准确 | 中 | 中 | 提供手动覆盖机制；用户可强制加载/卸载 Skill；收集反馈优化匹配规则 | W3 |
| 翻译不完整/错误 | 高 | 低 | 使用 i18n 扫描工具检测缺失翻译；社区贡献翻译；Phase 2 完善 | W4 |
| 测试时间不足 | 中 | 高 | W4 前两天专注测试；P0 bug 必须修复，P1 bug 可延期；自动化测试减少人工 | W4 |

### 每日站会建议

- **时间**：每天上午 10:00，15 分钟
- **参与**：全团队（前端/后端/测试）
- **内容**：昨日完成 / 今日计划 / 阻塞问题
- **输出**：阻塞问题当日必须解决或升级

### 周回顾建议

- **时间**：每周五下午 4:00，1 小时
- **内容**：本周成果演示 / 验收标准检查 / 下周计划调整
- **输出**：周总结文档 + 下周详细任务分配
