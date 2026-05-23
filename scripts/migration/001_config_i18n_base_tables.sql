-- ============================================================
-- 迁移脚本: 001_config_i18n_base_tables.sql
-- 用途: 全局配置元模型 + 多语言参数化模板 基础表结构
-- 版本: v1.0
-- 日期: 2026-05-22
-- 关联文档:
--   - docs/prd/global-config-i18n-implementation-plan.md §5
--   - ADR-009 (待创建)
-- 兼容性: MySQL 8.0+ / SQLite 3.35.0+
-- ============================================================

-- ---- 1. sys_config_schema 配置字段元数据注册表 ----
CREATE TABLE IF NOT EXISTS sys_config_schema (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  module_key    TEXT NOT NULL,
  field_key     TEXT NOT NULL,
  field_type    TEXT NOT NULL DEFAULT 'string',
  default_value TEXT,
  validator     TEXT,
  is_required   INTEGER DEFAULT 0,
  is_secret     INTEGER DEFAULT 0,
  description   TEXT,
  client_scope  TEXT DEFAULT 'all',
  min_app_ver   TEXT,
  sort_order    INTEGER DEFAULT 0,
  is_enabled    INTEGER DEFAULT 1,
  created_at    TEXT DEFAULT (datetime('now')),
  updated_at    TEXT DEFAULT (datetime('now')),
  UNIQUE (module_key, field_key)
);

-- ---- 2. sys_config_version 应用配置版本主表 ----
CREATE TABLE IF NOT EXISTS sys_config_version (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  module_key    TEXT NOT NULL,
  env           TEXT NOT NULL DEFAULT 'dev',
  version       INTEGER NOT NULL DEFAULT 1,
  config_json   TEXT NOT NULL,
  is_published  INTEGER NOT NULL DEFAULT 0,
  published_at  TEXT,
  created_at    TEXT DEFAULT (datetime('now')),
  updated_at    TEXT DEFAULT (datetime('now')),
  create_by     TEXT,
  update_by     TEXT
);
CREATE INDEX IF NOT EXISTS idx_cfg_ver_module_env_version
  ON sys_config_version (module_key, env, version);
CREATE INDEX IF NOT EXISTS idx_cfg_ver_latest
  ON sys_config_version (module_key, env, is_published, version DESC);

-- ---- 3. sys_lang_pack 语言包主表 ----
CREATE TABLE IF NOT EXISTS sys_lang_pack (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  pack_name     TEXT NOT NULL DEFAULT '',
  env           TEXT NOT NULL DEFAULT 'dev',
  version       INTEGER NOT NULL DEFAULT 1,
  lang_code     TEXT NOT NULL,
  description   TEXT,
  is_published  INTEGER NOT NULL DEFAULT 0,
  published_at  TEXT,
  published_by  INTEGER,
  created_at    TEXT DEFAULT (datetime('now')),
  updated_at    TEXT DEFAULT (datetime('now')),
  UNIQUE (pack_name, env, lang_code)
);
CREATE INDEX IF NOT EXISTS idx_lang_pack_published
  ON sys_lang_pack (is_published);

-- ---- 4. sys_lang_string 语言字符串明细表（含参数化模板支持）----
CREATE TABLE IF NOT EXISTS sys_lang_string (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  pack_id        INTEGER NOT NULL,
  string_key     TEXT NOT NULL,
  string_value   TEXT NOT NULL,
  group_name     TEXT NOT NULL DEFAULT 'common',
  version        INTEGER NOT NULL DEFAULT 1,
  operation_type TEXT NOT NULL DEFAULT 'ADD',
  prev_value     TEXT,
  template_type  TEXT DEFAULT 'plain',
  params_schema  TEXT DEFAULT NULL,
  preview_sample TEXT DEFAULT NULL,
  created_at     TEXT DEFAULT (datetime('now')),
  updated_at     TEXT DEFAULT (datetime('now')),
  UNIQUE (pack_id, string_key)
);
CREATE INDEX IF NOT EXISTS idx_lang_str_pack_id ON sys_lang_string (pack_id);
CREATE INDEX IF NOT EXISTS idx_lang_str_operation ON sys_lang_string (pack_id, operation_type);
CREATE INDEX IF NOT EXISTS idx_lang_str_string_key ON sys_lang_string (string_key);

-- ---- 种子数据：初始模块 schema 注册 ----
INSERT OR IGNORE INTO sys_config_schema (module_key, field_key, field_type, default_value, description, sort_order) VALUES
('base_cfg', 'domain_root', 'string', '', 'API 根域名', 1),
('base_cfg', 'domain_wap', 'string', '', 'WAP 页面域名', 2),
('base_cfg', 'sign_rand', 'string', '', '签名随机盐值', 3),
('base_cfg', 'construct_email', 'string', '', '反馈联系邮箱', 4),
('wap_cfg', 'user_agreement_url', 'string', '', '用户协议 URL', 1),
('wap_cfg', 'privacy_policy_url', 'string', '', '隐私政策 URL', 2),
('regex_cfg', 'regex_email', 'string', '', '邮箱正则表达式', 1),
('regex_cfg', 'regex_password', 'string', '', '密码正则表达式', 2),
('regex_cfg', 'regex_phone', 'string', '', '手机号正则表达式', 3),
('regex_cfg', 'regex_nick', 'string', '', '昵称正则表达式', 4),
('regex_cfg', 'regex_circle_name', 'string', '', '圈子名称正则表达式', 5),
('oss_cfg', 'oss_host', 'string', '', 'OSS 主机地址', 1),
('oss_cfg', 'oss_domain', 'string', '', 'OSS 域名', 2),
('oss_cfg', 'cdn_domain', 'string', '', 'CDN 域名', 3),
('lang_cfg', 'lang_code', 'string', 'zh-CN', '默认语言代码', 1),
('mute_cfg', 'durations', 'json', '[]', '静音时长选项列表', 1),
('group_cfg', 'group_config_pay_notice', 'string', '', '群组支付公告文案', 1);

-- ---- 种子数据：初始语言包注册 ----
INSERT OR IGNORE INTO sys_lang_pack (pack_name, env, lang_code, version, description, is_published) VALUES
('webp', 'dev', 'zh-CN', 1, '简体中文', 1),
('webp', 'dev', 'en', 1, 'English', 1);

-- ---- 种子数据：示例多语言字符串（含 plain 和 named 类型演示）----
INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_common_ok', '确定', 'common', 1, 'ADD', 'plain', NULL FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';
INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_common_cancel', '取消', 'common', 1, 'ADD', 'plain', NULL FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';
INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_msg_welcome', '欢迎 {name}，你有 {count} 条新消息', 'app', 1, 'ADD', 'named',
       '[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_common_ok', 'OK', 'common', 1, 'ADD', 'plain', NULL FROM sys_lang_pack WHERE lang_code = 'en' AND env = 'dev';
INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_common_cancel', 'Cancel', 'common', 1, 'ADD', 'plain', NULL FROM sys_lang_pack WHERE lang_code = 'en' AND env = 'dev';
INSERT OR IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'svc_msg_welcome', 'Welcome {name}, you have {count} new messages', 'app', 1, 'ADD', 'named',
       '[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]'
FROM sys_lang_pack WHERE lang_code = 'en' AND env = 'dev';
