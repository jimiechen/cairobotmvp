# deploy/tarscloud

本目录存放 CaiRobot MVP 在 TarsCloud 上的部署说明、配置模板和部署脚本。

## 1. 文件定位

本目录不负责：
- 业务代码实现
- 具体 CI/CD pipeline
- 容器镜像构建（由各自服务目录负责）

本目录负责：
- TarsCloud App / Server / Servant 的部署规范
- 部署配置模板
- 部署目录结构说明
- 各服务在 TarsCloud 上的注册关系

## 2. 目录结构

```text
deploy/tarscloud/
├── README.md                          # 本文件
├── configs/
│   └── README.md                      # 部署配置说明
└── templates/
    ├── README.md                      # 模板使用说明
    └── user-center-server.template.md # UserCenterServer 部署模板示例
```

## 3. TarsCloud 部署规范

### 3.1 命名规范

| 层级 | 命名规则 | 示例 |
|---|---|---|
| App | 统一为 `CaiRobot` | CaiRobot |
| Server | 使用 `XxxServer` | SystemServer、AuthServer |
| Servant | 使用 `XxxObj` | SystemObj、AuthObj |
| IDL module | 使用 `CaiRobotXxxApp` | CaiRobotSystemApp |
| IDL interface | 使用 `XxxObj` | SystemObj、AuthObj |

### 3.2 每个 Servant 必须暴露的接口

每个 Tars servant 必须实现：

- `Health`：轻量健康探测
- `HealthCheck`：完整健康检查

所有业务方法统一使用 bytes 签名：

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

### 3.3 部署模板内容

每个部署模板必须说明：

- app / server / servant 对应关系
- module / interface / method 对应关系
- 端口配置
- 超时配置
- 日志路径
- 配置路径

## 4. 相关文档

- [docs/api/tars规范.md](../../docs/api/tars规范.md)
- [docs/adr/ADR-0008-use-tarscloud-routing-layer.md](../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
- [tars/protocol/tars/](../../tars/protocol/tars/)
