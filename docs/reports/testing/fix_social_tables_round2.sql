-- ============================================
-- Social 域表结构修复 SQL (Round 2)
-- 目标：对齐 Go Model 定义的完整字段
-- 执行环境：MySQL 192.168.1.6:3306 go_biz
-- 前置：Round 1 已添加 users.uid + member_stats.id
-- ============================================

-- member_blocks: Model 使用 blocker_id/blocked_id，DB 使用 user_id/blocked_user_id
-- 方案：添加新列并建索引（保留原列兼容）
ALTER TABLE `member_blocks`
ADD COLUMN `blocker_id` CHAR(32) DEFAULT NULL AFTER `id`,
ADD COLUMN `blocked_id` CHAR(32) DEFAULT NULL AFTER `blocker_id`,
ADD INDEX `idx_blocker_id` (`blocker_id`),
ADD INDEX `idx_blocked_id` (`blocked_id`);

-- member_stats: Model 使用 replies_count/likes_received/groups_joined
-- DB 现有: topics_count/followers_count/following_count
-- 方案：添加 Model 需要的列
ALTER TABLE `member_stats`
ADD COLUMN `replies_count` INT NOT NULL DEFAULT 0 AFTER `topics_count`,
ADD COLUMN `likes_received` INT NOT NULL DEFAULT 0 AFTER `replies_count`,
ADD COLUMN `groups_joined` INT NOT NULL DEFAULT 0 AFTER `likes_received`;
