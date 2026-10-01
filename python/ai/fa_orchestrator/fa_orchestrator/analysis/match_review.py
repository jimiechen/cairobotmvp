# -*- coding: utf-8 -*-
"""复盘对照模块（PRD-12 ISS-FA5）：D10 赛果 vs 推荐结构对照与规则命中核验。

本文件负责：
- D10 赛果 fixture 加载（MySQL 层无现成 D10 查询契约，已登记遗留事项）
- HAD/HHAD 推荐单元 vs 赛果命中判定（hit/partial/miss/given_up/missing_result）
- 纪律条款命中核验：撤离硬门槛（SKILL v7.9/v7.11）、过热处置（v7.9）

不负责：报告渲染（html_report.py）、CLI 编排（cli.py）、图执行（graph_nodes.py）。

相关文档：
- PRD-12 §12 数据块 D10、§14 验收标准
- docs/drafts/PRD-12-M1-M5-Issue拆分草案.md ISS-FA5 节
"""
import json
from pathlib import Path

# 赛果行必需字段：缺任一即契约破损，禁止静默跳过
_REQUIRED_RESULT_FIELDS = ("fid", "had_result", "hhad_result")

VERDICT_HIT = "hit"
VERDICT_PARTIAL = "partial"
VERDICT_MISS = "miss"
VERDICT_GIVEN_UP = "given_up"
VERDICT_MISSING = "missing_result"

# 核验项 ①：撤离/反向信号 ≥3 重硬门槛（v7.11 机构动作优先于情报层）
_CHECK_EVACUATION = {"id": "①", "name": "撤离硬门槛合规（v7.9/v7.11）"}
# 核验项 ②：过热场处置标注（v7.9 偏差 ≥+20pp 强制分档处置）
_CHECK_OVERHEAT = {"id": "②", "name": "过热处置合规（v7.9）"}


def load_match_results(path: str) -> dict:
    """加载 D10 赛果 fixture，返回 {fid: 赛果dict} 映射。

    输入：赛果 JSON 路径（含 match_results 列表）。
    输出：fid → 赛果行（home_goals/away_goals/had_result/hhad_result）。
    失败行为：文件缺失抛 FileNotFoundError；赛果行缺必需字段抛 ValueError。
    """
    raw = json.loads(Path(path).read_text(encoding="utf-8"))
    results = {}
    for row in raw.get("match_results") or []:
        missing = [field for field in _REQUIRED_RESULT_FIELDS if field not in row]
        if missing:
            raise ValueError(f"赛果行缺字段 {missing}: {row}")
        results[row["fid"]] = row
    return results


def _unit_hit(unit: dict | None, actual: str | None) -> bool | None:
    """单层推荐单元命中判定。

    输入：unit 为 had_unit/hhand_unit（None 表示放弃或无该层单元），
         actual 为该层赛果方向（win/draw/lose，None 表示赛果缺失）。
    输出：True/False 命中；None 表示无法判定（无单元或无赛果）。
    """
    if unit is None or actual is None:
        return None
    if unit.get("form") == "D":
        return actual in (unit.get("legs") or [])
    return unit.get("selection") == actual


def _verdict(had_hit: bool | None, hhand_hit: bool | None, has_result: bool) -> str:
    """两层命中组合 → 总判定。

    前置：had_hit/hhand_hit 来自 _unit_hit。
    语义：赛果缺失优先判 missing_result；两层均无单元（放弃场）判 given_up；
         有单元层全中 hit、部分中 partial、全不中 miss。
    """
    if not has_result:
        return VERDICT_MISSING
    if had_hit is None and hhand_hit is None:
        return VERDICT_GIVEN_UP
    hits = [hit for hit in (had_hit, hhand_hit) if hit is not None]
    if all(hits):
        return VERDICT_HIT
    if any(hits):
        return VERDICT_PARTIAL
    return VERDICT_MISS


def _check_evacuation(rec: dict) -> dict:
    """撤离硬门槛核验（SKILL v7.9 放弃规则 > 反大众阻断；v7.11 机构动作优先）。

    硬门槛触发时期望：HAD 单元放弃 / 双选 / 配平负——禁止保留主推单选。
    输出：{id, name, result: pass/violation/not_triggered, note}。
    """
    base = dict(_CHECK_EVACUATION)
    hard_gate = ((rec.get("synthesis") or {}).get("evacuation") or {}).get("hard_gate")
    if not hard_gate:
        return {**base, "result": "not_triggered", "note": ""}
    had_unit = (rec.get("recommendation") or {}).get("had_unit")
    compliant = (
        had_unit is None
        or had_unit.get("form") == "D"
        or had_unit.get("selection") in ("draw", "lose")
    )
    note = "" if compliant else "撤离≥3重仍保留主推单选（v7.11 违规）"
    return {**base, "result": "pass" if compliant else "violation", "note": note}


def _check_overheat(rec: dict) -> dict:
    """过热处置核验（v7.9：偏差 ≥+20pp 场必须分档处置并强制标注）。

    合规口径：处置标注在推荐单元 tag，或生成在 sentiment.overheat.tag 且
    disposition 分档给出（FA3 实际链路：过热层为 hhand 时标注先落在
    overheat 层，下沉推荐单元属后续补强）。
    输出：{id, name, result: pass/violation/not_triggered, note}。
    """
    base = dict(_CHECK_OVERHEAT)
    overheat = (rec.get("sentiment") or {}).get("overheat") or {}
    if not overheat.get("triggered"):
        return {**base, "result": "not_triggered", "note": ""}
    had_unit = (rec.get("recommendation") or {}).get("had_unit")
    unit_tagged = bool(had_unit and "过热" in (had_unit.get("tag") or ""))
    layer_tagged = (
        bool(overheat.get("tag") and "过热" in overheat["tag"])
        and bool(overheat.get("disposition"))
    )
    compliant = unit_tagged or layer_tagged
    note = "" if compliant else "过热场未生成处置标注或档位（v7.9 违规）"
    return {**base, "result": "pass" if compliant else "violation", "note": note}


def review_recommendation(rec: dict, result: dict | None) -> dict:
    """单场复盘对照：推荐单元 vs D10 赛果 + 纪律条款核验。

    输入：rec 为推荐结构产物单场（含 recommendation/synthesis/sentiment），
         result 为 D10 赛果行（None 表示该场赛果缺失）。
    输出：{fid, match_no, verdict, had_hit, hhand_hit, rule_checks}。
    前置：rec 与 FA4 推荐结构产物同构。
    """
    recommendation = rec.get("recommendation") or {}
    had_hit = _unit_hit(recommendation.get("had_unit"),
                        (result or {}).get("had_result"))
    hhand_hit = _unit_hit(recommendation.get("hhand_unit"),
                          (result or {}).get("hhad_result"))
    return {
        "fid": rec.get("fid"),
        "match_no": rec.get("match_no"),
        "verdict": _verdict(had_hit, hhand_hit, result is not None),
        "had_hit": had_hit,
        "hhand_hit": hhand_hit,
        "rule_checks": [_check_evacuation(rec), _check_overheat(rec)],
    }


def build_review_report(reviews: list, match_date: str) -> dict:
    """汇总单场复盘结果为复盘报告。

    输入：reviews 为 review_recommendation 输出列表；match_date 为竞彩日。
    输出：{match_date, total, hits/partial/misses/given_up/missing_result,
          rule_violations, matches}。
    """
    verdicts = (VERDICT_HIT, VERDICT_PARTIAL, VERDICT_MISS,
                VERDICT_GIVEN_UP, VERDICT_MISSING)
    counts = {key: 0 for key in verdicts}
    violations = []
    for item in reviews:
        counts[item["verdict"]] += 1
        for check in item["rule_checks"]:
            if check["result"] == "violation":
                violations.append({
                    "fid": item["fid"],
                    "rule": f"{check['id']}{check['name']}",
                    "note": check["note"],
                })
    return {
        "match_date": match_date,
        "total": len(reviews),
        "hits": counts[VERDICT_HIT],
        "partial": counts[VERDICT_PARTIAL],
        "misses": counts[VERDICT_MISS],
        "given_up": counts[VERDICT_GIVEN_UP],
        "missing_result": counts[VERDICT_MISSING],
        "rule_violations": violations,
        "matches": reviews,
    }
