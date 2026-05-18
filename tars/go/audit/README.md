# Audit Server

## 概述

Audit Server 是 CaiRobot MVP 项目中的审计服务，负责操作日志记录、审计追踪等功能。

## 职责

- 操作日志记录
- 审计追踪
- 审计报告生成
- 审计数据查询
- 日志归档

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | AuditServer |
| Servant | AuditObj |

## 目录结构

```
tars/go/audit/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── Audit.tars             # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── AuditObj/
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
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
