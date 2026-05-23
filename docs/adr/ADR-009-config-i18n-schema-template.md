# ADR-009: Config Schema Registry 与 i18n 参数化模板架构决策

## 状态

已采纳

## 背景

CaiRobot MVP 需要支持运营团队自助新增配置字段和多语言 key，而不需要修改 Go 代码并重新编译部署。传统做法是将配置字段硬编码在 Go 结构体中或 Protobuf 定义中，每次新增字段都需要：

1. 修改 Go 结构体/Protobuf 定义
2. 重新编译并部署服务
3. 客户端同步更新

这种模式无法满足运营敏捷性要求，特别是在 MVP 阶段需要快速迭代配置项。

## 决策

### 1. 引入 sys_config_schema 元数据注册表

运营通过 `sys_config_schema` 表定义配置字段的元数据，包括：

- 字段名（field_key）
- 字段类型（field_type：string/int/bool/json/array）
- 默认值（default_value）
- 校验规则（validation_rule）
- 分组（module_name）
- 排序权重（sort_order）

新模块通过 `dynamic_modules` 下发，客户端根据 schema 动态渲染配置 UI。

### 2. 引入 DynamicConfigModule 自描述容器

```go
type DynamicConfigModule struct {
    ModuleName    string                 `json:"module_name"`
    DisplayName   string                 `json:"display_name"`
    SchemaVersion int                    `json:"schema_version"`
    Fields        []ConfigFieldSchema    `json:"fields"`
}

type ConfigFieldSchema struct {
    FieldKey       string          `json:"field_key"`
    FieldType      string          `json:"field_type"`      // string/int/bool/json/array
    DefaultValue   interface{}     `json:"default_value"`
    ValidationRule string          `json:"validation_rule"`
    SortOrder      int             `json:"sort_order"`
}
```

### 3. 多语言参数化模板架构

多语言支持三种模板类型：

| template_type | 说明 | 示例 |
|---|---|---|
| plain | 纯文本，无参数 | `"欢迎回来"` |
| named | 命名参数占位符 | `"你好，{name}！剩余{count}次机会"` |
| icu | ICU MessageFormat | `"你好，{name, select, male {先生} female {女士} other {}}，您有{count, plural, =0 {没有新消息} one {#条新消息} other {#条新消息}}"` |

每个翻译 key 关联 `params_schema` 参数描述：

```go
type I18nTemplate struct {
    Key           string            `json:"key"`
    TemplateType  string            `json:"template_type"`   // plain/named/icu
    ParamsSchema  []ParamSchema     `json:"params_schema"`
    Translations  map[string]string `json:"translations"`    // locale → template
}

type ParamSchema struct {
    ParamName  string `json:"param_name"`
    ParamType  string `json:"param_type"`   // string/int/float/date
    Required   bool   `json:"required"`
    Example    string `json:"example"`
}
```

### 4. 兼容策略

- **老客户端**只收到强类型字段和 plain 模板，不包含 dynamic_modules 和 params_schema
- **新客户端**通过 clientVersion 或能力协商获取完整 schema 信息
- Gateway 根据 clientVersion 过滤响应字段

## 后果

### 正面影响

1. **运营可自助扩展**：新增配置字段或多语言 key 无需改代码、无需部署
2. **零代码部署**：schema 变更通过 Admin 后台操作即可生效
3. **类型安全**：schema 定义了字段类型和校验规则，可自动生成校验逻辑
4. **文档自描述**：schema 本身就是 API 文档，可用于生成 OpenAPI 规范

### 负面影响

1. **schema 变更需要同步到客户端**：新增字段类型后，老版本客户端无法识别
2. **运行时校验开销**：动态 schema 需要在运行时解析和校验，比编译时类型检查略慢
3. **调试复杂度增加**：配置错误可能在运行时才发现，而非编译期

## 替代方案

### 方案 A: 纯 JSON 无 schema

直接使用 JSON 存储配置值，不做元数据定义。

**否决原因**：
- 无法做字段类型校验
- 无法生成文档
- 运营人员不知道有哪些可用字段
- 容易出现拼写错误导致配置不生效

### 方案 B: Protobuf 每次改定义重新编译

每次新增配置字段都修改 .proto 文件，重新生成代码并部署。

**否决原因**：
- 发布周期长（需走完整 CI/CD 流程）
- 运营无法自助操作
- 频繁变更会导致 Protobuf 版本膨胀
- 客户端必须与服务端同步发布

## 相关文档

- [CODE-WIKI.md](../wiki/CODE-WIKI.md) §5.1 配置与多语言协议路由示例
- [ADR-010-admin-boundary-sdk.md](ADR-010-admin-boundary-sdk.md)
- 协议编号注册表 6000 段

## 参考实现

- [services/config](../../go/services/config) — Schema Registry 实现
- [services/i18n](../../go/services/i18n) — 参数化模板服务实现
- [tars/config](../../go/tars/config) — Config Tars Servant
- [tars/i18n](../../go/tars/i18n) — I18n Tars Servant
