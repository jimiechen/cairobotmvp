# TarsCloud 部署

## 概述

本目录包含 CaiRobot MVP 项目的 TarsCloud 相关部署配置和模板文件。

## 目录结构

```
deploy/tarscloud/
├── README.md                # 本文档
├── apps/                    # 应用配置
│   └── CaiRobot/            # CaiRobot 应用配置
├── servers/                 # 服务配置
│   ├── UserCenterServer/
│   ├── ProviderAdminServer/
│   ├── OpenPlatformServer/
│   ├── DeviceGatewayServer/
│   ├── AuthServer/
│   ├── AuditServer/
│   └── AiBridgeServer/
├── templates/               # 部署模板
└── configs/                 # 配置文件
```

## 服务列表

| 服务名 | 说明 |
|--------|------|
| UserCenterServer | 用户中心服务 |
| ProviderAdminServer | 服务商后台服务 |
| OpenPlatformServer | 开放平台服务 |
| DeviceGatewayServer | 设备网关服务 |
| AuthServer | 认证服务 |
| AuditServer | 审计服务 |
| AiBridgeServer | AI 桥接服务 |

## 当前状态

- ✅ 目录结构已建立
- ⏳ 需要添加服务配置模板
- ⏳ 需要添加部署脚本
- ⏳ 需要添加部署文档

## 相关文档

- [Code Wiki](../../docs/wiki/CODE-WIKI.md)
- [ADR-0008](../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
