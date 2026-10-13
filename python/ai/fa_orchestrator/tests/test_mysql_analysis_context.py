# -*- coding: utf-8 -*-
"""MySqlAnalysisDataSource 基本面与赛果类查询单元测试（Q5/Q9/Q10/Q11/Q12/D10）。

全部使用 Fake 连接（conftest.patch_mysql_connect），禁止真实连库；
字段口径以 workbench 真实库为准（R-04 校准）：
- vote_snapshot 池代码为 HAD（胜平负）/ HHAD（让球胜平负），取最新快照
- intelligence_text 以【主队】/【客队】分段解析 party；sentiment 库内无维度，不臆造
- team_recent_form 的 scored/conceded 按 is_home 视角解析，友谊赛按赛程名判定
- D10 赛果 crawler_match_result.site_match_id == fid 直连
"""
from datetime import date

import pytest

from fa_orchestrator.data.mysql_analysis_source import MySqlAnalysisDataSource
from fa_orchestrator.data.sql_contracts import SQL_D10_RESULT


def _make_source(match_ids=None, **overrides):
    params = {
        "host": "db-host", "port": 3306, "user": "tester",
        "password": "secret", "database": "workbench",
    }
    params.update(overrides)
    return MySqlAnalysisDataSource(match_ids=match_ids, **params)


HAD_ROWS = [
    {"pool_code": "HAD", "goal_line": None, "h_odds": 2.2, "d_odds": 3.16,
     "a_odds": 2.79, "h_support_rate": 31.0, "d_support_rate": 31.0,
     "a_support_rate": 38.0, "snapshot_time": "2026-09-28 12:00:00"},
    {"pool_code": "HAD", "goal_line": None, "h_odds": 2.1, "d_odds": 3.2,
     "a_odds": 2.9, "h_support_rate": 33.0, "d_support_rate": 30.0,
     "a_support_rate": 37.0, "snapshot_time": "2026-09-29 20:10:00"},
]
HHAD_ROWS = [
    {"pool_code": "HHAD", "goal_line": -1.0, "h_odds": 4.6, "d_odds": 3.75,
     "a_odds": 1.55, "h_support_rate": 16.0, "d_support_rate": 22.0,
     "a_support_rate": 62.0, "snapshot_time": "2026-09-29 20:10:00"},
]
BF_ROWS = [
    # 模拟 SQL 全类型返回（不再过滤 trade_type）：standard 多时点出方向/趋势，
    # large 行占比列库内恒空、仅大额成交量有值（2026-10-12 真库探查口径）
    {"trade_type": "standard", "total_volume": 10266.0, "home_pct": 61.36,
     "draw_pct": 1.44, "away_pct": 37.2, "snapshot_time": "2026-09-28 12:00:00"},
    {"trade_type": "large", "total_volume": 3000.0, "home_pct": None,
     "draw_pct": None, "away_pct": None, "snapshot_time": "2026-09-28 12:00:00"},
    {"trade_type": "standard", "total_volume": 10266.0, "home_pct": 55.0,
     "draw_pct": 5.0, "away_pct": 40.0, "snapshot_time": "2026-09-29 21:47:08"},
    {"trade_type": "large", "total_volume": 9000.0, "home_pct": None,
     "draw_pct": None, "away_pct": None, "snapshot_time": "2026-09-29 21:47:08"},
]


class TestQ5Vote:
    def test_HAD池出支持率与SP取最新快照(self, patch_mysql_connect):
        patch_mysql_connect([HAD_ROWS + HHAD_ROWS, BF_ROWS])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["support"] == {"win": 33.0, "draw": 30.0, "lose": 37.0}
        assert vote["sp"] == {"win": 2.1, "draw": 3.2, "lose": 2.9}

    def test_HHAD池出让球块(self, patch_mysql_connect):
        patch_mysql_connect([HAD_ROWS + HHAD_ROWS, BF_ROWS])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["hhand"]["line"] == -1.0
        assert vote["hhand"]["odds"] == {"win": 4.6, "draw": 3.75, "lose": 1.55}
        assert vote["hhand"]["support"] == {"win": 16.0, "draw": 22.0, "lose": 62.0}

    def test_必发direction按最新占比且trend按方向占比首尾差(self, patch_mysql_connect):
        # 2026-10-12 库内探查：standard 多时点占比可判趋势（2949 home 68.50→79.18）
        patch_mysql_connect([HAD_ROWS, BF_ROWS])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["bf"]["total_volume"] == 10266.0
        assert vote["bf"]["direction"] == "home"      # 最新占比最大者
        assert vote["bf"]["trend"] == "falling"       # home 占比 61.36→55.0 回落
        assert vote["bf"]["large_volume"] == 9000.0   # large 大额成交量（占比列恒空）

    def test_必发trend按方向占比走高rising(self, patch_mysql_connect):
        rows = [
            dict(BF_ROWS[0], home_pct=40.0, draw_pct=10.0, away_pct=50.0),
            dict(BF_ROWS[2], home_pct=55.0, draw_pct=10.0, away_pct=35.0),
        ]
        patch_mysql_connect([HAD_ROWS, rows])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["bf"]["direction"] == "home"
        assert vote["bf"]["trend"] == "rising"

    def test_必发单时点trend为None(self, patch_mysql_connect):
        patch_mysql_connect([HAD_ROWS, [BF_ROWS[0]]])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["bf"]["trend"] is None   # 单时点快照，趋势不可判（缺口标注）

    def test_必发首尾等值trend为None(self, patch_mysql_connect):
        # 真实库 2948 两时点占比完全相同：无趋势变化，不臆造方向
        rows = [BF_ROWS[0], dict(BF_ROWS[2], home_pct=61.36,
                                 draw_pct=1.44, away_pct=37.2)]
        patch_mysql_connect([HAD_ROWS, rows])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["bf"]["trend"] is None

    def test_必发无large行时large_volume为None(self, patch_mysql_connect):
        patch_mysql_connect([HAD_ROWS, [BF_ROWS[0], BF_ROWS[2]]])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["bf"]["large_volume"] is None

    def test_缺HHAD池时hhand为None(self, patch_mysql_connect):
        patch_mysql_connect([HAD_ROWS, []])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["hhand"] is None

    def test_无投票行时support与sp为None(self, patch_mysql_connect):
        patch_mysql_connect([[], []])
        vote = _make_source({"1398605": 2635}).query_vote("1398605")
        assert vote["support"] is None and vote["sp"] is None


INTEL_ROWS = [
    {"source": "qianzhan500", "status": "filled", "title": "500com前瞻",
     "content": "【主队】\n哥伦布机员全队身价高于对手。", "published_at": None},
    {"source": "zqcf", "status": "empty", "title": "足球财富",
     "content": None, "published_at": None},
    {"source": "qianzhan500", "status": "filled", "title": "500com前瞻",
     "content": "【客队】\n圣路易斯城近6场4胜2平。", "published_at": None},
    {"source": "qianzhan500", "status": "filled", "title": "500com前瞻",
     "content": "综合分析：本场看好主队。", "published_at": None},
]


class TestQ9Intel:
    def test_按主客标记分段解析party(self, patch_mysql_connect):
        patch_mysql_connect([INTEL_ROWS])
        result = _make_source({"1398605": 2635}).query_fundamentals("1398605")
        items = result["intelligence"]["items"]
        assert items[0] == {"party": "home", "text": "哥伦布机员全队身价高于对手。"}
        assert items[1] == {"party": "away", "text": "圣路易斯城近6场4胜2平。"}
        # 无主客标记的段落 party 归 None；empty 状态行跳过（第 2 行）
        assert items[2] == {"party": None, "text": "综合分析：本场看好主队。"}
        assert len(items) == 3

    def test_无情报行时items为空(self, patch_mysql_connect):
        patch_mysql_connect([[]])
        result = _make_source({"1398605": 2635}).query_fundamentals("1398605")
        assert result["intelligence"]["items"] == []


H2H_ROWS = [
    {"h2h_date": date(2026, 6, 1), "competition": "欧国联", "home_team": "亚美尼亚",
     "away_team": "黑山", "score": "2:1", "score_home": 2, "score_away": 1,
     "result_for_current_home": "W"},
    {"h2h_date": date(2025, 3, 2), "competition": "友谊赛", "home_team": "黑山",
     "away_team": "亚美尼亚", "score": "0:0", "score_home": 0, "score_away": 0,
     "result_for_current_home": "D"},
]


class TestQ10H2h:
    def test_按日期倒序输出year_result_score(self, patch_mysql_connect):
        patch_mysql_connect([[], H2H_ROWS])
        result = _make_source({"1398605": 2635}).query_fundamentals("1398605")
        records = result["h2h"]["records"]
        assert records[0] == {"year": 2026, "result": "W", "score": "2-1"}
        assert records[1]["score"] == "0-0"

    def test_decay_flags恒空(self, patch_mysql_connect):
        patch_mysql_connect([[], H2H_ROWS])
        result = _make_source({"1398605": 2635}).query_fundamentals("1398605")
        assert result["h2h"]["decay_flags"] == []   # 库内无衰减标记维度（缺口标注）


# 预置顺序对齐契约排序语义（side 字母序 + form_date 倒序）
FORM_ROWS = [
    {"side": "away", "form_date": date(2026, 8, 30), "competition": "欧国联",
     "home_team": "威尔士", "away_team": "黑山", "score": "2:0",
     "result": "W", "is_home": 0},
    {"side": "home", "form_date": date(2026, 8, 30), "competition": "欧国联",
     "home_team": "亚美尼亚", "away_team": "直布罗陀", "score": "3:1",
     "result": "W", "is_home": 1},
    {"side": "home", "form_date": date(2026, 7, 24), "competition": "球会友谊",
     "home_team": "亚美尼亚", "away_team": "巴黎FC", "score": "2:2",
     "result": "D", "is_home": 1},
]


class TestQ11Recent:
    def test_按side分组且友谊赛判定(self, patch_mysql_connect):
        patch_mysql_connect([[], [], FORM_ROWS])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        home_matches = recent["home"]["matches"]
        assert home_matches[0]["is_friendly"] == 0
        assert home_matches[1]["is_friendly"] == 1

    def test_进失球按is_home视角解析(self, patch_mysql_connect):
        patch_mysql_connect([[], [], FORM_ROWS])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        # side=home 且 is_home=1：score 3:1 → 进 3 失 1
        assert recent["home"]["matches"][0] == {"is_friendly": 0, "scored": 3, "conceded": 1}
        # side=away 且 is_home=0：score 2:0 → 进 0 失 2
        assert recent["away"]["matches"][0] == {"is_friendly": 0, "scored": 0, "conceded": 2}

    def test_无近期行时matches为空列表(self, patch_mysql_connect):
        patch_mysql_connect([[], [], []])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        assert recent == {"home": {"matches": []}, "away": {"matches": []}}

    def test_比分异常行按缺口跳过不计入(self, patch_mysql_connect):
        # 真实库存在比分缺失/异常行（e2e 2026-10-08 发现）：进失球无法统计，
        # 该行不输出（违反 {scored, conceded} 数值契约），由 agent 场均统计跳过
        bad_rows = [dict(FORM_ROWS[0], score="vs")]
        patch_mysql_connect([[], [], bad_rows])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        assert recent["away"]["matches"] == []

    def test_home行比分异常同样跳过不误伤away(self, patch_mysql_connect):
        # BUG-FA-101 对称性：缺口剔除只按行判定，不得影响另一侧
        bad_rows = [dict(FORM_ROWS[1], score="vs"), FORM_ROWS[0]]
        patch_mysql_connect([[], [], bad_rows])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        assert recent["home"]["matches"] == []
        assert len(recent["away"]["matches"]) == 1

    def test_异常行与有效行混合时仅剔除异常行(self, patch_mysql_connect):
        # 剔除按行粒度：同侧有效行必须保留，不得整侧丢弃
        bad_rows = [dict(FORM_ROWS[0], score="vs"), FORM_ROWS[0]]
        patch_mysql_connect([[], [], bad_rows])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        assert recent["away"]["matches"] == [{"is_friendly": 0, "scored": 0, "conceded": 2}]

    def test_score为NULL或空串的行按缺口跳过(self, patch_mysql_connect):
        # 真实库 score 列可能为 NULL：split_score(None/"") → (None, None)，同缺口口径
        bad_rows = [dict(FORM_ROWS[0], score=None), dict(FORM_ROWS[0], score="")]
        patch_mysql_connect([[], [], bad_rows])
        recent = _make_source({"1398605": 2635}).query_fundamentals("1398605")["recent_form"]
        assert recent["away"]["matches"] == []


SCHEDULE_ROWS = [
    {"id": 2635, "league": "欧国联", "match_date": date(2026, 9, 29)},
    {"id": 2680, "league": "欧国联", "match_date": date(2026, 10, 2)},
    {"id": 2701, "league": "球会友谊", "match_date": date(2026, 10, 6)},
]


class TestQ12Schedule:
    def test_本场competition与未来赛程天数(self, patch_mysql_connect):
        patch_mysql_connect([[], [], [], SCHEDULE_ROWS])
        schedule = _make_source({"1398605": 2635}).query_fundamentals("1398605")["schedule"]
        assert schedule["competition"] == "欧国联"
        assert schedule["future"] == [{"days": 3, "competition": "欧国联"},
                                      {"days": 7, "competition": "球会友谊"}]

    def test_level_gap恒False(self, patch_mysql_connect):
        patch_mysql_connect([[], [], [], SCHEDULE_ROWS])
        schedule = _make_source({"1398605": 2635}).query_fundamentals("1398605")["schedule"]
        assert schedule["level_gap"] is False   # 级差判定缺库内数据（缺口标注）

    def test_本场自身不计入future(self, patch_mysql_connect):
        patch_mysql_connect([[], [], [], SCHEDULE_ROWS])
        schedule = _make_source({"1398605": 2635}).query_fundamentals("1398605")["schedule"]
        assert all(row["competition"] for row in schedule["future"])
        assert len(schedule["future"]) == len(SCHEDULE_ROWS) - 1


RESULT_ROWS = [{
    "site_match_id": "1398605", "match_no": "周一003", "home_team": "亚美尼亚",
    "away_team": "黑山", "league": "欧国联", "match_time": "2026-09-29 00:00:00",
    "home_score": 2, "away_score": 3, "half_score": "1:1", "result_flag": "0",
    "play_type": "jingcai",
}]


class TestD10Result:
    def test_按site_match_id直连输出赛果原始列(self, patch_mysql_connect):
        patch_mysql_connect([RESULT_ROWS])
        result = _make_source().query_match_result("1398605")
        assert result["home_score"] == 2 and result["away_score"] == 3
        assert result["half_score"] == "1:1"

    def test_无赛果行返回None(self, patch_mysql_connect):
        patch_mysql_connect([[]])
        assert _make_source().query_match_result("1398605") is None

    def test_执行契约SQL且参数为fid(self, patch_mysql_connect):
        conn = patch_mysql_connect([[]])
        _make_source().query_match_result("1398605")
        sql, params = conn.cursor_obj.executed[0]
        assert sql == SQL_D10_RESULT.strip()
        assert params == ("1398605",)
