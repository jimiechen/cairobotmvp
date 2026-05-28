// Package admin 提供 config 模块的管理后台写入层
//
// 职责：
//   - 接收管理后台的 CRUD 请求（Schema 定义 / 配置值发布）
//   - 委托 service 层执行校验（ValidateFieldValue / ValidateConfigMap）
//   - 通过 repository 层落库
//   - 记录审计日志
//   - 触发缓存失效（redisx.Client.Invalidate）
//   - 广播变更通知（redisx.PubSubClient.Publish + InvalidateEvent JSON）
//
// 不负责：
//   - 字段级校验逻辑（由 service.ValidateFieldValue / Validate.ValidateConfigMap 负责）
//   - HTTP 路由 / 参数解析（由 go-admin 插件层负责）
//   - SDK 端缓存刷新（由 SDK pub/sub consumer 负责）
//
// 架构约束（PRD-10 §10）：
//   - admin 子包持有 inner *service.SchemaService 引用
//   - admin 校验必须调用 inner 包级函数，禁止自行实现字段级校验
//   - admin 禁止出现 field_type== / switch field_type / validator JSON 解析
//
// 相关文档：
//   - PRD-10 §十：admin 子包复用 service 层校验的精确实现
//   - ADR-010：Admin 边界与 SDK 引用规范
package admin
