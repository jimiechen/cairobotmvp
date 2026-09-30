# -*- coding: utf-8 -*-
"""CLI 入口（PRD-12 F-05）：`run --date` 日级数据准备与两阶段分析。

数据源通道（R-04 联调后双通道）：--source fixture（默认，测试/演示）
与 --source mysql（真实 workbench 库，配置走 FA_MYSQL_* 环境变量）。
MySQL 通道仅覆盖数据准备（Q1/Q2b/解析）；两阶段分析与推荐结构产出
仍仅支持 fixture（分析类 SQL 契约未归档，经主控确认不纳入）。
replay/show/graph 子命令由 ISS-FA4 实现。
"""
import argparse
import json
import sys
from pathlib import Path

from fa_orchestrator.analysis_runner import AnalysisRunner
from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.data_source import CHANNEL_FIXTURE, CHANNEL_MYSQL, FixtureDataSource
from fa_orchestrator.mysql_data_source import MySqlConfigError, MySqlDataSource


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
    run_parser.add_argument(
        "--source",
        choices=(CHANNEL_FIXTURE, CHANNEL_MYSQL),
        default=CHANNEL_FIXTURE,
        help="数据源通道：fixture（默认）或 mysql（需 FA_MYSQL_* 环境变量）",
    )
    run_parser.add_argument("--output", help="就绪率报告 JSON 输出路径")
    run_parser.add_argument(
        "--analysis-fixture", help="两阶段分析数据块 fixture 路径（ISS-FA2）"
    )
    run_parser.add_argument(
        "--analysis-output", help="两阶段中间结论 JSON 输出路径（ISS-FA2）"
    )
    run_parser.add_argument(
        "--recommendation-output", help="推荐结构 JSON 输出路径（ISS-FA3）"
    )
    return parser


def _build_source(args):
    """按 --source 构建数据源。

    返回：data_source 实例。
    失败行为：mysql 通道缺 FA_MYSQL_* 配置抛 MySqlConfigError（调用方转退出码 2）。
    """
    if args.source == CHANNEL_MYSQL:
        return MySqlDataSource()
    if not args.fixture:
        raise ValueError("fixture 通道必须提供 --fixture 路径")
    return FixtureDataSource(args.fixture, analysis_path=args.analysis_fixture)


def _run(args) -> int:
    """执行 run 子命令：取数 → 核查 → 就绪率报告（→ 两阶段分析）。"""
    if args.source == CHANNEL_MYSQL and (args.analysis_output or args.recommendation_output):
        print(
            "错误：MySQL 通道暂不支持两阶段分析与推荐结构产出"
            "（分析类 SQL 契约未归档，R-04 范围限定）",
            file=sys.stderr,
        )
        return 2
    try:
        data_source = _build_source(args)
    except MySqlConfigError as exc:
        print(f"错误：{exc}", file=sys.stderr)
        return 2
    except ValueError as exc:
        print(f"错误：{exc}", file=sys.stderr)
        return 2

    report = DataAgent(data_source, channel=args.source).prepare(args.date)
    _print_ready_report(args, report)

    if args.output:
        _save_json(args.output, report, "就绪率报告")
    conclusion = None
    if args.analysis_output or args.recommendation_output:
        conclusion = AnalysisRunner(data_source).run(args.date)
    if args.analysis_output:
        _print_phase_conclusion(conclusion)
        _save_json(args.analysis_output, conclusion, "两阶段中间结论")
    if args.recommendation_output:
        _print_recommendation(conclusion)
        _save_json(args.recommendation_output, conclusion, "推荐结构产物")
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


def _print_recommendation(conclusion: dict) -> None:
    """输出推荐结构摘要到终端（ISS-FA3：决策 + 两层单元形态）。"""
    for item in conclusion["matches"]:
        units = item["recommendation"]
        print(f"  [推荐] {item['match_no']} 决策={item['review']['decision']}"
              f" 胜平负={_unit_desc(units['had_unit'])}"
              f" 让球={_unit_desc(units['hhand_unit'])}")


def _unit_desc(unit: dict | None) -> str:
    """推荐单元终端摘要：单选显示方向与星级，双选显示腿组合与路径。"""
    if not unit:
        return "空"
    if unit["form"] == "S":
        return f"S:{unit['selection']}({unit['stars']}星)"
    return f"D:{'+'.join(unit['legs'])}(路径{unit.get('path') or '无'})"


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
