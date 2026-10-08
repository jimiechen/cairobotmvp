# -*- coding: utf-8 -*-
"""DataAgent：Q1 取清单 → ID 四路径解析 → Q2b 完整性核查 → 就绪率报告。

职责边界（PRD-12 F-06~F-09）：
- 负责：编排取数流程（Q1/Q2b）、fid→match.id 解析、日级就绪率报告、溯源快照。
- 不负责：SQL 契约文本（sql_contracts.py）、数据源实现（data_source.py）、
  解析算法细节（id_resolver.py）、缺口补抓执行（后续 Issue）。
"""
from fa_orchestrator.data.data_source import (
    BLOCK_D1_MATCH_LIST,
    BLOCK_D2B_INTEGRITY,
    CHANNEL_FIXTURE,
    make_fetch_trace,
)
from fa_orchestrator.data.id_resolver import MatchResolveError, resolve_match_id
from fa_orchestrator.data.integrity import build_integrity_report, check_module_integrity


class DataAgent:
    """数据准备 Agent：输入竞彩日，输出就绪率报告。

    状态约束：无状态；数据源注入（测试/演示用 fixture，联调用 MySQL 通道）。
    """

    def __init__(self, data_source, channel: str = CHANNEL_FIXTURE):
        self.data_source = data_source
        self.channel = channel

    def prepare(self, match_date: str) -> dict:
        """执行数据准备并产出日级就绪率报告。

        输入：竞彩日（如 2026-09-29）。
        输出：就绪率报告（含 total_matches/ready_matches/ready_rate/matches）。
        失败行为：单场 ID 解析失败不中断流程，该场标记 unresolved 保留在报告中
        （禁止静默跳过）；数据源异常向上抛出。
        """
        matches = self.data_source.query_match_list(match_date)
        match_ids = self._resolve_all(matches, match_date)
        # mysql 组装路径衔接（方案 A 批次 2）：模块表与分析表 ID 域均为
        # `match`.id，解析结果回注支持该接口的数据源，供后续分析查询定位
        # （fixture 等无此接口的数据源跳过，保持向后兼容）
        inject = getattr(self.data_source, "set_match_ids", None)
        if inject is not None:
            inject(match_ids)
        per_match_counts = self._fetch_q2b_counts(matches, match_ids)

        report = build_integrity_report(per_match_counts)
        report["match_date"] = match_date
        self._enrich_report(report, matches, match_ids)
        return report

    def _resolve_all(self, matches: list, match_date: str) -> dict:
        """批量四路径解析：fid → match.id；失败记 None（unresolved）。"""
        candidates = self.data_source.query_jc_candidates()
        match_ids = {}
        for row in matches:
            try:
                match_ids[row["fid"]] = resolve_match_id(
                    match_no=row["match_no"],
                    match_date=match_date,
                    home_team=row["home_team"],
                    away_team=row["away_team"],
                    rows=candidates,
                )
            except MatchResolveError:
                match_ids[row["fid"]] = None
        return match_ids

    def _fetch_q2b_counts(self, matches: list, match_ids: dict = None) -> dict:
        """批量 Q2b 六模块行数核查。

        match_ids（fid → match.id）透传给数据源：MySQL 通道模块表的
        ID 域是 `match`.id（非 fid），必须以解析结果定位；fixture 忽略。
        """
        fids = [row["fid"] for row in matches]
        return self.data_source.query_q2b_counts(fids, match_ids=match_ids)

    def _enrich_report(self, report: dict, matches: list, match_ids: dict) -> None:
        """重建每场明细：基础信息 + integrity 子对象 + 溯源快照。"""
        integrity_by_fid = {item["fid"]: item for item in report["matches"]}
        enriched = []
        for row in matches:
            fid = row["fid"]
            integrity_detail = integrity_by_fid.get(fid, check_module_integrity(fid, {}))
            enriched.append({
                "fid": fid,
                "match_no": row["match_no"],
                "home_team": row["home_team"],
                "away_team": row["away_team"],
                "match_id": match_ids.get(fid),
                "integrity": {
                    "ready": integrity_detail["ready"],
                    "status": integrity_detail["status"],
                    "missing_modules": integrity_detail["missing_modules"],
                },
                "fetch_trace": make_fetch_trace(
                    channel=self.channel,
                    blocks=[BLOCK_D1_MATCH_LIST, BLOCK_D2B_INTEGRITY],
                ),
            })
        report["matches"] = enriched
