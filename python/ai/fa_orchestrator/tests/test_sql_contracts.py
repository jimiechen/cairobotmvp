# -*- coding: utf-8 -*-
"""SQL 契约测试（PRD-12 F-06）：Q1/Q2b SQL 与需求文档 §四 逐字对齐，六模块表名映射正确。

契约要点（需求文档硬性规则）：
- Q1 固定 site_code='500com' + is_active=1 + 参数化 %s
- Q2b 六模块：ouzhi/yazhi/daxiao/vote/stats/intel 同构替换表名
- `match` 是 MySQL 保留字，SQL 中必须反引号
"""
import pytest

from fa_orchestrator.sql_contracts import (
    INTEGRITY_MODULES,
    SQL_Q1_MATCH_LIST,
    SQL_JC_MATCH_CANDIDATES,
    build_q2b_sql,
)


class TestQ1MatchList:
    """Q1 比赛清单 SQL 契约。"""

    def test_使用500com活跃映射条件(self):
        assert "s.site_code = '500com'" in SQL_Q1_MATCH_LIST
        assert "s.is_active = 1" in SQL_Q1_MATCH_LIST

    def test_日期参数化防注入(self):
        assert "m.match_date = %s" in SQL_Q1_MATCH_LIST
        # 禁止任何字面量日期写死进契约 SQL
        assert "2026" not in SQL_Q1_MATCH_LIST

    def test_按场次编号排序(self):
        assert "ORDER BY m.match_no" in SQL_Q1_MATCH_LIST

    def test_返回fid别名列(self):
        assert "s.site_match_id AS fid" in SQL_Q1_MATCH_LIST


class TestQ2bIntegritySql:
    """Q2b 六模块完整性核查 SQL 契约。"""

    def test_六模块表名映射齐全(self):
        assert INTEGRITY_MODULES == {
            "ouzhi": "odds_europe",
            "yazhi": "odds_asian",
            "daxiao": "odds_overunder",
            "vote": "vote_snapshot",
            "stats": "team_season_stats",
            "intel": "intelligence_text",
        }

    def test_批量占位符按数量生成(self):
        sql = build_q2b_sql("odds_europe", 3)
        assert sql == (
            "SELECT match_id, COUNT(*) c FROM odds_europe "
            "WHERE match_id IN (%s,%s,%s) GROUP BY match_id"
        )

    def test_单场占位符(self):
        sql = build_q2b_sql("vote_snapshot", 1)
        assert "IN (%s)" in sql

    def test_表名必须来自白名单(self):
        with pytest.raises(ValueError):
            build_q2b_sql("user_passwords", 3)


class TestJcMatchCandidatesSql:
    """四路径解析的 `match` 表候选查询 SQL 契约（需求文档 §2.3）。"""

    def test_match保留字反引号(self):
        # `match` 是 MySQL 保留字，必须反引号
        assert "`match`" in SQL_JC_MATCH_CANDIDATES

    def test_竞彩玩法过滤(self):
        assert "play_type = 'jingcai'" in SQL_JC_MATCH_CANDIDATES

    def test_全参数化(self):
        assert SQL_JC_MATCH_CANDIDATES.count("%s") >= 3
