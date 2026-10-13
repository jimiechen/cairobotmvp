# -*- coding: utf-8 -*-
"""touzhu 文件解析器单元测试（立项四）。

解析 parse/results/{日期}/{match_id}.json 的 data.touzhu 块（SKILL.md 契约：
bf_trading/bf_large/bf_details）为可直接入库的行结构。

⚠️ 契约假设（宿主未提供 touzhu 文件样例，2026-10-13 主控确认"契约假设先行"）：
- bf_trading.summary 为三方向占比 {home, draw, away}（与 betfair_trading 占比列对齐）
- bf_details 为价位档列表，每档 {price, back_volume, lay_volume}
样例到手后假设不符仅调整解析层，测试随契约同步修订。
"""
from fa_orchestrator.data.touzhu_file_parse import touzhu_block

# 契约假设的最小样例块（结构假设见文件头注释）
TOUZHU_SAMPLE = {
    "bf_trading": {
        "total_volume": 728.0,
        "currency": "HKD",
        "summary": {"home": 22.66, "draw": 17.86, "away": 59.48},
        "bookmaker_pnl": -123.45,
    },
    "bf_large": {"total_volume": 3000.0},
    "bf_details": [
        {"price": 2.0, "back_volume": 100.0, "lay_volume": 50.0},
        {"price": 2.02, "back_volume": 80.5, "lay_volume": 30.0},
    ],
}


class TestTouzhuBlock:
    def test_完整touzhu块解析出standard与large成交行(self):
        result = touzhu_block(TOUZHU_SAMPLE)
        standard = next(r for r in result["trading_rows"]
                        if r["trade_type"] == "standard")
        large = next(r for r in result["trading_rows"]
                     if r["trade_type"] == "large")
        assert standard["total_volume"] == 728.0
        assert standard["home_pct"] == 22.66
        assert standard["draw_pct"] == 17.86
        assert standard["away_pct"] == 59.48
        assert standard["bookmaker_pnl"] == -123.45
        assert large["total_volume"] == 3000.0

    def test_bf_details解析出价位档行(self):
        result = touzhu_block(TOUZHU_SAMPLE)
        assert result["ladder_rows"] == [
            {"trade_type": "standard", "price": 2.0,
             "back_volume": 100.0, "lay_volume": 50.0},
            {"trade_type": "standard", "price": 2.02,
             "back_volume": 80.5, "lay_volume": 30.0},
        ]

    def test_bookmaker_pnl缺失时成交行pnl为None且记missing(self):
        block = {"bf_trading": {"total_volume": 728.0}}
        result = touzhu_block(block)
        assert result["trading_rows"][0]["bookmaker_pnl"] is None
        assert any("bookmaker_pnl" in m for m in result["missing"])

    def test_summary缺失时占比列为None不臆造(self):
        block = {"bf_trading": {"total_volume": 728.0,
                                 "bookmaker_pnl": -1.0}}
        result = touzhu_block(block)
        row = result["trading_rows"][0]
        assert row["home_pct"] is None
        assert row["draw_pct"] is None
        assert row["away_pct"] is None

    def test_无bf_details时ladder_rows为空(self):
        result = touzhu_block({"bf_trading": {"total_volume": 1.0}})
        assert result["ladder_rows"] == []
        assert any("bf_details" in m for m in result["missing"])

    def test_空touzhu块输出空rows且整块缺失标注(self):
        result = touzhu_block({})
        assert result["trading_rows"] == []
        assert result["ladder_rows"] == []
        assert result["missing"] == ["touzhu"]

    def test_None输入输出空rows(self):
        result = touzhu_block(None)
        assert result["trading_rows"] == []
        assert result["ladder_rows"] == []
        assert result["missing"] == ["touzhu"]

    def test_bf_details档位字段缺失容错为None(self):
        block = {"bf_details": [{"price": 2.0}]}
        result = touzhu_block(block)
        assert result["ladder_rows"] == [
            {"trade_type": "standard", "price": 2.0,
             "back_volume": None, "lay_volume": None}]

    def test_bf_large缺失时仅standard成交行(self):
        result = touzhu_block({"bf_trading": TOUZHU_SAMPLE["bf_trading"]})
        assert [r["trade_type"] for r in result["trading_rows"]] == ["standard"]
        assert any("bf_large" in m for m in result["missing"])
