# -*- coding: utf-8 -*-
"""StateGraph 图构建器（PRD-12 F-01）。

职责边界：
- 负责：按 OrchestrationConfig（YAML）把注册表中的 agent 函数装配为 LangGraph 图。
- 不负责：agent 实现（planner_agent.py / data_agent.py）、配置解析（config.py）。
"""
from langgraph.graph import END, START, StateGraph
from langgraph.graph.state import CompiledStateGraph

from fa_orchestrator.state import AnalysisState


class GraphBuildError(Exception):
    """图构建失败（注册表缺 agent 等）时抛出。"""


def _collect_terminal_nodes(config, outgoing: set) -> list:
    """收集无出边节点（需连接 END，保证图可终止）。"""
    return [node.name for node in config.nodes if node.name not in outgoing]


def build_graph(config, agent_registry) -> CompiledStateGraph:
    """按编排配置构建并编译 StateGraph。

    输入：OrchestrationConfig（YAML 加载产物）、agent_registry（agent 名 → 节点函数）。
    输出：编译后的 CompiledStateGraph（支持 invoke 与 Mermaid 导出）。
    失败行为：任一节点的 agent 未在注册表注册时抛 GraphBuildError。
    """
    builder = StateGraph(AnalysisState)
    for node in config.nodes:
        agent_fn = agent_registry.get(node.agent)
        if agent_fn is None:
            raise GraphBuildError(f"节点 {node.name} 的 agent 未注册: {node.agent}")
        builder.add_node(node.name, agent_fn)

    outgoing = set()
    for edge in config.edges:
        if edge.condition_key:
            builder.add_conditional_edges(
                edge.from_node, _make_condition_router(edge.condition_key, edge.branches)
            )
        else:
            builder.add_edge(edge.from_node, edge.to_node)
        outgoing.add(edge.from_node)

    builder.add_edge(START, config.entry_point)
    for terminal in _collect_terminal_nodes(config, outgoing):
        builder.add_edge(terminal, END)
    return builder.compile()


def _make_condition_router(condition_key: str, branches: dict):
    """构造条件边路由函数（F-14）：按 state 值路由，未知/缺失值兜底 manual 分支。"""

    def _route(state: AnalysisState) -> str:
        decision = state.get(condition_key) or "manual"
        return branches.get(decision, branches["manual"])

    return _route
