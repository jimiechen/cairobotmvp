# -*- coding: utf-8 -*-
"""FundamentalAgent 测试（PRD-12 F-10）：v7.12 第一阶段规则节点化。

案例锚点回归（v7.2/v7.7/v7.8/v7.10/v7.11/v7.12）：拉科场攻防主锚、弗赖堡场火力
保留、均势约束、压制力分级（周六012 弱压制）、杯赛/国际赛修正。全程纯规则计算，
不调用 LLM 与外部服务。
"""
import pytest

from fa_orchestrator.agents.fundamental_agent import (
    BALANCE_CONCLUSION,
    PHASE1_WEIGHTS,
    FundamentalAgent,
)


def _intel(items):
    """构造情报输入：items 为 (party, sentiment) 元组列表。"""
    return {"items": [{"party": p, "sentiment": s, "text": f"{p}-{s}"} for p, s in items]}


def _h2h(results, decay=None):
    """构造交锋输入：results 按年份升序的 (year, result) 列表。"""
    return {"records": [{"year": y, "result": r, "score": "1-0"} for y, r in results],
            "decay_flags": decay or []}


def _form(home, away):
    """构造近期状态输入：home/away 为 (is_friendly, scored, conceded) 列表。"""
    def _side(pairs):
        return {"matches": [{"is_friendly": bool(f), "scored": s, "conceded": c}
                            for f, s, c in pairs]}
    return {"home": _side(home), "away": _side(away)}


def _schedule(competition="league", level_gap=False):
    return {"competition": competition, "level_gap": level_gap,
            "future": [{"days": 3, "competition": competition}]}


def _base_match_data(**overrides):
    """标准联赛场基线：情报均衡、平手盘、无压制、平淡攻防。"""
    data = {
        "intelligence": _intel([("home", "favor"), ("away", "favor")]),
        "initial_position": {
            "initial_handicap": 0.0, "current_handicap": 0.0,
            "initial_home_water": 0.90, "initial_away_water": 0.95,
            "current_home_water": 0.88, "current_away_water": 0.97,
            "same_handicap": None,
        },
        "h2h": _h2h([(2024, "W"), (2025, "L"), (2026, "D")]),
        "recent_form": _form(
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        ),
        "schedule": _schedule(),
    }
    data.update(overrides)
    return data


class TestWeights:
    def test_默认权重符合v712配置且总和为100(self):
        assert PHASE1_WEIGHTS == {
            "intelligence": 30, "initial_position": 25, "h2h": 15,
            "recent_form": 20, "schedule": 10,
        }
        assert sum(PHASE1_WEIGHTS.values()) == 100

    def test_分析输出携带权重配置(self):
        result = FundamentalAgent().analyze(_base_match_data())
        assert result["weights"] == PHASE1_WEIGHTS
        assert result["phase"] == "fundamental"


class TestIntelligence:
    def test_归属方缺失的利弊条目按缺口忽略不编造计数(self):
        # 真库场景：立项二语义判读可能产出 party=None + sentiment=favor/against
        data = _base_match_data(intelligence=_intel([
            ("home", "favor"), (None, "against"), (None, "favor"),
        ]))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["home_favor"] == 1
        assert result["home_against"] == 0
        assert result["away_against"] == 0
        assert result["away_favor"] == 0

    def test_利弊比量化按条目计数输出五向计数(self):
        data = _base_match_data(intelligence=_intel([
            ("home", "favor"), ("home", "favor"), ("home", "favor"),
            ("away", "favor"), ("home", "against"),
            ("away", "against"), ("away", "against"), ("both", "general"),
        ]))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["home_favor"] == 3
        assert result["away_favor"] == 1
        assert result["home_against"] == 1
        assert result["away_against"] == 2
        assert result["general"] == 1

    def test_利弊比差值小于等于1时判均势并输出均势结论(self):
        # v7.7 均势约束锚点：3/3/2/2/1 → 禁用"利好天平"类结论
        data = _base_match_data(intelligence=_intel([
            ("home", "favor"), ("home", "favor"), ("home", "favor"),
            ("away", "favor"), ("away", "favor"), ("away", "favor"),
            ("home", "against"), ("home", "against"),
            ("away", "against"), ("away", "against"), ("both", "general"),
        ]))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["balance"] is True
        assert result["conclusion"] == BALANCE_CONCLUSION

    def test_利弊比倾斜时输出利好天平倾向结论(self):
        data = _base_match_data(intelligence=_intel(
            [("home", "favor")] * 4 + [("away", "favor")]))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["balance"] is False
        assert "主队" in result["conclusion"]

    def test_单方向利多情报大于等于3条且对方不足3条时标注密集方向(self):
        data = _base_match_data(intelligence=_intel(
            [("away", "favor")] * 3 + [("home", "favor")]))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["dense_direction"] == "away"

    def test_双方利多均大于等于3条时互抵不标密集方向(self):
        data = _base_match_data(intelligence=_intel(
            [("home", "favor")] * 3 + [("away", "favor")] * 3))
        result = FundamentalAgent().analyze(data)["intelligence"]
        assert result["dense_direction"] is None


class TestInitialPosition:
    def test_主队同盘胜率达80以上时输出深盘适应性强提示(self):
        data = _base_match_data(initial_position={
            "initial_handicap": 0.5, "current_handicap": 0.75,
            "initial_home_water": 0.90, "initial_away_water": 0.95,
            "current_home_water": 0.85, "current_away_water": 1.00,
            "same_handicap": {"home": {"win": 8, "draw": 1, "lose": 1},
                              "away": {"win": 2, "draw": 2, "lose": 2}},
        })
        result = FundamentalAgent().analyze(data)["initial_position"]
        assert "深盘适应性强" in result["same_handicap_hint"]

    def test_客队同盘输盘率达60以上时输出下盘疲软提示(self):
        data = _base_match_data(initial_position={
            "initial_handicap": 0.5, "current_handicap": 0.5,
            "initial_home_water": 0.90, "initial_away_water": 0.95,
            "current_home_water": 0.88, "current_away_water": 0.97,
            "same_handicap": {"home": {"win": 2, "draw": 2, "lose": 2},
                              "away": {"win": 1, "draw": 2, "lose": 7}},
        })
        result = FundamentalAgent().analyze(data)["initial_position"]
        assert "下盘疲软" in result["same_handicap_hint"]

    def test_同盘路历史缺失时提示缺失且不编造(self):
        result = FundamentalAgent().analyze(_base_match_data())["initial_position"]
        assert result["same_handicap_hint"] is None
        assert "同盘路" in result["note"] or "缺失" in result["note"]

    def test_盘口变化方向按即时对初盘判定(self):
        data = _base_match_data(initial_position={
            "initial_handicap": 0.0, "current_handicap": 0.25,
            "initial_home_water": 0.90, "initial_away_water": 0.95,
            "current_home_water": 0.78, "current_away_water": 1.08,
            "same_handicap": None,
        })
        result = FundamentalAgent().analyze(data)["initial_position"]
        assert result["handicap_movement"] == "升盘"


class TestSuppression:
    def test_最近连续3胜判强压制(self):
        data = _base_match_data(h2h=_h2h([
            (2021, "D"), (2022, "L"), (2023, "W"), (2024, "W"), (2025, "W"), (2026, "W"),
        ]))
        result = FundamentalAgent().analyze(data)["h2h"]
        assert result["suppression"] == "strong"

    def test_近6次交锋胜率达70以上判强压制(self):
        data = _base_match_data(h2h=_h2h([
            (2021, "W"), (2022, "W"), (2023, "L"), (2024, "W"), (2025, "W"), (2026, "W"),
        ]))
        result = FundamentalAgent().analyze(data)["h2h"]
        assert result["suppression"] == "strong"

    def test_不败型战绩降级为弱压制不构成强压制(self):
        # v7.10 周六012 反例固化：3胜3平0负（胜率50%）→ 弱压制
        data = _base_match_data(h2h=_h2h([
            (2021, "W"), (2022, "D"), (2023, "W"), (2024, "D"), (2025, "W"), (2026, "D"),
        ]))
        result = FundamentalAgent().analyze(data)["h2h"]
        assert result["suppression"] == "weak"

    def test_普通交锋不输出压制结论(self):
        result = FundamentalAgent().analyze(_base_match_data())["h2h"]
        assert result["suppression"] == "none"

    def test_当期衰减旗标存在时强压制降级为弱压制(self):
        data = _base_match_data(h2h=_h2h([
            (2023, "W"), (2024, "W"), (2025, "W"), (2026, "W"),
        ], decay=["opponent_win_streak"]))
        result = FundamentalAgent().analyze(data)["h2h"]
        assert result["decayed"] is True
        assert result["suppression"] == "weak"

    def test_仅统计近5年内交锋(self):
        # 2020 年的旧压制不入窗口：近5年（≥2021）仅 L/L/D → 无压制
        data = _base_match_data(h2h=_h2h([
            (2020, "W"), (2020, "W"), (2020, "W"), (2025, "L"), (2026, "L"), (2026, "D"),
        ]))
        result = FundamentalAgent().analyze(data, current_year=2026)["h2h"]
        assert result["suppression"] == "none"
        assert result["window_records_count"] == 3
