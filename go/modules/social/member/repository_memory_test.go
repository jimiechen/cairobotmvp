// repository_memory_test.go — MemoryRepository 单元测试
// 覆盖 MemoryRepository 全部 Repository 接口方法：
//   - User CRUD（创建、多索引查询、更新、批量查询）
//   - MemberBlock（创建、删除、分页、检查、计数）
//   - MemberStats（懒初始化、更新）
//   - 边界场景（空 ID、初始化状态）

package member

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMemoryRepo 创建用于测试的 MemoryRepository 实例
// 每个测试独立调用，确保测试间无状态污染
func newTestMemoryRepo(t *testing.T) *MemoryRepository {
	t.Helper()
	return NewMemoryRepository()
}

// newTestUser 创建标准测试用户，减少各测试中的重复构造代码
func newTestUser(id string) *User {
	return &User{
		ID:              id,
		UID:             fmt.Sprintf("10000000%s", id),
		Username:        fmt.Sprintf("user_%s", id),
		Password:        "hashed_pwd",
		Email:           fmt.Sprintf("user_%s@test.com", id),
		Phone:           fmt.Sprintf("1380001%s", id),
		Nickname:        fmt.Sprintf("Nick_%s", id),
		Status:          1,
		MembershipLevel: "normal",
		CreatedAt:       1700000000000,
		UpdatedAt:       1700000000000,
	}
}

// ========== User CRUD 测试 ==========

func TestCreateUser_成功创建并建立所有索引(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u001")
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// 通过 ID 索引查回
	gotByID, err := repo.GetUserByID(ctx, "u001")
	require.NoError(t, err)
	require.NotNil(t, gotByID)
	assert.Equal(t, "u001", gotByID.ID)

	// 通过 UID 索引查回
	gotByUID, err := repo.GetUserByUID(ctx, user.UID)
	require.NoError(t, err)
	require.NotNil(t, gotByUID)
	assert.Equal(t, "u001", gotByUID.ID)

	// 通过 Username 索引查回
	gotByName, err := repo.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.NotNil(t, gotByName)
	assert.Equal(t, "u001", gotByName.ID)

	// 通过 Email 索引查回
	gotByEmail, err := repo.GetUserByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, gotByEmail)
	assert.Equal(t, "u001", gotByEmail.ID)

	// 通过 Phone 索引查回
	gotByPhone, err := repo.GetUserByPhone(ctx, user.Phone)
	require.NoError(t, err)
	require.NotNil(t, gotByPhone)
	assert.Equal(t, "u001", gotByPhone.ID)
}

func TestGetUserByID_存在返回用户(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u002")
	repo.CreateUser(ctx, user)

	got, err := repo.GetUserByID(ctx, "u002")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u002", got.ID)
	assert.Equal(t, "Nick_u002", got.Nickname)
}

func TestGetUserByID_不存在返回错误(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	got, err := repo.GetUserByID(ctx, "nonexistent_id")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetUserByUID_正确查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u003")
	repo.CreateUser(ctx, user)

	got, err := repo.GetUserByUID(ctx, user.UID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u003", got.ID)
	assert.Equal(t, user.UID, got.UID)

	// 查不存在的 UID
	gotNil, err := repo.GetUserByUID(ctx, "999999999")
	require.NoError(t, err)
	assert.Nil(t, gotNil)
}

func TestGetUserByUsername_正确查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u004")
	repo.CreateUser(ctx, user)

	got, err := repo.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u004", got.ID)

	// 查不存在的 username
	gotNil, err := repo.GetUserByUsername(ctx, "no_such_user")
	require.NoError(t, err)
	assert.Nil(t, gotNil)
}

func TestGetUserByEmail_正确查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u005")
	repo.CreateUser(ctx, user)

	got, err := repo.GetUserByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u005", got.ID)

	// 查不存在的 email
	gotNil, err := repo.GetUserByEmail(ctx, "noone@nowhere.com")
	require.NoError(t, err)
	assert.Nil(t, gotNil)
}

func TestGetUserByPhone_正确查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u006")
	repo.CreateUser(ctx, user)

	got, err := repo.GetUserByPhone(ctx, user.Phone)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u006", got.ID)

	// 查不存在的 phone
	gotNil, err := repo.GetUserByPhone(ctx, "19999999999")
	require.NoError(t, err)
	assert.Nil(t, gotNil)
}

func TestUpdateUser_更新字段(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	user := newTestUser("u007")
	repo.CreateUser(ctx, user)

	// 修改昵称和头像
	user.Nickname = "UpdatedNick"
	user.Avatar = "https://example.com/new_avatar.png"
	user.UpdatedAt = 1700100000000
	err := repo.UpdateUser(ctx, user)
	require.NoError(t, err)

	// 验证 ID 索引已更新
	got, err := repo.GetUserByID(ctx, "u007")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "UpdatedNick", got.Nickname)
	assert.Equal(t, "https://example.com/new_avatar.png", got.Avatar)
	assert.Equal(t, int64(1700100000000), got.UpdatedAt)

	// 验证 Username 索引仍可访问
	gotByName, err := repo.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.NotNil(t, gotByName)
	assert.Equal(t, "UpdatedNick", gotByName.Nickname)
}

func TestBatchGetUsersByID_批量查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	// 创建 3 个用户
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("batch_u%d", i)
		user := newTestUser(id)
		repo.CreateUser(ctx, user)
	}

	// 批量查询：2 个存在 + 1 个不存在
	users, err := repo.BatchGetUsersByID(ctx, []string{"batch_u0", "batch_u2", "nonexistent"})
	require.NoError(t, err)
	// 只返回找到的用户，不存在的静默跳过
	assert.Len(t, users, 2)

	// 验证返回的是正确的用户
	ids := make(map[string]bool)
	for _, u := range users {
		ids[u.ID] = true
	}
	assert.True(t, ids["batch_u0"])
	assert.True(t, ids["batch_u2"])
	assert.False(t, ids["nonexistent"])

	// 全部不存在的场景
	usersEmpty, err := repo.BatchGetUsersByID(ctx, []string{"x", "y", "z"})
	require.NoError(t, err)
	assert.Len(t, usersEmpty, 0)
}

func TestGetUserByID_空ID返回错误(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	got, err := repo.GetUserByID(ctx, "")
	require.NoError(t, err)
	assert.Nil(t, got)
}

// ========== Block 操作测试 ==========

func TestCreateBlock_创建拉黑关系(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	block := &MemberBlock{
		ID:         "blk_001",
		BlockerID:  "blocker_a",
		BlockedID:  "blocked_b",
		Reason:     "spam",
		CreatedAt:  1700000001000,
	}
	err := repo.CreateBlock(ctx, block)
	require.NoError(t, err)

	// 通过 IsBlocked 验证关系存在
	blocked, err := repo.IsBlocked(ctx, "blocker_a", "blocked_b")
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestDeleteBlock_删除拉黑关系(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	block := &MemberBlock{
		ID:         "blk_del",
		BlockerID:  "del_user",
		BlockedID:  "del_target",
		CreatedAt:  1700000002000,
	}
	repo.CreateBlock(ctx, block)

	// 删除前确认存在
	blocked, _ := repo.IsBlocked(ctx, "del_user", "del_target")
	assert.True(t, blocked)

	// 执行删除
	err := repo.DeleteBlock(ctx, "del_user", "del_target")
	require.NoError(t, err)

	// 确认已删除
	blockedAfter, err := repo.IsBlocked(ctx, "del_user", "del_target")
	require.NoError(t, err)
	assert.False(t, blockedAfter)

	// 删除不存在的记录应静默成功（不报错）
	err = repo.DeleteBlock(ctx, "nobody", "nothing")
	require.NoError(t, err)
}

func TestIsBlocked_已拉黑返回true(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	block := &MemberBlock{
		ID:         "blk_check_true",
		BlockerID:  "checker",
		BlockedID:  "bad_guy",
		Reason:     "harassment",
		CreatedAt:  1700000003000,
	}
	repo.CreateBlock(ctx, block)

	blocked, err := repo.IsBlocked(ctx, "checker", "bad_guy")
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestIsBlocked_未拉黑返回false(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()

	// 未创建任何拉黑记录
	blocked, err := repo.IsBlocked(ctx, "user_x", "user_y")
	require.NoError(t, err)
	assert.False(t, blocked)

	// 创建 A→B 的拉黑后，B→A 应仍为 false（单向关系）
	block := &MemberBlock{
		ID:         "blk_one_way",
		BlockerID:  "user_x",
		BlockedID:  "user_y",
		CreatedAt:  1700000004000,
	}
	repo.CreateBlock(ctx, block)

	blockedXY, err := repo.IsBlocked(ctx, "user_x", "user_y")
	require.NoError(t, err)
	assert.True(t, blockedXY)

	blockedYX, err := repo.IsBlocked(ctx, "user_y", "user_x")
	require.NoError(t, err)
	assert.False(t, blockedYX)
}

func TestListBlocks_分页查询(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()
	blockerID := "pager_owner"

	// 创建 5 条拉黑记录
	for i := 0; i < 5; i++ {
		b := &MemberBlock{
			ID:         fmt.Sprintf("pager_blk_%d", i),
			BlockerID:  blockerID,
			BlockedID:  fmt.Sprintf("paged_target_%d", i),
			Reason:     fmt.Sprintf("reason_%d", i),
			CreatedAt:  int64(1700000005000 + int64(i)*1000),
		}
		repo.CreateBlock(ctx, b)
	}

	// 第 1 页：每页 2 条 → 返回 2 条，总数 5
	blocks1, total1, err := repo.ListBlocks(ctx, blockerID, 1, 2)
	require.NoError(t, err)
	assert.Len(t, blocks1, 2)
	assert.Equal(t, int64(5), total1)

	// 第 2 页：每页 2 条 → 返回 2 条
	blocks2, total2, err := repo.ListBlocks(ctx, blockerID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, blocks2, 2)
	assert.Equal(t, int64(5), total2)

	// 第 3 页：每页 2 条 → 剩余 1 条
	blocks3, total3, err := repo.ListBlocks(ctx, blockerID, 3, 2)
	require.NoError(t, err)
	assert.Len(t, blocks3, 1)
	assert.Equal(t, int64(5), total3)

	// 超出范围：第 4 页应返回空列表但总数仍为 5
	blocks4, total4, err := repo.ListBlocks(ctx, blockerID, 4, 2)
	require.NoError(t, err)
	assert.Len(t, blocks4, 0)
	assert.Equal(t, int64(5), total4)

	// 无拉黑记录的用户
	emptyBlocks, emptyTotal, err := repo.ListBlocks(ctx, "no_blocks_user", 1, 10)
	require.NoError(t, err)
	assert.Len(t, emptyBlocks, 0)
	assert.Equal(t, int64(0), emptyTotal)
}

func TestGetBlockCount_计数正确(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()
	blockerID := "count_owner"

	// 初始计数为 0
	count, err := repo.GetBlockCount(ctx, blockerID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 创建 3 条拉黑记录
	for i := 0; i < 3; i++ {
		b := &MemberBlock{
			ID:         fmt.Sprintf("cnt_blk_%d", i),
			BlockerID:  blockerID,
			BlockedID:  fmt.Sprintf("cnt_target_%d", i),
			CreatedAt:  int64(1700000006000 + int64(i)*1000),
		}
		repo.CreateBlock(ctx, b)
	}

	count, err = repo.GetBlockCount(ctx, blockerID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 不同用户的拉黑数不受影响
	otherCount, err := repo.GetBlockCount(ctx, "other_user")
	require.NoError(t, err)
	assert.Equal(t, int64(0), otherCount)
}

// ========== Stats 操作测试 ==========

func TestGetOrCreateStats_不存在则创建默认值(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()
	userID := "stats_new_user"

	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, 0, stats.TopicsCount)
	assert.Equal(t, 0, stats.RepliesCount)
	assert.Equal(t, 0, stats.LikesReceived)
	assert.Equal(t, 0, stats.GroupsJoined)
}

func TestGetOrCreateStats_已存在则返回已有记录(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()
	userID := "stats_existing_user"

	// 第一次调用 — 创建默认值
	stats1, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats1)

	// 手动修改统计字段（模拟业务逻辑修改）
	stats1.TopicsCount = 42
	stats1.RepliesCount = 100
	stats1.LikesReceived = 200

	// 第二次调用 — 应返回同一对象（非新创建）
	stats2, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats2)
	assert.Equal(t, stats1.UserID, stats2.UserID)
	assert.Equal(t, 42, stats2.TopicsCount)
	assert.Equal(t, 100, stats2.RepliesCount)
	assert.Equal(t, 200, stats2.LikesReceived)
}

func TestUpdateStats_更新统计字段(t *testing.T) {
	repo := newTestMemoryRepo(t)
	ctx := context.Background()
	userID := "stats_update_user"

	// 先获取默认统计
	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)

	// 更新全部统计字段
	stats.TopicsCount = 10
	stats.RepliesCount = 25
	stats.LikesReceived = 50
	stats.GroupsJoined = 3
	stats.UpdatedAt = 1700000007000
	err = repo.UpdateStats(ctx, stats)
	require.NoError(t, err)

	// 重新获取验证写入结果
	updated, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 10, updated.TopicsCount)
	assert.Equal(t, 25, updated.RepliesCount)
	assert.Equal(t, 50, updated.LikesReceived)
	assert.Equal(t, 3, updated.GroupsJoined)
	assert.Equal(t, int64(1700000007000), updated.UpdatedAt)
}

// ========== 边界测试 ==========

func TestNewMemoryRepository_初始化非nil(t *testing.T) {
	repo := NewMemoryRepository()

	// 所有内部 map 必须已分配，不能为 nil
	require.NotNil(t, repo.users)
	require.NotNil(t, repo.userUID)
	require.NotNil(t, repo.username)
	require.NotNil(t, repo.email)
	require.NotNil(t, repo.phone)
	require.NotNil(t, repo.blocks)
	require.NotNil(t, repo.stats)

	// 初始状态下所有 map 为空
	assert.Empty(t, repo.users)
	assert.Empty(t, repo.userUID)
	assert.Empty(t, repo.username)
	assert.Empty(t, repo.email)
	assert.Empty(t, repo.phone)
	assert.Empty(t, repo.blocks)
	assert.Empty(t, repo.stats)
}
