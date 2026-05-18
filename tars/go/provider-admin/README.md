# Provider Admin Server

## 概述

Provider Admin Server 是 CaiRobot MVP 项目中的服务商后台服务，负责服务商账号管理、设备批次管理、工单管理等功能。

## 职责

- 服务商账号管理
- 租户/渠道管理
- 设备批次管理
- 设备绑定与解绑管理
- 用户服务状态查看
- 售后工单管理
- 设备异常查看
- 服务套餐或权益管理
- 运营数据看板
- 系统配置管理

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | ProviderAdminServer |
| Servant | ProviderAdminObj |

## 目录结构

```
tars/go/provider-admin/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── ProviderAdmin.tars      # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── ProviderAdminObj/
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
- [PRD-01-服务商后台系统](../../../docs/prd/PRD-01-服务商后台系统.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
