# -*- coding: utf-8 -*-
"""报告输出层：Markdown 报告组装与 Standalone HTML 渲染。

职责边界：
- 负责：ReportAgent 按 OUTPUT_TEMPLATE.md 组装报告（溯源声明 + 异常告警）、
  html_report 渲染自包含 HTML。
- 不负责：分析决策（agents/）、赛果对照（analysis/match_review.py）、CLI 编排。

相关文档：
- PRD-12-足球情报分析多Agent可视化编排系统
- ADR-017-FA情报分析采用LangGraph多Agent编排
"""
