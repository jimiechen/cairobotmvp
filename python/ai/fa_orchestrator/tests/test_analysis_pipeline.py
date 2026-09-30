# -*- coding: utf-8 -*-
"""两阶段分析管线端到端测试（PRD-12 F-10/F-11/F-16，ISS-FA2 交付物）。

覆盖：analysis fixture 数据源扩展查询、AnalysisRunner 七场中间结论产出、
CLI --analysis-output 落盘。全程 fixture 数据源，禁止真实 MySQL 调用。
"""
import json
from pathlib import Path

from fa_orchestrator.analysis_runner import AnalysisRunner
from fa_orchestrator.cli import main
from fa_orchestrator.data_source import FixtureDataSource

_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929.json"
_ANALYSIS = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929_analysis.json"


def _make_source():
    return FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)


class TestAnalysisFixtureSource:
    def test_Q3返回单场三盘赔率明细(self):
        odds = _make_source().query_match_odds("1398594")
        assert odds is not None
        assert odds["kickoff_time"] == "09-30 18:00"
        assert {"euro", "asian", "overunder"} <= set(odds)
        assert len(odds["euro"]) == 4
        assert odds["asian"][0]["current"]["handicap"] == 0.25

    def test_Q5投票数据_缺失场返回None(self):
        source = _make_source()
        assert source.query_vote("1398594")["support"]["win"] == 55.5
        # 周二013（1398599）无投票，与 q2b_counts.vote=0 一致
        assert source.query_vote("1398599") is None

    def test_Q9至Q12返回基本面四块数据(self):
        data = _make_source().query_fundamentals("1398594")
        assert {"intelligence", "h2h", "recent_form", "schedule"} <= set(data)
        assert data["schedule"]["competition"] == "league"

    def test_历史推演数据缺失场返回None(self):
        source = _make_source()
        assert source.query_history_tracks("1398594") is not None
        assert source.query_history_tracks("1398600") is None
        assert source.query_heat_drift("1398597") is None

    def test_未配置analysis时查询返回None保持向后兼容(self):
        source = FixtureDataSource(_FIXTURE)
        assert source.query_match_odds("1398594") is None
        assert source.query_fundamentals("1398594") == {}
        assert source.query_vote("1398594") is None


class TestAnalysisRunner:
    def test_七场全部产出两阶段中间结论(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        assert conclusion["match_date"] == "2026-09-29"
        assert conclusion["total_matches"] == 7
        assert len(conclusion["matches"]) == 7
        for item in conclusion["matches"]:
            assert {"fid", "match_no", "fundamental", "odds_verify", "history_deduction"} <= set(item)
            assert item["fundamental"]["phase"] == "fundamental"
            assert item["odds_verify"]["phase"] == "odds_verify"

    def test_每场基本面含第一阶段五维结论(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        item = {m["fid"]: m for m in conclusion["matches"]}["1398594"]
        fundamental = item["fundamental"]
        assert {"intelligence", "initial_position", "h2h", "recent_form", "schedule"} <= set(fundamental)
        # 周二008 情报 3/1/1/2/1 → 利好天平倾向主队
        assert fundamental["intelligence"]["home_favor"] == 3
        assert fundamental["intelligence"]["balance"] is False

    def test_周二008升盘低水判真信号(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        item = {m["fid"]: m for m in conclusion["matches"]}["1398594"]
        euro_asia = item["odds_verify"]["euro_asia"]
        assert euro_asia["raise_count"] == 4
        assert euro_asia["interpretation"] == "真信号"

    def test_周二013两层背离被识别且投票热度数据缺失(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        by_fid = {m["fid"]: m for m in conclusion["matches"]}
        assert by_fid["1398599"]["odds_verify"]["euro_asia"]["two_layer_divergence"] is True
        assert by_fid["1398599"]["vote_available"] is False

    def test_周二014单家临场背离与无法推演时序标注(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        by_fid = {m["fid"]: m for m in conclusion["matches"]}
        deviation = by_fid["1398600"]["odds_verify"]["anomalies"]["single_house_deviation"]
        assert deviation["present"] is True
        assert deviation["pct"] == 6.7
        trend = by_fid["1398600"]["history_deduction"]["trend"]
        assert trend["available"] is False
        assert "无法推演时序" in trend["note"]

    def test_杯赛与国际赛场次输出对应修正(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        by_fid = {m["fid"]: m for m in conclusion["matches"]}
        assert by_fid["1398596"]["fundamental"]["recent_form"]["correction"] == "cup"
        assert by_fid["1398597"]["fundamental"]["recent_form"]["correction"] == "international"

    def test_周二012级别断层压平降级为利润平衡(self):
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        by_fid = {m["fid"]: m for m in conclusion["matches"]}
        anti_cold = by_fid["1398598"]["odds_verify"]["anomalies"]["anti_cold"]
        assert anti_cold["tag"] == "[利润平衡，降级]"

    def test_实力背书由基本面结论联动注入(self):
        # 周二013 交锋连续3胜 → strong 压制 → 注入 odds 层 strength_endorsed
        conclusion = AnalysisRunner(_make_source()).run("2026-09-29")
        by_fid = {m["fid"]: m for m in conclusion["matches"]}
        euro_asia = by_fid["1398599"]["odds_verify"]["euro_asia"]
        assert euro_asia["strength"]["h2h_suppression"] == "strong"
        assert euro_asia["strength_endorsed"] is True


class TestCliAnalysisOutput:
    def test_run产出两阶段中间结论JSON并落盘(self, tmp_path):
        output = tmp_path / "phase_conclusion.json"
        exit_code = main([
            "run", "--date", "2026-09-29",
            "--fixture", str(_FIXTURE),
            "--analysis-fixture", str(_ANALYSIS),
            "--analysis-output", str(output),
        ])
        assert exit_code == 0
        saved = json.loads(output.read_text(encoding="utf-8"))
        assert saved["total_matches"] == 7
        assert saved["matches"][0]["fundamental"]["phase"] == "fundamental"
