# -*- coding: utf-8 -*-
"""推荐结构管线集成测试（PRD-12 F-12/F-13/F-14，ISS-FA3 交付物）。

覆盖：样本日七场"热度→综合判定→矛盾检查"全链路产出、条件回退环
（矛盾检查不通过回退 SynthesisAgent，最多 2 次后转人工裁决）、
CLI --recommendation-output 完整推荐结构落盘、MySQL 通道限制。
单元测试全部使用 fixture，禁止真实外部服务。
"""
import json
from pathlib import Path

from fa_orchestrator.analysis_runner import AnalysisRunner
from fa_orchestrator.cli import main
from fa_orchestrator.data_source import FixtureDataSource
from fa_orchestrator.synthesis_agent import SynthesisAgent

_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929.json"
_ANALYSIS = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929_analysis.json"


class _CountingSynthesisAgent:
    """包装真实 SynthesisAgent：记录每次回退重跑使用的 attempt 值。"""

    def __init__(self, inner):
        self.inner = inner
        self.attempts = []

    def analyze(self, data, attempt=0):
        self.attempts.append(attempt)
        return self.inner.analyze(data, attempt=attempt)


class _ScriptedReviewAgent:
    """脚本化 ReviewAgent：按调用次序返回预置决策，用于验证回退环与转人工路径。"""

    def __init__(self, decisions):
        self.decisions = list(decisions)
        self.calls = []

    def review(self, payload):
        self.calls.append(payload.get("attempt"))
        decision = self.decisions.pop(0) if self.decisions else "retry"
        return {
            "phase": "review",
            "checks": [{"id": "⑦", "name": "推荐结构互斥合规", "result": "na" if decision == "pass" else "violation"}],
            "decision": decision,
            "violations": [] if decision == "pass" else ["⑦"],
        }


def _make_runner(**agent_overrides):
    source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
    return AnalysisRunner(source, **agent_overrides)


class TestFullPipeline:
    def test_七场全部产出热度综合与矛盾检查结论(self):
        result = _make_runner().run("2026-09-29")
        assert result["total_matches"] == 7
        for match in result["matches"]:
            assert match["sentiment"]["phase"] == "sentiment"
            assert match["synthesis"]["phase"] == "synthesis"
            assert match["review"]["phase"] == "review"
            assert len(match["review"]["checks"]) == 7

    def test_推荐结构互斥合法且单元齐备(self):
        result = _make_runner().run("2026-09-29")
        for match in result["matches"]:
            recommendation = match["recommendation"]
            assert recommendation["mutual_exclusion"]["legal"] is True
            assert set(recommendation) == {"had_unit", "hhand_unit", "notes", "mutual_exclusion"}

    def test_无投票场次热度降级且管线不中断(self):
        # 样本日 1398599 无投票数据：热度验证降级标注，推荐结构仍产出
        result = _make_runner().run("2026-09-29")
        match = next(m for m in result["matches"] if m["fid"] == "1398599")
        assert match["sentiment"]["had"]["popular"] is None
        assert "热度数据缺失" in match["sentiment"]["note"]


class TestReviewRetryLoop:
    def _analyze_match(self, runner, fid="1398594"):
        result = runner.run("2026-09-29")
        return next(m for m in result["matches"] if m["fid"] == fid), result

    def test_首轮矛盾回退一次后通过(self):
        source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
        counting = _CountingSynthesisAgent(SynthesisAgent())
        scripted = _ScriptedReviewAgent(["retry", "pass"])
        runner = AnalysisRunner(source, synthesis_agent=counting, review_agent=scripted)
        match, _ = self._analyze_match(runner)
        # 脚本 agent 为全场次共享：仅校验首场（1398594）消费的前缀调用序列
        assert counting.attempts[:2] == [0, 1]
        assert scripted.calls[:2] == [0, 1]
        assert match["review"]["decision"] == "pass"

    def test_重试2次仍矛盾转人工裁决并登记(self):
        scripted = _ScriptedReviewAgent(["retry", "retry", "retry"])
        runner = _make_runner(review_agent=scripted)
        match, result = self._analyze_match(runner)
        # 首场消耗 0/1/2 三轮回退预算后转人工；后续场次脚本耗尽默认 retry 同样转人工
        assert scripted.calls[:3] == [0, 1, 2]
        assert match["review"]["decision"] == "manual"
        assert match["fid"] in result["manual_review_matches"]


class TestCliRecommendationOutput:
    def test_推荐结构产物落盘(self, tmp_path):
        out = tmp_path / "recommendation.json"
        exit_code = main([
            "run", "--date", "2026-09-29",
            "--fixture", str(_FIXTURE), "--analysis-fixture", str(_ANALYSIS),
            "--recommendation-output", str(out),
        ])
        assert exit_code == 0
        payload = json.loads(out.read_text(encoding="utf-8"))
        assert payload["total_matches"] == 7
        for match in payload["matches"]:
            assert "recommendation" in match
            assert "sentiment" in match
            assert "review" in match

    def test_mysql通道禁止推荐结构输出(self, tmp_path, monkeypatch, capsys):
        import os

        for key in [k for k in os.environ if k.startswith("FA_MYSQL_")]:
            monkeypatch.delenv(key, raising=False)
        exit_code = main([
            "run", "--date", "2026-09-29", "--source", "mysql",
            "--recommendation-output", str(tmp_path / "rec.json"),
        ])
        assert exit_code == 2
