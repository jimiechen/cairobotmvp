按照数据的等级优化存储，1级数据需要主动操作协议才变化，2级数据是被动变化，包括统计，被高权限的用户操作后通知，制定不同的缓存策略，是否合理，当前的CaiRobot MVP 是一个**运营管理综合平台**，成员协议组是1000，群组协议组是2000，主题协议组是3000，请按照平台规范，设计优化协议，架构关系，社交关系等，指导本地trae完成相应的prd需求文档撰写

### **这个分级思路是合理的：1级数据作为“协议主动写入才变化”的强一致主数据，2级数据作为“由事件、统计、高权限操作被动变化”的派生数据，可以显著降低缓存失效复杂度，并让成员、群组、主题三个协议组形成清晰的数据边界。**

建议在 CaiRobot MVP 运营管理综合平台里，把社交域的数据治理明确成一套“数据等级 + 协议组 + 缓存策略 + 事件通知 + 权限边界”的规范。当前协议组可以固定为：**成员协议组 1000、群组协议组 2000、主题协议组 3000**。其中 1000 负责用户身份、成员关系、关注粉丝；2000 负责群组主体、付费方案、圈主管理；3000 负责主题帖子、阅读、评论、互动。这个划分符合当前单网关、MessagePacket、Protobuf、routes.yaml、TarsGo 的架构要求。

---

### **一、数据等级设计是否合理**

你提出的规则可以正式写入 PRD 和 CODE-WIKI：

```text
1级数据：只有用户或系统通过明确业务协议主动操作后才变化。
2级数据：由 1级数据变更、统计任务、阅读互动、高权限用户操作、系统事件被动触发变化。
```

这个设计合理，原因是 1级数据代表业务事实，必须强一致、可审计、可回放；2级数据代表业务状态、计数、聚合、缓存视图，允许延迟一致。社交 App 最怕把“事实数据”和“统计数据”混在一起，例如把成员数、粉丝数、阅读数、点赞数当成强一致主数据处理，会导致每次读写都需要复杂事务，缓存也很难维护。

建议进一步明确：

| 数据等级 | 定义 | 一致性要求 | 典型数据 | 更新方式 |
|---|---|---|---|---|
| 1级数据 | 主业务事实数据 | 强一致 | 用户、群组、成员关系、帖子、订单、关注关系 | 必须通过协议主动写入 |
| 2级数据 | 派生状态、统计、索引、缓存视图 | 最终一致 | 成员数、粉丝数、阅读数、点赞数、热度、榜单、未读数 | 由事件、任务、高权限操作触发更新 |

这个规则的关键是：**所有 2级数据都必须能从 1级数据重新计算出来**。如果某个字段丢失后无法重建，它就不应该被定义为 2级数据。

---

### **二、协议组重新规划**

根据你指定的协议组，建议统一规划如下：

```text
1000 = 成员协议组 Member Domain
2000 = 群组协议组 Group Domain
3000 = 主题协议组 Topic Domain
```

这里的“协议组”建议对应 MessagePacket 的 `maxType`，具体业务动作对应 `minType`。

#### **成员协议组：maxType = 1000**

成员协议组负责账号、登录、成员关系、关注粉丝、身份权限。

| minType | 协议名称 | 数据等级 | 说明 |
|---:|---|---|---|
| 1001 | MemberRegister | 1级 | 用户注册 |
| 1002 | MemberLogin | 1级 | 用户登录 |
| 1003 | MemberLogout | 1级 | 用户登出 |
| 1004 | GetMemberProfile | 1级读取 | 获取用户资料 |
| 1005 | UpdateMemberProfile | 1级 | 修改用户资料 |
| 1010 | FollowMember | 1级 | 关注用户/圈主 |
| 1011 | UnfollowMember | 1级 | 取消关注 |
| 1012 | ListFollowers | 2级读取 | 粉丝列表，可走缓存 |
| 1013 | ListFollowings | 2级读取 | 关注列表，可走缓存 |
| 1020 | GetMemberStats | 2级读取 | 粉丝数、关注数、发帖数 |
| 1030 | UpdateMemberStatus | 1级，高权限 | 平台管理员禁用/恢复用户 |

成员协议组的核心 1级表包括 `users`、`user_follows`、`member_sessions`、`member_auth_logs`。2级数据可以放到 `member_stats`、Redis 统计缓存、搜索索引中。

#### **群组协议组：maxType = 2000**

群组协议组负责圈子主体、入群、付费方案、成员管理、圈主管理。

| minType | 协议名称 | 数据等级 | 说明 |
|---:|---|---|---|
| 2001 | CreateGroup | 1级 | 创建群组 |
| 2002 | UpdateGroup | 1级 | 修改群组资料 |
| 2003 | GetGroupDetail | 1级 + 2级读取 | 群组详情 + 统计 |
| 2004 | ListGroups | 2级读取 | 群组列表、推荐列表 |
| 2010 | JoinGroup | 1级 | 加入免费群/申请入群 |
| 2011 | LeaveGroup | 1级 | 退出群组 |
| 2012 | ListGroupMembers | 2级读取 | 群成员列表 |
| 2013 | UpdateGroupMemberStatus | 1级，高权限 | 禁言、移除、恢复成员 |
| 2020 | CreateGroupPlan | 1级 | 创建付费方案 |
| 2021 | UpdateGroupPlan | 1级 | 修改付费方案 |
| 2022 | ListGroupPlans | 1级读取 | 查询付费方案 |
| 2030 | CreateGroupOrder | 1级 | 创建付费订单 |
| 2031 | ConfirmGroupPayment | 1级，高权限/系统 | 支付确认并开通权益 |
| 2040 | GetOwnerDashboard | 2级读取 | 圈主看板 |
| 2050 | AuditGroup | 1级，高权限 | 平台审核群组 |

群组协议组的核心 1级表包括 `groups`、`group_members`、`group_plans`、`group_orders`、`group_admin_actions`。2级数据包括 `groups.members_count`、`groups.topics_count`、`group_stats`、圈主看板缓存、群组推荐索引。

#### **主题协议组：maxType = 3000**

主题协议组负责帖子、阅读、评论、点赞、收藏、内容权限。

| minType | 协议名称 | 数据等级 | 说明 |
|---:|---|---|---|
| 3001 | CreateTopic | 1级 | 发布帖子 |
| 3002 | UpdateTopic | 1级 | 修改帖子 |
| 3003 | DeleteTopic | 1级 | 删除/下架帖子 |
| 3004 | GetTopicDetail | 1级 + 2级读取 | 帖子详情，必须鉴权 |
| 3005 | ListGroupTopics | 2级读取 | 群组帖子列表 |
| 3010 | MarkTopicRead | 2级写入 | 阅读记录，可异步 |
| 3011 | GetReadHistory | 2级读取 | 阅读历史 |
| 3020 | CreateComment | 1级 | 评论 |
| 3021 | DeleteComment | 1级 | 删除评论 |
| 3030 | ReactTopic | 1级或2级 | 点赞、收藏、分享 |
| 3031 | CancelReactTopic | 1级或2级 | 取消点赞、收藏 |
| 3040 | AuditTopic | 1级，高权限 | 审核/下架帖子 |
| 3050 | GetTopicStats | 2级读取 | 阅读数、评论数、点赞数 |

这里要特别注意：`CreateComment` 建议作为 1级数据，因为评论本身是用户生成内容，需要持久化、审核和审计；但 `comments_count` 是 2级统计。`ReactTopic` 如果只作为点赞关系表，也属于 1级数据；点赞数属于 2级数据。`MarkTopicRead` 可以定义为 2级写入，因为阅读记录本身偏行为数据，允许最终一致。

---

### **三、1级数据与2级数据的存储边界**

建议在 PRD 中明确以下规则。

#### **1级数据存储规则**

1级数据必须进入 MySQL/InnoDB 主库，使用事务保证一致性。它们是业务事实，不允许只存在 Redis 或搜索引擎中。

典型 1级表：

| 业务域 | 1级表 | 说明 |
|---|---|---|
| 成员 | `users` | 用户身份主表 |
| 成员 | `user_follows` | 关注关系事实表 |
| 成员 | `member_sessions` | 登录会话 |
| 群组 | `groups` | 群组主表 |
| 群组 | `group_members` | 群成员关系 |
| 群组 | `group_plans` | 付费方案 |
| 群组 | `group_orders` | 群组订单 |
| 群组 | `group_admin_actions` | 圈主管理审计 |
| 主题 | `topics` | 帖子主表 |
| 主题 | `topic_comments` | 评论主表 |
| 主题 | `topic_reactions` | 点赞/收藏关系 |
| 主题 | `topic_audit_logs` | 内容审核日志 |

1级数据更新后必须发布领域事件，例如：

```text
MemberRegistered
MemberFollowed
MemberUnfollowed
GroupCreated
GroupJoined
GroupMemberRemoved
GroupPlanCreated
GroupOrderPaid
TopicCreated
TopicDeleted
TopicCommentCreated
TopicReacted
```

这些事件用于驱动 2级数据变化。

#### **2级数据存储规则**

2级数据可以存储在 MySQL 冗余字段、Redis、搜索索引、排行榜缓存、统计表中，但必须能从 1级数据重建。

典型 2级数据：

| 数据 | 来源 | 存储建议 |
|---|---|---|
| 用户粉丝数 | `user_follows` | Redis + `member_stats` |
| 用户关注数 | `user_follows` | Redis + `member_stats` |
| 用户发帖数 | `topics` | Redis + `member_stats` |
| 群组成员数 | `group_members` | Redis + `groups.members_count` |
| 群组帖子数 | `topics` | Redis + `groups.topics_count` |
| 帖子阅读数 | `topic_read_records` | Redis HyperLog/Counter + `topic_stats` |
| 帖子点赞数 | `topic_reactions` | Redis Counter + `topic_stats` |
| 帖子评论数 | `topic_comments` | Redis Counter + `topic_stats` |
| 圈主看板 | 多表聚合 | Redis Snapshot + 定时刷新 |
| 群组推荐列表 | 群组与活跃度 | Redis Sorted Set / 搜索索引 |
| 热门帖子列表 | 阅读、点赞、评论 | Redis Sorted Set |

这里建议保留 `groups.members_count`、`groups.topics_count` 作为冗余快照字段，但不要把它们当成唯一真实来源。真实来源仍然是 `group_members` 和 `topics`。

---

### **四、缓存策略设计**

缓存策略要跟数据等级绑定，不能所有数据都用一个 TTL。

#### **1级数据缓存策略**

1级数据是强一致主数据，缓存策略建议采用 **Cache Aside + 主动失效**。

| 数据类型 | 缓存 Key 示例 | TTL | 失效方式 |
|---|---|---:|---|
| 用户资料 | `member:profile:{userId}` | 10 \~ 30 分钟 | `UpdateMemberProfile` 后主动删除 |
| 群组详情 | `group:detail:{groupId}` | 5 \~ 15 分钟 | `UpdateGroup` / `AuditGroup` 后主动删除 |
| 付费方案 | `group:plans:{groupId}` | 10 \~ 30 分钟 | `CreateGroupPlan` / `UpdateGroupPlan` 后主动删除 |
| 帖子详情 | `topic:detail:{topicId}` | 5 \~ 15 分钟 | `UpdateTopic` / `DeleteTopic` / `AuditTopic` 后主动删除 |
| 成员关系 | `group:member:{groupId}:{userId}` | 5 \~ 10 分钟 | `JoinGroup` / `LeaveGroup` / `UpdateGroupMemberStatus` 后主动删除 |
| 关注关系 | `follow:rel:{followerId}:{followingId}` | 5 \~ 10 分钟 | `FollowMember` / `UnfollowMember` 后主动删除 |

对 1级数据，不建议使用过长 TTL，也不建议只靠 TTL 等自然过期。只靠 TTL 会导致用户刚被禁言、移除、封禁后仍然能访问部分内容，这是社交和付费场景的大风险。

#### **2级数据缓存策略**

2级数据允许最终一致，缓存策略建议采用 **事件驱动更新 + TTL 兜底 + 定时校准**。

| 数据类型 | 缓存 Key 示例 | TTL | 更新方式 |
|---|---|---:|---|
| 用户统计 | `member:stats:{userId}` | 30 \~ 120 分钟 | 关注/发帖事件增量更新 |
| 群组统计 | `group:stats:{groupId}` | 10 \~ 60 分钟 | 入群/退群/发帖事件增量更新 |
| 帖子统计 | `topic:stats:{topicId}` | 5 \~ 30 分钟 | 阅读/评论/点赞事件增量更新 |
| 群组帖子列表 | `group:topics:{groupId}:{page}` | 1 \~ 5 分钟 | 发帖/删帖后删除第一页和相关页 |
| 粉丝列表 | `member:followers:{userId}:{page}` | 1 \~ 10 分钟 | 关注/取关后删除前几页 |
| 圈主看板 | `owner:dashboard:{ownerId}` | 1 \~ 10 分钟 | 事件刷新 + 定时重建 |
| 热门主题 | `topic:hot:{groupId}` | 1 \~ 5 分钟 | Sorted Set 增量计算 |
| 推荐群组 | `group:recommend:{category}` | 5 \~ 30 分钟 | 定时任务重建 |

2级数据需要允许短时间不一致。例如用户刚关注圈主，粉丝数可以延迟几秒更新；但关注关系本身必须立即写入 `user_follows`，且 `IsFollowing` 判断要查 1级关系或短 TTL 关系缓存。

---

### **五、高权限操作后的通知与缓存失效**

你提到“被高权限的用户操作后通知”，这个点非常重要，建议单独作为架构规范。

高权限操作包括：

```text
UpdateMemberStatus
AuditGroup
UpdateGroupMemberStatus
ConfirmGroupPayment
AuditTopic
DeleteTopic
AdminRemoveComment
AdminBanMember
OwnerMuteMember
OwnerRemoveMember
```

这些操作必须触发三类动作：

```text
1. 写入审计日志
2. 发布领域事件
3. 主动失效相关缓存
```

例如圈主移除群成员：

```text
UpdateGroupMemberStatus
        ↓
更新 group_members.status = removed
        ↓
写入 group_admin_actions
        ↓
发布 GroupMemberRemoved 事件
        ↓
删除 group:member:{groupId}:{userId}
        ↓
删除 group:members:{groupId}:*
        ↓
更新/失效 group:stats:{groupId}
        ↓
通知被移除用户
```

建议通知分三类：

| 通知类型 | 用途 | 是否必须 |
|---|---|---|
| 站内通知 | 用户可见的系统消息 | MVP 建议保留 |
| WebSocket/IM 推送 | 在线即时提醒 | 可后续 |
| 领域事件 | 系统内部缓存、统计、审计使用 | 必须 |

---

### **六、架构关系设计**

在现有 CaiRobot MVP 架构下，建议把社交域拆成三个协议域，但底层可以先实现为一个 Go social module，未来再拆服务。

```text
Client / Admin Web / Proto Tester
        ↓
POST /api/hello
        ↓
MessagePacket
        ↓
maxType = 1000 / 2000 / 3000
        ↓
routes.yaml
        ↓
TarsGo LocalInvoker / TarsInvoker
        ↓
go/modules/social
        ├── member
        ├── group
        ├── topic
        ├── permission
        ├── event
        ├── cache
        └── repository
        ↓
MySQL / Redis / MQ / Search Index
```

推荐目录：

```text
go/modules/social/
  member/
    handler.go
    service.go
    usecase.go
    repository.go
  group/
    handler.go
    service.go
    usecase.go
    repository.go
  topic/
    handler.go
    service.go
    usecase.go
    repository.go
  permission/
    service.go
  cache/
    keys.go
    invalidator.go
    policy.go
  event/
    publisher.go
    subscriber.go
    events.go
  model/
    member.go
    group.go
    topic.go
  migration/
    *.sql
```

协议文件建议：

```text
proto/social/member.proto
proto/social/group.proto
proto/social/topic.proto
proto/social/common.proto
```

文档文件建议：

```text
docs/prd/PRD-social-app-mvp.md
docs/adr/ADR-social-data-level-and-cache-strategy.md
docs/api/social-openapi.yaml
docs/api/协议编号注册表.md
docs/wiki/CODE-WIKI.md
```

---

### **七、OpenAPI 协议设计规范**

OpenAPI 仍然只描述统一网关，不新增绕过网关的 REST 业务接口。

建议 OpenAPI 中写：

```yaml
openapi: 3.0.3
info:
  title: CaiRobot Social Domain API
  version: 1.0.0
  description: >
    CaiRobot MVP 社交域协议文档。所有业务请求统一通过 POST /api/hello 进入，
    使用 MessagePacket.maxType 区分成员、群组、主题协议组，
    使用 MessagePacket.minType 区分具体业务动作，
    payload 使用对应 Protobuf Request 序列化后的 bytes。

paths:
  /api/hello:
    post:
      summary: Unified MessagePacket Gateway
      operationId: callMessagePacket
      requestBody:
        required: true
        content:
          application/x-protobuf:
            schema:
              type: string
              format: binary
      responses:
        "200":
          description: MessagePacket response
          content:
            application/x-protobuf:
              schema:
                type: string
                format: binary

x-cairobot-protocol-groups:
  - maxType: 1000
    name: MemberProtocolGroup
    description: 用户、成员、关注粉丝、身份状态协议组
  - maxType: 2000
    name: GroupProtocolGroup
    description: 群组、付费方案、成员管理、圈主管理协议组
  - maxType: 3000
    name: TopicProtocolGroup
    description: 主题、帖子、阅读、评论、互动协议组
```

然后用扩展字段列出协议：

```yaml
x-cairobot-protocols:
  - name: MemberRegister
    maxType: 1000
    minType: 1001
    dataLevel: 1
    requestProto: cairobot.social.v1.MemberRegisterRequest
    responseProto: cairobot.social.v1.MemberRegisterResponse
    cachePolicy: no-cache
    events:
      - MemberRegistered

  - name: GetGroupDetail
    maxType: 2000
    minType: 2003
    dataLevel: mixed
    requestProto: cairobot.social.v1.GetGroupDetailRequest
    responseProto: cairobot.social.v1.GetGroupDetailResponse
    cachePolicy: cache-aside
    cacheKeys:
      - group:detail:{groupId}
      - group:stats:{groupId}

  - name: GetTopicDetail
    maxType: 3000
    minType: 3004
    dataLevel: mixed
    requestProto: cairobot.social.v1.GetTopicDetailRequest
    responseProto: cairobot.social.v1.GetTopicDetailResponse
    cachePolicy: cache-aside-with-permission-check
    permission:
      - CanReadTopic
```

---

### **八、社交关系模型优化**

建议把社交关系拆成三层，不要混在一起：

```text
关注关系：user_follows
群成员关系：group_members
内容互动关系：topic_reactions / topic_comments / topic_read_records
```

#### **关注关系**

关注关系属于用户对用户的关系，服务于粉丝体系、圈主主页、推荐分发。

```text
user_follows
- follower_id
- following_id
- status
- source
- created_at
- updated_at
```

关注/取关是 1级数据变化；粉丝数、关注数是 2级数据变化。

#### **群成员关系**

群成员关系属于用户对群组的关系，服务于入群、付费权限、圈主管理。

```text
group_members
- group_id
- user_id
- role
- status
- join_source
- expired_at
- muted_until
- created_at
- updated_at
```

入群、退群、移除、禁言、恢复是 1级数据变化；成员数、活跃成员数、付费成员数是 2级统计。

#### **内容互动关系**

内容互动关系属于用户对主题的行为。

```text
topic_read_records
topic_comments
topic_reactions
```

评论和点赞关系建议按 1级事实保存；阅读记录可以作为 2级行为数据处理。阅读数、点赞数、评论数都是 2级统计。

---

### **九、权限与缓存的关键约束**

权限判断不能依赖 2级缓存的统计结果。

例如：

```text
能否阅读付费帖子：必须查 group_members 有效权益，不看 group_stats.paid_member_count
能否管理群成员：必须查 group_members.role，不看 owner_dashboard 缓存
能否访问私密群：必须查 group_members.status，不看 group.members_count
能否修改帖子：必须查 topics.author_id 或 group_members.role
```

也就是说，2级缓存只能用于展示和列表排序，不能作为安全决策依据。安全决策必须来自 1级数据或 1级数据的短 TTL 关系缓存，并且高权限操作后必须主动失效。

---

### **十、给本地 Trae 的 PRD 撰写指令**

下面这段可以直接发给本地 Trae，让它按照项目规范生成 PRD 文档。

```text
你现在接手的是 CaiRobot MVP 运营管理综合平台项目，请按照当前 CODE-WIKI 的单网关、MessagePacket、Protobuf、routes.yaml、TarsCloud/TarsGo、多语言 Monorepo 架构规范，撰写社交 App MVP 的 PRD 需求文档。

本次任务不是直接写代码，而是先完成 PRD 和架构需求说明文档。

一、文档目标

请新增或更新以下文档：

1. docs/prd/PRD-social-app-mvp.md
2. docs/adr/ADR-social-data-level-and-cache-strategy.md
3. docs/api/social-openapi.yaml
4. docs/api/协议编号注册表.md
5. docs/wiki/CODE-WIKI.md 中增加社交域、数据等级、缓存策略、协议组说明

二、项目背景

CaiRobot MVP 是一个运营管理综合平台，当前采用：

1. 单网关入口：POST /api/hello
2. MessagePacket：使用 maxType/minType 表示协议身份
3. Protobuf：业务请求和响应使用 Protobuf 定义
4. routes.yaml：负责协议号到内部服务目标的路由映射
5. TarsCloud/TarsGo：内部服务治理
6. Go + TypeScript + Python 多语言 Monorepo
7. Go 模块按 handler/service/usecase/repository 分层

本次 PRD 要设计一个社交 App MVP，支持用户注册、付费群组、阅读帖子、用户关注圈主、圈主管理粉丝和成员。

三、基础数据模型

请参考已有 basemodel.md 中的基础表：

1. users：用户主表
2. groups：群组/圈子主表
3. topics：主题/帖子主表

在 PRD 中说明这三张表是社交域基础模型，并补充建议新增的数据表：

1. user_follows：用户关注/粉丝关系表
2. group_members：群组成员关系表
3. group_plans：付费群组方案表
4. group_orders：群组付费订单表
5. topic_read_records：帖子阅读记录表
6. topic_comments：帖子评论表
7. topic_reactions：帖子点赞/收藏/分享互动表
8. group_admin_actions：圈主管理操作日志表
9. member_stats：成员统计快照表，可选
10. group_stats：群组统计快照表，可选
11. topic_stats：主题统计快照表，可选

四、数据等级规范

请在 PRD 中定义数据等级：

1级数据：
- 只有通过明确业务协议主动操作后才变化；
- 属于业务事实；
- 必须落 MySQL 主库；
- 必须强一致；
- 必须有审计或事件记录；
- 示例：users、groups、topics、user_follows、group_members、group_plans、group_orders、topic_comments、topic_reactions。

2级数据：
- 由 1级数据变更、统计任务、阅读互动、高权限操作、系统事件被动触发变化；
- 属于派生状态、统计、列表、缓存视图；
- 允许最终一致；
- 必须能从 1级数据重新计算；
- 示例：粉丝数、关注数、成员数、帖子数、阅读数、点赞数、评论数、圈主看板、热门帖子、推荐群组。

请明确规则：
- 2级数据不能作为权限判断依据；
- 2级数据丢失后必须可以重建；
- 高权限操作必须触发缓存失效和通知事件；
- 1级数据更新后必须发布领域事件驱动 2级数据更新。

五、协议组设计

请按照以下协议组规划 PRD：

成员协议组：
- maxType = 1000
- 负责用户注册、登录、资料、关注、粉丝、成员状态。

群组协议组：
- maxType = 2000
- 负责群组创建、群组详情、入群、付费方案、订单、圈主管理成员。

主题协议组：
- maxType = 3000
- 负责帖子发布、帖子详情、帖子列表、阅读记录、评论、点赞、内容审核。

请在 docs/api/协议编号注册表.md 中规划以下 minType，若已有冲突则重新分配：

成员协议组：
1001 MemberRegister
1002 MemberLogin
1003 MemberLogout
1004 GetMemberProfile
1005 UpdateMemberProfile
1010 FollowMember
1011 UnfollowMember
1012 ListFollowers
1013 ListFollowings
1020 GetMemberStats
1030 UpdateMemberStatus

群组协议组：
2001 CreateGroup
2002 UpdateGroup
2003 GetGroupDetail
2004 ListGroups
2010 JoinGroup
2011 LeaveGroup
2012 ListGroupMembers
2013 UpdateGroupMemberStatus
2020 CreateGroupPlan
2021 UpdateGroupPlan
2022 ListGroupPlans
2030 CreateGroupOrder
2031 ConfirmGroupPayment
2040 GetOwnerDashboard
2050 AuditGroup

主题协议组：
3001 CreateTopic
3002 UpdateTopic
3003 DeleteTopic
3004 GetTopicDetail
3005 ListGroupTopics
3010 MarkTopicRead
3011 GetReadHistory
3020 CreateComment
3021 DeleteComment
3030 ReactTopic
3031 CancelReactTopic
3040 AuditTopic
3050 GetTopicStats

六、缓存策略设计

请在 PRD 和 ADR 中定义缓存策略：

1级数据缓存：
- 使用 Cache Aside；
- TTL 较短；
- 协议写入后主动失效；
- 适用于用户资料、群组详情、帖子详情、群成员关系、关注关系、付费方案。

2级数据缓存：
- 使用事件驱动更新 + TTL 兜底 + 定时校准；
- 允许最终一致；
- 适用于成员统计、群组统计、帖子统计、粉丝列表、群组帖子列表、圈主看板、热门帖子、推荐群组。

请设计缓存 key 命名规范：

member:profile:{userId}
member:stats:{userId}
member:followers:{userId}:{page}
follow:rel:{followerId}:{followingId}

group:detail:{groupId}
group:member:{groupId}:{userId}
group:members:{groupId}:{page}
group:plans:{groupId}
group:stats:{groupId}
owner:dashboard:{ownerId}

topic:detail:{topicId}
topic:stats:{topicId}
topic:read:{userId}:{topicId}
group:topics:{groupId}:{page}
topic:hot:{groupId}

七、高权限操作规范

请在 PRD 中定义高权限操作必须执行：

1. 权限校验
2. 写入审计日志
3. 更新 1级数据
4. 发布领域事件
5. 主动失效相关缓存
6. 通知受影响用户或系统模块

高权限操作包括：
- UpdateMemberStatus
- AuditGroup
- UpdateGroupMemberStatus
- ConfirmGroupPayment
- AuditTopic
- DeleteTopic
- AdminRemoveComment
- OwnerMuteMember
- OwnerRemoveMember

八、权限规则

请在 PRD 中定义统一权限服务，不允许各 handler 自行散写权限判断。

权限能力包括：

CanViewGroup(userID, groupID)
CanJoinGroup(userID, groupID)
CanReadTopic(userID, topicID)
CanManageGroup(operatorID, groupID)
CanManageMember(operatorID, groupID, targetUserID)
CanPublishTopic(userID, groupID)
CanAuditContent(operatorID)

请明确：
- 权限判断不得依赖 2级统计缓存；
- 付费帖子阅读必须检查 group_members 有效权益；
- 圈主管理成员必须检查 group_members.role；
- 管理员审核必须检查平台管理权限；
- 被禁言用户不得发布帖子或评论；
- 被移除或过期成员不得阅读成员/付费内容。

九、OpenAPI 文档要求

请新增 docs/api/social-openapi.yaml。

OpenAPI 只描述统一网关入口 POST /api/hello，不得新增绕过网关的 REST 业务接口。

需要包含：

1. paths./api/hello.post
2. x-cairobot-protocol-groups
3. x-cairobot-protocols
4. 每个协议的 maxType、minType、dataLevel、requestProto、responseProto、cachePolicy、events、permission。

十、PRD 文档结构要求

docs/prd/PRD-social-app-mvp.md 请按以下结构撰写：

1. 背景与目标
2. 当前项目架构约束
3. 社交 App MVP 范围
4. 用户角色与使用场景
5. 核心业务流程
6. 基础数据模型
7. 数据等级规范
8. 协议组规划
9. 功能需求
10. 权限规则
11. 缓存策略
12. 高权限操作与通知机制
13. OpenAPI 协议映射
14. 非功能需求
15. 测试验收标准
16. 后续迭代范围

十一、验收标准

完成后请确保：

1. PRD 明确说明 1级数据和 2级数据的边界。
2. PRD 明确说明成员协议组 1000、群组协议组 2000、主题协议组 3000。
3. PRD 明确所有外部请求仍然走 POST /api/hello。
4. PRD 明确 OpenAPI 不新增真实业务 REST 路径。
5. PRD 明确 Protobuf Request/Response 由 proto/social/*.proto 承载。
6. PRD 明确 routes.yaml 需要映射 maxType/minType 到 TarsGo 服务。
7. PRD 明确缓存 key、TTL、主动失效、事件驱动更新策略。
8. PRD 明确高权限操作后的审计、事件、缓存失效和通知。
9. PRD 明确权限判断不能依赖 2级统计缓存。
10. PRD 明确测试验收标准。
```

---

### **十一、建议写入 PRD 的核心业务流程**

PRD 里最好把几个关键流程写清楚，方便后续 Trae 开发时不跑偏。

#### **用户关注圈主**

```text
FollowMember
        ↓
写入 user_follows，1级数据
        ↓
发布 MemberFollowed 事件
        ↓
删除 follow:rel:{followerId}:{followingId}
        ↓
更新 member:stats:{followingId}.followers_count
        ↓
删除 member:followers:{followingId}:*
        ↓
给圈主发送粉丝新增通知
```

#### **用户加入付费群组**

```text
CreateGroupOrder
        ↓
创建 group_orders，状态 pending
        ↓
支付完成
        ↓
ConfirmGroupPayment
        ↓
更新 group_orders = paid
        ↓
新增或续期 group_members
        ↓
发布 GroupOrderPaid / GroupMemberActivated
        ↓
删除 group:member:{groupId}:{userId}
        ↓
更新 group:stats:{groupId}.paid_members_count
        ↓
通知用户权益已开通
```

#### **用户阅读付费帖子**

```text
GetTopicDetail
        ↓
读取 topic 1级数据
        ↓
CanReadTopic 权限判断
        ↓
检查 group_members.status 和 expired_at
        ↓
有权限：返回完整内容
无权限：只返回摘要和购买提示
        ↓
MarkTopicRead 可异步写入阅读行为
        ↓
更新 topic:stats:{topicId}.read_count
```

#### **圈主移除成员**

```text
UpdateGroupMemberStatus
        ↓
CanManageMember 权限判断
        ↓
更新 group_members.status = removed
        ↓
写入 group_admin_actions
        ↓
发布 GroupMemberRemoved
        ↓
删除 group:member:{groupId}:{userId}
        ↓
删除 group:members:{groupId}:*
        ↓
更新 group:stats:{groupId}
        ↓
通知被移除用户
```

---

### **十二、最终主控口径**

这次优化可以定成 CaiRobot MVP 社交域的核心工程原则：

> 以协议组划分业务边界，以数据等级划分一致性边界。成员协议组 1000、群组协议组 2000、主题协议组 3000 分别承载用户关系、群组运营、内容互动。1级数据必须通过明确 Protobuf 协议主动变更，并作为业务事实强一致落库；2级数据由事件、统计任务和高权限操作被动变化，作为缓存、统计、列表和看板使用，允许最终一致但必须可重建。所有外部访问仍统一通过 `POST /api/hello`，由 `MessagePacket.maxType/minType` 路由到 TarsGo 服务，OpenAPI 只作为协议映射文档，不新增绕过网关的业务 REST 接口。

这个口径可以直接写进 PRD、ADR 和 CODE-WIKI，后续 Trae 只要按这个规则写文档和代码，就不会偏离当前平台架构。

*内容由 AI 生成仅供参考*