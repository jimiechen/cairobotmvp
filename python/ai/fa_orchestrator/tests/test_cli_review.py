# -*- coding: utf-8 -*-
"""CLI review 子命令端到端测试（PRD-12 ISS-FA5 复盘对照入口）。

覆盖：推荐 JSON + D10 赛果 fixture → 复盘 JSON 落盘、HTML 报告落盘、
推荐文件缺失错误退出码。
"""
import json
from pathlib import Path

from fa_orchestrator.cli import main

_RESULTS = Path(__file__).parent / "fixtures" / "sample_day_20260929_results.json"


def _rec_file(tmp_path: Path) -> str:
    """构造两场推荐产物（与 FA4 推荐结构同构）。"""
    recs = {
        "match_date": "2026-09-29",
        "matches": [
            {
                "fid": "1398594", "match_no": "周二008",
                "recommendation": {
                    "had_unit": {"form": "S", "selection": "win", "legs": None,
                                 "path": None, "stars": 3, "tag": None},
                    "hhand_unit": {"form": "D", "legs": ["draw", "lose"],
                                   "path": "A", "stars": 3, "tag": None},
                },
                "synthesis": {"evacuation": {"count": 1, "signals": [],
                                             "hard_gate": False}},
                "sentiment": {"overheat": {"triggered": False, "layer": None,
                                           "deviation_pp": None,
                                           "support_intel": None,
                                           "disposition": None, "tag": None}},
            },
            {
                "fid": "1398595", "match_no": "周二009",
                "recommendation": {
                    "had_unit": {"form": "S", "selection": "lose", "legs": None,
                                 "path": None, "stars": 2, "tag": None},
                    "hhand_unit": None,
                },
                "synthesis": {"evacuation": {"count": 3, "signals": ["x"],
                                             "hard_gate": True}},
                "sentiment": {"overheat": {"triggered": False, "layer": None,
                                           "deviation_pp": None,
                                           "support_intel": None,
                                           "disposition": None, "tag": None}},
            },
        ],
    }
    path = tmp_path / "rec.json"
    path.write_text(json.dumps(recs, ensure_ascii=False), encoding="utf-8")
    return str(path)


class TestCliReview:
    def test_复盘对照端到端落盘(self, tmp_path, capsys):
        output = tmp_path / "review.json"
        code = main([
            "review",
            "--results", str(_RESULTS),
            "--recommendation-file", _rec_file(tmp_path),
            "--output", str(output),
        ])
        assert code == 0
        report = json.loads(output.read_text(encoding="utf-8"))
        assert report["total"] == 2
        assert report["matches"][0]["fid"] == "1398594"
        capsys.readouterr()

    def test_复盘同时产出html报告(self, tmp_path, capsys):
        output = tmp_path / "review.json"
        html_out = tmp_path / "review.html"
        code = main([
            "review",
            "--results", str(_RESULTS),
            "--recommendation-file", _rec_file(tmp_path),
            "--output", str(output),
            "--html", str(html_out),
        ])
        assert code == 0
        assert html_out.read_text(encoding="utf-8").startswith("<!DOCTYPE html>")
        capsys.readouterr()

    def test_推荐文件缺失退出码2(self, tmp_path, capsys):
        code = main([
            "review",
            "--results", str(_RESULTS),
            "--recommendation-file", str(tmp_path / "no_such.json"),
            "--output", str(tmp_path / "review.json"),
        ])
        assert code == 2
        capsys.readouterr()
