# Social Service 开发规范文档

> **文档编号**: DEV-GUIDE-SOCIAL-001
> **版本**: v1.0
> **创建日期**: 2026-06-16
> **适用范围**: go/modules/social/ 全域开发
> **关联协议**: user_base.proto (1000) / group_base.proto (2000) / topic_base.proto (3000) / third_base.proto (4000) / inbox_base.proto (5000)
> **关联 PRD**: PRD-social-app-mvp.md / PRD-admin-backend.md
> **关联 ADR**: ADR-social-data-level-and-cache-strategy.md / ADR-plaza-virtual-membership.md

---

## 1. 三条不可违反规则

### Rule 1：一协议一 svc 文件
每个 Request/Response 协议对，对应一个独立的 `svc_{action}.go` 文件。
Trae **每次只允许创建或修改一个** `svc_*.go` 文件，禁止在一个文件里实现多个协议。

### Rule 2：proto 文件冻结
客户端已接入。`*Request` / `*Response` 的字段名、字段编号、package 名全部不可变。
Service 文件内部使用内部 `model.*` 类型，通过转换函数与 proto 类型解耦。

### Rule 3：数据库 model 层独立于 proto
`model.go` 中的结构体字段按 `basemodel.md` + PRD 设计，与 proto 字段无关。
proto 字段 → model 字段的映射在 `svc_*.go` 中的转换函数内完成。

---

## 2. 完整调用链（7 层架构）

```
Client App
   │  POST /api/hello
   │  Body: MessagePacket{ maxType, minType, extend, platform, data(proto bytes) }
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Gateway  go/gateway/proto-gateway/                  │
│  ① Decode MessagePacket → {maxType, minType, extend, data}  │
│  ② 查 routes.yaml → Target{App,Server,Servant,Method}       │
│  ③ extend["minType"] = "1021"  (注入 minType 字符串)         │
│  ④ TarsInvoker.Invoke(ctx, target, data, extend)             │
└──────────────────────────┬──────────────────────────────────┘
                           │ TarsInvoker.Invoke()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 2: TarsInvoker  tarsclient/invoker.go                  │
│  MVP S1 单体: LocalInvoker                                   │
│    handlers map[TargetKey]LocalHandler                       │
│    按 TargetKey{App,Server,Servant,Method} 查 map            │
│    → handler(ctx, reqBytes, extend)                          │
│  MVP S2+ 微服务: TarsGoInvoker (暂未启用)                     │
│    通过 TarsCloud 网络调用远端 Servant                         │
└──────────────────────────┬──────────────────────────────────┘
                           │ LocalHandler(ctx, reqBytes, extend)
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 3: TarsGo Servant  {domain}/servant.go                 │
│                                                              │
│  TarsGo 标准 bytes 接口签名:                                  │
│  Handle(ctx context.Context,                                 │
│         req     []byte,                                      │
│         extend  map[string]string,                           │
│  ) (retCode int, respBytes []byte, err error)                │
│                                                              │
│  职责（严格限定）:                                             │
│  ① 模块初始化时向 LocalInvoker.Register() 注册自身            │
│  ② 从 extend["minType"] 提取 minType 字符串转 int            │
│  ③ 调用 Handler.Dispatch(ctx, minType, reqBytes, extend)     │
│  ④ 本文件不包含任何业务逻辑                                    │
│  ⑤ 本文件不解析 proto bytes                                   │
└──────────────────────────┬──────────────────────────────────┘
                           │ Handler.Dispatch()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 4: Handler  {domain}/handler.go                        │
│                                                              │
│  职责（严格限定）:                                             │
│  ① 持有所有 svc_*.go Service 的引用                           │
│  ② switch minType → 找到对应 svc 的 Handle 方法              │
│  ③ proto.Unmarshal(reqBytes, req)                            │
│  ④ resp, err := svc.Handle(ctx, req)                         │
│  ⑤ proto.Marshal(resp) → respBytes                           │
│  ⑥ 本文件不包含任何业务逻辑                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │ svc.Handle()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 5: svc_*.go  (一协议一文件)                             │
│                                                              │
│  每个文件只处理一个协议(minType)，固定五步:                     │
│  ① 参数校验 → common.proto UserErrorCode/GroupErrorCode      │
│  ② 权限校验 → permission.Service (不直接查表)                 │
│  ③ 1级数据读写 → MySQL 事务，通过 Repository 接口             │
│  ④ 发布领域事件 → 2级数据异步更新 Redis/stats                  │
│  ⑤ 返回 proto Response (字段不可改，错误通过 Result 表达)      │
└──────────────────────────┬──────────────────────────────────┘
                           │ repo.XXX()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 6: Repository + Model                                  │
│  repository.go        ← DB 操作接口定义                       │
│  repository_gorm.go   ← GORM 实现                            │
│  model.go             ← 内部 DB 模型（非 proto 类型）          │
│       → MySQL   (1级强一致数据)                               │
│       → Redis   (2级最终一致缓存，事件驱动更新)                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 完整目录结构

```
go/modules/social/
├── module.go                           # 模块注册入口，聚合所有域 Servant
├── permission/
│   └── service.go                      # 跨域权限服务（8个方法，唯一权限入口）
│
├── member/                             # maxType=1000
│   ├── servant.go                      # [Layer 3] TarsGo Servant → 注册+转发
│   ├── handler.go                      # [Layer 4] minType switch dispatch
│   ├── repository.go         