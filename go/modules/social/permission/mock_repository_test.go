package permission

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
)

// mockMemberRepo 实现 member.Repository 接口的 Mock（仅 permission 测试所需方法）
type mockMemberRepo struct {
	users map[string]*member.User // key: user.ID
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{users: make(map[string]*member.User)}
}

func (m *mockMemberRepo) GetUserByID(ctx context.Context, id string) (*member.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}

// 以下方法为满足接口编译要求，permission service 不调用
func (m *mockMemberRepo) CreateUser(ctx context.Context, user *member.User) error                               { return nil }
func (m *mockMemberRepo) GetUserByUID(ctx context.Context, uid string) (*member.User, error)                  { return nil, nil }
func (m *mockMemberRepo) GetUserByUsername(ctx context.Context, username string) (*member.User, error)        { return nil, nil }
func (m *mockMemberRepo) GetUserByEmail(ctx context.Context, email string) (*member.User, error)              { return nil, nil }
func (m *mockMemberRepo) GetUserByPhone(ctx context.Context, phone string) (*member.User, error)              { return nil, nil }
func (m *mockMemberRepo) UpdateUser(ctx context.Context, user *member.User) error                            { return nil }
func (m *mockMemberRepo) BatchGetUsersByID(ctx context.Context, ids []string) ([]*member.User, error)         { return nil, nil }
func (m *mockMemberRepo) CreateBlock(ctx context.Context, block *member.MemberBlock) error                   { return nil }
func (m *mockMemberRepo) DeleteBlock(ctx context.Context, blockerID, blockedID string) error                 { return nil }
func (m *mockMemberRepo) IsBlocked(ctx context.Context, userID, targetUserID string) (bool, error)           { return false, nil }
func (m *mockMemberRepo) ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*member.MemberBlock, int64, error) {
	return nil, 0, nil
}
func (m *mockMemberRepo) GetBlockCount(ctx context.Context, blockerID string) (int64, error) { return 0, nil }
func (m *mockMemberRepo) GetOrCreateStats(ctx context.Context, userID string) (*member.MemberStats, error)  { return nil, nil }
func (m *mockMemberRepo) UpdateStats(ctx context.Context, stats *member.MemberStats) error                     { return nil }

// mockGroupRepo 实现 group.Repository 接口的 Mock（仅 permission 测试所需方法）
type mockGroupRepo struct {
	members   map[string]*group.GroupMember // key: groupID:userID
}

func newMockGroupRepo() *mockGroupRepo {
	return &mockGroupRepo{members: make(map[string]*group.GroupMember)}
}

func (m *mockGroupRepo) GetMember(ctx context.Context, groupID, userID string) (*group.GroupMember, error) {
	if mem, ok := m.members[groupID+":"+userID]; ok {
		return mem, nil
	}
	return nil, nil
}

// 以下方法为满足接口编译要求，permission service 不调用
func (m *mockGroupRepo) CreateGroup(ctx context.Context, g *group.Group) error                                    { return nil }
func (m *mockGroupRepo) GetGroupByID(ctx context.Context, id string) (*group.Group, error)                       { return nil, nil }
func (m *mockGroupRepo) GetGroupBySlug(ctx context.Context, slug string) (*group.Group, error)                  { return nil, nil }
func (m *mockGroupRepo) UpdateGroup(ctx context.Context, g *group.Group) error                                  { return nil }
func (m *mockGroupRepo) DeleteGroup(ctx context.Context, id string) error                                        { return nil }
func (m *mockGroupRepo) ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*group.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*group.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) CreateMember(ctx context.Context, mem *group.GroupMember) error                         { return nil }
func (m *mockGroupRepo) GetMemberByID(ctx context.Context, id string) (*group.GroupMember, error)               { return nil, nil }
func (m *mockGroupRepo) UpdateMember(ctx context.Context, mem *group.GroupMember) error                         { return nil }
func (m *mockGroupRepo) DeleteMember(ctx context.Context, id string) error                                      { return nil }
func (m *mockGroupRepo) ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*group.GroupMember, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*group.GroupMember, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) CountActiveMembers(ctx context.Context, groupID string) (int64, error)                  { return 0, nil }
func (m *mockGroupRepo) CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error)       { return 0, nil }
func (m *mockGroupRepo) IsUserMember(ctx context.Context, groupID, userID string) (bool, error)                 { return false, nil }
func (m *mockGroupRepo) GetPayConfigByGroupID(ctx context.Context, groupID string) (*group.GroupPayConfig, error) {
	return nil, nil
}
func (m *mockGroupRepo) CreatePayConfig(ctx context.Context, config *group.GroupPayConfig) error                { return nil }
func (m *mockGroupRepo) UpdatePayConfig(ctx context.Context, config *group.GroupPayConfig) error                { return nil }
func (m *mockGroupRepo) GetOrCreateStats(ctx context.Context, groupID string) (*group.GroupStats, error)        { return nil, nil }
func (m *mockGroupRepo) UpdateStats(ctx context.Context, stats *group.GroupStats) error                        { return nil }

// mockTopicRepo 实现 topic.Repository 接口的 Mock（仅 permission 测试所需方法）
type mockTopicRepo struct {
	topics map[string]*topic.Topic // key: topic.ID
}

func newMockTopicRepo() *mockTopicRepo {
	return &mockTopicRepo{topics: make(map[string]*topic.Topic)}
}

func (m *mockTopicRepo) GetTopicByID(ctx context.Context, id string) (*topic.Topic, error) {
	if t, ok := m.topics[id]; ok {
		return t, nil
	}
	return nil, nil
}

// 以下方法为满足接口编译要求，permission service 不调用
func (m *mockTopicRepo) CreateTopic(ctx context.Context, t *topic.Topic) error                                       { return nil }
func (m *mockTopicRepo) UpdateTopic(ctx context.Context, t *topic.Topic) error                                     { return nil }
func (m *mockTopicRepo) DeleteTopic(ctx context.Context, id string) error                                          { return nil }
func (m *mockTopicRepo) ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error)                  { return 0, nil }
func (m *mockTopicRepo) CreateReply(ctx context.Context, r *topic.TopicReply) error                                { return nil }
func (m *mockTopicRepo) GetReplyByID(ctx context.Context, id string) (*topic.TopicReply, error)                    { return nil, nil }
func (m *mockTopicRepo) DeleteReply(ctx context.Context, id string) error                                           { return nil }
func (m *mockTopicRepo) ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*topic.TopicReply, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*topic.TopicReply, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) UpdateReply(ctx context.Context, r *topic.TopicReply) error                                { return nil }
func (m *mockTopicRepo) CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error)                  { return 0, nil }
func (m *mockTopicRepo) CreateLike(ctx context.Context, l *topic.TopicLike) error                                 { return nil }
func (m *mockTopicRepo) DeleteLike(ctx context.Context, topicID, userID string) error                              { return nil }
func (m *mockTopicRepo) IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error)                   { return false, nil }
func (m *mockTopicRepo) CountLikesByTopicID(ctx context.Context, topicID string) (int64, error)                   { return 0, nil }
func (m *mockTopicRepo) CreateFavorite(ctx context.Context, f *topic.TopicFavorite) error                         { return nil }
func (m *mockTopicRepo) DeleteFavorite(ctx context.Context, topicID, userID string) error                          { return nil }
func (m *mockTopicRepo) IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error)               { return false, nil }
func (m *mockTopicRepo) ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*topic.TopicFavorite, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error)               { return 0, nil }
func (m *mockTopicRepo) CreateReplyLike(ctx context.Context, rl *topic.ReplyLike) error                           { return nil }
func (m *mockTopicRepo) DeleteReplyLike(ctx context.Context, replyID, userID string) error                        { return nil }
func (m *mockTopicRepo) IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error)                  { return false, nil }
func (m *mockTopicRepo) CountLikesByReplyID(ctx context.Context, replyID string) (int64, error)                  { return 0, nil }
func (m *mockTopicRepo) GetReadRecord(ctx context.Context, topicID, userID string) (*topic.TopicRead, error)      { return nil, nil }
func (m *mockTopicRepo) UpsertReadRecord(ctx context.Context, r *topic.TopicRead) error                           { return nil }
func (m *mockTopicRepo) ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*topic.TopicRead, int64, error) {
	return nil, 0, nil
}
func (m *mockTopicRepo) CountDistinctReaders(ctx context.Context, topicID string) (int64, error)                 { return 0, nil }
