# PRD-12 ISS-FA6 fa_orchestrator 目录分层重构 测试报告

## 1. 任务信息

| 项 | 值 |
|---|---|
| 日期 | 2026-10-01 |
| 分支 | feature/fa-orchestrator-skeleton |
| 任务性质 | 结构性重构（refactor）：外部行为不变，仅包内目录分层 |
| 重构前 | 28 个 .py 文件全部平铺在 `fa_orchestrator/` 包根，7 类职责混杂 |
| 方案确认 | AskUserQuestion 用户确认：方案 A 六层职责分层 + CLI 留根 + report_agent 入 report/ |

## 2. 重构后目录结构

```text
python/ai/fa_orchestrator/fa_orchestrator/
├── __init__.py / __main__.py / cli.py / cli_trace.py   # 入口层（包根保留，4 个）
├── configs/        # orchestrator.yaml + OUTPUT_TEMPLATE.md（留包根）
├── graph/          # 图编排内核（4）：state / config / graph_builder / graph_nodes
├── agents/         # 分析决策 Agent（7）：planner / data / fundamental / odds_verify / sentiment / synthesis / review
├── analysis/       # 分析引擎（4）：analysis_runner / odds_math / history_deduction / match_review
├── data/           # 数据接入（5）：data_source / mysql_data_source / sql_contracts / id_resolver / integrity
├── infra/          # 基础设施（2）：llm_gateway / sqlite_checkpointer
└── report/         # 报告输出（2）：report_agent / html_report
```

28 = 4 + 4 + 7 + 4 + 5 + 2 + 2，六个子包各含带中文职责边界注释的 `__init__.py`。

## 3. 变更内容

| 变更类型 | 数量 | 说明 |
|---|---|---|
| git mv 文件移动 | 24 | git mv 保留历史（R 状态） |
| 新增子包 __init__ | 6 | 中文职责边界 + 相关文档引用 |
| import 路径重写 | 76 处 / 38 文件 | 脚本批量重写（包内 31 处 + 测试 45 处） |
| 手工修正 | 3 处 | cli.py 多名 import 拆分 + html_report 导入补漏 + report_agent 模板路径锚点 |
| configs 路径锚点 | 1 处 | report_agent：`Path(__file__).parent` → `parents[1]`（report/ 子包向上取包根） |
| 不变项 | — | cli_trace yaml 锚点（留根不变）、configs/ 位置、全部测试文件位置、测试 fixtures |

合计：52 文件变更，+86 / -83 行。

## 4. 回归测试结果

### 4.1 单元测试（基线不回退）

```bash
make -C python unit
```

```text
304 passed, 4 skipped in 11.40s
```

与重构前基线（304 passed + 4 skipped）完全一致，无新增/丢失用例。

### 4.2 import 冒烟

六子包全链路导入 + compileall 通过。

### 4.3 CLI 端到端冒烟

```bash
cd python/ai/fa_orchestrator
python3 -m fa_orchestrator graph --output /tmp/fa6_graph.mmd          # Mermaid 导出，yaml 锚点 ✓
python3 -m fa_orchestrator run --date 2026-09-29 --fixture …          # fixture 通道，exit 0
python3 -m fa_orchestrator run … --checkpoint-db … --report-dir …     # 图驱动首跑，7 份报告落盘
python3 -m fa_orchestrator run … --checkpoint-db … --resume           # 断点续跑，exit 0
```

报告内容抽验：数据来源声明/模板章节完整（验证 OUTPUT_TEMPLATE.md 加载路径正确）。
人工裁决中断退出码 3 路径由 integration/test_checkpoint_resume.py 覆盖（manual 决策中断续跑 PASSED）。

### 4.4 make rules

```text
错误数: 0  警告数: 0（测试用例注册表检查通过，测试文件位置未变故注册表零改动）
```

## 5. 影响与兼容性

- 包导入路径变更：`fa_orchestrator.X` → `fa_orchestrator.<pkg>.X`（cli/cli_trace 除外）；
  包外无任何代码 import fa_orchestrator（全仓扫描确认，仅历史报告文档提及）。
- CLI 入口 `python -m fa_orchestrator` 不变；`from fa_orchestrator.cli import main` 不变。
- hatchling 打包 `packages = ["fa_orchestrator"]` 自动包含子包，pyproject 无需改动。
- ADR-017 仅约束顶层落位，包内分层不违反 ADR。

## 6. 结论

重构完成，全部回归绿。根目录从 28 个散落文件收敛为 4 个入口文件 + 6 个职责子包 + configs/。
