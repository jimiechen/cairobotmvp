# -*- coding: utf-8 -*-
"""Q2b 六模块完整性核查测试（PRD-12 F-08）。

核查口径：ouzhi/yazhi/daxiao/vote/stats/intel 六模块行数 > 0 视为就绪；
不完整必须列出缺口并触发补抓指令，全缺/无记录必须落 blocked 状态，禁止静默跳过。
"""
from fa_orchestrator.integrity import (
    INTEGRITY_MODULES,
    build_integrity_report,
    check_module_integrity,
)


class TestSingleMatchIntegrity:
    """单场六模块核查。"""

    def test_六模块全部大于零为就绪(self):
        counts = {m: 1 for m in INTEGRITY_MODULES}
        report = check_module_integrity("1398594", counts)
        assert report["ready"] is True
        assert report["fid"] == "1398594"
        assert report["missing_modules"] == []
        assert report["status"] == "ready"

    def test_部分模块为零列出缺口(self):
        counts = dict.fromkeys(INTEGRITY_MODULES, 5)
        counts["vote"] = 0
        report = check_module_integrity("1398599", counts)
        assert report["ready"] is False
        assert report["missing_modules"] == ["vote"]
        assert report["status"] == "incomplete"

    def test_Q2b无记录视为全缺(self):
        # Q2b 按行分组返回，若某场完全无行则 counts 为空
        report = check_module_integrity("9999999", {})
        assert report["ready"] is False
        assert report["status"] == "blocked"
        assert set(report["missing_modules"]) == set(INTEGRITY_MODULES)

    def test_blocked状态禁止静默跳过(self):
        counts = {"ouzhi": 0, "yazhi": 0, "daxiao": 0, "vote": 0, "stats": 0, "intel": 0}
        report = check_module_integrity("1398599", counts)
        assert report["status"] == "blocked"
        assert report["ready"] is False

    def test_未知模块被拒绝(self):
        counts = {"ouzhi": 5, "unknown_module": 3}
        report = check_module_integrity("1398594", counts)
        assert report["ready"] is False
        assert report["status"] == "incomplete"


class TestDayReport:
    """日级就绪率报告（ISS-FA1 交付物）。"""

    def _full_counts(self):
        return {m: 5 for m in INTEGRITY_MODULES}

    def test_就绪率计算_七场六场就绪(self):
        per_match = {
            str(1398594 + i): dict(self._full_counts()) for i in range(7)
        }
        # 第 6 场（1398599）缺 vote
        per_match["1398599"] = dict(self._full_counts(), vote=0)
        report = build_integrity_report(per_match)
        assert report["total_matches"] == 7
        assert report["ready_matches"] == 6
        assert abs(report["ready_rate"] - 6 / 7) < 1e-9

    def test_报告含每场明细与缺口(self):
        per_match = {
            "1398594": dict(self._full_counts()),
            "1398599": dict(self._full_counts(), vote=0, intel=0),
        }
        report = build_integrity_report(per_match)
        by_fid = {m["fid"]: m for m in report["matches"]}
        # 缺口按字典序输出（确定性）
        assert by_fid["1398599"]["missing_modules"] == ["intel", "vote"]

    def test_空日报告(self):
        report = build_integrity_report({})
        assert report["total_matches"] == 0
        assert report["ready_rate"] == 0.0
