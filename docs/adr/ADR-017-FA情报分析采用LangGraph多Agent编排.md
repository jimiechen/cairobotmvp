# ADR-017：FA 情报分析采用 LangGraph 多 Agent 编排

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | ADR-017 |
| 名称 | FA 情报分析采用 LangGraph 多 Agent 编排 |
| 状态 | 已确认 |
| 创建日期 | 2026-09-30 |
| 最后更新 | 2026-09-30 |
| 创建人 | 项目团队（Trae 起草，项目主控批准） |
| 关联 PRD | [PRD-12-足球情报分析多Agent可视化编排系统.md](../prd/PRD-12-足球情报分析多Agent可视化编排系统.md) |

## 2. 背景

PRD-12 已评审通过：将 `.trae/skills/fa-football-analyst`（v7.12）两阶段分析法工程化为 FA-Orchestrator 编排系统，解决人工编排效率低、纪律条款（反大众检查、撤离信号硬门槛、七项矛盾检查）执行不可度量、过程不可追溯、单场串行四个问题。

开工前必须解除 PRD-12 §2.2-1 的 R0 阻断点：**LangGraph（含 langchain-core）属大型依赖引入，需 ADR 决策**。同时确认 §2.2-2 的"agent-plan 外接大模型"语义。

## 3. 决策

**引入 LangGraph 作为 FA-Orchestrator 编排内核**，具体决策如下：

1. **编排框架**：采用 LangGraph `StateGraph`（显式图模型：节点 + 条件边 + 子图），承载 v7.12 两阶段分析法的 DAG 与条件分支。
2. **依赖范围**：仅引入 `langgraph` 与其最小依赖 `langchain-core`；**不引入** LangChain 全家桶（agents/chains/memory 等高层封装），分析逻辑全部自研节点。
3. **agent-plan 语义（确认）**：Planner 采用 Plan-and-Execute 模式产出结构化分析计划，支持 Re-Plan；大模型经 OpenAI 兼容 API 外部接入，`base_url/api_key/model` 全配置化（环境变量注入），不内置本地模型，不硬编码密钥。
4. **可视化**：MVP 采用 StateGraph 原生 Mermaid 导出 + 运行轨迹渲染，不自研图形化编辑器。
5. **可追溯**：每场比赛一个 thread，SQLite checkpointer 持久化节点级输入/输出/耗时/数据来源（D1-D12 数据块溯源），支持断点续跑与轨迹回放。
6. **部署形态**：Python 包落位 `python/ai/fa_orchestrator/`（遵循 ADR-0012 多语言 Monorepo 布局，与 hello 同级），CLI 为 MVP 唯一入口，不新增 Protobuf 协议。

## 4. 为什么选 LangGraph

| 维度 | 契合证据 |
|---|---|
| 显式图模型 | v7.12 纪律条款（过热分档、撤离≥3重硬门槛、互斥合法性、七项矛盾检查回退）天然是条件边，StateGraph 保证 100% 执行而非依赖提示词自觉 |
| 状态管理 | 两阶段分析中间态（基本面结论/赔率验证/热度偏差/推荐单元）以 TypedState 在节点间流转，类型安全 |
| 并行子图 | 多场次并行分析 = 每场一个子图实例，单场失败不扩散 |
| 可视化原生 | `draw_mermaid` 零成本满足 PRD-12 F-17 |
| Checkpoint 原生 | 内置 checkpointer 抽象满足 PRD-12 F-18/F-19（断点续跑、轨迹回放） |
| Python 生态 | 与 ADR-0004（AI 服务使用 Python）一致，`python/` 子 Makefile（pytest/ruff）直接复用 |

## 5. 替代方案

### 5.1 方案 A：自研编排内核（asyncio + 显式状态机）

- 优点：零外部依赖、完全可控、无版本演进风险。
- 缺点：可视化导出、checkpoint 持久化、断点恢复、并行子图调度均需从零自建，预估多投入 60%+ 工作量且重复造轮子；可追溯性（PRD 核心 P1 需求）实现质量难保证。
- 结论：❌ 不采用。

### 5.2 方案 B：CrewAI / AutoGen 等对话式多 Agent 框架

- 优点：多角色协作开箱即用。
- 缺点：角色对话驱动与 v7.12"显式 DAG + 硬性纪律条款"模型不符，条件分支与强制检查难以图级保证；偏离 PRD-12 F-01（图拓扑配置化）验收标准。
- 结论：❌ 不采用。

### 5.3 方案 C：维持人工 Skill 对话现状

- 优点：零改造。
- 缺点：PRD-12 §3 所列四个问题原样保留，v7.10-v7.12 连续多轮误判事故已证明人工执行规则遗漏的代价。
- 结论：❌ 不采用。

## 6. 正面影响

1. 纪律条款由图结构强制执行，规则执行可度量、可回归（案例锚点测试集）。
2. 数据块级 + 节点级双层追溯落地，复盘可精确复现分析路径。
3. 多场并行 + 断点续跑，7 场总时长约束（≤ 串行单场 × 3）可达成。
4. 图拓扑 YAML 配置化，v7.x 规则演进可经配置调整而非改核心代码。

## 7. 负面影响

1. 引入 langgraph + langchain-core 依赖树（安装体积与依赖面扩大），Python 环境管理要求提高。
2. LangGraph 版本演进较快，存在 API 变更风险，须锁定版本。
3. 团队需学习 StateGraph 抽象（节点/边/reducer/checkpointer）。
4. LLM 网关外接引入外部服务依赖（费用、时延、可用性），需 Mock 分层隔离。

## 8. 风险与缓解措施

| 风险 | 等级 | 缓解措施 |
|---|---|---|
| LangGraph API 变更导致升级成本 | 中 | pyproject 锁定精确版本；升级前跑全量回归测试 |
| 规则工程化偏差（提示词 → 节点逻辑） | 中 | v7.2-v7.12 全部案例锚点建回归测试集，偏差即测试失败 |
| LLM 外部依赖不稳 | 中 | 网关层重试 → 降级 → 人工裁决链；测试一律 Mock，禁止真实调用 |
| 依赖引入超范围（误入 LangChain 全家桶） | 低 | 依赖清单评审把关：仅 langgraph + langchain-core |

## 9. 约束

1. 依赖白名单：`langgraph`、`langchain-core`；新增其他 AI 框架依赖须重新走 ADR。
2. 版本锁定：pyproject.toml 精确版本号，禁止浮动版本。
3. 单元测试一律 Mock LLM 与外部服务，禁止真实大模型调用（TDD 规则红线）。
4. 密钥/连接串走环境变量，不进代码、日志、Git。
5. 单文件 ≤300 行、单方法 ≤30 行，Agent 按单一职责拆分。
6. 中文注释按 `.trae/rules/commenting.md` 执行，核心节点类说明职责边界。
7. 图拓扑由 YAML 配置定义，核心代码不硬编码图结构。

## 10. 通过判据

```bash
# 1. 依赖白名单核查（pyproject 中不得出现白名单外 AI 框架）
grep -E "langchain(?!\-core)|crewai|autogen" python/ai/fa_orchestrator/pyproject.toml  # 应为空

# 2. 单元测试（Mock LLM）
cd python/ai && make unit  # 全部通过

# 3. 工程规范检查
make rules  # 六项检查通过

# 4. 端到端验收（M5）
python -m fa_orchestrator run --date 2026-09-29  # 七场报告 + 轨迹落盘
```

## 11. 相关文档

- [PRD-12-足球情报分析多Agent可视化编排系统.md](../prd/PRD-12-足球情报分析多Agent可视化编排系统.md) - 需求与验收
- [ADR-0004-AI服务使用Python.md](ADR-0004-AI服务使用Python.md) - 技术栈前提
- [ADR-0012-polyglot-monorepo-directory-layout.md](ADR-0012-polyglot-monorepo-directory-layout.md) - 目录布局依据
- [ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md](ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md) - 工程入口
- `.trae/skills/fa-football-analyst/SKILL.md`（v7.12）- 行为基准规范
