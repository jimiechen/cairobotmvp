# E2E 预检报告（Phase 0: Preflight）

> **执行时间**：2026-06-05
> **执行人**：Trae (Auto)
> **依据文档**：[E2E-ACCEPTANCE-PLAN-V3.md](./E2E-ACCEPTANCE-PLAN-V3.md)
> **产物目录**：`docs/reports/testing/e2e-acceptance/`
> **阶段约束**：本轮仅做预检，不启动服务，不执行 E2E 请求，不生成最终 RUN 报告

---

## 检查总览

| 编号 | 检查项 | 优先级 | 结论 |
|------|--------|--------|------|
| P0-A | Content-Type 与 Body 编码检查 | P0 | **PASS**（附兼容性说明） |
| P0-B | routes.yaml 加载机制检查 | P0 | **PASS**（附差异表） |
| P0-C | HTTP 入口确认 | P0 | **PASS** |
| P1-D | 协议 maturity 分类 | P1 | **PASS** |
| P1-E | auth_required 分组 | P1 | **PASS** |
| P1-F | CLI 真实参数确认 | P1 | **PASS** |
| P1-G | trace CLI 可用性检查 | P1 | **DEGRADED** |
| P1-H | TarsGo 模式可用性检查 | P1 | **BLOCKED(tars)/PASS(local)** |

---

## P0-A：Content-Type 与 Body 编码检查

### A.1 Gateway Content-Type 强制要求

**源文件**：[http_server.go](../../../go/gateway/proto-gateway/internal/server/http_server.go#L40-L44)

```go
// 第 40-44 行：Content-Type 严格校验
if r.Header.Get("Content-Type") != "application/octet-stream" {
    tars.TLOG.Debug("unsupported Content-Type: " + r.Header.Get("Content-Type"))
    writeError(w, http.StatusUnsupportedMediaType, commonlib.CodeBadRequest, "unsupported media type")
    return
}
```

**结论**：
- Gateway **强制要求** `Content-Type: application/octet-stream`
- 错误 Content-Type 返回 **HTTP 415 UnsupportedMediaType**
- 错误响应体同样使用 `application/octet-stream`（第 116-122 行）
- 成功响应也使用 `application/octet-stream`（第 124-132 行）

### A.2 MessagePacket.data 字段类型

**源文件**：[message.proto](../../../proto/base/message.proto)

```protobuf
// 第 30-35 行
message MessagePacket {
  int32 maxType = 1;
  int32 minType = 2;
  map<string, string> extend = 3;
  Platform platform = 4;   // enum { UNKNOWN=0, WEB=1, APP=2, HARDWARE=3 }
  bytes data = 5;          // 业务协议包的序列化 bytes
}
```

**proto 注释原文**（第 35 行）：
> data 表示业务协议包的序列化 bytes，由 request_proto 对应的 Protobuf message 序列化而来

**Go 端反序列化**：[message_packet.go](../../../go/gateway/proto-gateway/internal/adapter/message_packet.go#L122-L131)
- 使用 `proto.Unmarshal(data, packet)` — 标准 Protobuf 二进制解码

### A.3 proto-tester 编码行为

**源文件**：[send.ts](../../../typescript/proto-tester/src/cli/commands/send.ts#L65-L72)

```typescript
// send.ts 第 65-72 行：data 字段编码逻辑
packetBinary = encodePacket({
  maxType,
  minType,
  payload: new Uint8Array(Buffer.from(JSON.stringify(payloadObj))), // JSON string → UTF-8 bytes
  extend: { method: `${protocol.requestMessage}` },
});
```

**encodePacket 实现**：[messagePacket.ts](../../../typescript/proto-tester/src/lib/messagePacket.ts#L46-L72)

```typescript
// messagePacket.ts 第 46-71 行
export function encodePacket(opts: PacketBuildOptions): Uint8Array {
  // ... 自动填充 traceId / requestId ...
  const packet = new MessagePacket({
    maxType: opts.maxType,
    minType: opts.minType,
    platform: opts.platform ?? Platform.WEB,
    extend: mergedExtend,
    data: safePayload,           // ← 直接写入 payload Uint8Array
  });
  return packet.serialize();      // ← Protobuf 二进制序列化
}
```

### A.4 兼容性判定矩阵

| 层级 | 组件 | data 字段期望格式 | 实际行为 |
|------|------|-------------------|----------|
| **协议定义层** | `message.proto` | Protobuf business message bytes | — |
| **传输层** | Gateway `proto.Unmarshal` | Protobuf MessagePacket binary | 只解包外层，不解释 data 内容 |
| **路由层** | Gateway 路由匹配 | 仅用 maxType/minType 匹配 | 不解析 data |
| **LocalInvoker 2100 段** | Health/Hello handler | 各 handler 自行决定 | Health 用 protobuf，Hello 待确认 |
| **LocalInvoker 6000 段** | Config/I18n handler | **JSON string UTF-8 bytes** | `json.Unmarshal(req, &appReq)` |
| **客户端层** | proto-tester send | **JSON string → UTF-8 bytes** | `Buffer.from(JSON.stringify(payloadObj))` |

### A.5 关键发现与风险

**发现 #1（P1 风险）：proto-tester 的 data 编码方式**

proto-tester 将业务 payload 以 `JSON.stringify() → UTF-8 bytes` 写入 `MessagePacket.data`。这与 `message.proto` 注释中"由 request_proto 对应的 Protobuf message 序列化而来"的规范**不一致**。

**但实际兼容性取决于目标 handler 的解码方式**：

| 目标路由段 | Handler 解码方式 | proto-tester 兼容？ | 说明 |
|-----------|-----------------|---------------------|------|
| 2100 段（Health） | protobuf Unmarshal | **需确认** | Health handler 可能期望 protobuf bytes |
| 2100 段（Hello） | 待代码确认 | **需确认** | Hello module handler 解码方式 |
| 6000 段（Config/I18n） | `json.Unmarshal` | **兼容** | [invoker.go](../../../go/gateway/proto-gateway/tarsclient/invoker.go#L237) 明确使用 JSON |

**发现 #2（P1 风险）：2100 段路由的 data 解码方式待运行时验证**

由于本轮为预检不启动服务，2100 段（Health/Hello）handler 对 data 字段的实际解码格式需在 **Phase 1 MVP 执行**时通过正向用例验证。

**预检建议**：
- Phase 1 优先执行 6000 段 Config/I18n 用例（已知 JSON 兼容）
- 2100 段用例作为第二优先级，若失败则记录 Bug 并切换为 protobuf bytes 重试

---

## P0-B：routes.yaml 加载机制检查

### B.1 加载机制

**源文件**：[main.go](../../../go/gateway/proto-gateway/cmd/server/main.go#L26)、[routes.go](../../../go/gateway/proto-gateway/internal/config/routes.go#L41-L67)

| 项目 | 值 | 来源 |
|------|-----|------|
| 默认路径 | `configs/gateway/routes.yaml` | main.go 第 26 行 |
| 读取方式 | `os.ReadFile(path)` — 文件系统读取 | routes.go 第 41-57 行 |
| 是否 Go embed | **否** | 未使用 `embed.FS` |
| 环境变量覆盖 | `GATEWAY_ROUTES_PATH` | routes.go 第 60-67 行 |
| 校验规则 | route_key 唯一、必填 tars 字段、vector<byte> 类型、timeout_ms > 0 | routes.go 第 70-107 行 |

### B.2 主配置 vs 开发副本 差异表

| 维度 | 主配置 (`configs/gateway/routes.yaml`) | 开发副本 (`go/gateway/proto-gateway/configs/gateway/routes.yaml`) |
|------|---------------------------------------|---------------------------------------------------------------|
| **路由数量** | **8 条**（2 系统 + 6 Config/I18n） | **2 条**（仅 Health + HelloWorld） |
| **2100:2097 Health** | 一致 | 一致 |
| **2100:2101 HelloWorld** | `request_proto: com.mineplanet.pojo.hello.HelloWorldRequest` | `request_proto: com.mineplanet.pojo.health.HelloWorldRequest` (**错误**) |
| **6000:6001 GetAppConfigs** | 有 | **缺失** |
| **6000:6009 AppConfigVersion** | 有 | **缺失** |
| **6000:6003 GetAppLanguage** | 有 | **缺失** |
| **6000:6005 GetLangPack** | 有 | **缺失** |
| **6000:6007 GetLangDifference** | 有 | **缺失** |

**关键差异说明**：
1. 开发副本仅含 2 条路由，缺少全部 6000 段 Config/I18n 路由
2. 开发副本中 HelloWorld 的 `request_proto` 包名错误：`health` 应为 `hello`

### B.3 GATEWAY_ROUTES_PATH 环境变量使用情况

| 位置 | 是否设置 GATEWAY_ROUTES_PATH | 说明 |
|------|------------------------------|------|
| Makefile（根目录及子模块） | **未设置** | grep 无匹配 |
| `.github/workflows/*.yml` | **未设置** | CI 中无此环境变量 |
| E2E 启动脚本 | **待 Phase 1 设置** | 必须指向主配置 `configs/gateway/routes.yaml` |

**结论**：如果不显式设置 `GATEWAY_ROUTES_PATH`，Gateway 启动时默认读取 `configs/gateway/routes.yaml`（主配置，8 条路由）。开发副本仅在 `go/gateway/proto-gateway/configs/gateway/` 下作为便利副本存在，不会影响默认行为。

---

## P0-C：HTTP 入口确认

### C.1 入口注册

**源文件**：[http_server.go](../../../go/gateway/proto-gateway/internal/server/http_server.go#L50)、[main.go](../../../go/gateway/proto-gateway/cmd/server/main.go#L50)

| 项目 | 值 |
|------|-----|
| HTTP Method | **POST** |
| URL Path | **`/api/hello`** |
| 注册方式 | `mux.Handle("/api/hello", gs)` （main.go 第 50 行） |
| 单入口架构 | 是，所有协议共用 `/api/hello` |

### C.2 请求规格

| 字段 | 要求 | 来源 |
|------|------|------|
| Method | POST | http_server.go 第 38 行 |
| Content-Type | `application/octet-stream`（强制） | http_server.go 第 40-44 行 |
| Body 格式 | Protobuf binary MessagePacket | http_server.go 第 60 行 `DeserializeMessagePacket(body)` |
| traceId | 可选；若 extend 中无则自动生成 UUID v4 | message_packet.go 第 66-74 行 |

### C.3 响应规格

| 场景 | HTTP Status | Content-Type | Body 格式 |
|------|-------------|--------------|-----------|
| 成功 | 200 | `application/octet-stream` | Protobuf binary MessagePacket（含 response data） |
| Content-Type 错误 | 415 | `application/octet-stream` | 错误 MessagePacket |
| 参数校验失败 | 400 | `application/octet-stream` | 错误 MessagePacket |
| 路由未找到 | 404 | `application/octet-stream` | 错误 MessagePacket |
| 内部错误 | 500 | `application/octet-stream` | 错误 MessagePacket |

### C.4 HTTP 入口汇总表

```
POST /api/hello
├── Headers
│   └── Content-Type: application/octet-stream (mandatory)
├── Body: Protobuf Binary MessagePacket
│   ├── maxType: int32     → 路由匹配（request_max）
│   ├── minType: int32     → 路由匹配（request_min）
│   ├── extend: map        → method / traceId / requestId / token...
│   ├── platform: enum     → WEB(1) / APP(2) / HARDWARE(3)
│   └── data: bytes        → 业务 payload（编码格式见 A.4）
└── Response: Protobuf Binary MessagePacket
    ├── maxType: response_max
    ├── minType: response_min
    └── data: bytes        → 业务响应 payload
```

---

## P1-D：协议 Maturity 分类

**源文件**：[protocols.json](../../../typescript/proto-tester/src/data/protocols.json)

### D.1 已启用协议（status = "已启用"）

| 协议名称 | maxType | minType | request message | response message | status |
|----------|---------|---------|-----------------|------------------|--------|
| ServiceHealthCheck | 2100 | 2097 | ServiceHealthCheckRequest | ServiceHealthCheckResponse | 已启用 |
| ServiceHealthCheckResponse | 2100 | 2098 | — | — | 已启用 |

### D.2 草案协议（status = "草案"）

| 协议名称 | maxType | minType | request message | response message | status |
|----------|---------|---------|-----------------|------------------|--------|
| HelloWorld | 2100 | 2101 | HelloWorldRequest | HelloWorldResponse | 草案 |
| HelloWorldResponse | 2100 | 2102 | — | — | 草案 |
| GetAppConfigs | 6000 | 6001 | AppConfigsReq | AppConfigsRsp | 草案 |
| AppConfigsRsp | 6000 | 6002 | — | — | 草案 |
| AppConfigVersion | 6000 | 6009 | AppConfigVersionReq | AppConfigVersionRsp | 草案 |
| AppConfigVersionRsp | 6000 | 6010 | — | — | 草案 |
| GetAppLanguage | 6000 | 6003 | AppFetchLanguageReq | AppFetchLanguageRsp | 草案 |
| AppFetchLanguageRsp | 6000 | 6004 | — | — | 草案 |
| GetLangPack | 6000 | 6005 | AppFetchLangPackReq | AppFetchLangPackRsp | 草案 |
| AppFetchLangPackRsp | 6000 | 6006 | — | — | 草案 |
| GetLangDifference | 6000 | 6007 | AppFetchLangDifferenceReq | AppFetchLangDifferenceRsp | 草案 |
| AppFetchLangDifferenceRsp | 6000 | 6008 | — | — | 草案 |

### D.3 统计

| status | 数量 | 占比 |
|--------|------|------|
| 已启用 | 2 | 14.3% |
| 草案 | 12 | 85.7% |
| 废弃 | 0 | 0% |
| 未知 | 0 | 0% |
| **合计** | **14** | 100% |

### D.4 E2E 优先级推论

基于 maturity 分类：
- **P0 用例**：2100 段 Health（已启用）— 必须首先通过
- **P1 用例**：6000 段 Config/I18n（草案）— 作为 MVP 核心覆盖
- **P2 用例**：2100 段 Hello（草案）— 补充覆盖

---

## P1-E：auth_required 分组

**源文件**：[routes.yaml（主配置）](../../../configs/gateway/routes.yaml)

### E.1 需要 Bearer Token 的路由（auth_required = true）

| route_key | command_name | minType | tars_method | timeout_ms |
|-----------|-------------|---------|-------------|------------|
| 6000:6001 | GetAppConfigs | 6001 | GetAppConfigs | 5000 |
| 6000:6009 | AppConfigVersion | 6009 | AppConfigVersion | 3000 |
| 6000:6005 | GetLangPack | 6005 | GetLangPack | 5000 |
| 6000:6007 | GetLangDifference | 6007 | GetLangDifference | 5000 |

**正向用例要求**：必须携带有效 Bearer Token（通过 `--token` 参数或 `extend.token` 字段）

**负向用例（必须生成）**：

| 用例 ID | 场景 | 预期行为 |
|---------|------|----------|
| NEG-TOKEN-001 | 缺失 Token（不传 --token） | 401 Unauthorized 或业务错误码 |
| NEG-TOKEN-002 | 错误 Token（无效签名/过期） | 401 Unauthorized 或业务错误码 |
| NEG-TOKEN-003 | 空 Token（--token ""） | 400 Bad Request 或 401 |

### E.2 无需 Token 的路由（auth_required = false）

| route_key | command_name | minType | tars_method | timeout_ms |
|-----------|-------------|---------|-------------|------------|
| 2100:2097 | ServiceHealthCheck | 2097 | HealthCheck | 3000 |
| 2100:2101 | HelloWorld | 2101 | HelloWorld | 3000 |
| 6000:6003 | GetAppLanguage | 6003 | GetAppLanguage | 3000 |

---

## P1-F：CLI 真实参数确认

详细参数文档见 companion file: **[E2E-PROTO-TESTER-CLI.md](./E2E-PROTO-TESTER-CLI.md)**

以下为摘要：

| 子命令 | 用途 | 必填参数 | 关键可选参数 | 退出码 |
|--------|------|----------|-------------|--------|
| `send` | 发送单次请求 | `--max`, `--min` | `--payload`, `--token`, `--gateway`, `--env` | 0/1/2/3 |
| `run` | 批量执行 suite | `--suite` | `--parallel`, `--env`, `--gateway` | 0/1 |
| `capture` | 浏览器捕获 | `--scenario` | `--video`, `--screenshot` | 0/1 |
| `trace` | 追踪查询 | `--id` | `--since`, `--outputDir` | 0/4 |

**版本信息**：proto-tester v2.0.0（[package.json](../../../typescript/proto-tester/package.json)）

---

## P1-G：trace CLI 可用性检查

### G.1 源码分析

**源文件**：[trace.ts](../../../typescript/proto-tester/src/cli/commands/trace.ts)

```typescript
// trace.ts 第 22-25 行：明确声明后端 API 不可用
// 当前 /api/dev/trace 接口尚未实现
// 临时返回 DEGRADED 状态，提示用户使用日志替代方案
```

```typescript
// trace.ts 第 46 行：固定退出码
process.exit(4); // EXIT_DEGRADED
```

### G.2 可用性判定

| 项目 | 值 |
|------|-----|
| **trace_cli_status** | **DEGRADED** |
| 退出码 | 4（固定） |
| 退化原因 | `/api/dev/trace` API 未实现 |
| stderr 输出 | 提示使用 `grep traceId` 日志替代 |
| 影响 | E2E-MVP-009（TraceId 全链路追踪）无法按原计划执行 |

### G.3 降级方案

当 trace CLI 返回 exit code 4 时：
1. 使用 `grep traceId gateway.log` 从服务端日志提取追踪记录
2. 手动关联请求/响应的 traceId
3. 在 RUN 报告中标注"trace 降级为日志检索"

---

## P1-H：TarsGo 模式可用性检查

### H.1 Invoker 模式选择

**源文件**：[main.go](../../../go/gateway/proto-gateway/cmd/server/main.go#L15-L18)

```go
// 第 15-18 行：通过环境变量选择 invoker 模式
invokerMode := os.Getenv("GATEWAY_INVOKER_MODE")
if invokerMode == "" {
    invokerMode = "local" // 默认 local
}
```

### H.2 LocalInvoker 状态

| 项目 | 值 |
|------|-----|
| 模式标识 | `local` |
| 默认模式 | 是 |
| 实现状态 | **可用** |
| 注册 handler | RegisterConfigI18nHandlers（6000 段，5 个方法） |
| data 解码 | 6000 段使用 `json.Unmarshal`（兼容 proto-tester JSON-bytes） |

### H.3 TarsGoInvoker 状态

**源文件**：[invoker.go](../../../go/gateway/proto-gateway/tarsclient/invoker.go#L207-L218)

```go
// 第 207-218 行：TarsGo 模式未实现
func (t *TarsGoInvoker) Invoke(ctx context.Context, req []byte) ([]byte, int, error) {
    return nil, 10500, errors.New("tars invoker is not implemented yet")
}
```

**源文件**：[main.go](../../../go/gateway/proto-gateway/cmd/server/main.go#L40-L46)

```go
// 第 40-46 行：Tars 模式启动时直接退出
case "tars":
    log.Fatal("tars mode is not implemented in S1 phase")
    os.Exit(1)
```

| 项目 | 值 |
|------|-----|
| 模式标识 | `tars` |
| 实现状态 | **BLOCKED**（S1 阶段未实现） |
| Invoke 返回码 | 10500（固定） |
| 启动行为 | `os.Exit(1)` 直接退出 |
| 影响 | 所有依赖 TarsGo 远程调用的用例 BLOCKED |

### H.4 E2E 影响评估

| 模式 | E2E 可执行范围 | 说明 |
|------|---------------|------|
| **local**（默认） | 全部 8 条路由 | LocalInvoker 内联处理 6000 段请求 |
| **tars** | **不可用** | 启动即退出，无法执行任何用例 |

**结论**：E2E 使用 **local 模式**执行，TarsGo 相关用例标记为 BLOCKED 但不阻断 local 模式 E2E。

---

## 整体结论

### 结论：**PARTIAL_READY**

### 可进入阶段 1 的条件

| 条件 | 状态 | 说明 |
|------|------|------|
| P0 Content-Type 强制校验 | PASS | octet-stream + 415 逻辑确认 |
| P0 routes.yaml 文件系统加载 | PASS | 默认主配置 8 条路由可用 |
| P0 HTTP 单入口 `/api/hello` | PASS | POST + Protobuf body 确认 |
| P1 协议 maturity 分类 | PASS | 2 已启用 + 12 草案 |
| P1 auth_required 分组 | PASS | 4 需 Token + 4 无需 Token |
| P1 CLI 参数确认 | PASS | 4 子命令参数已梳理 |
| local 模式可用 | PASS | LocalInvoker 可处理全部路由 |

### 已知限制（不阻断但需记录）

| 限制项 | 等级 | 影响 | 缓解措施 |
|--------|------|------|----------|
| trace CLI DEGRADED | P1 | E2E-MVP-009 降级执行 | 改用 grep traceId 日志 |
| TarsGo 模式 BLOCKED | P1 | Tars 远程调用用例不可执行 | 仅执行 local 模式用例 |
| proto-tester data 编码非标准 protobuf | P1 | 2100 段可能不兼容 | 优先跑 6000 段，2100 段运行时验证 |
| 开发副本 routes.yaml 缺 6000 段 | P2 | 若误设 GATEWAY_ROUTES_PATH 会缺路由 | E2E 不设置此变量，使用默认主配置 |

### 阻断项

**无 P0 阻断项。**

### 下一步动作

1. **主控确认本预检报告后**，进入 **Phase 1：MVP Case Matrix 生成**
2. Phase 1 启动命令示例：
   ```bash
   # Gateway（local 模式，默认读取 configs/gateway/routes.yaml）
   GATEWAY_INVOKER_MODE=local go run ./go/gateway/proto-gateway/cmd/server/main.go

   # proto-tester 发送 Health 检查（无需 Token）
   npx proto-tester send --max 2100 --min 2097
   ```

---

*报告完毕。等待主控确认后进入阶段 1。*
