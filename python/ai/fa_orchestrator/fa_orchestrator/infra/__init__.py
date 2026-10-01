# -*- coding: utf-8 -*-
"""基础设施层：LLM 统一网关与 LangGraph 持久化。

职责边界：
- 负责：LlmGateway 统一网关（超时/重试/脱敏）、SQLiteCheckpointer 断点持久化。
- 不负责：业务决策（agents/）、图拓扑（graph/）、数据接入（data/）。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
