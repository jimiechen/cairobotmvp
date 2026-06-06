# proto-tester CLI 参考手册

> **版本**：proto-tester v2.0.0
> **生成时间**：2026-06-05
> **来源**：Phase 0 Preflight 静态分析
> **源码位置**：`typescript/proto-tester/src/cli/commands/`
> **配套文档**：[E2E-PREFLIGHT.md](./E2E-PREFLIGHT.md)

---

## 目录

1. [概览](#1-概览)
2. [全局约定](#2-全局约定)
3. [send - 发送单次请求](#3-send---发送单次请求)
4. [run - 批量执行测试套件](#4-run---批量执行测试套件)
5. [capture - 浏览器场景捕获](#5-capture---浏览器场景捕获)
6. [trace - 追踪查询](#6-trace---追踪查询)
7. [退出码参考](#7-退出码参考)
8. [Artifact 输出能力](#8-artifact-输出能力)

---

## 1. 概览

| 项目 | 值 |
|------|-----|
| 包名 | proto-tester |
| 版本 | 2.0.0 |
| 类型 | ESM (type: "module") |
| 入口 | `dist/cli/index.js`（编译后）/ `src/cli/index.ts`（源码） |
| 运行方式 | `npx proto-tester <command>` 或 `tsx src/cli/index.ts <command>` |
| 依赖框架 | Commander.js v12 |
| Protobuf 库 | google-protobuf v3.21.2 |
| HTTP 客户端 | axios v1.7.2 |
| 可选依赖 | Playwright v1.49.0（capture 子命令） |

### 子命令列表

| 命令 | 功能 | 状态 |
|------|------|------|
| `send` | 发送单次 Protobuf 请求到 Gateway | **READY** |
| `run` | 按 YAML suite 定义批量执行测试用例 | **READY** |
| `capture` | 启动浏览器捕获 UI 场景 | **READY**（需 vite dev server） |
| `trace` | 按 traceId 查询全链路追踪 | **DEGRADED**（exit code 4） |

---

## 2. 全局约定

### 2.1 编码流水线

```
用户输入 payload (JSON object)
    │
    ▼
JSON.stringify(payloadObj)          ← 序列化为 JSON 字符串
    │
    ▼
Buffer.from(jsonString)             ← 转为 Node.js Buffer (UTF-8)
    │
    ▼
new Uint8Array(buffer)              ← 转为 Uint8Array
    │
    ▼
encodePacket({ ..., payload })      ← 写入 MessagePacket.data
    │
    ▼
packet.serialize()                  ← google-protobuf Protobuf 二进制序列化
    │
    ▼
HTTP POST /api/hello                ← Content-Type: application/octet-stream
```

### 2.2 自动填充字段

每次 `encodePacket()` 自动填充以下 extend 字段：

| 字段 | 来源 | 示例值 |
|------|------|--------|
| `method` | 从 protocol.requestMessage 或手动指定 | `com.mineplanet.pojo.health.ServiceHealthCheckRequest` |
| `traceId` | 参数传入或自动 UUID v4 | `a1b2c3d4-e5f6-7890-abcd-ef1234567890` |
| `requestId` | 自动 UUID v4 | `f7e6d5c4-b3a2-1098-fedc-ba0987654321` |

### 2.3 环境选择

`--env` 参数控制 Gateway 基础 URL：

| env 值 | Gateway URL | 说明 |
|--------|-------------|------|
| `dev` | `http://127.0.0.1:8080` | 本地开发 |
| `prod` | 生产地址（受控） | 需审批才能使用 |

可通过 `--gateway <url>` 显式覆盖。

---

## 3. send - 发送单次请求

### 3.1 命令签名

```bash
proto-tester send \
  --max <number>          \   # 必填：request_max（协议大类型）
  --min <number>          \   # 必填：request_min（协议小类型）
  [--payload <json>]      \   # 可选：业务 payload（JSON 字符串或对象）
  [--user <string>]       \   # 可选：用户标识
  [--env <dev|prod>]      \   # 可选：环境（默认 dev）
  [--gateway <url>]       \   # 可选：自定义 Gateway URL
  [--token <string>]      \   # 可选：Bearer Token
  [--outputDir <path>]        # 可选：输出目录
```

### 3.2 参数详解

| 参数 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--max` | **是** | number | — | 协议大类型（maxType），对应 routes.yaml 的 `request_max` |
| `--min` | **是** | number | — | 协议小类型（minType），对应 routes.yaml 的 `request_min` |
| `--payload` | 否 | JSON string/object | `{}` | 业务请求体。将被 `JSON.stringify()` 后写入 `MessagePacket.data` |
| `--user` | 否 | string | `"e2e-tester"` | 用户标识，写入 extend |
| `--env` | 否 | `dev\|prod` | `dev` | 目标环境 |
| `--gateway` | 否 | URL | 由 env 决定 | 显式指定 Gateway 地址（覆盖 env） |
| `--token` | 否 | string | — | Bearer Token，用于 auth_required=true 的路由 |
| `--outputDir` | 否 | path | 当前目录 | 响应 artifact 输出目录 |

### 3.3 退出码

| 退出码 | 含义 | 触发条件 |
|--------|------|----------|
| 0 | 成功 | HTTP 200 且业务返回成功 |
| 1 | 业务失败 | HTTP 200 但业务逻辑返回错误 |
| 2 | 传输失败 | 网络错误、HTTP 非 200、超时等 |
| 3 | 参数错误或生产阻断 | 缺少必填参数、env=prod 未授权等 |

### 3.4 Artifact 输出

send 命令在 `--outputDir` 目录下输出：

| 文件 | 格式 | 说明 |
|------|------|------|
| `response-<max>-<min>-<timestamp>.bin` | Protobuf binary | 原始响应二进制 |
| `response-<max>-<min>-<timestamp>.json` | JSON | 解码后的结构化响应（含 maxType/minType/extend/data） |

### 3.5 使用示例

```bash
# Health 检查（无需 Token）
npx proto-tester send --max 2100 --min 2097

# GetAppConfigs（需要 Token）
npx proto-tester send --max 6000 --min 6001 --token "Bearer eyJhbGci..."

# 自定义 payload
npx proto-tester send --max 2100 --min 2101 --payload '{"name":"test"}'

# 指定输出目录
npx proto-tester send --max 2100 --min 2097 --outputDir ./e2e-results
```

---

## 4. run - 批量执行测试套件

### 4.1 命令签名

```bash
proto-tester run \
  --suite <file.yaml>     \   # 必填：测试套件 YAML 文件
  [--parallel]            \   # 可选：并行执行
  [--env <dev|prod>]      \   # 可选：环境（默认 dev）
  [--gateway <url>]       \   # 可选：自定义 Gateway URL
```

### 4.2 参数详解

| 参数 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--suite` | **是** | path (.yaml/.yml) | — | 测试套件定义文件路径 |
| `--parallel` | 否 | flag | false（串行） | 是否并行执行用例 |
| `--env` | 否 | `dev\|prod` | `dev` | 目标环境 |
| `--gateway` | 否 | URL | 由 env 决定 | 显式指定 Gateway 地址 |

### 4.3 Suite YAML 格式

```yaml
# e2e-suite-mvp.yaml 示例结构
name: "E2E MVP Suite"
version: "1.0"
cases:
  - id: "E2E-MVP-001"
    name: "Health 正向检查"
    maxType: 2100
    minType: 2097
    payload: {}
    expect:
      httpStatus: 200
      # ... 更多断言规则
```

### 4.4 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 全部用例通过 |
| 1 | 任一用例失败 |

### 4.5 Artifact 输出

run 命令输出：

| 文件 | 格式 | 说明 |
|------|------|------|
| `run-report-<timestamp>.json` | JSON | 完整测试报告（每用例结果、耗时、断言详情） |
| 各用例响应文件 | .bin + .json | 同 send 命令的输出格式 |

---

## 5. capture - 浏览器场景捕获

### 5.1 命令签名

```bash
proto-tester capture \
  --scenario <name>       \   # 必填：场景名称
  [--video]               \   # 可选项：录制视频
  [--screenshot]               # 可选项：截图
```

### 5.2 参数详解

| 参数 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--scenario` | **是** | string | — | 场景名称（对应预定义场景脚本） |
| `--video` | 否 | flag | false | 录制 MP4 视频 |
| `--screenshot` | 否 | flag | false | 截取 PNG 截图 |

### 5.3 前置依赖

| 依赖 | 版本 | 说明 |
|------|------|------|
| Vite Dev Server | v5.x | 必须在 `http://127.0.0.1:3001` 运行 |
| Playwright | ^1.49.0 | optionalDependencies，需单独安装 |
| Chromium/Firefox/WebKit | Playwright browser | 需 `npx playwright install` |

### 5.4 Artifact 输出

capture 命令输出：

| 文件 | 格式 | 说明 |
|------|------|------|
| `capture-<scenario>-<timestamp>.mp4` | video | 录屏（--video 时） |
| `capture-<scenario>-<timestamp>.png` | image | 截图（--screenshot 时） |
| `capture-<scenario>-<timestamp>.json` | JSON | 结构化捕获报告 |

---

## 6. trace - 追踪查询

### 6.1 命令签名

```bash
proto-tester trace \
  --id <traceId>          \   # 必填：追踪 ID
  [--since <duration>]    \   # 可选：时间窗口起始偏移
  [--outputDir <path>]        # 可选：输出目录
```

### 6.2 参数详解

| 参数 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--id` | **是** | string (UUID) | — | 要查询的 traceId（通常从 send 响应的 extend 中获取） |
| `--since` | 否 | duration string | `"1h"` | 时间窗口（如 `"30m"`, `"2h"`） |
| `--outputDir` | 否 | path | 当前目录 | 输出目录 |

### 6.3 当前状态：DEGRADED

| 项目 | 值 |
|------|-----|
| **状态** | **DEGRADED** |
| 退出码 | **4**（固定） |
| 退化原因 | `/api/dev/trace` 后端 API 尚未实现 |
| 源码证据 | [trace.ts L22-25](../../../typescript/proto-tester/src/cli/commands/trace.ts#L22-L25), [trace.ts L46](../../../typescript/proto-tester/src/cli/commands/trace.ts#L46) |
| 替代方案 | `grep <traceId> gateway.log` 从服务端日志检索 |

### 6.4 预期行为（API 就绪后）

当 `/api/dev/trace` API 实现后，trace 命令将：
1. 向 Gateway 发送 GET `/api/dev/trace?id=<traceId>&since=<duration>`
2. 接收该 traceId 的全链路追踪记录（HTTP→Gateway→Invoker→Handler→响应）
3. 输出结构化追踪报告 JSON

---

## 7. 退出码参考

### 7.1 全局退出码汇总

| 退出码 | 名称 | 适用命令 | 含义 |
|--------|------|----------|------|
| 0 | EXIT_SUCCESS | 全部 | 操作成功 |
| 1 | EXIT_FAIL | send, run, capture | 业务逻辑失败 / 用例失败 |
| 2 | EXIT_TRANSPORT_ERROR | send | 网络错误 / HTTP 非 200 / 超时 |
| 3 | EXIT_PARAM_ERROR / EXIT_PROD_BLOCKED | send | 参数缺失或生产环境未授权 |
| 4 | EXIT_DEGRADED | trace | 功能降级（API 未就绪） |

### 7.2 Shell 集成示例

```bash
# 基本成功判断
npx proto-tester send --max 2100 --min 2097
if [ $? -eq 0 ]; then echo "PASS"; else echo "FAIL"; fi

# 区分失败类型
exit_code=$?
case $exit_code in
  0) echo "SUCCESS" ;;
  1) echo "BUSINESS_FAIL" ;;
  2) echo "TRANSPORT_ERROR" ;;
  3) echo "PARAM_OR_BLOCKED" ;;
  4) echo "DEGRADED" ;;
esac

# trace 降级检测
npx proto-tester trace --id "$trace_id"
if [ $? -eq 4 ]; then
  echo "TRACE_DEGRADED: using fallback grep"
  grep "$trace_id" /var/log/gateway.log
fi
```

---

## 8. Artifact 输出能力

### 8.1 输出文件矩阵

| 命令 | 文件类型 | 格式 | 触发条件 |
|------|----------|------|----------|
| send | `response-*.bin` | Protobuf binary | 每次执行均输出 |
| send | `response-*.json` | JSON (decoded) | 每次执行均输出 |
| run | `run-report-*.json` | JSON report | suite 执行完输出 |
| run | `response-*.bin/json` | 同 send | 每个用例输出 |
| capture | `capture-*.mp4` | video | `--video` 参数 |
| capture | `capture-*.png` | image | `--screenshot` 参数 |
| capture | `capture-*.json` | JSON report | 每次执行均输出 |
| trace | `trace-*.json` | JSON report | API 就绪时输出 |

### 8.2 输出目录结构（推荐）

```
e2e-artifacts/
├── preflight/
│   └── E2E-PREFLIGHT.md              # 本文件
├── mvp/
│   ├── run-report-YYYYMMDD-HHMMSS.json
│   ├── response-2100-2097-*.bin
│   ├── response-2100-2097-*.json
│   ├── response-6000-6001-*.bin
│   └── response-6000-6001-*.json
├── regression/
│   └── run-report-YYYYMMDD-HHMMSS.json
├── capture/
│   ├── capture-login-*.mp4
│   └── capture-login-*.png
└── trace/
    └── trace-<traceId>.json          #（API 就绪后）
```

---

## 附录 A：快速参考卡

```bash
# === Health 检查（最常用）===
npx proto-tester send --max 2100 --min 2097

# === Config/I18n（需 Token）===
npx proto-tester send --max 6000 --min 6001 --token "Bearer <jwt>"
npx proto-tester send --max 6000 --min 6005 --token "Bearer <jwt>"
npx proto-tester send --max 6000 --min 6003                    # 无需 Token

# === 批量执行===
npx proto-tester run --suite ./e2e-suite-mvp.yaml
npx proto-tester run --suite ./e2e-suite-mvp.yaml --parallel

# === 浏览器捕获（需 vite dev server）===
npx proto-tester capture --scenario login --video --screenshot

# === 追踪查询（当前 DEGRADED）===
npx proto-tester trace --id <trace-id-from-response>
# 退出码 4 → 使用替代方案: grep <trace-id> gateway.log
```

---

*本文档由 Phase 0 Preflight 静态分析自动生成，如 CLI 行为变更需重新执行 preflight 更新。*
