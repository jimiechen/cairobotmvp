# -*- coding: utf-8 -*-
"""历史推演（PRD-12 F-16，SKILL v7.4 M5 能力）：升降盘时序与投票热度漂移推导。

职责边界：
- 负责：把 MySQL 快照时序（tracks/drift）推导为可读的趋势结论，供第二阶段
  赔率验证判断升降盘趋势与热度走向。
- 不负责：时序数据获取（data_source.py）、任何盘口/赔率判定（odds_verify_agent.py）。

数据缺失（文件模式无快照、历史推演表无记录）时明确标注"无法推演时序"，
禁止编造趋势结论。
"""

# 热度漂移平稳带（百分点）：|home_pp| ≤ 该值视为热度平稳
HEAT_DRIFT_STABLE_BAND = 1.0

# 无法推演时的固定标注（v7.4 M5：缺失时明确标注，禁止编造）
NOTE_UNAVAILABLE = "无法推演时序"


def deduce_trend(tracks) -> dict:
    """推导升降盘时序趋势。

    输入：tracks（快照轨迹列表，元素含 kind: asian/overunder 与 change 数值）。
    输出：趋势结论 dict（available + 各类计数）。
    失败行为：tracks 为 None 或空列表时返回 available=False 并标注无法推演。
    """
    if not tracks:
        return {"available": False, "note": NOTE_UNAVAILABLE}
    asian_deepen = sum(1 for t in tracks if t.get("kind") == "asian" and t.get("change", 0) > 0)
    overunder_raise = sum(
        1 for t in tracks if t.get("kind") == "overunder" and t.get("change", 0) > 0
    )
    return {
        "available": True,
        "asian_deepen_count": asian_deepen,
        "overunder_raise_count": overunder_raise,
        "total": len(tracks),
        "note": "",
    }


def deduce_heat_drift(drift) -> dict:
    """推导投票热度漂移走向。

    输入：drift（{"home_pp": 百分点, "away_pp": 百分点}，正值为热度上升）。
    输出：走向结论 dict（available + home_pp/away_pp + trend）。
    失败行为：drift 为 None 时返回 available=False 并标注无法推演。
    """
    if not drift:
        return {"available": False, "note": NOTE_UNAVAILABLE}
    home_pp = drift.get("home_pp", 0.0)
    away_pp = drift.get("away_pp", 0.0)
    return {
        "available": True,
        "home_pp": home_pp,
        "away_pp": away_pp,
        "trend": _classify_drift(home_pp),
        "note": "",
    }


def _classify_drift(home_pp: float) -> str:
    """按主队热度漂移百分点分类热度走向。"""
    if home_pp > HEAT_DRIFT_STABLE_BAND:
        return "home_heat_rising"
    if home_pp < -HEAT_DRIFT_STABLE_BAND:
        return "home_heat_cooling"
    return "stable"
