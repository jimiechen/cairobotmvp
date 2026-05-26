-- Hello 模块配置 Schema 和多语言 Seed 数据
-- 用于模块首启时自动注入默认配置和文案
-- 与代码一同发布，确保"光有代码没有数据"的问题不会发生

-- ============================================================
-- 1. 配置 Schema 注入（sys_config_schema）
-- ============================================================

INSERT INTO sys_config_schema (module_key, field_key, field_type, default_value, validator, scope, description, created_at, updated_at) VALUES
('hello_cfg', 'server_name', 'string', 'CaiRobot', NULL, 'all', '服务名称，用于问候语渲染', NOW(), NOW()),
('hello_cfg', 'max_name_length', 'int', '32', '{"min":1,"max":256}', 'all', '用户名最大长度限制，超出返回 10400 错误', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  updated_at = NOW();

-- ============================================================
-- 2. 多语言文案注入（sys_lang_string）
-- ============================================================

INSERT INTO sys_lang_string (lang_key, template_type, params_schema, lang_code, template_text, description, created_at, updated_at) VALUES
('svc_hello_greeting', 'named', '[{"name":"name","type":"string","required":true},{"name":"server_name","type":"string","required":true}]', 'zh-CN', '你好，{name}！欢迎使用 {server_name}。', 'Hello 模块中文问候语模板', NOW(), NOW()),
('svc_hello_greeting', 'named', '[{"name":"name","type":"string","required":true},{"name":"server_name","type":"string","required":true}]', 'en', 'Hello, {name}! Welcome to {server_name}.', 'Hello 模块英文问候语模板', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  updated_at = NOW();

-- ============================================================
-- 3. 验证查询（可选，用于确认数据已正确插入）
-- ============================================================

-- SELECT * FROM sys_config_schema WHERE module_key = 'hello_cfg';
-- SELECT * FROM sys_lang_string WHERE lang_key = 'svc_hello_greeting';
