// Package config_admin 提供 go-admin 插件层的配置管理 API
//
// 职责：
//   - 接收 HTTP 请求（Gin handler），转换为 DTO
//   - 委托 admin.AdminConfigService 执行业务逻辑
//   - 返回 go-admin 标准响应格式
//   - 接入 RBAC 权限控制（Casbin）
//   - 记录操作日志到 sys_oper_log
//
// 不负责：
//   - 字段级校验（由 services/config/admin 负责）
//   - 缓存失效 / pub/sub 广播（由 services/config/admin 负责）
//   - 直接操作 sys_config_schema / sys_config_version 表
//
// 架构约束（PRD-10 §10 + M2' grep 自检）：
//   - 本包内禁止出现 "sys_config_schema" / "sys_config_version" 字面量
//   - 所有写操作必须通过 AdminConfigService 代理
//
// 路由前缀：/api/admin/v1/config/schema/* 和 /api/admin/v1/config/value/*
// RBAC 权限点：config:schema:read/write/delete、config:value:read/write
//
// 相关文档：
//   - PRD-10 §十二 M2'：go-admin config_admin 插件规格
//   - M1' 交付物：services/config/admin 包
package config_admin
