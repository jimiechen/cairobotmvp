package group

import (
	"context"
)

// mockRepository 实现 Repository 接口的 Mock，使用 map 存储隔离数据库
type mockRepository struct {
	groups     map[string]*Group          // key: group.ID 或 group.Slug
	members    map[string]*GroupMember    // key: member.ID
	memberIdx  map[string]*GroupMember    // key: groupID:userID
	payConfigs map[string]*GroupPayConfig // key: payConfig.GroupID
	stats      map[string]*GroupStats     // key: stats.GroupID
}

// newMockRepository 创建空的 mock repository 实例
func newMockRepository() *mockRepository {
	return &mockRepository{
		groups:     make(map[string]*Group),
		members:    make(map[string]*GroupMember),
		memberIdx:  make(map[string]*GroupMember),
		payConfigs: make(map[string]*GroupPayConfig),
		stats:      make(map[string]*GroupStats),
	}
}

// ========== Group CRUD 实现 ==========

func (m *mockRepository) CreateGroup(ctx context.Context, group *Group) error {
	m.groups[group.ID] = group
	m.groups[group.Slug] = group // slug 索引
	return nil
}

func (m *mockRepository) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	return m.groups[id], nil
}

func (m *mockRepository) GetGroupBySlug(ctx context.Context, slug string) (*Group, error) {
	return m.groups[slug], nil
}

func (m *mockRepository) UpdateGroup(ctx context.Context, group *Group) error {
	if _, ok := m.groups[group.ID]; ok {
		m.groups[group.ID] = group
	}
	return nil
}

func (m *mockRepository) DeleteGroup(ctx context.Context, id string) error {
	g := m.groups[id]
	if g != nil {
		delete(m.groups, g.Slug)
	}
	delete(m.groups, id)
	return nil
}

func (m *mockRepository) ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*Group, int64, error) {
	var result []*Group
	for k, v := range m.groups {
		// 避免重复（slug 索引也会存入）
		if k == v.ID {
			result = append(result, v)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*Group, int64, error) {
	var result []*Group
	for k, v := range m.groups {
		if k == v.ID && v.OwnerID == ownerID {
			result = append(result, v)
		}
	}
	return result, int64(len(result)), nil
}

// ========== GroupMember 操作实现 ==========

func (m *mockRepository) CreateMember(ctx context.Context, member *GroupMember) error {
	m.members[member.ID] = member
	m.memberIdx[member.GroupID+":"+member.UserID] = member
	return nil
}

func (m *mockRepository) GetMember(ctx context.Context, groupID, userID string) (*GroupMember, error) {
	return m.memberIdx[groupID+":"+userID], nil
}

func (m *mockRepository) GetMemberByID(ctx context.Context, id string) (*GroupMember, error) {
	return m.members[id], nil
}

func (m *mockRepository) UpdateMember(ctx context.Context, member *GroupMember) error {
	if _, ok := m.members[member.ID]; ok {
		m.members[member.ID] = member
		m.memberIdx[member.GroupID+":"+member.UserID] = member
	}
	return nil
}

func (m *mockRepository) DeleteMember(ctx context.Context, id string) error {
	member := m.members[id]
	if member != nil {
		delete(m.memberIdx, member.GroupID+":"+member.UserID)
	}
	delete(m.members, id)
	return nil
}

func (m *mockRepository) ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*GroupMember, int64, error) {
	var result []*GroupMember
	for k, v := range m.memberIdx {
		if len(k) > 0 && k[:len(groupID)] == groupID {
			if role != nil && v.Role != *role {
				continue
			}
			if status != nil && v.Status != *status {
				continue
			}
			result = append(result, v)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*GroupMember, int64, error) {
	var result []*GroupMember
	for k, v := range m.memberIdx {
		if len(k) > 0 && v.UserID == userID {
			result = append(result, v)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) CountActiveMembers(ctx context.Context, groupID string) (int64, error) {
	var count int64
	for k, v := range m.memberIdx {
		if len(k) > 0 && k[:len(groupID)] == groupID && v.Status == 1 {
			count++
		}
	}
	return count, nil
}

func (m *mockRepository) CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error) {
	var count int64
	for k, v := range m.memberIdx {
		if len(k) > 0 && k[:len(groupID)] == groupID && v.Role == role {
			count++
		}
	}
	return count, nil
}

func (m *mockRepository) IsUserMember(ctx context.Context, groupID, userID string) (bool, error) {
	_, ok := m.memberIdx[groupID+":"+userID]
	return ok, nil
}

// ========== PayConfig 实现 ==========

func (m *mockRepository) GetPayConfigByGroupID(ctx context.Context, groupID string) (*GroupPayConfig, error) {
	return m.payConfigs[groupID], nil
}

func (m *mockRepository) CreatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	m.payConfigs[config.GroupID] = config
	return nil
}

func (m *mockRepository) UpdatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	m.payConfigs[config.GroupID] = config
	return nil
}

// ========== Stats 实现 ==========

func (m *mockRepository) GetOrCreateStats(ctx context.Context, groupID string) (*GroupStats, error) {
	if s, ok := m.stats[groupID]; ok {
		return s, nil
	}
	s := &GroupStats{GroupID: groupID}
	m.stats[groupID] = s
	return s, nil
}

func (m *mockRepository) UpdateStats(ctx context.Context, stats *GroupStats) error {
	m.stats[stats.GroupID] = stats
	return nil
}
