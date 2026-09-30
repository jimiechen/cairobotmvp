# -*- coding: utf-8 -*-
"""AnalysisRunner：两阶段分析管线（PRD-12 F-10/F-11/F-16，ISS-FA2 交付物）。

职责边界：
- 负责：逐场取数 → FundamentalAgent（第一阶段）→ OddsVerifyAgent（第二阶段）
  → 历史推演（F-16）→ 两阶段中间结论日级汇总与落盘前组装。
- 不负责：数据源实现（data_source.py）、分析规则本身（fundamental_agent.py /
  odds_verify_agent.py）、热度偏差计算（ISS-FA3 SentimentAgent）、
  推荐结构判定（ISS-FA3 SynthesisAgent）、图编排（graph_builder.py）。

输入数据块缺失时按降级标注保留场次（禁止静默跳过），不编造结论。
"""

from fa_orchestrator.data_agent import DataAgent
from fa_orchestrator.fundamental_agent import FundamentalAgent
from fa_orchestrator.history_deduction import deduce_heat_drift, deduce_trend
from fa_orchestrator.odds_verify_agent import OddsVerifyAgent

# 近期交锋平局率 ≥ 该值判"平局方向有基本面支持"（draw_support，v7.2 防冷口径）
DRAW_SUPPORT_RATE = 0.4


class AnalysisRunner:
    """两阶段分析运行器：输入竞彩日，输出全场次中间结论汇总。

    状态约束：无状态；数据源与 agent 均注入（测试可替换为 fake）。
    """

    def __init__(self, data_source, fundamental_agent=None, odds_agent=None):
        self.data_source = data_source
        self.fundamental_agent = fundamental_agent or FundamentalAgent()
        self.odds_agent = odds_agent or OddsVerifyAgent()

    def run(self, match_date: str) -> dict:
        """执行两阶段分析并产出日级中间结论。

        输入：竞彩日（如 2026-09-29）。
        输出：{"match_date","channel","total_matches","matches":[...]}，
        每场含 integrity/vote_available/fundamental/odds_verify/history_deduction。
        失败行为：单场分析数据缺失按降级标注保留场次；数据源异常向上抛出。
        """
        data_agent = DataAgent(self.data_source)
        report = data_agent.prepare(match_date)
        matches = [self._analyze_match(row) for row in report["matches"]]
        return {
            "match_date": match_date,
            "channel": data_agent.channel,
            "total_matches": len(matches),
            "matches": matches,
        }

    def _analyze_match(self, row: dict) -> dict:
        """单场两阶段分析：基本面 → 赔率验证 → 历史推演（F-10/F-11/F-16）。"""
        fid = row["fid"]
        odds_data = self.data_source.query_match_odds(fid) or {}
        fundamentals = self.data_source.query_fundamentals(fid)
        fundamental = self.fundamental_agent.analyze({
            **fundamentals,
            "initial_position": self._extract_initial_position(odds_data),
        })
        return {
            "fid": fid,
            "match_no": row["match_no"],
            "home_team": row["home_team"],
            "away_team": row["away_team"],
            "integrity": row["integrity"],
            "vote_available": self.data_source.query_vote(fid) is not None,
            "fundamental": fundamental,
            "odds_verify": self.odds_agent.verify(
                self._build_odds_input(odds_data, fundamental, fundamentals)
            ),
            "history_deduction": self._deduce_history(fid),
        }

    @staticmethod
    def _extract_initial_position(odds_data: dict) -> dict:
        """初盘定位提取（Q3b）：亚盘澳门行优先，无澳门取第一家；无亚盘返回空。"""
        asian = odds_data.get("asian") or []
        row = next((r for r in asian if r.get("bk") == "澳门"), asian[0] if asian else None)
        if not row:
            return {}
        return {
            "initial_handicap": row.get("initial", {}).get("handicap"),
            "current_handicap": row.get("current", {}).get("handicap"),
        }

    def _build_odds_input(self, odds_data: dict, fundamental: dict, fundamentals: dict) -> dict:
        """组装第二阶段输入：三盘赔率 + 状态基线 + 基本面强度背书 + 平局支持。"""
        schedule = fundamentals.get("schedule") or {}
        return {
            "euro": odds_data.get("euro") or [],
            "asian": odds_data.get("asian") or [],
            "overunder": odds_data.get("overunder") or [],
            "state_baseline": {
                "home_avg_goals": fundamental["recent_form"]["home_avg_scored"],
            },
            "strength": {
                "value_ratio": None,
                "level_gap": schedule.get("level_gap", False),
                "h2h_suppression": fundamental["h2h"]["suppression"],
            },
            "draw_support": self._has_draw_support(fundamentals.get("h2h") or {}),
        }

    def _deduce_history(self, fid: str) -> dict:
        """F-16 历史推演：升降盘时序 + 热度漂移；缺失时标注无法推演时序。"""
        return {
            "trend": deduce_trend(self.data_source.query_history_tracks(fid)),
            "heat_drift": deduce_heat_drift(self.data_source.query_heat_drift(fid)),
        }

    @staticmethod
    def _has_draw_support(h2h: dict) -> bool:
        """近期交锋平局率 ≥ DRAW_SUPPORT_RATE 判平局方向有基本面支持（v7.2）。"""
        records = h2h.get("records") or []
        if not records:
            return False
        draw_count = sum(1 for r in records if r.get("result") == "D")
        return draw_count / len(records) >= DRAW_SUPPORT_RATE
