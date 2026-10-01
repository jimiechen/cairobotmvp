# -*- coding: utf-8 -*-
"""fid → match.id 四路径兜底解析（PRD-12 F-07，需求文档 §2.3）。

解析链：fid → site_id_mappings → matches → `match` 表落点。
本模块只负责纯内存的候选行匹配逻辑（四路径兜底），
不负责数据库查询（由 data_source 提供 rows）。

兜底顺序（与上游 resolve_match_id_by_500com 对齐）：
1. 场次编号 + 日期窗口（match_date ±2 天，play_type='jingcai'）
2. 仅场次编号
3. 队名归一（去空格 + 忽略大小写，双队同时匹配）
4. 别名学习（home_alias / away_alias 命中）
全部未命中抛 MatchResolveError，禁止返回 None 静默通过。
"""
from datetime import date

# 路径1 日期窗口边界（需求文档 §2.3：match_date±2 天）
DATE_WINDOW_DAYS = 2

# 允许的玩法过滤（需求文档 §2.3：play_type='jingcai'）
ALLOWED_PLAY_TYPE = "jingcai"


class MatchResolveError(Exception):
    """四路径全部未命中时抛出，错误信息必须含场次编号便于排查。"""


def _normalize_team_name(team_name: str) -> str:
    """队名归一：去首尾与内部空格、忽略大小写。"""
    return team_name.strip().replace(" ", "").casefold()


def _parse_date(date_text: str) -> date:
    return date.fromisoformat(str(date_text)[:10])


def _by_cc_and_window(rows, match_no, match_date):
    """路径1：cc + match_date ±2 天窗口，多条命中取日期最近。"""
    window_center = _parse_date(match_date)
    hits = [
        row for row in rows
        if row.get("cc") == match_no
        and row.get("play_type") == ALLOWED_PLAY_TYPE
        and abs((_parse_date(row["match_date"]) - window_center).days) <= DATE_WINDOW_DAYS
    ]
    if not hits:
        return None
    return min(hits, key=lambda row: abs((_parse_date(row["match_date"]) - window_center).days))


def _by_cc_only(rows, match_no):
    """路径2：仅按场次编号匹配（无日期窗口约束）。"""
    hits = [
        row for row in rows
        if row.get("cc") == match_no and row.get("play_type") == ALLOWED_PLAY_TYPE
    ]
    return hits[0] if hits else None


def _by_team_names(rows, home_team, away_team):
    """路径3：队名归一后主客双队同时匹配。"""
    want_home = _normalize_team_name(home_team)
    want_away = _normalize_team_name(away_team)
    for row in rows:
        if row.get("play_type") != ALLOWED_PLAY_TYPE:
            continue
        got_home = _normalize_team_name(row.get("home_team", ""))
        got_away = _normalize_team_name(row.get("away_team", ""))
        if got_home == want_home and got_away == want_away:
            return row
    return None


def _by_alias(rows, home_team, away_team):
    """路径4：别名表命中（主或客任一队命中即视为同场，最后兜底）。"""
    want_home = _normalize_team_name(home_team)
    want_away = _normalize_team_name(away_team)
    for row in rows:
        if row.get("play_type") != ALLOWED_PLAY_TYPE:
            continue
        home_aliases = {_normalize_team_name(a) for a in row.get("home_alias", [])}
        away_aliases = {_normalize_team_name(a) for a in row.get("away_alias", [])}
        if want_home in home_aliases or want_away in away_aliases:
            return row
    return None


def resolve_match_id(match_no, match_date, home_team, away_team, rows) -> int:
    """按四路径兜底解析 `match` 表落点 id。

    输入：match_no（如 周二008）、match_date（竞彩日或开球日）、主客队名、
          `match` 表候选行（由 data_source 查询提供）。
    输出：match.id（int）。
    失败行为：四路径全部未命中抛 MatchResolveError（含场次编号）。
    """
    ordered_paths = (
        _by_cc_and_window(rows, match_no, match_date),
        _by_cc_only(rows, match_no),
        _by_team_names(rows, home_team, away_team),
        _by_alias(rows, home_team, away_team),
    )
    for hit in ordered_paths:
        if hit is not None:
            return int(hit["id"])
    raise MatchResolveError(
        f"fid 解析失败：场次编号={match_no} 日期={match_date} "
        f"主队={home_team} 客队={away_team}（四路径全部未命中）"
    )
