# -*- coding: utf-8 -*-
"""SynthesisAgent：综合判定与推荐结构产出（PRD-12 F-13，ISS-FA3）。

职责边界：
- 负责：撤离信号收集与硬门槛（≥3 重无条件放弃，v7.11 / ≥2 重且情报零支撑，v7.9）、
  反大众前置检查（①压制锚点（复用 FA2 分级）+②实力差+③情报幅度词+语义解耦）、
  让球单选判定（v7.3 P 值/赔率/深盘门槛/过热方向）、低赔补位条款（v7.12 路径 C）、
  推荐结构互斥（v7.12 S/D 单元：每层 ≤1、跨层 ≤2、D+D 仅撤离 ≥3 重）。
- 不负责：热度偏差计算（sentiment_agent.py）、矛盾检查回退（review_agent.py）、
  数据取数与图编排。

语义解耦原则：禁止反买 ≠ 主推保留；v7.9 放弃规则优先于反大众阻断。
"""
from fa_orchestrator.analysis.odds_math import dual_path, implied_probabilities
from fa_orchestrator.agents.sentiment_agent import OVERHEAT_HHAND_PP

# 撤离/反向信号清单（SKILL v7.2/v7.8/v7.11 口径，来源见行内注释）
SIG_TWO_LAYER = "两层背离"           # v7.11：欧亚两层结论背离（单家背离联动）
SIG_SINGLE_HOUSE = "单家临场背离"    # v7.11：单家主流临场逆势抬升
SIG_TRAP_RAISE = "升盘诱盘"          # v7.2：升盘高水诱盘判定
SIG_BF_FALLING = "必发成交回落"      # v7.8 检查D/004 案例：必发方向成交回落
EVACUATION_HARD_GATE = 3             # v7.11：≥3 重 → 无条件放弃主推（周日007 案例）
EVACUATION_ZERO_SUPPORT = 2          # v7.9：≥2 重 且 情报层主利 0 条 → 放弃（周六012 案例）

# 反大众锚点（SKILL §3）：②实力差身价比 ≥2 倍；③情报幅度词
VALUE_RATIO_STRONG = 2.0
SWING_WORDS = ("崩盘", "击穿", "反弹")
# ①时效衰减补充口径：压制方近5场场均失球 ≥1.5 → 强压制降级弱压制（SKILL §3）
DECAY_CONCEDE = 1.5

# 让球单选判定（v7.3）：P ≥50% 且赔率 ≥1.60 单选；40-50% 仅双选；<40% 禁止；
# 深盘（让球线绝对值 ≥2）单选门槛提升至 55%
SINGLE_SELECT_MIN_P = 0.50
SINGLE_SELECT_MIN_ODDS = 1.60
DOUBLE_ONLY_P_LOW = 0.40
DEEP_LINE_ABS = 2.0
DEEP_MIN_P = 0.55
# 低赔补位（v7.12）：主推 SP <1.5 → 让球层强制双选；不合规按互斥砍剩单选
LOW_SP_THRESHOLD = 1.5

# 推荐单元星级：单选 4 / 双选 3 / 降星 2（保方向，v7.9 009 案例）
STAR_SINGLE_SELECT = 4
STAR_DOUBLE_ONLY = 3
STAR_DOWNGRADED = 2
# 方向 → 双选腿映射（撤离放弃改配平/负双选；客胜向改胜/平）
DUAL_LEGS = {
    "win": ["draw", "lose"],
    "lose": ["win", "draw"],
    "draw": ["win", "lose"],
}
# 市场方向 → 场地方向（情报主利口径；平局方向无确定性利多，计 0）
SIDE_OF_DIRECTION = {"win": "home", "lose": "away", "draw": None}


class SynthesisAgent:
    """综合判定 Agent：输入三阶段结论与竞彩投票块，输出推荐结构。

    状态约束：无状态；attempt 由调用方（AnalysisRunner 回退环）传入。
    """

    def analyze(self, data: dict, attempt: int = 0) -> dict:
        """产出单场推荐结构。

        输入：{"fundamental","odds","sentiment","vote","market_direction",
        "greed_fear"(可选，立项三贪婪恐惧层结论)}。
        输出：{"phase","evacuation","anti_public","main_action","recommendation",
        "attempt"[,"trap_combination"]}（贪婪恐惧输入存在时透传陷阱组合）。
        失败行为：数据块缺失按降级标注保留结构，不抛异常。
        """
        fundamental = data.get("fundamental") or {}
        odds = data.get("odds") or {}
        sentiment = data.get("sentiment") or {}
        vote = data.get("vote")
        market_direction = data.get("market_direction") or "win"
        greed_fear = data.get("greed_fear") or {}

        evacuation = self._collect_evacuation(odds, sentiment)
        anti_public = self._anti_public(fundamental, odds, market_direction)
        intel_favor = self._intel_favor_count(fundamental, market_direction)
        main_action = self._decide_main_action(evacuation, intel_favor, sentiment, anti_public)

        had_unit = self._build_had_unit(market_direction, main_action, sentiment, vote)
        hhand_unit, notes = self._build_hhand_unit(vote, sentiment, anti_public, market_direction)
        notes.extend(self._cover_note(had_unit, vote, market_direction))
        notes.extend(self._greed_fear_note(greed_fear))
        mutual = self.enforce_mutual_exclusion(had_unit, hhand_unit, evacuation["hard_gate"])
        if mutual["adjusted"]:
            hhand_unit = None
            notes.append("互斥修正：D+D 仅撤离≥3重允许，砍剩让球层")
        recommendation = {
            "had_unit": had_unit,
            "hhand_unit": hhand_unit,
            "notes": notes,
            "mutual_exclusion": mutual,
        }
        result = {"phase": "synthesis", "evacuation": evacuation, "anti_public": anti_public,
                  "main_action": main_action, "recommendation": recommendation, "attempt": attempt}
        if greed_fear:
            # 陷阱组合透传给报告层；警示口径见 _greed_fear_note（待主控确认）
            result["trap_combination"] = greed_fear.get("trap_combination")
        return result

    # ------------------------------------------------------------------ 撤离
    @staticmethod
    def _collect_evacuation(odds: dict, sentiment: dict) -> dict:
        """收集撤离/反向信号并判定硬门槛（≥3 重，v7.11）。"""
        euro_asia = odds.get("euro_asia") or {}
        anomalies = odds.get("anomalies") or {}
        bf = sentiment.get("bf") or {}
        signals = []
        if euro_asia.get("two_layer_divergence"):
            signals.append(SIG_TWO_LAYER)
        if (anomalies.get("single_house_deviation") or {}).get("present"):
            signals.append(SIG_SINGLE_HOUSE)
        if euro_asia.get("interpretation") == "诱盘":
            signals.append(SIG_TRAP_RAISE)
        if bf.get("available") and bf.get("trend") == "falling":
            signals.append(SIG_BF_FALLING)
        return {"count": len(signals), "signals": signals,
                "hard_gate": len(signals) >= EVACUATION_HARD_GATE}

    # ---------------------------------------------------------------- 反大众
    def _anti_public(self, fundamental: dict, odds: dict, market_direction: str) -> dict:
        """反大众前置检查（SKILL §3 ①②③ + 时效衰减 + 语义解耦）。"""
        suppression, decay_applied = self._anchor1_suppression(fundamental)
        strength = (odds.get("euro_asia") or {}).get("strength") or {}
        anchor2 = ((strength.get("value_ratio") or 0) >= VALUE_RATIO_STRONG
                   or bool((fundamental.get("schedule") or {}).get("level_gap")))
        items = (fundamental.get("intelligence") or {}).get("items") or []
        anchor3 = any(word in (item.get("text") or "") for item in items for word in SWING_WORDS)
        allow_counter = suppression == "none" and not anchor2 and not anchor3
        note = ("三条锚点全不命中，允许反买（仓位减半）" if allow_counter
                else "反大众锚点命中，禁止反买输出；禁止反买≠主推保留（语义解耦）")
        return {"anchor1": suppression, "decay_applied": decay_applied, "anchor2": anchor2,
                "anchor3": anchor3, "allow_counter": allow_counter, "note": note}

    @staticmethod
    def _anchor1_suppression(fundamental: dict) -> tuple:
        """①压制锚点：复用 FA2 分级（主队视角）；近5场失球达标补足时效衰减。"""
        h2h = fundamental.get("h2h") or {}
        suppression = h2h.get("suppression") or "none"
        decayed = bool(h2h.get("decayed"))
        avg_concede = (fundamental.get("recent_form") or {}).get("avg_concede")
        if avg_concede is not None and avg_concede >= DECAY_CONCEDE:
            decayed = True
        if suppression == "strong" and decayed:
            return "weak", True
        return suppression, decayed

    @staticmethod
    def _intel_favor_count(fundamental: dict, market_direction: str) -> int:
        """情报层主利条数（v7.9 联动口径）：主推方向场地同向利多计数。

        优先读原始 items；FA1 结论仅含五向计数时按 {side}_favor 回退。
        """
        side = SIDE_OF_DIRECTION.get(market_direction)
        if side is None:
            return 0
        intel = fundamental.get("intelligence") or {}
        items = intel.get("items")
        if items is not None:
            return sum(1 for item in items
                       if item.get("party") == side and item.get("sentiment") == "favor")
        return intel.get(f"{side}_favor", 0)

    # ---------------------------------------------------------------- 主推
    @staticmethod
    def _decide_main_action(evacuation: dict, intel_favor: int, sentiment: dict,
                            anti_public: dict) -> str:
        """主推动作裁决：撤离硬门槛 > v7.9 放弃 > 过热处置 > 弱压制降星 > 保留。"""
        if evacuation["count"] >= EVACUATION_HARD_GATE:
            return "abandon"
        if evacuation["count"] >= EVACUATION_ZERO_SUPPORT and intel_favor == 0:
            return "abandon"
        disposition = (sentiment.get("overheat") or {}).get("disposition")
        if disposition == "abandon":
            return "abandon"
        if disposition == "downgrade_keep" or anti_public["anchor1"] == "weak":
            return "downgrade"
        return "keep"

    def _build_had_unit(self, market_direction: str, main_action: str,
                        sentiment: dict, vote: dict) -> dict | None:
        """胜平负层单元：放弃 → 平/负双选（反爆冷保护）；保留/降星 → 主推单选。"""
        if main_action == "abandon":
            legs = DUAL_LEGS[market_direction]
            overheat_abandon = (sentiment.get("overheat") or {}).get("disposition") == "abandon"
            tag = "[过热放弃｜防守向双选]" if overheat_abandon else "[撤离放弃｜平负双选]"
            return {"form": "D", "legs": legs, "path": self._dual_path_for_legs(legs, vote, False),
                    "stars": STAR_DOUBLE_ONLY, "tag": tag}
        stars = STAR_DOWNGRADED if main_action == "downgrade" else 3
        return {"form": "S", "selection": market_direction, "legs": None,
                "path": None, "stars": stars, "tag": None}

    # ---------------------------------------------------------------- 让球层
    def _build_hhand_unit(self, vote: dict, sentiment: dict, anti_public: dict,
                          market_direction: str) -> tuple:
        """让球层单元（v7.3 单选判定 + v7.12 低赔补位），返回 (单元, 备注)。"""
        hhand = (vote or {}).get("hhand") or {}
        if not hhand.get("odds"):
            return None, ["竞彩让球赔率缺失，无法判定让球单元"]
        probability = implied_probabilities([{"current": hhand["odds"]}])
        if not probability:
            return None, ["竞彩让球赔率缺失，无法判定让球单元"]
        top = max(probability, key=probability.get)
        top_p = probability[top] / 100.0
        line = hhand.get("line") or 0.0
        gate = DEEP_MIN_P if abs(line) >= DEEP_LINE_ABS else SINGLE_SELECT_MIN_P
        sp_main = ((vote or {}).get("sp") or {}).get(market_direction)
        if sp_main is not None and sp_main < LOW_SP_THRESHOLD:
            return self._cover_unit(hhand, probability, market_direction)
        if anti_public["anchor1"] == "strong":
            return None, ["①强压制命中，放弃让球单选"]
        overheat_hhand = sentiment.get("hhand") or {}
        if (overheat_hhand.get("available")
                and (overheat_hhand.get("deviation_pp") or 0) >= OVERHEAT_HHAND_PP
                and overheat_hhand.get("popular") == top):
            return None, [f"让球方向过热(+{overheat_hhand['deviation_pp']}pp)，禁止单选"]
        if top_p < DOUBLE_ONLY_P_LOW:
            return None, [f"让球方向P={top_p:.0%}，禁止单选"]
        if abs(line) >= DEEP_LINE_ABS and top_p < DEEP_MIN_P:
            return None, [f"深盘让球线{line}，单选门槛{DEEP_MIN_P:.0%}未达标，禁止单选"]
        if top_p >= gate and hhand["odds"].get(top, 0) >= SINGLE_SELECT_MIN_ODDS:
            return {"form": "S", "selection": top, "legs": None, "path": None,
                    "stars": STAR_SINGLE_SELECT, "tag": None}, []
        second = max((key for key in probability if key != top), key=probability.get)
        legs = [top, second]
        path = dual_path({leg: hhand["odds"][leg] for leg in legs}, probability, cover=False)
        return {"form": "D", "legs": legs, "path": path, "stars": STAR_DOUBLE_ONLY,
                "tag": None}, []

    def _cover_unit(self, hhand: dict, probability: dict, market_direction: str) -> tuple:
        """低赔补位单元（v7.12）：强制双选；路径不合规砍剩单选并写明量化原因。"""
        legs = DUAL_LEGS[market_direction]
        leg_odds = {leg: hhand["odds"][leg] for leg in legs}
        path = dual_path(leg_odds, probability, cover=True)
        if path:
            unit = {"form": "D", "legs": legs, "path": path, "stars": STAR_DOUBLE_ONLY,
                    "tag": f"[低赔补位｜路径{path}]"}
            return unit, []
        keep_leg = max(legs, key=lambda leg: leg_odds[leg])
        min_odds = min(leg_odds.values())
        note = f"低赔补位路径不合规（最低腿{min_odds}，EV未达标），砍剩单选并写明原因"
        return {"form": "S", "selection": keep_leg, "legs": None, "path": None,
                "stars": STAR_DOWNGRADED, "tag": "[低赔补位｜砍剩单选]"}, [note]

    @staticmethod
    def _dual_path_for_legs(legs: list, vote: dict, cover: bool) -> str | None:
        """按竞彩 SP 判定双选收益比路径；SP 缺失返回 None（检查⑥按缺失处理）。"""
        sp = (vote or {}).get("sp") or {}
        if not all(sp.get(leg) for leg in legs):
            return None
        probability = implied_probabilities([{"current": sp}])
        return dual_path({leg: sp[leg] for leg in legs}, probability, cover=cover)

    @staticmethod
    def _cover_note(had_unit: dict | None, vote: dict, market_direction: str) -> list:
        """低赔补位条款结论标注：触发补位时由补位单元标注，其余写明无补位需求。"""
        if not had_unit or had_unit.get("form") != "S":
            return []
        sp_main = ((vote or {}).get("sp") or {}).get(market_direction)
        if sp_main is None:
            return ["竞彩SP缺失，低赔补位判定降级"]
        if sp_main >= LOW_SP_THRESHOLD:
            return ["无补位需求（主推SP≥1.5）"]
        return []

    # ------------------------------------------------- 贪婪恐惧消费（立项三）
    @staticmethod
    def _greed_fear_note(greed_fear: dict) -> list:
        """贪婪恐惧陷阱警示（立项三最小闭环）：仅双侧强陷阱追加警示备注。

        口径约束：强贪婪（诱多）+ 强恐惧（诱空）并存提示双向诱盘风险；
        只写警示不改推荐结构（调整口径待主控确认，PRD-12 §8.3 F-21）。
        输入为空（上游缺失/未接入）返回空列表，行为向后兼容。
        """
        if not greed_fear:
            return []
        greed_level = (greed_fear.get("greed") or {}).get("level")
        fear_level = (greed_fear.get("fear") or {}).get("level")
        if greed_level == "强" and fear_level == "强":
            combination = greed_fear.get("trap_combination") or ""
            return [f"贪婪恐惧陷阱警示：{combination}，警惕双向诱盘"]
        return []

    # ---------------------------------------------------------------- 互斥
    @staticmethod
    def enforce_mutual_exclusion(had_unit: dict | None, hhand_unit: dict | None,
                                 evacuation_hard_gate: bool) -> dict:
        """推荐结构互斥（v7.12）：每层 ≤1 单元、跨层 ≤2、D+D 仅撤离 ≥3 重允许。

        输入：两单元（可为 None）、撤离硬门槛标志。
        输出：{"legal","form","adjusted"}；D+D 无硬门槛时砍剩让球层保留胜平负主层。
        """
        had_form = had_unit["form"] if had_unit else "空"
        hhand_form = hhand_unit["form"] if hhand_unit else "空"
        if had_form == "D" and hhand_form == "D" and not evacuation_hard_gate:
            return {"legal": True, "form": "D+空", "adjusted": True}
        return {"legal": True, "form": f"{had_form}+{hhand_form}", "adjusted": False}
