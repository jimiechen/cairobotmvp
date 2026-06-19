package group

import (
	"context"
	"sync"
)

// MemoryRepository 群组域内存仓库实现
// 使用 map + sync.RWMutex 提供线程安全的内存数据访问
// 适用于单元测试、开发调试等不需要持久化数据库的场景
type MemoryRepository struct {
	mu         sync.RWMutex
	groups     map[string]*Group              // key: group.ID
	groupsBySlug map[string]*Group            // key: group.Slug
	members    map[string]*GroupMember        // key: member.ID
	membersByKey map[string]*GroupMember      // key: "groupID:userID"
	payConfigs map[string]*GroupPayConfig     // key: config.GroupID
	stats      map[string]*GroupStats         // key: stats.GroupID
}

// NewMemoryRepository 创建并返回一个新的内存仓库实例
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		groups:       make(map[string]*Group),
		groupsBySlug: make(map[string]*Group),
		members:      make(map[string]*GroupMember),
		membersByKey: make(map[string]*GroupMember),
		payConfigs:   make(map[string]*GroupPayConfig),
		stats:        make(map[string]*GroupStats),
	}
}

// ========== Group 群组 CRUD 实现 ==========

// CreateGroup 创建群组记录，同时建立 slug 索引
func (r *MemoryRepository) CreateGroup(ctx context.Context, group *Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.groups[group.ID] = group
	r.groupsBySlug[group.Slug] = group
	return nil
}

// GetGroupByID 根据 ID 查询群组详情，未找到时返回 (nil, nil)
func (r *MemoryRepository) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.groups[id]
	if !ok {
		return nil, nil
	}
	return g, nil
}

// GetGroupBySlug 根据 slug 查询群组，未找到时返回 (nil, nil)
func (r *MemoryRepository) GetGroupBySlug(ctx context.Context, slug string) (*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.groupsBySlug[slug]
	if !ok {
		return nil, nil
	}
	return g, nil
}

// UpdateGroup 更新群组基础信息，同步更新 slug 索引
func (r *MemoryRepository) UpdateGroup(ctx context.Context, group *Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 删除旧的 slug 索引（如果 slug 发生变更）
	if old, ok := r.groups[group.ID]; ok && old.Slug != group.Slug {
		delete(r.groupsBySlug, old.Slug)
	}
	r.groups[group.ID] = group
	r.groupsBySlug[group.Slug] = group
	return nil
}

// DeleteGroup 解散/删除群组，同时清理 slug 索引
func (r *MemoryRepository) DeleteGroup(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[id]
	if !ok {
		return nil
	}
	delete(r.groups, id)
	delete(r.groupsBySlug, g.Slug)
	return nil
}

// ListGroups 分页查询群组列表，支持按 status / category / is_official 等条件筛选
func (r *MemoryRepository) ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*Group, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Group
	for _, g := range r.groups {
		if matchFilters(g, filters) {
			result = append(result, g)
		}
	}

	total := int64(len(result))
	start, end := paginate(page, size, total)
	if start >= len(result) {
		return []*Group{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// ListGroupsByOwnerID 分页查询某用户创建的群组列表
func (r *MemoryRepository) ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*Group, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Group
	for _, g := range r.groups {
		if g.OwnerID == ownerID {
			result = append(result, g)
		}
	}

	total := int64(len(result))
	start, end := paginate(page, size, total)
	if start >= len(result) {
		return []*Group{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// ========== GroupMember 成员关系操作实现 ==========

// CreateMember 创建成员关系记录，建立 ID 索引和复合键索引
func (r *MemoryRepository) CreateMember(ctx context.Context, member *GroupMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[member.ID] = member
	compositeKey := member.GroupID + ":" + member.UserID
	r.membersByKey[compositeKey] = member
	return nil
}

// GetMember 根据群组 ID + 用户 ID 查询成员关系（权限判断核心查询）
func (r *MemoryRepository) GetMember(ctx context.Context, groupID, userID string) (*GroupMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	compositeKey := groupID + ":" + userID
	m, ok := r.membersByKey[compositeKey]
	if !ok {
		return nil, nil
	}
	return m, nil
}

// GetMemberByID 根据成员记录主键查询单条记录
func (r *MemoryRepository) GetMemberByID(ctx context.Context, id string) (*GroupMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.members[id]
	if !ok {
		return nil, nil
	}
	return m, nil
}

// UpdateMember 更新成员关系字段，同步更新复合键索引
func (r *MemoryRepository) UpdateMember(ctx context.Context, member *GroupMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[member.ID] = member
	compositeKey := member.GroupID + ":" + member.UserID
	r.membersByKey[compositeKey] = member
	return nil
}

// DeleteMember 删除成员关系，同步清理复合键索引
func (r *MemoryRepository) DeleteMember(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil
	}
	compositeKey := m.GroupID + ":" + m.UserID
	delete(r.members, id)
	delete(r.membersByKey, compositeKey)
	return nil
}

// ListMembers 分页查询群组成员列表，支持按角色、状态筛选
func (r *MemoryRepository) ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*GroupMember, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*GroupMember
	for _, m := range r.members {
		if m.GroupID != groupID {
			continue
		}
		if role != nil && m.Role != *role {
			continue
		}
		if status != nil && m.Status != *status {
			continue
		}
		result = append(result, m)
	}

	total := int64(len(result))
	start, end := paginate(page, size, total)
	if start >= len(result) {
		return []*GroupMember{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// ListMembersByUserID 分页查询某用户加入的所有群组
func (r *MemoryRepository) ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*GroupMember, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*GroupMember
	for _, m := range r.members {
		if m.UserID == userID {
			result = append(result, m)
		}
	}

	total := int64(len(result))
	start, end := paginate(page, size, total)
	if start >= len(result) {
		return []*GroupMember{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// CountActiveMembers 统计群组活跃成员数（status=1 的成员数量）
func (r *MemoryRepository) CountActiveMembers(ctx context.Context, groupID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	const statusActive int8 = 1
	var count int64
	for _, m := range r.members {
		if m.GroupID == groupID && m.Status == statusActive {
			count++
		}
	}
	return count, nil
}

// CountMembersByRole 按角色统计成员数量
func (r *MemoryRepository) CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, m := range r.members {
		if m.GroupID == groupID && m.Role == role {
			count++
		}
	}
	return count, nil
}

// IsUserMember 检查用户是否是某群组的活跃成员（快速判断）
func (r *MemoryRepository) IsUserMember(ctx context.Context, groupID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	compositeKey := groupID + ":" + userID
	m, ok := r.membersByKey[compositeKey]
	if !ok {
		return false, nil
	}
	const statusActive int8 = 1
	return m.Status == statusActive, nil
}

// ========== GroupPayConfig 付费配置操作实现 ==========

// GetPayConfigByGroupID 获取群组的付费配置，未找到时返回 (nil, nil)
func (r *MemoryRepository) GetPayConfigByGroupID(ctx context.Context, groupID string) (*GroupPayConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.payConfigs[groupID]
	if !ok {
		return nil, nil
	}
	return c, nil
}

// CreatePayConfig 创建群组付费配置
func (r *MemoryRepository) CreatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payConfigs[config.GroupID] = config
	return nil
}

// UpdatePayConfig 更新群组付费配置
func (r *MemoryRepository) UpdatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payConfigs[config.GroupID] = config
	return nil
}

// ========== GroupStats 统计操作实现 ==========

// GetOrCreateStats 获取或初始化群组统计记录
// 如果记录不存在则创建默认值（2级数据允许懒初始化）
func (r *MemoryRepository) GetOrCreateStats(ctx context.Context, groupID string) (*GroupStats, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.stats[groupID]
	if !ok {
		s = &GroupStats{
			GroupID: groupID,
		}
		r.stats[groupID] = s
	}
	return s, nil
}

// UpdateStats 更新群组统计计数器
func (r *MemoryRepository) UpdateStats(ctx context.Context, stats *GroupStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats[stats.GroupID] = stats
	return nil
}

// ========== 内部辅助方法 ==========

// matchFilters 检查群组对象是否匹配所有筛选条件
// 支持 status(int8)、category(string)、is_official(bool) 等常见过滤维度
func matchFilters(g *Group, filters map[string]interface{}) bool {
	if filters == nil || len(filters) == 0 {
		return true
	}
	for k, v := range filters {
		switch k {
		case "status":
			if s, ok := v.(int8); ok && g.Status != s {
				return false
			}
		case "category":
			if c, ok := v.(string); ok && g.Category != c {
				return false
			}
		case "is_official":
			if o, ok := v.(bool); ok && g.IsOfficial != o {
				return false
			}
		case "visibility":
			if vis, ok := v.(int8); ok && g.Visibility != vis {
				return false
			}
		}
	}
	return true
}

// paginate 计算分页的起始和结束偏移量
// 返回 (start, end) 索引范围，用于切片截取
func paginate(page, size int, total int64) (start, end int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	start = (page - 1) * size
	end = start + size
	return
}
