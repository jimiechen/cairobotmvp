# -*- coding: utf-8 -*-
"""ReviewAgent：七项矛盾检查与条件回退决策（PRD-12 F-14，ISS-FA3/ISS-FA4）。

职责边界：
- 负责：七项矛盾检查（CONSISTENCY_CHECKS.md v7.12：检查⑦推荐结构互斥合规、
  检查⑥扩为 A/B/C 三路径；ISS-FA4 解锁检查①③④大小球/比分池实检）逐项执行
  与结果落盘；决策产出（pass / retry ≤MAX_REVIEW_RETRIES 次 / manual 转人工裁决）。
- 不负责：执行修正动作（由 AnalysisRunner 回退环重跑 SynthesisAgent）、
  人工裁决交互（human-in-the-loop interrupt 由 ISS-FA4 F-20 实现）。

未产出层处理：payload 无 goal_layer 或为空层（无 phase）时，检查①③④标注
"na 不适用"如实落盘；产出后按 CONSISTENCY_CHECKS.md 实检，不编造结论。
"""
from fa_orchestrator.agents.goal_layer_agent import AMPLITUDE_WORDS_RISK

# F-14：条件回退最多 2 次，仍不过转人工裁决
MAX_REVIEW_RETRIES = 2
# 深盘口径：亚盘让球线绝对值 ≥1.0（v7 决策树盘口口径，来源 SKILL §3）
DEEP_LINE_ABS = 1.0
# 穿盘净胜口径：比分净胜球绝对值 ≥3（与 goal_layer_agent 比分池规则一致）
PENETRATION_NET_MIN = 3
CHECK_NAMES = {
    "①": "风险提示与大小球方向一致",
    "②": "次选降级替换（深盘次选保护位）",
    "③": "盘口与比分池净胜分布一致",
    "④": "穿盘证据 S1-S4 幅度联动",
    "⑤": "次选决策树（深盘禁平局次选）",
    "⑥": "双选收益比路径合规（A/B/C）",
    "⑦": "推荐结构互斥合规",
}
_NA_NOTE = "大小球/比分池层未产出，标注不适用"


class ReviewAgent:
    """矛盾检查 Agent：输入 SynthesisAgent 结论，输出七项检查结果与决策。

    状态约束：无状态；attempt 由调用方传入用于回退计数。
    """

    def review(self, payload: dict) -> dict:
        """执行七项矛盾检查并给出决策。

        输入：{"synthesis": SynthesisAgent 结论, "handicap_line": 亚盘让球线|None,
               "attempt": 当前回退轮次, "goal_layer": GoalLayerAgent 结论|{}(可选)}。
        输出：{"phase","checks","decision","violations"}。
        决策：无违规 → pass；有违规且 attempt <2 → retry；仍违规 → manual。
        大小球层已产出（phase=goal_layer）时检查①③④实检，否则标注 na。
        """
        synthesis = payload.get("synthesis") or {}
        recommendation = synthesis.get("recommendation") or {}
        deep = self._is_deep(payload.get("handicap_line"))
        goal_layer = payload.get("goal_layer") or {}
        produced = goal_layer.get("phase") == "goal_layer"
        checks = [
            (self._check_risk_vs_ou(goal_layer, recommendation) if produced
             else self._check_na("①")),
            self._check_next_protection(deep, recommendation),
            (self._check_pool_consistency(goal_layer, deep) if produced
             else self._check_na("③")),
            (self._check_penetration(goal_layer, recommendation) if produced
             else self._check_na("④")),
            self._check_deep_draw_second(deep, recommendation, synthesis),
            self._check_dual_path(recommendation),
            self._check_mutual_exclusion(recommendation),
        ]
        violations = [check["id"] for check in checks if check["result"] == "violation"]
        if not violations:
            decision = "pass"
        elif payload.get("attempt", 0) < MAX_REVIEW_RETRIES:
            decision = "retry"
        else:
            decision = "manual"
        return {"phase": "review", "checks": checks, "decision": decision,
                "violations": violations}

    @staticmethod
    def _is_deep(handicap_line) -> bool:
        """深盘判定：让球线绝对值 ≥1.0；线缺失（中文盘口文本未解析）按非深盘。"""
        return handicap_line is not None and abs(handicap_line) >= DEEP_LINE_ABS

    @staticmethod
    def _check_na(check_id: str) -> dict:
        """未产出层检查项：标注 na 不适用，不触发回退。"""
        return {"id": check_id, "name": CHECK_NAMES[check_id], "result": "na",
                "action": None, "note": _NA_NOTE}

    @staticmethod
    def _pass_note(check_id: str, note: str) -> dict:
        """构造带说明的通过项结果（如"未配置仅记录"场景）。"""
        return {"id": check_id, "name": CHECK_NAMES[check_id], "result": "pass",
                "action": None, "note": note}

    def _check_risk_vs_ou(self, goal_layer: dict, recommendation: dict) -> dict:
        """检查①（ISS-FA4 解锁）：风险提示幅度词与大小球方向一致。

        规则（CONSISTENCY_CHECKS.md 检查①）：notes 命中幅度词（击穿/崩盘/大胜/
        反弹/剪刀差）时大小球方向不得为"小"，反向即矛盾。
        大小球未配置（观望/缺口）时仅记录不校验。
        """
        ou_unit = goal_layer.get("ou_unit")
        if not ou_unit:
            return self._pass_note("①", "大小球未配置，风险提示仅记录不校验")
        notes = " ".join(recommendation.get("notes") or [])
        if (ou_unit.get("direction") == "under"
                and any(word in notes for word in AMPLITUDE_WORDS_RISK)):
            return {"id": "①", "name": CHECK_NAMES["①"], "result": "violation",
                    "action": "回退修正大小球方向或调整风险提示幅度词",
                    "note": "风险提示与大小球反向"}
        return self._pass("①")

    def _check_pool_consistency(self, goal_layer: dict, deep: bool) -> dict:
        """检查③（ISS-FA4 解锁）：盘口与比分池净胜分布一致。

        规则（CONSISTENCY_CHECKS.md 检查③）：深盘 + S1+S2 双穿证据时池须含
        净胜 3+ 保护位；大球方向池不得全低进球；小球方向池不得含净胜 3+。
        """
        pool = goal_layer.get("score_pool") or []
        penetration = goal_layer.get("penetration") or {}
        direction = (goal_layer.get("ou_unit") or {}).get("direction")
        net_high = [i for i in pool
                    if abs(i.get("net_win") or 0) >= PENETRATION_NET_MIN]
        if (deep and penetration.get("S1") and penetration.get("S2")
                and not net_high):
            return {"id": "③", "name": CHECK_NAMES["③"], "result": "violation",
                    "action": "回退补充净胜三比分", "note": "深盘双穿证据池缺净胜保护"}
        if direction == "over" and pool and all((i.get("total") or 0) <= 2 for i in pool):
            return {"id": "③", "name": CHECK_NAMES["③"], "result": "violation",
                    "action": "回退补充高进球比分", "note": "大球方向池全低进球"}
        if direction == "under" and net_high:
            return {"id": "③", "name": CHECK_NAMES["③"], "result": "violation",
                    "action": "回退剔除净胜三比分", "note": "小球方向池含净胜三"}
        return self._pass("③")

    def _check_penetration(self, goal_layer: dict, recommendation: dict) -> dict:
        """检查④（ISS-FA4 解锁）：穿盘证据 S1-S4 与幅度联动。

        规则（CONSISTENCY_CHECKS.md 检查四）：S1+S2 双穿证据须联动幅度配置
        （池含净胜 3+ 且大小球非"小"）；S3 幅度词命中禁止降级配小；
        证据命中 ≥3 时主推须全额星级不降星。
        """
        penetration = goal_layer.get("penetration") or {}
        direction = (goal_layer.get("ou_unit") or {}).get("direction")
        pool = goal_layer.get("score_pool") or []
        has_net_high = any(abs(i.get("net_win") or 0) >= PENETRATION_NET_MIN
                           for i in pool)
        if (penetration.get("S1") and penetration.get("S2")
                and (direction == "under" or not has_net_high)):
            return {"id": "④", "name": CHECK_NAMES["④"], "result": "violation",
                    "action": "回退补充净胜比分或修正大小球方向",
                    "note": "双穿证据未联动幅度配置"}
        if penetration.get("S3") and direction == "under":
            return {"id": "④", "name": CHECK_NAMES["④"], "result": "violation",
                    "action": "回退修正大小球方向或幅度提示",
                    "note": "幅度层命中禁止小方向降级"}
        if (penetration.get("hits") or 0) >= 3:
            had_unit = recommendation.get("had_unit") or {}
            if (had_unit.get("stars") or 0) <= 2:
                return {"id": "④", "name": CHECK_NAMES["④"], "result": "violation",
                        "action": "回退恢复主推全额星级", "note": "穿盘证据满三禁止降星"}
        return self._pass("④")

    def _check_next_protection(self, deep: bool, recommendation: dict) -> dict:
        """检查②：深盘主推单选时次选保护位不得静默缺省（缺省须写明原因）。"""
        had_unit = recommendation.get("had_unit")
        hhand_unit = recommendation.get("hhand_unit")
        notes = recommendation.get("notes") or []
        if not (deep and had_unit and had_unit.get("form") == "S" and hhand_unit is None):
            return self._pass("②")
        if any("让球" in note for note in notes):
            return self._pass("②")
        return {"id": "②", "name": CHECK_NAMES["②"], "result": "violation",
                "action": "回退补充双选保护位或写明不配置原因", "note": "次选保护位静默缺省"}

    def _check_deep_draw_second(self, deep: bool, recommendation: dict,
                                synthesis: dict) -> dict:
        """检查⑤：深盘双选含平局腿须有反爆冷依据（撤离放弃场景豁免，v7.11）。"""
        had_unit = recommendation.get("had_unit")
        if not (deep and had_unit and had_unit.get("form") == "D"
                and "draw" in (had_unit.get("legs") or [])):
            return self._pass("⑤")
        if synthesis.get("main_action") == "abandon":
            return self._pass("⑤")
        return {"id": "⑤", "name": CHECK_NAMES["⑤"], "result": "violation",
                "action": "回退将平局腿替换为同向让平或补充反爆冷依据", "note": "深盘平局次选"}

    def _check_dual_path(self, recommendation: dict) -> dict:
        """检查⑥（v7.12 扩为 A/B/C）：每个双选单元必须携带合规收益比路径。"""
        for layer in ("had_unit", "hhand_unit"):
            unit = recommendation.get(layer)
            if unit and unit.get("form") == "D" and unit.get("path") not in ("A", "B", "C"):
                return {"id": "⑥", "name": CHECK_NAMES["⑥"], "result": "violation",
                        "action": f"回退补充 {layer} 收益比路径或改单选", "note": "双选缺路径"}
        return self._pass("⑥")

    @staticmethod
    def _check_mutual_exclusion(recommendation: dict) -> dict:
        """检查⑦（v7.12 新增）：推荐结构互斥合规。"""
        mutual = recommendation.get("mutual_exclusion") or {}
        if mutual.get("legal") is False:
            return {"id": "⑦", "name": CHECK_NAMES["⑦"], "result": "violation",
                    "action": "回退按互斥原则砍剩单元", "note": "互斥违规"}
        return {"id": "⑦", "name": CHECK_NAMES["⑦"], "result": "pass",
                "action": None, "note": ""}

    @staticmethod
    def _pass(check_id: str) -> dict:
        """构造通过项结果。"""
        return {"id": check_id, "name": CHECK_NAMES[check_id], "result": "pass",
                "action": None, "note": ""}
