# -*- coding: utf-8 -*-
"""DataAgent 与 CLI 端到端测试（PRD-12 F-05/F-06~F-09，ISS-FA1 骨架）。

全程使用 fixture 数据源（tests/fixtures/sample_day_20260929.json），
禁止真实 MySQL（192.168.1.6:3306）与 HTTP（192.168.1.6:18000）调用。
"""
import json
from pathlib import Path

from fa_orchestrator.cli import main
from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.data_source import FixtureDataSource

_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929.json"


class TestFixtureDataSource:
    def test_Q1返回七场清单(self):
        source = FixtureDataSource(_FIXTURE)
        matches = source.query_match_list("2026-09-29")
        assert len(matches) == 7
        assert all("fid" in m and "match_no" in m for m in matches)

    def test_Q2b返回六模块行数(self):
        source = FixtureDataSource(_FIXTURE)
        counts = source.query_q2b_counts(["1398594"])
        assert counts["1398594"]["ouzhi"] == 30
        assert set(counts["1398594"]) == {
            "ouzhi", "yazhi", "daxiao", "vote", "stats", "intel",
        }

    def test_候选行供四路径解析(self):
        source = FixtureDataSource(_FIXTURE)
        rows = source.query_jc_candidates()
        assert any(r["id"] == 9001 for r in rows)


class TestDataAgentPrepare:
    def test_产出日级就绪率报告(self):
        source = FixtureDataSource(_FIXTURE)
        report = DataAgent(source).prepare("2026-09-29")
        assert report["match_date"] == "2026-09-29"
        assert report["total_matches"] == 7
        # fixture 中周二013（1398599）缺 vote → 就绪 6/7
        assert report["ready_matches"] == 6
        assert report["matches"][0]["fid"] == "1398594"

    def test_每场含解析后的match_id与溯源(self):
        source = FixtureDataSource(_FIXTURE)
        report = DataAgent(source).prepare("2026-09-29")
        by_fid = {m["fid"]: m for m in report["matches"]}
        # fid → match.id 四路径解析落点（路径1：cc+窗口）
        assert by_fid["1398594"]["match_id"] == 9001
        # 溯源快照（F-09）：数据块编号 + 通道 + 快照时间
        trace = by_fid["1398594"]["fetch_trace"]
        assert trace["blocks"] == ["D1", "D2b"]
        assert trace["channel"] == "fixture"

    def test_解析失败场次不得静默跳过(self):
        source = FixtureDataSource(_FIXTURE)
        report = DataAgent(source).prepare("2026-09-29")
        by_fid = {m["fid"]: m for m in report["matches"]}
        # 全部七场都必须出现在报告中（含未就绪场）
        assert len(by_fid) == 7
        assert by_fid["1398599"]["integrity"]["status"] == "incomplete"


class TestCli:
    def test_run_date输出就绪率报告(self, capsys, tmp_path):
        output = tmp_path / "ready_report.json"
        exit_code = main([
            "run", "--date", "2026-09-29",
            "--fixture", str(_FIXTURE),
            "--output", str(output),
        ])
        assert exit_code == 0
        saved = json.loads(output.read_text(encoding="utf-8"))
        assert saved["total_matches"] == 7
        assert saved["ready_matches"] == 6

        stdout_text = capsys.readouterr().out
        assert "2026-09-29" in stdout_text
        assert "6/7" in stdout_text
