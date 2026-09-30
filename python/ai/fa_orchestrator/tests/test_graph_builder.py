# -*- coding: utf-8 -*-
"""StateGraph 图构建测试（PRD-12 F-01）：按 YAML 配置构建图拓扑，支持 Mermaid 导出。

用 fake agent（仅往 state["trace"] 追加节点名）验证节点流转顺序，
不引入任何 LLM 与外部服务。
"""
import pytest
from langchain_core.runnables import RunnableConfig
from langgraph.graph.state import CompiledStateGraph

from fa_orchestrator.config import load_orchestration_config
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
    return {
        "planner_agent": _make_trace_agent("planner"),
        "data_agent": _make_trace_agent("data_agent"),
        "ready_report": _make_trace_agent("ready_report"),
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

    def test_节点按拓扑顺序流转(self, registry):
        config = load_orchestration_config(_CONFIG_PATH)
        graph = build_graph(config, registry)
        result = graph.invoke(
            {"match_date": "2026-09-29", "trace": []}, config=RunnableConfig()
        )
        # planner → data_agent → ready_report 顺序执行
        assert result["trace"] == ["planner", "data_agent", "ready_report"]
