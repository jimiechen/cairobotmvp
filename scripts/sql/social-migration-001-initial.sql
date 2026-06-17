-- ============================================================
-- 迁移脚本: social-migration-001-initial.sql
-- 用途: 社交域 MVP 初始表结构（10 张表）
-- 版本: v1.0
-- 日期: 2026-06-16
-- 关联文档:
--   - 社交域 PRD 数据模型
--   - proto/social/*.proto（GroupInfo / TopicInfo / ReplyInfo 等消息定义）
-- 目标数据库: MySQL 8.0+
--
-- 建表顺序（按依赖关系分批）:
--   Batch 1: groups, topics, topic_replies（无外键依赖）
--   Batch 2: group_members, group_pay_configs（依赖 groups + users）
--   Batch 3: topic_reads, topic_likes, topic_favorites, reply_likes（依赖 topics/group_members）
--   Batch 4: audit_logs（审计日志，可延后）
-- ============================================================

-- ============================================================
-- Batch 1: 无外键依赖的基础表
-- ============================================================

-- ---- 1. groups 群组表 ----
-- 对应 proto GroupInfo 消息 + GroupStats 统计字段
CREATE TABLE IF NOT EXISTS groups (
    id              VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '群组 ID（ULID）',
    name            VARCHAR(200) NOT NULL COMMENT '群组名称',
    slug            VARCHAR(200) NOT NULL COMMENT '群组唯一标识 slug',
    description     TEXT COMMENT '群组描述',
    avatar          VARCHAR(500) COMMENT '群组头像 URL',
    cover_image     VARCHAR(500) COMMENT '群组封面图 URL',
    category        VARCHAR(100) COMMENT '群组分类',
    tags            JSON COMMENT '群组标签 JSON 数组',
    owner_id        VARCHAR(64) NOT NULL COMMENT '群主用户 ID',
    status          TINYINT NOT NULL DEFAULT 1 COMMENT '群组状态：1-正常 2-已归档 3-已解散（GroupStatus 枚举）',
    visibility      TINYINT NOT NULL DEFAULT 1 COMMENT '可见性：1-公开 2-私密 3-邀请可见（GroupVisibility 枚举）',
    join_mode       TINYINT NOT NULL DEFAULT 1 COMMENT '加入方式：1-自由加入 2-需要审核 3-仅邀请（JoinMode 枚举）',
    is_official     BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否官方群组',
    is_featured     BOOLEAN NOT NULL DEFAULT FALSE '是否推荐/精选',
    max_members     INT NOT NULL DEFAULT 500 COMMENT '最大成员数上限',
    created_at      BIGINT COMMENT '创建时间戳（毫秒）',
    updated_at      BIGINT COMMENT '更新时间戳（毫秒）',
    UNIQUE KEY uk_groups_slug (slug),
    INDEX idx_groups_owner_id (owner_id),
    INDEX idx_groups_status (status),
    INDEX idx_groups_category (category),
    INDEX idx_groups_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组表';

-- ---- 2. topics 主题/帖子表 ----
-- 对应 proto TopicInfo 消息
CREATE TABLE IF NOT EXISTS topics (
    id                  VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '主题 ID（ULID）',
    title               VARCHAR(500) NOT NULL COMMENT '主题标题',
    content             TEXT COMMENT '主题内容（富文本/Markdown）',
    author_id           VARCHAR(64) NOT NULL COMMENT '作者用户 ID',
    author_name         VARCHAR(200) COMMENT '作者昵称（冗余字段，避免联表）',
    group_id            VARCHAR(64) COMMENT '所属群组 ID（NULL 表示全局帖子）',
    topic_type          TINYINT NOT NULL DEFAULT 1 COMMENT '主题类型：1-普通帖 2-问答帖 3-投票帖（TopicType 枚举）',
    content_format      TINYINT NOT NULL DEFAULT 1 COMMENT '内容格式：1-纯文本 2-Markdown 3-富文本（ContentFormat 枚举）',
    visibility          TINYINT NOT NULL DEFAULT 1 COMMENT '可见性：1-公开 2-仅成员（Visibility 枚举）',
    summary             VARCHAR(500) COMMENT '主题摘要',
    author_avatar       VARCHAR(500) COMMENT '作者头像 URL（冗余字段）',
    question_text       VARCHAR(300) COMMENT '问答帖的问题文本',
    qa_private          BOOLEAN NOT NULL DEFAULT FALSE COMMENT '问答是否私密',
    answered_at         BIGINT COMMENT '被回答时间戳',
    cover_image         VARCHAR(500) COMMENT '封面图 URL',
    member_id           VARCHAR(64) COMMENT '关联的成员记录 ID',
    draft_id            VARCHAR(64) COMMENT '关联草稿 ID',
    nav_types           JSON COMMENT '导航类型 JSON 数组',
    qa_target_user_id   VARCHAR(64) COMMENT '问答指定目标用户 ID',
    has_media           BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否包含媒体附件',
    has_docs            BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否包含文档附件',
    status              TINYINT NOT NULL DEFAULT 2 COMMENT '生命周期状态：1-草稿 2-已发布 3-已关闭 4-已删除（TopicLifecycleStatus 枚举）',
    created_at          BIGINT COMMENT '创建时间戳（毫秒）',
    updated_at          BIGINT COMMENT '更新时间戳（毫秒）',
    INDEX idx_topics_author_id (author_id),
    INDEX idx_topics_group_id (group_id),
    INDEX idx_topics_topic_type (topic_type),
    INDEX idx_topics_status (status),
    INDEX idx_topics_created_at (created_at),
    CONSTRAINT fk_topics_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主题/帖子表';

-- ---- 3. topic_replies 评论表 ----
-- 对应 proto ReplyInfo 消息
CREATE TABLE IF NOT EXISTS topic_replies (
    id                  VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '评论 ID（ULID）',
    topic_id            VARCHAR(64) NOT NULL COMMENT '所属主题 ID',
    content             TEXT NOT NULL COMMENT '评论内容',
    author_id           VARCHAR(64) NOT NULL COMMENT '评论者用户 ID',
    author_name         VARCHAR(200) COMMENT '评论者昵称（冗余字段）',
    parent_reply_id     VARCHAR(64) COMMENT '父评论 ID（用于嵌套回复，NULL 表示顶层评论）',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '评论状态：1-正常 2-已删除 3-已屏蔽（ReplyStatus 枚举）',
    like_count          INT NOT NULL DEFAULT 0 COMMENT '点赞数',
    created_at          BIGINT COMMENT '创建时间戳（毫秒）',
    updated_at          BIGINT COMMENT '更新时间戳（毫秒）',
    replies_count       INT NOT NULL DEFAULT 0 COMMENT '子回复数量',
    is_pinned           BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否置顶',
    is_liked            BOOLEAN NOT NULL DEFAULT FALSE COMMENT '当前用户是否已点赞（查询时动态填充）',
    level               TINYINT NOT NULL DEFAULT 1 COMMENT '回复层级：1-顶层 2-二级及以下',
    reply_to_user_id    VARCHAR(64) COMMENT '被回复的用户 ID',
    reply_to_user_name  VARCHAR(200) COMMENT '被回复的用户昵称',
    INDEX idx_replies_topic_id (topic_id),
    INDEX idx_replies_author_id (author_id),
    INDEX idx_replies_parent_reply_id (parent_reply_id),
    INDEX idx_replies_created_at (created_at),
    CONSTRAINT fk_replies_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_replies_parent FOREIGN KEY (parent_reply_id) REFERENCES topic_replies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主题评论表';

-- ============================================================
-- Batch 2: 依赖 groups + users 的表
-- ============================================================

-- ---- 4. group_members 群组成员关系表 ----
-- 对应 proto MemberInfo 消息
-- 注意：users 表由用户域维护，此处仅引用外键
CREATE TABLE IF NOT EXISTS group_members (
    id                      VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '成员关系 ID（ULID）',
    group_id                VARCHAR(64) NOT NULL COMMENT '群组 ID',
    user_id                 VARCHAR(64) NOT NULL COMMENT '用户 ID',
    role                    TINYINT NOT NULL DEFAULT 3 COMMENT '角色：1-群主 2-管理员 3-普通成员 4-待审核（GroupMemberRole 枚举）',
    status                  TINYINT NOT NULL DEFAULT 1 COMMENT '成员状态：1-正常 2-已退出 3-已移除 4-已禁言 5-待审核（MemberStatus 枚举）',
    join_reason             VARCHAR(500) COMMENT '申请加入理由',
    invited_by              VARCHAR(64) COMMENT '邀请人用户 ID',
    joined_at               BIGINT COMMENT '加入时间戳（毫秒）',
    last_activity_at        BIGINT COMMENT '最后活跃时间戳（毫秒）',
    bio                     VARCHAR(500) COMMENT '成员个人简介',
    answered_questions_count INT NOT NULL DEFAULT 0 COMMENT '已回答问题数',
    remaining_quota         INT NOT NULL DEFAULT 0 COMMENT '剩余提问配额',
    payment_cycle           TINYINT COMMENT '付费周期类型：NULL-免费 1-月付 2-季付 3-年付',
    membership_expires_at   BIGINT COMMENT '会员到期时间戳（毫秒）',
    muted_until             BIGINT COMMENT '禁言到期时间戳（毫秒）',
    ban_expires_at          BIGINT COMMENT '封禁到期时间戳（毫秒）',
    approved_by             VARCHAR(64) COMMENT '审核通过操作人 ID',
    banned_by               VARCHAR(64) COMMENT '封禁操作人 ID',
    ban_reason              VARCHAR(500) COMMENT '封禁原因',
    banned_at               BIGINT COMMENT '封禁时间戳（毫秒）',
    created_at              BIGINT COMMENT '创建时间戳（毫秒）',
    updated_at              BIGINT COMMENT '更新时间戳（毫秒）',
    UNIQUE KEY uk_group_user (group_id, user_id),
    INDEX idx_members_group_id (group_id),
    INDEX idx_members_user_id (user_id),
    INDEX idx_members_role (role),
    INDEX idx_members_status (status),
    CONSTRAINT fk_members_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组成员关系表';

-- ---- 5. group_pay_configs 群组付费配置表 ----
CREATE TABLE IF NOT EXISTS group_pay_configs (
    id              VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '付费配置 ID（ULID）',
    group_id        VARCHAR(64) NOT NULL COMMENT '群组 ID',
    price_monthly   DECIMAL(10,2) COMMENT '月付价格',
    price_quarterly DECIMAL(10,2) COMMENT '季付价格',
    price_yearly    DECIMAL(10,2) COMMENT '年付价格',
    currency        VARCHAR(10) NOT NULL DEFAULT 'CNY' COMMENT '货币单位',
    trial_days      INT NOT NULL DEFAULT 0 COMMENT '试用天数',
    is_enabled      BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用付费',
    created_at      BIGINT COMMENT '创建时间戳（毫秒）',
    updated_at      BIGINT COMMENT '更新时间戳（毫秒）',
    UNIQUE KEY uk_pay_config_group (group_id),
    CONSTRAINT fk_pay_config_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组付费配置表';

-- ============================================================
-- Batch 3: 依赖 topics + group_members + users 的交互表
-- ============================================================

-- ---- 6. topic_reads 阅读记录表 ----
CREATE TABLE IF NOT EXISTS topic_reads (
    id          VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '阅读记录 ID（ULID）',
    topic_id    VARCHAR(64) NOT NULL COMMENT '主题 ID',
    user_id     VARCHAR(64) NOT NULL COMMENT '用户 ID',
    read_at     BIGINT COMMENT '阅读时间戳（毫秒）',
    read_duration INT NOT NULL DEFAULT 0 COMMENT '阅读时长（秒）',
    UNIQUE KEY uk_topic_read (topic_id, user_id),
    INDEX idx_reads_user_id (user_id),
    INDEX idx_reads_read_at (read_at),
    CONSTRAINT fk_reads_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_reads_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主题阅读记录表';

-- ---- 7. topic_likes 点赞表 ----
CREATE TABLE IF NOT EXISTS topic_likes (
    id          VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '点赞记录 ID（ULID）',
    topic_id    VARCHAR(64) NOT NULL COMMENT '主题 ID',
    user_id     VARCHAR(64) NOT NULL COMMENT '用户 ID',
    created_at  BIGINT COMMENT '点赞时间戳（毫秒）',
    UNIQUE KEY uk_topic_like (topic_id, user_id),
    INDEX idx_likes_user_id (user_id),
    INDEX idx_likes_created_at (created_at),
    CONSTRAINT fk_likes_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主题点赞表';

-- ---- 8. topic_favorites 收藏表 ----
CREATE TABLE IF NOT EXISTS topic_favorites (
    id          VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '收藏记录 ID（ULID）',
    topic_id    VARCHAR(64) NOT NULL COMMENT '主题 ID',
    user_id     VARCHAR(64) NOT NULL COMMENT '用户 ID',
    created_at  BIGINT COMMENT '收藏时间戳（毫秒）',
    UNIQUE KEY uk_topic_favorite (topic_id, user_id),
    INDEX idx_favorites_user_id (user_id),
    INDEX idx_favorites_created_at (created_at),
    CONSTRAINT fk_favorites_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_favorites_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主题收藏表';

-- ---- 9. reply_likes 评论点赞表 ----
CREATE TABLE IF NOT EXISTS reply_likes (
    id          VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '评论点赞记录 ID（ULID）',
    reply_id    VARCHAR(64) NOT NULL COMMENT '评论 ID',
    user_id     VARCHAR(64) NOT NULL COMMENT '用户 ID',
    created_at  BIGINT COMMENT '点赞时间戳（毫秒）',
    UNIQUE KEY uk_reply_like (reply_id, user_id),
    INDEX idx_reply_likes_user_id (user_id),
    INDEX idx_reply_likes_created_at (created_at),
    CONSTRAINT fk_reply_likes_reply FOREIGN KEY (reply_id) REFERENCES topic_replies(id) ON DELETE CASCADE,
    CONSTRAINT fk_reply_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论点赞表';

-- ============================================================
-- Batch 4: 审计日志（可延后执行）
-- ============================================================

-- ---- 10. audit_logs 审计日志表 ----
CREATE TABLE IF NOT EXISTS audit_logs (
    id              VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '审计日志 ID（ULID）',
    actor_id        VARCHAR(64) NOT NULL COMMENT '操作人用户 ID',
    actor_name      VARCHAR(200) COMMENT '操作人昵称',
    action          VARCHAR(50) NOT NULL COMMENT '操作类型：create/update/delete/join/leave/ban/unban 等',
    target_type     VARCHAR(50) NOT NULL COMMENT '目标资源类型：group/topic/reply/member 等',
    target_id       VARCHAR(64) NOT NULL COMMENT '目标资源 ID',
    group_id        VARCHAR(64) COMMENT '关联群组 ID（方便按群筛选）',
    extra_data      JSON COMMENT '额外信息 JSON（变更前后快照等）',
    ip_address      VARCHAR(45) COMMENT '操作 IP 地址',
    user_agent      VARCHAR(500) COMMENT '客户端 User-Agent',
    created_at      BIGINT COMMENT '操作时间戳（毫秒）',
    INDEX idx_audit_actor (actor_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_target (target_type, target_id),
    INDEX idx_audit_group_id (group_id),
    INDEX audit_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='社交域审计日志表';
