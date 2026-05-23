# 多语言服务模块 (i18n-service)

## 模块概述

**职责**: 参数化多语言模板的管理与查询服务，支持全量拉取、增量同步和版本轮询。

**不负责**:
- 服务端模板渲染（由客户端 i18nsdk 负责）
- HTTP 路由处理（由 tars/provider-admin 负责）
- UI 语言切换（由 App 前端负责）

**关键输入**:
- `langCode`: 语言代码（zh-CN/en-US）
- `clientVersion`: 客户端版本号（用于兼容性过滤）
- `env`: 环境标识

**关键输出**:
- `LangPackResponse`: 全量语言包，包含所有字符串条目
- `LangDiffResponse`: 增量差异，包含新增和删除的 key

## 目录结构

```
go/services/i18n/
├── domain/              # 领域实体与值对象
│   ├── lang_pack.go     # LangPack 语言包实体
│   ├── lang_string.go   # LangString 字符串条目实体
│   ├── operation.go     # OperationType 操作类型枚举
│   ├── template.go      # TemplateType / LangParam / ValidateTemplate
│   └── version.go       # 版本相关工具函数
├── repository/          # 数据访问层
│   ├── interface.go     # I18nRepository 接口定义
│   ├── mysql_repo.go    # MySQL 持久化实现
│   ├── sqlite_repo.go   # SQLite 实现（测试用）
│   └── mock_repo.go     # Mock 实现（测试用）
├── cache/               # 缓存层
│   ├── interface.go     # Cache 接口定义
│   └── mock_cache.go    # Mock 实现（测试用）
├── service/             # 业务逻辑层
│   ├── interface.go     # I18nService 接口 + I18nServiceImpl 实现
│   ├── pack.go          # 全量语言包组装
│   ├── diff.go          # 增量差异计算
│   ├── language.go      # 语言列表管理
│   ├── template_validator.go  # 模板校验（质量门）
│   └── compat_filter.go      # 客户端版本兼容性过滤
└── sdk/                 # 客户端 SDK
    ├── client.go        # Client 接口 + clientImpl 实现
    ├── translate.go     # T() 渲染 / Raw() 原始 / BatchT 批量
    ├── batch.go         # 批量翻译优化
    ├── watch.go         # Watch 变更订阅
    ├── cache_lru.go     # LRU 缓存实现
    ├── pubsub.go        # Redis Pub/Sub 订阅
    └── remote.go        # 远程调用（TarsGo 占位）
```

**行数约束**: 单文件不超过 200 行，单方法不超过 30 行。

## 核心数据流

### 全量拉取 (Proto 6005/6006)

**场景**: 首次安装或本地缓存失效时使用。

```
Client App
    ↓ Proto 6005 LangPackReq (lang_code, client_version, env)
API Gateway (TarsGo)
    ↓ 调用 I18nService.GetLangPack()
I18nServiceImpl
    ↓ 1. 查 Cache (L2 Redis)
    ↓ 2. miss 则查 Repository (MySQL sys_lang_string)
    ↓ 3. ApplyCompatFilter() 按 client_version 过滤
    ↓ 4. 组装 LangPackResponse
LangPackRsp (Proto 6006)
    ↓ pack_version, strings[]
Client App 缓存完整语言包
```

### 增量拉取 (Proto 6007/6008)

**场景**: 定期增量同步，减少网络流量。

```
Client App
    ↓ Proto 6007 LangDiffReq (lang_code, since_version, client_version)
I18nServiceImpl
    ↓ 1. 查询 since_version 之后的所有变更
    ↓ 2. 计算 additions[] + deletions[]
    ↓ 3. ApplyCompatFilter() 过滤新增条目
LangDiffRsp (Proto 6008)
    ↓ current_version, additions[], deletions[]
Client App 合并到本地缓存
```

### 版本轮询 (Proto 6009/6010)

**场景**: 低频检测是否有新版本，避免无效的全量拉取。

```
Client App
    ↓ Proto 6009 LangVersionReq (lang_code, env)
I18nServiceImpl
    ↓ 对比当前版本与客户端 known_versions
LangVersionRsp (Proto 6010)
    ↓ has_changes: true/false
    ↓ latest_version: int64
Client App 决定是否触发全量/增量拉取
```

## 模板体系

### TemplateType 枚举

```go
type TemplateType string

const (
    TemplatePlain TemplateType = "plain"   // 纯文本，原样返回
    TemplateNamed TemplateType = "named"   // 命名参数模板，{param} 替换
    TemplateIcu   TemplateType = "icu"     // ICU MessageFormat（MVP 未实现）
)
```

### plain 类型

纯文本模板，无占位符，直接返回 value。

**示例**:
- Key: `app.title`
- Value: `CaiRobot 智能学习助手`
- 渲染结果: `CaiRobot 智能学习助手`

### named 类型

命名参数模板，使用 `{paramName}` 占位符，客户端渲染时替换。

**示例**:
- Key: `greeting.welcome`
- Value: `欢迎 {name}，你有 {count} 条新消息`
- Params: `[{name: string, required: true}, {count: int, required: true}]`
- 渲染参数: `{name: "张三", count: 42}`
- 渲染结果: `欢迎 张三，你有 42 条新消息`

### icu 类型（MVP 占位）

ICU MessageFormat 复杂模板，支持复数、性别、选择格式等。MVP 阶段仅保留类型定义，未实现渲染引擎。

**状态**: 返回 `ErrICUNotSupported` 错误。

## template_validator 质量门

`ValidateTemplate()` 是保存前的质量门，确保模板一致性。

### 校验规则

#### plain 类型

```go
func validatePlainTemplate(value string) error {
    // 不应包含任何 {xxx} 占位符
    re := regexp.MustCompile(`\{[^}]+\}`)
    matches := re.FindAllString(value, -1)
    if len(matches) > 0 {
        return errors.New("plain template should not contain placeholders")
    }
    return nil
}
```

#### named 类型

```go
func validateNamedTemplate(value string, params []LangParam) error {
    // 规则1：value 中的占位符必须在 params 中定义
    // 规则2：params 中 required=true 的参数必须出现在 value 中
    valuePlaceholders := ExtractPlaceholders(value)
    paramMap := make(map[string]LangParam)

    for _, p := range params {
        paramMap[p.Name] = p
        if p.Required {
            requiredParams[p.Name] = true
        }
    }

    // 检查未定义的占位符
    for _, ph := range valuePlaceholders {
        if _, exists := paramMap[ph]; !exists {
            return errors.New("placeholder " + ph + " not defined in params schema")
        }
    }

    // 检查缺失的必需参数
    for reqParam := range requiredParams {
        if !contains(valuePlaceholders, reqParam) {
            return errors.New("required param " + reqParam + " missing in value")
        }
    }
    return nil
}
```

#### icu 类型

基础校验：如果有占位符则必须定义 params。

### 多语言一致性校验

`ValidateCrossLanguagePlaceholders()` 确保同一 key 的不同语言版本占位符集合一致。

```go
// 同一 key 的 zh-CN 和 en-US 必须使用相同的占位符
entries := []LangStringEntry{
    {Key: "greeting", Value: "欢迎 {name}", LangCode: "zh-CN"},
    {Key: "greeting", Value: "Welcome {name}", LangCode: "en-US"},
}
err := ValidateCrossLanguagePlaceholders(entries) // nil → 通过
```

## compat_filter 兼容性

`ApplyCompatFilter()` 按客户端版本过滤模板类型，保证老客户端不会因新模板类型崩溃。

### 过滤规则

| 客户端版本 | 返回的模板类型 |
|---|---|
| >= 2.0.0 | plain + named + icu |
| < 2.0.0 | 仅 plain |

### 版本比较算法

```go
func compareVersions(v1, v2 string) int {
    // 简单的分段数值比较
    // 返回值：-1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
}
```

**设计理由**: named 和 icu 模板需要客户端支持渲染引擎，老版本不具备此能力。

## 核心实体

### LangPack

对应 `sys_lang_pack` 表，表示一个语言包的元信息。

```go
type LangPack struct {
    ID          int64
    PackName    string       // 包名
    Env         string       // 环境
    Version     int          // 包版本号
    LangCode    string       // 语言代码（zh-CN/en-US）
    Description string       // 描述
    IsPublished bool         // 是否已发布
    PublishedAt *time.Time   // 发布时间
}
```

**关键方法**: `IsPublishedStatus()` 判断是否可对外提供服务。

### LangString

对应 `sys_lang_string` 表，表示一个语言包中的字符串条目。

```go
type LangString struct {
    ID            int64
    PackID        int64
    StringKey     StringKey           // 字符串键
    StringValue   string              // 模板值
    GroupName     string              // 分组名
    Version       int                 // 条目版本
    OperationType OperationType       // 操作类型（add/update/delete）
    PrevValue     *string             // 更新前的值（用于 diff）
    TemplateType  TemplateType        // 模板类型
    ParamsSchema  string              // 参数 schema（JSON）
    PreviewSample string              // 预览示例
}
```

**关键方法**:
- `IsPlain()`: 判断是否为纯文本模板
- `IsNamed()`: 判断是否为命名参数模板
- `IsIcu()`: 判断是否为 ICU 模板
- `GetParams()`: 解析并返回参数列表

### LangStringEntry

业务视图，用于 API 响应。

```go
type LangStringEntry struct {
    Key          string
    Value        string
    TemplateType string           // plain/named/icu
    Params       []LangParamEntry // 参数描述列表
    OperationType string          // add/update/delete
}

type LangParamEntry struct {
    Name     string
    Type     string
    Required bool
    DefaultV string
}
```

### TemplateType

```go
type TemplateType string

const (
    TemplatePlain TemplateType = "plain"   // 纯文本
    TemplateNamed TemplateType = "named"   // 命名参数
    TemplateIcu   TemplateType = "icu"     // ICU（MVP 未实现）
)
```

### LangParam

参数描述，定义单个占位符的元信息。

```go
type LangParam struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Required bool   `json:"required"`
    DefaultV string `json:"default_v,omitempty"`
}
```

## API 接口

### I18nService 接口

```go
type I18nService interface {
    // 获取支持的语言列表（元数据）
    GetLanguages(clientVersion string) ([]LanguageMeta, error)

    // 获取全量语言包（首次加载或缓存失效时使用）
    GetLangPack(langCode string, clientVersion string, env string) (*LangPackResponse, error)

    // 获取增量语言包（定期增量同步）
    GetLangDifference(langCode string, sinceVersion int64, clientVersion string, env string) (*LangDiffResponse, error)

    // 校验模板一致性（保存前必须通过的质量门）
    ValidateTemplate(value string, templateType domain.TemplateType, params []domain.LangParam) error
}
```

### LanguageMeta

```go
type LanguageMeta struct {
    Code       string  // 语言代码（zh-CN）
    Name       string  // 英文名称（Chinese Simplified）
    NativeName string  // 本地名称（简体中文）
    IsDefault  bool    // 是否默认语言
}
```

### LangPackResponse

```go
type LangPackResponse struct {
    PackVersion int64              // 包版本号
    Strings     []LangStringEntry  // 字符串条目列表
}
```

### LangDiffResponse

```go
type LangDiffResponse struct {
    CurrentVersion int64              // 当前最新版本
    Additions      []LangStringEntry  // 新增/修改的条目
    Deletions      []string           // 删除的 key 列表
}
```

## 种子数据

### zh-CN（简体中文）

| Key | Value | TemplateType | Params |
|---|---|---|---|
| app.name | CaiRobot 智能学习助手 | plain | - |
| greeting.welcome | 欢迎 {userName}，今天有 {taskCount} 个任务待完成 | named | userName(string,required), taskCount(int,required) |
| common.confirm | 确定 | plain | - |

### en（English）

| Key | Value | TemplateType | Params |
|---|---|---|---|
| app.name | CaiRobot Smart Learning Assistant | plain | - |
| greeting.welcome | Welcome {userName}, you have {taskCount} tasks today | named | userName(string,required), taskCount(int,required) |
| common.confirm | Confirm | plain | - |

## 相关文档

- PRD: [PRD-01-服务商后台系统.md](../prd/PRD-01-服务商后台系统.md)
- ADR: [ADR-0012-polyglot-monorepo-directory-layout.md](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- Proto: [i18n.proto](../../../proto/i18n.proto)
- 协议编号: 6005-6010（见 [协议编号注册表.md](../api/协议编号注册表.md)）
