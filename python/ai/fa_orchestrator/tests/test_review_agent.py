# -*- coding: utf-8 -*-
"""ReviewAgent 单元测试（PRD-12 F-14，ISS-FA3）。

覆盖：七项矛盾检查（CONSISTENCY_CHECKS.md v7.12）逐项执行与结果落盘、
未产出层标注"不适用"、深盘次选保护位缺失/平局次选命中、双选收益比路径缺失命中、
互斥违规命中、条件回退决策（retry ≤2 次，仍不过转人工裁决）。
纯规则计算，禁止真实 LLM 与外部服务。
"""
import pytest

from fa_orchestrator.review_agent import MAX_REVIEW_RETRIES, ReviewAgent


def _unit(form="S", selection="win", legs=None, path=None, tag=None):
    """构造推荐单元最小形态（与 SynthesisAgent 输出同构）。"""
    return {"form": form, "selection": selection, "legs": legs, "path": path, "tag": tag, "stars": 3}


_HHAND_DEFAULT = object()  # 让球单元缺省哨兵：区分"未传参"与显式 None（让球层为空）


def _synthesis(had=None, hhand=_HHAND_DEFAULT, notes=(), legal=True):
    """构造 SynthesisAgent 结论：默认主胜单选 + 让球单选、互斥合法。

    hhand 显式传 None 表示让球层为空单元（触发检查②的静默缺省场景）。
    """
    return {
        "main_action": "keep",
        "recommendation": {
            "had_unit": had or _unit(),
            "hhand_unit": _unit(selection="win") if hhand is _HHAND_DEFAULT else hhand,
            "notes": list(notes),
            "mutual_exclusion": {"legal": legal, "form": "S+S", "adjusted": False},
        },
    }


def _payload(synthesis=None, line=-0.5, attempt=0):
    return {"synthesis": synthesis or _synthesis(), "handicap_line": line, "attempt": attempt}


def _check(result, check_id):
    return next(c for c in result["checks"] if c["id"] == check_id)


class TestSevenChecks:
    def test_七项检查全部执行并落盘(self):
        result = ReviewAgent().review(_payload())
        assert [c["id"] for c in result["checks"]] == ["①", "②", "③", "④", "⑤", "⑥", "⑦"]
        assert all("result" in c for c in result["checks"])

    def test_全部合规判通过(self):
        result = ReviewAgent().review(_payload())
        assert result["decision"] == "pass"
        assert result["violations"] == []

    def test_未产出层标注不适用且不触发回退(self):
        result = ReviewAgent().review(_payload())
        for check_id in ("①", "③", "④"):
            assert _check(result, check_id)["result"] == "na"


class TestDeepLineChecks:
    def test_深盘次选保护位静默缺失命中检查二(self):
        # 深盘主推单选但让球层为空且未写明原因 → 次选保护位静默缺省
        payload = _payload(synthesis=_synthesis(hhand=None, notes=[]), line=-1.5)
        result = ReviewAgent().review(payload)
        assert _check(result, "②")["result"] == "violation"
        assert _check(result, "②")["action"]
        assert result["decision"] == "retry"
        assert "②" in result["violations"]

    def test_深盘次选缺失但写明不配置原因则通过(self):
        payload = _payload(synthesis=_synthesis(hhand=None, notes=["让球方向禁止单选"]), line=-1.5)
        result = ReviewAgent().review(payload)
        assert _check(result, "②")["result"] == "pass"

    def test_深盘双选含平局腿命中检查五(self):
        had = _unit(form="D", legs=["draw", "lose"])
        payload = _payload(synthesis=_synthesis(had=had), line=-1.5)
        result = ReviewAgent().review(payload)
        assert _check(result, "⑤")["result"] == "violation"

    def test_撤离放弃的平负双选豁免检查五(self):
        # v7.11 撤离≥3重 → 放弃主推改配平/负双选，属反爆冷保护而非次选降级
        had = _unit(form="D", legs=["draw", "lose"])
        synthesis = _synthesis(had=had)
        synthesis["main_action"] = "abandon"
        result = ReviewAgent().review(_payload(synthesis=synthesis, line=-1.5))
        assert _check(result, "⑤")["result"] == "pass"


class TestDualPathAndExclusion:
    def test_双选缺收益比路径命中检查六(self):
        hhand = _unit(form="D", legs=["win", "draw"], path=None)
        result = ReviewAgent().review(_payload(synthesis=_synthesis(hhand=hhand)))
        assert _check(result, "⑥")["result"] == "violation"
        assert result["decision"] == "retry"

    @pytest.mark.parametrize("path", ["A", "B", "C"])
    def test_双选路径ABC均合规(self, path):
        hhand = _unit(form="D", legs=["win", "draw"], path=path)
        result = ReviewAgent().review(_payload(synthesis=_synthesis(hhand=hhand)))
        assert _check(result, "⑥")["result"] == "pass"

    def test_互斥违规命中检查七(self):
        payload = _payload(synthesis=_synthesis(legal=False))
        result = ReviewAgent().review(payload)
        assert _check(result, "⑦")["result"] == "violation"


class TestRetryDecision:
    def test_重试1次仍矛盾继续回退(self):
        result = ReviewAgent().review(_payload(synthesis=_synthesis(legal=False), attempt=1))
        assert result["decision"] == "retry"

    def test_重试2次仍矛盾转人工裁决(self):
        result = ReviewAgent().review(_payload(synthesis=_synthesis(legal=False), attempt=2))
        assert result["decision"] == "manual"
        assert MAX_REVIEW_RETRIES == 2

    def test_重试2次但全合规判通过(self):
        result = ReviewAgent().review(_payload(synthesis=_synthesis(), attempt=2))
        assert result["decision"] == "pass"
