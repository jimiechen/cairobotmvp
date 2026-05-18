# UserCenterServer 部署模板

本文件是 TarsCloud 部署模板示例，说明 UserCenterServer 在 TarsCloud 上的部署配置。

## 1. 服务对应关系

| 层级 | 名称 | 说明 |
|---|---|---|
| App | CaiRobot | TarsCloud App 统一名称 |
| Server | UserCenterServer | 用户中心服务 |
| Servant | UserCenterObj | 用户中心 servant |
| IDL module | CaiRobotUserCenterApp | Tars IDL module |
| IDL interface | UserCenterObj | Tars IDL interface |

## 2. 部署配置模板

```yaml
app: CaiRobot
server: UserCenterServer
module: CaiRobotUserCenterApp
servants:
  - name: UserCenterObj
    interface: UserCenterObj
    protocol: tars
    port: 10001
    methods:
      - Health
      - HealthCheck
      - BindDevice
      - UnbindDevice
      - ListFamilyDevices
      - ListLearningRecords
    request_type: vector<byte>
    response_type: vector<byte>
    timeout_ms: 3000
node: local
log_path: /data/logs/cairobot/user-center
config_path: /data/configs/cairobot/user-center
```

## 3. 方法说明

| 方法 | 说明 | 协议编号状态 |
|---|---|---|
| Health | 轻量健康探测 | 基础接口，无业务编号 |
| HealthCheck | 完整健康检查 | 基础接口，无业务编号 |
| BindDevice | 绑定设备 | 待登记 |
| UnbindDevice | 解绑设备 | 待登记 |
| ListFamilyDevices | 列出家庭设备 | 待登记 |
| ListLearningRecords | 列出学习记录 | 待登记 |

## 4. 重要约束

- 所有方法请求/响应类型统一为 `vector<byte>`
- 业务字段由 Protobuf message 定义，Tars IDL 不定义业务 struct
- 只有协议编号注册表中已登记且写入 routes.yaml 的方法才是当前可路由方法
- 未登记协议编号的方法仅为服务能力规划，不可通过网关路由

## 5. 相关文件

- [tars/protocol/tars/user_center.tars](../../../tars/protocol/tars/user_center.tars)
- [docs/api/协议编号注册表.md](../../../docs/api/协议编号注册表.md)
- [gateway/proto-gateway/configs/routes.yaml](../../../gateway/proto-gateway/configs/routes.yaml)
