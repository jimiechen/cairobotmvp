# Proto Gateway

## 职责

Proto Gateway 是 CaiRobot MVP 的单网关入口实现。

- 只暴露 `POST /api/hello`
- Content-Type 为 `application/octet-stream`
- 请求体是 MessagePacket bytes
- 主路由键是 `maxType:minType`
- 不按 URL path 路由业务
- 不以 proto package/service/method 作为主路由
- 不承载复杂业务逻辑
- 不直接访问业务数据库

## 负责

- MessagePacket 解析
- maxType/minType 校验
- routes.yaml 查找
- 协议编号注册表校验
- data 反序列化为业务 Protobuf
- 鉴权上下文处理
- extend 构造
- Protobuf marshal/unmarshal
- Tars bytes 接口调用
- Tars return code 转换
- 响应 MessagePacket 封装
- 日志、trace、metrics

## 不负责

- 具体业务逻辑（由 TarsGo 服务处理）
- 业务数据库访问
- 复杂状态管理

## 目录结构

```text
gateway/proto-gateway/
├── README.md
├── configs/
│   └── routes.yaml
└── internal/
    ├── router/
    │   └── README.md
    ├── adapter/
    │   └── README.md
    └── tarsclient/
        └── README.md
```

## 相关文档

- [router README](internal/router/README.md)
- [adapter README](internal/adapter/README.md)
- [tarsclient README](internal/tarsclient/README.md)
- [docs/api/tars规范.md](../../docs/api/tars规范.md)
- [docs/api/http-gateway规范.md](../../docs/api/http-gateway规范.md)
