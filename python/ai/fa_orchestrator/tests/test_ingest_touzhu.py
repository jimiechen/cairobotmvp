# -*- coding: utf-8 -*-
"""touzhu 文件入库 CLI 单元测试（立项四）。

验证 build_statements 纯函数的 SQL 组装（upsert/幂等）与 main 的
dry-run/写入双模式。写入模式使用 Fake 连接，禁止真实连库。
"""
import json

import pytest

from fa_orchestrator.ingest_touzhu import build_statements, main

PARSED = {
    "trading_rows": [
        {"trade_type": "standard", "total_volume": 728.0, "home_pct": 22.66,
         "draw_pct": 17.86, "away_pct": 59.48, "bookmaker_pnl": -123.45},
        {"trade_type": "large", "total_volume": 3000.0, "home_pct": None,
         "draw_pct": None, "away_pct": None, "bookmaker_pnl": None},
    ],
    "ladder_rows": [
        {"trade_type": "standard", "price": 2.0,
         "back_volume": 100.0, "lay_volume": 50.0},
    ],
    "missing": [],
}
SNAPSHOT = "2026-09-29 21:47:08"


class TestBuildStatements:
    def test_成交行走upsert且含pnl列(self):
        statements = build_statements(PARSED, match_id=2635,
                                      snapshot_time=SNAPSHOT)
        upsert_sql, upsert_params = statements[0]
        assert "INSERT INTO `betfair_trading`" in upsert_sql
        assert "bookmaker_pnl" in upsert_sql
        assert "ON DUPLICATE KEY UPDATE" in upsert_sql
        standard = upsert_params
        assert standard[0] == 2635          # match_id
        assert standard[1] == "standard"    # trade_type
        assert standard[6] == -123.45       # pnl 列
        assert standard[7] == SNAPSHOT

    def test_价位档先删后插保证幂等(self):
        statements = build_statements(PARSED, match_id=2635,
                                      snapshot_time=SNAPSHOT)
        sqls = [s for s, _ in statements]
        delete_idx = next(i for i, s in enumerate(sqls)
                          if "DELETE FROM `betfair_price_ladder`" in s)
        insert_idx = next(i for i, s in enumerate(sqls)
                          if "INSERT INTO `betfair_price_ladder`" in s)
        assert delete_idx < insert_idx
        delete_params = statements[delete_idx][1]
        assert delete_params == (2635, "standard", SNAPSHOT)

    def test_空解析结果仅返回空语句列表(self):
        parsed = {"trading_rows": [], "ladder_rows": [], "missing": []}
        assert build_statements(parsed, 2635, SNAPSHOT) == []


class FakeWriteConn:
    """最小写连接：记录语句，支撑 with/commit 语义。"""

    def __init__(self):
        self.executed = []
        self.committed = False
        self.closed = False

    def cursor(self):
        return self

    def execute(self, sql, params=None):
        self.executed.append((sql, params))

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False

    def commit(self):
        self.committed = True

    def close(self):
        self.closed = True


class TestMain:
    @pytest.fixture
    def touzhu_file(self, tmp_path):
        payload = {
            "site_match_id": 1398605, "date": "2026-09-29",
            "parsed_at": "2026-09-29T21:47:08",
            "data": {"touzhu": {
                "bf_trading": {"total_volume": 728.0,
                                "bookmaker_pnl": -123.45},
            }},
        }
        path = tmp_path / "1398605.json"
        path.write_text(json.dumps(payload), encoding="utf-8")
        return str(path)

    def test_默认dry_run只打印不连库(self, touzhu_file, capsys, monkeypatch):
        def fail_connect(**kwargs):
            raise AssertionError("dry-run 模式禁止建连")
        monkeypatch.setattr("fa_orchestrator.ingest_touzhu.pymysql.connect",
                            fail_connect)
        code = main(["--file", touzhu_file, "--match-id", "2635",
                     "--snapshot-time", SNAPSHOT])
        assert code == 0
        out = capsys.readouterr().out
        assert "betfair_trading" in out
        assert "dry-run" in out

    def test_write模式执行语句并提交(self, touzhu_file, monkeypatch):
        # 写入前有 FA_MYSQL_* 配置校验（缺失报 MySqlConfigError），测试注入假配置
        for name in ("FA_MYSQL_HOST", "FA_MYSQL_USER",
                     "FA_MYSQL_PASSWORD", "FA_MYSQL_DB"):
            monkeypatch.setenv(name, "test-" + name)
        conn = FakeWriteConn()
        monkeypatch.setattr("fa_orchestrator.ingest_touzhu.pymysql.connect",
                            lambda **kwargs: conn)
        code = main(["--file", touzhu_file, "--match-id", "2635",
                     "--snapshot-time", SNAPSHOT, "--write"])
        assert code == 0
        assert conn.committed and conn.closed
        assert any("betfair_trading" in sql for sql, _ in conn.executed)

    def test_缺match_id时报错退出(self, touzhu_file):
        with pytest.raises(SystemExit):
            main(["--file", touzhu_file, "--match-id", ""])
