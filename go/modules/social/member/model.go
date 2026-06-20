package member

// 内部数据库模型，按 PRD + basemodel.md 设计
// 不引入 proto 包，proto ↔ model 映射在 svc 中完成
// 具体字段按 PRD 数据模型章节定义

// User 用户身份主表（对应 users 表，1级数据）
// 负责存储用户注册信息、登录凭证和基础资料
// 不负责社交关系（group_members）、统计（member_stats）等派生数据
type User struct {
	// ID 用户内部主键，全系统关联的标识符（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// UID 用户对外展示的唯一编号（9位数字）
	UID string `gorm:"uniqueIndex;size:20" json:"uid"`
	// Username 登录用户名，全局唯一
	Username string `gorm:"uniqueIndex;size:50" json:"username"`
	// Password 加密后的密码（salt + hash），不存储明文
	Password string `gorm:"size:255" json:"-"`
	// Email 邮箱地址，全局唯一
	Email string `gorm:"uniqueIndex;size:100" json:"email"`
	// Phone 手机号，全局唯一
	Phone string `gorm:"uniqueIndex;size:20" json:"phone"`
	// Nickname 用户昵称，可重复
	Nickname string `gorm:"size:50" json:"nickname"`
	// Avatar 头像 URL
	Avatar string `gorm:"size:500" json:"avatar"`
	// Status 用户状态：active/inactive/banned/deleted
	Status int8 `gorm:"default:1" json:"status"`
	// MembershipLevel 平台会员等级（≠群组付费会员）
	MembershipLevel string `gorm:"size:32;default:normal" json:"membership_level"`
	// LastLoginAt 最后登录时间戳（毫秒）
	LastLoginAt int64 `json:"last_login_at"`
	// LastLoginIP 最后登录 IP 地址
	LastLoginIP string `gorm:"size:45" json:"-"`
	// LoginCount 累计登录次数
	LoginCount int `gorm:"default:0" json:"login_count"`
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (User) TableName() string {
	return "users"
}

// MemberBlock 用户拉黑关系表（1级数据）
// 记录用户之间的拉黑关系，用于权限判断和内容过滤
// 拉黑是单向关系：A 拉黑 B 不等于 B 拉黑 A
type MemberBlock struct {
	// ID 主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// BlockerID 发起拉黑操作的用户 ID
	BlockerID string `gorm:"size:64;index:idx_blocker" json:"blocker_id"`
	// BlockedID 被拉黑的用户 ID
	BlockedID string `gorm:"size:64;index:idx_blocked" json:"blocked_id"`
	// Reason 拉黑原因（可选）
	Reason string `gorm:"size:500" json:"reason"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
	// CreatedAt 拉黑操作时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (MemberBlock) TableName() string {
	return "member_blocks"
}

// MemberStats 用户统计快照表（2级数据，事件驱动更新）
// 存储用户的派生统计数据，允许最终一致，可从 1级数据重建
// 不用于权限判断，仅用于展示
type MemberStats struct {
	// UserID 用户 ID，作为主键（一个用户一条统计记录）
	UserID string `gorm:"primaryKey;size:64;uniqueIndex" json:"user_id"`
	// TopicsCount 已发布帖子数（从 topics 表聚合）
	TopicsCount int `gorm:"default:0" json:"topics_count"`
	// RepliesCount 已发表评论数（从 topic_replies 表聚合）
	RepliesCount int `gorm:"default:0" json:"replies_count"`
	// LikesReceived 被点赞总数（从 topic_likes + reply_likes 聚合）
	LikesReceived int `gorm:"default:0" json:"likes_received"`
	// GroupsJoined 已加入群组数（从 group_members 表聚合）
	GroupsJoined int `gorm:"default:0" json:"groups_joined"`
	// UpdatedAt 统计最后更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (MemberStats) TableName() string {
	return "member_stats"
}
