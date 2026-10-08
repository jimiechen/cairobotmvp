# -*- coding: utf-8 -*-
"""AnalysisRunner：日级分析管线（PRD-12 F-10~F-14/F-16，ISS-FA2/FA3 交付物）。

职责边界：
- 负责：逐场取数 → FundamentalAgent（第一阶段）→ OddsVerifyAgent（第二阶段）
  → 历史推演（F-16）→ SentimentAgent（F-12）→ SynthesisAgent（F-13）
  → ReviewAgent 七项矛盾检查与条件回退环（F-14）→ 日级汇总与落盘前组装。
- 不负责：数据源实现（data_source.py）、分析规则本身（各 agent 文件）、
  图编排（graph_builder.py，ISS-FA4 起以 StateGraph 节点复用同一批 agent）。

输入数据块缺失时按降级标注保留场次（禁止静默跳过），不编造结论。
"""

from fa_orchestrator.agents.data_agent import DataAgent
from fa_orchestrator.data.data_source import CHANNEL_FIXTURE
from fa_orchestrator.agents.fundamental_agent import FundamentalAgent
from fa_orchestrator.analysis.history_deduction import deduce_heat_drift, deduce_trend
from fa_orchestrator.analysis.odds_math import implied_probabilities
from fa_orchestrator.agents.odds_verify_agent import OddsVerifyAgent
from fa_orchestrator.agents.review_agent import MAX_REVIEW_RETRIES, ReviewAgent
from fa_orchestrator.agents.sentiment_agent import SentimentAgent
from fa_orchestrator.agents.synthesis_agent import SynthesisAgent

# 近期交锋平局率 ≥ 该值判"平局方向有基本面支持"（draw_support，v7.2 防冷口径）
DRAW_SUPPORT_RATE = 0.4


class AnalysisRunner:
    """日级分析运行器：输入竞彩日，输出全场次结论与推荐结构汇总。

    状态约束：无状态；数据源与 agent 均注入（测试可替换为 fake）。
    """

    def __init__(self, data_source, channel: str = CHANNEL_FIXTURE, fundamental_agent=None,
                 odds_agent=None, sentiment_agent=None, synthesis_agent=None, review_agent=None):
        self.data_source = data_source
        self.channel = channel
        self.fundamental_agent = fundamental_agent or FundamentalAgent()
        self.odds_agent = odds_agent or OddsVerifyAgent()
        self.sentiment_agent = sentiment_agent or SentimentAgent()
        self.synthesis_agent = synthesis_agent or SynthesisAgent()
        self.review_agent = review_agent or ReviewAgent()

    def run(self, match_date: str) -> dict:
        """执行日级分析并产出全场次结论（两阶段 + 热度/综合判定/矛盾检查）。

        输入：竞彩日（如 2026-09-29）。
        输出：{"match_date","channel","total_matches","matches","manual_review_matches"}，
        每场含 fundamental/odds_verify/history_deduction/sentiment/synthesis/
        review/recommendation。
        失败行为：单场分析数据缺失按降级标注保留场次；数据源异常向上抛出。
        """
        data_agent = DataAgent(self.data_source, channel=self.channel)
        report = data_agent.prepare(match_date)
        matches = [self._analyze_match(row) for row in report["matches"]]
        return {
            "match_date": match_date,
            "channel": data_agent.channel,
            "total_matches": len(matches),
            "matches": matches,
            "manual_review_matches": [
                m["fid"] for m in matches if m["review"]["decision"] == "manual"
            ],
            **self._degradation_notes(),
        }

    def _degradation_notes(self) -> dict:
        """数据源降级标注探测：支持自描述的数据源（如 mysql 通道）随结论透出。"""
        notes = getattr(self.data_source, "degradation_notes", None)
        return {"degradation_notes": list(notes)} if notes else {}

    def _analyze_match(self, row: dict) -> dict:
        """单场分析：基本面 → 赔率验证 → 历史推演 → 热度 → 综合判定+矛盾检查。"""
        fid = row["fid"]
        odds_data = self.data_source.query_match_odds(fid) or {}
        fundamentals = self.data_source.query_fundamentals(fid)
        fundamental = self.fundamental_agent.analyze({
            **fundamentals,
            "initial_position": self._extract_initial_position(odds_data),
        })
        vote = self.data_source.query_vote(fid)
        odds_verify = self.odds_agent.verify(
            self._build_odds_input(odds_data, fundamental, fundamentals)
        )
        sentiment = self.sentiment_agent.analyze({
            "vote": vote,
            "euro": odds_data.get("euro") or [],
            "fundamental": fundamental,
        })
        synthesis, review = self._synthesize_with_review(
            fundamental, odds_verify, sentiment, vote, odds_data
        )
        return {
            "fid": fid,
            "match_no": row["match_no"],
            "home_team": row["home_team"],
            "away_team": row["away_team"],
            "integrity": row["integrity"],
            "vote_available": vote is not None,
            "fundamental": fundamental,
            "odds_verify": odds_verify,
            "history_deduction": self._deduce_history(fid),
            "sentiment": sentiment,
            "synthesis": synthesis,
            "review": review,
            "recommendation": synthesis["recommendation"],
        }

    def _synthesize_with_review(self, fundamental: dict, odds_verify: dict,
                                sentiment: dict, vote: dict | None,
                                odds_data: dict) -> tuple:
        """综合判定 + 矛盾检查条件回退环（F-14）。

        retry 决策时重跑 SynthesisAgent（attempt 递增），最多 MAX_REVIEW_RETRIES 次；
        预算耗尽仍不通过 → 转人工裁决（manual），由 run() 登记 manual_review_matches。
        """
        synthesis_input = {
            "fundamental": fundamental,
            "odds": odds_verify,
            "sentiment": sentiment,
            "vote": vote,
            "market_direction": self._market_direction(odds_data.get("euro") or []),
        }
        payload = {"handicap_line": self._numeric_handicap(odds_data)}
        attempt = 0
        synthesis = self.synthesis_agent.analyze(synthesis_input, attempt=attempt)
        review = self.review_agent.review(
            {**payload, "synthesis": synthesis, "attempt": attempt}
        )
        while review["decision"] == "retry" and attempt < MAX_REVIEW_RETRIES:
            attempt += 1
            synthesis = self.synthesis_agent.analyze(synthesis_input, attempt=attempt)
            review = self.review_agent.review(
                {**payload, "synthesis": synthesis, "attempt": attempt}
            )
        if review["decision"] == "retry":
            # 回退预算耗尽仍不通过：按 F-14 契约转人工裁决
            review = {**review, "decision": "manual"}
        return synthesis, review

    @staticmethod
    def _market_direction(euro: list) -> str | None:
        """市场方向：欧赔去水隐含概率最高项；欧赔缺失/非法返回 None（综合判定兜底）。"""
        implied = implied_probabilities(euro)
        return max(implied, key=implied.get) if implied else None

    @staticmethod
    def _numeric_handicap(odds_data: dict) -> float | None:
        """亚盘现盘让球线（仅数值型透传）：中文盘口文本不解析，None 按非深盘。"""
        asian = odds_data.get("asian") or []
        row = next((r for r in asian if r.get("bk") == "澳门"), asian[0] if asian else None)
        handicap = (row or {}).get("current", {}).get("handicap")
        return handicap if isinstance(handicap, (int, float)) else None

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
