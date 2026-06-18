-- ============================================================
-- 迁移脚本: 004_social_domain_tables.sql
-- 用途: Social 域 MVP 全量建表（16 表）
-- 来源: PRD-social-app-mvp.md 数据模型章节
-- 目标数据库: MySQL 8.0+ / InnoDB
-- 数据库: go_biz
-- 版本: v1.0
-- 日期: 2026-06-17
--
-- 枚举约束:
--   svc 层必须使用 protobuf 枚举常量或包级常量
--   禁止在代码中硬编码裸数字或魔法字符串（参见 coding.md §6）
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

USE go_biz;

-- ============================================================
-- 1. users — 用户身份主表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` char(32) NOT NULL COMMENT '内部主键，全系统关联',
  `username` varchar(50) NOT NULL COMMENT '登录用户名',
  `password` varchar(255) NOT NULL COMMENT '加密密码',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像 URL',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '用户状态：1=active 2=inactive 3=banned 4=deleted',
  `membership_level` varchar(32) NOT NULL DEFAULT 'normal' COMMENT '平台会员等级',
  `last_login_at` bigint(20) DEFAULT NULL COMMENT '最近登录时间',
  `last_login_ip` varchar(64) DEFAULT NULL COMMENT '最近登录 IP',
  `login_count` bigint(20) NOT NULL DEFAULT '0' COMMENT '登录次数',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户身份主表（1级数据）';

-- ============================================================
-- 2. groups — 群组/圈子主表（1级数据 + 部分2级快照字段）
-- ============================================================
DROP TABLE IF EXISTS `groups`;
CREATE TABLE `groups` (
  `id` char(32) NOT NULL COMMENT '群组主键',
  `name` varchar(100) NOT NULL COMMENT '群组名称',
  `slug` varchar(100) NOT NULL COMMENT 'URL 标识',
  `owner_id` char(32) NOT NULL COMMENT '圈主用户 ID',
  `description` varchar(1000) NOT NULL DEFAULT '' COMMENT '群组描述',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '群组头像',
  `cover_image` varchar(500) NOT NULL DEFAULT '' COMMENT '群组封面图',
  `category` varchar(64) NOT NULL DEFAULT '' COMMENT '群组分类',
  `tags` json DEFAULT NULL COMMENT '标签列表 JSON',
  `type` varchar(20) NOT NULL DEFAULT 'free' COMMENT '群组类型：free/paid/mixed/invite',
  `visibility` tinyint(4) NOT NULL DEFAULT '1' COMMENT '可见性：1=公开 2=链接可见 3=私密',
  `join_mode` tinyint(4) NOT NULL DEFAULT '1' COMMENT '加入方式：1=直接 2=审核 3=付费 4=邀请',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '群组状态：1=active 2=inactive 3=pending 4=rejected 5=deleted',
  `members_count` bigint(20) NOT NULL DEFAULT '0' COMMENT '成员数冗余快照（2级）',
  `topics_count` bigint(20) NOT NULL DEFAULT '0' COMMENT '帖子数冗余快照（2级）',
  `max_members` bigint(20) NOT NULL DEFAULT '0' COMMENT '人数上限，0表示不限制',
  `rules` text COMMENT '群组规则',
  `welcome_message` varchar(1000) NOT NULL DEFAULT '' COMMENT '欢迎语',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_owner_status` (`owner_id`, `status`),
  KEY `idx_type_status` (`type`, `status`),
  KEY `idx_visibility_status` (`visibility`, `status`),
  KEY `idx_category_status` (`category`, `status`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_groups_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组/圈子主表（1级数据）';

-- ============================================================
-- 3. topics — 帖子主表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `topics`;
CREATE TABLE `topics` (
  `id` char(32) NOT NULL COMMENT '帖子主键',
  `title` varchar(200) NOT NULL COMMENT '标题',
  `content` longtext NOT NULL COMMENT '正文内容',
  `summary` varchar(500) NOT NULL DEFAULT '' COMMENT '摘要，无权限时返回',
  `author_id` char(32) NOT NULL COMMENT '作者 ID',
  `group_id` char(32) NOT NULL COMMENT '所属群组 ID',
  `type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子类型：1=normal 2=article 3=notice 4=qa',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态：1=draft 2=pending 3=published 4=rejected 5=banned 6=deleted',
  `visibility` tinyint(4) NOT NULL DEFAULT '1' COMMENT '可见性：1=PUBLIC 2=GROUP_MEMBER 3=PAID_MEMBER 4=OWNER_ONLY',
  `is_pinned` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否置顶',
  `is_featured` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否精选',
  `allow_comments` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否允许评论',
  `tags` json DEFAULT NULL COMMENT '帖子标签 JSON',
  `published_at` bigint(20) DEFAULT NULL COMMENT '发布时间',
  `last_activity_at` bigint(20) DEFAULT NULL COMMENT '最后活跃时间（2级派生）',
  `cover_image` varchar(500) NOT NULL DEFAULT '' COMMENT '封面图',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_status_published` (`group_id`, `status`, `published_at`),
  KEY `idx_author_status_created` (`author_id`, `status`, `created_at`),
  KEY `idx_visibility_status` (`visibility`, `status`),
  KEY `idx_group_pinned` (`group_id`, `is_pinned`, `published_at`),
  KEY `idx_group_featured` (`group_id`, `is_featured`, `published_at`),
  KEY `idx_last_activity` (`last_activity_at`),
  CONSTRAINT `fk_topics_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_topics_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子主表（1级数据）';

-- ============================================================
-- 4. group_members — 群组成员关系表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `group_members`;
CREATE TABLE `group_members` (
  `id` char(32) NOT NULL COMMENT '主键 UUID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `role` tinyint(4) NOT NULL DEFAULT '3' COMMENT '成员角色：1=owner 2=admin 3=member 4=moderator',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '成员状态：1=active 2=pending 3=banned 4=left 5=expired',
  `join_source` varchar(20) NOT NULL DEFAULT 'free' COMMENT '加入来源：free/paid/invite/admin/import/follow',
  `joined_at` bigint(20) DEFAULT NULL COMMENT '加入时间',
  `expired_at` bigint(20) DEFAULT NULL COMMENT '付费权益过期时间',
  `muted_until` bigint(20) DEFAULT NULL COMMENT '禁言截止时间戳',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_group_status_role` (`group_id`, `status`, `role`),
  KEY `idx_group_expired` (`group_id`, `expired_at`),
  KEY `idx_user_joined_at` (`user_id`, `joined_at`),
  CONSTRAINT `fk_group_members_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_group_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员关系表（1级数据）';

-- ============================================================
-- 5. group_plans — 付费群组方案表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `group_plans`;
CREATE TABLE `group_plans` (
  `id` char(32) NOT NULL COMMENT '方案主键',
  `group_id` char(32) NOT NULL COMMENT '所属群组',
  `name` varchar(100) NOT NULL COMMENT '方案名称：月卡/季卡/年卡',
  `plan_type` varchar(20) NOT NULL COMMENT '方案类型：monthly/quarterly/yearly/lifetime',
  `price_cent` bigint(20) NOT NULL DEFAULT '0' COMMENT '价格，单位：分',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `duration_days` int(11) NOT NULL DEFAULT '30' COMMENT '权益天数',
  `benefits` json DEFAULT NULL COMMENT '权益说明 JSON',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '方案状态：1=上架 2=下架',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_status` (`group_id`, `status`),
  KEY `idx_plan_type_status` (`plan_type`, `status`),
  CONSTRAINT `fk_group_plans_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='付费群组方案表（1级数据）';

-- ============================================================
-- 6. group_orders — 群组付费订单表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `group_orders`;
CREATE TABLE `group_orders` (
  `id` char(32) NOT NULL COMMENT '订单主键',
  `order_no` varchar(64) NOT NULL COMMENT '对外订单号',
  `user_id` char(32) NOT NULL COMMENT '购买用户',
  `group_id` char(32) NOT NULL COMMENT '购买群组',
  `plan_id` char(32) DEFAULT NULL COMMENT '购买方案',
  `amount_cent` bigint(20) NOT NULL DEFAULT '0' COMMENT '实付金额，单位：分',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '订单状态：1=pending 2=paid 3=cancelled 4=refunded 5=failed',
  `pay_channel` varchar(30) DEFAULT NULL COMMENT '支付渠道',
  `paid_at` bigint(20) DEFAULT NULL COMMENT '支付时间',
  `expired_at` bigint(20) DEFAULT NULL COMMENT '权益过期时间',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status_time` (`user_id`, `status`, `created_at`),
  KEY `idx_group_status_time` (`group_id`, `status`, `created_at`),
  KEY `idx_plan_status` (`plan_id`, `status`),
  CONSTRAINT `fk_group_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_group_orders_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_group_orders_plan` FOREIGN KEY (`plan_id`) REFERENCES `group_plans` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组付费订单表（1级数据）';

-- ============================================================
-- 7. topic_read_records — 帖子阅读记录表（2级行为数据）
-- ============================================================
DROP TABLE IF EXISTS `topic_read_records`;
CREATE TABLE `topic_read_records` (
  `id` char(32) NOT NULL COMMENT '阅读记录主键',
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `user_id` char(32) NOT NULL COMMENT '阅读用户',
  `group_id` char(32) DEFAULT NULL COMMENT '冗余群组 ID',
  `read_at` bigint(20) DEFAULT NULL COMMENT '最近阅读时间',
  `read_count` bigint(20) NOT NULL DEFAULT '0' COMMENT '阅读次数',
  `duration_sec` int(11) NOT NULL DEFAULT '0' COMMENT '累计阅读时长，单位：秒',
  `progress` int(11) NOT NULL DEFAULT '0' COMMENT '阅读进度，0-100',
  `created_at` bigint(20) NOT NULL COMMENT '首次阅读时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_topic` (`user_id`, `topic_id`),
  KEY `idx_topic_read_time` (`topic_id`, `read_at`),
  KEY `idx_user_read_time` (`user_id`, `read_at`),
  KEY `idx_group_read_time` (`group_id`, `read_at`),
  CONSTRAINT `fk_topic_read_records_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`),
  CONSTRAINT `fk_topic_read_records_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_topic_read_records_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子阅读记录表（2级行为数据）';

-- ============================================================
-- 8. topic_comments — 帖子评论表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `topic_comments`;
CREATE TABLE `topic_comments` (
  `id` char(32) NOT NULL COMMENT '评论主键',
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `user_id` char(32) NOT NULL COMMENT '评论用户',
  `parent_id` char(32) DEFAULT NULL COMMENT '父评论 ID，支持楼中楼',
  `content` text NOT NULL COMMENT '评论内容',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '评论状态：1=正常 2=审核中 3=隐藏 4=删除',
  `is_pinned` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否置顶',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_topic_status` (`topic_id`, `status`, `created_at`),
  KEY `idx_user_comments` (`user_id`, `created_at`),
  KEY `idx_group_status_time` (`group_id`, `status`, `created_at`),
  KEY `idx_parent_id` (`parent_id`),
  CONSTRAINT `fk_topic_comments_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`),
  CONSTRAINT `fk_topic_comments_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_topic_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_topic_comments_parent` FOREIGN KEY (`parent_id`) REFERENCES `topic_comments` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子评论表（1级数据）';

-- ============================================================
-- 9. topic_reactions — 帖子互动表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `topic_reactions`;
CREATE TABLE `topic_reactions` (
  `id` char(32) NOT NULL COMMENT '互动记录主键',
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `reaction_type` varchar(20) NOT NULL COMMENT '互动类型：like/favorite/share',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '互动状态：1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_topic_reaction` (`user_id`, `topic_id`, `reaction_type`),
  KEY `idx_topic_reaction` (`topic_id`, `reaction_type`, `status`),
  KEY `idx_user_reaction_time` (`user_id`, `reaction_type`, `created_at`),
  CONSTRAINT `fk_topic_reactions_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`),
  CONSTRAINT `fk_topic_reactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子互动表（1级数据）';

-- ============================================================
-- 10. topic_reply_likes — 评论/回复点赞表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `topic_reply_likes`;
CREATE TABLE `topic_reply_likes` (
  `id` char(32) NOT NULL COMMENT '回复点赞主键',
  `reply_id` char(32) NOT NULL COMMENT '评论/回复 ID',
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `user_id` char(32) NOT NULL COMMENT '点赞用户 ID',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_reply` (`user_id`, `reply_id`),
  KEY `idx_reply_status` (`reply_id`, `status`),
  KEY `idx_topic_status` (`topic_id`, `status`),
  CONSTRAINT `fk_topic_reply_likes_reply` FOREIGN KEY (`reply_id`) REFERENCES `topic_comments` (`id`),
  CONSTRAINT `fk_topic_reply_likes_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`),
  CONSTRAINT `fk_topic_reply_likes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论/回复点赞表（1级数据）';

-- ============================================================
-- 11. group_admin_actions — 圈主管理操作日志表（1级审计数据）
-- ============================================================
DROP TABLE IF EXISTS `group_admin_actions`;
CREATE TABLE `group_admin_actions` (
  `id` char(32) NOT NULL COMMENT '操作日志主键',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `operator_id` char(32) NOT NULL COMMENT '操作人 ID',
  `target_user_id` char(32) NOT NULL COMMENT '被操作用户 ID',
  `action_type` varchar(30) NOT NULL COMMENT '操作类型：approve/ban/mute/remove/set_admin/recover/unban/unmute',
  `reason` varchar(500) DEFAULT NULL COMMENT '操作原因',
  `created_at` bigint(20) NOT NULL COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_operator` (`group_id`, `operator_id`, `created_at`),
  KEY `idx_target` (`target_user_id`, `created_at`),
  KEY `idx_action_type` (`action_type`, `created_at`),
  CONSTRAINT `fk_group_admin_actions_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_group_admin_actions_operator` FOREIGN KEY (`operator_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_group_admin_actions_target` FOREIGN KEY (`target_user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='圈主管理操作日志表（1级审计数据）';

-- ============================================================
-- 12. topic_audit_logs — 帖子审核日志表（1级审计数据）
-- ============================================================
DROP TABLE IF EXISTS `topic_audit_logs`;
CREATE TABLE `topic_audit_logs` (
  `id` char(32) NOT NULL COMMENT '审核日志主键',
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `operator_id` char(32) NOT NULL COMMENT '审核操作人 ID',
  `author_id` char(32) NOT NULL COMMENT '帖子作者 ID',
  `action` varchar(30) NOT NULL COMMENT '审核动作：approve/reject/ban/delete',
  `old_status` tinyint(4) NOT NULL COMMENT '审核前状态',
  `new_status` tinyint(4) NOT NULL COMMENT '审核后状态',
  `reason` varchar(500) DEFAULT NULL COMMENT '审核原因',
  `created_at` bigint(20) NOT NULL COMMENT '审核时间',
  PRIMARY KEY (`id`),
  KEY `idx_topic_time` (`topic_id`, `created_at`),
  KEY `idx_group_time` (`group_id`, `created_at`),
  KEY `idx_operator_time` (`operator_id`, `created_at`),
  KEY `idx_author_time` (`author_id`, `created_at`),
  CONSTRAINT `fk_topic_audit_logs_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`),
  CONSTRAINT `fk_topic_audit_logs_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_topic_audit_logs_operator` FOREIGN KEY (`operator_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_topic_audit_logs_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子审核日志表（1级审计数据）';

-- ============================================================
-- 13. member_blocks — 用户拉黑关系表（1级数据）
-- ============================================================
DROP TABLE IF EXISTS `member_blocks`;
CREATE TABLE `member_blocks` (
  `id` char(32) NOT NULL COMMENT '拉黑记录主键',
  `user_id` char(32) NOT NULL COMMENT '发起拉黑的用户 ID',
  `blocked_user_id` char(32) NOT NULL COMMENT '被拉黑用户 ID',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_blocked` (`user_id`, `blocked_user_id`),
  KEY `idx_blocked_user` (`blocked_user_id`, `status`),
  KEY `idx_user_status` (`user_id`, `status`),
  CONSTRAINT `fk_member_blocks_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_member_blocks_blocked_user` FOREIGN KEY (`blocked_user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户拉黑关系表（1级数据）';

-- ============================================================
-- 14. member_stats — 成员统计快照表（2级，可重建）
-- ============================================================
DROP TABLE IF EXISTS `member_stats`;
CREATE TABLE `member_stats` (
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `topics_count` int(11) NOT NULL DEFAULT '0' COMMENT '发帖数',
  `followers_count` int(11) NOT NULL DEFAULT '0' COMMENT '粉丝数',
  `following_count` int(11) NOT NULL DEFAULT '0' COMMENT '关注数',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`user_id`),
  CONSTRAINT `fk_member_stats_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成员统计快照表（2级，可重建）';

-- ============================================================
-- 15. group_stats — 群组统计快照表（2级，可重建）
-- ============================================================
DROP TABLE IF EXISTS `group_stats`;
CREATE TABLE `group_stats` (
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `members_count` int(11) NOT NULL DEFAULT '0' COMMENT '成员数',
  `active_members_count` int(11) NOT NULL DEFAULT '0' COMMENT '活跃成员数',
  `paid_members_count` int(11) NOT NULL DEFAULT '0' COMMENT '付费成员数',
  `topics_count` int(11) NOT NULL DEFAULT '0' COMMENT '帖子数',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`group_id`),
  CONSTRAINT `fk_group_stats_group` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组统计快照表（2级，可重建）';

-- ============================================================
-- 16. topic_stats — 主题统计快照表（2级，可重建）
-- ============================================================
DROP TABLE IF EXISTS `topic_stats`;
CREATE TABLE `topic_stats` (
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `read_count` int(11) NOT NULL DEFAULT '0' COMMENT '阅读数',
  `comments_count` int(11) NOT NULL DEFAULT '0' COMMENT '评论数',
  `likes_count` int(11) NOT NULL DEFAULT '0' COMMENT '点赞数',
  `favorites_count` int(11) NOT NULL DEFAULT '0' COMMENT '收藏数',
  `shares_count` int(11) NOT NULL DEFAULT '0' COMMENT '分享数',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`topic_id`),
  CONSTRAINT `fk_topic_stats_topic` FOREIGN KEY (`topic_id`) REFERENCES `topics` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题统计快照表（2级，可重建）';

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 完成：Social 域 16 张表已创建
-- ============================================================
