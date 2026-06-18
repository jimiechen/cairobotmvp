// Package event 社交域领域事件基础设施
// 提供 DomainEvent 结构、Publisher/Subscriber 接口、MemoryBus/NoopPublisher 实现
// 不包含业务逻辑，不依赖 member/group/topic repository（避免循环依赖）
//
// 设计依据：
//   - PRD-social-app-mvp.md §7 数据等级规范：1级数据变更必须发布领域事件
//   - ADR-social-data-level-and-cache-strategy.md：事件驱动缓存与2级数据更新
//
// 架构选型：
//   - 生产环境：Redis Pub/Sub（跨实例传播）
//   - 单测/本地开发：MemoryBus（同步、可断言）
//   - 未配置时降级：NoopPublisher（不阻塞业务）
//   - 后续演进：Outbox + Redis Streams（可靠事件投递）
package event

// ===== 事件类型常量 =====
// 所有事件类型统一定义在此，svc 层禁止直接写事件名字符串

const (
	// === 成员域事件 ===

	// EventMemberRegistered 用户注册成功后发布
	// 触发点：UserRegister 写入 users 表成功后
	EventMemberRegistered = "MemberRegistered"

	// EventUserStatusChanged 用户状态变更后发布（管理员封禁/解封/注销等）
	// 触发点：UpdateMemberStatus / AdminUpdateMemberStatus 事务提交后
	EventUserStatusChanged = "UserStatusChanged"

	// === 群组域事件 ===

	// EventGroupCreated 群组创建成功后发布
	// 触发点：CreateGroup 写入 groups + owner member 成功后
	EventGroupCreated = "GroupCreated"

	// EventGroupJoined 用户加入群组成功后发布
	// 触发点：JoinGroup 写入 group_members 成功后
	EventGroupJoined = "GroupJoined"

	// EventGroupLeft 用户退出群组成功后发布
	// 触发点：LeaveGroup 更新 member.status=left 且状态确实发生变化后
	EventGroupLeft = "GroupLeft"

	// EventGroupMemberRemoved 成员被移除后发布
	// 触发点：RemoveMember 更新 member.status 后
	EventGroupMemberRemoved = "GroupMemberRemoved"

	// EventGroupMemberBanned 成员被封禁后发布
	// 触发点：BanMember 更新 member.status=banned 后
	EventGroupMemberBanned = "GroupMemberBanned"

	// EventGroupMemberMuted 成员被禁言后发布
	// 触发点：MuteMember 设置 muted_until 后
	EventGroupMemberMuted = "GroupMemberMuted"

	// EventGroupMemberRecovered 成员状态恢复后发布（解封/解禁/解除禁言）
	// 触发点：UnbanMember / UnmuteMember 恢复成员状态后
	EventGroupMemberRecovered = "GroupMemberRecovered"

	// EventGroupPlanCreated 付费方案创建成功后发布
	// 触发点：CreateGroupPlan 写入 group_plans 成功后
	EventGroupPlanCreated = "GroupPlanCreated"

	// EventGroupOrderPaid 订单支付确认后发布
	// 触发点：ConfirmGroupPayment / RenewMember 事务提交后
	EventGroupOrderPaid = "GroupOrderPaid"

	// EventGroupMemberActivated 成员权益激活后发布（续期/手动开通）
	// 触发点：RenewMember 续期或 ConfirmGroupPayment 权益 upsert 后
	EventGroupMemberActivated = "GroupMemberActivated"

	// === 主题域事件 ===

	// EventTopicCreated 帖子创建成功后发布
	// 触发点：CreateTopic 写入 topics 表成功后
	EventTopicCreated = "TopicCreated"

	// EventTopicDeleted 帖子删除成功后发布
	// 触发点：DeleteTopic 软删除状态写入成功后
	EventTopicDeleted = "TopicDeleted"

	// EventTopicCommentCreated 评论/回复创建成功后发布
	// 触发点：AddTopicReply / CreateComment 写入 topic_replies/topic_comments 成功后
	EventTopicCommentCreated = "TopicCommentCreated"

	// EventTopicReacted 帖子互动状态变更后发布（点赞/收藏/分享）
	// 触发点：LikeTopic / FavoriteTopic upsert 或取消后
	EventTopicReacted = "TopicReacted"

	// EventTopicApproved 帖子审核通过后发布
	// 触发点：AuditTopic action=approve 后
	EventTopicApproved = "TopicApproved"

	// EventTopicRejected 帖子审核拒绝后发布
	// 触发点：AuditTopic action=reject 后
	EventTopicRejected = "TopicRejected"

	// EventTopicBanned 帖子审核封禁后发布
	// 触发点：AuditTopic action=ban 后
	EventTopicBanned = "TopicBanned"
)

// ===== 聚合根类型常量 =====

const (
	AggregateMember  = "member"
	AggregateGroup   = "group"
	AggregateTopic   = "topic"
	AggregateOrder   = "group_order"
	AggregateComment = "topic_comment"
)

// ===== 事件版本 =====

// EventVersionCurrent 当前事件 schema 版本
// payload 字段变更时需升级版本号，消费者按 Type+Version 处理
const EventVersionCurrent = "1.0"

// ===== 动作类型常量（payload 中 action 字段使用）=====

const (
	ActionBan     = "ban"
	ActionRemove  = "remove"
	ActionMute    = "mute"
	ActionRecover = "recover"
)

// ===== 互动类型常量 =====

const (
	ReactionTypeLike     = "like"
	ReactionTypeFavorite = "favorite"
	ReactionTypeShare    = "share"
)
