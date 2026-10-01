# -*- coding: utf-8 -*-
"""赔率数学工具（PRD-12 ISS-FA3）：欧赔/让球赔率去水隐含概率计算。

职责边界：仅放无状态纯函数；不含业务规则（规则在 sentiment/synthesis agent）。
"""


def implied_probabilities(odds_rows: list) -> dict:
    """计算去水隐含概率（%）。

    输入：赔率行列表，每行 {"current": {"win","draw","lose"}}（欧赔各家/让球单一行）。
    输出：{"win": p, "draw": p, "lose": p}，三项归一后百分数，保留 1 位。
    失败行为：行缺失或赔率非法时返回空 dict（调用方按缺失降级，不编造）。
    """
    raw_totals = {"win": 0.0, "draw": 0.0, "lose": 0.0}
    counted = 0
    for row in odds_rows:
        current = (row or {}).get("current") or {}
        if not all(current.get(key) for key in raw_totals):
            continue
        for key, value in current.items():
            if key in raw_totals:
                raw_totals[key] += 1.0 / float(value)
        counted += 1
    if not counted:
        return {}
    total_sum = sum(raw_totals.values())
    if total_sum <= 0:
        return {}
    return {key: round(value / total_sum * 100.0, 1) for key, value in raw_totals.items()}


def dual_path(leg_odds: dict, implied: dict, cover: bool) -> str | None:
    """双选收益比路径判定（v7.12 检查⑥：A/B/C 三路径）。

    输入：leg_odds 双选腿赔率（{腿: 赔率}）、implied 对应去水概率、cover 是否低赔补位场景。
    输出："A"（两腿 ≥2.0 自动合规）/"C"（补位：最低腿 ≥1.8 且 EV ≥1.0）/"B"（EV ≥2.1）/
    None（不合规，调用方按互斥原则砍剩单选并写明量化原因）。
    """
    legs = list(leg_odds)
    if len(legs) < 2:
        return None
    if all(leg_odds[leg] >= PATH_A_MIN_LEG for leg in legs):
        return "A"
    expected_value = sum(implied.get(leg, 0.0) / 100.0 * leg_odds[leg] for leg in legs)
    if cover and min(leg_odds[leg] for leg in legs) >= PATH_C_MIN_LEG \
            and expected_value >= PATH_C_MIN_EV:
        return "C"
    if expected_value >= PATH_B_MIN_EV:
        return "B"
    return None


# 双选路径阈值（v7.12）：A 两腿≥2.0；B EV≥2.1；C 低赔补位最低腿≥1.8 且 EV≥1.0
PATH_A_MIN_LEG = 2.0
PATH_B_MIN_EV = 2.1
PATH_C_MIN_LEG = 1.8
PATH_C_MIN_EV = 1.0
