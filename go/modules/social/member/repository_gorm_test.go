// repository_gorm_test.go — GormRepository 冒烟测试
// 覆盖核心 CRUD 路径：用户创建/查询、拉黑操作、统计初始化
// 使用 SQLite in-memory 数据库，每个测试独立环境

package member

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库并自动迁移表结构
// 每个测试调用此函数获得独立的数据库实例，避免测试间干扰
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&User{}, &MemberBlock{}, &MemberStats{})
	require.NoError(t, err)
	return db
}

// ========== User CRUD 测试 ==========

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:              "u001",
		UID:             "100000001",
		Username:        "testuser",
		Password:        "hashed_pwd",
		Email:           "test@example.com",
		Phone:           "13800138000",
		Nickname:        "TestNick",
		Status:          UserStatusActive,
		MembershipLevel: "normal",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 创建
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// 按 ID 查回
	got, err := repo.GetUserByID(ctx, "u001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u001", got.ID)
	assert.Equal(t, "100000001", got.UID)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "test@example.com", got.Email)
	assert.Equal(t, "13800138000", got.Phone)
	assert.Equal(t, "TestNick", got.Nickname)
	assert.Equal(t, UserStatusActive, got.Status)
}

func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 查不存在的 ID → nil, nil
	user, err := repo.GetUserByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, user)

	// 查不存在的 UID
	user, err = repo.GetUserByUID(ctx, "999999999")
	require.NoError(t, err)
	assert.Nil(t, user)

	// 查不存在的 username
	user, err = repo.GetUserByUsername(ctx, "nouser")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetUserByEmailAndPhone(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:        "u002",
		UID:       "100000002",
		Username:  "emailuser",
		Password:  "pwd",
		Email:     "findme@test.com",
		Phone:     "13900139000",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// 按邮箱查
	got, err := repo.GetUserByEmail(ctx, "findme@test.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u002", got.ID)

	// 按手机号查
	got, err = repo.GetUserByPhone(ctx, "13900139000")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "u002", got.ID)
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:        "u003",
		UID:       "100000003",
		Username:  "updateuser",
		Password:  "oldpwd",
		Nickname:  "OldNick",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// 更新昵称和头像
	user.Nickname = "NewNick"
	user.Avatar = "https://example.com/avatar.png"
	user.UpdatedAt = time.Now().UnixMilli()
	err = repo.UpdateUser(ctx, user)
	require.NoError(t, err)

	// 验证更新结果
	got, err := repo.GetUserByID(ctx, "u003")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "NewNick", got.Nickname)
	assert.Equal(t, "https://example.com/avatar.png", got.Avatar)
}

func TestBatchGetUsersByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	for i := 0; i < 3; i++ {
		u := &User{
			ID:        fmt.Sprintf("batch_%d", i),
			UID:       fmt.Sprintf("10001000%d", i),
			Username:  fmt.Sprintf("batchuser%d", i),
			Password:  "pwd",
			Email:     fmt.Sprintf("batch%d@test.com", i),
			Phone:     fmt.Sprintf("1390001%04d", i),
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.CreateUser(ctx, u))
	}

	users, err := repo.BatchGetUsersByID(ctx, []string{"batch_0", "batch_2", "nonexistent"})
	require.NoError(t, err)
	assert.Len(t, users, 2) // nonexistent 不存在，只返回找到的
}

// ========== MemberBlock 测试 ==========

func TestCreateAndListBlocks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	blockerID := "blocker_01"

	// 创建 3 条拉黑记录
	for i := 0; i < 3; i++ {
		b := &MemberBlock{
			ID:        fmt.Sprintf("blk_%d", i),
			BlockerID: blockerID,
			BlockedID: fmt.Sprintf("blocked_%d", i),
			Reason:    fmt.Sprintf("reason_%d", i),
			CreatedAt: int64(1700000000000 + int64(i)*1000),
		}
		require.NoError(t, repo.CreateBlock(ctx, b))
	}

	// 分页查询第 1 页（每页 2 条）
	blocks, total, err := repo.ListBlocks(ctx, blockerID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, blocks, 2)

	// 第 2 页应剩余 1 条
	blocks2, _, err := repo.ListBlocks(ctx, blockerID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, blocks2, 1)
}

func TestIsBlocked(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 未拉黑时返回 false
	blocked, err := repo.IsBlocked(ctx, "a", "b")
	require.NoError(t, err)
	assert.False(t, blocked)

	// 创建拉黑关系
	b := &MemberBlock{
		ID:         "blk_test",
		BlockerID:  "a",
		BlockedID:  "b",
		Reason:     "spam",
		CreatedAt:  time.Now().UnixMilli(),
	}
	require.NoError(t, repo.CreateBlock(ctx, b))

	// 再次检查应为 true
	blocked, err = repo.IsBlocked(ctx, "a", "b")
	require.NoError(t, err)
	assert.True(t, blocked)

	// 反向检查应为 false（拉黑是单向的）
	blocked, err = repo.IsBlocked(ctx, "b", "a")
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestDeleteBlock(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	b := &MemberBlock{
		ID:         "blk_del",
		BlockerID:  "x",
		BlockedID:  "y",
		CreatedAt:  time.Now().UnixMilli(),
	}
	require.NoError(t, repo.CreateBlock(ctx, b))

	// 确认存在
	blocked, err := repo.IsBlocked(ctx, "x", "y")
	require.NoError(t, err)
	assert.True(t, blocked)

	// 删除
	err = repo.DeleteBlock(ctx, "x", "y")
	require.NoError(t, err)

	// 确认已删除
	blocked, err = repo.IsBlocked(ctx, "x", "y")
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestGetBlockCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.GetBlockCount(ctx, "counter_user")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 创建 2 条拉黑记录
	for i := 0; i < 2; i++ {
		b := &MemberBlock{
			ID:         fmt.Sprintf("cnt_blk_%d", i),
			BlockerID:  "counter_user",
			BlockedID:  fmt.Sprintf("target_%d", i),
			CreatedAt:  time.Now().UnixMilli(),
		}
		require.NoError(t, repo.CreateBlock(ctx, b))
	}

	count, err = repo.GetBlockCount(ctx, "counter_user")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// ========== MemberStats 测试 ==========

func TestGetOrCreateStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := "stats_user_01"

	// 首次获取 — 应创建默认统计记录
	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, 0, stats.TopicsCount)
	assert.Equal(t, 0, stats.RepliesCount)
	assert.Equal(t, 0, stats.LikesReceived)

	// 再次获取 — 应返回同一记录（幂等）
	stats2, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats2)
	assert.Equal(t, stats.UserID, stats2.UserID)
}

func TestUpdateStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := "stats_user_02"

	// 先创建默认统计
	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)

	// 更新计数器
	stats.TopicsCount = 5
	stats.RepliesCount = 10
	stats.LikesReceived = 20
	stats.GroupsJoined = 1
	stats.UpdatedAt = time.Now().UnixMilli()
	err = repo.UpdateStats(ctx, stats)
	require.NoError(t, err)

	// 重新获取验证
	updated, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 5, updated.TopicsCount)
	assert.Equal(t, 10, updated.RepliesCount)
	assert.Equal(t, 20, updated.LikesReceived)
	assert.Equal(t, 1, updated.GroupsJoined)
}
