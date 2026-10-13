# -*- coding: utf-8 -*-
"""CLI 可视化与追溯实现（PRD-12 F-17/F-18/F-19/F-20，ISS-FA4）。

职责边界：
- 负责：graph（Mermaid 导出 + 一致性断言）、replay（轨迹时序回放）、
  show（节点输出查询）、run 图驱动模式（--checkpoint-db 断点续跑 +
  --resume 人工裁决恢复 + 产物落盘）。
- 不负责：run 的 runner 版分析路径（cli.py）、agent 规则、图构建。

replay/show 数据源为 SQLiteCheckpointer 中最近一次 checkpoint 的
trace_events（operator.add 累积，含全部已完成节点事件）。
"""
import json
import sys
from pathlib import Path

from langgraph.types import Command

from fa_orchestrator.graph.config import load_orchestration_config
from fa_orchestrator.graph.graph_builder import build_graph, export_mermaid, mermaid_matches_config
from fa_orchestrator.graph.graph_nodes import GraphNodes
from fa_orchestrator.infra.sqlite_checkpointer import SQLiteCheckpointer

_YAML_PATH = Path(__file__).resolve().parent / "configs" / "orchestrator.yaml"
_INTERRUPT_EXIT = 3
_USAGE_EXIT = 2
# 图驱动 run 的人工裁决恢复值：MVP 统一"确认放行"，细化裁决表单为后续迭代
_RESUME_VALUE = {"approved": True, "note": "CLI 人工裁决确认"}


def cmd_graph(args) -> int:
    """graph 子命令：导出 Mermaid 源码并做与编排配置的一致性断言（F-17 验收 6）。"""
    config = load_orchestration_config(_YAML_PATH)
    graph = build_graph(config, GraphNodes(None).registry())
    mermaid = export_mermaid(graph)
    if not mermaid_matches_config(mermaid, config):
        print("错误：Mermaid 导出与编排配置不一致（F-17 验收 6）", file=sys.stderr)
        return _USAGE_EXIT
    output_path = Path(args.output)
    output_path.write_text(mermaid, encoding="utf-8")
    print(f" Mermaid 源码已写入: {output_path}")
    return 0


def _load_events(db_path: str, thread_id: str) -> list | None:
    """读取 thread 最近 checkpoint 的 trace_events；无记录返回 None。"""
    checkpointer = SQLiteCheckpointer(db_path)
    config = {"configurable": {"thread_id": thread_id}}
    checkpoint = checkpointer.get_tuple(config)
    if checkpoint is None:
        return None
    channel_values = checkpoint.checkpoint.get("channel_values") or {}
    return list(channel_values.get("trace_events") or [])


def cmd_replay(args) -> int:
    """replay 子命令：按执行时序渲染节点轨迹（F-18：耗时/状态/来源可查）。"""
    events = _load_events(args.checkpoint_db, args.thread)
    if events is None:
        print(f"错误：checkpoint 库中无 thread={args.thread} 的轨迹记录", file=sys.stderr)
        return _USAGE_EXIT
    print(f"轨迹回放 thread={args.thread} 共 {len(events)} 个节点事件:")
    for idx, event in enumerate(events, start=1):
        output_keys = ",".join(sorted((event.get("output") or {}).keys())) or "-"
        print(
            f"  {idx:>2}. {event['node']}  耗时 {event['duration_ms']}ms"
            f"  状态 {event['status']}  来源 {event.get('source') or '-'}"
            f"  输出[{output_keys}]"
        )
    return 0


def cmd_show(args) -> int:
    """show 子命令：JSON 输出指定节点的轨迹事件（输入/输出/耗时/来源）。"""
    events = _load_events(args.checkpoint_db, args.thread) or []
    matched = [event for event in events if event["node"] == args.node]
    if not matched:
        print(f"错误：thread={args.thread} 无节点 {args.node} 的轨迹事件", file=sys.stderr)
        return _USAGE_EXIT
    print(json.dumps(matched, ensure_ascii=False, indent=2))
    return 0


def _save_reports(report_dir: str, reports: dict) -> None:
    """按 fid 落盘单场报告 md（F-15）。"""
    dir_path = Path(report_dir)
    dir_path.mkdir(parents=True, exist_ok=True)
    for fid, content in reports.items():
        (dir_path / f"report-{fid}.md").write_text(content, encoding="utf-8")
    print(f"单场报告已写入: {dir_path}（{len(reports)} 份）")


def _save_graph_outputs(args, result: dict) -> None:
    """图驱动 run 的产物落盘：推荐结构 JSON + 单场报告目录。"""
    conclusion = result.get("phase_conclusion") or {}
    if args.recommendation_output and conclusion:
        _save_json_output(args.recommendation_output, conclusion)
    if args.report_dir and result.get("reports"):
        _save_reports(args.report_dir, result["reports"])


def _save_json_output(output, payload: dict) -> None:
    """落盘 JSON 产物并打印路径（与 cli._save_json 同格式）。"""
    output_path = Path(output)
    output_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"推荐结构产物已写入: {output_path}")


def _build_graph(args, data_source):
    """构建带 SQLite checkpointer 的图驱动管线（channel 与 run 子命令一致）。"""
    config = load_orchestration_config(_YAML_PATH)
    registry = GraphNodes(data_source, channel=args.source).registry()
    return build_graph(config, registry, checkpointer=SQLiteCheckpointer(args.checkpoint_db))


def _run_graph(args, data_source) -> int:
    """run 图驱动模式：--checkpoint-db 首跑（manual 中断退出 3）+ --resume 续跑。

    mysql 通道与 fixture 通道同构放行：分析类 SQL 契约已归档
    （ANALYSIS_CONTRACTS 14 契约）并经 2026-10-08 批次 v4 真库验证。
    """
    graph = _build_graph(args, data_source)
    invoke_config = {"configurable": {"thread_id": args.date}}
    if args.resume:
        result = graph.invoke(Command(resume=_RESUME_VALUE), invoke_config)
        _save_graph_outputs(args, result)
        return 0
    result = graph.invoke(
        {"match_date": args.date, "trace": [], "trace_events": []}, invoke_config
    )
    if result.get("__interrupt__"):
        fids = [m["fid"] for m in result["__interrupt__"][0].value["matches"]]
        print(
            f"人工裁决：竞彩日 {args.date} 有矛盾检查未通过场次 {fids}（F-20），"
            f"graph 驱动已暂停；使用 --resume --checkpoint-db 续跑"
        )
        return _INTERRUPT_EXIT
    _save_graph_outputs(args, result)
    return 0
