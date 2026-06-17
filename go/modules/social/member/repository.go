package member

import (
	"context"
)

// Repository 成员域数据库操作接口定义
// 覆盖成员协议组（maxType=1000）全部 MVP 白名单协议的数据访问需求
// GORM 实现放在 repository_gorm.go
type Repository interface {
	// ========== User 用户 CRUD（对应协议 minType 1021-1032）==========

	// CreateUser 创建用户记录，用于注册流程（minType=1021 UserRegister）
	// 前置条件：username/email/phone 未被占用
	CreateUser(ctx context.Context, user *User) error

	// GetUserByID 根据 ID 查询用户信息
	GetUserByID(ctx context.Context, id string) (*User, error)

	// GetUserByUID 根据对外编号查询用户（用于展示和关联查询）
	GetUserByUID(ctx context.Context, uid string) (*User, error)

	// GetUserByUsername 根据登录用户名查询（用于登录流程 minType=1023）
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByEmail 根据邮箱查询（登录流程备用字段）
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// GetUserByPhone 根据手机号查询（登录流程备用字段）
	GetUserByPhone(ctx context.Context, phone string) (*User, error)

	// UpdateUser 更新用户资料（minType=1031 UpdateUserInfo）
	UpdateUser(ctx context.Context, user *User) error

	// BatchGetUsersByID 批量获取用户信息（minType=1049 BatchGetUserInfo）
	// 用于帖子列表、评论列表等场景的作者信息批量加载
	BatchGetUsersByID(ctx context.Context, ids []string) ([]*User, error)

	// ========== MemberBlock 拉黑操作（对应协议 minType 1039-1048）==========

	// CreateBlock 创建拉黑关系（minType=1039 BlockUser）
	// 同一 blocker→blocked 对只能存在一条记录
	CreateBlock(ctx context.Context, block *MemberBlock) error

	// DeleteBlock 删除拉黑关系（minType=1041 UnblockUser）
	DeleteBlock(ctx context.Context, blockerID, blockedID string) error

	// ListBlocks 分页查询某用户的拉黑列表（minType=1043 GetBlockList）
	// 返回拉黑记录列表、总数、错误
	ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*MemberBlock, int64, error)

	// IsBlocked 检查 A 是否拉黑了 B（用于内容过滤和权限判断）
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)

	// GetBlockCount 获取用户的拉黑人数（minType=1047 GetBlockCount）
	GetBlockCount(ctx context.Context, blockerID string) (int64, error)

	// ========== MemberStats 统计操作（对应协议 minType 1045-1046）==========

	// GetOrCreateStats 获取或初始化用户统计记录（minType=1045 GetUserStats）
	// 如果记录不存在则创建默认值（2级数据允许懒初始化）
	GetOrCreateStats(ctx context.Context, userID string) (*MemberStats, error)

	// UpdateStats 更新用户统计计数器（事件驱动调用）
	// 仅更新传入的非零字段，用于发帖/评论/点赞等事件的增量更新
	UpdateStats(ctx context.Context, stats *MemberStats) error
}
