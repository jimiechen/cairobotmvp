# -*- coding: utf-8 -*-
"""ReportAgent：逐场 Markdown 报告产出（PRD-12 F-15，ISS-FA4）。

职责边界：
- 负责：按 OUTPUT_TEMPLATE.md 模板组装报告；数据来源声明强制标注
  模式（MySQL/file）与快照时间（F-15/验收 2）；模拟数据检测告警
  （整数赔率/全同赔/缺主流公司/欧赔缺失）；推荐结构/反大众/矛盾检查章节。
- 不负责：分析结论产出（各 agent）、报告文件落盘（CLI/编排层）、图编排。

模拟数据告警是验收 2 的硬要求：无任何模拟数据告警遗漏，无告警时明示"未触发"。
"""
from pathlib import Path

# 主流公司名单（F-15 检测口径）：两者均缺失判数据完整性存疑
MAIN_BOOKMAKERS = ("威廉希尔", "立博")
# 胜平负选项中文映射（报告表格用）
_SELECTION_ZH = {"win": "胜", "draw": "平", "lose": "负"}
_FORM_ZH = {"S": "单选", "D": "双选"}
_ANOMALY_NONE = "未触发模拟数据告警。"

_DEFAULT_TEMPLATE = Path(__file__).parent / "configs" / "OUTPUT_TEMPLATE.md"


def _is_whole(value) -> bool:
    """整数判定：float 整数值（2.0）与 int 均视为整数（模拟赔率典型特征）。"""
    return isinstance(value, (int, float)) and float(value).is_integer()


def detect_data_anomalies(euro: list) -> list[str]:
    """模拟数据检测（F-15）：返回告警文案列表。

    输入：欧赔行列表（{"bk": 公司, "current": {"win","draw","lose"}}）。
    输出：告警列表；正常数据返回空列表。
    检测项：欧赔缺失 → 缺主流公司 → 整数赔率（逐公司）→ 全同赔。
    """
    if not euro:
        return ["欧赔缺失，无法核验市场结构（模拟数据高危信号）"]
    warnings: list[str] = []
    books = {row.get("bk") for row in euro}
    if not any(name in books for name in MAIN_BOOKMAKERS):
        warnings.append(
            f"缺主流公司（{'/'.join(MAIN_BOOKMAKERS)}），数据完整性存疑"
        )
    for row in euro:
        current = row.get("current") or {}
        if current and all(_is_whole(current.get(k)) for k in ("win", "draw", "lose")):
            warnings.append(f"整数赔率疑似模拟数据: {row.get('bk')}")
    wins = [row.get("current", {}).get("win") for row in euro]
    if len(euro) >= 2 and all(w == wins[0] for w in wins):
        warnings.append("全同赔疑似模拟数据（各公司主胜赔完全一致）")
    return warnings


class ReportAgent:
    """报告生成 Agent：输入单场完整结论，输出模板化 Markdown。

    状态约束：无状态；模板路径可注入（默认包内 OUTPUT_TEMPLATE.md）。
    """

    def __init__(self, template_path=None):
        self._template = Path(template_path or _DEFAULT_TEMPLATE)

    def generate(self, match: dict, channel: str, snapshot_time: str) -> str:
        """生成单场报告。

        输入：match（AnalysisRunner/图节点产出的单场结论 dict，含
        odds_snapshot.euro）、channel（数据模式）、snapshot_time（快照时间）。
        输出：Markdown 报告文本。
        失败行为：模板缺失抛 FileNotFoundError；缺章节键按空渲染不抛。
        """
        synthesis = match.get("synthesis") or {}
        review = match.get("review") or {}
        euro = (match.get("odds_snapshot") or {}).get("euro") or []
        fills = {
            "title": (
                f"{match.get('match_no', '-')} "
                f"{match.get('home_team', '-')} vs {match.get('away_team', '-')}"
            ),
            "source_declaration": self._source_declaration(match, channel, snapshot_time),
            "anomaly_warnings": self._anomaly_section(detect_data_anomalies(euro)),
            "prediction_table": self._prediction_table(synthesis.get("recommendation") or {}),
            "anti_public": self._anti_public(synthesis.get("anti_public") or {}),
            "consistency_checks": self._checks_section(review),
        }
        report = self._template.read_text(encoding="utf-8")
        for key, value in fills.items():
            report = report.replace("{{" + key + "}}", value)
        return report

    @staticmethod
    def _source_declaration(match: dict, channel: str, snapshot_time: str) -> str:
        """溯源声明（验收 2）：模式 + 快照时间 + 场次标识，强制标注。"""
        return (
            f"- 数据模式：{channel}\n"
            f"- 快照时间：{snapshot_time}\n"
            f"- 场次：{match.get('match_no', '-')}（fid={match.get('fid', '-')})"
        )

    @staticmethod
    def _anomaly_section(warnings: list) -> str:
        """告警章节：无告警必须明示未触发，禁止静默省略。"""
        if not warnings:
            return _ANOMALY_NONE
        return "\n".join(f"- {w}" for w in warnings)

    @staticmethod
    def _unit_row(layer: str, unit: dict) -> str:
        """推荐单元渲染为表格行；空单元渲染为未配置行。"""
        if not unit:
            return f"| {layer} | - | - | - | - |"
        selection = unit.get("selection")
        selection_zh = _SELECTION_ZH.get(selection, selection or "-")
        legs = "/".join(unit.get("legs") or []) or "-"
        path = unit.get("path") or "-"
        form = _FORM_ZH.get(unit.get("form") or "", unit.get("form") or "-")
        return f"| {layer} | {form} | {selection_zh}{legs if legs != '-' else ''} | {path} | {unit.get('stars', '-')} |"

    def _prediction_table(self, recommendation: dict) -> str:
        """竞彩让球胜平负预测表（验收 1：推荐结构 S/D 单元）。"""
        header = "| 层 | 形式 | 选项 | 路径 | 星级 |\n|---|---|---|---|---|"
        rows = [
            self._unit_row("胜平负", recommendation.get("had_unit")),
            self._unit_row("让球", recommendation.get("hhand_unit")),
        ]
        return "\n".join([header, *rows])

    @staticmethod
    def _anti_public(anti_public: dict) -> str:
        """反大众检查章节：引用 SynthesisAgent 结论原文。"""
        note = anti_public.get("note") or "无反大众检查结论"
        return f"- {note}"

    @staticmethod
    def _checks_section(review: dict) -> str:
        """矛盾检查章节（验收 5）：七项结果逐项落盘 + 决策。"""
        checks = review.get("checks") or []
        lines = [f"决策：{review.get('decision', '-')}",
                 "| 检查项 | 名称 | 结果 | 修正动作 |", "|---|---|---|---|"]
        for check in checks:
            action = check.get("action") or "-"
            lines.append(
                f"| {check.get('id', '-')} | {check.get('name', '-')}"
                f" | {check.get('result', '-')} | {action} |"
            )
        return "\n".join(lines)
