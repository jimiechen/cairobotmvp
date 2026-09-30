# -*- coding: utf-8 -*-
"""SQLite checkpointer 断点续跑集成测试（PRD-12 F-19/F-20，ISS-FA4）。

使用真实 orchestrator.yaml 拓扑 + fake 节点验证 LangGraph 持久化机制：
进程级中断（节点抛错）后从最近 checkpoint 续跑；人工裁决 interrupt 后
Command(resume) 恢复；轨迹可从 checkpoint 读回（replay 数据源）。
SQLite 为本地文件存储（非外部服务），无需门控变量。
"""
from pathlib import Path

import pytest
from langgraph.types import Command, interrupt

from fa_orchestrator.config import load_orchestration_config
from fa_orchestrator.graph_builder import build_graph
from fa_orchestrator.sqlite_checkpointer import SQLiteCheckpointer

_PKG = Path(__file__).resolve().parent.parent.parent / "fa_orchestrator"
_YAML = _PKG / "configs" / "orchestrator.yaml"
_THREAD = {"configurable": {"thread_id": "t-integration"}}


def _trace_agent(name: str):
    def _node(state: dict) -> dict:
        # 与 GraphNodes._traced 同契约：trace + trace_events 双通道（F-18 回放数据源）
        return {
            "trace": [name],
            "trace_events": [{
                "node": name, "duration_ms": 0, "status": "ok",
                "output": {}, "source": "fake",
            }],
        }
    _node.__name__ = name
    return _node


def _registry(review_decisions: list, human_review_fn) -> dict:
    """构造覆盖全部 yaml agent 键的 fake 注册表。"""
    registry = {
        "planner_agent": _trace_agent("planner"),
        "data_agent": _trace_agent("data_agent"),
        "ready_report": _trace_agent("ready_report"),
        "fundamental_agent": _trace_agent("fundamental_agent"),
        "odds_verify_agent": _trace_agent("odds_verify_agent"),
        "sentiment_agent": _trace_agent("sentiment_agent"),
        "synthesis_agent": _trace_agent("synthesis_agent"),
        "report_agent": _trace_agent("report_agent"),
        "phase_report": _trace_agent("phase_report"),
        "review_agent": _review_agent(review_decisions),
        "human_review": human_review_fn,
    }
    return registry


def _review_agent(decisions: list):
    def _node(state: dict) -> dict:
        decision = decisions.pop(0) if decisions else "pass"
        update: dict = {"trace": ["review_agent"], "review_decision": decision}
        if decision == "manual":
            update["manual_review"] = ["1398597"]
        return update
    _node.__name__ = "review_agent"
    return _node


def _interrupt_human_review(state: dict) -> dict:
    """真实 interrupt 语义的人工裁决节点（与 graph_nodes 实现同构）。"""
    fids = state.get("manual_review") or []
    if not fids:
        return {"trace": ["human_review"]}
    answer = interrupt({"reason": "矛盾检查未通过", "matches": fids})
    return {"trace": ["human_review"],
            "human_decisions": {fid: answer for fid in fids}}


def _build(tmp_path, registry) -> object:
    config = load_orchestration_config(_YAML)
    return build_graph(
        config, registry,
        checkpointer=SQLiteCheckpointer(str(tmp_path / "ckpt.db")),
    )


class TestCrashResume:
    def test_节点抛错后从最近checkpoint续跑完成(self, tmp_path):
        """节点级中断（模拟进程崩溃）：同一 thread 再次 invoke(None) 续跑至完成。"""
        calls = {"fundamental_agent": 0}

        def _flaky(state: dict) -> dict:
            calls["fundamental_agent"] += 1
            if calls["fundamental_agent"] == 1:
                raise RuntimeError("模拟进程中断")
            return {"trace": ["fundamental_agent"]}

        registry = _registry(["pass"], _interrupt_human_review)
        registry["fundamental_agent"] = _flaky
        graph = _build(tmp_path, registry)
        with pytest.raises(RuntimeError):
            graph.invoke({"match_date": "2026-09-29"}, _THREAD)
        result = graph.invoke(None, _THREAD)
        # 续跑后走完全图：fundamental_agent 重试成功，trace 无重复补齐缺口
        assert result["trace"][-1] == "phase_report"
        assert calls["fundamental_agent"] == 2


class TestHumanInterruptResume:
    def test_manual决策中断后续跑完成(self, tmp_path):
        graph = _build(tmp_path, _registry(["manual"], _interrupt_human_review))
        result = graph.invoke({"match_date": "2026-09-29"}, _THREAD)
        assert result.get("__interrupt__"), "manual 场应中断等待人工裁决"
        result = graph.invoke(
            Command(resume={"approved": True, "note": "确认"}), _THREAD
        )
        assert result["trace"][-1] == "phase_report"
        assert result["human_decisions"]["1398597"]["note"] == "确认"

    def test_轨迹可从checkpoint读回(self, tmp_path):
        graph = _build(tmp_path, _registry(["pass"], _interrupt_human_review))
        result = graph.invoke({"match_date": "2026-09-29"}, _THREAD)
        checkpointer = SQLiteCheckpointer(str(tmp_path / "ckpt.db"))
        tp = checkpointer.get_tuple(_THREAD)
        channel_values = tp.checkpoint["channel_values"]
        # replay 数据源：最新 checkpoint 必须携带完整 trace 与 trace_events
        assert channel_values["trace"] == result["trace"]
        assert isinstance(channel_values["trace_events"], list)
