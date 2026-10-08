# -*- coding: utf-8 -*-
"""DataAgent 与 CLI 端到端测试（PRD-12 F-05/F-06~F-09，ISS-FA1 骨架）。

全程使用 fixture 数据源（tests/fixtures/sample_day_20260929.json），
禁止真实 MySQL（192.168.1.6:3306）与 HTTP（192.168.1.6:18000）调用。
"""
import json
from pathlib import Path

import pytest

from fa_orchestrator.cli import main
from fa_orchestrator.agents.data_agent import DataAgent
from fa_orchestrator.data.data_source import FixtureDataSource

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


class TestFixtureDayWindow:
    """Q1 竞彩日跨日窗口（PRD-12 §12.3）：[D 12:00, D+1 12:00) 左闭右开。

    fixture 通道与 Q1 SQL 契约口径一致，行 match_time 为 'HH:MM' 零填充字符串。
    """

    def _row(self, internal_id, match_date, match_time=None):
        row = {
            "internal_id": internal_id, "match_no": f"测试{internal_id:03d}",
            "home_team": "甲队", "away_team": "乙队",
            "match_date": match_date, "fid": str(1489000 + internal_id),
        }
        if match_time is not None:
            row["match_time"] = match_time
        return row

    def _query(self, tmp_path, rows):
        payload = {
            "match_date": "2026-10-08", "q1_matches": rows,
            "jc_match_rows": [], "q2b_counts": {},
        }
        path = tmp_path / "window_fixture.json"
        path.write_text(json.dumps(payload, ensure_ascii=False), encoding="utf-8")
        return FixtureDataSource(path).query_match_list("2026-10-08")

    def test_当日1159排除_1200含(self, tmp_path):
        rows = [self._row(1, "2026-10-08", "11:59"), self._row(2, "2026-10-08", "12:00")]
        matches = self._query(tmp_path, rows)
        assert [m["internal_id"] for m in matches] == [2]

    def test_次日1159含_1200排除(self, tmp_path):
        rows = [self._row(3, "2026-10-09", "11:59"), self._row(4, "2026-10-09", "12:00")]
        matches = self._query(tmp_path, rows)
        assert [m["internal_id"] for m in matches] == [3]

    def test_当日晚间场次含窗口(self, tmp_path):
        # 真实形态锚点：周四001 类当日 23:00 开球场次必须落入当日窗口
        rows = [self._row(5, "2026-10-08", "23:00")]
        matches = self._query(tmp_path, rows)
        assert [m["internal_id"] for m in matches] == [5]

    def test_窗口外日期排除且缺match_time不臆测(self, tmp_path):
        # D+2 / D-1 不落窗口；缺 match_time 的行判不在窗口（禁止静默放宽口径）
        rows = [
            self._row(6, "2026-10-10", "03:00"),
            self._row(7, "2026-10-07", "23:00"),
            self._row(8, "2026-10-08"),
        ]
        assert self._query(tmp_path, rows) == []


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


class TestCliMysqlSource:
    """--source mysql 真实通道（R-04）：配置缺失必须明确报错，不静默。"""

    _MYSQL_ENVS = ("FA_MYSQL_HOST", "FA_MYSQL_PORT", "FA_MYSQL_USER",
                   "FA_MYSQL_PASSWORD", "FA_MYSQL_DB")

    def test_source_mysql缺配置时报错退出码2(self, monkeypatch, capsys):
        for key in self._MYSQL_ENVS:
            monkeypatch.delenv(key, raising=False)
        exit_code = main(["run", "--date", "2026-09-29", "--source", "mysql"])
        assert exit_code == 2
        assert "FA_MYSQL" in capsys.readouterr().err

    def test_source_mysql分析产出不再提前拒绝(self, monkeypatch):
        # 方案 A 批次 3：分析类契约已归档，mysql 通道放行两阶段分析；
        # 不再提前返回 2，流程推进到真实连接（'db' 主机不可达 → 连接异常）
        import pymysql

        for key, val in (("FA_MYSQL_HOST", "db"), ("FA_MYSQL_USER", "u"),
                         ("FA_MYSQL_PASSWORD", "p"), ("FA_MYSQL_DB", "workbench")):
            monkeypatch.setenv(key, val)
        with pytest.raises(pymysql.err.OperationalError):
            main([
                "run", "--date", "2026-09-29", "--source", "mysql",
                "--analysis-output", "out.json",
            ])
