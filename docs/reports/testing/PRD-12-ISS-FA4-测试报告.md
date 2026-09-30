# PRD-12 ISS-FA4 测试报告：可视化、追溯与人工裁决

## 1. 测试环境

| 项 | 值 |
|---|---|
| 日期 | 2026-09-30（跨 2026-10-01 凌晨） |
| 执行者 | Trae IDE |
| 分支 | dev（本地，未推送） |
| Python | 3.11.0（macOS） |
| 核心依赖 | langgraph==1.2.12、langchain-core==1.6.6（ADR-017 白名单，未新增依赖） |
| 数据通道 | fixture（sample_day_20260929），全程无真实外部服务 |

## 2. 测试对象（ISS-FA4 交付范围）

| PRD-12 需求 | 交付物 | 说明 |
|---|---|---|
| F-19 断点续跑 | `fa_orchestrator/sqlite_checkpointer.py` | 基于 langgraph 自带 BaseCheckpointSaver 自实现（依赖白名单禁止 langgraph-checkpoint-sqlite）；实例级互斥锁串行化全部连接操作 |
| F-20 人工裁决点 | `fa_orchestrator/graph_nodes.py` `_human_review` | ReviewAgent manual 决策 → human_review 节点 interrupt；`Command(resume)` 恢复后 human_decisions 落 state |
| F-17 Mermaid 导出 | `graph_builder.export_mermaid` + `mermaid_matches_config` + CLI `graph` | 条件边渲染 `-->|label|`；导出与编排配置一致性断言，不一致 exit 2 |
| F-18 轨迹回放 | state `trace_events`（operator.add）+ CLI `replay`/`show` | 每节点事件含 node/duration_ms/status/output/source |
| F-15 报告输出 | `fa_orchestrator/report_agent.py` + `configs/OUTPUT_TEMPLATE.md` + CLI `--report-dir` | 溯源声明 + 模拟数据异常告警（整数赔率/全同赔/缺主流/欧赔缺失） |
| 图驱动 run | CLI `run --checkpoint-db [--resume] --report-dir` | 首跑 manual 中断 exit 3；resume 后落盘推荐结构 JSON + 单场报告 md |

## 3. 测试用例与结果

### 3.1 单元测试（make unit）

```bash
cd python && make unit
```

```text
264 passed, 4 skipped（4 skipped 为 FA_MYSQL_INTEGRATION 门控的真实连库用例）
基线 217（ISS-FA3）→ 264，新增 47
```

| 测试文件 | 用例数 | 覆盖点 | 注册表 ID |
|---|---:|---|---|
| tests/test_sqlite_checkpointer.py | 9 | put/get_tuple/list/put_writes 契约、thread 隔离、历史版本读取、serde 往返、重读文件 | TC-AI-0022 |
| tests/test_graph_nodes.py | 9 | registry 11 节点、首跑 interrupt、resume 日级结论、回退环有界、trace_events、图驱动与 runner 对齐、报告产物 | TC-AI-0023 |
| tests/test_report_agent.py | 12 | 异常检测 4 类、报告溯源声明/预测表/反大众/矛盾检查/告警 | TC-AI-0024 |
| tests/test_cli_trace.py | 9 | graph/replay/show 子命令、图驱动中断与 resume 落盘 | TC-AI-0025 |
| tests/test_config.py、tests/test_graph_builder.py | 更新 | 11 节点拓扑断言、path_map、checkpointer 透传 | TC-AI-0005/0006/0020 |

### 3.2 集成测试（make integration，checkpoint 免门控）

```bash
cd python && make integration
```

```text
checkpoint 断点续跑集成测试：3 passed（免门控）
真实连库集成测试：跳过（未设置 FA_MYSQL_INTEGRATION=1，R-04 门控保持）
```

| 用例 | 验证点 |
|---|---|
| 节点抛错后从最近 checkpoint 续跑完成 | flaky 节点首跑 RuntimeError → `invoke(None)` 续跑至 phase_report，节点恰执行 2 次 |
| manual 决策中断后续跑完成 | interrupt → `Command(resume={"approved":...})` → 恢复完成，human_decisions 记录裁决 |
| 轨迹可从 checkpoint 读回 | 最新 checkpoint channel_values 含完整 trace 与 trace_events（replay 数据源契约） |

### 3.3 安全用例（tests/safety-cases）

```bash
cd python/ai/fa_orchestrator && python3 -m pytest tests/safety-cases -q
```

```text
4 passed：注入假密钥 FA_LLM_API_KEY 后，终端输出 / checkpoint db 二进制 /
replay+show 输出 / 报告 md 均不含密钥字样（密钥不落日志）
```

### 3.4 CLI 端到端演示（fixture 通道）

| 步骤 | 命令 | 结果 |
|---|---|---|
| graph 导出 | `python3 -m fa_orchestrator graph --output …Mermaid导出.mmd` | exit 0，11 节点 + 条件边 `-->|retry|` 等，一致性断言通过 |
| 图驱动首跑 | `run --date 2026-09-29 --fixture … --checkpoint-db ckpt.db` | exit 3，提示"人工裁决：竞彩日 2026-09-29 有矛盾检查未通过场次 ['1398597']（F-20）" |
| replay 回放 | `replay --checkpoint-db … --thread 2026-09-29` | 12 个节点事件时序渲染：8 个主链节点 + retry 回退环 3 轮（synthesis/review 各 3 次），耗时/状态/来源/输出键全可查 |
| show 查询 | `show --node review_agent` | JSON 数组输出 review 事件（含 attempt 0/1/2 与七项检查结论） |
| resume 续跑 | `run … --checkpoint-db … --resume --recommendation-output … --report-dir …` | exit 0；推荐结构 JSON `total_matches=7`、`manual_review_matches=['1398597']`；7 份单场报告 md（含数据来源声明、告警节） |

e2e 产物：

| 文件 | 说明 |
|---|---|
| docs/reports/testing/PRD-12-ISS-FA4-图驱动Mermaid导出.mmd | F-17 导出源码 |
| docs/reports/testing/PRD-12-ISS-FA4-图驱动中断演示.txt | F-20 中断提示 |
| docs/reports/testing/PRD-12-ISS-FA4-轨迹回放演示.txt | F-18 时序回放 |
| docs/reports/testing/PRD-12-ISS-FA4-轨迹show-review.json | F-18 节点事件查询样例 |
| docs/reports/testing/PRD-12-ISS-FA4-图驱动推荐结构.json | resume 后日级结论 |
| docs/reports/testing/PRD-12-ISS-FA4-单场报告/ | F-15 七场报告 |

## 4. 修复的缺陷（测试红阶段确认失败原因符合预期）

| 编号 | 等级 | 现象 | 根因 | 修复 |
|---|---|---|---|---|
| BUG-FA4-01 | P1 | 图驱动 run 疑似死锁挂起（进程 0 CPU） | 复盘为**无限回退环**：review 结果不含 attempt 字段，节点用 `prev.get("attempt",-1)+1` 回读恒为 0 → 决策永远 retry → 超步无限增长撑爆 checkpointer 写入链 | attempt 计入 review 结果回写 state（0→1→2，与 runner 语义对齐），stream 模式验证 13 超步后稳定中断 |
| BUG-FA4-02 | P1 | langgraph 并发 checkpointer 写入时 put/put_writes 阻塞 | 单 sqlite3 连接多线程交错 execute/commit（check_same_thread=False 仅解除线程检查，不提供互斥） | 按官方 SqliteSaver 模式：实例级 threading.Lock 串行化全部连接操作；list 生成器锁内物化后再迭代 |
| BUG-FA4-03 | P2 | 集成测试 3 用例 ConfigError 路径不存在 | integration 目录比 tests/ 多一层，`parent.parent` 解析错误 | 修正为 `parent.parent.parent` |
| BUG-FA4-04 | P2 | 集成测试轨迹读回 KeyError: trace_events | 假节点只写 trace 未模拟真实节点契约 | 假节点按 GraphNodes._traced 同契约补 trace_events |

## 5. 风险与限制

- 复现依赖（langgraph 1.2.12 + Python 3.11.0）下 observe 到 checkpointer 写入链在并发+中断场景的调度脆弱性；attempt 有界化后本用例集稳定，但**人工裁决恢复的长轨迹（多轮 retry 后 interrupt）**仅覆盖 fixture 样本，真实 LLM 节点接入后需回归（R2）。
- `--resume` 恢复值固定 `{"approved": True}`（MVP 语义"确认放行"），细化裁决表单（部分采纳/驳回）为后续迭代（R2）。
- replay 的 duration_ms 在 fixture 通道为 0ms（本地计算瞬时完成），LLM 接入后才有实际耗时意义（R3）。
- fixture 为演示数据，报告结论不构成分析口径背书。

## 6. 结论

ISS-FA4 验收标准（PRD-12 §14 第 3/6 条、§15）全部满足：make unit 264 passed（4 skipped 为门控跳过）、checkpoint 集成免门控 3 passed、安全用例 4 passed、CLI e2e 五步演示全通过且产物落盘。TDD 8 步闭环完成。
