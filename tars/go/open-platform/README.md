# Open Platform Server

## 概述

Open Platform Server 是 CaiRobot MVP 项目中的开放平台服务，负责第三方API管理、认证、配额、Webhook管理等功能。

## 职责

- 开放平台认证
- API Key / AppSecret 管理
- API 调用配额管理
- API 调用日志
- Webhook 回调管理
- 第三方开发者管理

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | OpenPlatformServer |
| Servant | OpenPlatformObj |

## 目录结构

```
tars/go/open-platform/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── OpenPlatform.tars       # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── OpenPlatformObj/
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
- [PRD-03-开放平台API](../../../docs/prd/PRD-03-开放平台API.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
