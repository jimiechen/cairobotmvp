# -*- coding: utf-8 -*-
"""YAML 编排配置加载测试（PRD-12 F-01：图拓扑配置化，核心代码不硬编码图结构）。"""
from pathlib import Path

import pytest

from fa_orchestrator.config import ConfigError, load_orchestration_config

_CONFIG_PATH = (
    Path(__file__).resolve().parent.parent / "fa_orchestrator" / "configs" / "orchestrator.yaml"
)


class TestLoadOrchestrationConfig:
    def test_加载默认编排配置(self):
        config = load_orchestration_config(_CONFIG_PATH)
        assert config.entry_point == "planner"
        node_names = [n.name for n in config.nodes]
        assert node_names == ["planner", "data_agent", "ready_report"]

    def test_加载边定义(self):
        config = load_orchestration_config(_CONFIG_PATH)
        edge_pairs = [(e.from_node, e.to_node) for e in config.edges]
        assert ("planner", "data_agent") in edge_pairs
        assert ("data_agent", "ready_report") in edge_pairs

    def test_文件不存在报错(self, tmp_path):
        with pytest.raises(ConfigError):
            load_orchestration_config(tmp_path / "not_exist.yaml")

    def test_缺entry_point报错(self, tmp_path):
        bad = tmp_path / "bad.yaml"
        bad.write_text(
            "graph:\n  nodes:\n    - name: a\n      agent: a\n"
            "  edges: []\n",
            encoding="utf-8",
        )
        with pytest.raises(ConfigError):
            load_orchestration_config(bad)

    def test_边引用未定义节点报错(self, tmp_path):
        bad = tmp_path / "bad_edge.yaml"
        bad.write_text(
            "graph:\n  entry_point: a\n"
            "  nodes:\n    - name: a\n      agent: a\n"
            "  edges:\n    - from: a\n      to: ghost\n",
            encoding="utf-8",
        )
        with pytest.raises(ConfigError):
            load_orchestration_config(bad)

    def test_入口必须是已定义节点(self, tmp_path):
        bad = tmp_path / "bad_entry.yaml"
        bad.write_text(
            "graph:\n  entry_point: ghost\n"
            "  nodes:\n    - name: a\n      agent: a\n  edges: []\n",
            encoding="utf-8",
        )
        with pytest.raises(ConfigError):
            load_orchestration_config(bad)
