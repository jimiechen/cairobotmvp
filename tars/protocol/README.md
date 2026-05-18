# Tars Protocol

## 职责

本目录存放 Tars IDL 文件和 Protobuf 适配说明。

## 目录结构

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

## 规范

- 所有 `.tars` 文件必须采用统一 bytes 接口
- 不得定义业务 struct
- 每个 interface 都包含 Health 和 HealthCheck
- module 使用 `CaiRobotXxxApp`
- interface 使用 `XxxObj`
- 方法名使用 UpperCamelCase

## 相关文档

- [docs/api/tars规范.md](../../docs/api/tars规范.md)
