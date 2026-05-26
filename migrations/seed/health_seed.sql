-- Health 模块配置 Schema 和多语言 Seed 数据
-- 用于模块首启时自动注入默认配置和文案
-- 与代码一同发布，确保"光有代码没有数据"的问题不会发生

-- ============================================================
-- 1. 配置 Schema 注入（sys_config_schema）
-- ============================================================

INSERT INTO sys_config_schema (module_key, field_key, field_type, default_value, validator, scope, description, created_at, updated_at) VALUES
('system_cfg', 'build_version', 'string', '0.0.0-dev', NULL, 'all', '构建版本号，通过 CI/CD 注入', NOW(), NOW()),
('health_cfg', 'max_depth', 'int', '2', '{"min":0,"max":3}', 'all', '健康检查最大深度：0=仅自身存活，1=依赖 ping，2=依赖+实际查询', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  updated_at = NOW();

-- ============================================================
-- 2. 多语言文案注入（sys_lang_string）
-- 使用 ICU plural 模板演示高级 i18n 能力
-- ============================================================

INSERT INTO sys_lang_string (lang_key, template_type, params_schema, lang_code, template_text, description, created_at, updated_at) VALUES
('svc_health_status_summary', 'icu', '[{"name":"healthy","type":"int","required":true},{"name":"total","type":"int","required":true}]', 'zh-CN', '{healthy, plural,\n  =0 {所有依赖均不可用（{total} 项）}\n  other {# / {total} 项依赖正常}\n}', 'Health 模块中文状态摘要（ICU plural 模板）', NOW(), NOW()),
('svc_health_status_summary', 'icu', '[{"name":"healthy","type":"int","required":true},{"name":"total","type":"int","required":true}]', 'en', '{healthy, plural,\n  =0 {All {total} dependencies are down}\n  other {# of {total} dependencies healthy}\n}', 'Health 模块英文状态摘要（ICU plural 模板）', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  updated_at = NOW();

-- ============================================================
-- 3. 验证查询（可选，用于确认数据已正确插入）
-- ============================================================

-- SELECT * FROM sys_config_schema WHERE module_key IN ('system_cfg', 'health_cfg');
-- SELECT * FROM sys_lang_string WHERE lang_key = 'svc_health_status_summary';
