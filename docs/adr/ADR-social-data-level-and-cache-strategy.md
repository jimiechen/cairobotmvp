# ADR-social-data-level-and-cache-strategy.md

> **状态**: 已采纳
> **日期**: 2026-06-15
> **决策者**: 项目主控
> **相关 PRD**: [PRD-social-app-mvp](../prd/PRD-social-app-mvp.md)
> **相关 API**: [social-openapi.yaml](../api/social-openapi.yaml)

---

## 1. 背景

CaiRobot MVP 社交域涉及用户、群组、帖子、评论、互动、支付等多种数据类型。如果将所有数据等同对待——都使用相同的缓存策略、一致性要求和更新方式——会导致以下问题：

1. **缓存失效复杂度爆炸**：成员数、帖子数等频繁变化的统计数据与用户资料、群组名称等稳定数据使用相同 TTL，要么稳定数据缓存命中率低，要么统计数据不准确
2. **事务范围过大**：如果每次发帖都要同步更新"群组帖子数+作者发帖数+标签帖子数"，事务持有时间过长，锁竞争激烈
3. **权限判断风险**：如果权限判断依赖了可能延迟或不准确的统计缓存，会出现安全漏洞（如已过期付费用户仍能访问付费内容）
4. **审计边界不清**：哪些变更需要记录审计日志、哪些不需要，缺乏统一标准

因此需要建立一套**数据等级（Data Level）治理体系**，明确区分"业务事实数据"和"派生统计数据"，并为不同等级的数据设计不同的存储、一致性和缓存策略。

## 2. 决策

**采纳"1级/2级数据分级 + 协议组划分 + 事件驱动缓存"的综合策略。**

### 核心规则

```text
1级数据 = 业务事实，协议主动写入才变化，强一致，必须落 MySQL
2级数据 = 派生统计，由事件/任务被动触发变化，最终一致，可存 Redis/统计表
```

### 三层保障

```
数据等级划分 → 缓存策略绑定 → 权限安全约束
     ↓                ↓               ↓
  事实vs统计      Cache Aside vs   1级数据才能
  的明确边界    事件驱动更新       用于权限判断
```

## 3. 方案详述

### 3.1 数据等级定义

#### 1级数据：业务主事实（Strongly Consistent Facts）

**定义**：只有通过明确的业务协议（Protobuf Request）主动操作后才发生变化的数据。代表业务领域的"真理来源"（Source of Truth）。

**特征**：
- 属于业务事实，不是计算结果
- 变更必须有明确的业务操作触发（用户点击、管理员操作、系统协议调用）
- 必须落盘到 MySQL / InnoDB 主库
- 使用数据库事务保证强一致性（ACID）
- 更新后必须发布领域事件（Domain Event）
- 必须有审计追踪能力（对高权限操作）

**社交域 1级数据清单**：

| 表 | 关键字段 | 示例操作 |
|---|---------|---------|
| users | 全部字段 | MemberRegister, UpdateMemberProfile, UpdateMemberStatus |
| groups | 除 members_count/topics_count | CreateGroup, UpdateGroup, AuditGroup |
| topics | 全部字段 | CreateTopic, UpdateTopic, DeleteTopic, AuditTopic |
| group_members | 成员关系和角色 | JoinGroup, LeaveGroup, UpdateGroupMemberStatus, ConfirmGroupPayment |
| group_plans | 方案定义 | CreateGroupPlan, UpdateGroupPlan |
| group_orders | 订单状态 | CreateGroupOrder, ConfirmGroupPayment |
| topic_comments | 评论内容 | CreateComment, DeleteComment, AdminRemoveComment |
| topic_reactions | 互动关系 | ReactTopic, CancelReactTopic |
| group_admin_actions | 操作日志 | 所有高权限操作自动追加 |

**不变性约束**：1级数据一旦写入，在未被新的业务操作覆盖前，其值视为正确。不允许后台定时任务"修正"1级数据。

#### 2级数据：派生状态与统计（Eventually Consistent Derivatives）

**定义**：由 1级数据变更、统计任务、用户行为（阅读、点赞）、高权限操作、系统事件被动触发变化的数据。属于聚合、计数、索引、缓存视图等派生结果。

**特征**：
- 可以从 1级数据完整重新计算出来（可重建性）
- 允许最终一致（容忍 1-5 秒延迟）
- 可存储在 Redis、MySQL 冗余字段、搜索引擎、统计表中
- 通过领域事件驱动更新，或通过定时任务校准
- **绝对不能作为权限判断的依据**

**社交域 2级数据清单**：

| 数据 | 来源 1级表 | 存储建议 | 允许延迟 |
|-----|-----------|---------|---------|
| 用户发帖数 | topics | Redis Counter + member_stats.topics_count | 秒级 |
| 群组成员数 | group_members | Redis Counter + group_stats.members_count + groups.members_count | 秒级 |
| 群组帖子数 | topics | Redis Counter + group_stats.topics_count + groups.topics_count | 秒级 |
| 帖子阅读数 | topic_read_records | Redis HyperLogLog/Counter + topic_stats.read_count | 分钟级 |
| 帖子点赞数 | topic_reactions | Redis Counter + topic_stats.likes_count | 秒级 |
| 帖子评论数 | topic_comments | Redis Counter + topic_stats.comments_count | 秒级 |
| 圈主看板 | 多表聚合 | Redis Snapshot | 分钟级 |
| 群组帖子列表（分页） | topics | Redis Sorted Set + DB 分页 | 秒级 |
| 热门帖子排行 | 多维行为 | Redis Sorted Set（ZSET 加权评分） | 分钟级 |
| 推荐群组 | 活跃度数据 | Redis/Search Index | 小时级 |

**可重建性验证**：每个 2级数据必须能写出对应的 SQL 重建语句。例如：

```sql
-- 重建某群组的活跃成员数
SELECT COUNT(*) FROM group_members WHERE group_id = ? AND status = 1;

-- 重建某帖子的点赞数
SELECT COUNT(*) FROM topic_reactions WHERE topic_id = ? AND reaction_type = 'like' AND status = 1;
```

### 3.2 缓存策略设计

#### 1级数据缓存：Cache Aside + 主动失效

**适用对象**：用户资料、群组详情、帖子详情、成员关系、付费方案

**读写流程**：

```
读取（Cache Aside Read）:
  Client → 检查 Redis → HIT → 返回缓存
                      → MISS → 查 MySQL → 写入 Redis（带 TTL）→ 返回

写入（Cache Aside Write - 主动失效）:
  业务操作 → 更新 MySQL（事务提交成功）
          → DELETE 对应 Redis Key（不是更新缓存值！）
          → 下次读取时 Cache Miss → 回源重建
```

**TTL 设计原则**：

| 数据敏感度 | TTL 范围 | 说明 |
|-----------|---------|------|
| 高（安全相关） | 5-10 min | 成员关系、禁言状态 |
| 中（业务相关） | 10-20 min | 用户资料、群组详情、帖子详情 |
| 低（配置相关） | 20-30 min | 付费方案列表 |

**为什么是 Delete 而非 Update？**
- 避免并发写导致缓存值不一致（Lost Update 问题）
- 删除后下次回源保证拿到最新值
- 配合短 TTL 兜底，即使删除失败也能自然过期

**关键安全约束**：对于安全敏感的 1级数据缓存（如成员关系 `group:member:{groupId}:{userId}`），TTL 必须 ≤ 10 分钟。因为用户被禁言/移除后，旧缓存可能导致其仍能短暂执行操作。

#### 2级数据缓存：事件驱动更新 + TTL 兜底 + 定时校准

**适用对象**：各类统计、列表、排行榜、推荐数据

**更新流程**：

```
正常路径（事件驱动）:
  1级数据变更 → 发布领域事件
             → Event Consumer 监听
             → 增量更新 Redis（INCR/DECR/ZADD/ZREM 等）
             → 异步刷新统计快照表

兜底路径（TTL 过期）:
  缓存 TTL 到期 → 下次访问时 Cache Miss
              → 触发 rebuild（从 DB 或其他缓存聚合计算）

校正路径（定时任务）:
  定时 Job（如每小时）→ 全量对比 2级缓存 vs 1级数据计算结果
                     → 不一致则修正（修复 drift）
```

**TTL 设计原则**：

| 数据类型 | TTL 范围 | 说明 |
|---------|---------|------|
| 实时计数器 | 5-30 min | 点赞数、评论数、阅读数 |
| 分页列表 | 1-10 min | 帖子列表 |
| 排行榜 | 1-5 min | 热门帖子 |
| 推荐数据 | 5-30 min | 推荐群组 |
| 看板/聚合 | 1-10 min | 圈主看板 |

### 3.3 领域事件设计

#### 事件驱动的数据流

```
1级数据写入（事务内）
       ↓
事务提交成功
       ↓
发布领域Event（同步或同一 DB 事务内的 outbox）
       ↓
┌──────────────┬───────────────┬──────────────┐
│  缓存消费者   │  统计消费者    │  通知消费者    │
│  (Redis)     │  (Stats Table) │ (Notification)│
│              │               │              │
│ DEL key      │ UPDATE count  │ INSERT msg   │
│ INCR counter │ UPSERT snapshot│ PUSH ws      │
└──────────────┴───────────────┴──────────────┘
```

#### 社交域核心事件清单

| 事件名 | 触发条件 | 消费者动作 |
|-------|---------|-----------|
| MemberRegistered | 用户注册成功 | 初始化 member_stats |
| GroupCreated | 创建群组成功 | 初始化 group_stats |
| GroupJoined | 加入群组成功 | group.members_count+1; 删 group:member 缓存 |
| GroupLeft | 退出群组成功 | group.members_count-1; 删 group:member 缓存 |
| GroupMemberRemoved | 成员被移除 | 同 GroupJoined 逆向; 通知被移除用户 |
| GroupMemberBanned | 成员被禁言 | 删 group:member 缓存（立即生效）; 通知 |
| GroupPlanCreated | 创建付费方案 | 刷新 group:plans 缓存 |
| GroupOrderPaid | 支付确认成功 | group.paid_members+1; 开通权益; 通知 |
| TopicCreated | 发帖成功 | group.topics_count+1; author.topics+1 |
| TopicDeleted | 删帖成功 | group.topics_count-1; 清理关联 2级数据 |
| TopicCommentCreated | 评论成功 | topic.comments_count+1 |
| TopicReacted | 点赞/收藏成功 | topic.likes/favorites+1 |
| TopicAudited | 审核完成 | 删 topic:detail 缓存; 通知作者 |

### 3.4 高权限操作的增强处理流水线

高权限操作指影响其他用户权益的操作，除满足 1级数据的常规要求外，还需额外执行审计和通知步骤。

**标准六步流水线**：

```
Step 1: 权限预检（PermissionService.CanXxx）
        ↓ 通过
Step 2: 写入审计日志（group_admin_actions 或系统审计表）
        ↓
Step 3: 更新 1级数据（MySQL 事务内，含业务字段 + 审计字段）
        ↓ 事务提交成功
Step 4: 发布领域事件（EventPublisher）
        ↓
Step 5: 主动失效相关缓存（CacheInvalidator，批量 DEL）
        ↓
Step 6: 通知受影响用户（站内信 / WebSocket / 外部推送）
        ↓
返回成功响应
```

**任何一步失败的处理策略**：

| 失败步骤 | 处理策略 | 原因 |
|---------|---------|------|
| Step 1 权限不通过 | 返回错误，终止 | 安全底线 |
| Step 2 审计写失败 | 记录错误日志，继续（审计异步化兜底） | 不阻塞主流程 |
| Step 3 事务失败 | 回滚，返回错误 | 数据一致性优先 |
| Step 4 事件发布失败 | 重试 3 次 + 落地本地 outbox 表 | 最终一致性保证 |
| Step 5 缓存失效失败 | 记录警告，依赖 TTL 自然过期 | 缓存不一致可接受（短期） |
| Step 6 通知失败 | 重试 + 落地通知队列 | 通知可以延迟 |

## 4. 后果

### 4.1 正面影响

1. **缓存管理清晰化**：不同数据使用不同策略，不再"一把梭"
2. **安全边界明确化**：权限只基于 1级数据，消除了 2级缓存导致的安全漏洞风险
3. **性能优化方向明确**：1级数据走 Cache Aside（读多写少），2级数据走事件驱动（写多读频变）
4. **故障恢复能力**：2级数据全部可重建，Redis 故障不影响 1级数据正确性
5. **审计合规性**：高权限操作强制四件套（权限+审计+事件+通知），满足运营审计要求

### 4.2 负面影响

1. **架构复杂度增加**：需要维护事件发布/消费机制、缓存失效策略、统计重建任务
2. **开发工作量增大**：每个 1级数据写入操作都需要考虑事件发布和缓存失效
3. **调试难度增加**：2级数据不一致时需要排查事件消费是否正常、缓存失效是否到位
4. **团队学习成本**：开发者需要理解数据等级划分并正确应用

### 4.3 缓解措施

- 事件框架封装为通用 SDK（EventPublisher / EventSubscriber），降低接入成本
- 缓存失效逻辑封装为 CacheInvalidator，按数据等级自动选择失效策略
- 提供统计重建 CLI 命令（`make social-rebuild-stats`），支持全量/增量重建
- 在 CODE-WIKI 中补充完整的数据等级开发指南和 Checklist

## 5. 替代方案

### 方案 A：全量强一致（未采纳）

所有数据（包括统计计数）都在事务内同步更新。

**优点**：简单，数据始终一致。
**缺点**：事务范围大、锁竞争激烈、性能差；统计查询仍然需要 COUNT 聚合；无法利用 Redis 缓存优势。
**结论**：不适合社交场景的高并发读写模式。

### 方案 B：纯缓存驱动（未采纳）

所有数据都先写 Redis，再异步落盘 DB（Write-Behind 模式）。

**优点**：写入性能极高。
**缺点**：Redis 故障时数据丢失风险；不符合 1级数据"必须落 MySQL 主库"的审计要求；权限判断依赖缓存有安全隐患。
**结论**：可用于 2级数据的部分场景，但不能作为整体策略。

### 方案 C：CQRS（命令查询职责分离）（备选但暂不采纳）

读写分离，写模型（1级）和读模型（2级）完全分离，通过事件同步。

**优点**：最灵活的读写优化空间；天然支持多维度读模型。
**缺点**：MVP 阶段过重；需要维护双写一致性；团队需要 CQRS 经验。
**结论**：作为远期演进方向保留。当前采用本 ADR 的简化版本（本质上是轻量 CQRS），后续可在 2级数据层逐步引入 CQRS 读模型。

## 6. 与现有架构的关系

### 与 MessagePacket / Protobuf 的关系

数据等级划分不改变现有的 MessagePacket + Protobuf 通信方式。但在 Protobuf Response 定义中，可以标注数据等级：

```protobuf
message GetMemberProfileResponse {
  // 1级数据：来自 users 表
  UserDTO user = 1;           // Cache Aside, TTL 10-30min

  // 2级数据：来自聚合计算
  MemberStatsDTO stats = 2;   // Event-driven, TTL 30-120min
}
```

### 与 routes.yaml 的关系

数据等级不影响路由配置。路由仅关心 maxType/minType → Tars 目标的映射。但可以在 routes.yaml 中新增可选字段 `data_level` 作为文档标注：

```yaml
- request_max: 1000
  request_min: 1010
  data_level: "1-write"       # 1级写入操作
  cache_invalidate:
    - "group:member:{groupId}:{userId}"
    - "group:stats:{groupId}"
```

### 与 TarsGo 服务的关系

社交域 Tars 服务内部按以下分包实现：

```
go/modules/social/
  ├── member/          # 成员协议组 handler/service/usecase/repository
  │   └── event.go    # MemberRegistered/GroupCreated 等事件定义
  ├── group/           # 群组协议组
  │   └── event.go    # GroupCreated/GroupJoined 等事件定义
  ├── topic/           # 主题协议组
  │   └── event.go    # TopicCreated/TopicReacted 等事件定义
  ├── permission/      # 统一权限服务（只查 1级数据）
  ├── cache/           # 缓存 Key 定义 + 失效策略
  │   ├── keys.go     # 缓存 Key 命名常量
  │   ├── invalidator.go  # 按 1级/2级策略执行失效
  │   └── policy.go   # TTL 策略配置
  └── event/
      ├── publisher.go    # 事件发布
      ├── subscriber.go   # 事件消费
      └── events.go       # 全部事件枚举
```

## 7. 实施计划

### MVP 首期（当前 PRD 范围）

1. 实现 1级数据的 Cache Aside 缓存层（keys.go + invalidator.go）
2. 实现核心领域事件的发布（publisher.go）
3. 实现统一权限服务（permission/service.go），强制只查 1级数据
4. 实现高权限操作的审计日志写入（复用 group_admin_actions 表）
5. 2级数据首期使用"DB 实时查询 + groups/topics 表冗余字段"方案，暂不做独立的事件驱动更新

### 第二期（P1 迭代）

6. 引入 Redis 缓存层，实现 2级数据的事件驱动增量更新
7. 实现统计快照表的定时校准 Job
8. 引入 PubSub 事件总线（替代首期的内存事件）

### 第三期（P2 迭代）

9. 评估引入 CQRS 读模型的必要性
10. 实现完整的缓存监控和 drift 告警
11. 实现统计数据的全量重建工具链

---

*本 ADR 由 [PRD-social-app-mvp](../prd/PRD-social-app-mvp.md) §7 数据等级规范 和 §11 缓存策略章节展开而来。*
