-- 立项四：庄家 pnl 数据源扩容 —— 前向 DDL（宿主侧执行，执行前需主控确认）
-- 范围：touzhu 文件数据（庄家 pnl 指数、必发价位聚集）→ MySQL 表
-- 依据：docs/reports/立项评审-FA分析能力补齐-20261013.md 立项四；
--       SKILL.md data.touzhu 契约 {bf_trading:{total_volume,currency,summary,bookmaker_pnl}, bf_large, bf_details}
--
-- ⚠️ 契约假设标注（宿主未提供 touzhu 文件样例，2026-10-13 主控确认"契约假设先行"）：
--   1. bookmaker_pnl 为数值型庄家 pnl 指数（DECIMAL(14,2)，可负）
--   2. bf_details 为价位档成交列表，每档 {price, back_volume, lay_volume}
--   样例到手后如结构不符，仅需调整 ingest_touzhu 解析层与本表列语义，表名/主键结构不变。
--
-- 回滚：见同目录 down.sql

-- 1. 庄家 pnl 指数：betfair_trading 加列（与 total_volume 同快照行内聚，
--    复用 uk_bf 唯一键，不新建表；NULL 兼容宿主未回填的历史行）
ALTER TABLE `betfair_trading`
  ADD COLUMN `bookmaker_pnl` decimal(14,2) DEFAULT NULL
    COMMENT '庄家pnl指数（touzhu.bf_trading.bookmaker_pnl，立项四；NULL=未回填）'
    AFTER `away_pct`;

-- 2. 必发价位聚集（bf_details）：新表（一档一行，多行结构无法内聚到快照表）
CREATE TABLE `betfair_price_ladder` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `match_id` int(11) NOT NULL COMMENT '比赛id（对齐 betfair_trading.match_id 域）',
  `trade_type` varchar(8) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'standard' COMMENT '类型: standard标准盘/large大额（对齐 betfair_trading 口径）',
  `price` decimal(10,2) NOT NULL COMMENT '价位档',
  `back_volume` decimal(14,2) DEFAULT NULL COMMENT '该价位成交（back 侧，立项四）',
  `lay_volume` decimal(14,2) DEFAULT NULL COMMENT '该价位成交（lay 侧，立项四）',
  `snapshot_time` datetime NOT NULL COMMENT '抓取快照时间（对齐 betfair_trading.snapshot_time）',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_bpl` (`match_id`,`trade_type`,`price`,`snapshot_time`),
  KEY `idx_bpl_match` (`match_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='必发价位聚集快照（touzhu.bf_details，立项四）';
