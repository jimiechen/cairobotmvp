# Device Gateway Server

## 概述

Device Gateway Server 是 CaiRobot MVP 项目中的设备网关服务，负责设备连接管理、设备状态上报、设备指令下发等功能。

## 职责

- 设备连接管理
- 设备认证
- 设备状态上报
- 设备指令下发
- 设备事件推送
- 设备生命周期管理

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | DeviceGatewayServer |
| Servant | DeviceGatewayObj |

## 目录结构

```
tars/go/device-gateway/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── DeviceGateway.tars     # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── DeviceGatewayObj/
├── cmd/
│   └── server/
│       └── main.go            # 服务启动入口
├── internal/
│   ├── domain/                # 领域层
│   ├── application/          # 应用层
│   ├── infrastructure/       # 基础设施层
│   └── interfaces/           # 接口层
│       └── servant/          # Tars Servant 实现
├── configs/
│   └── config.conf           # Tars 服务配置
└── tests/                    # 测试文件
```

## 当前状态

- ✅ 目录结构已建立
- ⏳ 需要创建 Tars 接口定义文件
- ⏳ 需要实现业务逻辑
- ⏳ 需要编写单元测试

## 相关文档

- [Code Wiki](../../../docs/wiki/CODE-WIKI.md)
- [PRD-06-设备通信与协议](../../../docs/prd/PRD-06-设备通信与协议.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
