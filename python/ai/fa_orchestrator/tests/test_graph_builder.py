# -*- coding: utf-8 -*-
"""StateGraph 图构建测试（PRD-12 F-01）：按 YAML 配置构建图拓扑，支持 Mermaid 导出。

用 fake agent（仅往 state["trace"] 追加节点名）验证节点流转顺序，
不引入任何 LLM 与外部服务。
"""
import pytest
from langchain_core.runnables import RunnableConfig
from langgraph.graph.state import CompiledStateGraph

from fa_orchestrator.config import ConfigError, load_orchestration_config
from fa_orchestrator.graph_builder import GraphBuildError, build_graph
from fa_orchestrator.state import AnalysisState

from pathlib import Path

_CONFIG_PATH = (
    Path(__file__).resolve().parent.parent / "fa_orchestrator" / "configs" / "orchestrator.yaml"
)


def _make_trace_agent(node_name: str):
    """构造记录执行顺序的 fake agent 节点函数。"""

    def _node(state: AnalysisState) -> dict:
        return {"trace": [node_name]}

    _node.__name__ = node_name
    return _node


@pytest.fixture()
def registry():
    def _pass_review_agent(state):
        # 拓扑流转测试固定走 pass 分支（manual 分支由 TestConditionalEdge 覆盖）
        return {"trace": ["review_agent"], "review_decision": "pass"}

    _pass_review_agent.__name__ = "review_agent"
    return {
        "planner_agent": _make_trace_agent("planner"),
        "data_agent": _make_trace_agent("data_agent"),
        "ready_report": _make_trace_agent("ready_report"),
        "fundamental_agent": _make_trace_agent("fundamental_agent"),
        "odds_verify_agent": _make_trace_agent("odds_verify_agent"),
        "sentiment_agent": _make_trace_agent("sentiment_agent"),
        "synthesis_agent": _make_trace_agent("synthesis_agent"),
        "review_agent": _pass_review_agent,
        "human_review": _make_trace_agent("human_review"),
        "report_agent": _make_trace_agent("report_agent"),
        "phase_report": _make_trace_agent("phase_report"),
    }


class TestBuildGraph:
    def test_返回编译后的StateGraph(self, registry):
        config = load_orchestration_config(_CONFIG_PATH)
        graph = build_graph(config, registry)
        assert isinstance(graph, CompiledStateGraph)

    def test_注册agent缺失报错(self):
        config = load_orchestration_config(_CONFIG_PATH)
        with pytest.raises(GraphBuildError):
            build_graph(config, {})

    def test_Mermaid导出包含全部节点(self, registry):
        config = load_orchestration_config(_CONFIG_PATH)
        graph = build_graph(config, registry)
        mermaid = graph.get_graph().draw_mermaid()
        assert "planner" in mermaid
        assert "data_agent" in mermaid
        assert "ready_report" in mermaid
        assert "fundamental_agent" in mermaid
        assert "odds_verify_agent" in mermaid
        assert "sentiment_agent" in mermaid
        assert "synthesis_agent" in mermaid
        assert "review_agent" in mermaid
        assert "human_review" in mermaid
        assert "report_agent" in mermaid
        assert "phase_report" in mermaid

    def test_节点按拓扑顺序流转(self, registry):
        config = load_orchestration_config(_CONFIG_PATH)
        graph = build_graph(config, registry)
        result = graph.invoke(
            {"match_date": "2026-09-29", "trace": []}, config=RunnableConfig()
        )
        # planner → data_agent → ready_report → fundamental_agent → odds_verify_agent
        # → sentiment_agent → synthesis_agent → review_agent（pass）
        # → report_agent → phase_report（F-12/F-13/F-14/F-15）
        assert result["trace"] == [
            "planner", "data_agent", "ready_report",
            "fundamental_agent", "odds_verify_agent",
            "sentiment_agent", "synthesis_agent", "review_agent",
            "report_agent", "phase_report",
        ]

    def test_checkpointer参数传递给编译图(self, registry, tmp_path):
        from fa_orchestrator.sqlite_checkpointer import SQLiteCheckpointer
        config = load_orchestration_config(_CONFIG_PATH)
        graph = build_graph(
            config, registry,
            checkpointer=SQLiteCheckpointer(str(tmp_path / "ckpt.db")),
        )
        result = graph.invoke(
            {"match_date": "2026-09-29", "trace": []},
            config={"configurable": {"thread_id": "t-ckpt"}},
        )
        assert result["trace"][-1] == "phase_report"


class TestConditionalEdge:
    """矛盾检查条件回退（F-14）：review 决策经条件边路由，纪律条款图级强制。"""

    def _review_agent(self, decisions):
        """构造按脚本返回决策的 fake review 节点。"""

        def _node(state: AnalysisState) -> dict:
            decision = decisions.pop(0) if decisions else "pass"
            return {"trace": ["review_agent"], "review_decision": decision}

        _node.__name__ = "review_agent"
        return _node

    def _invoke(self, decisions, extra_registry=None):
        config = load_orchestration_config(_CONFIG_PATH)
        registry = {
            "planner_agent": _make_trace_agent("planner"),
            "data_agent": _make_trace_agent("data_agent"),
            "ready_report": _make_trace_agent("ready_report"),
            "fundamental_agent": _make_trace_agent("fundamental_agent"),
            "odds_verify_agent": _make_trace_agent("odds_verify_agent"),
            "sentiment_agent": _make_trace_agent("sentiment_agent"),
            "synthesis_agent": _make_trace_agent("synthesis_agent"),
            "review_agent": self._review_agent(decisions),
            "human_review": _make_trace_agent("human_review"),
            "report_agent": _make_trace_agent("report_agent"),
            "phase_report": _make_trace_agent("phase_report"),
        }
        registry.update(extra_registry or {})
        graph = build_graph(config, registry)
        return graph.invoke({"match_date": "2026-09-29", "trace": []}, config=RunnableConfig())

    def test_重试决策路由回synthesis_形成回退环(self):
        result = self._invoke(["retry", "pass"])
        assert result["trace"].count("synthesis_agent") == 2
        assert result["trace"].count("review_agent") == 2
        assert result["trace"][-1] == "phase_report"

    def test_通过决策直达报告节点不再回退(self):
        result = self._invoke(["pass"])
        assert result["trace"].count("synthesis_agent") == 1
        assert result["trace"][-1] == "phase_report"

    def test_人工决策经human_review直达报告节点(self):
        result = self._invoke(["manual"])
        assert result["trace"].count("synthesis_agent") == 1
        # ISS-FA4：manual 分支路由到人工裁决节点（F-20）再到报告
        assert "human_review" in result["trace"]
        assert result["trace"][-1] == "phase_report"

    def test_未知决策值兜底转人工路由_保证有界终止(self):
        result = self._invoke(["boom"])
        assert result["trace"].count("synthesis_agent") == 1
        assert "human_review" in result["trace"]
        assert result["trace"][-1] == "phase_report"

    def test_条件边缺少manual兜底分支报错(self, tmp_path):
        config_path = tmp_path / "bad.yaml"
        config_path.write_text(
            "graph:\n"
            "  entry_point: review_agent\n"
            "  nodes:\n"
            "    - name: review_agent\n"
            "      agent: review_agent\n"
            "    - name: synthesis_agent\n"
            "      agent: synthesis_agent\n"
            "    - name: phase_report\n"
            "      agent: phase_report\n"
            "  edges:\n"
            "    - from: review_agent\n"
            "      to: phase_report\n"
            "      condition:\n"
            "        key: review_decision\n"
            "        branches: {retry: synthesis_agent}\n",
            encoding="utf-8",
        )
        with pytest.raises(ConfigError):
            load_orchestration_config(config_path)
