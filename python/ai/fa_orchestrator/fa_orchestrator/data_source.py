# -*- coding: utf-8 -*-
"""取数数据源（PRD-12 F-06/F-09）：MySQL / fixture 双通道 + 数据溯源快照。

职责边界：
- 负责：Q1/Q2b/候选行查询的数据源抽象与 fixture 实现、溯源快照结构。
- 不负责：SQL 契约文本（sql_contracts.py）、完整性判定（integrity.py）。

MySQL 通道（MySqlDataSource）依赖 pymysql（仓库外联调环境，R-04 未就绪），
ISS-FA1 仅保留接口骨架；单元测试一律使用 FixtureDataSource，禁止真实连库。
"""
import json
from datetime import datetime, timezone
from pathlib import Path

from fa_orchestrator.sql_contracts import SQL_JC_MATCH_CANDIDATES, SQL_Q1_MATCH_LIST, build_q2b_sql

# 数据通道标识（F-09 溯源快照必填）
CHANNEL_FIXTURE = "fixture"
CHANNEL_MYSQL = "mysql"

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

    状态约束：只读；fixture 结构对齐需求文档 §2.3/§四。
    """

    def __init__(self, fixture_path):
        self.fixture_path = Path(fixture_path)
        self.fixture = json.loads(self.fixture_path.read_text(encoding="utf-8"))
        self.match_date = self.fixture.get("match_date")

    def query_match_list(self, match_date: str) -> list:
        """Q1（D1 比赛清单）。

        fixture 本身即样本日数据集；真实窗口口径（12:00 → 次日 12:00，
        跨日场次归属开球日）由调用方保证（对齐 Q1 SQL 契约注释）。
        """
        if match_date != self.match_date:
            return []
        return list(self.fixture.get("q1_matches", []))

    def query_q2b_counts(self, fids: list) -> dict:
        """Q2b（六模块行数）：按 fid 过滤返回 {fid: {模块: 行数}}。"""
        all_counts = self.fixture.get("q2b_counts", {})
        return {fid: all_counts[fid] for fid in fids if fid in all_counts}

    def query_jc_candidates(self) -> list:
        """`match` 表候选行（四路径解析输入）。"""
        return list(self.fixture.get("jc_match_rows", []))

    def get_sql_contracts(self) -> dict:
        """暴露本数据源对应的 SQL 契约（供联调核对与审计）。"""
        return {
            "Q1": SQL_Q1_MATCH_LIST.strip(),
            "Q2b_ouzhi": build_q2b_sql("odds_europe", 1),
            "jc_candidates": SQL_JC_MATCH_CANDIDATES.strip(),
        }

    @staticmethod
    def _in_day_window(row: dict, match_date: str) -> bool:
        """竞彩日窗口判定：跨日场次按开球日归属（fixture 简化：同月即命中）。"""
        return str(row.get("match_date", "")).startswith(match_date[:7])


class MySqlDataSource:
    """MySQL 直连数据源（骨架）。

    依赖 pymysql 与联调环境（R-04：192.168.1.6:3306/workbench），
    ISS-FA1 不实现连接逻辑；实现时必须：全参数化 SQL、`match` 反引号、
    连接串走环境变量、快照溯源落盘。
    """

    def __init__(self, *args, **kwargs):
        raise NotImplementedError(
            "MySQL 通道待联调环境就绪（R-04）后实现，当前请使用 FixtureDataSource"
        )
