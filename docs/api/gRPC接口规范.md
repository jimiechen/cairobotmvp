# gRPC 接口规范

> 状态说明：根据 ADR-0008，CaiRobot MVP 内部核心服务调用主链路已调整为 TarsCloud/TarsGo。本文件保留为历史设计与兼容参考，不再作为内部核心服务调用的主规范。新的内部 RPC 标准见 [docs/api/tars规范.md](tars规范.md)。

## 1. 服务定义

- 历史规划中内部服务间通信使用 gRPC；根据 ADR-0008，当前内部核心服务主链路调整为 TarsCloud/TarsGo。
- 服务定义放在 proto/[service]/v1/ 目录下
- 使用 proto3 语法

## 2. 接口设计原则

- 单一职责：每个 RPC 方法只做一件事
- 幂等性：读操作必须幂等，写操作尽量幂等
- 批量操作：支持批量操作以减少 round-trip
- 分页：列表操作必须支持分页

## 3. 错误处理

- 使用 `google.rpc.Status` 或自定义 Error 类型
- 错误码必须明确
- 错误信息必须可读

## 4. 元数据

- 使用 gRPC metadata 传递认证信息
- 使用 gRPC metadata 传递 trace-id
- 使用 gRPC metadata 传递 tenant-id（如有）

## 5. 超时和重试

- 客户端必须设置合理的超时
- 服务端必须处理超时
- 幂等操作可以重试

## 6. 相关文档

- [protobuf规范.md](protobuf规范.md)
- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)
