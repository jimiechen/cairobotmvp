# -*- coding: utf-8 -*-
"""历史推演（F-16）测试：升降盘时序与热度漂移推导，MySQL 专属数据缺失时明确标注。

依据：SKILL v7.4 M5 能力（tracks：change>0 主让加深/大球升线；drift 单位百分点）；
PRD-12 F-16：缺失时明确标注"无法推演时序"。全程无外部服务调用。
"""
from fa_orchestrator.analysis.history_deduction import (
    TREND_UNKNOWN,
    deduce_heat_drift,
    deduce_trend,
)


class TestDeduceTrend:
    def test_轨迹缺失时标注无法推演时序(self):
        result = deduce_trend(None)
        assert result["available"] is False
        assert "无法推演时序" in result["note"]

    def test_轨迹为空列表时同样标注无法推演(self):
        result = deduce_trend([])
        assert result["available"] is False
        assert "无法推演时序" in result["note"]

    def test_正变化按类型分别计数主让加深与大球升线(self):
        tracks = [
            {"kind": "asian", "change": 1},
            {"kind": "asian", "change": 1},
            {"kind": "overunder", "change": 1},
            {"kind": "overunder", "change": 0},
        ]
        result = deduce_trend(tracks)
        assert result["available"] is True
        assert result["asian_deepen_count"] == 2
        assert result["overunder_raise_count"] == 1
        assert result["total"] == 4

    def test_负变化与零变化不计入加深或升线(self):
        tracks = [
            {"kind": "asian", "change": -1},
            {"kind": "overunder", "change": -1},
            {"kind": "asian", "change": 0},
        ]
        result = deduce_trend(tracks)
        assert result["available"] is True
        assert result["asian_deepen_count"] == 0
        assert result["overunder_raise_count"] == 0


class TestDeduceHeatDrift:
    def test_漂移缺失时标注无法推演时序(self):
        result = deduce_heat_drift(None)
        assert result["available"] is False
        assert "无法推演时序" in result["note"]

    def test_漂移正常返回百分点与热度走向(self):
        result = deduce_heat_drift({"home_pp": 4.2, "away_pp": -4.2})
        assert result["available"] is True
        assert result["home_pp"] == 4.2
        assert result["trend"] == "home_heat_rising"

    def test_漂移回落至平稳带判平稳(self):
        result = deduce_heat_drift({"home_pp": 0.5, "away_pp": -0.5})
        assert result["trend"] == "stable"

    def test_漂移大幅下行判热度降温(self):
        result = deduce_heat_drift({"home_pp": -3.0, "away_pp": 3.0})
        assert result["trend"] == "home_heat_cooling"

    def test_热度维度缺失时按缺口标注不崩溃(self):
        # BUG-FA-103：BUG-FA-102 引入的 None 语义（该侧无有效热度快照）到达消费端，
        # 走向按缺口标注，禁止 TypeError、禁止编造（真实库 2026-10-08 周四001 触发）
        result = deduce_heat_drift({"home_pp": None, "away_pp": None})
        assert result["available"] is True   # 数据块存在，仅热度维度缺失
        assert result["trend"] == TREND_UNKNOWN
        assert "热度" in result["note"]

    def test_主侧热度缺失时走向按缺口标注且有效侧透传(self):
        result = deduce_heat_drift({"home_pp": None, "away_pp": 5.0})
        assert result["trend"] == TREND_UNKNOWN
        assert result["away_pp"] == 5.0

    def test_仅客侧热度缺失时走向仍按主侧判定(self):
        # 走向按主队侧分类（v7.4 M5 口径）：客侧缺失降级标注但不影响主侧走向
        result = deduce_heat_drift({"home_pp": 4.2, "away_pp": None})
        assert result["trend"] == "home_heat_rising"
        assert "热度" in result["note"]
