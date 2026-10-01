# -*- coding: utf-8 -*-
"""分析决策 Agent 层：图节点的各领域 Agent 实现。

职责边界：
- 负责：Planner/Data/Fundamental/OddsVerify/Sentiment/Synthesis/Review 七类
  分析决策 Agent。
- 不负责：报告组装与渲染（report/）、数据源实现（data/）、图拓扑（graph/）。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
