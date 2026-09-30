# ISS-FA2 测试报告：基本面与赔率验证 Agent（TDD）

| 字段 | 值 |
|---|---|
| 日期 | 2026-09-30 |
| 分支 | feature/fa-orchestrator-skeleton |
| Issue | ISS-FA2（docs/drafts/PRD-12-M1-M5-Issue拆分草案.md） |
| 相关 PRD | PRD-12（F-10/F-11/F-16） |
| 相关 ADR | ADR-017 |
| 行为基准 | .trae/skills/fa-football-analyst/SKILL.md（v7.12）+ OUTPUT_TEMPLATE.md + CONSISTENCY_CHECKS.md |
| 执行人 | Trae |

## 1. 测试环境

- macOS（arm64）/ Python 3.11.0 / pytest 7.4.3
- 依赖（精确锁定，ADR-017 白名单内）：langgraph==1.2.12、langchain-core==1.6.6、pyyaml>=6.0.1
- 外部服务：无。两个分析 Agent 为纯规则计算（不调用 LLM）；全部测试使用 fixture 数据源，禁止真实 MySQL（192.168.1.6:3306）调用

## 2. 测试对象

`python/ai/fa_orchestrator/`（ISS-FA2 新增 4 文件 + 扩展 4 文件）：

| 模块 | 职责 | PRD 功能项 |
|---|---|---|
| fundamental_agent.py | 第一阶段基本面五维分析：情报利弊比与均势约束（v7.7）、初盘定位与同盘路提示（v7.4）、压制力分级与时效衰减（v7.10）、进球外推与火热纪律（v7.2）、杯赛/国际赛修正（v7.8/v7.11）；权重按 v7.12 固定配置 | F-10 |
| odds_verify_agent.py | 第二阶段赔率验证：结构验证（v7.0）、欧亚转换（升盘真信号边界 v7.11、两层背离信让球层）、防冷分级（v7.2）、立博独开分流（v7.9）、单家临场背离（v7.11） | F-11 |
| history_deduction.py | 历史推演：升降盘时序与投票热度漂移推导，缺失标注"无法推演时序" | F-16 |
| analysis_runner.py | 两阶段管线运行器：逐场组装（初盘定位从亚盘澳门行提取、状态基线/实力背书/平局支持联动注入）→ 日级中间结论汇总 | F-10/F-11/F-16 |
| data_source.py（扩展） | analysis fixture 可选加载 + Q3/Q5/Q9~Q12/历史推演查询（未配置时向后兼容返回 None/空） | F-06/F-10 |
| state.py（扩展） | 新增 fundamentals/odds_verification/phase_conclusion 中间态字段 | F-01 |
| orchestrator.yaml（扩展） | 拓扑追加 ready_report→fundamental_agent→odds_verify_agent→phase_report | F-01 |
| cli.py（扩展） | --analysis-fixture/--analysis-output 两阶段结论落盘 | F-05 |

范围说明：热度偏差计算属 F-12（SentimentAgent，ISS-FA3），本 Issue 不含；MarketHeatAgent 相关查询仅保留 heat_drift 原始透出。

## 3. 测试用例与结果

命令：

```bash
cd python && make unit
# 等价：cd python/ai && python3 -m pytest fa_orchestrator/tests/ -v
```

结果：**131 passed**（fa_orchestrator 130 = FA1 60 + FA2 新增 70，hello 既有 1），0 failed，6.31s。

ISS-FA2 分组覆盖：

| 测试文件 | 用例数 | 覆盖点 |
|---|---|---|
| test_history_deduction.py | 8 | 时序计数（asian 深化/大小球升线）、None/空列表 → available=False + "无法推演时序"、热度漂移三态、平稳带 ±1pp 边界 |
| test_fundamental_agent.py | 17 | v7.12 五维权重逐项、利弊五向计数（yl/wl/null 口径）、均势约束（差 ≤1）、密集方向（单方 ≥3 且对方 <3）、同盘路提示（80%/60% 阈值）、压制力分级（连续 3 胜/胜率 70%/不败型 weak/衰减降级/近 5 年窗口） |
| test_fundamental_form.py | 10 | (2+1)×0.9=2.7 外推、热身赛排除、近 5 场均值口径、火热 hot_sides 与 30% 折扣文案、杯赛锚点降级 + 对攻崩坏不得降权、国际赛 0.3 单位轻仓、拉科场锚点回归（外推 3.42 不受降盘影响） |
| test_odds_verify_agent.py | 21 | 结构验证主流公司缺失告警、升盘解读优先级（低水真信号[拉科孤立升盘锚点]→深盘+背书[004]→状态背离诱盘[v7.2]→均势需验证[008]）、两层背离（威廉 0.72 vs 升盘 1.16-1.19）、防冷三级（有效/降级/不构成）、立博独开两分流、单家临场背离（6.7% 触发/5.0% 边界/4.5% 不触发/非临场不触发） |
| test_analysis_pipeline.py | 14 | analysis 数据源扩展查询与向后兼容、AnalysisRunner 七场全部产出（含未就绪场不静默跳过）、案例锚点回归（周二008 升盘低水真信号/周二013 两层背离 + vote 缺失/周二014 单家背离 6.7% + 无法推演时序/周二010 杯赛修正/周二011 国际赛修正/周二012 级别断层压平降级）、CLI --analysis-output 落盘 |

案例锚点回归说明：v7.2-v7.12 锚点（拉科场/弗赖堡场/周六012/周日007/008）以行为断言形式内联在测试中（真实复盘原始数据未经授权不入 fixture）；样本日 fixture 七场按差异化语义构造，覆盖升盘/均势/杯赛/国际赛/级别断层压平/两层背离/单家背离等全部规则分支。

## 4. 端到端验证（CLI）

```bash
cd python/ai/fa_orchestrator && python3 -m fa_orchestrator.cli run --date 2026-09-29 \
  --fixture tests/fixtures/sample_day_20260929.json \
  --analysis-fixture tests/fixtures/sample_day_20260929_analysis.json \
  --analysis-output docs/reports/testing/PRD-12-ISS-FA2-样本日两阶段中间结论.json
```

输出（exit=0，七场差异化结论）：

```text
[分析] 周二008 升盘解读=真信号 标签=- 时序=可推演          # 4 家共振升盘+低水
[分析] 周二009 升盘解读=无升盘 标签=- 时序=可推演          # 均势
[分析] 周二010 升盘解读=无升盘 标签=- 时序=可推演          # 杯赛（升线动作在大小球层）
[分析] 周二011 升盘解读=无升盘 标签=- 时序=可推演          # 国际赛
[分析] 周二012 升盘解读=无升盘 标签=- 时序=可推演          # 级别断层+压平（anomalies.anti_cold=降级）
[分析] 周二013 升盘解读=真信号 标签=[两层背离｜信让球层] 时序=可推演
[分析] 周二014 升盘解读=无升盘 标签=- 时序=无法推演        # 推演数据缺失标注
```

交付物 JSON 已落盘：`docs/reports/testing/PRD-12-ISS-FA2-样本日两阶段中间结论.json`（每场含 integrity/vote_available/fundamental/odds_verify/history_deduction）。

## 5. TDD 过程记录

1. 测试红：实现前 5 个测试文件收集失败（ModuleNotFoundError），失败原因符合预期。
2. 实现绿：history_deduction/fundamental（含 form）/odds_verify 首轮 35/35 通过；管线首轮 74/74 通过。
3. 回归：make unit 首轮 130/131——test_config 对 YAML 节点断言仍为 FA1 三节点（FA2 计划内拓扑扩展），同步更新断言后 131 passed。
4. 规范整改：fundamental_agent 盘口方向魔法字符串（"未知/升盘/降盘/持平"）提为具名常量（rule 6.2）。

## 6. Bug 列表

| Bug | 等级 | 状态 | 说明 |
|---|---|---|---|
| TEST-FN-PCT-CHAR | P2 | 已修复 | 中文测试函数名含 `%` 字符导致 SyntaxError 收集失败（`%` 非合法标识符字符），改为"以上/个点"表述；教训：首轮失败应先逐文件 py_compile 定位，避免被首个收集错误误导 |

## 7. 工程规范核查

- 文件行数：fundamental_agent 280 / odds_verify_agent 252 / analysis_runner 119 / data_source 126 / cli 95，全部 ≤300；方法均 ≤30 行
- 中文注释：核心类/方法均说明职责边界、输入输出与失败行为
- 禁魔法数字/魔法字符串：阈值（低水 1.00/超高水 1.15/超低水 0.75/背离 5%/压平 3%/均势差 1 等）与文案/方向字符串全部具名常量化
- 纯规则计算：两 Agent 不调用 LLM 与外部服务；测试全 fixture、全 Mock
- 数据缺失降级：无升盘→解读 None、无投票→vote_available=False、无时序→"无法推演时序"，禁止编造

## 8. 风险与遗留

- R1（承接 FA1）：dev 分支不存在，分支自 main 创建，待主控确认是否补建 dev
- R2（承接 FA1）：MySqlDataSource 仍为骨架（R-04 未就绪），ISS-FA3 前需确认 MySQL 环境与真实数据口径
- R3（承接 FA1）：`make rules` 308 项失败为存量问题（与本 Issue 无关）
- R4：真实复盘原始数据（拉科/弗赖堡等）未经授权不入 fixture，锚点以行为断言内联；如需数据级回归需主控确认数据授权
- R5：value_ratio（身价差）数据源未定（fixture 恒 None），实力背书目前依赖 level_gap 与 h2h_suppression 两路，待 R-04 就绪后补 Q13 类数据块

## 9. 结论

ISS-FA2 验收标准达成：FundamentalAgent 五维规则与 OddsVerifyAgent 全部判定分支按 v7.12 基准实现并有测试锚定，历史推演缺失降级正确，样本日七场两阶段中间结论 JSON 可产出。可提交评审。
