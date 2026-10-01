# -*- coding: utf-8 -*-
"""Standalone HTML 报告生成测试（PRD-12 ISS-FA5 / AGENTS.md §15）。

覆盖：自包含约束（无外链可离线打开）、复盘汇总表、验收标准核验表、
Bug 与风险节、落盘行为。
"""
import json

from fa_orchestrator.html_report import render_html_report, write_html_report


def _review_report():
    """构造最小复盘汇总（与 build_review_report 产物同构）。"""
    return {
        "match_date": "2026-09-29",
        "total": 7, "hits": 3, "partial": 1, "misses": 2,
        "given_up": 1, "missing_result": 0,
        "rule_violations": [],
        "matches": [
            {"fid": "1398594", "match_no": "周二008", "verdict": "hit",
             "had_hit": True, "hhand_hit": True, "rule_checks": []},
            {"fid": "1398595", "match_no": "周二009", "verdict": "miss",
             "had_hit": False, "hhand_hit": False, "rule_checks": []},
        ],
    }


def _acceptance():
    """构造 §14 验收标准核验行（8 条）。"""
    return [
        {"id": idx, "item": f"验收条目{idx}", "result": "pass", "evidence": "见测试报告"}
        for idx in range(1, 9)
    ]


class TestRenderHtmlReport:
    def test_自包含无外链可离线打开(self):
        html = render_html_report(_review_report(), _acceptance(), title="验收报告")
        assert html.startswith("<!DOCTYPE html>")
        assert "<link" not in html
        assert "<script src" not in html
        assert "http://" not in html and "https://" not in html

    def test_含复盘汇总与场次明细(self):
        html = render_html_report(_review_report(), _acceptance(), title="验收报告")
        assert "1398594" in html and "周二008" in html
        assert "命中" in html and "2026-09-29" in html

    def test_含验收八条核验表(self):
        html = render_html_report(_review_report(), _acceptance(), title="验收报告")
        assert "验收条目1" in html and "验收条目8" in html

    def test_含遗留事项与风险节(self):
        meta = {
            "leftovers": ["D9 变盘史无独立表（维持上游结论，仅登记）"],
            "risks": ["D10 赛果为演示数据"],
        }
        html = render_html_report(_review_report(), _acceptance(),
                                  title="验收报告", meta=meta)
        assert "D9 变盘史无独立表" in html and "演示数据" in html

    def test_规则违规场次高亮(self):
        report = _review_report()
        report["rule_violations"] = [{"fid": "1398595", "rule": "①撤离硬门槛"}]
        html = render_html_report(report, _acceptance(), title="验收报告")
        assert "①撤离硬门槛" in html


class TestWriteHtmlReport:
    def test_落盘utf8(self, tmp_path):
        out = tmp_path / "report.html"
        write_html_report(
            render_html_report(_review_report(), _acceptance(), title="验收报告"),
            str(out),
        )
        content = out.read_text(encoding="utf-8")
        assert content.startswith("<!DOCTYPE html>")

    def test_落盘内容为合法自包含文档(self, tmp_path):
        out = tmp_path / "report.html"
        write_html_report(
            render_html_report(_review_report(), _acceptance(), title="验收报告"),
            str(out),
        )
        assert "</html>" in out.read_text(encoding="utf-8")
