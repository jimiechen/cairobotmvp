package topic

import (
	"context"
	"sync"
)

// MemoryRepository 主题域 Repository 接口的纯内存实现
// 使用 map + sync.RWMutex 提供线程安全的数据存储
// 适用于单元测试、集成测试和开发阶段快速验证，不适用于生产环境
type MemoryRepository struct {
	mu         sync.RWMutex
	topics     map[string]*Topic           // key: Topic.ID
	replies    map[string]*TopicReply      // key: TopicReply.ID
	likes      map[string]*TopicLike       // key: userID:topicID
	favorites  map[string]*TopicFavorite   // key: userID:topicID
	replyLikes map[string]*ReplyLike       // key: userID:replyID
	reads      map[string]*TopicRead       // key: userID:topicID
}

// NewMemoryRepository 创建 MemoryRepository 实例并初始化所有内部存储 map
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		topics:     make(map[string]*Topic),
		replies:    make(map[string]*TopicReply),
		likes:      make(map[string]*TopicLike),
		favorites:  make(map[string]*TopicFavorite),
		replyLikes: make(map[string]*ReplyLike),
		reads:      make(map[string]*TopicRead),
	}
}

// ========== Topic 帖子 CRUD ==========

func (r *MemoryRepository) CreateTopic(ctx context.Context, topic *Topic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.topics[topic.ID] = topic
	return nil
}

func (r *MemoryRepository) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.topics[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *MemoryRepository) UpdateTopic(ctx context.Context, topic *Topic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.topics[topic.ID]; ok {
		r.topics[topic.ID] = topic
	}
	return nil
}

func (r *MemoryRepository) DeleteTopic(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.topics, id)
	return nil
}

func (r *MemoryRepository) ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*Topic, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Topic
	for _, t := range r.topics {
		if t.Status == TopicStatusActive {
			result = append(result, t)
		}
	}

	total := int64(len(result))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(result) {
		return []*Topic{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (r *MemoryRepository) ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*Topic, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Topic
	for _, t := range r.topics {
		if t.GroupID == groupID && t.Status == TopicStatusActive {
			result = append(result, t)
		}
	}
	total := int64(len(result))
	paged := paginateTopics(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*Topic, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Topic
	for _, t := range r.topics {
		if t.AuthorID == authorID && t.Status == TopicStatusActive {
			result = append(result, t)
		}
	}
	total := int64(len(result))
	paged := paginateTopics(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, t := range r.topics {
		if t.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

// ========== TopicReply 评论操作 ==========

func (r *MemoryRepository) CreateReply(ctx context.Context, reply *TopicReply) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replies[reply.ID] = reply
	return nil
}

func (r *MemoryRepository) GetReplyByID(ctx context.Context, id string) (*TopicReply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	repl, ok := r.replies[id]
	if !ok {
		return nil, nil
	}
	return repl, nil
}

func (r *MemoryRepository) DeleteReply(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.replies, id)
	return nil
}

func (r *MemoryRepository) ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*TopicReply, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TopicReply
	for _, repl := range r.replies {
		if repl.TopicID != topicID || repl.Status != ReplyStatusActive {
			continue
		}
		// 按 parent_reply_id 过滤获取子回复（非空时仅返回该父评论下的直接子回复）
		if parentReplyID != nil && repl.ParentReplyID != *parentReplyID {
			continue
		}
		result = append(result, repl)
	}

	total := int64(len(result))
	paged := paginateReplies(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*TopicReply, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TopicReply
	for _, repl := range r.replies {
		if repl.AuthorID == authorID {
			result = append(result, repl)
		}
	}
	total := int64(len(result))
	paged := paginateReplies(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) UpdateReply(ctx context.Context, reply *TopicReply) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.replies[reply.ID]; ok {
		r.replies[reply.ID] = reply
	}
	return nil
}

func (r *MemoryRepository) CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, repl := range r.replies {
		if repl.TopicID == topicID && repl.Status == ReplyStatusActive {
			count++
		}
	}
	return count, nil
}

// ========== TopicLike 点赞操作 ==========

func (r *MemoryRepository) CreateLike(ctx context.Context, like *TopicLike) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := compositeKey(like.UserID, like.TopicID)
	if _, exists := r.likes[key]; exists {
		return nil // 幂等：已存在则不重复写入
	}
	r.likes[key] = like
	return nil
}

func (r *MemoryRepository) DeleteLike(ctx context.Context, topicID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.likes, compositeKey(userID, topicID))
	return nil
}

func (r *MemoryRepository) IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.likes[compositeKey(userID, topicID)]
	return ok, nil
}

func (r *MemoryRepository) CountLikesByTopicID(ctx context.Context, topicID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for k := range r.likes {
		if matchSuffix(k, topicID) {
			count++
		}
	}
	return count, nil
}

// ========== TopicFavorite 收藏操作 ==========

func (r *MemoryRepository) CreateFavorite(ctx context.Context, favorite *TopicFavorite) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := compositeKey(favorite.UserID, favorite.TopicID)
	r.favorites[key] = favorite
	return nil
}

func (r *MemoryRepository) DeleteFavorite(ctx context.Context, topicID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.favorites, compositeKey(userID, topicID))
	return nil
}

func (r *MemoryRepository) IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.favorites[compositeKey(userID, topicID)]
	return ok, nil
}

func (r *MemoryRepository) ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*TopicFavorite, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TopicFavorite
	for _, f := range r.favorites {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	total := int64(len(result))
	paged := paginateFavorites(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for k := range r.favorites {
		if matchSuffix(k, topicID) {
			count++
		}
	}
	return count, nil
}

// ========== ReplyLike 评论点赞操作 ==========

func (r *MemoryRepository) CreateReplyLike(ctx context.Context, replyLike *ReplyLike) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := compositeKey(replyLike.UserID, replyLike.ReplyID)
	r.replyLikes[key] = replyLike
	return nil
}

func (r *MemoryRepository) DeleteReplyLike(ctx context.Context, replyID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.replyLikes, compositeKey(userID, replyID))
	return nil
}

func (r *MemoryRepository) IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.replyLikes[compositeKey(userID, replyID)]
	return ok, nil
}

func (r *MemoryRepository) CountLikesByReplyID(ctx context.Context, replyID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for k := range r.replyLikes {
		if matchSuffix(k, replyID) {
			count++
		}
	}
	return count, nil
}

// ========== TopicRead 阅读记录操作 ==========

func (r *MemoryRepository) GetReadRecord(ctx context.Context, topicID, userID string) (*TopicRead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	read, ok := r.reads[compositeKey(userID, topicID)]
	if !ok {
		return nil, nil
	}
	return read, nil
}

func (r *MemoryRepository) UpsertReadRecord(ctx context.Context, read *TopicRead) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := compositeKey(read.UserID, read.TopicID)

	// UPSERT 语义：存在则更新 read_at 和累加 duration
	if existing, ok := r.reads[key]; ok {
		read.ReadDuration += existing.ReadDuration
	}
	r.reads[key] = read
	return nil
}

func (r *MemoryRepository) ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*TopicRead, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*TopicRead
	for _, rd := range r.reads {
		if rd.UserID == userID {
			result = append(result, rd)
		}
	}
	total := int64(len(result))
	paged := paginateReads(result, page, size)
	return paged, total, nil
}

func (r *MemoryRepository) CountDistinctReaders(ctx context.Context, topicID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	readers := make(map[string]bool)
	for k := range r.reads {
		userID, targetID := splitCompositeKey(k)
		if targetID == topicID {
			readers[userID] = true
		}
	}
	return int64(len(readers)), nil
}

// ========== 内部辅助方法 ==========

// compositeKey 生成复合键，格式为 "part1:part2"
// 用于 likes、favorites、reads 等关系型数据的 map 存储
func compositeKey(part1, part2 string) string {
	return part1 + ":" + part2
}

// splitCompositeKey 将复合键拆分为两个部分
func splitCompositeKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

// matchSuffix 检查复合键的后半部分是否匹配目标值
// 用于 CountLikesByTopicID、CountFavoritesByTopicID 等按目标 ID 聚合计数的场景
func matchSuffix(key, target string) bool {
	idx := indexOfSeparator(key)
	if idx < 0 || idx+1 >= len(key) {
		return false
	}
	return key[idx+1:] == target
}

// indexOfSeparator 返回 ':' 在字符串中的位置索引，未找到返回 -1
func indexOfSeparator(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return i
		}
	}
	return -1
}

// paginateTopics 对帖子列表执行分页切片
func paginateTopics(list []*Topic, page, size int) []*Topic {
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(list) {
		return []*Topic{}
	}
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// paginateReplies 对评论列表执行分页切片
func paginateReplies(list []*TopicReply, page, size int) []*TopicReply {
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(list) {
		return []*TopicReply{}
	}
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// paginateFavorites 对收藏列表执行分页切片
func paginateFavorites(list []*TopicFavorite, page, size int) []*TopicFavorite {
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(list) {
		return []*TopicFavorite{}
	}
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// paginateReads 对阅读记录列表执行分页切片
func paginateReads(list []*TopicRead, page, size int) []*TopicRead {
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(list) {
		return []*TopicRead{}
	}
	end := start + size
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}
