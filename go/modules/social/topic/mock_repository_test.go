package topic

import (
	"context"
)

// mockRepository 实现 Repository 接口的 Mock，用于隔离数据库
type mockRepository struct {
	topics      map[string]*Topic
	replies     map[string]*TopicReply
	likes       map[string]*TopicLike    // key: userID:topicID
	favorites  map[string]*TopicFavorite // key: userID:topicID
	reads       map[string]*TopicRead     // key: userID:topicID
	replyLikes  map[string]*ReplyLike     // key: userID:replyID
}

// newMockRepository 创建 Mock 实例，所有 map 初始化为空
func newMockRepository() *mockRepository {
	return &mockRepository{
		topics:      make(map[string]*Topic),
		replies:     make(map[string]*TopicReply),
		likes:       make(map[string]*TopicLike),
		favorites:   make(map[string]*TopicFavorite),
		reads:       make(map[string]*TopicRead),
		replyLikes:  make(map[string]*ReplyLike),
	}
}

// ========== Topic CRUD ==========

func (m *mockRepository) CreateTopic(ctx context.Context, topic *Topic) error {
	m.topics[topic.ID] = topic
	return nil
}

func (m *mockRepository) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	if t, ok := m.topics[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockRepository) UpdateTopic(ctx context.Context, topic *Topic) error {
	if _, ok := m.topics[topic.ID]; ok {
		m.topics[topic.ID] = topic
	}
	return nil
}

func (m *mockRepository) DeleteTopic(ctx context.Context, id string) error {
	delete(m.topics, id)
	return nil
}

func (m *mockRepository) ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*Topic, int64, error) {
	var result []*Topic
	for _, t := range m.topics {
		if t.Status == 2 { // 只返回已发布状态
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

func (m *mockRepository) ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*Topic, int64, error) {
	var result []*Topic
	for _, t := range m.topics {
		if t.GroupID == groupID && t.Status == 2 {
			result = append(result, t)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*Topic, int64, error) {
	var result []*Topic
	for _, t := range m.topics {
		if t.AuthorID == authorID && t.Status == 2 {
			result = append(result, t)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error) {
	var count int64
	for _, t := range m.topics {
		if t.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

// ========== Reply ops ==========

func (m *mockRepository) CreateReply(ctx context.Context, reply *TopicReply) error {
	m.replies[reply.ID] = reply
	return nil
}

func (m *mockRepository) GetReplyByID(ctx context.Context, id string) (*TopicReply, error) {
	if r, ok := m.replies[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockRepository) DeleteReply(ctx context.Context, id string) error {
	delete(m.replies, id)
	return nil
}

func (m *mockRepository) ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*TopicReply, int64, error) {
	var result []*TopicReply
	for _, r := range m.replies {
		if r.TopicID == topicID && r.Status == 1 {
			result = append(result, r)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*TopicReply, int64, error) {
	var result []*TopicReply
	for _, r := range m.replies {
		if r.AuthorID == authorID {
			result = append(result, r)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) UpdateReply(ctx context.Context, reply *TopicReply) error {
	if _, ok := m.replies[reply.ID]; ok {
		m.replies[reply.ID] = reply
	}
	return nil
}

func (m *mockRepository) CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	for _, r := range m.replies {
		if r.TopicID == topicID && r.Status == 1 {
			count++
		}
	}
	return count, nil
}

// ========== Like ops ==========

func (m *mockRepository) CreateLike(ctx context.Context, like *TopicLike) error {
	key := like.UserID + ":" + like.TopicID
	m.likes[key] = like
	return nil
}

func (m *mockRepository) DeleteLike(ctx context.Context, topicID, userID string) error {
	key := userID + ":" + topicID
	delete(m.likes, key)
	return nil
}

func (m *mockRepository) IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error) {
	key := userID + ":" + topicID
	_, ok := m.likes[key]
	return ok, nil
}

func (m *mockRepository) CountLikesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	for k := range m.likes {
		if len(k) > len(topicID) && k[len(k)-len(topicID):] == topicID {
			count++
		}
	}
	return count, nil
}

// ========== Favorite ops ==========

func (m *mockRepository) CreateFavorite(ctx context.Context, fav *TopicFavorite) error {
	key := fav.UserID + ":" + fav.TopicID
	m.favorites[key] = fav
	return nil
}

func (m *mockRepository) DeleteFavorite(ctx context.Context, topicID, userID string) error {
	key := userID + ":" + topicID
	delete(m.favorites, key)
	return nil
}

func (m *mockRepository) IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error) {
	key := userID + ":" + topicID
	_, ok := m.favorites[key]
	return ok, nil
}

func (m *mockRepository) ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*TopicFavorite, int64, error) {
	var result []*TopicFavorite
	for _, f := range m.favorites {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error) {
	var count int64
	for k := range m.favorites {
		if len(k) > len(topicID) && k[len(k)-len(topicID):] == topicID {
			count++
		}
	}
	return count, nil
}

// ========== ReplyLike ops ==========

func (m *mockRepository) CreateReplyLike(ctx context.Context, rl *ReplyLike) error {
	key := rl.UserID + ":" + rl.ReplyID
	m.replyLikes[key] = rl
	return nil
}

func (m *mockRepository) DeleteReplyLike(ctx context.Context, replyID, userID string) error {
	key := userID + ":" + replyID
	delete(m.replyLikes, key)
	return nil
}

func (m *mockRepository) IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error) {
	key := userID + ":" + replyID
	_, ok := m.replyLikes[key]
	return ok, nil
}

func (m *mockRepository) CountLikesByReplyID(ctx context.Context, replyID string) (int64, error) {
	var count int64
	for k := range m.replyLikes {
		if len(k) > len(replyID) && k[len(k)-len(replyID):] == replyID {
			count++
		}
	}
	return count, nil
}

// ========== Read ops ==========

func (m *mockRepository) GetReadRecord(ctx context.Context, topicID, userID string) (*TopicRead, error) {
	key := userID + ":" + topicID
	if r, ok := m.reads[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockRepository) UpsertReadRecord(ctx context.Context, read *TopicRead) error {
	key := read.UserID + ":" + read.TopicID
	m.reads[key] = read
	return nil
}

func (m *mockRepository) ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*TopicRead, int64, error) {
	var result []*TopicRead
	for _, r := range m.reads {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) CountDistinctReaders(ctx context.Context, topicID string) (int64, error) {
	readers := make(map[string]bool)
	for k := range m.reads {
		idx := -1
		for i, c := range k {
			if c == ':' {
				idx = i
				break
			}
		}
		if idx >= 0 && k[idx+1:] == topicID {
			readers[k[:idx]] = true
		}
	}
	return int64(len(readers)), nil
}
