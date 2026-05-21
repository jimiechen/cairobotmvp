# Proto Gateway

## 职责

Proto Gateway 是 CaiRobot MVP 的 **TarsGo HTTP Servant**，作为单网关入口实现。

- 基于 TarsCloud/TarsGo v1.4.6 框架构建
- 使用 `tars.TarsHttpMux` + `tars.AddHttpServant` + `tars.Run` 承载 HTTP 入口
- 只暴露 `POST /api/hello`
- Content-Type 为 `application/octet-stream`
- 请求体是 MessagePacket bytes
- 主路由键是 `maxType:minType`
- 不按 URL path 路由业务
- 不以 proto package/service/method 作为主路由
- 不承载复杂业务逻辑
- 不直接访问业务数据库

## 技术基线

| 依赖 | 版本 | 位置 |
|---|---|---|
| TarsCloud/TarsGo | v1.4.6 | `go/third_party/TarsGo/TarsGo-1.4.6/`（replace 指向） |

## 运行模式

Proto Gateway 基于同一套 TarsGo 技术基线支持两种**部署拓扑**：

### 单体部署模式（默认）

```bash
GATEWAY_INVOKER_MODE=local
```

- 本地开发、测试、演示使用
- **使用 TarsGo 框架运行**（TarsHttpMux / AddHttpServant / Run）
- **不连接远程 TarsCloud 注册中心**（locator 为空），但不是不依赖 TarsGo
- 通过 **LocalInvoker**（本进程 TarsGo servant adapter）调用同部署单元内的业务 servant
- 所有 TarsGo servant（SystemServer 等）在同一进程或同一部署单元中
- 仍然严格走 routes.yaml
- 严格遵守 Tars bytes 契约：request/response 均为 Protobuf bytes

### 微服务部署模式

```bash
GATEWAY_INVOKER_MODE=tars
```

- 正式部署或集成环境使用
- **使用 TarsGo 框架运行**（与单体模式相同的技术基线）
- 连接远程 TarsCloud 注册中心，通过 **TarsGoInvoker**（远程 TarsGo client）调用独立部署的 TarsCloud servant
- GatewayServer、SystemServer 等独立部署为不同进程
- 当前 **TarsGoInvoker 远程调用尚未实现**，启动会报错（S1 阶段）

### 两种模式的共同点

无论单体还是微服务，都必须遵守：

- Tars bytes 统一方法签名：`int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response)`
- MessagePacket 作为唯一入口报文
- maxType:minType 主路由键
- Protobuf bytes 作为 data 承载
- routes.yaml 路由配置
- 协议编号注册表校验
- 项目统一状态码体系

区别仅在于**部署拓扑**：servant 在同进程 vs 跨进程远程调用。

## 负责

- MessagePacket 解析
- maxType/minType 校验
- routes.yaml 查找
- 协议编号注册表校验
- data 反序列化为业务 Protobuf
- 鉴权上下文处理
- extend 构造
- Protobuf marshal/unmarshal
- Tars bytes 接口调用（通过 TarsInvoker 接口）
- Tars return code 转换
- 响应 MessagePacket 封装
- 日志、trace、metrics

## 不负责

- 具体业务逻辑（由 TarsGo servant 处理）
- 业务数据库访问
- 复杂状态管理

## 目录结构

```text
go/gateway/proto-gateway/
├── README.md
├── go.mod                          # 引入 github.com/TarsCloud/TarsGo v1.4.6
├── cmd/
│   └── server/
│       └── main.go                 # TarsGo 入口：TarsHttpMux + AddHttpServant + Run
├── configs/
│   ├── gateway/
│   │   └── gateway.local.conf      # TarsGo 单体部署本地配置（locator 为空）
│   └── routes.yaml
└── internal/
    ├── config/
    │   ├── routes.go
    │   └── routes_test.go
    ├── router/
    │   ├── router.go
    │   └── router_test.go
    ├── adapter/
    │   ├── message_packet.go
    │   └── message_packet_test.go
    ├── tarsclient/
    │   ├── invoker.go              # TarsInvoker 接口 + LocalInvoker + TarsGoInvoker
    │   └── invoker_test.go
    └── server/
        ├── http_server.go          # GatewayServer（http.Handler，注册到 TarsHttpMux）
        └── http_server_test.go
```

## 如何运行

```bash
# 单体部署模式（默认，使用 TarsGo 框架，不连接远程 TarsCloud）
cd go/gateway/proto-gateway
go run cmd/server/main.go

# 微服务模式（当前未实现，S1 阶段）
GATEWAY_INVOKER_MODE=tars go run cmd/server/main.go
```

## 如何测试

```bash
# Gateway 测试
cd go/gateway/proto-gateway && go test ./...

# System 模块测试
cd go/tars/system && go test ./...

# 所有 Go 模块测试
cd go && bash ../scripts/dev/go-test.sh
```

## Module Path

```
github.com/jimiechen/mineplanet/go/gateway/proto-gateway
```

## 相关文档

- [docs/api/tars规范.md](../../docs/api/tars规范.md)
- [docs/api/http-gateway规范.md](../../docs/api/http-gateway规范.md)
- [docs/adr/ADR-0008-use-tarscloud-routing-layer.md](../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
- [go/third_party/TarsGo/README.md](../third_party/TarsGo/README.md)
