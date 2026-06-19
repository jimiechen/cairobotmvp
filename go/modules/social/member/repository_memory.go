package member

import (
	"context"
	"sync"
)

// MemoryRepository 使用内存数据结构（map）实现 Repository 接口
// 适用于单元测试、开发环境或无数据库依赖的场景
// 线程安全：所有读写操作均通过 RWMutex 保护
type MemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]*User              // key: user.ID
	userUID  map[string]*User              // key: user.UID
	username map[string]*User              // key: user.Username
	email    map[string]*User              // key: user.Email
	phone    map[string]*User              // key: user.Phone
	blocks   map[string]*MemberBlock       // key: "blockerID:blockedID"
	stats    map[string]*MemberStats       // key: stats.UserID
}

// NewMemoryRepository 创建并初始化内存仓库实例
// 所有内部 map 均已分配，可直接使用
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:    make(map[string]*User),
		userUID:  make(map[string]*User),
		username: make(map[string]*User),
		email:    make(map[string]*User),
		phone:    make(map[string]*User),
		blocks:   make(map[string]*MemberBlock),
		stats:    make(map[string]*MemberStats),
	}
}

// ========== User 用户 CRUD 实现 ==========

// CreateUser 创建用户记录，同时建立 ID/UID/Username/Email/Phone 的反向索引
// 前置条件：username/email/phone 未被占用（由上层 Service 保证）
func (r *MemoryRepository) CreateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	if user.UID != "" {
		r.userUID[user.UID] = user
	}
	if user.Username != "" {
		r.username[user.Username] = user
	}
	if user.Email != "" {
		r.email[user.Email] = user
	}
	if user.Phone != "" {
		r.phone[user.Phone] = user
	}
	return nil
}

// GetUserByID 根据主键 ID 查询用户
// 未找到时返回 (nil, nil)，不返回错误
func (r *MemoryRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// GetUserByUID 根据对外编号查询用户
// 用于展示和关联查询场景
func (r *MemoryRepository) GetUserByUID(ctx context.Context, uid string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.userUID[uid]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// GetUserByUsername 根据登录用户名查询用户
// 用于登录流程（minType=1023）
func (r *MemoryRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.username[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// GetUserByEmail 根据邮箱地址查询用户
// 登录流程备用字段
func (r *MemoryRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.email[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// GetUserByPhone 根据手机号查询用户
// 登录流程备用字段
func (r *MemoryRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.phone[phone]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// UpdateUser 更新用户资料，同步更新各索引字段
// 仅当用户存在时执行更新；不存在则静默跳过
func (r *MemoryRepository) UpdateUser(ctx context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, exists := r.users[user.ID]
	if !exists {
		return nil
	}

	// 清理旧索引值（如果字段发生了变更）
	if old.Username != user.Username && old.Username != "" {
		delete(r.username, old.Username)
	}
	if old.Email != user.Email && old.Email != "" {
		delete(r.email, old.Email)
	}
	if old.Phone != user.Phone && old.Phone != "" {
		delete(r.phone, old.Phone)
	}
	if old.UID != user.UID && old.UID != "" {
		delete(r.userUID, old.UID)
	}

	// 写入新数据和新索引
	r.users[user.ID] = user
	if user.Username != "" {
		r.username[user.Username] = user
	}
	if user.Email != "" {
		r.email[user.Email] = user
	}
	if user.Phone != "" {
		r.phone[user.Phone] = user
	}
	if user.UID != "" {
		r.userUID[user.UID] = user
	}
	return nil
}

// BatchGetUsersByID 批量获取用户信息
// 用于帖子列表、评论列表等场景的作者信息批量加载
// 只返回存在的用户，不存在的 ID 静默跳过
func (r *MemoryRepository) BatchGetUsersByID(ctx context.Context, ids []string) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*User, 0, len(ids))
	for _, id := range ids {
		if u, ok := r.users[id]; ok {
			result = append(result, u)
		}
	}
	return result, nil
}

// ========== MemberBlock 拉黑操作实现 ==========

// CreateBlock 创建拉黑关系，使用复合键 "blockerID:blockedID" 保证唯一性
// 同一 blocker→blocked 对只能存在一条记录
func (r *MemoryRepository) CreateBlock(ctx context.Context, block *MemberBlock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := block.BlockerID + ":" + block.BlockedID
	r.blocks[key] = block
	return nil
}

// DeleteBlock 删除拉黑关系
// 不存在时静默返回成功
func (r *MemoryRepository) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := blockerID + ":" + blockedID
	delete(r.blocks, key)
	return nil
}

// ListBlocks 分页查询某用户的拉黑列表
// page 从 1 开始，size 为每页条数
// 返回当前页数据、总记录数和错误信息
func (r *MemoryRepository) ListBlocks(ctx context.Context, blockerID string, page, size int) ([]*MemberBlock, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := blockerID + ":"
	var matched []*MemberBlock
	for key, b := range r.blocks {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			matched = append(matched, b)
		}
	}

	total := int64(len(matched))
	if total == 0 {
		return []*MemberBlock{}, 0, nil
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if offset >= len(matched) {
		return []*MemberBlock{}, total, nil
	}

	end := offset + size
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

// IsBlocked 检查 blockerID 是否拉黑了 blockedID
// 用于内容过滤和权限判断
func (r *MemoryRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := blockerID + ":" + blockedID
	_, ok := r.blocks[key]
	return ok, nil
}

// GetBlockCount 获取指定用户的拉黑人数统计
func (r *MemoryRepository) GetBlockCount(ctx context.Context, blockerID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := blockerID + ":"
	count := int64(0)
	for key := range r.blocks {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			count++
		}
	}
	return count, nil
}

// ========== MemberStats 统计操作实现 ==========

// GetOrCreateStats 获取用户统计记录，若不存在则创建默认值
// 2级数据允许懒初始化，用于发帖/评论等事件前的统计读取
func (r *MemoryRepository) GetOrCreateStats(ctx context.Context, userID string) (*MemberStats, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.stats[userID]
	if ok {
		return s, nil
	}

	// 不存在时创建默认统计记录
	defaultStats := &MemberStats{
		UserID: userID,
	}
	r.stats[userID] = defaultStats
	return defaultStats, nil
}

// UpdateStats 更新用户统计计数器（全量覆盖写入）
// 由事件驱动调用，用于发帖/评论/点赞等事件的增量更新
func (r *MemoryRepository) UpdateStats(ctx context.Context, stats *MemberStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats[stats.UserID] = stats
	return nil
}
