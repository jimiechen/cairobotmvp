// repository_gorm_test.go — GormRepository 集成测试
// 直连真实 MySQL 数据库（go_biz），验证 User/Block/Stats 全部 CRUD 路径
// 运行条件：需设置 MYSQL_HOST 环境变量（source .env.local 后执行）

package member

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const testIDPrefix = "gt_mbr_" // 集成测试 ID 前缀，避免与生产数据冲突

// setupTestDB 连接真实 MySQL 数据库
// 环境变量：MYSQL_HOST/MYSQL_PORT/MYSQL_USER/MYSQL_PASSWORD/MYSQL_DATABASE
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		t.Skip("跳过集成测试：未设置 MYSQL_HOST（需 source .env.local）")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5000ms",
		getEnv("MYSQL_USER", "root"),
		getEnv("MYSQL_PASSWORD", ""),
		host,
		getEnv("MYSQL_PORT", "3306"),
		getEnv("MYSQL_DATABASE", "go_biz"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	require.NoError(t, sqlDB.PingContext(context.Background()))
	return db
}

// cleanupTestData 按 ID 前缀删除测试数据（测试结束后调用）
// 注意：必须先删 stats（FK→users），再删 blocks，最后删 users
// stats 表用 user_id 匹配（因为 PK=user_id 可能与 id 前缀不同）
func cleanupUserData(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM member_stats WHERE user_id LIKE ?", testIDPrefix+"%")
	db.Exec("DELETE FROM member_blocks WHERE id LIKE ?", testIDPrefix+"%")
	db.Exec("DELETE FROM users WHERE id LIKE ?", testIDPrefix+"%")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ========== User CRUD 测试 ==========

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:              testIDPrefix + "u001",
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

	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	got, err := repo.GetUserByID(ctx, testIDPrefix+"u001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testIDPrefix+"u001", got.ID)
	assert.Equal(t, "100000001", got.UID)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "test@example.com", got.Email)
	assert.Equal(t, "13800138000", got.Phone)
	assert.Equal(t, "TestNick", got.Nickname)
	assert.Equal(t, UserStatusActive, got.Status)
}

func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	user, err := repo.GetUserByID(ctx, testIDPrefix+"_nonexistent_99999")
	require.NoError(t, err)
	assert.Nil(t, user)

	user, err = repo.GetUserByUID(ctx, "999999999")
	require.NoError(t, err)
	assert.Nil(t, user)

	user, err = repo.GetUserByUsername(ctx, testIDPrefix+"_nouser_xyz")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetUserByEmailAndPhone(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:        testIDPrefix + "u002",
		UID:       "100000002",
		Username:  "emailuser",
		Password:  "pwd",
		Email:     testIDPrefix + "findme@test.com",
		Phone:     "13900139000",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	got, err := repo.GetUserByEmail(ctx, testIDPrefix+"findme@test.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testIDPrefix+"u002", got.ID)

	got, err = repo.GetUserByPhone(ctx, "13900139000")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testIDPrefix+"u002", got.ID)
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	user := &User{
		ID:        testIDPrefix + "u003",
		UID:       "100000003",
		Username:  "updateuser",
		Password:  "oldpwd",
		Nickname:  "OldNick",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	user.Nickname = "NewNick"
	user.Avatar = "https://example.com/avatar.png"
	user.UpdatedAt = time.Now().UnixMilli()
	err = repo.UpdateUser(ctx, user)
	require.NoError(t, err)

	got, err := repo.GetUserByID(ctx, testIDPrefix+"u003")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "NewNick", got.Nickname)
	assert.Equal(t, "https://example.com/avatar.png", got.Avatar)
}

func TestBatchGetUsersByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	for i := 0; i < 3; i++ {
		u := &User{
			ID:        fmt.Sprintf("%sbatch_%d", testIDPrefix, i),
			UID:       fmt.Sprintf("10001000%d", i),
			Username:  fmt.Sprintf("%sbatchuser%d", testIDPrefix, i),
			Password:  "pwd",
			Email:     fmt.Sprintf("%sbatch%d@test.com", testIDPrefix, i),
			Phone:     fmt.Sprintf("1390001%04d", i),
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.CreateUser(ctx, u))
	}

	users, err := repo.BatchGetUsersByID(ctx, []string{testIDPrefix + "batch_0", testIDPrefix + "batch_2", "_nonexistent_"})
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// ========== MemberBlock 测试 ==========

func TestCreateAndListBlocks(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	blockerID := testIDPrefix + "blocker_01"

	for i := 0; i < 3; i++ {
		b := &MemberBlock{
			ID:         fmt.Sprintf("%sblk_%d", testIDPrefix, i),
			BlockerID:  blockerID,
			BlockedID:  fmt.Sprintf("%sblocked_%d", testIDPrefix, i),
			Reason:    fmt.Sprintf("reason_%d", i),
			CreatedAt: int64(1700000000000 + int64(i)*1000),
		}
		require.NoError(t, repo.CreateBlock(ctx, b))
	}

	blocks, total, err := repo.ListBlocks(ctx, blockerID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, blocks, 2)

	blocks2, _, err := repo.ListBlocks(ctx, blockerID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, blocks2, 1)
}

func TestIsBlocked(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	blocked, err := repo.IsBlocked(ctx, testIDPrefix+"_a", testIDPrefix+"_b")
	require.NoError(t, err)
	assert.False(t, blocked)

	b := &MemberBlock{
		ID:         testIDPrefix + "blk_test",
		BlockerID:  testIDPrefix + "_a",
		BlockedID:  testIDPrefix + "_b",
		Reason:     "spam",
		CreatedAt:  time.Now().UnixMilli(),
	}
	require.NoError(t, repo.CreateBlock(ctx, b))

	blocked, err = repo.IsBlocked(ctx, testIDPrefix+"_a", testIDPrefix+"_b")
	require.NoError(t, err)
	assert.True(t, blocked)

	blocked, err = repo.IsBlocked(ctx, testIDPrefix+"_b", testIDPrefix+"_a")
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestDeleteBlock(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	b := &MemberBlock{
		ID:         testIDPrefix + "blk_del",
		BlockerID:  testIDPrefix + "_x",
		BlockedID:  testIDPrefix + "_y",
		CreatedAt:  time.Now().UnixMilli(),
	}
	require.NoError(t, repo.CreateBlock(ctx, b))

	blocked, err := repo.IsBlocked(ctx, testIDPrefix+"_x", testIDPrefix+"_y")
	require.NoError(t, err)
	assert.True(t, blocked)

	err = repo.DeleteBlock(ctx, testIDPrefix+"_x", testIDPrefix+"_y")
	require.NoError(t, err)

	blocked, err = repo.IsBlocked(ctx, testIDPrefix+"_x", testIDPrefix+"_y")
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestGetBlockCount(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	count, err := repo.GetBlockCount(ctx, testIDPrefix+"_counter_user")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	for i := 0; i < 2; i++ {
		b := &MemberBlock{
			ID:         fmt.Sprintf("%scnt_blk_%d", testIDPrefix, i),
			BlockerID:  testIDPrefix + "_counter_user",
			BlockedID:  fmt.Sprintf("%starget_%d", testIDPrefix, i),
			CreatedAt:  time.Now().UnixMilli(),
		}
		require.NoError(t, repo.CreateBlock(ctx, b))
	}

	count, err = repo.GetBlockCount(ctx, testIDPrefix+"_counter_user")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// ========== MemberStats 测试 ==========

func TestGetOrCreateStats(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := testIDPrefix + "stats_user_01"

	// 前置：创建用户记录（member_stats 有 FK 约束引用 users.id）
	now := time.Now().UnixMilli()
	repo.CreateUser(ctx, &User{
		ID: userID, UID: "900000001", Username: "stats_user_01",
		Password: "hashed", Email: "stats_01@test.local", Status: UserStatusActive,
		CreatedAt: now, UpdatedAt: now,
	})

	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, 0, stats.TopicsCount)
	assert.Equal(t, 0, stats.RepliesCount)
	assert.Equal(t, 0, stats.LikesReceived)

	stats2, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, stats2)
	assert.Equal(t, stats.UserID, stats2.UserID)
}

func TestUpdateStats(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupUserData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := testIDPrefix + "stats_user_02"

	// 前置：创建用户记录（member_stats 有 FK 约束引用 users.id）
	now := time.Now().UnixMilli()
	repo.CreateUser(ctx, &User{
		ID: userID, UID: "900000002", Username: "stats_user_02",
		Password: "hashed", Email: "stats_02@test.local", Phone: "9000002002",
		Status: UserStatusActive, CreatedAt: now, UpdatedAt: now,
	})

	stats, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)

	stats.TopicsCount = 5
	stats.RepliesCount = 10
	stats.LikesReceived = 20
	stats.GroupsJoined = 1
	stats.UpdatedAt = time.Now().UnixMilli()
	err = repo.UpdateStats(ctx, stats)
	require.NoError(t, err)

	updated, err := repo.GetOrCreateStats(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 5, updated.TopicsCount)
	assert.Equal(t, 10, updated.RepliesCount)
	assert.Equal(t, 20, updated.LikesReceived)
	assert.Equal(t, 1, updated.GroupsJoined)
}
