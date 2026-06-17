package group

import (
	"context"
)

// Repository 群组域数据库操作接口定义
// 覆盖群组协议组（maxType=2000）全部 MVP 白名单协议的数据访问需求
// GORM 实现放在 repository_gorm.go
type Repository interface {
	// ========== Group 群组 CRUD（对应协议 minType 2005-2012）==========

	// CreateGroup 创建群组记录（minType=2005 CreateGroup）
	// 创建后需自动将创建者加入 group_members（role=owner）
	CreateGroup(ctx context.Context, group *Group) error

	// GetGroupByID 根据 ID 查询群组详情
	GetGroupByID(ctx context.Context, id string) (*Group, error)

	// GetGroupBySlug 根据 slug 查询群组（用于 URL 访问场景）
	GetGroupBySlug(ctx context.Context, slug string) (*Group, error)

	// UpdateGroup 更新群组基础信息（minType=2009 UpdateGroup）
	UpdateGroup(ctx context.Context, group *Group) error

	// DeleteGroup 解散/删除群组（minType=2011 DeleteGroup，高权限操作）
	DeleteGroup(ctx context.Context, id string) error

	// ListGroups 分页查询群组列表（支持按状态、分类、官方等条件筛选）
	// 返回群组列表、总数、错误
	ListGroups(ctx context.Context, page, size int, filters map[string]interface{}) ([]*Group, int64, error)

	// ListGroupsByOwnerID 分页查询某用户创建的群组列表
	ListGroupsByOwnerID(ctx context.Context, ownerID string, page, size int) ([]*Group, int64, error)

	// ========== GroupMember 成员关系操作（对应协议 minType 2013-2038）==========

	// CreateMember 创建成员关系记录（minType=2013 JoinGroup）
	// 前置条件：未达到 max_members 上限、用户尚未加入该群组
	CreateMember(ctx context.Context, member *GroupMember) error

	// GetMember 根据群组 ID + 用户 ID 查询成员关系（权限判断核心查询）
	// 用于 CanReadTopic / CanManageMember / CanPublishTopic 等 1级数据判断
	GetMember(ctx context.Context, groupID, userID string) (*GroupMember, error)

	// GetMemberByID 根据成员记录主键查询单条记录
	GetMemberByID(ctx context.Context, id string) (*GroupMember, error)

	// UpdateMember 更新成员关系字段（状态变更、角色变更等）
	UpdateMember(ctx context.Context, member *GroupMember) error

	// DeleteMember 删除成员关系（minType=2015 LeaveGroup）
	DeleteMember(ctx context.Context, id string) error

	// ListMembers 分页查询群组成员列表（圈主管理成员页 minType=2013+action=query）
	// 支持按角色、状态筛选；返回成员列表、总数、错误
	ListMembers(ctx context.Context, groupID string, page, size int, role, status *int8) ([]*GroupMember, int64, error)

	// ListMembersByUserID 分页查询某用户加入的所有群组（我的圈子列表）
	ListMembersByUserID(ctx context.Context, userID string, page, size int) ([]*GroupMember, int64, error)

	// CountActiveMembers 统计群组活跃成员数（status=active 的成员数量）
	CountActiveMembers(ctx context.Context, groupID string) (int64, error)

	// CountMembersByRole 按角色统计成员数量（如管理员数量校验）
	CountMembersByRole(ctx context.Context, groupID string, role int8) (int64, error)

	// IsUserMember 检查用户是否是某群组的活跃成员（快速判断，不返回完整记录）
	IsUserMember(ctx context.Context, groupID, userID string) (bool, error)

	// ========== GroupPayConfig 付费配置操作 ==========

	// GetPayConfigByGroupID 获取群组的付费配置
	GetPayConfigByGroupID(ctx context.Context, groupID string) (*GroupPayConfig, error)

	// CreatePayConfig 创建群组付费配置（创建付费群时调用）
	CreatePayConfig(ctx context.Context, config *GroupPayConfig) error

	// UpdatePayConfig 更新群组付费配置（价格调整、启用/禁用付费）
	UpdatePayConfig(ctx context.Context, config *GroupPayConfig) error

	// ========== GroupStats 统计操作（对应协议 minType=2041）==========

	// GetOrCreateStats 获取或初始化群组统计记录（minType=2041 RefreshGroupStat）
	// 如果记录不存在则创建默认值（2级数据允许懒初始化）
	GetOrCreateStats(ctx context.Context, groupID string) (*GroupStats, error)

	// UpdateStats 更新群组统计计数器（事件驱动调用）
	// 入群/退群/发帖事件触发增量更新
	UpdateStats(ctx context.Context, stats *GroupStats) error
}
