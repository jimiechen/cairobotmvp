# Gateway internal/server 编译错误 API 兼容性分析报告

## 1. 问题概述

**问题现象**: `go/gateway/proto-gateway/internal/server` 包编译失败：

```
internal/server/e2e_config_i18n_test.go:35:27: undefined: repository.NewSchemaRepo
internal/server/e2e_config_i18n_test.go:36:20: undefined: cache.NewLRUCache
internal/server/e2e_config_i18n_test.go:40:28: undefined: i18ncache.NewLRUCache
```

**影响范围**: `go/gateway/proto-gateway/internal/server/e2e_config_i18n_test.go` 测试文件无法编译

**严重程度**: P2 - 影响 E2E 测试，不阻断主流程

---

## 2. 根因分析

### 2.1 问题本质

**E2E 测试代码使用了已不存在的 API**，与 config/i18n 服务的实际实现不匹配。

这是 **TDD 开发过程中的正常中间状态**：测试先写，但对应的实现 API 尚未完成或已重构。

### 2.2 具体 API 不匹配项

| 测试代码使用的 API | 实际存在的 API | 状态 |
|-------------------|---------------|------|
| `repository.NewSchemaRepo(db)` | `repository.NewSQLiteSchemaRepo(db)` | ❌ 名称不匹配 |
| `cache.NewLRUCache(100)` | `sdk.newLRUCache(capacity, ttlSec)` | ❌ 包路径+可见性不匹配 |
| `i18ncache.NewLRUCache(100)` | `sdk.newLRUCache(capacity)` | ❌ 包路径+可见性不匹配 |

### 2.3 代码对比分析

#### 2.3.1 SchemaRepo

**测试代码期望** (`e2e_config_i18n_test.go:35`):
```go
schemaRepo := repository.NewSchemaRepo(db)
```

**实际实现** (`go/services/config/repository/schema_repo.go:28`):
```go
// NewSQLiteSchemaRepo 基于已有 DB 连接创建 Schema 仓库
func NewSQLiteSchemaRepo(db *sql.DB) *SQLiteSchemaRepo {
    return &SQLiteSchemaRepo{db: db}
}
```

**问题**: 函数名从 `NewSchemaRepo` 改为 `NewSQLiteSchemaRepo`，明确标识实现类型。

#### 2.3.2 Config LRU Cache

**测试代码期望** (`e2e_config_i18n_test.go:36`):
```go
lruCache := cache.NewLRUCache(100)
```

**实际实现** (`go/services/config/sdk/cache_lru.go:31`):
```go
// newLRUCache 创建 LRU 缓存实例（包内私有）
func newLRUCache(capacity int, ttlSec ...int) *lruCache {
    // ...
}
```

**问题**:
1. 函数在 `sdk` 包中，不在 `cache` 包
2. 函数是小写的 `newLRUCache`（包内私有），不是导出的 `NewLRUCache`
3. 参数不同：测试期望单参数，实际实现支持可变参数 `ttlSec ...int`

#### 2.3.3 I18n LRU Cache

**测试代码期望** (`e2e_config_i18n_test.go:40`):
```go
i18nLRUCache := i18ncache.NewLRUCache(100)
```

**实际实现** (`go/services/i18n/sdk/cache_lru.go:59`):
```go
// newLRUCache 创建 LRU 缓存实例（包内私有）
func newLRUCache(capacity int) *lruCache {
    // ...
}
```

**问题**:
1. 函数在 `sdk` 包中，不在 `cache` 包
2. 函数是小写的 `newLRUCache`（包内私有）
3. config 和 i18n 的 `newLRUCache` 签名不一致（config 有 ttlSec 参数，i18n 没有）

### 2.4 设计意图分析

#### 为什么 cache 构造函数是私有的？

查看 `go/services/config/sdk/client.go` 中的使用方式：

```go
// configClient 配置 SDK 客户端
type configClient struct {
    options ClientOptions
    cache   *lruCache  // 直接使用私有类型
    // ...
}

// NewClient 创建配置客户端（公开 API）
func NewClient(opts ClientOptions) (*configClient, error) {
    // ...
    cache := newLRUCache(opts.CacheCapacity, opts.CacheTTL)
    // ...
}
```

**设计意图**: 
- `lruCache` 是 SDK 内部实现细节，不对外暴露
- 用户通过 `NewClient` 创建客户端，缓存由 SDK 内部管理
- 符合封装原则，避免外部直接操作缓存

#### 为什么 SchemaRepo 改名？

```go
// 从：
func NewSchemaRepo(db *sql.DB) *SchemaRepo

// 改为：
func NewSQLiteSchemaRepo(db *sql.DB) *SQLiteSchemaRepo
```

**设计意图**:
- 明确标识是 SQLite 实现（区别于 MySQL 实现 `NewMySQLConfigRepo`）
- 符合命名规范：构造函数应体现具体实现类型

---

## 3. 影响评估

### 3.1 直接影响

- `internal/server/e2e_config_i18n_test.go` 无法编译
- 3 个 E2E 测试无法运行：
  - `TestE2E_GetAppConfigs_FullChain`
  - `TestE2E_GetLangPack_FullChain`
  - `TestE2E_NewFieldWithoutCodeChange`

### 3.2 间接影响

- Gateway 的 Config/I18n E2E 测试覆盖率缺失
- 无法验证端到端的配置/多语言链路

---

## 4. 解决方案

### 方案 A：修复 E2E 测试代码以匹配实际 API（推荐）

修改 `e2e_config_i18n_test.go`，使用正确的 API：

```go
// 修改前
schemaRepo := repository.NewSchemaRepo(db)
lruCache := cache.NewLRUCache(100)
i18nLRUCache := i18ncache.NewLRUCache(100)

// 修改后
schemaRepo := repository.NewSQLiteSchemaRepo(db)
// cache 需要通过 sdk.NewClient 创建，或使用 mock
// 或者提供测试专用的导出构造函数
```

**优点**:
- 符合实际 API 设计
- 保持封装性

**缺点**:
- 需要重构测试的依赖注入方式
- 可能需要暴露测试专用的构造函数

### 方案 B：为测试暴露导出构造函数

在 `sdk` 包中增加导出的测试构造函数：

```go
// go/services/config/sdk/export_test.go
//go:build test
package sdk

func NewLRUCacheForTest(capacity int, ttlSec ...int) *lruCache {
    return newLRUCache(capacity, ttlSec...)
}
```

**优点**:
- 保持生产代码的封装性
- 测试可以访问内部实现

**缺点**:
- 需要条件编译（build tag）
- 增加维护复杂度

### 方案 C：使用 Mock 替代真实依赖

E2E 测试中使用 Mock Cache 和 Mock Repo：

```go
mockCache := cache.NewMockCache()
mockSchemaRepo := repository.NewMockSchemaRepository()
```

**优点**:
- 测试更稳定（不依赖具体实现）
- 测试运行更快

**缺点**:
- E2E 测试变成集成测试，失去端到端验证价值

---

## 5. 推荐方案

**采用方案 A + 补充方案 B**：

1. **修改 E2E 测试**：使用正确的 API 名称（`NewSQLiteSchemaRepo`）
2. **提供测试导出函数**：为 `sdk` 包的私有构造函数提供 `export_test.go`
3. **统一 cache 接口**：考虑统一 config 和 i18n 的 cache 构造函数签名

---

## 6. 实施步骤

### 步骤 1: 修改 E2E 测试中的 SchemaRepo 调用

```go
// e2e_config_i18n_test.go
// 修改前
schemaRepo := repository.NewSchemaRepo(db)

// 修改后
schemaRepo := repository.NewSQLiteSchemaRepo(db)
```

### 步骤 2: 创建测试导出文件

创建 `go/services/config/sdk/export_test.go`:
```go
//go:build test

package sdk

// NewLRUCacheForTest 为测试提供 LRU 缓存实例
func NewLRUCacheForTest(capacity int, ttlSec ...int) *lruCache {
    return newLRUCache(capacity, ttlSec...)
}
```

创建 `go/services/i18n/sdk/export_test.go`:
```go
//go:build test

package sdk

// NewLRUCacheForTest 为测试提供 LRU 缓存实例
func NewLRUCacheForTest(capacity int) *lruCache {
    return newLRUCache(capacity)
}
```

### 步骤 3: 修改 E2E 测试中的 Cache 调用

```go
// e2e_config_i18n_test.go
// 修改前
lruCache := cache.NewLRUCache(100)
i18nLRUCache := i18ncache.NewLRUCache(100)

// 修改后
lruCache := configsdk.NewLRUCacheForTest(100)
i18nLRUCache := i18nsdk.NewLRUCacheForTest(100)
```

### 步骤 4: 验证编译

```bash
cd go/gateway/proto-gateway && go test ./internal/server/...
```

---

## 7. 长期建议

### 7.1 API 稳定性

- E2E 测试应使用公开稳定的 API
- 内部实现重构时，应同步更新 E2E 测试
- 考虑将 E2E 测试纳入 CI 阻断项

### 7.2 测试分层

| 测试类型 | 职责 | 使用的 API |
|---------|------|-----------|
| 单元测试 | 验证单个函数/类 | 包内私有 + 公开 API |
| 集成测试 | 验证模块协作 | 公开 API |
| E2E 测试 | 验证完整链路 | 公开 API + 测试导出函数 |

### 7.3 文档同步

- PRD/ADR 中应记录公开 API 契约
- 重构内部实现时，检查 E2E 测试兼容性

---

## 8. 相关文件

| 文件 | 说明 |
|------|------|
| `go/gateway/proto-gateway/internal/server/e2e_config_i18n_test.go` | 需要修复的 E2E 测试 |
| `go/services/config/repository/schema_repo.go` | SchemaRepo 实际实现 |
| `go/services/config/sdk/cache_lru.go` | Config LRU Cache 实现 |
| `go/services/i18n/sdk/cache_lru.go` | I18n LRU Cache 实现 |
| `go/services/config/service/interface.go` | AppConfigService 构造函数 |

---

*报告生成时间: 2026-05-23*
*分析人: Trae*
*严重程度: P2 - 影响 E2E 测试，不阻断主流程*
