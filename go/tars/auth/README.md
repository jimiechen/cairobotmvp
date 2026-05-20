# AuthServer

## 服务定位

认证服务，提供登录、Token 校验、刷新、登出等认证能力。

## Tars 标识

- TarsCloud App: CaiRobot
- Server: AuthServer
- Servant: AuthObj
- Object: CaiRobot.AuthServer.AuthObj
- Module: CaiRobotAuthApp
- Interface: AuthObj
- IDL: tars/protocol/tars/auth.tars

## 标准方法签名

所有方法统一使用：

```tars
int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

## 调用来源

外部请求通过 `POST /api/hello` 进入 Gateway，由 MessagePacket.maxType/minType 命中 routes.yaml 后转发到本服务。

## 当前状态

S0 阶段：只定义骨架和规范，不实现复杂业务逻辑。
