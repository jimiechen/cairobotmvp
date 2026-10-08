# 2026-10-08 竞彩比赛数据完整性检查报告

## 1. 基本信息

- 检查日期：2026-10-08
- 竞彩日对象：2026-10-08（当日库内清单 8 场；另 2026-10-09 已预入库 5 场，不在本次范围）
- 分支：feature/fa-orchestrator-skeleton（本次伴随提交：d3a7345）
- 相关 PRD：PRD-12 §8.2 F-05~F-16
- 数据源：workbench 库 192.168.1.6:3306（只读探测与查询，凭据 tmp/fa.env，不入库不入日志）
- 检查链路：Q1 场次清单 → 四路径解析（fid→match.id）→ Q2b 六模块完整性 → 分析数据块行数覆盖（契约 SQL 只读统计）→ mysql 通道全链路分析组装验证

## 2. 结论摘要

| 指标 | 结果 |
|---|---|
| Q1 清单场次 | 8 场 |
| 四路径解析命中 | 7 场（87.5%） |
| 六模块完整性就绪 | 7/7（可核查场次口径，缺口 0） |
| 未解析场次 | 1 场（`75` 米内罗美洲 vs 福塔莱萨，fid=1384258） |
| 分析数据块覆盖 | 三盘/情报/交锋/状态/赛程全齐；让球投票、必发、热度维度存在缺口（见 §5） |
| 全链路分析组装 | 8 场全部产出（BUG-FA-103 修复后），review 6 通过 / 2 转人工 |

**总体判断：数据可支撑当日分析流程，但存在 1 场解析缺口、3 类维度性缺口与 2 项数据质量隐患，已如实标注、不影响"缺口不编造"原则的落地。**

## 3. 数据准备层结果（Q1 + 解析 + Q2b）

命令：`python3 -m fa_orchestrator run --date 2026-10-08 --source mysql`

| 场次 | 对阵 | fid | match.id | 状态 |
|---|---|---|---|---|
| 75 | 米内罗美洲 vs 福塔莱萨 | 1384258 | 未解析 | blocked（六模块全缺） |
| 周三001 | 赫尔辛基火花 vs 国际图尔 | 1489042 | 2942 | ready |
| 周三002 | 巴西国际 vs 科林蒂安 | 1362483 | 2943 | ready |
| 周三003 | 里莫 vs 格雷米奥 | 1362403 | 2944 | ready |
| 周三004 | 布拉干蒂诺RB vs 米拉索尔 | 1362459 | 2945 | ready |
| 周三005 | 维多利亚 vs 沙佩科 | 1362482 | 2946 | ready |
| 周三006 | 博塔弗戈 vs 达伽马 | 1362412 | 2947 | ready |
| 周四001 | 赫尔辛基 vs 瓦萨 | 1489044 | 2948 | ready |

就绪率口径说明：CLI 报告 `7/7` 的分母仅为解析成功场次；按 Q1 清单全量口径有效就绪率为 **7/8（87.5%）**。建议后续 Issue 将未解析场次纳入就绪率分母，避免口径偏乐观。

### 3.1 未解析场次根因（三表对照，只读排查）

- `matches`（500com 清单）：match_no=`75`，米内罗美洲 vs 福塔莱萨，2026-10-08
- `match`（采集表）：存在同场行 id=2913（cc_id=`75`、日期吻合），但 **play_type=`beidan`，缺 `jingcai` 行**
- 四路径解析按需求文档 §2.3 约束过滤 `play_type='jingcai'`，全部未命中（队名亦异写：库内 away=福塔雷萨 vs 清单 福塔莱萨）
- 该场 crawler_match_result 已有赛果（0:1），确认是真实竞彩场次而非噪声

**定性：采集/入库侧缺口**（竞彩玩法行未入库），非解析逻辑缺陷。需上游补抓该场 jingcai 行后重跑。

## 4. 分析数据块行数覆盖（7 场已解析场次，契约 SQL 只读统计）

| 数据块（契约） | 覆盖情况 | 缺口说明 |
|---|---|---|
| Q3 开球时间（`match`） | 7/7 | - |
| Q3 欧盘 odds_europe | 7/7（54~580 行/场，均含初盘） | - |
| Q3 亚盘 odds_asian | 7/7（32~237 行/场，均含初盘） | - |
| Q3 大小盘 odds_overunder | 7/7（32 行/场，均含初盘） | - |
| Q5 投票 vote_snapshot | 7/7 有 had 池（6 快照/场；周四001 仅 2） | **hhand 让球池 7 场全部 0 行** |
| Q5 必发 betfair_trading(standard) | 6/7 | 周四001（2948）无行 |
| Q9 情报 intelligence_text(filled) | 7/7（3~4 条/场） | - |
| Q10 交锋 h2h_record | 7/7（2~7 条/场） | - |
| Q11 状态 team_recent_form | 7/7（22 行/场 = 主客各 11） | 每场 2 行异常比分（见 §5.4） |
| Q12 赛程（自联结） | 依赖 `match` jingcai 行 | 7 场可查 |
| D12 热度 heat_analysis | 7/7 有快照（3 行/场） | **周四001 3 快照 cold_hot 全部 NULL** |
| D9 澳门变盘轨迹（亚盘/大小） | 7/7（亚盘 2~14、大小 2 行/场） | - |
| D10 赛果 crawler_match_result | 7 个 fid 有行（多场已完赛） | 同 fid 存在重复行（见 §5.5） |

## 5. 数据质量发现清单

| # | 发现 | 等级 | 影响与处置 |
|---|---|---|---|
| 5.1 | fid=1384258 `match` 表缺 jingcai 行（仅 beidan） | P1（数据缺口） | 该场无法进入竞彩分析链路；需采集侧补抓后重跑 |
| 5.2 | vote_snapshot 让球池（hhand）7 场全缺 | P2 | 热度验证让球维度降级为均势兜底（既有降级口径），方向判定依赖 had 池 |
| 5.3 | 周四001 heat_analysis 3 快照 cold_hot 全 NULL + 必发无行 | P2 | 热度走向按缺口标注 unknown（本次修复后语义，见 §6）；赛果未出属正常 |
| 5.4 | team_recent_form 每场 2 行异常比分（"vs"等） | P3 | 采集解析噪声；BUG-FA-101 缺口跳过口径已兜住，不计入进失球统计 |
| 5.5 | crawler_match_result 同 fid 重复行（如 1362403 两条 1:1） | P3 | D10 契约按 updated_at/id 倒序取最新不阻断；建议上游去重 |
| 5.6 | 队名漂移：500com 清单 vs `match` 表（赫尔辛基火花/国际图尔 vs 格尼斯坦/图尔库国际；瓦斯科达伽马 vs 达伽马；米拉索尔 vs 米拉索 等） | P2 | 四路径之路径 3（双队名归一）失效，全靠路径 1（cc+日期窗口）命中；路径 4 别名机制在真实表无别名列、实际不可用。一旦 cc 断号即无法解析，建议采集侧统一队名源或落别名表 |

## 6. 伴随修复：BUG-FA-103（TDD，d3a7345）

- 现象：mysql 通道分析组装在周四001（热度全缺场次）抛 `TypeError: '>' not supported between instances of 'NoneType' and 'float'`，当日分析中断
- 根因：BUG-FA-102 在数据源层引入"侧向无有效热度快照 → None"缺口语义，消费端 `history_deduction._classify_drift` 未适配
- 修复：`deduce_heat_drift` 对侧向 None 按缺口标注——主侧缺失时 trend=unknown（不编造走向），有效侧数值照常透传，note 标注缺口；顺手移除缺失键默认 0.0 的编造路径
- 测试：tests/test_history_deduction.py 新增 3 例（红→绿）；`make unit` **365 passed, 4 skipped**
- 真实库验证：重跑后 8 场分析全部产出，周四001 热度走向=unknown 且流程不中断

## 7. 全链路分析组装结果（修复后）

命令：`run --date 2026-10-08 --source mysql --analysis-output ...`（channel=mysql，degradation_notes 3 条随结论落盘）

| 场次 | integrity | review | 热度走向 | 胜平负单元 | 让球单元 |
|---|---|---|---|---|---|
| 75 | blocked | pass | 无法推演（无映射） | S:win(3星) | 空 |
| 周三001 | ready | pass | stable | S:lose(3星) | D:win+lose（路径A） |
| 周三002 | ready | manual | stable | S:win(3星) | D:lose+draw |
| 周三003 | ready | pass | stable | S:win(2星) | S:lose(4星) |
| 周三004 | ready | pass | stable | D:draw+lose（路径A） | 空 |
| 周三005 | ready | pass | stable | D:draw+lose（路径A） | 空 |
| 周三006 | ready | manual | stable | S:lose(3星) | D:win+draw |
| 周四001 | ready | pass | unknown（缺口标注） | D:draw+lose（路径A） | 空 |

转人工 2 场（1362483、1362412）为 review 七项矛盾检查硬门槛正常触发，非缺陷。

## 8. 风险与遗留

| # | 事项 | 等级 | 建议 |
|---|---|---|---|
| 1 | blocked 场次（75）在数据全缺下仍产出推荐单元（S:win 3星）且 review 通过 | R2 | 违反"缺口不编造"精神，建议后续 Issue：integrity 非 ready 场次的推荐结果强制标记"不可采信"或直接抑制输出 |
| 2 | 就绪率分母不含未解析场次（7/7 vs 实际 7/8） | R2 | 后续 Issue 调整 build_integrity_report 口径 |
| 3 | §5.1/5.2/5.6 三类采集侧缺口 | R1 | 需上游（采集管道）确认补抓与队名源统一，建议单独立 Issue 跟踪 |
| 4 | 2026-10-09 已预入库 5 场，预计同样存在 hhand/热度类缺口 | R3 | 明日检查时验证 |

## 9. 检查命令与可复现性

```bash
# 环境凭据（不入库不入日志）
set -a; source tmp/fa.env; set +a

# 数据准备核查（Q1+解析+Q2b）
cd python/ai/fa_orchestrator
PYTHONPATH=fa_orchestrator python3 -m fa_orchestrator run --date 2026-10-08 --source mysql

# 全链路分析组装（含两阶段结论）
PYTHONPATH=fa_orchestrator python3 -m fa_orchestrator run --date 2026-10-08 --source mysql --analysis-output <path>

# 回归基线
make unit   # 365 passed, 4 skipped
```

原始探测数据：tmp/fa_data_check_20261008.json（就绪率报告）、tmp/fa_analysis_check_20261008.json（两阶段结论）、tmp/fa_analysis_blocks_20261008.json（分析块行数统计）。

## 10. 结论

**2026-10-08 竞彩数据完整性检查完成：7/8 场可解析且六模块齐备，分析链路经 BUG-FA-103 修复后全量产出。** 采集侧三类缺口（jingcai 行缺失、hhand 投票全缺、队名漂移）与两项管线口径风险已如实登记，待上游与后续 Issue 处理。
