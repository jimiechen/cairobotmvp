# Proto 协议定义

本目录包含 CaiRobot MVP 项目的所有 Protocol Buffers 协议定义。

## 目录结构

```
proto/
├── common/          # 通用定义（错误码、分页、时间等）
├── provider_admin/  # 服务商后台协议
├── user_center/     # 终端用户中台协议
├── open_platform/   # 开放平台协议
├── ai_service/      # AI 服务协议
└── device/          # 设备通信协议
```

## 规范

详见 [docs/api/protobuf规范.md](../docs/api/protobuf规范.md)

## 相关文档

- [ADR-0003-服务协议使用Protobuf.md](../docs/adr/ADR-0003-服务协议使用Protobuf.md)
