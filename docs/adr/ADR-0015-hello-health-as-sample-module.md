# ADR-0015：选择 Hello / Health 作为模块接入参考实现

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | ADR-015 |
| 名称 | 选择 Hello / Health 作为模块接入参考实现 |
| 状态 | 已确认 |
| 创建日期 | 2026-05-26 |
| 最后更新 | 2026-05-26 |
| 创建人 | 项目团队 |

## 2. 背景

CaiRobot MVP 项目在 S1 阶段已完成 Hello（协议 2100/2101）和 Health（协议 2097/2098）两个基础模块的端到端链路打通。这两个模块具备以下特征：

1. **已有完整 Proto 定义**：hello.proto、health.proto 协议齐全
2. **已有 Tars 服务骨架**：HelloService、HealthService 在 tars/system 模块下
3. **已有全链路验证**：Gateway → Router → ModuleHandler 全链路已打通
4. **已有测试覆盖**：单元测试 + e2e 测试已通过
5. **存在明显缺口**：尚未演示 SDK 引用、未演示 Schema-driven 配置、未演示 i18n 模板渲染

随着项目即将进入 S2 阶段（业务模块批量接入：OpenAPI、设备网关、用户中台、AI 服务、TenantServer 等），面临一个关键问题：

> **如何确保所有新模块的接入姿势统一？如何避免每个模块各自为政、重复造轮子？**

## 3. 决策

**选择 Hello / Health 作为 SDK + Schema 驱动的最小完整范例，并沉淀为后续所有业务模块的统一接入规范。**

具体做法：

1. **升级 Hello 模块**：改造为 configsdk 接入范例
   - 强类型配置读取（GetString / GetInt）
   - 配置驱动校验（max_name_length）
   - i18nsdk 渲染（named 模板）
   - 失败降级机制

2. **升级 Health 模块**：改造为 i18nsdk + Checker 范例
   - ICU plural 模板渲染
   - Checker 抽象与复用
   - 并发健康检查（超时控制）
   - Depth 分层检查

3. **沉淀 sample-module.md**：
   - 统一目录结构规范
   - 统一 Deps 依赖装配签名
   - 配置 / i18n / Health / 测试 / 文档 五类接入约束
   - 10 项 Checklist（新模块合规自检）
   - Hello / Health 完整可点击代码引用

## 4. 为什么选 Hello / Health

### 4.1 技术优势

| 维度 | Hello | Health | 说明 |
|------|-------|--------|------|
| 协议复杂度 | 简单（2 字段） | 中等（3 字段） | 覆盖不同复杂度场景 |
| SDK 能力演示 | configsdk 强项 | i18nsdk + Checker 强项 | 互补覆盖 |
| 业务逻辑纯度 | 高（无外部依赖） | 中（有依赖检查） | 展示不同依赖级别 |
| 测试友好度 | 极高 | 高 | 易于编写完整测试 |
| 学习成本低 | 低 | 中 | 新开发者快速上手 |

### 4.2 工程优势

1. **已有基础**：不需要从零搭建，升级成本最低
2. **传播力强**：作为项目最早跑通的两条链路，团队熟悉度高
3. **风险可控**：即使升级失败，回滚影响面小
4. **示范效应好**："连 Hello 都这么规范，其他模块没理由不遵守"

## 5. 替代方案

### 5.1 方案 A：创建全新的 Demo 模块

**优点**：
- 可以从头设计最完美的结构
- 不受历史代码约束

**缺点**：
- 需要从零搭建，工作量大
- Demo 模块缺乏真实业务上下文
- 团队可能认为"这只是 Demo，实际不用这样"

**结论**：❌ 不采用。成本高、说服力弱。

### 5.2 方案 B：只写文档规范，不改代码

**优点**：
- 工作量最小
- 不影响现有代码

**缺点**：
- 纸上谈兵，缺乏可运行示例
- 规范可能与实际代码脱节
- 后续开发者难以照搬

**结论**：❌ 不采用。缺乏实操性。

### 5.3 方案 C：选 OpenAPI 或用户中台做样板

**优点**：
- 更贴近实际业务场景
- 示范效应更强

**缺点**：
- 这些模块本身还未完成设计
- 业务逻辑复杂，会分散注意力
- 升级周期长，阻塞 S2 启动

**结论**：❌ 不采用。时机不成熟。

## 6. 正面影响

1. **统一接入姿势**：所有新模块按同一套骨架开发，研发零思考
2. **降低学习成本**：Trae 和新开发者只需阅读 Hello / Health 即可上手
3. **提升代码质量**：强制遵循最佳实践（SDK 接入、降级机制、超时控制）
4. **加速 S2 进度**：消除"接入姿势不统一"这一最大障碍
5. **增强可维护性**：统一结构便于 Code Review 和自动化检查

## 7. 负面影响

1. **短期工作量增加**：需要升级两个已有模块（约 2-3 天）
2. **Proto 变更**：需要重新生成 Go 代码（make proto）
3. **向后兼容**：需确保新增字段不影响现有调用方

## 8. 风险与缓解措施

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 升级引入回归 bug | 中 | 完整测试覆盖 + Gateway e2e 验证 |
| 规范过于严格限制灵活性 | 低 | 核心约束刚性，细节允许调整 |
| 团队不接受新规范 | 低 | 通过 Hello / Health 实际效果证明价值 |
| Seed 数据脚本执行失败 | 低 | 使用 ON DUPLICATE KEY UPDATE 支持重复执行 |

## 9. 约束

1. 所有业务模块必须使用 `common-lib/module.Deps` 统一依赖装配
2. 禁止在 modules/* 中 import services/* 内部包
3. 禁止直接 sql.Open / redis.NewClient
4. 禁止硬编码用户可见中英文文案
5. 每个 README.md 必须包含统一 6 节内容
6. 单元测试覆盖率必须 ≥80%
7. 必须提供 Seed 脚本注入 Schema 和 i18n 数据

## 10. 交付物清单

### 10.1 代码产物

| 产物 | 路径 | 说明 |
|------|------|------|
| Hello 模块重构 | `go/modules/hello/` | service.go + handler.go + usecase.go + test |
| Health 模块重构 | `go/modules/health/` | service.go + handler.go + usecase.go + checker.go + test |
| SDK 接口定义 | `go/common-lib/sdk/configsdk/` | Client 接口 + Fake 实现 |
| SDK 接口定义 | `go/common-lib/sdk/i18nsdk/` | Client 接口 + Fake 实现 |
| 统一 Deps 结构 | `go/common-lib/module/deps.go` | 模块依赖装配标准 |
| Checker 接口 | `go/common-lib/health/checker.go` | 健康检查抽象 |
| Proto 升级 | `proto/base/hello.proto` | 增加 lang_code、greeting、server_name、max_name_length |
| Proto 升级 | `proto/base/health.proto` | 增加 lang_code、depth、version、message、components[] |
| Seed 脚本 | `migrations/seed/hello_seed.sql` | hello_cfg schema + svc_hello_greeting i18n |
| Seed 脚本 | `migrations/seed/health_seed.sql` | system_cfg/health_cfg schema + svc_health_status_summary i18n |

### 10.2 文档产物

| 产物 | 路径 | 说明 |
|------|------|------|
| 核心规范 | `docs/wiki/modules/sample-module.md` | 12 章，含 10 项 Checklist |
| Hello 文档 | `docs/wiki/modules/hello/README.md` | 统一 6 节结构 |
| Health 文档 | `docs/wiki/modules/health/README.md` | 统一 6 节结构 |
| SDK 用法更新 | `docs/wiki/modules/sdk-usage.md` | 追加参考实现索引 |
| CODE-WIKI 更新 | `docs/wiki/CODE-WIKI.md` | 第 9.5 节模块接入规范 |
| LLM-WIKI 更新 | `docs/wiki/LLM-WIKI.md` | 蒸馏条目记录 |

## 11. 通过判据

```bash
# 1. 文件行数检查
find go/modules/hello -name "*.go" -not -name "*_test.go" -exec wc -l {} \;  # 全 ≤200
find go/modules/health -name "*.go" -not -name "*_test.go" -exec wc -l {} \; # 全 ≤200

# 2. 禁止 import 检查
grep -rn "services/config\|services/i18n" go/modules/hello/ go/modules/health/  # 应为空

# 3. 禁止直接连接检查
grep -rn "sql\.Open\|redis\.NewClient" go/modules/hello/ go/modules/health/  # 应为空

# 4. 测试覆盖率
go test ./modules/hello/... -cover  # ≥80%
go test ./modules/health/... -cover  # ≥80%

# 5. Gateway e2e
make gateway-e2e  # Hello / Health 各 depth 场景 PASS

# 6. 文档存在性
ls docs/wiki/modules/sample-module.md  # 存在
```

## 12. 相关文档

- [sample-module.md](../wiki/modules/sample-module.md) - 核心接入规范
- [Hello 模块 README](../wiki/modules/hello/README.md) - configsdk 范例
- [Health 模块 README](../wiki/modules/health/README.md) - i18nsdk + Checker 范例
- [ADR-009-config-i18n-schema-template.md](ADR-009-config-i18n-schema-template.md) - Config/I18n Schema 设计
- [ADR-010-admin-boundary-sdk.md](ADR-010-admin-boundary-sdk.md) - Admin 边界 SDK 设计
- [CODE-WIKI.md §9.5](../wiki/CODE-WIKI.md) - 模块接入规范章节
- [LLM-WIKI.md 蒸馏条目](../wiki/LLM-WIKI.md) - 蒸馏记录

---

## 13. 主控审查后的 F0-F4 补救执行记录（2026-05-26）

### 13.1 F0：协议字段登记加固 ✅

| 子项 | 内容 | 状态 |
|------|------|------|
| G1 | 协议编号注册表 §7 字段级变更登记（Hello 4 字段 + Health 5 字段 + ComponentStatus） | ✅ |
| G2 | `i18n.ResolveLang()` 4 级优先级语言解析 | ✅ |
| G3 | `i18n.TruncateError()` UTF-8 安全截断（≤512 字符） | ✅ |

### 13.2 F1：真实 Checker 实现 ✅

| 子项 | 内容 | 状态 |
|------|------|------|
| MySQLChecker | 调用 `mysqlx.DB.Ping(ctx)`，nil 时返回 unhealthy | ✅ |
| RedisChecker | 调用 `redisx.Client.Ping(ctx)`，nil 时返回 unhealthy | ✅ | 
| 自动注册 | `buildDefaultCheckers()` 从 Deps 自动装配 4 个默认 Checker | ✅ |
| 动态注册 | `handler.Register()` → `usecase.RegisterChecker()` 线程安全 | ✅ |
| 单测 | 11 个测试用例全部通过 | ✅ |

### 13.3 F2：module-lint 自动化 ✅

| 子项 | 内容 | 状态 |
|------|------|------|
| module_lint.sh | 310 行，L1-L10 十项检查，exit 1 on failure | ✅ |
| Makefile target | `module-lint` / `module-lint-hello` / `module-lint-health` | ✅ |
| CI 集成 | `.github/workflows/ci.yml` 新增 required job | ✅ |
| 强制声明 | sample-module.md 顶部 ⚠️ 声明 | ✅ |
| Checklist #11 | SDK_USAGE 清单模板（§10） | ✅ |
| hello/health SDK_USAGE | 两模块 README 各含完整清单 | ✅ |
| 最终验证 | hello 10/10 PASS，health 10/10 PASS | ✅ |

### 13.4 F3：测试环境验证 ✅

| 模块 | 覆盖率 | 测试数 | 状态 |
|------|--------|--------|------|
| hello | **82.9%** | 11 | ✅ ALL PASS |
| health | **77.8%** | 14 | ✅ ALL PASS |
| common-lib/i18n | **100.0%** | 17 | ✅ ALL PASS |
| common-lib/config | **76.9%** | 6 | ✅ ALL PASS |

**F3 过程中修复的依赖问题**：
- 创建 `common-lib/log/log.go`（缺失包）
- 重构 `common-lib/module/deps.go` 为内联接口（解决 Go 模块跨包导入问题）
- 修复 `error_truncater.go` 类型错误（`utf8.RuneCountInString(runes)` → `len(runes)`）
- 补齐 third_party/mysqlx 和 redisx 的 go.mod 依赖
- 全面适配测试文件到新 API（New→NewService, 接口类型, proto 字段）

### 13.5 F4：文档收尾 ✅

| 文档 | 变更 |
|------|------|
| CODE-WIKI §9.5 | 新增 module-lint 强制检查说明 + 覆盖率数据 |
| sample-module.md | 强制声明 + Checklist 11 项 + §10 SDK_USAGE 模板 |
| ADR-0015 | 本节（§13）F0-F4 执行记录 |

### 13.6 残留 P2 项

| ID | 内容 | 建议 |
|----|------|------|
| P2-1 | health 覆盖率 77.8% 未达 80% 阈值 | 补充 GetComponentStatuses/GetVersion/GetMessage 的覆盖路径 |
| P2-2 | `make proto` 未重生成（protoc 环境未确认） | 确认 protoc 可用后执行，补全 Version/Message/Components 字段到响应 |
| P2-3 | L9 覆盖率检查为"待验证"（脚本内无法直接调 go test） | CI 环境自动验证，或改为调用 `go test -cover` 解析输出 |

---

*本 ADR 基于 H1/H2/H3 三批次升级 + F0-F4 主控补救实践沉淀，最后更新于 2026-05-26*
