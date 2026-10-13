# -*- coding: utf-8 -*-
"""ReviewAgent 单元测试（PRD-12 F-14，ISS-FA3）。

覆盖：七项矛盾检查（CONSISTENCY_CHECKS.md v7.12）逐项执行与结果落盘、
未产出层标注"不适用"、深盘次选保护位缺失/平局次选命中、双选收益比路径缺失命中、
互斥违规命中、条件回退决策（retry ≤2 次，仍不过转人工裁决）。
纯规则计算，禁止真实 LLM 与外部服务。
"""
import pytest

from fa_orchestrator.agents.review_agent import MAX_REVIEW_RETRIES, ReviewAgent


def _unit(form="S", selection="win", legs=None, path=None, tag=None, stars=3):
    """构造推荐单元最小形态（与 SynthesisAgent 输出同构）。"""
    return {"form": form, "selection": selection, "legs": legs, "path": path,
            "tag": tag, "stars": stars}


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


_GOAL_LAYER_DEFAULT = object()  # 大小球层缺省哨兵：未传参时 payload 不带 goal_layer 键（保持 na 兼容）


def _payload(synthesis=None, line=-0.5, attempt=0, goal_layer=_GOAL_LAYER_DEFAULT):
    payload = {"synthesis": synthesis or _synthesis(), "handicap_line": line,
               "attempt": attempt}
    if goal_layer is not _GOAL_LAYER_DEFAULT:
        payload["goal_layer"] = goal_layer
    return payload


def _check(result, check_id):
    return next(c for c in result["checks"] if c["id"] == check_id)


def _ou(direction="over", line=2.5, stake=0.5, tag=None):
    """构造大小球单元（与 GoalLayerAgent._build_ou_unit 输出同构）。"""
    return {"direction": direction, "line": line, "stars": 3,
            "stake": stake, "tag": tag, "note": ""}


def _pool_item(score, tag=None):
    """比分字符串 → 池单元（与 GoalLayerAgent._pool_item 同构）。"""
    home, away = (int(part) for part in score.split("-"))
    return {"score": score, "net_win": home - away, "total": home + away, "tag": tag}


def _goal_layer(ou=None, pool=None, penetration=None):
    """构造大小球/比分池层输出（与 GoalLayerAgent.analyze 输出同构，phase 必填）。"""
    return {
        "phase": "goal_layer",
        "expected_total_goals": 2.8,
        "line": 2.5,
        "ou_unit": ou,
        "ou_note": None,
        "score_pool": pool or [],
        "pool_note": None,
        "penetration": penetration or {"S1": False, "S2": False, "S3": False,
                                       "S4": False, "hits": 0},
        "n4_triggered": False,
    }


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


class TestGoalLayerChecks:
    """ISS-FA4 解锁：goal_layer 产出后检查①③④转为实检，缺失/空层保持 na 兼容。"""

    def test_图侧空大小球层视为未产出保持na(self):
        # 图驱动节点缺失场传 {}，不得误判为已产出层
        result = ReviewAgent().review(_payload(goal_layer={}))
        for check_id in ("①", "③", "④"):
            assert _check(result, check_id)["result"] == "na"
        assert result["decision"] == "pass"

    # ------------------------------------------------------------ 检查①
    def test_风险提示幅度词命中而大小球为小命中检查一(self):
        synthesis = _synthesis(notes=["0.75 盘口被击穿风险"])
        goal_layer = _goal_layer(ou=_ou(direction="under"))
        result = ReviewAgent().review(_payload(synthesis=synthesis, goal_layer=goal_layer))
        assert _check(result, "①")["result"] == "violation"
        assert result["decision"] == "retry"

    def test_风险提示幅度词命中而大小球为大通过检查一(self):
        synthesis = _synthesis(notes=["0.75 盘口被击穿风险"])
        goal_layer = _goal_layer(ou=_ou(direction="over"))
        result = ReviewAgent().review(_payload(synthesis=synthesis, goal_layer=goal_layer))
        assert _check(result, "①")["result"] == "pass"

    def test_无风险提示词时检查一通过(self):
        goal_layer = _goal_layer(ou=_ou(direction="under"))
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "①")["result"] == "pass"

    def test_大小球未配置时检查一注明通过(self):
        # 观望/缺口场景：ou_unit 为 None，仅记录不校验，note 须注明
        goal_layer = _goal_layer(ou=None, pool=[_pool_item("1-0"), _pool_item("1-1")])
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        check = _check(result, "①")
        assert check["result"] == "pass"
        assert "未配置" in (check["note"] or "")

    # ------------------------------------------------------------ 检查③
    def test_深盘双穿证据而池无净胜三命中检查三(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("1-0"), _pool_item("2-1")],
            penetration={"S1": True, "S2": True, "S3": False, "S4": True, "hits": 3},
        )
        result = ReviewAgent().review(_payload(line=-1.0, goal_layer=goal_layer))
        assert _check(result, "③")["result"] == "violation"

    def test_深盘双穿证据且池含净胜三通过检查三(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("3-0", tag="[穿盘净胜]"), _pool_item("1-0")],
            penetration={"S1": True, "S2": True, "S3": False, "S4": True, "hits": 3},
        )
        result = ReviewAgent().review(_payload(line=-1.0, goal_layer=goal_layer))
        assert _check(result, "③")["result"] == "pass"

    def test_大球方向而池全低进球命中检查三(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("1-0"), _pool_item("0-1")],
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "③")["result"] == "violation"

    def test_小球方向而池含净胜三命中检查三(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="under"),
            pool=[_pool_item("3-0"), _pool_item("1-0")],
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "③")["result"] == "violation"

    # ------------------------------------------------------------ 检查④
    def test_双穿证据而池无净胜三命中检查四(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("1-0"), _pool_item("2-1")],
            penetration={"S1": True, "S2": True, "S3": False, "S4": True, "hits": 3},
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "④")["result"] == "violation"

    def test_双穿证据而大小球为小命中检查四(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="under"),
            pool=[_pool_item("3-0"), _pool_item("1-0")],
            penetration={"S1": True, "S2": True, "S3": False, "S4": False, "hits": 2},
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "④")["result"] == "violation"

    def test_S3幅度词而大小球为小命中检查四(self):
        # 幅度层命中禁止降级配小（CONSISTENCY_CHECKS.md 检查四）
        goal_layer = _goal_layer(
            ou=_ou(direction="under"),
            pool=[_pool_item("1-1"), _pool_item("0-1")],
            penetration={"S1": False, "S2": False, "S3": True, "S4": False, "hits": 1},
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "④")["result"] == "violation"

    def test_证据命中满三而主推降星命中检查四(self):
        had = _unit(stars=2)
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("2-1"), _pool_item("3-1")],
            penetration={"S1": False, "S2": True, "S3": True, "S4": True, "hits": 3},
        )
        result = ReviewAgent().review(_payload(synthesis=_synthesis(had=had), goal_layer=goal_layer))
        assert _check(result, "④")["result"] == "violation"

    def test_证据命中满三主推全额星级通过检查四(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("2-1"), _pool_item("3-1")],
            penetration={"S1": False, "S2": True, "S3": True, "S4": True, "hits": 3},
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert _check(result, "④")["result"] == "pass"

    # ------------------------------------------------------------ 全链路
    def test_合法大小球层三项检查全过无na(self):
        goal_layer = _goal_layer(
            ou=_ou(direction="over"),
            pool=[_pool_item("2-1"), _pool_item("3-1"), _pool_item("3-0")],
        )
        result = ReviewAgent().review(_payload(goal_layer=goal_layer))
        assert result["decision"] == "pass"
        assert result["violations"] == []
        for check_id in ("①", "③", "④"):
            assert _check(result, check_id)["result"] == "pass"
