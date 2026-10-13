-- 立项四：庄家 pnl 数据源扩容 —— 回滚 DDL（与 up.sql 配对，先验证依赖再执行）
-- 幂等性：MySQL 5.7 不支持 IF EXISTS 的 DROP COLUMN，5.7 环境需人工确认列存在；
-- 8.0+ 可手工替换为 ALTER TABLE ... DROP COLUMN IF EXISTS。

-- 1. 回滚价位聚集表（含真库数据确认：DROP 前先 COUNT 备份再删）
DROP TABLE IF EXISTS `betfair_price_ladder`;

-- 2. 回滚庄家 pnl 列
ALTER TABLE `betfair_trading` DROP COLUMN `bookmaker_pnl`;
