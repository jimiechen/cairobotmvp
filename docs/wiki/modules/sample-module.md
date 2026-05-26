# 模块接入规范（Sample Module）

> **本文档是后续所有业务模块的统一接入模板。**
>
> **任何业务模块（OpenAPI、设备网关、用户中台、AI 服务、TenantServer 等）只需照抄 Hello / Health 骨架，再填业务逻辑即可。**

---

## 📋 新模块接入 Checklist（10 项）

| # | 检查项 | 通过判据 |
|---|--------|----------|
| 1 | 协议号已在 [协议编号注册表](../../api/协议编号注册表.md) 登记 | grep 命中 |
| 2 | proto 文件放在 `proto/<场景>/<name>.proto` | 文件存在 |
| 3 | 模块目录 `go/modules/<name>/` 含 5 个标准文件 | ls 通过 |
| 4 | service.go 使用 `common-lib/module.Deps` 装配 | grep 通过 |
| 5 | usecase.go 仅通过 configsdk / i18nsdk 取数 | grep 自检 |
| 6 | 0 直接 sql.Open / redis.NewClient | grep 为空 |
| 7 | Schema / i18n key 已通过 seed 脚本注入 | SQL 验证 |
| 8 | 实现 Checker 并注册到 modules/health | 单测通过 |
| 9 | usecase_test.go 覆盖率 ≥80% | go cover |
| 10 | README.md 含统一 6 节内容 | 人工 review |

**任何业务模块只要能完整打钩这 10 项，就算合规。**

---

## 🏗️ 一、模块标准目录结构

```
go/modules/<name>/                ← 业务模块（被 Gateway 路由调用）
├── service.go                    依赖装配，输出 *Service 单例
├── handler.go                    bytes ↔ proto 转换 + 路由分发
├── usecase.go                    业务逻辑（多个，按用例命名）
├── checker.go                    可选：health 检查器
├── *_test.go                     单元测试，覆盖率 ≥80%
└── README.md                     模块文档（含 schema/i18n key 清单）
```

### 文件规模约束

| 文件 | 推荐上限 | 绝对上限 |
|------|----------|----------|
| service.go | 80 行 | 200 行 |
| handler.go | 120 行 | 200 行 |
| usecase.go | 150 行 | 200 行 |
| checker.go | 150 行 | 200 行 |
| 测试文件 | 300 行 | 500 行 |

---

## 🔌 二、统一依赖装配（Deps 结构）

所有业务模块必须通过 `module.Deps` 接收依赖，**禁止**：
- ❌ 直接 import services/config 或 services/i18n 内部包
- ❌ 直接 sql.Open / redis.NewClient

### Deps 结构定义

[deps.go](../../../go/common-lib/module/deps.go)

```go
type Deps struct {
    Config configsdk.Client     // 必选：配置 SDK
    I18n   i18nsdk.Client       // 可选：国际化 SDK（无文案场景为 nil）
    DB     interface{}           // 可选：数据库连接（预留 mysqlx.DB 类型）
    Cache  interface{}           // 可选：缓存客户端（预留 redisx.Client 类型）
    Logger *log.Logger           // 必选：日志组件
}
```

### main.go 中统一装配示例

```go
deps := module.Deps{
    Config: configsdk.Default(),
    I18n:   i18nsdk.Default(),
    DB:     mysqlx.Default(),
    Cache:  redisx.Default(),
    Logger: log.Default(),
}

helloSvc  := hello.New(deps)
healthSvc := health.New(deps, checkers)
```

### Service 标准签名

```go
func New(deps module.Deps) *Service {
    // 1. 创建 Usecase（注入 SDK 客户端）
    // 2. 创建 Handler（注入 Usecase + Logger）
    // 3. 返回 Service 实例
}
```

---

## ⚙️ 三、配置接入约束

### 3.1 基本规则

1. **任何配置读取必须通过 configsdk**
2. **module_key 命名**：`<模块名>_cfg`（如 hello_cfg、device_cfg、openapi_cfg）
3. **配置字段必须先在 sys_config_schema 注册后才能读取**
4. **启动时 SDK 调用 GetInt/GetString 不应阻塞主流程**
   - 读不到时降级到 default（schema 提供）
   - 仅记录 warn 日志，不抛异常
5. **热更场景使用 Watch 订阅**，不要在每次请求里轮询

### 3.2 configsdk 接口定义

[configsdk/client.go](../../../go/common-lib/sdk/configsdk/client.go)

```go
type Client interface {
    GetString(ctx context.Context, moduleKey string, fieldKey string) (string, error)
    GetInt(ctx context.Context, moduleKey string, fieldKey string) (int64, error)
    GetBool(ctx context.Context, moduleKey string, fieldKey string) (bool, error)
    Watch(ctx context.Context, moduleKey string, callback func(fieldKey string, oldValue, newValue interface{})) error
    Ping(ctx context.Context) error
}
```

### 3.3 配置读取最佳实践

✅ **推荐写法**：

```go
serverName, err := u.cfg.GetString(ctx, "hello_cfg", "server_name")
if err != nil {
    return nil, fmt.Errorf("读取配置失败: %w", err)
}
if serverName == "" {
    serverName = "CaiRobot" // 降级默认值
}
```

❌ **禁止写法**：

```go
// 禁止：直接访问配置存储
cfg, _ := config.Load("/etc/cairobot/config.yaml")
serverName cfg.MySQL.Host
```

---

## 🌍 四、多语言接入约束

### 4.1 基本规则

1. **任何用户可见文案必须通过 i18nsdk.T() 渲染**
2. **key 命名**：`svc_<模块名>_<场景>`（如 svc_hello_greeting、svc_health_status_summary）
3. **服务端 T() 仅用于**：推送、短信、邮件、Webhook、错误响应文案
4. **客户端可渲染的场景使用 Raw()**，由客户端按 template_type 渲染
5. **任何 T() 调用必须有降级文案（fallback string）**，不允许直接抛错
6. **必须在 sys_lang_string 中注册中英文两条记录**，否则 CI 拒绝合入

### 4.2 i18nsdk 接口定义

[i18nsdk/client.go](../../../go/common-lib/sdk/i18nsdk/client.go)

```go
type Client interface {
    T(ctx context.Context, lang string, key string, params map[string]any) (string, error)
    Raw(ctx context.Context, lang string, key string) (string, string, error)
    Ping(ctx context.Context) error
}
```

### 4.3 模板类型

| Type | 说明 | 适用场景 |
|------|------|----------|
| TemplateTypeNamed | 命名参数模板 `{name}` | 大多数场景 |
| TemplateTypeICU | ICU MessageFormat | 复数、性别、选择等复杂场景 |

### 4.4 多语言渲染最佳实践

✅ **推荐写法**：

```go
greeting, err := u.i18n.T(ctx, lang, "svc_hello_greeting", map[string]any{
    "name":        req.Name,
    "server_name": serverName,
})
if err != nil {
    greeting = fmt.Sprintf("Hello, %s! Welcome to %s.", name, serverName) // 降级 fallback
}
```

❌ **禁止写法**：

```go
// 禁止：硬编码用户可见文案
message := "你好，" + name + "！欢迎使用 CaiRobot。"
```

---

## ❤️ 五、Health 接入约束

### 5.1 基本规则

1. **每个模块若涉及外部依赖，必须实现 Checker 并注册到 health 模块**
2. **checker 的 Check() 必须有超时（默认 1s）**，不能拖垮 health 协议
3. **失败状态写入 ComponentStatus.Error 字段**，便于运维定位

### 5.2 Checker 接口定义

[health/checker.go](../../../go/common-lib/health/checker.go)

```go
type Checker interface {
    Name() string
    Check(ctx context.Context) ComponentStatus
}

type ComponentStatus struct {
    Name      string
    Healthy   bool
    LatencyMs int64
    Error     string
}
```

### 5.3 内置 Checker 实现

| Checker | 名称 | 用途 |
|---------|------|------|
| ConfigChecker | config | configsdk.Ping |
| I18nChecker | i18n | i18nsdk.Ping |
| MySQLChecker | mysql | mysqlx.DB Ping（预留） |
| RedisChecker | redis | redisx.Client Ping（预留） |

### 5.4 使用示例

```go
checkers := []health.Checker{
    NewConfigChecker(deps.Config),
    NewI18nChecker(deps.I18n),
    NewMySQLChecker(),
    NewRedisChecker(),
}
healthSvc := health.New(deps, checkers)
```

---

## 🧪 六、测试接入约束

### 6.1 基本规则

1. **usecase_test.go 必须用 SDK Fake**，不依赖真实 Redis / MySQL
2. **集成测试用 build tag = integration**
3. **模块覆盖率必须 ≥80%**
4. **每个模块必须有一个端到端用例**（gateway → handler → usecase → SDK → 返回）

### 6.2 Fake SDK 使用示例

[configsdk/fake.go](../../../go/common-lib/sdk/configsdk/fake.go)
[i18nsdk/fake.go](../../../go/common-lib/sdk/i18nsdk/fake.go)

```go
func TestUsecase_Greet_NormalCase(t *testing.T) {
    cfg := configsdk.NewFakeClient()
    cfg.Set("hello_cfg", "server_name", "CaiRobot")
    cfg.Set("hello_cfg", "max_name_length", int64(32))

    i18n := i18nsdk.NewFakeClient()
    i18n.SetTranslation("zh-CN", "svc_hello_greeting", "你好，{name}！欢迎使用 {server_name}。")

    usecase := NewUsecase(cfg, i18n)
    // ... 测试逻辑
}
```

---

## 📝 七、文档接入约束

每个模块 README.md 必须包含 **6 节**：

1. **模块职责**（一句话）
2. **协议清单**（maxType/minType + req/rsp message）
3. **配置 Schema 清单**（module_key + 字段表）
4. **多语言 Key 清单**（key + 各语言模板）
5. **依赖关系**（依赖哪些 SDK / 基础设施）
6. **健康检查**（Checker 实现说明）

参考范例：
- [Hello 模块 README](../hello/README.md)
- [Health 模块 README](../health/README.md)

---

## 🎯 八、命名规范

### 8.1 module_key 命名

格式：`<模块名>_cfg`

| 模块 | module_key | 示例字段 |
|------|------------|----------|
| Hello | hello_cfg | server_name, max_name_length |
| Health | system_cfg, health_cfg | build_version, max_depth |
| Device | device_cfg | timeout_ms, retry_count |
| OpenAPI | openapi_cfg | rate_limit, api_key_ttl |

### 8.2 i18n key 命名

格式：`svc_<模块名>_<场景>`

| 场景 | key 示例 |
|------|----------|
| Hello 问候 | svc_hello_greeting |
| Health 状态摘要 | svc_health_status_summary |
| 错误提示 | svc_device_offline_warning |
| 成功通知 | svc_user_welcome_message |

---

## 📦 九、Seed 数据规范

每个模块必须提供 Seed 脚本，位于 `migrations/seed/<name>_seed.sql`。

### Seed 脚本内容

1. **sys_config_schema 注入**：模块所有配置字段的 Schema 定义
2. **sys_lang_string 注入**：模块所有 i18n key 的中英文模板
3. **使用 ON DUPLICATE KEY UPDATE**：支持重复执行

### Seed 脚本示例

参考：
- [hello_seed.sql](../../../migrations/seed/hello_seed.sql)
- [health_seed.sql](../../../migrations/seed/health_seed.sql)

---

## 🔗 十、参考实现

### 10.1 Hello 模块（configsdk 接入范例）

**演示能力**：
- ✅ 强类型配置读取（GetString / GetInt）
- ✅ 配置驱动校验（max_name_length）
- ✅ 服务端 i18n 渲染（named 模板）
- ✅ 失败降级机制

**代码引用**：
- [service.go (33行)](../../../go/modules/hello/service.go)
- [handler.go (51行)](../../../go/modules/hello/handler.go)
- [usecase.go (97行)](../../../go/modules/hello/usecase.go)
- [usecase_test.go (247行)](../../../go/modules/hello/usecase_test.go)
- [README.md](../hello/README.md)

### 10.2 Health 模块（i18nsdk ICU + Checker 范例）

**演示能力**：
- ✅ ICU plural 模板渲染
- ✅ Checker 抽象与复用
- ✅ 并发健康检查（超时控制）
- ✅ Depth 分层检查

**代码引用**：
- [service.go (34行)](../../../go/modules/health/service.go)
- [handler.go (51行)](../../../go/modules/health/handler.go)
- [usecase.go (144行)](../../../go/modules/health/usecase.go)
- [checker.go (96行)](../../../go/modules/health/checker.go)
- [usecase_test.go (280行)](../../../go/modules/health/usecase_test.go)
- [README.md](../health/README.md)

---

## ⚠️ 十一、全程禁止事项

1. ❌ **不允许 git push**（本次任务范围内）
2. ❌ **不允许在 modules/* 中 import services/* 的内部包**，必须只用 SDK
3. ❌ **不允许直接 sql.Open / redis.NewClient**
4. ❌ **不允许硬编码任何用户可见中英文文案**（错误码字符串可豁免）
5. ❌ **不允许跳过 H3**，沉淀文档与代码升级同等重要
6. ❌ **不允许把"统一 Deps"放在 modules/* 内部**，必须放 common-lib/module/

---

## 📊 十二、终验收标准

完成新模块开发后，请执行以下自检：

```bash
# 1. 文件行数检查
find go/modules/<name> -name "*.go" -not -name "*_test.go" -exec wc -l {} \;

# 2. 禁止 import 检查
grep -rn "services/config\|services/i18n" go/modules/<name>/;

# 3. 禁止直接连接检查
grep -rn "sql\.Open\|redis\.NewClient" go/modules/<name>/;

# 4. 硬编码文案检查
grep -rn "[\u4e00-\u9fa5]" go/modules/<name>/usecase.go | grep -v "//";

# 5. 测试覆盖率
go test ./modules/<name>/... -cover;

# 6. Gateway e2e（如有）
make gateway-e2e;
```

**所有检查项通过后，该模块才算合规。**

---

## 🎓 总结

**三句话收敛**：

1. **Hello 升级演示 configsdk**：强类型读取 + 服务端 i18n 渲染 + 配置驱动校验。
2. **Health 升级演示 i18nsdk**：ICU plural 模板 + Checker 抽象 + 多依赖健康汇总。
3. **sample-module.md 把这两次升级沉淀为模板**，后续所有模块按 10 项 Checklist 接入，研发零思考、Trae 零跑偏。

**这一步做完后，CaiRobot MVP 就具备了"加一个新业务模块只是抄 Hello / Health 一遍"的工程效率。从 S1 进入 S2（业务模块批量接入）的最大障碍——接入姿势不统一——也就被消除了。**

---

*本文档基于 H1/H2 升级实践沉淀，最后更新于 2026-05-26*
