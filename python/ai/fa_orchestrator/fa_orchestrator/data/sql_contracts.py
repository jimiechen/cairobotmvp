# -*- coding: utf-8 -*-
"""SQL 契约常量（PRD-12 F-06）：取数 SQL 与需求文档 §四 对齐 + 分析类契约按真实库归档。

职责边界：
- 负责：固化局域网取数 SQL 契约文本与六模块表名映射，供 DataAgent 拼装查询；
  分析类契约（Q3/Q5/Q9~Q12/D9/D10/D12）以 workbench 真实库结构为准归档
  （方案 A 批次 1，R-04 需求文档列名漂移已校准）。
- 不负责：数据库连接与执行（data_source.py / mysql_data_source.py /
  mysql_analysis_source.py）、结果解析与完整性判定（integrity.py / id_resolver.py）。

约束来源：《需求文档 - 局域网数据获取》§四 + 2026-10-08 workbench 只读探测；
`match` 是 MySQL 保留字，SQL 中必须反引号；全部参数化 %s 占位，禁止拼接，
禁止字面量日期/投票池名/轨迹公司名写死进契约。
"""

# 六模块完整性核查映射（PRD-12 F-08）：模块名 → Q2b 同构替换表名
INTEGRITY_MODULES = {
    "ouzhi": "odds_europe",
    "yazhi": "odds_asian",
    "daxiao": "odds_overunder",
    "vote": "vote_snapshot",
    "stats": "team_season_stats",
    "intel": "intelligence_text",
}

# 竞彩日窗口起点（PRD-12 §12.3：[D 12:00, D+1 12:00) 左闭右开）。
# Q1 契约 SQL 内 '12:00' 字面量与本常量同源；代码侧候选窗口（mysql 通道
# kickoff_time 参数化 / fixture 通道 match_time 过滤）统一引用本常量
DAY_WINDOW_START_TIME = "12:00"

# Q1：竞彩日比赛清单（500com 活跃映射，按编号排序）
# 窗口口径（PRD-12 §12.3）：竞彩日 D = 开球窗口 [D 12:00, D+1 12:00) 左闭右开，
# match_date 归属开球日——当日 12:00 起（含）或次日 12:00 前（不含）开球的
# 场次归属竞彩日 D。两个 %s 传同一竞彩日；match_time 为 'HH:MM' 零填充字符串
# （2026-10-08 workbench 探测口径），字典序比较与时间序一致。
SQL_Q1_MATCH_LIST = """
SELECT m.internal_id, m.match_no, m.home_team, m.away_team, m.match_date,
       s.site_match_id AS fid
FROM matches m
JOIN site_id_mappings s ON s.internal_id = m.internal_id
  AND s.site_code = '500com' AND s.is_active = 1
WHERE (m.match_date = %s AND m.match_time >= '12:00')
   OR (m.match_date = DATE_ADD(%s, INTERVAL 1 DAY) AND m.match_time < '12:00')
ORDER BY m.match_no, m.internal_id;
"""

# 四路径解析的 `match` 表候选查询（需求文档 §2.3：cc + match_date 窗口 + play_type='jingcai'）
# `match` 为 MySQL 保留字，必须反引号；窗口边界（±2 天）由调用方传入起止日期
SQL_JC_MATCH_CANDIDATES = """
SELECT id, cc, match_date, play_type, home_team, away_team
FROM `match`
WHERE play_type = 'jingcai'
  AND (cc = %s OR home_team = %s OR away_team = %s);
"""

_Q2B_SQL_TEMPLATE = (
    "SELECT match_id, COUNT(*) c FROM {table} "
    "WHERE match_id IN {placeholders} GROUP BY match_id"
)


def build_q2b_sql(table: str, id_count: int) -> str:
    """构建 Q2b 模块行数核查 SQL。

    输入：table（六模块白名单表名）、id_count（批量占位符数量）。
    输出：参数化 SQL 文本。
    失败行为：表名不在白名单时抛 ValueError，防止任意表名拼入。
    """
    allowed_tables = set(INTEGRITY_MODULES.values())
    if table not in allowed_tables:
        raise ValueError(f"表名不在六模块白名单内: {table}")
    placeholders = "(" + ",".join(["%s"] * id_count) + ")"
    return _Q2B_SQL_TEMPLATE.format(table=table, placeholders=placeholders)


# ---- 分析类契约（Q3/Q5/Q9~Q12/D9/D10/D12，2026-10-08 按 workbench 真实库归档）----
# 与需求文档 §四 列名漂移点（cc/home_team → cc_id/home/away、pool_code 取值
# HAD/HHAD）已按实际库校准；数据域约束：模块表 match_id 域 = `match`.id，
# D10 赛果例外（crawler_match_result.site_match_id == 500com fid 直连）。

POOL_HAD = "HAD"          # 胜平负投票池（vote_snapshot.pool_code 取值）
POOL_HHAD = "HHAD"        # 让球胜平负投票池
# D9 变盘轨迹参考公司集（多公司口径，参数化传入）：四家主流定价基准，
# 库内均有 ≥2 快照（2026-10-12 探测：2948 皇冠 42/Bet365 25/澳门 3 轮）
TRACK_BOOKMAKERS = ("澳门", "皇冠", "Bet365", "Pinnacle")

SQL_Q3_KICKOFF = """
SELECT kickoff_time FROM `match` WHERE id = %s;
"""

SQL_Q3_EUROPE = """
SELECT b.name AS bk, e.win, e.draw, e.lose, e.payout,
       e.is_initial, e.snapshot_time
FROM odds_europe e
JOIN bookmaker b ON b.id = e.bookmaker_id
WHERE e.match_id = %s
ORDER BY b.sort_order, b.id, e.snapshot_time, e.id;
"""

SQL_Q3_ASIAN = """
SELECT b.name AS bk, a.home_water, a.handicap, a.away_water,
       a.is_initial, a.snapshot_time
FROM odds_asian a
JOIN bookmaker b ON b.id = a.bookmaker_id
WHERE a.match_id = %s
ORDER BY b.sort_order, b.id, a.snapshot_time, a.id;
"""

SQL_Q3_OVERUNDER = """
SELECT b.name AS bk, o.over_odds, o.line, o.under_odds,
       o.is_initial, o.snapshot_time
FROM odds_overunder o
JOIN bookmaker b ON b.id = o.bookmaker_id
WHERE o.match_id = %s
ORDER BY b.sort_order, b.id, o.snapshot_time, o.id;
"""

# 投票池：HAD 出支持率/SP，HHAD 出让球块；每池取最新快照（排序后取末行）
SQL_Q5_VOTE = """
SELECT pool_code, goal_line, h_odds, d_odds, a_odds,
       h_support_rate, d_support_rate, a_support_rate, snapshot_time
FROM vote_snapshot
WHERE match_id = %s AND pool_code IN (%s, %s)
ORDER BY pool_code, snapshot_time, id;
"""

# 必发交易：全类型读取——standard 行出方向与占比趋势（多时点），large 行出
# 大额成交量（2026-10-12 库内探查：large 与 standard 同频快照、占比列恒空，
# 不用于方向判定）；trade_type 列保留供解析层分组
SQL_Q5_BF = """
SELECT trade_type, total_volume, home_pct, draw_pct, away_pct, snapshot_time
FROM betfair_trading
WHERE match_id = %s
ORDER BY snapshot_time, id;
"""

# 立项四扩展：庄家 pnl 指数（bookmaker_pnl 列由宿主执行
# migrations/betfair_pnl_expand/up.sql 后存在；列未建时本查询报
# OperationalError，数据源层捕获降级 None，不破坏原 Q5_bf 口径）。
# pnl 属 bf_trading（标准盘）属性，仅取 standard 行。
SQL_Q5_BF_PNL = """
SELECT trade_type, bookmaker_pnl, snapshot_time
FROM betfair_trading
WHERE match_id = %s AND trade_type = 'standard'
ORDER BY snapshot_time, id;
"""

# 立项四扩展：必发价位聚集（betfair_price_ladder 表由宿主建表后生效；
# 表未建时降级 None）。取最新快照全档，解析层出聚集度摘要。
SQL_Q5_LADDER = """
SELECT price, back_volume, lay_volume
FROM betfair_price_ladder
WHERE match_id = %s
  AND snapshot_time = (SELECT MAX(snapshot_time) FROM betfair_price_ladder
                       WHERE match_id = %s)
ORDER BY price;
"""

# 立项四扩展（宿主答复 2026-10-14）：500com 投注倾向庄家盈亏——
# heat_analysis.pnl 为庄家该方向盈亏（正=庄家赢，side 三方向：主队名/
# 平局/客队名），作为庄家 pnl 指数的现实数据源先行接入（必发 pnl 回填后
# 两源并存）；side 与 match 表主客队名匹配映射由数据源层完成。
# 表未建报 ProgrammingError(1146)，与 pnl/ladder 同口径降级 None。
SQL_Q5_HEAT_PNL = """
SELECT h.side, h.pnl, h.cold_hot, h.snapshot_time, m.home, m.away
FROM heat_analysis h
JOIN `match` m ON m.id = h.match_id
WHERE h.match_id = %s
ORDER BY h.snapshot_time, h.id;
"""

# 情报：仅取已填充行；party 由正文【主队】/【客队】/【真队名】标记解析，
# sentiment 由 parse 层关键词判读（立项二②）；JOIN match 取主客队名供真队名标记匹配
# （match 表主客列为 home/away，别名对齐 parse 层参数命名）
SQL_Q9_INTEL = """
SELECT i.source, i.title, i.content, i.status, i.published_at,
       m.home AS home_team, m.away AS away_team
FROM intelligence_text i
JOIN `match` m ON m.id = i.match_id
WHERE i.match_id = %s AND i.status = 'filled'
ORDER BY i.id;
"""

# 历史交锋：按日期倒序（近年优先）；result_for_current_home 为当前主队视角
SQL_Q10_H2H = """
SELECT h2h_date, competition, home_team, away_team,
       score, score_home, score_away, result_for_current_home
FROM h2h_record
WHERE match_id = %s
ORDER BY h2h_date DESC, id DESC;
"""

# 近期战绩：side 分组倒序（近期优先）；进失球按 is_home 视角解析
SQL_Q11_RECENT = """
SELECT side, form_date, competition, home_team, away_team,
       score, result, is_home
FROM team_recent_form
WHERE match_id = %s
ORDER BY side, form_date DESC, id DESC;
"""

# 赛程：本场 + 双方未来场次（子查询锚定本场日期与主客队名，LIMIT 防全表）
SQL_Q12_SCHEDULE = """
SELECT m2.id, m2.league, m2.match_date
FROM `match` m2
WHERE m2.play_type = 'jingcai'
  AND m2.match_date >= (SELECT m1.match_date FROM `match` m1 WHERE m1.id = %s)
  AND (m2.home = (SELECT m1.home FROM `match` m1 WHERE m1.id = %s)
       OR m2.away = (SELECT m1.home FROM `match` m1 WHERE m1.id = %s)
       OR m2.home = (SELECT m1.away FROM `match` m1 WHERE m1.id = %s)
       OR m2.away = (SELECT m1.away FROM `match` m1 WHERE m1.id = %s))
ORDER BY m2.match_date, m2.id
LIMIT 10;
"""

# 热度快照：side 为队名（含'平局'），JOIN `match` 取主客队名做匹配
SQL_D12_HEAT = """
SELECT h.side, h.cold_hot, h.snapshot_time, m.home, m.away
FROM heat_analysis h
JOIN `match` m ON m.id = h.match_id
WHERE h.match_id = %s
ORDER BY h.snapshot_time, h.id;
"""

# 变盘轨迹（D9）：四家参考公司亚盘/大小盘快照序列，公司名参数化传入；
# 输出 bk 列供按公司分组推演，公司内按快照时间升序（全局排序下相对序保持）
SQL_D9_ASIAN_TRACKS = """
SELECT b.name AS bk, a.handicap, a.is_initial
FROM odds_asian a
JOIN bookmaker b ON b.id = a.bookmaker_id
WHERE a.match_id = %s AND b.name IN (%s, %s, %s, %s)
ORDER BY b.name, a.snapshot_time, a.id;
"""

SQL_D9_OVERUNDER_TRACKS = """
SELECT b.name AS bk, o.line, o.is_initial
FROM odds_overunder o
JOIN bookmaker b ON b.id = o.bookmaker_id
WHERE o.match_id = %s AND b.name IN (%s, %s, %s, %s)
ORDER BY b.name, o.snapshot_time, o.id;
"""

# D10 赛果：site_match_id == 500com fid 直连，取最新更新行
SQL_D10_RESULT = """
SELECT site_match_id, match_no, home_team, away_team, league,
       match_time, home_score, away_score, half_score, result_flag, play_type
FROM crawler_match_result
WHERE site_match_id = %s
ORDER BY updated_at DESC, id DESC;
"""

# 分析契约注册表：键名供 get_sql_contracts 暴露与审计核对
ANALYSIS_CONTRACTS = {
    "Q3_kickoff": SQL_Q3_KICKOFF.strip(),
    "Q3_europe": SQL_Q3_EUROPE.strip(),
    "Q3_asian": SQL_Q3_ASIAN.strip(),
    "Q3_overunder": SQL_Q3_OVERUNDER.strip(),
    "Q5_vote": SQL_Q5_VOTE.strip(),
    "Q5_bf": SQL_Q5_BF.strip(),
    "Q5_bf_pnl": SQL_Q5_BF_PNL.strip(),
    "Q5_ladder": SQL_Q5_LADDER.strip(),
    "Q5_heat_pnl": SQL_Q5_HEAT_PNL.strip(),
    "Q9_intel": SQL_Q9_INTEL.strip(),
    "Q10_h2h": SQL_Q10_H2H.strip(),
    "Q11_recent": SQL_Q11_RECENT.strip(),
    "Q12_schedule": SQL_Q12_SCHEDULE.strip(),
    "D9_asian_tracks": SQL_D9_ASIAN_TRACKS.strip(),
    "D9_overunder_tracks": SQL_D9_OVERUNDER_TRACKS.strip(),
    "D12_heat": SQL_D12_HEAT.strip(),
    "D10_result": SQL_D10_RESULT.strip(),
}
