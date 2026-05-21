# ADR-0014: MessagePacket.data 使用 Protobuf binary bytes

## 状态

已接受（2026-05-20）

## 背景

CaiRobot MVP Gateway 需要在客户端和业务模块之间传递数据。MessagePacket 作为 Gateway envelope，其 data 字段的格式选择直接影响：

1. Gateway 编解码逻辑复杂度
2. 业务模块接口设计
3. 前端对接方式
4. Tars 微服务兼容性
5. 调试和可观测性

## 决策

**MessagePacket.data 固定使用 Protobuf binary bytes，不采用 JSON payload。**

### 核心规则

```text
MessagePacket.data = proto.Marshal(ProtobufMessage) → []byte
```

### 数据流

```
Client HTTP Request
    ↓
Gateway 解析 MessagePacket envelope
    ↓
Gateway 提取 data 字段（[]byte，不解码）
    ↓
LocalInvoker / TarsGoInvoker 透传 data 给业务模块
    ↓
业务模块接收 []byte，proto.Unmarshal 为具体 message
    ↓
业务模块处理，返回 proto.Marshal(resp)
    ↓
Gateway 将 resp bytes 写回 MessagePacket.data
    ↓
返回 Client
```

### 关键约束

1. **Gateway 不做 JSON 编解码**
   - Gateway 只负责透传 Protobuf bytes
   - 不引入 JSON 序列化依赖
   
2. **业务模块接收 Protobuf bytes**
   - 模块接口签名：`func(ctx, requestBytes []byte) ([]byte, error)`
   - 模块内部自行 Unmarshal/Marshal
   
3. **测试构造使用 Protobuf**
   - 单元测试：直接构造 Protobuf message 并 Marshal
   - 集成测试：构造完整 MessagePacket（data = proto.Marshal(req)）

## 理由

### 选择 Protobuf 的原因

| # | 理由 | 说明 |
|---|:---|---|
| 1 | **唯一协议契约** | CaiRobot MVP 当前所有协议定义均使用 Protobuf |
| 2 | **协议编号绑定** | maxType + minType 已绑定到 Protobuf message 定义 |
| 3 | **Tars 天然兼容** | Tars bytes 接口原生承载 Protobuf bytes |
| 4 | **实现已收敛** | 当前 Gateway/LocalInvoker/SystemService 已向 Protobuf bytes 收口 |
| 5 | **避免二义性** | 引入 JSON 会产生第二套语义，增加维护成本 |

### 否决 JSON 的原因

| # | 风险 | 说明 |
|---|:---|---|
| 1 | **双重序列化** | Protobuf → JSON → Protobuf 的转换开销 |
| 2 | **Schema 演进弱** | JSON 缺乏编译时类型检查 |
| 3 | **调试误导** | 可读性高但掩盖了真实的协议边界 |
| 4 | **前端耦合** | 强制前端理解内部协议细节 |
| 5 | **MVP 阶段过度工程** | 当前阶段无 JSON 必要场景 |

## 后果

### 正面影响

1. **架构一致性**
   - 全链路统一使用 Protobuf
   - 无需维护两套序列化逻辑
   
2. **性能优势**
   - 二进制体积小（比 JSON 小 30-50%）
   - 序列化/反序列化速度快
   
3. **类型安全**
   - 编译时检查字段类型
   - Proto 文件即文档
   
4. **Tars 兼容性**
   - 直接对接 TarsGo 服务端
   - 无需额外的适配层

### 负面影响

1. **调试复杂度**
   - 需要工具查看 Protobuf bytes（不能直接读）
   - 缓解方案：日志中记录关键字段、使用 protoc --decode
   
2. **前端对接**
   - 前端需要 protobuf.js 或后端提供 JSON 转换层
   - 缓解方案：Gateway 提供 debug 端点（可选，非必须）
   
3. **测试编写**
   - 测试代码需要 import protobuf 包
   - 缓解方案：封装 TestHelper 简化测试构造

### 兼容性影响

| 组件 | 影响 | 应对措施 |
|---|---|---|
| Gateway HTTP Server | 无需修改 | 已正确透传 bytes |
| LocalInvoker | 无需修改 | 已正确调用 handler |
| SystemService | 无需修改 | 已返回 proto.Marshal() |
| LocalHandler | 无需修改 | 已透传 bytes |
| TarsGoInvoker | 待实现 | S1 阶段处理 |
| 前端 App | 需适配 | 使用 protobuf.js 或转换 API |

## 替代方案

### 方案 A：JSON payload（已否决）

**描述：** MessagePacket.data 使用 UTF-8 JSON 字符串

**优点：**
- 可读性强，调试方便
- 前端原生支持
- 无需额外工具

**缺点：**
- 与现有 Protobuf 协议体系冲突
- 需要维护两套 Schema（Proto + JSON DTO）
- 性能较差
- 类型安全性降低

**否决理由：** 主控明确否决，MVP 阶段不接受第二套协议语义

---

### 方案 B：混合模式（部分场景 JSON）（已否决）

**描述：** 根据路由选择 JSON 或 Protobuf

**优点：** 灵活适应不同场景

**缺点：**
- 增加复杂性
- 难以预测行为
- 维护成本高

**否决理由：** 违背"单一协议契约"原则

---

## 未来演进

### S0 阶段（当前）✅

- [x] MessagePacket.data = Protobuf bytes
- [x] local 模式闭环验证通过
- [x] 所有测试使用 Protobuf 构造

### S1 阶段（规划中）

- [ ] 实现 TarsGoInvoker 真实远程调用
- [ ] 验证跨进程 Protobuf 传输
- [ ] 补充网络异常测试

### S2 阶段（远期）

- [ ] 评估是否需要 JSON debug 端点
- [ ] 评估是否需要 GraphQL/REST adapter
- [ ] 基于 S1 实际数据决定

## 相关决策

- [ADR-0003: 服务协议使用 Protobuf](../ADR-0003-%E6%9C%8D%E5%8A%A1%E5%8D%8F%E8%AE%AE%E4%BD%BF%E7%94%A8Protobuf.md)
- [ADR-0012: 多语言 Monorepo 目录布局](../ADR-0012-polyglot-monorepo-directory-layout.md)

## 参考证据

### 代码证据

**SystemService 返回 Protobuf bytes：**
[go/tars/system/internal/service/system_service.go#L23-L33](../../go/tars/system/internal/service/system_service.go#L23-L33)

```go
func (s *SystemService) HealthCheck(ctx context.Context, serviceName string) ([]byte, error) {
    resp := &pb.ServiceHealthCheckResponse{...}
    return proto.Marshal(resp)  // 返回 Protobuf bytes
}
```

**Gateway 透传 data：**
[go/gateway/proto-gateway/internal/server/http_server.go#L94](../../go/gateway/proto-gateway/internal/server/http_server.go#L94)

```go
returnCode, responseBytes, err := gs.invoker.Invoke(r.Context(), target, packet.Data, extend)
// packet.Data 是 []byte，直接透传
```

**E2E 测试构造 Protobuf data：**
[go/gateway/proto-gateway/internal/server/http_server_test.go#L160-L167](../../go/gateway/proto-gateway/internal/server/http_server_test.go#L160-L167)

```go
healthReq := &pb.ServiceHealthCheckRequest{ServiceName: "gateway-e2e-test"}
reqData, _ := proto.Marshal(healthReq)  // 序列化为 Protobuf bytes
reqPacket.Data = reqData
```

### 测试证据

```
Gateway proto-gateway: 40/40 子测试 PASS ✅
Tars System:          7/7 子测试 PASS   ✅
总计：                47/47 子测试 PASS (100%) ✅
```

## 决策记录

| 属性 | 值 |
|---|---|
| **决策日期** | 2026-05-20 |
| **决策人** | 项目主控 |
| **决策状态** | 已接受 |
| **Source Record** | SRC-TRAE-20260520-143000-gateway-s0-local-closure |
| **关联 Issue** | （待创建） |
| **评审报告** | [TabAI会话_1779258912884.md](../../docs/tabbit/inbox/2026/05/TabAI会话_1779258912884.md) |

## 附录

### A. Protobuf vs JSON 对比矩阵

| 维度 | Protobuf Binary | JSON |
|---|:---|:---|
| **体积** | 小（Varint 编码） | 大（字段名重复） |
| **速度** | 快（二进制操作） | 慢（字符串解析） |
| **可读性** | 低（需工具） | 高（可直接读） |
| **类型安全** | 高（编译时检查） | 低（运行时错误） |
| **Schema 演进** | 强（向前兼容） | 弱（易破坏） |
| **工具生态** | 成熟（protoc） | 极其成熟 |
| **Tars 兼容性** | **原生支持** | 需要转换层 |
| **前端支持** | 需要 protobuf.js | 原生支持 |
| **调试体验** | 较差 | 优秀 |

### B. FAQ

**Q1: 为什么不提供 JSON debug 端点？**

A: MVP 阶段优先保证架构简洁。如果后续确实需要，可在 S2 阶段评估添加 `/debug/decode` 端点。

**Q2: 前端如何对接？**

A: 前端有两种选择：
1. 使用 protobuf.js 直接处理 Protobuf
2. 通过 Gateway 提供的 REST adapter（如需要，S2 阶段实现）

**Q3: 是否永远不能用 JSON？**

A: 本 ADR 锁定的是 **MVP 阶段**的决策。S2 阶段可根据实际数据重新评估。
