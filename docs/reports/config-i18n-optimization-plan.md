# Config + I18n 全量优化实施方案

> **版本**: v1.0  
> **日期**: 2026-05-24  
> **状态**: 待主控评审  
> **约束条件**: 单文件 ≤ 400 行，单函数 ≤ 400 行

---

## 一、执行摘要

### 1.1 当前状态快照

| 检测项 | 约束 | 当前值 | 状态 |
|--------|------|--------|------|
| 超限文件数 | ≤400 行 | **0 个** | ✅ 通过 |
| 超限函数数 | ≤400 行 | **0 个** | ✅ 通过 |
| 200-400 行文件 | 关注区 | **10 个** | ⚠️ 需优化 |
| 50-400 行函数 | 关注区 | **10 个** | ⚠️ 需关注 |

### 1.2 阻断项清单（来自验收报告）

| ID | 问题 | 严重度 | 来源 |
|----|------|--------|------|
| B-01 | compose.go 硬编码 base_cfg/pay_cfg switch | P0 | 架构红线 |
| B-02 | third_party/{mysqlx,redisx,sqlitex} 缺失 | P0 | 基础设施 |
| B-03 | Tars IDL (config.tars/i18n.tars) 缺失 | P0 | 编译依赖 |
| B-04 | ADR-008 文档缺失 | P1 | 决策落档 |
| B-05 | sdk-usage.md 缺失 | P1 | 接入文档 |

### 1.3 实施批次总览

```
R0 ✅ 检测口径核对（已完成）
R1 ⏳ 第三方基础设施补齐
R2 ⏳ 架构红线修复（compose.go 去硬编码）
R3 ⏳ 文件/函数全量拆分优化
R4 ⏳ Tars IDL + 路径纠偏 + 协议号处理
R5 ⏳ 测试覆盖率提升 + 文档补齐
终验 ⏳ 全量复测
```

---

## 二、R0 检测口径核对结果

### 2.1 函数行数疑点澄清

**原问题**：之前检测报告中出现 "redis_cache.go::Invalidate=785 行"、"value.go::Float=1292 行" 的异常数据。

**根因分析**：awk 统计脚本将**文件绝对行号**误报为**函数体行数**。

**修正后实测结果（Top 10 超过 50 行的函数）**：

| 排名 | 文件 | 函数名 | 起始行 | 结束行 | **实际行数** |
|:----:|------|--------|:------:|:------:|:----------:|
| 1 | config/sdk/remote.go | (远程调用) | 48 | 126 | **79** |
| 2 | i18n/sdk/remote.go | (远程调用) | 39 | 100 | **62** |
| 3 | i18n/sdk/remote.go | (解码) | 101 | 164 | **64** |
| 4 | i18n/repository/sqlite_repo.go | GetStringsByPackID | 77 | 139 | **63** |
| 5 | i18n/repository/sqlite_repo.go | GetPackByLangCode | 23 | 76 | **54** |
| 6 | i18n/repository/sqlite_repo.go | (字符串查询) | 140 | 202 | **63** |
| 7 | i18n/repository/sqlite_repo.go | (发布操作) | 203 | 260 | **58** |
| 8 | i18n/repository/sqlite_repo.go | (增量查询) | 261 | 314 | **54** |
| 9 | config/sdk/remote.go | (解码) | 150 | 206 | **57** |
| 10 | i18n/service/diff.go | CalculateDiff | 9 | 61 | **52**

✅ **结论**：在 ≤400 行约束下，所有函数均未超限，无需强制拆分。

### 2.2 routes.yaml 路径决策

**主控决策**：最终路径 = `configs/gateway/routes.yaml`

**当前冲突点**：

| 位置 | 引用路径 | 需要修改 |
|------|----------|----------|
| configs/go/README.md | `gateway/routes.yaml` | ✅ 无需改（相对路径） |
| deploy/tarscloud/templates/*.md | `go/gateway/proto-gateway/configs/routes.yaml` | ⚠️ 需同步 |
| go/gateway/proto-gateway/configs/gateway/routes.yaml | （副本） | ⚠️ 确认是否保留 |

### 2.3 协议号 6002/6004/6006/6008/6010 决策

**主控决策**：**保留方案（请求/响应对应）**

**理由**：
- 请求与响应协议号一一对应，便于追踪和调试
- 注册表规则第 5 条保持不变："Request 和 Response 应分别拥有独立编号"
- 6001↔6002、6003↔6004、6005↔6006、6007↔6008、6009↔6010 成对登记

**影响范围**：
- 协议编号注册表：**保留全部 10 条记录**（5 对请求/响应）
- ADR-008：记录此决策及理由

---

## 三、200-400 行文件详细清单与拆分方案

### 3.1 当前超 200 行的文件（按行数降序）

| # | 文件路径 | 当前行数 | 目标 | 优先级 |
|:-:|----------|:-------:|:----:|:------:|
| 1 | services/config/service/compose.go | 340 | ≤250 | **P0** |
| 2 | services/i18n/repository/sqlite_repo.go | 339 | ≤250 | P1 |
| 3 | services/config/sdk/remote.go | 266 | ≤250 | P1 |
| 4 | services/i18n/sdk/remote.go | 238 | ≤200 | P2 |
| 5 | services/config/repository/sqlite_repo.go | 238 | ≤200 | P2 |
| 6 | services/config/repository/mem_repo.go | 209 | ≤200 | P2 |
| 7 | services/i18n/sdk/translate.go | 207 | ≤200 | P2 |
| 8 | services/i18n/repository/mysql_repo.go | 206 | ≤200 | P2 |
| 9 | services/config/repository/mysql_repo.go | 206 | ≤200 | P2 |

---

## 四、各批次详细实施方案

### 批次 R1：第三方基础设施补齐

#### 4.1.1 目标

创建 `go/third_party/{mysqlx,redisx,sqlitex}` 三个基础设施包，为后续 SDK 完善提供依赖。

#### 4.1.2 任务清单

**mysqlx 包（MySQL 连接池封装）**

```
go/third_party/mysqlx/
├── mysqlx.go          # 连接池封装（健康检查、慢查询日志、超时）≤150 行
├── mysqlx_test.go     # 单元测试（sqlmock 或集成 build tag）
└── go.mod             # 子模块定义
```

核心接口：
```go
type DB interface {
    Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
    Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...any) *sql.Row
    BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
    Ping(ctx context.Context) error
    Close() error
}

func NewDB(cfg *Config) (DB, error)
```

**redisx 包（Redis 客户端封装）**

```
go/third_party/redisx/
├── redisx.go          # 客户端封装（健康检查、命名空间隔离）≤150 行
├── pubsub.go          # pub/sub 发布订阅封装 ≤120 行
├── redisx_test.go     # 单元测试（miniredis）
└── go.mod
```

核心接口：
```go
type Client interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Scan(ctx context.Context, prefix string) ([]string, error)
    Subscribe(ctx context.Context, channel string, handler func(msg string)) (CancelFunc, error)
    Publish(ctx context.Context, channel string, msg string) error
    Ping(ctx context.Context) error
    Close() error
}
```

**sqlitex 包（SQLite 封装）**

```
go/third_party/sqlitex/
├── sqlitex.go         # modernc.org/sqlite 封装（in-memory + 文件双模式）≤120 行
├── schema_adapter.go  # MySQL schema → SQLite schema 适配 ≤100 行
├── sqlitex_test.go
└── go.mod
```

#### 4.1.3 迁移规则

现有代码中以下直接调用需迁移到 third_party 包：

| 当前写法 | 迁移到 |
|----------|--------|
| `sql.Open("mysql", ...)` | `mysqlx.NewDB(cfg)` |
| `redis.NewClient(...)` | `redisx.NewClient(cfg)` |
| `sql.Open("sqlite", ...)` | `sqlitex.Open(dsn, mode)` |

#### 4.1.4 通过判据

```bash
# 目录存在性
ls go/third_party/{mysqlx,redisx,sqlitex}/

# 测试通过
go test ./third_party/... -v

# 无残留直接调用
grep -rn "sql\.Open\|redis\.NewClient" go/services/ go/tars/

# 文件行数合规
find go/third_party -name "*.go" -not -name "*_test.go" -exec wc -l {} \; | awk '$1>400'
# 输出必须为空
```

---

### 批次 R2：架构红线修复 - compose.go 去硬编码

#### 4.2.1 问题诊断

当前 [compose.go:109-134](go/services/config/service/compose.go#L109-L134) 存在硬编码 switch：

```go
switch moduleKey {
case "base_cfg":   // ← 硬编码
    protoRsp.BaseCfg = mapToAppBaseConfigs(typedMap)
case "wap_cfg":    // ← 硬编码
    protoRsp.WapCfg = mapToAppWapUrlConfigs(typedMap)
// ... 共 8 个 case
}
```

**违反规则**：Schema-driven 架构要求不允许 module_key 字符串字面量做分支判断。

#### 4.2.2 解决方案：注册表映射模式

**方案设计**：用 map[moduleKey]mapperFunc 替代 switch/case。

```go
// compose_static.go - 静态模块映射注册表

// StaticModuleMapper 定义静态模块 → Protobuf 映射函数签名
type StaticModuleMapper func(map[string]*domain.TypedValue) proto.Message

// staticMapperRegistry 静态模块映射注册表
var staticMapperRegistry = map[string]StaticModuleMapper{
    domain.ModuleKeyBase:  mapToAppBaseConfigs,
    domain.ModuleKeyWap:   mapToAppWapUrlConfigs,
    domain.ModuleKeyRegex: mapToAppRegexConfigs,
    domain.ModuleKeyPay:   mapToAppPayConfigs,
    domain.ModuleKeyOss:   mapToAppOssConfigs,
    domain.ModuleKeyLang:  mapToAppLanguageConfigs,
    domain.ModuleKeyMute:  mapToAppMuteConfigs,
    domain.ModuleKeyGroup: mapToAppGroupConfigs,
}

// MapStaticModulesToProtoFields 使用注册表映射（无硬编码 switch）
func MapStaticModulesToProtoFields(staticModules map[string]map[string]*domain.TypedValue) *pb.AppConfigsRsp {
    protoRsp := &pb.AppConfigsRsp{}
    
    for moduleKey, typedMap := range staticModules {
        mapper, ok := staticMapperRegistry[moduleKey]
        if !ok {
            continue // 未注册的静态模块跳过
        }
        
        msg := mapper(typedMap)
        setProtoField(protoRsp, moduleKey, msg)
    }
    
    return protoRsp
}
```

#### 4.2.3 文件拆分计划

compose.go 当前 340 行，拆分为：

| 新文件 | 职责 | 预估行数 |
|--------|------|:-------:|
| service/compose.go | 主流程入口 + ClassifyModules + ToJSONMap | ~80 行 |
| service/compose_dynamic.go | BuildDynamicModule + DynamicModuleView 组装 | ~70 行 |
| service/compose_static.go | MapStaticModulesToProtoFields + 注册表 | ~90 行 |
| service/compose_mappers.go | 8 个 mapToApp* 映射函数 | ~130 行 |
| service/compose_helpers.go | toString/toFloat64/toBool 辅助函数 | ~40 行 |

**总计**: ~410 行（分布在 5 个文件中，每个均 < 400 行 ✅）

#### 4.2.4 通过判据

```bash
# 无硬编码 case
grep -rn 'case "base_cfg"\|case "member_cfg"\|case "pay_cfg"' go/services/config/service/
# 输出必须为空

# 文件行数合规
find go/services/config/service -name "compose*.go" -exec wc -l {} \;
# 全部 ≤ 400

# 测试通过
go test ./services/config/service/... -count=1 -v
```

---

### 批次 R3：全量整改 - 文件/函数优化

#### 4.3.1 拆分优先级矩阵

| 文件 | 当前行数 | 拆分策略 | 预期产出 |
|------|:-------:|---------|----------|
| i18n/repository/sqlite_repo.go | 339 | 按 CRUD 操作拆分 | sqlite_repo.go(60) + sqlite_pack.go(90) + sqlite_string.go(100) + sqlite_release.go(89) |
| config/sdk/remote.go | 266 | 按职责拆分 | remote.go(120) + remote_invoker.go(80) + remote_decoder.go(66) |
| i18n/sdk/remote.go | 238 | 按职责拆分 | remote.go(110) + remote_decoder.go(128) |
| config/repository/sqlite_repo.go | 238 | 同上 | sqlite_repo.go(80) + sqlite_schema.go(80) + sqlite_value.go(78) |
| config/repository/mem_repo.go | 209 | 按数据域拆分 | mem_repo.go(110) + mem_schema.go(50) + mem_value.go(49) |
| i18n/sdk/translate.go | 207 | 按模板类型拆分 | translate.go(80) + render_named.go(70) + render_helpers.go(57) |
| i18n/repository/mysql_repo.go | 206 × 2 | 同 sqlite 拆法 | 各拆为 3 个文件 |

#### 4.3.2 长函数处理原则

对于 10 个超过 50 行的函数（均在 52-79 行范围内）：

**处理策略**：在 ≤400 行约束下，这些函数**无需强制拆分**。

但如果拆分文件时自然切分了函数边界，可顺势提取 helper：
- 提取 stage 函数（validate / normalize / persist / publish）
- 提取公共逻辑为小写开头的内部函数

**禁止事项**：
- ❌ 不允许通过"加注释折叠"绕过
- ❌ 不允许为了拆分而破坏内聚性
- ❌ 不允许修改对外接口签名

#### 4.3.3 通过判据

```bash
# 所有文件 ≤ 400 行
find go/services -name "*.go" -not -name "*_test.go" -exec wc -l {} \; | awk '$1>400'
# 输出必须为空

# 所有函数 ≤ 400 行（使用正确统计脚本）
# （已验证当前无超限函数）

# 测试不回归
go test ./services/... -count=1
```

---

### 批次 R4：Tars IDL 补齐 + 路径纠偏 + 协议号处理

#### 4.4.1 Tars IDL 创建

**config.tars**（对应 routes.yaml 中 ConfigServer 的两个方法）:
```tars
module CaiRobotConfigApp
{
    interface ConfigObj
    {
        int GetAppConfigs(vector<byte> req, out vector<byte> rsp);
        int AppConfigVersion(vector<byte> req, out vector<byte> rsp);
    };
};
```

**i18n.tars**（对应 routes.yaml 中 I18nServer 的三个方法）:
```tars
module CaiRobotI18nApp
{
    interface I18nObj
    {
        int GetAppLanguage(vector<byte> req, out vector<byte> rsp);
        int GetLangPack(vector<byte> req, out vector<byte> rsp);
        int GetLangDifference(vector<byte> req, out vector<byte> rsp);
    };
};
```

方法名与 `configs/gateway/routes.yaml` 中的 `tars_method` 字段一一对应：
- ConfigObj: GetAppConfigs (6001/6002), AppConfigVersion (6009/6010)
- I18nObj: GetAppLanguage (6003/6004), GetLangPack (6005/6006), GetLangDifference (6007/6008)

命名规范遵循 CODE-WIKI 第 9 章 Tars 规范。

#### 4.4.2 routes.yaml 路径统一

**决策**：以 `configs/gateway/routes.yaml` 为唯一权威源

需要同步的位置：
- deploy/tarscloud/templates/ 中的引用文档
- go/gateway/proto-gateway/ 下的副本（确认是否删除或作为 symlink）

#### 4.4.3 协议号确认

**决策**：保留全部 10 条协议编号记录（5 对请求/响应）

当前注册表状态（无需变更）：

| max | min | 方向 | Message | 状态 |
|-----|-----|------|---------|------|
| 6000 | 6001 | C→S | AppConfigsReq | 保留 |
| 6000 | 6002 | S→C | AppConfigsRsp | **保留** |
| 6000 | 6003 | C→S | AppFetchLanguageReq | 保留 |
| 6000 | 6004 | S→C | AppFetchLanguageRsp | **保留** |
| 6000 | 6005 | C→S | AppFetchLangPackReq | 保留 |
| 6000 | 6006 | S→C | AppFetchLangPackRsp | **保留** |
| 6000 | 6007 | C→S | AppFetchLangDifferenceReq | 保留 |
| 6000 | 6008 | S→C | AppFetchLangDifferenceRsp | **保留** |
| 6000 | 6009 | C→S | AppConfigVersionReq | 保留 |
| 6000 | 6010 | S→C | AppConfigVersionRsp | **保留** |

注册表规则第 5 条保持不变："Request 和 Response 应分别拥有独立编号"

#### 4.4.4 通过判据

```bash
# IDL 文件存在
ls tars/protocol/tars/{config,i18n}.tars

# 路径引用唯一
grep -rn "routes\.yaml" --include="*.go" --include="*.yaml" .
# 应只有一种路径格式

# 协议号完整保留（10 条记录）
grep -E "600[1-10]" docs/api/协议编号注册表.md | wc -l
# 输出应为 10（5 对请求/响应）
```

---

### 批次 R5：测试覆盖率提升 + 文档补齐

#### 4.5.1 测试覆盖率现状与目标

| 包 | 当前覆盖率 | 目标 | 差距 | 重点补充用例 |
|---|:---------:|:---:|:---:|-------------|
| config/domain | 85.5% | ≥80% | ✅ 已达标 | - |
| i18n/domain | 89.7% | ≥80% | ✅ 已达标 | - |
| i18n/service | 89.7% | ≥80% | ✅ 已达标 | - |
| i18n/sdk | 80.0% | ≥80% | ✅ 已达标 | - |
| config/sdk | 72.2% | ≥80% | -7.8% | Watch 回调、pub/sub 失效场景 |
| config/service | 53.1% | ≥80% | -26.9% | compose 各分支、过滤逻辑错误路径 |
| config/cache | 37.8% | ≥80% | -42.2% | Redis 未命中/TTL/pub/sub 失效/LRU 边界 |
| i18n/cache | 39.7% | ≥80% | -40.3% | 同上 |
| config/repository | 38.3% | ≥80% | -41.7% | 并发安全、版本自增、schema CRUD 错误路径 |
| **i18n/repository** | **8.5%** | **≥80%** | **-71.5%** | **P0 重点** |

#### 4.5.2 i18n/repository 补测重点（P0）

这是最危险的低覆盖率包，需重点补充：

**正常路径**：
- GetPackByLangCode 命中/未命中
- GetStringsByPackID 正常返回
- SavePack / UpdatePack 成功

**边界条件**：
- 空 result set（返回 nil, nil）
- 超长字符串字段
- 并发写入同一 pack

**错误路径**：
- DB 连接断开
- JSON 解析失败
- SQL 语法错误（mock 触发）

**契约一致性**：
- SQLite 实现与 MySQL 实现对相同输入应返回相同输出
- 用 table-driven test 验证

#### 4.5.3 文档补齐

**ADR-008-config-i18n-migration.md**:

```markdown
# ADR-008: 全局配置与多语言迁移

## Status: Accepted
## Date: 2026-05-24

## Context
旧项目存在三大问题：
1. 双重存储（config_query.go 直接 SQL + sys_config_version）
2. 单文件超限（历史遗留 563 行文件）
3. 缓存硬编码版本号

## Decision
采用 Schema Registry + SDK + admin 边界三层架构...

## Consequences
### Positive
- 单一数据源，消除双重读路径
- Schema-driven，新增配置零代码变更

### Negative
- 强类型字段兼容层增加了复杂度
- 迁移期需维护镜像写入

### 协议号决策
6002/6004/6006/6008/6010 **保留**，
请求与响应协议号一一对应（6001↔6002、6003↔6004...），
便于追踪调试，注册表规则第 5 条保持不变。
```

**sdk-usage.md**:

```markdown
# SDK 使用规范

## 引入方式（使用完整模块路径）

### Config SDK
```go
import configsdk "github.com/jimiechen/mineplanet/go/services/config/sdk"
```

### I18N SDK
```go
import i18nsdk "github.com/jimiechen/mineplanet/go/services/i18n/sdk"
```

## InProcess vs Remote 选择
- 开发/测试环境：ModeInProcess
- 生产环境：ModeRemote

## 业务接入示例
### OpenAPI 服务
### 设备网关服务
### 用户中台服务

## 禁止事项
❌ 不得 import services/config/service 或 repository（必须只通过 sdk 包访问）
❌ 不得调用 admin-server API（铁律，业务服务 0 引用）
❌ 不得使用相对路径导入（必须使用上述完整 module path）
```

#### 4.5.4 通过判据

```bash
# 覆盖率达标
go test ./services/... -coverprofile=cover.out
go tool cover -func=cover.out | awk '{if($3+0<80 && $1!="total:") print}'
# 输出必须为空

# 文档存在
ls docs/wiki/adr/ADR-008*.md docs/wiki/modules/sdk-usage.md
```

---

## 五、全程禁止事项

1. ❌ 不允许 git push
2. ❌ 不允许新增 admin-server 依赖（铁律）
3. ❌ 不允许破坏 6001/6003/6005/6007 wire 兼容
4. ❌ 不允许跳过任何一批，必须 R0→R5 串行执行
5. ❌ 不允许擅自决策 R0 中的 3 个疑点（已由主控决策）
6. ❌ 不允许通过"屏蔽 lint 规则"或"加注释"绕过行数限制
7. ❌ 不允许在 compose.go 用任何 module_key 字符串字面量做分支判断
8. ❌ 不允许扩大改动范围到本指令未列项
9. ❌ 单文件/单函数不得超过 400 行

---

## 六、风险预估与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|:----:|:----:|---------|
| R2 反射方案性能回退 | 中 | 低 | benchmark 对比，必要时缓存 reflect.Type |
| R3 拆分导致 import 循环 | 低 | 高 | 提前画依赖图，确保单向依赖 |
| R5 测试补写引入 mock 过多 | 中 | 中 | 优先用真实 SQLite（in-memory），减少 mock |
| R4 协议号删除影响已有客户端 | 低 | 高 | 保留 reserved 标记，不立即复用编号 |

---

## 七、终验 checklist

完成 R0~R5 后，逐项确认：

- [ ] **模块 1**：所有文件 ≤ 400 行，所有函数 ≤ 400 行
- [ ] **模块 4**：`grep -rn 'case "xxx_cfg"' go/services/config/service/` 输出为空
- [ ] **模块 5**：SDK 接口完整，三层缓存实现到位
- [ ] **模块 6**：admin 边界铁律仍然成立（grep 为空）
- [ ] **模块 8**：所有包覆盖率 ≥ 80%
- [ ] **模块 9**：ADR-008/009/010 + sdk-usage.md 全部齐备
- [ ] **模块 10**：禁止事项全部 NO

---

**文档编制**: Trae AI Assistant  
**提交时间**: 2026-05-24  
**等待**: 主控评审意见
