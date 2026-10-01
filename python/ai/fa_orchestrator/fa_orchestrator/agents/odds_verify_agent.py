# -*- coding: utf-8 -*-
"""OddsVerifyAgent：第二阶段赔率验证分析（PRD-12 F-11，SKILL v7.12 规则节点化）。

职责边界：
- 负责：赔率结构验证、欧亚转换（升盘真信号边界 v7.11、两层背离信让球层）、
  异常信号排查（防冷分级 v7.2、立博独开分流 v7.9、单家临场背离 v7.11）。
- 不负责：市场热度偏差计算与过热处置（ISS-FA3 SentimentAgent）、推荐单元
  结构判定（ISS-FA3 SynthesisAgent）、数据获取（data_source.py）。

判定优先级（v7.2/v7.11）：升盘+低水按真信号（含孤立升盘）；高水时深盘+实力背书
→ 真信号（004 锚点），状态背离 → 诱盘（v7.2），均势无背书 → 需实力差验证（008）。
纯规则计算，不调用 LLM 与外部服务。
"""
from statistics import median

# 主流公司清单（SKILL v7.0 数据质量检查：缺主流公司须告警）
MAINSTREAM_BOOKS = ("澳门", "立博", "威廉希尔", "Bet365")

LOW_WATER_THRESHOLD = 1.00        # 上盘水位 <1.00 判低水（真信号资金背书形态）
EXTREME_HIGH_WATER = 1.15         # 让球层极端超高水（无人承接，两层背离右侧条件）
SUPER_LOW_WATER = 0.75            # 胜平负层超低水（008 威廉 0.72 级别"资金极强"）
DEEP_EURO_WIN_ODDS = 2.0          # 主胜欧赔均值 <2.0 判深盘（v7.11 边界）
STATE_DIVERGE_GOALS = 1.0         # 场均进球 <1 且获升盘 → 状态背离（v7.2 诱盘条件）
SINGLE_HOUSE_RAISE_MIN_PCT = 5.0  # 单家主流临场逆势抬升 ≥5% 强制降星/减半仓（v7.11）
ANTI_COLD_DRAW_DROP_MIN_PCT = 3.0 # 平赔较初赔下调 ≥3% 计"压平"动作
ANTI_COLD_MIN_COMPANIES = 2       # ≥2 家同向压平才构成压平动作
STRENGTH_VALUE_RATIO = 2.0        # 身价差 ≥2 倍计实力背书（反大众检查②同源口径）

TRUE_SIGNAL = "真信号"
TRAP_SIGNAL = "诱盘"
NEED_STRENGTH_VERIFY = "需实力差验证，禁止重仓"
RAISE_FORM_TOGETHER = "共振升盘"
RAISE_FORM_SOLO = "孤立升盘"
RAISE_FORM_NONE = "无升盘"
STATE_ALIGNED = "同向"
STATE_DIVERGED = "背离"
TAG_TWO_LAYER = "[两层背离｜信让球层]"
TAG_ANTI_COLD_VALID = "[防冷有效]"
TAG_ANTI_COLD_DOWNGRADE = "[利润平衡，降级]"
TAG_SOLO_DEPTH = "[降深度]"
TAG_SOLO_VALUE = "[价值方向]"
TAG_SINGLE_HOUSE_TEMPLATE = "[单家背离-{pct}%｜降星/减半仓]"
NOTE_LADBROKES_ABSENT = "亚盘无立博数据，独开信号不可判定"


class OddsVerifyAgent:
    """赔率验证 Agent：输入单场三盘赔率数据块，输出第二阶段结论。

    状态约束：无状态；纯函数式规则计算，可并发调用。
    """

    def verify(self, match_data: dict) -> dict:
        """执行第二阶段赔率验证。

        输入：match_data（euro/asian/overunder 三盘 + state_baseline/strength/draw_support）。
        输出：第二阶段结论 dict（structure + euro_asia + anomalies）。
        失败行为：盘种缺失按空列表处理，对应检查降级标注，不抛异常、不编造。
        """
        euro = match_data.get("euro") or []
        asian = match_data.get("asian") or []
        strength = match_data.get("strength") or {}
        return {
            "phase": "odds_verify",
            "structure": self._verify_structure(euro),
            "euro_asia": self._verify_euro_asia(euro, asian, match_data, strength),
            "anomalies": self._check_anomalies(euro, asian, match_data, strength),
        }

    @staticmethod
    def _verify_structure(euro: list) -> dict:
        """赔率结构验证：公司数量、主流公司覆盖、返还率均值（v7.0）。"""
        books = [row.get("bk") for row in euro]
        payouts = [row["payout"] for row in euro if row.get("payout") is not None]
        return {
            "company_count": len(euro),
            "mainstream_missing": [bk for bk in MAINSTREAM_BOOKS if bk not in books],
            "avg_payout": round(sum(payouts) / len(payouts), 3) if payouts else None,
        }

    def _verify_euro_asia(self, euro: list, asian: list, match_data: dict, strength: dict) -> dict:
        """欧亚转换验证：升盘解读（v7.2/v7.11）+ 两层背离（008 锚点）。"""
        raise_rows = [r for r in asian
                      if r.get("current", {}).get("handicap", 0)
                      > r.get("initial", {}).get("handicap", 0)]
        avg_home_water = _mean([r["current"]["home_water"] for r in raise_rows])
        state_alignment = self._state_alignment(match_data.get("state_baseline") or {})
        deep_market = self._is_deep_market(euro)
        endorsed = self._is_strength_endorsed(strength)
        interpretation = self._interpret_raise(
            len(raise_rows), avg_home_water, state_alignment, deep_market, endorsed
        )
        divergent = self._detect_two_layer(asian)
        tags = [TAG_TWO_LAYER] if divergent else []
        return {
            "raise_count": len(raise_rows),
            "raise_form": self._raise_form(len(raise_rows)),
            "avg_home_water": avg_home_water,
            "state_alignment": state_alignment,
            "deep_market": deep_market,
            "strength": {
                "value_ratio": strength.get("value_ratio"),
                "level_gap": bool(strength.get("level_gap")),
                "h2h_suppression": strength.get("h2h_suppression", "none"),
            },
            "strength_endorsed": endorsed,
            "interpretation": interpretation,
            "two_layer_divergence": divergent,
            "tags": tags,
            "note": "",
        }

    @staticmethod
    def _state_alignment(state_baseline: dict):
        """状态同向性：主队近期场均进球 ≥1 判同向，<1 判背离（v7.2）。"""
        if not state_baseline or state_baseline.get("home_avg_goals") is None:
            return None
        if state_baseline["home_avg_goals"] < STATE_DIVERGE_GOALS:
            return STATE_DIVERGED
        return STATE_ALIGNED

    @staticmethod
    def _is_deep_market(euro: list) -> bool:
        """深盘判定：主胜欧赔均值 <2.0（v7.11 升盘真信号边界前置条件）。"""
        wins = [row["current"]["win"] for row in euro if row.get("current")]
        return bool(wins) and _mean(wins) < DEEP_EURO_WIN_ODDS

    @staticmethod
    def _is_strength_endorsed(strength: dict) -> bool:
        """基本面硬度背书：身价差≥2倍 / 级别断层 / 历史强压制任一命中（v7.11 004）。"""
        value_ratio = strength.get("value_ratio")
        return bool(
            (value_ratio is not None and value_ratio >= STRENGTH_VALUE_RATIO)
            or strength.get("level_gap")
            or strength.get("h2h_suppression") == "strong"
        )

    def _interpret_raise(self, raise_count: int, avg_home_water, state_alignment,
                         deep_market: bool, endorsed: bool):
        """升盘动作解读（v7.2 + v7.11 边界）：低水真信号 → 深盘背书 → 背离诱盘 → 需验证。"""
        if raise_count == 0:
            return None
        if avg_home_water is not None and avg_home_water < LOW_WATER_THRESHOLD:
            return TRUE_SIGNAL  # 升盘+低水（资金背书），孤立升盘低水同判（拉科锚点）
        if deep_market and endorsed:
            return TRUE_SIGNAL  # 深盘+基本面硬背书，高水孤立升盘按真信号（004 锚点）
        if state_alignment == STATE_DIVERGED:
            return TRAP_SIGNAL  # 升盘+高水+状态背离 → 诱盘（v7.2 双条件）
        return NEED_STRENGTH_VERIFY  # 均势盘无背书，禁止按真信号重仓（008 锚点）

    @staticmethod
    def _raise_form(raise_count: int) -> str:
        """升盘形态：≥2 家同向升盘为共振，单家为孤立（v7.2 案例两形态）。"""
        if raise_count >= 2:
            return RAISE_FORM_TOGETHER
        if raise_count == 1:
            return RAISE_FORM_SOLO
        return RAISE_FORM_NONE

    @staticmethod
    def _detect_two_layer(asian: list) -> bool:
        """两层背离检测（v7.11 008 固化）：胜平负层超低水与让球层升盘极端高水并存。"""
        super_low = [r for r in asian
                     if r.get("current", {}).get("handicap", 0) <= 0
                     and r.get("current", {}).get("home_water", 2.0) <= SUPER_LOW_WATER]
        extreme_raise = [r for r in asian
                         if r.get("current", {}).get("handicap", 0)
                         > r.get("initial", {}).get("handicap", 0)
                         and r.get("current", {}).get("home_water", 0.0) >= EXTREME_HIGH_WATER]
        return bool(super_low) and bool(extreme_raise)

    def _check_anomalies(self, euro: list, asian: list, match_data: dict, strength: dict) -> dict:
        """异常信号排查：防冷分级 + 立博独开分流 + 单家临场背离。"""
        return {
            "anti_cold": self._classify_anti_cold(
                euro, bool(strength.get("level_gap")),
                bool(match_data.get("draw_support")),
            ),
            "ladbrokes_solo": self._classify_ladbrokes_solo(asian),
            "single_house_deviation": self._detect_single_house(euro),
        }

    @staticmethod
    def _classify_anti_cold(euro: list, level_gap: bool, draw_support: bool) -> dict:
        """防冷分级（v7.2）：条件齐备才判防冷有效，否则降级为正常利润平衡。"""
        flattened = [
            row for row in euro
            if row.get("initial", {}).get("draw") and row.get("current", {}).get("draw")
            and (row["initial"]["draw"] - row["current"]["draw"])
            / row["initial"]["draw"] * 100 >= ANTI_COLD_DRAW_DROP_MIN_PCT
        ]
        if len(flattened) < ANTI_COLD_MIN_COMPANIES:
            return {"present": False, "tag": None, "reasons": []}
        if level_gap or not draw_support:
            reasons = _anti_cold_downgrade_reasons(level_gap, draw_support)
            return {"present": True, "tag": TAG_ANTI_COLD_DOWNGRADE, "reasons": reasons}
        return {"present": True, "tag": TAG_ANTI_COLD_VALID, "reasons": []}

    @staticmethod
    def _classify_ladbrokes_solo(asian: list) -> dict:
        """立博反向独开分流（v7.9）：同向浅盘→[降深度]；指向受让方→[价值方向]。"""
        ladbrokes = next((r for r in asian if r.get("bk") == "立博"), None)
        others = [r for r in asian if r.get("bk") != "立博"]
        if ladbrokes is None or not others:
            return {"present": False, "tag": None, "note": NOTE_LADBROKES_ABSENT}
        lad_hc = ladbrokes.get("current", {}).get("handicap", 0)
        others_hc = [r.get("current", {}).get("handicap", 0) for r in others]
        majority_hc = median(others_hc)
        if lad_hc > 0 and majority_hc > 0 and abs(lad_hc) < abs(majority_hc):
            return {"present": True, "tag": TAG_SOLO_DEPTH, "note": ""}
        if lad_hc < 0 < majority_hc:
            return {"present": True, "tag": TAG_SOLO_VALUE, "note": ""}
        return {"present": False, "tag": None, "note": ""}

    @staticmethod
    def _detect_single_house(euro: list) -> dict:
        """单家主流临场逆势抬升（v7.11 008 固化）：临场快照主胜抬升 ≥5% 强制降星。"""
        candidates = []
        for row in euro:
            if not row.get("is_final_snapshot"):
                continue
            initial_win = row.get("initial", {}).get("win")
            current_win = row.get("current", {}).get("win")
            if not initial_win or current_win is None:
                continue
            raise_pct = (current_win - initial_win) / initial_win * 100
            if raise_pct >= SINGLE_HOUSE_RAISE_MIN_PCT:
                candidates.append((raise_pct, row))
        if not candidates:
            return {"present": False, "pct": None, "bk": None, "tag": None}
        raise_pct, row = max(candidates, key=lambda item: item[0])
        pct = round(raise_pct, 1)
        return {
            "present": True,
            "pct": pct,
            "bk": row.get("bk"),
            "tag": TAG_SINGLE_HOUSE_TEMPLATE.format(pct=pct),
        }


def _mean(values: list):
    """空安全均值：空列表返回 None（调用方降级处理）。"""
    return sum(values) / len(values) if values else None


def _anti_cold_downgrade_reasons(level_gap: bool, draw_support: bool) -> list:
    """防冷降级原因（v7.2 三条禁止反买依据的对应项）。"""
    reasons = []
    if level_gap:
        reasons.append("存在联赛级别断层（实力进一步倾斜，防冷降噪）")
    if not draw_support:
        reasons.append("平局方向无基本面支持/无水位配合（纯价格信号，无量能验证）")
    return reasons
