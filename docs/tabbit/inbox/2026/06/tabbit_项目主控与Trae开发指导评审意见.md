@PRD-social-app-mvp.md 评审prd文档，输出评审意见，输出拆分后的数据库表设计，补充社交规则：当前需求没有关注需求，是群组关系维护，成员有5种，游客，普通成员，嘉宾，管理员，群组所有人，成员间可以相互拉黑，拉黑后无法获取对应主题，群组所有人可以控制付费周期，折扣周期，成员可以付费加入和续费，也可以免费加入，嘉宾是被邀请加入，主题和群主，嘉宾邀请都可以分享，消息中心可以看到各种消息，私信，群组状态通知，订单变更通知，点赞和评论，请基于redis + go +will消息队列设计

### **评审结论：当前 PRD 架构方向正确，但业务模型需要从“关注型社交”调整### **评审结论：当前 PRD 架构方向基本正确，但业务关系需要从“关注粉丝社交”调整为“群组成员关系社交”，并补齐黑名单、嘉宾邀请、付费周期/折扣周期、分享、消息中心、私信、通知与 Redis + Go + 消息队列事件架构。**

`PRD-social-app-mvp.md` 已经很好地覆盖了 CaiRobot MVP 的平台约束：单网关 `POST /api/hello`、`MessagePacket`、Protobuf、`routes.yaml`、TarsGo、数据等级、缓存策略、协议组 1000/2000/3000 等内容。但当前文档仍然偏向“关注圈主/粉丝关系”的内容型社区模型，而你现在补充的真实需求更接近 **群组关系维护 + 付费社群 + 嘉宾邀请 + 成员黑名单 + 主题阅读控制 + 消息中心**。所以 PRD 需要做一次产品模型纠偏：**删除或降级关注需求，把核心社交关系从 `user_follows` 改为 `group_members`、`member_blocks`、`group_invites`、`topic_shares`、`message_center` 等关系模型。**

---

### **一、PRD 评审意见**

#### **1. 当前 PRD 的优点**

当前 PRD 最大的优点是已经把 CaiRobot MVP 的工程规则固定下来了，尤其是以下部分应该保留：

| 项目 | 评审意见 |
|---|---|
| 单网关约束 | 保留，所有社交域请求继续走 `POST /api/hello` |
| 协议组规划 | 保留，成员 1000、群组 2000、主题 3000 的边界清晰 |
| 数据等级设计 | 保留，1级强一致事实数据、2级最终一致派生数据是合理的 |
| 缓存策略 | 保留，但需要补充消息、黑名单、邀请、订单、分享的缓存策略 |
| 高权限操作四件套 | 保留，权限校验、审计、事件、缓存失效、通知必须坚持 |
| OpenAPI 约束 | 保留，OpenAPI 只描述统一网关，不新增真实 REST 业务接口 |
| 权限服务 | 保留，但需要扩展为群组角色 + 黑名单 + 付费权益 + 嘉宾邀请综合判断 |

这说明 PRD 的“平台架构层”是合格的，但“社交产品层”需要重构。

#### **2. 当前 PRD 的核心问题**

当前 PRD 里最大的问题是把“关注圈主/粉丝关系”作为核心社交关系，但你现在明确说 **当前需求没有关注需求，是群组关系维护**。因此以下内容需要调整：

| 当前 PRD 内容 | 问题 | 建议 |
|---|---|---|
| `user_follows` 作为核心表 | 与当前需求不符 | 从 P0 移除，改为后续扩展 |
| `FollowMember / UnfollowMember` | 当前无关注需求 | 从协议组 1000 删除或标记 P2 |
| `ListFollowers / ListFollowings` | 当前无粉丝关系需求 | 改为 `ListGroupMembers`、`ListBlockedMembers`、`ListGuestInvites` |
| 圈主粉丝管理 | 当前应是群组成员管理 | 改为群组成员、嘉宾、黑名单、付费成员管理 |
| 用户关注圈主流程 | 业务不成立 | 替换为免费入群、付费入群、嘉宾邀请、拉黑流程 |
| 圈主看板偏粉丝统计 | 统计口径不准 | 改为群组成员数、付费成员数、嘉宾数、续费率、主题数、消息互动数 |

也就是说，PRD 的社交关系应从：

```text
用户关注用户 / 用户关注圈主
```

调整为：

```text
用户加入群组 / 用户续费群组 / 嘉宾受邀加入 / 成员之间拉黑 / 群主和管理员管理成员 / 成员围绕主题互动
```

#### **3. 成员角色模型需要重写**

你定义了 5 种成员身份：

```text
游客
普通成员
嘉宾
管理员
群组所有人
```

这应该成为 PRD 的核心权限模型。建议统一建模为：

| 角色 | role 值 | 是否入群 | 核心权限 |
|---|---|---|---|
| 游客 | `guest_viewer` 或无成员记录 | 否 | 浏览公开群组、公开主题摘要、注册登录 |
| 普通成员 | `member` | 是 | 查看成员可见主题、评论、点赞、付费/续费 |
| 嘉宾 | `guest` | 是，被邀请 | 查看被授权内容、分享主题、参与互动，权限由邀请规则限制 |
| 管理员 | `admin` | 是 | 管理成员、审核主题、处理消息与状态通知 |
| 群组所有人 | `owner` | 是 | 群组最高权限，设置付费周期、折扣周期、邀请嘉宾、管理管理员 |

注意：**游客不是 `group_members` 的正式成员角色**。游客通常没有成员记录，权限由 `groups.visibility` 和 `topics.visibility` 判断。否则会导致每个浏览者都写入成员表，数据膨胀严重。

#### **4. 黑名单规则必须补充为安全规则**

你补充“成员间可以相互拉黑，拉黑后无法获取对应主题”。这个规则要非常明确，否则会出现权限漏洞。

建议规则：

```text
A 拉黑 B 后：
1. B 不能查看 A 发布的主题完整内容；
2. B 不能评论 A 的主题；
3. B 不能给 A 发送私信；
4. B 不能通过群组主题列表看到 A 的部分受限主题，或只能看到脱敏摘要；
5. A 也可以选择是否不看 B 的内容，默认双向屏蔽互动；
6. 管理员和群组所有人处理审核/管理时不受普通黑名单限制，但操作需要审计。
```

这里要区分两类黑名单：

| 类型 | 表 | 说明 |
|---|---|---|
| 用户间拉黑 | `member_blocks` | 成员 A 拉黑成员 B |
| 群组级封禁 | `group_members.status = banned` 或 `group_member_restrictions` | 群主/管理员封禁成员 |

两者不能混用。用户拉黑是私人关系；群组封禁是管理行为。

#### **5. 付费周期与折扣周期需要单独建模**

当前 PRD 里只有 `group_plans`，还不够。你补充“群组所有人可以控制付费周期，折扣周期”，这意味着价格体系至少要拆成：

```text
付费方案 group_plans
付费周期 group_plan_periods
折扣活动 group_discounts
订单 group_orders
权益 group_entitlements 或 group_members.expired_at
```

如果只把周期和折扣写进 `group_plans.benefits JSON`，后续会很难做查询、运营、审计和订单回溯。

建议原则：

```text
group_plans 定义商品/权益类型；
group_plan_periods 定义月卡、季卡、年卡、终身卡；
group_discounts 定义限时折扣、首购折扣、续费折扣、邀请码折扣；
group_orders 记录用户实际购买快照；
group_member_entitlements 记录用户当前权益。
```

#### **6. 消息中心需要独立成为基础能力**

你补充消息中心包括：

```text
各种消息
私信
群组状态通知
订单变更通知
点赞和评论
```

这不应该散落在各业务表里，建议新增 **消息中心协议组或作为成员协议组子域**。由于你当前协议组已固定为成员 1000、群组 2000、主题 3000，可以先放在成员协议组 1000 内，后续独立成 4000。

建议 P0 先定义为：

| 消息类型 | 来源 |
|---|---|
| 私信 | 用户对用户 |
| 群组状态通知 | 入群、退群、禁言、移除、嘉宾邀请、管理员变更 |
| 订单变更通知 | 创建订单、支付成功、续费成功、权益过期提醒 |
| 点赞通知 | 主题被点赞 |
| 评论通知 | 主题被评论、评论被回复 |
| 系统通知 | 管理员审核、平台公告 |

消息中心的数据一定是 1级数据，因为用户要能可靠看到历史消息；未读数是 2级数据。

---

### **二、建议调整后的 PRD 结构**

建议把 `PRD-social-app-mvp.md` 重构为下面结构：

```text
1. 背景与目标
2. 当前项目架构约束
3. 社交 App MVP 范围
4. 核心产品模型
   4.1 群组关系模型
   4.2 成员角色模型
   4.3 付费权益模型
   4.4 黑名单模型
   4.5 嘉宾邀请模型
   4.6 消息中心模型
5. 用户角色与使用场景
6. 核心业务流程
   6.1 用户免费加入群组
   6.2 用户付费加入群组
   6.3 用户续费群组
   6.4 群主邀请嘉宾
   6.5 成员拉黑成员
   6.6 用户阅读主题
   6.7 分享主题/邀请
   6.8 消息中心收发与未读统计
7. 数据等级规范
8. 数据库表设计
9. 协议组规划
10. 功能需求
11. 权限规则
12. 缓存策略
13. Redis + Go + 消息队列事件架构
14. OpenAPI 协议映射
15. 非功能需求
16. 测试验收标准
17. 后续迭代范围
```

---

### **三、拆分后的数据库表设计**

下面是基于当前需求重新拆分后的数据库表设计。已有的 `users`、`groups`、`topics` 继续保留，但语义要调整。

---

## **1. 成员域表设计**

### **1.1 `users` 用户主表，已有，保留**

用途：账号身份主表。

建议调整：

```text
users 只保存账号身份、资料、状态；
不保存粉丝数、群组数、未读数、付费权益；
不承载关注关系，因为当前需求没有关注模型。
```

保留核心字段：

| 字段 | 等级 | 说明 |
|---|---|---|
| `id` | 1级 | 内部用户 ID |
| `username` | 1级 | 用户名 |
| `password` | 1级 | 加密密码 |
| `email` | 1级 | 邮箱 |
| `phone` | 1级 | 手机号 |
| `nickname` | 1级 | 昵称 |
| `avatar` | 1级 | 头像 |
| `status` | 1级 | active/inactive/banned/deleted |
| `membership_level` | 1级 | 平台会员等级，不等于群组付费权益 |

---

### **1.2 `member_blocks` 成员拉黑关系表**

用途：成员之间相互拉黑，影响私信、主题可见性、评论互动。

```sql
CREATE TABLE `member_blocks` (
  `id` char(32) NOT NULL COMMENT '主键 UUID',
  `blocker_id` char(32) NOT NULL COMMENT '拉黑人用户 ID',
  `blocked_id` char(32) NOT NULL COMMENT '被拉黑人用户 ID',
  `scope` varchar(20) NOT NULL DEFAULT 'global' COMMENT 'global/group',
  `group_id` char(32) DEFAULT NULL COMMENT 'scope=group 时生效的群组',
  `reason` varchar(255) DEFAULT NULL COMMENT '拉黑原因',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_block_pair_scope` (`blocker_id`, `blocked_id`, `scope`, `group_id`),
  KEY `idx_blocker_status` (`blocker_id`, `status`, `created_at`),
  KEY `idx_blocked_status` (`blocked_id`, `status`, `created_at`),
  KEY `idx_group_block` (`group_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成员拉黑关系表（1级数据）';
```

数据等级：1级。

关键规则：

```text
1. blocker_id 不能等于 blocked_id；
2. active 状态下，被拉黑人不能私信拉黑人；
3. active 状态下，被拉黑人不能获取拉黑人发布的受限主题完整内容；
4. 管理员执行群组管理和审核时可绕过普通拉黑，但必须审计；
5. 拉黑关系变化后必须失效 block:rel 缓存。
```

---

### **1.3 `member_sessions` 用户会话表**

用途：登录态、token 管理、设备管理。

```sql
CREATE TABLE `member_sessions` (
  `id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `token_id` varchar(64) NOT NULL COMMENT 'JWT jti 或会话 ID',
  `device_id` varchar(128) DEFAULT NULL,
  `device_type` varchar(30) DEFAULT NULL COMMENT 'ios/android/web/admin',
  `ip` varchar(45) DEFAULT NULL,
  `user_agent` varchar(500) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=revoked 3=expired',
  `expired_at` bigint(20) NOT NULL,
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token_id` (`token_id`),
  KEY `idx_user_status` (`user_id`, `status`, `expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表（1级数据）';
```

---

## **2. 群组域表设计**

### **2.1 `groups` 群组主表，已有，保留并优化**

用途：群组基础信息。

建议字段语义：

| 字段 | 等级 | 说明 |
|---|---|---|
| `id` | 1级 | 群组 ID |
| `name` | 1级 | 群组名称 |
| `slug` | 1级 | 群组短标识 |
| `owner_id` | 1级 | 群组所有人 |
| `type` | 1级 | free/paid/mixed/invite |
| `visibility` | 1级 | public/link/private |
| `join_mode` | 1级 | free/apply/paid/invite |
| `status` | 1级 | active/inactive/auditing/banned/deleted |
| `members_count` | 2级 | 成员数快照 |
| `topics_count` | 2级 | 主题数快照 |

---

### **2.2 `group_members` 群组成员关系表**

用途：群组成员身份、角色、状态、权益基础。

```sql
CREATE TABLE `group_members` (
  `id` char(32) NOT NULL COMMENT '主键 UUID',
  `group_id` char(32) NOT NULL COMMENT '群组 ID',
  `user_id` char(32) NOT NULL COMMENT '用户 ID',
  `role` varchar(20) NOT NULL DEFAULT 'member' COMMENT 'member/guest/admin/owner',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=pending 3=muted 4=banned 5=left 6=expired',
  `join_source` varchar(20) NOT NULL DEFAULT 'free' COMMENT 'free/paid/invite/admin/import',
  `invited_by` char(32) DEFAULT NULL COMMENT '邀请人',
  `joined_at` bigint(20) DEFAULT NULL COMMENT '加入时间',
  `left_at` bigint(20) DEFAULT NULL COMMENT '离开时间',
  `muted_until` bigint(20) DEFAULT NULL COMMENT '禁言截止时间',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_group_status_role` (`group_id`, `status`, `role`),
  KEY `idx_invited_by` (`invited_by`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员关系表（1级数据）';
```

注意：建议将 `expired_at` 从 `group_members` 拆出去，放到 `group_member_entitlements`，因为一个用户可能有多次续费、折扣、赠送权益，需要完整追踪。

---

### **2.3 `group_member_entitlements` 群组成员权益表**

用途：记录成员当前有效权益，支持付费加入、续费、赠送、嘉宾权益。

```sql
CREATE TABLE `group_member_entitlements` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `member_id` char(32) NOT NULL COMMENT 'group_members.id',
  `entitlement_type` varchar(20) NOT NULL COMMENT 'free/paid/guest/gift/admin_grant',
  `source_order_id` char(32) DEFAULT NULL COMMENT '来源订单',
  `source_invite_id` char(32) DEFAULT NULL COMMENT '来源邀请',
  `started_at` bigint(20) NOT NULL COMMENT '权益开始时间',
  `expired_at` bigint(20) DEFAULT NULL COMMENT '权益过期时间，NULL 表示永久',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=expired 3=revoked',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_user_status` (`group_id`, `user_id`, `status`),
  KEY `idx_member_status` (`member_id`, `status`),
  KEY `idx_expired_at` (`expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员权益表（1级数据）';
```

阅读付费主题时，应检查这张表，而不是只看 `group_members.role`。

---

### **2.4 `group_plans` 群组付费方案表**

用途：群主配置付费商品。

```sql
CREATE TABLE `group_plans` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `name` varchar(100) NOT NULL COMMENT '方案名称',
  `description` varchar(500) DEFAULT NULL,
  `plan_scope` varchar(20) NOT NULL DEFAULT 'group' COMMENT 'group/topic/all',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=inactive 3=deleted',
  `created_by` char(32) NOT NULL COMMENT '创建人，通常是群主',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_status` (`group_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组付费方案表（1级数据）';
```

---

### **2.5 `group_plan_periods` 付费周期表**

用途：群主控制月卡、季卡、年卡、永久卡等周期。

```sql
CREATE TABLE `group_plan_periods` (
  `id` char(32) NOT NULL,
  `plan_id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `period_type` varchar(20) NOT NULL COMMENT 'monthly/quarterly/yearly/lifetime/custom',
  `duration_days` int(11) DEFAULT NULL COMMENT '周期天数，lifetime 可为空',
  `price_cent` bigint(20) NOT NULL DEFAULT '0',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY',
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=inactive',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_plan_status` (`plan_id`, `status`),
  KEY `idx_group_status` (`group_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组付费周期表（1级数据）';
```

---

### **2.6 `group_discounts` 群组折扣周期表**

用途：群主设置折扣周期、首购折扣、续费折扣、邀请码折扣。

```sql
CREATE TABLE `group_discounts` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `plan_id` char(32) DEFAULT NULL COMMENT '为空表示适用于全部方案',
  `period_id` char(32) DEFAULT NULL COMMENT '为空表示适用于全部周期',
  `discount_type` varchar(20) NOT NULL COMMENT 'percent/fixed/first_buy/renew/invite',
  `discount_value` bigint(20) NOT NULL COMMENT 'percent=折扣百分比，fixed=减免金额分',
  `start_at` bigint(20) NOT NULL,
  `end_at` bigint(20) NOT NULL,
  `max_uses` int(11) DEFAULT NULL,
  `used_count` int(11) NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=inactive 3=expired',
  `created_by` char(32) NOT NULL,
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_time_status` (`group_id`, `start_at`, `end_at`, `status`),
  KEY `idx_plan_period` (`plan_id`, `period_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组折扣周期表（1级数据）';
```

---

### **2.7 `group_invites` 群组邀请表**

用途：群主、管理员、嘉宾发起邀请，支持嘉宾邀请分享。

```sql
CREATE TABLE `group_invites` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `inviter_id` char(32) NOT NULL COMMENT '邀请人',
  `invitee_id` char(32) DEFAULT NULL COMMENT '被邀请用户，可为空',
  `invite_code` varchar(64) NOT NULL COMMENT '邀请码',
  `invite_type` varchar(20) NOT NULL COMMENT 'guest/member/admin',
  `share_channel` varchar(30) DEFAULT NULL COMMENT 'link/wechat/system/private_message',
  `max_uses` int(11) NOT NULL DEFAULT '1',
  `used_count` int(11) NOT NULL DEFAULT '0',
  `expired_at` bigint(20) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=used 3=expired 4=revoked',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invite_code` (`invite_code`),
  KEY `idx_group_status` (`group_id`, `status`, `created_at`),
  KEY `idx_inviter` (`inviter_id`, `created_at`),
  KEY `idx_invitee` (`invitee_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组邀请表（1级数据）';
```

---

### **2.8 `group_admin_actions` 群组管理审计表**

用途：群主/管理员操作审计。

在当前 PRD 基础上扩展：

```sql
CREATE TABLE `group_admin_actions` (
  `id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `operator_id` char(32) NOT NULL,
  `target_user_id` char(32) DEFAULT NULL,
  `target_resource_id` char(32) DEFAULT NULL COMMENT '主题/评论/订单/邀请等资源 ID',
  `target_resource_type` varchar(30) DEFAULT NULL COMMENT 'member/topic/comment/order/invite',
  `action_type` varchar(30) NOT NULL COMMENT 'approve/ban/mute/remove/set_admin/invite_guest/update_plan/update_discount',
  `reason` varchar(500) DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  `created_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_operator` (`group_id`, `operator_id`, `created_at`),
  KEY `idx_target_user` (`target_user_id`, `created_at`),
  KEY `idx_resource` (`target_resource_type`, `target_resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组管理审计表（1级数据）';
```

---

## **3. 订单与付费域表设计**

虽然订单属于群组业务，但建议单独成子域，便于后续支付扩展。

### **3.1 `group_orders` 群组订单表**

```sql
CREATE TABLE `group_orders` (
  `id` char(32) NOT NULL,
  `order_no` varchar(64) NOT NULL,
  `user_id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `plan_id` char(32) NOT NULL,
  `period_id` char(32) NOT NULL,
  `discount_id` char(32) DEFAULT NULL,
  `order_type` varchar(20) NOT NULL COMMENT 'join/renew/gift/admin_grant',
  `original_amount_cent` bigint(20) NOT NULL DEFAULT '0',
  `discount_amount_cent` bigint(20) NOT NULL DEFAULT '0',
  `pay_amount_cent` bigint(20) NOT NULL DEFAULT '0',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=pending 2=paid 3=cancelled 4=refunded 5=failed 6=expired',
  `pay_channel` varchar(30) DEFAULT NULL,
  `paid_at` bigint(20) DEFAULT NULL,
  `entitlement_started_at` bigint(20) DEFAULT NULL,
  `entitlement_expired_at` bigint(20) DEFAULT NULL,
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status_time` (`user_id`, `status`, `created_at`),
  KEY `idx_group_status_time` (`group_id`, `status`, `created_at`),
  KEY `idx_plan_period` (`plan_id`, `period_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组订单表（1级数据）';
```

订单支付成功后，必须写入或续期 `group_member_entitlements`。

---

### **3.2 `group_order_events` 订单事件表**

用途：订单状态流转审计。

```sql
CREATE TABLE `group_order_events` (
  `id` char(32) NOT NULL,
  `order_id` char(32) NOT NULL,
  `order_no` varchar(64) NOT NULL,
  `from_status` tinyint(4) DEFAULT NULL,
  `to_status` tinyint(4) NOT NULL,
  `event_type` varchar(30) NOT NULL COMMENT 'created/paid/cancelled/refunded/expired/renewed',
  `operator_id` char(32) DEFAULT NULL,
  `reason` varchar(500) DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  `created_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order` (`order_id`, `created_at`),
  KEY `idx_order_no` (`order_no`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组订单事件表（1级审计数据）';
```

---

## **4. 主题域表设计**

### **4.1 `topics` 主题主表，已有，保留并优化**

建议新增或明确字段：

| 字段 | 等级 | 说明 |
|---|---|---|
| `group_id` | 1级 | 所属群组 |
| `author_id` | 1级 | 作者 |
| `visibility` | 1级 | public/group_member/paid_member/guest/admin_owner |
| `share_enabled` | 1级 | 是否允许分享 |
| `comment_enabled` | 1级 | 是否允许评论 |
| `status` | 1级 | draft/published/auditing/rejected/deleted |
| `last_activity_at` | 2级 | 最后互动时间 |

如果不能直接改已有表，就在 PRD 中定义字段映射：`allow_comments` 对应评论开关，新增 `share_enabled` 或使用扩展字段。

---

### **4.2 `topic_comments` 评论表**

保留当前 PRD 设计，建议增加被回复用户字段：

```sql
CREATE TABLE `topic_comments` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `parent_id` char(32) DEFAULT NULL,
  `reply_to_user_id` char(32) DEFAULT NULL,
  `content` text NOT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=normal 2=pending 3=hidden 4=deleted',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_topic_status` (`topic_id`, `status`, `created_at`),
  KEY `idx_user_comments` (`user_id`, `created_at`),
  KEY `idx_reply_to_user` (`reply_to_user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题评论表（1级数据）';
```

评论产生消息：通知主题作者或被回复用户。

---

### **4.3 `topic_reactions` 主题互动表**

保留当前 PRD 设计，支持点赞、收藏、分享行为记录。

```sql
CREATE TABLE `topic_reactions` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `reaction_type` varchar(20) NOT NULL COMMENT 'like/favorite/share',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=cancelled',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_topic_reaction` (`user_id`, `topic_id`, `reaction_type`),
  KEY `idx_topic_reaction` (`topic_id`, `reaction_type`, `status`),
  KEY `idx_group_reaction` (`group_id`, `reaction_type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题互动表（1级数据）';
```

---

### **4.4 `topic_shares` 主题分享表**

用途：主题和群主、嘉宾邀请都可以分享。主题分享必须单独记录，便于统计和权限控制。

```sql
CREATE TABLE `topic_shares` (
  `id` char(32) NOT NULL,
  `topic_id` char(32) NOT NULL,
  `group_id` char(32) NOT NULL,
  `sharer_id` char(32) NOT NULL COMMENT '分享人',
  `share_code` varchar(64) NOT NULL COMMENT '分享码',
  `share_channel` varchar(30) DEFAULT NULL COMMENT 'link/wechat/private_message/system',
  `visibility_snapshot` tinyint(4) NOT NULL COMMENT '分享时主题可见性快照',
  `expired_at` bigint(20) DEFAULT NULL,
  `view_count` bigint(20) NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=expired 3=revoked',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_share_code` (`share_code`),
  KEY `idx_topic_status` (`topic_id`, `status`, `created_at`),
  KEY `idx_sharer` (`sharer_id`, `created_at`),
  KEY `idx_group` (`group_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题分享表（1级数据）';
```

关键规则：

```text
分享不等于绕过权限。
用户通过分享链接访问主题时，仍然必须经过 CanReadTopic。
如果主题是付费成员可见，分享页只能展示摘要和付费/加入提示。
```

---

### **4.5 `topic_read_records` 阅读记录表**

保留当前 PRD 设计，但需要补充黑名单判断：

```text
MarkTopicRead 只有在 CanReadTopic 通过后才允许写入；
被作者拉黑的用户不能写入该作者主题的阅读记录。
```

---

## **5. 消息中心表设计**

消息中心是新需求，需要完整补充。

### **5.1 `conversations` 会话表**

用途：私信、系统消息、群组通知会话聚合。

```sql
CREATE TABLE `conversations` (
  `id` char(32) NOT NULL,
  `conversation_type` varchar(20) NOT NULL COMMENT 'private/system/group/order/interaction',
  `group_id` char(32) DEFAULT NULL,
  `owner_user_id` char(32) DEFAULT NULL COMMENT '单用户消息盒子归属',
  `peer_user_id` char(32) DEFAULT NULL COMMENT '私信对端用户',
  `last_message_id` char(32) DEFAULT NULL,
  `last_message_at` bigint(20) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=active 2=archived 3=deleted',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_owner_type_time` (`owner_user_id`, `conversation_type`, `last_message_at`),
  KEY `idx_group_time` (`group_id`, `last_message_at`),
  KEY `idx_peer` (`peer_user_id`, `last_message_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息会话表（1级数据）';
```

---

### **5.2 `messages` 消息主表**

用途：统一存储私信、通知、订单变更、点赞评论消息。

```sql
CREATE TABLE `messages` (
  `id` char(32) NOT NULL,
  `conversation_id` char(32) DEFAULT NULL,
  `message_type` varchar(30) NOT NULL COMMENT 'private_text/group_status/order_status/like/comment/system',
  `sender_id` char(32) DEFAULT NULL COMMENT '发送人，系统消息可为空',
  `receiver_id` char(32) NOT NULL COMMENT '接收人',
  `group_id` char(32) DEFAULT NULL,
  `topic_id` char(32) DEFAULT NULL,
  `comment_id` char(32) DEFAULT NULL,
  `order_id` char(32) DEFAULT NULL,
  `title` varchar(200) DEFAULT NULL,
  `content` text DEFAULT NULL,
  `payload` json DEFAULT NULL COMMENT '结构化扩展内容',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=normal 2=recalled 3=deleted',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_receiver_type_time` (`receiver_id`, `message_type`, `created_at`),
  KEY `idx_conversation_time` (`conversation_id`, `created_at`),
  KEY `idx_group_time` (`group_id`, `created_at`),
  KEY `idx_topic_time` (`topic_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息主表（1级数据）';
```

---

### **5.3 `message_receipts` 消息状态表**

用途：已读、未读、删除、归档。

```sql
CREATE TABLE `message_receipts` (
  `id` char(32) NOT NULL,
  `message_id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `conversation_id` char(32) DEFAULT NULL,
  `read_at` bigint(20) DEFAULT NULL,
  `deleted_at` bigint(20) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=unread 2=read 3=deleted',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_user` (`message_id`, `user_id`),
  KEY `idx_user_status_time` (`user_id`, `status`, `created_at`),
  KEY `idx_conversation_user_status` (`conversation_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息回执表（1级数据）';
```

未读数是 2级数据，可以放 Redis。

---

### **5.4 `message_delivery_logs` 消息投递日志表**

用途：记录消息队列投递、失败重试、幂等。

```sql
CREATE TABLE `message_delivery_logs` (
  `id` char(32) NOT NULL,
  `event_id` char(64) NOT NULL,
  `message_id` char(32) DEFAULT NULL,
  `receiver_id` char(32) NOT NULL,
  `channel` varchar(30) NOT NULL COMMENT 'inbox/ws/push/email',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=pending 2=sent 3=failed 4=ignored',
  `retry_count` int(11) NOT NULL DEFAULT '0',
  `last_error` varchar(500) DEFAULT NULL,
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_event_receiver_channel` (`event_id`, `receiver_id`, `channel`),
  KEY `idx_receiver_status` (`receiver_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息投递日志表（1级审计数据）';
```

---

## **6. 统计快照表设计**

### **6.1 `member_stats` 成员统计表**

关注需求取消后，统计口径改成群组社交相关。

```sql
CREATE TABLE `member_stats` (
  `user_id` char(32) NOT NULL,
  `joined_groups_count` int(11) NOT NULL DEFAULT '0',
  `owned_groups_count` int(11) NOT NULL DEFAULT '0',
  `guest_groups_count` int(11) NOT NULL DEFAULT '0',
  `topics_count` int(11) NOT NULL DEFAULT '0',
  `comments_count` int(11) NOT NULL DEFAULT '0',
  `unread_messages_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成员统计快照表（2级数据，可重建）';
```

---

### **6.2 `group_stats` 群组统计表**

```sql
CREATE TABLE `group_stats` (
  `group_id` char(32) NOT NULL,
  `members_count` int(11) NOT NULL DEFAULT '0',
  `paid_members_count` int(11) NOT NULL DEFAULT '0',
  `guest_members_count` int(11) NOT NULL DEFAULT '0',
  `admin_members_count` int(11) NOT NULL DEFAULT '0',
  `topics_count` int(11) NOT NULL DEFAULT '0',
  `comments_count` int(11) NOT NULL DEFAULT '0',
  `orders_count` int(11) NOT NULL DEFAULT '0',
  `renew_orders_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组统计快照表（2级数据，可重建）';
```

---

### **6.3 `topic_stats` 主题统计表**

```sql
CREATE TABLE `topic_stats` (
  `topic_id` char(32) NOT NULL,
  `read_count` int(11) NOT NULL DEFAULT '0',
  `unique_read_count` int(11) NOT NULL DEFAULT '0',
  `comments_count` int(11) NOT NULL DEFAULT '0',
  `likes_count` int(11) NOT NULL DEFAULT '0',
  `favorites_count` int(11) NOT NULL DEFAULT '0',
  `shares_count` int(11) NOT NULL DEFAULT '0',
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主题统计快照表（2级数据，可重建）';
```

---

### **四、补充后的社交规则**

#### **1. 成员身份规则**

群组内正式角色只有 4 种，游客不入成员表：

```text
member  普通成员
guest   嘉宾
admin   管理员
owner   群组所有人
```

游客是访问态：

```text
未登录用户或未加入群组用户 = 游客访问态
```

角色权限建议如下：

| 权限 | 游客 | 普通成员 | 嘉宾 | 管理员 | 群组所有人 |
|---|---:|---:|---:|---:|---:|
| 浏览公开群组 | 是 | 是 | 是 | 是 | 是 |
| 查看公开主题摘要 | 是 | 是 | 是 | 是 | 是 |
| 查看成员主题 | 否 | 是 | 视邀请权限 | 是 | 是 |
| 查看付费主题 | 否 | 付费后 | 视邀请权限 | 是 | 是 |
| 发布主题 | 否 | 是 | 可配置 | 是 | 是 |
| 评论点赞 | 否 | 是 | 可配置 | 是 | 是 |
| 分享主题 | 否 | 是 | 是 | 是 | 是 |
| 邀请嘉宾 | 否 | 否 | 可配置 | 是 | 是 |
| 管理成员 | 否 | 否 | 否 | 是 | 是 |
| 设置付费周期 | 否 | 否 | 否 | 否 | 是 |
| 设置折扣周期 | 否 | 否 | 否 | 否 | 是 |
| 转让群组 | 否 | 否 | 否 | 否 | 是 |

---

#### **2. 黑名单规则**

```text
成员 A 拉黑成员 B 后：
1. B 不能给 A 发私信；
2. B 不能查看 A 发布的非公开主题完整内容；
3. B 不能评论 A 的主题；
4. B 不能点赞/收藏 A 的主题；
5. A 的主题在 B 的列表中应被过滤，或仅展示“内容不可见”占位；
6. 群组所有人和管理员进行群组治理时不受普通拉黑限制；
7. 黑名单不得影响订单、系统通知、违规处理等平台必要消息。
```

权限服务中新增：

```go
CanSendPrivateMessage(senderID, receiverID string) error
CanViewAuthorTopic(viewerID, authorID, topicID string) error
IsBlockedBetween(userA, userB string) (bool, error)
```

---

#### **3. 付费与续费规则**

```text
1. 群组所有人可以创建付费方案；
2. 群组所有人可以配置不同付费周期；
3. 群组所有人可以配置折扣周期；
4. 成员可以付费加入；
5. 成员可以续费；
6. 续费时如果当前权益未过期，则从原 expired_at 往后延长；
7. 续费时如果当前权益已过期，则从当前时间重新计算；
8. 支付成功后必须生成订单事件和消息通知；
9. 权益开通必须写入 group_member_entitlements；
10. 权益变化后必须失效 group:member、group:entitlement、topic 权限相关缓存。
```

---

#### **4. 嘉宾邀请规则**

```text
1. 嘉宾必须通过邀请加入；
2. 邀请人可以是群组所有人、管理员，或被授权的嘉宾；
3. 嘉宾是否能发帖、评论、分享、继续邀请，由邀请权限或群组配置决定；
4. 嘉宾邀请链接必须有过期时间、使用次数、状态；
5. 嘉宾加入后写入 group_members(role=guest)；
6. 嘉宾权益写入 group_member_entitlements(entitlement_type=guest)；
7. 嘉宾邀请和接受邀请都要产生消息通知。
```

---

#### **5. 主题分享规则**

```text
1. 主题作者可以分享自己的主题；
2. 群组所有人可以分享群组内主题；
3. 嘉宾在权限允许时可以分享主题；
4. 分享链接不绕过主题权限；
5. 付费主题分享后，未付费用户只能看到摘要和加入/付费提示；
6. 分享行为写入 topic_shares；
7. 分享次数写入 topic_stats.shares_count；
8. 分享成功可以产生消息或事件。
```

---

#### **6. 消息中心规则**

消息中心需要支持：

```text
私信
群组状态通知
订单变更通知
点赞通知
评论通知
系统通知
```

消息类型建议：

| message_type | 说明 |
|---|---|
| `private_text` | 私信文本 |
| `group_joined` | 入群成功 |
| `group_invited` | 被邀请加入群组 |
| `group_removed` | 被移除群组 |
| `group_muted` | 被禁言 |
| `order_created` | 订单创建 |
| `order_paid` | 支付成功 |
| `order_renewed` | 续费成功 |
| `order_expiring` | 即将过期 |
| `topic_liked` | 主题被点赞 |
| `topic_commented` | 主题被评论 |
| `comment_replied` | 评论被回复 |
| `system_notice` | 系统通知 |

---

### **五、协议组优化设计**

当前协议组保持：

```text
成员协议组 maxType = 1000
群组协议组 maxType = 2000
主题协议组 maxType = 3000
```

但需要删除关注协议，补充黑名单、消息、邀请、续费、分享协议。

---

## **1. 成员协议组 1000**

| minType | 协议 | 数据等级 | 说明 |
|---:|---|---|---|
| 1001 | MemberRegister | 1级 | 注册 |
| 1002 | MemberLogin | 1级 | 登录 |
| 1003 | MemberLogout | 1级 | 登出 |
| 1004 | GetMemberProfile | 1级读取 | 获取资料 |
| 1005 | UpdateMemberProfile | 1级 | 修改资料 |
| 1010 | BlockMember | 1级 | 拉黑成员 |
| 1011 | UnblockMember | 1级 | 取消拉黑 |
| 1012 | ListBlockedMembers | 1级读取 | 黑名单列表 |
| 1020 | GetMemberStats | 2级读取 | 成员统计 |
| 1030 | UpdateMemberStatus | 1级高权限 | 平台禁用/恢复用户 |
| 1040 | SendPrivateMessage | 1级 | 发送私信 |
| 1041 | ListConversations | 1级+2级读取 | 会话列表 |
| 1042 | ListMessages | 1级读取 | 消息列表 |
| 1043 | MarkMessageRead | 1级写入 + 2级统计 | 标记已读 |
| 1044 | GetUnreadMessageCount | 2级读取 | 未读数 |

建议从 PRD 删除：

```text
FollowMember
UnfollowMember
ListFollowers
ListFollowings
```

如果产品未来要关注，可移到 P2。

---

## **2. 群组协议组 2000**

| minType | 协议 | 数据等级 | 说明 |
|---:|---|---|---|
| 2001 | CreateGroup | 1级 | 创建群组 |
| 2002 | UpdateGroup | 1级 | 修改群组 |
| 2003 | GetGroupDetail | 1级+2级读取 | 群组详情 |
| 2004 | ListGroups | 2级读取 | 群组列表 |
| 2010 | JoinGroupFree | 1级 | 免费加入 |
| 2011 | LeaveGroup | 1级 | 退出群组 |
| 2012 | ListGroupMembers | 2级读取 | 成员列表 |
| 2013 | UpdateGroupMemberStatus | 1级高权限 | 禁言/移除/恢复 |
| 2014 | InviteGroupGuest | 1级 | 邀请嘉宾 |
| 2015 | AcceptGroupInvite | 1级 | 接受邀请 |
| 2016 | ListGroupInvites | 1级读取 | 邀请列表 |
| 2020 | CreateGroupPlan | 1级 | 创建付费方案 |
| 2021 | UpdateGroupPlan | 1级 | 修改付费方案 |
| 2022 | ListGroupPlans | 1级读取 | 方案列表 |
| 2023 | CreateGroupPlanPeriod | 1级 | 创建付费周期 |
| 2024 | UpdateGroupPlanPeriod | 1级 | 修改付费周期 |
| 2025 | CreateGroupDiscount | 1级 | 创建折扣周期 |
| 2026 | UpdateGroupDiscount | 1级 | 修改折扣周期 |
| 2030 | CreateGroupOrder | 1级 | 创建入群订单 |
| 2031 | ConfirmGroupPayment | 1级高权限 | 支付确认 |
| 2032 | RenewGroupMembership | 1级 | 续费 |
| 2033 | GetGroupEntitlement | 1级读取 | 查询当前权益 |
| 2040 | GetOwnerDashboard | 2级读取 | 群主看板 |
| 2050 | AuditGroup | 1级高权限 | 审核群组 |

---

## **3. 主题协议组 3000**

| minType | 协议 | 数据等级 | 说明 |
|---:|---|---|---|
| 3001 | CreateTopic | 1级 | 发布主题 |
| 3002 | UpdateTopic | 1级 | 修改主题 |
| 3003 | DeleteTopic | 1级 | 删除/下架主题 |
| 3004 | GetTopicDetail | 1级+2级读取 | 主题详情，含权限和黑名单判断 |
| 3005 | ListGroupTopics | 2级读取 | 群组主题列表 |
| 3010 | MarkTopicRead | 2级写入 | 标记阅读 |
| 3011 | GetReadHistory | 2级读取 | 阅读历史 |
| 3020 | CreateComment | 1级 | 评论 |
| 3021 | DeleteComment | 1级 | 删除评论 |
| 3030 | ReactTopic | 1级 | 点赞/收藏 |
| 3031 | CancelReactTopic | 1级 | 取消点赞/收藏 |
| 3032 | ShareTopic | 1级 | 分享主题 |
| 3033 | GetTopicShare | 1级读取 | 获取分享信息 |
| 3040 | AuditTopic | 1级高权限 | 审核主题 |
| 3050 | GetTopicStats | 2级读取 | 主题统计 |

---

### **六、Redis + Go + 消息队列架构设计**

你提出基于 Redis + Go + 消息队列设计。建议采用：

```text
Go Social Module
  ↓
MySQL 存储 1级事实数据
  ↓
Redis 缓存 1级短 TTL 数据 + 2级统计/列表/未读数
  ↓
消息队列处理领域事件、通知投递、统计更新、缓存失效
```

如果当前项目没有引入独立 MQ，可以先用：

```text
MVP 阶段：Redis Streams
后续阶段：Kafka / Pulsar / RabbitMQ
```

这里推荐 **Redis Streams**，因为当前架构已有 Redis 缓存语义，Go 端集成简单，适合 MVP。

---

## **1. Go 模块架构**

```text
go/modules/social/
  member/
    handler.go
    usecase.go
    service.go
    repository.go
  group/
    handler.go
    usecase.go
    service.go
    repository.go
  topic/
    handler.go
    usecase.go
    service.go
    repository.go
  message/
    handler.go
    usecase.go
    service.go
    repository.go
  permission/
    service.go
  payment/
    service.go
  invite/
    service.go
  cache/
    keys.go
    policy.go
    invalidator.go
  event/
    publisher.go
    subscriber.go
    events.go
    stream.go
  mq/
    redis_stream.go
    consumer_group.go
  model/
    *.go
```

---

## **2. Redis Key 设计**

### **成员相关**

```text
member:profile:{userId}
member:session:{tokenId}
member:stats:{userId}
member:block:rel:{blockerId}:{blockedId}
member:block:list:{userId}:{page}
```

### **群组相关**

```text
group:detail:{groupId}
group:member:{groupId}:{userId}
group:entitlement:{groupId}:{userId}
group:members:{groupId}:{role}:{page}
group:plans:{groupId}
group:periods:{planId}
group:discounts:{groupId}
group:invite:{inviteCode}
group:stats:{groupId}
owner:dashboard:{ownerId}
```

### **主题相关**

```text
topic:detail:{topicId}
topic:permission:{topicId}:{userId}
topic:stats:{topicId}
topic:read:{userId}:{topicId}
topic:share:{shareCode}
group:topics:{groupId}:{page}
```

### **消息中心相关**

```text
message:unread:{userId}
message:unread:{userId}:{conversationType}
message:conversation:list:{userId}:{page}
message:conversation:{conversationId}
message:recent:{userId}
```

---

## **3. 消息队列 Topic / Stream 设计**

使用 Redis Streams 时，建议按领域拆 stream：

```text
stream:social:member
stream:social:group
stream:social:topic
stream:social:order
stream:social:message
stream:social:cache
```

消费者组：

```text
cg:stats-updater
cg:cache-invalidator
cg:message-dispatcher
cg:notification-writer
cg:audit-writer
```

---

## **4. 领域事件设计**

### **成员事件**

```text
MemberRegistered
MemberBlocked
MemberUnblocked
MemberStatusChanged
PrivateMessageSent
MessageRead
```

### **群组事件**

```text
GroupCreated
GroupUpdated
GroupJoinedFree
GroupLeft
GroupGuestInvited
GroupInviteAccepted
GroupMemberMuted
GroupMemberRemoved
GroupMemberRecovered
GroupPlanCreated
GroupPeriodUpdated
GroupDiscountCreated
GroupAudited
```

### **订单事件**

```text
GroupOrderCreated
GroupOrderPaid
GroupMembershipRenewed
GroupEntitlementActivated
GroupEntitlementExpired
```

### **主题事件**

```text
TopicCreated
TopicUpdated
TopicDeleted
TopicRead
TopicCommentCreated
TopicLiked
TopicUnliked
TopicShared
TopicAudited
```

### **消息事件**

```text
MessageCreated
MessageDelivered
MessageRead
UnreadCountChanged
```

---

## **5. 事件处理链路示例**

### **成员拉黑**

```text
BlockMember
  ↓
写入 member_blocks
  ↓
发布 MemberBlocked
  ↓
CacheInvalidator 删除：
  member:block:rel:{blocker}:{blocked}
  topic:permission:*:{blocked}
  ↓
MessageDispatcher 可选通知
  ↓
后续 blocked 用户访问 blocker 主题时 CanReadTopic 拒绝
```

---

### **付费加入群组**

```text
CreateGroupOrder
  ↓
写入 group_orders(status=pending)
  ↓
发布 GroupOrderCreated
  ↓
写入订单消息 order_created
  ↓
ConfirmGroupPayment
  ↓
更新 group_orders(status=paid)
  ↓
写入 group_member_entitlements
  ↓
更新 group_members(active)
  ↓
发布 GroupOrderPaid / GroupEntitlementActivated
  ↓
CacheInvalidator 删除：
  group:member:{groupId}:{userId}
  group:entitlement:{groupId}:{userId}
  topic:permission:*:{userId}
  ↓
StatsUpdater 更新 group:stats
  ↓
MessageDispatcher 写入 order_paid 消息
```

---

### **主题被评论**

```text
CreateComment
  ↓
CanReadTopic + CanCommentTopic
  ↓
写入 topic_comments
  ↓
发布 TopicCommentCreated
  ↓
StatsUpdater 更新 topic:stats.comments_count
  ↓
CacheInvalidator 删除 topic:detail:{topicId}
  ↓
MessageDispatcher 给主题作者/被回复用户写消息
  ↓
UnreadCounter 更新 message:unread:{receiverId}
```

---

### **主题分享**

```text
ShareTopic
  ↓
CanShareTopic
  ↓
写入 topic_shares
  ↓
发布 TopicShared
  ↓
StatsUpdater 更新 topic:stats.shares_count
  ↓
MessageDispatcher 可选发送分享消息
  ↓
访问分享链接时仍执行 CanReadTopic
```

---

### **七、权限服务补充设计**

现有 PRD 的 PermissionService 需要扩展：

```go
type PermissionService interface {
    CanViewGroup(userID, groupID string) error
    CanJoinGroup(userID, groupID string) error
    CanReadTopic(userID, topicID string) (AccessResult, error)
    CanCommentTopic(userID, topicID string) error
    CanReactTopic(userID, topicID string) error
    CanShareTopic(userID, topicID string) error

    CanManageGroup(operatorID, groupID string) error
    CanManageMember(operatorID, groupID, targetUserID string) error
    CanInviteGuest(operatorID, groupID string) error
    CanManagePlan(operatorID, groupID string) error
    CanManageDiscount(operatorID, groupID string) error

    CanSendPrivateMessage(senderID, receiverID string) error
    IsBlockedBetween(userA, userB string) (bool, error)
}
```

`CanReadTopic` 新规则：

```text
1. 读取 topic 1级数据；
2. 检查 topic.status；
3. 检查用户是否被作者拉黑；
4. 检查 group_members 角色和状态；
5. 检查 group_member_entitlements 是否有效；
6. 检查游客是否只允许看公开内容；
7. 返回 ALLOWED / NEED_LOGIN / NEED_JOIN / NEED_PAY / BLOCKED / FORBIDDEN。
```

---

### **八、需要 Trae 修改 PRD 的具体指令**

下面这段可以直接发给本地 Trae。

```text
请评审并重构 docs/prd/PRD-social-app-mvp.md。当前文档方向基本正确，但产品模型需要从“关注粉丝社交”调整为“群组关系维护社交”。

一、必须修改的核心点

1. 删除或降级关注需求：
   - 删除 P0 中的 FollowMember、UnfollowMember、ListFollowers、ListFollowings。
   - user_follows 不再作为 P0 核心表。
   - 如需保留，标记为 P2 后续扩展。

2. 明确当前社交关系核心是群组关系维护：
   - 用户免费加入群组。
   - 用户付费加入群组。
   - 用户续费群组。
   - 嘉宾被邀请加入群组。
   - 群组所有人和管理员管理成员。
   - 成员之间可以相互拉黑。
   - 主题阅读权限受群组成员、付费权益、嘉宾身份、黑名单共同影响。

3. 成员身份改为 5 类访问/角色：
   - 游客：未加入群组的访问态，不写入 group_members。
   - 普通成员：role=member。
   - 嘉宾：role=guest，通过邀请加入。
   - 管理员：role=admin。
   - 群组所有人：role=owner。

4. 补充黑名单规则：
   - 新增 member_blocks 表。
   - BlockMember / UnblockMember / ListBlockedMembers 协议。
   - 拉黑后，被拉黑人不能获取拉黑人发布的受限主题完整内容。
   - 拉黑后，被拉黑人不能私信、评论、点赞拉黑人内容。
   - 管理员和群组所有人执行治理操作时可绕过普通拉黑，但必须审计。

5. 补充付费周期和折扣周期：
   - 群组所有人可以配置付费方案、付费周期、折扣周期。
   - 新增 group_plans、group_plan_periods、group_discounts。
   - 成员可以付费加入和续费。
   - 续费要正确计算权益时间。
   - 新增 group_member_entitlements 表记录权益。

6. 补充嘉宾邀请：
   - 新增 group_invites 表。
   - 群主、管理员、被授权嘉宾可以邀请嘉宾。
   - 嘉宾加入后写入 group_members(role=guest)。
   - 嘉宾权益写入 group_member_entitlements(entitlement_type=guest)。

7. 补充主题和邀请分享：
   - 新增 topic_shares 表。
   - 主题作者、群组所有人、嘉宾在权限允许时可以分享主题。
   - 分享链接不能绕过权限。
   - 付费主题分享后，未付费用户只能看到摘要和购买/加入提示。

8. 补充消息中心：
   - 新增 conversations、messages、message_receipts、message_delivery_logs。
   - 支持私信、群组状态通知、订单变更通知、点赞通知、评论通知、系统通知。
   - 未读数作为 2级数据存入 Redis。
   - 消息历史作为 1级数据写入 MySQL。

二、数据库表设计需要调整

请在 PRD 的“数据库表设计”章节中按以下域拆分：

成员域：
- users：已有，保留。
- member_sessions：用户会话。
- member_blocks：成员拉黑关系。

群组域：
- groups：已有，保留。
- group_members：群组成员关系，role=member/guest/admin/owner。
- group_member_entitlements：群组成员权益。
- group_plans：群组付费方案。
- group_plan_periods：付费周期。
- group_discounts：折扣周期。
- group_invites：嘉宾邀请。
- group_admin_actions：群组管理审计。

订单域：
- group_orders：群组订单，支持 join/renew/gift/admin_grant。
- group_order_events：订单状态流转事件。

主题域：
- topics：已有，保留。
- topic_comments：评论。
- topic_reactions：点赞/收藏/分享行为。
- topic_shares：主题分享。
- topic_read_records：阅读记录。

消息域：
- conversations：会话。
- messages：消息主表。
- message_receipts：消息回执。
- message_delivery_logs：消息投递日志。

统计域：
- member_stats：成员统计。
- group_stats：群组统计。
- topic_stats：主题统计。

三、协议组调整

保持当前协议组：
- 成员协议组 maxType=1000
- 群组协议组 maxType=2000
- 主题协议组 maxType=3000

成员协议组需要调整为：
1001 MemberRegister
1002 MemberLogin
1003 MemberLogout
1004 GetMemberProfile
1005 UpdateMemberProfile
1010 BlockMember
1011 UnblockMember
1012 ListBlockedMembers
1020 GetMemberStats
1030 UpdateMemberStatus
1040 SendPrivateMessage
1041 ListConversations
1042 ListMessages
1043 MarkMessageRead
1044 GetUnreadMessageCount

群组协议组需要调整为：
2001 CreateGroup
2002 UpdateGroup
2003 GetGroupDetail
2004 ListGroups
2010 JoinGroupFree
2011 LeaveGroup
2012 ListGroupMembers
2013 UpdateGroupMemberStatus
2014 InviteGroupGuest
2015 AcceptGroupInvite
2016 ListGroupInvites
2020 CreateGroupPlan
2021 UpdateGroupPlan
2022 ListGroupPlans
2023 CreateGroupPlanPeriod
2024 UpdateGroupPlanPeriod
2025 CreateGroupDiscount
2026 UpdateGroupDiscount
2030 CreateGroupOrder
2031 ConfirmGroupPayment
2032 RenewGroupMembership
2033 GetGroupEntitlement
2040 GetOwnerDashboard
2050 AuditGroup

主题协议组需要调整为：
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
3032 ShareTopic
3033 GetTopicShare
3040 AuditTopic
3050 GetTopicStats

四、Redis + Go + 消息队列架构

请在 PRD 中新增“Redis + Go + 消息队列事件架构”章节。

MVP 阶段建议使用 Redis Streams 作为消息队列：

Streams：
- stream:social:member
- stream:social:group
- stream:social:topic
- stream:social:order
- stream:social:message
- stream:social:cache

消费者组：
- cg:stats-updater
- cg:cache-invalidator
- cg:message-dispatcher
- cg:notification-writer
- cg:audit-writer

Go 模块建议：
go/modules/social/member
go/modules/social/group
go/modules/social/topic
go/modules/social/message
go/modules/social/permission
go/modules/social/payment
go/modules/social/invite
go/modules/social/cache
go/modules/social/event
go/modules/social/mq

五、Redis Key 规范

请补充以下 key：

member:profile:{userId}
member:session:{tokenId}
member:stats:{userId}
member:block:rel:{blockerId}:{blockedId}
member:block:list:{userId}:{page}

group:detail:{groupId}
group:member:{groupId}:{userId}
group:entitlement:{groupId}:{userId}
group:members:{groupId}:{role}:{page}
group:plans:{groupId}
group:periods:{planId}
group:discounts:{groupId}
group:invite:{inviteCode}
group:stats:{groupId}
owner:dashboard:{ownerId}

topic:detail:{topicId}
topic:permission:{topicId}:{userId}
topic:stats:{topicId}
topic:read:{userId}:{topicId}
topic:share:{shareCode}
group:topics:{groupId}:{page}

message:unread:{userId}
message:unread:{userId}:{conversationType}
message:conversation:list:{userId}:{page}
message:conversation:{conversationId}
message:recent:{userId}

六、权限服务扩展

请扩展 PermissionService：

CanViewGroup
CanJoinGroup
CanReadTopic
CanCommentTopic
CanReactTopic
CanShareTopic
CanManageGroup
CanManageMember
CanInviteGuest
CanManagePlan
CanManageDiscount
CanSendPrivateMessage
IsBlockedBetween

权限规则必须明确：
- 权限判断不得依赖 2级统计数据。
- 主题阅读必须同时检查 topic 状态、群组成员关系、付费权益、嘉宾身份、黑名单。
- 分享不能绕过权限。
- 私信必须检查黑名单。
- 群组所有人和管理员治理行为必须写审计日志。

七、文档同步

请同步更新：
- docs/prd/PRD-social-app-mvp.md
- docs/adr/ADR-social-data-level-and-cache-strategy.md
- docs/api/social-openapi.yaml
- docs/api/协议编号注册表.md
- docs/wiki/CODE-WIKI.md

八、输出要求

完成后请输出：
1. PRD 修改摘要。
2. 删除/降级的关注需求清单。
3. 新增的群组关系、黑名单、嘉宾、付费周期、折扣周期、消息中心需求。
4. 数据库表设计清单。
5. 协议编号调整清单。
6. Redis Key 设计。
7. 消息队列事件设计。
8. 权限规则变更。
9. 测试验收标准。
```

---

### **九、评审后的主控结论**

这份 PRD 不建议直接进入开发，需要先做一次需求纠偏。纠偏方向是：

```text
从“关注/粉丝型社交”
调整为
“群组成员关系型社交”
```

重构后的核心模型应是：

```text
群组是社交容器；
成员关系是核心关系；
付费权益决定内容访问；
嘉宾通过邀请进入；
成员之间可拉黑；
主题阅读受角色、权益、黑名单共同控制；
消息中心承载私信、状态通知、订单通知、点赞评论；
Redis 负责缓存与未读数；
Go 服务负责协议、权限、事务和事件发布；
消息队列负责统计、通知、缓存失效和异步投递。
```

按这个方向改完后，PRD 才能真正匹配当前 CaiRobot MVP 运营管理综合平台的产品目标和工程架构。

*内容由 AI 生成仅供参考*