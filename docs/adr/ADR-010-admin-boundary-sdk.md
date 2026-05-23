# ADR-010: Admin 边界与 SDK 引用规范

## 状态

已采纳

## 背景

CaiRobot MVP 需要明确 admin-server（管理后台）和业务服务（如 AI 服务、设备网关等）的职责边界，避免循环依赖和职责混乱。

当前存在的问题：

1. **边界模糊**：admin-server 是否应该提供读 API 给业务服务调用？
2. **性能问题**：业务服务每次需要配置时都调 admin HTTP API，延迟高、耦合重
3. **缓存一致性**：多个业务服务缓存同一份数据，如何保证一致性？
4. **依赖方向**：如果业务服务依赖 admin，那 admin 又可能依赖业务服务（如用户权限校验），形成循环

## 决策

### 1. Admin 职责边界

**Admin 只负责写路径（CRUD），不提供读 API 给业务服务**：

| 操作类型 | 调用方 | 说明 |
|---|---|---|
| Schema 创建/更新/删除 | 运营人员 → Admin HTTP | 管理配置元数据 |
| 配置值写入 | 运营人员 → Admin HTTP | 管理配置实际值 |
| 多语言翻译录入 | 运营人员 → Admin HTTP | 管理翻译内容 |
| 配置读取 | 业务服务 → SDK | 通过 SDK 读取，不经过 Admin |

### 2. SDK 引用架构

业务服务通过 SDK（configsdk / i18nsdk）引用配置和多语言能力：

```text
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  业务服务 A   │     │  业务服务 B   │     │  业务服务 C   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────┬───────┴───────────────────┘
                   │
          ┌────────▼────────┐
          │   configsdk     │  ← L1 LRU 缓存
          │   i18nsdk       │
          └────────┬────────┘
                   │
          ┌────────▼────────┐
          │    Redis        │  ← L2 共享缓存
          │   (pub/sub)     │
          └────────┬────────┘
                   │
          ┌────────▼────────┐
          │ Config Service  │  ← L3 远程兜底
          │ I18n Service    │
          └─────────────────┘
```

### 3. SDK 三层缓存架构

| 层级 | 存储 | TTL | 说明 |
|---|---|---|---|
| L1 | 进程内 LRU | 5s | 单实例缓存，零网络开销 |
| L2 | Redis | 60s | 跨实例共享，支持 pub/sub 失效 |
| L3 | 远程服务调用 | - | 兜底保障，Redis 不可用时直连 |

### 4. 缓存失效机制

Admin 写入操作后主动触发缓存失效：

```text
Admin 写入配置
    ↓
1. 持久化到 DB（SQLite/MySQL）
2. 删除 Redis 对应 key
3. 发布 pub/sub 失效通知
    ↓
各业务服务 SDK 收到通知
    ↓
清除 L1 本地缓存
```

## 后果

### 正面影响

1. **边界清晰**：Admin 专注写操作，业务服务通过 SDK 读，职责单一
2. **解耦彻底**：业务服务不直接依赖 Admin HTTP API，可独立部署和扩缩容
3. **性能优异**：L1 缓存命中时延迟 < 1ms，无需跨网络调用
4. **可观测性强**：SDK 内置 metrics，可监控缓存命中率和调用延迟

### 负面影响

1. **SDK 维护成本**：需要维护缓存一致性逻辑、重试机制、熔断降级
2. **多语言 SDK 同步**：如果未来有 Python/TS 业务服务，需要实现对应 SDK
3. **最终一致性**：写入后短暂延迟（通常 < 100ms）才能在所有实例生效

## 替代方案

### 方案 A: 业务直接调 Admin 读 API

业务服务每次需要配置时，直接 HTTP 调用 Admin 的读接口。

**否决原因**：
- **耦合严重**：业务服务强依赖 Admin 的 HTTP 接口格式
- **性能差**：每次读取都需要跨网络 HTTP 调用，延迟 10-50ms
- **无法水平扩展**：Admin 成为所有业务的瓶颈
- **单点故障**：Admin 不可用时，所有业务功能受影响

### 方案 B: 共享数据库

业务服务直接查询配置数据库。

**否决原因**：
- **违反分层原则**：业务服务不应知道存储细节
- **安全性差**：数据库连接信息暴露给所有服务
- **无法加缓存层**：每个服务自行实现缓存，难以保证一致性

## 相关文档

- [CODE-WIKI.md](../wiki/CODE-WIKI.md) §13.2 Go 语言资产目录树 — sdk/ 目录说明
- [ADR-009-config-i18n-schema-template.md](ADR-009-config-i18n-schema-template.md) — Schema Registry 架构
- [services/config/sdk](../../go/services/config/sdk) — configsdk 实现
- [services/i18n/sdk](../../go/services/i18n/sdk) — i18nsdk 实现

## 参考实现

- [tars/provider-admin](../../go/tars/provider-admin) — Admin Gin HTTP 服务
- [go/services/config/sdk](../../go/services/config/sdk) — configsdk 三层缓存
- [go/services/i18n/sdk](../../go/services/i18n/sdk) — i18nsdk 三层缓存
