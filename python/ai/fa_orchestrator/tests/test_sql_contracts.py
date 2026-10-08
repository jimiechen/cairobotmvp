# -*- coding: utf-8 -*-
"""SQL 契约测试（PRD-12 F-06）：Q1/Q2b SQL 与需求文档 §四 逐字对齐，六模块表名映射正确。

契约要点（需求文档硬性规则）：
- Q1 固定 site_code='500com' + is_active=1 + 参数化 %s
- Q2b 六模块：ouzhi/yazhi/daxiao/vote/stats/intel 同构替换表名
- `match` 是 MySQL 保留字，SQL 中必须反引号
- 分析类契约（Q3/Q5/Q9~Q12/D9/D10/D12）以 workbench 真实库为准
  （R-04 需求文档列名漂移已校准），全参数化，禁止字面量日期/池名/公司名
"""
import pytest

from fa_orchestrator.data.sql_contracts import (
    INTEGRITY_MODULES,
    ANALYSIS_CONTRACTS,
    POOL_HAD,
    POOL_HHAD,
    SQL_D9_ASIAN_TRACKS,
    SQL_D10_RESULT,
    SQL_Q1_MATCH_LIST,
    SQL_Q3_ASIAN,
    SQL_Q3_KICKOFF,
    SQL_JC_MATCH_CANDIDATES,
    TRACK_BOOKMAKER_NAME,
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


ANALYSIS_SQL_KEYS = (
    "Q3_kickoff", "Q3_europe", "Q3_asian", "Q3_overunder",
    "Q5_vote", "Q5_bf", "Q9_intel", "Q10_h2h", "Q11_recent",
    "Q12_schedule", "D9_asian_tracks", "D9_overunder_tracks",
    "D12_heat", "D10_result",
)


class TestAnalysisContractSql:
    """分析类契约（Q3/Q5/Q9~Q12/D9/D10/D12）：以真实库结构为准，全参数化。"""

    def test_十四个分析契约齐全且非空(self):
        for key in ANALYSIS_SQL_KEYS:
            assert key in ANALYSIS_CONTRACTS, f"缺少契约 {key}"
            assert ANALYSIS_CONTRACTS[key].strip(), f"契约 {key} 为空"

    def test_全部参数化禁止字面量日期(self):
        for key, sql in ANALYSIS_CONTRACTS.items():
            assert "%s" in sql, f"契约 {key} 未参数化"
            assert "2026" not in sql, f"契约 {key} 含字面量日期"

    def test_match保留字反引号(self):
        assert "`match`" in ANALYSIS_CONTRACTS["Q3_kickoff"]
        assert "`match`" in ANALYSIS_CONTRACTS["Q12_schedule"]

    def test_三盘SQL按公司解名并按快照排序(self):
        for key in ("Q3_europe", "Q3_asian", "Q3_overunder"):
            sql = ANALYSIS_CONTRACTS[key]
            assert "JOIN bookmaker" in sql, f"{key} 缺少公司解名 JOIN"
            assert "is_initial" in sql, f"{key} 缺少初盘标志"
            assert "snapshot_time" in sql, f"{key} 缺少快照排序"

    def test_投票池过滤参数化(self):
        sql = ANALYSIS_CONTRACTS["Q5_vote"]
        assert "IN (%s, %s)" in sql
        assert (POOL_HAD, POOL_HHAD) == ("HAD", "HHAD")

    def test_变盘轨迹公司名参数化(self):
        assert ANALYSIS_CONTRACTS["D9_asian_tracks"].count("%s") == 2
        assert "澳门" not in ANALYSIS_CONTRACTS["D9_asian_tracks"]
        assert TRACK_BOOKMAKER_NAME == "澳门"

    def test_D10按site_match_id直连(self):
        sql = ANALYSIS_CONTRACTS["D10_result"]
        assert "site_match_id = %s" in sql
        assert "`match`" not in SQL_D9_ASIAN_TRACKS  # 盘口轨迹不依赖 match 表

    def test_D12热度JOIN主客队名(self):
        sql = ANALYSIS_CONTRACTS["D12_heat"]
        assert "JOIN `match`" in sql
