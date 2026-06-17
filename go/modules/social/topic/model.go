package topic

// 内部数据库模型，按 PRD + basemodel.md 设计
// 不引入 proto 包，proto ↔ model 映射在 svc 中完成
// 具体字段按 PRD 数据模型章节定义

// Topic 主题/帖子主表（对应 topics 表，1级数据）
// 负责存储帖子内容、可见性配置和生命周期状态
// 不负责评论（topic_replies）、互动（topic_likes 等）、统计（topic_stats）等关联数据
type Topic struct {
	// ID 帖子主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// Title 帖子标题
	Title string `gorm:"size:500;not null" json:"title"`
	// Content 帖子正文内容（支持纯文本/Markdown/富文本）
	Content string `gorm:"type:text" json:"content"`
	// AuthorID 作者用户 ID
	AuthorID string `gorm:"size:64;not null;index:idx_topics_author_id" json:"author_id"`
	// AuthorName 作者昵称（冗余字段，避免联表查询 users）
	AuthorName string `gorm:"size:200" json:"author_name"`
	// GroupID 所属群组 ID（NULL 表示全局帖子）
	GroupID string `gorm:"size:64;index:idx_topics_group_id" json:"group_id"`
	// TopicType 主题类型：1-普通帖 2-问答帖 3-投票帖
	TopicType int8 `gorm:"default:1;index:idx_topics_topic_type" json:"topic_type"`
	// ContentFormat 内容格式：1-纯文本 2-Markdown 3-富文本
	ContentFormat int8 `gorm:"default:1" json:"content_format"`
	// Visibility 可见性：1-公开 2-仅成员（权限判断核心字段）
	Visibility int8 `gorm:"default:1" json:"visibility"`
	// Summary 摘要（无权限时返回此字段代替完整内容）
	Summary string `gorm:"size:500" json:"summary"`
	// AuthorAvatar 作者头像 URL（冗余字段）
	AuthorAvatar string `gorm:"size:500" json:"author_avatar"`
	// QuestionText 问答帖的问题文本
	QuestionText string `gorm:"size:300" json:"question_text"`
	// QAPrivate 问答是否私密（仅指定用户可回答）
	QAPrivate bool `gorm:"default:false" json:"qa_private"`
	// AnsweredAt 被回答时间戳（毫秒）
	AnsweredAt int64 `json:"answered_at"`
	// CoverImage 封面图 URL
	CoverImage string `gorm:"size:500" json:"cover_image"`
	// MemberID 关联的成员记录 ID（冗余引用）
	MemberID string `gorm:"size:64" json:"member_id"`
	// DraftID 关联草稿 ID
	DraftID string `gorm:"size:64" json:"draft_id"`
	// NavTypes 导航类型 JSON 数组
	NavTypes string `gorm:"type:json" json:"nav_types"`
	// QATargetUserID 问答指定目标用户 ID
	QATargetUserID string `gorm:"size:64" json:"qa_target_user_id"`
	// HasMedia 是否包含媒体附件
	HasMedia bool `gorm:"default:false" json:"has_media"`
	// HasDocs 是否包含文档附件
	HasDocs bool `gorm:"default:false" json:"has_docs"`
	// Status 生命周期状态：1-草稿 2-已发布 3-已关闭 4-已删除
	Status int8 `gorm:"default:2;index:idx_topics_status" json:"status"`
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (Topic) TableName() string {
	return "topics"
}

// TopicReply 评论/回复表（对应 topic_replies 表，1级数据）
// 支持嵌套回复（楼中楼），通过 parent_reply_id 构建树形结构
// 评论是用户生成内容，需要持久化、审核和审计
type TopicReply struct {
	// ID 评论主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// TopicID 所属主题/帖子 ID
	TopicID string `gorm:"size:64;not null;index:idx_replies_topic_id" json:"topic_id"`
	// Content 评论内容
	Content string `gorm:"type:text;not null" json:"content"`
	// AuthorID 评论者用户 ID
	AuthorID string `gorm:"size:64;not null;index:idx_replies_author_id" json:"author_id"`
	// AuthorName 评论者昵称（冗余字段，避免联表查询）
	AuthorName string `gorm:"size:200" json:"author_name"`
	// ParentReplyID 父评论 ID（NULL 表示顶层评论，非 NULL 表示嵌套回复）
	ParentReplyID string `gorm:"size:64;index:idx_replies_parent_reply_id" json:"parent_reply_id"`
	// Status 评论状态：1-正常 2-已删除 3-已屏蔽
	Status int8 `gorm:"default:1" json:"status"`
	// LikeCount 点赞数（冗余计数，以 reply_likes 实际数据为准）
	LikeCount int `gorm:"default:0" json:"like_count"`
	// CreatedAt 创建时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt 更新时间戳（毫秒）
	UpdatedAt int64 `json:"updated_at"`
	// RepliesCount 子回复数量（用于展示折叠状态）
	RepliesCount int `gorm:"default:0" json:"replies_count"`
	// IsPinned 是否置顶（圈主/管理员操作）
	IsPinned bool `gorm:"default:false" json:"is_pinned"`
	// IsLiked 当前用户是否已点赞（查询时动态填充，不持久化）
	IsLiked bool `gorm:"default:false" json:"is_liked"`
	// Level 回复层级：1-顶层 2-二级及以下
	Level int8 `gorm:"default:1" json:"level"`
	// ReplyToUserID 被回复的用户 ID
	ReplyToUserID string `gorm:"size:64" json:"reply_to_user_id"`
	// ReplyToUserName 被回复的用户昵称
	ReplyToUserName string `gorm:"size:200" json:"reply_to_user_name"`
}

// TableName 返回 GORM 对应的数据库表名
func (TopicReply) TableName() string {
	return "topic_replies"
}

// TopicLike 帖子点赞表（对应 topic_likes 表，1级事实数据）
// 记录用户对帖子的点赞关系；同一 user→topic 只能有一条记录
// 点赞数聚合属于 2级统计数据，存在 topic_stats 中
type TopicLike struct {
	// ID 点赞记录主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// TopicID 被点赞的主题/帖子 ID
	TopicID string `gorm:"size:64;not null;uniqueIndex:uk_topic_like" json:"topic_id"`
	// UserID 点赞用户 ID
	UserID string `gorm:"size:64;not null;uniqueIndex:uk_topic_like;index:idx_likes_user_id" json:"user_id"`
	// CreatedAt 点赞时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (TopicLike) TableName() string {
	return "topic_likes"
}

// TopicFavorite 帖子收藏表（对应 topic_favorites 表，1级事实数据）
// 记录用户对帖子的收藏关系；同一 user→topic 只能有一条记录
// 收藏数聚合属于 2级统计数据
type TopicFavorite struct {
	// ID 收藏记录主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// TopicID 被收藏的主题/帖子 ID
	TopicID string `gorm:"size:64;not null;uniqueIndex:uk_topic_favorite" json:"topic_id"`
	// UserID 收藏用户 ID
	UserID string `gorm:"size:64;not null;uniqueIndex:uk_topic_favorite;index:idx_favorites_user_id" json:"user_id"`
	// CreatedAt 收藏时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (TopicFavorite) TableName() string {
	return "topic_favorites"
}

// TopicRead 阅读记录表（对应 topic_reads 表，2级行为数据）
// 记录用户阅读行为，用于已读状态、阅读历史和圈主数据分析
// 允许最终一致，允许异步写入（不阻塞读帖响应）
type TopicRead struct {
	// ID 阅读记录主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// TopicID 被阅读的主题/帖子 ID
	TopicID string `gorm:"size:64;not null;index:idx_reads_read_at;uniqueIndex:uk_topic_read,priority:1" json:"topic_id"`
	// UserID 阅读用户 ID
	UserID string `gorm:"size:64;not null;uniqueIndex:uk_topic_read,priority:2;index:idx_reads_user_id" json:"user_id"`
	// ReadAt 最近阅读时间戳（毫秒）
	ReadAt int64 `json:"read_at"`
	// ReadDuration 累计阅读时长（秒），多次阅读累加
	ReadDuration int `gorm:"default:0" json:"read_duration"`
}

// TableName 返回 GORM 对应的数据库表名
func (TopicRead) TableName() string {
	return "topic_reads"
}

// ReplyLike 评论点赞表（对应 reply_likes 表，1级事实数据）
// 记录用户对评论的点赞关系；同一 user→reply 只能有一条记录
type ReplyLike struct {
	// ID 评论点赞记录主键（ULID 风格字符串）
	ID string `gorm:"primaryKey;size:64" json:"id"`
	// ReplyID 被点赞的评论 ID
	ReplyID string `gorm:"size:64;not null;uniqueIndex:uk_reply_like" json:"reply_id"`
	// UserID 点赞用户 ID
	UserID string `gorm:"size:64;not null;uniqueIndex:uk_reply_like;index:idx_reply_likes_user_id" json:"user_id"`
	// CreatedAt 点赞时间戳（毫秒）
	CreatedAt int64 `json:"created_at"`
}

// TableName 返回 GORM 对应的数据库表名
func (ReplyLike) TableName() string {
	return "reply_likes"
}
