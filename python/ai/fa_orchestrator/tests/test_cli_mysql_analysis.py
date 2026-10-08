# -*- coding: utf-8 -*-
"""CLI mysql 分析放行测试（PRD-12 F-05/F-10~F-14，方案 A 批次 3）。

覆盖：--source mysql 时两阶段分析与推荐产出放行、_build_source 返回分析数据源、
结论降级标注（D9/D10 缺口如实透出）。全程 Fake 连接，禁止真实连库；
真实联调 e2e 走验收报告流程（tmp/fa.env 凭据）。
"""
import json
from datetime import datetime

from fa_orchestrator.cli import _build_source
from fa_orchestrator.cli import main
from fa_orchestrator.data.mysql_analysis_source import MySqlAnalysisDataSource

MATCH_DATE = "2026-09-29"

Q1_ROWS = [{"internal_id": "match_c8e30a24", "match_no": "周一003",
            "home_team": "亚美尼亚", "away_team": "黑山",
            "match_date": MATCH_DATE, "fid": "1398605"}]
CAND_ROWS = [{"id": 2635, "cc_id": "周一003", "match_date": MATCH_DATE,
              "play_type": "jingcai", "home": "亚美尼亚", "away": "黑山"}]
KICKOFF_ROWS = [{"kickoff_time": datetime(2026, 9, 29, 18, 0, 0)}]
EUROPE_ROWS = [
    {"bk": "澳门", "win": 2.1, "draw": 3.25, "lose": 3.4, "payout": 0.9,
     "is_initial": 1, "snapshot_time": datetime(2026, 9, 28, 12, 0, 0)},
    {"bk": "澳门", "win": 2.02, "draw": 3.3, "lose": 3.55, "payout": 0.905,
     "is_initial": 0, "snapshot_time": datetime(2026, 9, 29, 20, 10, 0)},
]


def _mysql_env(monkeypatch):
    for key, val in (("FA_MYSQL_HOST", "db-host"), ("FA_MYSQL_PORT", "3306"),
                     ("FA_MYSQL_USER", "tester"), ("FA_MYSQL_PASSWORD", "secret"),
                     ("FA_MYSQL_DB", "workbench")):
        monkeypatch.setenv(key, val)


class TestBuildSource:
    def test_mysql通道返回分析数据源(self, monkeypatch):
        _mysql_env(monkeypatch)

        class _Args:
            source = "mysql"
            fixture = None
            analysis_fixture = None

        assert isinstance(_build_source(_Args()), MySqlAnalysisDataSource)


class TestCliMysqlAnalysis:
    """run --source mysql 全链路（Fake 连接）：数据准备 → 两阶段分析 → 推荐产出。"""

    def _fake_result_sets(self):
        module_rows = [{"match_id": 2635, "c": n} for n in (231, 272, 32, 20, 6, 4)]
        prepare_sets = [Q1_ROWS, CAND_ROWS] + [[row] for row in module_rows]
        analyze_sets = [
            KICKOFF_ROWS, EUROPE_ROWS,
            [], [],            # asian/overunder 空（结构完整空集）
            [], [], [], [],    # intel/h2h/recent/schedule 空
            [], [],            # vote/bf 空
            [], [],            # tracks 空
            [],                # heat 空
        ]
        # CLI 先 prepare 一次，AnalysisRunner.run 内部再 prepare 一次 → 序列 ×2
        return prepare_sets * 2 + analyze_sets

    def test_两阶段分析与推荐产出放行(self, patch_mysql_connect, tmp_path, monkeypatch):
        _mysql_env(monkeypatch)
        patch_mysql_connect(self._fake_result_sets())
        analysis_out = tmp_path / "phase_conclusion.json"
        rec_out = tmp_path / "recommendation.json"
        exit_code = main([
            "run", "--date", MATCH_DATE, "--source", "mysql",
            "--analysis-output", str(analysis_out),
            "--recommendation-output", str(rec_out),
        ])
        assert exit_code == 0
        saved = json.loads(analysis_out.read_text(encoding="utf-8"))
        assert saved["channel"] == "mysql"
        assert saved["total_matches"] == 1
        item = saved["matches"][0]
        assert {"fundamental", "odds_verify", "history_deduction",
                "synthesis", "review", "recommendation"} <= set(item)
        assert json.loads(rec_out.read_text(encoding="utf-8"))["matches"][0]["recommendation"] \
            is not None

    def test_结论透出降级标注(self, patch_mysql_connect, tmp_path, monkeypatch):
        _mysql_env(monkeypatch)
        patch_mysql_connect(self._fake_result_sets())
        analysis_out = tmp_path / "phase_conclusion.json"
        exit_code = main([
            "run", "--date", MATCH_DATE, "--source", "mysql",
            "--analysis-output", str(analysis_out),
        ])
        assert exit_code == 0
        saved = json.loads(analysis_out.read_text(encoding="utf-8"))
        notes = " ".join(saved["degradation_notes"])
        assert "D9" in notes and "参考公司" in notes
        assert "D10" in notes
