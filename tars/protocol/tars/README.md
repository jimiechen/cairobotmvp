# Tars IDL Files

## 说明

本目录存放所有 Tars IDL 文件。所有 `.tars` 文件必须采用统一 bytes 接口，不得定义业务 struct。

## 文件列表

| 文件 | module | interface | 说明 |
|---|---|---|---|
| system.tars | CaiRobotSystemApp | SystemObj | 系统服务 |
| auth.tars | CaiRobotAuthApp | AuthObj | 认证服务 |
| provider_admin.tars | CaiRobotProviderAdminApp | ProviderAdminObj | 服务商后台 |
| user_center.tars | CaiRobotUserCenterApp | UserCenterObj | 用户中心 |
| open_platform.tars | CaiRobotOpenPlatformApp | OpenPlatformObj | 开放平台 |
| ai_bridge.tars | CaiRobotAiBridgeApp | AiBridgeObj | AI 桥接 |
| device_gateway.tars | CaiRobotDeviceGatewayApp | DeviceGatewayObj | 设备网关 |
| audit.tars | CaiRobotAuditApp | AuditObj | 审计服务 |

## 规范

- 每个 interface 都包含 Health 和 HealthCheck
- 所有方法使用统一 bytes 签名：`int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response)`
- 不定义业务 struct

## 相关文档

- [docs/api/tars规范.md](../../../docs/api/tars规范.md)
