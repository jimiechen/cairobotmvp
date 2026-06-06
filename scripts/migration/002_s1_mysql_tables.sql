-- ============================================================
-- 迁移脚本: 002_s1_mysql_tables.sql
-- 用途: S1 阶段 MySQL 表结构（从 001 SQLite 版本适配）
-- 版本: v1.0
-- 日期: 2026-06-05
-- 关联文档:
--   - docs/superpowers/specs/2026-06-05-s1-real-services-auth-design.md §3
--   - docs/prd/global-config-i18n-implementation-plan.md §5
--   - ADR-009-config-i18n-schema-template
-- 目标数据库: MySQL 8.0+
-- 数据库: cairobot_db (由 gateway.local.conf 配置)
-- ============================================================

-- ---- 前置：确保数据库存在 ----
CREATE DATABASE IF NOT EXISTS cairobot_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cairobot_db;

-- ---- 1. sys_config_schema 配置字段元数据注册表 ----
-- 与 001 迁移结构一致，使用 MySQL 类型
CREATE TABLE IF NOT EXISTS sys_config_schema (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    module_key    VARCHAR(64) NOT NULL,
    field_key     VARCHAR(128) NOT NULL,
    field_type    VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT 'string/int/bool/json/regex',
    default_value TEXT,
    validator     TEXT COMMENT '正则表达式或 JSON Schema',
    is_required   TINYINT(1) NOT NULL DEFAULT 0,
    is_secret     TINYINT(1) NOT NULL DEFAULT 0,
    description   VARCHAR(512),
    client_scope  VARCHAR(64) DEFAULT 'all',
    min_app_ver   VARCHAR(16),
    sort_order    INT NOT NULL DEFAULT 0,
    is_enabled    TINYINT(1) NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_module_field (module_key, field_key),
    INDEX idx_module_key (module_key),
    INDEX idx_enabled (is_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置字段元数据注册表';

-- ---- 2. sys_config_version 应用配置版本主表 ----
-- config_data 使用 JSON 类型存储完整配置快照
CREATE TABLE IF NOT EXISTS sys_config_version (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    module_key    VARCHAR(64) NOT NULL,
    env           VARCHAR(16) NOT NULL DEFAULT 'dev',
    version       BIGINT NOT NULL DEFAULT 1,
    config_data   JSON NOT NULL COMMENT '配置内容 JSON',
    is_published TINYINT(1) NOT NULL DEFAULT 0,
    published_at  DATETIME DEFAULT NULL,
    publisher     VARCHAR(64) DEFAULT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    create_by     VARCHAR(64) DEFAULT NULL,
    update_by     VARCHAR(64) DEFAULT NULL,
    UNIQUE KEY uk_module_env_version (module_key, env, version),
    INDEX idx_cfg_ver_module_env (module_key, env),
    INDEX idx_cfg_ver_published (module_key, env, is_published, version DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用配置版本快照表';

-- ---- 3. sys_lang_pack 语言包主表 ----
CREATE TABLE IF NOT EXISTS sys_lang_pack (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    pack_name     VARCHAR(64) NOT NULL DEFAULT '',
    env           VARCHAR(16) NOT NULL DEFAULT 'dev',
    version       BIGINT NOT NULL DEFAULT 1,
    lang_code     VARCHAR(10) NOT NULL,
    description   VARCHAR(256),
    is_published  TINYINT(1) NOT NULL DEFAULT 0,
    published_at  DATETIME DEFAULT NULL,
    published_by  VARCHAR(64) DEFAULT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pack_name_env_lang (pack_name, env, lang_code),
    INDEX idx_lang_pack_published (is_published),
    INDEX idx_lang_pack_code (lang_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语言包主表';

-- ---- 4. sys_lang_string 语言字符串明细表（含参数化模板支持）----
CREATE TABLE IF NOT EXISTS sys_lang_string (
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    pack_id        BIGINT NOT NULL,
    string_key     VARCHAR(256) NOT NULL,
    string_value   TEXT NOT NULL,
    group_name     VARCHAR(64) NOT NULL DEFAULT 'common',
    version        BIGINT NOT NULL DEFAULT 1,
    operation_type ENUM('ADD','UPDATE','DELETE') NOT NULL DEFAULT 'ADD',
    prev_value     TEXT,
    template_type  ENUM('plain','named','icu') DEFAULT 'plain' COMMENT '参数化模板类型',
    params_schema JSON DEFAULT NULL COMMENT '参数定义 JSON Schema',
    preview_sample TEXT DEFAULT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pack_string_key (pack_id, string_key),
    INDEX idx_lang_str_pack_id (pack_id),
    INDEX idx_lang_str_operation (pack_id, operation_type),
    INDEX idx_lang_str_group (group_name),
    CONSTRAINT fk_lang_string_pack FOREIGN KEY (pack_id) REFERENCES sys_lang_pack(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语言字符串明细表';
