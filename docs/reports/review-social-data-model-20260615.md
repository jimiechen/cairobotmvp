# 数据模型评审意见：简化关注模型、专注群组社交

> **评审日期**：2026-06-15
> **评审对象**：`tabbit_项目主控与Trae开发指导对接.md`（v1 + v2）+ `basemodel.md`
> **评审视角**：简化关注模型 / 专注群组社交 / MVP 可落地性
> **关联文档**：PRD-social-app-mvp.md（已产出）、ADR-social-data-level-and-cache-strategy.md（已产出）

---

## 一、评审结论

| 维度 | 评价 | 说明 |
|---|---|---|
| 整体架构合理性 | **通过** | 单网关 + MessagePacket + 协议组(1000/2000/3000) + 1级/2级数据分级，与 CaiRobot MVP 架构完全对齐 |
| 基础表复用 | **通过** | users/groups/topics 三张基础表边界清晰，不堆砌动态关系 |
| 群组核心模型 | **通过** | group_members/group_plans/group_orders/group_admin_actions 四张扩展表设计合理 |
| 内容互动模型 | **通过** | topic_read_records/topic_comments/topic_reactions 职责分明 |
| **关注模型复杂度** | **需优化** | user_follows 作为独立一级表 + 4 个协议 + 事件链路，在"群组为核心"的定位下偏重 |
| **粉丝管理与成员管理重叠** | **需明确** | 圈主同时管理 user_follows（粉丝）和 group_members（成员），两个管理面存在语义混淆 |
| **协议数量** | **可精简** | 当前 22 对协议（44 个编号），关注相关协议占 ~18%，MVP 阶段可裁剪 |

---

## 二、核心问题：关注模型与群组社交定位的张力

### 2.1 问题陈述

当前设计将社交关系拆为三层：

```
关注关系：user_follows          ← 用户↔用户
群成员关系：group_members        ← 用户↔群组
内容互动关系：topic_reactions等  ← 用户↔内容
```

产品定位是 **"以圈子/群组为核心的内容型社交产品"**。在这个定位下：

- 用户的核心行为是 **加入群组 → 阅读帖子 → 互动**
- 圈主的核心行为是 **创建群组 → 发布帖子 → 管理成员**
- "关注圈主"是一个 **衍生行为**，不是主路径

但当前设计中，`user_follows` 占据了和 `group_members` 同等的模型地位：
- 独立 1 级表（6 字段 + 3 索引）
- 独立协议 4 个（FollowMember/UnfollowMember/ListFollowers/ListFollowings）
- 独立事件链路（MemberFollowed → 缓存失效 → 统计更新 → 通知）
- 独立缓存 key（`follow:rel:{f}:{t}`、`member:followers:{uid}:{page}` 等）

**这导致 MVP 的数据模型复杂度被一个非核心功能拉高了约 20%。**

### 2.2 具体矛盾点

| # | 矛盾 | 当前设计 | 群组社交定位下的期望 |
|---|---|---|---|
| 1 | **圈主"粉丝"vs"成员"语义重叠** | user_follows 存粉丝关系，group_members 存成员关系；圈主需要两套管理界面 | 圈主只关心"谁在我的群里"，粉丝=成员的一种视图 |
| 2 | **关注来源 source 字段过度设计** | source 支持 group/topic/profile/search 五种来源 | MVP 只需一种入口，source 可延后 |
| 3 | **ListFollowers/ListFollowings 与 ListGroupMembers 功能重叠** | 三个列表接口返回的数据高度相似（都是用户 ID 列表 + 分页） | 合并为统一的成员/关注者查询 |
| 4 | **粉丝数统计作为独立 2 级数据** | followers_count 由 user_follows 事件驱动更新 | 可从 group_members 聚合，或直接延后到 P1 |
| 5 | **FollowMember 协议独立于 JoinGroup** | 关注圈主 ≠ 加入群组，两条独立路径 | 关注 ≈ 加入圈主的公开/免费群，一条路径 |

---

## 三、优化建议

### 3.1 方案对比

#### 方案 A：保留 user_follows 但降级为 P1（推荐用于 v2 迭代）

**思路**：MVP 不实现独立关注功能，所有社交关系通过群组成员表达。

| 变更项 | 当前 | 优化后 |
|---|---|---|
| user_follows 表 | 1 级表，6 字段 | **移至 P1**，MVP 不建表 |
| FollowMember 协议 | minType=1011 | **移除** |
| UnfollowMember 协议 | minType=1013 | **移除** |
| ListFollowers 协议 | minType=1015 | **移除**，用 ListGroupMembers 替代 |
| ListFollowings 协议 | minType=1017 | **移除** |
| GetMemberStats 协议 | minType=1019 | **移除或合并入 GetMemberProfile** |
| MemberFollowed 事件 | 14 个事件之一 | **移除** |
| 相关缓存 key | 4 个 | **移除** |

**"关注圈主"的替代实现**：用户点击"关注"时，自动执行 `JoinGroup(group_id=圈主的公开默认群)`。这样：
- 关注关系 = 群组成员关系（join_source=follow）
- 粉丝列表 = 群组成员列表（过滤 role=member）
- 取消关注 = 退出群组

**优点**：
- 减少 1 张表、4~5 个协议、3+ 个事件、4+ 个缓存 key
- 模型更纯粹：用户只有一种社交关系——"我在哪些群里"
- 圈主只有一个管理面："我的群组成员"

**缺点**：
- 无法关注一个圈主但不加入其任何群组
- 如果圈主有多个群组，关注哪个存在歧义（需要一个"默认关注群"概念）

**适用场景**：MVP 阶段快速验证群组社交核心闭环。

---

#### 方案 B：轻量化 user_follows 为"订阅书签"（折中方案）

**思路**：保留 user_follows 表，但大幅简化字段和协议，将其定位为"订阅/书签"而非完整的粉丝体系。

**表结构简化**：

```sql
-- 精简版：去掉 source、status 复杂状态，只保留 active/deleted
CREATE TABLE `user_follows` (
  `id` char(32) NOT NULL,
  `follower_id` char(32) NOT NULL,
  `following_id` char(32) NOT NULL,
  `group_id` char(32) DEFAULT NULL COMMENT '关注的触发来源群组，可为空',
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_follow` (`follower_id`, `following_id`),
  KEY `idx_following_time` (`following_id`, `created_at`)
) ENGINE=InnoDB COMMENT='用户关注/订阅（轻量版）';
```

**变更点**：

| 变更项 | 当前(v2) | 优化后(B) |
|---|---|---|
| 字段数 | 7（id/follower/following/status/source/created/updated） | **5**（id/follower/following/group_id/created） |
| status 枚举 | active/cancelled/blocked | **删除**，物理删除即可 |
| source 枚举 | group/topic/profile/search | **替换为 group_id**，记录来源群组 |
| updated_at | 有 | **删除**，关注行为不需要更新 |
| 协议数 | 4（follow/unfollow/list-f/list-g） | **2**（follow/unfollow），列表合并入成员接口 |
| 事件 | MemberFollowed + MemberUnfollowed | **保留 1 个** MemberFollowed |

**协议精简**：

```
保留：
  1011 FollowMember       # 关注（自动关联可选群组）
  1013 UnfollowMember     # 取关

移除/合并：
  1015 ListFollowers      → 合并入 ListGroupMembers（加 filter=followers_only）
  1017 ListFollowings     → 合并入 ListGroups（加 filter=following_owners_groups）
  1019 GetMemberStats     → 合并入 GetMemberProfile（附加 stats 字段）
```

**优点**：
- 保留了"关注但不入群"的灵活性
- 大幅减少协议和事件数量
- group_id 冗余字段让"从哪个群关注"可追溯

**缺点**：
- 仍需维护两张关系表的语义区分
- 圈主仍有两个管理面（虽然粉丝面已简化）

---

#### 方案 C：合并为统一关系表（激进方案）

**思路**：用一张 `social_relations` 统一表达"用户与任何实体的关系"。

```sql
CREATE TABLE `social_relations` (
  `id` char(32) NOT NULL,
  `user_id` char(32) NOT NULL,
  `target_type` enum('group','user') NOT NULL,
  `target_id` char(32) NOT NULL,
  `relation_type` enum('member','follow','owner','admin','moderator') NOT NULL,
  `status` tinyint DEFAULT 1,
  `metadata` json DEFAULT NULL,  -- expired_at/muted_until/reason 等变长属性
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_relation` (`user_id`, `target_type`, `target_id`, `relation_type`),
  KEY `idx_target` (`target_type`, `target_id`, `status`),
  KEY `idx_user_type` (`user_id`, `target_type`, `status`)
) ENGINE=InnoDB COMMENT='统一社交关系表';
```

**优点**：极致简洁，一张表覆盖所有关系。

**缺点**：
- metadata JSON 牺牲了 SQL 类型安全
- 不同 relation_type 的字段差异大（成员有 expired_at，关注没有），JSON 不利于索引
- 与 basemodel.md 已有 groups/users/topics 的 ER 关系不一致
- **不推荐 MVP 采用**，过度抽象

---

### 3.1 推荐方案

**MVP 推荐：方案 A（移除独立关注，用群组成员替代）**
- 如果产品确认"群组是唯一社交单元"
- 快速落地，减少 20% 模型复杂度

**如果必须保留关注功能：方案 B（轻量化 user_follows）**
- 保留关注灵活性的同时大幅瘦身
- 减少约 50% 的关注相关协议和事件

---

## 四、其他数据模型评审意见

### 4.1 通过项

| # | 项 | 评价 |
|---|---|---|
| 1 | **users 不堆砌粉丝数/关注数** | 正确。计数字段走 2 级或 Redis |
| 2 | **groups 不保存动态成员关系** | 正确。members_count 仅作为冗余快照 |
| 3 | **topics.access_level 权限分级** | 正确。PUBLIC/GROUP_MEMBER/PAID_MEMBER/OWNER_ONLY 四级清晰 |
| 4 | **group_members.expired_at 付费过期** | 正确。权益过期时间直接挂在成员关系上 |
| 5 | **group_plans 独立于 groups** | 正确。一个群组可有多个付费方案 |
| 6 | **group_admin_actions 审计日志** | 正确。圈主操作必须有审计轨迹 |
| 7 | **topic_reactions 统一互动表** | 正确。like/favorite/share 不拆表，MVP 够用 |
| 8 | **1级/2级数据分级治理** | 正确。这是本次设计的最大亮点 |
| 9 | **权限铁律：不能依赖2级数据做权限判断** | 正确。防止缓存一致性漏洞 |
| 10 | **Cache Aside + 主动失效 vs 事件驱动** | 正确。两级缓存策略匹配两级数据 |

### 4.2 建议修改项

| # | 项 | 当前设计 | 建议 | 优先级 |
|---|---|---|---|---|
| 1 | **groups.members_count 写入时机** | 文档说"事务内同步更新，后续改异步" | MVP 直接用 **事件驱动异步更新**，避免写放大 | P1 |
| 2 | **topics.author_id 类型不一致** | basemodel 用 char(36)，users.id 用 char(32) | **统一为 char(32)**，FK 类型必须一致 | P0 |
| 3 | **topics.group_id 类型不一致** | topics 用 char(36)，groups.id 用 char(32) | **统一为 char(32)** | P0 |
| 4 | **topic_read_records.group_id 冗余** | 设计中有冗余 group_id | **保留**，有利于按群组查询阅读热度 | 通过 |
| 5 | **group_orders.order_no 无长度说明** | varchar(64) | 建议明确生成规则（如 `{yyyyMMdd}{HHmmss}{随机6位}`） | P2 |
| 6 | **user_follows.source 枚举值过多** | group/topic/profile/search | 如采用方案 B，**改为 group_id 外键** | P1 |
| 7 | **缺少 group_invites 表** | 邀请制群组的邀请码/链接管理未建模 | MVP 可用 join_mode=4 + invite_code 字段暂代，P1 再拆表 | P2 |
| 8 | **缺少 content_moderation 表** | 内容审核状态散落在 topics.status | MVP 用 topics.status 足够，P1 再拆分审核工作流表 | P2 |

### 4.3 与 PRD-social-app-mvp.md 已产出文档的差异

以下是在已产出 PRD 中发现的需要同步修正的点：

| # | 差异 | PRD 当前值 | 应修正为 |
|---|---|---|---|
| 1 | topics 主键类型 | PRD 未明确定义 DDL 类型 | 与 basemodel.md 对齐为 char(32) |
| 2 | user_follows 定位 | PRD 按 v2 设计为完整 1 级表 | 根据本评审结论决定是否降级/简化 |
| 3 | 成员协议组协议数 | PRD 包含 11 对（含关注 4 对） | 根据选定方案调整 |

---

## 五、协议影响评估

若采用 **方案 A（移除独立关注）**，协议组变化如下：

### 成员协议组（maxType=1000）：22 对 → 17 对

| minType | 协议名称 | 处置 |
|---|---|---|
| 1001/1002 | MemberRegister / Response | **保留** |
| 1003/1004 | MemberLogin / Response | **保留** |
| 1005/1006 | MemberLogout / Response | **保留** |
| 1007/1008 | GetMemberProfile / Response | **保留**，可内嵌基础统计 |
| 1009/1010 | UpdateMemberProfile / Response | **保留** |
| 1011/1012 | FollowMember / Response | **移除** 或改为 JoinDefaultGroup |
| 1013/1014 | UnfollowMember / Response | **移除** 或改为 LeaveDefaultGroup |
| 1015/1016 | ListFollowers / Response | **移除** |
| 1017/1018 | ListFollowings / Response | **移除** |
| 1019/1020 | GetMemberStats / Response | **合并入 GetMemberProfile** |
| 1021/1022 | UpdateMemberStatus / Response | **保留** |

**净减少**：5 对协议（10 个编号），从 11 对缩减为 **6 对（12 个编号）**。

### 群组协议组（maxType=2000）：15 对 → 不变

群组协议组不受关注模型简化的影响。

### 主题协议组（maxType=3000）：13 对 → 不变

主题协议组不受关注模型简化的影响。

### 总量变化

| 指标 | 当前（v2 已修正） | 方案A后 | 变化率 |
|---|---|---|---|
| 总协议对数 | 22 对 | **17 对** | -23% |
| 总编号数 | 44 个 | **34 个** | -23% |
| 1级表数量 | 11 张 | **10 张** | -9%（减 user_follows） |
| 领域事件数 | 14 个 | **11 个** | -21%（减 Follow/Unfollow/Stats） |
| 缓存 key 数 | 20+ 个 | **16 个** | -20% |

---

## 六、总结建议

### 必须修改（P0）

1. **统一 topics 表 FK 类型**：author_id 和 group_id 从 char(36) 改为 char(32)，与 users.id / groups.id 一致
2. **明确关注模型策略**：选择方案 A/B/C 之一，不要维持当前的"全量设计"

### 建议修改（P1）

3. **按选定方案调整 PRD 中的协议组表格**（减少 5 对关注协议）
4. **按选定方案调整 ADR 中的 1 级/2 级数据分类表**
5. **按选定方案调整 OpenAPI 中的 x-cairobot-protocols 条目**
6. **按选定方案调整 CODE-WIKI §26 中的协议组说明**
7. **groups.members_count 在 MVP 阶段直接使用事件驱动异步更新**，不做事务内同步

### 可以延后（P2）

8. group_invites 邀请表
9. content_moderation 审核工作流表
10. 推荐流、搜索索引等 2 级数据的高级策略

---

*评审完成。以上意见基于"简化关注模型、专注群组社交"的评审视角输出，供项目主控决策参考。*
