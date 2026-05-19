# Tars 规范

## 1. 核心原则

TarsCloud/TarsGo 是内部 RPC 与服务治理层，不是外部 API 协议。

- 外部业务入口只有 `POST /api/hello`
- 外部入口报文是 `MessagePacket`
- 业务字段结构由 Protobuf message 定义
- 协议身份由 Protobuf message 内部 `Type.max + Type.min` 定义
- Tars IDL 不重复定义业务字段
- Tars 方法统一使用 `vector<byte> request`、`map<string,string> extend`、`out vector<byte> response`

## 2. 与单网关 MessagePacket 的关系

```text
Client → POST /api/hello → MessagePacket → Gateway → routes.yaml → TarsGo servant
```

- Gateway 解析 MessagePacket，提取 `maxType/minType`
- Gateway 使用 `maxType:minType` 查询 routes.yaml
- Gateway 根据路由目标调用 TarsGo servant
- TarsGo 服务处理业务逻辑，返回 Protobuf Response bytes
- Gateway 封装响应 MessagePacket 返回客户端

## 3. 与 Protobuf 规范的关系

- Protobuf 定义业务字段结构和协议编号
- Tars IDL 不定义业务 struct，只定义统一方法签名
- Tars 方法中的 `vector<byte> request/response` 承载 Protobuf 序列化结果
- `Type.max + Type.min` 是接口报文唯一身份，由 Protobuf message 内部 enum Type 声明

## 4. 与协议编号注册表的关系

- routes.yaml 中的 `request_max/request_min` 必须存在于协议编号注册表
- routes.yaml 中的 `response_max/response_min` 必须存在于协议编号注册表
- 未登记协议编号不得写入真实 routes.yaml
- Tars 方法名与协议编号无直接绑定关系，由 routes.yaml 映射

## 5. 与 routes.yaml 的关系

routes.yaml 负责把 `request_max/request_min` 映射到内部 Tars 目标：

```yaml
routes:
  - request_max: 2100
    request_min: 2097
    route_key: "2100:2097"
    request_proto: com.mineplanet.pojo.health.ServiceHealthCheckRequest
    response_max: 2100
    response_min: 2098
    response_proto: com.mineplanet.pojo.health.ServiceHealthCheckResponse
    tars_app: CaiRobot
    tars_server: SystemServer
    tars_servant: SystemObj
    tars_module: CaiRobotSystemApp
    tars_interface: SystemObj
    tars_method: HealthCheck
    tars_request_type: vector<byte>
    tars_response_type: vector<byte>
```

## 6. TarsCloud App / Server / Servant 命名规范

### 6.1 TarsCloud App

统一为：

```text
CaiRobot
```

### 6.2 Tars Server

使用 UpperCamelCase，并以 Server 结尾：

```text
SystemServer
AuthServer
ProviderAdminServer
UserCenterServer
OpenPlatformServer
AiBridgeServer
DeviceGatewayServer
AuditServer
```

### 6.3 Tars Servant

使用 UpperCamelCase，并以 Obj 结尾：

```text
SystemObj
AuthObj
ProviderAdminObj
UserCenterObj
OpenPlatformObj
AiBridgeObj
DeviceGatewayObj
AuditObj
```

完整对象名：

```text
CaiRobot.UserCenterServer.UserCenterObj
CaiRobot.AuthServer.AuthObj
CaiRobot.DeviceGatewayServer.DeviceGatewayObj
```

## 7. Tars IDL module / interface / method 命名规范

### 7.1 module

使用：

```text
CaiRobot[Domain]App
```

示例：

```text
CaiRobotSystemApp
CaiRobotAuthApp
CaiRobotProviderAdminApp
CaiRobotUserCenterApp
CaiRobotOpenPlatformApp
CaiRobotAiBridgeApp
CaiRobotDeviceGatewayApp
CaiRobotAuditApp
```

### 7.2 interface

使用 Servant 名，统一以 Obj 结尾：

```text
SystemObj
AuthObj
ProviderAdminObj
UserCenterObj
OpenPlatformObj
AiBridgeObj
DeviceGatewayObj
AuditObj
```

### 7.3 method

使用 UpperCamelCase：

```text
Health
HealthCheck
Login
BindDevice
SendCommand
```

## 8. 统一 Tars bytes 方法签名

所有 Tars 方法必须使用：

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

含义：
- `request`：Protobuf Request message 序列化 bytes
- `extend`：上下文透传 map
- `response`：Protobuf Response message 序列化 bytes
- `return int`：Tars 方法返回状态码

禁止使用业务 Tars struct：

```text
BindDeviceReq
BindDeviceRsp
LoginReq
LoginRsp
SendCommandReq
SendCommandRsp
RequestContext
TarsResult
```

## 9. Health / HealthCheck 基础接口

每个 interface 都必须包含：

```tars
int Health(vector<byte> request, map<string,string> extend, out vector<byte> response);
int HealthCheck(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

说明：
- Health：轻量健康探测
- HealthCheck：完整健康检查，可检查依赖组件
- 二者都使用统一 bytes 签名
- 返回码遵守项目统一状态码

## 10. extend map 标准字段

| key | 说明 | 来源 |
|---|---|---|
| traceId | 链路追踪 ID | 客户端或 Gateway |
| requestId | 请求唯一 ID | 客户端或 Gateway |
| token | Token | MessagePacket.extend 或 Header |
| caller | 调用方 | Gateway |
| clientIp | 客户端 IP | Gateway |
| userId | 用户 ID | AuthServer 校验结果 |
| tenantId | 租户 / 服务商 ID | AuthServer 校验结果 |
| locale | 语言区域 | Header 或 MessagePacket.extend |
| maxType | 请求 maxType | MessagePacket |
| minType | 请求 minType | MessagePacket |
| requestProto | 请求 Protobuf 类型 | routes.yaml |
| responseProto | 响应 Protobuf 类型 | routes.yaml |
| authRequired | 是否需要鉴权 | routes.yaml |
| auditRequired | 是否需要审计 | routes.yaml |
| platform | 平台类型 | MessagePacket.platform |

要求：
- key 使用 lowerCamelCase
- value 统一为 string
- 客户端传入的 userId / tenantId 不可信
- userId / tenantId 必须来自 AuthServer 校验结果
- traceId / requestId 必须全链路透传

## 11. Tars return 与 Result.code 分层规则

Tars 方法 return 与 Protobuf Result.code 使用同一项目状态码体系，但处于不同层级。

- Tars return 表示本次内部 Tars 方法调用的处理状态，用于 Gateway 判断调用是否成功、是否需要转换错误
- Protobuf Response message 内部的 Result.code 表示业务响应状态
- Gateway 对外返回 MessagePacket 时，应在 MessagePacket.extend 中写入 code/message，优先来自业务 Response.Result；如果 Tars 调用失败或无法解析业务 Response，则使用 Tars return 或 Gateway 错误映射结果

规则：
1. Tars return = 10200 且 Response.Result 存在时，对外 code 优先使用 Response.Result.code
2. Tars return = 10200 但 Response.Result 不存在时，对外 code 使用 10200
3. Tars return != 10200 时，Gateway 不直接透出底层异常，按统一错误码写入 MessagePacket.extend.code/message
4. Tars 框架异常（如超时、服务不存在、调用失败）由 Gateway 映射为 10504、10500 等项目统一错误码
5. 不再使用 `code=0` 表示成功

基础状态码：

| code | 含义 |
|---:|---|
| 10200 | 成功 |
| 10400 | 请求参数错误 |
| 10401 | 无权限 |
| 10404 | 资源不存在 |
| 10429 | 请求过于频繁 |
| 10500 | 失败 / 内部错误 |
| 10504 | 上游超时 / Tars 调用超时 |

## 12. Tars 错误处理规范

- TarsCloud 框架异常、节点地址、servant 路径、panic 堆栈不得返回给客户端
- 内部日志可以记录详细错误
- 对外 message 必须稳定、安全、可文档化
- Gateway 负责将 Tars 异常转换为项目统一错误码

## 13. Tars IDL 文件目录规范

```text
tars/protocol/
├── README.md
├── tars/
│   ├── README.md
│   ├── system.tars
│   ├── auth.tars
│   ├── provider_admin.tars
│   ├── user_center.tars
│   ├── open_platform.tars
│   ├── ai_bridge.tars
│   ├── device_gateway.tars
│   └── audit.tars
└── proto-adapter/
    └── README.md
```

## 14. TarsGo 服务目录规范

```text
tars/go/
├── system/
│   └── README.md
├── auth/
│   └── README.md
├── provider-admin/
│   └── README.md
├── user-center/
│   └── README.md
├── open-platform/
│   └── README.md
├── ai-bridge/
│   └── README.md
├── device-gateway/
│   └── README.md
└── audit/
    └── README.md
```

## 15. Gateway 调用 Tars 完整流程

### 15.1 微服务模式流程

1. 客户端请求 `POST /api/hello`
2. 请求体是 `MessagePacket` bytes
3. Gateway 反序列化 MessagePacket
4. Gateway 校验 maxType/minType/data
5. Gateway 使用 `maxType:minType` 查询 routes.yaml
6. Gateway 校验协议编号是否已登记
7. Gateway 根据 `request_proto` 将 MessagePacket.data 反序列化为业务 Protobuf Request
8. Gateway 执行基础参数校验
9. Gateway 根据 `auth_required` 判断是否需要鉴权
10. Gateway 合并 MessagePacket.extend、Header、鉴权结果、路由元信息，构造 Tars extend
11. Gateway 将业务 Protobuf Request 重新序列化为 `vector<byte> request`
12. Gateway 根据路由目标调用 Tars：

```text
CaiRobot.SystemServer.SystemObj.HealthCheck
```

13. TarsGo 服务反序列化 request bytes 为 Protobuf Request
14. TarsGo 服务执行业务逻辑
15. TarsGo 服务构造 Protobuf Response
16. TarsGo 服务序列化 Response 为 response bytes
17. TarsGo 方法 return 项目统一状态码
18. Gateway 接收 return code 和 response bytes
19. Gateway 根据 `response_proto` 反序列化 response bytes
20. Gateway 使用 `response_max/response_min` 封装响应 MessagePacket
21. Gateway 在响应 MessagePacket.extend 写入 code/message/traceId/requestId
22. 返回客户端

### 15.2 单体模式流程

单体模式（`GATEWAY_INVOKER_MODE=local`）不走真实 TarsCloud，流程差异：

1-11 步与微服务模式相同
12. Gateway 通过 `LocalInvoker` 查找本地注册的 handler
13. Local handler 接收 request bytes 和 extend
14. Local handler 反序列化 request bytes 为 Protobuf Request
15. Local handler 调用本地 SystemService 执行业务逻辑
16. Local handler 构造 Protobuf Response
17. Local handler 序列化 Response 为 response bytes
18. Local handler 返回 return code 和 response bytes 给 Gateway
19-22 步与微服务模式相同

单体模式下，System 模块通过 `localhandler` 包对外暴露 bytes 接口，内部调用 `internal/service` 业务逻辑。

## 16. Gateway 运行模式

Gateway 支持通过环境变量切换运行模式：

### 16.1 单体模式（默认）

```bash
GATEWAY_INVOKER_MODE=local
```

- 本地开发、测试、演示使用
- 不依赖真实 TarsCloud
- 通过 `LocalInvoker` 调用本地业务模块 handler
- 仍然严格走 routes.yaml
- System 模块通过 `localhandler` 包暴露 bytes 接口

### 16.2 微服务模式

```bash
GATEWAY_INVOKER_MODE=tars
```

- 正式部署或集成环境使用
- 通过 `TarsGoInvoker` 调用 TarsCloud 服务
- 当前尚未实现，启动会报错

### 16.3 TarsInvoker 接口

统一调用接口：

```go
type TarsInvoker interface {
    Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}
```

实现：
- `LocalInvoker`：单体模式，本地 handler 注册表
- `TarsGoInvoker`：微服务模式，调用 TarsCloud（未实现）

## 17. 启动校验要求

Gateway 启动时必须校验：

1. request_max/request_min 不重复
2. route_key 与 request_max/request_min 一致
3. request_proto 存在于协议编号注册表
4. response_proto 存在于协议编号注册表
5. response_max/response_min 存在于协议编号注册表
6. request_proto 的 Type.max/min 与 request_max/request_min 一致
7. response_proto 的 Type.max/min 与 response_max/response_min 一致
8. tars_app/tars_server/tars_servant/tars_module/tars_interface/tars_method 不为空
9. tars_request_type 必须是 vector<byte>
10. tars_response_type 必须是 vector<byte>
11. tars_method 必须存在于对应 .tars 文件 interface 中
12. 未登记协议编号不得启动成功

## 18. 测试要求

### 18.1 Gateway 测试

| 测试项 | 要求 |
|---|---|
| MessagePacket 解析测试 | 验证 application/octet-stream 请求体可反序列化 |
| maxType/minType 路由测试 | 验证 route_key 正确命中 |
| routes.yaml 重复路由测试 | 重复 request_max/request_min 应启动失败 |
| 协议编号注册表一致性测试 | routes.yaml 中编号必须存在于注册表 |
| request_proto/response_proto 一致性测试 | proto 类型与 Type.max/min 一致 |
| .tars 方法存在性测试 | tars_method 必须存在于 interface |
| Tars bytes 调用测试 | request bytes / response bytes 完整透传 |
| Tars return 映射测试 | 10200/10401/10500/10504 正确映射 |
| extend 透传测试 | traceId/requestId/userId/tenantId 正确传递 |
| LocalInvoker 注册测试 | handler 正确注册和调用 |
| TarsGoInvoker 未实现测试 | 微服务模式启动应报错 |

### 18.2 System 模块测试

| 测试项 | 要求 |
|---|---|
| SystemService 业务逻辑测试 | HealthCheck / HelloWorld 返回正确 |
| LocalHandler bytes 适配测试 | 请求 bytes 正确反序列化，响应正确序列化 |
| LocalHandler 错误处理测试 | 非法 bytes / 未知 maxType/minType 返回错误 |

## 19. 与 ADR-0008 的关系

本规范是 ADR-0008 的技术实现细则。ADR-0008 决策了 TarsCloud/TarsGo 作为内部 RPC 与服务治理层，本规范定义了具体的命名、接口、流程和校验规则。

## 20. Module Path 规范

所有 Go module 使用统一 path 前缀：

```text
github.com/jimiechen/mineplanet/go/...
```

当前模块：

| 模块 | Path |
|---|---|
| gateway/proto-gateway | github.com/jimiechen/mineplanet/go/gateway/proto-gateway |
| tars/system | github.com/jimiechen/mineplanet/go/tars/system |

Go Workspace 位于 `go/go.work`，执行 Go 命令前需 `cd go/`。

## 21. 相关文档

- [ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)
- [protobuf规范.md](protobuf规范.md)
- [协议编号注册表.md](协议编号注册表.md)
- [http-gateway规范.md](http-gateway规范.md)
