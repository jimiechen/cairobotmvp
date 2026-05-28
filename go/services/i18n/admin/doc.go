// Package admin 提供 i18n 模块的管理后台写入层
//
// 职责：
//   - 接收管理后台的语言字符串 CRUD 请求
//   - 委托 service 层执行模板校验（ValidateTemplate）
//   - 通过 repository 层落库
//   - 记录审计日志
//   - 触发缓存失效与变更广播
//
// 不负责：
//   - 模板校验逻辑（由 service.ValidateTemplate 负责）
//   - HTTP 路由 / 参数解析（由 go-admin 插件层负责）
//
// 架构约束（PRD-10 §10）：
//   - admin 子包持有 inner service.I18nService 引用（用于校验）
//   - admin 校验必须调用 inner.ValidateTemplate()，禁止自行实现模板解析
//   - admin 禁止出现模板参数提取 / 占位符替换等业务逻辑
//
// 相关文档：
//   - PRD-10 §十：admin 子包复用 service 层校验的精确实现
//   - docs/reports/i18n-sdk-pubsub-current.md：i18n SDK 升级路径分析
package admin
