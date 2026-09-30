# -*- coding: utf-8 -*-
"""Q2b 六模块完整性核查（PRD-12 F-08）。

核查口径：ouzhi/yazhi/daxiao/vote/stats/intel 六模块行数 > 0 视为就绪。
不完整必须列出缺口触发补抓；全缺/无记录落 blocked 状态，禁止静默跳过
（继承上游硬性规则：抓取失败必须回写 error/blocked）。
"""
from fa_orchestrator.sql_contracts import INTEGRITY_MODULES

# 状态机：ready（全就绪）/ incomplete（部分缺口）/ blocked（全缺，禁止静默跳过）
STATUS_READY = "ready"
STATUS_INCOMPLETE = "incomplete"
STATUS_BLOCKED = "blocked"


def check_module_integrity(fid: str, counts: dict) -> dict:
    """单场六模块完整性核查。

    输入：fid（500com 场次 ID）、counts（Q2b 返回的 {模块名: 行数}，可为空）。
    输出：{fid, ready, status, missing_modules, unknown_modules}。
    失败行为：出现白名单外模块或缺口均判 not ready，不抛异常（核查是信息聚合）。
    """
    known_modules = set(INTEGRITY_MODULES)
    unknown_modules = sorted(set(counts) - known_modules)

    missing_modules = sorted(
        module for module in INTEGRITY_MODULES
        if counts.get(module, 0) == 0
    )

    # Q2b 无记录或六模块全缺 → blocked（继承上游：禁止静默跳过）
    if not counts or len(missing_modules) == len(INTEGRITY_MODULES):
        status = STATUS_BLOCKED
    elif missing_modules or unknown_modules:
        status = STATUS_INCOMPLETE
    else:
        status = STATUS_READY

    return {
        "fid": fid,
        "ready": status == STATUS_READY,
        "status": status,
        "missing_modules": missing_modules,
        "unknown_modules": unknown_modules,
    }


def build_integrity_report(per_match_counts: dict) -> dict:
    """构建日级就绪率报告（ISS-FA1 交付物）。

    输入：per_match_counts（{fid: {模块名: 行数}}）。
    输出：{match_date 由调用方补充, total_matches, ready_matches, ready_rate, matches}。
    失败行为：空输入返回零报告，不抛异常。
    """
    match_reports = [
        check_module_integrity(fid, counts)
        for fid, counts in per_match_counts.items()
    ]
    total = len(match_reports)
    ready = sum(1 for item in match_reports if item["ready"])
    return {
        "total_matches": total,
        "ready_matches": ready,
        "ready_rate": (ready / total) if total else 0.0,
        "matches": match_reports,
    }
