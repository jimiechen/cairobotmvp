# AI Bridge Server

## 概述

AI Bridge Server 是 CaiRobot MVP 项目中的 AI 桥接服务，负责与 Python AI Service 进行通信，封装意图分类、提示词改写、回答审核等功能。

## 职责

- 与 Python AI Service 通信
- 意图分类请求封装
- 提示词改写请求封装
- 回答审核请求封装
- OCR 结果理解请求封装
- 模型网关请求封装

## Tars 信息

| 配置项 | 值 |
|--------|-----|
| App | CaiRobot |
| Server | AiBridgeServer |
| Servant | AiBridgeObj |

## 目录结构

```
tars/go/ai-bridge/
├── README.md                    # 本文档
├── tars/                       # Tars 定义文件
│   └── AiBridge.tars          # Tars 接口定义
├── go/                        # Tars 生成代码
│   └── AiBridgeObj/
├── cmd/
│   └── server/
│       └── main.go            # 服务启动入口
├── internal/
│   ├── domain/                # 领域层
│   ├── application/          # 应用层
│   ├── infrastructure/       # 基础设施层
│   │   └── external/         # 外部服务调用（Python AI Service）
│   └── interfaces/           # 接口层
│       └── servant/          # Tars Servant 实现
├── configs/
│   └── config.conf           # Tars 服务配置
└── tests/                    # 测试文件
```

## 当前状态

- ✅ 目录结构已建立
- ⏳ 需要创建 Tars 接口定义文件
- ⏳ 需要实现与 Python AI Service 的通信
- ⏳ 需要实现业务逻辑
- ⏳ 需要编写单元测试

## 相关文档

- [Code Wiki](../../../docs/wiki/CODE-WIKI.md)
- [PRD-04-AI服务系统](../../../docs/prd/PRD-04-AI服务系统.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
- [AI Service](../../../ai/service/README.md)
