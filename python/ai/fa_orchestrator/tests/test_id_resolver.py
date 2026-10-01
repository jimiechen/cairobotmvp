# -*- coding: utf-8 -*-
"""fid → match.id 四路径兜底解析测试（PRD-12 F-07，需求文档 §2.3）。

兜底顺序：场次编号+日期窗口 → 场次编号 → 队名归一 → 别名学习。
全部未命中必须抛异常，禁止返回 None 静默通过。
"""
import pytest

from fa_orchestrator.data.id_resolver import MatchResolveError, resolve_match_id


def _row(row_id, cc, match_date, home, away, home_alias=None, away_alias=None):
    """构造 `match` 表候选行。"""
    return {
        "id": row_id,
        "cc": cc,
        "match_date": match_date,
        "play_type": "jingcai",
        "home_team": home,
        "away_team": away,
        "home_alias": home_alias or [],
        "away_alias": away_alias or [],
    }


class TestPath1CcAndDateWindow:
    """路径1：cc（场次编号）+ match_date ±2 天窗口 + play_type='jingcai'。"""

    def test_窗口内唯一命中(self):
        rows = [_row(1, "周二008", "2026-09-30", "北城红鹰", "南岭蓝鲨")]
        got = resolve_match_id(
            match_no="周二008", match_date="2026-09-30",
            home_team="北城红鹰", away_team="南岭蓝鲨", rows=rows,
        )
        assert got == 1

    def test_跨日竞彩日仍在窗口内命中(self):
        # 竞彩日 09-29 查询、开球日 09-30 入库：差 1 天 ≤ 2 天窗口
        rows = [_row(1, "周二008", "2026-09-30", "北城红鹰", "南岭蓝鲨")]
        got = resolve_match_id(
            match_no="周二008", match_date="2026-09-29",
            home_team="北城红鹰", away_team="南岭蓝鲨", rows=rows,
        )
        assert got == 1

    def test_窗口边界差两天命中(self):
        rows = [_row(7, "周二014", "2026-10-01", "长风队", "短雨队")]
        got = resolve_match_id(
            match_no="周二014", match_date="2026-09-29",
            home_team="长风队", away_team="短雨队", rows=rows,
        )
        assert got == 7

    def test_多条候选取日期最近(self):
        rows = [
            _row(11, "周二012", "2026-10-01", "星城流星", "月城潮汐"),
            _row(12, "周二012", "2026-09-30", "星城流星", "月城潮汐"),
        ]
        got = resolve_match_id(
            match_no="周二012", match_date="2026-09-29",
            home_team="星城流星", away_team="月城潮汐", rows=rows,
        )
        assert got == 12

    def test_非竞彩玩法不参与命中(self):
        rows = [{
            "id": 1, "cc": "周二008", "match_date": "2026-09-30",
            "play_type": "zqdc", "home_team": "北城红鹰", "away_team": "南岭蓝鲨",
            "home_alias": [], "away_alias": [],
        }]
        with pytest.raises(MatchResolveError):
            resolve_match_id(
                match_no="周二008", match_date="2026-09-30",
                home_team="北城红鹰", away_team="南岭蓝鲨", rows=rows,
            )


class TestPath2CcOnly:
    """路径2：cc 命中但日期超出窗口。"""

    def test_超窗口回落到仅场次编号(self):
        rows = [_row(2, "周二009", "2026-10-05", "东川闪电", "西湖巨浪")]
        got = resolve_match_id(
            match_no="周二009", match_date="2026-09-29",
            home_team="东川闪电", away_team="西湖巨浪", rows=rows,
        )
        assert got == 2


class TestPath3TeamNormalize:
    """路径3：cc 不命中，按队名归一（去空格 + 忽略大小写）匹配。"""

    def test_队名归一双队命中(self):
        rows = [_row(3, "周六003", "2026-09-30", " 中原猛虎 ", "滨海飞鱼")]
        got = resolve_match_id(
            match_no="周二010", match_date="2026-09-29",
            home_team="中原猛虎", away_team="滨海飞鱼", rows=rows,
        )
        assert got == 3

    def test_仅主队匹配不通过(self):
        rows = [_row(3, "周六003", "2026-09-30", "中原猛虎", "另一队")]
        with pytest.raises(MatchResolveError):
            resolve_match_id(
                match_no="周二010", match_date="2026-09-29",
                home_team="中原猛虎", away_team="滨海飞鱼", rows=rows,
            )


class TestPath4Alias:
    """路径4：队名不匹配，按别名（自动学习）命中。"""

    def test_别名命中主队(self):
        rows = [_row(4, "周二011", "2026-09-30", "高原之鹰", "平原野牛",
                     home_alias=["高原雄鹰"])]
        got = resolve_match_id(
            match_no="周二099", match_date="2026-09-29",
            home_team="高原雄鹰", away_team="平原野牛", rows=rows,
        )
        assert got == 4

    def test_别名命中客队(self):
        rows = [_row(5, "周二012", "2026-09-30", "星城流星", "月城潮汐",
                     away_alias=["潮汐队"])]
        got = resolve_match_id(
            match_no="周二098", match_date="2026-09-29",
            home_team="星城队伍", away_team="潮汐队", rows=rows,
        )
        assert got == 5


class TestResolveFailure:
    """全部路径未命中必须抛 MatchResolveError。"""

    def test_空候选行(self):
        with pytest.raises(MatchResolveError):
            resolve_match_id(
                match_no="周二008", match_date="2026-09-29",
                home_team="不存在队", away_team="无此队", rows=[],
            )

    def test_错误信息含场次编号便于排查(self):
        with pytest.raises(MatchResolveError) as exc_info:
            resolve_match_id(
                match_no="周二008", match_date="2026-09-29",
                home_team="不存在队", away_team="无此队", rows=[],
            )
        assert "周二008" in str(exc_info.value)
