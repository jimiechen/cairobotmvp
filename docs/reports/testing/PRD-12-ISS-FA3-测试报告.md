# PRD-12 ISS-FA3 测试报告（热度、综合判定与矛盾检查）

## 1. 基本信息

| 字段 | 值 |
|---|---|
| 任务 | ISS-FA3：SentimentAgent（F-12）/ SynthesisAgent（F-13）/ ReviewAgent（F-14）+ 条件回退环 |
| 分支 | feature/fa-orchestrator-skeleton |
| 依据 PRD | [PRD-12](../prd/PRD-12-足球情报分析多Agent可视化编排系统.md) |
| 依据 ADR | [ADR-017](../adr/ADR-017-FA情报分析采用LangGraph多Agent编排.md) |
| 行为基准 | `.trae/skills/fa-football-analyst/SKILL.md`（v7.12）与 `CONSISTENCY_CHECKS.md`（只读） |
| 执行人 | Trae IDE |
| 执行日期 | 2026-09-30 |

## 2. 测试范围与用例分布

本轮新增 66 个用例（注册表编号 TC-AI-0017 ~ TC-AI-0021）：

| 用例 ID | 测试文件 | 用例数 | 覆盖需求 |
|---|---|---|---|
| TC-AI-0017 | tests/test_sentiment_agent.py | 13 | F-12：支持率偏差 pp、20pp/15pp 过热边界、必发形态、过热两档处置 |
| TC-AI-0018 | tests/test_synthesis_agent.py | 26 | F-13：反大众前置①②③、撤离 ≥3 重硬门槛、让球单选 P 值门槛、低赔补位路径 A/B/C、互斥 7 组参数化 |
| TC-AI-0019 | tests/test_review_agent.py | 15 | F-14：七项矛盾检查命中与豁免、pass/retry/manual 决策、回退上限 2 次 |
| TC-AI-0020 | tests/test_graph_builder.py::TestConditionalEdge | 5 | F-01/F-14：条件回退边、manual 兜底有界终止 |
| TC-AI-0021 | tests/test_recommendation_pipeline.py | 7 | 端到端：七场全产出、回退环 attempt 序列、manual_review_matches 登记、CLI 落盘 |

## 3. 测试命令与结果

```bash
cd python && make unit
# 结果：217 passed, 4 skipped（基线 151 passed + 4 skipped；FA3 新增 66 用例全绿；
# skip 均为 R-04 真实 MySQL 联调门控用例，符合预期）
```

```bash
cd python/ai/fa_orchestrator && python3 -m fa_orchestrator run --date 2026-09-29 \
  --fixture tests/fixtures/sample_day_20260929.json \
  --analysis-fixture tests/fixtures/sample_day_20260929_analysis.json \
  --recommendation-output docs/reports/testing/PRD-12-ISS-FA3-样本日推荐结构.json
# 结果：exit 0，七场全部产出热度/综合/矛盾检查/推荐结构
```

## 4. 端到端产物：样本日七场推荐结构

产物：[PRD-12-ISS-FA3-样本日推荐结构.json](PRD-12-ISS-FA3-样本日推荐结构.json)（total_matches=7，manual_review_matches=[1398597]）

| fid | 对阵 | 决策 | 胜平负 | 让球 | 互斥 | 过热 | 撤离信号 |
|---|---|---|---|---|---|---|---|
| 1398594 | 北城红鹰 vs 南岭蓝鲨 | pass | S:win(3星) | D:draw+lose(路径A) | S+D | False | 1 |
| 1398595 | 东川闪电 vs 西湖巨浪 | pass | S:win(3星) | D:win+draw(路径A) | S+D | False | 0 |
| 1398596 | 中原猛虎 vs 滨海飞鱼 | pass | S:lose(3星) | S:lose(4星) | S+S | False | 0 |
| 1398597 | 高原雄鹰 vs 平原野牛 | **manual** | S:win(3星) | D:win+draw(路径None) | S+D | False | 0 |
| 1398598 | 星城流星 vs 月城潮汐 | pass | S:win(**2星**) | 空 | S+空 | True(downgrade_keep) | 0 |
| 1398599 | 火焰城 vs 冰晶城 | pass | S:win(3星) | 空 | S+空 | False | 1 |
| 1398600 | 长风队 vs 短雨队 | pass | S:win(3星) | S:win(4星) | S+S | False | 1 |

行为锚点验证：

1. **低赔补位条款**（1398594）：竞彩 SP 1.45 低赔 → 让球层低赔补位 D 双选路径 A，不与胜平负单选互斥冲突。
2. **过热两档处置**（1398598）：让球方向支持率过热 → 让球层弃单（备注写明"让球"原因）+ 胜平负单元降星保方向（downgrade_keep）。
3. **矛盾检查转人工**（1398597）：让球双选路径无法达标 → 检查⑥命中 → 回退 2 次仍不过 → manual 并登记 manual_review_matches（F-14 条件回退契约真实命中）。
4. **数据缺失降级**（1398599）：无投票块 → 热度 popular=None + "热度数据缺失"标注，让球层空单元，不编造结论。

## 5. Bug 与偏差修复记录

| # | 现象 | 定性 | 处置 |
|---|---|---|---|
| 1 | test_config 节点/边断言停留在 FA2 拓扑 | 测试未随拓扑演进同步 | 更新断言至 FA3 九节点 + 质检链路边 |
| 2 | 条件边缺 manual 兜底用例 YAML 节点写成标量列表 → TypeError 而非 ConfigError | 测试 YAML 缺陷 | 节点改 {name, agent} 字典，走预期校验路径 |
| 3 | 回退环测试断言共享脚本 agent 的全量 calls | 测试契约缺陷（agent 为全场共享） | 改为校验首场消费的前缀序列 calls[:n] |
| 4 | test_review 的 _synthesis helper 无法表达 hhand=None 空单元 | 测试 helper 哨兵缺失 | 引入 _HHAND_DEFAULT 哨兵区分缺省与显式空 |
| 5 | 过热 20pp 用例期望偏差 20.0，实际 72%-50%=22.0 | 测试期望值笔误 | 修正为 22.0（19.9 边界用例仍守门槛） |
| 6 | 强压制用例断言"放弃单选"，实现备注为"放弃让球单选" | 断言字符串不匹配 | 改断言"强压制"+"让球"（备注含"让球"为检查②弱契约，须保留） |
| 7 | SentimentAgent 交锋同向佐证未映射 W/D/L 赛果口径 | **实现缺陷** | side home→W / away→L 映射，占比门槛具名常量 H2H_ALIGNED_RATE |
| 8 | FA1 基本面输出仅含五向计数（无原始 items），FA3 两 Agent 按 items 契约取数恒为 0 | **实现集成缺口** | _intel_favor_count/_has_independent_support 增加 {side}_favor 计数回退 |
| 9 | CLI 推荐产物用例 --fixture 未传值，argparse 解析失败 | 测试缺陷 | 显式传 fixture/analysis-fixture 路径（与 FA2 CLI 约定一致） |

## 6. 风险与待确认

- **检查②弱契约**：让球层空单元是否"静默缺省"依赖备注文本含"让球"字样，建议后续演进为结构化 flag（recorded_reason）。
- **回退环幂等性**：SynthesisAgent 为纯函数，重跑 attempt 输入相同则结论不变（除 attempt 字段）；ISS-FA4 接入 StateGraph 后如引入 Re-Plan/LLM 节点需重新评估重跑语义。
- **fixture 为演示数据**：支持率/P 值分布非真实盘口，七场结论仅用于管线验证，不构成分析口径背书。
- **ISS-FA4 待办**：三 Agent 接入 StateGraph 节点（analyze/verify 签名与节点函数适配）、manual_review 写入 state、checkpointer 轨迹落盘。

## 7. 结论

ISS-FA3 交付完成：三 Agent + odds_math + 条件回退环 + CLI 推荐产物全部落地，`make unit` 217 passed / 4 skipped，端到端七场推荐结构产出且四个行为锚点全部命中。测试用例注册表（TC-AI-0017~0021）、Raw 归档、任务索引已同步。
