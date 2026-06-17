// repository_gorm.go 群组域 Repository 接口的 GORM 数据库实现
// 使用 GORM ORM 框架操作 SQLite / MySQL 等关系数据库
// 负责群组、成员、付费配置、统计四张表的 CRUD 操作
// 不负责业务逻辑校验（由 Service 层完成），不负责缓存
//
// 相关文档：
// - PRD 社交域 MVP-P0 Step 8：Repository 接口 GORM 实现

package group

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormRepository Repository 接口的 GORM 实现
// 持有 *gorm.DB 实例，所有方法通过此实例执行数据库操作
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository 创建 GORM Repository 实例
// 参数 db 为已初始化的 GORM 数据库连接（含 AutoMigrate）
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// ========== Group 群组 CRUD 实现 ==========

// CreateGroup 创建群组记录，插入 groups 表
func (r *GormRepository) CreateGroup(ctx context.Context, group *Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据 ID 查询单条群组记录
// 记录不存在时返回 nil, nil（不返回 gorm.ErrRecordNotFound）
func (r *GormRepository) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	var group Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询群组失败(id=%s): %w", id, err)
	}
	return &group, nil
}

// GetGroupBySlug 根据 slug 查询群组（用于 URL 友好访问场景）
func (r *GormRepository) GetGroupBySlug(ctx context.Context, slug string) (*Group, error) {
	var group Group
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&group).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询群组失败(slug=%s): %w", slug, err)
	}
	return &group, nil
}

// UpdateGroup 更新群组基础信息，使用 Select("*") 确保零值字段（如 false/0）也能被更新
func (r *GormRepository) UpdateGroup(ctx context.Context, group *Group) error {
	result := r.db.WithContext(ctx).
		Model(&Group{}).
		Select("*").
		Where("id = ?", group.ID).
		Updates(group)
	if result.Error != nil {
		return fmt.Errorf("更新群组失败(id=%s): %w", group.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("更新群组失败: 群组不存在(id=%s)", group.ID)
	}
	return nil
}

// DeleteGroup 删除群组记录（软删除或物理删除取决于模型配置）
func (r *GormRepository) DeleteGroup(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Group{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("删除群组失败(id=%s): %w", id, result.Error)
	}
	return nil
}

// ListGroups 分页查询群组列表，支持动态 filters 条件筛选
// filters map 的 key 对应 Group 结构体字段名（snake_case 或 gorm column name）
// 返回当前页数据、总数、错误
func (r *GormRepository) ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*Group, int64, error) {
	var groups []*Group
	var total int64

	query := r.db.WithContext(ctx).Model(&Group{})

	// 遍历 filters 动态构建 WHERE 条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// 先查总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询群组总数失败: %w", err)
	}

	// 分页查询：page 从 1 开始，offset = (page-1)*size
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("查询群组列表失败: %w", err)
	}

	return groups, total, nil
}

// ListGroupsByOwnerID 分页查询某用户创建的所有群组
func (r *GormRepository) ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*Group, int64, error) {
	var groups []*Group
	var total int64

	query := r.db.WithContext(ctx).Model(&Group{}).Where("owner_id = ?", ownerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户群组总数失败(owner=%s): %w", ownerID, err)
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户群组列表失败(owner=%s): %w", ownerID, err)
	}

	return groups, total, nil
}

// ========== GroupMember 成员关系操作实现 ==========

// CreateMember 创建成员关系记录，插入 group_members 表
func (r *GormRepository) CreateMember(ctx context.Context, member *GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember 根据群组 ID + 用户 ID 查询成员关系（权限判断核心查询）
// 用于 CanReadTopic / CanManageMember / CanPublishTopic 等 1级数据判断
func (r *GormRepository) GetMember(ctx context.Context, groupID, userID string) (*GroupMember, error) {
	var member GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询成员关系失败(group=%s,user=%s): %w", groupID, userID, err)
	}
	return &member, nil
}

// GetMemberByID 根据成员记录主键查询单条成员记录
func (r *GormRepository) GetMemberByID(ctx context.Context, id string) (*GroupMember, error) {
	var member GroupMember
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询成员记录失败(id=%s): %w", id, err)
	}
	return &member, nil
}

// UpdateMember 更新成员关系字段（状态变更、角色变更等），使用 Select("*") 确保零值字段也能更新
func (r *GormRepository) UpdateMember(ctx context.Context, member *GroupMember) error {
	result := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Select("*").
		Where("id = ?", member.ID).
		Updates(member)
	if result.Error != nil {
		return fmt.Errorf("更新成员记录失败(id=%s): %w", member.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("更新成员记录失败: 成员记录不存在(id=%s)", member.ID)
	}
	return nil
}

// DeleteMember 删除成员关系记录
func (r *GormRepository) DeleteMember(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&GroupMember{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("删除成员记录失败(id=%s): %w", id, result.Error)
	}
	return nil
}

// ListMembers 分页查询群组成员列表，支持按角色和状态可选过滤
// role/status 为 nil 时不加过滤条件；非 nil 时追加 WHERE 子句
func (r *GormRepository) ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*GroupMember, int64, error) {
	var members []*GroupMember
	var total int64

	query := r.db.WithContext(ctx).Model(&GroupMember{}).Where("group_id = ?", groupID)

	if role != nil {
		query = query.Where("role = ?", *role)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询成员总数失败(group=%s): %w", groupID, err)
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("joined_at ASC").Find(&members).Error; err != nil {
		return nil, 0, fmt.Errorf("查询成员列表失败(group=%s): %w", groupID, err)
	}

	return members, total, nil
}

// ListMembersByUserID 分页查询某用户加入的所有群组成员关系（我的圈子列表）
func (r *GormRepository) ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*GroupMember, int64, error) {
	var members []*GroupMember
	var total int64

	query := r.db.WithContext(ctx).Model(&GroupMember{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户成员关系总数(user=%s): %w", userID, err)
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("joined_at DESC").Find(&members).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户成员关系列表失败(user=%s): %w", userID, err)
	}

	return members, total, nil
}

// CountActiveMembers 统计群组活跃成员数（status=1 正常状态的成员数量）
func (r *GormRepository) CountActiveMembers(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND status = ?", groupID, GroupMemberStatusActive).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计活跃成员数失败(group=%s): %w", groupID, err)
	}
	return count, nil
}

// CountMembersByRole 按角色统计成员数量（如管理员数量上限校验）
func (r *GormRepository) CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND role = ?", groupID, role).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("按角色统计成员数失败(group=%s,role=%d): %w", groupID, role, err)
	}
	return count, nil
}

// IsUserMember 检查用户是否是某群组的活跃成员（快速判断，不返回完整记录）
// 存在且 status=1 返回 true，否则返回 false
func (r *GormRepository) IsUserMember(ctx context.Context, groupID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, userID, GroupMemberStatusActive).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查成员身份失败(group=%s,user=%s): %w", groupID, userID, err)
	}
	return count > 0, nil
}

// ========== GroupPayConfig 付费配置操作实现 ==========

// GetPayConfigByGroupID 获取群组的付费配置
func (r *GormRepository) GetPayConfigByGroupID(ctx context.Context, groupID string) (*GroupPayConfig, error) {
	var config GroupPayConfig
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询付费配置失败(group=%s): %w", groupID, err)
	}
	return &config, nil
}

// CreatePayConfig 创建群组付费配置（创建付费群时调用）
func (r *GormRepository) CreatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// UpdatePayConfig 更新群组付费配置（价格调整、启用/禁用付费），使用 Select("*") 确保零值字段也能更新
func (r *GormRepository) UpdatePayConfig(ctx context.Context, config *GroupPayConfig) error {
	result := r.db.WithContext(ctx).
		Model(&GroupPayConfig{}).
		Select("*").
		Where("group_id = ?", config.GroupID).
		Updates(config)
	if result.Error != nil {
		return fmt.Errorf("更新付费配置失败(group=%s): %w", config.GroupID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("更新付费配置失败: 配置不存在(group=%s)", config.GroupID)
	}
	return nil
}

// ========== GroupStats 统计操作实现 ==========

// GetOrCreateStats 获取或初始化群组统计记录
// 如果记录不存在则创建默认值的统计记录（2级数据允许懒初始化）
// 使用 clause.OnConflict 实现幂等的 get-or-create 语义
func (r *GormRepository) GetOrCreateStats(ctx context.Context, groupID string) (*GroupStats, error) {
	var stats GroupStats
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).First(&stats).Error
	if err == nil {
		return &stats, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询群组统计失败(group=%s): %w", groupID, err)
	}

	// 记录不存在，创建默认统计记录
	stats = GroupStats{GroupID: groupID}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&stats).Error; err != nil {
		return nil, fmt.Errorf("创建群组统计记录失败(group=%s): %w", groupID, err)
	}
	return &stats, nil
}

// UpdateStats 更新群组统计计数器（事件驱动调用），使用 Select("*") 确保零值字段也能更新
func (r *GormRepository) UpdateStats(ctx context.Context, stats *GroupStats) error {
	result := r.db.WithContext(ctx).
		Model(&GroupStats{}).
		Select("*").
		Where("group_id = ?", stats.GroupID).
		Updates(stats)
	if result.Error != nil {
		return fmt.Errorf("更新群组统计失败(group=%s): %w", stats.GroupID, result.Error)
	}
	return nil
}
