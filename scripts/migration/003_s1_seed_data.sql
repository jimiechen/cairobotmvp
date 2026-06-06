-- ============================================================
-- 迁移脚本: 003_s1_seed_data.sql
-- 用途: S1 阶段种子数据（配置模块 + 语言包 + 示例字符串）
-- 版本: v1.0
-- 日期: 2026-06-05
-- 前置: 必须先执行 002_s1_mysql_tables.sql
-- ============================================================

USE cairobot_db;

-- ---- 种子数据：初始配置模块 schema 注册 ----
INSERT IGNORE INTO sys_config_schema (module_key, field_key, field_type, default_value, description, sort_order) VALUES
('base_config', 'app_name', 'string', 'CaiRobot', '应用名称', 1),
('base_config', 'debug', 'bool', 'true', '调试模式开关', 2),
('base_config', 'max_connections', 'int', '100', '最大并发连接数', 3),
('base_config', 'domain_root', 'string', '', 'API 根域名', 4),
('base_config', 'sign_salt', 'string', '', '签名随机盐值（生产环境必须设置）', 5);

-- ---- 种子数据：初始语言包注册 ----
INSERT IGNORE INTO sys_lang_pack (pack_name, env, lang_code, version, description, is_published, published_at)
VALUES ('webp', 'dev', 'zh-CN', 1, '简体中文 Web 端', 1, NOW()),
       ('webp', 'dev', 'en-US', 1, 'English Web Pack', 1, NOW());

-- ---- 种子数据：简体中文字符串 ----
INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.ok', '确定', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.cancel', '取消', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.loading', '加载中...', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'error.network', '网络连接失败，请检查网络设置', 'error', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'error.unauthorized', '未授权访问，请重新登录', 'error', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'greeting.welcome', '欢迎 {name}，你有 {count} 条新消息', 'app', 1, 'ADD', 'named',
       '[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]'
FROM sys_lang_pack WHERE lang_code = 'zh-CN' AND env = 'dev';

-- ---- 种子数据：英文字符串 ----
INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.ok', 'OK', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.cancel', 'Cancel', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'common.loading', 'Loading...', 'common', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'error.network', 'Network error. Please check your connection.', 'error', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type)
SELECT id, 'error.unauthorized', 'Unauthorized. Please log in again.', 'error', 1, 'ADD', 'plain'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

INSERT IGNORE INTO sys_lang_string (pack_id, string_key, string_value, group_name, version, operation_type, template_type, params_schema)
SELECT id, 'greeting.welcome', 'Welcome {name}, you have {count} new messages', 'app', 1, 'ADD', 'named',
       '[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]'
FROM sys_lang_pack WHERE lang_code = 'en-US' AND env = 'dev';

-- ---- 种子数据：示例配置版本 ----
INSERT IGNORE INTO sys_config_version (module_key, env, version, config_data, is_published, publisher, published_at)
VALUES ('base_config', 'dev', 1,
        '{"app_name":"CaiRobot","debug":true,"max_connections":100,"domain_root":"","sign_salt":""}',
        1, 'system', NOW());

-- ---- 验证：确认种子数据已写入 ----
SELECT '=== Schema 注册 ===' AS info;
SELECT module_key, COUNT(*) AS field_count FROM sys_config_schema GROUP BY module_key;

SELECT '=== 语言包 ===' AS info;
SELECT lang_code, pack_name, version, is_published FROM sys_lang_pack;

SELECT '=== 字符串统计 ===' AS info;
SELECT lp.lang_code, COUNT(ls.id) AS string_count
FROM sys_lang_pack lp LEFT JOIN sys_lang_string ls ON ls.pack_id = lp.id
GROUP BY lp.lang_code;

SELECT '=== 配置版本 ===' AS info;
SELECT module_key, env, version, is_published FROM sys_config_version;
