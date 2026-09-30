# -*- coding: utf-8 -*-
"""FA-Orchestrator：足球情报分析多 Agent 可视化编排系统。

职责边界：
- 负责：LangGraph StateGraph 编排内核、PlannerAgent、DataAgent（Q1/Q2b 取数与完整性核查）、
  LLM 统一网关、CLI 入口。
- 不负责：分析类 Agent（基本面/赔率/热度/综合判定，ISS-FA2/FA3）、
  可视化与 checkpointer（ISS-FA4）、Skill 本身的行为规范。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""

__version__ = "0.1.0"
