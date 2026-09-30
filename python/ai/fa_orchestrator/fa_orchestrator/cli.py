# -*- coding: utf-8 -*-
"""CLI 入口（PRD-12 F-05）：`run --date` 日级数据准备与两阶段分析。

ISS-FA1 范围：run 子命令（fixture 数据源），产出并落盘就绪率报告。
ISS-FA2 扩展：--analysis-fixture/--analysis-output 触发两阶段分析管线，
产出七场中间结论 JSON。replay/show/graph 子命令由 ISS-FA4 实现。
"""
import argparse
import json
import sys
from pathlib import Path

from fa_orchestrator.analysis_runner import AnalysisRunner
from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.data_source import FixtureDataSource


def build_parser() -> argparse.ArgumentParser:
    """构建 CLI 参数解析器。"""
    parser = argparse.ArgumentParser(
        prog="fa_orchestrator",
        description="FA 足球情报分析多 Agent 编排系统（PRD-12）",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    run_parser = subparsers.add_parser("run", help="日级数据准备与两阶段分析")
    run_parser.add_argument("--date", required=True, help="竞彩日，如 2026-09-29")
    run_parser.add_argument("--fixture", help="fixture 数据源路径（测试/演示）")
    run_parser.add_argument("--output", help="就绪率报告 JSON 输出路径")
    run_parser.add_argument(
        "--analysis-fixture", help="两阶段分析数据块 fixture 路径（ISS-FA2）"
    )
    run_parser.add_argument(
        "--analysis-output", help="两阶段中间结论 JSON 输出路径（ISS-FA2）"
    )
    return parser


def _run(args) -> int:
    """执行 run 子命令：取数 → 核查 → 就绪率报告（→ 两阶段分析）。"""
    if not args.fixture:
        print("错误：当前仅支持 --fixture 数据源（MySQL 通道待 R-04 就绪）", file=sys.stderr)
        return 2

    data_source = FixtureDataSource(args.fixture, analysis_path=args.analysis_fixture)
    report = DataAgent(data_source).prepare(args.date)
    _print_ready_report(args, report)

    if args.output:
        _save_json(args.output, report, "就绪率报告")
    if args.analysis_output:
        conclusion = AnalysisRunner(data_source).run(args.date)
        _print_phase_conclusion(conclusion)
        _save_json(args.analysis_output, conclusion, "两阶段中间结论")
    return 0


def _print_ready_report(args, report: dict) -> None:
    """输出就绪率报告摘要到终端。"""
    ready = report["ready_matches"]
    total = report["total_matches"]
    print(f"竞彩日 {args.date} 数据就绪率: {ready}/{total}")
    for item in report["matches"]:
        status = item["integrity"]["status"]
        missing = ",".join(item["integrity"]["missing_modules"]) or "-"
        print(f"  {item['match_no']} {item['home_team']} vs {item['away_team']}"
              f" fid={item['fid']} match_id={item['match_id']} 状态={status} 缺口={missing}")


def _print_phase_conclusion(conclusion: dict) -> None:
    """输出两阶段中间结论摘要到终端。"""
    for item in conclusion["matches"]:
        euro_asia = item["odds_verify"]["euro_asia"]
        print(f"  [分析] {item['match_no']} 升盘解读={euro_asia['interpretation'] or '无升盘'}"
              f" 标签={''.join(euro_asia['tags']) or '-'}"
              f" 时序={'可推演' if item['history_deduction']['trend']['available'] else '无法推演'}")


def _save_json(output, payload: dict, label: str) -> None:
    """落盘 JSON 产物并打印路径。"""
    output_path = Path(output)
    output_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"{label}已写入: {output_path}")


def main(argv=None) -> int:
    """CLI 主入口。返回 0 表示成功，非 0 表示失败。"""
    args = build_parser().parse_args(argv)
    if args.command == "run":
        return _run(args)
    return 2


if __name__ == "__main__":
    sys.exit(main())
