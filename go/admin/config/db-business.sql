-- ============================================================
-- Admin MVP 业务种子数据
-- 用途：添加"配置中心"和"多语言管理"菜单 + 业务表 + 测试数据
-- 执行方式: cd go/admin && sqlite3 go-admin-db.db < config/db-business.sql
-- ============================================================

BEGIN TRANSACTION;

-- ==================== 1. 业务表 DDL ====================

-- 1.1 config_schema 配置字段定义表
CREATE TABLE IF NOT EXISTS config_schema (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    module_key    TEXT    NOT NULL DEFAULT '',
    field_key     TEXT    NOT NULL DEFAULT '',
    field_type    TEXT    NOT NULL DEFAULT 'string',
    default_value TEXT    NOT NULL DEFAULT '',
    validator     TEXT    NOT NULL DEFAULT '',
    is_required   INTEGER NOT NULL DEFAULT 0,
    is_secret     INTEGER NOT NULL DEFAULT 0,
    description   TEXT    NOT NULL DEFAULT '',
    client_scope  TEXT    NOT NULL DEFAULT '',
    sort_order    INTEGER NOT NULL DEFAULT 0,
    is_enabled    INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME,
    updated_at    DATETIME,
    deleted_at    DATETIME
);

-- 1.2 config_value 配置值发布表
CREATE TABLE IF NOT EXISTS config_value (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    version     INTEGER NOT NULL DEFAULT 1,
    module_key  TEXT    NOT NULL DEFAULT '',
    env         TEXT    NOT NULL DEFAULT 'dev',
    field_key   TEXT    NOT NULL DEFAULT '',
    value       TEXT    NOT NULL DEFAULT '',
    operator    TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
);

-- 1.3 i18n_pack 语言包表
CREATE TABLE IF NOT EXISTS i18n_pack (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    pack_name   TEXT    NOT NULL DEFAULT '',
    lang_code   TEXT    NOT NULL DEFAULT 'zh-CN',
    description TEXT    NOT NULL DEFAULT '',
    version     INTEGER NOT NULL DEFAULT 1,
    status      TEXT    NOT NULL DEFAULT 'active',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
);

-- 1.4 i18n_string 语言字符串表
CREATE TABLE IF NOT EXISTS i18n_string (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    pack_id        INTEGER NOT NULL DEFAULT 0,
    string_key     TEXT    NOT NULL DEFAULT '',
    string_value   TEXT    NOT NULL DEFAULT '',
    group_name     TEXT    NOT NULL DEFAULT '',
    template_type  TEXT    NOT NULL DEFAULT 'plain',
    operation_type TEXT    NOT NULL DEFAULT '',
    params_schema  TEXT    NOT NULL DEFAULT '',
    preview_sample TEXT    NOT NULL DEFAULT '',
    version        INTEGER NOT NULL DEFAULT 1,
    created_at     DATETIME,
    updated_at     DATETIME,
    deleted_at     DATETIME
);


-- ==================== 2. sys_api 接口注册 ====================
-- 从 200 开始分配业务 API ID，避免与系统 API 冲突（系统最大 ~135）

INSERT INTO sys_api VALUES (200, 'go-admin/app/admin/config_admin/apis.SchemaApi.GetSchemaList-fm', '配置Schema列表', '/api/admin/v1/config/schema', 'BUS', 'GET', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (201, 'go-admin/app/admin/config_admin/apis.SchemaApi.CreateSchema-fm', '新增配置Schema', '/api/admin/v1/config/schema', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (202, 'go-admin/app/admin/config_admin/apis.SchemaApi.UpdateSchema-fm', '更新配置Schema', '/api/admin/v1/config/schema', 'BUS', 'PUT', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (203, 'go-admin/app/admin/config_admin/apis.SchemaApi.DeleteSchema-fm', '删除配置Schema', '/api/admin/v1/config/schema', 'BUS', 'DELETE', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (204, 'go-admin/app/admin/config_admin/apis.ValueApi.PublishValue-fm', '发布配置值', '/api/admin/v1/config/value/publish', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (205, 'go-admin/app/admin/config_admin/apis.ValueApi.GetValueVersions-fm', '配置版本列表', '/api/admin/v1/config/value/versions', 'BUS', 'GET', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (210, 'go-admin/app/admin/i18n_admin/apis.StringApi.ListStrings-fm', '字符串列表', '/api/admin/v1/i18n/string', 'BUS', 'GET', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (211, 'go-admin/app/admin/i18n_admin/apis.StringApi.CreateString-fm', '新增字符串', '/api/admin/v1/i18n/string', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (212, 'go-admin/app/admin/i18n_admin/apis.StringApi.UpdateString-fm', '更新字符串', '/api/admin/v1/i18n/string', 'BUS', 'PUT', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (213, 'go-admin/app/admin/i18n_admin/apis.StringApi.DeleteString-fm', '删除字符串', '/api/admin/v1/i18n/string', 'BUS', 'DELETE', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (214, 'go-admin/app/admin/i18n_admin/apis.PackApi.PublishPack-fm', '发布语言包', '/api/admin/v1/i18n/pack/publish', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (215, 'go-admin/app/admin/i18n_admin/apis.PackApi.RollbackPack-fm', '回滚语言包', '/api/admin/v1/i18n/pack/rollback', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (216, 'go-admin/app/admin/i18n_admin/apis.ImportExportApi.ImportStringsFromCSV-fm', 'CSV导入', '/api/admin/v1/i18n/import/csv', 'BUS', 'POST', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);
INSERT INTO sys_api VALUES (217, 'go-admin/app/admin/i18n_admin/apis.ExportStringsToCSV-fm', 'CSV导出', '/api/admin/v1/i18n/export/csv', 'BUS', 'GET', '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL, 0, 0);


-- ==================== 3. sys_menu 菜单注册 ====================
-- 从 600 开始分配 MenuId，避免与系统菜单冲突（系统最大 542）

-- 3.1 配置中心（一级目录 M）
INSERT INTO sys_menu VALUES (600, 'ConfigCenter', '配置中心', 'setting', '/config-center', '/0/600', 'M', '无', '', 0, false, '', 'Layout', 15, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);

-- 3.2 配置中心子菜单
INSERT INTO sys_menu VALUES (601, 'ConfigSchemaList', 'Schema列表', 'list', '/config/schema-list', '/0/600/601', 'C', '无', 'config:schema:list', 600, false, '', '/config/schema-list', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (602, 'ConfigValuePublish', '配置发布', 'upload', '/config/value-publish', '/0/600/602', 'C', '无', 'config:value:list', 600, false, '', '/config/value-publish', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (603, 'ConfigVersionHistory', '版本历史', 'time-range', '/config/version-history', '/0/600/603', 'C', '无', 'config:version:list', 600, false, '', '/config/version-history', 30, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);

-- 3.3 配置中心按钮权限（F 类型）
INSERT INTO sys_menu VALUES (610, '', '查询Schema', 'app-group-fill', '', '/0/600/601/610', 'F', 'GET', 'config:schema:list', 601, false, '', '', 40, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (611, '', '新增Schema', 'app-group-fill', '', '/0/600/601/611', 'F', 'POST', 'config:schema:add', 601, false, '', '', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (612, '', '修改Schema', 'app-group-fill', '', '/0/600/601/612', 'F', 'PUT', 'config:schema:edit', 601, false, '', '', 30, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (613, '', '删除Schema', 'app-group-fill', '', '/0/600/601/613', 'F', 'DELETE', 'config:schema:delete', 601, false, '', '', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (614, '', '发布配置', 'app-group-fill', '', '/0/600/602/614', 'F', 'POST', 'config:value:publish', 602, false, '', '', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (615, '', '查询版本', 'app-group-fill', '', '/0/600/603/615', 'F', 'GET', 'config:version:list', 603, false, '', '', 40, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);

-- 3.4 多语言管理（一级目录 M）
INSERT INTO sys_menu VALUES (700, 'I18nManage', '多语言管理', 'language', '/i18n', '/0/700', 'M', '无', '', 0, false, '', 'Layout', 16, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);

-- 3.5 多语言管理子菜单
INSERT INTO sys_menu VALUES (701, 'I18nStringManage', '字符串管理', 'edit-pen', '/i18n/string-list', '/0/700/701', 'C', '无', 'i18n:string:list', 700, false, '', '/i18n/string-list', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (702, 'I18nPackManage', '语言包管理', 'folder-opened', '/i18n/pack-manage', '/0/700/702', 'C', '无', 'i18n:pack:list', 700, false, '', '/i18n/pack-manage', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (703, 'I18nImportExport', 'CSV导入导出', 'document', '/i18n/import-export', '/0/700/703', 'C', '无', 'i18n:import:list', 700, false, '', '/i18n/import-export', 30, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);

-- 3.6 多语言管理按钮权限（F 类型）
INSERT INTO sys_menu VALUES (710, '', '查询字符串', 'app-group-fill', '', '/0/700/701/710', 'F', 'GET', 'i18n:string:list', 701, false, '', '', 40, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (711, '', '新增字符串', 'app-group-fill', '', '/0/700/701/711', 'F', 'POST', 'i18n:string:add', 701, false, '', '', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (712, '', '修改字符串', 'app-group-fill', '', '/0/700/701/712', 'F', 'PUT', 'i18n:string:edit', 701, false, '', '', 30, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (713, '', '删除字符串', 'app-group-fill', '', '/0/700/701/713', 'F', 'DELETE', 'i18n:string:delete', 701, false, '', '', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (714, '', '发布语言包', 'app-group-fill', '', '/0/700/702/714', 'F', 'POST', 'i18n:pack:publish', 702, false, '', '', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (715, '', '回滚语言包', 'app-group-fill', '', '/0/700/702/715', 'F', 'POST', 'i18n:pack:rollback', 702, false, '', '', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (716, '', 'CSV导入', 'app-group-fill', '', '/0/700/703/716', 'F', 'POST', 'i18n:import:add', 703, false, '', '', 10, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);
INSERT INTO sys_menu VALUES (717, '', 'CSV导出', 'app-group-fill', '', '/0/700/703/717', 'F', 'GET', 'i18n:export:list', 703, false, '', '', 20, '0', '1', 1, 1, '2026-05-29 00:00:00.000', '2026-05-29 00:00:00.000', NULL);


-- ==================== 4. sys_menu_api_rule 菜单-接口关联 ====================
-- 配置中心菜单 → API 关联
INSERT INTO sys_menu_api_rule VALUES (601, 200);
INSERT INTO sys_menu_api_rule VALUES (601, 205);
INSERT INTO sys_menu_api_rule VALUES (610, 200);
INSERT INTO sys_menu_api_rule VALUES (611, 201);
INSERT INTO sys_menu_api_rule VALUES (612, 202);
INSERT INTO sys_menu_api_rule VALUES (613, 203);
INSERT INTO sys_menu_api_rule VALUES (602, 204);
INSERT INTO sys_menu_api_rule VALUES (602, 205);
INSERT INTO sys_menu_api_rule VALUES (614, 204);
INSERT INTO sys_menu_api_rule VALUES (603, 205);
INSERT INTO sys_menu_api_rule VALUES (615, 205);

-- 多语言管理菜单 → API 关联
INSERT INTO sys_menu_api_rule VALUES (701, 210);
INSERT INTO sys_menu_api_rule VALUES (701, 211);
INSERT INTO sys_menu_api_rule VALUES (701, 212);
INSERT INTO sys_menu_api_rule VALUES (701, 213);
INSERT INTO sys_menu_api_rule VALUES (710, 210);
INSERT INTO sys_menu_api_rule VALUES (711, 211);
INSERT INTO sys_menu_api_rule VALUES (712, 212);
INSERT INTO sys_menu_api_rule VALUES (713, 213);
INSERT INTO sys_menu_api_rule VALUES (702, 214);
INSERT INTO sys_menu_api_rule VALUES (702, 215);
INSERT INTO sys_menu_api_rule VALUES (714, 214);
INSERT INTO sys_menu_api_rule VALUES (715, 215);
INSERT INTO sys_menu_api_rule VALUES (703, 216);
INSERT INTO sys_menu_api_rule VALUES (703, 217);
INSERT INTO sys_menu_api_rule VALUES (716, 216);
INSERT INTO sys_menu_api_rule VALUES (717, 217);


-- ==================== 5. 角色菜单关联（管理员角色 ID=1 获得全部菜单）====================
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 600);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 601);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 602);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 603);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 610);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 611);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 612);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 613);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 614);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 615);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 700);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 701);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 702);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 703);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 710);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 711);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 712);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 713);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 714);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 715);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 716);
INSERT OR IGNORE INTO sys_role_menu (role_id, menu_id) VALUES (1, 717);


-- ==================== 6. Casbin 权限规则（角色ID=1 管理员拥有所有业务权限）====================
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/config/schema', 'GET', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/config/schema', 'POST', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/config/schema', 'PUT', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/config/schema', 'DELETE', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/config/value/*', '*', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/i18n/string', 'GET', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/i18n/string', 'POST', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/i18n/string', 'PUT', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/i18n/string', 'DELETE', '', '', '');
INSERT OR IGNORE INTO casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) VALUES (NULL, 'p', '1', '/api/admin/v1/i18n/*', '*', '', '', '');


-- ==================== 7. 业务测试数据 ====================

-- 7.1 config_schema 测试数据（5条，匹配 E2E 测试期望）
INSERT INTO config_schema (module_key, field_key, field_type, default_value, validator, is_required, is_secret, description, client_scope, sort_order, is_enabled, created_at, updated_at) VALUES
('app.server', 'port', 'int', '8080', 'range(1,65535)', 1, 0, '服务监听端口', 'all', 1, 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
('app.server', 'host', 'string', '0.0.0.0', '', 1, 0, '服务绑定地址', 'all', 2, 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
('app.server', 'mode', 'string', 'dev', 'in(dev,test,prod)', 1, 0, '运行模式', 'all', 3, 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
('app.database', 'max_open_conns', 'int', '100', 'range(1,1000)', 0, 0, '数据库最大连接数', 'backend', 1, 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
('app.database', 'max_idle_conns', 'int', '10', 'range(1,100)', 0, 0, '数据库最大空闲连接数', 'backend', 2, 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00');

-- 7.2 i18n_pack 测试数据（2个语言包）
INSERT INTO i18n_pack (pack_name, lang_code, description, version, status, created_at, updated_at) VALUES
('Web前端中文包', 'zh-CN', 'Web前端界面中文语言包', 1, 'active', '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
('Web前端英文包', 'en-US', 'Web前端界面英文语言包', 1, 'active', '2026-05-29 00:00:00', '2026-05-29 00:00:00');

-- 7.3 i18n_string 测试数据（5条，匹配 E2E 测试期望）
INSERT INTO i18n_string (pack_id, string_key, string_value, group_name, template_type, operation_type, params_schema, preview_sample, version, created_at, updated_at) VALUES
(1, 'greeting.hello', '你好，世界！', '通用', 'plain', '', '', '', 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
(1, 'greeting.goodbye', '再见！', '通用', 'plain', '', '', '', 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
(1, 'error.not_found', '资源未找到：{resourceId}', '错误提示', 'named', '', '{"resourceId":"string"}', '资源未找到：USER-001', 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
(1, 'auth.login_title', '请登录', '认证', 'plain', '', '', '', 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00'),
(1, 'common.confirm', '确认', '通用', 'plain', '', '', '', 1, '2026-05-29 00:00:00', '2026-05-29 00:00:00');

COMMIT;
