# -*- coding: utf-8 -*-
"""MySqlDataSource 真实 MySQL 联调（R-04，环境门控执行）。

运行方式：FA_MYSQL_INTEGRATION=1 make integration
（需同时通过 FA_MYSQL_HOST/PORT/USER/PASSWORD/DB 提供连接配置，凭据不入库）
默认（未设 FA_MYSQL_INTEGRATION）整文件 skip 并输出原因：
单元测试禁止真实连库（testing.md §2/§6），联调按门控放行。

样本锚点（2026-09-30 探测）：workbench 库，样本日 2026-09-29 共 13 场，
周一003（fid=1398605）四路径解析落点 match.id=2635，六模块全有数据。
若联调库数据发生清理，需重新探测并同步本文件锚点。
"""
import os

import pytest

from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.data_source import CHANNEL_MYSQL
from fa_orchestrator.id_resolver import resolve_match_id
from fa_orchestrator.mysql_data_source import MySqlDataSource

_GATE_ENV = "FA_MYSQL_INTEGRATION"

pytestmark = pytest.mark.skipif(
    os.environ.get(_GATE_ENV) != "1",
    reason=(
        "真实 MySQL 联调需 FA_MYSQL_INTEGRATION=1 与 FA_MYSQL_* 连接配置"
        "（单元测试禁止真实连库，联调按门控放行）"
    ),
)

SAMPLE_DATE = "2026-09-29"
SAMPLE_FID = "1398605"        # 周一003 亚美尼亚 vs 黑山
SAMPLE_MATCH_NO = "周一003"


def _build_source() -> MySqlDataSource:
    """从 FA_MYSQL_* 环境变量构建真实数据源。"""
    return MySqlDataSource()


class TestMySqlIntegration:
    """真实通道联调：只读查询，禁止任何写入。"""

    def test_Q1返回样本日清单(self):
        matches = _build_source().query_match_list(SAMPLE_DATE)
        assert len(matches) >= 7
        assert SAMPLE_FID in {m["fid"] for m in matches}

    def test_四路径解析真实落点(self):
        source = _build_source()
        row = next(
            m for m in source.query_match_list(SAMPLE_DATE) if m["fid"] == SAMPLE_FID
        )
        match_id = resolve_match_id(
            match_no=row["match_no"],
            match_date=SAMPLE_DATE,
            home_team=row["home_team"],
            away_team=row["away_team"],
            rows=source.query_jc_candidates(),
        )
        assert isinstance(match_id, int)
        assert match_id > 0

    def test_Q2b六模块真实行数(self):
        source = _build_source()
        row = next(
            m for m in source.query_match_list(SAMPLE_DATE) if m["fid"] == SAMPLE_FID
        )
        match_id = resolve_match_id(
            match_no=row["match_no"],
            match_date=SAMPLE_DATE,
            home_team=row["home_team"],
            away_team=row["away_team"],
            rows=source.query_jc_candidates(),
        )
        counts = source.query_q2b_counts([SAMPLE_FID], match_ids={SAMPLE_FID: match_id})
        assert set(counts[SAMPLE_FID]) == {"ouzhi", "yazhi", "daxiao", "vote", "stats", "intel"}
        assert all(v > 0 for v in counts[SAMPLE_FID].values())

    def test_DataAgent真实通道端到端(self):
        report = DataAgent(_build_source(), channel=CHANNEL_MYSQL).prepare(SAMPLE_DATE)
        assert report["total_matches"] >= 7
        assert report["ready_matches"] >= 1
        assert report["matches"][0]["fetch_trace"]["channel"] == CHANNEL_MYSQL
