# -*- coding: utf-8 -*-
"""SentimentAgent 单元测试（PRD-12 F-12，ISS-FA3）。

覆盖：HAD/HHAD 支持率偏差计算、过热判定（胜平负 ≥+20pp / 让球 ≥+15pp）边界、
必发成交形态输入与缺失降级、过热两档处置（有独立支撑降星保方向 / 无支撑放弃）。
纯规则计算，禁止真实 LLM 与外部服务（SKILL v7.9 过热场处置口径）。
"""
import pytest

from fa_orchestrator.agents.sentiment_agent import INTEL_SUPPORT_MIN, SentimentAgent


# 两家对称欧赔：1/2.0+1/4.0+1/4.0=1.0，去水后主胜隐含概率恰为 50.0%
_EURO_50 = [
    {"current": {"win": 2.0, "draw": 4.0, "lose": 4.0}},
    {"current": {"win": 2.0, "draw": 4.0, "lose": 4.0}},
]


def _fundamental(intel_items=(), h2h_results=(), recent_form=None):
    """构造基本面结论（与 FA2 AnalysisRunner 输出同构的最小形态）。"""
    return {
        "intelligence": {"items": list(intel_items)},
        "h2h": {"records": [{"result": r} for r in h2h_results]},
        "recent_form": recent_form or {},
    }


def _make_input(**overrides):
    """构造 SentimentAgent 标准输入：主胜支持率 72% vs 隐含 50% → +22pp 过热。"""
    base = {
        "vote": {"support": {"win": 72.0, "draw": 14.0, "lose": 14.0}},
        "euro": _EURO_50,
        "fundamental": _fundamental(),
    }
    base.update(overrides)
    return base


class TestHadDeviation:
    def test_支持率偏差为支持率减去水隐含概率(self):
        result = SentimentAgent().analyze(_make_input())
        assert result["phase"] == "sentiment"
        assert result["had"]["popular"] == "win"
        assert result["had"]["deviation_pp"] == pytest.approx(22.0)

    def test_无投票数据时热度验证降级且不编造(self):
        result = SentimentAgent().analyze(_make_input(vote=None))
        assert result["had"]["popular"] is None
        assert result["had"]["deviation_pp"] is None
        assert "热度数据缺失" in result["note"]


class TestOverheatThreshold:
    def test_胜平负偏差达20pp判过热(self):
        result = SentimentAgent().analyze(_make_input())
        assert result["overheat"]["triggered"] is True
        assert result["overheat"]["layer"] == "had"
        assert result["overheat"]["deviation_pp"] == pytest.approx(22.0)

    def test_胜平负偏差19点9pp不过热_边界(self):
        vote = {"support": {"win": 69.9, "draw": 15.0, "lose": 15.1}}
        result = SentimentAgent().analyze(_make_input(vote=vote))
        assert result["overheat"]["triggered"] is False

    def test_让球偏差15pp判让球过热(self):
        vote = {
            "support": {"win": 40.0, "draw": 30.0, "lose": 30.0},
            "hhand": {
                "line": -1.0,
                "odds": {"win": 2.0, "draw": 4.0, "lose": 4.0},
                "support": {"win": 65.0, "draw": 17.5, "lose": 17.5},
            },
        }
        result = SentimentAgent().analyze(_make_input(vote=vote))
        assert result["hhand"]["available"] is True
        assert result["hhand"]["deviation_pp"] == pytest.approx(15.0)
        assert result["overheat"]["triggered"] is True
        assert result["overheat"]["layer"] == "hhand"

    def test_让球投票缺失标注降级不编造(self):
        result = SentimentAgent().analyze(_make_input(vote={"support": {"win": 50.0, "draw": 25.0, "lose": 25.0}}))
        assert result["hhand"]["available"] is False
        assert "让球热度数据缺失" in result["hhand"]["note"]


class TestBetfairInput:
    def test_必发成交形态输入被记录(self):
        bf = {"total_volume": 120000.0, "direction": "home", "trend": "falling"}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert result["bf"]["available"] is True
        assert result["bf"]["trend"] == "falling"

    def test_必发数据缺失标注缺失(self):
        result = SentimentAgent().analyze(_make_input())
        assert result["bf"]["available"] is False
        assert "必发" in result["bf"]["summary"]

    def test_庄家pnl与价位聚集透传且summary含描述(self):
        # 立项四：pnl/价位聚集为输入记录（skill 未定义判定口径，语义解读待主控确认）
        bf = {"total_volume": 120000.0, "direction": "home", "trend": "falling",
              "bookmaker_pnl": -123.45,
              "ladder": {"top_price": 2.0, "top_ratio": 0.7273, "levels": 2}}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert result["bf"]["bookmaker_pnl"] == -123.45
        assert result["bf"]["ladder"]["top_ratio"] == 0.7273
        assert "庄家pnl" in result["bf"]["summary"]

    def test_pnl缺失时summary不含pnl片段(self):
        bf = {"total_volume": 120000.0, "direction": "home", "trend": "rising"}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert result["bf"]["bookmaker_pnl"] is None
        assert "庄家pnl" not in result["bf"]["summary"]

    # ---- 立项四收口：口径经主控确认（2026-10-14），交叉验证标注 + 聚集度三档 ----

    def test_pnl为负时标注庄家赔付风险为大众方向赢面信号(self):
        bf = {"direction": "home", "trend": "falling", "bookmaker_pnl": -123.45}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert "大众方向赢面倾向" in result["bf"]["pnl_note"]
        assert "庄家pnl -123.45" in result["bf"]["summary"]

    def test_pnl非负时标注庄家风险可控(self):
        bf = {"direction": "home", "trend": "rising", "bookmaker_pnl": 50.0}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert "庄家风险可控" in result["bf"]["pnl_note"]

    def test_pnl缺失时pnl_note为None(self):
        result = SentimentAgent().analyze(_make_input(vote=_make_input()["vote"]))
        assert result["bf"]["pnl_note"] is None

    def test_聚集度达高档阈值标注高聚集(self):
        bf = {"direction": "home", "trend": "rising",
              "ladder": {"top_price": 2.0, "top_ratio": 0.6, "levels": 3}}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert "高聚集" in result["bf"]["ladder_note"]

    def test_聚集度中等档标注(self):
        bf = {"direction": "home", "trend": "rising",
              "ladder": {"top_price": 2.0, "top_ratio": 0.4, "levels": 3}}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert "中等" in result["bf"]["ladder_note"]

    def test_聚集度分散档标注大众观点不一致(self):
        bf = {"direction": "home", "trend": "rising",
              "ladder": {"top_price": 2.0, "top_ratio": 0.29, "levels": 5}}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        assert "分散" in result["bf"]["ladder_note"]
        assert "大众观点不一致" in result["bf"]["ladder_note"]

    def test_ladder缺失时ladder_note为None(self):
        result = SentimentAgent().analyze(_make_input(vote=_make_input()["vote"]))
        assert result["bf"]["ladder_note"] is None

    # ---- 立项四扩展（宿主答复 2026-10-14）：heat pnl（500com 投注倾向
    # 庄家盈亏，正=庄家赢）作为庄家 pnl 现实数据源先行接入，口径同 pnl_note
    # 体系（负=赔付风险=大众该方向赢面倾向信号）；仅输入记录，不进加减分 ----

    def test_热度盈亏三方向标注庄家解读与来源(self):
        bf = {"direction": "home", "trend": "falling",
              "heat_pnl": {"home": 3712.0, "draw": 33329.0, "away": -50808.0,
                           "snapshot_time": "2026-10-10 10:27:59"}}
        result = SentimentAgent().analyze(_make_input(vote={**_make_input()["vote"], "bf": bf}))
        note = result["bf"]["heat_pnl_note"]
        assert note is not None
        assert "500com" in note                      # 来源标注，区分必发 pnl
        assert "主 3712" in note and "庄家赢" in note
        assert "客 -50808" in note and "大众该方向赢面倾向" in note
        assert "平 33329" in note
        assert "500com" in result["bf"]["summary"]   # summary 追加

    def test_热度盈亏缺失时note为None且summary不含片段(self):
        result = SentimentAgent().analyze(_make_input(vote=_make_input()["vote"]))
        assert result["bf"]["heat_pnl_note"] is None
        assert "500com" not in result["bf"]["summary"]


class TestOverheatDisposition:
    def test_独立利多情报达标时降星保方向(self):
        items = [
            {"party": "home", "sentiment": "favor", "text": "主力前锋复出"},
            {"party": "home", "sentiment": "favor", "text": "主场近5战4胜"},
        ]
        result = SentimentAgent().analyze(_make_input(fundamental=_fundamental(intel_items=items)))
        assert result["overheat"]["support_intel"] is True
        assert result["overheat"]["disposition"] == "downgrade_keep"
        assert "情报支撑：有" in result["overheat"]["tag"]
        assert "22" in result["overheat"]["tag"]

    def test_无独立支撑时主推放弃(self):
        result = SentimentAgent().analyze(_make_input())
        assert result["overheat"]["support_intel"] is False
        assert result["overheat"]["disposition"] == "abandon"
        assert "情报支撑：无" in result["overheat"]["tag"]

    def test_利多情报少于阈值不计独立支撑(self):
        items = [{"party": "home", "sentiment": "favor", "text": "仅一条利多"}]
        result = SentimentAgent().analyze(_make_input(fundamental=_fundamental(intel_items=items)))
        assert result["overheat"]["support_intel"] is False
        assert INTEL_SUPPORT_MIN == 2

    def test_交锋同向佐证计独立支撑(self):
        result = SentimentAgent().analyze(
            _make_input(fundamental=_fundamental(h2h_results=("W", "W", "D")))
        )
        assert result["overheat"]["support_intel"] is True

    def test_未过热时处置与标注为空(self):
        vote = {"support": {"win": 50.0, "draw": 25.0, "lose": 25.0}}
        result = SentimentAgent().analyze(_make_input(vote=vote))
        assert result["overheat"]["triggered"] is False
        assert result["overheat"]["disposition"] is None
        assert result["overheat"]["tag"] is None
