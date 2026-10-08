# 方案 A MySQL 通道端到端验收报告

## 1. 基本信息

- 日期：2026-10-08
- 分支：feature/fa-orchestrator-skeleton
- 相关 PRD：PRD-12 §8.2 F-05~F-14
- 相关 ADR：ADR-017（LangGraph 多 Agent 编排）
- 执行批次：批次 1 be51786 → 批次 2 a2899a1 → 批次 3（本报告）

## 2. 测试环境

| 项 | 值 |
|---|---|
| MySQL workbench 库 | 192.168.1.6（只读探测与查询） |
| LLM 接入点 | https://ark.cn-beijing.volces.com/api/plan/v3（agent-plan 语义，未使用 /api/v3） |
| 模型 | glm-5.3-flash |
| 凭据 | tmp/fa.env（FA_MYSQL_* / FA_LLM_*，不入库不入日志） |
| Python | 3.11 |
| 竞彩日样本 | 2026-09-29（13 场） |

## 3. 验收链路与结果

### 3.1 数据准备链路（批次 1/2 产出）

命令：`python3 -m fa_orchestrator run --date 2026-09-29 --source mysql`

结果：**13/13 ready**，四路径解析（fid → site_id_mappings → `match`.id）全部命中
（如 1398605→2635、1495540→2655），六模块完整性无缺口。

### 3.2 两阶段分析链路（批次 3 放行）

命令：`run --source mysql --analysis-output ... --recommendation-output ...`

结果：exit 0，13 场全量产出两阶段结论（channel=mysql）：

- review 判定：8 通过 / 5 转人工（manual_review_matches 正常回填）
- 推荐结构：S/D 单元产出，互斥检查 legal=true（如周一003 S+D 双选）
- 降级标注：degradation_notes 3 条随结论落盘（见 §5）

产物：

- docs/reports/testing/mysql_e2e_conclusion_20260929.json
- docs/reports/testing/mysql_e2e_recommendation_20260929.json

### 3.3 LLM agent-plan 冒烟（F-02/F-03 真实接入）

PlannerAgent + LlmGateway.from_env() 真连 agent-plan 端点：

- base_url 确认为 /api/plan/v3（OpenAI 兼容），模型 glm-5.3-flash
- make_plan("2026-09-29") 返回合法计划 JSON：五必填字段齐全，
  6 场次 / 3 并行分组 / data_preflight 含实质预检结论
- 注意：agent-plan 为长响应模式，网关默认 30s 超时不足，本次冒烟按 240s 通过
  （网关默认超时调整列入遗留事项）

## 4. e2e 发现并修复的真实库缺陷（TDD 补测）

| Bug | 现象 | 根因 | 修复 | 回归 |
|---|---|---|---|---|
| BUG-FA-101 | _side_stats TypeError（None 累加） | team_recent_form.score 存在解析失败行，parse 契约 (None, None) 直传违反 agent 数值契约 | _recent_form_block 比分异常行按缺口跳过，不输出 None 项 | 356 passed |
| BUG-FA-102 | drift_of TypeError（None 相减） | heat_analysis.cold_hot 存在 None（热度未计算快照） | drift_of 排除 None 快照，取有效快照首尾差；单有效快照 0.0 | 356 passed |

两缺陷均为 fixture 干净数据掩盖、真实库才暴露的数据口径问题；修复遵循
"缺口如实标注、不编造" 原则，各补 1 个失败测试先行。

## 5. 降级标注（随结论落盘）

1. D9 变盘轨迹为参考公司（澳门）快照序列降级口径，非全市场统计
2. Q9 情报利弊计数的 sentiment 维度库内缺失，均势判定为缺口兜底
3. D10 赛果经 site_match_id 直连，让球折算由复盘层完成（review 子命令暂为 fixture 输入）

## 6. 测试与 CI 本地等价

- `make unit`：356 passed, 4 skipped（批次 3 前基线 346 → +6 批次 2 → +4 批次 3 及缺陷回归）
- lint：ruff 未安装，跳过（与既有基线一致）
- 本地等价命令均在 python/ai/fa_orchestrator 目录执行

## 7. 遗留风险与待确认

| # | 事项 | 等级 | 说明 |
|---|---|---|---|
| 1 | --recommendation-output 与 --analysis-output 落盘同一完整结论 JSON | R2 | 既有实现（无测试锚定），推荐精简结构属后续 Issue |
| 2 | review 复盘子命令仍为 fixture 输入 | R2 | D10 折算属复盘层，方案 A 范围外，需单独 Issue |
| 3 | LlmGateway 默认 30s 超时不足 agent-plan 长响应 | R2 | 冒烟以 240s 通过；网关超时配置化待调整 |
| 4 | CLI 数据准备与 AnalysisRunner.run 内部重复 prepare（六模块重复查询） | R3 | 效率优化，非正确性问题 |
| 5 | LLM planner 未接入 CLI 图驱动（graph_nodes 预留 planner_fn 注入口） | R3 | 按既有设计后置 |
| 6 | 周一003 等 5 场转人工裁决 | 提示 | review 硬门槛正常触发，非缺陷 |

## 8. 结论

**MySQL 通道端到端验收通过**：数据准备（Q1/Q2b/四路径解析）→ 两阶段分析
（Q3/Q5/Q9~Q12/D9/D12 组装与规则计算）→ 推荐结构产出全链路真库打通；
LLM agent-plan 接入点验证合法；缺口与降级口径如实标注。
