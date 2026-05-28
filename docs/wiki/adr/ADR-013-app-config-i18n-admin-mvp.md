# ADR-013: App Config & I18n Admin MVP 架构决策

## 状态

已接受（2026-05-27）

## 背景

CaiRobot MVP 项目需要新增管理后台能力：
- 配置管理：Schema 定义 + 值发布 + 版本追踪
- 国际化管理：字符串 CRUD + 语言包发布/回滚 + CSV 批量导入导出

历史方案 `provider-admin`（基于 go-admin 团队旧版）存在以下问题：
- 与项目现有服务层（services/config, services/i18n）耦合过深
- 直接操作数据库表（sys_config_schema, sys_lang_string），绕过业务校验层
- 无法独立测试，Mock 困难
- 前端与后端强绑定，无法替换 UI 框架

## 决策

采用 **「Admin Boundary Pattern」**（管理边界模式）：

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────┐
│  go-admin   │────▶│  {config,i18n}/  │────▶│  {config,i18n}/ │
│   plugin    │     │     admin/       │     │    service/   │
│ (HTTP handler│     │  (写层+审计+缓存) │     │  (校验层)      │
│  DTO转换)   │     └──────────────────┘     └──────────────┘
└─────────────┘              │                       │
                            ▼                       ▼
                    ┌──────────────┐        ┌──────────────┐
                    │  repository/ │        │   domain/    │
                    │  (落库层)    │        │  (领域模型)   │
                    └──────────────┘        └──────────────┘
```

核心原则：

1. **admin 插件层只做 HTTP 协议适配**（接收请求→DTO转换→调用 admin 服务接口→返回响应）
2. **admin 服务层负责业务编排**（校验→落库→审计→缓存失效→pub/sub 广播）
3. **service 层只做纯校验**（ValidateConfigMap / ValidateTemplate，无副作用）
4. **所有写操作必须经过 admin 服务代理**，禁止插件直连 repository

## 详细设计

### 3.1 接口驱动注入（Interface-based DI）

```go
// admin 包定义统一接口，插件层依赖接口而非具体结构体
type ConfigAdminService interface {
    ConfigSchemaService  // CreateSchema/UpdateSchema/DeleteSchema/ListSchemas
    ConfigValueService    // PublishValue/GetValueVersions
}

type I18nAdminService interface {
    I18nStringService    // CreateString/UpdateString/DeleteString/ListStrings
    I18nPackService      // PublishPack/RollbackPack/ImportStringsFromCSV/ExportStringsToCSV
}
```

优势：
- 插件层测试可用 fake 实现（fakeConfigSvc / fakeI18nSvc），无需启动 Redis/DB
- admin 服务层测试可 Mock repository 和 service 层
- 未来替换 UI 框架（如从 gin 切换到 echo）不影响业务逻辑

### 3.2 本地 Publisher 接口解耦

```go
// config/admin 和 i18n/admin 各自定义 Publisher 接口
type Publisher interface {
    Publish(ctx context.Context, topic string, payload []byte) error
}
```

避免直接依赖 `redisx.PubSubClient` 具体类型，使测试可以注入 NoopPublisher。

### 3.3 10400 校验错误码协议

后端统一格式：

```json
{
    "code": 10400,
    "message": "参数校验失败",
    "errors": [
        {"field": "port", "message": "不能大于 65535"},
        {"field": "timeout", "message": "不能为负数"}
    ]
}
```

前端处理：
- HTTP Handler 层检测 error message 包含「模板」或「template」关键字 → 返回 400 + 10400 body
- Vue 组件展示字段级错误提示（红色文字）+ 弹窗汇总详情

### 3.4 AES-256-CBC DSN 加密

数据库连接串不以明文存储在环境变量中：

```go
// 加密流程：plainDSN → PKCS7 padding → CBC encrypt → Base64 编码
// 解密流程：Base64 decode → CBC decrypt → PKCS7 unpadding → plainDSN
// 密钥来源：ADMIN_DSN_KEY 环境变量（32 字节 hex）
```

### 3.5 Pub/Sub InvalidateEvent JSON Payload

替代旧的逗号分隔格式，采用结构化 JSON：

```json
{
    "tenant_id": "default",
    "scope": "config",
    "module_keys": ["app.server", "app.cache"]
}
```

三阶向下兼容策略：
1. 有 tenant_id → 结构化处理
2. 无 tenant_id 但有 module_keys → 结构化处理（无租户过滤）
3. 非 JSON → 逗号分隔降级（向后兼容旧 SDK）

## 后果

### 正面影响

| 维度 | 效果 |
|------|------|
| 可测试性 | 插件层 32 个测试全部使用 fake 实现，无需外部依赖 |
| 可维护性 | admin 服务与 provider-admin 完全隔离，可独立演进 |
| 安全性 | 所有写操作经校验层 + 审计日志 + 权限中间件 |
| 可扩展性 | 接口驱动设计支持替换底层实现（如从 SQLite 切换 MySQL） |
| 向后兼容 | Pub/Sub payload 双格式兼容，不影响存量 SDK |

### 负面影响

| 维度 | 影响 | 应对 |
|------|------|------|
| 间接层数量 | 多了一层 admin 服务包装 | 单文件不超过 200 行，职责单一 |
| 学习成本 | 新成员需理解三层架构 | doc.go 文件头注释 + ADR 文档 |
| ImportStringsFromCSV reader 参数 | 使用 interface{} 而非 io.Reader | 标记为 P2 遗留风险，后续迭代修正 |

## 替代方案

### 方案 A：直接在 provider-admin 上扩展

- ❌ provider-admin 与 services 层耦合深，改动面大
- ❌ 无法独立测试
- ❌ 表名直写散落在各处，grep 自检无法拦截

### 方案 B：使用 gRPC 连接 services 层

- ✅ 进程隔离好
- ❌ 引入 gRPC 依赖复杂度高
- ❌ MVP 阶段过度工程化
- ❌ 本地开发需要额外启动 gRPC server

### 方案 C（采纳）：Admin Boundary Pattern

- ✅ 同进程内轻量级分层
- ✅ 接口驱动，测试友好
- ✅ grep 自检可强制边界约束
- ✅ 与现有代码风格一致（Go workspace module 隔离）

## 相关文档

- [PRD-10](../prd/PRD-10-Admin管理后台MVP.md)：完整需求规格
- [ADR-008](ADR-008-config-i18n-migration.md)：config/i18n SDK 迁移决策
- [测试报告](../testing/admin-mvp-test-report.md)：56/56 测试通过证据
- [go-admin 可行性报告](../reports/go-admin-feasibility-report.md)：技术选型论证
