# -*- coding: utf-8 -*-
"""OddsVerifyAgent 测试（PRD-12 F-11）：v7.12 第二阶段规则节点化。

案例锚点回归（v7.2/v7.11）：拉科场升盘低水真信号、深盘背书（09-27 004 丹麦场）、
均势盘需验证（008 挪威场）、诱盘边界、两层背离信让球层、防冷分级、立博独开分流、
单家临场背离。全程纯规则计算，不调用 LLM 与外部服务。
"""
from fa_orchestrator.agents.odds_verify_agent import (
    NEED_STRENGTH_VERIFY,
    TAG_ANTI_COLD_DOWNGRADE,
    TAG_ANTI_COLD_VALID,
    TAG_SINGLE_HOUSE_TEMPLATE,
    TAG_SOLO_DEPTH,
    TAG_SOLO_VALUE,
    TAG_TWO_LAYER,
    TRUE_SIGNAL,
    TRAP_SIGNAL,
    OddsVerifyAgent,
)


def _euro(bk, iw, cw, final=False, idr=3.20, cd=3.25, il=3.50, cl=3.55, payout=0.91):
    return {"bk": bk, "initial": {"win": iw, "draw": idr, "lose": il},
            "current": {"win": cw, "draw": cd, "lose": cl}, "payout": payout,
            "is_final_snapshot": final}


def _asian(bk, ih, ihc, ia, ch, chc, ca, final=False):
    return {"bk": bk, "initial": {"home_water": ih, "handicap": ihc, "away_water": ia},
            "current": {"home_water": ch, "handicap": chc, "away_water": ca},
            "is_final_snapshot": final}


def _data(euro, asian, **overrides):
    data = {
        "euro": euro, "asian": asian, "overunder": [],
        "state_baseline": {"home_avg_goals": 1.8},
        "strength": {"value_ratio": None, "level_gap": False, "h2h_suppression": "none"},
        "draw_support": False,
    }
    data.update(overrides)
    return data


def _agent():
    return OddsVerifyAgent()


class TestStructure:
    def test_主流公司覆盖检查(self):
        result = _agent().verify(_data(
            [_euro("澳门", 2.10, 2.02), _euro("威廉希尔", 2.12, 2.04)],
            [_asian("澳门", 0.90, 0.0, 0.95, 0.88, 0.0, 0.97)],
        ))["structure"]
        assert result["company_count"] == 2
        assert set(result["mainstream_missing"]) == {"立博", "Bet365"}

    def test_主流公司齐全时无缺失告警(self):
        euro = [_euro(bk, 2.10, 2.02) for bk in ("澳门", "立博", "威廉希尔", "Bet365")]
        result = _agent().verify(_data(euro, []))["structure"]
        assert result["mainstream_missing"] == []


class TestStructureMarketDirection:
    """structure 增强（立项二①）：全市场升降赔家数 + 基准去水概率 + 凯利离群。"""

    def test_主胜升降赔家数统计(self):
        # skill"30 家全升/29 家降"口径：主胜赔率初→即方向家数
        euro = [
            _euro("澳门", 2.10, 2.20),        # 升
            _euro("立博", 2.10, 2.00),        # 降
            _euro("威廉希尔", 2.10, 2.10),    # 平
            _euro("Bet365", 2.10, 2.30),      # 升
        ]
        result = _agent().verify(_data(euro, []))["structure"]
        assert result["market_direction"] == {"win_rise": 2, "win_drop": 1, "win_flat": 1}

    def test_基准去水概率以立博为准(self):
        euro = [
            _euro("澳门", 2.10, 2.02, cd=3.30, cl=3.60),
            _euro("立博", 2.10, 2.00, cd=3.20, cl=3.40),
        ]
        result = _agent().verify(_data(euro, []))["structure"]
        probs = result["baseline_probs"]
        assert probs["bk"] == "立博"
        # 立博 current：1/2.00+1/3.20+1/3.40 归一
        assert probs["win"] == round((1 / 2.00) / (1 / 2.00 + 1 / 3.20 + 1 / 3.40), 3)
        assert probs["draw"] == round((1 / 3.20) / (1 / 2.00 + 1 / 3.20 + 1 / 3.40), 3)
        assert probs["lose"] == round((1 / 3.40) / (1 / 2.00 + 1 / 3.20 + 1 / 3.40), 3)

    def test_无立博时基准回退主流公司(self):
        euro = [_euro("澳门", 2.10, 2.02), _euro("威廉希尔", 2.12, 2.04)]
        result = _agent().verify(_data(euro, []))["structure"]
        assert result["baseline_probs"]["bk"] == "威廉希尔"

    def test_凯利离群检测按最大值输出(self):
        # 立博基准主胜概率 p；澳门主胜 kelly = 2.02×p；构造某公司 >1.0 的引诱形态
        euro = [
            _euro("立博", 2.10, 2.00, cd=3.20, cl=3.40),
            _euro("澳门", 2.10, 2.05, cd=3.30, cl=3.60),
            _euro("诱盘公司", 2.10, 2.60, cd=3.10, cl=3.20),
        ]
        result = _agent().verify(_data(euro, []))["structure"]
        outlier = result["kelly_outlier"]
        assert outlier["present"] is True
        assert outlier["bk"] == "诱盘公司"
        assert outlier["result"] == "win"
        assert outlier["count_over"] == 1
        assert outlier["max"] > 1.0

    def test_无凯利离群时如实输出(self):
        euro = [_euro("立博", 2.10, 2.00), _euro("澳门", 2.10, 2.02)]
        result = _agent().verify(_data(euro, []))["structure"]
        outlier = result["kelly_outlier"]
        assert outlier["present"] is False
        assert outlier["count_over"] == 0

    def test_无即时盘时结构降级不崩溃(self):
        euro = [{"bk": "澳门", "initial": {"win": 2.1, "draw": 3.25, "lose": 3.5},
                 "current": None, "payout": None, "is_final_snapshot": False}]
        result = _agent().verify(_data(euro, []))["structure"]
        assert result["market_direction"] == {"win_rise": 0, "win_drop": 0, "win_flat": 0}
        assert result["baseline_probs"] is None
        assert result["kelly_outlier"]["present"] is False


class TestRaiseInterpretation:
    def test_拉科场锚点_孤立升盘但低水按真信号处理(self):
        # v7.2 拉科场（2026-08-30）：Bet365 平半 0.770 低水→3-1 主胜；v7.1"孤立升盘=诱盘"已废止
        asian = [
            _asian("Bet365", 0.85, 0.0, 1.01, 0.77, 0.25, 1.09),
            _asian("澳门", 0.90, 0.0, 0.95, 0.90, 0.0, 0.95),
        ]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["euro_asia"]
        assert result["raise_count"] == 1
        assert result["raise_form"] == "孤立升盘"
        assert result["interpretation"] == TRUE_SIGNAL

    def test_连续共振升盘且状态同向判真信号(self):
        # 弗赖堡场锚点：连续激进升盘+状态数据同向（场均4球）→ 真信号
        asian = [
            _asian("澳门", 0.88, 0.25, 0.98, 0.80, 0.75, 1.06),
            _asian("Bet365", 0.87, 0.25, 0.99, 0.79, 0.75, 1.07),
        ]
        euro = [_euro("澳门", 1.85, 1.82)]
        data = _data(euro, asian, state_baseline={"home_avg_goals": 4.0})
        result = _agent().verify(data)["euro_asia"]
        assert result["raise_count"] == 2
        assert result["raise_form"] == "共振升盘"
        assert result["interpretation"] == TRUE_SIGNAL

    def test_深盘加实力背书时孤立升盘高水按真信号处理(self):
        # v7.11 升盘真信号边界（09-27 004 丹麦场）：深盘+身价差≥2倍+历史强压制 → 真信号
        asian = [_asian("Bet365", 0.85, 0.5, 1.01, 1.05, 0.75, 0.80)]
        euro = [_euro("澳门", 1.85, 1.82), _euro("立博", 1.88, 1.85)]
        data = _data(euro, asian, strength={
            "value_ratio": 2.35, "level_gap": False, "h2h_suppression": "strong",
        })
        result = _agent().verify(data)["euro_asia"]
        assert result["deep_market"] is True
        assert result["strength_endorsed"] is True
        assert result["interpretation"] == TRUE_SIGNAL

    def test_均势盘无实力背书时升盘信号需验证禁止重仓(self):
        # v7.11（09-27 008 挪威场）：均势盘（主胜≥2.0）升盘信号天然弱，必须叠加实力差验证
        asian = [_asian("Bet365", 0.85, 0.0, 1.01, 1.05, 0.25, 0.80)]
        euro = [_euro("澳门", 2.60, 2.58)]
        data = _data(euro, asian, state_baseline={"home_avg_goals": 1.6})
        result = _agent().verify(data)["euro_asia"]
        assert result["deep_market"] is False
        assert result["interpretation"] == NEED_STRENGTH_VERIFY

    def test_升盘高水且状态背离判诱盘(self):
        # v7.2：升盘伴随高水 + 升盘方向与球队近期状态数据背离（场均<1）→ 诱盘
        asian = [_asian("Bet365", 0.85, 0.0, 1.01, 1.05, 0.25, 0.80)]
        euro = [_euro("澳门", 2.20, 2.18)]
        data = _data(euro, asian, state_baseline={"home_avg_goals": 0.7})
        result = _agent().verify(data)["euro_asia"]
        assert result["state_alignment"] == "背离"
        assert result["interpretation"] == TRAP_SIGNAL

    def test_无升盘时不输出解读(self):
        asian = [_asian("澳门", 0.90, 0.0, 0.95, 0.88, 0.0, 0.97)]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["euro_asia"]
        assert result["raise_count"] == 0
        assert result["interpretation"] is None


class TestTwoLayerDivergence:
    def test_超低水与升盘极端高水并存时标注两层背离信让球层(self):
        # v7.11（09-27 008 挪威场固化）：胜平负层超低水资金信号与让球层升盘挂
        # 极端超高水无人承接同时出现 → 优先信让球层机构定价
        asian = [
            _asian("威廉希尔", 0.72, 0.0, 1.14, 0.72, 0.0, 1.14),
            _asian("澳门", 0.88, 0.0, 0.98, 1.18, 0.25, 0.68),
        ]
        euro = [_euro("澳门", 1.85, 1.82)]
        result = _agent().verify(_data(euro, asian))["euro_asia"]
        assert result["two_layer_divergence"] is True
        assert TAG_TWO_LAYER in result["tags"]

    def test_普通低水升盘不触发两层背离(self):
        asian = [
            _asian("威廉希尔", 0.85, 0.0, 1.01, 0.80, 0.25, 1.06),
            _asian("澳门", 0.90, 0.0, 0.95, 0.82, 0.25, 1.05),
        ]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["euro_asia"]
        assert result["two_layer_divergence"] is False


class TestAntiCold:
    def test_同级且平局有支持且多家压平判防冷有效(self):
        euro = [
            _euro("澳门", 2.05, 2.02, idr=3.30, cd=3.05),
            _euro("立博", 2.08, 2.05, idr=3.28, cd=3.02),
        ]
        data = _data(euro, [], draw_support=True)
        result = _agent().verify(data)["anomalies"]["anti_cold"]
        assert result["present"] is True
        assert result["tag"] == TAG_ANTI_COLD_VALID

    def test_级别断层时压平动作降级为利润平衡(self):
        # v7.2 奥格斯堡 3-0 锚点：存在联赛级别断层 → 禁止作为反买依据
        euro = [
            _euro("澳门", 2.05, 2.02, idr=3.30, cd=3.05),
            _euro("立博", 2.08, 2.05, idr=3.28, cd=3.02),
        ]
        data = _data(euro, [], draw_support=True,
                     strength={"value_ratio": None, "level_gap": True, "h2h_suppression": "none"})
        result = _agent().verify(data)["anomalies"]["anti_cold"]
        assert result["present"] is True
        assert result["tag"] == TAG_ANTI_COLD_DOWNGRADE
        assert any("级别断层" in r for r in result["reasons"])

    def test_平局方向无基本面支持时同样降级(self):
        euro = [
            _euro("澳门", 2.05, 2.02, idr=3.30, cd=3.05),
            _euro("立博", 2.08, 2.05, idr=3.28, cd=3.02),
        ]
        result = _agent().verify(_data(euro, [], draw_support=False))["anomalies"]["anti_cold"]
        assert result["tag"] == TAG_ANTI_COLD_DOWNGRADE

    def test_无压平动作时不触发防冷检查(self):
        euro = [_euro("澳门", 2.05, 2.02)]
        result = _agent().verify(_data(euro, [], draw_support=True))["anomalies"]["anti_cold"]
        assert result["present"] is False


class TestLadbrokesSolo:
    def test_立博独开与热门同向但盘位更浅标注降深度(self):
        # v7.9（09-25 007 亚美尼亚场）：独开与主流同向但盘位更浅 → 砍主推深度
        asian = [
            _asian("澳门", 0.90, 0.25, 0.95, 0.88, 0.5, 0.97),
            _asian("皇冠", 0.91, 0.25, 0.94, 0.89, 0.5, 0.96),
            _asian("Bet365", 0.88, 0.25, 0.98, 0.87, 0.5, 0.99),
            _asian("立博", 0.92, 0.25, 0.93, 0.90, 0.25, 0.95),
        ]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["anomalies"]["ladbrokes_solo"]
        assert result["present"] is True
        assert result["tag"] == TAG_SOLO_DEPTH

    def test_立博独开直接指向受让方标注价值方向(self):
        # v7.9（09-25 008 意大利场）：独开指向受让/冷门方 → 走让球三条件判定后可主推
        asian = [
            _asian("澳门", 0.90, 0.25, 0.95, 0.88, 0.5, 0.97),
            _asian("皇冠", 0.91, 0.25, 0.94, 0.89, 0.5, 0.96),
            _asian("Bet365", 0.88, 0.25, 0.98, 0.87, 0.5, 0.99),
            _asian("立博", 1.00, -0.25, 0.85, 0.98, -0.25, 0.87),
        ]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["anomalies"]["ladbrokes_solo"]
        assert result["present"] is True
        assert result["tag"] == TAG_SOLO_VALUE

    def test_立博与主流一致时不触发独开信号(self):
        asian = [
            _asian("澳门", 0.90, 0.25, 0.95, 0.88, 0.5, 0.97),
            _asian("立博", 0.91, 0.25, 0.94, 0.89, 0.5, 0.96),
        ]
        result = _agent().verify(_data([_euro("澳门", 2.10, 2.05)], asian))["anomalies"]["ladbrokes_solo"]
        assert result["present"] is False


class TestSingleHouseDeviation:
    def test_单家主流临场逆势抬升超5个点强制标注降星减半仓(self):
        # v7.11（09-27 008 挪威场）：B365 主胜 2.25→2.40（+6.7%）被判"单家试盘"是直接失误
        euro = [
            _euro("澳门", 2.30, 2.29),
            _euro("立博", 2.33, 2.32),
            _euro("Bet365", 2.25, 2.40, final=True),
        ]
        result = _agent().verify(_data(euro, []))["anomalies"]["single_house_deviation"]
        assert result["present"] is True
        assert result["bk"] == "Bet365"
        assert result["pct"] == 6.7
        assert result["tag"] == TAG_SINGLE_HOUSE_TEMPLATE.format(pct=6.7)

    def test_临场抬升不足5个点不触发(self):
        euro = [_euro("Bet365", 2.00, 2.09, final=True)]
        result = _agent().verify(_data(euro, []))["anomalies"]["single_house_deviation"]
        assert result["present"] is False

    def test_涨幅恰为5个点边界触发(self):
        euro = [_euro("Bet365", 2.00, 2.10, final=True)]
        result = _agent().verify(_data(euro, []))["anomalies"]["single_house_deviation"]
        assert result["present"] is True
        assert result["pct"] == 5.0

    def test_非临场快照的抬升不触发(self):
        # is_final_snapshot=False 的常规抬赔属于正常价格波动，不属于临场背离
        euro = [_euro("Bet365", 2.00, 2.40, final=False)]
        result = _agent().verify(_data(euro, []))["anomalies"]["single_house_deviation"]
        assert result["present"] is False
