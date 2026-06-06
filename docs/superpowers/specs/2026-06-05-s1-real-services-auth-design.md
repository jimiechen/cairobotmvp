# S1 阶段设计规范: noop Stub → 真实服务 + Auth 中间件

> 版本: v1.0
> 日期: 2026-06-05
> 状态: 待评审
> 方案: A — 单体内嵌（Gateway 直接集成 MySQL 服务 + JWT Auth）
> 前置: BUG-E2E-001 (已修复) + BUG-E2E-002 HealthCheck (已修复)
> 关联: IMP-CONFIG-I18N-001 (Phase D), ADR-009, ADR-012

## 1. 设计目标

将 Gateway local 模式下的 noop stub 替换为真实 MySQL 后端服务，并实现 JWT Auth 中间件，使 6000 段 Config/I18n 接口具备完整的数据读写和鉴权能力。

### 1.1 成功标准

| 标准 | 验证方式 |
|------|----------|
| GetAppConfigs (6001) 返回 MySQL 中的真实配置数据 | E2E 测试 |
| GetAppLanguage (6003) 返回 MySQL 中的语言列表 | E2E 测试 |
| GetLangPack (6005) 返回 MySQL 中的语言包 | E2E 测试 |
| auth_required=true 路由缺失 Token 时返回鉴权错误（非 10404） | E2E 负向测试 |
| auth_required=true 路由错误 Token 时返回鉴权失败 | E2E 负向测试 |
| auth_required=false 路由无 Token 正常访问 | E2E 正向测试 |
| HealthCheck 在 MySQL 连接正常时返回 OK | 单元测试 |

### 1.2 非目标

- 不做微服务拆分（S2 阶段）
- 不实现 Admin 后台的 Config/I18n 管理界面
- 不实现 Redis 缓存层（S1 用 MockCache，后续按需加 Redis）
- 不修改 proto 文件或协议编号

---

## 2. 架构设计

### 2.1 整体架构

```
┌──────────────────────────────────────────────────────┐
│                   Gateway Binary                     │
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐   │
│  │ ConfigSvc  │  │ I18nSvc    │  │ AuthService  │   │
│  │(MySQL Repo)│  │(MySQL Repo)│  │(JWT 签发/校验)│   │
│  └─────┬──────┘  └─────┬──────┘  └──────┬───────┘   │
│        │               │                │            │
│  ┌─────▼───────────────▼────────────────▼───────┐    │
│  │          RegisterAllLocalHandlers()           │    │
│  │   RegisterSystemHandlers(deps)               │    │
│  │   RegisterConfigI18nHandlers(cfg, i18n)       │    │
│  └─────────────────────┬─────────────────────────┘    │
│                        │                              │
│  ┌─────────────────────▼─────────────────────────┐    │
│  │         Auth Middleware (ServeHTTP 层)         │    │
│  │  1. 路由匹配 → route.AuthRequired?            │    │
│  │  2. 提取 extend["token"]                      │    │
│  │  3. JWT Parse + Validate                      │    │
│  │  4. 通过 → Invoke / 拒绝 → 40101              │    │
│  └─────────────────────┬─────────────────────────┘    │
│                        │                              │
│  ┌─────────────────────▼─────────────────────────┐    │
│  │        HTTP Server :8080 /api/hello           │    │
│  └───────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────┘
                           │
                    MySQL :3306 (cairobot_db)
                    ┌──────────────────┐
                    │ sys_config_schema │
                    │ sys_config_version│
                    │ sys_lang_pack     │
                    │ sys_lang_string   │
                    └──────────────────┘
```

### 2.2 请求处理流程（auth_required=true 时）

```
POST /api/hello (MessagePacket protobuf binary)
  │
  ├─ [Auth MW] Content-Type 检查
  ├─ [Auth MW] DeserializeMessagePacket → packet
  ├─ [Auth MW] 路由匹配 → Target{route}
  │     └─ route.AuthRequired == true?
  │           ├─ Yes → 提取 packet.Extend["token"]
  │           │      └─ JWT Validate
  │           │            ├─ Valid → 放行 (注入 userId 到 extend)
  │           │            └─ Invalid → 返回 40101 (Unauthorized)
  │           └─ No → 放行
  │
  ├─ [Server] invoker.Invoke(target, packet) → handler
  │     └─ handler(reqData) → json.Unmarshal → 业务逻辑 → resp
  │
  └─ [Server] SerializeMessagePacket(resp) → HTTP 200
```

### 2.3 请求处理流程（auth_required=false 时）

与当前流程一致，无变化。HealthCheck、GetAppLanguage 保持免鉴权。

---

## 3. DDL 设计

### 3.1 数据库: `cairobot_db`（使用已有 MySQL 实例）

```sql
-- 1. 配置 Schema 元数据表
CREATE TABLE IF NOT EXISTS sys_config_schema (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    module_key  VARCHAR(64) NOT NULL UNIQUE COMMENT '模块标识',
    name        VARCHAR(128) NOT NULL COMMENT '显示名称',
    description TEXT DEFAULT NULL,
    env         VARCHAR(16) NOT NULL DEFAULT 'dev' COMMENT '环境',
    status      TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_env_status (env, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. 配置版本快照表
CREATE TABLE IF NOT EXISTS sys_config_version (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    module_key    VARCHAR(64) NOT NULL,
    env           VARCHAR(16) NOT NULL DEFAULT 'dev',
    version       BIGINT NOT NULL DEFAULT 1,
    config_data   JSON NOT NULL COMMENT '配置内容 JSON',
    publisher     VARCHAR(64) DEFAULT NULL,
    published_at  DATETIME DEFAULT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_module_env_version (module_key, env, version),
    INDEX idx_module_env (module_key, env)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 3. 语言包表
CREATE TABLE IF NOT EXISTS sys_lang_pack (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    lang_code     VARCHAR(10) NOT NULL COMMENT '语言代码',
    client_min    VARCHAR(16) NOT NULL DEFAULT '1.0.0',
    pack_version  BIGINT NOT NULL DEFAULT 1,
    env           VARCHAR(16) NOT NULL DEFAULT 'dev',
    status        TINYINT NOT NULL DEFAULT 1,
    published_at  DATETIME DEFAULT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_lang_pack (lang_code, pack_version, env),
    INDEX idx_lang_code (lang_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 4. 语言字符串表
CREATE TABLE IF NOT EXISTS sys_lang_string (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    pack_id       BIGINT NOT NULL,
    string_key    VARCHAR(256) NOT NULL COMMENT '翻译键',
    value         TEXT NOT NULL COMMENT '翻译值',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pack_key (pack_id, string_key),
    INDEX idx_pack_id (pack_id),
    CONSTRAINT fk_lang_string_pack FOREIGN KEY (pack_id) REFERENCES sys_lang_pack(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3.2 种子数据

```sql
-- 初始化一个示例配置模块
INSERT INTO sys_config_schema (module_key, name, description, env) VALUES
('base_config', '基础配置', '应用全局基础配置', 'dev');

-- 初始化版本 V1
INSERT INTO sys_config_version (module_key, env, version, config_data, publisher, published_at) VALUES
('base_config', 'dev', 1, '{"app_name":"CaiRobot","debug":true,"max_connections":100}', 'system', NOW());

-- 初始化语言包
INSERT INTO sys_lang_pack (lang_code, client_min, pack_version, env, status, published_at) VALUES
('zh-CN', '1.0.0', 1, 'dev', 1, NOW()),
('en-US', '1.0.0', 1, 'dev', 1, NOW());

-- 初始化中文语言字符串
SET @zh_pack = (SELECT id FROM sys_lang_pack WHERE lang_code='zh-CN' AND pack_version=1);
INSERT INTO sys_lang_string (pack_id, string_key, value) VALUES
(@zh_pack, 'common.ok', '确定'),
(@zh_pack, 'common.cancel', '取消'),
(@zh_pack, 'common.loading', '加载中...'),
(@zh_pack, 'error.network', '网络连接失败'),
(@zh_pack, 'error.unauthorized', '未授权访问');

-- 初始化英文语言字符串
SET @en_pack = (SELECT id FROM sys_lang_pack WHERE lang_code='en-US' AND pack_version=1);
INSERT INTO sys_lang_string (pack_id, string_key, value) VALUES
(@en_pack, 'common.ok', 'OK'),
(@en_pack, 'common.cancel', 'Cancel'),
(@en_pack, 'common.loading', 'Loading...'),
(@en_pack, 'error.network', 'Network error'),
(@en_pack, 'error.unauthorized', 'Unauthorized');
```

---

## 4. 模块设计

### 4.1 MySQL Repository 实现

#### 4.1.1 Config MySQL Repository

**新文件**: `go/services/config/repository/mysql_repo.go`

基于已有的 `repository.ConfigRepository` 接口，实现 MySQL 版本：

```go
// MySQLConfigRepo 基于 mysqlx.DB 的 Config Repository 实现
type MySQLConfigRepo struct {
    db     mysqlx.DB
    logger Logger
}

func NewMySQLConfigRepo(db mysqlx.DB, logger Logger) *MySQLConfigRepo
// 实现: GetLatestVersion, GetByModuleAndVersion, ListPublishedVersions, Save
```

关键方法：
- `ListPublishedVersions(moduleKey, env)` → 查询 `sys_config_version` 表
- `GetByModuleAndVersion(moduleKey, env, version)` → 读取 `config_data` JSON 字段
- `GetLatestVersion(moduleKey, env)` → MAX(version)

#### 4.1.2 Schema MySQL Repository

**新文件**: `go/services/config/repository/mysql_schema_repo.go`

基于已有的 `repository.SchemaRepository` 接口：

```go
// MySQLSchemaRepo 基于 mysqlx.DB 的 Schema 元数据 Repository
type MySQLSchemaRepo struct {
    db     mysqlx.DB
    logger Logger
}

func NewMySQLSchemaRepo(db mysqlx.DB, logger Logger) *MySQLSchemaRepo
// 实现: ListActiveSchemas, GetByKey, Save
```

#### 4.1.3 I18n MySQL Repository

**新文件**: `go/services/i18n/repository/mysql_repo.go`（或扩展现有文件）

基于已有的 `repository.I18nRepository` 接口：

```go
// MySQLI18nRepo 基于 mysqlx.DB 的 I18n Repository 实现
type MySQLI18nRepo struct {
    db     mysqlx.DB
    logger Logger
}

func NewMySQLI18nRepo(db mysqlx.DB, logger Logger) *MySQLI18nRepo
// 实现: GetPackByLangCode, GetStringsByPackID, GetDiffSince, ListPacks, FindStringByKey
```

### 4.2 Cache 层

S1 使用 **MockCache**（已有实现），不引入 Redis 依赖。后续可替换为 RedisCache。

```go
// 使用已有的 MockCache（内存缓存，开发测试够用）
cache := config.NewMockCache()
i18nCache := i18n.NewMockCache()
```

### 4.3 服务组装函数

**修改文件**: `go/gateway/proto-gateway/tarsclient/invoker.go`

新增函数，构建真实的 Config 和 I18n 服务实例：

```go
// buildRealServices 构建 MySQL 后端的真实 Config 和 I18n 服务
// 依赖: Deps.DB (mysqlx.DB), Deps.Logger (Logger)
func buildRealServices(deps module.Deps) (configservice.ConfigService, i18nservice.I18nService) {
    // Config Service
    configRepo := configrepo.NewMySQLConfigRepo(deps.DB.(mysqlx.DB), deps.Logger)
    schemaRepo := configrepo.NewMySQLSchemaRepo(deps.DB.(mysqlx.DB), deps.Logger)
    cfgCache := configcache.NewMockCache()
    configSvc := configservice.NewAppConfigService(configRepo, schemaRepo, cfgCache)

    // I18n Service
    i18nRepo := i18nrepo.NewMySQLI18nRepo(deps.DB.(mysqlx.DB), deps.Logger)
    i18nCache := i18ncache.NewMockCache()
    i18nSvc := i18nservice.NewI18nService(i18nRepo, i18nCache, "dev")

    return configSvc, i18nSvc
}
```

### 4.4 启动模式切换

**修改文件**: `go/gateway/proto-gateway/cmd/server/main.go`

```go
if mode == "local" {
    invoker = tarsclient.NewLocalInvoker()

    if useRealServices {  // 新增：通过环境变量控制
        deps := buildDepsWithDB()  // 包含 MySQL 连接
        tarsclient.RegisterSystemHandlersWithDeps(invoker.(*tarsclient.LocalInvoker), deps)
        cfg, i18n := tarsclient.BuildRealServices(deps)
        tarsclient.RegisterConfigI18nHandlers(invoker.(*tarsclient.LocalInvoker), cfg, i18n)
    } else {
        // 当前行为：noop stub（保持向后兼容）
        tarsclient.RegisterAllLocalHandlers(invoker.(*tarsclient.LocalInvoker))
    }
}
```

环境变量: `GATEWAY_USE_REAL_SERVICES=true` 或 `GATEWAY_INVOKER_MODE=mysql-local`

---

## 5. Auth 中间件设计

### 5.1 架构位置

在 `http_server.go` 的 `ServeHTTP` 方法中，路由匹配之后、`invoker.Invoke` 之前插入 Auth 校验。

### 5.2 JWT Auth Service

**新文件**: `go/tars/auth/service.go` + `go/tars/auth/jwt.go`

```go
// AuthService JWT 认证服务
type AuthService struct {
    secret     []byte           // JWT 签名密钥（从配置读取）
    issuer     string           // 签发者标识 "cairobot"
    expiration time.Duration    // Token 有效期（默认 24h）
}

// TokenClaims JWT 载荷
type TokenClaims struct {
    UserID   string `json:"user_id"`
    Role     string `json:"role"`       // parent / child / admin
    DeviceID string `json:"device_id"`  // 可选
    jwt.RegisteredClaims
}

// 核心方法
func (s *AuthService) GenerateToken(userID, role string) (string, error)   // 签发
func (s *AuthService) ValidateToken(tokenStr string) (*TokenClaims, error) // 校验
```

### 5.3 Gateway Auth 中间件

**新文件**: `go/gateway/proto-gateway/internal/middleware/auth.go`

```go
// AuthMiddleware Gateway 层 Auth 中间件
type AuthMiddleware struct {
    authService *auth.Service
    logger      Logger
}

// Intercept 在 Invoker.Invoke 前执行 Token 校验
func (m *AuthMiddleware) Intercept(ctx context.Context, target Target, packet *pb.MessagePacket) (*pb.MessagePacket, error) {
    route := GetRoute(target)  // 从 RouteTable 查找

    if !route.AuthRequired {
        return nil, nil  // 免鉴权，放行
    }

    tokenStr := packet.Extend["token"]
    if tokenStr == "" {
        return BuildErrorResponse(40101, "missing token", packet.Extend["traceId"]), nil
    }

    claims, err := m.authService.ValidateToken(tokenStr)
    if err != nil {
        return BuildErrorResponse(40102, "invalid token: "+err.Error(), packet.Extend["traceId"]), nil
    }

    // 将 userId 注入 extend，供下游 handler 使用
    packet.Extend["user_id"] = claims.UserID
    packet.Extend["user_role"] = claims.Role

    return nil, nil  // 校验通过，放行
}
```

### 5.4 错误码定义

| 错误码 | 含义 | HTTP Status |
|--------|------|-------------|
| 40101 | 缺少 Token | 200 (业务码) |
| 40102 | Token 无效/过期 | 200 (业务码) |
| 40103 | 权限不足 | 200 (业务码) |

> 注意：Gateway 协议层面始终返回 HTTP 200 + MessagePacket，业务码在 extend["code"] 中携带。

### 5.5 Token 传递协议约定

- **Key 名称**: `"token"`
- **位置**: `MessagePacket.Extend` map
- **格式**: JWT Bearer Token（`eyJhbGci...`）
- **proto-tester 发送方式**: `--token "eyJhbGci..."` 参数写入 extend

---

## 6. 文件变更清单

### 6.1 新增文件

| 文件 | 说明 |
|------|------|
| `scripts/sql/s1_init_ddl.sql` | DDL + 种子数据 |
| `scripts/sql/s1_seed_data.sql` | 示例配置和语言数据 |
| `go/services/config/repository/mysql_repo.go` | Config MySQL Repository |
| `go/services/config/repository/mysql_schema_repo.go` | Schema MySQL Repository |
| `go/services/i18n/repository/mysql_repo.go` | I18n MySQL Repository |
| `go/tars/auth/service.go` | AuthService 核心 |
| `go/tars/auth/jwt.go` | JWT 工具（签发/校验） |
| `go/tars/auth/service_test.go` | AuthService 单元测试 |
| `go/gateway/proto-gateway/internal/middleware/auth.go` | Gateway Auth 中间件 |
| `go/gateway/proto-gateway/internal/middleware/auth_test.go` | Auth 中间件测试 |

### 6.2 修改文件

| 文件 | 变更 |
|------|------|
| `go/gateway/proto-gateway/tarsclient/invoker.go` | 新增 `BuildRealServices()` 函数 |
| `go/gateway/proto-gateway/cmd/server/main.go` | 增加 MySQL 连接初始化 + 服务组装分支 |
| `go/gateway/proto-gateway/internal/server/http_server.go` | ServeHTTP 中插入 AuthMiddleware.Intercept 调用 |
| `configs/gateway/gateway.local.conf` | 新增 MySQL 连接配置 + JWT secret 配置 |

### 6.3 不修改

- Proto 文件、协议编号
- routes.yaml（auth_required 已有字段）
- System handler 逻辑（HealthCheck/HelloWorld）
- proto-tester 编码逻辑

---

## 7. 测试策略

### 7.1 单元测试

| 测试文件 | 内容 |
|----------|------|
| `mysql_repo_test.go` | Config/I18n Repository CRUD（需要 testcontainers 或 mock DB） |
| `jwt_test.go` | JWT 签发/校验/过期/篡改 |
| `auth_middleware_test.go` | 免鉴权放行 / 缺失 Token / 无效 Token / 有效 Token |
| `invoker_test.go` 扩展 | TestBuildRealServices 验证 MySQL 服务组装正确 |

### 7.2 集成测试

| 测试文件 | 内容 |
|----------|------|
| `e2e_mysql_services_test.go` | MySQL 直连下 6001/6003/6005/6007/6009 全链路 |
| `e2e_auth_test.go` | Auth 中间件正负向（MVP-005/006 场景恢复为真实鉴权验证） |

### 7.3 E2E 复测

修复后重新执行 Phase 3 的全部 MVP 用例，预期从 PARTIAL_PASS 提升到 ALL_PASS：

| 用例 | 修复前 | S1 修复后期望 |
|------|--------|---------------|
| MVP-001 HealthCheck | PASS | PASS |
| MVP-002 HelloWorld | PASS | PASS |
| MVP-003 System traceId | PASS | PASS |
| MVP-004 Config 正向+Token | PASS (10200, 空) | **PASS (10200, 真实数据)** |
| MVP-005 Config 缺失 Token | PASS (10200, noop) | **40101 (鉴权拒绝)** |
| MVP-006 Config 错误 Token | PASS (10200, noop) | **40102 (鉴权拒绝)** |
| MVP-007 I18n 无 Token | PASS (10200, 固定数据) | **PASS (10200, MySQL 数据)** |
| MVP-008 I18n 有 Token | PASS (10200, 固定数据) | **PASS (10200, MySQL 数据)** |
| MVP-010 traceId 贯穿 | PASS | PASS |

---

## 8. 实施步骤（建议顺序）

```
Step 1: DDL + 种子数据脚本 → 执行到本地 MySQL
Step 2: MySQL Repository 实现 (config + schema + i18n)
Step 3: Repository 单元测试（testcontainers 或 SQLite 兼容模式）
Step 4: BuildRealServices 组装函数 + main.go 启动改造
Step 5: 替换 noop 为真实服务 → 运行 6001/6003/6005 E2E 验证数据返回
Step 6: AuthService (JWT) 实现 + 单元测试
Step 7: AuthMiddleware 实现 + http_server.go 集成
Step 8: Auth 负向 E2E (MVP-005/006) 验证鉴权拦截
Step 9: 全量 MVP 用例复测
Step 10: 报告输出 + commit + push
```

---

## 9. 风险与依赖

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| MySQL 连接串泄露到 Git | P0 | 通过 gateway.local.conf 加载，不入库；.env.example 提供模板 |
| JWT secret 硬编码 | P0 | 同上，运行时从配置文件读取 |
| DDL 与现有 Repository 接口不兼容 | P1 | 先读接口定义再实现，接口不变则无需改 |
| testcontainers 依赖增加 CI 复杂度 | P2 | S1 可先用 mock DB 或 SQLite 兼容模式跑单元测试 |
| Gateway 二进制体积增大 | P3 | 可接受，Go 静态编译本就如此 |

---

## 10. 关联文档

- [IMP-CONFIG-I18N-001](../../prd/global-config-i18n-implementation-plan.md) — 总体实施方案
- [ADR-009](../../adr/ADR-009-config-i18n-schema-template.md) — Schema 模板架构决策
- [ADR-012](../../adr/ADR-0012-polyglot-monorepo-directory-layout.md) — 目录布局规范
- [E2E-FIX-REPORT-6000-HANDLERS.md](../testing/e2e-acceptance/E2E-FIX-REPORT-6000-HANDLERS.md) — BUG-E2E-001 修复报告
