# -*- coding: utf-8 -*-
"""FundamentalAgent：第一阶段基本面深度分析（PRD-12 F-10，SKILL v7.12 规则节点化）。

职责边界：
- 负责：情报利弊比量化与均势约束（v7.7）、初盘定位与同盘路提示（v7.4）、
  历史交锋压制力分级与时效衰减（v7.10）、近期状态进球外推与火热纪律（v7.2）、
  杯赛/国际赛大小球修正（v7.8/v7.11）、赛程影响；五维权重按 v7.12 固定配置。
- 不负责：赔率验证（odds_verify_agent.py）、数据获取（data_source.py）、
  推荐结构判定（ISS-FA3 SynthesisAgent）。

纯规则计算，不调用 LLM 与外部服务；输入数据块缺失时按 v7.6 降级标注，禁止编造。
"""

# 第一阶段五维权重（SKILL v7.12：情报30/初盘25/交锋15/状态20/赛程10）
PHASE1_WEIGHTS = {
    "intelligence": 30,
    "initial_position": 25,
    "h2h": 15,
    "recent_form": 20,
    "schedule": 10,
}

GOAL_EXTRAPOLATION_FACTOR = 0.9      # 预期总进球 = 两队场均进球之和 × 0.9（v7.2）
HOT_GOAL_THRESHOLD = 3.0             # 近5场场均进球 ≥3 判"状态火热"
FIRE_DISCOUNT_CAP_PCT = 30           # 火热方成色折扣上限（v7.2 弗赖堡场固化）
ATTACK_COLLAPSE_THRESHOLD = 3.5      # 对攻崩坏形态：近5场场均总进球 ≥3.5（v7.8）
BALANCE_MARGIN = 1                   # 利弊比差值 ≤1 判均势（v7.7 013 场事故固化）
DENSE_DIRECTION_COUNT = 3            # 单方向利多同向 ≥3 条标注"战力线索密集"（v7.7）
SAME_HANDICAP_HOME_WIN_RATE = 0.80   # 主队同盘赢盘占比 ≥80% → 深盘适应性强（v7.4）
SAME_HANDICAP_AWAY_LOSE_RATE = 0.60  # 客队同盘输盘占比 ≥60% → 下盘疲软（v7.4）
SUPPRESS_WIN_STREAK = 3              # 连续胜 ≥3 场 → 强压制（v7.10，含平即断）
SUPPRESS_WIN_RATE = 0.70             # 近6次交锋胜率 ≥70%（胜不含平）→ 强压制（v7.10）
H2H_RECENT_YEARS = 5                 # 历史压制力仅统计近5年内交锋（v7.10 时效衰减）

# 盘口变化方向（v7.4 初盘定位）
MOVEMENT_UNKNOWN = "未知"
MOVEMENT_RAISE = "升盘"
MOVEMENT_DROP = "降盘"
MOVEMENT_FLAT = "持平"

BALANCE_CONCLUSION = "均势，方向交由赔率验证层"
HOME_TILT_CONCLUSION = "利好天平倾向主队"
AWAY_TILT_CONCLUSION = "利好天平倾向客队"
NOTE_SAME_HANDICAP_MISSING = "ℹ️ 7M 同盘路统计缺失"
NOTE_DECAY_DOWNGRADE = "压制方当期出现衰减项，压制力降级为弱压制/参考项"
FIRE_NOTE = f"状态火热一方火力保留（对手成色折扣上限{FIRE_DISCOUNT_CAP_PCT}%）"
CUP_ANCHOR_NOTE = "杯赛/锦标赛：数据主锚（场均进球和×0.9）降级为参考项，市场升线动作升为共同主锚"
CUP_COLLAPSE_NOTE = "对攻崩坏形态（近5场场均≥3.5球）不得降权"
INTERNATIONAL_ANCHOR_NOTE = (
    "国际赛/友谊赛：0.9 外推仅作参考项（国际赛方差大），大小球轻仓≤0.3 单位或弃配"
)


class FundamentalAgent:
    """基本面分析 Agent：输入单场基本面数据块，输出第一阶段五维结论。

    状态约束：无状态；纯函数式规则计算，可并发调用。
    """

    def analyze(self, match_data: dict, current_year: int = 2026) -> dict:
        """执行第一阶段分析。

        输入：match_data（intelligence/initial_position/h2h/recent_form/schedule 五块）。
        输出：第一阶段结论 dict（五维子结论 + phase + weights）。
        失败行为：数据块缺失按空数据处理并降级标注，不抛异常、不编造数据。
        """
        return {
            "phase": "fundamental",
            "weights": dict(PHASE1_WEIGHTS),
            "intelligence": self._analyze_intelligence(match_data.get("intelligence") or {}),
            "initial_position": self._analyze_initial_position(
                match_data.get("initial_position") or {}
            ),
            "h2h": self._analyze_h2h(match_data.get("h2h") or {}, current_year),
            "recent_form": self._analyze_recent_form(
                match_data.get("recent_form") or {},
                (match_data.get("schedule") or {}).get("competition"),
            ),
            "schedule": self._analyze_schedule(match_data.get("schedule") or {}),
        }

    def _analyze_intelligence(self, intel: dict) -> dict:
        """情报利弊比量化（v7.4 五向计数 + v7.7 均势约束）。"""
        items = intel.get("items") or []
        counts = {
            "home_favor": 0, "away_favor": 0,
            "home_against": 0, "away_against": 0, "general": 0,
        }
        for item in items:
            self._count_intel_item(counts, item)
        net_favor = (counts["home_favor"] + counts["away_against"]) - (
            counts["away_favor"] + counts["home_against"]
        )
        balance = abs(net_favor) <= BALANCE_MARGIN
        return {
            **counts,
            "net_favor": net_favor,
            "balance": balance,
            "dense_direction": self._dense_direction(counts),
            "conclusion": BALANCE_CONCLUSION if balance else self._tilt_conclusion(net_favor),
        }

    @staticmethod
    def _count_intel_item(counts: dict, item: dict) -> None:
        """按 7M 情报规则归类单条情报：yl→利、wl→弊、null→通用（v7.4）。"""
        party = item.get("party")
        sentiment = item.get("sentiment")
        if sentiment == "general":
            counts["general"] += 1
        elif sentiment == "favor":
            counts[f"{party}_favor"] += 1
        elif sentiment == "against":
            counts[f"{party}_against"] += 1

    @staticmethod
    def _dense_direction(counts: dict):
        """单方向利多同向≥3 条且对方不足 3 条时标注密集方向（v7.7）。"""
        home_dense = counts["home_favor"] >= DENSE_DIRECTION_COUNT
        away_dense = counts["away_favor"] >= DENSE_DIRECTION_COUNT
        if home_dense and not away_dense:
            return "home"
        if away_dense and not home_dense:
            return "away"
        return None

    @staticmethod
    def _tilt_conclusion(net_favor: int) -> str:
        """非均势时按利好净差方向输出天平结论。"""
        return HOME_TILT_CONCLUSION if net_favor > 0 else AWAY_TILT_CONCLUSION

    @staticmethod
    def _analyze_initial_position(pos: dict) -> dict:
        """初盘定位（v7.4）：盘口变化方向 + 同盘路历史适应提示。"""
        initial_handicap = pos.get("initial_handicap")
        current_handicap = pos.get("current_handicap")
        if initial_handicap is None or current_handicap is None:
            movement = MOVEMENT_UNKNOWN
        elif current_handicap > initial_handicap:
            movement = MOVEMENT_RAISE
        elif current_handicap < initial_handicap:
            movement = MOVEMENT_DROP
        else:
            movement = MOVEMENT_FLAT
        hint, note = _same_handicap_hint(pos.get("same_handicap"))
        return {
            "initial_handicap": initial_handicap,
            "current_handicap": current_handicap,
            "handicap_movement": movement,
            "same_handicap_hint": hint,
            "note": note,
        }

    def _analyze_h2h(self, h2h: dict, current_year: int) -> dict:
        """历史交锋压制力分级（v7.10：强/弱/无 + 近5年窗口 + 时效衰减）。"""
        window = [r for r in (h2h.get("records") or [])
                  if r.get("year", 0) >= current_year - H2H_RECENT_YEARS]
        window.sort(key=lambda r: r.get("year", 0))
        results = [r.get("result") for r in window]
        suppression = self._classify_suppression(results, h2h.get("decay_flags") or [])
        return {
            "window_records_count": len(window),
            "streak_wins": _latest_win_streak(results),
            "win_rate": _win_rate(results),
            "suppression": suppression,
            "decayed": bool(h2h.get("decay_flags")),
            "note": NOTE_DECAY_DOWNGRADE if h2h.get("decay_flags") else "",
        }

    @staticmethod
    def _classify_suppression(results: list, decay_flags: list) -> str:
        """压制力分级：连续胜≥3 或 胜率≥70% → strong；不败型 → weak（v7.10）。"""
        if not results:
            return "none"
        strong = (_latest_win_streak(results) >= SUPPRESS_WIN_STREAK
                  or _win_rate(results) >= SUPPRESS_WIN_RATE)
        if strong and decay_flags:
            return "weak"  # 时效衰减：强压制降级为弱压制/参考项
        if strong:
            return "strong"
        if "L" not in results:
            return "weak"  # 无负但不满足强压制的"不败型"战绩
        return "none"

    def _analyze_recent_form(self, form: dict, competition) -> dict:
        """近期状态与进球外推（v7.2 主锚 + v7.8/v7.11 修正）。"""
        home_stats = _side_stats((form.get("home") or {}).get("matches") or [])
        away_stats = _side_stats((form.get("away") or {}).get("matches") or [])
        expected_total = (
            (home_stats["avg_scored"] + away_stats["avg_scored"]) * GOAL_EXTRAPOLATION_FACTOR
        )
        hot_sides = [side for side, stats in
                     (("home", home_stats), ("away", away_stats))
                     if stats["recent5_avg"] >= HOT_GOAL_THRESHOLD]
        correction, anchor_note = self._competition_correction(
            competition, home_stats, away_stats
        )
        return {
            "home_avg_scored": home_stats["avg_scored"],
            "home_avg_conceded": home_stats["avg_conceded"],
            "away_avg_scored": away_stats["avg_scored"],
            "away_avg_conceded": away_stats["avg_conceded"],
            "home_recent5_avg": home_stats["recent5_avg"],
            "away_recent5_avg": away_stats["recent5_avg"],
            "expected_total_goals": expected_total,
            "hot_sides": hot_sides,
            "fire_note": FIRE_NOTE if hot_sides else "",
            "correction": correction,
            "anchor_note": anchor_note,
        }

    @staticmethod
    def _competition_correction(competition, home_stats: dict, away_stats: dict) -> tuple:
        """赛事类型修正（v7.8 杯赛 / v7.11 国际赛），返回 (correction, anchor_note)。"""
        if competition == "cup":
            note = CUP_ANCHOR_NOTE
            collapsed = any(s["recent5_avg"] >= ATTACK_COLLAPSE_THRESHOLD
                            for s in (home_stats, away_stats))
            if collapsed:
                note += "；" + CUP_COLLAPSE_NOTE
            return "cup", note
        if competition == "international":
            return "international", INTERNATIONAL_ANCHOR_NOTE
        return None, ""

    @staticmethod
    def _analyze_schedule(schedule: dict) -> dict:
        """赛程影响（10%）：赛事类型 + 未来赛程密度透出。"""
        future = schedule.get("future") or []
        return {
            "competition": schedule.get("competition"),
            "level_gap": bool(schedule.get("level_gap")),
            "future_count": len(future),
            "note": "",
        }


def _same_handicap_hint(same_handicap):
    """同盘路历史提示（v7.4）：主深盘适应 / 客下盘疲软；缺失返回 (None, 缺失标注)。"""
    if not same_handicap:
        return None, NOTE_SAME_HANDICAP_MISSING
    home = same_handicap.get("home") or {}
    away = same_handicap.get("away") or {}
    home_total = sum(home.values())
    away_total = sum(away.values())
    if home_total and home.get("win", 0) / home_total >= SAME_HANDICAP_HOME_WIN_RATE:
        return "主队深盘适应性强，支持上盘", ""
    if away_total and away.get("lose", 0) / away_total >= SAME_HANDICAP_AWAY_LOSE_RATE:
        return "客队下盘疲软，支持上盘", ""
    return None, ""


def _side_stats(matches: list) -> dict:
    """单队近期状态统计：外推与火热判定均以正赛口径统计（热身赛分开，v7.2）。"""
    official = [m for m in matches if not m.get("is_friendly")]
    scored = [m.get("scored", 0) for m in official]
    conceded = [m.get("conceded", 0) for m in official]
    recent5 = scored[-5:]
    count = len(scored)
    return {
        "avg_scored": sum(scored) / count if count else 0.0,
        "avg_conceded": sum(conceded) / count if conceded else 0.0,
        "recent5_avg": sum(recent5) / len(recent5) if recent5 else 0.0,
    }


def _latest_win_streak(results: list) -> int:
    """从最近一场往回数连续胜场数（含平即断，v7.10 口径）。"""
    streak = 0
    for result in reversed(results):
        if result != "W":
            break
        streak += 1
    return streak


def _win_rate(results: list) -> float:
    """近6次交锋胜率（胜不含平，v7.10 口径）。"""
    if not results:
        return 0.0
    return sum(1 for r in results if r == "W") / len(results)
