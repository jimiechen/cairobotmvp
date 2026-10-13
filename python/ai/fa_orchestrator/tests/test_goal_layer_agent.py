# -*- coding: utf-8 -*-
"""GoalLayerAgent 单元测试（PRD-12 ISS-FA4 比分池/大小球层）。

覆盖：N6 外推 → 大小球方向判定（盘口差 ≥0.5 才给方向）、比分池生成
（方向×进球档位矩阵 + 大/小修正）、杯赛/国际赛修正（v7.8/v7.11）、
穿盘证据 S1-S4、N4 升深减仓（0.5→0.25）。
纯规则计算，禁止真实 LLM 与外部服务。
"""
import pytest

from fa_orchestrator.agents.goal_layer_agent import GoalLayerAgent


def _fundamental(expected=2.7, correction=None, home_recent5=1.6, away_recent5=1.1,
                 items=()):
    """构造 fundamental 结论（近期状态块 + 情报 items）。"""
    return {
        "recent_form": {
            "expected_total_goals": expected,
            "home_recent5_avg": home_recent5,
            "away_recent5_avg": away_recent5,
            "correction": correction,
        },
        "intelligence": {"items": list(items)},
    }


def _overunder(line=2.5, over=1.85, under=1.95, initial_line=None):
    """构造竞彩大小球行（pair_odds 结构，澳门行优先）。"""
    return [{"bk": "澳门",
             "initial": {"over_odds": over, "line": initial_line or line,
                         "under_odds": under},
             "current": {"over_odds": over, "line": line, "under_odds": under}}]


def _synthesis(main_action="keep", anchor2=False, notes=()):
    """构造 synthesis 结论（S2 透传 + 幅度词源 notes）。"""
    return {"main_action": main_action,
            "anti_public": {"anchor2": anchor2},
            "recommendation": {"notes": list(notes),
                               "had_unit": {"form": "S", "stars": 3}}}


def _data(expected=2.7, line=2.5, initial_line=None, direction="win",
          movement="持平", correction=None, anchor2=False, notes=(),
          baseline=None, items=()):
    return {
        "fundamental": _fundamental(expected=expected, correction=correction,
                                    items=items),
        "overunder": _overunder(line=line, initial_line=initial_line),
        "market_direction": direction,
        "handicap_line": -0.5,
        "handicap_movement": movement,
        "synthesis": _synthesis(anchor2=anchor2, notes=notes),
        "odds_verify": {"structure": {"baseline_probs": baseline or {}}},
    }


def _ou(result):
    return result["ou_unit"]


class TestOuDirection:
    def test_外推显著高于盘口判大球(self):
        result = GoalLayerAgent().analyze(_data(expected=3.2, line=2.5))
        assert _ou(result)["direction"] == "over"
        assert _ou(result)["stake"] == 0.5

    def test_外推显著低于盘口判小球(self):
        result = GoalLayerAgent().analyze(_data(expected=1.8, line=2.5))
        assert _ou(result)["direction"] == "under"

    def test_外推与盘口贴近观望不配置(self):
        result = GoalLayerAgent().analyze(_data(expected=2.7, line=2.5))
        assert _ou(result) is None
        assert "观望" in result["ou_note"]

    def test_大小球盘缺失无法配置(self):
        data = _data(expected=3.2)
        data["overunder"] = []
        result = GoalLayerAgent().analyze(data)
        assert _ou(result) is None
        assert "大小球盘" in result["ou_note"]

    def test_无外推数据降级不编造(self):
        data = _data(expected=None, line=2.5)
        result = GoalLayerAgent().analyze(data)
        assert _ou(result) is None
        assert "外推" in result["ou_note"]

    def test_结论透出外推与盘口线(self):
        result = GoalLayerAgent().analyze(_data(expected=3.2, line=2.5))
        assert result["expected_total_goals"] == 3.2
        assert result["line"] == 2.5


class TestScorePool:
    def test_主胜方向比分池全主胜比分(self):
        result = GoalLayerAgent().analyze(_data(direction="win"))
        assert result["score_pool"]
        assert all(item["net_win"] > 0 for item in result["score_pool"])

    def test_客胜方向比分池全客胜比分(self):
        result = GoalLayerAgent().analyze(_data(direction="lose"))
        assert all(item["net_win"] < 0 for item in result["score_pool"])

    def test_大球配置池含总进球三以上比分(self):
        # 外推 2.0 vs 盘口 1.5 → 大球；基础池为低进球档，须修正含 total≥3 比分
        result = GoalLayerAgent().analyze(_data(expected=2.0, line=1.5))
        assert _ou(result)["direction"] == "over"
        assert any(item["total"] >= 3 for item in result["score_pool"])

    def test_小球配置池不含净胜三球比分(self):
        result = GoalLayerAgent().analyze(_data(expected=1.8, line=2.5))
        assert _ou(result)["direction"] == "under"
        assert all(abs(item["net_win"]) < 3 for item in result["score_pool"])

    def test_方向缺失比分池为空并标注(self):
        result = GoalLayerAgent().analyze(_data(direction=None))
        assert result["score_pool"] == []
        assert "方向" in result["ou_note"] or result.get("pool_note")


class TestCompetitionCorrection:
    def test_杯赛外推与市场升线冲突弃配(self):
        # 外推判小但盘口升线（市场倾向大）→ 主锚失效冲突，不配置
        result = GoalLayerAgent().analyze(
            _data(expected=1.8, line=2.5, initial_line=2.0,
                  correction="cup", movement="升盘"))
        assert _ou(result) is None
        assert "主锚失效冲突" in result["ou_note"]

    def test_杯赛外推与市场动作同向正常配置(self):
        result = GoalLayerAgent().analyze(
            _data(expected=1.8, line=2.5, initial_line=3.0,
                  correction="cup", movement="降盘"))
        assert _ou(result)["direction"] == "under"

    def test_国际赛大小球轻仓不超0点3(self):
        result = GoalLayerAgent().analyze(_data(expected=3.2, line=2.5,
                                                correction="international"))
        assert _ou(result)["direction"] == "over"
        assert _ou(result)["stake"] <= 0.3
        assert "国际赛方差大" in (_ou(result)["tag"] or "")

    def test_联赛场不触发修正仓位(self):
        result = GoalLayerAgent().analyze(_data(expected=3.2, line=2.5))
        assert _ou(result)["stake"] == 0.5
        assert not result["n4_triggered"]


class TestPenetration:
    def test_S1_基准去水概率达64命中(self):
        result = GoalLayerAgent().analyze(
            _data(expected=2.0, line=2.5, baseline={"win": 0.66, "draw": 0.2, "lose": 0.14}))
        assert result["penetration"]["S1"] is True

    def test_S2_反大众实力差透传命中(self):
        result = GoalLayerAgent().analyze(_data(expected=2.0, line=2.5, anchor2=True))
        assert result["penetration"]["S2"] is True

    def test_S3_备注与情报幅度词命中(self):
        result = GoalLayerAgent().analyze(
            _data(expected=2.0, line=2.5, notes=["0.75 盘口被击穿风险"],
                  items=[{"party": "home", "sentiment": "favor", "text": "客队恐崩盘"}]))
        assert result["penetration"]["S3"] is True

    def test_S4_大球方向或外推三球命中(self):
        result = GoalLayerAgent().analyze(_data(expected=3.2, line=2.5))
        assert result["penetration"]["S4"] is True

    def test_S1加S2双命中比分池追加净胜三与反向保护(self):
        # 外推 2.8 → mid 档基础池无净胜 3+，S1+S2 双命中强制追加并标注
        result = GoalLayerAgent().analyze(
            _data(expected=2.8, line=2.5, direction="win",
                  baseline={"win": 0.66, "draw": 0.2, "lose": 0.14}, anchor2=True))
        scores = [item["score"] for item in result["score_pool"]]
        tags = [item["tag"] for item in result["score_pool"]]
        assert "3-0" in scores and "0-1" in scores
        assert "[穿盘净胜]" in tags and "[反向爆冷保护]" in tags

    def test_S1加S2双命中基础池已含净胜三不重复追加(self):
        result = GoalLayerAgent().analyze(
            _data(expected=3.2, line=2.5, direction="win",
                  baseline={"win": 0.66, "draw": 0.2, "lose": 0.14}, anchor2=True))
        scores = [item["score"] for item in result["score_pool"]]
        assert any(abs(item["net_win"]) >= 3 for item in result["score_pool"])
        assert scores.count("3-0") == 1


class TestN4:
    def test_升盘深盘触发穿盘预警减仓(self):
        data = _data(expected=3.2, line=2.5, movement="升盘")
        data["handicap_line"] = -1.25
        result = GoalLayerAgent().analyze(data)
        assert _ou(result)["stake"] == 0.25
        assert "穿盘预警" in _ou(result)["tag"]
        assert result["n4_triggered"] is True

    def test_升盘浅盘不触发减仓(self):
        data = _data(expected=3.2, line=2.5, movement="升盘")
        data["handicap_line"] = -0.5
        result = GoalLayerAgent().analyze(data)
        assert _ou(result)["stake"] == 0.5
        assert result["n4_triggered"] is False

    def test_深盘未升盘不触发减仓(self):
        data = _data(expected=3.2, line=2.5, movement="持平")
        data["handicap_line"] = -1.25
        result = GoalLayerAgent().analyze(data)
        assert _ou(result)["stake"] == 0.5
