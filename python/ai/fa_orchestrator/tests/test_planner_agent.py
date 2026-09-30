# -*- coding: utf-8 -*-
"""PlannerAgent Plan-and-Execute 测试（PRD-12 F-02）。

LLM 一律 Mock（fake gateway 返回预设文本），验证：
- 产出结构化分析计划（场次/并行分组/数据预检）
- LLM 输出非法 JSON / markdown 包裹 JSON 的容错
"""
import json

import pytest

from fa_orchestrator.planner_agent import PlannerAgent, PlannerError


class _FakeGateway:
    """假网关：complete() 返回预设文本，记录调用。"""

    def __init__(self, reply_text):
        self.reply_text = reply_text
        self.calls = []

    def complete(self, messages):
        self.calls.append(messages)
        from langchain_core.messages import AIMessage

        return AIMessage(content=self.reply_text)


_PLAN = {
    "match_date": "2026-09-29",
    "matches": [
        {"fid": "1398594", "match_no": "周二008", "group": 1},
        {"fid": "1398595", "match_no": "周二009", "group": 1},
    ],
    "parallel_groups": 1,
    "data_preflight": "Q2b 预检通过",
    "fetch_instructions": [],
}


class TestMakePlan:
    def test_产出结构化计划(self):
        gateway = _FakeGateway(json.dumps(_PLAN, ensure_ascii=False))
        plan = PlannerAgent(gateway).make_plan("2026-09-29")
        assert plan["match_date"] == "2026-09-29"
        assert len(plan["matches"]) == 2
        assert "parallel_groups" in plan

    def test_计划必填字段校验(self):
        broken = {"match_date": "2026-09-29"}  # 缺 matches/parallel_groups
        gateway = _FakeGateway(json.dumps(broken, ensure_ascii=False))
        with pytest.raises(PlannerError):
            PlannerAgent(gateway).make_plan("2026-09-29")

    def test_markdown包裹的JSON可解析(self):
        wrapped = "```json\n" + json.dumps(_PLAN, ensure_ascii=False) + "\n```"
        gateway = _FakeGateway(wrapped)
        plan = PlannerAgent(gateway).make_plan("2026-09-29")
        assert plan["match_date"] == "2026-09-29"

    def test_非法JSON报错(self):
        gateway = _FakeGateway("这不是JSON")
        with pytest.raises(PlannerError):
            PlannerAgent(gateway).make_plan("2026-09-29")

    def test_提示词包含竞彩日(self):
        gateway = _FakeGateway(json.dumps(_PLAN, ensure_ascii=False))
        PlannerAgent(gateway).make_plan("2026-09-29")
        prompt_text = json.dumps(gateway.calls[0], ensure_ascii=False)
        assert "2026-09-29" in prompt_text
