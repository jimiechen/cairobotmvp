# -*- coding: utf-8 -*-
"""CLI 入口（PRD-12 F-05）：`run --date` 日级数据准备骨架。

ISS-FA1 范围：仅 run 子命令（fixture 数据源），产出并落盘就绪率报告。
replay/show/graph 子命令由 ISS-FA4 实现。
"""
import argparse
import json
import sys
from pathlib import Path

from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.data_source import FixtureDataSource


def build_parser() -> argparse.ArgumentParser:
    """构建 CLI 参数解析器。"""
    parser = argparse.ArgumentParser(
        prog="fa_orchestrator",
        description="FA 足球情报分析多 Agent 编排系统（PRD-12）",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    run_parser = subparsers.add_parser("run", help="日级数据准备与就绪率报告（ISS-FA1）")
    run_parser.add_argument("--date", required=True, help="竞彩日，如 2026-09-29")
    run_parser.add_argument("--fixture", help="fixture 数据源路径（测试/演示）")
    run_parser.add_argument("--output", help="就绪率报告 JSON 输出路径")
    return parser


def _run(args) -> int:
    """执行 run 子命令：取数 → 核查 → 就绪率报告。"""
    if not args.fixture:
        print("错误：ISS-FA1 仅支持 --fixture 数据源（MySQL 通道待 R-04 就绪）", file=sys.stderr)
        return 2

    data_source = FixtureDataSource(args.fixture)
    report = DataAgent(data_source).prepare(args.date)

    ready = report["ready_matches"]
    total = report["total_matches"]
    print(f"竞彩日 {args.date} 数据就绪率: {ready}/{total}")
    for item in report["matches"]:
        status = item["integrity"]["status"]
        missing = ",".join(item["integrity"]["missing_modules"]) or "-"
        print(f"  {item['match_no']} {item['home_team']} vs {item['away_team']}"
              f" fid={item['fid']} match_id={item['match_id']} 状态={status} 缺口={missing}")

    if args.output:
        output_path = Path(args.output)
        output_path.write_text(
            json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        print(f"就绪率报告已写入: {output_path}")
    return 0


def main(argv=None) -> int:
    """CLI 主入口。返回 0 表示成功，非 0 表示失败。"""
    args = build_parser().parse_args(argv)
    if args.command == "run":
        return _run(args)
    return 2


if __name__ == "__main__":
    sys.exit(main())
