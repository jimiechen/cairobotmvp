# -*- coding: utf-8 -*-
"""DataAgent mysql 分析组装路径测试（PRD-12 F-06~F-09，方案 A 批次 2）。

覆盖：解析映射回注契约（DataAgent.prepare → set_match_ids → 分析查询）、
AnalysisRunner mysql 通道标记、Fake 连接单场端到端组装。
全程 Fake 连接，禁止真实连库（testing.md 红线）；
组装目标结构对齐 fixture 契约（sample_day_*_analysis.json）。
"""
from datetime import date, datetime

from fa_orchestrator.agents.data_agent import DataAgent
from fa_orchestrator.analysis.analysis_runner import AnalysisRunner
from fa_orchestrator.data.data_source import CHANNEL_FIXTURE, CHANNEL_MYSQL
from fa_orchestrator.data.mysql_analysis_source import MySqlAnalysisDataSource

MATCH_DATE = "2026-09-29"
FID = "1398605"
MATCH_ID = 2635

Q1_ROWS = [{"internal_id": "match_c8e30a24", "match_no": "周一003",
            "home_team": "亚美尼亚", "away_team": "黑山",
            "match_date": MATCH_DATE, "fid": FID}]
CAND_ROWS = [{"id": MATCH_ID, "cc_id": "周一003", "match_date": MATCH_DATE,
              "play_type": "jingcai", "home": "亚美尼亚", "away": "黑山"}]


def _make_source():
    return MySqlAnalysisDataSource(host="db-host", port=3306, user="tester",
                                   password="secret", database="workbench")


def _module_rows(count):
    return [{"match_id": MATCH_ID, "c": count}]


KICKOFF_ROWS = [{"kickoff_time": datetime(2026, 9, 29, 18, 0, 0)}]
EUROPE_ROWS = [
    {"bk": "澳门", "win": 2.1, "draw": 3.25, "lose": 3.4, "payout": 0.9,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
    {"bk": "澳门", "win": 2.02, "draw": 3.3, "lose": 3.55, "payout": 0.905,
     "is_initial": 0, "snapshot_time": datetime(2026, 9, 29, 20, 10, 0)},
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
INTEL_ROWS = [{"source": "500com", "title": "前瞻", "status": "filled",
               "content": "【主队】伤停恢复齐整。\n【客队】主力中卫停赛。",
               "published_at": datetime(2026, 9, 28, 8, 0, 0)}]
H2H_ROWS = [
    {"h2h_date": date(2025, 10, 12), "competition": "世预赛", "home_team": "黑山",
     "away_team": "亚美尼亚", "score": "0:2", "score_home": 0, "score_away": 2,
     "result_for_current_home": "W"},
    {"h2h_date": date(2024, 6, 8), "competition": "友谊赛", "home_team": "亚美尼亚",
     "away_team": "黑山", "score": "1:1", "score_home": 1, "score_away": 1,
     "result_for_current_home": "D"},
]
FORM_ROWS = [
    {"side": "away", "form_date": date(2026, 8, 30), "competition": "欧国联",
     "home_team": "对手A", "away_team": "黑山", "score": "1:0", "result": "W",
     "is_home": 0},
    {"side": "home", "form_date": date(2026, 8, 30), "competition": "欧国联",
     "home_team": "亚美尼亚", "away_team": "对手B", "score": "3:1", "result": "W",
     "is_home": 1},
    {"side": "home", "form_date": date(2026, 7, 24), "competition": "友谊赛",
     "home_team": "亚美尼亚", "away_team": "对手C", "score": "2:2", "result": "D",
     "is_home": 1},
]
SCHEDULE_ROWS = [
    {"id": MATCH_ID, "league": "league", "match_date": date(2026, 9, 29)},
    {"id": 2700, "league": "league", "match_date": date(2026, 10, 6)},
]
VOTE_ROWS = [
    {"pool_code": "HAD", "goal_line": None, "h_odds": 2.05, "d_odds": 3.2,
     "a_odds": 3.5, "h_support_rate": 55.5, "d_support_rate": 25.0,
     "a_support_rate": 19.5, "snapshot_time": datetime(2026, 9, 29, 10, 0, 0)},
    {"pool_code": "HHAD", "goal_line": -0.25, "h_odds": 2.1, "d_odds": 3.3,
     "a_odds": 3.4, "h_support_rate": 52.0, "d_support_rate": 27.0,
     "a_support_rate": 21.0, "snapshot_time": datetime(2026, 9, 29, 10, 0, 0)},
]
BF_ROWS = [{"trade_type": "standard", "total_volume": 86000, "home_pct": 58.0,
            "draw_pct": 24.0, "away_pct": 18.0,
            "snapshot_time": datetime(2026, 9, 29, 10, 0, 0)}]
# 立项四扩展结果集：query_vote 在 BF_ROWS 后追加 pnl/ladder 两次查询，
# E2E 序列必须同步插入（否则后续 D9/D12 结果集错位）
BF_PNL_ROWS = [{"trade_type": "standard", "bookmaker_pnl": -88.0,
                "snapshot_time": datetime(2026, 9, 29, 10, 0, 0)}]
LADDER_ROWS = [{"price": 2.3, "back_volume": 70.0, "lay_volume": 30.0}]
TRACK_ASIAN_ROWS = [{"bk": "澳门", "handicap": 0.0, "is_initial": 1},
                    {"bk": "澳门", "handicap": 0.25, "is_initial": 0}]
TRACK_OU_ROWS = [{"bk": "澳门", "line": 2.5, "is_initial": 1}]
HEAT_ROWS = [
    {"side": "亚美尼亚", "cold_hot": 9, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "黑山", "cold_hot": -30, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "亚美尼亚", "cold_hot": -12, "snapshot_time": datetime(2026, 9, 29, 15, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
    {"side": "黑山", "cold_hot": 41, "snapshot_time": datetime(2026, 9, 29, 15, 0, 0),
     "home": "亚美尼亚", "away": "黑山"},
]


class TestMatchIdsInjection:
    """解析映射回注契约：prepare 解析结果必须送达分析数据源（ID 域衔接点）。"""

    def test_prepare后解析映射回注分析查询可用(self, patch_mysql_connect):
        patch_mysql_connect(
            [Q1_ROWS, CAND_ROWS]
            + [_module_rows(n) for n in (231, 272, 32, 20, 6, 4)]
            + [KICKOFF_ROWS, EUROPE_ROWS, ASIAN_ROWS, OVERUNDER_ROWS]
        )
        source = _make_source()
        report = DataAgent(source, channel=CHANNEL_MYSQL).prepare(MATCH_DATE)
        assert report["matches"][0]["match_id"] == MATCH_ID
        odds = source.query_match_odds(FID)
        assert odds["kickoff_time"] == "09-29 18:00"
        assert odds["euro"][0]["bk"] == "澳门"

    def test_解析失败场次回注后分析查询不臆测(self, patch_mysql_connect):
        # 候选行缺失 → 四路径未命中 → 该场 match_id=None → 分析查询 None/空
        patch_mysql_connect([Q1_ROWS, []])
        source = _make_source()
        report = DataAgent(source, channel=CHANNEL_MYSQL).prepare(MATCH_DATE)
        assert report["matches"][0]["match_id"] is None
        assert source.query_match_odds(FID) is None
        assert source.query_fundamentals(FID) == {}
        assert source.query_vote(FID) is None

    def test_数据源无回注接口时prepare不受影响(self):
        # fixture 通道数据源无 set_match_ids 方法，DataAgent 必须兼容（向后兼容）
        stub = _StubSource()
        report = DataAgent(stub, channel=CHANNEL_FIXTURE).prepare(MATCH_DATE)
        assert report["total_matches"] == 1


class _StubSource:
    """最小数据源桩：验证无 set_match_ids 的数据源走 prepare 不报错。"""

    def query_match_list(self, match_date):
        return [dict(Q1_ROWS[0])]

    def query_jc_candidates(self, match_date):
        return [dict(CAND_ROWS[0], cc=CAND_ROWS[0]["cc_id"], home_team=CAND_ROWS[0]["home"],
                     away_team=CAND_ROWS[0]["away"])]

    def query_q2b_counts(self, fids, match_ids=None):
        return {fid: {"ouzhi": 30, "yazhi": 10, "daxiao": 5,
                      "vote": 10, "stats": 4, "intel": 2} for fid in fids}

    def query_match_odds(self, fid):
        return None

    def query_vote(self, fid):
        return None

    def query_fundamentals(self, fid):
        return {}

    def query_history_tracks(self, fid):
        return None

    def query_heat_drift(self, fid):
        return None


class TestAnalysisRunnerChannel:
    """AnalysisRunner 通道标记：mysql 通道结论不得误标为 fixture。"""

    def test_mysql通道结论标记为mysql(self):
        conclusion = AnalysisRunner(_StubSource(), channel=CHANNEL_MYSQL).run(MATCH_DATE)
        assert conclusion["channel"] == CHANNEL_MYSQL

    def test_默认通道仍为fixture(self):
        conclusion = AnalysisRunner(_StubSource()).run(MATCH_DATE)
        assert conclusion["channel"] == CHANNEL_FIXTURE


class TestMysqlAnalysisE2E:
    """单场端到端组装：DataAgent 解析回注 → 两阶段分析 → 推荐结构（Fake 连接）。"""

    def test_mysql通道单场端到端组装(self, patch_mysql_connect):
        patch_mysql_connect(
            [Q1_ROWS, CAND_ROWS]
            + [_module_rows(n) for n in (231, 272, 32, 20, 6, 4)]
            + [KICKOFF_ROWS, EUROPE_ROWS, ASIAN_ROWS, OVERUNDER_ROWS,
               INTEL_ROWS, H2H_ROWS, FORM_ROWS, SCHEDULE_ROWS,
               VOTE_ROWS, BF_ROWS, BF_PNL_ROWS, LADDER_ROWS,
               TRACK_ASIAN_ROWS, TRACK_OU_ROWS, HEAT_ROWS]
        )
        source = _make_source()
        conclusion = AnalysisRunner(source, channel=CHANNEL_MYSQL).run(MATCH_DATE)

        assert conclusion["channel"] == CHANNEL_MYSQL
        assert conclusion["total_matches"] == 1
        item = conclusion["matches"][0]
        assert {"fid", "match_no", "fundamental", "odds_verify", "history_deduction",
                "sentiment", "synthesis", "review", "recommendation"} <= set(item)
        # mysql 行 → fixture 契约结构的分析结论确定性断言
        assert item["fundamental"]["initial_position"]["handicap_movement"] == "升盘"
        assert item["fundamental"]["h2h"]["suppression"] == "weak"
        assert item["fundamental"]["intelligence"]["balance"] is True  # sentiment 无库维度→均势
        assert item["vote_available"] is True
        assert item["sentiment"]["bf"]["bookmaker_pnl"] == -88.0  # 立项四端到端透传
        assert item["sentiment"]["bf"]["ladder"]["top_price"] == 2.3
        assert item["history_deduction"]["trend"]["available"] is True
        assert item["history_deduction"]["trend"]["asian_deepen_count"] == 1
        assert item["history_deduction"]["trend"]["total"] == 1
        assert item["history_deduction"]["heat_drift"]["available"] is True
        assert item["history_deduction"]["heat_drift"]["home_pp"] == -21.0
        assert item["history_deduction"]["heat_drift"]["away_pp"] == 71.0
