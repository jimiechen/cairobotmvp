# 真实 MySQL + Redis 底座测试方案（简化版）

> **版本**：v1.1（根据主控反馈修订）
> **日期**：2026-06-19
> **原则**：单元测试直连真实数据库，跳过建表阶段，无内存模式双切换

---

## 一、目标

将现有 GORM 单元测试从 **SQLite 内存库** 切换到 **真实 MySQL (go_biz)**，新增 Redis TokenStore 集成，验证登录闭环和端到端功能。

---

## 二、基础设施

| 组件 | 地址 | 用途 |
|------|------|------|
| MySQL | 192.168.1.6:3306 root/123456 go_biz | Social 全部业务数据 |
| Redis | 192.168.1.6:6379 无密码 DB=2 | Token 黑名单 |

> 建表由用户/运维负责，本方案跳过建表步骤。

---

## 三、实施步骤

### Step 1：修改 3 个 `repository_gorm_test.go` 直连 MySQL

**核心改动**：`setupTestDB(t)` 从 SQLite in-memory 改为真实 MySQL。

**修改文件清单**：

| 文件 | 当前实现 | 改为 |
|------|---------|------|
| `member/repository_gorm_test.go` | `gorm.Open(sqlite.Open(":memory:"))` | `gorm.Open(mysql.Open(dsn))` |
| `group/repository_gorm_test.go` | 同上 | 同上 |
| `topic/repository_gorm_test.go` | 同上 | 同上 |

**统一 `setupTestDB` 实现**：

```go
// setupTestDB 连接真实 MySQL 数据库
// 环境变量：MYSQL_HOST/MYSQL_PORT/MYSQL_USER/MYSQL_PASSWORD/MYSQL_DATABASE
func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5000ms",
        getEnv("MYSQL_HOST", "127.0.0.1"),
        getEnv("MYSQL_PASSWORD", ""),
        getEnv("MYSQL_HOST", "127.0.0.1"),
        getEnv("MYSQL_PORT", "3306"),
        getEnv("MYSQL_DATABASE", "go_biz"),
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent), // 测试时关闭 SQL 日志
    })
    require.NoError(t, err)

    sqlDB, _ := db.DB()
    require.NoError(t, sqlDB.PingContext(context.Background()))

    return db
}
```

**ID 冲突解决**：每个测试用唯一 ID 前缀（如 `gt_member_` / `gt_group_` / `gt_topic_`），测试结束后 DELETE 清理。

### Step 2：新增 RedisTokenStore + 测试

**新建文件**：`member/token_store_redis.go`

- 使用 `github.com/redis/go-redis/v9`
- `Blacklist`: SETEX 写入黑名单 key
- `IsBlacklisted`: EXISTS 检查
- key 格式：`social:token:blacklist:{token_hash}`

**新建测试**：`member/token_store_redis_test.go`

- 直连 192.168.1.6:6379 DB=2
- 测试 Blacklist → IsBlacklisted → TTL 过期自动清理
- 测试后 DEL 清理 key

### Step 3：RegisterSocialHandlers 改用 GORM+Redis

**修改文件**：`tarsclient/invoker.go` 的 `RegisterSocialHandlers`

```go
func RegisterSocialHandlers(invoker *LocalInvoker) {
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "cairobot-mvp-p0-dev-secret-key-32bytes-min!!"
    }

    // 直接使用真实数据库和 Redis（无 memory 切换）
    db := openMySQLFromEnv()   // 新增辅助函数
    rdb := openRedisFromEnv()   // 新增辅助函数

    memberRepo := member.NewGormRepository(db)
    groupRepo := group.NewGormRepository(db)
    topicRepo := topic.NewGormRepository(db)
    tokenStore := member.NewRedisTokenStore(rdb, "social:tl:")

    jwtMgr, _ := member.NewJWTManager(member.DefaultJWTConfig().SetSecretKey(jwtSecret))

    socialMod := socialmodule.NewModule(
        memberRepo, groupRepo, topicRepo,
        socialmodule.WithJWTManager(jwtMgr),
    )
    socialMod.MemberServant.InjectTokenStore(tokenStore)

    // ... handler 注册不变 ...
}
```

### Step 4：新增登录链路集成测试

**新建文件**：`member/svc_login_integration_test.go`

| # | 用例 | 验证点 |
|---|------|--------|
| 1 | Register → Login → GetUserInfo | JWT 签发/解析/用户信息正确 |
| 2 | Login → RefreshToken | refresh_token 兑换新 token |
| 3 | Login → Logout → 旧 token 被拒 | Redis 黑名单生效 |
| 4 | 错误密码登录 → 10401 | bcrypt 校验失败 |
| 5 | 重复注册 → 10612 | 用户名唯一性 |

直接使用 GormRepository + RedisTokenStore 创建 Servant，不走 Gateway 层。

### Step 5：E2E mysql 模式验证

**修改文件**：`social_e2e_smoke/main.go`

- 增加 `SOCIAL_REPO_MODE=mysql` 标记输出
- Suffix 时间戳保证数据唯一性（已有）
- 报告中标注使用真实 DB

---

## 四、运行命令

```bash
# 加载环境变量
export $(grep -v '^#' go/gateway/proto-gateway/configs/gateway/.env.local | xargs)

# 运行全部 Social 单元测试（直连 MySQL + Redis）
cd go/modules/social && go test ./... -count=1 -v -timeout 120s

# 仅运行登录链路测试
cd go/modules/social/member && go test ./... -count=1 -v -run "TestLogin" -timeout 60s

# 启动 Gateway + E2E
make gateway-restart
cd go/gateway/proto-gateway && go run cmd/social_e2e_smoke/main.go
```

---

## 五、文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `member/repository_gorm_test.go` | **修改** | SQLite → 真实 MySQL |
| `group/repository_gorm_test.go` | **修改** | SQLite → 真实 MySQL |
| `topic/repository_gorm_test.go` | **修改** | SQLite → 真实 MySQL |
| `member/token_store_redis.go` | **新建** | RedisTokenStore |
| `member/token_store_redis_test.go` | **新建** | Redis 集成测试 |
| `member/svc_login_integration_test.go` | **新建** | 登录闭环 5 条用例 |
| `tarsclient/invoker.go` | **修改** | RegisterSocialHandlers 使用 GORM+Redis |
| `social_e2e_smoke/main.go` | **修改** | E2E 增加 mysql 模式标记 |

---

## 六、验收标准

| # | 标准 | 要求 |
|---:|------|------|
| 1 | 单元测试全绿 | `go test ./go/modules/social/...` PASS，直连 MySQL |
| 2 | 登录闭环 | Register→Login→GetInfo→Logout→Refresh 全部 PASS |
| 3 | Token 黑名单 | Logout 后旧 token 在 Redis 中被拒绝 |
| 4 | E2E 通过 | 34 条用例 POS+NEG >= 30，EXIT_CODE=0 |
| 5 | 编译 | `go build ./...` 零错误 |
