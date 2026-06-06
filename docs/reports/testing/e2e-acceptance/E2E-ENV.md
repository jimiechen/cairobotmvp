# E2E 环境记录 (E2E-ENV)

> 生成时间: 2026-06-05T04:48:00+08:00
> 执行阶段: Phase 2 环境就绪 + Phase 3 MVP 执行

## 1. 基础环境

| 项目 | 版本/值 | 状态 |
|------|---------|------|
| OS | macOS (Darwin) | OK |
| Go | 1.25.5 | OK |
| Node.js | v22.20.0 | OK |
| tsx | 4.22.4 | OK |
| npm | 已安装 (proto-tester 依赖 531 packages) | OK |
| protoc + protoc-gen-ts | 已安装 (proto TS 生成 OK) | OK |

## 2. Gateway 服务

| 项目 | 值 | 状态 |
|------|-----|------|
| 二进制路径 | `build/proto-gateway-server` | OK |
| PID | 49549 | Running |
| 监听端口 | 8080 | OK |
| 启动模式 | local (GATEWAY_INVOKER_MODE=local) | OK |
| 路由配置 | `configs/gateway/routes.yaml` (8 routes) | OK |
| 已注册 handler | RegisterSystemHandlers only (HealthCheck + HelloWorld) | **部分** |
| 未注册 handler | RegisterConfigI18nHandlers (5 handlers) | **缺失** |

## 3. proto-tester CLI

| 项目 | 值 | 状态 |
|------|-----|------|
| 源码目录 | `typescript/proto-tester/` | OK |
| 启动命令 | `npx tsx src/cli/index.ts <subcommand>` | OK |
| 可用子命令 | send, run, trace | OK |
| Protobuf TS 生成 | `proto/generated/ts/*.ts` (已生成) | OK |
| google-protobuf | 已安装 (项目根 node_modules) | OK |
| 报告输出目录 | `typescript/proto-tester/proto-tester-reports/` | OK |

## 4. 关键发现

### 4.1 Config/I18n Handler 未注册 (P1 Bug)

**文件**: `go/gateway/proto-gateway/cmd/server/main.go` 第 38 行

**现象**: Gateway 启动时仅调用 `RegisterSystemHandlers()`，未调用 `RegisterConfigI18nHandlers()`

**影响**: 所有 6000 段协议（GetAppConfigs/GetAppLanguage/GetLangPack 等）返回 10404 NOT_FOUND

**修复方案**: 在 main.go 的 local 分支中增加 `tarsclient.RegisterConfigI18nHandlers(...)` 调用，需传入 configSvc 和 i18nSvc 实例

### 4.2 Auth Channel 为 STUB

**现象**: Gateway 不读取 Authorization Header，不校验 extend["token"]
auth_required 字段仅写入 extend map，无实际鉴权逻辑

**影响**: MVP-005 (缺失Token) / MVP-006 (错误Token) 无法验证鉴权行为

### 4.3 trace CLI 状态 DEGRADED

**现象**: `trace list` 子命令不存在，trace 需要 `-i <traceId>` 参数
服务端 `/api/dev/trace` API 未实现

**影响**: MVP-010 traceId 贯穿只能通过代码分析确认，无法端到端验证

## 5. Artifact 目录结构

```
docs/reports/testing/e2e-acceptance/
├── E2E-PREFLIGHT.md              # Phase 0: 前置检查报告
├── E2E-PROTO-TESTER-CLI.md       # Phase 0: CLI 参考手册
├── E2E-CASE-MATRIX.md            # Phase 1: 用例矩阵 (13+14)
├── E2E-MVP-SUITE.yaml            # Phase 1: MVP 无Token 套件
├── E2E-MVP-SUITE-TOKEN.yaml      # Phase 1: MVP 有Token 套件
├── E2E-REGRESSION-SUITE.yaml     # Phase 1: 回归套件 (14 cases)
├── E2E-AUTH-CHANNEL.md           # Phase 1: 鉴权通道分析
├── E2E-ARTIFACT-SPEC.md          # Phase 1: Artifact 规范
├── E2E-ENV.md                    # Phase 2: 本文档 — 环境记录
└── E2E-RUN-REPORT.md             # Phase 3: 执行报告

typescript/proto-tester/proto-tester-reports/
├── send-2100-2097-*.md           # MVP-002 Health 报告
├── send-2100-2101-*.md           # MVP-003/010 Hello 报告 (多份)
```
