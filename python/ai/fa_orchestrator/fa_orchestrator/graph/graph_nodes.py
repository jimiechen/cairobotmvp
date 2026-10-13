# -*- coding: utf-8 -*-
"""图节点注册表：把分析 Agent 包装为 StateGraph 节点函数（PRD-12 F-18/F-20，ISS-FA4）。

职责边界：
- 负责：日级批处理节点（planner/data/四类分析/review/human_review/report/phase_report）、
  结构化轨迹事件记录（F-18：耗时/状态/输出/数据来源）、人工裁决 interrupt（F-20）、
  复用 AnalysisRunner 单场输入组装逻辑保证图驱动与 runner 版结论对齐。
- 不负责：agent 规则实现、图构建（graph_builder.py）、checkpoint 存储、CLI。

human_review 使用 langgraph.types.interrupt：需配合 checkpointer 才能恢复；
无 manual 场时不触发，pass 全流程不受影响。retry 回退环（F-14）由
review_agent 条件边（retry → synthesis_agent）图级强制，回退次数有界（≤2）。
"""
import time

from langgraph.types import interrupt

from fa_orchestrator.analysis.analysis_runner import AnalysisRunner
from fa_orchestrator.agents.data_agent import CHANNEL_FIXTURE, DataAgent
from fa_orchestrator.report.report_agent import ReportAgent

# 本地静态计划阶段（fixture 模式无 LLM 依赖；联调时可注入 LLM 版 planner）
_LOCAL_PLAN_STAGES = ("fundamental", "odds_verify", "sentiment", "greed_fear",
                      "synthesis", "review")


class GraphNodes:
    """日级图节点注册表：输入数据源与各 Agent，输出 agent 键 → 节点函数映射。

    状态约束：无状态；单场结论经 state 逐节点传递，节点间不共享可变对象。
    单场数据（odds/vote/fundamentals）按节点职责独立取数，保证轨迹中
    数据来源（F-18）真实标注而非透传。
    """

    def __init__(self, data_source, fundamental_agent=None, odds_agent=None,
                 sentiment_agent=None, synthesis_agent=None, review_agent=None,
                 report_agent=None, planner_fn=None, channel=CHANNEL_FIXTURE,
                 greed_fear_agent=None):
        self.data_source = data_source
        self.channel = channel
        self._runner = AnalysisRunner(
            data_source, channel=channel,
            fundamental_agent=fundamental_agent, odds_agent=odds_agent,
            sentiment_agent=sentiment_agent, synthesis_agent=synthesis_agent,
            review_agent=review_agent, greed_fear_agent=greed_fear_agent,
        )
        self._report_agent = report_agent or ReportAgent()
        self._planner_fn = planner_fn or self._local_plan

    def registry(self) -> dict:
        """返回 agent 键 → 节点函数映射（覆盖 orchestrator.yaml 全部 agent）。"""
        return {
            "planner_agent": self._traced(self._planner, "planner"),
            "data_agent": self._traced(self._data, "data_agent"),
            "ready_report": self._traced(self._ready_report, "ready_report"),
            "fundamental_agent": self._traced(self._fundamental, "fundamental_agent"),
            "odds_verify_agent": self._traced(self._odds_verify, "odds_verify_agent"),
            "sentiment_agent": self._traced(self._sentiment, "sentiment_agent"),
            "greed_fear_agent": self._traced(self._greed_fear, "greed_fear_agent"),
            "synthesis_agent": self._traced(self._synthesis, "synthesis_agent"),
            "review_agent": self._traced(self._review, "review_agent"),
            "human_review": self._traced(self._human_review, "human_review"),
            "report_agent": self._traced(self._report, "report_agent"),
            "phase_report": self._traced(self._phase_report, "phase_report"),
        }

    def _traced(self, fn, name):
        """节点包装：记录结构化轨迹事件（F-18），并追加 trace 节点名。"""
        def _node(state: dict) -> dict:
            started = time.perf_counter()
            update = fn(state) or {}
            event = {
                "node": name,
                "duration_ms": int((time.perf_counter() - started) * 1000),
                "status": "ok",
                "output": {k: v for k, v in update.items()
                           if k not in ("trace", "trace_events")},
                "source": update.get("channel") or self.channel,
            }
            return {"trace": [name], "trace_events": [event], **update}
        _node.__name__ = name
        return _node

    # ---------- 编排前段：计划与数据准备 ----------

    @staticmethod
    def _local_plan(match_date: str) -> dict:
        """本地静态计划（fixture 模式）：按 PRD 两阶段+质检流程排布执行阶段。"""
        return {"match_date": match_date, "stages": list(_LOCAL_PLAN_STAGES)}

    def _planner(self, state: dict) -> dict:
        return {"plan": self._planner_fn(state["match_date"])}

    def _data(self, state: dict) -> dict:
        data_agent = DataAgent(self.data_source, channel=self.channel)
        report = data_agent.prepare(state["match_date"])
        return {"matches": report["matches"], "integrity_report": report,
                "channel": data_agent.channel}

    @staticmethod
    def _ready_report(state: dict) -> dict:
        return {"ready_report": state.get("integrity_report") or {}}

    # ---------- 分析链：四类 Agent 逐场批处理 ----------

    def _fundamental(self, state: dict) -> dict:
        fundamentals = {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            odds_data = self.data_source.query_match_odds(fid) or {}
            inputs = {
                **(self.data_source.query_fundamentals(fid) or {}),
                "initial_position": self._runner._extract_initial_position(odds_data),
            }
            fundamentals[fid] = self._runner.fundamental_agent.analyze(inputs)
        return {"fundamentals": fundamentals}

    def _odds_verify(self, state: dict) -> dict:
        verification = {}
        history = {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            odds_data = self.data_source.query_match_odds(fid) or {}
            fundamentals_row = self.data_source.query_fundamentals(fid) or {}
            verification[fid] = self._runner.odds_agent.verify(
                self._runner._build_odds_input(
                    odds_data, (state.get("fundamentals") or {}).get(fid) or {},
                    fundamentals_row,
                )
            )
            history[fid] = self._runner._deduce_history(fid)
        return {"odds_verification": verification, "history_deduction": history}

    def _sentiment(self, state: dict) -> dict:
        sentiment_map = {}
        vote_available = {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            vote = self.data_source.query_vote(fid)
            vote_available[fid] = vote is not None
            odds_data = self.data_source.query_match_odds(fid) or {}
            sentiment_map[fid] = self._runner.sentiment_agent.analyze({
                "vote": vote,
                "euro": odds_data.get("euro") or [],
                "fundamental": (state.get("fundamentals") or {}).get(fid) or {},
            })
        return {"sentiment": sentiment_map, "vote_availability": vote_available}

    def _greed_fear(self, state: dict) -> dict:
        """贪婪恐惧节点（立项三）：sentiment 后逐场产出贪婪/恐惧评分与陷阱组合。

        输入复用 runner 口径（sentiment/fundamental/odds_verify/欧亚盘）；
        结论只作为 synthesis 警示输入，不参与重试环重算。
        """
        greed_map = {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            odds_data = self.data_source.query_match_odds(fid) or {}
            greed_map[fid] = self._runner.greed_fear_agent.analyze({
                "sentiment": (state.get("sentiment") or {}).get(fid) or {},
                "fundamental": (state.get("fundamentals") or {}).get(fid) or {},
                "odds_verify": (state.get("odds_verification") or {}).get(fid) or {},
                "euro": odds_data.get("euro") or [],
                "asian": odds_data.get("asian") or [],
            })
        return {"greed_fear": greed_map}

    def _synthesis(self, state: dict) -> dict:
        """综合判定节点：synthesis 后即时产出大小球/比分池层（ISS-FA4）。

        goal_layer 依赖 synthesis 结论（S2/S3 证据），retry 回退时随 synthesis
        一并重算；pass/manual 场保留既有结论不重算。
        """
        synthesis_map = dict(state.get("synthesis") or {})
        goal_layer_map = dict(state.get("goal_layer") or {})
        attempts = dict(state.get("synthesis_attempts_by_match") or {})
        review_map = state.get("review") or {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            prev_decision = (review_map.get(fid) or {}).get("decision")
            if prev_decision not in (None, "retry"):
                # pass/manual 场保留既有结论；仅 retry 场按回退预算重合成
                continue
            attempt = attempts.get(fid, 0)
            synthesis_map[fid] = self._runner.synthesis_agent.analyze(
                self._synthesis_input(state, row), attempt=attempt
            )
            attempts[fid] = attempt + 1
            goal_layer_map[fid] = self._goal_layer(state, row, synthesis_map[fid])
        return {"synthesis": synthesis_map, "goal_layer": goal_layer_map,
                "synthesis_attempts_by_match": attempts}

    def _goal_layer(self, state: dict, row: dict, synthesis: dict) -> dict:
        """组装并执行单场大小球/比分池层（复用 runner 输入组装保证对齐）。"""
        fid = row["fid"]
        odds_data = self.data_source.query_match_odds(fid) or {}
        return self._runner.goal_layer_agent.analyze({
            **self._runner._goal_layer_input(
                (state.get("fundamentals") or {}).get(fid) or {},
                (state.get("odds_verification") or {}).get(fid) or {},
                odds_data,
                self._runner._market_direction(odds_data.get("euro") or []),
            ),
            "synthesis": synthesis,
        })

    def _review(self, state: dict) -> dict:
        """矛盾检查节点（F-14）：attempt 计入结果回写 state，保证回退环有界。

        回退语义与 AnalysisRunner._synthesize_with_review 对齐：
        attempt 0→1→2，attempt 达 MAX_REVIEW_RETRIES 仍违规转 manual，
        防止 review 返回值缺失 attempt 字段导致的无限回退环。
        """
        review_map = {}
        prev_review = state.get("review") or {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            odds_data = self.data_source.query_match_odds(fid) or {}
            attempt = (prev_review.get(fid) or {}).get("attempt", -1) + 1
            result = self._runner.review_agent.review({
                "synthesis": (state.get("synthesis") or {}).get(fid) or {},
                "goal_layer": (state.get("goal_layer") or {}).get(fid) or {},
                "handicap_line": self._runner._numeric_handicap(odds_data),
                "attempt": attempt,
            })
            review_map[fid] = {**result, "attempt": attempt}
        decisions = [item["decision"] for item in review_map.values()]
        if "retry" in decisions:
            decision = "retry"
        elif "manual" in decisions:
            decision = "manual"
        else:
            decision = "pass"
        manual = [fid for fid, item in review_map.items()
                  if item["decision"] == "manual"]
        return {"review": review_map, "review_decision": decision,
                "manual_review": manual}

    # ---------- 人工裁决与报告产出 ----------

    def _human_review(self, state: dict) -> dict:
        """人工裁决节点（F-20）：manual 场 interrupt 等待分析师确认后继续。"""
        fids = list(state.get("manual_review") or [])
        if not fids:
            return {}
        review_map = state.get("review") or {}
        answer = interrupt({
            "reason": "矛盾检查未通过，需人工裁决（F-20）",
            "matches": [
                {"fid": fid,
                 "violations": (review_map.get(fid) or {}).get("violations") or []}
                for fid in fids
            ],
        })
        return {"human_decisions": {fid: answer for fid in fids}}

    def _report(self, state: dict) -> dict:
        reports = {}
        for row in state.get("matches") or []:
            fid = row["fid"]
            odds_data = self.data_source.query_match_odds(fid) or {}
            euro = odds_data.get("euro") or []
            snapshot_time = (euro[0].get("snapshot_time") if euro else None) or "-"
            reports[fid] = self._report_agent.generate(
                self._assemble(state, row, odds_data),
                channel=self.channel, snapshot_time=snapshot_time,
            )
        return {"reports": reports}

    def _phase_report(self, state: dict) -> dict:
        matches = [
            self._assemble(
                state, row, self.data_source.query_match_odds(row["fid"]) or {}
            )
            for row in state.get("matches") or []
        ]
        return {"phase_conclusion": {
            "match_date": state["match_date"],
            "channel": self.channel,
            "total_matches": len(matches),
            "matches": matches,
            "manual_review_matches": [
                fid for fid, item in (state.get("review") or {}).items()
                if item["decision"] == "manual"
            ],
        }}

    # ---------- 内部组装（复用 AnalysisRunner 保证结论对齐） ----------

    def _synthesis_input(self, state: dict, row: dict) -> dict:
        fid = row["fid"]
        odds_data = self.data_source.query_match_odds(fid) or {}
        return {
            "fundamental": (state.get("fundamentals") or {}).get(fid) or {},
            "odds": (state.get("odds_verification") or {}).get(fid) or {},
            "sentiment": (state.get("sentiment") or {}).get(fid) or {},
            "vote": self.data_source.query_vote(fid),
            "market_direction": self._runner._market_direction(
                odds_data.get("euro") or []
            ),
            "greed_fear": (state.get("greed_fear") or {}).get(fid) or {},
        }

    def _assemble(self, state: dict, row: dict, odds_data: dict) -> dict:
        """组装单场完整结论（键结构与 AnalysisRunner._analyze_match 对齐）。"""
        fid = row["fid"]
        synthesis = (state.get("synthesis") or {}).get(fid) or {}
        return {
            "fid": fid,
            "match_no": row["match_no"],
            "home_team": row["home_team"],
            "away_team": row["away_team"],
            "integrity": row["integrity"],
            "vote_available": (state.get("vote_availability") or {}).get(fid, False),
            "fundamental": (state.get("fundamentals") or {}).get(fid) or {},
            "odds_verify": (state.get("odds_verification") or {}).get(fid) or {},
            "history_deduction": (state.get("history_deduction") or {}).get(fid) or {},
            "sentiment": (state.get("sentiment") or {}).get(fid) or {},
            "greed_fear": (state.get("greed_fear") or {}).get(fid) or {},
            "synthesis": synthesis,
            "goal_layer": (state.get("goal_layer") or {}).get(fid) or {},
            "review": (state.get("review") or {}).get(fid) or {},
            "recommendation": synthesis.get("recommendation") or {},
            "odds_snapshot": {"euro": odds_data.get("euro") or []},
        }
