# -*- coding: utf-8 -*-
"""GoalLayerAgent：大小球与比分池层（PRD-12 ISS-FA4，skill v7.2/v7.8/v7.11 规则节点化）。

职责边界：
- 负责：N6 进球外推消费（expected_total_goals）→ 大小球方向判定与推荐单元、
  比分池生成与预测输出、杯赛/国际赛大小球修正（v7.8/v7.11）、穿盘证据
  S1-S4 计算、N4 升深穿盘预警减仓（0.5→0.25）。
- 不负责：外推计算本身（fundamental_agent.recent_form）、胜平负/让球推荐
  （synthesis_agent.py）、矛盾检查回退（review_agent.py）、数据取数。

未产出处理：盘口/外推/方向缺失时如实标注缺口并返回 None 单元，不编造结论。
"""

# 外推与盘口差 ≥0.5 才给方向，贴近盘口观望（skill 检查④ S4"多球外推"口径）
OU_MARGIN = 0.5
# 大小球标准仓位单位（skill 价值小注 0.5 单位口径）
OU_STAKE = 0.5
# 国际赛/友谊赛轻仓上限（v7.11：≤0.3 单位）
INTERNATIONAL_STAKE_CAP = 0.3
# N4 穿盘预警减仓仓位（立项评审：0.5→0.25）
N4_STAKE = 0.25
# N4 升深判定：亚盘升盘且现盘让球线绝对值 ≥1.0（深盘口径，同 review_agent）
DEEP_LINE_ABS = 1.0
# S1 穿盘证据：基准去水主胜/客胜 ≥64%（CONSISTENCY_CHECKS.md 检查四）
S1_MIN_PROB = 0.64
# 比分池进球档位阈值（平局入池硬条件"预期<2.2"与高进球 3.0 分界）
POOL_LOW_THRESHOLD = 2.2
POOL_HIGH_THRESHOLD = 3.0
# 比分池基础矩阵：方向 × 进球档位（比分为主-客格式，win 净胜为正）
SCORE_POOL_BASE = {
    ("win", "low"): ["1-0", "2-1", "2-0"],
    ("win", "mid"): ["2-1", "2-0", "3-1"],
    ("win", "high"): ["2-1", "3-1", "3-0"],
    ("draw", "low"): ["0-0", "1-1"],
    ("draw", "mid"): ["1-1", "2-2"],
    ("draw", "high"): ["1-1", "2-2", "3-3"],
    ("lose", "low"): ["0-1", "1-2", "0-2"],
    ("lose", "mid"): ["1-2", "0-2", "2-3"],
    ("lose", "high"): ["1-2", "2-3", "0-3"],
}
# 比分池最低单元数（skill 池一般 3 个左右，over/under 修正后不足则补同向比分）
POOL_MIN_SIZE = 2
# S3 情报幅度词（CONSISTENCY_CHECKS.md 检查四：崩盘/击穿/反弹）
AMPLITUDE_WORDS_S3 = ("崩盘", "击穿", "反弹")
# 检查①幅度词（风险提示类：击穿/崩盘/大胜/反弹/剪刀差）
AMPLITUDE_WORDS_RISK = ("击穿", "崩盘", "大胜", "反弹", "剪刀差")
TAG_INTERNATIONAL = "[国际赛方差大，轻仓]"
TAG_N4 = "[穿盘预警｜减仓]"
TAG_PENETRATION_HIGH = "[穿盘净胜]"
TAG_UPSET_GUARD = "[反向爆冷保护]"
NOTE_CUP_CONFLICT = "杯赛/锦标赛主锚失效冲突，不配置"
NOTE_OVERLOOK = "外推与盘口贴近，观望不配置"
NOTE_NO_LINE = "竞彩大小球盘缺失，无法配置"
NOTE_NO_EXPECTED = "近期状态外推缺失，无法配置"
NOTE_NO_DIRECTION = "市场方向缺失，比分池不适用"


class GoalLayerAgent:
    """大小球/比分池 Agent：输入基本面外推与大小球盘，输出 ou_unit 与比分池。

    状态约束：无状态；纯函数式规则计算，可并发调用。
    """

    def analyze(self, data: dict) -> dict:
        """产出大小球推荐单元与比分池。

        输入：{"fundamental","overunder","market_direction","handicap_line",
               "handicap_movement","synthesis","odds_verify"}。
        输出：{"phase","expected_total_goals","line","ou_unit","ou_note",
               "score_pool","penetration","n4_triggered"}。
        失败行为：数据块缺失按缺口标注返回 None/空结构，不抛异常、不编造。
        """
        fundamental = data.get("fundamental") or {}
        recent = fundamental.get("recent_form") or {}
        expected = recent.get("expected_total_goals")
        overunder = data.get("overunder") or []
        line_row = self._pick_line_row(overunder)
        line = ((line_row or {}).get("current") or {}).get("line")
        ou_unit, ou_note = self._decide_ou(
            expected, line, line_row, recent, data.get("handicap_movement"),
            data.get("handicap_line"),
        )
        penetration = self._penetration(data, ou_unit)
        direction = data.get("market_direction")
        pool = self._build_score_pool(direction, expected, ou_unit, penetration)
        return {
            "phase": "goal_layer",
            "expected_total_goals": expected,
            "line": line,
            "ou_unit": ou_unit,
            "ou_note": ou_note or (None if ou_unit else NOTE_OVERLOOK),
            "score_pool": pool,
            "pool_note": NOTE_NO_DIRECTION if direction not in ("win", "draw", "lose") else None,
            "penetration": penetration,
            "n4_triggered": bool(ou_unit and ou_unit.get("tag")
                                 and TAG_N4 in ou_unit["tag"]),
        }

    # ---------------------------------------------------------------- 大小球
    @staticmethod
    def _pick_line_row(overunder: list) -> dict | None:
        """澳门行优先取盘口行（对齐 runner _numeric_handicap 模式），无则取首行。"""
        return next((r for r in overunder if r.get("bk") == "澳门"),
                    overunder[0] if overunder else None)

    def _decide_ou(self, expected, line, line_row, recent, movement,
                   handicap_line) -> tuple:
        """大小球方向判定：N6 外推主锚 + 杯赛/国际赛/N4 修正，返回 (单元, 缺口备注)。

        失败行为：盘口/外推缺失 → (None, 缺口备注)；外推贴近盘口 → (None, 观望)。
        """
        if line is None or line_row is None:
            return None, NOTE_NO_LINE
        if expected is None:
            return None, NOTE_NO_EXPECTED
        diff = expected - line
        if abs(diff) < OU_MARGIN:
            return None, NOTE_OVERLOOK
        direction = "over" if diff > 0 else "under"
        correction = recent.get("correction")
        if correction == "cup" and self._cup_conflict(direction, line_row, recent):
            return None, NOTE_CUP_CONFLICT
        return self._build_ou_unit(direction, line, correction, movement,
                                   handicap_line), None

    @staticmethod
    def _cup_conflict(direction: str, line_row: dict, recent: dict) -> bool:
        """杯赛冲突判定（v7.8）：外推方向与市场升/降线动作矛盾 → 主锚失效冲突。

        对攻崩坏形态（一方近5场场均总进球 ≥3.5）不降权：仍按外推方向保留。
        """
        initial_line = (line_row.get("initial") or {}).get("line")
        current_line = (line_row.get("current") or {}).get("line")
        if isinstance(initial_line, (int, float)) and initial_line != current_line:
            market_direction = "over" if current_line > initial_line else "under"
            if market_direction != direction:
                return True
        return False

    def _build_ou_unit(self, direction: str, line, correction, movement,
                       handicap_line) -> dict:
        """构造大小球单元：杯赛标注 / 国际赛轻仓 / N4 升深减仓（0.5→0.25）。"""
        stake = OU_STAKE
        tag = None
        if correction == "international":
            stake = min(stake, INTERNATIONAL_STAKE_CAP)
            tag = TAG_INTERNATIONAL
        if (movement == "升盘" and handicap_line is not None
                and abs(handicap_line) >= DEEP_LINE_ABS):
            stake = N4_STAKE
            tag = f"{tag}；{TAG_N4}" if tag else TAG_N4
        return {"direction": direction, "line": line, "stars": 3,
                "stake": stake, "tag": tag, "note": ""}

    # ---------------------------------------------------------------- 比分池
    def _build_score_pool(self, direction, expected, ou_unit, penetration) -> list:
        """比分池生成：方向 × 进球档位矩阵 + 大/小修正 + 强穿盘证据追加。

        输入 penetration 为 analyze 内已算好的穿盘证据（避免重复计算）。
        失败行为：市场方向缺失返回空池（pool_note 标注）。
        """
        if direction not in ("win", "draw", "lose"):
            return []
        tier = self._pool_tier(expected, ou_unit)
        pool = [self._pool_item(score) for score in SCORE_POOL_BASE[(direction, tier)]]
        pool = self._adjust_pool_for_ou(pool, direction, ou_unit)
        pool = self._append_penetration_scores(pool, direction, penetration)
        return pool

    @staticmethod
    def _pool_tier(expected, ou_unit) -> str:
        """池进球档位：小球强制低档；无大小球按外推分档；大球至少中档。"""
        if ou_unit and ou_unit.get("direction") == "under":
            return "low"
        if expected is None or expected < POOL_LOW_THRESHOLD:
            return "low"
        if expected >= POOL_HIGH_THRESHOLD:
            return "high"
        return "mid"

    @staticmethod
    def _pool_item(score: str) -> dict:
        """比分字符串 → 池单元 {score, net_win, total}。"""
        home, away = (int(part) for part in score.split("-"))
        return {"score": score, "net_win": home - away, "total": home + away,
                "tag": None}

    def _adjust_pool_for_ou(self, pool: list, direction: str, ou_unit) -> list:
        """大小球一致性修正（检查③生成侧保证）：大球补高进球比分，小球滤净胜 3+。"""
        if not ou_unit:
            return pool
        if ou_unit.get("direction") == "over":
            if not any(item["total"] >= 3 for item in pool):
                pool = pool + [self._pool_item(score)
                               for score in self._high_goal_scores(direction)
                               if all(item["score"] != score for item in pool)]
        else:
            pool = [item for item in pool if abs(item["net_win"]) < 3]
            while len(pool) < POOL_MIN_SIZE:
                fallback = SCORE_POOL_BASE[(direction, "low")]
                for score in fallback:
                    item = self._pool_item(score)
                    if all(p["score"] != score for p in pool):
                        pool.append(item)
                    if len(pool) >= POOL_MIN_SIZE:
                        break
                else:
                    break
        return pool

    @staticmethod
    def _high_goal_scores(direction: str) -> tuple:
        """大球修正用的方向内高进球比分（相邻净胜，禁止跳向拼比分）。"""
        return {"win": ("2-1",), "draw": ("2-2",), "lose": ("1-2",)}[direction]

    def _append_penetration_scores(self, pool: list, direction: str,
                                   penetration: dict) -> list:
        """S1+S2 双命中：追加净胜 3+ 比分 + 反向冷门比分保护（0-1/1-0 类）。"""
        if not (penetration.get("S1") and penetration.get("S2")):
            return pool
        high, guard = (("3-0", "0-1") if direction == "win"
                       else ("0-3", "1-0") if direction == "lose" else (None, None))
        if high is None:
            return pool
        existing = {item["score"] for item in pool}
        additions = []
        if high not in existing:
            item = self._pool_item(high)
            item["tag"] = TAG_PENETRATION_HIGH
            additions.append(item)
        if guard not in existing:
            item = self._pool_item(guard)
            item["tag"] = TAG_UPSET_GUARD
            additions.append(item)
        return pool + additions

    # ---------------------------------------------------------------- 穿盘证据
    def _penetration(self, data: dict, ou_unit) -> dict:
        """穿盘证据 S1-S4（CONSISTENCY_CHECKS.md 检查四）与命中数。

        S1 基准去水主胜/客胜 ≥64%；S2 反大众实力差（anti_public.anchor2）；
        S3 幅度词（notes + 情报条目）；S4 大球方向或外推 ≥3 球。
        """
        synthesis = data.get("synthesis") or {}
        baseline = ((data.get("odds_verify") or {}).get("structure") or {}
                    ).get("baseline_probs") or {}
        recent = (data.get("fundamental") or {}).get("recent_form") or {}
        s1 = any((baseline.get(key) or 0) >= S1_MIN_PROB for key in ("win", "lose"))
        s2 = bool(((synthesis.get("anti_public") or {}).get("anchor2")))
        notes = " ".join(((synthesis.get("recommendation") or {}).get("notes") or []))
        intel_text = " ".join(
            (item.get("text") or "")
            for item in ((data.get("fundamental") or {}).get("intelligence")
                         or {}).get("items") or []
        )
        s3 = any(word in notes or word in intel_text
                 for word in AMPLITUDE_WORDS_S3)
        s4 = bool(ou_unit and ou_unit.get("direction") == "over") or (
            (recent.get("expected_total_goals") or 0) >= 3)
        signals = {"S1": s1, "S2": s2, "S3": s3, "S4": s4}
        return {**signals, "hits": sum(1 for v in signals.values() if v)}
