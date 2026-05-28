// Package i18n_admin 提供 go-admin 插件层的国际化管理 API
//
// 职责：
//   - 接收 HTTP 请求（Gin handler），转换为 DTO
//   - 委托 admin.AdminI18nService 执行业务逻辑
//   - 返回 go-admin 标准响应格式
//   - 接入 RBAC 权限控制（Casbin）
//
// 不负责：
//   - 模板校验（由 services/i18n/admin 负责）
//   - 缓存失效 / pub/sub 广播（由 services/i18n/admin 负责）
//   - 直接操作 sys_lang_pack / sys_lang_string 表
//
// 架构约束（PRD-10 §10 + M3' grep 自检）：
//   - 本包内禁止出现 "sys_lang_pack" / "sys_lang_string" 字面量
//   - 所有写操作必须通过 AdminI18nService 代理
//
// 路由前缀：/api/admin/v1/i18n/string/* 和 /api/admin/v1/i18n/pack/*
// RBAC 权限点：i18n:string:read/write/delete、i18n:pack:read/write
//
// 相关文档：
//   - PRD-10 §十二 M3'：go-admin i18n_admin 插件规格
//   - M1' 交付物：services/i18n/admin 包
package i18n_admin
