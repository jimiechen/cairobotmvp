# -*- coding: utf-8 -*-
"""ReviewAgent：七项矛盾检查与条件回退决策（PRD-12 F-14，ISS-FA3）。

职责边界：
- 负责：七项矛盾检查（CONSISTENCY_CHECKS.md v7.12：新增检查⑦推荐结构互斥合规、
  检查⑥扩为 A/B/C 三路径）逐项执行与结果落盘；决策产出
  （pass / retry ≤MAX_REVIEW_RETRIES 次 / manual 转人工裁决）。
- 不负责：执行修正动作（由 AnalysisRunner 回退环重跑 SynthesisAgent）、
  人工裁决交互（human-in-the-loop interrupt 由 ISS-FA4 F-20 实现）。

未产出层处理：大小球/比分池相关检查（①③④）在推荐结构仅含胜平负/让球两层的
MVP 阶段标注"na 不适用"，如实落盘，不编造结论。
"""
# F-14：条件回退最多 2 次，仍不过转人工裁决
MAX_REVIEW_RETRIES = 2
# 深盘口径：亚盘让球线绝对值 ≥1.0（v7 决策树盘口口径，来源 SKILL §3）
DEEP_LINE_ABS = 1.0
CHECK_NAMES = {
    "①": "风险提示与大小球方向一致",
    "②": "次选降级替换（深盘次选保护位）",
    "③": "盘口与比分池净胜分布一致",
    "④": "穿盘证据 S1-S4 幅度联动",
    "⑤": "次选决策树（深盘禁平局次选）",
    "⑥": "双选收益比路径合规（A/B/C）",
    "⑦": "推荐结构互斥合规",
}
_NA_NOTE = "大小球/比分池层 ISS-FA4 前未产出，标注不适用"


class ReviewAgent:
    """矛盾检查 Agent：输入 SynthesisAgent 结论，输出七项检查结果与决策。

    状态约束：无状态；attempt 由调用方传入用于回退计数。
    """

    def review(self, payload: dict) -> dict:
        """执行七项矛盾检查并给出决策。

        输入：{"synthesis": SynthesisAgent 结论, "handicap_line": 亚盘让球线|None,
               "attempt": 当前回退轮次}。
        输出：{"phase","checks","decision","violations"}。
        决策：无违规 → pass；有违规且 attempt <2 → retry；仍违规 → manual。
        """
        synthesis = payload.get("synthesis") or {}
        recommendation = synthesis.get("recommendation") or {}
        deep = self._is_deep(payload.get("handicap_line"))
        checks = [
            self._check_na("①"),
            self._check_next_protection(deep, recommendation),
            self._check_na("③"),
            self._check_na("④"),
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
