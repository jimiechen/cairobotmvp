# -*- coding: utf-8 -*-
"""MySqlAnalysisDataSource 盘口与推演类查询单元测试（Q3/D9/D12）。

全部使用 Fake 连接（conftest.patch_mysql_connect），禁止真实连库；
字段口径以 workbench 真实库为准（R-04 需求文档列名漂移已校准）：
- 三盘赔率按公司归并初盘（is_initial=1）与即时盘（最新快照）
- D9 变盘轨迹为降级实现（单一公司快照序列推演变盘方向）
- D12 热度漂移 = 首尾快照 cold_hot 差（side 为队名，需 JOIN `match` 匹配主客）
"""
from datetime import date, datetime

import pytest

from fa_orchestrator.data.mysql_analysis_source import MySqlAnalysisDataSource
from fa_orchestrator.data.mysql_data_source import MySqlDataSource
from fa_orchestrator.data.sql_contracts import (
    SQL_D9_ASIAN_TRACKS,
    SQL_D9_OVERUNDER_TRACKS,
    SQL_D12_HEAT,
    SQL_Q3_ASIAN,
    SQL_Q3_EUROPE,
    SQL_Q3_KICKOFF,
    SQL_Q3_OVERUNDER,
)


def _make_source(match_ids=None, **overrides):
    params = {
        "host": "db-host", "port": 3306, "user": "tester",
        "password": "secret", "database": "workbench",
    }
    params.update(overrides)
    return MySqlAnalysisDataSource(match_ids=match_ids, **params)


KICKOFF_ROWS = [{"kickoff_time": datetime(2026, 9, 29, 18, 0, 0)}]

# 同一公司（澳门）初盘 + 两次即时盘快照
EUROPE_ROWS = [
    {"bk": "澳门", "win": 2.1, "draw": 3.25, "lose": 3.4, "payout": 0.9,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
    {"bk": "澳门", "win": 2.02, "draw": 3.3, "lose": 3.55, "payout": 0.905,
     "is_initial": 0, "snapshot_time": datetime(2026, 9, 29, 20, 10, 0)},
    {"bk": "立博", "win": 2.15, "draw": 3.2, "lose": 3.45, "payout": 0.918,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
]
ASIAN_ROWS = [
    {"bk": "澳门", "home_water": 0.9, "handicap": 0.0, "away_water": 0.95,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
    {"bk": "澳门", "home_water": 0.78, "handicap": 0.25, "away_water": 1.08,
     "is_initial": 0, "snapshot_time": datetime(2026, 9, 29, 20, 10, 0)},
]
OVERUNDER_ROWS = [
    {"bk": "澳门", "over_odds": 0.92, "line": 2.5, "under_odds": 0.94,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
]


class TestResolve:
    """fid → match.id 解析映射：无映射不臆测（对齐 Q2b 口径）。"""

    def test_无映射时分析查询返回None或空(self):
        source = _make_source()
        assert source.query_match_odds("1398605") is None
        assert source.query_vote("1398605") is None
        assert source.query_fundamentals("1398605") == {}
        assert source.query_history_tracks("1398605") is None
        assert source.query_heat_drift("1398605") is None

    def test_映射后查询参数为match_id(self, patch_mysql_connect):
        conn = patch_mysql_connect([KICKOFF_ROWS, [], [], []])
        _make_source({"1398605": 2635}).query_match_odds("1398605")
        sql, params = conn.cursor_obj.executed[0]
        assert params == (2635,)

    def test_继承基础数据源配置(self):
        assert isinstance(_make_source(), MySqlDataSource)

    def test_D10赛果按fid直连无需映射(self, patch_mysql_connect):
        conn = patch_mysql_connect([[{
            "site_match_id": "1398605", "match_no": "周一003", "home_team": "亚美尼亚",
            "away_team": "黑山", "league": "欧国联", "match_time": "2026-09-29 00:00:00",
            "home_score": 2, "away_score": 3, "half_score": "1:1",
            "result_flag": "0", "play_type": "jingcai",
        }]])
        result = _make_source().query_match_result("1398605")
        sql, params = conn.cursor_obj.executed[0]
        assert params == ("1398605",)
        assert result["home_score"] == 2


class TestQ3Odds:
    def test_三盘初即盘按公司归并(self, patch_mysql_connect):
        patch_mysql_connect([KICKOFF_ROWS, EUROPE_ROWS, ASIAN_ROWS, OVERUNDER_ROWS])
        data = _make_source({"1398605": 2635}).query_match_odds("1398605")
        assert data["kickoff_time"] == "09-29 18:00"
        euro = {row["bk"]: row for row in data["euro"]}
        assert euro["澳门"]["initial"] == {"win": 2.1, "draw": 3.25, "lose": 3.4}
        assert euro["澳门"]["current"] == {"win": 2.02, "draw": 3.3, "lose": 3.55}
        assert euro["澳门"]["payout"] == 0.905
        assert euro["立博"]["current"] is None

    def test_亚盘与大小盘口径(self, patch_mysql_connect):
        patch_mysql_connect([KICKOFF_ROWS, [], ASIAN_ROWS, OVERUNDER_ROWS])
        data = _make_source({"1398605": 2635}).query_match_odds("1398605")
        asian = data["asian"][0]
        assert asian["initial"] == {"home_water": 0.9, "handicap": 0.0, "away_water": 0.95}
        assert asian["current"]["handicap"] == 0.25
        assert "is_final_snapshot" in asian
        ou = data["overunder"][0]
        assert ou["initial"]["line"] == 2.5
        assert "is_final_snapshot" not in ou   # 对齐 fixture：大小盘无终盘标志

    def test_无数据行时结构完整(self, patch_mysql_connect):
        patch_mysql_connect([[{"kickoff_time": None}], [], [], []])
        data = _make_source({"1398605": 2635}).query_match_odds("1398605")
        assert data == {"kickoff_time": None, "euro": [], "asian": [], "overunder": []}

    def test_执行契约SQL且三盘JOIN解名(self, patch_mysql_connect):
        conn = patch_mysql_connect([KICKOFF_ROWS, [], [], []])
        _make_source({"1398605": 2635}).query_match_odds("1398605")
        sqls = [sql for sql, _ in conn.cursor_obj.executed]
        assert SQL_Q3_KICKOFF.strip() in sqls
        assert SQL_Q3_EUROPE.strip() in sqls
        assert SQL_Q3_ASIAN.strip() in sqls
        assert SQL_Q3_OVERUNDER.strip() in sqls


HEAT_ROWS = [
    {"side": "亚美尼亚", "cold_hot": 9, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "黑山", "cold_hot": -30, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "平局", "cold_hot": -93, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "亚美尼亚", "cold_hot": -12, "snapshot_time": datetime(2026, 9, 29, 15, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "黑山", "cold_hot": 41, "snapshot_time": datetime(2026, 9, 29, 15, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
]


class TestD12HeatDrift:
    def test_漂移为首尾快照cold_hot差(self, patch_mysql_connect):
        patch_mysql_connect([HEAT_ROWS])
        drift = _make_source({"1398605": 2635}).query_heat_drift("1398605")
        assert drift == {"home_pp": -21, "away_pp": 71}   # 平局行不计

    def test_单快照漂移为0(self, patch_mysql_connect):
        patch_mysql_connect([[HEAT_ROWS[0], HEAT_ROWS[1]]])
        drift = _make_source({"1398605": 2635}).query_heat_drift("1398605")
        assert drift == {"home_pp": 0.0, "away_pp": 0.0}

    def test_队名匹配不上时该侧为None(self, patch_mysql_connect):
        rows = [dict(HEAT_ROWS[0], side="未知队"), dict(HEAT_ROWS[1], side="未知队2")]
        patch_mysql_connect([rows])
        drift = _make_source({"1398605": 2635}).query_heat_drift("1398605")
        assert drift == {"home_pp": None, "away_pp": None}

    def test_热度未计算快照行按缺口跳过(self, patch_mysql_connect):
        # 真实库存在 cold_hot=None 快照行（e2e 2026-10-08 发现）：
        # 热度未计算的快照不计入漂移，按有效快照首尾差计算
        rows = [
            dict(HEAT_ROWS[0], cold_hot=None),  # 亚美尼亚快照1：热度未计算
            dict(HEAT_ROWS[1], cold_hot=None),  # 黑山快照1：热度未计算
            HEAT_ROWS[3],                       # 亚美尼亚快照2：唯一有效 home 值
        ]
        patch_mysql_connect([rows])
        drift = _make_source({"1398605": 2635}).query_heat_drift("1398605")
        assert drift == {"home_pp": 0.0, "away_pp": None}

    def test_执行契约SQL且JOIN主客队名(self, patch_mysql_connect):
        conn = patch_mysql_connect([[]])
        _make_source({"1398605": 2635}).query_heat_drift("1398605")
        sql, params = conn.cursor_obj.executed[0]
        assert sql == SQL_D12_HEAT.strip()
        assert params == (2635,)


class TestD9Tracks:
    """变盘轨迹（降级口径：单一公司快照序列推演变盘方向）。"""

    def test_亚盘盘口升降记轨迹(self, patch_mysql_connect):
        rows = [
            {"handicap": 0.0, "is_initial": 1},
            {"handicap": 0.25, "is_initial": 0},
            {"handicap": 0.0, "is_initial": 0},
            {"handicap": -0.25, "is_initial": 0},
        ]
        patch_mysql_connect([rows, []])
        tracks = _make_source({"1398605": 2635}).query_history_tracks("1398605")
        assert tracks == [{"kind": "asian", "change": 1},
                          {"kind": "asian", "change": -1},
                          {"kind": "asian", "change": -1}]

    def test_大小盘球线变化记轨迹(self, patch_mysql_connect):
        rows = [{"line": 2.5, "is_initial": 1}, {"line": 3.0, "is_initial": 0}]
        patch_mysql_connect([[], rows])
        tracks = _make_source({"1398605": 2635}).query_history_tracks("1398605")
        assert {"kind": "overunder", "change": 1} in tracks

    def test_盘口无变化时轨迹为空(self, patch_mysql_connect):
        rows = [{"handicap": 0.25, "is_initial": 1}, {"handicap": 0.25, "is_initial": 0}]
        patch_mysql_connect([rows, []])
        assert _make_source({"1398605": 2635}).query_history_tracks("1398605") == []

    def test_执行契约SQL且公司名参数化(self, patch_mysql_connect):
        conn = patch_mysql_connect([[], []])
        _make_source({"1398605": 2635}).query_history_tracks("1398605")
        sqls = [sql for sql, _ in conn.cursor_obj.executed]
        assert SQL_D9_ASIAN_TRACKS.strip() in sqls
        assert SQL_D9_OVERUNDER_TRACKS.strip() in sqls
        # 轨迹公司名不得写死进 SQL，必须参数化传入
        assert all("澳门" not in sql for sql in sqls)


class TestAnalysisContracts:
    def test_契约暴露供审计(self):
        contracts = _make_source().get_sql_contracts()
        for key in ("Q3_kickoff", "Q3_europe", "Q3_asian", "Q3_overunder",
                    "Q5_vote", "Q5_bf", "Q9_intel", "Q10_h2h", "Q11_recent",
                    "Q12_schedule", "D9_asian_tracks", "D9_overunder_tracks",
                    "D12_heat", "D10_result"):
            assert key in contracts, f"缺少契约 {key}"
