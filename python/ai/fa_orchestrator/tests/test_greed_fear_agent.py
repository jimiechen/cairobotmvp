# -*- coding: utf-8 -*-
"""GreedFearAgent 测试（PRD-12 立项三：人性贪婪恐惧分析）。

覆盖：市场倾向判定、贪婪/恐惧各 4 信号评分档位、缺口信号权重重归一、
等级阈值（≥70 强/≥40 中/≥15 弱）、陷阱组合、投票缺失与平局热门降级。
纯规则计算，禁止真实 LLM 与外部服务（testing.md 红线）。
"""
from fa_orchestrator.agents.greed_fear_agent import GreedFearAgent


def _euro(rows):
    """构造欧赔行：rows 为 (bk, 初胜赔, 即胜赔, 初负赔, 即负赔) 元组列表。"""
    return [{"bk": bk,
             "initial": {"win": w0, "draw": 3.2, "lose": l0},
             "current": {"win": w1, "draw": 3.2, "lose": l1}}
            for bk, w0, w1, l0, l1 in rows]


def _asian(handicap_init=0.0, handicap_cur=0.0, home_water=0.90, away_water=0.92):
    """构造亚盘澳门行（单行：初盘/即时盘同公司）。"""
    return [{"bk": "澳门",
             "initial": {"home_water": 0.92, "handicap": handicap_init,
                         "away_water": 0.92},
             "current": {"home_water": home_water, "handicap": handicap_cur,
                         "away_water": away_water}}]


def _sentiment(popular="win", support=55.0, deviation=8.0):
    """构造热度结论（与 SentimentAgent.had 同构）。"""
    return {"had": {"popular": popular, "support": support,
                    "implied": 45.0, "deviation_pp": deviation, "note": ""},
            "overheat": {"triggered": deviation >= 20, "layer": "had",
                         "deviation_pp": deviation}}


# 缺省欧赔：胜/负方向均无升降（1 平 1 降 1 升 → 占比 1/3 < 0.4 记 0）
_EURO_NEUTRAL = _euro([("澳门", 2.10, 2.10, 3.10, 3.10),
                       ("威廉", 2.12, 2.11, 3.15, 3.15),
                       ("立博", 2.08, 2.12, 3.05, 3.05)])
# 胜方向 2/3 降赔（贪婪满分档）
_EURO_WIN_DROP = _euro([("澳门", 2.10, 2.05, 3.10, 3.10),
                        ("威廉", 2.12, 2.02, 3.15, 3.15),
                        ("立博", 2.08, 2.12, 3.05, 3.05)])


def _data(**overrides):
    """标准场基线：主队热门、平手盘、无贪婪/恐惧信号。"""
    data = {
        "sentiment": _sentiment(),
        "fundamental": {"intelligence": {"home_against": 0, "away_against": 0}},
        "odds_verify": {"euro_asia": {"two_layer_divergence": False}},
        "euro": _EURO_NEUTRAL,
        "asian": _asian(),
    }
    data.update(overrides)
    return data


class TestMarketTendency:
    def test_主队热门时弱势方判定为客队并透传支持率偏差(self):
        result = GreedFearAgent().analyze(_data())
        assert result["phase"] == "greed_fear"
        assert result["market_tendency"] == {
            "popular": "win", "unpopular": "lose",
            "support": 55.0, "deviation_pp": 8.0}

    def test_客队热门时弱势方判定为主队(self):
        result = GreedFearAgent().analyze(
            _data(sentiment=_sentiment(popular="lose", support=48.0)))
        assert result["market_tendency"]["popular"] == "lose"
        assert result["market_tendency"]["unpopular"] == "win"

    def test_投票缺失整体降级不编造评分(self):
        result = GreedFearAgent().analyze(_data(sentiment={
            "had": {"popular": None, "support": None, "implied": None,
                    "deviation_pp": None, "note": "热度数据缺失"}}))
        assert result["market_tendency"]["popular"] is None
        assert result["greed"]["score"] is None
        assert result["fear"]["level"] == "无数据"
        assert "缺失" in result["note"]

    def test_平局热门判市场均势贪婪恐惧降级(self):
        # 支持率热门为平局 → 无明确优势方/弱势方，贪婪恐惧不适用
        result = GreedFearAgent().analyze(
            _data(sentiment=_sentiment(popular="draw")))
        assert result["greed"]["score"] is None
        assert result["fear"]["score"] is None
        assert "均势" in result["note"]


class TestGreedSignals:
    def test_热门方向降赔家数占比过半记贪婪赔率变化满分(self):
        result = GreedFearAgent().analyze(_data(euro=_EURO_WIN_DROP))
        assert result["greed"]["signals"]["odds_change"]["score"] == 100

    def test_热门方向降赔家数占比四成记半分(self):
        # 5 家中 2 家降赔（占比 0.4）→ 半分档
        euro = _euro([("澳门", 2.10, 2.05, 3.10, 3.10),
                      ("威廉", 2.12, 2.02, 3.15, 3.15),
                      ("立博", 2.08, 2.08, 3.05, 3.05),
                      ("竞彩", 2.15, 2.15, 3.20, 3.20),
                      ("365", 2.09, 2.09, 3.12, 3.12)])
        result = GreedFearAgent().analyze(_data(euro=euro))
        assert result["greed"]["signals"]["odds_change"]["score"] == 50

    def test_升盘朝热门方向记贪婪盘口变化满分(self):
        # 主队热门 + 让球线 0 → -0.5（主让加深）→ 贪婪
        result = GreedFearAgent().analyze(_data(asian=_asian(0.0, -0.5)))
        assert result["greed"]["signals"]["handicap_change"]["score"] == 100

    def test_热门方低水记水位诱惑满分(self):
        result = GreedFearAgent().analyze(
            _data(asian=_asian(home_water=0.78)))
        assert result["greed"]["signals"]["water_lure"]["score"] == 100

    def test_支持率过热偏差记市场背离满分(self):
        result = GreedFearAgent().analyze(
            _data(sentiment=_sentiment(deviation=22.0)))
        assert result["greed"]["signals"]["market_divergence"]["score"] == 100

    def test_两层背离记市场背离半分(self):
        result = GreedFearAgent().analyze(_data(
            sentiment=_sentiment(deviation=5.0),
            odds_verify={"euro_asia": {"two_layer_divergence": True}}))
        assert result["greed"]["signals"]["market_divergence"]["score"] == 50


class TestFearSignals:
    def test_深盘朝热门方向记门槛提升满分(self):
        # 主队热门 + 深盘 -1.25 → 弱势方（客队）门槛提升
        result = GreedFearAgent().analyze(_data(asian=_asian(0.0, -1.25)))
        assert result["fear"]["signals"]["gate_raise"]["score"] == 100

    def test_浅盘不记门槛提升(self):
        result = GreedFearAgent().analyze(_data(asian=_asian(0.0, -0.5)))
        assert result["fear"]["signals"]["gate_raise"]["score"] == 0

    def test_弱势方向升赔家数占比过半记赔率压制满分(self):
        # 客队为弱势方：5 家中 3 家客胜升赔（占比 0.6）
        euro = _euro([("澳门", 2.10, 2.05, 3.10, 3.10),
                      ("威廉", 2.12, 2.12, 3.15, 3.30),
                      ("立博", 2.08, 2.08, 3.05, 3.25),
                      ("竞彩", 2.15, 2.15, 3.20, 3.40),
                      ("365", 2.09, 2.09, 3.12, 3.12)])
        result = GreedFearAgent().analyze(_data(euro=euro))
        assert result["fear"]["signals"]["odds_suppress"]["score"] == 100

    def test_弱势方高水记水位压制满分(self):
        result = GreedFearAgent().analyze(
            _data(asian=_asian(away_water=1.02)))
        assert result["fear"]["signals"]["water_suppress"]["score"] == 100

    def test_弱势方利空情报两条记基本面恐惧满分(self):
        result = GreedFearAgent().analyze(_data(
            fundamental={"intelligence": {"home_against": 0, "away_against": 2}}))
        assert result["fear"]["signals"]["fundamental_fear"]["score"] == 100

    def test_弱势方利空情报一条记半分(self):
        result = GreedFearAgent().analyze(_data(
            fundamental={"intelligence": {"home_against": 0, "away_against": 1}}))
        assert result["fear"]["signals"]["fundamental_fear"]["score"] == 50


class TestScoreAndLevel:
    def test_缺口信号权重重归一不按零分摊薄(self):
        # 无亚盘 → 盘口/水位信号缺口，仅按剩余信号权重加权：
        # 赔率变化 100×30、背离 0×20 → 100×30/(30+20) = 60
        result = GreedFearAgent().analyze(_data(euro=_EURO_WIN_DROP, asian=[]))
        greed = result["greed"]
        assert greed["signals"]["handicap_change"]["score"] is None
        assert greed["signals"]["water_lure"]["score"] is None
        assert greed["score"] == 60.0

    def test_等级阈值强中弱无(self):
        # 贪婪：降赔满分 + 低水满分 + 过热满分 → 75 强；恐惧：全零 → 无
        result = GreedFearAgent().analyze(_data(
            euro=_EURO_WIN_DROP, asian=_asian(home_water=0.78),
            sentiment=_sentiment(deviation=22.0)))
        assert result["greed"]["level"] == "强"
        assert result["fear"]["level"] == "无"
        assert result["trap_combination"] == "强贪婪+无恐惧"

    def test_中等评分记中等级(self):
        # 赔率 50×30 + 盘口 0 + 水位 50×25 + 背离 100×20 → 47.5 → 中
        euro = _euro([("澳门", 2.10, 2.05, 3.10, 3.10),
                      ("威廉", 2.12, 2.02, 3.15, 3.15),
                      ("立博", 2.08, 2.08, 3.05, 3.05),
                      ("竞彩", 2.15, 2.15, 3.20, 3.20),
                      ("365", 2.09, 2.09, 3.12, 3.12)])
        result = GreedFearAgent().analyze(_data(
            euro=euro, asian=_asian(home_water=0.86),
            sentiment=_sentiment(deviation=22.0)))
        assert result["greed"]["score"] == 47.5
        assert result["greed"]["level"] == "中"

    def test_评分低于阈值记无陷阱(self):
        result = GreedFearAgent().analyze(_data())
        assert result["greed"]["level"] == "无"
        assert result["fear"]["level"] == "无"
        assert result["trap_combination"] == "无贪婪+无恐惧"

    def test_全部信号缺口时输出无数据(self):
        result = GreedFearAgent().analyze(_data(
            euro=[], asian=[],
            sentiment={"had": {"popular": "win", "support": 50.0,
                               "implied": None, "deviation_pp": None, "note": ""},
                       "overheat": {"triggered": False}},
            fundamental={"intelligence": {}}))
        assert result["greed"]["score"] is None
        assert result["fear"]["level"] == "无数据"
        assert result["trap_combination"] == "无数据"
