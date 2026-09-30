# -*- coding: utf-8 -*-
"""安全用例测试（PRD-12 §9 安全，ISS-FA4：密钥不落日志）。

验证：LLM 网关密钥（FA_LLM_API_KEY）不得出现在 checkpoint 数据库、
报告产物、replay/show 终端输出中。密钥通过 monkeypatch 注入假值，
全程无真实外部服务调用。
"""
import json
from pathlib import Path

from fa_orchestrator.cli import main

_FIXTURE = Path(__file__).resolve().parent.parent / "fixtures" / "sample_day_20260929.json"
_ANALYSIS = Path(__file__).resolve().parent.parent / "fixtures" / "sample_day_20260929_analysis.json"
_DATE = "2026-09-29"
# 假密钥（monkeypatch 注入，非真实凭证）
_FAKE_KEY = "test-secret-FA4-must-not-leak"


def _run_to_interrupt(tmp_path, monkeypatch, capsys) -> Path:
    """图驱动跑至人工裁决中断，返回 (db, stdout)。"""
    monkeypatch.setenv("FA_LLM_API_KEY", _FAKE_KEY)
    db = tmp_path / "ckpt.db"
    code = main([
        "run", "--date", _DATE, "--fixture", str(_FIXTURE),
        "--analysis-fixture", str(_ANALYSIS), "--checkpoint-db", str(db),
    ])
    assert code == 3
    return db, capsys.readouterr().out


class TestSecretNotLeaked:
    def test_终端输出不落密钥(self, tmp_path, monkeypatch, capsys):
        _, out = _run_to_interrupt(tmp_path, monkeypatch, capsys)
        assert _FAKE_KEY not in out

    def test_checkpoint数据库不落密钥(self, tmp_path, monkeypatch, capsys):
        db, _ = _run_to_interrupt(tmp_path, monkeypatch, capsys)
        raw = db.read_bytes()
        assert _FAKE_KEY.encode() not in raw

    def test_replay与show输出不落密钥(self, tmp_path, monkeypatch, capsys):
        db, _ = _run_to_interrupt(tmp_path, monkeypatch, capsys)
        for sub in ("replay", "show"):
            main([sub, "--checkpoint-db", str(db), "--thread", _DATE,
                  "--node", "review_agent"] if sub == "show"
                 else ["replay", "--checkpoint-db", str(db), "--thread", _DATE])
            assert _FAKE_KEY not in capsys.readouterr().out

    def test_报告产物不落密钥(self, tmp_path, monkeypatch, capsys):
        db, _ = _run_to_interrupt(tmp_path, monkeypatch, capsys)
        report_dir = tmp_path / "reports"
        main([
            "run", "--date", _DATE, "--fixture", str(_FIXTURE),
            "--analysis-fixture", str(_ANALYSIS),
            "--checkpoint-db", str(db), "--resume",
            "--report-dir", str(report_dir),
        ])
        capsys.readouterr()
        for md in report_dir.glob("*.md"):
            assert _FAKE_KEY not in md.read_text(encoding="utf-8")
