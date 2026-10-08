# -*- coding: utf-8 -*-
"""tests 共享夹具：MySQL Fake 连接（单元测试禁止真实连库，testing.md 红线）。

FakeCursor/FakeConnection 提供 DictCursor 语义的最小实现，供
test_mysql_data_source / test_mysql_analysis_* 复用；结果集按 execute
顺序弹出。真实 MySQL 联调走 tests/integration（FA_MYSQL_INTEGRATION 门控）。
"""
import pytest

import fa_orchestrator.data.mysql_data_source as mysql_data_source_mod


class FakeCursor:
    """记录执行语句并按序返回预置结果集（DictCursor 语义）。"""

    def __init__(self, result_sets):
        self.result_sets = list(result_sets)
        self.executed = []

    def execute(self, sql, params=None):
        self.executed.append((sql, params))

    def __iter__(self):
        return iter(self.result_sets.pop(0) if self.result_sets else [])

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False


class FakeConnection:
    """假连接：仅支撑 with 语义与单游标访问。"""

    def __init__(self, result_sets):
        self.cursor_obj = FakeCursor(result_sets)
        self.closed = False

    def cursor(self):
        return self.cursor_obj

    def close(self):
        self.closed = True

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        self.closed = True
        return False


@pytest.fixture
def patch_mysql_connect(monkeypatch):
    """把 pymysql.connect 替换为返回 FakeConnection 的工厂。

    用法：conn = patch_mysql_connect([结果集1, 结果集2, ...])，
    conn.cursor_obj.executed 可断言执行语句与参数。
    数据源族通过 MySqlDataSource._connect 统一建连，patch 一处即全覆盖。
    """

    def _install(result_sets):
        conn = FakeConnection(result_sets)

        def fake_connect(**kwargs):
            return conn

        monkeypatch.setattr(mysql_data_source_mod.pymysql, "connect", fake_connect)
        return conn

    return _install
