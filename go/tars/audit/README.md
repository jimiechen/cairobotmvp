# AuditServer

## 服务定位

审计服务，提供审计日志写入、查询等审计能力。

## Tars 标识

- TarsCloud App: CaiRobot
- Server: AuditServer
- Servant: AuditObj
- Object: CaiRobot.AuditServer.AuditObj
- Module: CaiRobotAuditApp
- Interface: AuditObj
- IDL: tars/protocol/tars/audit.tars

## 标准方法签名

所有方法统一使用：

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

## 调用来源

外部请求通过 `POST /api/hello` 进入 Gateway，由 MessagePacket.maxType/minType 命中 routes.yaml 后转发到本服务。

## 当前状态

S0 阶段：只定义骨架和规范，不实现复杂业务逻辑。
