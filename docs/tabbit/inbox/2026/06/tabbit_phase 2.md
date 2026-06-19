@social-phase1p0-closeout-p1p2-plan.md 评审方案，当前阶段只完成基础user，group，topic三模块组件，thrid和inbox先挂起

### **评审结论：方案方向正确，但 P1/P2 范围仍偏大；当前阶段应收敛为 user/member、group、topic 三个基础模块闭环，third 和 inbox 明确挂起，不进入本轮排期。**

我建议将当前方案从“P0 收口 + P1/P2 全量进阶方案”调整为“社交基础三域收口方案”。当前阶段的目标不应是扩展新域，而是把已有 user/member、group、topic 三个模块做成稳定、可测、可接入网关的基础业务组件。third、inbox 可以保留在路线图里，但不应出现在当前阶段的执行任务、验收标准和工期估算中。

### **一、总体评审意见**

`social-phase1p0-closeout-p1p2-plan.md` 的 P0 收口报告比较完整，能够清楚说明当前交付状态、测试覆盖、协议覆盖、遗留 Stub、事件系统 TODO 和后续风险。作为阶段收口文档，它的质量是合格的。

但作为下一阶段执行方案，目前有三个问题需要调整。

第一，P1 范围仍然偏大。文档虽然写了“P1 不做 third/inbox 两域”，但 P1 同时包含认证闭环、权限补全、Topic Stub 修复、事件注入链路、Redis Pub/Sub、6 个补充协议，实际已经接近一个小版本迭代。对于当前只要求完成基础 user/group/topic 三模块组件的阶段来说，应该进一步压缩。

第二，Phase 2 中把 third/inbox 写入 2D/2E，容易让执行团队误解为它们是自然顺延的下一阶段任务。既然当前已明确 third 和 inbox 挂起，就应从 Phase 2 主计划中移出，改为“冻结项 / 后续专项”。

第三，事件系统在当前阶段应谨慎定位。P0 已经有 event 包、MemoryBus、NoopPublisher 和 Redis Pub/Sub TODO，但如果当前目标只是基础三模块组件闭环，事件系统可以保留接口和发布点，不必强制完成完整 Redis 消费者、StatsHandler、NotifyHandler。否则会把基础模块交付拖入基础设施建设。

---

### **二、建议重新定义当前阶段目标**

当前阶段建议命名为：

```text
Phase 1 P1：Social Core 三域基础闭环
```

目标是：

```text
完成 user/member、group、topic 三个基础模块的核心协议、核心 Repository、核心 Service、Handler Dispatch、Gateway Route、单元测试和基础鉴权闭环。
```

不建议当前阶段承担：

```text
third 域
inbox 域
完整通知系统
完整统计消费者
完整缓存生产部署
复杂 CQRS 读模型
收益、钱包、支付渠道、IM 签名
```

也就是说，当前阶段应该追求：

```text
基础业务能跑通
协议号对齐
Stub 明显减少
认证链路可用
权限判断可信
事件接口不阻塞后续演进
```

而不是追求：

```text
所有社交增强能力完整落地
完整事件驱动统计体系
完整缓存体系
完整消息通知体系
```

---

### **三、P1 范围建议调整**

原 P1 的批次 A、B、C、D、E 中，我建议做如下裁剪。

#### **保留：批次 A 认证闭环**

认证闭环应该保留，而且优先级最高。

当前 `member/svc_login.go` 仍是伪 token，`svc_logout.go` 缺少 Redis 黑名单，RefreshToken 未实现。这个问题会直接影响后续所有接口的真实鉴权能力。

保留任务：

```text
P1-A1 实现 JWT 令牌生成
P1-A2 实现 Token 黑名单
P1-A3 实现 RefreshToken
P1-A4 补充认证测试
```

但建议把 Redis 黑名单做成接口抽象：

```go
type TokenStore interface {
    Blacklist(ctx context.Context, token string, ttl time.Duration) error
    IsBlacklisted(ctx context.Context, token string) (bool, error)
}
```

实现可以分为：

```text
MemoryTokenStore：单测/本地
RedisTokenStore：生产
```

这样即使 Redis 环境尚未确认，也不影响 P1 基础模块测试闭环。

#### **保留：批次 B 权限服务补全**

权限服务是 user/group/topic 三模块之间的核心粘合层，应保留。

尤其是：

```text
CanManageGroup
CanManageMember
CanReadTopic
CanOperateTopic
```

这些能力决定圈主、成员、普通用户之间的数据边界。根据 ADR，权限判断只能基于 1级数据，不能依赖 2级统计或缓存，所以当前阶段应该优先保证权限服务基于 MySQL/Repository 的正确性。

保留任务：

```text
P1-B1 CanManageGroup 接入真实数据源
P1-B2 平台管理员/管理角色查询接入
P1-B3 权限边界测试补充
```

#### **保留：批次 C Topic Stub 修复**

Topic 当前 Stub 会影响基础三域完整性，应保留。

尤其是：

```text
CreateReport 不能复用 UpdateTopic stub
UnfavoriteTopic 不能复用 FavoriteTopic 反向占位
```

保留任务：

```text
P1-C1 新建 SvcCreateReport
P1-C2 新建 SvcUnfavoriteTopic
P1-C3 更新 topic/handler.go
P1-C4 补充测试
```

不过需要注意，举报功能虽然属于 topic 域，但如果它后续依赖审核后台或 inbox 通知，当前阶段只需要完成“举报记录落库 + 返回 report_id”，不要扩展到通知和审核流。

#### **调整：批次 D 事件注入链路**

事件系统建议降级处理。

原计划中 P1-D 包含：

```text
ModuleOptions
Servant/Handler/Svc 注入链路
6 个核心事件发布
Redis Pub/Sub 基础实现
事件发布测试
```

我建议当前阶段只保留：

```text
事件接口注入链路
NoopPublisher 默认实现
Memory/FakePublisher 测试实现
6 个核心事件发布点
事件发布失败不阻塞业务测试
```

暂缓：

```text
Redis Pub/Sub 生产实现
RedisSubscriber
StatsHandler 真实统计更新
NotifyHandler
完整 CacheHandler
```

原因是当前目标是基础 user/group/topic 三域组件，而不是事件基础设施生产化。Redis Pub/Sub 当然是后续生产必须项，但可以放到“事件基础设施专项”中，不阻塞当前三域基础闭环。

调整后的 P1-D 应为：

| 任务ID | 调整后任务 | 是否当前阶段执行 |
|---|---|---|
| P1-D1 | ModuleOptions + Option 模式 | 执行 |
| P1-D2 | Servant 构造函数注入 event.Publisher | 执行 |
| P1-D3 | Handler 透传 publisher 到 svc | 执行 |
| P1-D4 | 6 个核心事件发布点 | 执行 |
| P1-D5 | 事件发布失败不阻塞主业务测试 | 执行 |
| P1-D6 | Redis Pub/Sub 基础实现 | 暂缓到事件专项 |

这样可以保证业务代码不需要二次大改，同时避免当前阶段被 Redis 运行环境、订阅生命周期、消费者幂等、统计更新拖住。

#### **保留但压缩：批次 E 补充 P0 遗漏核心协议**

批次 E 是必要的，但建议从 7 个协议压缩为“基础三域最小闭环协议”。

优先保留：

```text
GetGroupInfo      2001
GetGroupMembers   2017
GetMembership     2039
UpdateMemberStatus 1033
ListUserGroups    1045
GetReplyList      3025
```

这 6 个协议确实都属于 user/group/topic 基础能力，建议当前阶段实现。

但要注意 `UpdateMemberStatus` 的定位。如果它涉及平台管理员操作、封禁、解封、审计日志，那么当前阶段至少要做到：

```text
状态更新落库
发布 UserStatusChanged 事件，默认 Noop
清理鉴权/用户资料缓存的接口预留
权限检查最小可用
测试覆盖 active/banned/deleted 等状态流转
```

不要扩展到完整风控、通知、后台操作台。

---

### **四、third 和 inbox 应该如何处理**

当前应明确写入方案：

```text
third 域：挂起，不进入 Phase 1 P1，不进入当前 Phase 2 主计划。
inbox 域：挂起，不进入 Phase 1 P1，不进入当前 Phase 2 主计划。
```

文档中 Phase 2D 和 Phase 2E 建议删除出主执行计划，改为“冻结池”。

可以改成：

```text
冻结项 F1：Third 域，等待支付、OSS、OAuth、分享链路产品需求确认后单独立项。
冻结项 F2：Inbox 域，等待通知系统、IM 能力、站内信模型确认后单独立项。
```

这样做的好处是很明确：当前社交域不会因为 future scope 而膨胀。

third 域涉及：

```text
OSS 上传
分享链接
OAuth 第三方登录
钱包支付
```

这些都不是 user/group/topic 基础闭环的必要条件。

inbox 域涉及：

```text
消息查询
已读标记
系统通知
IM 签名
```

这些依赖通知系统和消息模型，也不应该与基础三域一起交付。

---

### **五、建议调整后的 P1 任务清单**

当前阶段建议保留以下任务。

| 批次 | 任务 | 当前阶段建议 |
|---|---|---|
| A | 认证闭环 | 保留，必须做 |
| B | 权限服务补全 | 保留，必须做 |
| C | Topic Stub 修复 | 保留，必须做 |
| D | 事件注入链路 | 只做接口注入和发布点，Redis 实现暂缓 |
| E | 基础协议补齐 | 保留 6 个核心协议 |
| F | third/inbox | 删除出当前阶段，挂起 |

调整后，当前阶段的核心交付应该是：

```text
1. Login / Logout / RefreshToken 可用
2. User 状态变更可用
3. Group 详情、成员列表、会员信息可查
4. Topic 评论列表、举报、取消收藏可用
5. 权限服务能支撑圈主/成员/普通用户边界
6. event.Publisher 注入链路完成，默认 Noop，不阻塞业务
7. routes.yaml 与 proto/base 协议号继续保持一致
8. 所有 Stub/占位从 handler 主链路中清除
```

---

### **六、建议调整后的验收标准**

原文档验收标准里包含“6 个核心事件可发布到 Redis”。我建议当前阶段改为：

```text
6 个核心事件可通过 event.Publisher 接口发布；
默认 NoopPublisher 不影响业务；
测试中可通过 FakePublisher 断言事件类型和 payload；
RedisPublisher 不作为当前阶段强制验收项。
```

调整后的验收标准建议如下：

| 检查项 | 当前阶段通过标准 |
|---|---|
| 编译 | `go build ./...` PASS |
| 测试 | `go test ./... -count=1` PASS，测试数 >= 260 |
| 静态检查 | `go vet ./...` 无新增阻塞项 |
| 协议对齐 | base proto / social proto / routes.yaml 三源一致 |
| 认证链路 | Login → JWT → Logout → RefreshToken 流程可测 |
| 权限链路 | CanManageGroup / CanManageMember / CanReadTopic 核心边界可测 |
| User 基础协议 | Register/Login/Get/Update/Status/ListGroups 可用 |
| Group 基础协议 | Create/Join/Leave/GetInfo/GetMembers/GetMembership 可用 |
| Topic 基础协议 | Create/List/Detail/Reply/GetReplyList/Like/Favorite/Unfavorite/Report 可用 |
| 事件链路 | event.Publisher 注入完成，Noop 默认，Fake 可测 |
| Stub 清理 | 当前基础三域 handler 主链路不再复用 stub/占位 svc |
| third/inbox | 不创建、不排期、不纳入验收 |

---

### **七、建议调整后的时间线**

原文档 P1 预估 5-8 个工作日。如果暂缓 Redis Pub/Sub 和 third/inbox，当前阶段可以压缩为 5 个工作日左右。

```text
Day 1:
认证闭环
- JWT
- Logout 黑名单接口
- RefreshToken
- 测试

Day 2:
权限补全 + Topic Stub 修复
- CanManageGroup / CanManageMember
- CreateReport
- UnfavoriteTopic
- 测试

Day 3:
Group 补充协议
- GetGroupInfo
- GetGroupMembers
- GetMembership
- routes.yaml
- 测试

Day 4:
Member/Topic 补充协议
- UpdateMemberStatus
- ListUserGroups
- GetReplyList
- routes.yaml
- 测试

Day 5:
事件注入链路 + 全量收口
- ModuleOptions
- Publisher 注入
- 6 个事件发布点
- FakePublisher 测试
- go build / go test / go vet
- 文档同步
```

如果中途发现 JWT、权限或 Repository 改动超出预期，可以扩展到 6-7 个工作日。但不建议因为 Redis Pub/Sub 或 eventhandler 统计更新拉长当前阶段。

---

### **八、文档中建议修改的具体位置**

#### **1. 修改 4.2 P1 范围定义**

原文：

```text
P1 不做：third/inbox 两域、缓存层生产部署、事件系统完整落地、UI/前端。
P1 专注：认证闭环、权限完善、事件注入链路、剩余 P0 协议补全。
```

建议改为：

```text
P1 不做：third/inbox 两域、缓存层生产部署、Redis Pub/Sub 生产化、完整事件消费者、通知系统、UI/前端。
P1 专注：user/member、group、topic 三个基础模块的认证闭环、权限完善、Stub 清理、核心协议补齐、事件接口注入链路。
```

#### **2. 修改批次 D**

将 `P1-D6 Redis Pub/Sub 基础实现（生产可用）` 从 P1 移出，改为：

```text
P1-D6 FakePublisher / MemoryPublisher 测试支持
```

新增挂起项：

```text
P2-Event-1 Redis Pub/Sub Publisher/Subscriber 生产实现
```

#### **3. 修改 Phase 2 章节**

原 Phase 2D / 2E 建议从主计划移除，改为：

```text
冻结域：Third / Inbox

Third 域暂不启动，等待外部支付、OSS、OAuth、分享能力需求确认。
Inbox 域暂不启动，等待通知系统、站内信、IM 签名需求确认。
```

#### **4. 修改验收标准**

原验收标准：

```text
事件系统：6 个核心事件可发布到 Redis；发布失败不阻塞业务
```

建议改为：

```text
事件系统：6 个核心事件可通过 event.Publisher 接口发布；默认 NoopPublisher；FakePublisher 可断言；发布失败不阻塞业务。Redis Pub/Sub 不作为当前阶段强制验收项。
```

#### **5. 修改风险项 R-P1-002**

原风险：

```text
Redis 依赖引入（auth 黑名单 + 事件 Pub/Sub）
```

建议拆成两条：

```text
R-P1-002A：认证黑名单依赖 Redis。缓解：TokenStore 接口隔离，Memory 实现用于测试，本地/生产使用 Redis。
R-P1-002B：事件 Pub/Sub 生产实现暂缓。缓解：当前只完成 Publisher 接口注入和 Fake 测试，RedisPublisher 进入事件专项。
```

---

### **九、给主控的建议决策**

建议主控确认以下决策：

```text
1. 当前阶段只交付 user/member、group、topic 三个基础模块。
2. third 和 inbox 全部挂起，不进入当前 P1/P2 主执行计划。
3. Redis Pub/Sub 事件生产实现暂缓，不阻塞当前基础三域收口。
4. 当前事件系统只完成接口注入、Noop 默认实现、Fake 测试和核心发布点。
5. P1 验收重点从“事件基础设施生产化”调整为“三域基础业务可用”。
```

按这个调整后，方案会更贴合当前阶段目标，也能避免社交域在基础能力尚未完全闭环前过早扩展到 third、inbox、通知、统计和完整事件基础设施。

*内容由 AI 生成仅供参考*