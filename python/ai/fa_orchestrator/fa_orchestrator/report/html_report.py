# -*- coding: utf-8 -*-
"""Standalone HTML 报告生成（PRD-12 §16 / AGENTS.md §15，ISS-FA5）。

本文件负责：把复盘对照汇总 + 验收标准核验结果渲染为自包含 HTML
（内联 CSS、零外链、离线可打开），并落盘。
不负责：复盘对照逻辑（match_review.py）、验收结论判定（人工/主控评审）。

输出目录约束：docs/reports/html/（AGENTS.md §15）。
"""
import html
from pathlib import Path

# 内联样式：离线自包含约束下禁用任何外链资源（字体/脚本/样式表）
_CSS = """
body { font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
       margin: 24px; color: #1f2328; }
h1 { border-bottom: 2px solid #0969da; padding-bottom: 8px; }
h2 { margin-top: 28px; }
table { border-collapse: collapse; width: 100%; margin: 12px 0; }
th, td { border: 1px solid #d0d7de; padding: 6px 10px; text-align: left; }
th { background: #f6f8fa; }
.pass { color: #1a7f37; font-weight: 600; }
.violation { color: #cf222e; font-weight: 600; }
.na { color: #57606a; }
.summary { background: #f6f8fa; padding: 12px 16px; border-radius: 6px; }
"""

_VERDICT_LABELS = {
    "hit": "命中", "partial": "部分命中", "miss": "未命中",
    "given_up": "已放弃", "missing_result": "赛果缺失",
}


def _esc(value) -> str:
    """HTML 转义（None 安全），防注入并保证合法文档。"""
    return html.escape(str(value)) if value is not None else "-"


def _bool_label(value) -> str:
    """布尔命中 → 中文标签（None → 未判定）。"""
    if value is None:
        return '<span class="na">未判定</span>'
    return "命中" if value else "未命中"


def _table(headers: list, rows: list) -> str:
    """渲染简单表格（内容全部转义）。"""
    head = "".join(f"<th>{_esc(h)}</th>" for h in headers)
    body = "".join(
        "<tr>" + "".join(f"<td>{cell}</td>" for cell in row) + "</tr>"
        for row in rows
    )
    return f"<table><thead><tr>{head}</tr></thead><tbody>{body}</tbody></table>"


def _render_matches(review_report: dict) -> str:
    """渲染复盘汇总表与场次明细（含规则违规高亮）。"""
    summary = (
        f'<div class="summary">竞彩日 {_esc(review_report["match_date"])}｜'
        f"共 {_esc(review_report['total'])} 场："
        f"命中 {_esc(review_report['hits'])}｜部分命中 {_esc(review_report['partial'])}｜"
        f"未命中 {_esc(review_report['misses'])}｜已放弃 {_esc(review_report['given_up'])}｜"
        f"赛果缺失 {_esc(review_report['missing_result'])}</div>"
    )
    rows = []
    for item in review_report["matches"]:
        violation_texts = [
            f'{_esc(c["id"])}{_esc(c["name"])}'
            for c in item.get("rule_checks") or [] if c["result"] == "violation"
        ]
        rows.append([
            _esc(item["fid"]), _esc(item["match_no"]),
            f'<span class="{"violation" if violation_texts else "pass"}">'
            f'{_esc(_VERDICT_LABELS.get(item["verdict"], item["verdict"]))}</span>',
            _bool_label(item["had_hit"]), _bool_label(item["hhand_hit"]),
            "<br>".join(violation_texts) or '<span class="na">无</span>',
        ])
    table = _table(
        ["fid", "场次", "判定", "胜平负单元", "让球单元", "规则违规"], rows)
    violations = review_report.get("rule_violations") or []
    if violations:
        v_rows = [[_esc(v["fid"]), _esc(v["rule"]), _esc(v.get("note", ""))]
                  for v in violations]
        table += "<h3>规则违规汇总</h3>" + _table(["fid", "规则", "说明"], v_rows)
    return summary + table


def _render_acceptance(acceptance: list) -> str:
    """渲染 §14 验收标准核验表（8 条逐项）。"""
    rows = []
    for item in acceptance:
        css = {"pass": "pass", "violation": "violation"}.get(item["result"], "na")
        rows.append([
            _esc(item["id"]), _esc(item["item"]),
            f'<span class="{css}">{_esc(item["result"])}</span>',
            _esc(item["evidence"]),
        ])
    return _table(["编号", "验收标准", "结论", "证据"], rows)


def _render_meta(meta: dict | None) -> str:
    """渲染遗留事项与风险节（ISS-FA5 登记：D9 无独立表、checkGate 口径）。"""
    if not meta:
        return ""
    parts = []
    if meta.get("leftovers"):
        items = "".join(f"<li>{_esc(x)}</li>" for x in meta["leftovers"])
        parts.append(f"<h2>遗留事项登记</h2><ul>{items}</ul>")
    if meta.get("risks"):
        items = "".join(f"<li>{_esc(x)}</li>" for x in meta["risks"])
        parts.append(f"<h2>风险</h2><ul>{items}</ul>")
    return "".join(parts)


def render_html_report(review_report: dict, acceptance: list | None,
                       title: str, meta: dict | None = None) -> str:
    """渲染自包含 HTML 验收报告。

    输入：review_report 为 build_review_report 产物；acceptance 为 §14
         核验行列表（None 表示不渲染该节）；meta 携带遗留事项与风险。
    输出：完整 HTML 字符串（零外链，离线可打开）。
    """
    sections = [f"<h1>{_esc(title)}</h1>",
                "<h2>复盘对照汇总（D10）</h2>",
                _render_matches(review_report)]
    if acceptance:
        sections += ["<h2>PRD-12 §14 验收标准核验</h2>",
                     _render_acceptance(acceptance)]
    sections.append(_render_meta(meta))
    body = "".join(sections)
    return ("<!DOCTYPE html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\">"
            f"<title>{_esc(title)}</title><style>{_CSS}</style></head>"
            f"<body>{body}</body></html>")


def write_html_report(content: str, output_path: str) -> None:
    """HTML 报告落盘（UTF-8）。失败行为：路径不可写抛 OSError。"""
    Path(output_path).write_text(content, encoding="utf-8")
