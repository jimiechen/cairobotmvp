# -*- coding: utf-8 -*-
"""MySQL 直连数据源（PRD-12 F-06 联调通道，R-04）。

职责边界：
- 负责：workbench 库（192.168.1.6:3306）数据准备通道——Q1 清单、候选行、
  Q2b 六模块行数核查、连接配置管理（环境变量 + 构造参数，凭据不入库）。
- 不负责：fixture 通道（data_source.py）、分析类查询
  （mysql_analysis_source.py，方案 A 批次 1 起归档实现）、SQL 契约文本
  （sql_contracts.py）、解析与完整性判定（id_resolver.py / integrity.py）。

联调确认的环境事实（2026-09-30 / 2026-10-08 只读探测，详见测试报告）：
- workbench 库九张契约表齐全；Q1 契约 SQL 原样可运行
- 六模块表 match_id 域 = `match`.id（非 500com fid），Q2b 须传入解析后 match_ids
- 契约 SQL_JC_MATCH_CANDIDATES 列名与实际 schema 不符（cc→cc_id 等），本层适配
- 分析类契约（Q3/Q5/Q9~Q12/D9/D10/D12）已按真实库结构归档至 sql_contracts.py，
  由 MySqlAnalysisDataSource 执行（本模块不再保留桩方法）

单元测试一律使用 Fake 连接，禁止真实连库；
真实联调走 tests/integration/test_mysql_integration.py（FA_MYSQL_INTEGRATION 门控）。
"""
import os

import pymysql
import pymysql.cursors

from fa_orchestrator.data.sql_contracts import (
    INTEGRITY_MODULES,
    SQL_JC_MATCH_CANDIDATES,
    SQL_Q1_MATCH_LIST,
    build_q2b_sql,
)

# MySQL 连接配置环境变量（凭据只走环境变量/构造参数，禁止入库入日志）
ENV_MYSQL_HOST = "FA_MYSQL_HOST"
ENV_MYSQL_PORT = "FA_MYSQL_PORT"
ENV_MYSQL_USER = "FA_MYSQL_USER"
ENV_MYSQL_PASSWORD = "FA_MYSQL_PASSWORD"
ENV_MYSQL_DB = "FA_MYSQL_DB"

# 连接默认值：端口走 MySQL 标准端口；超时从短，避免编排层长时间挂起
DEFAULT_MYSQL_PORT = 3306
DEFAULT_CONNECT_TIMEOUT_S = 5

# 联调适配（R-04 探测发现）：需求文档契约 SQL_JC_MATCH_CANDIDATES 的列名
# （cc/home_team/away_team）与 workbench 实际列（cc_id/home/away）不一致。
# 契约文本保留原样供审计对照；本常量为实际执行版本，列名归一由
# MySqlDataSource._adapt_candidate_row 完成。ORDER BY id：`match` 表存在
# 同场次重复行且数据集中于小 id（联调探测 8 组），保证四路径解析确定性。
SQL_JC_CANDIDATES_MYSQL = (
    "SELECT id, cc_id, match_date, play_type, home, away "
    "FROM `match` WHERE play_type = 'jingcai' ORDER BY id"
)


def _iso_day(value) -> str:
    """日期值归一为 ISO 日期字符串：兼容 date/datetime/str（DB 返回 date 对象）。"""
    return str(value)[:10]


class MySqlConfigError(ValueError):
    """MySQL 连接配置缺失或不完整时抛出（连接串不入库、不落日志）。"""


class MySqlDataSource:
    """MySQL 直连数据源（R-04 联调通道，只读）。

    状态约束：只读查询，禁止任何写入；连接按查询懒建立、用毕即关；
    凭据仅经构造参数或 FA_MYSQL_* 环境变量传入，repr 一律脱敏。

    实现约束（需求文档硬性规则）：
    - Q1 使用 sql_contracts 契约 SQL 原样执行，全参数化
    - 候选查询使用 schema 适配版 SQL（契约列名与实际库不一致，见常量注释）
    - Q2b 六模块表名必须来自 INTEGRITY_MODULES 白名单（build_q2b_sql 保证）
    """

    def __init__(self, host=None, port=None, user=None, password=None, database=None):
        """构建数据源；配置优先级：显式参数 > 环境变量。

        失败行为：host/user/password/database 任一缺失抛 MySqlConfigError。
        懒连接：构造期不建立连接，首次查询时才连接。
        """
        env = os.environ
        self.host = host or env.get(ENV_MYSQL_HOST)
        self.port = int(port or env.get(ENV_MYSQL_PORT) or DEFAULT_MYSQL_PORT)
        self.user = user or env.get(ENV_MYSQL_USER)
        self.password = password or env.get(ENV_MYSQL_PASSWORD)
        self.database = database or env.get(ENV_MYSQL_DB)

        missing = [
            name for name, value in (
                (ENV_MYSQL_HOST, self.host),
                (ENV_MYSQL_USER, self.user),
                (ENV_MYSQL_PASSWORD, self.password),
                (ENV_MYSQL_DB, self.database),
            ) if not value
        ]
        if missing:
            raise MySqlConfigError(
                f"MySQL 连接配置缺失: {', '.join(missing)}"
                "（请传参或设置对应环境变量）"
            )

    def __repr__(self):
        # 密码脱敏：repr 可能进入异常链/调试输出，禁止泄露凭据
        return (
            f"MySqlDataSource(host={self.host!r}, port={self.port}, "
            f"user={self.user!r}, database={self.database!r}, password=***)"
        )

    def _connect(self):
        """建立只读用途连接（DictCursor 便于按列名取值）。"""
        return pymysql.connect(
            host=self.host,
            port=self.port,
            user=self.user,
            password=self.password,
            database=self.database,
            charset="utf8mb4",
            connect_timeout=DEFAULT_CONNECT_TIMEOUT_S,
            cursorclass=pymysql.cursors.DictCursor,
        )

    def query_match_list(self, match_date: str) -> list:
        """Q1（D1 比赛清单）：契约 SQL 原样执行，仅参数化日期。

        输出：{internal_id, match_no, home_team, away_team, match_date, fid}
        （match_date/fid 归一为字符串，与 fixture 行结构对齐）。
        失败行为：连接/执行异常向上抛出，不吞掉。
        """
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(SQL_Q1_MATCH_LIST.strip(), (match_date,))
            rows = list(cur)
        return [
            {
                "internal_id": row["internal_id"],
                "match_no": str(row["match_no"]),
                "home_team": row["home_team"],
                "away_team": row["away_team"],
                "match_date": _iso_day(row["match_date"]),
                "fid": str(row["fid"]),
            }
            for row in rows
        ]

    def query_jc_candidates(self) -> list:
        """`match` 表候选行（四路径解析输入），列名归一为解析器口径。

        输出行结构：{id, cc, match_date, play_type, home_team, away_team}
        （alias 字段缺省：实际表无别名列，路径4 自然不触发）。
        """
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(SQL_JC_CANDIDATES_MYSQL)
            rows = list(cur)
        return [self._adapt_candidate_row(row) for row in rows]

    @staticmethod
    def _adapt_candidate_row(row: dict) -> dict:
        """workbench 实际列名 → id_resolver 契约口径（cc_id→cc、home→home_team 等）。"""
        return {
            "id": int(row["id"]),
            "cc": row["cc_id"],
            "match_date": _iso_day(row["match_date"]),
            "play_type": row["play_type"],
            "home_team": row["home"],
            "away_team": row["away"],
        }

    def query_q2b_counts(self, fids: list, match_ids: dict = None) -> dict:
        """Q2b（六模块行数）：按解析后的 match.id 批量核查。

        输入：fids（500com 场次号清单）、match_ids（fid → match.id，
        DataAgent 四路径解析结果——模块表 ID 域即 `match`.id）。
        输出：{fid: {六模块: 行数}}；无任何可用映射时返回 {}（不臆测）。
        失败行为：连接/执行异常向上抛出。
        """
        id_by_fid = {
            fid: mid for fid, mid in (match_ids or {}).items()
            if mid is not None and fid in set(fids)
        }
        if not id_by_fid:
            return {}

        ids = tuple(id_by_fid.values())
        with self._connect() as conn, conn.cursor() as cur:
            counts_by_module = {}
            for module, table in INTEGRITY_MODULES.items():
                cur.execute(build_q2b_sql(table, len(ids)), ids)
                counts_by_module[module] = {row["match_id"]: row["c"] for row in cur}

        return {
            fid: {
                module: counts_by_module[module].get(mid, 0)
                for module in INTEGRITY_MODULES
            }
            for fid, mid in id_by_fid.items()
        }

    def get_sql_contracts(self) -> dict:
        """暴露数据源对应的 SQL 契约（供联调核对与审计）。

        jc_candidates 保留需求文档契约原文供差异对照，
        jc_candidates_mysql_adapted 为实际执行版本。
        """
        return {
            "Q1": SQL_Q1_MATCH_LIST.strip(),
            "Q2b_ouzhi": build_q2b_sql("odds_europe", 1),
            "jc_candidates": SQL_JC_MATCH_CANDIDATES.strip(),
            "jc_candidates_mysql_adapted": SQL_JC_CANDIDATES_MYSQL,
        }
