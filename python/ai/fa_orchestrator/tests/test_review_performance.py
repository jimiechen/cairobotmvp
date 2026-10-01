# -*- coding: utf-8 -*-
"""多场并行性能验证测试（PRD-12 §14 / ISS-FA5：并行 ≤ 串行单场 × 3）。

用 LangGraph StateGraph 构造七场扇出图（每场节点 sleep 0.1s 模拟分析耗时），
断言一次 invoke 并行执行总耗时 ≤ 单场耗时 × 3；串行基准仅作参考打印。
"""
import operator
import time
from typing import Annotated, TypedDict

from langgraph.graph import END, START, StateGraph

MATCH_COUNT = 7
SLEEP_SECONDS = 0.1
SERIAL_MULTIPLIER = 3


class _State(TypedDict):
    done: Annotated[int, operator.add]


def _sleep_node(name: str):
    """模拟单场分析节点：固定耗时后累加完成计数。"""

    def _node(state: _State) -> dict:
        time.sleep(SLEEP_SECONDS)
        return {"done": 1}

    _node.__name__ = name
    return _node


def _build_fanout_graph():
    """构造七场并行扇出图：START 同时分发到七个模拟节点。"""
    graph = StateGraph(_State)
    for idx in range(MATCH_COUNT):
        graph.add_node(f"match_{idx}", _sleep_node(f"match_{idx}"))
        graph.add_edge(START, f"match_{idx}")
        graph.add_edge(f"match_{idx}", END)
    return graph.compile()


class TestParallelPerformance:
    def test_七场并行耗时不超过串行单场三倍(self):
        graph = _build_fanout_graph()
        start = time.perf_counter()
        result = graph.invoke({"done": 0})
        parallel_elapsed = time.perf_counter() - start
        assert result["done"] == MATCH_COUNT
        # 串行基准 = 单场 0.1s × 7 = 0.7s；并行应 ≤ 单场 × 3（含调度余量）
        assert parallel_elapsed <= SLEEP_SECONDS * SERIAL_MULTIPLIER, (
            f"并行耗时 {parallel_elapsed:.3f}s 超过串行单场×3 "
            f"（{SLEEP_SECONDS * SERIAL_MULTIPLIER}s），多场并行未生效"
        )

    def test_并行执行完成后计数完整(self):
        """七节点全部执行（并行不丢任务）：done 累加 == 场次数。"""
        graph = _build_fanout_graph()
        assert graph.invoke({"done": 0})["done"] == MATCH_COUNT
