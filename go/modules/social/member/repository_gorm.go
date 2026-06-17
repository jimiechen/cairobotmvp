// repository_gorm.go — Repository 接口的 GORM 数据库实现
// 职责：将 Repository 接口定义的全部 17 个方法映射为 GORM 数据库操作
// 不负责：业务校验（如用户名唯一性）、缓存、事务编排（由 svc 层处理）
// 依赖：gorm.io/gorm，使用 SQLite/MySQL 等方言

package member

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// GormRepository Repository 接口的 GORM 实现
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository 创建基于 GORM 的成员域仓库实例
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// ========== User 用户 CRUD ==========

// CreateUser 创建用户记录，ID 由调用方预填充
func (r *GormRepository) CreateUser(ctx context.Context, user *User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return fmt.Errorf("CreateUser(id=%s): %w", user.ID, err)
	}
	return nil
}

// GetUserByID 根据 ID 查询用户，不存在返回 (nil, nil)
func (r *GormRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserByID(id=%s): %w", id, err)
	}
	return &user, nil
}

// GetUserByUID 根据对外编号查询用户，不存在返回 (nil, nil)
func (r *GormRepository) GetUserByUID(ctx context.Context, uid string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserByUID(uid=%s): %w", uid, err)
	}
	return &user, nil
}

// GetUserByUsername 根据登录用户名查询用户，不存在返回 (nil, nil)
func (r *GormRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserByUsername(username=%s): %w", username, err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱查询用户，不存在返回 (nil, nil)
func (r *GormRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserByEmail(email=%s): %w", email, err)
	}
	return &user, nil
}

// GetUserByPhone 根据手机号查询用户，不存在返回 (nil, nil)
func (r *GormRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("GetUserByPhone(phone=%s): %w", phone, err)
	}
	return &user, nil
}

// UpdateUser 更新用户资料，仅更新指定的可变字段
func (r *GormRepository) UpdateUser(ctx context.Context, user *User) error {
	err := r.db.WithContext(ctx).
		Model(&User{ID: user.ID}).
		Updates(map[string]interface{}{
			"nickname":         user.Nickname,
			"avatar":           user.Avatar,
			"status":           user.Status,
			"membership_level": user.MembershipLevel,
			"last_login_at":    user.LastLoginAt,
			"last_login_ip":    user.LastLoginIP,
			"login_count":      user.LoginCount,
			"updated_at":       user.UpdatedAt,
		}).Error
	if err != nil {
		return fmt.Errorf("UpdateUser(id=%s): %w", user.ID, err)
	}
	return nil
}

// BatchGetUsersByID 批量获取用户信息，按 ID 列表查询
func (r *GormRepository) BatchGetUsersByID(ctx context.Context, ids []string) ([]*User, error) {
	var users []*User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("BatchGetUsersByID(count=%d): %w", len(ids), err)
	}
	return users, nil
}

// ========== MemberBlock 拉黑操作 ==========

// CreateBlock 创建拉黑关系记录
func (r *GormRepository) CreateBlock(ctx context.Context, block *MemberBlock) error {
	err := r.db.WithContext(ctx).Create(block).Error
	if err != nil {
		return fmt.Errorf("CreateBlock(blocker=%s, blocked=%s): %w", block.BlockerID, block.BlockedID, err)
	}
	return nil
}

// DeleteBlock 删除指定拉黑关系
func (r *GormRepository) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	err := r.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Delete(&MemberBlock{}).Error
	if err != nil {
		return fmt.Errorf("DeleteBlock(blocker=%s, blocked=%s): %w", blockerID, blockedID, err)
	}
	return nil
}

// ListBlocks 分页查询某用户的拉黑列表，返回记录列表和总数
func (r *GormRepository) ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*MemberBlock, int64, error) {
	var blocks []*MemberBlock
	var total int64

	query := r.db.WithContext(ctx).Model(&MemberBlock{}).Where("blocker_id = ?", blockerID)

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ListBlocks(blocker=%s) count: %w", blockerID, err)
	}

	// 分页查询
	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&blocks).Error
	if err != nil {
		return nil, 0, fmt.Errorf("ListBlocks(blocker=%s, page=%d, size=%d): %w", blockerID, page, size, err)
	}

	return blocks, total, nil
}

// IsBlocked 检查 blockerID 是否已拉黑 blockedID
func (r *GormRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	var block MemberBlock
	err := r.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		First(&block).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("IsBlocked(blocker=%s, blocked=%s): %w", blockerID, blockedID, err)
	}
	return true, nil
}

// GetBlockCount 获取用户的拉黑人数
func (r *GormRepository) GetBlockCount(ctx context.Context, blockerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&MemberBlock{}).
		Where("blocker_id = ?", blockerID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("GetBlockCount(blocker=%s): %w", blockerID, err)
	}
	return count, nil
}

// ========== MemberStats 统计操作 ==========

// GetOrCreateStats 获取或初始化用户统计记录，不存在则创建默认值
func (r *GormRepository) GetOrCreateStats(ctx context.Context, userID string) (*MemberStats, error) {
	var stats MemberStats
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 记录不存在，创建默认统计
			newStats := &MemberStats{
				UserID:    userID,
				UpdatedAt: 0,
			}
			if createErr := r.db.WithContext(ctx).Create(newStats).Error; createErr != nil {
				return nil, fmt.Errorf("GetOrCreateStats(user_id=%s) create: %w", userID, createErr)
			}
			return newStats, nil
		}
		return nil, fmt.Errorf("GetOrCreateStats(user_id=%s): %w", userID, err)
	}
	return &stats, nil
}

// UpdateStats 更新用户统计计数器，仅更新指定的计数字段
func (r *GormRepository) UpdateStats(ctx context.Context, stats *MemberStats) error {
	err := r.db.WithContext(ctx).
		Model(&MemberStats{UserID: stats.UserID}).
		Updates(map[string]interface{}{
			"topics_count":   stats.TopicsCount,
			"replies_count":  stats.RepliesCount,
			"likes_received": stats.LikesReceived,
			"groups_joined":  stats.GroupsJoined,
			"updated_at":     stats.UpdatedAt,
		}).Error
	if err != nil {
		return fmt.Errorf("UpdateStats(user_id=%s): %w", stats.UserID, err)
	}
	return nil
}
