# -*- coding: utf-8 -*-
"""touzhu 文件入库 CLI（立项四）。

文件职责：读取 parse/results/{日期}/{match_id}.json 的 data.touzhu 块，
解析为 betfair_trading / betfair_price_ladder 行并写入 MySQL。

不负责：touzhu 块字段解析规则（data/touzhu_file_parse.py）、分析层消费。

用法：
    # dry-run（默认）：仅打印解析结果与将执行语句，不建连不写库
    python -m fa_orchestrator.ingest_touzhu --file <json路径> --match-id <match.id> \
        [--snapshot-time "YYYY-MM-DD HH:MM:SS"]
    # 写入模式：需显式 --write
    python -m fa_orchestrator.ingest_touzhu --file <json路径> --match-id 2635 --write

快照时间优先级：--snapshot-time > 文件顶层 parsed_at > 执行时刻。
幂等口径：betfair_trading 走 uk_bf 唯一键 upsert；
betfair_price_ladder 同键先 DELETE 后 INSERT。
连接配置复用 FA_MYSQL_* 环境变量（凭据不入库不落日志）。
"""
import argparse
import json
import logging
import os
import sys
from datetime import datetime

import pymysql

from fa_orchestrator.data.mysql_data_source import (
    DEFAULT_CONNECT_TIMEOUT_S,
    DEFAULT_MYSQL_PORT,
    ENV_MYSQL_DB,
    ENV_MYSQL_HOST,
    ENV_MYSQL_PASSWORD,
    ENV_MYSQL_PORT,
    ENV_MYSQL_USER,
    MySqlConfigError,
)
from fa_orchestrator.data.touzhu_file_parse import touzhu_block

logger = logging.getLogger(__name__)

# betfair_trading upsert：uk_bf (match_id, trade_type, snapshot_time) 冲突时
# 更新量/占比/pnl 列（VALUES() 函数对齐 MySQL 5.7，真库为 5.7 系）
SQL_TRADING_UPSERT = """
INSERT INTO `betfair_trading`
  (match_id, trade_type, total_volume, home_pct, draw_pct, away_pct,
   bookmaker_pnl, snapshot_time)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
ON DUPLICATE KEY UPDATE
  total_volume = VALUES(total_volume),
  home_pct = VALUES(home_pct),
  draw_pct = VALUES(draw_pct),
  away_pct = VALUES(away_pct),
  bookmaker_pnl = VALUES(bookmaker_pnl)
""".strip()

# 价位档幂等：同 (match_id, trade_type, snapshot_time) 先删后插
SQL_LADDER_DELETE = """
DELETE FROM `betfair_price_ladder`
WHERE match_id = %s AND trade_type = %s AND snapshot_time = %s
""".strip()

SQL_LADDER_INSERT = """
INSERT INTO `betfair_price_ladder`
  (match_id, trade_type, price, back_volume, lay_volume, snapshot_time)
VALUES (%s, %s, %s, %s, %s, %s)
""".strip()


def build_statements(parsed: dict, match_id: int, snapshot_time: str) -> list:
    """解析结果 → [(sql, params), ...] 执行语句序列（纯函数，供审计/dry-run）。

    输入：parsed（touzhu_block 输出）、match_id（`match`.id 域）、
    snapshot_time（快照时间字符串）。
    输出：成交行 upsert + 价位档先删后插的语句序列；无数据行返回 []。
    """
    statements = []
    for row in parsed.get("trading_rows") or []:
        statements.append((SQL_TRADING_UPSERT, (
            match_id, row["trade_type"], row["total_volume"],
            row["home_pct"], row["draw_pct"], row["away_pct"],
            row["bookmaker_pnl"], snapshot_time,
        )))
    for row in parsed.get("ladder_rows") or []:
        trade_type = row["trade_type"]
        if not any("DELETE FROM `betfair_price_ladder`" in sql
                   and params[0] == match_id and params[1] == trade_type
                   for sql, params in statements):
            statements.append((SQL_LADDER_DELETE,
                               (match_id, trade_type, snapshot_time)))
        statements.append((SQL_LADDER_INSERT, (
            match_id, trade_type, row["price"],
            row["back_volume"], row["lay_volume"], snapshot_time,
        )))
    return statements


def _resolve_snapshot_time(args, payload: dict) -> str:
    """快照时间归一：--snapshot-time > parsed_at > 执行时刻。"""
    if args.snapshot_time:
        return args.snapshot_time
    parsed_at = payload.get("parsed_at")
    if parsed_at:
        try:
            return datetime.fromisoformat(str(parsed_at)).strftime(
                "%Y-%m-%d %H:%M:%S")
        except ValueError:
            logger.warning("parsed_at 无法解析: %r，回退执行时刻", parsed_at)
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def main(argv=None) -> int:
    """CLI 入口：默认 dry-run，--write 才连库执行。

    失败行为：文件不存在/JSON 非法/配置缺失向上抛出（argparse/connect 异常），
    不吞异常；dry-run 禁止建连（测试断言覆盖）。
    """
    parser = argparse.ArgumentParser(
        prog="python -m fa_orchestrator.ingest_touzhu",
        description="touzhu 文件（庄家 pnl/价位聚集）入 MySQL（立项四）")
    parser.add_argument("--file", required=True, help="touzhu JSON 文件路径")
    parser.add_argument("--match-id", required=True, type=int,
                        help="`match`.id（betfair_trading.match_id 域）")
    parser.add_argument("--snapshot-time",
                        help="快照时间 YYYY-MM-DD HH:MM:SS（默认取 parsed_at）")
    parser.add_argument("--write", action="store_true",
                        help="实际写库（缺省 dry-run 只打印）")
    args = parser.parse_args(argv)

    with open(args.file, encoding="utf-8") as fh:
        payload = json.load(fh)
    snapshot_time = _resolve_snapshot_time(args, payload)
    parsed = touzhu_block((payload.get("data") or {}).get("touzhu"))
    statements = build_statements(parsed, args.match_id, snapshot_time)


    if parsed.get("missing"):
        print(f"缺失标注（不臆造，列取 None）: {parsed['missing']}")
    if not args.write:
        print(f"dry-run：共 {len(statements)} 条语句，未连库")
        for sql, params in statements:
            print(f"SQL: {sql}\n参数: {params}")
        return 0

    # 写库为唯一的仓库内 FA 入库通道，与只读数据源（MySqlDataSource）区分；
    # 配置复用同一组 FA_MYSQL_* 环境变量，缺失时报 MySqlConfigError
    env = os.environ
    missing = [name for name in (ENV_MYSQL_HOST, ENV_MYSQL_USER,
                                 ENV_MYSQL_PASSWORD, ENV_MYSQL_DB)
               if not env.get(name)]
    if missing:
        raise MySqlConfigError(
            f"MySQL 连接配置缺失: {', '.join(missing)}（请设置对应环境变量）")
    conn = pymysql.connect(
        host=env[ENV_MYSQL_HOST], port=int(env.get(ENV_MYSQL_PORT) or DEFAULT_MYSQL_PORT),
        user=env[ENV_MYSQL_USER], password=env[ENV_MYSQL_PASSWORD],
        database=env[ENV_MYSQL_DB], charset="utf8mb4",
        connect_timeout=DEFAULT_CONNECT_TIMEOUT_S)
    try:
        with conn.cursor() as cur:
            for sql, params in statements:
                cur.execute(sql, params)
        conn.commit()
        print(f"写入完成：{len(statements)} 条语句已提交")
    finally:
        conn.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
