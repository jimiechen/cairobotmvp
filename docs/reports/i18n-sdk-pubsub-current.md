# i18n SDK Pub/Sub 现状摘要

> 生成日期：2026-05-26
> 对比基准：config SDK 已完成的 InvalidateEvent 升级（M0'.0.7-B）
> 目的：为 M0'.0.7-C 决策提供输入，供项目主控选择升级路径（方案 A/B/C）

## 1. 文件位置

| 文件 | 路径 |
|---|---|
| pub/sub 实现 | `go/services/i18n/sdk/pubsub.go` |
| RedisClient 接口定义 | `go/services/i18n/sdk/pubsub.go:33-40` |
| onMessage 处理逻辑 | `go/services/i18n/sdk/pubsub.go:108-121` |
| LRU 缓存（含 InvalidateByLangCode） | `go/services/i18n/sdk/cache_lru.go` |
| 现有测试 | `go/services/i18n/sdk/pubsub_test.go` |

## 2. 架构差异对比

### 2.1 Redis 抽象层

| 维度 | config SDK | i18n SDK |
|---|---|---|
| 接口类型 | `redisx.Client`（第三方库统一接口） | 自定义 `RedisClient`（包内接口） |
| 接口文件 | `go/third_party/redisx/redisx.go` | `go/services/i18n/sdk/pubsub.go:33-40` |
| Get/Set/Delete | ✅ | ✅ |
| Subscribe 签名 | `Subscribe(channel, handler MessageHandler) (CancelFunc, error)` | `Subscribe(channel string, handler func(string)) (func(), error)` |
| Publish 签名 | `Publish(ctx, channel, msg) error`（需 ctx） | `Publish(channel, message string) error`（无 ctx） |
| Scan | ✅ | ❌ |
| Invalidate | ✅（M0'.0.5 新增） | ❌ |
| Close | ✅ | ✅ |

**关键差异：**
- i18n SDK 的 `Subscribe` handler 是 `func(string)` 原始签名，config SDK 使用了 `MessageHandler` 类型别名
- i18n SDK 的 `Publish` **没有 context.Context 参数**
- i18n SDK **没有 Scan 方法**，无法直接使用 redisx 的批量失效能力

### 2.2 Channel 常量

| SDK | Channel 名称 |
|---|---|
| config SDK | `cairobot.config.invalidate` |
| i18n SDK | `cairobot.i18n.invalidate` |

两个 SDK 使用 **不同的 channel**，互不干扰。

### 2.3 消息处理逻辑

**config SDK（已升级）：**

```go
// 三级降级策略
func (p *pubsubManager) onMessage(msg string) {
    var evt InvalidateEvent
    if err := json.Unmarshal([]byte(msg), &evt); err == nil {
        if evt.TenantID != "" { p.handleStructured(evt); return }
        if len(evt.ModuleKeys) > 0 { p.handleStructured(evt); return } // 无 tenant_id 但有 keys
    }
    // 降级到逗号分隔
    moduleKeys := strings.Split(msg, ",")
    p.handleLegacy(moduleKeys)
}
```

**i18n SDK（当前原始状态）：**

```go
// 仅支持逗号分隔格式
func (p *pubSubClient) onMessage(msg string, watchers *watcherManager, cache *lruCache) {
    langCodes := strings.Split(msg, ",")
    for _, langCode := range langCodes {
        langCode = strings.TrimSpace(langCode)
        if langCode == "" { continue }
        cache.InvalidateByLangCode(langCode)
        watchers.Trigger(langCode, 0)
    }
}
```

### 2.4 缓存失效方式

| 维度 | config SDK | i18n SDK |
|---|---|---|
| 缓存 Key 类型 | `string`（"sdk:{module_key}"） | `cacheKey` struct（env:langCode:version） |
| 失效方法 | `cache.delete(cacheKey)` 精确删除单条 | `cache.InvalidateByLangCode(langCode)` 遍历匹配 |
| 匹配策略 | 按 module_key 前缀精确匹配 | 按 langCode 字段遍历所有条目 |
| 批量失效 | 逐 key 删除 | 遍历全表 O(n) |

### 2.5 Watcher 回调

| 维度 | config SDK | i18n SDK |
|---|---|---|
| 触发方法 | `watcher.notify(key, snapshot)` | `watchers.Trigger(langCode, packVersion)` |
| 回调参数 | `*ModuleSnapshot`（含 Fields） | `int64`（仅版本号） |
| 注册 API | `register(moduleKey, func(*ModuleSnapshot))` | 内部 map 操作 |

## 3. 现有测试覆盖

i18n SDK `pubsub_test.go` 已有 **11 个测试**：

| 测试名 | 覆盖场景 | 状态 |
|---|---|---|
| TestPubSubClient_New | 构造函数 | PASS |
| TestPubSubClient_StartStop | 启停生命周期 | PASS |
| TestPubSubClient_Start_NoRedisConfig | 缺少 Redis 配置报错 | PASS |
| TestPubSubClient_Publish | 发布消息 | PASS |
| TestPubSubClient_Publish_NotInitialized | 未初始化发布报错 | PASS |
| TestPubSubClient_OnMessage | 单语言码消息处理 | PASS |
| TestPubSubClient_OnMessage_Batch | 多语言码逗号分隔 | PASS |
| TestPubSubClient_IsConnected_False | 初始未连接 | PASS |
| TestInvalidateChannel_Constant | Channel 常量值 | PASS |
| TestRedisConfig_DefaultValues | 配置默认值 | PASS |
| TestPubSubClient_Stop_NilRedis | nil redis 停止 | PASS |
| TestGoRedisClient_Interface | 接口实现验证 | PASS |

**缺失覆盖：**
- JSON 格式消息解析（InvalidateEvent 格式）
- 空 / 异常消息不 panic
- 无效 JSON 降级行为

## 4. 升级路径选项

### 方案 A：完全对齐 config SDK（推荐用于 M2' 之后）

**改动范围：**
1. 将 i18n SDK 的 `RedisClient` 替换为 `redisx.Client` + `redisx.PubSubClient`
2. 复用 `InvalidateEvent` 类型（或创建 i18n 专用子类型）
3. `onMessage` 升级为三级降级（与 config SDK 一致）
4. 补充 JSON 格式测试用例

**优点：**
- 两套 SDK 行为完全一致，维护成本最低
- 复用 redisx 基础设施（Scan、Invalidate、Publish）
- 未来新增 SDK 可直接复制模式

**缺点：**
- 改动较大，涉及接口替换和所有调用点
- Publish 签名变化（增加 ctx）影响上层调用方
- 需要同步修改 service 层的 Redis 注入方式

**工作量估计：** 中等（~200 行改动 + ~8 个新测试）

### 方案 B：最小改动（仅升级 onMessage 解析）

**改动范围：**
1. 在现有 `onMessage` 中添加 JSON 尝试解析
2. 降级后仍走原有 `strings.Split` + `InvalidateByLangCode`
3. 不改接口、不改 Redis 抽象层
4. 补充 3~4 个 JSON 相关测试

**优点：**
- 改动极小（~30 行）
- 零风险，不影响现有功能
- 可快速交付

**缺点：**
- 与 config SDK 的 Redis 抽象层仍不一致
- 未来如果要统一基础设施，还是要做方案 A
- Publish 签名仍然不同（无 ctx）

**工作量估计：** 小（~30 行 + ~4 个新测试）

### 方案 C：暂不升级（仅 Admin 发送端使用 JSON）

**改动范围：**
1. i18n SDK 保持现状不变
2. Admin 后端发送时同时发两种格式（JSON 给未来版，逗号分隔给当前版）
3. 或仅在 Admin 端发 JSON，i18n SDK 通过"无效 JSON 降级"路径自动兼容

**优点：**
- 零改动
- 当前功能不受影响

**缺点：**
- 技术债持续累积
- 两套 SDK 行为差异永远存在
- 违背 PRD §9 的 pub/sub 协议统一要求

**工作量估计：** 零

## 5. 建议

**M0' 阶段推荐方案 B**（最小改动），原因：
1. M0' 的目标是打通基础链路，验证可行性
2. 方案 B 可在 30 分钟内完成，不阻塞后续批次
3. 完整对齐（方案 A）留到 M2'（Admin 写入层实现时）统一进行

**M2' 阶段应执行方案 A**，在实现 `services/i18n/admin/` 写入包时一并完成接口对齐。

## 6. 风险项

| 风险 ID | 等级 | 描述 | 建议 |
|---|---|---|---|
| R-I18N-1 | R2 | i18n SDK Publish 无 ctx 参数，与 redisx.PubSubClient.Publish 签名不一致 | M2' 统一时补上 ctx |
| R-I18N-2 | R2 | InvalidateByLangCode 是 O(n) 遍历，大量语言包时性能堪忧 | M2' 时考虑改为索引结构或前缀匹配 |
| R-I18N-3 | R3 | watcher.Trigger 只传版本号不传完整数据，与 config SDK 的 notify(snapshot) 语义不同 | M2' 时评估是否需要对齐 |
