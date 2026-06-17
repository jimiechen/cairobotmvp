// Package permission 跨域权限服务，统一权限判断入口
// 所有权限判断必须通过此 Service，禁止在 svc 中直接查表
package permission

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
)

// Service 权限服务接口，8 个能力方法 + 广场虚拟成员特化
type Service interface {
	CanViewGroup(ctx context.Context, groupID, userID string) bool
	CanJoinGroup(ctx context.Context, groupID, userID string) bool
	CanReadTopic(ctx context.Context, groupID, topicID, userID string) bool
	CanViewTopicSummary(ctx context.Context, topicID, userID string) bool
	CanPublishTopic(ctx context.Context, groupID, userID string) bool
	CanManageGroup(ctx context.Context, groupID, userID string) bool
	CanManageMember(ctx context.Context, groupID, operatorID, targetUserID string) bool
	CanAuditContent(ctx context.Context, userID string) bool
}

// service 权限服务实现体，持有群组和成员域的 Repository 引用
//
// 不负责：
// - 缓存读写（权限判断只用 MySQL 1 级数据）
// - 业务逻辑执行（只做权限判定，返回 bool）
// - 审计日志记录（由调用方 svc 负责）
type service struct {
	groupRepo    group.Repository   // 注入群组 Repository（查 group_members 表）
	memberRepo   member.Repository  // 注入成员 Repository（查 users 表用户状态）
	topicRepo    topic.Repository   // 注入帖子 Repository（查 topics 表可见性）
	plazaGroupID string             // 广场虚拟群组固定 ID，从配置中心读取
}

// NewService 创建权限服务实例
//
// 参数:
//   - groupRepo: 群组域 Repository，用于查 group_members
//   - memberRepo: 成员域 Repository，用于查 users 表状态
//   - topicRepo: 帖子域 Repository，用于查 topics 表可见性（getTopicVisibility）
//   - plazaGroupID: 广场群组 ID，从配置中心 social.plaza_group_id 读取
func NewService(groupRepo group.Repository, memberRepo member.Repository, topicRepo topic.Repository, plazaGroupID string) Service {
	return &service{
		groupRepo:    groupRepo,
		memberRepo:   memberRepo,
		topicRepo:    topicRepo,
		plazaGroupID: plazaGroupID,
	}
}

// isPlazaGroup 判断是否为广场群组
func (s *service) isPlazaGroup(groupID string) bool {
	return groupID == s.plazaGroupID
}

// isUserActive 判断用户是否为 active 状态（广场虚拟成员推导基础）
//
// 广场普通成员的"成员身份"由 user.status=active 推导，
// 不依赖 group_members 表中的物理记录。
// 返回 false 时默认拒绝权限（保守策略：DB 异常不越权）。
func (s *service) isUserActive(ctx context.Context, userID string) bool {
	user, err := s.memberRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return false // DB 异常或用户不存在 → 保守拒绝
	}
	return user.Status == member.UserStatusActive
}

// getMemberRole 查询用户在指定群组中的成员角色和状态
//
// 用于非广场群组的标准权限判断。
// 返回 nil 表示用户不是该群组成员（或 DB 异常）。
func (s *service) getMemberRole(ctx context.Context, groupID, userID string) *memberRole {
	m, err := s.groupRepo.GetMember(ctx, groupID, userID)
	if err != nil || m == nil {
		return nil
	}
	// int8 role → string 映射（group/model.go 定义：1=群主 2=管理员 3=普通 4=待审核）
	roleMap := map[int8]string{1: "owner", 2: "admin", 3: "member", 4: "guest"}
	roleStr, ok := roleMap[m.Role]
	if !ok {
		roleStr = RoleStrGuest // 未知角色降级为 guest
	}
	return &memberRole{Role: roleStr, Status: m.Status}
}

// getTopicVisibility 查询帖子的可见性等级和所属群组
//
// 用于 CanReadTopic / CanViewTopicSummary 的可见性判断。
// 通过 topicRepo.GetTopicByID 查询 topics 表（topics 属于 topic 域）。
func (s *service) getTopicVisibility(ctx context.Context, topicID string) *topicVisibility {
	t, err := s.topicRepo.GetTopicByID(ctx, topicID)
	if err != nil || t == nil {
		return nil // DB 异常或帖子不存在 → 保守拒绝
	}
	return &topicVisibility{
		GroupID:    t.GroupID,
		Visibility: t.Visibility,
	}
}

// ── 轻量值对象（避免跨域引入完整 model）────────────────────────────────

// memberRole 用户在群组中的角色与状态
type memberRole struct {
	Role   string // owner / admin / guest / member
	Status int8   // 1=active 2=muted 3=banned 4=left 5=expired
}

// topicVisibility 帖子可见性元数据
type topicVisibility struct {
	GroupID    string
	Visibility int8 // 使用 topic.TopicVisibility* 常量
}

// ── 8 个能力方法实现 ────────────────────────────────────────────────────

// CanViewGroup 判断用户是否可查看群组
//
// 广场特化: active 用户即具备广场查看权限，无需 group_members 记录
// 非广场: 需是群组 active 成员
func (s *service) CanViewGroup(ctx context.Context, groupID, userID string) bool {
	if s.isPlazaGroup(groupID) {
		// 广场群组: active 用户 = 虚拟普通成员，直接放行
		return s.isUserActive(ctx, userID)
	}
	m := s.getMemberRole(ctx, groupID, userID)
	return m != nil && m.Status == 1
}

// CanJoinGroup 判断用户是否可加入群组
//
// 广场特化: 返回 false（所有 active 用户已是虚拟成员，不需主动 Join）
// 非广场: 未加入或已退出/过期 → 可加入
func (s *service) CanJoinGroup(ctx context.Context, groupID, userID string) bool {
	if s.isPlazaGroup(groupID) {
		// 广场群组不需要主动加入，虚拟成员机制自动覆盖
		return false
	}
	m := s.getMemberRole(ctx, groupID, userID)
	// 未加入 或 已离开/过期 → 允许加入
	return m == nil || (m.Status != group.GroupMemberStatusActive)
}

// CanReadTopic 判断用户是否可读取帖子完整内容
//
// 可见性规则:
//   - PUBLIC(1): 所有人可读
//   - GROUP_MEMBER(2): 走群组成员判断（广场走虚拟成员逻辑）
//   - PAID_MEMBER(3): 需付费权益有效
//   - OWNER_ONLY(4): 仅 owner/admin 可读
//
// 铁律: 只查 MySQL 1级数据，禁止使用 Redis 缓存做权限决策
func (s *service) CanReadTopic(ctx context.Context, groupID, topicID, userID string) bool {
	tv := s.getTopicVisibility(ctx, topicID)
	if tv == nil {
		return false
	}
	switch tv.Visibility {
	case topic.TopicVisibilityPublic: // PUBLIC: 直接放行
		return true
	case topic.TopicVisibilityGroupMember: // GROUP_MEMBER: 复用 CanViewGroup 逻辑（含广场虚拟成员特化）
		return s.CanViewGroup(ctx, tv.GroupID, userID)
	case topic.TopicVisibilityPaidMember: // PAID_MEMBER: 需检查权益有效性
		// TODO(group, MVP): 调用 groupRepo.GetEntitlementByGroupAndUser
		// 检查 entitlement.status==1 && entitlement.expired_at > now()
		return false
	case topic.TopicVisibilityOwnerOnly: // OWNER_ONLY: 仅 owner/admin
		m := s.getMemberRole(ctx, tv.GroupID, userID)
		return m != nil && (m.Role == RoleStrOwner || m.Role == RoleStrAdmin)
	}
	return false
}

// CanViewTopicSummary 判断用户是否可查看帖子摘要
//
// 权限比 CanReadTopic 宽松: visibility >= 1 即可查看摘要
// 广场和非广场行为一致
func (s *service) CanViewTopicSummary(ctx context.Context, topicID, userID string) bool {
	tv := s.getTopicVisibility(ctx, topicID)
	if tv == nil {
		return false
	}
	// visibility >= 1 表示帖子存在且可见（所有级别均可看摘要）
	return tv.Visibility >= 1
}

// CanPublishTopic 判断用户是否可在群组中发布帖子或回复
//
// 广场特化:
//   - active 且未被 ban 则可发帖
//   - ban 记录落在 group_members 表（物理记录），找不到记录视为普通虚拟成员
// 非广场: 需是 active 成员
func (s *service) CanPublishTopic(ctx context.Context, groupID, userID string) bool {
	if s.isPlazaGroup(groupID) {
		// 先检查是否有 ban 记录（ban 是物理记录，必须查 group_members）
		m := s.getMemberRole(ctx, groupID, userID)
		if m != nil && m.Status == 3 { // 3=banned
			return false
		}
		// 无 ban 记录 + 用户 active = 可发帖
		return s.isUserActive(ctx, userID)
	}
	// 非广场: 必须是 active 成员
	m := s.getMemberRole(ctx, groupID, userID)
	return m != nil && m.Status == 1
}

// CanManageGroup 判断用户是否可管理群组（圈主/管理员鉴权）
//
// 广场特化: 普通用户返回 false（广场管理需显式 admin/owner 角色）
// 非广场: role 为 owner 或 admin
func (s *service) CanManageGroup(ctx context.Context, groupID, userID string) bool {
	m := s.getMemberRole(ctx, groupID, userID)
	if m == nil {
		return false
	}
	return m.Role == "owner" || m.Role == "admin"
}

// CanManageMember 判断操作者是否可管理目标成员
//
// 角色层级: owner > admin > member，同级不可互操作
// 广场特化: 返回 false（广场成员管理由平台管理员处理）
//
// 层级规则:
//   - owner 可管理 admin 和 member（不可管理其他 owner）
//   - admin 仅可管理普通 member
//   - member 无权管理任何人
func (s *service) CanManageMember(ctx context.Context, groupID, operatorID, targetUserID string) bool {
	op := s.getMemberRole(ctx, groupID, operatorID)
	if op == nil {
		return false
	}
	target := s.getMemberRole(ctx, groupID, targetUserID)
	if target == nil {
		return false
	}
	switch op.Role {
	case RoleStrOwner:
		// owner 可管理除自己以外的所有角色
		return target.Role != RoleStrOwner
	case RoleStrAdmin:
		// admin 仅可管理普通 member
		return target.Role == RoleStrMember
	}
	// member / guest 无权管理其他成员
	return false
}

// CanAuditContent 判断用户是否可审核内容（平台管理员专属）
//
// 不依赖群组上下文，仅检查用户的全局管理员角色。
// 通过 memberRepo 查询用户角色标识。
func (s *service) CanAuditContent(ctx context.Context, userID string) bool {
	// TODO(member, MVP): 调用 memberRepo 查询用户是否具有平台管理员角色
	// 可能的实现方式:
	//   1. users 表有 is_platform_admin 字段
	//   2. 或独立的 platform_admins 表
	//   3. 或通过 membership_level / 特殊 role 判断
	// 待确认具体方案后补充查询逻辑
	return false
}
