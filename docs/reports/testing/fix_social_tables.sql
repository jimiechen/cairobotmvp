-- ============================================
-- Social 域表结构修复 SQL
-- 目标：使 go_biz 库表结构与 Go Model 一致
-- 执行环境：MySQL 192.168.1.6:3306 go_biz
-- ============================================

-- 1. users 表：添加 uid 列（9位数字对外编号）
ALTER TABLE `users`
ADD COLUMN `uid` VARCHAR(20) DEFAULT NULL AFTER `id`,
ADD UNIQUE INDEX `idx_uid` (`uid`);

-- 2. member_stats 表：添加 id 主键列（char(32) 全局唯一标识符）
ALTER TABLE `member_stats`
ADD COLUMN `id` CHAR(32) NOT NULL DEFAULT '' FIRST,
ADD PRIMARY KEY (`id`);

-- 验证
-- SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='go_biz' AND TABLE_NAME='users' ORDER BY ORDINAL_POSITION;
-- SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='go_biz' AND TABLE_NAME='member_stats' ORDER BY ORDINAL_POSITION;
