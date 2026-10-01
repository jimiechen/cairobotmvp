# -*- coding: utf-8 -*-
"""FundamentalAgent 近期状态与修正条款测试（PRD-12 F-10，按场景拆分文件）。

案例锚点：进球外推 0.9 系数（v7.2 大小球主锚）、拉科场攻防主锚、弗赖堡场火力
保留（v7.2 折扣上限30%）、杯赛修正（v7.8）、国际赛修正（v7.11）。
"""
from fa_orchestrator.agents.fundamental_agent import FundamentalAgent

from test_fundamental_agent import _form, _schedule


def _form_agent():
    return FundamentalAgent()


def _data(recent_form, schedule=None):
    from test_fundamental_agent import _base_match_data
    return _base_match_data(recent_form=recent_form, schedule=schedule or _schedule())


class TestGoalExtrapolation:
    def test_进球外推按两队正赛场均进球和乘以09(self):
        form = _form(
            [(0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["home_avg_scored"] == 2.0
        assert result["away_avg_scored"] == 1.0
        assert round(result["expected_total_goals"], 2) == 2.7

    def test_外推只统计正赛排除热身赛(self):
        form = _form(
            [(0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1), (1, 5, 0)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (1, 5, 0)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["home_avg_scored"] == 2.0
        assert result["away_avg_scored"] == 1.0

    def test_输出两队近5场场均进球供火热判定(self):
        form = _form(
            [(0, 0, 2), (0, 1, 2), (0, 3, 1), (0, 3, 0), (0, 4, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["home_recent5_avg"] == 2.2
        assert result["away_recent5_avg"] == 1.0

    def test_近5场场均3球以上判火热并输出火力保留提示(self):
        # v7.2 弗赖堡场锚点：近5场场均4球被折扣是直接失误——火热方折扣上限30%
        form = _form(
            [(0, 4, 1), (0, 4, 1), (0, 4, 1), (0, 4, 1), (0, 4, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["hot_sides"] == ["home"]
        assert "30%" in result["fire_note"]

    def test_无火热一方时不输出火力保留提示(self):
        form = _form(
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["hot_sides"] == []
        assert result["fire_note"] == ""


class TestCompetitionCorrection:
    def test_联赛场无修正标注且外推为主锚(self):
        form = _form(
            [(0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(_data(form))["recent_form"]
        assert result["correction"] is None
        assert result["anchor_note"] == ""

    def test_杯赛修正输出数据主锚降级与共同主锚标注(self):
        # v7.8：杯赛场数据主锚降级为参考项，市场升线动作升为共同主锚
        form = _form(
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(
            _data(form, _schedule(competition="cup", level_gap=True))
        )["recent_form"]
        assert result["correction"] == "cup"
        assert "参考项" in result["anchor_note"]
        assert "共同主锚" in result["anchor_note"]

    def test_杯赛对攻崩坏形态标注不得降权(self):
        # v7.8：一方近5场场均总进球≥3.5 的"对攻崩坏形态"不得降权（004 场 4.8 球被降权是直接失误）
        form = _form(
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
            [(0, 4, 1), (0, 4, 2), (0, 4, 1), (0, 3, 2), (0, 4, 1)],
        )
        result = _form_agent().analyze(
            _data(form, _schedule(competition="cup", level_gap=True))
        )["recent_form"]
        assert result["away_recent5_avg"] == pytest_approx(3.8)
        assert "不得降权" in result["anchor_note"]

    def test_国际赛修正输出外推仅参考与轻仓弃配标注(self):
        # v7.11：国际赛 0.9 外推仅作参考项，大小球轻仓≤0.3 单位或弃配
        form = _form(
            [(0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1), (0, 2, 1)],
            [(0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1), (0, 1, 1)],
        )
        result = _form_agent().analyze(
            _data(form, _schedule(competition="international"))
        )["recent_form"]
        assert result["correction"] == "international"
        assert "参考" in result["anchor_note"]
        assert "0.3" in result["anchor_note"]


class TestLakataniAnchor:
    def test_拉科场锚点_外推只由攻防数据决定(self):
        # v7.2 拉科场（2026-08-30）：全市场降 2 球盘，实际打出 4 球——
        # "降盘=坚定信号"逻辑失效，外推只由近期攻防数据计算，盘口动作不影响基本面主锚
        form = _form(
            [(0, 2, 1), (0, 2, 0), (0, 3, 1), (0, 2, 1), (0, 3, 0)],
            [(0, 1, 2), (0, 2, 1), (0, 1, 1), (0, 2, 2), (0, 1, 1)],
        )
        data = _data(form)
        # 即便初盘定位显示降盘，基本面外推不受其影响
        data["initial_position"] = dict(
            data["initial_position"], initial_handicap=2.5, current_handicap=2.0,
        )
        result = _form_agent().analyze(data)["recent_form"]
        # 主队正赛场均 (2+2+3+2+3)/5=2.4，客队 (1+2+1+2+1)/5=1.4 → (2.4+1.4)*0.9=3.42
        assert round(result["expected_total_goals"], 2) == 3.42


def pytest_approx(value):
    """round 断言辅助（避免在本文件重复导入 pytest）。"""
    import pytest

    return pytest.approx(value, abs=0.01)
