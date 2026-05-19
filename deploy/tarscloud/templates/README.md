# deploy/tarscloud/templates

本目录存放 TarsCloud 服务部署模板。

## 1. 文件定位

本目录负责：
- 提供 Tars Server 部署模板示例
- 说明 app / server / servant / module / interface / method 的对应关系

本目录不负责：
- 真实生产部署
- 环境特定的配置值

## 2. 模板规范

每个部署模板必须包含：

- app / server / servant 对应关系
- IDL module / interface 对应关系
- 暴露的方法列表
- 请求/响应类型（统一为 `vector<byte>`）
- 超时配置
- 日志和配置路径

## 3. 模板列表

| 模板文件 | 对应服务 | 对应 .tars 文件 |
|---|---|---|
| user-center-server.template.md | UserCenterServer | user_center.tars |

## 4. 重要说明

模板中的业务方法是服务能力规划；只有协议编号注册表中已登记且写入 routes.yaml 的方法才是当前可路由方法。

## 5. 相关文档

- [../README.md](../README.md)
- [../../docs/api/tars规范.md](../../docs/api/tars规范.md)
- [../../docs/api/协议编号注册表.md](../../docs/api/协议编号注册表.md)
- [../../go/gateway/proto-gateway/configs/routes.yaml](../../go/gateway/proto-gateway/configs/routes.yaml)
