# -*- coding: utf-8 -*-
"""图编排内核层：LangGraph StateGraph 的状态、拓扑配置、图构建与节点函数。

职责边界：
- 负责：AnalysisState 定义、orchestrator.yaml 拓扑配置加载（F-01）、build_graph、
  GraphNodes 节点函数注册。
- 不负责：具体分析逻辑（agents/）、数据接入（data/）、CLI 交互（包根 cli.py）。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
