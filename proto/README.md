# Proto 协议定义

本目录包含 CaiRobot MVP 项目的所有 Protocol Buffers 协议定义。

## 目录结构

```
proto/
├── base/            # 基础协议（网关入口、通用返回等）
│   ├── message.proto # 网关统一入口 MessagePacket
│   ├── result.proto  # 通用返回 Result、PageInfo、ErrorDetail
│   └── health.proto  # 健康检查协议
├── common/          # 通用定义（错误码、分页、时间等）
├── provider_admin/  # 服务商后台协议
├── user_center/     # 终端用户中台协议
├── open_platform/   # 开放平台协议
├── ai_service/      # AI 服务协议
└── device/          # 设备通信协议
```

## 核心规则

在 CaiRobot MVP 中，Protobuf 协议编号 `max + min` 是接口报文的唯一身份。

1. 每个业务 Request/Response/Event 内部必须声明 enum Type
2. Type.max 表示协议大类
3. Type.min 表示协议小类
4. Type.max + Type.min 必须唯一，并登记到 [协议编号注册表.md](../docs/api/协议编号注册表.md)
5. Protobuf 是协议契约事实来源，OpenAPI 是对外说明

## 规范

详见 [docs/api/protobuf规范.md](../docs/api/protobuf规范.md)

## 相关文档

- [ADR-0003-服务协议使用Protobuf.md](../docs/adr/ADR-0003-服务协议使用Protobuf.md)
- [协议编号注册表.md](../docs/api/协议编号注册表.md)
- [openapi-protobuf映射规范.md](../docs/api/openapi-protobuf映射规范.md)
