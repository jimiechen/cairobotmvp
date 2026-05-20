# Proto Gateway

## 职责

Proto Gateway 是 CaiRobot MVP 的单网关入口实现。

- 只暴露 `POST /api/hello`
- Content-Type 为 `application/octet-stream`
- 请求体是 MessagePacket bytes
- 主路由键是 `maxType:minType`
- 不按 URL path 路由业务
- 不以 proto package/service/method 作为主路由
- 不承载复杂业务逻辑
- 不直接访问业务数据库

## 运行模式

Gateway 支持两种运行模式：

### 单体模式（默认）

```bash
GATEWAY_INVOKER_MODE=local
```

- 本地开发、测试、演示使用
- 不依赖真实 TarsCloud
- 通过 LocalInvoker 调用本地业务模块 handler
- 仍然严格走 routes.yaml

### 微服务模式

```bash
GATEWAY_INVOKER_MODE=tars
```

- 正式部署或集成环境使用
- 通过 TarsGoInvoker 调用 TarsCloud 服务
- 当前尚未实现，启动会报错

## 负责

- MessagePacket 解析
- maxType/minType 校验
- routes.yaml 查找
- 协议编号注册表校验
- data 反序列化为业务 Protobuf
- 鉴权上下文处理
- extend 构造
- Protobuf marshal/unmarshal
- Tars bytes 接口调用
- Tars return code 转换
- 响应 MessagePacket 封装
- 日志、trace、metrics

## 不负责

- 具体业务逻辑（由 TarsGo 服务处理）
- 业务数据库访问
- 复杂状态管理

## 目录结构

```text
go/gateway/proto-gateway/
├── README.md
├── go.mod
├── cmd/
│   └── server/
│       └── main.go
├── configs/
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
    │   ├── invoker.go
    │   └── invoker_test.go
    └── server/
        ├── http_server.go
        └── http_server_test.go
```

## 如何运行

```bash
# 单体模式（默认）
cd go/gateway/proto-gateway
go run cmd/server/main.go

# 微服务模式（当前未实现）
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
