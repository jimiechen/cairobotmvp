# AGENTS.md

## 1. 文件定位

本文件是 CaiRobot MVP 仓库的最高级工程协作规范，用于指导 Trae、AI 编码助手和开发者按照统一流程进行开发。

本文件只约束工程开发流程。详细规则以 `.trae/rules/*.md` 为单一事实来源，本文件为索引和摘要。

> **冲突处理规则**：如本文件摘要与 `.trae/rules/*.md` 详版规则存在冲突，以 `.trae/rules/*.md` 为准。

## 2. 适用阶段

当前项目处于 **S0：规则与目录骨架** 阶段。

| 阶段 | 说明 | 适用规则 |
|---|---|---|
| S0 | 规则建设、目录骨架 | 全部强制（日报可按需启用） |
| S1 | PRD + ADR 落地 | 全部强制 |
| S2 | 核心功能开发 | 全部强制 |
| S3 | 稳定运营 | 全部强制 |

> S0 阶段可由项目主控决定是否启用日报机制。

## 3. 项目开发原则

本项目采用：

- PRD 驱动开发
- Issue 驱动任务拆分
- TDD 测试驱动开发
- 小步提交
- Pull Request 评审
- 文档与代码同步更新

任何**代码类开发任务**都必须先明确需求、再写测试、再写实现。

> **例外**：非代码类变更（`docs/`、`chore/`、`ci/`）可不写测试，但仍需说明变更目的与影响范围。

## 4. 默认语言规范

本仓库默认使用中文进行工程协作，包括：

- PRD
- ADR
- Issue
- Pull Request 描述
- 评审意见
- 测试说明
- 开发任务说明
- 文档说明

以下内容可以使用英文：

- 代码变量名
- 函数名
- 类名
- 文件路径
- 目录名
- JSON 字段
- 命令行
- 第三方库名
- Git commit 类型前缀（详见第 10 节）

## 5. 编码规范摘要

详细编码规范见 [`.trae/rules/coding.md`](.trae/rules/coding.md)，摘要如下：

### 5.1 命名规范

- 变量、参数：lowerCamelCase
- 常量：UPPER_SNAKE_CASE
- 类、接口：UpperCamelCase
- 布尔变量：以 is/has/can/should 开头
- 禁止拼音变量名

> **语言特定规则**：Go 遵循官方 `gofmt` + `golint`，Python 遵循 PEP8 + `ruff`，TypeScript 遵循 ESLint + Prettier。项目级命名规范与语言官方规范冲突时，以语言官方为准。

### 5.2 文件规模限制

- 单文件：推荐 ≤ 300 行，绝对 ≤ 500 行
- 单类：推荐 ≤ 200 行
- 单方法：推荐 ≤ 30 行，绝对 ≤ 50 行
- 单个 PR：推荐 ≤ 10 个文件，推荐 ≤ 300 行代码，超出后必须在 PR 中说明原因

> **自动生成代码不计入上限**：Protobuf 生成代码、ORM 模型、前端路由表、迁移脚本等自动生成文件不计入此上限。

### 5.3 注释规范

- 默认中文注释
- 核心类、核心方法必须有注释
- 注释解释"为什么"和"约束"
- 禁止无意义注释
- TODO/FIXME 必须带说明，格式：`TODO(姓名或模块, 里程碑): 说明原因和后续动作`

### 5.4 类与职责

- 单一职责原则
- 不允许上帝类
- 状态机类只做状态迁移
- 工具类只放纯函数

### 5.5 目录与分支

详见 [.trae/rules/git.md](.trae/rules/git.md)。目录必须按以下分类：

| 分类 | 目录 |
|---|---|
| 必备 | `docs/`、`proto/`、`services/`、`ai/`、`web/` |
| 可选 | `app/`、`hardware/`、`firmware/`、`mechanical/`、`tests/`、`scripts/` |

分支名：`feature/*` `fix/*` `docs/*` `test/*` `hardware/*` `refactor/*` `chore/*` `ci/*`

一个分支只做一类事情。

在开始任何代码或硬件相关实现前，必须满足：

1. 有明确的 PRD 或 ADR 依据。
2. 有明确的 Issue 或任务描述。
3. Issue 中定义了目标、非目标和验收标准。
4. 明确本次任务涉及的目录和模块。
5. 明确需要新增或修改的测试。

如果以上条件不满足，Trae 不应直接实现代码，应先要求补充需求或创建 Issue。

## 7. 标准开发流程

每个任务必须按以下流程执行：

1. 阅读相关 PRD。
2. 阅读相关 ADR。
3. 阅读当前 Issue。
4. 确认任务范围。
5. 编写失败测试。
6. 运行测试并确认失败原因符合预期。
7. 编写最小实现代码。
8. 运行测试并通过。
9. 小范围重构。
10. 再次运行测试。
11. 更新相关文档。
12. 提交 Pull Request。

## 8. TDD 强制要求

所有可测试代码必须遵守红-绿-重构流程：

- 红：先写失败测试。
- 绿：写最小实现让测试通过。
- 重构：在测试通过的前提下优化结构。

禁止：

- 先写实现再补测试。
- 删除失败测试来通过构建。
- 修改测试含义来迁就实现。
- 在没有测试的情况下实现 P0 功能。
- 用 Mock 掩盖核心业务逻辑。

## 9. Git 协作规则

默认分支策略：

- `main`：稳定分支，只合并通过评审的代码。
- `dev`：日常集成分支。
- `feature/*`：功能开发分支。
- `fix/*`：缺陷修复分支。
- `docs/*`：文档变更分支。
- `hardware/*`：硬件相关变更分支。

提交信息建议格式：

- `docs: 初始化项目规范`
- `docs(prd): 添加MVP总纲`
- `test(ai): 添加意图分类测试`
- `feat(ai): 实现基础意图分类`
- `fix(app): 修复设备断连状态处理`
- `refactor(firmware): 简化锁仓状态机`

## 10. 变更边界

Trae 每次任务应保持小范围变更。

禁止：

- 一次性生成大规模项目代码。
- 修改与当前 Issue 无关的文件。
- 擅自新增模块。
- 擅自调整目录结构。
- 擅自修改 PRD 的产品范围。
- 擅自引入大型依赖。
- 擅自改变架构决策。

如确需调整架构，必须新增或更新 ADR。

## 11. 文档同步规则

如果代码行为变化，必须同步更新相关文档。

文档位置：

- 产品需求：`docs/prd/`
- 架构决策：`docs/adr/`
- 接口协议：`docs/api/`
- 测试策略：`docs/testing/`
- 硬件说明：`docs/hardware/`
- 固件说明：`docs/firmware/`
- App 说明：`docs/app/`
- AI 安全策略：`docs/ai-safety/`

## 12. Pull Request 要求

每个 PR 必须说明：

1. 关联 Issue。
2. 本次改动内容。
3. 修改了哪些文件。
4. 运行了哪些测试。
5. 是否更新文档。
6. 是否有风险或未完成事项。

没有测试结果的 PR 不应合并。

## 13. AI 编码助手行为约束

Trae 或其他 AI 编码助手必须：

- 先读规则，再执行任务。
- 先问清楚范围，再写代码。
- 优先生成测试。
- 不擅自扩大功能范围。
- 不生成无关文件。
- 不覆盖已有重要文档。
- 输出修改文件清单。
- 输出测试命令和测试结果。
- 对不确定事项主动提问。

## 14. 当前阶段建议

当前项目处于多系统架构设计阶段，优先工作顺序为：

1. 完善 PRD-00 到 PRD-06。
2. 完善 ADR-0001 到 ADR-0007。
3. 建立 Protobuf 协议定义。
4. 建立 Issue 与 PR 模板。
5. 建立各服务测试框架。
6. 再开始逐个服务实现。

## 15. 每日汇报、Markdown 蒸馏与 LLM Wiki

本项目要求 Trae 每天 22:00 前提交日报。

日报必须记录：

- 今日完成内容。
- 修改文件清单。
- 新增或修改的 PRD。
- 新增或修改的测试。
- 测试命令和结果。
- 测试报告路径。
- 截图证据路径。
- 视频证据路径。
- Bug 列表。
- 事故说明。
- 风险事项。
- 阻塞事项。
- 明日计划。
- 需要项目主控确认的问题。

日报完成后，必须将当天内容蒸馏为 Markdown，保存到：
`docs/reports/distilled/`

蒸馏后的关键信息必须同步更新到：
`docs/wiki/LLM-WIKI.md`

如果当天产生测试报告，应同时输出 Standalone HTML 报告，保存到：
`docs/reports/html/`

Standalone HTML 必须支持离线打开，并包含测试步骤、测试结果、截图、视频、Bug、风险和结论。

### 每日固定工作流

1. 21:30 - 停止扩大开发范围
2. 21:30-21:45 - 整理今日修改文件
3. 21:45-22:00 - 生成日报
4. 22:00 前 - 提交日报
5. 22:00 后 - 将日报蒸馏为 Markdown
6. 同步更新 LLM Wiki

### 任务汇报机制摘要

本项目采用项目主控双评审机制。

Trae 是任务执行者，项目主控 A 和项目主控 B 负责需求拆解、任务评审、风险判断和方向控制。

如果当天没有代码变更，也必须提交无代码变更日报。

发生 Bug、事故、阻塞或测试失败时，必须如实上报。重大事故必须立即上报，不得等到日报。

详细规则见：
`./.trae/rules/reporting.md`

## 16. 用户任务不完整时的主动补齐规则

用户任务描述不一定完整。Trae 不得只机械执行用户显式描述的内容。

当用户提出开发、测试、协议、文档或交付任务时，Trae 必须主动根据本仓库工程规范补齐：

- PRD
- ADR
- Issue
- Protobuf / OpenAPI
- 协议编号注册表
- 单元测试
- 集成测试
- 契约测试
- 边界测试
- CI / Workflow
- 测试报告
- Bug 追溯
- 日报
- Markdown 蒸馏
- LLM Wiki

如果某一项不适用，必须说明"不适用原因"。如果某一项无法完成，必须说明阻塞原因和后续计划。

## 17. 工程级 TDD 流程

所有开发任务必须遵守工程级 TDD 流程：

1. **需求红**：先明确 PRD 验收标准
2. **协议红**：先定义 Protobuf / OpenAPI 契约
3. **测试红**：先写失败测试
4. **实现绿**：写最小实现
5. **报告绿**：输出测试报告和证据
6. **CI 绿**：GitHub Actions 或本地等价命令通过
7. **重构**：测试与 CI 通过后再优化
8. **沉淀**：日报、Markdown 蒸馏、LLM Wiki 更新

### 完成标准

没有 CI 通过或本地等价测试结果，不得宣称任务完成。

如果 CI 暂时无法运行，必须说明：
- 无法运行原因
- 本地等价命令
- 本地结果
- 后续补齐计划

## 18. GitHub CI / Workflow 规则

### 18.1 CI 必须通过

所有 PR 必须通过 GitHub Actions CI 才能合并。

CI 至少包含以下检查：
- `docs-check`：检查关键文档是否存在
- `proto-check`：检查协议编号唯一性和注册表同步
- `go-test`：运行 Golang 测试
- `python-test`：运行 Python 测试
- `web-test`：运行 ReactJS App 测试
- `admin-web-test`：运行 AdminWeb 测试
- `report-check`：检查测试报告、日报、蒸馏、LLM Wiki

### 18.2 跳过规则

如果某一端暂未实现，必须有明确的跳过原因输出，不能静默跳过。

### 18.3 本地等价

如果 CI 无法运行，必须提供本地等价命令和结果。

### 18.4 CI 文件位置

- Workflow 定义：`.github/workflows/ci.yml`
- 检查脚本：`scripts/ci/`

## 19. Skill 索引

本项目的工程规范通过 Skill 体系强制执行。详细规范参见 `.trae/skills/`。

### 19.1 Skill 列表

| Skill 名称 | 触发条件 | Priority | Blocking | 对应 Rules |
|---|---|---|---|---|
| cairobot-active-gap-filling | 任何开发任务、Issue、PR | highest | true | reporting.md |
| cairobot-engineering-workflow | 任务启动、Issue、PR | high | true | - |
| cairobot-tdd-loop | 实现、新增功能、修复 Bug、涉及代码目录 | high | true | tdd.md, testing.md |
| cairobot-proto-registry-guard | 定义协议、新增接口、涉及 proto/ | high | true | docs.md |
| cairobot-ci-gatekeeper | 创建 PR、合并 PR | high | true | tdd.md, testing.md |
| cairobot-coding-standard | 编写代码、涉及代码目录 | medium | false | coding.md |
| cairobot-daily-report | 提交日报、任务完成 | high | true | reporting.md |
| cairobot-git-discipline | 创建分支、提交、创建 PR | medium | false | git.md |
| cairobot-doc-placement | 新增文件、修改目录 | medium | false | docs.md |
| cairobot-html-distillation | 蒸馏、生成报告 | medium | false | reporting.md |

### 19.2 Skill 与 Rules 关系

| 层级 | 文件 | 作用 |
|---|---|---|
| 知识层 | `.trae/rules/*.md` | 详细规范，可读性优先 |
| 执行层 | `.trae/skills/*/SKILL.md` | 触发条件 + 步骤 + 校验 + 阻断 |

Skill 内部用相对路径引用 Rules，例如：

```markdown
详细规则参见：[.trae/rules/tdd.md](../../.trae/rules/tdd.md)
```

### 19.3 Skill 激活顺序

任务启动时的标准激活顺序：

```
cairobot-active-gap-filling
    ↓（缺口扫描完成后）
cairobot-engineering-workflow
    ↓（根据任务类型激活）
cairobot-tdd-loop / cairobot-proto-registry-guard / cairobot-coding-standard / ...
    ↓（任务完成后）
cairobot-daily-report
    ↓（PR 创建前）
cairobot-ci-gatekeeper
```

### 19.4 Skill 违规阻断

以下 Skill 具有 blocking=true，必须在继续执行前完成：

- cairobot-active-gap-filling
- cairobot-engineering-workflow
- cairobot-tdd-loop
- cairobot-proto-registry-guard
- cairobot-ci-gatekeeper
- cairobot-daily-report

### 19.5 本地验证

Skill 可通过以下命令本地验证：

```bash
# 文档检查
python3 scripts/ci/check_required_docs.py

# 协议检查
python3 scripts/ci/check_proto_registry.py

# 报告检查
python3 scripts/ci/check_reports.py
```
