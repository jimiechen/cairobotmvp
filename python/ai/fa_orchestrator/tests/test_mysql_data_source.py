# -*- coding: utf-8 -*-
"""MySqlDataSource 单元测试（PRD-12 F-06 联调，R-04 退役）。

全部使用 Fake 连接（conftest.patch_mysql_connect），禁止真实连库；
真实 MySQL 联调见 tests/integration/test_mysql_integration.py（FA_MYSQL_INTEGRATION 门控）。

联调已确认的环境事实（2026-09-30 只读探测，详见测试报告）：
- workbench 库九张契约表齐全；Q1 契约 SQL 原样可运行
- 六模块表 match_id 域 = `match`.id（非 500com fid），Q2b 需传入解析后的 match_id
- 契约 SQL_JC_MATCH_CANDIDATES 列名与实际 schema 不符（cc→cc_id 等），数据源内做适配
- 分析类查询已迁至 MySqlAnalysisDataSource（mysql_analysis_source.py），本文件不再覆盖
"""
from datetime import date

import pytest

import fa_orchestrator.data.mysql_data_source as mysql_data_source_mod
from fa_orchestrator.agents.data_agent import DataAgent
from fa_orchestrator.data.data_source import CHANNEL_MYSQL
from fa_orchestrator.data.mysql_data_source import (
    SQL_JC_CANDIDATES_MYSQL,
    MySqlConfigError,
    MySqlDataSource,
)
from fa_orchestrator.data.sql_contracts import SQL_JC_MATCH_CANDIDATES, SQL_Q1_MATCH_LIST, build_q2b_sql


def _make_source(**overrides):
    params = {
        "host": "db-host", "port": 3306, "user": "tester",
        "password": "secret", "database": "workbench",
    }
    params.update(overrides)
    return MySqlDataSource(**params)


Q1_ROWS = [
    {"internal_id": "match_c8e30a24", "match_no": "周一003", "home_team": "亚美尼亚",
     "away_team": "黑山", "match_date": "2026-09-29", "fid": "1398605"},
]
CAND_ROWS = [
    {"id": 2635, "cc_id": "周一003", "match_date": "2026-09-29",
     "play_type": "jingcai", "home": "亚美尼亚", "away": "黑山"},
]


class TestMySqlConfig:
    """连接配置：构造参数 > 环境变量；缺失必填项抛 MySqlConfigError。"""

    def test_缺失database配置时抛错(self):
        with pytest.raises(MySqlConfigError):
            MySqlDataSource(host="db-host", user="tester", password="secret")

    def test_环境变量兜底完整配置(self, monkeypatch):
        for key, val in (("FA_MYSQL_HOST", "env-host"), ("FA_MYSQL_USER", "env-u"),
                         ("FA_MYSQL_PASSWORD", "env-p"), ("FA_MYSQL_DB", "workbench")):
            monkeypatch.setenv(key, val)
        source = MySqlDataSource()
        assert source.host == "env-host" and source.database == "workbench"

    def test_构造不触发连接(self):
        # 懒连接：构造期绝不建立真实连接（host 不可达也不报错）
        _make_source(host="unreachable-host")

    def test_repr不泄露密码(self):
        assert "secret" not in repr(_make_source())


class TestQ1MatchList:
    def test_执行契约SQL原样并返回清单(self, patch_mysql_connect):
        conn = patch_mysql_connect([Q1_ROWS])
        rows = _make_source().query_match_list("2026-09-29")
        assert rows[0]["fid"] == "1398605"
        assert rows[0]["match_no"] == "周一003"
        sql, params = conn.cursor_obj.executed[0]
        assert sql == SQL_Q1_MATCH_LIST.strip()
        # 窗口口径（[D 12:00, D+1 12:00) 左闭右开）内置于契约 SQL：
        # 当日分支与次日分支两个 %s 传同一竞彩日
        assert params == ("2026-09-29", "2026-09-29")

    def test_日期与fid归一为字符串(self, patch_mysql_connect):
        patch_mysql_connect([[dict(Q1_ROWS[0], match_date=date(2026, 9, 29), fid=1398605)]])
        rows = _make_source().query_match_list("2026-09-29")
        assert rows[0]["match_date"] == "2026-09-29"
        assert rows[0]["fid"] == "1398605"


class TestJcCandidates:
    def test_候选行列名适配为解析器口径(self, patch_mysql_connect):
        patch_mysql_connect([CAND_ROWS])
        rows = _make_source().query_jc_candidates("2026-09-29")
        assert rows[0] == {
            "id": 2635, "cc": "周一003", "match_date": "2026-09-29",
            "play_type": "jingcai", "home_team": "亚美尼亚", "away_team": "黑山",
        }
        assert "cc_id" not in rows[0] and "home" not in rows[0]

    def test_候选SQL窗口化且放开玩法过滤(self, patch_mysql_connect):
        # 宿主速览口径（2026-10-09）：候选按 kickoff_time ∈ [D 12:00, D+1 12:00)
        # 直查，beidan 与 jingcai 同入候选（编号体系不同靠队名路径兜底）
        conn = patch_mysql_connect([CAND_ROWS])
        _make_source().query_jc_candidates("2026-09-29")
        sql, params = conn.cursor_obj.executed[0]
        assert sql == SQL_JC_CANDIDATES_MYSQL
        assert "kickoff_time >= %s" in sql and "kickoff_time < %s" in sql
        assert "WHERE play_type" not in sql    # 玩法过滤由解析路径承担，SQL 不再限制
        assert "ORDER BY id" in sql            # 重复场次下解析确定性（取小 id）
        assert "`match`" in sql                # 保留字反引号
        assert params == ("2026-09-29 12:00", "2026-09-30 12:00")

    def test_契约文本保留原样供审计对照(self):
        # 需求文档契约不改写：差异由适配层消化，文本可对照审计
        assert "cc = %s" in SQL_JC_MATCH_CANDIDATES


class TestQ2bCounts:
    def test_按解析后match_id查询六模块(self, patch_mysql_connect):
        module_counts = [
            [{"match_id": 2635, "c": 231}], [{"match_id": 2635, "c": 272}],
            [{"match_id": 2635, "c": 32}], [{"match_id": 2635, "c": 20}],
            [{"match_id": 2635, "c": 6}], [{"match_id": 2635, "c": 4}],
        ]
        conn = patch_mysql_connect(module_counts)
        counts = _make_source().query_q2b_counts(["1398605"], match_ids={"1398605": 2635})
        assert counts["1398605"] == {
            "ouzhi": 231, "yazhi": 272, "daxiao": 32,
            "vote": 20, "stats": 6, "intel": 4,
        }
        # 六模块各执行一次白名单参数化查询，占位符数量 = id 数
        executed = conn.cursor_obj.executed
        assert len(executed) == 6
        assert executed[0] == (build_q2b_sql("odds_europe", 1), (2635,))

    def test_无match_id映射时返回空不臆测(self, patch_mysql_connect):
        patch_mysql_connect([[]])
        assert _make_source().query_q2b_counts(["1398605"]) == {}

    def test_部分fid缺映射时跳过该fid(self, patch_mysql_connect):
        patch_mysql_connect([[{"match_id": 2635, "c": 231}]] + [[]] * 5)
        counts = _make_source().query_q2b_counts(
            ["1398605", "1398610"], match_ids={"1398605": 2635}
        )
        assert "1398605" in counts and "1398610" not in counts

    def test_模块无行计为0(self, patch_mysql_connect):
        patch_mysql_connect([[{"match_id": 2635, "c": 231}], [], [], [], [], []])
        counts = _make_source().query_q2b_counts(["1398605"], match_ids={"1398605": 2635})
        assert counts["1398605"]["yazhi"] == 0


class TestContractAndTrace:
    def test_通道标识与SQL契约暴露(self, patch_mysql_connect):
        patch_mysql_connect([[]])
        source = _make_source()
        contracts = source.get_sql_contracts()
        assert contracts["Q1"] == SQL_Q1_MATCH_LIST.strip()
        assert contracts["jc_candidates_mysql_adapted"] == SQL_JC_CANDIDATES_MYSQL
        assert contracts["jc_candidates"] == SQL_JC_MATCH_CANDIDATES.strip()
        assert CHANNEL_MYSQL == "mysql"


class TestDataAgentMySqlChannel:
    """DataAgent 真实通道编排：解析结果必须传给 Q2b（ID 域差异的衔接点）。"""

    def test_解析结果作为match_ids传入Q2b(self):
        stub = StubSource()
        DataAgent(stub, channel=CHANNEL_MYSQL).prepare("2026-09-29")
        fids, match_ids = stub.q2b_calls[0]
        assert fids == ["1398605"]
        assert match_ids == {"1398605": 2635}

    def test_报告溯源通道标记为mysql(self):
        report = DataAgent(StubSource(), channel=CHANNEL_MYSQL).prepare("2026-09-29")
        assert report["matches"][0]["fetch_trace"]["channel"] == CHANNEL_MYSQL


class StubSource:
    """最小数据源桩：验证 DataAgent → Q2b 的 match_ids 传参契约。"""

    def __init__(self):
        self.q2b_calls = []

    def query_match_list(self, match_date):
        return [{"fid": "1398605", "match_no": "周一003",
                 "home_team": "亚美尼亚", "away_team": "黑山"}]

    def query_jc_candidates(self, match_date):
        return [{"id": 2635, "cc": "周一003", "match_date": "2026-09-29",
                 "play_type": "jingcai", "home_team": "亚美尼亚", "away_team": "黑山"}]

    def query_q2b_counts(self, fids, match_ids=None):
        self.q2b_calls.append((list(fids), match_ids))
        return {fid: {"ouzhi": 30, "yazhi": 10, "daxiao": 5,
                      "vote": 10, "stats": 4, "intel": 2} for fid in fids}


class TestErrorPropagation:
    def test_数据库异常不吞掉向上抛(self, monkeypatch):
        def broken_connect(**kwargs):
            raise RuntimeError("db down")

        monkeypatch.setattr(mysql_data_source_mod.pymysql, "connect", broken_connect)
        with pytest.raises(RuntimeError):
            _make_source().query_match_list("2026-09-29")
