# PRD-10: Admin 管理后台 MVP（app_config + sys_lang_string 最小闭环）

## 文档信息

| 项目 | 内容 |
|---|---|
| PRD 编号 | PRD-10 |
| 标题 | Admin 管理后台 MVP：配置管理 + 多语言管理 CRUD + 前端页面 + SDK 联动 |
| 状态 | 已评审通过（三轮评审 + 主控确认） |
| 相关 ADR | [ADR-010](../adr/ADR-010-admin-boundary-sdk.md)（Admin 边界与 SDK 引用规范） |
| 相关 Issue | （待创建） |
| 总工期 | 9~10 个工作日（M0' ~ M7'） |
| 评审历史 | 原方案 → 一审(5阻塞) → 二审(3偏差) → 三审(2行伪代码修正) → 放行 |

---

## 一、背景与目标

### 1.1 背景

CaiRobot MVP 阶段需要运营人员能够自主完成以下操作，不再依赖研发手动改库：

- 新增/修改/删除配置字段定义（Schema）
- 按环境切换配置值并发布（Value）
- 新增/修改/删除多语言文案（String）
- 发布语言包版本（Pack）
- CSV 批量导入导出多语言文案

当前仓库中存在两套占位实现需要清理：
- `go/tars/provider-admin/` — Gin 框架的后端 Handler（**本轮废弃**）
- `web/provider-admin/` — React + Vite 的前端页面（**本轮废弃**）

### 1.2 目标

以最小可用闭环为目标，在 go-admin 开源框架上构建完整的「后端 CRUD + 前端操作页 + SDK 联动」三件套。

### 1.3 非目标（本轮不做）

```text
✗ 多租户（tenant_id 字段保留但全部默认 default）
✗ ABAC 三维权限（先用 RBAC：admin / operator / viewer）
✗ 审核工单（先直接落库，敏感字段 S2 再加审核流）
✗ 联邦视图、跨租户漂移
✗ 数据源 Registry（仅连 ops_db + config_db + i18n_db 三个固定 DSN）
✗ 高级文件管理 / OSS 离线包
```

所有表结构 + API 签名预留 `tenant_id`、`reviewable` 字段，S2 接入时不破坏现有数据。

---

## 二、技术决策总表（D1~D7）

| 编号 | 决策内容 | 决策值 | 依据 |
|---|---|---|---|
| D1 | go-admin 版本 | **v2.2.0**（commit: `8f8a197d`，单体应用模板非库） | POC 验证后由主控签字确认 |
| D2 | 前端框架 | 官方 go-admin-ui **v2.0.9**（**Vue 2.6.11 + Element UI**，vue-cli-service） | 保持默认主题，不调品牌色 |
| D3 | 数据源 | 固定 3 个 DSN（ops_db / config_db / i18n_db），写配置文件 | 不做 sys_data_source 表 |
| D4 | 租户字段 | 所有表 `tenant_id VARCHAR(64) NOT NULL DEFAULT 'default'`，SQL 默认带过滤 | 顶部不放租户切换栏 |
| D5 | 写路径 | admin 控制器 → `services/{config,i18n}/admin` → MySQL → Redis | **禁止** admin 插件直写 sys_config_* / sys_lang_* 表 |
| D6 | 失效广播 | channel = `cairobot.config.invalidate` / `cairobot.i18n.invalidate`；payload = JSON 格式 InvalidateEvent | 向后兼容旧逗号分隔格式 |
| D7 | 测试标准 | 单元 SQLite ≥80%、集成 MySQL build tag = integration、端到端联动必跑 | 缺一不可 |

---

## 三、架构铁律

### 3.1 写入唯一合法路径

```
Admin HTTP 请求
    → go/admin/plugins/{config,i18n}_admin/apis/     ← 控制器层（参数校验 + DTO 转换）
    → services/{config,i18n}/admin/                   ← 写入层（校验 + 落库 + 审计 + 失效 + 广播）
        → 复用 services/{config,i18n}/service/ 校验   ← 校验层（不重写）
        → repository                                   ← 持久化层
        → redisx.Client.Invalidate                     ← 缓存失效
        → redisx.PubSubClient.Publish                  ← 变更广播
```

### 3.2 启动架构（go-admin 安装后）

```
┌──────────────────────────────────────────────────────┐
│                    开发者本地环境                       │
│                                                      │
│  ┌─────────────────┐    ┌─────────────────────────┐  │
│  │ 终端 A           │    │ 终端 B                   │  │
│  │                 │    │                          │  │
│  │ cd go/admin     │    │ cd typescript/admin-web  │  │
│  │ go run cmd/     │    │ pnpm dev                 │  │
│  │ admin/main.go   │    │                          │  │
│  │                 │    │ :3000 (Vite dev server)   │  │
│  │ :8000 (Gin) ───┼──→│ proxy → :8000             │  │
│  │ go-admin engine │    │                          │  │
│  └────────┬────────┘    └────────▲─────────────────┘  │
│           │                      │                    │
│           ▼                      │                    │
│  ┌─────────────────────────────────┴────────────────┐  │
│  │              go/admin/ (独立 Go module)            │  │
│  │                                                  │  │
│  │  go.mod ← github.com/go-admin-team/go-admin      │  │
│  │  ├── config/settings.yml                         │  │
│  │  │   ├── DSN (AES-256-CBC 加密)                   │  │
│  │  │   └── Redis 连接配置                            │  │
│  │  ├── cmd/admin/main.go                            │  │
│  │  │   └── engine.Default() + gin.Server{}          │  │
│  │  └── plugins/                                     │  │
│  │      ├── config_admin/  (我们写的)                │  │
│  │      └── i18n_admin/     (我们写的)                │  │
│  └──────────────┬───────────────────────────────────┘  │
│                 │                                       │
│         ┌───────┴───────┐                             │
│         │ Go workspace  │ ← go/work 注册 ./admin       │
│         └───────────────┘                             │
│                                                      │
│  依赖来源（M0'.0.1 安装）：                           │
│  ├─ go-admin 核心包     → Go module proxy 自动拉取     │
│  ├─ Gin / GORM / Casbin → 随 go-admin 引入（隔离在     │
│  │                        go/admin/go.mod 内）          │
│  └─ go-admin-ui 源码    → git clone + 整合到          │
│                          typescript/admin-web/src/    │
└──────────────────────────────────────────────────────┘

启动顺序：
  ① M0'.0.1: git clone tag → mkdir骨架 → go mod init → go mod tidy
  ② M0'.0.2: clone ui → pnpm install → 编辑 main.go → go run → 浏览器验证
  ③ M0'.0.3: 配置 DSN 加密 → 设置 ADMIN_DSN_KEY → 启动验证连接
```

### 3.3 禁止事项（全程 18 条）

```text
1.  不允许 git push
2.  不允许 admin 插件直接写 sys_config_* / sys_lang_* 表
3.  不允许业务服务调用 admin
4.  不允许 admin 暴露公网
5.  不允许跳过单测或集成测试
6.  不允许在 admin-web 复用业务 sys_lang_string 做自身 i18n
7.  不允许把多租户、ABAC、审核工单提到本轮做
8.  不允许擅自更换 go-admin 版本
9.  不允许在 admin 子包中复现校验逻辑（必须复用 service 层包级函数）
10. 不允许 pub/sub 仍发送逗号分隔格式（必须 JSON InvalidateEvent）
11. 不允许 provider-admin 删除前不创建归档分支
12. 不允许 go-admin 可行性报告跳过 POC 验证
13. 不允许 redisx.Invalidate 无单元测试
14. 不允许 SDK 消费端去掉降级分支（向后兼容保留至 S2）
15. 不允许在 redisx.Client 接口上扩展 Publish（保持职责单一）
16. 不允许 CSV 导入跳过 dry-run 直接进事务
17. 不允许跳过 i18n SDK 现状摘要直接改代码
18. 不允许以"与 config SDK 同样逻辑"作为 i18n SDK 实施依据
```

### 3.4 边界隔离自检（每批必须通过）

```bash
# admin 插件不得直写业务表
grep -rn "sys_config_schema\|sys_config_version\|sys_lang_pack\|sys_lang_string" \
  go/admin/plugins/*/  # 必须为空

# 业务服务不得引用 admin
grep -rn "services/config/admin\|services/i18n/admin" \
  go/services/openapi/ go/services/devicegw/ \
  go/services/usercenter/ go/services/aiservice/ go/modules/  # 必须为空

# admin 子包必须复用 service 层校验（不得自行实现字段级校验）
grep -E "field_type ==|switch.*field_type|validator JSON|default_value 解析" \
  go/services/config/admin/*.go go/services/i18n/admin/*.go  # 必须为空（排除 _test.go）

# admin 子包必须命中 service 层 API
grep "ValidateSchema\|ValidateValue\|ValidateLangString\|ValidateTemplate" \
  go/services/config/admin/*.go go/services/i18n/admin/*.go  # 必须命中
```

---

## 四、交付物清单

### 4.1 配置侧（app_config）

```text
✓ Schema 字段 CRUD（增/删/改/查、按 module_key 过滤、按 type 校验）
✓ Value 值 CRUD（按 env 切换、按 schema 动态渲染表单）
✓ 启用/禁用、发布版本、实时预览
✓ 写入即失效缓存 + Redis pub/sub 广播（JSON payload）
✓ admin-web 两个页面：Schema 列表+表单、Value 列表+表单
```

### 4.2 i18n 侧（sys_lang_string）

```text
✓ String CRUD（key + 多语言 value + template_type + params_schema）
✓ 多语言并排编辑（中英对照）
✓ 占位符与 params_schema 一致性强校验
✓ preview_sample 实时预览
✓ CSV 简易导入导出（dry-run + 两阶段事务）
✓ Pack 版本号自增 + Redis pub/sub 广播（JSON payload）
✓ admin-web 两个页面：String 列表+并排编辑、Pack 简易管理
```

### 4.3 通用

```text
✓ go-admin 自带 RBAC（admin / operator / viewer 三角色）
✓ 操作日志（go-admin sys_oper_log）
✓ 端到端联动测试（admin 改完 → SDK 100ms 内感知）
```

---

## 五、目录与文件骨架

```text
go/admin/
├── core/                       go-admin 原版
├── plugins/
│   ├── config_admin/
│   │   ├── apis/
│   │   │   ├── schema.go              ≤180 行
│   │   │   └── value.go               ≤180 行
│   │   ├── service/
│   │   │   ├── schema_service.go      仅 DTO 转换 + 调 services 层
│   │   │   └── value_service.go
│   │   ├── models/
│   │   │   ├── schema_dto.go
│   │   │   └── value_dto.go
│   │   ├── router/router.go
│   │   └── plugin.go
│   └── i18n_admin/
│       ├── apis/
│       │   ├── string.go
│       │   ├── pack.go
│       │   └── import_export.go
│       ├── service/
│       ├── models/
│       ├── router/
│       └── plugin.go
├── cmd/admin/main.go
└── config/settings.yml

go/services/config/admin/              ← 真正的写入层
├── service.go                          AdminService 接口
├── schema_service.go                  DTO→domain + 调校验 + 落库 + 审计 + 失效 + 广播
├── value_service.go                   同上
├── audit.go                           审计日志
└── *_test.go                          覆盖率 ≥80%

go/services/i18n/admin/
├── service.go
├── string_service.go                  template_validator + 落库 + 广播
├── pack_service.go                    版本号自增 + 失效
├── import_export.go                   CSV 解析（两阶段 dry-run→事务）
├── audit.go
└── *_test.go                          覆盖率 ≥80%

typescript/admin-web/                  ← go-admin-ui **v2.0.9** fork（**Vue 2.6.11 + Element UI**，vue-cli-service）
└── src/views/
    ├── config/
    │   ├── schema-list.vue
    │   ├── schema-form.vue
    │   ├── value-list.vue
    │   └── value-form.vue
    └── i18n/
        ├── string-list.vue            含多语言并排
        ├── string-form.vue
        ├── pack-list.vue
        └── import-export.vue
```

每个文件 ≤200 行、每个函数 ≤50 行。

---

## 六、API 设计

### 6.1 配置 Schema

```
GET    /api/admin/v1/config/schema?module_key=&page=&size=
POST   /api/admin/v1/config/schema
PUT    /api/admin/v1/config/schema/:id
DELETE /api/admin/v1/config/schema/:id
GET    /api/admin/v1/config/schema/:id
```

请求体（POST/PUT）：

```json
{
  "tenant_id": "default",
  "module_key": "hello_cfg",
  "field_key": "max_name_length",
  "field_type": "string",
  "default_value": "32",
  "validator": "{\"min\":1,\"max\":256}",
  "is_required": true,
  "is_secret": false,
  "client_scope": "all",
  "min_app_ver": "",
  "description": "用户名最大长度"
}
```

校验规则（`services/config/admin` 层调用 `service.ValidateFieldValue()`）：
- field_type 合法（string/int/bool/float/enum/json/list）
- default_value 能按 field_type 解析
- validator JSON 合法且自洽（required / regex: / range: / enum:）
- module_key + field_key 唯一（同 tenant 下）
- 删除前必须先把 sys_config_version 中对应字段清空

### 6.2 配置 Value

```
GET    /api/admin/v1/config/value?env=prod&module_key=
POST   /api/admin/v1/config/value         批量保存某 module 的所有字段
GET    /api/admin/v1/config/value/:env/:module_key
```

请求体：

```json
{
  "tenant_id": "default",
  "env": "prod",
  "module_key": "hello_cfg",
  "fields": {
    "max_name_length": "32",
    "server_name": "CaiRobot"
  },
  "publish": true
}
```

行为：
- `services/config/admin.ValueService.Save()`
- 校验每个字段必须存在 schema、值符合 type/validator（调用 `service.ValidateConfigMap()`）
- 写 sys_config_version（version 自增整数）
- 失效 Redis: `config:value:{tenant}:{env}:{module_key}:*`（调用 `redisx.Client.Invalidate()`）
- 发布 pub/sub: `cairobot.config.invalidate`（payload 为 JSON InvalidateEvent）

### 6.3 i18n String

```
GET    /api/admin/v1/i18n/string?lang_code=&group=&key_like=&page=&size=
POST   /api/admin/v1/i18n/string                    新增（含全部语言一次性提交）
PUT    /api/admin/v1/i18n/string/:key                修改
DELETE /api/admin/v1/i18n/string/:key                软删除（operation_type=DEL）
GET    /api/admin/v1/i18n/string/:key                详情
```

请求体（多语言并排）：

```json
{
  "tenant_id": "default",
  "key": "svc_hello_greeting",
  "group": "app",
  "template_type": "named",
  "params_schema": [
    {"name":"name","type":"string","required":true},
    {"name":"server_name","type":"string","required":true}
  ],
  "preview_sample": {"name":"张三","server_name":"CaiRobot"},
  "values": {
    "zh-CN": "你好，{name}！欢迎使用 {server_name}。",
    "en":    "Hello, {name}! Welcome to {server_name}."
  }
}
```

强校验（调用 `service.ValidateTemplate()` + `ValidateCrossLanguagePlaceholders()`）：
- value 中 {xxx} 占位符必须出现在 params_schema
- params_schema 中 required=true 的参数必须出现在所有语言 value 中
- 不同语言 value 占位符集合必须一致
- preview_sample 必须能 plain/named/icu 渲染成功
- ICU 模板用现成库做语法校验

### 6.4 i18n Pack

```
GET  /api/admin/v1/i18n/pack
POST /api/admin/v1/i18n/pack/publish     按 tenant + lang_code 自增 version + 失效广播
```

请求体：

```json
{
  "tenant_id": "default",
  "env": "prod",
  "lang_code": "zh-CN",
  "description": "新增 svc_demo_msg"
}
```

行为：
- sys_lang_pack.version 自增 1
- 失效 Redis: `i18n:pack:{tenant}:{env}:{lang_code}:*` 与 `i18n:diff:{tenant}:{env}:{lang_code}:*`
- 发布 `cairobot.i18n.invalidate`（payload 为 JSON InvalidateEvent）

### 6.5 i18n CSV 导入导出

```
POST /api/admin/v1/i18n/string/import     multipart/form-data，CSV
GET  /api/admin/v1/i18n/string/export?lang_code=
```

CSV 列：`key,group,template_type,params_schema_json,zh-CN,en`

行为（两阶段）：
- **阶段一 Dry-Run**：解析 CSV → 内存构造 LangString → 逐条调 `ValidateTemplate()` → 任一失败返回行级错误列表
- **阶段二事务写入**：仅 dry-run 全通过才执行，单事务批量 INSERT，上限 100 条/事务，任一失败整体回滚
- 性能约束：100 条 ≤2 秒

---

## 七、pub/sub 协议升级

### 7.1 新 payload 格式（InvalidateEvent）

```json
{
  "tenant_id": "default",
  "scope": "config",
  "env": "prod",
  "module_keys": ["hello_cfg", "base_cfg"],
  "lang_codes": [],
  "version": 42,
  "timestamp": 1716739200,
  "trace_id": "uuid-v4"
}
```

类型定义位置：`go/services/config/sdk/types.go`

```go
type InvalidateEvent struct {
    TenantID   string   `json:"tenant_id"`
    Scope      string   `json:"scope"`                 // "config" | "i18n"
    Env        string   `json:"env"`
    ModuleKeys []string `json:"module_keys,omitempty"` // config 用
    LangCodes  []string `json:"lang_codes,omitempty"`   // i18n 用
    Version    int64    `json:"version"`
    Timestamp  int64    `json:"timestamp"`
    TraceID    string   `json:"trace_id,omitempty"`
}
```

### 7.2 SDK 消费端升级（向后兼容）

**config SDK** (`go/services/config/sdk/pubsub.go` onMessage)：

```go
func (p *pubsubManager) onMessage(msg string) {
    var evt InvalidateEvent
    if err := json.Unmarshal([]byte(msg), &evt); err == nil && evt.TenantID != "" {
        p.handleStructured(evt)
        return
    }
    // 向后兼容：降级解析逗号分隔字符串（旧格式）
    log.Warnw("received legacy pub/sub format, migrating to JSON", "msg", msg[:min(len(msg), 80)])
    moduleKeys := strings.Split(msg, ",")
    p.handleLegacy(moduleKeys)
}
```

**i18n SDK** 升级方式待 C.1 现状摘要后由主控拍板（候选方案 A/B/C，默认倾向方案 A：对齐 redisx.PubSubClient）。

### 7.3 版本切换策略

- M0' 阶段：SDK 同时支持新旧两种格式
- M6' 验收：admin 端只发 JSON 格式
- S2 阶段：观察 1 个月无降级日志后，移除旧格式分支

---

## 八、Redis 扩展：Invalidate 方法

### 8.1 接口扩展

在 `go/third_party/redisx/redisx.go` 的 `Client` 接口新增：

```go
type Client interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Scan(ctx context.Context, pattern string) ([]string, error)
    Invalidate(ctx context.Context, pattern string) error    // ← 新增
    Ping(ctx context.Context) error
    Close() error
}
```

### 8.2 实现

```go
func (r *redisClient) Invalidate(ctx context.Context, pattern string) error {
    keys, err := r.Scan(ctx, pattern)
    if err != nil {
        return fmt.Errorf("redisx invalidate scan: %w", err)
    }
    if len(keys) == 0 {
        return nil
    }
    const batchSize = 500
    for i := 0; i < len(keys); i += batchSize {
        end := i + batchSize
        if end > len(keys) {
            end = len(keys)
        }
        if err := r.Delete(ctx, keys[i:end]...); err != nil {
            return fmt.Errorf("redisx invalidate delete batch [%d:%d]: %w", i, end, err)
        }
    }
    return nil
}
```

### 8.3 测试覆盖（miniredis，≥80%）

| 用例 | 说明 |
|---|---|
| 空匹配 | pattern 命中 0 key |
| 单 key | 命中 1 个 |
| 多 key < 500 | 单批删除 |
| 多 key > 500 | 分批删除（建议 1200 个 key 验证 3 批） |
| Scan 失败 | Redis 连接异常 |
| Delete 失败 | 部分删除异常 |
| ctx 取消 | 超时取消传播 |

---

## 九、依赖注入约定

admin 子包的标准依赖结构：

```go
// services/config/admin/schema_service.go
type AdminSchemaService struct {
    inner   *service.SchemaService          // 复用 service 层校验（核心！）
    repo    domain.SchemaRepository          // 持久化
    cache   redisx.Client                    // 缓存失效（走 Invalidate）
    bus     redisx.PubSubClient              // 变更广播（走 Publish）
    audit   audit.Logger                     // 审计日志
}

type AdminValueService struct {
    inner   *service.SchemaService
    repo    domain.ConfigRepository
    cache   redisx.Client
    bus     redisx.PubSubClient
    audit   audit.Logger
}

// services/i18n/admin/string_service.go
type AdminStringService struct {
    inner   *service.I18nServiceImpl         // 复用 service 层 TemplateValidator
    repo    domain.StringRepository
    cache   redisx.Client
    bus     redisx.PubSubClient
    audit   audit.Logger
}
```

关键约束：
- **失效走 `redisx.Client.Invalidate(ctx, pattern)`**
- **广播走 `redisx.PubSubClient.Publish(ctx, channel, payloadJSON)`**
- **不在 `redisx.Client` 接口上扩展 Publish**

---

## 十、admin 子包复用 service 层校验的精确实现

### 10.1 config 侧：SchemaService 新增方法

在 `go/services/config/service/schema_service.go` 新增（M0' 末或 M1' 起手）：

```go
// ValidateSchema 仅做校验，不落库不广播
// 供 admin 子包及未来其它写入路径复用
// 内部委托给 validate.go 包级函数 ValidateFieldValue
//
// 校验内容：
//   1. field_type 合法性（string/int/bool/float/enum/json/list）
//   2. default_value 能按 field_type 解析为 TypedValue
//   3. validator JSON 合法且自洽
func (s *SchemaService) ValidateSchema(schema *domain.FieldSchema) error {
    if schema == nil {
        return ErrInvalidSchemaInput
    }

    if !isValidFieldType(schema.FieldType) {
        return ValidationError{Field: schema.FieldKey, Reason: "非法字段类型: " + string(schema.FieldType)}
    }

    tv, err := parseTypedValue(schema.FieldType, schema.DefaultValue)
    if err != nil {
        return ValidationError{Field: schema.FieldKey, Reason: "默认值解析失败: " + err.Error()}
    }

    if schema.Validator != "" {
        if err = ValidateFieldValue(tv, schema); err != nil {
            return err
        }
    }
    return nil
}

// isValidFieldType 判断字段类型是否在合法枚举中
func isValidFieldType(ft domain.FieldType) bool {
    switch ft {
    case domain.FieldTypeString, domain.FieldTypeInt, domain.FieldTypeBool,
         domain.FieldTypeFloat, domain.FieldTypeEnum, domain.FieldTypeJSON,
         domain.FieldTypeList:
        return true
    }
    return false
}

// parseTypedValue 按 fieldType 将原始字符串解析为 TypedValue
func parseTypedValue(ft domain.FieldType, raw string) (*domain.TypedValue, error) {
    var val any
    switch ft {
    case domain.FieldTypeString:
        val = raw
    case domain.FieldTypeInt:
        var n int64
        if err := json.Unmarshal([]byte(raw), &n); err != nil {
            return nil, err
        }
        val = n
    case domain.FieldTypeBool:
        val = raw == "true" || raw == "1"
    case domain.FieldTypeFloat:
        var f float64
        if err := json.Unmarshal([]byte(raw), &f); err != nil {
            return nil, err
        }
        val = f
    case domain.FieldTypeEnum, domain.FieldTypeJSON, domain.FieldTypeList:
        val = raw
    default:
        return nil, fmt.Errorf("unknown field type: %s", ft)
    }
    return domain.NewTypedValue(ft, val), nil
}
```

> **注意**：以上代码调用的是 `validate.go` 中已有的包级纯函数 `ValidateFieldValue()`，**不是** `s.validator.Validate()`（SchemaService 没有 validator 字段）。

### 10.2 config 侧：admin 子包调用示例

```go
// services/config/admin/schema_service.go
func (s *AdminSchemaService) Create(ctx context.Context, req CreateSchemaDTO) error {
    schema := req.ToDomain()

    // ① DTO → domain 完成

    // ② 复用 service 层校验（调用 SchemaService.ValidateSchema，
    //     其内部委托给 validate.go 包级函数 ValidateFieldValue）
    if err := s.inner.ValidateSchema(schema); err != nil {
        return ErrValidation.Wrap(err)
    }

    // ③ 落库
    if err := s.repo.Create(ctx, schema); err != nil {
        return err
    }

    // ④ 审计日志
    s.audit.Log(ctx, "create_schema", schema)

    // ⑤ 失效缓存（不阻塞主流程）
    pattern := fmt.Sprintf("config:schema:%s:*", schema.TenantID)
    if err := s.cache.Invalidate(ctx, pattern); err != nil {
        log.Warnw("invalidate failed after create schema", "err", err, "pattern", pattern)
    }

    // ⑥ 广播（构造 InvalidateEvent JSON）
    payload, _ := json.Marshal(InvalidateEvent{
        TenantID:   schema.TenantID,
        Scope:      "config",
        Env:        "",
        ModuleKeys: []string{schema.ModuleKey},
        Version:    0,
        Timestamp:  time.Now().Unix(),
    })
    if err := s.bus.Publish(ctx, "cairobot.config.invalidate", string(payload)); err != nil {
        log.Warnw("publish failed after create schema", "err", err)
    }

    return nil
}
```

### 10.3 i18n 侧：I18nServiceImpl 新增方法

在 `go/services/i18n/service/template_validator.go` 或新建 `lang_service.go`：

```go
// ValidateLangString 校验完整的多语言字符串条目
// 组合调用已有的 ValidateTemplate + ValidateCrossLanguagePlaceholders
// 供 admin 子包复用
func (s *I18nServiceImpl) ValidateLangString(ls *domain.LangString) error {
    params, err := ls.GetParams()
    if err != nil {
        return err
    }

    if err := s.ValidateTemplate(ls.StringValue, ls.TemplateType, params); err != nil {
        return err
    }

    return nil
}
```

---

## 十一、前端页面设计（admin-web）

基于 go-admin-ui **v2.0.9**（**Vue 2.6.11 + Element UI**），不改主题。i18n 自身文案用 vue-i18n + 静态 JSON（zh-CN + en），不复用业务 sys_lang_string。

### 11.1 Schema 列表页（schema-list.vue）

```
顶部筛选：[模块下拉] [字段类型下拉] [是否必填] [关键字搜索] [刷新] [新增]
表格列：module_key | field_key | type | default | required | secret | scope | desc | 操作
操作：编辑 | 删除（二次确认）
特殊：按 module_key 分组折叠（可选）
```

### 11.2 Schema 表单页（schema-form.vue）

```
字段：tenant_id(隐藏) | module_key | field_key | field_type(select) |
      default_value(动态渲染) | validator(textarea+模板) | is_required(switch) |
      is_secret(switch) | client_scope(select) | min_app_ver | description
提交前前端先校验一遍，后端再校验一遍
```

### 11.3 Value 列表页（value-list.vue）

```
顶部：[env 切换 dev/test/prod] [module_key 选择] [刷新]
表格：field_key | 当前值 | 类型 | 默认值 | 是否覆盖默认 | 描述
底部：[批量保存] [发布]，保存按钮置灰直到 dirty
```

### 11.4 Value 表单页（value-form.vue）

```
按 schema 动态渲染：string→input | int/float→number | bool→switch |
                       enum→select | json→JSON编辑器 | list→标签输入
实时校验（schema.validator），右侧固定预览面板展示完整 JSON
```

### 11.5 i18n String 列表页（string-list.vue）

```
顶部筛选：[group 下拉] [template_type 下拉] [key 模糊搜索] [刷新] [新增] [导入] [导出]
表格（多语言并排）：key | group | type | zh-CN | en | params 数量 | 操作
操作：编辑 | 删除 | 复制
行内编辑：plain 类型可双击快速改；named/icu 必须进完整表单
```

### 11.6 i18n String 表单页（string-form.vue）

```
左半区：key | group(select) | template_type(radio) |
         params_schema(动态表格) | preview_sample(JSON编辑器)
右半区：多语言 Tab/并排(zh-CN/en textarea) |
         占位符 {xxx} 高亮 | 实时校验（红=缺失/黄=不一致） |
         实时预览（preview_sample 渲染结果）
底部：[取消] [保存草稿] [保存并发布]
```

独立组件：
- `components/PlaceholderHighlight.vue` — 占位符高亮
- `components/IcuPreview.vue` — ICU 渲染预览（messageformat 库）
- `components/LangSideBySide.vue` — 多语言并排编辑

### 11.7 Pack 管理页 + 导入导出页

```
pack-list.vue：表格(env | lang_code | version | 最近发布时间 | 描述 | 操作)
import-export.vue：
  导入：CSV 上传 → 预览前10行 → 行级校验状态 → 全通过才出现"确认导入"
  导出：选 lang_code(group多选) → 下载 CSV
```

---

## 十二、执行批次（M0' ~ M7'）

### M0'：环境锁定 + go-admin 安装配置 + 5 个阻塞问题整改（1 天）

#### 前置现状

执行前仓库状态：
- `go/admin/` 目录 **不存在**（需新建）
- `typescript/admin-web/` 仅有一个空 README.md（需填充）
- `go/work` 中无 admin 模块（仍为旧的 `./tars/provider-admin`）
- go-admin / go-admin-ui 依赖 **未引入**

---

#### 0.1：锁定并安装 go-admin 后端框架

**步骤 1：查询最新稳定版本**

```bash
git ls-remote --tags https://github.com/go-admin-team/go-admin \
  | grep -v '{}' | grep -v 'beta\|rc\|alpha' | tail -10
```

输出示例：
```
abc1234...    refs/tags/v2.0.0
def5678...    refs/tags/v2.0.1
```

记录最终选定的 **tag 名**（如 `v2.x.x`）和 **commit hash**（如 `a1b2c3d`），写入可行性报告。

**步骤 2：创建 go/admin 目录骨架**

```bash
mkdir -p go/admin/{cmd/admin,config,plugins/{config_admin,i18n_admin}/{apis,service,models,router}}
cd go/admin

# 初始化 Go module（module 路径与项目一致）
go mod init github.com/jimiechen/mineplanet/go/admin
```

创建最小入口文件 `cmd/admin/main.go`：

```go
package main

import (
    "log"
    _ "github.com/go-admin-team/go-admin/core"
)

func main() {
    log.Println("go-admin entry point (M0' placeholder)")
}
```

**步骤 3：引入 go-admin 核心依赖**

在 `go/admin/go.mod` 中暂不手动写 require，先通过 import 触发自动拉取：

编辑 `cmd/admin/main.go` 为：

```go
package main

import (
    "github.com/gin-gonic/gin"
    _ "github.com/go-admin-team/go-admin/core"
    _ "github.com/go-admin-team/go-admin/modules/db"
    _ "github.com/go-admin-team/go-admin/modules/language"
    _ "github.com/go-admin-team/go-admin/modules/menu"
    _ "github.com/go-admin-team/go-admin/plugins/admin"
)

func main() {}
```

然后执行：

```bash
cd go/admin
go mod tidy
```

`go mod tidy` 会自动从 Go module proxy 拉取 go-admin 及其全部依赖（Gin/GORM/Casbin/golang-jwt 等）。

**步骤 4：验证依赖拉取成功**

```bash
cd go/admin
go list -m all | grep go-admin
# 预期输出类似：
# github.com/go-admin-team/go-admin v2.x.x
# github.com/go-admin-team/go-admin/core v2.x.x
# github.com/go-admin-team/go-admin/plugins/admin v2.x.x
```

**步骤 5：依赖冲突检查**

对比 go-admin 引入的依赖与项目现有依赖：

| 依赖 | 项目当前版本 | go-admin 默认版本 | 冲突？ |
|---|---|---|---|
| Gin | TarsGo 内置 | go-admin 自带 | 需确认 |
| GORM | mysqlx 使用 | go-admin 自带 | 需确认 |
| Casbin | — | go-admin RBAC 用 | 新增 |
| golang-jwt | — | go-admin 认证用 | 新增 |

将结果填入 `docs/reports/go-admin-feasibility-report.md` 的依赖冲突矩阵表。若存在大版本冲突（如 Gin v1 vs 不兼容），记录到报告中并提出解决方案（replace / upgrade / vendor）。

> **注意**：go/admin 是独立 Go module（有自己的 go.mod），不与 go/services/config 或 go/tars/ 共享 go.mod，因此依赖隔离天然由 Go workspace 保证。

**通过判据**：

- □ 给出 go-admin 最终 tag + commit hash
- □ `go/admin/go.mod` 已创建且 `go mod tidy` 成功
- □ `go list -m all \| grep go-admin` 输出非空
- □ 依赖冲突矩阵已填写（有/无冲突均需说明）

---

#### 0.2：安装配置 go-admin-ui 前端 + 启动验证

**步骤 1：获取 go-admin-ui 源码**

go-admin-ui 是独立前端仓库，需要确认官方发布方式：

```bash
# 方式 A：检查 npm 是否有官方包
npm view go-admin-ui version 2>/dev/null && echo "npm 包可用" || echo "npm 包不可用"

# 方式 B：直接 clone 官方 Git 仓库（推荐）
git clone --depth 1 \
  -b $(git ls-remote --tags https://github.com/go-admin-team/go-admin-ui \
       2>/dev/null | grep -v '{}' | grep -v 'beta\|rc\|alpha' | tail -1 | awk '{print $2}' | sed 's/refs.tags//') \
  https://github.com/go-admin-team/go-admin-ui.git \
  typescript/admin-ui-temp
```

**步骤 2：将前端源码整合到 typescript/admin-web/**

```bash
# 如果 typescript/admin-web/ 已有空壳内容，先清空（保留 .gitignore 如有）
rm -rf typescript/admin-web/src typescript/admin-web/package.json \
      typescript/admin-web/index.html typescript/admin-web/vite.config.* \
      typescript/admin-web/tsconfig.*

# 将 clone 的源码核心内容复制过来
cp -r typescript/admin-ui-temp/src   typescript/admin-web/
cp    typescript/admin-ui-temp/package.json   typescript/admin-web/
cp    typescript/admin-ui-temp/index.html     typescript/admin-web/
cp    typescript/admin-ui-temp/vite.config.*  typescript/admin-web/
cp    typescript/admin-ui-temp/tsconfig.*     typescript/admin-web/

# 清理临时目录
rm -rf typescript/admin-ui-temp
```

**步骤 3：安装前端依赖**

```bash
cd typescript/admin-web
pnpm install
```

**步骤 4：更新 go.work 注册 admin 模块**

编辑 `go/go.work`，将：
```
./tars/provider-admin
```
替换为：
```
./admin
```

完整 go.work 内容应为：

```
go 1.25.5

use (
    ../proto/generated/go
    ./common-lib
    ./gateway/proto-gateway
    ./modules/health
    ./modules/hello
    ./services/config
    ./services/i18n
    ./tars/config
    ./tars/i18n
    ./admin                              ← 替换原来的 provider-admin
    ./tars/system
    ./third_party/mysqlx
    ./third_party/redisx
    ./third_party/sqlitex
)
```

然后同步：

```bash
cd go
go work sync
```

**步骤 5：后端启动验证**

创建 `go/admin/config/settings.yml`（M0' 最小配置）：

```yaml
url_prefix: /admin
store:
  path: ./data/admin.db
language:
    languages:
        {zh-CN: 简体中文,en: English}
```

创建 `go/admin/cmd/admin/main.go` 最小可启动版本：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/go-admin-team/go-admin/core"
    "github.com/go-admin-team/go-admin/engine"
    "github.com/go-admin-team/go-admin/plugins/admin"
    "github.com/go-admin-team/go-admin/modules/config"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.ReadConfig(config.ConfigPath("config/settings.yml"))

    eng := engine.Default()

    eng.Use(gin.Recovery())
    eng.UsePublic()

    eng.AddConfigFromPath("config/settings.yml")

    eng.InjectRouter(
        admin.Admin(ctx.Background()),
    )

    eng.SetMenu(admin.GetAdminMenu())

    srv := &http.Server{
        Addr:         ":8000",
        Handler:      eng.RunUnix(),
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    log.Printf("[go-admin] server starting at :8000")
    log.Printf("[go-admin] default login: admin / admin")

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
    fmt.Println("Server exiting")
}
```

启动后端：

```bash
cd go/admin
go run cmd/admin/main.go
```

预期日志输出：
```
[go-admin] server starting at :8000
[go-admin] default login: admin / admin
[go-admin] init menu & permission data success
```

浏览器访问 `http://localhost:8000/admin` → 看到 go-admin 默认登录页 → 输入 `admin / admin` 登录成功 → 进入默认管理后台面板 ✅

**步骤 6：前端启动验证**

新开终端：

```bash
cd typescript/admin-web
pnpm dev
```

预期：前端 dev server 启动（通常 proxy 到后端 8000），浏览器展示 go-admin-ui 默认管理界面 ✅

**通过判据**：

- □ `go/work` 中已注册 `./admin`（不再有 provider-admin）
- □ `go run go/admin/cmd/admin/main.go` 启动无报错
- □ 浏览器 `http://localhost:8000/admin` 可访问
- □ `admin / admin` 登录成功进入默认面板
- □ `pnpm dev` 在 typescript/admin-web/ 下启动成功
- □ 前端页面正常渲染（菜单、表格、按钮可见）

---

#### 0.3：三个 DSN 配置 + 加密存储

**DSN 加密方案**：

- 算法：AES-256-CBC
- 密钥来源：环境变量 `ADMIN_DSN_KEY`（32 字节 hex 编码的 AES key）
- 存储位置：`go/admin/config/settings.yml` 中的 DSN 字段值为 Base64(AES加密(明文DSN))
- go-admin 自带 config.Encrypt 工具或自定义加解密函数

**settings.yml DSN 配置段**：

```yaml
database:
    driver: mysql
    # ops_dsn 为 AES 加密后的 Base64 字符串
    # 明文格式: user:password@tcp(host:3306)/ops_db?charset=utf8mb4&parseTime=True&loc=Local
    ops_dsn: "${ADMIN_OPS_DSN_ENCRYPTED}"
    config_dsn: "${ADMIN_CONFIG_DSN_ENCRYPTED}"
    i18n_dsn: "${ADMIN_I18N_DSN_ENCRYPTED}"

redis:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
```

**密钥初始化脚本**（一次性）：

```bash
# 生成随机 32 字节 AES-256 密钥
openssl rand -hex 32 > configs/keys/dev.key

# 设置环境变量指向密钥文件
export ADMIN_DSN_KEY=$(cat configs/keys/dev.key)
```

**DSN 加解密工具函数**（放在 `go/admin/config/dsn.go`）：

```go
package config

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
    "os"
    "strings"
)

func GetDSN(envKey string) (string, error) {
    encrypted := os.Getenv(envKey)
    if encrypted == "" {
        return "", fmt.Errorf("environment variable %s is empty", envKey)
    }

    keyHex := os.Getenv("ADMIN_DSN_KEY")
    if keyHex == "" {
        return "", fmt.Errorf("ADMIN_DSN_KEY environment variable not set")
    }

    key, err := hex.DecodeString(keyHex)
    if err != nil || len(key) != 32 {
        return "", fmt.Errorf("invalid ADMIN_DSN_KEY: must be 32-byte hex")
    }

    ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
    if err != nil {
        return "", fmt.Errorf("failed to decode base64 DSN: %w", err)
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }

    if len(ciphertext) < aes.BlockSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    iv := ciphertext[:aes.BlockSize]
    ciphertext = ciphertext[aes.BlockSize:]

    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(ciphertext, ciphertext)

    padding := ciphertext[len(ciphertext)-1]
    if int(padding) > len(ciphertext) || padding == 0 || padding > aes.BlockSize {
        return "", fmt.Errorf("invalid padding")
    }

    return string(ciphertext[:len(ciphertext)-int(padding)]), nil
}

func EncryptDSN(plainDSN string, keyHex string) (string, error) {
    key, err := hex.DecodeString(keyHex)
    if err != nil || len(key) != 32 {
        return "", fmt.Errorf("invalid key")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    plaintext := []byte(plainDSN)
    plaintext = PKCS7Pad(plaintext, block.BlockSize())

    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
        return "", err
    }
    copy(ciphertext[:aes.BlockSize], iv)

    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func PKCS7Pad(data []byte, blockSize int) []byte {
    pad := blockSize - len(data)%blockSize
    b := make([]byte, len(data)+pad)
    copy(b, data)
    for i := len(data); i < len(b); i++ {
        b[i] = byte(pad)
    }
    return b
}
```

**自检：DSN 不泄露到日志**

```bash
# 启动后 grep 确认 DSN 不出现在 stdout/stderr
go run cmd/admin/main.go 2>&1 | tee /tmp/admin-log.txt
grep -E "(mysql|password|@tcp)" /tmp/admin-log.txt
# 预期：无匹配（DSN 以加密形式存储，运行时不打印明文）
```

**密钥轮换文档**（M0'.0.3 输出）：`docs/runbook/admin-dsn-key-rotation.md`

**通过判据**：

- □ 三个 DSN 可正确解密连接 MySQL
- □ `grep` 日志无明文 DSN 泄露
- □ 密钥轮换文档已输出
- □ 多环境密钥文件存在：`configs/keys/{dev,test,prod}.key`

---

#### 0.4：provider-admin 废弃迁移（前后端一并归档+删除）

| 任务 | 内容 | 通过判据 |
|---|---|---|
| **0.4** | **provider-admin 废弃迁移（前后端一并归档+删除）** | grep 全局无残留；build 通过；归档分支含前后端 |
| **0.5** | **redisx.Invalidate API 实现（含接口声明+分批删除+单测≥80%）** | miniredis 测试全 PASS |
| 0.6 | Publish 接入约定文档化（CODE-WIKI §9） | 文档已更新 |
| **0.7** | **pub/sub payload 升级**：A) 定义 InvalidateEvent；B) config SDK onMessage 升级（JSON优先+降级）；**C) i18n SDK 先输出现状摘要等主控拍板** | config SDK 单测 PASS；i18n 摘要已交 |
| 0.8 | go-admin 可行性报告（依赖冲突矩阵 + POC 6 项 + 风险评估） | 结论"继续" |

**0.4 详细步骤**：

```
a. git checkout -b archive/provider-admin-v0
   归档内容必须同时包含：
   - go/tars/provider-admin/        整目录
   - web/provider-admin/             整目录
   一次性提交后切回主分支

b. rm -rf go/tars/provider-admin
   rm -rf web/provider-admin

c. 全局 grep 清理：
   grep -rn "provider-admin\|tars/provider-admin\|web/provider-admin" \
     go/ web/ docs/ scripts/ ci/ deploy/ \
     package.json pnpm-workspace.yaml turbo.json
   全部删除或替换

d. 重点检查清单：
   □ go.work
   □ Makefile
   □ docker-compose.yml / docker-compose.*.yml
   □ CI workflow
   □ 顶层 package.json / pnpm-workspace.yaml / turbo.json
   □ CODE-WIKI.md（第348/496/560/607/648行有引用需更新）
   □ README 与部署文档

e. 输出 docs/migration/provider-admin-to-go-admin.md
   含：前后端废弃理由、业务逻辑映射表、数据库表延续说明、归档分支位置、保留期至2026-11
```

**0.7.C i18n SDK 现状摘要要求**：

Trae 在执行 C.2 前必须输出 `docs/reports/i18n-sdk-pubsub-current.md`，包含：

```
- 订阅回调签名（参数类型与个数）
- 解析入口函数名与位置
- channel 常量定义
- 是否依赖 redisx.PubSubClient 还是另一套抽象
- 当前消息格式约定
- 外部依赖方数量评估
```

主控收到摘要后选择 A/B/C（默认倾向 A：对齐 redisx.PubSubClient）。

**M0' 整体通过判据**：
- 5 个阻塞问题全部修复
- redisx.Invalidate 单测 PASS（≥80%）
- config SDK pubsub 兼容单测 PASS（≥80%）
- provider-admin 全局无残留（含 web/ 和 docs/wiki/）
- 可行性报告结论"继续"
- DSN 加密 round-trip 验证通过
- **i18n SDK 现状摘要已提交（C.1），等待主控拍板后再实施 C.2/C.3**

---

### M1'：services/{config,i18n}/admin 子包（1.5 天）

| 任务 | 内容 | 通过判据 |
|---|---|---|
| 1.1 | config/admin/: schema_service + value_service + audit + service.go + doc.go + test ≥80% | 单文件≤200行；覆盖率≥80% |
| 1.2 | i18n/admin/: string_service + pack_service + import_export + audit + test ≥80% | 同上 |
| 1.3 | 失效+广播：cache=redisx.Client.Invalidate；bus=redisx.PubSubClient.Publish；payload=InvalidateEvent JSON | 端到端：admin写→SDK 100ms感知 |
| 1.4 | lint 隔离 grep 通过 | 业务服务不引用 admin |
| 1.5 | 单测 SQLite+miniredis；集成测试 build tag=integration | 集成测试 PASS |
| **1.6** | **职责边界自检脚本 scripts/lint/admin_boundary_check.sh** | **命中 ValidateSchema/ValidateValue/ValidateLangString；不含字段级校验关键字** |

**1.1 关键约束（偏差 1 最终订正）**：

```text
admin 子包必须持有 inner *service.SchemaService（或 I18nServiceImpl）
admin 校验调用 inner.ValidateSchema() / inner.ValidateLangString()
admin 层禁止出现：field_type== / switch field_type / validator JSON / default_value 解析
admin 仅负责：DTO转换 → inner校验 → repo落库 → audit → cache.Invalidate → bus.Publish
```

---

### M2'：go-admin config_admin 插件（1.5 天）

- apis/schema.go + apis/value.go（≤180行/个）
- service/（DTO 转换）、models/（请求响应 DTO）
- router/router.go + plugin.go
- RBAC 接入：菜单"系统配置→Schema管理/值管理"，权限点 config:schema:read/write/delete、config:value:read/write
- 操作日志接入 sys_oper_log
- 路由前缀 `/api/admin/v1/config/...`
- 单测≥80%；集成测试端到端 PASS
- **grep 自检：admin 插件内不得出现 sys_config_schema/sys_config_version 直写**

---

### M3'：go-admin i18n_admin 插件（2 天）

- apis/string.go + apis/pack.go + apis/import_export.go
- 强校验：全部保存动作先经 `service.TemplateValidator`，失败返回 10400 + 字段级错误
- CSV 导入两阶段（dry-run→事务写入），单事务≤100条
- RBAC：i18n:string:*、i18n:pack:*
- 单测≥80%
- 集成端到端：named key 渲染正确、ICU plural 三分支正确、CSV 100条≤2秒
- **grep 自检：admin 插件内不得出现 sys_lang_pack/sys_lang_string 直写**

---

### M4'：admin-web 配置页面（2 天）

- src/views/config/ 下 4 个 vue 页面
- schema-form：field_type 切换时 default_value/validator 动态渲染
- value-form：按 schema 动态渲染表单 + 实时校验
- 路由 + 菜单注册 + 权限点 UI 对齐
- admin-web 自身 i18n：vue-i18n + locales/{zh-CN,en}.json
- vitest 单测≥60%（schema-form 校验 + value-form 动态渲染）
- pnpm build 成功 + 人工验收 4 页面

---

### M5'：admin-web i18n 页面（2 天）

- src/views/i18n/ 下 4 个 vue 页面
- PlaceholderHighlight.vue / IcuPreview.vue / LangSideBySide.vue 独立组件（各单测≥60%）
- ICU 预览用 messageformat 库
- vitest 单测≥60%
- pnpm build 成功 + 人工验收 4 页面

---

### M6'：端到端联动验收（1 天）

| 用例 | 验证内容 |
|---|---|
| T-CONFIG | admin 新增 schema→新增 value→发布→SDK GetString 返回新值→改值→SDK 100ms 内返回 |
| T-I18N-NAMED | admin 新增 named key→发布 pack→SDK T() 渲染正确→改文案→SDK 100ms 内返回 |
| T-I18N-ICU | admin 新增 ICU plural key→三分支(=0/one/other)渲染正确 |
| T-PUBSUB-JSON | JSON payload 正确触发 SDK 失效（config + i18n 各验证） |
| T-PUBSUB-LEGACY | 旧逗号分隔格式仍能降级触发失效 |
| T-DEGRADE | Redis 不可用时 admin 仍落库成功（pub/sub 仅 Warn 不阻塞） |
| T-CONCURRENT | 3 个 admin 并发写同一 module_key（乐观锁验证） |
| T-LARGE | 单 module 50 字段 Value 表单渲染 ≤200ms |
| T-SECURITY | validator JSON 注入 / 超长 string / 嵌套深度攻击用例 |
| T-OPER-LOG | 上述操作在 sys_oper_log 中可查 |

边界铁律 grep（必须全部通过）：
```bash
grep -rn "go/admin\|admin-server" go/services/openapi/ go/services/devicegw/ \
  go/services/usercenter/ go/services/aiservice/ go/modules/  # 必须为空
grep -rn "sys_config_schema\|sys_config_version\|sys_lang_pack\|sys_lang_string" \
  go/admin/plugins/  # 必须为空
```

---

### M7'：文档 + ADR + 测试报告（1 天）

| 交付物 | 位置 |
|---|---|
| Admin Config Wiki | docs/wiki/modules/admin-config.md |
| Admin I18n Wiki | docs/wiki/modules/admin-i18n.md |
| ADR-013 | docs/wiki/adr/ADR-013-app-config-i18n-admin-mvp.md |
| CODE-WIKI §9 更新 | docs/wiki/CODE-WIKI.md（新增 admin 章节） |
| LLM-WIKI 蒸馏 | LLM-WIKI.md 新增 admin 条目 |
| 测试报告 | docs/testing/admin-mvp-test-report.md |
| provider-admin 迁移记录 | docs/migration/provider-admin-to-go-admin.md |
| go-admin 可行性报告 | docs/reports/go-admin-feasibility-report.md |
| DSN 密钥轮换 runbook | docs/runbook/admin-dsn-key-rotation.md |
| i18n SDK pubsub 现状摘要 | docs/reports/i18n-sdk-pubsub-current.md |
| 错误码规范 | docs/api/error-codes.md |
| API 版本化策略 | docs/api/admin-versioning.md |
| 部署共存图 | docs/deploy/go-admin-tarsgo-coexist.md（docker-compose + k8s 示例） |
| DDL 迁移脚本 | docs/migration/ddl-tenant-id.sql |

---

## 十三、工期总表

| 批次 | 内容 | 原估算 | 修订后 |
|---|---|---|---|
| M0' | 环境锁定 + 5阻塞整改 | 0.5天 | **1 天** |
| M1' | services/{config,i18n}/admin 子包 | 1天 | **1.5 天** |
| M2' | config_admin 插件 | 1天 | **1.5 天** |
| M3' | i18n_admin 插件 | 1.5天 | **2 天** |
| M4' | admin-web 配置页面 | 1.5天 | **2 天** |
| M5' | admin-web i18n 页面 | 1.5天 | **2 天** |
| M6' | 端到端验收 | 0.5天 | **1 天** |
| M7' | 文档 + ADR | 0.5天 | **1 天** |
| **合计** | | **7~8 天** | **9~10 天** |

加上的 1.5~2 天全部用于技术债清理 + 测试加强 + 评审迭代。

---

## 十四、风险登记册

| 风险ID | 等级 | 说明 | 应对措施 |
|---|---|---|---|
| R-ARCH-001 | R0 | provider-admin 废弃后不可逆 | M0' 必须建归档分支再删除 |
| R-API-002 | R0 | pub/sub payload JSON 与 SDK 旧格式兼容 | 双格式降级机制已设计 |
| R-CODE-003 | R1 | ValidateSchema 伪代码编译错误（FD-4.1/4.2） | **已在本 PRD 中修正为包级函数调用** |
| R-REDIS-004 | R1 | redisx.Invalidate 分批阈值合理性 | batchSize=500 经测试验证 |
| R-TIME-005 | R1 | 总工期 9~10 天仍紧凑 | 每批有明确通过判据，不通过不放行 |
| R-FRAMEWORK-006 | R2 | go-admin 选型经 POC 后仍有隐藏兼容性 | M0'.0.8 可行性报告必须包含风险评估 |
| R-I18N-SDK-007 | R2 | i18n SDK 抽象层与 config SDK 不同 | C.1 现状摘要强制前置，主控拍板 A/B/C |
| R-DUP-008 | R2 | admin 与 service 校验逻辑重复 | 1.6 自检脚本强制拦截 |
| R-FRONTEND-009 | R2 | web/provider-admin 清理遗漏 | 0.4 订正版已扩展清理范围 |

---

## 十四点五、版本锁定与升级路径（M0' 执行后固化）

### 实际锁定版本

| 组件 | 版本 | Commit Hash | 来源 | 关键特征 |
|---|---|---|---|---|
| **go-admin 后端** | **v2.2.0** | `8f8a197db1a50403d332519ad3caa743d65d5f1a` | [deploy/go-admin-2.2.0.zip](../deploy/go-admin-2.2.0.zip) ⬇️ 离线安装包 | 单体应用模板，module=`go-admin`，非 Go 库 |
| **go-admin-core** | **v1.5.3-rc.3** | `2763de5dcdf4` (via go.mod) | go-admin 依赖 | Gin v1.10 + GORM v1.25 + Casbin v2.104 |
| **go-admin-ui 前端** | **v2.0.9** | — | [deploy/go-admin-ui-2.0.9.zip](../deploy/go-admin-ui-2.0.9.zip) ⬇️ 离线安装包 | **Vue 2.6.11** + Element UI + vue-cli-service |
| **Gin** | v1.10.0 | — | go-admin 依赖 | 与 TarsGo 内置 Gin 隔离（不同 module） |
| **GORM** | v1.25.12 | — | go-admin 依赖 | 与 services/mysqlx 隔离（不同 module） |

> ⚠️ **源码包来源说明**：
> 上述两个 zip 包由项目主控提供，存放于 `deploy/` 目录。
> - `go-admin-2.2.0.zip` — 解压后得到完整 go-admin v2.2.0 后端源码
> - `go-admin-ui-2.0.9.zip` — 解压后得到完整 go-admin-ui v2.0.9 前端源码
>
> **后续升级操作**：
> 1. 从 [go-admin-team/go-admin](https://github.com/go-admin-team/go-admin) Releases 页下载新版本 zip
> 2. 从 [go-admin-team/go-admin-ui](https://github.com/go-admin-team/go-admin-ui) Releases 页下载新版本 zip
> 3. 替换 `deploy/` 目录下对应文件
> 4. 按本报告「升级路径」章节执行兼容性检查

### 架构发现（与 PRD 原假设的差异）

```text
PRD 原假设                          实际情况
─────────────────────────────────────────────────────
go-admin 是 Go 库                   → 是单体应用模板（fork 整个仓库）
import "github.com/go-admin-team/   → module 名为 "go-admin"
  go-admin/core"
前端 Vue3 + Element Plus             → Vue 2.6.11 + Element UI
前端 Vite 构建工具                    → vue-cli-service (webpack)
前端 dev 端口未指定                  → 默认 9527，API proxy → :8000
```

### 升级路径记录

```text
当前状态：go-admin v2.2.0 + go-admin-ui v2.0.9

后续升级注意事项：
1. go-admin 大版本升级（v2→v3）需评估：
   - module 路径是否变化
   - core 包接口是否 breaking change
   - 数据库迁移脚本是否兼容

2. go-admin-ui 升级到 Vue3 版本时需注意：
   - 当前为 Vue 2.6.11（v2.0.9 是最后一个 Vue2 版本）
   - Vue3 版本在独立分支或新仓库
   - 升级需要重写组件语法（Options API → Composition API）
   - 建议：本轮先用 v2.0.9 跑通闭环，S2 阶段再评估 Vue3 迁移成本

3. 源码包存放在 deploy/ 目录：
   - go-admin-2.2.0.zip    ← 后端源码备份
   - go-admin-ui-2.0.9.zip ← 前端源码备份
   - 后续升级时替换这两个包即可
```

---

## 十五、验收标准

### 15.1 功能验收

- [ ] 运营人员可通过浏览器完成 Schema CRUD + Value 发布
- [ ] 运营人员可通过浏览器完成 String CRUD + Pack 发布 + CSV 导入导出
- [ ] admin 改配置后，SDK 在 100ms 内感知变更
- [ ] admin 改文案后，SDK 在 100ms 内感知变更（含 named / icu plural）
- [ ] RBAC 三角色权限控制正常
- [ ] 操作日志可查

### 15.2 技术验收

- [ ] 单文件 ≤200 行，单函数 ≤50 行
- [ ] 后端单测覆盖率 ≥80%，前端 ≥60%
- [ ] 集成测试（MySQL + Redis）PASS
- [ ] 端到端三个闭环（配置 / named i18n / ICU i18n）PASS
- [ ] 边界铁律 grep 全部通过
- [ ] 职责边界自检脚本 PASS
- [ ] CSV 100 条导入 ≤2 秒
- [ ] 50 字段 Value 表单渲染 ≤200ms
- [ ] Redis 降级场景 admin 仍能落库
- [ ] 安全攻击用例（注入/超长/嵌套）全部拦截

### 15.3 文档验收

- [ ] 本 PRD 全部章节已覆盖
- [ ] ADR-013 已创建且 Status=Accepted
- [ ] CODE-WIKI / LLM-WIKI 已同步
- [ ] 测试报告含实测数据
- [ ] 迁移记录完整
- [ ] DDL 脚本可执行

---

## 十六、相关文档索引

| 文档 | 位置 | 说明 |
|---|---|---|
| ADR-010 | docs/adr/ADR-010-admin-boundary-sdk.md | Admin 边界与 SDK 引用规范 |
| ADR-013 | docs/wiki/adr/ADR-013-app-config-i18n-admin-mvp.md | 本方案架构决策记录（M7' 交付） |
| CODE-WIKI | docs/wiki/CODE-WIKI.md | Go 语言资产目录树 |
| LLM-WIKI | docs/wiki/LLM-WIKI.md | 大模型知识蒸馏 |
| config SDK | go/common-lib/sdk/configsdk/client.go | 配置 SDK 客户端接口 |
| i18n SDK | go/common-lib/sdk/i18nsdk/client.go | 国际化 SDK 客户端接口 |
| redisx | go/third_party/redisx/ | Redis 客户端封装 |
| config domain | go/services/config/domain/ | 配置领域模型 |
| i18n domain | go/services/i18n/domain/ | 国际化领域模型 |
| config service | go/services/config/service/ | 配置服务层（校验在此） |
| i18n service | go/services/i18n/service/ | 国际化服务层（TemplateValidator 在此） |
| provider-admin 归档 | archive/provider-admin-v0 分支 | 旧实现参考（保留至 2026-11） |
