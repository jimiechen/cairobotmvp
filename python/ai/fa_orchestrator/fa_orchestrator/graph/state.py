# -*- coding: utf-8 -*-
"""LangGraph 共享状态定义（PRD-12 §7：两阶段分析中间态在节点间流转）。

ISS-FA1 定义编排骨架字段；ISS-FA2 扩展两阶段分析中间态
（fundamentals/odds_verification/phase_conclusion）。
ISS-FA3 扩展质检层字段（sentiment/synthesis/review/条件回退计数）。
ISS-FA4 扩展追溯与人工裁决字段（trace_events/human_decisions/reports 等）。
trace/trace_events 使用 operator.add reducer 聚合。
"""
import operator
from typing import Annotated, TypedDict


class AnalysisState(TypedDict, total=False):
    """FA-Orchestrator 图状态容器。

    字段说明：
    - match_date：竞彩日（如 2026-09-29）
    - plan：PlannerAgent 产出的结构化分析计划（F-02）
    - matches：Q1 场次清单（D1 数据块）
    - integrity_report：日级数据就绪率报告（F-08）
    - ready_report：节点汇总后的最终就绪率报告
    - fundamentals：逐场第一阶段基本面结论（F-10，{fid: 结论}）
    - odds_verification：逐场第二阶段赔率验证结论（F-11，{fid: 结论}）
    - phase_conclusion：两阶段中间结论日级汇总（ISS-FA2 交付物）
    - sentiment：逐场热度结论（F-12，{fid: 结论}，ISS-FA3）
    - synthesis：逐场综合判定结论（F-13，{fid: 结论}，ISS-FA3）
    - review：逐场矛盾检查结果（F-14，{fid: 结果}，ISS-FA3）
    - review_decision：矛盾检查决策 pass/retry/manual（条件边路由键，ISS-FA3）
    - synthesis_attempts：SynthesisAgent 条件回退执行次数（有界：≤2，ISS-FA3）
    - manual_review：需人工裁决的场次清单（仍不过转人工，ISS-FA3）
    - trace：节点执行轨迹（reducer 聚合，供追溯/回放扩展）
    - trace_events：结构化轨迹事件（节点/耗时/输出/来源，F-18，ISS-FA4）
    - synthesis_attempts_by_match：逐场回退执行计数（{fid: 次数}，ISS-FA4）
    - vote_availability：逐场投票数据可用性（{fid: bool}，ISS-FA4）
    - history_deduction：逐场历史推演结论（{fid: 结论}，F-16，ISS-FA4）
    - human_decisions：人工裁决结论（{fid: 裁决值}，F-20，ISS-FA4）
    - reports：逐场 Markdown 报告（{fid: 报告文本}，F-15，ISS-FA4）
    """
    match_date: str
    plan: dict
    matches: list
    integrity_report: dict
    ready_report: dict
    fundamentals: dict
    odds_verification: dict
    phase_conclusion: dict
    sentiment: dict
    synthesis: dict
    review: dict
    review_decision: str
    synthesis_attempts: int
    manual_review: list
    trace: Annotated[list, operator.add]
    trace_events: Annotated[list, operator.add]
    synthesis_attempts_by_match: dict
    vote_availability: dict
    history_deduction: dict
    human_decisions: dict
    reports: dict
