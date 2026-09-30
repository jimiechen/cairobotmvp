# -*- coding: utf-8 -*-
"""CLI 可视化与追溯子命令测试（PRD-12 F-17/F-18/F-19/F-20，ISS-FA4）。

覆盖：graph（Mermaid 导出 + 一致性断言）、replay（轨迹时序回放）、
show（节点输入/输出/耗时/来源查询）、run 图驱动模式（--checkpoint-db 中断
提示人工裁决 + --resume 续跑完成）。全程 fixture，禁止真实外部服务。
"""
import json
from pathlib import Path

from fa_orchestrator.cli import main
from fa_orchestrator.graph_builder import mermaid_matches_config
from fa_orchestrator.config import load_orchestration_config

_PKG = Path(__file__).resolve().parent.parent / "fa_orchestrator"
_YAML = _PKG / "configs" / "orchestrator.yaml"
_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929.json"
_ANALYSIS = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929_analysis.json"
_DATE = "2026-09-29"
_EXPECT_NODES = [
    "planner", "data_agent", "ready_report", "fundamental_agent",
    "odds_verify_agent", "sentiment_agent", "synthesis_agent",
    "review_agent", "human_review", "report_agent", "phase_report",
]


def _checkpoint_run(tmp_path, capsys=None) -> Path:
    """先跑一次图驱动 run（interrupt 中断），返回 checkpoint db 路径。

    capsys 非空时清空缓冲区，避免 run 的提示污染后续 show 的纯 JSON 断言。
    """
    db = tmp_path / "ckpt.db"
    code = main([
        "run", "--date", _DATE, "--fixture", str(_FIXTURE),
        "--analysis-fixture", str(_ANALYSIS),
        "--checkpoint-db", str(db),
    ])
    assert code == 3, "含 manual 场的图驱动 run 应以人工裁决中断退出（F-20）"
    if capsys is not None:
        capsys.readouterr()
    return db


class TestGraphCommand:
    def test_graph导出Mermaid并写文件(self, tmp_path, capsys):
        output = tmp_path / "graph.mmd"
        code = main(["graph", "--output", str(output)])
        assert code == 0
        mermaid = output.read_text(encoding="utf-8")
        for node in _EXPECT_NODES:
            assert node in mermaid
        assert " Mermaid 源码已写入" in capsys.readouterr().out

    def test_graph一致性断言_导出与编排配置一致(self):
        """验收 6：Mermaid 导出与运行时 StateGraph 定义一致。"""
        from fa_orchestrator.graph_builder import build_graph, export_mermaid
        from fa_orchestrator.graph_nodes import GraphNodes
        config = load_orchestration_config(_YAML)
        graph = build_graph(config, GraphNodes(None).registry())
        assert mermaid_matches_config(export_mermaid(graph), config)

    def test_mermaid与配置不一致时判定失败(self):
        config = load_orchestration_config(_YAML)
        assert not mermaid_matches_config("graph TD\n  planner\n", config)

    def test_graph一致性断言失败退出码2(self, tmp_path, monkeypatch):
        import fa_orchestrator.cli_trace as cli_trace
        monkeypatch.setattr(cli_trace, "mermaid_matches_config", lambda *a: False)
        assert main(["graph", "--output", str(tmp_path / "g.mmd")]) == 2


class TestReplayAndShow:
    def test_replay渲染节点执行时序(self, tmp_path, capsys):
        db = _checkpoint_run(tmp_path, capsys)
        code = main(["replay", "--checkpoint-db", str(db), "--thread", _DATE])
        assert code == 0
        out = capsys.readouterr().out
        assert "planner" in out
        assert "耗时" in out
        assert "fixture" in out

    def test_show查询指定节点输出(self, tmp_path, capsys):
        db = _checkpoint_run(tmp_path, capsys)
        code = main([
            "show", "--checkpoint-db", str(db),
            "--thread", _DATE, "--node", "review_agent",
        ])
        assert code == 0
        events = json.loads(capsys.readouterr().out)
        assert events[0]["node"] == "review_agent"
        assert events[0]["output"], "show 必须可查节点输出（F-18）"

    def test_show节点不存在退出码2(self, tmp_path):
        db = _checkpoint_run(tmp_path)
        code = main([
            "show", "--checkpoint-db", str(db),
            "--thread", _DATE, "--node", "ghost",
        ])
        assert code == 2


class TestRunGraphDriven:
    def test_run图驱动中断时提示人工裁决(self, tmp_path, capsys):
        db = tmp_path / "ckpt.db"
        code = main([
            "run", "--date", _DATE, "--fixture", str(_FIXTURE),
            "--analysis-fixture", str(_ANALYSIS),
            "--checkpoint-db", str(db),
        ])
        out = capsys.readouterr().out
        assert code == 3
        assert "人工裁决" in out
        assert _DATE in out

    def test_run_resume后完成并落盘推荐与报告(self, tmp_path, capsys):
        db = _checkpoint_run(tmp_path)
        conclusion_path = tmp_path / "conclusion.json"
        report_dir = tmp_path / "reports"
        code = main([
            "run", "--date", _DATE, "--fixture", str(_FIXTURE),
            "--analysis-fixture", str(_ANALYSIS),
            "--checkpoint-db", str(db), "--resume",
            "--recommendation-output", str(conclusion_path),
            "--report-dir", str(report_dir),
        ])
        assert code == 0
        conclusion = json.loads(conclusion_path.read_text(encoding="utf-8"))
        assert conclusion["total_matches"] == 7
        assert conclusion["manual_review_matches"] == ["1398597"]
        reports = sorted(p.name for p in report_dir.glob("*.md"))
        assert len(reports) == 7
