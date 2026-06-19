package group

import (
	"context"
	"testing"
)

// TestNewMemoryRepository_初始化非nil 验证 NewMemoryRepository 返回的实例各字段均已正确初始化
func TestNewMemoryRepository_初始化非nil(t *testing.T) {
	repo := NewMemoryRepository()

	if repo == nil {
		t.Fatal("NewMemoryRepository 不应返回 nil")
	}
	if repo.groups == nil {
		t.Error("groups map 未初始化")
	}
	if repo.groupsBySlug == nil {
		t.Error("groupsBySlug map 未初始化")
	}
	if repo.members == nil {
		t.Error("members map 未初始化")
	}
	if repo.membersByKey == nil {
		t.Error("membersByKey map 未初始化")
	}
	if repo.payConfigs == nil {
		t.Error("payConfigs map 未初始化")
	}
	if repo.stats == nil {
		t.Error("stats map 未初始化")
	}
}

// ========== Group 群组 CRUD 测试 ==========

// TestCreateGroup_创建圈子并双索引查询 验证 CreateGroup 后可通过 ID 和 Slug 两种方式查到同一记录
func TestCreateGroup_创建圈子并双索引查询(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	group := &Group{
		ID:          "grp-001",
		Name:        "测试圈子",
		Slug:        "test-group",
		Description: "这是一个测试圈子",
		OwnerID:     "user-001",
		Status:      1,
	}

	err := repo.CreateGroup(ctx, group)
	if err != nil {
		t.Fatalf("CreateGroup 失败: %v", err)
	}

	// 通过 ID 查询
	byID, err := repo.GetGroupByID(ctx, "grp-001")
	if err != nil {
		t.Fatalf("GetGroupByID 返回错误: %v", err)
	}
	if byID == nil {
		t.Fatal("通过 ID 查询应返回记录")
	}
	if byID.Name != "测试圈子" {
		t.Errorf("Name 期望 '测试圈子', 实际 '%s'", byID.Name)
	}

	// 通过 Slug 查询
	bySlug, err := repo.GetGroupBySlug(ctx, "test-group")
	if err != nil {
		t.Fatalf("GetGroupBySlug 返回错误: %v", err)
	}
	if bySlug == nil {
		t.Fatal("通过 Slug 查询应返回记录")
	}
	if bySlug.ID != "grp-001" {
		t.Errorf("通过 Slug 查到的 ID 期望 'grp-001', 实际 '%s'", bySlug.ID)
	}
}

// TestUpdateGroup_更新圈子信息 验证更新后字段值变更且 slug 索引同步更新
func TestUpdateGroup_更新圈子信息(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	original := &Group{
		ID:   "grp-002",
		Name: "原始名称",
		Slug: "old-slug",
	}
	repo.CreateGroup(ctx, original)

	updated := &Group{
		ID:          "grp-002",
		Name:        "新名称",
		Slug:        "new-slug",
		Description: "更新后的描述",
		Status:      2,
	}

	err := repo.UpdateGroup(ctx, updated)
	if err != nil {
		t.Fatalf("UpdateGroup 失败: %v", err)
	}

	// 验证 ID 查询返回更新后的值
	g, _ := repo.GetGroupByID(ctx, "grp-002")
	if g.Name != "新名称" {
		t.Errorf("Name 期望 '新名称', 实际 '%s'", g.Name)
	}
	if g.Description != "更新后的描述" {
		t.Errorf("Description 期望 '更新后的描述', 实际 '%s'", g.Description)
	}

	// 旧 slug 应不可查到
	oldSlug, _ := repo.GetGroupBySlug(ctx, "old-slug")
	if oldSlug != nil {
		t.Error("旧 slug 应已被清理")
	}

	// 新 slug 应可查到
	newSlug, _ := repo.GetGroupBySlug(ctx, "new-slug")
	if newSlug == nil || newSlug.ID != "grp-002" {
		t.Error("新 slug 应可查到对应记录")
	}
}

// TestDeleteGroup_删除圈子 验证删除后 ID 和 slug 双索引均无法查到
func TestDeleteGroup_删除圈子(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	group := &Group{ID: "grp-003", Name: "待删除圈子", Slug: "delete-me"}
	repo.CreateGroup(ctx, group)

	err := repo.DeleteGroup(ctx, "grp-003")
	if err != nil {
		t.Fatalf("DeleteGroup 失败: %v", err)
	}

	byID, _ := repo.GetGroupByID(ctx, "grp-003")
	if byID != nil {
		t.Error("删除后通过 ID 不应查到记录")
	}

	bySlug, _ := repo.GetGroupBySlug(ctx, "delete-me")
	if bySlug != nil {
		t.Error("删除后通过 Slug 不应查到记录")
	}
}

// TestListGroups_分页列出圈子 验证分页参数和筛选条件正常工作
func TestListGroups_分页列出圈子(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 创建测试数据：3 个活跃 + 1 个已归档
	for i := 1; i <= 3; i++ {
		repo.CreateGroup(ctx, &Group{
			ID:    string(rune('0'+i)),
			Name:  string(rune('A' + rune(i))),
			Slug:  string(rune('a' + rune(i))),
			Status: 1,
		})
	}
	repo.CreateGroup(ctx, &Group{ID: "4", Name: "D", Slug: "d", Status: 2})

	// 无筛选：总数应为 4
	all, total, err := repo.ListGroups(ctx, 1, 10, nil)
	if err != nil {
		t.Fatalf("ListGroups 失败: %v", err)
	}
	if total != 4 {
		t.Errorf("无筛选时总数期望 4, 实际 %d", total)
	}
	if len(all) != 4 {
		t.Errorf("无筛选时返回数量期望 4, 实际 %d", len(all))
	}

	// 按 status=1 筛选（只验证总数）
	_, activeTotal, _ := repo.ListGroups(ctx, 1, 10, map[string]interface{}{"status": int8(1)})
	if activeTotal != 3 {
		t.Errorf("status=1 时总数期望 3, 实际 %d", activeTotal)
	}

	// 分页：第 1 页每页 2 条
	page1, _, _ := repo.ListGroups(ctx, 1, 2, nil)
	if len(page1) != 2 {
		t.Errorf("第 1 页期望 2 条, 实际 %d", len(page1))
	}

	// 分页：第 2 页每页 2 条
	page2, _, _ := repo.ListGroups(ctx, 2, 2, nil)
	if len(page2) != 2 {
		t.Errorf("第 2 页期望 2 条, 实际 %d", len(page2))
	}

	// 超出范围：第 3 页每页 2 条
	page3, total3, _ := repo.ListGroups(ctx, 3, 2, nil)
	if len(page3) != 0 {
		t.Errorf("超出范围的页期望空列表, 实际 %d 条", len(page3))
	}
	if total3 != 4 {
		t.Errorf("超出范围时总数仍应返回 4, 实际 %d", total3)
	}
}

// TestListGroupsByOwnerID_按圈主过滤 验证只返回指定 ownerID 的群组
func TestListGroupsByOwnerID_按圈主过滤(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateGroup(ctx, &Group{ID: "g1", Name: "A", Slug: "a", OwnerID: "owner-a"})
	repo.CreateGroup(ctx, &Group{ID: "g2", Name: "B", Slug: "b", OwnerID: "owner-a"})
	repo.CreateGroup(ctx, &Group{ID: "g3", Name: "C", Slug: "c", OwnerID: "owner-b"})

	list, total, err := repo.ListGroupsByOwnerID(ctx, "owner-a", 1, 10)
	if err != nil {
		t.Fatalf("ListGroupsByOwnerID 失败: %v", err)
	}
	if total != 2 {
		t.Errorf("owner-a 的群组数量期望 2, 实际 %d", total)
	}
	if len(list) != 2 {
		t.Errorf("返回列表长度期望 2, 实际 %d", len(list))
	}
}

// ========== GroupMember 成员操作测试 ==========

// TestCreateMember_创建成员并复合键查询 验证创建后可通过 ID 和 groupID:userID 复合键两种方式查到
func TestCreateMember_创建成员并复合键查询(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	member := &GroupMember{
		ID:      "mem-001",
		GroupID: "grp-001",
		UserID:  "user-001",
		Role:    1, // owner
		Status:  1,
	}

	err := repo.CreateMember(ctx, member)
	if err != nil {
		t.Fatalf("CreateMember 失败: %v", err)
	}

	// 通过复合键查询
	byKey, err := repo.GetMember(ctx, "grp-001", "user-001")
	if err != nil {
		t.Fatalf("GetMember 返回错误: %v", err)
	}
	if byKey == nil {
		t.Fatal("通过复合键查询应返回记录")
	}
	if byKey.Role != 1 {
		t.Errorf("Role 期望 1(Owner), 实际 %d", byKey.Role)
	}

	// 通过主键 ID 查询
	byID, err := repo.GetMemberByID(ctx, "mem-001")
	if err != nil {
		t.Fatalf("GetMemberByID 返回错误: %v", err)
	}
	if byID == nil {
		t.Fatal("通过 ID 查询应返回记录")
	}
	if byID.UserID != "user-001" {
		t.Errorf("UserID 期望 'user-001', 实际 '%s'", byID.UserID)
	}
}

// TestUpdateMember_更新成员角色和状态 验证更新后角色和状态值变更
func TestUpdateMember_更新成员角色和状态(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	original := &GroupMember{
		ID:      "mem-002",
		GroupID: "grp-001",
		UserID:  "user-002",
		Role:    3, // 普通成员
		Status:  1,
	}
	repo.CreateMember(ctx, original)

	updated := &GroupMember{
		ID:      "mem-002",
		GroupID: "grp-001",
		UserID:  "user-002",
		Role:    2, // 升级为管理员
		Status:  4, // 被禁言
	}

	err := repo.UpdateMember(ctx, updated)
	if err != nil {
		t.Fatalf("UpdateMember 失败: %v", err)
	}

	m, _ := repo.GetMember(ctx, "grp-001", "user-002")
	if m.Role != 2 {
		t.Errorf("Role 期望 2(Admin), 实际 %d", m.Role)
	}
	if m.Status != 4 {
		t.Errorf("Status 期望 4(Muted), 实际 %d", m.Status)
	}
}

// TestDeleteMember_删除成员 验证删除后 ID 索引和复合键索引均清理
func TestDeleteMember_删除成员(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	member := &GroupMember{
		ID:      "mem-003",
		GroupID: "grp-001",
		UserID:  "user-003",
	}
	repo.CreateMember(ctx, member)

	err := repo.DeleteMember(ctx, "mem-003")
	if err != nil {
		t.Fatalf("DeleteMember 失败: %v", err)
	}

	byID, _ := repo.GetMemberByID(ctx, "mem-003")
	if byID != nil {
		t.Error("删除后通过 ID 不应查到记录")
	}

	byKey, _ := repo.GetMember(ctx, "grp-001", "user-003")
	if byKey != nil {
		t.Error("删除后通过复合键不应查到记录")
	}
}

// TestListMembers_分页列成员 验证分页和角色/状态筛选功能
func TestListMembers_分页列成员(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 创建 3 个成员：1 owner + 1 admin + 1 member
	repo.CreateMember(ctx, &GroupMember{ID: "m1", GroupID: "grp-list", UserID: "u1", Role: 1, Status: 1})
	repo.CreateMember(ctx, &GroupMember{ID: "m2", GroupID: "grp-list", UserID: "u2", Role: 2, Status: 1})
	repo.CreateMember(ctx, &GroupMember{ID: "m3", GroupID: "grp-list", UserID: "u3", Role: 3, Status: 1})
	// 其他群组的成员不应出现
	repo.CreateMember(ctx, &GroupMember{ID: "m4", GroupID: "other-grp", UserID: "u4", Role: 3, Status: 1})

	// 无筛选（只验证总数）
	_, total, err := repo.ListMembers(ctx, "grp-list", 1, 10, nil, nil)
	if err != nil {
		t.Fatalf("ListMembers 失败: %v", err)
	}
	if total != 3 {
		t.Errorf("grp-list 成员总数期望 3, 实际 %d", total)
	}

	// 按角色筛选 admin
	admins, adminTotal, _ := repo.ListMembers(ctx, "grp-list", 1, 10, ptrInt8(2), nil)
	if adminTotal != 1 {
		t.Errorf("admin 数量期望 1, 实际 %d", adminTotal)
	}
	if admins[0].UserID != "u2" {
		t.Errorf("admin 的 UserID 期望 'u2', 实际 '%s'", admins[0].UserID)
	}

	// 分页
	page1, _, _ := repo.ListMembers(ctx, "grp-list", 1, 2, nil, nil)
	if len(page1) != 2 {
		t.Errorf("第 1 页期望 2 条, 实际 %d", len(page1))
	}
}

// TestListMembersByUserID_按用户反查群组 验证返回指定用户加入的所有群组成员记录
func TestListMembersByUserID_按用户反查群组(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateMember(ctx, &GroupMember{ID: "m1", GroupID: "grp-a", UserID: "user-x"})
	repo.CreateMember(ctx, &GroupMember{ID: "m2", GroupID: "grp-b", UserID: "user-x"})
	repo.CreateMember(ctx, &GroupMember{ID: "m3", GroupID: "grp-c", UserID: "other-user"})

	list, total, err := repo.ListMembersByUserID(ctx, "user-x", 1, 10)
	if err != nil {
		t.Fatalf("ListMembersByUserID 失败: %v", err)
	}
	if total != 2 {
		t.Errorf("user-x 加入的群组数期望 2, 实际 %d", total)
	}
	if len(list) != 2 {
		t.Errorf("返回列表长度期望 2, 实际 %d", len(list))
	}
}

// TestIsUserMember_判断是否为成员 验证活跃成员返回 true，非活跃或不存在返回 false
func TestIsUserMember_判断是否为成员(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateMember(ctx, &GroupMember{ID: "m1", GroupID: "grp-check", UserID: "active-user", Role: 3, Status: 1})
	repo.CreateMember(ctx, &GroupMember{ID: "m2", GroupID: "grp-check", UserID: "left-user", Role: 3, Status: 2}) // 已退出

	// 活跃成员
	isMember, err := repo.IsUserMember(ctx, "grp-check", "active-user")
	if err != nil {
		t.Fatalf("IsUserMember 返回错误: %v", err)
	}
	if !isMember {
		t.Error("活跃用户应返回 true")
	}

	// 已退出成员
	isLeft, _ := repo.IsUserMember(ctx, "grp-check", "left-user")
	if isLeft {
		t.Error("已退出用户应返回 false")
	}

	// 从未加入的用户
	isNever, _ := repo.IsUserMember(ctx, "grp-check", "never-user")
	if isNever {
		t.Error("从未加入的用户应返回 false")
	}
}

// TestCountActiveMembers_统计活跃成员数 验证只统计 status=1 的成员
func TestCountActiveMembers_统计活跃成员数(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateMember(ctx, &GroupMember{ID: "m1", GroupID: "grp-count", UserID: "u1", Status: 1})
	repo.CreateMember(ctx, &GroupMember{ID: "m2", GroupID: "grp-count", UserID: "u2", Status: 1})
	repo.CreateMember(ctx, &GroupMember{ID: "m3", GroupID: "grp-count", UserID: "u3", Status: 2}) // 已退出
	repo.CreateMember(ctx, &GroupMember{ID: "m4", GroupID: "grp-count", UserID: "u4", Status: 3}) // 已移除

	count, err := repo.CountActiveMembers(ctx, "grp-count")
	if err != nil {
		t.Fatalf("CountActiveMembers 返回错误: %v", err)
	}
	if count != 2 {
		t.Errorf("活跃成员数期望 2, 实际 %d", count)
	}
}

// TestCountMembersByRole_按角色统计 验证指定角色的成员计数准确
func TestCountMembersByRole_按角色统计(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	repo.CreateMember(ctx, &GroupMember{ID: "m1", GroupID: "grp-role", UserID: "u1", Role: 1}) // owner
	repo.CreateMember(ctx, &GroupMember{ID: "m2", GroupID: "grp-role", UserID: "u2", Role: 2}) // admin
	repo.CreateMember(ctx, &GroupMember{ID: "m3", GroupID: "grp-role", UserID: "u3", Role: 2}) // admin
	repo.CreateMember(ctx, &GroupMember{ID: "m4", GroupID: "grp-role", UserID: "u4", Role: 3}) // member

	ownerCount, _ := repo.CountMembersByRole(ctx, "grp-role", 1)
	if ownerCount != 1 {
		t.Errorf("owner 数量期望 1, 实际 %d", ownerCount)
	}

	adminCount, _ := repo.CountMembersByRole(ctx, "grp-role", 2)
	if adminCount != 2 {
		t.Errorf("admin 数量期望 2, 实际 %d", adminCount)
	}

	memberCount, _ := repo.CountMembersByRole(ctx, "grp-role", 3)
	if memberCount != 1 {
		t.Errorf("member 数量期望 1, 实际 %d", memberCount)
	}
}

// ========== PayConfig / Stats 测试 ==========

// TestPayConfigCRUD_付费配置完整生命周期 验证创建、查询、更新、再查询的完整流程
func TestPayConfigCRUD_付费配置完整生命周期(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	config := &GroupPayConfig{
		ID:            "pay-001",
		GroupID:       "grp-pay",
		PriceMonthly:  29.9,
		PriceQuarterly: 79.9,
		PriceYearly:   299.0,
		Currency:      "CNY",
		TrialDays:     7,
		IsEnabled:     true,
	}

	// 创建
	err := repo.CreatePayConfig(ctx, config)
	if err != nil {
		t.Fatalf("CreatePayConfig 失败: %v", err)
	}

	// 查询
	found, err := repo.GetPayConfigByGroupID(ctx, "grp-pay")
	if err != nil {
		t.Fatalf("GetPayConfigByGroupID 返回错误: %v", err)
	}
	if found == nil {
		t.Fatal("刚创建的 PayConfig 应可查到")
	}
	if found.PriceMonthly != 29.9 {
		t.Errorf("PriceMonthly 期望 29.9, 实际 %f", found.PriceMonthly)
	}
	if !found.IsEnabled {
		t.Error("IsEnabled 应为 true")
	}

	// 更新
	config.PriceMonthly = 39.9
	config.IsEnabled = false
	err = repo.UpdatePayConfig(ctx, config)
	if err != nil {
		t.Fatalf("UpdatePayConfig 失败: %v", err)
	}

	// 再次查询验证更新
	updated, _ := repo.GetPayConfigByGroupID(ctx, "grp-pay")
	if updated.PriceMonthly != 39.9 {
		t.Errorf("更新后 PriceMonthly 期望 39.9, 实际 %f", updated.PriceMonthly)
	}
	if updated.IsEnabled {
		t.Error("更新后 IsEnabled 应为 false")
	}
}

// TestGetOrCreateStats_获取或初始化统计 验证不存在时自动创建默认值，存在时直接返回
func TestGetOrCreateStats_获取或初始化统计(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 首次获取：不存在时应自动创建
	stats, err := repo.GetOrCreateStats(ctx, "grp-stats-new")
	if err != nil {
		t.Fatalf("GetOrCreateStats 失败: %v", err)
	}
	if stats == nil {
		t.Fatal("GetOrCreateStats 不应返回 nil")
	}
	if stats.GroupID != "grp-stats-new" {
		t.Errorf("GroupID 期望 'grp-stats-new', 实际 '%s'", stats.GroupID)
	}
	if stats.MembersCount != 0 {
		t.Errorf("新建 Stats 的 MembersCount 默认应为 0, 实际 %d", stats.MembersCount)
	}

	// 二次获取：应返回同一条记录（非新建）
	same, _ := repo.GetOrCreateStats(ctx, "grp-stats-new")
	if same != stats {
		t.Error("二次获取应返回同一对象引用")
	}
}

// TestUpdateStats_更新统计数据 验证更新后值持久化
func TestUpdateStats_更新统计数据(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 先初始化
	stats, _ := repo.GetOrCreateStats(ctx, "grp-stats-update")

	// 更新
	stats.MembersCount = 100
	stats.TopicsCount = 50
	stats.ActiveMembersCount = 80
	stats.PaidMembersCount = 20

	err := repo.UpdateStats(ctx, stats)
	if err != nil {
		t.Fatalf("UpdateStats 失败: %v", err)
	}

	// 重新获取验证
	fetched, _ := repo.GetOrCreateStats(ctx, "grp-stats-update")
	if fetched.MembersCount != 100 {
		t.Errorf("更新后 MembersCount 期望 100, 实际 %d", fetched.MembersCount)
	}
	if fetched.TopicsCount != 50 {
		t.Errorf("更新后 TopicsCount 期望 50, 实际 %d", fetched.TopicsCount)
	}
}

// ========== 边界情况测试 ==========

// TestQueryNonExistentData_查询不存在的数据 验证查询不存在的数据返回 (nil, nil) 而非 error
func TestQueryNonExistentData_查询不存在的数据(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Group 查询
	g, err := repo.GetGroupByID(ctx, "nonexistent-group")
	if err != nil {
		t.Errorf("GetGroupByID 对不存在的 ID 不应返回 error: %v", err)
	}
	if g != nil {
		t.Error("GetGroupByID 对不存在的 ID 应返回 nil")
	}

	gs, err := repo.GetGroupBySlug(ctx, "nonexistent-slug")
	if err != nil {
		t.Errorf("GetGroupBySlug 对不存在的 slug 不应返回 error: %v", err)
	}
	if gs != nil {
		t.Error("GetGroupBySlug 对不存在的 slug 应返回 nil")
	}

	// Member 查询
	m, err := repo.GetMember(ctx, "grp-xxx", "user-xxx")
	if err != nil {
		t.Errorf("GetMember 对不存在的复合键不应返回 error: %v", err)
	}
	if m != nil {
		t.Error("GetMember 对不存在的复合键应返回 nil")
	}

	mi, err := repo.GetMemberByID(ctx, "nonexistent-member-id")
	if err != nil {
		t.Errorf("GetMemberByID 对不存在的 ID 不应返回 error: %v", err)
	}
	if mi != nil {
		t.Error("GetMemberByID 对不存在的 ID 应返回 nil")
	}

	// PayConfig 查询
	pc, err := repo.GetPayConfigByGroupID(ctx, "nonexistent-pay-group")
	if err != nil {
		t.Errorf("GetPayConfigByGroupID 对不存在的 groupID 不应返回 error: %v", err)
	}
	if pc != nil {
		t.Error("GetPayConfigByGroupID 对不存在的 groupID 应返回 nil")
	}

	// IsUserMember 对不存在的用户
	isMem, err := repo.IsUserMember(ctx, "grp-xxx", "user-xxx")
	if err != nil {
		t.Errorf("IsUserMember 对不存在的用户不应返回 error: %v", err)
	}
	if isMem {
		t.Error("IsUserMember 对不存在的用户应返回 false")
	}
}

// ========== 辅助函数 ==========

// 注意：ptrInt8 函数已在 repository_gorm_test.go 中定义，此处不再重复声明
