# Proto Gateway

## 概述

Proto Gateway 是 CaiRobot MVP 项目中负责 Protobuf 请求解析、路由转发和协议适配的核心组件。它作为外部请求和内部 TarsGo 服务之间的桥梁，统一处理协议转换、上下文透传、服务发现和负载均衡。

## 职责

- **协议解析**：解析 Protobuf 请求，识别 package、service、method
- **路由查找**：根据路由配置查找目标 TarsGo 服务
- **协议适配**：Protobuf 与 Tars 协议之间的转换
- **上下文透传**：透传 trace-id、auth-token 等上下文信息
- **服务发现**：与 TarsCloud 集成，发现可用服务实例
- **负载均衡**：选择合适的服务实例进行调用
- **超时控制**：设置合理的超时时间，防止请求堆积
- **错误码转换**：将内部 Tars 错误转换为统一的 Protobuf 响应

## 非职责

- **不承载复杂业务逻辑**：业务逻辑由内部 TarsGo 服务处理
- **不直接访问业务数据库**：数据访问由业务服务处理
- **不绕过 AuthService/AuditService**：认证和审计由专门的服务处理
- **不修改 Protobuf 契约语义**：保持协议定义的一致性

## 目录结构

```
gateway/proto-gateway/
├── README.md                    # 本文档
├── cmd/
│   └── server/                  # 服务启动入口
│       └── main.go              # 主程序文件
├── internal/
│   ├── router/                  # 路由模块
│   │   ├── router.go            # 路由查找逻辑
│   │   └── config.go            # 路由配置加载
│   ├── adapter/                 # 协议适配模块
│   │   ├── proto_adapter.go     # Protobuf与Tars协议转换
│   │   └── result_adapter.go    # 结果转换
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # 鉴权中间件
│   │   ├── audit.go             # 审计中间件
│   │   └── tracing.go           # 链路追踪中间件
│   ├── tarsclient/              # Tars客户端封装
│   │   ├── client.go            # Tars客户端
│   │   └── pool.go              # 连接池
│   └── server/                  # 服务端
│       ├── grpc.go              # gRPC服务
│       └── http.go              # HTTP服务（可选）
├── configs/
│   └── routes.yaml              # 路由配置文件
└── api/                         # 生成的API代码
```

## 配置说明

路由配置文件 `configs/routes.yaml` 定义了 Protobuf 请求到 TarsGo 服务的映射关系。

详细配置说明见 [configs/routes.yaml](./configs/routes.yaml) 中的注释。

## 当前状态

- ✅ 目录结构已建立
- ⏳ 路由配置示例已创建
- ⏳ 需要实现路由查找逻辑
- ⏳ 需要实现协议转换逻辑
- ⏳ 需要实现 Tars 客户端集成
- ⏳ 需要实现中间件

## 后续工作

1. 实现路由加载和查找逻辑
2. 实现 Protobuf 与 Tars 协议转换
3. 集成 TarsGo 客户端
4. 实现鉴权、审计、链路追踪中间件
5. 编写单元测试
6. 集成 CI/CD 流程

## 相关文档

- [Code Wiki](../../docs/wiki/CODE-WIKI.md)
- [ADR-0008-使用TarsCloud作为Protobuf到TarsGo的内部路由转发层](../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
- [protobuf规范](../../docs/api/protobuf规范.md)
