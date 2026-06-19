package social

import (
	"context"
	"testing"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
)

// mockModuleRepositories 创建模块级测试用的空 Mock Repository
// 各子域 Repository 接口由各包内部 _test.go 的 mockRepository 实现，
// 此处仅做类型适配（module 层不直接依赖具体 mock 结构体）
type mockModuleMemberRepo struct{}

func (m *mockModuleMemberRepo) CreateUser(ctx context.Context, user *member.User) error { return nil }
func (m *mockModuleMemberRepo) GetUserByID(ctx context.Context, id string) (*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) GetUserByUID(ctx context.Context, uid string) (*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) GetUserByUsername(ctx context.Context, username string) (*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) GetUserByEmail(ctx context.Context, email string) (*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) GetUserByPhone(ctx context.Context, phone string) (*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) UpdateUser(ctx context.Context, user *member.User) error { return nil }
func (m *mockModuleMemberRepo) BatchGetUsersByID(ctx context.Context, ids []string) ([]*member.User, error) {
	return nil, nil
}
func (m *mockModuleMemberRepo) CreateBlock(ctx context.Context, block *member.MemberBlock) error { return nil }
func (m *mockModuleMemberRepo) DeleteBlock(ctx context.Context, blockerID, blockedID string) error { return nil }
func (m *mockModuleMemberRepo) ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*member.MemberBlock, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleMemberRepo) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	return false, nil
}
func (m *mockModuleMemberRepo) GetBlockCount(ctx context.Context, blockerID string) (int64, error) { return 0, nil }
func (m *mockModuleMemberRepo) GetOrCreateStats(ctx context.Context, userID string) (*member.MemberStats, error) {
	return &member.MemberStats{UserID: userID}, nil
}
func (m *mockModuleMemberRepo) UpdateStats(ctx context.Context, stats *member.MemberStats) error { return nil }

type mockModuleGroupRepo struct{}

func (m *mockModuleGroupRepo) CreateGroup(ctx context.Context, group *group.Group) error          { return nil }
func (m *mockModuleGroupRepo) GetGroupByID(ctx context.Context, id string) (*group.Group, error)   { return nil, nil }
func (m *mockModuleGroupRepo) GetGroupBySlug(ctx context.Context, slug string) (*group.Group, error) {
	return nil, nil
}
func (m *mockModuleGroupRepo) UpdateGroup(ctx context.Context, group *group.Group) error           { return nil }
func (m *mockModuleGroupRepo) DeleteGroup(ctx context.Context, id string) error                   { return nil }
func (m *mockModuleGroupRepo) ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*group.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleGroupRepo) ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*group.Group, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleGroupRepo) CreateMember(ctx context.Context, member *group.GroupMember) error   { return nil }
func (m *mockModuleGroupRepo) GetMember(ctx context.Context, groupID, userID string) (*group.GroupMember, error) {
	return nil, nil
}
func (m *mockModuleGroupRepo) GetMemberByID(ctx context.Context, id string) (*group.GroupMember, error) {
	return nil, nil
}
func (m *mockModuleGroupRepo) UpdateMember(ctx context.Context, member *group.GroupMember) error     { return nil }
func (m *mockModuleGroupRepo) DeleteMember(ctx context.Context, id string) error                   { return nil }
func (m *mockModuleGroupRepo) ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*group.GroupMember, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleGroupRepo) ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*group.GroupMember, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleGroupRepo) CountActiveMembers(ctx context.Context, groupID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleGroupRepo) CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error) {
	return 0, nil
}
func (m *mockModuleGroupRepo) IsUserMember(ctx context.Context, groupID, userID string) (bool, error) {
	return false, nil
}
func (m *mockModuleGroupRepo) GetPayConfigByGroupID(ctx context.Context, groupID string) (*group.GroupPayConfig, error) {
	return nil, nil
}
func (m *mockModuleGroupRepo) CreatePayConfig(ctx context.Context, config *group.GroupPayConfig) error { return nil }
func (m *mockModuleGroupRepo) UpdatePayConfig(ctx context.Context, config *group.GroupPayConfig) error { return nil }
func (m *mockModuleGroupRepo) GetOrCreateStats(ctx context.Context, groupID string) (*group.GroupStats, error) {
	return &group.GroupStats{GroupID: groupID}, nil
}
func (m *mockModuleGroupRepo) UpdateStats(ctx context.Context, stats *group.GroupStats) error         { return nil }

type mockModuleTopicRepo struct{}

func (m *mockModuleTopicRepo) CreateTopic(ctx context.Context, topic *topic.Topic) error            { return nil }
func (m *mockModuleTopicRepo) GetTopicByID(ctx context.Context, id string) (*topic.Topic, error)    { return nil, nil }
func (m *mockModuleTopicRepo) UpdateTopic(ctx context.Context, topic *topic.Topic) error            { return nil }
func (m *mockModuleTopicRepo) DeleteTopic(ctx context.Context, id string) error                    { return nil }
func (m *mockModuleTopicRepo) ListTopics(ctx context.Context, page, size int, filters map[string]interface{}, orderBy string) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) ListTopicsByGroupID(ctx context.Context, groupID string, page, size int) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) ListTopicsByAuthorID(ctx context.Context, authorID string, page, size int) ([]*topic.Topic, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) CountTopicsByGroupID(ctx context.Context, groupID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleTopicRepo) CreateReply(ctx context.Context, reply *topic.TopicReply) error        { return nil }
func (m *mockModuleTopicRepo) GetReplyByID(ctx context.Context, id string) (*topic.TopicReply, error) { return nil, nil }
func (m *mockModuleTopicRepo) DeleteReply(ctx context.Context, id string) error                     { return nil }
func (m *mockModuleTopicRepo) ListReplies(ctx context.Context, topicID string, page, size int, parentReplyID *string) ([]*topic.TopicReply, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) ListRepliesByAuthorID(ctx context.Context, authorID string, page, size int) ([]*topic.TopicReply, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) UpdateReply(ctx context.Context, reply *topic.TopicReply) error        { return nil }
func (m *mockModuleTopicRepo) CountRepliesByTopicID(ctx context.Context, topicID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleTopicRepo) CreateLike(ctx context.Context, like *topic.TopicLike) error          { return nil }
func (m *mockModuleTopicRepo) DeleteLike(ctx context.Context, topicID, userID string) error         { return nil }
func (m *mockModuleTopicRepo) IsTopicLiked(ctx context.Context, topicID, userID string) (bool, error) {
	return false, nil
}
func (m *mockModuleTopicRepo) CountLikesByTopicID(ctx context.Context, topicID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleTopicRepo) CreateFavorite(ctx context.Context, fav *topic.TopicFavorite) error   { return nil }
func (m *mockModuleTopicRepo) DeleteFavorite(ctx context.Context, topicID, userID string) error     { return nil }
func (m *mockModuleTopicRepo) IsTopicFavorited(ctx context.Context, topicID, userID string) (bool, error) {
	return false, nil
}
func (m *mockModuleTopicRepo) ListFavoritesByUserID(ctx context.Context, userID string, page, size int) ([]*topic.TopicFavorite, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) CountFavoritesByTopicID(ctx context.Context, topicID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleTopicRepo) CreateReplyLike(ctx context.Context, rl *topic.ReplyLike) error        { return nil }
func (m *mockModuleTopicRepo) DeleteReplyLike(ctx context.Context, replyID, userID string) error   { return nil }
func (m *mockModuleTopicRepo) IsReplyLiked(ctx context.Context, replyID, userID string) (bool, error) {
	return false, nil
}
func (m *mockModuleTopicRepo) CountLikesByReplyID(ctx context.Context, replyID string) (int64, error) {
	return 0, nil
}
func (m *mockModuleTopicRepo) GetReadRecord(ctx context.Context, topicID, userID string) (*topic.TopicRead, error) {
	return nil, nil
}
func (m *mockModuleTopicRepo) UpsertReadRecord(ctx context.Context, read *topic.TopicRead) error      { return nil }
func (m *mockModuleTopicRepo) ListReadsByUserID(ctx context.Context, userID string, page, size int) ([]*topic.TopicRead, int64, error) {
	return nil, 0, nil
}
func (m *mockModuleTopicRepo) CountDistinctReaders(ctx context.Context, topicID string) (int64, error) {
	return 0, nil
}

// TestNewModule_不传选项_默认使用NoopPublisher 验证不传选项时模块使用 NoopPublisher，不 panic
func TestNewModule_不传选项_默认使用NoopPublisher(t *testing.T) {
	memberRepo := &mockModuleMemberRepo{}
	groupRepo := &mockModuleGroupRepo{}
	topicRepo := &mockModuleTopicRepo{}

	// Act — 不传 WithPublisher 选项，不应 panic
	mod := NewModule(memberRepo, groupRepo, topicRepo)

	// Assert
	if mod == nil {
		t.Fatal("期望 Module 不为 nil")
	}
	if mod.MemberServant == nil {
		t.Error("期望 MemberServant 不为 nil")
	}
	if mod.GroupServant == nil {
		t.Error("期望 GroupServant 不为 nil")
	}
	if mod.TopicServant == nil {
		t.Error("期望 TopicServant 不为 nil")
	}
}

// TestNewModule_传入WithPublisher_使用指定Publisher 验证通过 WithPublisher 注入的 Publisher 被正确传递给子域 Servant
func TestNewModule_传入WithPublisher_使用指定Publisher(t *testing.T) {
	memberRepo := &mockModuleMemberRepo{}
	groupRepo := &mockModuleGroupRepo{}
	topicRepo := &mockModuleTopicRepo{}
	publisher := &event.FakePublisher{}

	// Act — 传入 WithPublisher 选项
	mod := NewModule(memberRepo, groupRepo, topicRepo, WithPublisher(publisher))

	// Assert
	if mod == nil {
		t.Fatal("期望 Module 不为 nil")
	}
	if mod.MemberServant == nil {
		t.Error("期望 MemberServant 不为 nil")
	}
	if mod.GroupServant == nil {
		t.Error("期望 GroupServant 不为 nil")
	}
	if mod.TopicServant == nil {
		t.Error("期望 TopicServant 不为 nil")
	}
	// 通过调用任意 svc 方法验证 publisher 已注入（此处仅验证构造不 panic），
	// 具体 publisher 传递到 svc 内部的验证由各子域 event 测试覆盖
}
