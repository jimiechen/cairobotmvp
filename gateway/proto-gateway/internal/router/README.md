# Router

## 职责

Router 负责根据 MessagePacket.maxType/minType 查询 routes.yaml，返回路由目标信息。

## 路由主键

- 路由主键是 `request_max:request_min`
- 来源是 `MessagePacket.maxType/minType`

## 路由结果

路由结果包含：

- request_proto
- response_proto
- response_max/response_min
- tars_app
- tars_server
- tars_servant
- tars_module
- tars_interface
- tars_method
- tars_request_type（固定为 vector<byte>）
- tars_response_type（固定为 vector<byte>）

## 启动校验

- 启动时检查重复路由（request_max/request_min 不重复）
- 启动时检查协议编号注册表一致性
- 启动时检查 tars_method 是否存在于 .tars 文件 interface 中

## 相关文档

- [docs/api/tars规范.md](../../../docs/api/tars规范.md)
