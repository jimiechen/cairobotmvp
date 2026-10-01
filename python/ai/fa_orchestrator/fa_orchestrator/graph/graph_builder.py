# -*- coding: utf-8 -*-
"""StateGraph 图构建器（PRD-12 F-01/F-17）。

职责边界：
- 负责：按 OrchestrationConfig（YAML）把注册表中的 agent 函数装配为 LangGraph 图；
  checkpointer 透传（F-19 断点续跑）；Mermaid 导出与编排配置一致性断言（F-17/验收 6）。
- 不负责：agent 实现（graph_nodes.py）、配置解析（config.py）、checkpoint 存储
  （sqlite_checkpointer.py）。
"""
from langgraph.graph import END, START, StateGraph
from langgraph.graph.state import CompiledStateGraph

from fa_orchestrator.graph.state import AnalysisState


class GraphBuildError(Exception):
    """图构建失败（注册表缺 agent 等）时抛出。"""


def _collect_terminal_nodes(config, outgoing: set) -> list:
    """收集无出边节点（需连接 END，保证图可终止）。"""
    return [node.name for node in config.nodes if node.name not in outgoing]


def build_graph(config, agent_registry, checkpointer=None) -> CompiledStateGraph:
    """按编排配置构建并编译 StateGraph。

    输入：OrchestrationConfig（YAML 加载产物）、agent_registry（agent 名 → 节点函数）、
    checkpointer（可选，F-19；传入后 invoke 需带 thread_id 并支持 interrupt/resume）。
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
            # path_map 让 langgraph 静态渲染条件边分支（draw_mermaid 一致性断言依赖）
            builder.add_conditional_edges(
                edge.from_node, _make_condition_router(edge.condition_key, edge.branches),
                path_map=dict(edge.branches),
            )
        else:
            builder.add_edge(edge.from_node, edge.to_node)
        outgoing.add(edge.from_node)

    builder.add_edge(START, config.entry_point)
    for terminal in _collect_terminal_nodes(config, outgoing):
        builder.add_edge(terminal, END)
    if checkpointer is not None:
        return builder.compile(checkpointer=checkpointer)
    return builder.compile()


def _make_condition_router(condition_key: str, branches: dict):
    """构造条件边路由函数（F-14）。

    返回分支值（decision），节点名映射由 path_map（branches）承担；
    未知/缺失值兜底 "manual" 分支（有界终止，人工裁决节点处理）。
    """

    def _route(state: AnalysisState) -> str:
        decision = state.get(condition_key) or "manual"
        return decision if decision in branches else "manual"

    return _route


def export_mermaid(compiled) -> str:
    """导出运行时 StateGraph 的 Mermaid 源码（F-17）。

    与 draw_mermaid 的差异：draw_mermaid 不渲染条件边分支，无法支撑
    验收 6 的一致性断言；本函数从 get_graph().edges 忠实渲染全部节点、
    普通边与条件边（携带分支 label，兼作 F-18 条件走向展示）。
    输入：CompiledStateGraph。输出：Mermaid 文本。
    """
    runtime = compiled.get_graph()
    lines = ["graph TD"]
    for name in runtime.nodes:
        lines.append(f"    {name}({name})")
    for edge in runtime.edges:
        if getattr(edge, "conditional", False):
            lines.append(f"    {edge.source} -->|{edge.data}| {edge.target}")
        else:
            lines.append(f"    {edge.source} --> {edge.target}")
    return "\n".join(lines)


def mermaid_matches_config(mermaid: str, config) -> bool:
    """Mermaid 导出与编排配置一致性断言（F-17/验收 6）。

    判定口径：导出串必须包含配置定义的全部节点与全部边
    （普通边 from --> to；条件边逐分支核对 from -->|decision| target）。
    输出：True=一致；False=存在节点或边缺失。
    """
    for node in config.nodes:
        if f"{node.name}({node.name})" not in mermaid:
            return False
    for edge in config.edges:
        if edge.condition_key:
            arrows = [
                f"{edge.from_node} -->|{decision}| {target}"
                for decision, target in edge.branches.items()
            ]
        else:
            arrows = [f"{edge.from_node} --> {edge.to_node}"]
        if any(arrow not in mermaid for arrow in arrows):
            return False
    return True
