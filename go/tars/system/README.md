# SystemServer

## 服务定位

系统服务，提供 HealthCheck 和 HelloWorld 等基础能力。

## Tars 标识

- TarsCloud App: CaiRobot
- Server: SystemServer
- Servant: SystemObj
- Object: CaiRobot.SystemServer.SystemObj
- Module: CaiRobotSystemApp
- Interface: SystemObj
- IDL: tars/protocol/tars/system.tars

## 标准方法签名

所有方法统一使用：

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

## 调用来源

外部请求通过 `POST /api/hello` 进入 Gateway，由 MessagePacket.maxType/minType 命中 routes.yaml 后转发到本服务。

## 当前状态

S0 阶段：只定义骨架和规范，不实现复杂业务逻辑。

## 目录结构

```text
go/tars/system/
├── go.mod
├── cmd/
│   └── main.go
├── internal/
│   └── service/
│       ├── system_service.go
│       └── system_service_test.go
└── localhandler/
    ├── local_handler.go
    └── local_handler_test.go
```

## 模块职责

### internal/service

业务逻辑层，负责：
- HealthCheck：服务健康检查
- HelloWorld：基础问候接口

### localhandler

单体模式下的 bytes 适配层，负责：
- 接收 request bytes 和 extend
- 反序列化 request bytes 为 Protobuf Request
- 调用 internal/service 业务逻辑
- 序列化 Response 为 response bytes
- 返回 return code 和 response bytes

## 运行模式

### 单体模式（Gateway GATEWAY_INVOKER_MODE=local）

Gateway 通过 LocalInvoker 直接调用本模块的 localhandler，不经过 TarsCloud。

```text
Gateway → LocalInvoker → localhandler.Invoke → internal/service → localhandler → Gateway
```

### 微服务模式（Gateway GATEWAY_INVOKER_MODE=tars）

Gateway 通过 TarsGoInvoker 调用 TarsCloud 上的 SystemServer 服务。

```text
Gateway → TarsGoInvoker → TarsCloud → SystemServer → SystemObj → Gateway
```

## 如何测试

```bash
# 独立测试 System 模块
cd go/tars/system && go test ./...

# 所有 Go 模块测试
cd go && bash ../scripts/dev/go-test.sh
```

## Module Path

```
github.com/jimiechen/mineplanet/go/tars/system
```
