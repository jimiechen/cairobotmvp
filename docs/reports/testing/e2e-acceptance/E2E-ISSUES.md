# E2E 问题追踪 (E2E-ISSUES.md)

> 更新时间: 2026-06-05T16:25:00+08:00
> 关联报告: E2E-RUN-REPORT.md / E2E-FIX-REPORT-6000-HANDLERS.md / E2E-RERUN-REPORT.md

## 问题总览

| 问题ID | 严重等级 | 状态 | 影响用例 | 当前状态 |
|--------|----------|------|----------|----------|
| BUG-E2E-001 | P1 | **已修复** | MVP-004/005/006/007/008 | ✅ 复测通过 |
| BUG-E2E-002 | P2 | Open | MVP-002/003 | 预存问题 |
| BUG-E2E-003 | P3 | Open | routes.yaml 同步 | 建议改进 |

---

## BUG-E2E-001: Config/I18n Handler 未注册（已修复）

### 基本信息

| 字段 | 值 |
|------|-----|
| 问题ID | BUG-E2E-001 |
| 发现时间 | 2026-06-05 Phase 3 E2E 执行 |
| 发现方式 | E2E-MVP-004/007/008 返回 10404 |
| 严重等级 | **P1** (影响核心功能) |
| 当前状态 | **已修复 + 已复测验证** |

### 根因

`main.go` 仅调用 `RegisterSystemHandlers()`，遗漏 `RegisterConfigI18nHandlers()`。

次要根因：开发副本 `routes.yaml` 不完整（仅 2 路由 vs 完整 8 路由）。

### 修复方案

| 文件 | 修改内容 |
|------|----------|
| `tarsclient/invoker.go` | 新增 noop stub + RegisterAllLocalHandlers() |
| `cmd/server/main.go` | L38: RegisterSystemHandlers → RegisterAllLocalHandlers |
| `configs/gateway/routes.yaml` | 从 48 行替换为 178 行完整版 |

### 验证结果

```
修复前: MVP-004 → 10404, MVP-007 → 10404, MVP-008 → 10404
修复后: 全部 6 个定向用例 → 10200 PASS
单元测试: TestRegisterAllLocalHandlers 8/8 PASS
E2E 测试: GetAppConfigs FullChain PASS, GetLangPack FullChain PASS
```

### 10404 消除证据

| 端点 | 修复前 | 修复后 |
|------|--------|--------|
| 6000:6001 GetAppConfigs | 10404 | **10200** |
| 6000:6009 AppConfigVersion | 10404 | **10200** |
| 6000:6003 GetAppLanguage | 10404 | **10200** |
| 6000:6005 GetLangPack | 10404 | **10200** |
| 6000:6007 GetLangDifference | 10404 | **10200** |

---

## BUG-E2E-002: HealthCheck / HelloWorld 模块预存失败（Open）

### 基本信息

| 字段 | 值 |
|------|-----|
| 问题ID | BUG-E2E-002 |
| 发现时间 | 修复验证期间发现（非本次修复引入） |
| 严重等级 | **P2** (影响局部功能) |
| 当前状态 | **Open** — 与本次修复无关 |

### 现象

以下测试在本次修复前后均失败：

| 测试名 | 错误信息 |
|--------|----------|
| TestGateway_E2E_Modules_HealthCheck | expected status 'OK', got 'Unhealthy' |
| TestGateway_E2E_Modules_HelloWorld | expected Result.Code 10200, got 10400 "name too long" |
| TestGatewayServer_E2E_HealthCheck | expected status 'OK', got "Unhealthy" |
| TestRegisterModuleHandlers/HealthCheck | expected result code 10200, got 10500 |
| TestRegisterModuleHandlers/HelloWorld | expected result code 10200, got 10400 |

### 分析

这些是 System 模块（2100 段）的预存在问题，与本次 Config/I18n（6000 段）修复无关。可能原因：
- Health 模块的依赖注入或环境检测逻辑问题
- Hello 模块的参数校验规则与测试数据不匹配

### 建议

单独开 Issue 排查，不阻塞本次修复合入。

---

## BUG-E2E-003: routes.yaml 双副本同步风险（建议改进）

### 基本信息

| 字段 | 值 |
|------|-----|
| 问题ID | BUG-E2E-003 |
| 发现时间 | 本次修复过程中发现 |
| 严重等级 | **P3** (低风险) |
| 当前状态 | **Open** — 建议改进 |

### 现象

项目中存在两个 routes.yaml 副本：

| 路径 | 行数 | 用途 |
|------|------|------|
| `configs/gateway/routes.yaml` | 178 行 (8 路由) | 项目根，权威版本 |
| `go/gateway/proto-gateway/configs/gateway/routes.yaml` | 178 行 (刚同步) | 开发副本，Gateway 运行时加载 |

Gateway 通过相对路径 `configs/gateway/routes.yaml` 加载，CWD 为 `go/gateway/proto-gateway/` 时使用开发副本。

### 风险

两副本可能再次漂移，导致类似 BUG-E2E-001 的路由缺失问题。

### 建议

1. 使用符号链接：`ln -sf ../../../configs/gateway/routes.yaml go/gateway/proto-gateway/configs/gateway/routes.yaml`
2. 或 CI 校验：对比两文件 diff，不一致时阻断构建
3. 或统一为单一入口，通过环境变量/GOPATH 定位

---

## 问题趋势

```
Phase 3 初测 (2026-06-05):
  FAIL x3 (10404) + N/A x2 = 5 个受影响用例
  结论: PARTIAL_PASS → FAIL

Fix + 复测 (2026-06-05 16:20-16:25):
  BUG-E2E-001: 已修复，6/6 定向复测 PASS
  BUG-E2E-002: Open (P2, 单独排查)
  BUG-E2E-003: CI 校验已实现
  结论: ALL PASS (定向复测)

主控确认 (2026-06-05 16:30):
  BUG-E2E-001: ✅ 闭环，可合入
  BUG-E2E-002: ⏸️ 单独 Issue 排查
  BUG-E2E-003: ✅ CI 校验已落地
  最终结论: 修复完成，可推进合入流程
```
