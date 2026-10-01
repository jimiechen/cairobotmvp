# -*- coding: utf-8 -*-
"""复盘对照模块测试（PRD-12 ISS-FA5，D10 赛果对照与规则命中核验）。

覆盖：D10 赛果加载契约、HAD/HHAD 推荐单元 vs 赛果判定、
纪律条款（撤离硬门槛/过热处置）规则命中核验、复盘汇总统计。
"""
import pytest
from pathlib import Path

from fa_orchestrator.analysis.match_review import (
    build_review_report,
    load_match_results,
    review_recommendation,
)

_RESULTS_PATH = str(
    Path(__file__).parent / "fixtures" / "sample_day_20260929_results.json")


def _had_unit(form="S", selection="win", tag=None):
    """构造 HAD 单元（S 单选 / D 双选 / None 放弃）。"""
    if form is None:
        return None
    return {
        "form": form, "selection": selection if form == "S" else None,
        "legs": ["draw", "lose"] if form == "D" else None,
        "path": "A" if form == "D" else None, "stars": 3, "tag": tag,
    }


def _hhand_unit(legs=("draw", "lose")):
    """构造 HHAD 双选单元（legs=None 表示无单元）。"""
    if legs is None:
        return None
    return {"form": "D", "legs": list(legs), "path": "A", "stars": 3, "tag": None}


def _rec(fid="1398594", had_unit="default", hhand_unit="default",
         hard_gate=False, overheat=False, overheat_tag=None):
    """构造复盘对照输入（与 FA4 推荐结构产物同构）。"""
    if had_unit == "default":
        # 过热处置标注（v7.9）落在最终推荐单元 tag 上
        had_unit = _had_unit(tag=overheat_tag)
    if hhand_unit == "default":
        hhand_unit = _hhand_unit()
    return {
        "fid": fid, "match_no": "周二008",
        "recommendation": {"had_unit": had_unit, "hhand_unit": hhand_unit},
        "synthesis": {
            "evacuation": {"count": 3 if hard_gate else 1,
                           "signals": ["机构大幅升水"] if hard_gate else ["必发成交回落"],
                           "hard_gate": hard_gate},
        },
        "sentiment": {"overheat": {
            "triggered": overheat, "layer": "had" if overheat else None,
            "deviation_pp": 22.0 if overheat else None,
            "support_intel": "yes" if overheat else None,
            "disposition": "keep_downgrade" if overheat else None,
            "tag": overheat_tag,
        }},
    }


def _result(had="win", hhad="win"):
    """构造 D10 赛果。"""
    return {"home_goals": 2, "away_goals": 1, "had_result": had, "hhad_result": hhad}


class TestLoadMatchResults:
    def test_加载赛果fixture返回fid映射(self):
        results = load_match_results(_RESULTS_PATH)
        assert set(results) == {str(fid) for fid in range(1398594, 1398601)}
        assert results["1398594"]["had_result"] == "win"

    def test_赛果文件缺失报错(self):
        with pytest.raises(FileNotFoundError):
            load_match_results("tests/fixtures/no_such_results.json")

    def test_赛果字段缺失报错(self, tmp_path):
        bad = tmp_path / "bad.json"
        bad.write_text('{"match_results": [{"fid": "1"}]}', encoding="utf-8")
        with pytest.raises(ValueError):
            load_match_results(str(bad))


class TestHadUnitReview:
    def test_S单选命中(self):
        out = review_recommendation(_rec(), _result(had="win"))
        assert out["had_hit"] is True

    def test_S单选未命中(self):
        out = review_recommendation(_rec(), _result(had="lose"))
        assert out["had_hit"] is False

    def test_D双选命中(self):
        out = review_recommendation(
            _rec(had_unit=_had_unit("D")), _result(had="lose"))
        assert out["had_hit"] is True

    def test_D双选未命中(self):
        out = review_recommendation(
            _rec(had_unit=_had_unit("D")), _result(had="win"))
        assert out["had_hit"] is False

    def test_放弃场不计胜负(self):
        out = review_recommendation(
            _rec(had_unit=None, hhand_unit=None), _result())
        assert out["had_hit"] is None and out["hhand_hit"] is None
        assert out["verdict"] == "given_up"


class TestHhandUnitReview:
    def test_让球双选命中腿(self):
        out = review_recommendation(_rec(), _result(hhad="lose"))
        assert out["hhand_hit"] is True

    def test_让球双选未命中(self):
        out = review_recommendation(_rec(), _result(hhad="win"))
        assert out["hhand_hit"] is False

    def test_让球单元缺失不判定(self):
        out = review_recommendation(_rec(hhand_unit=None), _result())
        assert out["hhand_hit"] is None


class TestVerdict:
    def test_两层全命中(self):
        out = review_recommendation(_rec(), _result(had="win", hhad="lose"))
        assert out["verdict"] == "hit"

    def test_两层全未命中(self):
        out = review_recommendation(
            _rec(), _result(had="lose", hhad="win"))
        assert out["verdict"] == "miss"

    def test_部分命中_一层命中一层未命中(self):
        out = review_recommendation(
            _rec(), _result(had="win", hhad="win"))
        assert out["verdict"] == "partial"

    def test_单层推荐命中即命中(self):
        out = review_recommendation(
            _rec(hhand_unit=None), _result(had="win"))
        assert out["verdict"] == "hit"

    def test_赛果缺失(self):
        out = review_recommendation(_rec(), None)
        assert out["verdict"] == "missing_result"
        assert out["had_hit"] is None


class TestRuleChecks:
    def test_撤离硬门槛合规_双选(self):
        out = review_recommendation(
            _rec(hard_gate=True, had_unit=_had_unit("D")), _result())
        check = out["rule_checks"][0]
        assert check["id"] == "①" and check["result"] == "pass"

    def test_撤离硬门槛合规_配平负(self):
        out = review_recommendation(
            _rec(hard_gate=True, had_unit=_had_unit("S", "draw")), _result())
        assert out["rule_checks"][0]["result"] == "pass"

    def test_撤离硬门槛合规_放弃(self):
        out = review_recommendation(
            _rec(hard_gate=True, had_unit=None, hhand_unit=None), _result())
        assert out["rule_checks"][0]["result"] == "pass"

    def test_撤离硬门槛违规_保留主胜单选(self):
        out = review_recommendation(_rec(hard_gate=True), _result())
        check = out["rule_checks"][0]
        assert check["id"] == "①" and check["result"] == "violation"

    def test_撤离未触发跳过(self):
        out = review_recommendation(_rec(hard_gate=False), _result())
        assert out["rule_checks"][0]["result"] == "not_triggered"

    def test_过热处置合规_带标签(self):
        out = review_recommendation(
            _rec(overheat=True, overheat_tag="[过热+22pp｜情报支撑：有]"), _result())
        check = out["rule_checks"][1]
        assert check["id"] == "②" and check["result"] == "pass"

    def test_过热处置合规_标注在过热层待下沉(self):
        # FA3 实际链路：过热层为 hhand 时 tag 生成在 sentiment.overheat.tag
        # （disposition 分档同时给出），标注下沉推荐单元属 FA3 后续补强
        rec = _rec(overheat=True)
        rec["sentiment"]["overheat"]["tag"] = "[过热+16.7pp｜情报支撑：有]"
        out = review_recommendation(rec, _result())
        assert out["rule_checks"][1]["result"] == "pass"

    def test_过热处置违规_无标注(self):
        out = review_recommendation(_rec(overheat=True), _result())
        assert out["rule_checks"][1]["result"] == "violation"

    def test_未过热跳过(self):
        out = review_recommendation(_rec(overheat=False), _result())
        assert out["rule_checks"][1]["result"] == "not_triggered"


class TestBuildReviewReport:
    def test_汇总统计命中未命中(self):
        recs = [
            _rec("1398594"),                       # had win + hhad lose → 双命中
            _rec("1398595", had_unit=_had_unit("S", "lose")),  # had miss
        ]
        results = {"1398594": _result("win", "lose"),
                   "1398595": _result("win", "win")}
        report = build_review_report(
            [review_recommendation(r, results.get(r["fid"])) for r in recs],
            match_date="2026-09-29",
        )
        assert report["total"] == 2 and report["hits"] == 1 and report["misses"] == 1
        assert report["match_date"] == "2026-09-29"

    def test_赛果缺失与放弃计数(self):
        recs = [_rec("1398596", had_unit=None, hhand_unit=None), _rec("1398599")]
        results = {"1398596": _result()}
        report = build_review_report(
            [review_recommendation(r, results.get(r["fid"])) for r in recs],
            match_date="2026-09-29",
        )
        assert report["given_up"] == 1 and report["missing_result"] == 1

    def test_规则违规汇总(self):
        recs = [_rec("1398594", hard_gate=True), _rec("1398595", overheat=True)]
        results = {r["fid"]: _result() for r in recs}
        report = build_review_report(
            [review_recommendation(r, results[r["fid"]]) for r in recs],
            match_date="2026-09-29",
        )
        assert len(report["rule_violations"]) == 2
