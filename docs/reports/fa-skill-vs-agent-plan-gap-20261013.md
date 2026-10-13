# FA Skill 报告 vs Agent-Plan 能力差距分析（研发进度反推）

> 日期：2026-10-13　|　分析人：Trae
> 输入：skill（FA v7.15 两阶段分析法）当日 3 份实盘报告（亚冠精英 3 场，
> 文件模式数据源）；agent-plan（PRD-12 FA-Orchestrator）v4 真库产物
> （2026-10-08 竞彩日 7 场，MySQL 通道）
> 目的：以 skill 终态报告为基准，反推 agent-plan 研发进度与下一阶段优先级

## 一、结论

- **agent-plan 已跑通"分析结构骨架"**：v4 两阶段链路 7/7 场产出
  fundamental → odds_verify → history_deduction → sentiment → synthesis →
  review → recommendation 全套结论，推荐结构（S/D 单元、路径合规、互斥
  修正）与撤离/反大众/过热判定已实测工作
- **距 skill 终态报告约 50%**：五部分中第三部分（人性贪婪恐惧）整体缺失，
  第四部分的比分池/大小球层（ISS-FA4）未产出（review ①③④ 全 na），
  第二部分的凯利/去水概率/全市场方向统计、情报语义利弊判读缺失
- skill 单报告由 LLM 一次性生成（人工级语义判读）；agent-plan 为确定性
  编排——**缺口不是"写报告"，而是分析算子与数据维度的工程化补齐**

## 二、板块映射（skill 五部分 → agent-plan v4 实测）

### 第一部分：数据溯源报告 —— 覆盖度 ~60%

| skill 能力 | agent-plan 现状 | 缺口 |
|---|---|---|
| 14 项完整性检查表 | 就绪率 6/7 + integrity 模块计数（Q2b 六模块） | 检查粒度粗（布尔/行数级，无"初/即时盘齐备""时间戳新鲜度"） |
| 数据时间戳标注 | fetch_trace 有查询痕迹，无 per-块时间戳 | P2 |
| 无数据告警（成色折扣标注） | DEGRADATION_NOTES 通道级 | 场次级成色标注缺 |

### 第二部分：v7.0 两阶段分析 —— 覆盖度 ~55%

| skill 能力 | agent-plan 现状（v4 实测字段） | 覆盖 | 缺口 |
|---|---|---|---|
| 情报利弊比（30%） | fundamental.intelligence：home_favor 等计数 | 结构有 | **语义判读缺**：v4 实测全 0 恒"均势"——skill 逐条判读利弊归属，agent 未实现条目级正负判定 |
| 初盘定位（25%） | fundamental.initial_position：初/即盘口、movement、同盘路 hint | 结构有 | 分歧盘逐家归因（检查 A：孤立深开 vs 退防）缺；主流 16 家逐家水位表缺 |
| 历史交锋（15%） | fundamental.h2h：win_rate/suppression/decay | **有** | — |
| 近期状态（20%） | fundamental.recent_form 数据齐 | 数据有 | **进球数外推（N6 分段系数→预期总进球）缺**——skill 大小球主锚 |
| 赛程影响（10%） | fundamental.schedule | 有 | 体能/战意语义解读属 LLM 层 |
| 赔率结构验证（40%） | odds_verify.structure：company_count/mainstream_missing/avg_payout | 部分 | **凯利指数、立博基准去水概率、全市场升赔/降赔家数统计缺**（skill 的"30 家全升/29 家降"与 Bet365 凯利 1.01 单家背离检测） |
| 欧亚转换验证（25%） | odds_verify.euro_asia：raise_form/state_alignment/deep_market/strength_endorsed/interpretation（真信号实测）/two_layer_divergence | **有** | N4 穿盘预警联动（升深→比分池/补位减仓）缺（依赖 ISS-FA4） |
| 市场热度验证（20%） | sentiment.had/hhand：deviation_pp + overheat（18.2pp 放弃判定实测）+ bf trend falling（v4 新增） | **有** | 庄家 pnl 指数缺（MySQL betfair_trading 无 pnl 列——**数据源缺口**）；必发价位聚集度分析缺 |
| 异常信号排查（15%） | odds_verify.anomalies（anti_cold/ladbrokes_solo）+ synthesis.evacuation | 部分 | 信号逐条判定表（6 类）未结构化输出 |

### 第三部分：人性贪婪与恐惧分析 —— 覆盖度 0%

- 贪婪/恐惧评分（0-100）、陷阱等级、四信号检测——**PRD-12 编排未规划
  对应 agent**，属范围外新增，需扩范围评审
- 反大众前置检查（v7.10）：synthesis.anti_public 三锚点（anchor1-3 +
  allow_counter 实测"全不命中允许反买"）——**已有**（属第四部分能力，
  skill 归入人性篇）

### 第四部分：综合预测结论 —— 覆盖度 ~65%

| skill 能力 | agent-plan 现状 | 覆盖 |
|---|---|---|
| v7.0 首选/次选 | synthesis.recommendation（had_unit form/legs/path/stars/tag） | 有 |
| 竞彩让球预测（v7.3） | sentiment.hhand 热度偏差 + synthesis.hhand_unit（本例互斥砍剩 null，机制在） | 有 |
| 撤离/过热/反大众联动 | evacuation + anti_public + overheat（实测 abandon 判定） | 有 |
| **比分预测（1:2/1:3/1:1）** | 无 | **缺（ISS-FA4）** |
| **大小球推荐** | 无 | **缺（ISS-FA4）** |
| 穿盘预警联动仓位（0.5→0.25） | 无 | 缺（依赖 ISS-FA4） |
| 报告矛盾检查（12 项） | review.checks ①-⑧+A-D+杯赛，pass/na 判定实测 | 框架有，na 项待 ISS-FA4 |

### 第五部分：投注建议 —— 覆盖度 ~40%

- 推荐方案单元/星级/路径：recommendation 有
- **亚盘推荐、大小球推荐、仓位单位、加仓/减仓操作建议：缺**
- **数据完整性评分（92/100 量化）：缺**（就绪率为布尔级）

## 三、缺口清单（按优先级）

### P0（阻断 skill 终态报告核心板块）

1. **ISS-FA4 比分池/大小球层**：解锁 review ①③④ na 项 + N6 进球外推 +
   N4 穿盘联动 + 大小球/比分推荐——当前最大缺口，一项解锁四类能力
2. **凯利/去水/全市场方向统计**：odds_verify.structure 增强（数据已在
   odds_europe，纯计算）
3. **情报语义利弊判读**：intelligence 条目级正负判定（可 LLM 辅助批注 +
   确定性聚合，或规则关键词先行）

### P1（提升报告质量）

4. 数据完整性量化评分（92/100 型）
5. 分歧盘逐家归因（检查 A）与主流水位逐家表
6. 庄家 pnl 指数——**需数据源扩容**：betfair_trading 无 pnl 列，与宿主确认
   touzhu 文件数据（pnl/价位聚集）是否入库
7. 亚盘/大小球推荐与仓位操作建议层（recommendation 扩展）

### P2（体验层）

8. per-块数据时间戳、场次级成色标注
9. 人性贪婪恐惧分析模块——**需先立项评审**（PRD-12 未规划，扩范围）

## 四、数据源差异备注

| 维度 | skill（文件模式） | agent-plan（MySQL 通道） |
|---|---|---|
| 欧赔 | 30 家 | 27 家（库内 bookmaker 覆盖） |
| 必发 | data.touzhu：成交+pnl+价位 | betfair_trading：成交占比+v7.4 后 trend/large_volume，**无 pnl** |
| 情报 | 7M/足球财富/500com 前瞻四源全文 | intelligence_text 四源（items 已通） |
| 同盘路历史 | 7M odds.history | 库内无——same_handicap_hint 缺失如实标注 |

## 五、研发进度反推

| 里程碑 | 进度 | 依据 |
|---|---|---|
| M1 数据通道（MySQL） | **100%** | v4 就绪率 6/7（80 缺 vote 为库内实情） |
| M2 数据完备性 | **100%**（库内已有数据口径） | D9 多公司轨迹 + 必发全类型 + 欧赔 history 三项改造，v4 矩阵核验 |
| M3 分析链路（两阶段） | **75%** | 6 agent 结论结构齐且实测工作；缺凯利/去水/语义利弊/进球外推算子 |
| M4 可视化编排（图驱动） | **90%** | fixture 通道全链路验证；mysql 通道被 cli_trace.py L119-121 拦截待放开 |
| M5 收尾（报告/复盘/CI） | **70%** | 就绪率/两阶段/单场 md 有；standalone v4、审计复核节、mysql 图驱动、commit 待做 |
| **skill 终态对齐度** | **~50%** | 五部分：一 60% / 二 55% / 三 0% / 四 65% / 五 40% |

**下一阶段建议顺序**：P0-1（ISS-FA4）→ P0-2（赔率统计算子）→ M4/M5 收尾
（放开 mysql 图驱动 + standalone v4）→ P0-3（语义利弊）→ P1 数据源扩容
评审。

## 六、待主控确认

1. 人性贪婪恐惧分析是否纳入 PRD-12 范围（新增第 12 agent 或并入 synthesis）？
2. 庄家 pnl/价位聚集数据是否由宿主侧入库（touzhu 文件 → MySQL 表）？
3. P0-1 ISS-FA4 是否即刻立项（与 M5 收尾的先后顺序）？
