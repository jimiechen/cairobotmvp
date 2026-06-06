# E2E 用例矩阵（Phase 1: Case Matrix）

> **生成时间**：2026-06-05
> **依据**：[E2E-PREFLIGHT.md](./E2E-PREFLIGHT.md) + 源码静态分析
> **执行阶段**：Phase 1 — 仅生成矩阵和 suite 文件，不执行请求
> **配套文件**：
> - [E2E-MVP-SUITE.yaml](./E2E-MVP-SUITE.yaml)
> - [E2E-REGRESSION-SUITE.yaml](./E2E-REGRESSION-SUITE.yaml)
> - [E2E-AUTH-CHANNEL.md](./E2E-AUTH-CHANNEL.md)
> - [E2E-ARTIFACT-SPEC.md](./E2E-ARTIFACT-SPEC.md)

---

## 判定规则总表

| 规则编号 | 规则描述 | 判定 |
|----------|----------|------|
| D-01 | Content-Type 错误返回 415 是 PASS 的负向用例 | 负向用例期望 415 即为 PASS |
| D-02 | trace CLI DEGRADED(exit=4) 不导致 traceId 用例失败 | 改用 grep 日志替代，记录 DEGRADED |
| D-03 | TarsGo 模式未实现导致相关用例 BLOCKED | 标记 BLOCKED，不影响 local MVP 验收 |
| D-04 | 草案协议 Handler 未实现时 BLOCKED | 不得判 FAIL |
| D-05 | 已启用协议外层 MessagePacket 或路由失败应判 FAIL | Health 是已启用协议，必须严格判定 |
| D-06 | 2100 段因 proto-tester data 编码与 handler 期望不一致失败 | 关联 E2E-REG-025，按项目规范决定 FAIL/BLOCKED |

---

## 一、MVP 必选集（13 条）

### E2E-MVP-001：Preflight 结果确认

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-001 |
| name | Preflight 结果确认 — 验证 Phase 0 预检结论可复现 |
| priority | P0 |
| executionOrder | 1 |
| route_key | N/A（元用例） |
| command_name | N/A |
| maxType | N/A |
| minType | N/A |
| responseMaxType | N/A |
| responseMinType | N/A |
| request_proto | N/A |
| response_proto | N/A |
| protocol_status | N/A |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | N/A |
| data_encoding_actual_from_proto_tester | N/A |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | N/A（元检查） |
| expected_gateway_behavior | 确认 Gateway 以 local 模式启动，routes.yaml 加载 8 条路由 |
| expected_service_behavior | LocalInvoker 已注册 System + Config/I18n handlers |
| trace_assertion | N/A |
| artifact_required | gateway 启动日志截图/副本 |
| blockedWhen | Gateway 无法启动或 routes.yaml 加载失败 |
| failWhen | 预检结论 PARTIAL_READY 的已知限制无法接受时 |

---

### E2E-MVP-012：错误 Content-Type 返回 415

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-012 |
| name | Content-Type 强制校验 — 非 octet-stream 返回 415 |
| priority | P0 |
| executionOrder | 2 |
| route_key | N/A（前置拦截） |
| command_name | N/A |
| maxType | N/A |
| minType | N/A |
| responseMaxType | 0 |
| responseMinType | 0 |
| request_proto | N/A |
| response_proto | N/A |
| protocol_status | N/A |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | N/A |
| data_encoding_actual_from_proto_tester | N/A |
| content_type | **application/json**（故意错误） |
| request_body_format | 任意（被 415 拦截前不解析） |
| expected_http_status | **415 UnsupportedMediaType** |
| expected_gateway_behavior | http_server.go 第 40-44 行拦截，返回 415 + error MessagePacket |
| expected_service_behavior | 不到达 service 层 |
| trace_assertion | 无 traceId（未进入路由匹配） |
| artifact_required | response.raw.json（含 code + message） |
| blockedWhen | Gateway 未启动 |
| failWhen | HTTP 状态码 ≠ 415 |

**子变体**：

| 变体 ID | Content-Type 值 | 预期 HTTP 状态 |
|---------|-----------------|----------------|
| MVP-012-a | `application/json` | 415 |
| MVP-012-b | `text/plain` | 415 |
| MVP-012-c | （空 / 缺失） | 415 |

---

### E2E-MVP-004：Config 正向 — 有效 Token

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-004 |
| name | GetAppConfigs 正向 — 有效 Bearer Token，6000:6001 |
| priority | P0 |
| executionOrder | 3 |
| route_key | 6000:6001 |
| command_name | GetAppConfigs |
| maxType | 6000 |
| minType | 6001 |
| responseMaxType | 6000 |
| responseMinType | 6002 |
| request_proto | com.mineplanet.pojo.AppConfigsReq |
| response_proto | com.mineplanet.pojo.AppConfigsRsp |
| protocol_status | 草案 |
| auth_required | **true** |
| token_required | **true** |
| token_channel | HTTP Header `Authorization: Bearer <jwt>`（proto-tester --token 写入 Header） |
| data_encoding_expected | Protobuf business message bytes（按 proto 规范） |
| data_encoding_actual_from_proto_tester | **JSON string → UTF-8 bytes**（`Buffer.from(JSON.stringify(payloadObj))`） |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket (data = JSON-bytes) |
| expected_http_status | 200 |
| expected_gateway_behavior | 路由匹配 6000:6001 → LocalInvoker → ConfigObj.GetAppConfigs |
| expected_service_behavior | json.Unmarshal(data, &appReq) → configSvc.GetAppConfigs() → json.Marshal(resp) |
| trace_assertion | 响应 extend 含 traceId；trace CLI DEGRADED 时用 grep |
| artifact_required | request.raw.json + response.raw.json + assertion.json |
| blockedWhen | ConfigService 未注入或 GetAppConfigs 方法不存在 |
| failWhen | HTTP ≠ 200 或业务码 ≠ 10200 或响应 data 为空 |

**payload 示例**：
```json
{}
```

---

### E2E-MVP-005：Config 缺失 Token

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-005 |
| name | GetAppConfigs 负向 — 缺失 Token，6000:6001 |
| priority | P0 |
| executionOrder | 4 |
| route_key | 6000:6001 |
| command_name | GetAppConfigs |
| maxType | 6000 |
| minType | 6001 |
| responseMaxType | 6000 |
| responseMinType | 6002 |
| request_proto | com.mineplanet.pojo.AppConfigsReq |
| response_proto | com.mineplanet.pojo.AppConfigsRsp |
| protocol_status | 草案 |
| auth_required | **true** |
| token_required | **true**（本用例故意不传） |
| token_channel | **无**（--token 参数省略） |
| data_encoding_expected | Protobuf business message bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | **待运行时确认**（见 AUTH-CHANNEL 分析） |
| expected_gateway_behavior | 当前 Gateway **未实现 token 校验**，预计正常到达 handler 并返回 200 |
| expected_service_behavior | 同 MVP-004（token 校验 stub） |
| trace_assertion | 响应含 traceId |
| artifact_required | response.raw.json + assertion.json |
| blockedWhen | 无 |
| failWhen | **取决于鉴权实现状态**：若 auth 为 stub 则预期 200（PASS）；若已实现则预期 401/业务错误码 |

**重要说明**：此用例用于验证鉴权逻辑是否已实现。当前源码分析显示 [http_server.go](../../../go/gateway/proto-gateway/internal/server/http_server.go) 中**无 Authorization Header 读取代码**，auth_required 仅作为元数据传递给 extend。详见 [E2E-AUTH-CHANNEL.md](./E2E-AUTH-CHANNEL.md)。

---

### E2E-MVP-006：Config 错误 Token

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-006 |
| name | GetAppConfigs 负向 — 错误/无效 Token，6000:6001 |
| priority | P0 |
| executionOrder | 5 |
| route_key | 6000:6001 |
| command_name | GetAppConfigs |
| maxType | 6000 |
| minType | 6001 |
| responseMaxType | 6000 |
| responseMinType | 6002 |
| request_proto | com.mineplanet.pojo.AppConfigsReq |
| response_proto | com.mineplanet.pojo.AppConfigsRsp |
| protocol_status | 草案 |
| auth_required | **true** |
| token_required | **true** |
| token_channel | HTTP Header `Authorization: Bearer invalid-token-12345` |
| data_encoding_expected | Protobuf business message bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | **待运行时确认**（同 MVP-005） |
| expected_gateway_behavior | 同 MVP-005（auth stub） |
| expected_service_behavior | 同 MVP-004 |
| trace_assertion | 响应含 traceId |
| artifact_required | response.raw.json + assertion.json |
| blockedWhen | 无 |
| failWhen | 取决于鉴权实现状态 |

---

### E2E-MVP-007：I18n 无 Token 正向

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-007 |
| name | GetAppLanguage 正向 — 无需 Token，6000:6003 |
| priority | P0 |
| executionOrder | 6 |
| route_key | 6000:6003 |
| command_name | GetAppLanguage |
| maxType | 6000 |
| minType | 6003 |
| responseMaxType | 6000 |
| responseMinType | 6004 |
| request_proto | com.mineplanet.pojo.AppFetchLanguageReq |
| response_proto | com.mineplanet.pojo.AppFetchLanguageRsp |
| protocol_status | 草案 |
| auth_required | **false** |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | Protobuf business message bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket (data = JSON-bytes) |
| expected_http_status | 200 |
| expected_gateway_behavior | 路由匹配 6000:6003 → LocalInvoker → I18nObj.GetAppLanguage |
| expected_service_behavior | json.Unmarshal(req, &langReq) → i18nSvc.GetLanguages() → json.Marshal(languages) |
| trace_assertion | 响应 extend 含 traceId |
| artifact_required | request.raw.json + response.raw.json + assertion.json |
| blockedWhen | I18nService 未注入 |
| failWhen | HTTP ≠ 200 或业务码 ≠ 10200 |

**payload 示例**：
```json
{"client_version": "1.0.0"}
```

---

### E2E-MVP-008：I18n 有 Token 正向

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-008 |
| name | GetLangPack 正向 — 需要 Token，6000:6005 |
| priority | P0 |
| executionOrder | 7 |
| route_key | 6000:6005 |
| command_name | GetLangPack |
| maxType | 6000 |
| minType | 6005 |
| responseMaxType | 6000 |
| responseMinType | 6006 |
| request_proto | com.mineplanet.pojo.AppFetchLangPackReq |
| response_proto | com.mineplanet.pojo.AppFetchLangPackRsp |
| protocol_status | 草案 |
| auth_required | **true** |
| token_required | **true** |
| token_channel | HTTP Header `Authorization: Bearer <valid-jwt>` |
| data_encoding_expected | Protobuf business message bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket (data = JSON-bytes) |
| expected_http_status | 200 |
| expected_gateway_behavior | 路由匹配 6000:6005 → LocalInvoker → I18nObj.GetLangPack |
| expected_service_behavior | json.Unmarshal(req, &packReq) → i18nSvc.GetLangPack() → json.Marshal(resp) |
| trace_assertion | 响应 extend 含 traceId |
| artifact_required | request.raw.json + response.raw.json + assertion.json |
| blockedWhen | I18nService 未注入 |
| failWhen | HTTP ≠ 200 或业务码 ≠ 10200 |

**payload 示例**：
```json
{"lang_code": "zh-CN", "client_version": "1.0.0", "env": "dev"}
```

---

### E2E-MVP-002：Health 正向

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-002 |
| name | ServiceHealthCheck 正向 — 已启用协议，2100:2097 |
| priority | P0 |
| executionOrder | 8 |
| route_key | 2100:2097 |
| command_name | ServiceHealthCheck |
| maxType | 2100 |
| minType | 2097 |
| responseMaxType | 2100 |
| responseMinType | 2098 |
| request_proto | com.mineplanet.pojo.health.ServiceHealthCheckRequest |
| response_proto | com.mineplanet.pojo.health.ServiceHealthCheckResponse |
| protocol_status | **已启用** |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | Protobuf ServiceHealthCheckRequest bytes |
| data_encoding_actual_from_proto_tester | **JSON string → UTF-8 bytes**（可能不兼容） |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | 200 |
| expected_gateway_behavior | 路由匹配 2100:2097 → LocalInvoker → SystemObj.HealthCheck |
| expected_service_behavior | healthSvc.Check(ctx, req) — **req 为 protobuf bytes or JSON bytes 待确认** |
| trace_assertion | 响应 extend 含 traceId；全链路一致 |
| artifact_required | request.raw.json + response.raw.json + trace.log + assertion.json |
| blockedWhen | Health module 未注册 |
| failWhen | HTTP ≠ 200 或业务码 ≠ 10200（D-05：已启用协议必须严格通过） |

**编码风险说明**：Health handler (`healthSvc.Check`) 的入参 `req []byte` 直接来自 `packet.Data`。若 healthSvc 内部使用 `protobuf.Unmarshal` 解析 req，而 proto-tester 发送的是 JSON-string-bytes，将导致 **protobuf 解码失败**。此时按 D-06 关联 E2E-REG-025。

**payload 示例**：
```json
{}
```

---

### E2E-MVP-003：Hello 正向

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-003 |
| name | HelloWorld 正向 — 草案协议，2100:2101 |
| priority | P1 |
| executionOrder | 9 |
| route_key | 2100:2101 |
| command_name | HelloWorld |
| maxType | 2100 |
| minType | 2101 |
| responseMaxType | 2100 |
| responseMinType | 2102 |
| request_proto | com.mineplanet.pojo.hello.HelloWorldRequest |
| response_proto | com.mineplanet.pojo.hello.HelloWorldResponse |
| protocol_status | 草案 |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | Protobuf HelloWorldRequest bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | 200 |
| expected_gateway_behavior | 路由匹配 2100:2101 → LocalInvoker → SystemObj.HelloWorld |
| expected_service_behavior | helloSvc.SayHello(ctx, req) — req 编码格式待确认 |
| trace_assertion | 响应 extend 含 traceId |
| artifact_required | request.raw.json + response.raw.json + assertion.json |
| blockedWhen | Hello module 未实现（D-04：草案 Handler 未实现 = BLOCKED） |
| failWhen | Handler 存在但返回非 10200 |

**payload 示例**：
```json
{"name": "e2e-tester"}
```

---

### E2E-MVP-009：不存在 maxType/minType

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-009 |
| name | 路由未找到 — 不存在的 maxType/minType 组合 |
| priority | P0 |
| executionOrder | 10 |
| route_key | 9999:9999（不存在） |
| command_name | N/A |
| maxType | 9999 |
| minType | 9999 |
| responseMaxType | 0 |
| responseMinType | 0 |
| request_proto | N/A |
| response_proto | N/A |
| protocol_status | N/A |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | N/A |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | 200（Gateway 返回 error Packet，HTTP 层仍为 200） |
| expected_gateway_behavior | http_server.go 第 82-88 行：route not found → BuildErrorPacket(code=CodeNotFound) |
| expected_service_behavior | 不到达 service 层 |
| trace_assertion | 响应 extend.code = CodeNotFound 值 |
| artifact_required | response.raw.json（含 code="not found" 类消息） |
| blockedWhen | 无 |
| failWhen | 响应不含 code 字段或 code ≠ NotFound 对应值 |

**注意**：Gateway 对"路由未找到"的 HTTP 状态码仍为 200（writePacket 不设置非 2xx 状态码），错误信息在 MessagePacket.extend["code"] 中传递。

---

### E2E-MVP-010：traceId 贯穿验证

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-010 |
| name | traceId 全链路贯穿 — 请求→Gateway→响应一致 |
| priority | P1 |
| executionOrder | 11 |
| route_key | 2100:2097 |
| command_name | ServiceHealthCheck |
| maxType | 2100 |
| minType | 2097 |
| responseMaxType | 2100 |
| responseMinType | 2098 |
| request_proto | com.mineplanet.pojo.health.ServiceHealthCheckRequest |
| response_proto | com.mineplanet.pojo.health.ServiceHealthCheckResponse |
| protocol_status | 已启用 |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | Protobuf bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | 200 |
| expected_gateway_behavior | EnsureTraceId 自动生成或透传 traceId |
| expected_service_behavior | extend 中携带 traceId 到 handler |
| trace_assertion | **核心断言**：响应 MessagePacket.extend["traceId"] == 请求中 traceId（或自动生成的 UUID v4 格式） |
| artifact_required | response.raw.json + trace.log（grep 结果或 trace CLI 输出） |
| blockedWhen | 无（DEGRADED 时降级为 grep） |
| failWhen | 响应中无 traceId 或格式非法（D-02：DEGRADED 不判 FAIL） |

**降级方案**（trace CLI exit=4 时）：
```bash
# 从 Gateway 日志中提取 traceId
grep "traceId" /path/to/gateway.log | grep "2100:2097"
```

---

### E2E-MVP-011：CLI 能力验证

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-011 |
| name | CLI run/capture/trace 子命令能力验证 |
| priority | P1 |
| executionOrder | 12 |
| route_key | N/A（CLI 元测试） |
| command_name | N/A |
| maxType | N/A |
| minType | N/A |
| responseMaxType | N/A |
| responseMinType | N/A |
| request_proto | N/A |
| response_proto | N/A |
| protocol_status | N/A |
| auth_required | false |
| token_required | false |
| token_channel | N/A |
| data_encoding_expected | N/A |
| data_encoding_actual_from_proto_tester | N/A |
| content_type | N/A |
| request_body_format | N/A |
| expected_http_status | N/A |
| expected_gateway_behavior | N/A |
| expected_service_behavior | N/A |
| trace_assertion | N/A |
| artifact_required | 各子命令 --help 输出 / 退出码记录 |
| blockedWhen | proto-tester 未编译（需 `npm run cli:build` 或 `tsx` 可用） |
| failWhen | 任一子命令无法执行或退出码异常 |

**子验证项**：

| 子项 | 命令 | 预期退出码 | 预期输出 |
|------|------|-----------|----------|
| MVP-011-a | `proto-tester send --help` | 0 | send 用法文本 |
| MVP-011-b | `proto-tester run --help` | 0 | run 用法文本 |
| MVP-011-c | `proto-tester capture --help` | 0 | capture 用法文本 |
| MVP-011-d | `proto-tester trace --help` | 0 | trace 用法文本 |
| MVP-011-e | `proto-tester trace --id test-001` | **4**（DEGRADED） | 降级提示信息 |
| MVP-011-f | `proto-tester run --suite ./E2E-MVP-SUITE.yaml` | 0 或 1 | suite 执行摘要 |

---

### E2E-MVP-013：Web UI Token 内存化 + 发送验证

| 字段 | 值 |
|------|-----|
| caseId | E2E-MVP-013 |
| name | Web UI Token 内存化输入 + 通过 UI 触发发送验证 |
| priority | P1 |
| executionOrder | 13 |
| route_key | 2100:2097（或 6000:6001） |
| command_name | ServiceHealthCheck（或 GetAppConfigs） |
| maxType | 2100（或 6000） |
| minType | 2097（或 6001） |
| responseMaxType | 2100（或 6000） |
| responseMinType | 2098（或 6002） |
| request_proto | 同上 |
| response_proto | 同上 |
| protocol_status | 已启用（或草案） |
| auth_required | false（或 true） |
| token_required | 取决于目标路由 |
| token_channel | Web UI 内存 → HTTP Authorization Header |
| data_encoding_expected | Protobuf bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| content_type | application/octet-stream |
| request_body_format | Protobuf binary MessagePacket |
| expected_http_status | 200 |
| expected_gateway_behavior | Web UI 调用 apiClient.sendRequest → POST /api/hello |
| expected_service_behavior | 同 MVP-002 或 MVP-004 |
| trace_assertion | UI 显示 traceId 和响应摘要 |
| artifact_required | 截图(video/screenshot) + capture JSON 报告 |
| blockedWhen | vite dev server 未启动（capture 前置依赖） |
| failWhen | UI 无法发送或响应显示异常 |

---

## 二、回归集（精选）

### E2E-REG-025：MessagePacket data 编码一致性

| 字段 | 值 |
|------|-----|
| caseId | E2E-REG-025 |
| name | data 编码一致性 — proto-tester vs handler 期望格式交叉验证 |
| priority | **P0** |
| route_key | 2100:2097 + 6000:6001 |
| data_encoding_expected | 按 message.proto 注释：Protobuf business message bytes |
| data_encoding_actual_from_proto_tester | JSON string → UTF-8 bytes |
| expected_gateway_behavior | 分别对 2100 段(protobuf期望?) 和 6000 段(JSON期望) 发送相同 payload |
| expected_service_behavior | 6000 段 json.Unmarshal 成功；2100 段结果待观察 |
| trace_assertion | 记录两段的实际解码行为差异 |
| blockedWhen | 无 |
| failWhen | 两段均失败且无法归因到编码问题时升级为 P0 Bug |

---

### E2E-REG-026：字段缺失边界

| caseId | name | 缺失字段 | 预期行为 |
|--------|------|----------|----------|
| REG-026-a | 缺失 maxType | maxType=0 或省略 | 400 + "maxType/minType must > 0" |
| REG-026-b | 缺失 minType | minType=0 或省略 | 400 + "maxType/minType must > 0" |
| REG-026-c | 缺失 data | data=[] | 400 + "data is empty" |
| REG-026-d | 空 body | body 长度=0 | 400 + "empty body" |

### E2E-REG-027：字段类型错误

| caseId | name | 错误类型 | 预期行为 |
|--------|------|----------|----------|
| REG-027-a | maxType 为字符串 | 非数字 | protobuf unmarshal 可能失败 → 400 |
| REG-027-b | body 为纯文本 | 非 protobuf 二进制 | "invalid message packet" → 400 |

### E2E-REG-028：下游服务不可达

| caseId | name | 场景 | 预期行为 |
|--------|------|------|----------|
| REG-028-a | LocalInvoker handler 未注册 | 使用未注册的 TargetKey | returnCode=10404 + "local handler not found" |
| REG-028-b | TarsGoInvoker 远程不可达 | tars 模式下网络不通 | returnCode=10500 + "tars invoker is not implemented yet" |

### E2E-REG-029：timeout_ms 超时处理

| caseId | name | timeout 设置 | 预期行为 |
|--------|------|-------------|----------|
| REG-029-a | Health 3s 内返回 | timeout_ms=3000 | 正常 200 |
| REG-029-b | Config 5s 内返回 | timeout_ms=5000 | 正常 200 |
| REG-029-c | 模拟超时场景 | handler sleep > timeout | 待确认是否有超时拦截机制 |

### E2E-REG-030：并发 traceId 不串用

| caseId | name | 场景 | 断言 |
|--------|------|------|------|
| REG-030-a | 并发 3 请求 traceId 不同 | `run --suite --parallel 3` | 每个响应的 traceId 互不相同 |
| REG-030-b | 同一请求 req/res traceId 一致 | 单次 send | 请求和响应 extend 中 traceId 相同 |

### E2E-REG-031：大 payload / 边界 payload

| caseId | name | payload 大小 | 预期行为 |
|--------|------|-------------|----------|
| REG-031-a | 空 payload `{}` | ~2 bytes (JSON) | 正常处理（data 非空即可） |
| REG-031-b | 大 payload 1MB | ~1MB JSON | 待确认 Gateway/Handler 是否有 body size limit |
| REG-031-c | 特殊字符 payload | 含 unicode/emoji/null byte | JSON 序列化后 UTF-8 编码正确性 |

### E2E-REG-032：invalidate 事件兼容性

| caseId | name | 场景 | 说明 |
|--------|------|------|------|
| REG-032-a | 配置变更后版本号递增 | AppConfigVersion 轮询 | version 变更检测 |
| REG-032-b | 语言包增量获取 | GetLangDifference | since_version 参数有效性 |

### E2E-REG-033：TarsGo 条件型 BLOCKED 用例

| caseId | name | blockedReason |
|--------|------|---------------|
| REG-033-a | TarsGo 远程 HealthCheck | TarsGoInvoker.Invoke() 返回 10500 + "not implemented yet" |
| REG-033-b | TarsGo 远程 GetAppConfigs | 同上 |
| REG-033-c | TarsGo 与 Local 双模式一致性对比 | 需 TarsGo 实现后才可执行 |

**标记**：全部 BLOCKED(D-03)，不纳入 MVP 可执行集。

### E2E-REG-034：Method Not Allowed

| caseId | name | Method | 预期 |
|--------|------|--------|------|
| REG-034-a | GET 请求 | GET | 405 Method Not Allowed |
| REG-034-b | PUT 请求 | PUT | 405 Method Not Allowed |

---

## 三、执行顺序总览

```
执行顺序  │ caseId          │ 名称                          │ 优先级 │ 路由        │ Token
──────────┼─────────────────┼───────────────────────────────┼────────┼────────────┼───────
    1     │ E2E-MVP-001     │ Preflight 结果确认             │ P0     │ 元用例      │ —
    2     │ E2E-MVP-012     │ Content-Type 415              │ P0     │ 前置拦截    │ —
    3     │ E2E-MVP-004     │ Config 正向+Token              │ P0     │ 6000:6001   │ ✓
    4     │ E2E-MVP-005     │ Config 缺失Token               │ P0     │ 6000:6001   │ ✗
    5     │ E2E-MVP-006     │ Config 错误Token               │ P0     │ 6000:6001   │ ✗
    6     │ E2E-MVP-007     │ I18n 无Token正向               │ P0     │ 6000:6003   │ —
    7     │ E2E-MVP-008     │ I18n 有Token正向               │ P0     │ 6000:6005   │ ✓
    8     │ E2E-MVP-002     │ Health 正向                    │ P0     │ 2100:2097   │ —
    9     │ E2E-MVP-003     │ Hello 正向                     │ P1     │ 2100:2101   │ —
   10     │ E2E-MVP-009     │ 路由未找到                     │ P0     │ 9999:9999   │ —
   11     │ E2E-MVP-010     │ traceId 贯穿                   │ P1     │ 2100:2097   │ —
   12     │ E2E-MVP-011     │ CLI 能力验证                   │ P1     │ 元测试      │ —
   13     │ E2E-MVP-013     │ Web UI Token+发送              │ P1     │ UI          │ ✓
──────────┼─────────────────┼───────────────────────────────┼────────┼────────────┼───────
         │ E2E-REG-025~034  │ 回归集（14 条）                 │ P0-P1  │ 多路由      │ 混合
```

---

## 四、优先级分布统计

| 优先级 | MVP 数量 | 回归数量 | 合计 |
|--------|----------|----------|------|
| P0 | 8 | 4 | 12 |
| P1 | 5 | 10 | 15 |
| **合计** | **13** | **14** | **27** |

## 五、BlockedWhen / FailWhen 汇总

| caseId | blockedWhen | failWhen |
|--------|-------------|----------|
| MVP-001 | Gateway 无法启动 | 预检结论不可接受 |
| MVP-012 | Gateway 未启动 | HTTP ≠ 415 |
| MVP-004 | ConfigService 未注入 | HTTP ≠ 200 或 业务码 ≠ 10200 |
| MVP-005 | 无 | 取决于 auth 实现（stub→200=PASS, impl→401=PASS） |
| MVP-006 | 无 | 同 MVP-005 |
| MVP-007 | I18nService 未注入 | HTTP ≠ 200 或 业务码 ≠ 10200 |
| MVP-008 | I18nService 未注入 | HTTP ≠ 200 或 业务码 ≠ 10200 |
| MVP-002 | Health module 未注册 | HTTP ≠ 200 或 业务码 ≠ 10200（D-05） |
| MVP-003 | Hello 未实现 | BLOCKED(D-04)；实现后 ≠ 10200 则 FAIL |
| MVP-009 | 无 | 响应无 code 字段 |
| MVP-010 | 无 | DEGRADED 时不 FAIL(D-02) |
| MVP-011 | proto-tester 未编译 | 子命令退出码异常 |
| MVP-013 | vite dev server 未启动 | UI 异常 |
| REG-025 | 无 | 两段均失败且非编码问题 |
| REG-033 | TarsGo 未实现 | BLOCKED(D-03) |

---

*矩阵完毕。配合 YAML suite 文件使用。*
