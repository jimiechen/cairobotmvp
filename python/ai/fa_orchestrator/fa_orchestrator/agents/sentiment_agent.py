# -*- coding: utf-8 -*-
"""SentimentAgent：投票热度偏差与过热判定（PRD-12 F-12，ISS-FA3）。

职责边界：
- 负责：HAD/HHAD 支持率偏差计算（支持率 − 对应盘口去水隐含概率）、过热判定
  （胜平负 ≥+20pp / 让球 ≥+15pp，SKILL v7.9 两档阈值）、必发成交形态输入记录、
  过热两档处置（有独立支撑 → 降星保方向；无支撑 → 主推放弃，009 案例）。
- 不负责：撤离信号门槛与推荐结构判定（synthesis_agent.py）、矛盾检查
  （review_agent.py）、数据取数（data_source.py）。

数据缺失处理：投票/让球投票/必发任一缺失时明确标注降级，不编造结论。
"""
from fa_orchestrator.analysis.odds_math import implied_probabilities

# 过热两档阈值（SKILL v7.9）：支持率偏差单位为百分点（pp）
OVERHEAT_HAD_PP = 20.0
OVERHEAT_HHAND_PP = 15.0
# "有独立支撑"判定：同向利多情报 ≥2 条（009 案例），或交锋/状态层同向佐证
INTEL_SUPPORT_MIN = 2
# 热门方向 → 场地方向映射（win=主胜方向 / lose=客胜方向；平局方向无确定性利多口径）
POPULAR_SIDE = {"win": "home", "lose": "away", "draw": None}
# 状态层同向佐证口径：热门为主胜方向且主队场均进球 ≥1.5（客胜向佐证数据缺，不编造）
RECENT_ALIGNED_MIN_GOALS = 1.5
# 交锋记录赛果口径（主队视角，与 FA2 draw_support 的平局 "D" 同源）
H2H_RESULT_HOME_WIN = "W"
H2H_RESULT_AWAY_WIN = "L"
# 交锋同向佐证占比门槛：该方向赛果占比过半
H2H_ALIGNED_RATE = 0.5
# ---- 立项四收口常量（口径经主控确认 2026-10-14，数值待主控后续调整）----
# 庄家 pnl 赔付风险阈值：pnl 低于此值视为庄家赔付风险（大众方向赢面倾向信号）
BOOKMAKER_PNL_RISK_THRESHOLD = 0.0
# 立项四扩展：heat pnl 三方向标签（键与 bf.heat_pnl 的方向键对齐）
HEAT_PNL_SIDE_LABELS = {"home": "主", "draw": "平", "away": "客"}
# 价位聚集度三档：top 档成交占比达高档为高聚集，达中档为中等，低于中档为分散
LADDER_CONCENTRATED_RATIO = 0.6
LADDER_MODERATE_RATIO = 0.4


class SentimentAgent:
    """热度分析 Agent：输入投票块与欧赔，输出偏差/过热/必发形态结论。

    状态约束：无状态；输入数据缺失时输出降级标注而非异常。
    """

    def analyze(self, data: dict) -> dict:
        """计算单场热度结论。

        输入：{"vote": 竞彩投票块(含 support/sp/hhand/bf) | None,
               "euro": 欧赔行列表, "fundamental": 第一阶段基本面结论}。
        输出：{"phase","had","hhand","bf","overheat","note"}。
        失败行为：投票缺失整体降级标注；子块缺失分别标注，不抛异常。
        """
        vote = data.get("vote")
        fundamental = data.get("fundamental") or {}
        had = self._analyze_had(vote, data.get("euro") or [])
        hhand = self._analyze_hhand(vote)
        bf = self._analyze_bf(vote)
        overheat = self._judge_overheat(had, hhand, fundamental)
        note = "" if vote else "ℹ️ 投票热度数据缺失，热度验证降级为赔率结构侧信息"
        return {
            "phase": "sentiment",
            "had": had,
            "hhand": hhand,
            "bf": bf,
            "overheat": overheat,
            "note": note,
        }

    @staticmethod
    def _analyze_had(vote: dict | None, euro: list) -> dict:
        """HAD 支持率偏差：支持率 − 欧赔去水隐含概率（百分点）。"""
        support = (vote or {}).get("support") or {}
        if not support:
            return {"popular": None, "support": None, "implied": None,
                    "deviation_pp": None, "note": "热度数据缺失"}
        popular = max(support, key=support.get)
        implied = implied_probabilities(euro)
        if popular not in implied:
            return {"popular": popular, "support": support[popular], "implied": None,
                    "deviation_pp": None, "note": "欧赔缺失，无法计算偏差"}
        deviation = round(support[popular] - implied[popular], 1)
        return {"popular": popular, "support": support[popular],
                "implied": implied[popular], "deviation_pp": deviation, "note": ""}

    @staticmethod
    def _analyze_hhand(vote: dict | None) -> dict:
        """HHAD 让球支持率偏差：让球支持率 − 竞彩让球赔率去水隐含概率。"""
        hhand = (vote or {}).get("hhand") or {}
        support = hhand.get("support") or {}
        if not support:
            return {"available": False, "popular": None, "deviation_pp": None,
                    "note": "让球热度数据缺失，热度验证降级为胜平负热度映射"}
        popular = max(support, key=support.get)
        implied = implied_probabilities(
            [{"current": hhand.get("odds") or {}}] if hhand.get("odds") else []
        )
        deviation = round(support[popular] - implied[popular], 1) if popular in implied else None
        return {"available": True, "popular": popular, "deviation_pp": deviation, "note": ""}

    @staticmethod
    def _analyze_bf(vote: dict | None) -> dict:
        """必发成交形态输入（F-12）：记录方向与回落形态，缺失时标注降级。

        立项四扩展（口径经主控确认 2026-10-14）：bookmaker_pnl 交叉验证
        标注——pnl < BOOKMAKER_PNL_RISK_THRESHOLD 视为庄家赔付风险（大众
        方向赢面倾向信号），否则庄家风险可控；价位聚集三档标注——
        top_ratio 达高档阈值高聚集、中档阈值中等、低于中档分散（大众观点
        不一致）。二者仅进 summary/pnl_note/ladder_note 标注，不参与
        加减分与推荐结构（skill 未定义分值，最小闭环）；数据结构为契约
        假设，真值校准待 touzhu 样例。
        """
        bf = (vote or {}).get("bf") or {}
        if not bf:
            return {"available": False, "trend": None, "bookmaker_pnl": None,
                    "ladder": None, "pnl_note": None, "ladder_note": None,
                    "heat_pnl_note": None,
                    "summary": "必发成交数据缺失，形态输入降级"}
        bookmaker_pnl = bf.get("bookmaker_pnl")
        ladder = bf.get("ladder")
        summary = f"必发成交方向 {bf.get('direction')}，形态 {bf.get('trend')}"
        pnl_note = None
        if bookmaker_pnl is not None:
            if bookmaker_pnl < BOOKMAKER_PNL_RISK_THRESHOLD:
                pnl_note = (f"庄家pnl {bookmaker_pnl}，庄家赔付风险"
                            f"（大众方向赢面倾向信号）")
            else:
                pnl_note = f"庄家pnl {bookmaker_pnl}，庄家风险可控"
            summary += f"，{pnl_note}"
        ladder_note = None
        if ladder and ladder.get("top_ratio") is not None:
            ratio = ladder["top_ratio"]
            if ratio >= LADDER_CONCENTRATED_RATIO:
                ladder_note = f"价位聚集度 {ratio}（高聚集）"
            elif ratio >= LADDER_MODERATE_RATIO:
                ladder_note = f"价位聚集度 {ratio}（中等）"
            else:
                ladder_note = f"价位聚集度 {ratio}（分散，大众观点不一致）"
            summary += f"，{ladder_note}"
        heat_note = SentimentAgent._heat_pnl_note(bf.get("heat_pnl"))
        if heat_note:
            summary += f"，{heat_note}"
        return {"available": True, "trend": bf.get("trend"),
                "bookmaker_pnl": bookmaker_pnl, "ladder": ladder,
                "heat_pnl": bf.get("heat_pnl"),
                "pnl_note": pnl_note, "ladder_note": ladder_note,
                "heat_pnl_note": heat_note,
                "summary": summary}

    @staticmethod
    def _heat_pnl_note(heat_pnl: dict | None) -> str | None:
        """heat pnl 三方向标注（立项四扩展，宿主答复 2026-10-14）。

        输入：bf.heat_pnl（{"home","draw","away","snapshot_time"}，500com
        投注倾向庄家盈亏，正=庄家赢）| None。
        输出：三方向解读字符串（口径同 pnl_note 体系：负=庄家赔付=大众该
        方向赢面倾向信号；正=庄家赢=大众该方向输面倾向），来源标注
        "500com"与必发 pnl 区分；None 输入返回 None。
        约束：仅输入记录标注，不参与加减分与推荐结构（最小闭环）。
        """
        if not heat_pnl:
            return None
        parts = []
        for key, label in HEAT_PNL_SIDE_LABELS.items():
            pnl = heat_pnl.get(key)
            if pnl is None:
                parts.append(f"{label} 缺失")
                continue
            if pnl < BOOKMAKER_PNL_RISK_THRESHOLD:
                parts.append(f"{label} {pnl}（大众该方向赢面倾向信号）")
            else:
                parts.append(f"{label} {pnl}（庄家赢）")
        return f"庄家盈亏(500com投注倾向,正=庄家赢)：{' / '.join(parts)}"

    def _judge_overheat(self, had: dict, hhand: dict, fundamental: dict) -> dict:
        """过热判定与两档处置（SKILL v7.9）：胜平负 ≥+20pp / 让球 ≥+15pp。"""
        layer, deviation = None, None
        if had.get("deviation_pp") is not None and had["deviation_pp"] >= OVERHEAT_HAD_PP:
            layer, deviation = "had", had["deviation_pp"]
        elif hhand.get("deviation_pp") is not None and hhand["deviation_pp"] >= OVERHEAT_HHAND_PP:
            layer, deviation = "hhand", hhand["deviation_pp"]
        if layer is None:
            return {"triggered": False, "layer": None, "deviation_pp": None,
                    "support_intel": None, "disposition": None, "tag": None}
        popular = had["popular"] if layer == "had" else hhand["popular"]
        support_intel = self._has_independent_support(popular, fundamental)
        disposition = "downgrade_keep" if support_intel else "abandon"
        support_label = "有" if support_intel else "无"
        tag = f"[过热+{deviation}pp｜情报支撑：{support_label}]"
        return {"triggered": True, "layer": layer, "deviation_pp": deviation,
                "support_intel": support_intel, "disposition": disposition, "tag": tag}

    def _has_independent_support(self, popular: str | None, fundamental: dict) -> bool:
        """独立支撑三路径：同向利多情报 ≥INTEL_SUPPORT_MIN / 交锋 / 状态层同向佐证。"""
        side = POPULAR_SIDE.get(popular)
        if side is None:
            return False
        intel = fundamental.get("intelligence") or {}
        items = intel.get("items")
        if items is not None:
            favor_count = sum(
                1 for item in items
                if item.get("party") == side and item.get("sentiment") == "favor"
            )
        else:
            # FA1 结论仅含五向计数（无原始 items）时按 {side}_favor 回退
            favor_count = intel.get(f"{side}_favor", 0)
        if favor_count >= INTEL_SUPPORT_MIN:
            return True
        if self._is_h2h_aligned(side, fundamental):
            return True
        return self._is_recent_aligned(popular, fundamental)

    @staticmethod
    def _is_h2h_aligned(side: str, fundamental: dict) -> bool:
        """交锋同向佐证：近期交锋中该方向赛果（W/D/L 主队视角）占比过半。"""
        records = (fundamental.get("h2h") or {}).get("records") or []
        if not records:
            return False
        target = H2H_RESULT_HOME_WIN if side == "home" else H2H_RESULT_AWAY_WIN
        aligned = sum(1 for record in records if record.get("result") == target)
        return aligned / len(records) >= H2H_ALIGNED_RATE

    @staticmethod
    def _is_recent_aligned(popular: str, fundamental: dict) -> bool:
        """状态层同向佐证：热门为主胜且主队场均进球达标；其余方向数据缺不判定。"""
        if popular != "win":
            return False
        avg_scored = (fundamental.get("recent_form") or {}).get("home_avg_scored")
        return avg_scored is not None and avg_scored >= RECENT_ALIGNED_MIN_GOALS
