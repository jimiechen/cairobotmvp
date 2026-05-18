# User Center Server

## 概述

User Center Server 是 CaiRobot MVP 项目中的用户中心服务，负责家庭、用户、设备绑定、学习记录等功能。

## 职责

- 用户账号管理
- 家庭空间管理
- 孩子档案管理
- 设备绑定关系管理
- 学习会话记录
- 用户隐私设置

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | UserCenterServer |
| Servant | UserCenterObj |

## 目录结构

```
tars/go/user-center/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── UserCenter.tars         # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── UserCenterObj/
├── cmd/
│   └── server/
│       └── main.go            # 服务启动入口
├── internal/
│   ├── domain/                # 领域层
│   │   ├── model/            # 领域模型
│   │   ├── repository/       # 仓储接口
│   │   └── service/          # 领域服务
│   ├── application/          # 应用层
│   │   ├── usecase/          # 用例编排
│   │   └── dto/              # 数据传输对象
│   ├── infrastructure/       # 基础设施层
│   │   ├── persistence/      # 持久化实现
│   │   ├── cache/            # 缓存实现
│   │   └── external/         # 外部服务调用
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
- [PRD-02-终端用户中台系统](../../../docs/prd/PRD-02-终端用户中台系统.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
