# 鉴权通道分析（E2E Auth Channel）

> **生成时间**：2026-06-05
> **阶段**：Phase 1 — 静态分析，不执行请求
> **关联文件**：[E2E-CASE-MATRIX.md](./E2E-CASE-Matrix.md) MVP-004/005/006

---

## 一、结论速查

| 问题 | 结论 |
|------|------|
| Gateway 是否读取 Authorization Header？ | **否** — http_server.go 中无 Header 读取代码 |
| Gateway 是否读取 MessagePacket.extend["token"]？ | **否** — 仅透传 extend 到 BuildTarsExtend，不做校验 |
| proto-tester --token 写入位置？ | **HTTP Authorization Header only**（apiClient.ts L119） |
| auth_required=true 路由当前是否拦截无 Token 请求？ | **否** — 当前为 stub，不拦截 |
| Token 校验逻辑实现状态 | **未实现（STUB）** |

---

## 二、源码证据链

### 2.1 Gateway 端：http_server.go（无 token 读取）

**文件**：[http_server.go](../../../go/gateway/proto-gateway/internal/server/http_server.go)

ServeHTTP 方法（第 33-114 行）的完整处理流程：

```
1. 检查 Method == POST                          (L34-38)
2. 检查 Content-Type == octet-stream             (L40-44)
3. 读取 Body                                     (L46-58)
4. 反序列化 MessagePacket                         (L60-65)
5. 检查 maxType > 0 && minType > 0               (L67-71)
6. 检查 len(data) > 0                            (L73-77)
7. 路由匹配 FindRoute                             (L82-88)
8. 构建 extend → Invoke                           (L92-113)
```

**关键发现**：整个 ServeHTTP 方法中：
- **无** `r.Header.Get("Authorization")` 调用
- **无** `packet.Extend["token"]` 读取或校验
- **无** 任何 if/switch 分支处理 auth_required 字段
- `auth_required` 仅通过 `route.AuthRequired` 传递给 [BuildTarsExtend](../../../go/gateway/proto-gateway/internal/adapter/message_packet.go#L88)（L105），写入 extend map 作为元数据

### 2.2 proto-tester 端：token 传递路径

**文件**：[apiClient.ts](../../../typescript/proto-tester/src/lib/apiClient.ts)

```typescript
// 第 117-120 行：token 仅写入 HTTP Header
headers: {
  'Content-Type': 'application/octet-stream',
  ...(req.token ? { Authorization: `Bearer ${req.token}` } : {}),
},
```

**文件**：[send.ts](../../../typescript/proto-tester/src/cli/commands/send.ts)

```typescript
// 第 81-87 行：--token 参数传递给 sendRequest
const response = await sendRequest({
  maxType,
  minType,
  payload: packetBinary,
  gatewayUrl: opts.gateway,
  token: opts.token,          // ← 来自 --token 命令行参数
});
```

**文件**：[runner.ts](../../../typescript/proto-tester/src/cli/runner.ts)

```typescript
// 第 128-134 行：suite 级别 token 传递给每个 case
const response = await sendRequest({
  maxType: tc.protocol[0],
  minType: tc.protocol[1],
  payload: packetBinary,
  gatewayUrl: opts.gateway,
  token: opts.token,          // ← 来自 run --token 全局参数
});
```

### 2.3 LocalInvoker Handler 层：无 token 检查

**文件**：[invoker.go](../../../go/gateway/proto-gateway/tarsclient/invoker.go) RegisterConfigI18nHandlers（第 229-327 行）

全部 5 个 handler 的模式一致：
```go
func(ctx context.Context, req []byte) ([]byte, error) {
    var appReq SomeStruct
    if err := json.Unmarshal(req, &appReq); err != nil {  // 只解码 data
        return nil, err
    }
    resp, err := someService.Method(&appReq)              // 直接调用业务
    // ...
}
```

**关键发现**：handler 不读取 `extend` map，因此即使 token 在 extend 中也不会被检查。

---

## 三、Token 传递完整链路图

```
用户输入
  │
  ▼
proto-tester send --token "eyJhbGci..."   或   run --suite X.yaml --token "eyJhbGci..."
  │
  ├─► apiClient.sendRequest({ token: "eyJhbGci..." })
  │     │
  │     ▼
  │   HTTP Request Headers:
  │     Authorization: Bearer eyJhbGci...    ◄── token 在这里
  │     Content-Type: application/octet-stream
  │
  │   HTTP Body:
  │     Protobuf MessagePacket {
  │       maxType: 6000, minType: 6001,
  │       data: JSON-string-bytes,          ◄── token 不在这里
  │       extend: { method, traceId, requestId }  ◄── token 也不在这里
  │     }
  │
  ▼
Gateway http_server.ServeHTTP()
  │
  ├─► r.Header.Get("Authorization")          ◄── 【未调用】token 在这里但无人读取
  ├─► r.Body → DeserializeMessagePacket()     ◄── 只解包外层 protobuf
  ├─► route.AuthRequired                      ◄── 读取 routes.yaml 元数据
  ├─► BuildTarsExtend(..., authRequired, ...) ◄── 写入 extend["authRequired"]="true"
  │                                             但不做任何 if 判断
  └─► invoker.Invoke(ctx, target, data, extend)
        │
        ▼
     LocalInvoker / TarsGoInvoker
        │
        ▼
     Handler (json.Unmarshal(req) → Service.Method())
        │                                    ◄── 【不读 extend】token 完全不可见
        ▼
     Response → BuildResponsePacket() → writePacket()
```

---

## 四、各场景预期行为

### 4.1 正向用例（有效 Token）

| 项目 | 值 |
|------|-----|
| 用例 | E2E-MVP-004, E2E-MVP-008 |
| 命令 | `send --max 6000 --min 6001 --token "valid-jwt"` |
| HTTP Header | `Authorization: Bearer valid-jwt` |
| Gateway 行为 | **忽略** Header，正常路由到 handler |
| Handler 行为 | json.Unmarshal(data) → 业务处理 |
| 预期结果 | HTTP 200 + businessCode=10200 |
| **实际鉴权效果** | **无** — Token 被发送但不被校验 |

### 4.2 负向用例 A：缺失 Token

| 项目 | 值 |
|------|-----|
| 用例 | E2E-MVP-005 |
| 命令 | `send --max 6000 --min 6001`（不传 --token） |
| HTTP Header | 无 Authorization |
| Gateway 行为 | **不检查**，正常路由到 handler |
| 预期结果 | HTTP 200 + businessCode=10200（与正向相同） |
| **判定** | 若返回 200 → 记录 "**AUTH-STUB**: 鉴权未实现" ，**不判 FAIL** |

### 4.3 负向用例 B：错误 Token

| 项目 | 值 |
|------|-----|
| 用例 | E2E-MVP-006 |
| 命令 | `send --max 6000 --min 6001 --token "invalid-token"` |
| HTTP Header | `Authorization: Bearer invalid-token` |
| Gateway 行为 | **不检查**，正常路由到 handler |
| 预期结果 | HTTP 200 + businessCode=10200（与正向相同） |
| **判定** | 同上，记录 AUTH-STUB |

### 4.4 负向用例 C：空 Token

| 项目 | 值 |
|------|-----|
| 命令 | `send --max 6000 --min 6001 --token ""` |
| HTTP Header | `Authorization: Bearer ` （空值） |
| Gateway 行为 | **不检查** |
| 预期结果 | HTTP 200 + businessCode=10200 |
| **判定** | 同上 |

---

## 五、预期响应码汇总表

| 场景 | Token 状态 | 当前实际预期 | 鉴权实现后预期 | 差异说明 |
|------|-----------|-------------|---------------|----------|
| 有效 Token | 有效 JWT | 200 / 10200 | 200 / 10200 | 一致 |
| 缺失 Token | 未传 | 200 / 10200 | **401** / 10401 或业务错误码 | **需实现后变更** |
| 错误 Token | 无效字符串 | 200 / 10200 | **401** / 10401 | **需实现后变更** |
| 空 Token | 空字符串 | 200 / 10200 | **400** / 10400 或 **401** | **需实现后变更** |

---

## 六、对 E2E 执行的影响和应对策略

### 6.1 MVP-005 / MVP-006 判定规则

```
IF 鉴权已实现（Gateway 返回非 200 或 businessCode ≠ 10200）:
    → 按 expected_http_status 判定 PASS/FAIL

IF 鉴权未实现（STUB）（Gateway 返回 200 / 10200）:
    → 判定 PASS with WARNING "AUTH-STUB: token validation not implemented"
    → 在 RUN 报告中标记 auth_required 路由的鉴权状态为 STUB
    → 不阻塞 MVP 验收，但作为 P1 遗留项记录
```

### 6.2 Phase 2 执行时应记录的信息

每条 auth_required=true 路由的请求必须记录：

1. 是否携带了 Authorization Header
2. Gateway 日志中是否有 auth 相关输出
3. 返回的 HTTP 状态码和业务码
4. 最终判定：AUTH-IMPLEMENTED / AUTH-STUB

### 6.3 建议的后续动作

| 优先级 | 动作 | 说明 |
|--------|------|------|
| P0 | 在 http_server.go ServeHTTP 中增加 token 读取和校验 | 路由匹配后、Invoke 前，根据 route.AuthRequired 决定是否拦截 |
| P1 | 定义 Token 格式规范（JWT? API Key? 自定义?） | 影响校验逻辑实现 |
| P1 | proto-tester 增加 per-case token 支持 | runner.ts 当前仅支持 suite 级别 token |
| P2 | 考虑 extend["token"] 作为备用通道 | 部分 App 端可能将 token 放入 MessagePacket.extend |

---

*分析完毕。当前结论：**鉴权逻辑为 STUB 状态，auth_required 仅作为元数据传递，不影响路由和调用。***
