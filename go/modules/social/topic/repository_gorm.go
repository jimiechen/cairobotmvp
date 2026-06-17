// repository_gorm.go — Topic 域 Repository 接口的 GORM 数据库实现
// 职责：将 Repository 接口定义的全部方法映射为 GORM 数据库操作，覆盖帖子、评论、点赞、收藏、阅读记录、评论点赞六张表
// 不负责：业务校验（如权限判断、内容审核）、缓存、事务编排（由 svc 层处理）
//
// 相关文档：
// - PRD 社交域 MVP-P0 Step 8：Topic 域 Repository 接口 GORM 实现

package topic

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormRepository Repository 接口的 GORM 实现
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository 创建基于 GORM 的主题域仓库实例
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// ========== Topic 帖子 CRUD 实现 ==========

// CreateTopic 创建帖子记录，插入 topics 表
func (r *GormRepository) CreateTopic(ctx context.Context, topic *Topic) error {
	err := r.db.WithContext(ctx).Create(topic).Error
	if err != nil {
		return fmt.Errorf("CreateTopic(id=%s): %w", topic.ID, err)
	}
	return nil
}

// GetTopicByID 根据 ID 查询帖子详情
// 记录不存在时返回 nil, nil（不返回 gorm.ErrRecordNotFound）
func (r *GormRepository) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	var topic Topic
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&topic).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetTopicByID(id=%s): %w", id, err)
	}
	return &topic, nil
}

// UpdateTopic 更新帖子字段，使用 Select("*") 确保零值字段（如 false/0）也能被更新
func (r *GormRepository) UpdateTopic(ctx context.Context, topic *Topic) error {
	result := r.db.WithContext(ctx).
		Model(&Topic{}).
		Select("*").
		Where("id = ?", topic.ID).
		Updates(topic)
	if result.Error != nil {
		return fmt.Errorf("UpdateTopic(id=%s): %w", topic.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("UpdateTopic: 帖子不存在(id=%s)", topic.ID)
	}
	return nil
}

// DeleteTopic 删除帖子记录（物理删除）
func (r *GormRepository) DeleteTopic(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Topic{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("DeleteTopic(id=%s): %w", id, result.Error)
	}
	return nil
}

// ListTopics 分页查询帖子列表，支持动态 filters 条件筛选和排序
// filters map 的 key 对应数据库列名；orderBy 为排序字段（默认 created_at DESC）
// 返回当前页数据、总数、错误
func (r *GormRepository) ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*Topic, int64, error) {
	var topics []*Topic
	var total int64

	query := r.db.WithContext(ctx).Model(&Topic{})

	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopics count: %w", err)
	}

	if orderBy == "" {
		orderBy = "created_at DESC"
	}
	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order(orderBy).Find(&topics).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopics(page=%d,size=%d): %w", page, size, err)
	}

	return topics, total, nil
}

// ListTopicsByGroupID 分页查询某群组的帖子列表
func (r *GormRepository) ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*Topic, int64, error) {
	var topics []*Topic
	var total int64

	query := r.db.WithContext(ctx).Model(&Topic{}).Where("group_id = ?", groupID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopicsByGroupID(group=%s) count: %w", groupID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&topics).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopicsByGroupID(group=%s,page=%d,size=%d): %w", groupID, page, size, err)
	}

	return topics, total, nil
}

// ListTopicsByAuthorID 分页查询某用户的帖子列表（个人主页）
func (r *GormRepository) ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*Topic, int64, error) {
	var topics []*Topic
	var total int64

	query := r.db.WithContext(ctx).Model(&Topic{}).Where("author_id = ?", authorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopicsByAuthorID(author=%s) count: %w", authorID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&topics).Error; err != nil {
		return nil, 0, fmt.Errorf("ListTopicsByAuthorID(author=%s,page=%d,size=%d): %w", authorID, page, size, err)
	}

	return topics, total, nil
}

// CountTopicsByGroupID 统计群组内已发布帖子数
func (r *GormRepository) CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Topic{}).
		Where("group_id = ? AND status = ?", groupID, TopicStatusActive).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountTopicsByGroupID(group=%s): %w", groupID, err)
	}
	return count, nil
}

// ========== TopicReply 评论操作实现 ==========

// CreateReply 创建评论/回复记录
func (r *GormRepository) CreateReply(ctx context.Context, reply *TopicReply) error {
	err := r.db.WithContext(ctx).Create(reply).Error
	if err != nil {
		return fmt.Errorf("CreateReply(id=%s,topic=%s): %w", reply.ID, reply.TopicID, err)
	}
	return nil
}

// GetReplyByID 根据评论 ID 查询单条评论
// 记录不存在时返回 nil, nil
func (r *GormRepository) GetReplyByID(ctx context.Context, id string) (*TopicReply, error) {
	var reply TopicReply
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&reply).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetReplyByID(id=%s): %w", id, err)
	}
	return &reply, nil
}

// DeleteReply 删除评论记录（物理删除）
func (r *GormRepository) DeleteReply(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&TopicReply{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("DeleteReply(id=%s): %w", id, result.Error)
	}
	return nil
}

// ListReplies 分页查询帖子的评论列表
// parentReplyID 非空时仅返回该评论的子回复（楼中楼），为空时返回顶层评论
func (r *GormRepository) ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*TopicReply, int64, error) {
	var replies []*TopicReply
	var total int64

	query := r.db.WithContext(ctx).Model(&TopicReply{}).Where("topic_id = ?", topicID)

	if parentReplyID != nil {
		query = query.Where("parent_reply_id = ?", *parentReplyID)
	} else {
		query = query.Where("parent_reply_id IS NULL OR parent_reply_id = ''")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListReplies(topic=%s) count: %w", topicID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("created_at ASC").Find(&replies).Error; err != nil {
		return nil, 0, fmt.Errorf("ListReplies(topic=%s,page=%d,size=%d): %w", topicID, page, size, err)
	}

	return replies, total, nil
}

// ListRepliesByAuthorID 分页查询某用户发表的评论列表
func (r *GormRepository) ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*TopicReply, int64, error) {
	var replies []*TopicReply
	var total int64

	query := r.db.WithContext(ctx).Model(&TopicReply{}).Where("author_id = ?", authorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListRepliesByAuthorID(author=%s) count: %w", authorID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&replies).Error; err != nil {
		return nil, 0, fmt.Errorf("ListRepliesByAuthorID(author=%s,page=%d,size=%d): %w", authorID, page, size, err)
	}

	return replies, total, nil
}

// UpdateReply 更新评论字段（状态变更等），使用 Select("*") 确保零值字段也能更新
func (r *GormRepository) UpdateReply(ctx context.Context, reply *TopicReply) error {
	result := r.db.WithContext(ctx).
		Model(&TopicReply{}).
		Select("*").
		Where("id = ?", reply.ID).
		Updates(reply)
	if result.Error != nil {
		return fmt.Errorf("UpdateReply(id=%s): %w", reply.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("UpdateReply: 评论不存在(id=%s)", reply.ID)
	}
	return nil
}

// CountRepliesByTopicID 统计帖子有效评论数（status=1 正常状态）
func (r *GormRepository) CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TopicReply{}).
		Where("topic_id = ? AND status = ?", topicID, ReplyStatusActive).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountRepliesByTopicID(topic=%s): %w", topicID, err)
	}
	return count, nil
}

// ========== TopicLike 点赞操作实现 ==========

// CreateLike 创建帖子点赞记录，使用 uniqueIndex 保证幂等（重复点赞不报错）
func (r *GormRepository) CreateLike(ctx context.Context, like *TopicLike) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(like).Error
	if err != nil {
		return fmt.Errorf("CreateLike(topic=%s,user=%s): %w", like.TopicID, like.UserID, err)
	}
	return nil
}

// DeleteLike 取消帖子点赞记录
func (r *GormRepository) DeleteLike(ctx context.Context, topicID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", topicID, userID).
		Delete(&TopicLike{})
	if result.Error != nil {
		return fmt.Errorf("DeleteLike(topic=%s,user=%s): %w", topicID, userID, result.Error)
	}
	return nil
}

// IsTopicLiked 检查用户是否已点赞某帖子
func (r *GormRepository) IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error) {
	var like TopicLike
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", topicID, userID).
		First(&like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("IsTopicLiked(topic=%s,user=%s): %w", topicID, userID, err)
	}
	return true, nil
}

// CountLikesByTopicID 统计帖子点赞数
func (r *GormRepository) CountLikesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TopicLike{}).
		Where("topic_id = ?", topicID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountLikesByTopicID(topic=%s): %w", topicID, err)
	}
	return count, nil
}

// ========== TopicFavorite 收藏操作实现 ==========

// CreateFavorite 创建收藏记录，使用 uniqueIndex 保证幂等
func (r *GormRepository) CreateFavorite(ctx context.Context, favorite *TopicFavorite) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(favorite).Error
	if err != nil {
		return fmt.Errorf("CreateFavorite(topic=%s,user=%s): %w", favorite.TopicID, favorite.UserID, err)
	}
	return nil
}

// DeleteFavorite 取消收藏记录
func (r *GormRepository) DeleteFavorite(ctx context.Context, topicID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", topicID, userID).
		Delete(&TopicFavorite{})
	if result.Error != nil {
		return fmt.Errorf("DeleteFavorite(topic=%s,user=%s): %w", topicID, userID, result.Error)
	}
	return nil
}

// IsTopicFavorited 检查用户是否已收藏某帖子
func (r *GormRepository) IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error) {
	var fav TopicFavorite
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", topicID, userID).
		First(&fav).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("IsTopicFavorited(topic=%s,user=%s): %w", topicID, userID, err)
	}
	return true, nil
}

// ListFavoritesByUserID 分页查询用户的收藏列表
func (r *GormRepository) ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*TopicFavorite, int64, error) {
	var favorites []*TopicFavorite
	var total int64

	query := r.db.WithContext(ctx).Model(&TopicFavorite{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListFavoritesByUserID(user=%s) count: %w", userID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&favorites).Error; err != nil {
		return nil, 0, fmt.Errorf("ListFavoritesByUserID(user=%s,page=%d,size=%d): %w", userID, page, size, err)
	}

	return favorites, total, nil
}

// CountFavoritesByTopicID 统计帖子收藏数
func (r *GormRepository) CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TopicFavorite{}).
		Where("topic_id = ?", topicID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountFavoritesByTopicID(topic=%s): %w", topicID, err)
	}
	return count, nil
}

// ========== ReplyLike 评论点赞操作实现 ==========

// CreateReplyLike 创建评论点赞记录，使用 uniqueIndex 保证幂等
func (r *GormRepository) CreateReplyLike(ctx context.Context, replyLike *ReplyLike) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(replyLike).Error
	if err != nil {
		return fmt.Errorf("CreateReplyLike(reply=%s,user=%s): %w", replyLike.ReplyID, replyLike.UserID, err)
	}
	return nil
}

// DeleteReplyLike 取消评论点赞记录
func (r *GormRepository) DeleteReplyLike(ctx context.Context, replyID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("reply_id = ? AND user_id = ?", replyID, userID).
		Delete(&ReplyLike{})
	if result.Error != nil {
		return fmt.Errorf("DeleteReplyLike(reply=%s,user=%s): %w", replyID, userID, result.Error)
	}
	return nil
}

// IsReplyLiked 检查用户是否已点赞某评论
func (r *GormRepository) IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error) {
	var rl ReplyLike
	err := r.db.WithContext(ctx).
		Where("reply_id = ? AND user_id = ?", replyID, userID).
		First(&rl).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("IsReplyLiked(reply=%s,user=%s): %w", replyID, userID, err)
	}
	return true, nil
}

// CountLikesByReplyID 统计评论点赞数
func (r *GormRepository) CountLikesByReplyID(ctx context.Context, replyID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReplyLike{}).
		Where("reply_id = ?", replyID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountLikesByReplyID(reply=%s): %w", replyID, err)
	}
	return count, nil
}

// ========== TopicRead 阅读记录操作实现 ==========

// GetReadRecord 获取用户对某帖子的阅读记录
// 记录不存在时返回 nil, nil
func (r *GormRepository) GetReadRecord(ctx context.Context, topicID, userID string) (*TopicRead, error) {
	var read TopicRead
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", topicID, userID).
		First(&read).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetReadRecord(topic=%s,user=%s): %w", topicID, userID, err)
	}
	return &read, nil
}

// UpsertReadRecord 创建或更新阅读记录（UPSERT 语义）
// 同一 user→topic 使用 clause.OnConflict 实现幂等的 upsert：存在则更新 read_at 和累加 duration
func (r *GormRepository) UpsertReadRecord(ctx context.Context, read *TopicRead) error {
	// 使用 find-or-create 模式确保 SQLite 兼容（ON CONFLICT 需要 UNIQUE 约束而非 index）
	var existing TopicRead
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND user_id = ?", read.TopicID, read.UserID).
		First(&existing).Error
	if err == nil {
		// 记录已存在，更新阅读时间和替换时长
		return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"read_at":       read.ReadAt,
			"read_duration": read.ReadDuration,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("UpsertReadRecord(topic=%s,user=%s): %w", read.TopicID, read.UserID, err)
	}
	return r.db.WithContext(ctx).Create(read).Error
}

// ListReadsByUserID 分页查询用户的阅读历史
func (r *GormRepository) ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*TopicRead, int64, error) {
	var reads []*TopicRead
	var total int64

	query := r.db.WithContext(ctx).Model(&TopicRead{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListReadsByUserID(user=%s) count: %w", userID, err)
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if err := query.Offset(offset).Limit(size).Order("read_at DESC").Find(&reads).Error; err != nil {
		return nil, 0, fmt.Errorf("ListReadsByUserID(user=%s,page=%d,size=%d): %w", userID, page, size, err)
	}

	return reads, total, nil
}

// CountDistinctReaders 统计帖子独立阅读人数
func (r *GormRepository) CountDistinctReaders(ctx context.Context, topicID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TopicRead{}).
		Where("topic_id = ?", topicID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountDistinctReaders(topic=%s): %w", topicID, err)
	}
	return count, nil
}
