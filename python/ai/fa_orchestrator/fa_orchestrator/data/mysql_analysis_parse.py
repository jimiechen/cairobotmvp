# -*- coding: utf-8 -*-
"""分析查询结果解析器（mysql_analysis_source 的纯函数层）。

职责边界：
- 负责：DB 行 → fixture 契约结构的纯函数归一——数值/时间格式化、比分解析、
  盘口初即盘配对、投票池解构、情报分段、变盘轨迹与热度漂移推算。
- 不负责：数据库连接与 SQL 执行（mysql_analysis_source.py）、
  分析计算与趋势判断（analysis/ 层）。

字段口径与降级标注（2026-10-08 workbench 只读探测校准，R-04）：
- bk 用 bookmaker.name 原名（与 fixture 口径一致；name_alias 为脱敏名）
- 库内无对应维度、禁止臆造的字段由调用方输出恒值并注明缺口：
  Q10 decay_flags 恒 []、Q12 level_gap 恒 False、
  Q3 is_final_snapshot 恒 False（无终盘标志）
"""
import re
from datetime import date, datetime

# 友谊赛判定关键词（team_recent_form.competition 含'友谊'计为热身赛）
FRIENDLY_KEYWORD = "友谊"

# 情报正文的主客标记（intelligence_text.content 分段解析依据）
INTEL_MARKERS = {"【主队】": "home", "【客队】": "away"}


def num(value):
    """Decimal/None → float/None（DB 数值列统一归一，对齐 fixture 数字口径）。"""
    return None if value is None else float(value)


def fmt_time(value):
    """datetime/str → 'MM-DD HH:MM'；None → None（对齐 fixture 时间格式）。"""
    if value is None:
        return None
    if isinstance(value, datetime):
        return value.strftime("%m-%d %H:%M")
    text = str(value)
    return text[5:16] if len(text) >= 16 else text


def to_day(value):
    """date/datetime/str → date（未来赛程天数差计算用）。"""
    if isinstance(value, datetime):
        return value.date()
    if isinstance(value, date):
        return value
    return date.fromisoformat(str(value)[:10])


def split_score(score):
    """'主:客' 比分字符串 → (主分, 客分)；格式异常返回 (None, None)。"""
    parts = str(score or "").split(":")
    if len(parts) != 2:
        return None, None
    try:
        return int(parts[0]), int(parts[1])
    except ValueError:
        return None, None


def pair_odds(rows, fields, final_flag, history=False) -> list:
    """按公司归并初盘（is_initial=1 首行）与即时盘（最新非初盘行）。

    输入 rows 已按公司、快照时间排序（契约 ORDER BY 保证）；
    payout/snapshot_time 取 current 优先（回退 initial）；无初盘或
    无即时盘时对应块 None；final_flag=True 时补 is_final_snapshot 恒 False。
    history=True（仅欧赔，ADR-018）时附全状态序列键。
    """
    by_bookmaker = {}
    for row in rows:
        by_bookmaker.setdefault(row["bk"], []).append(row)
    paired = []
    for bk, snapshots in by_bookmaker.items():
        initial = next((r for r in snapshots if r["is_initial"]), None)
        current = next((r for r in reversed(snapshots) if not r["is_initial"]), None)
        item = {
            "bk": bk,
            "initial": {f: num(initial[f]) for f in fields} if initial else None,
            "current": {f: num(current[f]) for f in fields} if current else None,
        }
        ref = current or initial
        if ref is not None:
            if "payout" in ref:
                item["payout"] = num(ref["payout"])
            item["snapshot_time"] = fmt_time(ref["snapshot_time"])
        if history:
            item["history"] = history_states(snapshots, fields)
        if final_flag:
            item["is_final_snapshot"] = False   # 库无终盘标志（缺口标注）
        paired.append(item)
    return paired


def history_states(snapshots, fields) -> list:
    """欧赔全状态序列（ADR-018）：按赔率值元组保序去重，每态取末次行。

    首态=初盘；序按各态**首次出现**定序，末态=最后新增态——若最新赔率
    回摆历史态（真实库 2951 澳门存在），current 对应 history 中间项
    （该态 payout/快照时间仍取末次行）；时间口径 fmt_time 对齐 fixture。
    """
    states = {}
    for row in snapshots:
        key = tuple(num(row[f]) for f in fields)
        states[key] = row   # dict 保插入序：首现定序、末次覆盖取值
    return [
        {**{f: num(row[f]) for f in fields},
         "payout": num(row["payout"]),
         "snapshot_time": fmt_time(row["snapshot_time"])}
        for row in states.values()
    ]


def had_block(row, suffix):
    """vote_snapshot h/d/a 前缀列 → {win, draw, lose}（suffix 为列后缀）。"""
    return {"win": num(row[f"h_{suffix}"]),
            "draw": num(row[f"d_{suffix}"]),
            "lose": num(row[f"a_{suffix}"])}


def bf_block(rows) -> dict:
    """必发全行序列 → {total_volume, direction, trend, large_volume}。

    standard 组：direction 按最新占比最大者（缺失视为 0）；trend 按
    direction 方向占比的有效首尾差（升 rising / 降 falling / 单时点或
    等值 None，口径对齐 D12 首尾差，None 为缺口标注不臆造）；total_volume
    取最新快照。large 组：占比列库内恒空（2026-10-12 探查），仅大额成交量
    有值，large_volume 取最新，无 large 行时为 None。
    """
    standard = [r for r in rows if r["trade_type"] == "standard"]
    large = [r for r in rows if r["trade_type"] == "large"]
    latest = standard[-1]
    pcts = {"home": num(latest["home_pct"]) or 0,
            "draw": num(latest["draw_pct"]) or 0,
            "away": num(latest["away_pct"]) or 0}
    direction = max(pcts, key=pcts.get)
    values = [v for r in standard
              for v in [num(r[f"{direction}_pct"])] if v is not None]
    if len(values) >= 2 and values[-1] != values[0]:
        trend = "rising" if values[-1] > values[0] else "falling"
    else:
        trend = None
    return {"total_volume": num(latest["total_volume"]),
            "direction": direction,
            "trend": trend,
            "large_volume": num(large[-1]["total_volume"]) if large else None}


def split_intel_items(content) -> list:
    """Q9 前瞻正文 → items（按【主队】/【客队】标记切分段落）。

    标记后的段落 party 与标记一致；无标记正文 party=None；
    sentiment 库内无维度不输出（缺口标注）。
    """
    items = []
    party = None
    for chunk in re.split(r"(【主队】|【客队】)", str(content or "")):
        if chunk in INTEL_MARKERS:
            party = INTEL_MARKERS[chunk]
        elif chunk.strip():
            items.append({"party": party, "text": chunk.strip()})
    return items


def track_changes(kind, field, rows) -> list:
    """D9 变盘轨迹：按参考公司分组，各以自身初盘为基线，盘口变化记 {kind, bk, change:±1}。

    多公司口径（四家参考公司）：分组保证公司间基线互不串扰；
    行需按公司内快照时间升序（SQL 已排序）；等值不计、基线跟随最近一次变化。
    """
    tracks = []
    baselines = {}
    for row in rows:
        bk = row.get("bk")
        value = num(row[field])
        if row["is_initial"]:
            baselines[bk] = value
            continue
        baseline = baselines.get(bk)
        if value is None or baseline is None:
            continue
        if value != baseline:
            tracks.append({"kind": kind, "bk": bk,
                           "change": 1 if value > baseline else -1})
            baselines[bk] = value
    return tracks


def drift_of(rows, side_name):
    """D12 热度漂移：指定队名侧的首尾快照 cold_hot 差；无匹配 → None。

    cold_hot=None（热度未计算的快照行，真实库存在）按缺口排除，
    取有效快照首尾差；有效快照仅 1 个时差为 0.0。
    """
    values = [row["cold_hot"] for row in rows
              if row["side"] == side_name and row["cold_hot"] is not None]
    if not values:
        return None
    if len(values) == 1:
        return 0.0
    return float(values[-1] - values[0])
