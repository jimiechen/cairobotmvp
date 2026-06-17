package topic

import (
	"context"
)

// Repository 主题域数据库操作接口定义
// 覆盖主题协议组（maxType=3000）全部 MVP 白名单协议的数据访问需求
// GORM 实现放在 repository_gorm.go
type Repository interface {
	// ========== Topic 帖子 CRUD（对应协议 minType 3001-3010）==========

	// CreateTopic 创建帖子记录（minType=3001 CreateTopic）
	// 前置条件：用户是群组成员且未被禁言（CanPublishTopic）
	CreateTopic(ctx context.Context, topic *Topic) error

	// GetTopicByID 根据 ID 查询帖子详情（minType=3005 GetTopicList 单条场景 + GetTopicDetail）
	GetTopicByID(ctx context.Context, id string) (*Topic, error)

	// UpdateTopic 更新帖子内容或状态（编辑、审核、上下架等）
	UpdateTopic(ctx context.Context, topic *Topic) error

	// DeleteTopic 删除帖子（minType=3009 DeleteTopic，软删除设置 status=deleted）
	DeleteTopic(ctx context.Context, id string) error

	// ListTopics 分页查询帖子列表（minType=3005 GetTopicList）
	// 支持按 group_id、status、topic_type、author_id 等条件筛选和排序
	// 返回帖子列表、总数、错误
	ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*Topic, int64, error)

	// ListTopicsByGroupID 分页查询某群组的帖子列表（群组内帖子浏览）
	ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*Topic, int64, error)

	// ListTopicsByAuthorID 分页查询某用户的帖子列表（个人主页）
	ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*Topic, int64, error)

	// CountTopicsByGroupID 统计群组内已发布帖子数（用于 2级 stats 更新）
	CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error)

	// ========== TopicReply 评论操作（对应协议 minType 3043-3066）==========

	// CreateReply 创建评论/回复（minType=3043 AddTopicReply）
	// 支持顶层评论（parent_reply_id 为空）和嵌套回复（楼中楼）
	CreateReply(ctx context.Context, reply *TopicReply) error

	// GetReplyByID 根据评论 ID 查询单条评论
	GetReplyByID(ctx context.Context, id string) (*TopicReply, error)

	// DeleteReply 删除评论（minType=3055 DeleteTopicReply，软删除设置 status=deleted）
	DeleteReply(ctx context.Context, id string) error

	// ListReplies 分页查询帖子的评论列表（minType=3065 GetReplyList）
	// 返回评论列表、总数、错误；支持按 parent_reply_id 过滤获取子回复
	ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*TopicReply, int64, error)

	// ListRepliesByAuthorID 分页查询某用户发表的评论列表
	ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*TopicReply, int64, error)

	// UpdateReply 更新评论字段（状态变更、置顶等）
	UpdateReply(ctx context.Context, reply *TopicReply) error

	// CountRepliesByTopicID 统计帖子有效评论数（用于 2级 stats 更新）
	CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error)

	// ========== TopicLike 点赞操作（对应协议 minType 3061-3062）==========

	// CreateLike 创建帖子点赞记录（minType=3061 LikeTopic）
	// 幂等：同一 user→topic 重复点赞返回已存在错误
	CreateLike(ctx context.Context, like *TopicLike) error

	// DeleteLike 取消帖子点赞（minType=3061 LikeTopic 的取消操作）
	DeleteLike(ctx context.Context, topicID, userID string) error

	// IsTopicLiked 检查用户是否已点赞某帖子（快速判断，用于前端按钮状态）
	IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error)

	// CountLikesByTopicID 统计帖子点赞数（用于 2级 stats 更新）
	CountLikesByTopicID(ctx context.Context, topicID string) (int64, error)

	// ========== TopicFavorite 收藏操作（对应协议 minType 3063-3064）==========

	// CreateFavorite 创建收藏记录（minType=3063 FavoriteTopic）
	CreateFavorite(ctx context.Context, favorite *TopicFavorite) error

	// DeleteFavorite 取消收藏（minType=3063 FavoriteTopic 的取消操作）
	DeleteFavorite(ctx context.Context, topicID, userID string) error

	// IsTopicFavorited 检查用户是否已收藏某帖子（快速判断）
	IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error)

	// ListFavoritesByUserID 分页查询用户的收藏列表（我的收藏页面）
	ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*TopicFavorite, int64, error)

	// CountFavoritesByTopicID 统计帖子收藏数（用于 2级 stats 更新）
	CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error)

	// ========== ReplyLike 评论点赞操作（对应协议 minType 3077-3078）==========

	// CreateReplyLike 创建评论点赞记录（minType=3077 LikeReply）
	CreateReplyLike(ctx context.Context, replyLike *ReplyLike) error

	// DeleteReplyLike 取消评论点赞
	DeleteReplyLike(ctx context.Context, replyID, userID string) error

	// IsReplyLiked 检查用户是否已点赞某评论
	IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error)

	// CountLikesByReplyID 统计评论点赞数（用于更新 reply.like_count 冗余字段）
	CountLikesByReplyID(ctx context.Context, replyID string) (int64, error)

	// ========== TopicRead 阅读记录操作（2级数据，异步写入）==========

	// GetReadRecord 获取用户对某帖子的阅读记录（判断是否已读）
	GetReadRecord(ctx context.Context, topicID, userID string) (*TopicRead, error)

	// UpsertReadRecord 创建或更新阅读记录（minType=3006 MarkTopicRead）
	// 同一 user→topic 使用 UPSERT 语义：存在则更新 read_at 和累加 duration
	UpsertReadRecord(ctx context.Context, read *TopicRead) error

	// ListReadsByUserID 分页查询用户的阅读历史（我的阅读页面）
	ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*TopicRead, int64, error)

	// CountDistinctReaders 统计帖子独立阅读人数（用于 2级 stats 更新）
	CountDistinctReaders(ctx context.Context, topicID string) (int64, error)
}
