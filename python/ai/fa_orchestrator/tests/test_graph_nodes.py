# -*- coding: utf-8 -*-
"""图驱动注册表测试（PRD-12 F-18/F-19/F-20，ISS-FA4）。

GraphNodes 将 planner/data/四类分析 Agent/ReviewAgent/human_review/ReportAgent
包装为 StateGraph 节点函数；全程 fixture 数据源，禁止真实 MySQL/HTTP/LLM。
验证：节点注册完整性、人工裁决 interrupt/resume（F-20）、条件回退环图级生效（F-14）、
结构化轨迹（F-18）、图驱动与 runner 版结论对齐。
"""
from pathlib import Path

from langgraph.types import Command

from fa_orchestrator.analysis.analysis_runner import AnalysisRunner
from fa_orchestrator.graph.config import load_orchestration_config
from fa_orchestrator.data.data_source import FixtureDataSource
from fa_orchestrator.graph.graph_builder import build_graph
from fa_orchestrator.graph.graph_nodes import GraphNodes
from fa_orchestrator.infra.sqlite_checkpointer import SQLiteCheckpointer

_PKG = Path(__file__).resolve().parent.parent / "fa_orchestrator"
_YAML = _PKG / "configs" / "orchestrator.yaml"
_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929.json"
_ANALYSIS = Path(__file__).resolve().parent / "fixtures" / "sample_day_20260929_analysis.json"
_DATE = "2026-09-29"
# fixture 样本日：1398597 两次回退仍违规 → manual（FA3 产物结论），其余 pass
_MANUAL_FID = "1398597"


def _invoke_config(thread_id: str = _DATE) -> dict:
    return {"configurable": {"thread_id": thread_id}}


def _build(tmp_path):
    """构建带 SQLite checkpointer 的图驱动分析管线。"""
    source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
    config = load_orchestration_config(_YAML)
    graph = build_graph(
        config, GraphNodes(source).registry(),
        checkpointer=SQLiteCheckpointer(str(tmp_path / "ckpt.db")),
    )
    return graph


def _initial_state() -> dict:
    return {"match_date": _DATE, "trace": [], "trace_events": []}


class TestGraphNodesRegistry:
    def test_registry覆盖编排配置全部agent键(self):
        source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
        config = load_orchestration_config(_YAML)
        registry = GraphNodes(source).registry()
        expected = {node.agent for node in config.nodes}
        assert set(registry) == expected

    def test_无checkpointer可构建图(self):
        source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
        config = load_orchestration_config(_YAML)
        graph = build_graph(config, GraphNodes(source).registry())
        assert graph is not None


class TestHumanInTheLoop:
    def test_图驱动首跑在人工裁决处中断(self, tmp_path):
        graph = _build(tmp_path)
        result = graph.invoke(_initial_state(), _invoke_config())
        interrupts = result.get("__interrupt__") or []
        assert interrupts, "manual 场应触发 human-in-the-loop interrupt（F-20）"
        payload = interrupts[0].value
        assert payload["matches"][0]["fid"] == _MANUAL_FID
        assert payload["matches"][0]["violations"], "裁决提示必须携带矛盾违规项"

    def test_resume后产出日级结论与人工裁决记录(self, tmp_path):
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(
            Command(resume={"approved": True, "note": "分析师复核确认"}),
            _invoke_config(),
        )
        conclusion = result["phase_conclusion"]
        assert conclusion["total_matches"] == 7
        assert conclusion["manual_review_matches"] == [_MANUAL_FID]
        assert result["human_decisions"][_MANUAL_FID]["note"] == "分析师复核确认"

    def test_manual场登记人工裁决清单于中断前(self, tmp_path):
        graph = _build(tmp_path)
        result = graph.invoke(_initial_state(), _invoke_config())
        # 中断发生在 human_review 节点：manual_review 已由 review 节点登记
        assert result["manual_review"] == [_MANUAL_FID]

    def test_retry回退环在图级生效(self, tmp_path):
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        # 1398597 两次回退后转 manual：synthesis/review 各执行 3 次（attempt 0/1/2）
        trace = result["trace"]
        assert trace.count("synthesis_agent") == 3
        assert trace.count("review_agent") == 3


class TestTraceAndAlignment:
    def test_trace_events记录节点耗时与数据来源(self, tmp_path):
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        events = result["trace_events"]
        by_node = {}
        for event in events:
            assert event["status"] == "ok"
            assert event["duration_ms"] >= 0
            by_node.setdefault(event["node"], []).append(event)
        # 数据来源（F-18）：分析类节点必须标注 channel
        assert by_node["fundamental_agent"][0]["source"] == "fixture"
        assert by_node["data_agent"][0]["source"] == "fixture"

    def test_图驱动结论与runner版对齐(self, tmp_path):
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
        runner_matches = AnalysisRunner(source).run(_DATE)["matches"]
        by_fid = {m["fid"]: m for m in runner_matches}
        graph_matches = {m["fid"]: m for m in result["phase_conclusion"]["matches"]}
        # 共享同一批 agent 与输入组装函数：推荐结构与矛盾决策必须逐场一致
        for fid in ("1398594", "1398597"):
            assert graph_matches[fid]["recommendation"] == by_fid[fid]["recommendation"]
            assert graph_matches[fid]["review"]["decision"] == by_fid[fid]["review"]["decision"]

    def test_图驱动矛盾检查明细与runner逐项对齐(self, tmp_path):
        """回归（立项三遗留差异根因）：AnalysisState 未声明 goal_layer/
        greed_fear 时，LangGraph 按 TypedDict schema 过滤节点输出，
        图通道 goal_layer 恒缺失 → 检查①③④恒 na，与 runner 实检不一致。
        修复后两通道检查明细必须逐项一致。"""
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        source = FixtureDataSource(_FIXTURE, analysis_path=_ANALYSIS)
        runner_matches = AnalysisRunner(source).run(_DATE)["matches"]
        by_fid = {m["fid"]: m for m in runner_matches}
        graph_matches = {m["fid"]: m for m in result["phase_conclusion"]["matches"]}
        for fid in ("1398594", "1398597"):
            runner_checks = {c["id"]: c["result"] for c in by_fid[fid]["review"]["checks"]}
            graph_checks = {c["id"]: c["result"] for c in graph_matches[fid]["review"]["checks"]}
            assert runner_checks == graph_checks, f"场 {fid} 矛盾检查明细两通道不一致"

    def test_比分池层与贪婪恐惧结论进入图状态(self, tmp_path):
        """AnalysisState 必须声明 goal_layer/greed_fear 字段，保证立项一/三
        层结论在图通道不被 schema 过滤而进入综合判定与报告组装。"""
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        matches = result["phase_conclusion"]["matches"]
        assert matches, "应有场次结论"
        for m in matches:
            assert (m.get("goal_layer") or {}).get("phase") == "goal_layer", \
                f"场 {m['fid']} 比分池层结论缺失（被 state schema 过滤）"
            assert (m.get("greed_fear") or {}).get("phase") == "greed_fear", \
                f"场 {m['fid']} 贪婪恐惧结论缺失（被 state schema 过滤）"

    def test_报告产物含溯源声明与矛盾检查章节(self, tmp_path):
        graph = _build(tmp_path)
        graph.invoke(_initial_state(), _invoke_config())
        result = graph.invoke(Command(resume={"approved": True}), _invoke_config())
        reports = result["reports"]
        assert set(reports) == {m["fid"] for m in result["phase_conclusion"]["matches"]}
        report = reports["1398594"]
        assert "数据来源声明" in report
        assert "fixture" in report
        assert "矛盾检查" in report
