# -*- coding: utf-8 -*-
"""YAML 编排配置加载（PRD-12 F-01：图拓扑配置化，核心代码不硬编码图结构）。

职责边界：
- 负责：解析与校验 orchestrator.yaml（节点/边/入口），产出强类型配置对象。
- 不负责：图构建（graph_builder.py）、节点函数注册与执行。
"""
from dataclasses import dataclass, field
from pathlib import Path

import yaml


class ConfigError(Exception):
    """编排配置缺失或校验不通过时抛出。"""


@dataclass
class NodeSpec:
    """图节点定义：name 为图内节点名，agent 为 agent_registry 注册键。"""
    name: str
    agent: str


@dataclass
class EdgeSpec:
    """图边定义：条件分支（condition）由后续 Issue 扩展。"""
    from_node: str
    to_node: str


@dataclass
class OrchestrationConfig:
    """编排配置对象：入口节点、节点清单、边清单。"""
    entry_point: str
    nodes: list = field(default_factory=list)
    edges: list = field(default_factory=list)


def load_orchestration_config(path) -> OrchestrationConfig:
    """加载并校验 YAML 编排配置。

    输入：YAML 配置文件路径。
    输出：OrchestrationConfig。
    失败行为：文件缺失、缺 graph/entry_point、入口未定义、边引用未定义节点均抛 ConfigError。
    """
    config_path = Path(path)
    if not config_path.exists():
        raise ConfigError(f"编排配置不存在: {config_path}")

    raw = yaml.safe_load(config_path.read_text(encoding="utf-8")) or {}
    graph = raw.get("graph") or {}
    entry_point = graph.get("entry_point")
    if not entry_point:
        raise ConfigError("编排配置缺少 graph.entry_point")

    nodes = [
        NodeSpec(name=item["name"], agent=item["agent"])
        for item in graph.get("nodes") or []
    ]
    node_names = {node.name for node in nodes}
    if entry_point not in node_names:
        raise ConfigError(f"entry_point 未在节点中定义: {entry_point}")

    edges = []
    for item in graph.get("edges") or []:
        for endpoint in (item.get("from"), item.get("to")):
            if endpoint not in node_names:
                raise ConfigError(f"边引用了未定义节点: {endpoint}")
        edges.append(EdgeSpec(from_node=item["from"], to_node=item["to"]))

    return OrchestrationConfig(entry_point=entry_point, nodes=nodes, edges=edges)
