# -*- coding: utf-8 -*-
"""SQL 契约常量（PRD-12 F-06）：Q1/Q2b 等取数 SQL 与需求文档 §四 逐字对齐。

职责边界：
- 负责：固化局域网取数 SQL 契约文本与六模块表名映射，供 DataAgent 拼装查询。
- 不负责：数据库连接与执行（data_source.py）、结果解析（integrity.py / id_resolver.py）。

约束来源：《需求文档 - 局域网数据获取》§四；
`match` 是 MySQL 保留字，SQL 中必须反引号；全部参数化 %s 占位，禁止拼接。
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

# Q1：当日比赛清单（500com 活跃映射，按编号排序）
# 窗口口径（12:00 → 次日 12:00 左闭右开）与跨日归属由调用方保证
SQL_Q1_MATCH_LIST = """
SELECT m.internal_id, m.match_no, m.home_team, m.away_team, m.match_date,
       s.site_match_id AS fid
FROM matches m
JOIN site_id_mappings s ON s.internal_id = m.internal_id
  AND s.site_code = '500com' AND s.is_active = 1
WHERE m.match_date = %s
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
