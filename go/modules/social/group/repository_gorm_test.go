// repository_gorm_test.go GormRepository 冒烟测试
// 使用真实 MySQL 数据库验证全部 22 个 Repository 方法的正确性
// 覆盖核心 CRUD、分页、过滤、统计等关键路径
// 环境变量未配置时自动 Skip

package group

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const testIDPrefix = "gt_grp_"

// getEnv 读取环境变量，未设置时返回 fallback 值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setupTestDB 连接真实 MySQL 数据库（从环境变量读取连接参数）
// 未配置 MYSQL_HOST 时自动 t.Skip
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	host := getEnv("MYSQL_HOST", "")
	if host == "" {
		t.Skip("跳过：未设置 MYSQL_HOST 环境变量")
	}
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")
	dbname := getEnv("MYSQL_DATABASE", "cairobot_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// cleanupGroupData 清理当前测试创建的群组相关数据（按 ID 前缀匹配）
func cleanupGroupData(t *testing.T, db *gorm.DB) {
	t.Helper()
	prefix := testIDPrefix + "%"
	db.Exec("DELETE FROM group_stats WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM group_pay_configs WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM group_members WHERE id LIKE ?", prefix)
	db.Exec("DELETE FROM `groups` WHERE id LIKE ?", prefix)
}

// 辅助：创建一个标准测试群组
func makeTestGroup(id, name, slug, ownerID string) *Group {
	now := time.Now().UnixMilli()
	return &Group{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: "测试圈子描述",
		Category:    "tech",
		OwnerID:     ownerID,
		Status:      GroupStatusActive,
		Visibility:  GroupVisibilityPublic,
		JoinMode:    GroupJoinModeOpen,
		MaxMembers:  500,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// 辅助：创建一个标准测试成员记录
func makeTestMember(id, groupID, userID string, role int8) *GroupMember {
	now := time.Now().UnixMilli()
	return &GroupMember{
		ID:        id,
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		Status:    GroupMemberStatusActive,
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ========== Group CRUD 测试 ==========

func TestCreateAndGetGroup(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	group := makeTestGroup(testIDPrefix+"g001", "Go技术圈", "go-tech", "user001")
	require.NoError(t, repo.CreateGroup(ctx, group))

	// 按 ID 查回，字段一致
	found, err := repo.GetGroupByID(ctx, testIDPrefix+"g001")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "Go技术圈", found.Name)
	require.Equal(t, "go-tech", found.Slug)
	require.Equal(t, "user001", found.OwnerID)

	// 按 Slug 查回
	foundBySlug, err := repo.GetGroupBySlug(ctx, "go-tech")
	require.NoError(t, err)
	require.NotNil(t, foundBySlug)
	require.Equal(t, testIDPrefix+"g001", foundBySlug.ID)

	// 不存在的 ID 返回 nil
	notFound, err := repo.GetGroupByID(ctx, "nonexist")
	require.NoError(t, err)
	require.Nil(t, notFound)
}

func TestUpdateGroup(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	group := makeTestGroup(testIDPrefix+"g002", "原始名称", "orig-slug", "user001")
	repo.CreateGroup(ctx, group)

	group.Name = "更新后名称"
	group.Description = "更新后描述"
	require.NoError(t, repo.UpdateGroup(ctx, group))

	updated, _ := repo.GetGroupByID(ctx, testIDPrefix+"g002")
	require.Equal(t, "更新后名称", updated.Name)
	require.Equal(t, "更新后描述", updated.Description)
}

func TestDeleteGroup(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	group := makeTestGroup(testIDPrefix+"g003", "待删除", "del-slug", "user001")
	repo.CreateGroup(ctx, group)

	require.NoError(t, repo.DeleteGroup(ctx, testIDPrefix+"g003"))
	deleted, err := repo.GetGroupByID(ctx, testIDPrefix+"g003")
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func TestListGroupsAndPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 插入 5 条数据
	for i := 0; i < 5; i++ {
		g := makeTestGroup(
			generateGroupID(),
			"圈子"+string(rune('A'+i)),
			"slug-"+string(rune('A'+i)),
			"user001",
		)
		repo.CreateGroup(ctx, g)
	}

	// 第一页 size=2 → 2条，total=5
	list, total, err := repo.ListGroups(ctx, 1, 2, nil)
	require.NoError(t, err)
	require.Equal(t, int64(5), total)
	require.Len(t, list, 2)

	// 第二页 size=2 → 2条
	list2, total2, _ := repo.ListGroups(ctx, 2, 2, nil)
	require.Equal(t, int64(5), total2)
	require.Len(t, list2, 2)

	// 带 filters 过滤
	filteredList, filteredTotal, _ := repo.ListGroups(ctx, 1, 10, map[string]interface{}{
		"owner_id": "user001",
	})
	require.GreaterOrEqual(t, filteredTotal, int64(5))
	require.NotEmpty(t, filteredList)
}

func TestListGroupsByOwnerID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g010", "A的群", "a-group", "owner_a"))
	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g011", "B的群", "b-group", "owner_b"))
	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g012", "A的第二个群", "a-group2", "owner_a"))

	list, total, err := repo.ListGroupsByOwnerID(ctx, "owner_a", 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
}

// ========== Member 操作测试 ==========

func TestCreateAndGetMemberByGroupAndUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	member := makeTestMember(testIDPrefix+"m001", testIDPrefix+"g001", "u001", GroupMemberRoleMember)
	require.NoError(t, repo.CreateMember(ctx, member))

	// 按 groupID + userID 联合查询
	found, err := repo.GetMember(ctx, testIDPrefix+"g001", "u001")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, testIDPrefix+"m001", found.ID)
	require.Equal(t, GroupMemberRoleMember, found.Role)

	// 按 ID 查询
	foundByID, err := repo.GetMemberByID(ctx, testIDPrefix+"m001")
	require.NoError(t, err)
	require.NotNil(t, foundByID)
	require.Equal(t, testIDPrefix+"g001", foundByID.GroupID)
}

func TestCreateGroupWithOwnerMember(t *testing.T) {
	// 验证 Bug 1 修复：创建群组后必须能查到 owner 成员记录
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 模拟 SvcCreate 的行为：先建群，再添加 owner 成员
	group := makeTestGroup(testIDPrefix+"g100", "带Owner的群", "with-owner", "owner_001")
	require.NoError(t, repo.CreateGroup(ctx, group))

	ownerMember := makeTestMember(testIDPrefix+"m_owner_100", testIDPrefix+"g100", "owner_001", GroupMemberRoleOwner)
	require.NoError(t, repo.CreateMember(ctx, ownerMember))

	// 验证 owner 成员可被查询到
	isMember, err := repo.IsUserMember(ctx, testIDPrefix+"g100", "owner_001")
	require.NoError(t, err)
	require.True(t, isMember, "创建者加入为群主后应被识别为活跃成员")

	// 验证角色正确
	member, err := repo.GetMember(ctx, testIDPrefix+"g100", "owner_001")
	require.NoError(t, err)
	require.NotNil(t, member)
	require.Equal(t, GroupMemberRoleOwner, member.Role)
}

func TestUpdateAndDeleteMember(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	member := makeTestMember(testIDPrefix+"m002", testIDPrefix+"g001", "u002", GroupMemberRoleMember)
	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g001", "测试群", "test-g", "owner_x"))
	repo.CreateMember(ctx, member)

	// 更新状态为禁言
	member.Status = GroupMemberStatusBanned
	require.NoError(t, repo.UpdateMember(ctx, member))

	updated, _ := repo.GetMemberByID(ctx, testIDPrefix+"m002")
	require.Equal(t, GroupMemberStatusBanned, updated.Status)

	// 删除成员
	require.NoError(t, repo.DeleteMember(ctx, testIDPrefix+"m002"))
	deleted, err := repo.GetMemberByID(ctx, testIDPrefix+"m002")
	require.NoError(t, err)
	require.Nil(t, deleted)
}

func TestListMembersWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g200", "分页测试群", "page-test", "owner_x"))

	// 插入 7 个成员
	for i := 0; i < 7; i++ {
		m := makeTestMember(
			generateMemberID(),
			testIDPrefix+"g200",
			"user_"+string(rune('0'+i)),
			GroupMemberRoleMember,
		)
		repo.CreateMember(ctx, m)
	}

	// 第一页 size=3 → 3条，total=7
	list, total, err := repo.ListMembers(ctx, testIDPrefix+"g200", 1, 3, nil, nil)
	require.NoError(t, err)
	require.Equal(t, int64(7), total)
	require.Len(t, list, 3)

	// 带 role 过滤：只查管理员（当前无管理员）
	adminList, adminTotal, _ := repo.ListMembers(ctx, testIDPrefix+"g200", 1, 10, ptrInt8(GroupMemberRoleAdmin), nil)
	require.Equal(t, int64(0), adminTotal)
	require.Empty(t, adminList)

	// 带 status 过滤
	activeList, activeTotal, _ := repo.ListMembers(ctx, testIDPrefix+"g200", 1, 10, nil, ptrInt8(GroupMemberStatusActive))
	require.Equal(t, int64(7), activeTotal)
	require.Len(t, activeList, 7)
}

func TestIsUserMember(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g300", "身份检查群", "identity-check", "owner_x"))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m300", testIDPrefix+"g300", "active_user", GroupMemberRoleMember))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m301", testIDPrefix+"g300", "banned_user", GroupMemberRoleMember))

	// 将 banned_user 设为封禁状态
	banned := &GroupMember{ID: testIDPrefix + "m301", Status: GroupMemberStatusBanned}
	repo.UpdateMember(ctx, banned)

	// 活跃成员 → true
	isActive, err := repo.IsUserMember(ctx, testIDPrefix+"g300", "active_user")
	require.NoError(t, err)
	require.True(t, isActive)

	// 封禁成员 → false（status != 1）
	isBanned, err := repo.IsUserMember(ctx, testIDPrefix+"g300", "banned_user")
	require.NoError(t, err)
	require.False(t, isBanned)

	// 非成员 → false
	isNonMember, err := repo.IsUserMember(ctx, testIDPrefix+"g300", "stranger")
	require.NoError(t, err)
	require.False(t, isNonMember)
}

func TestCountActiveMembersAndByRole(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g400", "统计测试群", "stats-test", "owner_x"))

	// 插入混合角色的成员
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m400", testIDPrefix+"g400", "owner", GroupMemberRoleOwner))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m401", testIDPrefix+"g400", "admin1", GroupMemberRoleAdmin))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m402", testIDPrefix+"g400", "admin2", GroupMemberRoleAdmin))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m403", testIDPrefix+"g400", "member1", GroupMemberRoleMember))

	// 活跃成员数 = 4
	activeCount, err := repo.CountActiveMembers(ctx, testIDPrefix+"g400")
	require.NoError(t, err)
	require.Equal(t, int64(4), activeCount)

	// 管理员数 = 2
	adminCount, err := repo.CountMembersByRole(ctx, testIDPrefix+"g400", GroupMemberRoleAdmin)
	require.NoError(t, err)
	require.Equal(t, int64(2), adminCount)

	// 群主数 = 1
	ownerCount, err := repo.CountMembersByRole(ctx, testIDPrefix+"g400", GroupMemberRoleOwner)
	require.NoError(t, err)
	require.Equal(t, int64(1), ownerCount)
}

func TestListMembersByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g500", "群A", "ga", "owner_x"))
	repo.CreateGroup(ctx, makeTestGroup(testIDPrefix+"g501", "群B", "gb", "owner_y"))

	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m500", testIDPrefix+"g500", "multi_user", GroupMemberRoleMember))
	repo.CreateMember(ctx, makeTestMember(testIDPrefix+"m501", testIDPrefix+"g501", "multi_user", GroupMemberRoleAdmin))

	list, total, err := repo.ListMembersByUserID(ctx, "multi_user", 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, list, 2)
}

// ========== PayConfig 测试 ==========

func TestPayConfigCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	config := &GroupPayConfig{
		ID:            testIDPrefix + "pc001",
		GroupID:       testIDPrefix + "g600",
		PriceMonthly:  29.9,
		PriceQuarterly: 79.9,
		PriceYearly:   299.0,
		Currency:      "CNY",
		TrialDays:     7,
		IsEnabled:     true,
		CreatedAt:     time.Now().UnixMilli(),
		UpdatedAt:     time.Now().UnixMilli(),
	}

	// 创建
	require.NoError(t, repo.CreatePayConfig(ctx, config))

	// 查询
	found, err := repo.GetPayConfigByGroupID(ctx, testIDPrefix+"g600")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, 29.9, found.PriceMonthly)
	require.True(t, found.IsEnabled)

	// 更新
	found.PriceMonthly = 39.9
	found.IsEnabled = false
	require.NoError(t, repo.UpdatePayConfig(ctx, found))

	updated, _ := repo.GetPayConfigByGroupID(ctx, testIDPrefix+"g600")
	require.Equal(t, 39.9, updated.PriceMonthly)
	require.False(t, updated.IsEnabled)

	// 不存在返回 nil
	noConfig, err := repo.GetPayConfigByGroupID(ctx, "nonexist")
	require.NoError(t, err)
	require.Nil(t, noConfig)
}

// ========== Stats 测试 ==========

func TestGetOrCreateStatsAndUpdateStats(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupGroupData(t, db)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// 首次获取 → 自动创建默认值
	stats, err := repo.GetOrCreateStats(ctx, testIDPrefix+"g700")
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Equal(t, testIDPrefix+"g700", stats.GroupID)
	require.Equal(t, int64(0), int64(stats.MembersCount))

	// 再次获取 → 返回已有记录（幂等）
	stats2, err := repo.GetOrCreateStats(ctx, testIDPrefix+"g700")
	require.NoError(t, err)
	require.NotNil(t, stats2)
	require.Equal(t, stats.GroupID, stats2.GroupID)

	// 更新统计
	stats2.MembersCount = 100
	stats2.ActiveMembersCount = 80
	stats2.TopicsCount = 50
	require.NoError(t, repo.UpdateStats(ctx, stats2))

	updated, _ := repo.GetOrCreateStats(ctx, testIDPrefix+"g700")
	require.Equal(t, 100, updated.MembersCount)
	require.Equal(t, 80, updated.ActiveMembersCount)
	require.Equal(t, 50, updated.TopicsCount)
}

// ========== 辅助函数 ==========

// ptrInt8 将 int8 值转为指针，用于可选参数传参
func ptrInt8(v int8) *int8 {
	return &v
}
