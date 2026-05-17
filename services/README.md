# Services

CaiRobot MVP 后端服务集合。

## 服务列表

| 服务名称 | 目录 | 说明 |
|---------|------|------|
| 服务商后台服务 | provider-admin/ | 服务商管理、设备批次、售后工单等 |
| 用户中台服务 | user-center/ | 用户账号、家庭、设备绑定、学习记录等 |
| 开放平台服务 | open-platform/ | 开放 API 认证、配额、日志等 |
| 设备网关服务 | device-gateway/ | 设备连接、通信、控制等 |
| API 网关服务 | api-gateway/ | 统一 API 入口、路由、限流等 |
| 认证服务 | auth-service/ | 统一认证、授权、Token 管理 |
| 审计服务 | audit-service/ | 操作日志、审计追踪 |

## 技术栈

- 语言：Golang
- 通信：gRPC + Protobuf
- Web：Gin / Echo（待定）
- ORM：待定

## 目录结构规范

每个服务遵循以下目录结构：

```
services/[service-name]/
├── README.md          # 服务说明
├── cmd/               # 入口程序
│   └── server/        # 服务启动入口
├── internal/          # 内部代码
│   ├── domain/        # 领域层（模型、仓储接口）
│   ├── application/   # 应用层（用例编排）
│   ├── infrastructure/ # 基础设施层（数据库、缓存、外部服务）
│   └── interfaces/    # 接口层（gRPC、HTTP handler）
├── api/               # API 定义（生成的代码）
├── configs/           # 配置文件
├── tests/             # 测试文件
└── go.mod             # Go 模块定义
```

## 相关文档

- [ADR-0002-后端使用Golang.md](../docs/adr/ADR-0002-后端使用Golang.md)
- [docs/api/](../docs/api/)
- [proto/](../proto/)
