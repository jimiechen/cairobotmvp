package topic

// TopicStatus 帖子状态（对应 topics.status 字段）
// 值与 proto TopicStatus 枚举一一对应
const (
	TopicStatusActive   int8 = 1 // 活跃       → proto TOPIC_STATUS_ACTIVE
	TopicStatusInactive int8 = 2 // 非活跃     → proto TOPIC_STATUS_INACTIVE
	TopicStatusDeleted  int8 = 3 // 已删除     → proto TOPIC_STATUS_DELETED
	TopicStatusLocked   int8 = 4 // 锁定       → proto TOPIC_STATUS_LOCKED
	TopicStatusPinned   int8 = 5 // 置顶       → proto TOPIC_STATUS_PINNED
)

// TopicType 帖子类型（对应 topics.topic_type 字段）
// 值与 proto TopicType 枚举一一对应（注意：4/5 为预留值，6=文章）
const (
	TopicTypeNormal       int8 = 1 // 普通帖子   → proto TOPIC_TYPE_NORMAL
	TopicTypeAnnouncement int8 = 2 // 公告       → proto TOPIC_TYPE_ANNOUNCEMENT
	TopicTypeQuestion     int8 = 3 // 问题       → proto TOPIC_TYPE_QUESTION
	TopicTypeArticle      int8 = 6 // 文章       → proto TOPIC_TYPE_ARTICLE
)

// ReplyStatus 回复状态（对应 topic_replies.status 字段）
// 值与 proto ReplyStatus 枚举一一对应
const (
	ReplyStatusActive  int8 = 1 // 活跃       → proto REPLY_STATUS_ACTIVE
	ReplyStatusDeleted int8 = 2 // 已删除     → proto REPLY_STATUS_DELETED
	ReplyStatusHidden  int8 = 3 // 隐藏       → proto REPLY_STATUS_HIDDEN
)

// TopicVisibility 帖子可见性等级（对应 topics.visibility 字段）
// 值与 proto Visibility 枚举对应（详见 topic_base.proto Visibility 枚举定义）
const (
	TopicVisibilityPublic      int8 = 1 // 公开，所有人可读
	TopicVisibilityGroupMember int8 = 2 // 仅圈子成员可读
	TopicVisibilityPaidMember  int8 = 3 // 仅付费成员可读（permission.CanReadTopic PAID 分支）
	TopicVisibilityOwnerOnly   int8 = 4 // 仅圈主可读
)
