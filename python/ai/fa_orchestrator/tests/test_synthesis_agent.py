# -*- coding: utf-8 -*-
"""SynthesisAgent 单元测试（PRD-12 F-13，ISS-FA3）。

覆盖：反大众前置检查（①强/弱压制+时效衰减+②实力差+③幅度词+语义解耦）、
v7.9 放弃规则优先级、撤离信号 2/3 重硬门槛边界、让球单选 P 值边界（v7.3）、
低赔补位条款（v7.12 路径 C）、推荐结构互斥（v7.12 S/D 单元）参数化断言。
纯规则计算，禁止真实 LLM 与外部服务。
"""
import pytest

from fa_orchestrator.agents.synthesis_agent import SynthesisAgent


def _fundamental(intel_items=(), h2h_suppression="none", h2h_decayed=False,
                 recent_form=None, schedule=None):
    """构造基本面结论最小形态（复用 FA2 h2h.suppression 压制力分级契约）。"""
    return {
        "intelligence": {"items": list(intel_items)},
        "h2h": {"suppression": h2h_suppression, "decayed": h2h_decayed},
        "recent_form": recent_form or {},
        "schedule": schedule or {},
    }


def _odds(two_layer=False, single_house=False, interpretation=None, value_ratio=None):
    """构造 odds_verify 结论：可开关两类撤离信号与实力背书。"""
    return {
        "euro_asia": {
            "two_layer_divergence": two_layer,
            "interpretation": interpretation,
            "strength": {"value_ratio": value_ratio, "level_gap": False, "h2h_suppression": "none"},
        },
        "anomalies": {"single_house_deviation": {"present": single_house, "pct": 6.7}},
    }


def _sentiment(overheat=None, hhand=None, bf=None):
    """构造 SentimentAgent 结论：默认不过热、必发缺失。"""
    return {
        "overheat": overheat or {"triggered": False, "layer": None, "disposition": None, "tag": None},
        "hhand": hhand or {"available": False, "deviation_pp": None, "popular": None, "note": ""},
        "bf": bf or {"available": False, "trend": None, "summary": ""},
    }


def _vote(sp_win=1.80, line=-1.0, hhand_odds=None):
    """构造竞彩投票块：默认让球 P(胜)=51.9% 且赔率 1.80（单选达标区间）。"""
    return {
        "sp": {"win": sp_win, "draw": 3.40, "lose": 4.20},
        "hhand": {"line": line, "odds": hhand_odds or {"win": 1.80, "draw": 3.60, "lose": 4.20}},
    }


def _make_input(**overrides):
    """标准输入：市场方向主胜，无撤离信号、无压制、不过热。"""
    base = {
        "fundamental": _fundamental(),
        "odds": _odds(),
        "sentiment": _sentiment(),
        "vote": _vote(),
        "market_direction": "win",
    }
    base.update(overrides)
    return base


def _greed_fear(greed_level="无", fear_level="无", combination=None):
    """构造 GreedFearAgent 结论：默认无陷阱（合成层消费为可选输入）。"""
    return {
        "phase": "greed_fear",
        "greed": {"score": 0, "level": greed_level, "signals": {}},
        "fear": {"score": 0, "level": fear_level, "signals": {}},
        "trap_combination": combination or f"{greed_level}贪婪+{fear_level}恐惧",
    }


def _agent():
    return SynthesisAgent()


class TestAntiPublicChecks:
    def test_连续胜3场判强压制禁止反买(self):
        result = _agent().analyze(_make_input(fundamental=_fundamental(h2h_suppression="strong")))
        assert result["anti_public"]["anchor1"] == "strong"
        assert result["anti_public"]["allow_counter"] is False

    def test_不败型战绩降级弱压制不阻断(self):
        # FA2 口径：不败型（无 L）→ weak，只降星不阻断（v7.10）
        result = _agent().analyze(_make_input(fundamental=_fundamental(h2h_suppression="weak")))
        assert result["anti_public"]["anchor1"] == "weak"
        # 语义解耦：弱压制只降星，不构成反买阻断，也不等于主推保留
        assert result["anti_public"]["allow_counter"] is False

    def test_压制方近5场失球达1点5触发时效衰减(self):
        fundamental = _fundamental(h2h_suppression="strong", recent_form={"avg_concede": 1.6})
        result = _agent().analyze(_make_input(fundamental=fundamental))
        assert result["anti_public"]["decay_applied"] is True
        assert result["anti_public"]["anchor1"] == "weak"

    def test_身价差2倍命中实力锚点(self):
        result = _agent().analyze(_make_input(odds=_odds(value_ratio=2.5)))
        assert result["anti_public"]["anchor2"] is True
        assert result["anti_public"]["allow_counter"] is False

    def test_情报幅度词命中锚点三(self):
        items = [{"party": "home", "sentiment": "favor", "text": "客队防线恐被击穿"}]
        result = _agent().analyze(_make_input(fundamental=_fundamental(intel_items=items)))
        assert result["anti_public"]["anchor3"] is True

    def test_三条全不命中允许反买但仓位减半(self):
        result = _agent().analyze(_make_input())
        assert result["anti_public"]["anchor1"] == "none"
        assert result["anti_public"]["allow_counter"] is True
        assert "仓位减半" in result["anti_public"]["note"]


class TestEvacuationGate:
    def test_撤离2重且情报零支撑_主推放弃(self):
        # v7.9 优先级：撤离/反向信号 ≥2 重有效 且 情报层主利 0 条 → 放弃（周六012 案例）
        items = [{"party": "away", "sentiment": "against", "text": "客队伤停"}]
        data = _make_input(
            fundamental=_fundamental(intel_items=items),
            odds=_odds(two_layer=True, single_house=True),
        )
        result = _agent().analyze(data)
        assert result["evacuation"]["count"] == 2
        assert result["evacuation"]["hard_gate"] is False
        assert result["main_action"] == "abandon"

    def test_撤离2重但情报有支撑_不触发放弃规则(self):
        items = [{"party": "home", "sentiment": "favor", "text": "主场5连胜"}]
        data = _make_input(
            fundamental=_fundamental(intel_items=items),
            odds=_odds(two_layer=True, single_house=True),
        )
        result = _agent().analyze(data)
        assert result["evacuation"]["count"] == 2
        assert result["main_action"] == "keep"

    def test_撤离3重硬门槛_情报支撑有仍放弃(self):
        # v7.11 硬门槛：≥3 重无条件放弃（周日007 案例），情报支撑不能否决
        items = [{"party": "home", "sentiment": "favor", "text": "利多一"},
                 {"party": "home", "sentiment": "favor", "text": "利多二"}]
        data = _make_input(
            fundamental=_fundamental(intel_items=items),
            odds=_odds(two_layer=True, single_house=True, interpretation="诱盘"),
        )
        result = _agent().analyze(data)
        assert result["evacuation"]["count"] == 3
        assert result["evacuation"]["hard_gate"] is True
        assert result["main_action"] == "abandon"
        # 放弃主推后改配平/负双选（胜平负层 D 单元）
        assert result["recommendation"]["had_unit"]["form"] == "D"
        assert result["recommendation"]["had_unit"]["legs"] == ["draw", "lose"]

    def test_必发成交回落计撤离信号(self):
        bf = {"available": True, "trend": "falling", "summary": "主胜成交回落"}
        data = _make_input(sentiment=_sentiment(bf=bf), odds=_odds(two_layer=True, single_house=True))
        result = _agent().analyze(data)
        assert "必发成交回落" in result["evacuation"]["signals"]


class TestHandicapSingleSelect:
    def test_概率达标且赔率达1点60判单选(self):
        result = _agent().analyze(_make_input())
        unit = result["recommendation"]["hhand_unit"]
        assert unit["form"] == "S"
        assert unit["selection"] == "win"
        assert unit["stars"] == 4

    def test_概率40至50区间仅双选(self):
        odds = {"win": 1.90, "draw": 3.40, "lose": 3.80}  # P(胜)=48.6%
        result = _agent().analyze(_make_input(vote=_vote(hhand_odds=odds)))
        unit = result["recommendation"]["hhand_unit"]
        assert unit["form"] == "D"
        assert unit["stars"] == 3

    def test_概率低于40禁止单选(self):
        odds = {"win": 2.80, "draw": 3.50, "lose": 3.90}  # P(胜)=39.7%
        result = _agent().analyze(_make_input(vote=_vote(hhand_odds=odds)))
        assert result["recommendation"]["hhand_unit"] is None
        assert any("禁止单选" in note for note in result["recommendation"]["notes"])

    def test_深盘让球线2及以上单选门槛提升至55(self):
        result = _agent().analyze(_make_input(vote=_vote(line=-2.0)))
        # P(胜)=51.9% 达常规门槛但低于深盘 55% 门槛 → 禁单选
        assert result["recommendation"]["hhand_unit"] is None
        assert any("深盘" in note for note in result["recommendation"]["notes"])

    def test_让球过热方向禁止单选(self):
        hhand = {"available": True, "deviation_pp": 16.0, "popular": "win", "note": ""}
        result = _agent().analyze(_make_input(sentiment=_sentiment(hhand=hhand)))
        assert result["recommendation"]["hhand_unit"] is None
        assert any("过热" in note for note in result["recommendation"]["notes"])

    def test_强压制命中放弃让球单选(self):
        data = _make_input(fundamental=_fundamental(h2h_suppression="strong"))
        result = _agent().analyze(data)
        assert result["recommendation"]["hhand_unit"] is None
        # 备注需含"让球"字样：供矛盾检查②区分"静默缺省"与"写明原因"
        assert any("强压制" in note and "让球" in note
                   for note in result["recommendation"]["notes"])


class TestLowSpCover:
    def test_主推SP低于1点5强制让球层双选补位(self):
        result = _agent().analyze(_make_input(vote=_vote(sp_win=1.45)))
        unit = result["recommendation"]["hhand_unit"]
        assert unit["form"] == "D"
        assert "低赔补位" in unit["tag"]
        assert unit["path"] == "A"  # 让平 3.60 / 让负 4.20 双腿 ≥2.0 自动合规

    def test_补位腿不足1点8走路径C不合规砍剩单选(self):
        odds = {"win": 1.50, "draw": 1.70, "lose": 1.90}
        result = _agent().analyze(_make_input(vote=_vote(sp_win=1.45, hhand_odds=odds)))
        unit = result["recommendation"]["hhand_unit"]
        assert unit["form"] == "S"
        assert any("不合规" in note and "1.7" in note for note in result["recommendation"]["notes"])

    def test_主推SP达1点5写明无补位需求(self):
        result = _agent().analyze(_make_input())
        assert any("无补位需求" in note for note in result["recommendation"]["notes"])


class TestMutualExclusion:
    @pytest.mark.parametrize(
        "had_form,hhand_form,hard_gate,expect_form,expect_adjusted",
        [
            ("S", "S", False, "S+S", False),
            ("S", "D", False, "S+D", False),
            ("D", "S", False, "D+S", False),
            ("D", "D", True, "D+D", False),    # 撤离≥3重放弃主推场 D+D 合法
            ("D", "D", False, "D+空", True),   # 无硬门槛 D+D → 砍让球层
            ("S", None, False, "S+空", False),
            (None, "D", False, "空+D", False),
        ],
    )
    def test_互斥合法组合清单逐项断言(self, had_form, hhand_form, hard_gate, expect_form, expect_adjusted):
        had = {"form": had_form} if had_form else None
        hhand = {"form": hhand_form} if hhand_form else None
        outcome = SynthesisAgent.enforce_mutual_exclusion(had, hhand, hard_gate)
        assert outcome["legal"] is True
        assert outcome["form"] == expect_form
        assert outcome["adjusted"] is expect_adjusted


class TestGreedFearConsumption:
    """贪婪恐惧层消费（立项三 Issue 2）：强陷阱组合警示 + 透传，不改变推荐结构。"""

    def test_强贪婪强恐惧组合追加警示并透传陷阱组合(self):
        result = _agent().analyze(_make_input(
            greed_fear=_greed_fear("强", "强", "强贪婪+强恐惧")))
        assert result["trap_combination"] == "强贪婪+强恐惧"
        assert any("贪婪恐惧" in note and "陷阱" in note
                   for note in result["recommendation"]["notes"])

    def test_无贪婪恐惧输入不透传不追加警示(self):
        result = _agent().analyze(_make_input())
        assert "trap_combination" not in result
        assert not any("贪婪恐惧" in note for note in result["recommendation"]["notes"])

    def test_弱陷阱组合透传但不追加警示(self):
        # 最小闭环口径：仅双侧强陷阱警示，弱/中陷阱只透传给报告层
        result = _agent().analyze(_make_input(
            greed_fear=_greed_fear("弱", "无", "弱贪婪+无恐惧")))
        assert result["trap_combination"] == "弱贪婪+无恐惧"
        assert not any("贪婪恐惧" in note for note in result["recommendation"]["notes"])
