# -*- coding: utf-8 -*-
"""分析引擎层：逐场分析流水线、赔率算法与赛后复盘对照。

职责边界：
- 负责：AnalysisRunner 逐场分析引擎、odds_math 赔率纯算法、history_deduction
  历史走势推演、match_review 赛后复盘对照（ISS-FA5）。
- 不负责：图编排（graph/）、Agent 决策实现（agents/）、数据源（data/）。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
