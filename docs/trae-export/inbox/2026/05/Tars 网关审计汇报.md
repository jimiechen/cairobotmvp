# Tars 网关审计汇报

> **Workspace:** /Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp

---

## User
先停止继续修改代码。当前主控认为 Tars 网关实现和 tars 目录结构可能没有按项目要求完成。现在进入“只读审计汇报”阶段。 
 
 本次任务只允许读取、检查、运行必要的只读命令和测试命令，不允许改代码、不允许移动文件、不允许提交。 
 
 请按以下要求向主控汇报，让主控判断当前实现是否符合 CaiRobot MVP 架构要求。 
 
 一、必须先读取的规范文件 
 
 请读取并摘取与 Tars/Gateway 相关的关键要求： 
 
 1. AGENTS.md 
 2. docs/wiki/LLM-WIKI.md 
 3. docs/wiki/CODE-WIKI.md 
 4. docs/wiki/工程规范索引.md 
 5. docs/adr/ADR-0008-use-tarscloud-routing-layer.md 
 6. docs/adr/ADR-0012-polyglot-monorepo-directory-layout.md 
 7. docs/adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md 
 8. docs/api/tars规范.md 
 9. docs/api/protobuf规范.md 
 10. docs/api/协议编号注册表.md 
 11. proto/base/message.proto 
 12. configs/gateway/routes.yaml，如存在 
 13. go/go.work 
 14. go/gateway/proto-gateway/README.md，如存在 
 15. go/tars/system/README.md，如存在 
 
 如果文件不存在，必须明确写“缺失”，不能跳过不报。 
 
 二、主控对齐基线 
 
 请按以下基线判断当前实现是否符合要求： 
 
 1. 项目是多语言 monorepo，Go 代码必须位于 `go/` 下。 
 2. `go.work` 必须位于 `go/go.work`，根目录不应放 `go.work`。 
 3. TarsGo 服务 Go 代码应该位于 `go/tars/` 下。 
 4. Gateway Go 代码应该位于 `go/gateway/proto-gateway/` 下。 
 5. 根目录 `tars/` 不应再混放 Go 源码；如果存在，应说明它是 IDL、部署配置、遗留目录还是错误残留。 
 6. 业务入口必须是单网关入口，不应新增多个 REST 业务 path。 
 7. 网关统一入口应使用 `MessagePacket`。 
 8. 路由主键应使用 `Type.max + Type.min`，不是 proto package/service/method。 
 9. 已登记协议编号包括： 
    - 2100:2097 ServiceHealthCheckRequest 
    - 2100:2098 ServiceHealthCheckResponse 
    - 2100:2101 HelloWorldRequest 
    - 2100:2102 HelloWorldResponse 
 10. Tars IDL 不应定义业务 struct 与 Protobuf 一一映射。 
 11. Tars 接口应使用统一 bytes 签名，例如： 
     - request: vector<byte> 
     - extend/context: map<string,string> 
     - response: out vector<byte> 
 12. Gateway 应支持 LocalInvoker 和 TarsGoInvoker。 
 13. `GATEWAY_INVOKER_MODE=tars` 未实现时不能静默回退 local。 
 14. AuditInvoker 应包裹 LocalInvoker 和 TarsGoInvoker。 
 15. 审计应记录 traceId、requestId、maxType、minType、target、method、returnCode、duration、request/response size+hash。 
 16. 审计默认不记录完整 payload。 
 17. token、authorization、password、secret、apiKey、accessToken、refreshToken 必须脱敏。 
 18. TLOG 应使用 TarsGo TLOG，不应引入外部 tlog 包。 
 19. 当前不允许引入第三方 AOP 框架。 
 20. 所有新增测试必须登记到 docs/testing/测试用例注册表.md。 
 
 三、请输出当前实际目录结构 
 
 请执行以下命令并汇报结果： 
 
 ```bash 
 pwd 
 git branch --show-current 
 git status --short 
 find . -maxdepth 3 -type d | sort 
 find go -maxdepth 4 -type f | sort 
 find tars -maxdepth 4 -type f | sort 2>/dev/null || true 
 find configs -maxdepth 4 -type f | sort 2>/dev/null || true 
 find proto -maxdepth 4 -type f | sort

## Assistant
# Tars/Gateway 只读审计汇报

## 一、规范文件读取结果

### 1.1 已读取文件摘要

| 序号 | 文件 | 状态 | 关键要求摘取 |
|---|---|---|---|
| 1 | AGENTS.md | 存在 | S0阶段；TDD红-绿-重构；根目录必备docs/proto/services/ai/web；禁止擅自新增模块/调整目录 |
| 2 | docs/wiki/LLM-WIKI.md | 存在 | 单网关架构；MessagePacket唯一入口；maxType/minType路由；Tars bytes接口；已登记4个协议编号 |
| 3 | docs/wiki/CODE-WIKI.md | 存在 | Gateway在go/gateway/proto-gateway/；TarsGo在go/tars/；tars/protocol/放IDL；routes.yaml以max+min为键；统一bytes签名 |
| 4 | docs/wiki/工程规范索引.md | 存在 | 按语言分层；Go进go/；根目录不保留语言特定构建文件 |
| 5 | docs/adr/ADR-0008-use-tarscloud-routing-layer.md | 存在 | 单网关POST /api/hello；Tars IDL不定义业务struct；Protobuf是唯一契约；routes.yaml只能配置已登记编号 |
| 6 | docs/adr/ADR-0012-polyglot-monorepo-directory-layout.md | 存在 | go.work必须在go/下；根目录禁止放go.work；Go代码进go/；去中心化构建 |
| 7 | docs/adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md | 存在 | 三层Makefile结构；根目录Makefile只做编排；scripts/承载复杂逻辑 |
| 8 | docs/api/tars规范.md | 存在 | Tars IDL不定义业务struct；统一bytes签名；extend标准字段；Health/HealthCheck必须存在；启动校验12项要求 |
| 9 | docs/api/protobuf规范.md | 存在 | max+min唯一身份；每个message必须声明enum Type；MessagePacket定义；协议变更必须同步注册表 |
| 10 | docs/api/协议编号注册表.md | 存在 | 已登记：2100:2097/2098/2101/2102 |
| 11 | proto/base/message.proto | 存在 | MessagePacket含maxType/minType/extend/platform/data |
| 12 | configs/gateway/routes.yaml | **缺失** | 无此文件 |
| 13 | go/go.work | 存在 | 管理gateway/proto-gateway、tars/system、services三个模块 |
| 14 | go/gateway/proto-gateway/README.md | 存在 | 职责清晰；单体/微服务双模式；目录结构说明 |
| 15 | go/tars/system/README.md | 存在 | Tars标识正确；标准bytes方法签名；localhandler职责说明 |

---

## 二、当前实际目录结构

### 2.1 基本信息
- **工作目录**: `/Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp`
- **当前分支**: `chore/monorepo-makefile-entrypoint`
- **Git状态**: 大量未跟踪文件（go/、docs/reports/coverage/、docs/services/ 等）

### 2.2 关键目录结构

```
.
├── go/
│   ├── go.work                    # Workspace 定义
│   ├── go.work.sum
│   ├── Makefile
│   ├── gateway/
│   │   └── proto-gateway/         # Gateway 实现
│   │       ├── README.md
│   │       ├── go.mod
│   │       ├── go.sum
│   │       ├── cmd/server/main.go
│   │       ├── configs/routes.yaml   # ← 路由配置在此，非 configs/gateway/
│   │       └── internal/
│   │           ├── adapter/
│   │           ├── config/
│   │           ├── router/
│   │           ├── server/
│   │           └── tarsclient/
│   ├── services/                  # 遗留/未迁移服务
│   │   ├── go.mod
│   │   ├── main.go
│   │   └── hello_service_test.go
│   ├── shared/                    # 空目录（无文件）
│   ├── tars/
│   │   ├── system/                # System 模块实现
│   │   │   ├── cmd/main.go
│   │   │   ├── go.mod
│   │   │   ├── internal/service/
│   │   │   └── localhandler/
│   │   ├── ai-bridge/ (README.md only)
│   │   ├── audit/ (README.md only)
│   │   ├── auth/ (README.md only)
│   │   ├── device-gateway/ (README.md only)
│   │   ├── open-platform/ (README.md only)
│   │   ├── provider-admin/ (README.md only)
│   │   └── user-center/ (README.md only)
│   └── third_party/               # 空目录
├── tars/
│   └── protocol/                  # Tars IDL 定义
│       ├── tars/
│       │   ├── system.tars
│       │   ├── auth.tars
│       │   ├── audit.tars
│       │   ├── ai_bridge.tars
│       │   ├── device_gateway.tars
│       │   ├── open_platform.tars
│       │   ├── provider_admin.tars
│       │   └── user_center.tars
│       └── proto-adapter/
├── proto/
│   ├── base/
│   │   ├── message.proto
│   │   ├── health.proto
│   │   ├── hello.proto
│   │   └── result.proto
│   └── generated/
│       ├── go/
│       ├── python/
│       ├── ts/
│       └── tarsgo/                # TarsGo 生成代码
├── configs/                       # 缺失（根目录无此目录）
├── docs/testing/测试用例注册表.md   # 存在
└── web/                           # 遗留前端目录（与typescript/并存）
```

---

## 三、按主控基线逐项审计

### 基线1：项目是多语言 monorepo，Go 代码必须位于 `go/` 下
**状态**: ⚠️ **部分不符合**

| 模块 | 位置 | 符合 |
|---|---|---|
| gateway/proto-gateway | go/gateway/proto-gateway | ✅ |
| tars/system | go/tars/system | ✅ |
| services (遗留) | go/services/ | ✅ |
| web (前端) | web/ 和 typescript/web/ 并存 | ⚠️ 重复/遗留 |

**问题**: 根目录存在 `web/` 目录，同时 `typescript/web/` 也存在。根据 ADR-0012，TypeScript 应进 `typescript/`，`web/` 应为遗留目录。

---

### 基线2：`go.work` 必须位于 `go/go.work`，根目录不应放 `go.work`
**状态**: ✅ **符合**

- `go/go.work` 存在
- 根目录无 `go.work`

---

### 基线3：TarsGo 服务 Go 代码应该位于 `go/tars/` 下
**状态**: ✅ **符合**

- `go/tars/system/` 存在且包含实际代码
- 其他服务目录仅有 README.md（S0 阶段骨架正常）

---

### 基线4：Gateway Go 代码应该位于 `go/gateway/proto-gateway/` 下
**状态**: ✅ **符合**

- 代码实际位于 `go/gateway/proto-gateway/`

---

### 基线5：根目录 `tars/` 不应再混放 Go 源码
**状态**: ✅ **符合**

- 根目录 `tars/` 仅包含：
  - `tars/protocol/tars/*.tars`（IDL 定义）
  - `tars/protocol/proto-adapter/`（协议适配文档）
- 无 Go 源码混入

**说明**: 根目录 `tars/` 是 Tars IDL 协议定义目录，与 `go/tars/` 职责分离清晰。

---

### 基线6：业务入口必须是单网关入口，不应新增多个 REST 业务 path
**状态**: ✅ **符合**

- `http_server.go` 仅暴露 `POST /api/hello`
- 无其他业务 path
- 代码中明确校验 Method 和 Content-Type

---

### 基线7：网关统一入口应使用 `MessagePacket`
**状态**: ⚠️ **部分不符合**

- 架构设计使用 `MessagePacket`
- 但 `adapter/message_packet.go` 中 `SerializeMessagePacket` 和 `DeserializeMessagePacket` 是**空实现/临时实现**：
  - `SerializeMessagePacket` 直接返回 `[]byte{}, nil`
  - `DeserializeMessagePacket` 直接返回错误 `"protobuf deserialization not implemented yet"`

**风险**: Gateway 目前无法真正处理 Protobuf 序列化的 MessagePacket，只能跑通测试中的 mock 路径。

---

### 基线8：路由主键应使用 `Type.max + Type.min`，不是 proto package/service/method
**状态**: ✅ **符合**

- `router.go` 使用 `fmt.Sprintf("%d:%d", maxType, minType)` 作为 key
- `routes.yaml` 中 `route_key` 格式为 `"2100:2097"`
- 代码明确不以 proto package/service/method 路由

---

### 基线9：已登记协议编号检查
**状态**: ✅ **符合**

协议编号注册表已登记：
- 2100:2097 ServiceHealthCheckRequest
- 2100:2098 ServiceHealthCheckResponse
- 2100:2101 HelloWorldRequest
- 2100:2102 HelloWorldResponse

`routes.yaml` 中配置的 request_max/request_min 均在注册表中。

---

### 基线10：Tars IDL 不应定义业务 struct 与 Protobuf 一一映射
**状态**: ✅ **符合**

- `tars/protocol/tars/system.tars` 仅定义统一 bytes 方法签名
- 无业务 struct 定义（如 BindDeviceReq/LoginReq 等）
- 注释中仅说明对应的 Protobuf message，无字段映射

---

### 基线11：Tars 接口应使用统一 bytes 签名
**状态**: ✅ **符合**

- `system.tars` 中所有方法签名：
  ```
  int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
  ```
- `routes.yaml` 中 `tars_request_type` 和 `tars_response_type` 均为 `vector<byte>`

---

### 基线12：Gateway 应支持 LocalInvoker 和 TarsGoInvoker
**状态**: ⚠️ **部分不符合**

| Invoker | 状态 | 说明 |
|---|---|---|
| LocalInvoker | ✅ 已实现 | `local_invoker.go` 完整实现，含注册/调用逻辑 |
| TarsGoInvoker | ⚠️ 占位实现 | 仅返回错误 `"tars invoker is not implemented yet"` |

---

### 基线13：`GATEWAY_INVOKER_MODE=tars` 未实现时不能静默回退 local
**状态**: ✅ **符合**

- `http_server.go` 中：
  ```go
  if mode == "tars" {
      return fmt.Errorf("tars invoker is not implemented yet")
  }
  ```
- 明确报错，不静默回退

---

### 基线14：AuditInvoker 应包裹 LocalInvoker 和 TarsGoInvoker
**状态**: ❌ **缺失**

- 代码中无 `AuditInvoker` 实现
- `invoker.go` 中只有 `LocalInvoker` 和 `TarsGoInvoker`，无审计包裹层

---

### 基线15：审计应记录 traceId、requestId、maxType、minType、target、method、returnCode、duration、request/response size+hash
**状态**: ❌ **缺失**

- 无审计逻辑实现
- 无 duration 记录
- 无 request/response size+hash 记录

---

### 基线16：审计默认不记录完整 payload
**状态**: ❌ **缺失**

- 无审计实现，此要求自然未满足

---

### 基线17：token、authorization、password、secret、apiKey、accessToken、refreshToken 必须脱敏
**状态**: ❌ **缺失**

- 无审计/日志脱敏逻辑
- `BuildTarsExtend` 中直接透传 extend 字段，无脱敏处理

---

### 基线18：TLOG 应使用 TarsGo TLOG，不应引入外部 tlog 包
**状态**: ⚠️ **待确认**

- 当前代码中**未使用任何 tlog 包**
- 也未使用 TarsGo TLOG
- 属于"尚未实现"而非"错误引入"

---

### 基线19：当前不允许引入第三方 AOP 框架
**状态**: ✅ **符合**

- 代码中无 AOP 框架引入
- 依赖仅 `github.com/google/uuid` 和 `gopkg.in/yaml.v3`

---

### 基线20：所有新增测试必须登记到 docs/testing/测试用例注册表.md
**状态**: ⚠️ **部分不符合**

- 测试用例注册表存在
- 但注册表中的**测试函数名称与实际代码不一致**：

| 注册表登记 | 实际代码 | 状态 |
|---|---|---|
| TC-GW-0001 TestMessagePacketEncodeDecode | 实际为 `TestBuildErrorPacket` 等 | ❌ 不符 |
| TC-GW-0002 TestRoutesLoad | 实际为 `TestLoadRoutes` | ⚠️ 近似 |
| TC-GW-0003 TestRouterMatch | 实际为 `TestFindRoute` | ❌ 不符 |
| TC-GW-0004 TestHTTPServerStart | 实际为 `TestGatewayServer_ServeHTTP` | ❌ 不符 |
| TC-GW-0005 TestTarsClientInvoke | 实际为 `TestLocalInvoker` / `TestTarsGoInvoker` | ❌ 不符 |
| TC-TARS-0001 TestSystemGetStatus | 实际为 `TestSystemService_HealthCheck` | ❌ 不符 |
| TC-TARS-0002 TestLocalHandlerProcess | 实际为 `TestHandler_Invoke` | ❌ 不符 |

**问题**: 注册表与实现不同步，存在"登记了但函数名对不上"的问题。

---

## 四、额外发现的问题

### 4.1 routes.yaml 位置不符合规范预期
- **规范预期**: `configs/gateway/routes.yaml`（根目录 configs/ 下）
- **实际位置**: `go/gateway/proto-gateway/configs/routes.yaml`
- **说明**: ADR-0012 规定根目录保留 `configs/` 作为工程级公共资产，但当前路由配置放在 gateway 模块内部

### 4.2 go.work 包含未迁移的 services 模块
```go
use (
    ./gateway/proto-gateway
    ./tars/system
    ./services          // ← go/services/ 是遗留模块
)
```
- `go/services/` 包含 `hello_service_test.go` 和 `main.go`
- 该模块 path 为 `github.com/jimiechen/mineplanet/go/services/hello-service`
- 与 ADR-0012 的 module path 规范（应带 `/go` 前缀）不一致，且是遗留代码

### 4.3 go/shared/ 目录为空
- CODE-WIKI 规划了 `go/shared/audit/`、`go/shared/config/` 等
- 实际为空目录，无文件

### 4.4 MessagePacket 序列化未实现
- `SerializeMessagePacket` 和 `DeserializeMessagePacket` 是空实现
- 导致 Gateway 无法真正处理请求（HTTP 测试中使用 mock 绕过）

### 4.5 System 模块未注册到 Gateway
- `http_server.go` 中：
  ```go
  invoker = tarsclient.NewLocalInvoker()
  // TODO: 注册 System handler
  ```
- LocalInvoker 创建后未注册任何 handler，Gateway 启动后无法实际调用 System 服务

### 4.6 LocalHandler 使用 JSON 而非 Protobuf
- `local_handler.go` 中：
  ```go
  resp := fmt.Sprintf(`{"status":"%s","timestamp":%d}`, status, timestamp)
  ```
- 返回的是 JSON 字符串，不是 Protobuf 序列化 bytes
- 与架构要求的 "Tars bytes 承载 Protobuf 序列化结果" 不符

---

## 五、审计结论汇总

| 类别 | 项目数 | 符合 | 部分符合 | 不符合 | 缺失 |
|---|---|---|---|---|---|
| 目录结构 | 5 | 4 | 1 | 0 | 0 |
| 架构设计 | 6 | 4 | 2 | 0 | 0 |
| 协议规范 | 4 | 4 | 0 | 0 | 0 |
| 实现完整性 | 5 | 2 | 1 | 0 | 2 |

### 关键风险项（必须修改）

| 优先级 | 问题 | 风险等级 |
|---|---|---|
| P0 | MessagePacket 序列化/反序列化未实现 | Gateway 无法运行 |
| P0 | System handler 未注册到 Gateway | 端到端链路不通 |
| P1 | AuditInvoker 缺失 | 审计能力缺失 |
| P1 | 审计字段/脱敏未实现 | 安全合规风险 |
| P1 | 测试用例注册表与实际代码不同步 | 可追溯性缺失 |
| P2 | routes.yaml 位置与规范预期不一致 | 目录结构规范 |
| P2 | go/services/ 遗留模块未清理 | 模块管理 |
| P2 | LocalHandler 返回 JSON 而非 Protobuf | 协议一致性 |

### 建议主控决策

1. **是否接受当前 S0 阶段的骨架实现**（大量 TODO 和空实现）？
2. **routes.yaml 位置**是否从 `go/gateway/proto-gateway/configs/` 迁移到根目录 `configs/gateway/`？
3. **go/services/ 遗留模块**是否删除或迁移？
4. **MessagePacket 序列化实现**是否优先排期（阻塞 Gateway 实际运行）？
5. **AuditInvoker 和审计脱敏**是否在 S1 阶段实现，还是 S0 补齐？

