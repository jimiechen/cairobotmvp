@ADR-social-data-level-and-cache-strategy.md @领域事件评审意见.md 请答复trae反馈的评审意见

### **总体答复：评审意见采纳，领域事件方案需要从“纯内存 MVP 方案”修正为“接口隔离 + Redis 优先 + Memory 测试实现 + Outbox 预留”的执行方案。**

Trae 的评审结论整体成立。结合 `ADR-social-data-level-and-cache-strategy.md` 已采纳内容，原方案中“首期只做内存事件总线、P1 再引入 PubSub”的安排需要调整。原因是 ADR 已明确要求 1级数据更新后发布领域事件，并且 2级数据、缓存失效、统计更新均依赖事件机制；同时 TarsCloud 服务存在多实例部署可能，纯内存事件总线无法跨实例传播事件，不能作为生产默认实现。

下面是对评审意见的正式答复，可直接给 Trae 作为修订指令。

### **评审结论确认**

接受 Trae 的总体评审结论：**建议修改后执行**。

本次领域事件方案的方向是正确的，但原方案与 PRD、ADR、实际代码结构之间存在若干不一致点。执行前必须修正事件清单、注入链路、文件引用、状态常量、并发安全和可观测性设计。

同时，需要补充一个更关键的架构修正：**生产环境不应以 MemoryBus 作为默认事件总线，而应默认使用 Redis 实现。MemoryBus 仅保留给单测、本地开发或无 Redis 环境的降级场景。**

---

### **一、关于事件基础设施选型的修正**

原方案中“内存事件总线 + Outbox 预留，不引入 Redis Pub/Sub”的范围控制过于保守。结合 ADR 中的事件驱动缓存策略，现调整为：

```text
生产/测试环境：RedisPublisher + RedisSubscriber
单测/本地开发：MemoryBus
未配置事件能力：NoopPublisher
后续强可靠事件：Outbox + Redis Streams / MQ
```

理由如下：

ADR 已明确规定：

```text
1级数据更新后必须发布领域事件
2级数据通过领域事件驱动更新
缓存消费者、统计消费者、通知消费者监听事件后执行处理
```

如果使用纯内存事件总线，只能在当前 Go 进程内传播事件。TarsCloud 多实例部署时，实例 A 写入 1级数据后，实例 B、实例 C 无法收到缓存失效和统计更新事件，会造成跨实例缓存不一致、统计漂移和通知遗漏。因此，MemoryBus 不适合作为生产默认实现。

修订后的执行原则是：

```text
event.Publisher / event.Subscriber 保持接口隔离
Redis 实现作为生产默认
MemoryBus 作为测试实现
NoopPublisher 仅用于未启用事件系统时兜底
Outbox 作为可靠事件的后续演进
```

MVP-P0 可以先使用 Redis Pub/Sub。如果需要任务重试、延迟处理、失败队列，可以在具体消费者内部转投 Asynq；如果后续事件不可丢，再升级为 Outbox + Redis Streams。

---

### **二、对 P0 必须修改项的答复**

#### **1. 文档重复内容：采纳**

接受评审意见。

`领域设计.md` 前 1-27 行重复摘要需要删除，保留正式设计正文作为唯一版本，避免执行时产生歧义。

修订要求：

```text
删除重复前言
保留正式文档
在文档开头补充 ADR 引用和本次修订说明
```

#### **2. 缺少 UserStatusChanged 事件：采纳**

接受评审意见。

PRD/ADR 中 1级数据清单包含 `users.status`，`UpdateMemberStatus` 属于明确业务操作，必须在状态变更后发布领域事件。原方案只覆盖 `MemberRegistered`，没有覆盖用户状态变更，这是遗漏。

需要补充事件：

```go
EventUserStatusChanged = "UserStatusChanged"
```

建议 payload：

```go
type UserStatusChangedPayload struct {
    UserID     string `json:"user_id"`
    OldStatus  int32  `json:"old_status"`
    NewStatus  int32  `json:"new_status"`
    OperatorID string `json:"operator_id,omitempty"`
    Reason     string `json:"reason,omitempty"`
    ChangedAt  int64  `json:"changed_at"`
}
```

触发点：

```text
UpdateMemberStatus / AdminUpdateMemberStatus 成功提交事务后发布
```

消费者动作：

```text
删除 user:profile:{userId}
删除 user:auth:{userId} 或相关会话/权限缓存
通知用户或后台审计系统
```

#### **3. TopicAudited 粒度过粗：采纳并修正为独立事件常量**

接受评审意见。

原方案使用 `TopicAudited` + `action` 字段表达审核结果，技术上可行，但与 PRD 中的 `TopicApproved / TopicRejected / TopicBanned` 不完全对齐。为避免消费者中大量 switch action，并让缓存策略更清晰，修正为独立事件常量。

需要补充：

```go
EventTopicApproved = "TopicApproved"
EventTopicRejected = "TopicRejected"
EventTopicBanned   = "TopicBanned"
```

三者可以共享 payload：

```go
type TopicAuditPayload struct {
    TopicID    string `json:"topic_id"`
    GroupID    string `json:"group_id"`
    AuthorID   string `json:"author_id"`
    OperatorID string `json:"operator_id"`
    Reason     string `json:"reason,omitempty"`
    AuditedAt  int64  `json:"audited_at"`
}
```

消费者策略：

```text
TopicApproved：刷新 topic detail/list，更新 group topics_count 或发布可见列表索引
TopicRejected：删除 topic detail/list 缓存，通知作者
TopicBanned：删除 topic detail/list/cache/rank/hot 相关缓存，通知作者，记录高权限操作
```

如仍需保留 `TopicAudited`，只能作为内部兼容事件，不作为 PRD 对外标准事件。

#### **4. 文件名和实际代码结构不符：采纳**

接受评审意见。

执行前必须以实际代码为准修正文档中的文件路径。方案中引用不存在文件会影响 Trae 执行准确性。

修订要求：

```text
删除或标注尚未实现的 svc_create_plan.go
删除或标注尚未实现的 svc_unban.go
删除或标注尚未实现的 topic/svc_pin.go
将 svc_favorite.go 修正为实际文件名 svc_favorite_topic.go
```

对于尚未实现但属于 PRD 范围的服务，不能在领域事件任务中假设其存在。应该分为两类：

```text
已存在 svc：本次接入事件发布
未存在 svc：只补充事件常量和 payload，不改业务代码
```

#### **5. svc_leave.go 硬编码状态值：采纳，纳入本次修复范围**

接受评审意见。

`member.Status = 2 // 已退出` 属于裸状态值，违反编码规范，也会影响事件幂等判断的可读性。需要在本次任务中一并修复。

建议新增或复用统一常量：

```go
const (
    GroupMemberStatusActive  = 1
    GroupMemberStatusLeft    = 2
    GroupMemberStatusBanned  = 3
    GroupMemberStatusRemoved = 4
    GroupMemberStatusExpired = 5
)
```

但具体值必须以当前 proto 或数据库枚举定义为准，不能凭文档直接改。执行前需要检查已有 pb 枚举、model 常量或数据库注释，优先复用已有定义。

修订要求：

```text
全局排查 group/member/topic svc 中的裸状态值
优先替换为 pb 枚举或 domain 常量
将状态常量纳入事件 payload 和幂等判断
```

#### **6. 注入链路与实际代码不匹配：采纳，并扩大 Step E6 改动范围**

接受评审意见。

原方案低估了注入链路改动范围。当前实际结构是：

```go
func NewModule(memberRepo member.Repository, groupRepo group.Repository, topicRepo topic.Repository) *Module
```

而 handler 构造函数目前只接收 repo：

```go
func NewHandler(repo Repository) *Handler
```

如果接入 `event.Publisher`，需要修改：

```text
go/modules/social/module.go
go/modules/social/member/handler.go
go/modules/social/group/handler.go
go/modules/social/topic/handler.go
go/modules/social/member/servant.go
go/modules/social/group/servant.go
go/modules/social/topic/servant.go
相关 svc 构造函数
相关单测 mock/fake
```

修订后的 Step E6 应改为“事件注入链路改造”，不是小范围 wiring。

建议目标结构：

```go
type ModuleOptions struct {
    EventPublisher event.Publisher
    EventSubscriber event.Subscriber
}
```

如果为了减少破坏性，也可以先采用兼容签名：

```go
func NewModule(
    memberRepo member.Repository,
    groupRepo group.Repository,
    topicRepo topic.Repository,
    opts ...Option,
) *Module
```

默认注入 `NoopPublisher`，避免未改完的调用点直接编译失败。

---

### **三、对 P1 建议修改项的答复**

#### **1. DomainEvent.Payload 类型问题：部分采纳**

接受“存在类型安全风险”的判断，但 MVP-P0 阶段可以先不强制上强类型泛型事件系统。

修订方案：

```go
type DomainEvent struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Version     string          `json:"version"`
    AggregateID string          `json:"aggregate_id,omitempty"`
    OccurredAt  int64           `json:"occurred_at"`
    Payload     json.RawMessage `json:"payload,omitempty"`
}
```

事件构造函数仍然支持传入强类型 payload：

```go
func NewDomainEvent(eventType string, aggregateID string, payload any) (DomainEvent, error) {
    b, err := json.Marshal(payload)
    if err != nil {
        return DomainEvent{}, err
    }

    return DomainEvent{
        ID:          NewEventID(),
        Type:        eventType,
        Version:     "1.0",
        AggregateID: aggregateID,
        OccurredAt:  time.Now().UnixMilli(),
        Payload:     b,
    }, nil
}
```

这样兼顾：

```text
事件总线统一传输 JSON
payload 定义仍然强类型
消费者按事件类型反序列化
后续 Outbox 可直接落 JSON
```

不建议继续使用裸 `map[string]interface{}` 作为长期结构。

#### **2. MemoryBus 同步模式性能风险：采纳**

接受评审意见。

MemoryBus 如果同步执行 handler，而 handler 内部写 DB 或访问 Redis，会拖慢主请求链路。修订如下：

```text
MemoryBus 仅用于单测/本地开发
生产 RedisSubscriber 独立 goroutine 消费
StatsHandler 首版避免在请求链路中同步重 IO
handler panic 必须 recover
handler error 必须结构化日志记录
```

如保留 MemoryBus，建议实现两种模式：

```go
type MemoryBusMode string

const (
    MemoryBusSync  MemoryBusMode = "sync"
    MemoryBusAsync MemoryBusMode = "async"
)
```

但 MVP-P0 更建议简单处理：

```text
单测用同步 MemoryBus，便于断言
本地开发可用异步 MemoryBus
生产不使用 MemoryBus
```

#### **3. GroupMemberChangedPayload 合并语义过多：部分采纳**

接受“字段必填性需要说明”的意见，但不强制拆成多个 payload。

建议事件常量保持明确：

```go
EventGroupMemberRemoved = "GroupMemberRemoved"
EventGroupMemberBanned  = "GroupMemberBanned"
EventGroupMemberMuted   = "GroupMemberMuted"
EventGroupMemberRecovered = "GroupMemberRecovered"
```

payload 可以共享：

```go
type GroupMemberChangedPayload struct {
    GroupID     string `json:"group_id"`
    UserID      string `json:"user_id"`
    OperatorID  string `json:"operator_id"`
    OldStatus   int32  `json:"old_status,omitempty"`
    NewStatus   int32  `json:"new_status,omitempty"`
    MutedUntil  int64  `json:"muted_until,omitempty"`
    Reason      string `json:"reason,omitempty"`
    ChangedAt   int64  `json:"changed_at"`
}
```

文档中必须明确：

```text
GroupMemberMuted 必须包含 muted_until
GroupMemberRemoved 必须包含 old_status/new_status
GroupMemberBanned 必须包含 reason/operator_id
GroupMemberRecovered 必须包含 old_status/new_status
```

这样比一个 `action` 字段更清晰，也与 PRD 事件清单更一致。

#### **4. 幂等规则缺少实现指导：采纳**

接受评审意见。

事件发布必须以“实际状态发生变化”为准，不能只以请求成功为准。否则重复请求会导致统计重复加减。

建议统一采用 repository 层 CAS 或 rowsAffected 模式。

示例：

```sql
UPDATE group_members
SET status = ?, updated_at = ?
WHERE group_id = ?
  AND user_id = ?
  AND status <> ?;
```

或者针对离群：

```sql
UPDATE group_members
SET status = ?, updated_at = ?
WHERE group_id = ?
  AND user_id = ?
  AND status = ?;
```

svc 层根据 `rowsAffected` 判断是否发布事件：

```go
changed, err := repo.MarkLeft(ctx, groupID, userID)
if err != nil {
    return err
}

if changed {
    _ = publisher.Publish(ctx, event)
}
```

规则：

```text
没有实际变更，不发布事件
事务失败，不发布事件
事务成功但事件发布失败，不回滚主业务，但记录错误并进入补偿机制
```

对于 JoinGroup、LeaveGroup、BanMember、ReactTopic、CancelReactTopic 都需要按这个模式处理。

#### **5. 事件版本管理机制：采纳**

接受评审意见。

`DomainEvent` 增加 `Version` 字段，默认 `"1.0"`。

```go
Version string `json:"version"`
```

事件消费者必须按 `Type + Version` 处理。MVP-P0 可以只支持 `"1.0"`，但结构上要预留，避免后续 Outbox 或 Redis Streams 落地后难以兼容。

---

### **四、对测试缺口的答复**

#### **1. MemoryBus 并发安全测试：采纳**

接受评审意见。

如果保留 MemoryBus，必须明确并发模型。建议实现为并发安全：

```go
type MemoryBus struct {
    mu       sync.RWMutex
    handlers map[string][]Handler
}
```

`Subscribe` 使用写锁，`Publish` 使用读锁复制 handler 切片后释放锁，再执行 handler，避免执行 handler 时持有锁。

测试补充：

```text
TestMemoryBus_ConcurrentSubscribePublish
TestMemoryBus_Publish_NoHandler
TestMemoryBus_Publish_HandlerError
TestMemoryBus_Publish_HandlerPanicRecovered
```

#### **2. 事件发布失败不影响业务：采纳**

接受评审意见。

必须补充 svc 级测试，验证事件发布失败时主业务不回滚。尤其是以下操作：

```text
MemberRegister
CreateGroup
JoinGroup
LeaveGroup
CreateTopic
ReactTopic
UpdateGroupMemberStatus
```

测试目标：

```text
Publisher 返回 error：业务响应仍成功
Publisher panic：recover 后业务响应仍成功
错误被记录日志
必要时写入 outbox 或补偿记录
```

但要注意：如果后续采用 Outbox 且 Outbox 写入是事务内的一部分，则 outbox 写失败的策略需要另行定义。当前 MVP-P0 可先采用“事件失败不阻塞主业务”。

#### **3. Handler 错误传播行为：采纳**

接受评审意见。

MemoryBus / RedisSubscriber 的 handler 执行策略需要明确。

建议采用 best-effort 模式：

```text
同一个事件有多个 handler 时，一个 handler 失败不阻止其他 handler 执行
所有错误统一收集并记录
Publish 返回 aggregated error，但业务方默认不因该错误回滚
```

原因是缓存失效、统计更新、通知发送属于不同消费者，不能因为通知失败导致缓存失效不执行。

---

### **五、对文档缺口的答复**

#### **1. 缺少 ADR 交叉引用：采纳**

接受评审意见。

领域事件设计文档开头必须显式引用：

```text
ADR-social-data-level-and-cache-strategy.md
PRD-social-app-mvp.md
协议编号注册表
social-openapi.yaml
```

并说明本方案是 ADR 中“事件驱动缓存”和“2级数据最终一致”机制的实现方案。

#### **2. 缺少事件迁移/变更流程：采纳**

接受评审意见。

新增事件必须遵循固定流程：

```text
1. 在 event/constants.go 新增事件常量
2. 在 event/payload.go 新增强类型 payload
3. 在事件注册表/文档中登记事件语义、触发点、消费者动作
4. 在对应 svc 事务提交后发布事件
5. 在 eventhandler 中新增或复用消费者
6. 补充单测：发布点测试、payload 测试、handler 幂等测试
7. 如涉及缓存，补充 CacheInvalidator key 清单
```

修改 payload 字段时：

```text
只允许向后兼容新增 optional 字段
不允许直接删除或改名已有字段
破坏性变更必须提升 Version
消费者必须兼容旧版本
```

#### **3. 缺少监控/可观测性：采纳**

接受评审意见。

事件系统需要结构化日志和指标字段。最低要求：

日志字段：

```text
event_id
event_type
event_version
aggregate_id
publisher
consumer
duration_ms
result
error
retry_count
```

建议指标：

```text
social_event_publish_total
social_event_publish_failed_total
social_event_consume_total
social_event_consume_failed_total
social_event_consume_duration_ms
social_event_handler_panic_total
```

MVP-P0 如果暂未接 metrics 系统，至少要保证结构化日志完整。

---

### **六、对风险提示的答复**

Trae 提出的四个风险均接受。

#### **R-EVENT-001：改动面大**

接受。执行时必须拆分提交，不要一次性改完整事件系统。

建议顺序：

```text
Commit 1：新增 event 包、常量、payload、DomainEvent、NoopPublisher、MemoryBus
Commit 2：新增 RedisPublisher/RedisSubscriber 或接口占位
Commit 3：改 module/handler/servant 注入链路，默认 NoopPublisher
Commit 4：接入 member 事件发布
Commit 5：接入 group 事件发布
Commit 6：接入 topic 事件发布
Commit 7：补充 eventhandler/cache invalidator
Commit 8：补齐测试和文档
```

#### **R-EVENT-002：同步 handler 拖慢请求**

接受。生产不使用同步 MemoryBus。RedisSubscriber 独立消费。StatsHandler 首版不在请求链路中同步做重 IO。

#### **R-EVENT-003：Payload 类型安全问题**

接受。将 `map[string]interface{}` 调整为 `json.RawMessage + 强类型 payload 构造函数`。

#### **R-EVENT-004：硬编码状态值**

接受。纳入本次修复范围，并补充全局排查。

---

### **七、修订后的执行版本**

最终执行方案应从原来的：

```text
MemoryBus + Outbox 预留
```

修正为：

```text
event 接口隔离
+ Redis Pub/Sub 生产实现
+ MemoryBus 单测/本地实现
+ NoopPublisher 兜底
+ Outbox/Redis Streams 后续演进
```

目录建议：

```text
go/modules/social/
  event/
    constants.go
    event.go
    payload.go
    publisher.go
    subscriber.go
    noop.go
    memory_bus.go
    redis_pubsub.go
  eventhandler/
    cache_handler.go
    stats_handler.go
    notify_handler.go
  cache/
    keys.go
    invalidator.go
```

核心接口：

```go
type Publisher interface {
    Publish(ctx context.Context, evt DomainEvent) error
}

type Subscriber interface {
    Subscribe(ctx context.Context, eventType string, handler Handler) error
}

type Handler interface {
    Handle(ctx context.Context, evt DomainEvent) error
}
```

`DomainEvent` 建议结构：

```go
type DomainEvent struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Version     string          `json:"version"`
    AggregateID string          `json:"aggregate_id,omitempty"`
    OccurredAt  int64           `json:"occurred_at"`
    Payload     json.RawMessage `json:"payload,omitempty"`
}
```

### **八、最终答复**

Trae 的评审意见整体采纳，并在其基础上进一步修正事件基础设施选型：**生产默认 Redis，不再将纯 MemoryBus 作为 MVP 生产方案。**

本次执行前必须完成以下修订：

1. 删除领域设计文档重复前言。
2. 显式引用 `ADR-social-data-level-and-cache-strategy.md`。
3. 补充 `UserStatusChanged` 事件。
4. 将 `TopicAudited` 拆分为 `TopicApproved / TopicRejected / TopicBanned`。
5. 修正与实际代码不一致的文件名。
6. 将 `svc_leave.go` 及其他裸状态值修复纳入任务范围。
7. 重新评估并扩大注入链路改造范围。
8. 将 `Payload` 从 `map[string]interface{}` 调整为 `json.RawMessage + 强类型 payload`。
9. 明确 MemoryBus 仅用于测试/本地，生产使用 Redis。
10. 补充幂等、并发安全、发布失败、handler 错误传播和可观测性测试。

按上述修改后，领域事件方案可以进入执行阶段。

*内容由 AI 生成仅供参考*