package permission

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
)

// 固定测试 ID，避免硬编码散落
const (
	testGroupID    = "grp_001"
	testPlazaGroupID = "plaza_000"
	testUserID     = "user_001"
	testOwnerID    = "user_owner"
	testAdminID    = "user_admin"
	testMemberID   = "user_member"
	testTopicID    = "topic_001"
	testOtherID    = "user_other"
)

// newTestService 创建带预填充数据的测试用 permission service
func newTestService(memberRepo *mockMemberRepo, groupRepo *mockGroupRepo, topicRepo *mockTopicRepo) Service {
	return NewService(groupRepo, memberRepo, topicRepo, testPlazaGroupID)
}

// ────────────────────────────────────────────────
// Helper: isUserActive 测试（2 个场景）
// ────────────────────────────────────────────────

func TestIsUserActive_ActiveUser_ReturnsTrue(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusActive}

	svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
	require.True(t, svc.(*service).isUserActive(context.Background(), testUserID))
}

func TestIsUserActive_NonActiveOrNotExist_ReturnsFalse(t *testing.T) {
	t.Run("banned_user", func(t *testing.T) {
		memberRepo := newMockMemberRepo()
		memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusSuspended}
		svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
		require.False(t, svc.(*service).isUserActive(context.Background(), testUserID))
	})

	t.Run("deleted_user", func(t *testing.T) {
		memberRepo := newMockMemberRepo()
		memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusDeleted}
		svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
		require.False(t, svc.(*service).isUserActive(context.Background(), testUserID))
	})

	t.Run("nonexistent_user", func(t *testing.T) {
		memberRepo := newMockMemberRepo()
		svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
		require.False(t, svc.(*service).isUserActive(context.Background(), testUserID))
	})
}

// ────────────────────────────────────────────────
// Helper: getMemberRole 测试（2 个场景）
// ────────────────────────────────────────────────

func TestGetMemberRole_ActiveMember_ReturnsCorrectRoleAndStatus(t *testing.T) {
	groupRepo := newMockGroupRepo()

	tests := []struct {
		name     string
		role     int8
		status   int8
		wantRole string
	}{
		{"owner_active", group.GroupMemberRoleOwner, group.GroupMemberStatusActive, RoleStrOwner},
		{"admin_active", group.GroupMemberRoleAdmin, group.GroupMemberStatusActive, RoleStrAdmin},
		{"member_active", group.GroupMemberRoleMember, group.GroupMemberStatusActive, RoleStrMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
				GroupID: testGroupID,
				UserID:  testUserID,
				Role:    tt.role,
				Status:  tt.status,
			}
			svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
			result := svc.(*service).getMemberRole(context.Background(), testGroupID, testUserID)
			require.NotNil(t, result)
			require.Equal(t, tt.wantRole, result.Role)
			require.Equal(t, tt.status, result.Status)
		})
	}
}

func TestGetMemberRole_MemberNotExists_ReturnsNil(t *testing.T) {
	groupRepo := newMockGroupRepo()
	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())

	result := svc.(*service).getMemberRole(context.Background(), testGroupID, "nonexistent")
	require.Nil(t, result)
}

// ────────────────────────────────────────────────
// Helper: getTopicVisibility 测试（2 个场景）
// ────────────────────────────────────────────────

func TestGetTopicVisibility_TopicExists_ReturnsGroupIDAndVisibility(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID:         testTopicID,
		GroupID:    testGroupID,
		Visibility: topic.TopicVisibilityPublic,
	}

	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)
	result := svc.(*service).getTopicVisibility(context.Background(), testTopicID)
	require.NotNil(t, result)
	require.Equal(t, testGroupID, result.GroupID)
	require.Equal(t, topic.TopicVisibilityPublic, result.Visibility)
}

func TestGetTopicVisibility_TopicNotExists_ReturnsNil(t *testing.T) {
	topicRepo := newMockTopicRepo()
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)

	result := svc.(*service).getTopicVisibility(context.Background(), "nonexistent_topic")
	require.Nil(t, result)
}

// ────────────────────────────────────────────────
// CanViewGroup 测试（广场特化 + 非广场）
// ────────────────────────────────────────────────

func TestCanViewGroup_Plaza_ActiveUser_Allow(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusActive}

	svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
	require.True(t, svc.CanViewGroup(context.Background(), testPlazaGroupID, testUserID))
}

func TestCanViewGroup_Plaza_InactiveUser_Deny(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusSuspended}

	svc := newTestService(memberRepo, newMockGroupRepo(), newMockTopicRepo())
	require.False(t, svc.CanViewGroup(context.Background(), testPlazaGroupID, testUserID))
}

func TestCanViewGroup_NormalGroup_ActiveMember_Allow(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testUserID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	require.True(t, svc.CanViewGroup(context.Background(), testGroupID, testUserID))
}

func TestCanViewGroup_NormalGroup_NonMember_Deny(t *testing.T) {
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
	require.False(t, svc.CanViewGroup(context.Background(), testGroupID, testUserID))
}

// ────────────────────────────────────────────────
// CanJoinGroup 测试
// ────────────────────────────────────────────────

func TestCanJoinGroup_Plaza_AlwaysDeny(t *testing.T) {
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
	require.False(t, svc.CanJoinGroup(context.Background(), testPlazaGroupID, testUserID))
}

func TestCanJoinGroup_NormalGroup_NonMember_Allow(t *testing.T) {
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
	require.True(t, svc.CanJoinGroup(context.Background(), testGroupID, testUserID))
}

func TestCanJoinGroup_NormalGroup_ActiveMember_Deny(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testUserID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	require.False(t, svc.CanJoinGroup(context.Background(), testGroupID, testUserID))
}

// ────────────────────────────────────────────────
// CanManageGroup 测试（2 个场景）
// ────────────────────────────────────────────────

func TestCanManageGroup_OwnerOrAdmin_Allow(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testOwnerID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testOwnerID,
		Role: group.GroupMemberRoleOwner, Status: group.GroupMemberStatusActive,
	}
	groupRepo.members[testGroupID+":"+testAdminID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testAdminID,
		Role: group.GroupMemberRoleAdmin, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	require.True(t, svc.CanManageGroup(context.Background(), testGroupID, testOwnerID))
	require.True(t, svc.CanManageGroup(context.Background(), testGroupID, testAdminID))
}

func TestCanManageGroup_MemberOrNonMember_Deny(t *testing.T) {
	t.Run("normal_member", func(t *testing.T) {
		groupRepo := newMockGroupRepo()
		groupRepo.members[testGroupID+":"+testMemberID] = &group.GroupMember{
			GroupID: testGroupID, UserID: testMemberID,
			Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
		}
		svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
		require.False(t, svc.CanManageGroup(context.Background(), testGroupID, testMemberID))
	})

	t.Run("non_member", func(t *testing.T) {
		svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
		require.False(t, svc.CanManageGroup(context.Background(), testGroupID, testOtherID))
	})
}

// ────────────────────────────────────────────────
// CanManageMember 测试（3 个场景）
// ────────────────────────────────────────────────

func TestCanManageMember_AdminOperatesNormalMember_Allow(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testAdminID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testAdminID,
		Role: group.GroupMemberRoleAdmin, Status: group.GroupMemberStatusActive,
	}
	groupRepo.members[testGroupID+":"+testMemberID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testMemberID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	require.True(t, svc.CanManageMember(context.Background(), testGroupID, testAdminID, testMemberID))
}

func TestCanManageMember_MemberOperatesOthers_Deny(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testMemberID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testMemberID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}
	groupRepo.members[testGroupID+":"+testOtherID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testOtherID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	require.False(t, svc.CanManageMember(context.Background(), testGroupID, testMemberID, testOtherID))
}

func TestCanManageMember_OperateOnOwner_Deny(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testAdminID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testAdminID,
		Role: group.GroupMemberRoleAdmin, Status: group.GroupMemberStatusActive,
	}
	groupRepo.members[testGroupID+":"+testOwnerID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testOwnerID,
		Role: group.GroupMemberRoleOwner, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
	// admin 不能操作 owner
	require.False(t, svc.CanManageMember(context.Background(), testGroupID, testAdminID, testOwnerID))
	// owner 不能操作另一个 owner（自己）
	require.False(t, svc.CanManageMember(context.Background(), testGroupID, testOwnerID, testOwnerID))
}

// ────────────────────────────────────────────────
// CanPublishTopic 测试（2 个场景）
// ────────────────────────────────────────────────

func TestCanPublishTopic_ActiveMemberNotMuted_Allow(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusActive}

	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testUserID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(memberRepo, groupRepo, newMockTopicRepo())
	require.True(t, svc.CanPublishTopic(context.Background(), testGroupID, testUserID))
}

func TestCanPublishTopic_DenyCases(t *testing.T) {
	t.Run("muted_member_in_normal_group", func(t *testing.T) {
		groupRepo := newMockGroupRepo()
		groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
			GroupID: testGroupID, UserID: testUserID,
			Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusMuted,
		}
		svc := newTestService(newMockMemberRepo(), groupRepo, newMockTopicRepo())
		require.False(t, svc.CanPublishTopic(context.Background(), testGroupID, testUserID))
	})

	t.Run("non_member_in_normal_group", func(t *testing.T) {
		svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
		require.False(t, svc.CanPublishTopic(context.Background(), testGroupID, testUserID))
	})

	t.Run("banned_user_in_plaza", func(t *testing.T) {
		memberRepo := newMockMemberRepo()
		memberRepo.users[testUserID] = &member.User{ID: testUserID, Status: member.UserStatusSuspended}

		groupRepo := newMockGroupRepo()
		// 有 ban 记录（status=banned）
		groupRepo.members[testPlazaGroupID+":"+testUserID] = &group.GroupMember{
			GroupID: testPlazaGroupID, UserID: testUserID,
			Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusBanned,
		}

		svc := newTestService(memberRepo, groupRepo, newMockTopicRepo())
		require.False(t, svc.CanPublishTopic(context.Background(), testPlazaGroupID, testUserID))
	})
}

// ────────────────────────────────────────────────
// CanReadTopic 测试（6 个场景）
// ────────────────────────────────────────────────

func TestCanReadTopic_Public_Allow(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityPublic,
	}

	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)
	require.True(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testUserID))
}

func TestCanReadTopic_GroupMember_ActiveMember_Allow(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityGroupMember,
	}

	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testUserID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testUserID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, topicRepo)
	require.True(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testUserID))
}

func TestCanReadTopic_GroupMember_NonMember_Deny(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityGroupMember,
	}

	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)
	require.False(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testUserID))
}

func TestCanReadTopic_OwnerOnly_OwnerOrAdmin_Allow(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityOwnerOnly,
	}

	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testOwnerID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testOwnerID,
		Role: group.GroupMemberRoleOwner, Status: group.GroupMemberStatusActive,
	}
	groupRepo.members[testGroupID+":"+testAdminID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testAdminID,
		Role: group.GroupMemberRoleAdmin, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, topicRepo)
	require.True(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testOwnerID))
	require.True(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testAdminID))
}

func TestCanReadTopic_OwnerOnly_NormalMember_Deny(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityOwnerOnly,
	}

	groupRepo := newMockGroupRepo()
	groupRepo.members[testGroupID+":"+testMemberID] = &group.GroupMember{
		GroupID: testGroupID, UserID: testMemberID,
		Role: group.GroupMemberRoleMember, Status: group.GroupMemberStatusActive,
	}

	svc := newTestService(newMockMemberRepo(), groupRepo, topicRepo)
	require.False(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testMemberID))
}

func TestCanReadTopic_PaidMember_TemporaryDenied(t *testing.T) {
	// 当前 MVP-P0 暂缓 PAID_MEMBER 权益判断，统一返回 false。
	// 此测试锁定已知暂缓行为，避免未来误以为已支持付费阅读。
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityPaidMember,
	}

	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)
	require.False(t, svc.CanReadTopic(context.Background(), testGroupID, testTopicID, testUserID))
}

// ────────────────────────────────────────────────
// CanViewTopicSummary 测试
// ────────────────────────────────────────────────

func TestCanViewTopicSummary_TopicExists_Allow(t *testing.T) {
	topicRepo := newMockTopicRepo()
	topicRepo.topics[testTopicID] = &topic.Topic{
		ID: testTopicID, GroupID: testGroupID, Visibility: topic.TopicVisibilityOwnerOnly,
	}

	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), topicRepo)
	require.True(t, svc.CanViewTopicSummary(context.Background(), testTopicID, testUserID))
}

func TestCanViewTopicSummary_TopicNotExists_Deny(t *testing.T) {
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
	require.False(t, svc.CanViewTopicSummary(context.Background(), "nonexistent", testUserID))
}

// ────────────────────────────────────────────────
// CanAuditContent 暂缓测试（当前固定返回 false）
// ────────────────────────────────────────────────

func TestCanAuditContent_AlwaysFalse(t *testing.T) {
	svc := newTestService(newMockMemberRepo(), newMockGroupRepo(), newMockTopicRepo())
	require.False(t, svc.CanAuditContent(context.Background(), testUserID))
}
