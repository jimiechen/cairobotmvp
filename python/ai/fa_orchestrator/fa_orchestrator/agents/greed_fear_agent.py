# -*- coding: utf-8 -*-
"""GreedFearAgent：人性贪婪恐惧分析（PRD-12 立项三，skill v7.10 第三部分规则节点化）。

职责边界：
- 负责：市场倾向判定（热门方/弱势方）、贪婪陷阱 4 信号评分（赔率变化/盘口变化/
  水位诱惑/市场背离）、恐惧陷阱 4 信号评分（门槛提升/赔率压制/水位压制/
  基本面恐惧）、陷阱等级（强/中/弱/无）与陷阱组合输出。
- 不负责：反大众前置检查与推荐决策（synthesis_agent.anti_public）、
  情报利弊判读（parse 层语义判读）、矛盾检查（review_agent）、取数。

未产出处理：投票缺失/平局热门/子块缺失时按缺口降级（评分 None + 等级"无数据"），
缺口信号剔除后权重重归一，不编造结论。
"""

# 贪婪信号权重（合计 100，OUTPUT_TEMPLATE 第三部分 4 信号同构）
GREED_WEIGHTS = {"odds_change": 30, "handicap_change": 25,
                 "water_lure": 25, "market_divergence": 20}
# 恐惧信号权重（合计 100）
FEAR_WEIGHTS = {"gate_raise": 30, "odds_suppress": 25,
                "water_suppress": 20, "fundamental_fear": 25}
# 升/降赔家数占比档位（全市场口径，与立项二 market_direction"30 家全升"同源）
RATIO_STRONG = 0.6
RATIO_MID = 0.4
# 水位诱惑档位：热门方水位越低，跟风重仓诱惑越强
WATER_LURE_STRONG = 0.80
WATER_LURE_MID = 0.88
# 水位压制档位：弱势方水位走高（机构不设防/恐吓位）
WATER_SUPPRESS_STRONG = 1.00
WATER_SUPPRESS_MID = 0.95
# 支持率偏差档位（skill 过热口径 ≥+20pp 同源）
DIVERGENCE_STRONG = 20.0
DIVERGENCE_MID = 10.0
# 深盘门槛口径（同 review_agent/goal_layer_agent）
DEEP_LINE_ABS = 1.0
# 基本面恐惧档位：弱势方利空情报条目数（立项二 party 判读产出）
INTEL_FEAR_STRONG = 2
INTEL_FEAR_MID = 1
# 陷阱等级阈值
LEVEL_STRONG = 70
LEVEL_MID = 40
LEVEL_WEAK = 15


class GreedFearAgent:
    """人性贪婪恐惧 Agent：输入热度/基本面/赔率验证/三盘行，输出贪婪恐惧评分与陷阱组合。

    状态约束：无状态；纯函数式规则计算，可并发调用。
    """

    def analyze(self, data: dict) -> dict:
        """产出市场倾向与贪婪/恐惧陷阱结论。

        输入：{"sentiment": SentimentAgent 结论, "fundamental": 基本面结论,
               "odds_verify": 赔率验证结论, "euro": 欧赔行列表, "asian": 亚盘行列表}。
        输出：{"phase","market_tendency","greed","fear","trap_combination","note"}。
        失败行为：投票缺失/平局热门整体降级，子块缺失按缺口剔除权重，不抛异常。
        """
        had = ((data.get("sentiment") or {}).get("had")) or {}
        popular = had.get("popular")
        tendency = {
            "popular": popular,
            "unpopular": self._unpopular(popular),
            "support": had.get("support"),
            "deviation_pp": had.get("deviation_pp"),
        }
        if popular is None:
            note = "投票热度缺失，贪婪恐惧分析降级"
            greed = self._blank()
            fear = self._blank()
        elif popular == "draw":
            note = "支持率热门为平局，市场倾向均势，贪婪恐惧不适用"
            greed = self._blank()
            fear = self._blank()
        else:
            note = ""
            greed = self._side(data, popular, GREED_WEIGHTS, is_greed=True)
            fear = self._side(data, tendency["unpopular"], FEAR_WEIGHTS,
                              is_greed=False)
        return {
            "phase": "greed_fear",
            "market_tendency": tendency,
            "greed": greed,
            "fear": fear,
            "trap_combination": self._combination(greed["level"], fear["level"]),
            "note": note,
        }

    # ---------------------------------------------------------------- 信号计算
    def _side(self, data: dict, direction: str, weights: dict,
              is_greed: bool) -> dict:
        """单侧（贪婪/恐惧）信号评分：逐信号算分 → 缺口重归一加权 → 等级。"""
        if is_greed:
            signals = {
                "odds_change": self._odds_change(data, direction),
                "handicap_change": self._handicap_change(data, direction),
                "water_lure": self._water_lure(data, direction),
                "market_divergence": self._market_divergence(data),
            }
        else:
            signals = {
                "gate_raise": self._gate_raise(data, direction),
                "odds_suppress": self._odds_suppress(data, direction),
                "water_suppress": self._water_suppress(data, direction),
                "fundamental_fear": self._fundamental_fear(data, direction),
            }
        scored = {name: item["score"] for name, item in signals.items()}
        score = self._weighted(scored, weights)
        return {"score": score, "level": self._level(score), "signals": signals}

    @staticmethod
    def _odds_change(data: dict, direction: str) -> dict:
        """贪婪-赔率变化：热门方向全市场降赔家数占比（大众追捧压赔）。"""
        ratio = GreedFearAgent._drop_ratio(data.get("euro") or [], direction)
        if ratio is None:
            return {"score": None, "note": "欧赔初/即盘缺失"}
        if ratio >= RATIO_STRONG:
            return {"score": 100, "note": f"降赔家数占比 {ratio:.0%}"}
        if ratio >= RATIO_MID:
            return {"score": 50, "note": f"降赔家数占比 {ratio:.0%}"}
        return {"score": 0, "note": ""}

    @staticmethod
    def _odds_suppress(data: dict, direction: str) -> dict:
        """恐惧-赔率压制：弱势方向全市场升赔家数占比（市场抛弃）。"""
        ratio = GreedFearAgent._rise_ratio(data.get("euro") or [], direction)
        if ratio is None:
            return {"score": None, "note": "欧赔初/即盘缺失"}
        if ratio >= RATIO_STRONG:
            return {"score": 100, "note": f"升赔家数占比 {ratio:.0%}"}
        if ratio >= RATIO_MID:
            return {"score": 50, "note": f"升赔家数占比 {ratio:.0%}"}
        return {"score": 0, "note": ""}

    def _handicap_change(self, data: dict, direction: str) -> dict:
        """贪婪-盘口变化：亚盘朝热门方向加深（引诱跟风）。"""
        row = self._asian_row(data.get("asian") or [])
        initial = (row or {}).get("initial") or {}
        current = (row or {}).get("current") or {}
        h0, h1 = initial.get("handicap"), current.get("handicap")
        if row is None or h0 is None or h1 is None:
            return {"score": None, "note": "亚盘缺失"}
        deepened = (direction == "win" and h1 < h0) or \
                   (direction == "lose" and h1 > h0)
        return {"score": 100 if deepened else 0, "note": ""}

    def _gate_raise(self, data: dict, direction: str) -> dict:
        """恐惧-门槛提升：深盘朝热门方向（direction 为弱势方，弱势方介入门槛加深）。"""
        row = self._asian_row(data.get("asian") or [])
        current = ((row or {}).get("current") or {}).get("handicap")
        if row is None or current is None:
            return {"score": None, "note": "亚盘缺失"}
        # direction 为弱势方：热门方让深（弱势方=lose → 主让深；=win → 客让深）
        deep_toward_popular = (direction == "lose" and current <= -DEEP_LINE_ABS) \
            or (direction == "win" and current >= DEEP_LINE_ABS)
        return {"score": 100 if deep_toward_popular else 0, "note": ""}

    def _water_lure(self, data: dict, direction: str) -> dict:
        """贪婪-水位诱惑：热门方即时水位被压低（低水引诱重仓跟买）。"""
        water = self._side_water(data.get("asian") or [], direction)
        if water is None:
            return {"score": None, "note": "亚盘缺失"}
        if water <= WATER_LURE_STRONG:
            return {"score": 100, "note": f"热门方水位 {water}"}
        if water <= WATER_LURE_MID:
            return {"score": 50, "note": f"热门方水位 {water}"}
        return {"score": 0, "note": ""}

    def _water_suppress(self, data: dict, direction: str) -> dict:
        """恐惧-水位压制：弱势方即时水位走高（恐吓位/不设防）。"""
        water = self._side_water(data.get("asian") or [], direction)
        if water is None:
            return {"score": None, "note": "亚盘缺失"}
        if water >= WATER_SUPPRESS_STRONG:
            return {"score": 100, "note": f"弱势方水位 {water}"}
        if water >= WATER_SUPPRESS_MID:
            return {"score": 50, "note": f"弱势方水位 {water}"}
        return {"score": 0, "note": ""}

    @staticmethod
    def _market_divergence(data: dict) -> dict:
        """贪婪-市场背离：支持率偏差过热（skill ≥+20pp 口径）/两层背离。"""
        sentiment = data.get("sentiment") or {}
        deviation = ((sentiment.get("had") or {}).get("deviation_pp"))
        two_layer = bool(((data.get("odds_verify") or {}).get("euro_asia") or {})
                         .get("two_layer_divergence"))
        if deviation is None and not two_layer:
            return {"score": None, "note": "热度偏差缺失"}
        if deviation is not None and deviation >= DIVERGENCE_STRONG:
            return {"score": 100, "note": f"支持率偏差 +{deviation}pp"}
        if (deviation is not None and deviation >= DIVERGENCE_MID) or two_layer:
            return {"score": 50, "note": "两层背离" if two_layer
                    else f"支持率偏差 +{deviation}pp"}
        return {"score": 0, "note": ""}

    @staticmethod
    def _fundamental_fear(data: dict, direction: str) -> dict:
        """恐惧-基本面恐惧：弱势方利空情报条目数（立项二 party 判读产出）。

        intelligence 计数键为 home/away 维度，win/lose 方向需映射（win=主/lose=客）。
        """
        intel = (data.get("fundamental") or {}).get("intelligence") or {}
        side = "home" if direction == "win" else "away"
        count = intel.get(f"{side}_against")
        if count is None:
            return {"score": None, "note": "情报利弊计数缺失"}
        if count >= INTEL_FEAR_STRONG:
            return {"score": 100, "note": f"弱势方利空 {count} 条"}
        if count >= INTEL_FEAR_MID:
            return {"score": 50, "note": f"弱势方利空 {count} 条"}
        return {"score": 0, "note": ""}

    # ---------------------------------------------------------------- 汇总与工具
    @staticmethod
    def _weighted(scored: dict, weights: dict):
        """缺口重归一加权：评分 None 的信号剔除，剩余权重归一后加权求和。"""
        available = {name: s for name, s in scored.items() if s is not None}
        if not available:
            return None
        total_weight = sum(weights[name] for name in available)
        total = sum(s * weights[name] for name, s in available.items())
        return round(total / total_weight, 1)

    @staticmethod
    def _level(score):
        """陷阱等级：≥70 强 / ≥40 中 / ≥15 弱 / <15 无 / 全缺口"无数据"。"""
        if score is None:
            return "无数据"
        if score >= LEVEL_STRONG:
            return "强"
        if score >= LEVEL_MID:
            return "中"
        if score >= LEVEL_WEAK:
            return "弱"
        return "无"

    @staticmethod
    def _combination(greed_level: str, fear_level: str) -> str:
        """陷阱组合：双侧均"无数据"时输出"无数据"，否则"X贪婪+Y恐惧"。"""
        if greed_level == "无数据" and fear_level == "无数据":
            return "无数据"
        return f"{greed_level}贪婪+{fear_level}恐惧"

    @staticmethod
    def _unpopular(popular):
        """弱势方判定：热门的反方向；平局热门/缺失时为 None。"""
        if popular == "win":
            return "lose"
        if popular == "lose":
            return "win"
        return None

    @staticmethod
    def _asian_row(asian: list):
        """澳门行优先取亚盘行（对齐 runner _numeric_handicap 模式），无则取首行。"""
        return next((r for r in asian if r.get("bk") == "澳门"),
                    asian[0] if asian else None)

    @staticmethod
    def _side_water(asian: list, direction: str):
        """方向对应的水位：win → 主水位，lose → 客水位。"""
        row = GreedFearAgent._asian_row(asian)
        if row is None:
            return None
        key = "home_water" if direction == "win" else "away_water"
        return (row.get("current") or {}).get(key)

    @staticmethod
    def _direction_ratios(euro: list, direction: str) -> dict | None:
        """单方向初→即升降家数占比：{drop, rise}；无可比公司返回 None。"""
        drop = rise = total = 0
        for row in euro:
            initial = (row.get("initial") or {}).get(direction)
            current = (row.get("current") or {}).get(direction)
            if initial is None or current is None:
                continue
            total += 1
            if current < initial:
                drop += 1
            elif current > initial:
                rise += 1
        if total == 0:
            return None
        return {"drop": drop / total, "rise": rise / total}

    @classmethod
    def _drop_ratio(cls, euro: list, direction: str):
        return (cls._direction_ratios(euro, direction) or {}).get("drop")

    @classmethod
    def _rise_ratio(cls, euro: list, direction: str):
        return (cls._direction_ratios(euro, direction) or {}).get("rise")

    @staticmethod
    def _blank() -> dict:
        """整体降级空结构：评分/等级/信号全缺口。"""
        return {"score": None, "level": "无数据", "signals": {}}
