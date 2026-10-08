# -*- coding: utf-8 -*-
"""取数数据源（PRD-12 F-06/F-09/F-10/F-11/F-16）：fixture 通道 + 数据溯源快照。

职责边界：
- 负责：通道标识、溯源快照结构、fixture 数据源实现（Q1/Q2b/候选行/
  两阶段分析数据块 Q3/Q5/Q9~Q12/历史推演）。
- 不负责：SQL 契约文本（sql_contracts.py）、MySQL 真实通道
  （mysql_data_source.py，R-04 联调）、完整性判定（integrity.py）、
  分析计算（fundamental_agent.py / odds_verify_agent.py）。

测试约束：单元测试一律使用 FixtureDataSource 或 Fake 连接，禁止真实连库；
真实联调走 tests/integration/test_mysql_integration.py（FA_MYSQL_INTEGRATION 门控）。
analysis fixture 为可选配置：未配置时分析类查询返回 None/空 dict（向后兼容）。
"""
import json
from datetime import date, datetime, timedelta, timezone
from pathlib import Path

from fa_orchestrator.data.sql_contracts import SQL_JC_MATCH_CANDIDATES, SQL_Q1_MATCH_LIST, build_q2b_sql

# 数据通道标识（F-09 溯源快照必填）
CHANNEL_FIXTURE = "fixture"
CHANNEL_MYSQL = "mysql"

# 竞彩日窗口起点（PRD-12 §12.3：[D 12:00, D+1 12:00) 左闭右开）；
# match_time 为 'HH:MM' 零填充字符串，字典序比较与时间序一致
DAY_WINDOW_START_TIME = "12:00"

# 溯源快照涉及的数据块编号（需求文档 §三）
BLOCK_D1_MATCH_LIST = "D1"
BLOCK_D2B_INTEGRITY = "D2b"


def make_fetch_trace(channel: str, blocks: list, extra: dict = None) -> dict:
    """构建数据溯源快照（F-09）：数据块编号 + 通道 + 快照时间 + 行数等。"""
    trace = {
        "blocks": blocks,
        "channel": channel,
        "snapshot_time": datetime.now(timezone.utc).isoformat(),
    }
    trace.update(extra or {})
    return trace


class FixtureDataSource:
    """fixture 数据源：从样本日 JSON 读取，供单元测试与 CLI 演示。

    状态约束：只读；fixture 结构对齐需求文档 §2.3/§四；
    analysis fixture（可选）结构对齐 ISS-FA2 分析数据块说明。
    """

    def __init__(self, fixture_path, analysis_path=None):
        self.fixture_path = Path(fixture_path)
        self.fixture = json.loads(self.fixture_path.read_text(encoding="utf-8"))
        self.match_date = self.fixture.get("match_date")
        self.analysis = {}
        if analysis_path is not None:
            self.analysis = json.loads(Path(analysis_path).read_text(encoding="utf-8"))

    def query_match_list(self, match_date: str) -> list:
        """Q1（D1 比赛清单）：按竞彩日跨日窗口过滤，与 Q1 SQL 契约口径一致。

        窗口口径（PRD-12 §12.3）：[D 12:00, D+1 12:00) 左闭右开——当日 12:00
        起（含）或次日 12:00 前（不含）开球的场次归属竞彩日 match_date；
        fixture 行 match_time 为 'HH:MM' 零填充开球时间（对齐 matches 表口径）。
        """
        if match_date != self.match_date:
            return []
        return [
            row for row in self.fixture.get("q1_matches", [])
            if self._in_day_window(row, match_date)
        ]

    def query_q2b_counts(self, fids: list, match_ids: dict = None) -> dict:
        """Q2b（六模块行数）：按 fid 过滤返回 {fid: {模块: 行数}}。

        match_ids（fid → match.id，DataAgent 四路径解析结果）供 MySQL 通道
        定位模块表数据块；fixture 数据块本身按 fid 组织，忽略该参数。
        """
        all_counts = self.fixture.get("q2b_counts", {})
        return {fid: all_counts[fid] for fid in fids if fid in all_counts}

    def query_jc_candidates(self) -> list:
        """`match` 表候选行（四路径解析输入）。"""
        return list(self.fixture.get("jc_match_rows", []))

    def query_match_odds(self, fid: str):
        """Q3（单场三盘赔率明细）：欧赔/亚盘/大小球，is_final_snapshot 标注临场快照。"""
        return self.analysis.get("q3_odds", {}).get(fid)

    def query_vote(self, fid: str):
        """Q5（投票支持率）：缺失场次返回 None（对齐 q2b_counts.vote=0 口径）。"""
        return self.analysis.get("q5_vote", {}).get(fid)

    def query_fundamentals(self, fid: str) -> dict:
        """Q9~Q12（基本面四块）：情报/历史交锋/近期状态/赛程；未配置 analysis 返回空。"""
        if not self.analysis:
            return {}
        return {
            "intelligence": self.analysis.get("q9_intelligence", {}).get(fid) or {},
            "h2h": self.analysis.get("q10_h2h", {}).get(fid) or {},
            "recent_form": self.analysis.get("q11_recent", {}).get(fid) or {},
            "schedule": self.analysis.get("q12_schedule", {}).get(fid) or {},
        }

    def query_history_tracks(self, fid: str):
        """M5 历史推演（升降盘时序轨迹）：缺失返回 None → 调用方标注无法推演时序。"""
        return self.analysis.get("history_tracks", {}).get(fid)

    def query_heat_drift(self, fid: str):
        """M5 历史推演（投票热度漂移百分点）：缺失返回 None。"""
        return self.analysis.get("heat_drift", {}).get(fid)

    def get_sql_contracts(self) -> dict:
        """暴露本数据源对应的 SQL 契约（供联调核对与审计）。"""
        return {
            "Q1": SQL_Q1_MATCH_LIST.strip(),
            "Q2b_ouzhi": build_q2b_sql("odds_europe", 1),
            "jc_candidates": SQL_JC_MATCH_CANDIDATES.strip(),
        }

    @staticmethod
    def _in_day_window(row: dict, match_date: str) -> bool:
        """竞彩日窗口判定（PRD-12 §12.3）：[D 12:00, D+1 12:00) 左闭右开。

        输入：row（match_date=开球日、match_time='HH:MM' 零填充）、
        match_date=竞彩日。
        输出：该场是否落入竞彩日 D 的开球窗口。
        失败行为：开球日或开球时间缺失判不在窗口——不臆测、不静默放宽口径。
        """
        kickoff_day = str(row.get("match_date") or "")
        kickoff_time = str(row.get("match_time") or "")
        if not kickoff_day or not kickoff_time:
            return False
        if kickoff_day == match_date:
            return kickoff_time >= DAY_WINDOW_START_TIME
        next_day = (date.fromisoformat(match_date) + timedelta(days=1)).isoformat()
        return kickoff_day == next_day and kickoff_time < DAY_WINDOW_START_TIME
