# PRD-social-app-mvp.md

> **文档编号**: PRD-20
> **版本**: v1.0-draft
> **创建日期**: 2026-06-15
> **状态**: 草案
> **相关 ADR**: [ADR-social-data-level-and-cache-strategy](../adr/ADR-social-data-level-and-cache-strategy.md)
> **相关 API**: [social-openapi.yaml](../api/social-openapi.yaml)
> **相关注册表**: [协议编号注册表](../api/协议编号注册表.md)

***

## 1. 背景与目标

### 1.1 背景

CaiRobot MVP 是一个运营管理综合平台，当前已完成系统基础能力（HealthCheck、HelloWorld）、全局配置服务（Config）、多语言服务（I18n）和 Admin 管理后台。项目采用单网关 + MessagePacket + Protobuf + TarsCloud/TarsGo 的统一架构。

平台需要扩展社交 App 能力，支持用户注册登录、付费群组、帖子阅读、用户关注圈主、圈主管理粉丝和成员等核心社交功能。这些功能将基于已有的 `users`、`groups`、`topics` 三张基础数据表进行扩展。

### 1.2 目标

在现有 CaiRobot MVP 架构约束下，设计并定义社交域的完整需求规范：

1. 建立以**协议组**划分的业务边界（成员 1000 / 群组 2000 / 主题 3000）
2. 建立**数据等级**治理体系（1级强一致事实数据 / 2级最终一致派生数据）
3. 定义全部业务协议接口（37 个 minType）
4. 定义缓存策略和领域事件驱动机制
5. 定义统一权限服务体系

### 1.3 非目标

- 不涉及即时通讯（IM）/聊天功能
- 不涉及推荐算法/Feed 流排序
- 不涉及支付渠道对接（MVP 首期使用管理员手动开通权益）
- 不涉及内容审核 AI 能力（首期仅人工审核）
- 不涉及分销/优惠券/创作者收益结算

***

## 2. 当前项目架构约束

### 2.1 强制架构约束

社交域必须严格遵循以下架构规则，不得偏离：

| 约束项     | 规则                                                     | 来源                                            |
| ------- | ------------------------------------------------------ | --------------------------------------------- |
| 网关入口    | 所有外部请求统一走 `POST /api/hello`                            | [CODE-WIKI §2](../wiki/CODE-WIKI.md)          |
| 报文格式    | 使用 MessagePacket（maxType/minType/extend/platform/data） | [CODE-WIKI §3](../wiki/CODE-WIKI.md)          |
| 协议定义    | 业务字段使用 Protobuf 定义，Request/Response 分离                 | [ADR-0003](../adr/ADR-0003-服务协议使用Protobuf.md) |
| 路由映射    | maxType/minType → Tars 目标通过 routes.yaml 配置             | [CODE-WIKI §5](../wiki/CODE-WIKI.md)          |
| 内部契约    | TarsGo 标准 bytes 接口签名                                   | [CODE-WIKI §7](../wiki/CODE-WIKI.md)          |
| Go 分层   | handler → usecase → service → repository               | [coding.md](../../.trae/rules/coding.md)      |
| OpenAPI | 只描述统一网关入口，不新增绕过网关的 REST 接口                             | 本 PRD §13                                     |
| 模块接入    | 遵循 common-lib/module.Deps + health.Checker + SDK       | [CODE-WIKI §9.5](../wiki/CODE-WIKI.md)        |
| 编号唯一性   | max+min 全局唯一，必须在注册表登记                                  | [协议编号注册表](../api/协议编号注册表.md)                  |

### 2.2 已有资产复用

| 资产         | 路径                                                   | 复用方式              |
| ---------- | ---------------------------------------------------- | ----------------- |
| users 表    | basemodel.md                                         | 作为用户身份主表，不堆砌社交关系  |
| groups 表   | basemodel.md                                         | 作为群组/圈子主表，不保存动态关系 |
| topics 表   | basemodel.md                                         | 作为帖子主表，阅读权限由关联表判断 |
| common-lib | go/common-lib/                                       | 复用错误码、类型定义        |
| Gateway    | go/gateway/proto-gateway/                            | 复用网关路由分发能力        |
| Tars 骨架    | proto/generated/tarsgo/CaiRobot{Auth,UserCenter}App/ | 参考或改造为社交域 Tars 服务 |
| Admin 后台   | go/admin/                                            | 扩展社交域运营管理插件       |

### 2.3 与预留模块的关系

CODE-WIKI 中已规划以下 MVP2 预留模块：

| 预留模块             | 与社交域的关系     | 处理策略                         |
| ---------------- | ----------- | ---------------------------- |
| `modules/users`  | 社交域依赖用户身份能力 | 社交域 member 子模块调用 users 接口    |
| `modules/auth`   | 社交域依赖认证鉴权能力 | 社交域通过 extend.token 走 auth 校验 |
| `modules/groups` | 社交域群组功能高度重叠 | 合并为社交域 group 子模块             |
| `modules/topics` | 社交域主题功能高度重叠 | 合并为社交域 topic 子模块             |

**决策：社交域作为独立模块** **`go/modules/social`** **实现，内部按 member/group/topic 拆分子包。预留的 users/auth modules 由社交域通过接口依赖，groups/topics 预留模块合并入 social 模块。**

***

## 3. 社交 App MVP 范围

### 3.1 产品定位

一个以"圈子/群组"为核心的内容型社交产品。普通用户可以注册账号、浏览圈子、加入免费群组或购买付费群组权限；圈主可以创建和运营自己的群组，发布帖子，管理粉丝、成员、内容和付费方案；运营后台可以管理用户、群组、帖子、订单、举报、审核和平台配置。

### 3.2 用户角色

| 角色    | 核心能力                        | 数据边界             |
| ----- | --------------------------- | ---------------- |
| 游客    | 浏览公开圈子、查看公开帖子摘要、注册登录        | 不能访问付费内容         |
| 普通用户  | 注册登录、加入群组、阅读帖子、评论互动、关注圈主    | 只能访问自己有权限的内容     |
| 圈主    | 创建/管理群组、发布帖子、管理成员和粉丝、配置付费方案 | 只能管理自己拥有或授权管理的群组 |
| 平台管理员 | 用户管理、群组审核、帖子审核、订单管理、风控配置    | 通过运营后台管理全局数据     |

### 3.3 核心概念区分

以下三组概念必须分表建模，不能混用：

| 概念    | 含义         | 数据表                                                       | 协议组        |
| ----- | ---------- | --------------------------------------------------------- | ---------- |
| 关注关系  | 用户→用户的社交关注 | user\_follows                                             | 成员协议组 1000 |
| 群成员关系 | 用户→群组的加入关系 | group\_members                                            | 群组协议组 2000 |
| 付费权益  | 用户→群组的交易权益 | group\_orders + group\_members.expired\_at                | 群组协议组 2000 |
| 内容互动  | 用户→帖子的行为   | topic\_reactions / topic\_comments / topic\_read\_records | 主题协议组 3000 |

***

## 4. 用户角色与使用场景

### 4.1 核心使用场景

#### 场景 1：用户注册与登录

```
游客 → 输入注册信息 → MemberRegister → 创建 users 记录 → 返回 token 和用户资料
已注册用户 → 输入凭据 → MemberLogin → 校验凭证 → 返回 token 和用户资料
```

#### 场景 2：用户关注圈主

```
普通用户 → 进入圈主主页 → 点击关注 → FollowMember → 写入 user_follows(1级)
→ 发布 MemberFollowed 事件 → 更新粉丝数(2级) → 通知圈主
```

#### 场景 3：用户加入付费群组

```
普通用户 → 浏览群组详情 → 选择付费方案 → CreateGroupOrder → 创建订单(1级)
→ 支付完成 → ConfirmGroupPayment → 更新订单=paid(1级) → 新增/续期 group_members(1级)
→ 发布 GroupOrderPaid 事件 → 更新群统计(2级) → 通知用户
```

#### 场景 4：用户阅读付费帖子

```
已登录用户 → 进入帖子详情 → GetTopicDetail → 读取 topics(1级)
→ CanReadTopic 权限判断 → 检查 group_members.status + expired_at
→ 有权限: 返回完整内容
→ 无权限: 返回摘要 + 购买提示
→ 异步 MarkTopicRead → 写入阅读记录(2级) → 更新阅读数(2级)
```

#### 场景 5：圈主管理成员

```
圈主 → 进入成员管理页 → 选择成员 → 操作(禁言/移除/恢复)
→ CanManageMember 权限判断 → UpdateGroupMemberStatus
→ 更新 group_members.status(1级) → 写入 group_admin_actions(1级)
→ 发布 GroupMemberRemoved 事件 → 失效缓存 → 通知被操作用户
```

### 4.2 MVP 第一期范围（P0）

| 功能               | 优先级 | 是否纳入首期   |
| ---------------- | --- | -------- |
| 用户注册/登录          | P0  | 必须       |
| 用户资料查询与修改        | P0  | 必须       |
| 创建群组             | P0  | 必须       |
| 群组列表/详情          | P0  | 必须       |
| 加入免费群组           | P0  | 必须       |
| 创建付费方案           | P0  | 必须       |
| 创建订单并开通权益（手动确认）  | P0  | 必须（简化支付） |
| 发布帖子             | P0  | 必须       |
| 帖子列表/详情          | P0  | 必须       |
| 付费帖子阅读权限判断       | P0  | 必须       |
| 用户关注/取关圈主        | P0  | 必须       |
| 圈主查看粉丝与群成员       | P0  | 必须       |
| 圈主管理成员（禁言/移除/恢复） | P1  | 建议       |
| 评论/点赞/收藏         | P1  | 建议       |
| 阅读记录             | P1  | 建议       |
| 平台审核（群组/帖子）      | P1  | 推荐       |
| 推荐流/搜索/消息通知      | P2  | 后续迭代     |
| 分销/优惠券/收益结算      | P2  | 后续迭代     |

***

## 5. 核心业务流程

### 5.1 用户关注圈主流程

```text
Client                    Gateway              Social Service           MySQL          Redis
  |                          |                      |                     |             |
  |-- FollowMember ---------->|                      |                     |             |
  | (maxType=1000,min=1011)   |                      |                     |             |
  |                          |-- routes.yaml ------->|                     |             |
  |                          |                      |-- BEGIN TX -------->|             |
  |                          |                      |-- INSERT user_follows            |
  |                          |                      |<-- TX OK -----------|             |
  |                          |                      |-- publish Event ----|--> update  |
  |                          |                      |   MemberFollowed     |    stats    |
  |                          |                      |-- DEL cache key ----------------->|
  |                          |                      |   follow:rel:{f}:{t}               |
  |<-- Response -------------|<-- encode protobuf ---|                     |             |
```

**数据等级标注**：

- INSERT user\_follows → **1级数据**（强一致事务写入）
- 更新粉丝数缓存 → **2级数据**（事件驱动异步更新）
- 删除关注关系缓存 → Cache Aside 主动失效

### 5.2 用户加入付费群组流程

```text
Client                    Gateway              Social Service           MySQL          Redis
  |                          |                      |                     |             |
  |-- CreateGroupOrder ----->|                      |                     |             |
  | (maxType=2000,min=2019)   |                      |                     |             |
  |                          |                      |-- INSERT group_orders            |
  |                          |                      | (status=pending)     |             |
  |<-- OrderCreated ---------|<----------------------|                     |             |
  |                          |                      |                     |             |
  |-- ConfirmGroupPayment -->|                      |                     |             |
  | (maxType=2000,min=2021)   |                      |                     |             |
  |                          |                      |-- BEGIN TX -------->|             |
  |                          |                      |-- UPDATE orders      |             |
  |                          |                      |   status=paid        |             |
  |                          |                      |-- UPSERT members     |             |
  |                          |                      |   (expired_at)       |             |
  |                          |                      |<-- TX OK -----------|             |
  |                          |                      |-- publish Event ----|--> update  |
  |                          |                      |   GroupOrderPaid     |    stats    |
  |                          |                      |-- DEL cache keys --->|             |
  |<-- PaymentConfirmed -----|<----------------------|                     |             |
```

**关键约束**：订单状态变更和成员权益更新必须在同一事务内完成，保证一致性。

### 5.3 用户阅读付费帖子流程

```text
Client                    Gateway              Social Service         Permission      MySQL
  |                          |                      |                   Service          |
  |-- GetTopicDetail ------->|                      |                   |                |
  | (maxType=3000,min=3003)   |                      |                   |                |
  |                          |                      |-- load topic ------|--> 1级查询     |
  |                          |                      |-- CanReadTopic? --|                |
  |                          |                      |   |-- check group_members  |        |
  |                          |                      |   |   status=active        |        |
  |                          |                      |   |-- check expired_at     |        |
  |                          |                      |   |   > now()              |        |
  |                          |                      |<-- permission result -|                |
  |                          |                      |-- if allowed:        |                |
  |                          |                      |   return full content|                |
  |                          |                      |-- else:              |                |
  |                          |                      |   return summary      |                |
  |                          |                      |   + purchase hint    |                |
  |<-- TopicDetail ----------|<-- (with can_read flag)                   |                |
  |                          |                      |                                |
  |-- [async] MarkTopicRead ->|                      |                   |                |
  | (maxType=3000,min=3011)   |                      |-- async write ---->|--> 2级写入     |
  |                          |                      |   topic_read_records |                |
```

**权限铁律**：CanReadTopic 判断只能查 1级数据（group\_members），不能依赖 2级缓存（group\_stats.paid\_member\_count）。

### 5.4 圈主移除群成员流程

```text
Operator                  Gateway            Social Service        Perm Svc        MySQL        Redis
  |                           |                    |                 |              |           |
  |--UpdateGroupMemberStatus->|                    |                 |              |           |
  | (maxType=2000,min=2013)    |                    |                 |              |           |
  |                           |                    |--CanManageMember|              |           |
  |                           |                    |  check role ----|              |           |
  |                           |                    |  check ownership |              |           |
  |                           |                    |<--allowed ------|              |           |
  |                           |                    |--BEGIN TX ----->|              |           |
  |                           |                    |--UPDATE members  |              |           |
  |                           |                    |  status=removed  |              |           |
  |                           |                    |--INSERT audit ->|              |           |
  |                           |                    |  admin_actions   |              |           |
  |                           |                    |<--TX OK --------|              |           |
  |                           |                    |--publish Event --|-->update stats|
  |                           |                    |  GroupMemberRemoved                |
  |                           |                    |--DEL cache keys -------------->|
  |                           |                    |--notify target user              |
  |<--Updated ----------------|<--------------------|                 |              |           |
```

**高权限操作四件套**：权限校验 → 审计日志 → 1级数据更新 → 事件发布 + 缓存失效 + 通知

***

## 6. 基础数据模型

### 6.1 已有基础表（来自 basemodel.md）

#### users — 用户身份主表

| 字段                | 类型              | 说明                             | 数据等级 |
| ----------------- | --------------- | ------------------------------ | ---- |
| id                | char(32) PK     | 内部主键，全系统关联                     | 1级   |
| username          | varchar(50) UK  | 登录用户名                          | 1级   |
| password          | varchar(255)    | 加密密码                           | 1级   |
| email             | varchar(100) UK | 邮箱                             | 1级   |
| phone             | varchar(20) UK  | 手机号                            | 1级   |
| nickname          | varchar(50)     | 昵称                             | 1级   |
| avatar            | varchar(255)    | 头像 URL                         | 1级   |
| status            | enum            | active/inactive/banned/deleted | 1级   |
| membership\_level | varchar(32)     | 平台会员等级（≠群组付费会员）                | 1级   |
| created\_at       | bigint          | 创建时间                           | 1级   |
| updated\_at       | bigint          | 更新时间                           | 1级   |

**不在 users 中增加的字段**：followers\_count、following\_count、group\_count 等频繁变化统计字段。

#### groups — 群组/圈子主表

| 字段             | 类型              | 说明                     | 数据等级 |
| -------------- | --------------- | ---------------------- | ---- |
| id             | char(32) PK     | 主键                     | 1级   |
| name           | varchar(100)    | 群组名称                   | 1级   |
| slug           | varchar(100) UK | URL 标识                 | 1级   |
| owner\_id      | char(32) FK     | 圈主用户 ID                | 1级   |
| type           | varchar(20)     | free/paid/mixed/invite | 1级   |
| visibility     | tinyint         | 1=公开 2=链接可见 3=私密       | 1级   |
| join\_mode     | tinyint         | 1=直接 2=审核 3=付费 4=邀请    | 1级   |
| status         | tinyint         | 状态                     | 1级   |
| members\_count | bigint          | 成员数冗余（2级快照）            | 2级   |
| topics\_count  | bigint          | 帖子数冗余（2级快照）            | 2级   |
| max\_members   | bigint          | 人数上限                   | 1级   |
| created\_at    | bigint          | 创建时间                   | 1级   |
| updated\_at    | bigint          | 更新时间                   | 1级   |

**type 枚举值定义**：

| 值      | 含义               |
| ------ | ---------------- |
| free   | 免费圈子，所有内容公开      |
| paid   | 付费圈子，需购买方案才能加入   |
| mixed  | 混合圈子，部分免费 + 部分付费 |
| invite | 邀请制圈子，仅邀请加入      |

**join\_mode 枚举值定义**：

| 值 | 含义    |
| - | ----- |
| 1 | 直接加入  |
| 2 | 申请后审核 |
| 3 | 付费加入  |
| 4 | 仅邀请加入 |

#### topics — 帖子主表

| 字段                 | 类型           | 说明                   | 数据等级 |
| ------------------ | ------------ | -------------------- | ---- |
| id                 | char(32) PK  | 主键                   | 1级   |
| title              | varchar(200) | 标题                   | 1级   |
| content            | longtext     | 正文内容                 | 1级   |
| summary            | varchar(500) | 摘要（无权限时返回此字段）        | 1级   |
| author\_id         | char(36) FK  | 作者 ID                | 1级   |
| group\_id          | char(36) FK  | 所属群组 ID              | 1级   |
| type               | tinyint      | 帖子类型                 | 1级   |
| status             | tinyint      | 草稿/待审核/已发布/已下架/已删除   | 1级   |
| visibility         | tinyint      | 公开/群成员可见/付费成员可见/圈主私密 | 1级   |
| is\_pinned         | tinyint      | 是否置顶                 | 1级   |
| is\_featured       | tinyint      | 是否精选                 | 1级   |
| published\_at      | bigint       | 发布时间                 | 1级   |
| last\_activity\_at | bigint       | 最后活跃时间               | 2级   |
| cover\_image       | varchar(500) | 封面图                  | 1级   |
| created\_at        | bigint       | 创建时间                 | 1级   |
| updated\_at        | bigint       | 更新时间                 | 1级   |

**visibility 枚举值定义**：

| 值 | 含义            | 权限要求      |
| - | ------------- | --------- |
| 1 | PUBLIC        | 所有人可读     |
| 2 | GROUP\_MEMBER | 群组成员可读    |
| 3 | PAID\_MEMBER  | 付费成员可读    |
| 4 | OWNER\_ONLY   | 仅圈主和管理员可读 |

### 6.2 新增数据表

#### 6.2.1 user\_follows — 用户关注/粉丝关系表（1级）

表达用户对用户的单向关注关系，服务于粉丝体系、圈主主页、推荐分发。

```sql
CREATE TABLE `user_follows` (
  `id` char(32) NOT NULL COMMENT '主键 UUID',
  `follower_id` char(32) NOT NULL COMMENT '关注者 ID',
  `following_id` char(32) NOT NULL COMMENT '被关注者 ID（通常是圈主/作者）',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=cancelled 3=blocked',
  `source` varchar(30) DEFAULT NULL COMMENT '来源: group/topic/profile/search',
  `created_at` bigint(20) NOT NULL COMMENT '关注时间',
  `updated_at` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follow` (`follower_id`, `following_id`) COMMENT '防重复关注',
  KEY `idx_following_status_time` (`following_id`, `status`, `created_at`),
  KEY `idx_follower_status_time` (`follower_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关注/粉丝关系表（1级数据）';
```

**注意**：圈主的"粉丝"来自 user\_follows，圈主的"群成员"来自 group\_members，两者不可混用。

#### 6.2.2 group\_members — 群组成员关系表（1级）

核心关系表，用于判断用户是否属于某个群组、在群组内的身份、是否被禁言、付费权益是否有效。

```sql
CREATE TABLE `group_members` (
  `id` char(32) NOT NULL COMMENT '主键 UUID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `role` varchar(20) NOT NULL DEFAULT 'member' COMMENT 'owner/admin/moderator/member',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=pending 3=banned 4=left 5=expired',
  `join_source` varchar(20) DEFAULT NULL COMMENT 'free/paid/invite/admin/import',
  `joined_at` bigint(20) DEFAULT NULL COMMENT '加入时间',
  `expired_at` bigint(20) DEFAULT NULL COMMENT '付费权益过期时间（免费群为空）',
  `muted_until` bigint(20) DEFAULT NULL COMMENT '禁言截止时间戳',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`) COMMENT '一个用户在一个群组只有一条记录',
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_group_status_role` (`group_id`, `status`, `role`),
  KEY `idx_group_expired` (`group_id`, `expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员关系表（1级数据）';
```

#### 6.2.3 group\_plans — 付费群组方案表（1级）

定义群组的付费方案（月卡/季卡/年卡/终身），包括价格、周期和权益说明。

```sql
CREATE TABLE `group_plans` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL COMMENT '所属群组',
  `name` varchar(100) NOT NULL COMMENT '方案名称: 月卡/季卡/年卡',
  `plan_type` varchar(20) NOT NULL COMMENT 'monthly/quarterly/yearly/lifetime',
  `price_cent` bigint(20) NOT NULL DEFAULT '0' COMMENT '价格（单位：分）',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `duration_days` int(11) NOT NULL DEFAULT '30' COMMENT '权益天数',
  `benefits` json DEFAULT NULL COMMENT '权益说明 JSON',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=上架 2=下架',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_status` (`group_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='付费群组方案表（1级数据）';
```

#### 6.2.4 group\_orders — 群组付费订单表（1级）

记录用户购买行为，连接用户、群组和付费方案。支付成功后由业务服务在同一事务内更新 group\_members 的 expired\_at。

```sql
CREATE TABLE `group_orders` (
  `id` char(32) NOT NULL,
  `order_no` varchar(64) NOT NULL COMMENT '对外订单号',
  `user_id` char(32) NOT NULL COMMENT '购买用户',
  `group_id` char(32) NOT NULL COMMENT '购买群组',
  `plan_id` char(32) DEFAULT NULL COMMENT '购买的方案',
  `amount_cent` bigint(20) NOT NULL DEFAULT '0' COMMENT '实付金额（分）',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=pending 2=paid 3=cancelled 4=refunded 5=failed',
  `pay_channel` varchar(30) DEFAULT NULL COMMENT '支付渠道',
  `paid_at` bigint(20) DEFAULT NULL COMMENT '支付时间',
  `expired_at` bigint(20) DEFAULT NULL COMMENT '权益过期时间',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status_time` (`user_id`, `status`, `created_at`),
  KEY `idx_group_status_time` (`group_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组付费订单表（1级数据）';
```

#### 6.2.5 topic\_read\_records — 帖子阅读记录表（2级行为数据）

记录用户阅读行为，用于已读状态、阅读历史、内容推荐和圈主数据分析。允许最终一致，允许异步写入。

```sql
CREATE TABLE `topic_read_records` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `user_id` char(32) NOT NULL COMMENT '阅读用户',
  `group_id` char(32) DEFAULT NULL COMMENT '冗余群组 ID（便于按群组查询）',
  `read_at` bigint(20) DEFAULT NULL COMMENT '最近阅读时间',
  `read_count` bigint(20) DEFAULT '0' COMMENT '阅读次数',
  `duration_sec` int(11) DEFAULT '0' COMMENT '累计阅读时长（秒）',
  `progress` int(11) DEFAULT '0' COMMENT '阅读进度 0-100',
  `created_at` bigint(20) NOT NULL COMMENT '首次阅读时间',
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_topic` (`user_id`, `topic_id`) COMMENT '同一用户同一帖子一条记录',
  KEY `idx_topic_read_time` (`topic_id`, `read_at`),
  KEY `idx_user_read_time` (`user_id`, `read_at`),
  KEY `idx_group_read_time` (`group_id`, `read_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子阅读记录表（2级行为数据）';
```

#### 6.2.6 topic\_comments — 帖子评论表（1级）

用户生成内容，需要持久化、审核和审计。

```sql
CREATE TABLE `topic_comments` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `user_id` char(32) NOT NULL COMMENT '评论用户',
  `parent_id` char(32) DEFAULT NULL COMMENT '父评论 ID（支持楼中楼）',
  `content` text NOT NULL COMMENT '评论内容',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=正常 2=审核中 3=隐藏 4=删除',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_topic_status` (`topic_id`, `status`, `created_at`),
  KEY `idx_user_comments` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子评论表（1级数据）';
```

#### 6.2.7 topic\_reactions — 帖子互动表（1级）

点赞、收藏、分享等互动行为的统一关系表。关系本身是 1级事实数据；点赞数/收藏数等聚合是 2级统计数据。

```sql
CREATE TABLE `topic_reactions` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL COMMENT '帖子 ID',
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `reaction_type` varchar(20) NOT NULL COMMENT 'like/favorite/share',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_topic_reaction` (`user_id`, `topic_id`, `reaction_type`),
  KEY `idx_topic_reaction` (`topic_id`, `reaction_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子互动表（1级数据）';
```

#### 6.2.8 group\_admin\_actions — 圈主管理操作日志表（1级）

记录圈主/管理员的管理操作，用于审计和运营分析。

```sql
CREATE TABLE `group_admin_actions` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `operator_id` char(32) NOT NULL COMMENT '操作人 ID（圈主/管理员）',
  `target_user_id` char(32) NOT NULL COMMENT '被操作用户 ID',
  `action_type` varchar(30) NOT NULL COMMENT 'approve/ban/mute/remove/set_admin',
  `reason` varchar(500) DEFAULT NULL COMMENT '操作原因',
  `created_at` bigint(20) NOT NULL COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_group_operator` (`group_id`, `operator_id`, `created_at`),
  KEY `idx_target` (`target_user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='圈主管理操作日志表（1级审计数据）';
```

#### 6.2.9 统计快照表（可选，2级）

以下三张表用于存储 2级统计快照，支持从 1级数据重建：

```sql
-- member_stats：成员统计快照（粉丝数/关注数/发帖数）
CREATE TABLE `member_stats` (
  `user_id` char(32) NOT NULL PRIMARY KEY,
  `followers_count` int(11) NOT NULL DEFAULT '0',
  `followings_count` int(11) NOT NULL DEFAULT '0',
  `topics_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成员统计快照表（2级，可重建）';

-- group_stats：群组统计快照
CREATE TABLE `group_stats` (
  `group_id` char(32) NOT NULL PRIMARY KEY,
  `members_count` int(11) NOT NULL DEFAULT '0',
  `active_members_count` int(11) NOT NULL DEFAULT '0',
  `paid_members_count` int(11) NOT NULL DEFAULT '0',
  `topics_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组统计快照表（2级，可重建）';

-- topic_stats：主题统计快照
CREATE TABLE `topic_stats` (
  `topic_id` char(32) NOT NULL PRIMARY KEY,
  `read_count` int(11) NOT NULL DEFAULT '0',
  `comments_count` int(11) NOT NULL DEFAULT '0',
  `likes_count` int(11) NOT NULL DEFAULT '0',
  `favorites_count` int(11) NOT NULL DEFAULT '0',
  `shares_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题统计快照表（2级，可重建）';
```

### 6.3 ER 关系总览

```
users ──1:N──┬── user_follows（关注关系）
           ├── group_members（群成员关系）
           ├── topic_reactions（互动关系）
           ├── topic_read_records（阅读记录）
           ├── topic_comments（评论）
           └── group_orders（订单）

groups ──1:N──┬── group_members（群成员）
             ├── group_plans（付费方案）
             ├── group_orders（订单）
             ├── topics（帖子）
             └── group_admin_actions（管理日志）

topics ──1:N──┬── topic_comments（评论）
             ├── topic_reactions（互动）
             └── topic_read_records（阅读记录）

group_plans ──1:N── group_orders（订单）
```

***

## 7. 数据等级规范

### 7.1 数据等级定义

详见 [ADR-social-data-level-and-cache-strategy](../adr/ADR-social-data-level-and-cache-strategy.md)。

本节列出社交域数据等级的完整分类清单。

### 7.2 1级数据清单

| 数据表                   | 字段/记录类型                               | 一致性要求    | 更新方式                                                             |
| --------------------- | ------------------------------------- | -------- | ---------------------------------------------------------------- |
| users                 | 全部字段                                  | 强一致，事务写入 | MemberRegister/UpdateMemberProfile/UpdateMemberStatus            |
| groups                | 除 members\_count/topics\_count 外的全部字段 | 强一致，事务写入 | CreateGroup/UpdateGroup/AuditGroup                               |
| topics                | 全部字段                                  | 强一致，事务写入 | CreateTopic/UpdateTopic/DeleteTopic/AuditTopic                   |
| user\_follows         | 关注关系记录                                | 强一致，事务写入 | FollowMember/UnfollowMember                                      |
| group\_members        | 成员关系记录                                | 强一致，事务写入 | JoinGroup/LeaveGroup/UpdateGroupMemberStatus/ConfirmGroupPayment |
| group\_plans          | 付费方案记录                                | 强一致，事务写入 | CreateGroupPlan/UpdateGroupPlan                                  |
| group\_orders         | 订单记录                                  | 强一致，事务写入 | CreateGroupOrder/ConfirmGroupPayment                             |
| topic\_comments       | 评论记录                                  | 强一致，事务写入 | CreateComment/DeleteComment/AdminRemoveComment                   |
| topic\_reactions      | 互动关系记录                                | 强一致，事务写入 | ReactTopic/CancelReactTopic                                      |
| group\_admin\_actions | 审计日志                                  | 强一致，事务写入 | 所有高权限操作自动追加                                                      |

**1级数据铁律**：

- 必须进入 MySQL/InnoDB 主库，使用事务保证一致性
- 不允许只存在于 Redis 或搜索引擎中
- 更新后必须发布领域事件
- 必须有审计追踪能力

### 7.3 2级数据清单

| 数据         | 来源 1级表                                  | 存储位置                                         | 允许延迟 | 重建方式                                                                                |
| ---------- | --------------------------------------- | -------------------------------------------- | ---- | ----------------------------------------------------------------------------------- |
| 用户粉丝数      | user\_follows                           | Redis + member\_stats                        | 秒级   | COUNT(user\_follows WHERE following\_id=? AND status=active)                        |
| 用户关注数      | user\_follows                           | Redis + member\_stats                        | 秒级   | COUNT(user\_follows WHERE follower\_id=? AND status=active)                         |
| 用户发帖数      | topics                                  | Redis + member\_stats                        | 秒级   | COUNT(topics WHERE author\_id=? AND status=published)                               |
| 群组成员数      | group\_members                          | Redis + group\_stats + groups.members\_count | 秒级   | COUNT(group\_members WHERE group\_id=? AND status=active)                           |
| 群组帖子数      | topics                                  | Redis + group\_stats + groups.topics\_count  | 秒级   | COUNT(topics WHERE group\_id=? AND status=published)                                |
| 帖子阅读数      | topic\_read\_records                    | Redis + topic\_stats                         | 分钟级  | COUNT(DISTINCT user\_id FROM topic\_read\_records) 或 HyperLogLog                    |
| 帖子点赞数      | topic\_reactions                        | Redis + topic\_stats                         | 秒级   | COUNT(topic\_reactions WHERE topic\_id=? AND reaction\_type=like AND status=active) |
| 帖子评论数      | topic\_comments                         | Redis + topic\_stats                         | 秒级   | COUNT(topic\_comments WHERE topic\_id=? AND status=normal)                          |
| 圈主看板       | 多表聚合                                    | Redis Snapshot                               | 分钟级  | 多表 JOIN 聚合查询                                                                        |
| 粉丝列表（分页）   | user\_follows                           | Redis List/Set                               | 秒级   | SELECT ... FROM user\_follows ORDER BY created\_at                                  |
| 群组帖子列表（分页） | topics                                  | Redis Sorted Set                             | 秒级   | SELECT ... FROM topics WHERE group\_id=? ORDER BY published\_at                     |
| 热门帖子       | topic\_reactions + topic\_read\_records | Redis Sorted Set                             | 分钟级  | 加权评分计算                                                                              |
| 推荐群组       | groups + group\_stats                   | Redis/Search                                 | 小时级  | 活跃度评分算法                                                                             |

**2级数据铁律**：

- 允许最终一致，不要求实时精确
- 必须能从 1级数据完整重建
- **绝对不能作为权限判断依据**
- 丢失后可通过重建任务恢复

### 7.4 数据等级转换规则

| 规则     | 说明                               |
| ------ | -------------------------------- |
| 1→2 禁止 | 不允许将 1级事实数据降级为 2级派生数据            |
| 2→1 禁止 | 不允许将 2级统计数据提升为 1级事实数据（除非重新定义为事实） |
| 1驱动2   | 1级数据变更必须发布事件，驱动对应 2级数据更新         |
| 2不影响1  | 2级数据的计算/缓存/展示错误不得影响 1级数据的正确性     |

***

## 8. 协议组规划

### 8.1 协议组总览

社交域划分为三个协议组，每个协议组对应一个 maxType：

| 协议组   | maxType  | 职责                      | Proto 文件                  | Tars Server  | Tars Servant |
| ----- | -------- | ----------------------- | ------------------------- | ------------ | ------------ |
| 成员协议组 | **1000** | 用户注册、登录、资料、关注、粉丝、成员状态   | proto/social/member.proto | MemberServer | MemberObj    |
| 群组协议组 | **2000** | 群组 CRUD、入群、付费方案、订单、圈主管理 | proto/social/group.proto  | GroupServer  | GroupObj     |
| 主题协议组 | **3000** | 帖子 CRUD、阅读记录、评论、互动、内容审核 | proto/social/topic.proto  | TopicServer  | TopicObj     |

### 8.2 成员协议组（maxType = 1000）

| minType | 协议名称                | 方向  | 数据等级  | 说明                |
| ------- | ------------------- | --- | ----- | ----------------- |
| 1001    | MemberRegister      | C→S | 1级写入  | 用户注册              |
| 1002    | MemberLogin         | C→S | 1级读取  | 用户登录              |
| 1003    | MemberLogout        | C→S | 1级写入  | 用户登出              |
| 1004    | GetMemberProfile    | C→S | 1级读取  | 获取用户资料            |
| 1005    | UpdateMemberProfile | C→S | 1级写入  | 修改用户资料            |
| 1010    | FollowMember        | C→S | 1级写入  | 关注用户/圈主           |
| 1011    | UnfollowMember      | C→S | 1级写入  | 取消关注              |
| 1012    | ListFollowers       | C→S | 2级读取  | 粉丝列表（可走缓存）        |
| 1013    | ListFollowings      | C→S | 2级读取  | 关注列表（可走缓存）        |
| 1020    | GetMemberStats      | C→S | 2级读取  | 用户统计（粉丝数/关注数/发帖数） |
| 1030    | UpdateMemberStatus  | C→S | 1级高权限 | 平台管理员禁用/恢复用户      |

### 8.3 群组协议组（maxType = 2000）

| minType | 协议名称                    | 方向  | 数据等级    | 说明             |
| ------- | ----------------------- | --- | ------- | -------------- |
| 2001    | CreateGroup             | C→S | 1级写入    | 创建群组           |
| 2002    | UpdateGroup             | C→S | 1级写入    | 修改群组资料         |
| 2003    | GetGroupDetail          | C→S | 1级+2级读取 | 群组详情（含统计信息）    |
| 2004    | ListGroups              | C→S | 2级读取    | 群组列表/推荐列表      |
| 2010    | JoinGroup               | C→S | 1级写入    | 加入群组（免费/申请/邀请） |
| 2011    | LeaveGroup              | C→S | 1级写入    | 退出群组           |
| 2012    | ListGroupMembers        | C→S | 2级读取    | 群成员列表          |
| 2013    | UpdateGroupMemberStatus | C→S | 1级高权限   | 禁言/移除/恢复成员     |
| 2020    | CreateGroupPlan         | C→S | 1级写入    | 创建付费方案         |
| 2021    | UpdateGroupPlan         | C→S | 1级写入    | 修改付费方案         |
| 2022    | ListGroupPlans          | C→S | 1级读取    | 查询群组付费方案       |
| 2030    | CreateGroupOrder        | C→S | 1级写入    | 创建付费订单         |
| 2031    | ConfirmGroupPayment     | C→S | 1级高权限   | 支付确认/手动开通权益    |
| 2040    | GetOwnerDashboard       | C→S | 2级读取    | 圈主看板数据         |
| 2050    | AuditGroup              | C→S | 1级高权限   | 平台审核群组         |

### 8.4 主题协议组（maxType = 3000）

| minType | 协议名称             | 方向  | 数据等级    | 说明              |
| ------- | ---------------- | --- | ------- | --------------- |
| 3001    | CreateTopic      | C→S | 1级写入    | 发布帖子            |
| 3002    | UpdateTopic      | C→S | 1级写入    | 修改帖子            |
| 3003    | DeleteTopic      | C→S | 1级写入    | 删除/下架帖子         |
| 3004    | GetTopicDetail   | C→S | 1级+2级读取 | 帖子详情（含权限判断）     |
| 3005    | ListGroupTopics  | C→S | 2级读取    | 群组帖子列表          |
| 3010    | MarkTopicRead    | C→S | 2级写入    | 标记阅读（可异步）       |
| 3011    | GetReadHistory   | C→S | 2级读取    | 阅读历史            |
| 3020    | CreateComment    | C→S | 1级写入    | 发表评论            |
| 3021    | DeleteComment    | C→S | 1级写入    | 删除评论            |
| 3030    | ReactTopic       | C→S | 1级或2级   | 点赞/收藏/分享        |
| 3031    | CancelReactTopic | C→S | 1级或2级   | 取消点赞/收藏         |
| 3040    | AuditTopic       | C→S | 1级高权限   | 审核/下架帖子         |
| 3050    | GetTopicStats    | C→S | 2级读取    | 帖子统计（阅读/评论/点赞数） |

### 8.5 编号冲突说明

当前 [协议编号注册表](../api/协议编号注册表.md) 的"编号分配建议"中，1000-1999 为"通用基础协议"、2000-2999 为"系统/健康检查/网关"、3000-3999 为"认证与权限"。但上述三个范围目前**无任何已占用编号**（仅有 2100 段被 health/hello 占用，属于 2000-2999 范围内的子集）。

**决策**：社交域三个协议组正式占用 1000/2000/3000 三个 maxType，后续新增通用基础协议从 1100 开始、系统协议从 2100 已占用的下一个可用号开始、认证协议从 3100 开始。本次 PRD 将同步更新协议编号注册表。

***

## 9. 功能需求

### 9.1 成员协议组功能需求

#### FR-Member-001：用户注册

**前置条件**：用户名、邮箱/手机号、密码未被注册

**输入**：username, password, email, phone, nickname

**处理流程**：

1. 校验 username 格式（3-50 字符，字母数字下划线）
2. 校验 email 格式和唯一性
3. 校验 phone 格式和唯一性（如果提供）
4. 密码加密存储（salt + hash）
5. 生成 char(32) UUID 作为 id
6. 生成 9 位 uid
7. 写入 users 表（status=active, membership\_level=normal）
8. 返回 user\_id, uid, nickname, token（JWT）

**输出**：MemberRegisterResponse { user\_id, uid, nickname, token }

**异常**：USERNAME\_EXISTS, EMAIL\_EXISTS, PHONE\_EXISTS, INVALID\_FORMAT

**领域事件**：MemberRegistered

***

#### FR-Member-002：用户登录

**输入**：login\_key（username/email/phone）, password

**处理流程**：

1. 根据 login\_key 类型匹配查询字段
2. 校验密码 hash
3. 检查用户状态（active 才允许登录）
4. 更新 last\_login\_at, last\_login\_ip, login\_count
5. 生成 JWT token
6. 返回用户基本信息

**输出**：MemberLoginResponse { user\_id, uid, nickname, avatar, token, membership\_level }

**异常**：INVALID\_CREDENTIALS, USER\_BANNED, USER\_DELETED

***

#### FR-Member-003：关注用户

**前置条件**：已登录，follower\_id ≠ following\_id

**输入**：target\_user\_id（被关注者）

**处理流程**：

1. CanFollowTarget 权限检查（目标用户未被屏蔽）
2. 查询是否已关注
3. 未关注：INSERT user\_follows（status=active）
4. 已关注且 cancelled：UPDATE status=active
5. 发布 MemberFollowed 事件
6. 失效 follow:rel 缓存
7. 异步更新被关注者粉丝数（2级）

**输出**：FollowMemberResponse { success, followed\_at }

**异常**：ALREADY\_FOLLOWED, CANNOT\_FOLLOW\_SELF, TARGET\_USER\_BANNED

**领域事件**：MemberFollowed

***

#### FR-Member-010：获取用户统计

**输入**：user\_id（可选，默认当前登录用户）

**处理流程**：

1. 优先从 Redis member:stats:{userId} 读取
2. 缓存未命中时查 member\_stats 表
3. member\_stats 也未命中时触发异步重建
4. 返回 followers\_count, followings\_count, topics\_count

**输出**：GetMemberStatsResponse { followers\_count, followings\_count, topics\_count }

**缓存策略**：Cache Aside, TTL 30-120min, 关注/取关事件驱动更新

***

### 9.2 群组协议组功能需求

#### FR-Group-001：创建群组

**前置条件**：已登录，用户未被禁言

**输入**：name, slug, description, avatar, cover\_image, category, tags\[], type, visibility, join\_mode, max\_members, rules, welcome\_message

**处理流程**：

1. 校验 name（1-100字符）、slug 格式和唯一性
2. 校验 type/visibility/join\_mode 枚举值合法
3. 如果 type=paid 但无默认 plan，创建默认免费 plan
4. 生成 char(32) UUID 作为 id
5. owner\_id = 当前用户 ID
6. 写入 groups 表
7. 自动将创建者加入 group\_members（role=owner, status=active）
8. 发布 GroupCreated 事件

**输出**：CreateGroupResponse { group(GroupDTO) }

**领域事件**：GroupCreated

***

#### FR-Group-010：加入群组

**前置条件**：已登录，群组 status=active

**输入**：group\_id, plan\_id（付费群必填）, invite\_code, message（申请理由）

**处理流程**：

1. CanJoinGroup 权限检查
2. 查询群组 join\_mode：
   - 直接加入 → 直接写 group\_members
   - 申请审核 → 写 group\_members(status=pending)，通知圈主
   - 付费加入 → 检查是否有有效 plan\_id，跳转支付流程
   - 邀请加入 → 校验 invite\_code
3. 检查 max\_members 限制
4. 检查是否已在群组中
5. 写入 group\_members
6. 发布 GroupJoined 事件
7. 异步更新群成员数（2级）

**输出**：JoinGroupResponse { member\_id, status, expired\_at }

**异常**：GROUP\_FULL, ALREADY\_JOINED, GROUP\_CLOSED, PLAN\_REQUIRED, INVALID\_INVITE\_CODE

**领域事件**：GroupJoined

***

#### FR-Group-013：更新群组成员状态（高权限）

**前置条件**：操作人具有 CanManageMember 权限

**输入**：group\_id, target\_user\_id, action(ban/mute/remove/recover), reason, duration（mute 时）

**处理流程**（高权限操作四件套）：

1. **权限校验**：CanManageMember(operator, group, target)
   - 操作人必须是 group\_owner 或 group\_admin
   - 不能对 owner 执行 ban/mute/remove
2. **审计日志**：写入 group\_admin\_actions
3. **1级数据更新**：更新 group\_members 对应字段
   - ban → status = banned
   - mute → muted\_until = now + duration
   - remove → status = left
   - recover → status = active, 清除 muted\_until
4. **领域事件**：发布 GroupMemberBanned / GroupMemberMuted / GroupMemberRemoved / GroupMemberRecovered
5. **缓存失效**：
   - DEL group:member:{groupId}:{userId}
   - DEL group:members:{groupId}:\*
   - 更新/失效 group:stats:{groupId}
6. **通知**：站内通知被操作用户

**输出**：UpdateGroupMemberStatusResponse { success, new\_status }

**异常**：NO\_PERMISSION, CANNOT\_OPERATE\_OWNER, MEMBER\_NOT\_FOUND

***

#### FR-Group-031：确认支付开通权益（高权限）

**前置条件**：订单存在且 status=pending，操作人有 ConfirmGroupPayment 权限

**输入**：order\_no

**处理流程**（事务内原子操作）：

1. 查询 order\_no 对应的订单
2. 校验订单状态 = pending
3. **BEGIN TRANSACTION**
4. UPDATE group\_orders SET status=paid, paid\_at=now()
5. 根据 plan\_id 计算 expired\_at = now + plan.duration\_days
6. UPSERT group\_members：
   - 已存在 → 更新 status=active, expired\_at=new\_expired\_at
   - 不存在 → INSERT status=active, expired\_at, role=member, join\_source=paid
7. **COMMIT**
8. 发布 GroupOrderPaid / GroupMemberActivated 事件
9. 失效 group:member 缓存
10. 更新 group:stats 中的 paid\_members\_count
11. 通知用户权益已开通

**输出**：ConfirmGroupPaymentResponse { success, expired\_at, member\_status }

**异常**：ORDER\_NOT\_FOUND, ORDER\_ALREADY\_PAID, PLAN\_NOT\_FOUND, GROUP\_CLOSED

**领域事件**：GroupOrderPaid, GroupMemberActivated

***

### 9.3 主题协议组功能需求

#### FR-Topic-001：发布帖子

**前置条件**：已登录，用户是群组成员且未被禁言，CanPublishTopic 通过

**输入**：group\_id, title, content, summary, tags\[], type, visibility, cover\_image

**处理流程**：

1. CanPublishTopic 权限检查：
   - 用户是 group\_members 且 status=active
   - 用户 muted\_until < now()
   - 群组 status=active
2. 校验 title（1-200字符）、content 非空
3. 根据 group.type 和 post.visibility 校验权限：
   - 付费帖子只能在 paid/mixed 群组发布
4. 生成 char(32) UUID
5. 写入 topics 表（status=待审核 or 已发布，取决于群组设置）
6. 发布 TopicCreated 事件
7. 异步更新群组 topics\_count（2级）

**输出**：CreateTopicResponse { topic(TopicDTO) }

**异常**：NOT\_GROUP\_MEMBER, USER\_MUTED, GROUP\_CLOSED, CONTENT\_EMPTY, VISIBILITY\_FORBIDDEN

**领域事件**：TopicCreated

***

#### FR-Topic-004：获取帖子详情（含权限判断）

**输入**：topic\_id

**处理流程**：

1. 查询 topics 表获取帖子基本信息（1级数据）
2. 检查帖子状态：已删除/已下架 → 返回摘要 + 提示
3. **CanReadTopic 权限判断**（核心安全逻辑）：
   ```
   switch (topic.visibility):
     PUBLIC:
       → 可读（无需群组成员身份）
     GROUP_MEMBER:
       → 查 group_members(group_id, user_id).status == active
     PAID_MEMBER:
       → 查 group_members(group_id, user_id).status == active
       AND group_members.expired_at > now()
     OWNER_ONLY:
       → 查 group_members(group_id, user_id).role IN (owner, admin)
       OR 用户是平台管理员
   ```
4. 有权限：返回完整 content + 附件列表
5. 无权限：返回 summary + visibility + need\_plan\_id（如果是付费帖）
6. **异步**：MarkTopicRead（2级写入，不阻塞响应）

**输出**：GetTopicDetailResponse {
topic(TopicDTO),
can\_read\_full\_content: bool,
access\_status: enum,  // ALLOWED / NEED\_JOIN\_GROUP / NEED\_PAY / FORBIDDEN
need\_plan\_id: string|null  // 付费帖返回推荐的 plan\_id
}

**异常**：TOPIC\_NOT\_FOUND, TOPIC\_DELETED

**缓存策略**：Cache Aside, TTL 5-15min, UpdateTopic/DeleteTopic/AuditTopic 后主动失效

***

#### FR-Topic-020：发表评论

**前置条件**：已登录，有帖子阅读权限，未被禁言

**输入**：topic\_id, content, parent\_id（可选，楼中楼）

**处理流程**：

1. CanReadTopic 快速检查（确保用户至少能看到帖子）
2. CanCommentTopic 检查：
   - 帖子 allow\_comments=true
   - 用户未被全局禁言
   - 用户在该群组未被禁言（muted\_until < now()）
3. 校验 content 非空且 ≤ 5000 字符
4. 写入 topic\_comments（status=正常 或 待审核）
5. 发布 TopicCommentCreated 事件
6. 异步更新帖子 comments\_count（2级）

**输出**：CreateCommentResponse { comment\_id, created\_at }

**异常**：TOPIC\_NOT\_FOUND, NO\_READ\_PERMISSION, COMMENTS\_DISABLED, USER\_MUTED, CONTENT\_TOO\_LONG

**领域事件**：TopicCommentCreated

***

#### FR-Topic-040：审核帖子（高权限）

**前置条件**：操作人具有 CanAuditContent 权限

**输入**：topic\_id, action(approve/reject/ban), reason

**处理流程**（高权限操作四件套）：

1. **权限校验**：CanAuditContent(operator)
2. **审计日志**：记录到 topic\_audit\_logs 或 group\_admin\_actions
3. **1级数据更新**：更新 topics.status
   - approve → status = published
   - reject → status = rejected
   - ban → status = banned
4. **领域事件**：TopicApproved / TopicRejected / TopicBanned
5. **缓存失效**：DEL topic:detail:{topicId}
6. **通知**：通知帖子作者

**输出**：AuditTopicResponse { success, new\_status }

**异常**：NO\_AUDIT\_PERMISSION, TOPIC\_NOT\_FOUND, DUPLICATE\_ACTION

***

## 10. 权限规则

### 10.1 统一权限服务

社交域实现统一的 PermissionService，位于 `go/modules/social/permission/service.go`。

**铁律：不允许各 handler 自行散写权限判断逻辑。**

### 10.2 权限能力清单

| 能力方法            | 签名                                                 | 说明         |
| --------------- | -------------------------------------------------- | ---------- |
| CanViewGroup    | `(userID, groupID string) error`                   | 能否查看群组详情   |
| CanJoinGroup    | `(userID, groupID string) error`                   | 能否加入群组     |
| CanReadTopic    | `(userID, topicID string) (bool, error)`           | 能否阅读帖子完整内容 |
| CanManageGroup  | `(operatorID, groupID string) error`               | 能否管理群组设置   |
| CanManageMember | `(operatorID, groupID, targetUserID string) error` | 能否管理指定成员   |
| CanPublishTopic | `(userID, groupID string) error`                   | 能否在群组发帖    |
| CanAuditContent | `(operatorID string) error`                        | 能否执行内容审核   |

### 10.3 权限判断规则详述

#### CanViewGroup(userID, groupID)

```
IF groups.visibility == public AND groups.status == active:
    → ALLOW
ELIF 用户是 group_members(groupID, userID).status IN (active, admin, owner):
    → ALLOW
ELIF 用户是平台管理员:
    → ALLOW
ELSE:
    → DENY (GROUP_PRIVATE 或 GROUP_NOT_JOINED)
```

**数据来源**：groups 表（1级）+ group\_members 表（1级）。禁止使用 group\_stats 缓存。

#### CanReadTopic(userID, topicID)

```
// 先加载 topic 基本信息
topic := repository.GetTopic(topicID)  // 1级数据

SWITCH topic.visibility:
    CASE PUBLIC:
        IF topic.status == published:
            RETURN true, nil
        ELSE:
            RETURN false, TOPIC_UNAVAILABLE

    CASE GROUP_MEMBER:
        member := repository.GetGroupMember(topic.groupID, userID)  // 1级数据
        IF member.status == active:
            RETURN true, nil
        ELSE:
            RETURN false, NEED_JOIN_GROUP

    CASE PAID_MEMBER:
        member := repository.GetGroupMember(topic.groupID, userID)  // 1级数据
        IF member.status == active AND member.expired_at > now():
            RETURN true, nil
        ELIF member.status == active AND member.expired_at <= now():
            RETURN false, NEED_RENEW  // 需要续费
        ELSE:
            RETURN false, NEED_JOIN_AND_PAY  // 需要加入并付费

    CASE OWNER_ONLY:
        member := repository.GetGroupMember(topic.groupID, userID)
        IF member.role IN (owner, admin):
            RETURN true, nil
        ELIF isPlatformAdmin(userID):
            RETURN true, nil
        ELSE:
            RETURN false, FORBIDDEN
```

**铁律**：CanReadTopic 的判断只查 group\_members 1级关系表，禁止使用以下 2级数据：

- ❌ group\_stats.paid\_member\_count
- ❌ groups.members\_count
- ❌ owner\_dashboard 缓存
- ❌ 任何 Redis 统计缓存

#### CanManageMember(operatorID, groupID, targetUserID)

```
operator := repository.GetGroupMember(groupID, operatorID)  // 1级数据
target := repository.GetGroupMember(groupID, targetUserID)    // 1级数据

// 操作人必须是 owner 或 admin
IF operator.role NOT IN (owner, admin):
    RETURN NO_PERMISSION

// 不能对自己操作（除非只是查看）
IF operator.user_id == target.user_id:
    RETURN CANNOT_OPERATE_SELF

// 不能操作 owner（除非操作人也是 owner 且是更高层级管理员）
IF target.role == owner:
    RETURN CANNOT_OPERATE_OWNER

RETURN nil  // 允许
```

#### CanPublishTopic(userID, groupID)

```
member := repository.GetGroupMember(groupID, userID)  // 1级数据
group := repository.GetGroup(groupID)                  // 1级数据

// 用户必须是活跃群成员
IF member.status != active:
    RETURN NOT_GROUP_MEMBER, nil

// 用户未被禁言
IF member.muted_until != NULL AND member.muted_until > now():
    RETURN USER_MUTED, nil

// 群组必须开放
IF group.status != active:
    RETURN GROUP_CLOSED, nil

RETURN nil  // 允许
```

### 10.4 权限与数据等级的安全约束

| 约束              | 说明                                    |
| --------------- | ------------------------------------- |
| 权限只能基于 1级数据     | CanXxx 方法内部只查 MySQL 主表或极短 TTL 的关系缓存   |
| 2级数据仅用于展示       | 统计数、排序列表、推荐分数等 2级数据只能用于前端展示和排序        |
| 高权限操作二次校验       | 即使前端隐藏了按钮，后端也必须做完整权限校验                |
| 缓存失效优先于容忍 stale | 安全相关的缓存（如成员关系）宁可多删也不能让 stale 数据影响权限判断 |

***

## 11. 缓存策略

### 11.1 总体原则

详见 [ADR-social-data-level-and-cache-strategy](../adr/ADR-social-data-level-and-cache-strategy.md) 第四章和第五章。

### 11.2 1级数据缓存策略（Cache Aside + 主动失效）

| 数据类型 | 缓存 Key                                  | TTL      | 失效时机                                        | 失效方式      |
| ---- | --------------------------------------- | -------- | ------------------------------------------- | --------- |
| 用户资料 | `member:profile:{userId}`               | 10-30min | UpdateMemberProfile                         | 主动 DELETE |
| 群组详情 | `group:detail:{groupId}`                | 5-15min  | UpdateGroup / AuditGroup                    | 主动 DELETE |
| 付费方案 | `group:plans:{groupId}`                 | 10-30min | CreatePlan / UpdatePlan                     | 主动 DELETE |
| 帖子详情 | `topic:detail:{topicId}`                | 5-15min  | UpdateTopic / DeleteTopic / AuditTopic      | 主动 DELETE |
| 成员关系 | `group:member:{groupId}:{userId}`       | 5-10min  | JoinGroup / LeaveGroup / UpdateMemberStatus | 主动 DELETE |
| 关注关系 | `follow:rel:{followerId}:{followingId}` | 5-10min  | FollowMember / UnfollowMember               | 主动 DELETE |

**1级数据缓存铁律**：

- TTL 不超过 30 分钟
- 不依赖 TTL 自然过期，必须主动失效
- 安全敏感场景（如成员关系、禁言状态）TTL 更短（5-10min）
- 失效后下次访问触发 Cache Miss → 回源 DB → 重建缓存

### 11.3 2级数据缓存策略（事件驱动 + TTL 兜底 + 定时校准）

| 数据类型   | 缓存 Key                             | TTL       | 更新方式            | 重建方式        |
| ------ | ---------------------------------- | --------- | --------------- | ----------- |
| 用户统计   | `member:stats:{userId}`            | 30-120min | 关注/发帖事件增量       | 全量 COUNT 重建 |
| 群组统计   | `group:stats:{groupId}`            | 10-60min  | 入群/退群/发帖事件增量    | 全量 COUNT 重建 |
| 帖子统计   | `topic:stats:{topicId}`            | 5-30min   | 阅读/评论/点赞事件增量    | 全量 COUNT 重建 |
| 粉丝列表   | `member:followers:{userId}:{page}` | 1-10min   | 关注/取关后删除前几页     | DB 分页查询重建   |
| 群组帖子列表 | `group:topics:{groupId}:{page}`    | 1-5min    | 发帖/删帖后删除第一页     | DB 分页查询重建   |
| 圈主看板   | `owner:dashboard:{ownerId}`        | 1-10min   | 事件刷新 + 定时重建     | 多表 JOIN 重建  |
| 热门帖子   | `topic:hot:{groupId}`              | 1-5min    | Sorted Set 增量计算 | 加权评分重算      |
| 推荐群组   | `group:recommend:{category}`       | 5-30min   | 定时任务重建          | 活跃度算法重建     |

### 11.4 缓存 Key 命名规范总表

```
# ===== 成员域 =====
member:profile:{userId}              # 用户资料
member:stats:{userId}                # 用户统计
member:followers:{userId}:{page}     # 粉丝列表（分页）
member:followings:{userId}:{page}     # 关注列表（分页）
follow:rel:{followerId}:{followingId} # 关注关系是否存在

# ===== 群组域 =====
group:detail:{groupId}               # 群组详情
group:member:{groupId}:{userId}      # 成员关系
group:members:{groupId}:{page}       # 成员列表（分页）
group:plans:{groupId}                # 付费方案列表
group:stats:{groupId}                # 群组统计
owner:dashboard:{ownerId}             # 圈主看板
group:recommend:{category}           # 推荐群组

# ===== 主题域 =====
topic:detail:{topicId}               # 帖子详情
topic:stats:{topicId}                # 帖子统计
topic:read:{userId}:{topicId}        # 阅读记录
group:topics:{groupId}:{page}        # 群组帖子列表（分页）
topic:hot:{groupId}                  # 热门帖子排行
```

***

## 12. 高权限操作与通知机制

### 12.1 高权限操作清单

| 操作                      | 协议                 | 权限要求        | 审计 | 事件                 | 缓存失效                                   | 通知     |
| ----------------------- | ------------------ | ----------- | -- | ------------------ | -------------------------------------- | ------ |
| UpdateMemberStatus      | 1030               | 平台管理员       | ✅  | UserStatusChanged  | member:profile:\*                      | 被操作用户  |
| AuditGroup              | 2050               | 平台管理员/审核员   | ✅  | GroupAudited       | group:detail:\*                        | 圈主     |
| UpdateGroupMemberStatus | 2013               | 圈主/管理员      | ✅  | GroupMember\*      | group:member/group:members/group:stats | 被操作用户  |
| ConfirmGroupPayment     | 2031               | 系统操作/管理员    | ✅  | GroupOrderPaid     | group:member/group:stats               | 购买用户   |
| AuditTopic              | 3040               | 平台管理员/审核员   | ✅  | TopicAudited       | topic:detail:\*                        | 帖子作者   |
| DeleteTopic             | 3003               | 作者本人/圈主/管理员 | ✅  | TopicDeleted       | topic:detail/group:topics/topic:hot    | 无      |
| AdminRemoveComment      | (扩展3021)           | 管理员         | ✅  | CommentRemoved     | topic:stats                            | 被删评论用户 |
| OwnerMuteMember         | 2013+action=mute   | 圈主/管理员      | ✅  | GroupMemberMuted   | group:member                           | 被禁言用户  |
| OwnerRemoveMember       | 2013+action=remove | 圈主/管理员      | ✅  | GroupMemberRemoved | group:member/group:members/group:stats | 被移除用户  |

### 12.2 高权限操作标准处理流水线

每个高权限操作必须按以下顺序执行，不可省略：

```
Step 1: 权限校验（PermissionService.CanXxx）
   ↓ 通过
Step 2: 写入审计日志（group_admin_actions 或系统审计表）
   ↓
Step 3: 更新 1级数据（MySQL 事务内）
   ↓ 事务提交成功
Step 4: 发布领域事件（EventPublisher）
   ↓
Step 5: 主动失效相关缓存（CacheInvalidator）
   ↓
Step 6: 通知受影响用户（NotificationService / 站内信）
   ↓
Step 7: 返回成功响应
```

### 12.3 领域事件清单

| 事件名称                | 触发来源                    | 消费者（2级数据更新）                    |
| ------------------- | ----------------------- | ------------------------------ |
| MemberRegistered    | MemberRegister          | 初始化 member\_stats              |
| MemberFollowed      | FollowMember            | 更新粉丝数(following)、关注数(follower) |
| MemberUnfollowed    | UnfollowMember          | 更新粉丝数、关注数                      |
| GroupCreated        | CreateGroup             | 初始化 group\_stats               |
| GroupJoined         | JoinGroup               | 更新成员数                          |
| GroupLeft           | LeaveGroup              | 更新成员数                          |
| GroupMemberRemoved  | UpdateGroupMemberStatus | 更新成员数、清理缓存                     |
| GroupMemberBanned   | UpdateGroupMemberStatus | 更新成员数、清理缓存                     |
| GroupPlanCreated    | CreateGroupPlan         | 刷新群组方案缓存                       |
| GroupOrderPaid      | ConfirmGroupPayment     | 更新付费成员数、开通权益                   |
| TopicCreated        | CreateTopic             | 更新群组帖子数                        |
| TopicDeleted        | DeleteTopic             | 更新群组帖子数、清理评论/互动                |
| TopicCommentCreated | CreateComment           | 更新帖子评论数                        |
| TopicReacted        | ReactTopic              | 更新帖子点赞/收藏数                     |
| TopicAudited        | AuditTopic              | 更新帖子状态缓存                       |

### 12.4 通知分级

| 通知级别             | 用途        | MVP 首期 | 实现方式                  |
| ---------------- | --------- | ------ | --------------------- |
| 领域事件（内部）         | 驱动 2级数据更新 | **必须** | Redis PubSub / 内存事件总线 |
| 站内通知             | 用户可见的系统消息 | **保留** | 写入 notifications 表    |
| WebSocket/IM 推送  | 在线即时提醒    | 后续     | WebSocket 服务          |
| 外部推送（短信/邮件/Mail） | 重要操作触达    | 后续     | 第三方推送服务               |

***

## 13. OpenAPI 协议映射

### 13.1 设计原则

OpenAPI 文档仅描述统一网关入口，不新增任何绕过网关的业务 REST 接口。

完整 OpenAPI 定义见 \[social-openapi.yaml]\(../api/social-openapi.yaml）。

### 13.2 网关入口定义

```yaml
paths:
  /api/hello:
    post:
      summary: CaiRobot Social Domain Unified Gateway
      description: >
        社交域所有业务请求统一通过此入口。
        使用 MessagePacket.maxType 区分成员(1000)/群组(2000)/主题(3000)协议组，
        使用 MessagePacket.minType 区分具体业务动作，
        payload 为对应 Protobuf Request 序列化后的 bytes。
      operationId: callSocialMessagePacket
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
```

### 13.3 协议组扩展声明

```yaml
x-cairobot-protocol-groups:
  - maxType: 1000
    name: MemberProtocolGroup
    description: 用户注册、登录、资料、关注粉丝、成员状态协议组
  - maxType: 2000
    name: GroupProtocolGroup
    description: 群组创建、详情、入群、付费方案、订单、圈主管理协议组
  - maxType: 3000
    name: TopicProtocolGroup
    description: 帖子发布、详情、列表、阅读、评论、互动、审核协议组
```

### 13.4 完整协议映射索引

每个协议包含以下元数据：maxType, minType, dataLevel, requestProto, responseProto, cachePolicy, events\[], permission\[]。

完整映射见 [social-openapi.yaml](../api/social-openapi.yaml) 中的 `x-cairobot-protocols` 扩展字段。

***

## 14. 非功能需求

### 14.1 性能要求

| 场景            | P50 延迟  | P99 延迟  | QPS 目标 |
| ------------- | ------- | ------- | ------ |
| 用户登录          | < 100ms | < 300ms | 500    |
| 获取帖子详情（含权限判断） | < 150ms | < 400ms | 1000   |
| 发布帖子          | < 300ms | < 800ms | 200    |
| 关注/取关         | < 150ms | < 400ms | 500    |
| 群组列表（分页）      | < 200ms | < 500ms | 800    |
| 粉丝列表（分页）      | < 200ms | < 500ms | 500    |
| 圈主看板          | < 300ms | < 800ms | 100    |

### 14.2 安全要求

- 所有写入操作必须经过认证（auth\_required: true，除 MemberRegister/MemberLogin）
- 密码必须 salt + hash 存储，不允许明文
- Token 使用 JWT，含 userId + role + exp
- 敏感操作（支付、权限变更）需要二次验证或操作确认
- SQL 注入防护：全部使用参数化查询（repository 层）
- XSS 防护：用户生成内容（帖子正文、评论）在输出时转义

### 14.3 可用性要求

- MySQL 主库可用性 ≥ 99.9%
- Redis 缓存可用性 ≥ 99.5%（缓存降级时直连 DB）
- 核心链路（读帖子、加群）降级策略：缓存 miss 时回源 DB
- 2级数据统计异常时不影响 1级数据读写

### 14.4 数据一致性要求

- 1级数据：强一致（ACID 事务）
- 2级数据：最终一致（容忍 1-5 秒延迟）
- 支付→权益开通：事务内原子操作（最强一致要求）
- 计数器：允许短暂不一致（±1 可接受）

***

## 15. 测试验收标准

### 15.1 正常路径测试用例

| 编号     | 场景          | 输入                                  | 期望输出                                            |
| ------ | ----------- | ----------------------------------- | ----------------------------------------------- |
| TC-001 | 用户正常注册      | 有效 username/email/password/nickname | 返回 user\_id + token，users 表新增记录                 |
| TC-002 | 用户正常登录      | 正确 credentials                      | 返回用户信息 + token                                  |
| TC-003 | 正常关注圈主      | follower ≠ following                | user\_follows 新增 active 记录                      |
| TC-004 | 正常取关        | 已存在 active 关注记录                     | user\_follows status → cancelled                |
| TC-005 | 创建免费群组      | 合法参数 + type=free                    | groups 新增记录 + 自动成为 owner                        |
| TC-006 | 正常加入免费群组    | group(join\_mode=direct)            | group\_members 新增 active 记录                     |
| TC-007 | 创建付费订单      | group + plan                        | group\_orders 新增 pending 记录                     |
| TC-008 | 手动确认支付开通权益  | order\_no                           | orders=paid + members=active + expired\_at 设置正确 |
| TC-009 | 发布帖子        | 有发帖权限的用户                            | topics 新增记录                                     |
| TC-010 | 读取公开帖子      | 任何人                                 | 返回完整内容，can\_read\_full\_content=true            |
| TC-011 | 读取群成员帖子     | 群组成员                                | 返回完整内容                                          |
| TC-012 | 读取付费帖子（有权益） | 付费成员且未过期                            | 返回完整内容                                          |
| TC-013 | 发表评论        | 有权限的用户                              | topic\_comments 新增记录                            |
| TC-014 | 点赞帖子        | 已登录用户                               | topic\_reactions 新增 like 记录                     |
| TC-015 | 圈主查看成员列表    | 圈主本人                                | 返回该群组所有成员                                       |
| TC-016 | 圈主禁言成员      | 圈主操作非 owner 成员                      | member.muted\_until 被设置 + 审计日志产生                |

### 15.2 异常路径测试用例（负面测试）

| 编号     | 场景         | 输入                             | 期望输出                                   |
| ------ | ---------- | ------------------------------ | -------------------------------------- |
| TE-001 | 注册用户名已存在   | 重复 username                    | USERNAME\_EXISTS 错误                    |
| TE-002 | 注册邮箱已存在    | 重复 email                       | EMAIL\_EXISTS 错误                       |
| TE-003 | 登录密码错误     | 错误密码                           | INVALID\_CREDENTIALS 错误                |
| TE-004 | 登录被禁用用户    | banned 状态用户                    | USER\_BANNED 错误                        |
| TE-005 | 自己关注自己     | follower == following          | CANNOT\_FOLLOW\_SELF 错误                |
| TE-006 | 重复关注       | 已 active                       | ALREADY\_FOLLOWED 错误                   |
| TE-007 | 加入已达上限群组   | groups.max\_members=100, 第101人 | GROUP\_FULL 错误                         |
| TE-008 | 重复加入已加入群组  | 已是 active 成员                   | ALREADY\_JOINED 错误                     |
| TE-009 | 无权限读取付费帖子  | 非成员访问 PAID\_MEMBER 帖子          | can\_read=false + NEED\_JOIN\_AND\_PAY |
| TE-010 | 付费权益已过期    | expired\_at < now              | can\_read=false + NEED\_RENEW          |
| TE-011 | 非圈主尝试管理成员  | 普通成员调用 UpdateGroupMemberStatus | NO\_PERMISSION 错误                      |
| TE-012 | 圈主尝试移除其他圈主 | owner 操作 owner                 | CANNOT\_OPERATE\_OWNER 错误              |
| TE-013 | 被禁言用户发帖    | muted\_until > now() 的用户       | USER\_MUTED 错误                         |
| TE-014 | 删除已删除帖子    | status=deleted 的帖子             | TOPIC\_NOT\_FOUND / TOPIC\_DELETED     |
| TE-015 | 在关闭群组发帖    | status=inactive 的群组            | GROUP\_CLOSED 错误                       |
| TE-016 | 并发点赞取消竞态   | 同时 React + CancelReact         | 最终状态确定（幂等保护）                           |

### 15.3 协议集成测试

| 编号     | 场景                | 验证点                                 |
| ------ | ----------------- | ----------------------------------- |
| TI-001 | maxType=1000 路由正确 | routes.yaml 匹配到 MemberServer        |
| TI-002 | maxType=2000 路由正确 | routes.yaml 匹配到 GroupServer         |
| TI-003 | maxType=3000 路由正确 | routes.yaml 匹配到 TopicServer         |
| TI-004 | 未登记协议号拒绝          | 返回 PROTOCOL\_NOT\_REGISTERED 错误     |
| TI-005 | Protobuf 编解码正确    | Request bytes ↔ Response bytes 往返一致 |
| TI-006 | extend 透传正确       | traceId/requestId/userId 全链路传递      |

### 15.4 缓存一致性测试

| 编号         | 场景                | 验证点                        |
| ---------- | ----------------- | -------------------------- |
| TCACHE-001 | 更新用户资料后缓存失效       | GET profile 返回新数据（TTL 内）   |
| TCACHE-002 | 关注后粉丝数最终一致        | 1-3 秒内 followers\_count +1 |
| TCACHE-003 | 禁言后立即生效           | 被禁言用户下一请求即无法发帖             |
| TCACHE-004 | Redis 不可用时降级直连 DB | 功能正常，延迟略有上升                |

***

## 16. 后续迭代范围

### 16.1 第二期（P1-P2）

| 功能         | 说明                 |
| ---------- | ------------------ |
| 消息通知系统     | 站内信 + WebSocket 推送 |
| 内容审核 AI 集成 | 自动审核帖子/评论          |
| 搜索引擎接入     | Elasticsearch 全文检索 |
| 推荐流        | 基于关注关系 + 行为的内容推荐   |
| 真实支付对接     | 微信支付/支付宝/IAP       |
| 退款流程       | 订单退款 → 权益回收        |
| 数据导出       | 圈主导出成员/收益报表        |
| 内容分类标签体系   | 自動打標 + 人工校正        |

### 16.2 第三期（远期）

| 功能       | 说明          |
| -------- | ----------- |
| 即时通讯（IM） | 群聊/私聊       |
| 直播       | 圈主直播        |
| 问答系统     | 付费问答        |
| 知识付费     | 专栏/课程       |
| 分销体系     | 推广佣金        |
| 多租户      | 服务商模式       |
| 国际化      | 多语言 UI + 内容 |

***

## 附录 A：术语表

| 术语            | 定义                               |
| ------------- | -------------------------------- |
| 1级数据          | 业务事实数据，强一致，协议主动写入才变化             |
| 2级数据          | 派生统计数据，最终一致，由事件/任务被动触发变化         |
| 圈主            | 群组的 owner 角色，拥有最高管理权限            |
| 群组            | 也称"圈子"，用户聚集和内容发布的容器              |
| 帖子            | 也称"主题"，群组内的内容单元                  |
| MessagePacket | 统一网关入口报文，包含 maxType/minType/data |
| 领域事件          | 1级数据变更后发布的结构化事件，驱动 2级数据更新        |
| Cache Aside   | 缓存读写策略：读时回源填充，写时主动失效             |
| 高权限操作         | 影响其他用户权益的操作，需额外审计和通知             |

## 附录 B：参考文档

| 文档                     | 路径                                                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| CODE-WIKI              | [docs/wiki/CODE-WIKI.md](../wiki/CODE-WIKI.md)                                                                                 |
| 协议编号注册表                | [docs/api/协议编号注册表.md](../api/协议编号注册表.md)                                                                                       |
| 数据等级 ADR               | [docs/adr/ADR-social-data-level-and-cache-strategy.md](../adr/ADR-social-data-level-and-cache-strategy.md)                     |
| Social OpenAPI         | [docs/api/social-openapi.yaml](../api/social-openapi.yaml)                                                                     |
| 基础数据模型                 | [docs/tabbit/inbox/2026/06/basemodel.md](../tabbit/inbox/2026/06/basemodel.md)                                                 |
| ADR-0003 Protobuf      | [docs/adr/ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)                                                       |
| ADR-0008 TarsCloud 路由层 | [docs/adr/ADR-0008-use-tarscloud-routing-layer.md](../adr/ADR-0008-use-tarscloud-routing-layer.md)                             |
| ADR-0014 MessagePacket | [docs/adr/ADR-0014-message-packet-data-format-protobuf-bytes.md](../adr/ADR-0014-message-packet-data-format-protobuf-bytes.md) |

