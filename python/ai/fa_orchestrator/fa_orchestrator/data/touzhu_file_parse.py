# -*- coding: utf-8 -*-
"""touzhu 文件块解析（立项四）。

文件职责：将 parse/results/{日期}/{match_id}.json 的 data.touzhu 块
（SKILL.md 契约：bf_trading/bf_large/bf_details）解析为可直接写入
MySQL（betfair_trading / betfair_price_ladder）的行结构。

不负责：文件读取与 JSON 装载（ingest_touzhu CLI）、SQL 语句组装与
写入、分析层解读（sentiment_agent.py）。

⚠️ 契约假设（2026-10-14 宿主定案：touzhu 数据由宿主直接写两表，
无需物理文件，本解析层降为备用文件通道；若宿主后续提供样例文件，
假设不符仅调整本层与测试）：
- bf_trading.summary / bf_large.summary 为三方向占比 {home, draw, away}
- bf_details 为价位档列表，每档 {price, back_volume, lay_volume}

缺失降级口径：子块/字段缺失时对应列取 None 并记入 missing，不臆造；
输出行结构与 betfair_trading（立项四加列 bookmaker_pnl）及
betfair_price_ladder 的列一一对应。
"""


def _num(value):
    """数值归一：int/float/str 数字 → float；None/非法 → None。"""
    if value is None:
        return None
    try:
        return float(value)
    except (TypeError, ValueError):
        return None


def _pct_block(summary):
    """summary 三方向占比 → (home_pct, draw_pct, away_pct)，缺失全 None。"""
    if not isinstance(summary, dict):
        return None, None, None
    return (_num(summary.get("home")), _num(summary.get("draw")),
            _num(summary.get("away")))


def _trading_row(trade_type, block):
    """bf_trading/bf_large 子块 → betfair_trading 行（缺失字段 None）。"""
    block = block if isinstance(block, dict) else {}
    home_pct, draw_pct, away_pct = _pct_block(block.get("summary"))
    return {
        "trade_type": trade_type,
        "total_volume": _num(block.get("total_volume")),
        "home_pct": home_pct,
        "draw_pct": draw_pct,
        "away_pct": away_pct,
        # bookmaker_pnl 仅 bf_trading（标准盘）携带；large 行库内无此维度
        "bookmaker_pnl": _num(block.get("bookmaker_pnl")),
    }


def _ladder_rows(details):
    """bf_details 价位档列表 → betfair_price_ladder 行列表。"""
    rows = []
    if not isinstance(details, list):
        return rows
    for item in details:
        if not isinstance(item, dict):
            continue
        rows.append({
            "trade_type": "standard",
            "price": _num(item.get("price")),
            "back_volume": _num(item.get("back_volume")),
            "lay_volume": _num(item.get("lay_volume")),
        })
    return [row for row in rows if row["price"] is not None]


def touzhu_block(touzhu):
    """data.touzhu 块 → 入库行结构。

    输入：data.touzhu 层 dict（可为 None）。
    输出：{"trading_rows": [betfair_trading 行],
           "ladder_rows": [betfair_price_ladder 行],
           "missing": [缺失子块/字段标注]}
    失败行为：输入 None/空时整块缺失标注 ["touzhu"]（不展开子块），
    子块/字段缺失分别标注，均不抛异常。
    """
    if not isinstance(touzhu, dict) or not touzhu:
        return {"trading_rows": [], "ladder_rows": [], "missing": ["touzhu"]}
    missing = []
    trading_rows = []
    bf_trading = touzhu.get("bf_trading")
    if bf_trading is None:
        missing.append("bf_trading")
    else:
        row = _trading_row("standard", bf_trading)
        if row["bookmaker_pnl"] is None:
            missing.append("bf_trading.bookmaker_pnl")
        trading_rows.append(row)

    bf_large = touzhu.get("bf_large")
    if bf_large is None:
        missing.append("bf_large")
    elif _num(bf_large.get("total_volume") if isinstance(bf_large, dict) else None) is not None:
        trading_rows.append(_trading_row("large", bf_large))

    details = touzhu.get("bf_details")
    if details is None:
        missing.append("bf_details")
    ladder_rows = _ladder_rows(details)

    return {"trading_rows": trading_rows, "ladder_rows": ladder_rows,
            "missing": missing}
