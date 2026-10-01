# -*- coding: utf-8 -*-
"""ReportAgent 单元测试（PRD-12 F-15，ISS-FA4）。

验证：模拟数据检测告警（整数赔率/全同赔/缺主流公司/欧赔缺失）、
报告模板输出（数据来源声明 + 预测表 + 反大众检查 + 矛盾检查章节）。
全程使用构造数据，不依赖外部服务。
"""
from fa_orchestrator.report.report_agent import ReportAgent, detect_data_anomalies

# 主流公司样例（fixture 同构）：威廉希尔/立博/澳门
_NORMAL_EURO = [
    {"bk": "威廉希尔", "current": {"win": 2.05, "draw": 3.3, "lose": 3.55}},
    {"bk": "立博", "current": {"win": 2.1, "draw": 3.25, "lose": 3.5}},
    {"bk": "澳门", "current": {"win": 2.02, "draw": 3.3, "lose": 3.55}},
]

# 构造最小推荐结构（与 FA3 SynthesisAgent 输出键对齐）
_HAD_UNIT = {"form": "S", "selection": "win", "legs": None, "path": None,
             "stars": 3, "tag": None}
_HHAND_UNIT = {"form": "D", "selection": None, "legs": ["让胜", "让平"],
               "path": "A", "stars": 2, "tag": None}


def _match(recommendation=None, checks=None, decision="pass") -> dict:
    """构造 per-match 报告输入（与 AnalysisRunner 产出键对齐）。"""
    return {
        "fid": "1398594",
        "match_no": "周二008",
        "home_team": "主队",
        "away_team": "客队",
        "synthesis": {
            "main_action": "keep",
            "anti_public": {"anchor1": "none", "anchor2": False, "anchor3": False,
                            "allow_counter": True, "note": "三条锚点全不命中，允许反买"},
            "recommendation": recommendation or {
                "had_unit": _HAD_UNIT, "hhand_unit": _HHAND_UNIT,
                "mutual_exclusion": {}, "notes": [],
            },
        },
        "review": {"phase": "review", "decision": decision,
                   "violations": [],
                   "checks": checks or [
                       {"id": "①", "name": "风险提示与大小球方向一致",
                        "result": "na", "action": None, "note": "未产出层"},
                       {"id": "②", "name": "次选降级替换", "result": "pass",
                        "action": None, "note": None},
                   ]},
    }


class TestDetectDataAnomalies:
    def test_整数赔率触发模拟数据告警(self):
        euro = [{"bk": "澳门", "current": {"win": 2.0, "draw": 3.0, "lose": 4.0}}]
        warnings = detect_data_anomalies(euro)
        assert any("整数赔率" in w and "澳门" in w for w in warnings)

    def test_全同赔触发告警(self):
        euro = [
            {"bk": "威廉希尔", "current": {"win": 2.1, "draw": 3.2, "lose": 3.4}},
            {"bk": "立博", "current": {"win": 2.1, "draw": 3.2, "lose": 3.4}},
        ]
        assert any("全同赔" in w for w in detect_data_anomalies(euro))

    def test_缺主流公司触发告警(self):
        euro = [{"bk": "澳门", "current": {"win": 2.02, "draw": 3.3, "lose": 3.55}}]
        assert any("缺主流公司" in w for w in detect_data_anomalies(euro))

    def test_欧赔缺失返回缺失告警(self):
        assert any("欧赔缺失" in w for w in detect_data_anomalies([]))

    def test_正常赔率无告警(self):
        assert detect_data_anomalies(_NORMAL_EURO) == []


class TestReportAgentGenerate:
    def _generate(self, tmp_path=None, **overrides) -> str:
        match = overrides.get("match", _match())
        euro = overrides.get("euro", _NORMAL_EURO)
        match["odds_snapshot"] = {"euro": euro}
        return ReportAgent().generate(
            match, channel="fixture", snapshot_time="09-29 20:10",
        )

    def test_报告含数据来源声明与快照时间(self):
        report = self._generate()
        assert "数据来源声明" in report
        assert "fixture" in report
        assert "09-29 20:10" in report

    def test_报告含竞彩让球胜平负预测表(self):
        report = self._generate()
        # 胜平负单选 win → 胜；让球双选 path A
        assert "胜平负" in report
        assert "胜" in report
        assert "让球" in report
        assert "A" in report

    def test_报告含反大众检查章节(self):
        assert "允许反买" in self._generate()

    def test_报告含矛盾检查章节与决策(self):
        report = self._generate()
        assert "矛盾检查" in report
        assert "①" in report and "②" in report
        assert "pass" in report

    def test_模拟数据告警写入报告(self):
        euro = [{"bk": "澳门", "current": {"win": 2.0, "draw": 3.0, "lose": 4.0}}]
        report = self._generate(euro=euro)
        assert "整数赔率" in report

    def test_无告警时报告明示未触发(self):
        assert "未触发" in self._generate()

    def test_模板占位符无残留(self):
        assert "{{" not in self._generate()
