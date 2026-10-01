# PRD-12 ISS-FA5 端到端验收与复盘对照 测试报告

## 1. 测试环境

| 项 | 值 |
|---|---|
| 日期 | 2026-10-01 |
| 分支 | feature/fa-orchestrator-skeleton（基线：FA4 完成态 264 passed） |
| Python | 3.11 |
| 依赖 | langgraph==1.2.12 / langchain-core==1.6.6 / pyyaml / pymysql（ADR-017 白名单，零新增） |
| 数据 | 样本日 2026-09-29（周二008-014 七场），D10 赛果为演示数据 |

## 2. 测试对象（ISS-FA5 交付物）

| 模块 | 文件 | 说明 |
|---|---|---|
| 复盘对照 | fa_orchestrator/match_review.py | D10 赛果加载 / HAD+HHAD 命中判定 / 撤离+过热规则核验 / 汇总 |
| HTML 报告 | fa_orchestrator/html_report.py | 自包含 Standalone HTML（零外链，离线可打开） |
| CLI | fa_orchestrator/cli.py（review 子命令） | `review --results --recommendation-file --output [--html]` |
| D10 fixture | tests/fixtures/sample_day_20260929_results.json | 赛果演示数据（hhad_result 为让球线折算预计算） |

## 3. 测试结果

### 3.1 单元测试

```bash
cd python && make unit
```

```text
304 passed, 4 skipped in 10.67s
```

新增 4 个测试文件（40 用例）：
- test_match_review.py（28）：赛果加载契约 3 / HAD 单元判定 6 / HHAD 判定 3 / verdict 组合 5 / 规则核验 9 / 汇总 3（含 BUG-FA5-01 修复回归用例）
- test_html_report.py（7）：自包含零外链 / 汇总明细 / 验收表 / 遗留风险节 / 违规高亮 / 落盘 2
- test_review_performance.py（2）：七场并行 ≤ 串行单场×3 / 并行计数完整
- test_cli_review.py（3）：review 端到端落盘 / HTML 产出 / 缺文件退出码 2

### 3.2 集成与端到端

```bash
cd python && make integration                    # checkpoint 断点续跑（免门控）
cd python/ai/fa_orchestrator && python3 -m fa_orchestrator run … --checkpoint-db /tmp/fa5_ckpt.db
python3 -m fa_orchestrator run … --checkpoint-db /tmp/fa5_ckpt.db --resume
python3 -m fa_orchestrator replay --checkpoint-db /tmp/fa5_ckpt.db --thread 2026-09-29
python3 -m fa_orchestrator review --results … --recommendation-file … --output … --html …
```

```text
checkpoint 集成测试：3 passed（免门控）
图驱动首跑：exit 3（人工裁决 1398597，F-20）
resume 续跑：exit 0，7 份单场报告落盘
replay：15 个节点事件按时序回放
review 复盘：命中 3/7（partial 2 / miss 2），规则违规 0 项
MySQL 真实连库：凭据未持久化无法复跑（见 §6 验收 8 与遗留登记）
```

## 4. PRD-12 §14 验收标准逐项核验（第 1-8 条）

| 编号 | 验收标准 | 结论 | 证据 |
|---|---|---|---|
| 1 | run 一条命令端到端 7 份报告，结构符合 OUTPUT_TEMPLATE.md | ✅ 通过 | FA5 e2e：resume 后 7 份报告（/tmp/fa5_reports），结构含预测表/反大众表/矛盾检查节（FA4 ReportAgent） |
| 2 | 溯源声明 100% 覆盖 + 模拟数据告警无遗漏 | ✅ 通过 | FA4 报告产物逐份带数据模式/块编号/快照时间；安全用例 4 passed |
| 3 | thread 可 replay、节点四要素可查、断点续跑 | ✅ 通过 | FA5 e2e：15 事件回放 + exit 3 → resume exit 0；FA4 checkpoint 集成 3 passed |
| 4 | 纪律条款图级强制（反大众前置 / 撤离 ≥3 重分支） | ✅ 通过 | FA3 案例锚点回归（周六012/周日007/008）；样本日无 ≥3 重撤离场（复盘核验 ① not_triggered 属实） |
| 5 | 七项矛盾检查全执行写入报告；互斥由条件边保证 | ✅ 通过 | FA4 推荐产物逐场 review.checks 7 项 + mutual_exclusion.legal |
| 6 | Mermaid 导出与运行时 StateGraph 一致 | ✅ 通过 | FA4 graph 子命令一致性断言（不一致 exit 2）+ 导出产物归档 |
| 7 | Mock/fixture 单测 + make unit 通过 + make rules 通过 | ⚠️ 部分通过 | make unit 304 passed, 4 skipped；make rules 308 项失败为历史遗留（非 PRD-12 引入，见遗留登记） |
| 8 | 需求文档 §七 第 1/3 条（SQL 行数一致 / 五模块入库 >0） | ⚠️ 证据引用 | R-04 联调报告已归档 4 integration passed（真实连库）；本次凭据未持久化无法复跑（凭据不入库为正确安全实践） |

## 5. 多场并行性能验证

| 项 | 值 |
|---|---|
| 方法 | StateGraph 七节点扇出（每节点 sleep 0.1s），一次 invoke 并行执行 |
| 断言 | 并行总耗时 ≤ 串行单场（0.1s）× 3 = 0.3s |
| 结果 | test_七场并行耗时不超过串行单场三倍 PASSED（实际 < 0.2s） |

## 6. 复盘对照结果（D10 演示赛果）

| 统计 | 值 |
|---|---|
| 命中 hit | 3（1398596/1398597/1398598） |
| 部分命中 partial | 2（1398594/1398595） |
| 未命中 miss | 2（1398599/1398600） |
| 规则违规 | 0 项（撤离硬门槛：样本日未触发；过热处置：1398598 处置合规） |

产物：
- [PRD-12-ISS-FA5-复盘对照.json](PRD-12-ISS-FA5-复盘对照.json)
- [PRD-12-ISS-FA5-复盘对照报告.html](../html/PRD-12-ISS-FA5-复盘对照报告.html)

## 7. Bug 列表

| Bug ID | 等级 | 状态 | 说明 |
|---|---|---|---|
| BUG-FA5-01 | P3 | 已修复 | 复盘核验项②口径过窄：仅检查 had_unit.tag，未覆盖 FA3 实际链路（过热层为 hhand 时 tag 生成于 sentiment.overheat.tag + disposition 分档）。修正合规口径为 unit_tagged or layer_tagged，补回归用例 1 个。 |

## 8. 遗留事项登记（草案要求 + 本次新增）

| 事项 | 处置 |
|---|---|
| D9 变盘史无独立表 | 维持上游结论，仅登记（草案既定） |
| checkGate 口径不匹配 | 维持上游结论，仅登记（草案既定） |
| D10 赛果无 MySQL 查询契约 | 新登记：复盘对照当前仅 fixture 通道；MySQL 层扩展需先归档 D10 SQL 契约（需求文档 §四 未定义） |
| FA3 过热标注下沉推荐单元 | 新登记：tag 目前在 overheat 层，下沉至最终推荐单元属 FA3 补强项（核验口径已兼容，不阻断） |
| make rules 308 项历史失败 | 非本次引入，维持既有风险登记，待专项处理 |

## 9. 风险说明

- D10 赛果为演示数据，复盘命中统计不构成对分析质量的评估结论；真实复盘需接入真实赛果源。
- MySQL 复盘扩展依赖 D10 SQL 契约归档（需求文档缺口，需上游补充）。

## 10. 结论

ISS-FA5 交付物全部完成并通过测试：单元 304 passed（+40）、checkpoint 集成 3 passed、端到端验收链（首跑中断→裁决→续跑→回放→复盘→HTML）全通；§14 验收 8 条中 6 条通过、2 条部分通过/证据引用（make rules 历史遗留、MySQL 凭据未持久化），均有既有归档证据或遗留登记支撑。验收报告提交项目主控双评审。
