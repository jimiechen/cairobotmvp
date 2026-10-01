# -*- coding: utf-8 -*-
"""数据接入层：取数通道、SQL 契约、ID 解析与完整性核查。

职责边界：
- 负责：Fixture/MySQL 数据源、SQL 语句契约、match_id 解析、模块完整性核查。
- 不负责：数据加工分析（analysis/）、Agent 决策（agents/）、LLM 调用（infra/）。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
