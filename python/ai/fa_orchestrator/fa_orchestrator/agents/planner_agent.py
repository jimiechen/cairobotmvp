# -*- coding: utf-8 -*-
"""PlannerAgent：Plan-and-Execute 产出结构化分析计划（PRD-12 F-02）。

职责边界：
- 负责：构建规划提示词、调用 LLM 网关、解析并校验结构化计划 JSON。
- 不负责：计划执行（data_agent.py）、LLM 协议细节（llm_gateway.py）。

agent-plan 语义（ADR-017 §3.3）：计划-执行模式，Re-Plan 由后续 Issue 扩展。
LLM 一律经网关注入，测试时注入 fake gateway，禁止真实调用。
"""
import json
import re

from fa_orchestrator.infra.llm_gateway import LlmGateway

# 计划必填字段（缺失即视为 LLM 输出不合规）
PLAN_REQUIRED_FIELDS = ("match_date", "matches", "parallel_groups")

# markdown 代码块包裹的 JSON（LLM 常见输出形态）
_MARKDOWN_JSON_PATTERN = re.compile(r"```(?:json)?\s*(.*?)\s*```", re.DOTALL)

PLANNER_SYSTEM_PROMPT = (
    "你是足球情报分析系统的规划器。给定竞彩日，产出 JSON 分析计划，"
    "必含字段：match_date（竞彩日）、matches（场次列表，含 fid/match_no/group）、"
    "parallel_groups（并行分组数）、data_preflight（数据预检结论）、"
    "fetch_instructions（补抓指令列表）。只输出 JSON，不要输出其他内容。"
)


class PlannerError(Exception):
    """LLM 输出不合规（非法 JSON / 缺必填字段）时抛出。"""


class PlannerAgent:
    """规划 Agent：输入竞彩日，输出结构化分析计划。

    状态约束：无状态；依赖注入 LlmGateway（测试注入 fake）。
    """

    def __init__(self, gateway: LlmGateway):
        self.gateway = gateway

    def make_plan(self, match_date: str) -> dict:
        """产出结构化分析计划。

        输入：竞彩日（如 2026-09-29）。
        输出：计划 dict（含 match_date/matches/parallel_groups 等必填字段）。
        失败行为：LLM 返回非法 JSON 或缺必填字段抛 PlannerError。
        """
        reply = self.gateway.complete([
            {"role": "system", "content": PLANNER_SYSTEM_PROMPT},
            {"role": "user", "content": f"竞彩日：{match_date}，请生成分析计划。"},
        ])
        plan = self._parse_plan_json(reply.content)
        self._validate_plan(plan, match_date)
        return plan

    def _parse_plan_json(self, reply_text: str) -> dict:
        """解析 LLM 回复中的计划 JSON，兼容 markdown 代码块包裹。"""
        text = reply_text.strip()
        markdown_match = _MARKDOWN_JSON_PATTERN.search(text)
        if markdown_match:
            text = markdown_match.group(1)
        try:
            plan = json.loads(text)
        except json.JSONDecodeError as error:
            raise PlannerError(f"计划 JSON 解析失败: {error}") from error
        if not isinstance(plan, dict):
            raise PlannerError("计划必须是 JSON 对象")
        return plan

    def _validate_plan(self, plan: dict, match_date: str) -> None:
        """校验计划必填字段与竞彩日一致性。"""
        missing = [field_name for field_name in PLAN_REQUIRED_FIELDS if field_name not in plan]
        if missing:
            raise PlannerError(f"计划缺少必填字段: {', '.join(missing)}")
        if plan.get("match_date") != match_date:
            raise PlannerError(
                f"计划竞彩日不一致: 期望 {match_date}, 实际 {plan.get('match_date')}"
            )
