package group

// 内部数据库模型，按 PRD + basemodel.md 设计
// 不引入 proto 包，proto ↔ model 映射在 svc 中完成
// 具体字段按 PRD 数据模型章节定义

// Group 群组/圈子主表（对应 groups 表，1级数据）
// 负责存储群组基础信息、可见性和加入方式配置
// 不负责成员关系（group_members）、统计（group_stats）等派生数据
type Group struct {
	// ID 群组主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// Name 群组名称
	Name string `gorm:"size:200;not null" json:"name"`
	// Slug 群组唯一标识 URL slug，全局唯一
	Slug string `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	// Description 群组描述
	Description string `gorm:"type:text" json:"description"`
	// Avatar 群组头像 URL
	Avatar string `gorm:"size:500" json:"avatar"`
	// CoverImage 群组封面图 URL
	CoverImage string `gorm:"size:500" json:"cover_image"`
	// Category 群组分类标签
	Category string `gorm:"size:100" json:"category"`
	// Tags 群组标签 JSON 数组
	Tags string `gorm:"type:json" json:"tags"`
	// OwnerID 圈主用户 ID（1级事实字段）
	OwnerID string `gorm:"size:64;not null;index" json:"owner_id"`
	// Status 群组状态：1-正常 2-已归档 3-已解散
	Status int8 `gorm:"default:1;index" json:"status"`
	// Visibility 可见性：1-公开 2-私密 3-邀请可见
	Visibility int8 `gorm:"default:1" json:"visibility"`
	// JoinMode 加入方式：1-自由加入 2-需要审核 3-仅邀请
	JoinMode int8 `gorm:"default:1" json:"join_mode"`
	// IsOfficial 是否官方群组
	IsOfficial bool `gorm:"default:false" json:"is_official"`
	// IsFeatured 是否推荐/精选
	IsFeatured bool `gorm:"default:false" json:"is_featured"`
	// MaxMembers 最大成员数上限
	MaxMembers int `gorm:"default:500" json:"max_members"`
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (Group) TableName() string {
	return "groups"
}

// GroupMember 群组成员关系表（对应 group_members 表，1级数据）
// 核心关系表，用于判断用户是否属于某个群组、在群组内的身份、是否被禁言、付费权益是否有效
// 权限判断（CanReadTopic/CanManageMember 等）必须查此表的 1级数据，禁止依赖缓存
type GroupMember struct {
	// ID 成员关系主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// GroupID 所属群组 ID
	GroupID string `gorm:"size:64;not null;index:idx_members_group_id" json:"group_id"`
	// UserID 用户 ID
	UserID string `gorm:"size:64;not null;index:idx_members_user_id" json:"user_id"`
	// Role 角色：1-群主 2-管理员 3-普通成员 4-待审核（对应 proto GroupMemberRole 枚举）
	Role int8 `gorm:"default:3;index:idx_members_role" json:"role"`
	// Status 成员状态：1-正常 2-已退出 3-已移除 4-已禁言 5-待审核
	Status int8 `gorm:"default:1;index:idx_members_status" json:"status"`
	// JoinReason 申请加入理由（审核模式时填写）
	JoinReason string `gorm:"size:500" json:"join_reason"`
	// InvitedBy 邀请人用户 ID（邀请制加入时填写）
	InvitedBy string `gorm:"size:64" json:"invited_by"`
	// JoinedAt 加入时间戳（毫秒）
	JoinedAt int64 `json:"joined_at"`
	// LastActivityAt 最后活跃时间戳（毫秒，用于活跃度计算）
	LastActivityAt int64 `json:"last_activity_at"`
	// Bio 成员个人简介（群内展示用）
	Bio string `gorm:"size:500" json:"bio"`
	// AnsweredQuestionsCount 已回答问题数（问答帖场景配额统计）
	AnsweredQuestionsCount int `gorm:"default:0" json:"answered_questions_count"`
	// RemainingQuota 剩余提问配额
	RemainingQuota int `gorm:"default:0" json:"remaining_quota"`
	// PaymentCycle 付费周期类型：NULL-免费 1-月付 2-季付 3-年付
	PaymentCycle *int8 `json:"payment_cycle"`
	// MembershipExpiresAt 会员权益到期时间戳（毫秒），免费成员为空
	MembershipExpiresAt int64 `json:"membership_expires_at"`
	// MutedUntil 禁言到期时间戳（毫秒），未禁言时为 0 或 NULL
	MutedUntil int64 `json:"muted_until"`
	// BanExpiresAt 封禁到期时间戳（毫秒），未封禁时为 0 或 NULL
	BanExpiresAt int64 `json:"ban_expires_at"`
	// ApprovedBy 审核通过操作人 ID
	ApprovedBy string `gorm:"size:64" json:"approved_by"`
	// BannedBy 封禁操作人 ID
	BannedBy string `gorm:"size:64" json:"banned_by"`
	// BanReason 封禁原因
	BanReason string `gorm:"size:500" json:"ban_reason"`
	// BannedAt 封禁时间戳（毫秒）
	BannedAt int64 `json:"banned_at"`
	// CreatedAt 记录创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 记录更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (GroupMember) TableName() string {
	return "group_members"
}

// GroupPayConfig 群组付费配置表（对应 group_pay_configs 表，1级数据）
// 定义群组的付费方案价格和试用策略
// 一个群组只有一条付费配置记录（uk_pay_config_group 唯一约束）
type GroupPayConfig struct {
	// ID 付费配置主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// GroupID 所属群组 ID，唯一约束
	GroupID string `gorm:"size:64;uniqueIndex:uk_pay_config_group;not null" json:"group_id"`
	// PriceMonthly 月付价格（单位：元）
	PriceMonthly float64 `gorm:"type:decimal(10,2)" json:"price_monthly"`
	// PriceQuarterly 季付价格（单位：元）
	PriceQuarterly float64 `gorm:"type:decimal(10,2)" json:"price_quarterly"`
	// PriceYearly 年付价格（单位：元）
	PriceYearly float64 `gorm:"type:decimal(10,2)" json:"price_yearly"`
	// Currency 货币单位
	Currency string `gorm:"size:10;default:CNY" json:"currency"`
	// TrialDays 试用天数
	TrialDays int `gorm:"default:0" json:"trial_days"`
	// IsEnabled 是否启用付费功能
	IsEnabled bool `gorm:"default:false" json:"is_enabled"`
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (GroupPayConfig) TableName() string {
	return "group_pay_configs"
}

// GroupStats 群组统计快照表（对应 group_stats 表，2级数据，事件驱动更新）
// 存储群组的派生统计数据，允许最终一致，可从 group_members + topics 等 1级表重建
// 不用于权限判断，仅用于展示和排序
type GroupStats struct {
	// GroupID 群组 ID，作为主键（一个群组一条统计记录）
	GroupID string `gorm:"primaryKey;size:64;uniqueIndex" json:"group_id"`
	// MembersCount 总成员数（从 group_members WHERE status=active 聚合）
	MembersCount int `gorm:"default:0" json:"members_count"`
	// ActiveMembersCount 活跃成员数（有近期活动的成员）
	ActiveMembersCount int `gorm:"default:0" json:"active_members_count"`
	// PaidMembersCount 付费成员数（membership_expires_at > now 的成员）
	PaidMembersCount int `gorm:"default:0" json:"paid_members_count"`
	// TopicsCount 帖子总数（从 topics WHERE status=published AND group_id=? 聚合）
	TopicsCount int `gorm:"default:0" json:"topics_count"`
	// UpdatedAt 统计最后更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (GroupStats) TableName() string {
	return "group_stats"
}
