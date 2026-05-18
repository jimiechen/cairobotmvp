# deploy/tarscloud/configs

本目录存放 TarsCloud 部署相关的配置文件说明。

## 1. 文件定位

本目录负责：
- TarsCloud 部署配置规范说明
- 配置项定义和约束

本目录不负责：
- 真实生产环境的敏感配置（如密码、密钥）
- 具体环境的 IP、端口清单

## 2. 配置规范

### 2.1 Tars 服务配置

每个 Tars Server 部署时必须包含：

| 配置项 | 说明 | 必填 |
|---|---|---|
| app | TarsCloud App 名称，统一为 `CaiRobot` | 是 |
| server | Server 名称，如 `UserCenterServer` | 是 |
| servant | Servant 名称，如 `UserCenterObj` | 是 |
| protocol | 协议类型，统一为 `tars` | 是 |
| port | 服务监听端口 | 是 |
| timeout | 调用超时时间（毫秒） | 是 |
| log_path | 日志存放路径 | 是 |
| config_path | 配置文件路径 | 是 |

### 2.2 环境区分

真实环境配置按以下方式管理：

- 开发环境：`configs/dev/`
- 测试环境：`configs/test/`
- 预发布环境：`configs/staging/`
- 生产环境：`configs/prod/`

当前 S0 阶段只提供配置规范说明，不存放真实环境配置。

## 3. 相关文档

- [../README.md](../README.md)
- [../../docs/api/tars规范.md](../../docs/api/tars规范.md)
