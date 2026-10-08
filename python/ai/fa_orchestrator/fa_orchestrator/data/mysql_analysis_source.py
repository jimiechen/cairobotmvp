# -*- coding: utf-8 -*-
"""MySQL 分析类数据源（PRD-12 F-06/F-09，方案 A 批次 1）。

职责边界：
- 负责：Q3/Q5/Q9~Q12/D9/D10/D12 分析类查询——消费 fid→match.id 解析映射、
  执行契约 SQL 并组织行解析，输出结构与 fixture 分析契约
  （tests/fixtures/sample_day_*_analysis.json）对齐。
- 不负责：SQL 契约文本（sql_contracts.py）、行→契约结构解析纯函数
  （mysql_analysis_parse.py）、数据准备查询（mysql_data_source.py）、
  分析计算与趋势判断（analysis/ 层）、复盘折算（match_review.py）。

字段口径与降级标注（2026-10-08 workbench 只读探测校准，R-04）：
- 数据域：模块表 match_id 域 = `match`.id；D10 赛果例外（site_match_id == fid 直连）
- D9 变盘轨迹为降级实现：仅按参考公司（TRACK_BOOKMAKER_NAME）快照序列推演变盘
  方向，非全市场口径；参考公司无快照返回 []（对齐 F-16 无法推演时序标注）
- D12 热度漂移 = 首尾快照 cold_hot 差；side 为队名，经 JOIN `match` 匹配主客
- 库内无对应维度的恒值字段（bf.trend/decay_flags/level_gap 等）见 parse 层说明

单元测试一律 Fake 连接；真实联调走 tests/integration（FA_MYSQL_INTEGRATION 门控）。
"""
from fa_orchestrator.data.mysql_analysis_parse import (
    FRIENDLY_KEYWORD,
    bf_block,
    drift_of,
    fmt_time,
    had_block,
    num,
    pair_odds,
    split_intel_items,
    split_score,
    to_day,
    track_changes,
)
from fa_orchestrator.data.mysql_data_source import MySqlDataSource
from fa_orchestrator.data.sql_contracts import (
    ANALYSIS_CONTRACTS,
    POOL_HAD,
    POOL_HHAD,
    SQL_D10_RESULT,
    SQL_D12_HEAT,
    SQL_D9_ASIAN_TRACKS,
    SQL_D9_OVERUNDER_TRACKS,
    SQL_Q10_H2H,
    SQL_Q11_RECENT,
    SQL_Q12_SCHEDULE,
    SQL_Q3_ASIAN,
    SQL_Q3_EUROPE,
    SQL_Q3_KICKOFF,
    SQL_Q3_OVERUNDER,
    SQL_Q5_BF,
    SQL_Q5_VOTE,
    SQL_Q9_INTEL,
    TRACK_BOOKMAKER_NAME,
)


class MySqlAnalysisDataSource(MySqlDataSource):
    """MySQL 分析类数据源：消费解析映射，输出 fixture 对齐的分析数据块。

    状态约束：只读；match_ids（fid → match.id）由 DataAgent 四路径解析结果
    注入（构造参数传入，批次 2 接 resolver）；无映射时分析查询返回 None/空，
    不臆测（与 Q2b 口径一致）。D10 赛果例外：site_match_id == fid 直连。
    """

    def __init__(self, *, match_ids=None, **kwargs):
        """构建分析数据源；match_ids 为四路径解析结果（fid → match.id）。"""
        super().__init__(**kwargs)
        self.match_ids = dict(match_ids or {})

    def _resolve(self, fid):
        """fid → match.id；无映射返回 None（调用方按缺口返回 None/空）。"""
        return self.match_ids.get(str(fid))

    def _fetch(self, sql, params) -> list:
        """单查询样板收敛：连接 → 执行 → 取全行；连接按调用懒建、用毕即关。"""
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(sql, params)
            return list(cur)

    def query_match_odds(self, fid: str):
        """Q3（三盘赔率明细）：初盘/即时盘按公司归并。

        输出：{kickoff_time, euro[], asian[], overunder[]}；无映射返回 None；
        无数据行返回结构完整空集。euro/asian 行含 is_final_snapshot（缺口恒
        False）；overunder 行无该字段（对齐 fixture）。
        """
        match_id = self._resolve(fid)
        if match_id is None:
            return None
        kickoff_rows = self._fetch(SQL_Q3_KICKOFF.strip(), (match_id,))
        kickoff = kickoff_rows[0]["kickoff_time"] if kickoff_rows else None
        return {
            "kickoff_time": fmt_time(kickoff),
            "euro": pair_odds(self._fetch(SQL_Q3_EUROPE.strip(), (match_id,)),
                              ("win", "draw", "lose"), final_flag=True),
            "asian": pair_odds(self._fetch(SQL_Q3_ASIAN.strip(), (match_id,)),
                               ("home_water", "handicap", "away_water"), final_flag=True),
            "overunder": pair_odds(self._fetch(SQL_Q3_OVERUNDER.strip(), (match_id,)),
                                   ("over_odds", "line", "under_odds"), final_flag=False),
        }

    def query_vote(self, fid: str):
        """Q5（投票支持率）：HAD 出支持率与 SP，HHAD 出让球块，必发出方向。

        输出：{support, sp, hhand, bf}；缺池/无映射时对应块 None；
        bf.trend 需多时点快照，库内单时点恒 None（缺口标注）。
        """
        match_id = self._resolve(fid)
        if match_id is None:
            return None
        pool_rows = self._fetch(SQL_Q5_VOTE.strip(), (match_id, POOL_HAD, POOL_HHAD))
        bf_rows = self._fetch(SQL_Q5_BF.strip(), (match_id,))
        latest = {}
        for row in pool_rows:   # SQL 按池 + 快照时间排序，取每池末行（最新）
            latest[row["pool_code"]] = row
        had = latest.get(POOL_HAD)
        hhad = latest.get(POOL_HHAD)
        return {
            "support": had_block(had, "support_rate") if had else None,
            "sp": had_block(had, "odds") if had else None,
            "hhand": ({
                "line": num(hhad["goal_line"]),
                "odds": had_block(hhad, "odds"),
                "support": had_block(hhad, "support_rate"),
            } if hhad else None),
            "bf": bf_block(bf_rows[-1]) if bf_rows else None,
        }

    def query_fundamentals(self, fid: str) -> dict:
        """Q9~Q12（基本面四块）：情报/历史交锋/近期状态/赛程。

        输出结构与 fixture 协议一致：{intelligence, h2h, recent_form, schedule}；
        无映射返回 {}；各子块无数据时为空结构（不臆造维度）。
        """
        match_id = self._resolve(fid)
        if match_id is None:
            return {}
        return {
            "intelligence": {"items": self._intel_items(match_id)},
            "h2h": self._h2h_block(match_id),
            "recent_form": self._recent_form_block(match_id),
            "schedule": self._schedule_block(match_id),
        }

    def _intel_items(self, match_id) -> list:
        """Q9：intelligence_text → items（parse 层按主客标记切段）。"""
        items = []
        for row in self._fetch(SQL_Q9_INTEL.strip(), (match_id,)):
            items.extend(split_intel_items(row["content"]))
        return items

    def _h2h_block(self, match_id) -> dict:
        """Q10：历史交锋记录（按日期倒序）；score '主:客' → '主-客' 对齐 fixture。

        decay_flags（衰减标记）库内无维度，恒 []（缺口标注）。
        """
        records = [
            {"year": row["h2h_date"].year if row["h2h_date"] else None,
             "result": row["result_for_current_home"],
             "score": str(row["score"]).replace(":", "-") if row["score"] else None}
            for row in self._fetch(SQL_Q10_H2H.strip(), (match_id,))
        ]
        return {"records": records, "decay_flags": []}

    def _recent_form_block(self, match_id) -> dict:
        """Q11：双方近期战绩（side 分组）；scored/conceded 按 is_home 视角解析。

        is_friendly 按赛程名含 FRIENDLY_KEYWORD 判定；side 仅取 home/away。
        """
        grouped = {"home": [], "away": []}
        for row in self._fetch(SQL_Q11_RECENT.strip(), (match_id,)):
            if row["side"] not in grouped:
                continue
            score_home, score_away = split_score(row["score"])
            scored, conceded = ((score_home, score_away) if row["is_home"]
                                else (score_away, score_home))
            grouped[row["side"]].append({
                "is_friendly": 1 if FRIENDLY_KEYWORD in str(row["competition"] or "") else 0,
                "scored": scored,
                "conceded": conceded,
            })
        return {"home": {"matches": grouped["home"]}, "away": {"matches": grouped["away"]}}

    def _schedule_block(self, match_id) -> dict:
        """Q12：本场赛事级别 + 双方未来赛程（days 为距本场天数）。

        level_gap（级差）库内无判定数据，恒 False（缺口标注）；
        未来场不含本场自身；本场行缺失时 competition/future 为空/None。
        """
        rows = self._fetch(SQL_Q12_SCHEDULE.strip(), (match_id,) * 5)
        competition = next((r["league"] for r in rows if r["id"] == match_id), None)
        reference = next((to_day(r["match_date"]) for r in rows if r["id"] == match_id), None)
        future = [
            {"days": (to_day(row["match_date"]) - reference).days,
             "competition": row["league"]}
            for row in rows
            if row["id"] != match_id and reference is not None
        ]
        return {"competition": competition, "level_gap": False, "future": future}

    def query_history_tracks(self, fid: str):
        """D9 降级（变盘轨迹）：按参考公司快照序列推演亚盘/大小盘变盘方向。

        change：+1 盘口升 / -1 盘口降；参考公司无快照返回 []（无法推演时序，
        调用方对齐 F-16 缺口标注）；无映射返回 None。
        """
        match_id = self._resolve(fid)
        if match_id is None:
            return None
        tracks = []
        for kind, sql, field in (("asian", SQL_D9_ASIAN_TRACKS, "handicap"),
                                 ("overunder", SQL_D9_OVERUNDER_TRACKS, "line")):
            rows = self._fetch(sql.strip(), (match_id, TRACK_BOOKMAKER_NAME))
            tracks.extend(track_changes(kind, field, rows))
        return tracks

    def query_heat_drift(self, fid: str):
        """D12（热度漂移）：主客侧首尾快照 cold_hot 差（'平局'行不计）。

        side 为队名，经 JOIN `match` 匹配主客；无匹配行该侧 None；
        单快照漂移 0.0；无映射/无快照返回 None。
        """
        match_id = self._resolve(fid)
        if match_id is None:
            return None
        rows = self._fetch(SQL_D12_HEAT.strip(), (match_id,))
        if not rows:
            return None
        home_name, away_name = rows[0]["home"], rows[0]["away"]
        return {"home_pp": drift_of(rows, home_name),
                "away_pp": drift_of(rows, away_name)}

    def query_match_result(self, fid: str) -> dict:
        """D10（赛果）：crawler_match_result.site_match_id == fid 直连（无需解析）。

        输出赛果原始列（比分未回填时 home_score/away_score 为 None 原样返回）；
        had_result/hhad_result 折算属复盘层职责，此处不做；无行返回 None。
        """
        rows = self._fetch(SQL_D10_RESULT.strip(), (str(fid),))
        if not rows:
            return None
        row = rows[0]
        return {
            "site_match_id": str(row["site_match_id"]),
            "match_no": row["match_no"],
            "home_team": row["home_team"],
            "away_team": row["away_team"],
            "league": row["league"],
            "match_time": str(row["match_time"]) if row["match_time"] else None,
            "home_score": row["home_score"],
            "away_score": row["away_score"],
            "half_score": row["half_score"],
            "result_flag": row["result_flag"],
            "play_type": row["play_type"],
        }

    def get_sql_contracts(self) -> dict:
        """暴露分析类 SQL 契约（供联调核对与审计）。"""
        return dict(ANALYSIS_CONTRACTS)
