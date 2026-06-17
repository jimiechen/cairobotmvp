-- ============================================================
-- 回滚脚本: social-migration-001-initial-rollback.sql
-- 用途: 回滚 social-migration-001-initial.sql 创建的 10 张表
-- 版本: v1.0
-- 日期: 2026-06-16
-- 关联迁移脚本: social-migration-001-initial.sql
--
-- 删除顺序（与建表顺序相反，先删依赖表再删被依赖表）:
--   Step 1: audit_logs（Batch 4，无外键依赖其他社交表）
--   Step 2: reply_likes, topic_favorites, topic_likes, topic_reads（Batch 3）
--   Step 3: group_pay_configs, group_members（Batch 2）
--   Step 4: topic_replies, topics, groups（Batch 1）
-- ============================================================

SET FOREIGN_KEY_CHECKS = 0;

-- ---- Batch 4 回滚: 审计日志 ----
DROP TABLE IF EXISTS audit_logs;

-- ---- Batch 3 回滚: 交互表（依赖 topics / topic_replies / users）----
DROP TABLE IF EXISTS reply_likes;
DROP TABLE IF EXISTS topic_favorites;
DROP TABLE IF EXISTS topic_likes;
DROP TABLE IF EXISTS topic_reads;

-- ---- Batch 2 回滚: 群组成员与付费配置（依赖 groups / users）----
DROP TABLE IF EXISTS group_pay_configs;
DROP TABLE IF EXISTS group_members;

-- ---- Batch 1 回滚: 基础表（topic_replies 依赖 topics，topics 可选依赖 groups）----
DROP TABLE IF EXISTS topic_replies;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS groups;

SET FOREIGN_KEY_CHECKS = 1;
