package group

// GroupMemberRole 圈子成员角色（对应 group_members.role 字段）
// 值与 proto GroupMemberRole 枚举一一对应，converter.go 可直接使用此常量进行映射
const (
	GroupMemberRoleOwner   int8 = 1 // 群主      → proto GROUP_ROLE_OWNER
	GroupMemberRoleAdmin   int8 = 2 // 管理员    → proto GROUP_ROLE_ADMIN
	GroupMemberRoleMember  int8 = 3 // 普通成员  → proto GROUP_ROLE_MEMBER
	GroupMemberRoleGuest   int8 = 4 // 嘉宾      → proto GROUP_ROLE_GUEST
	GroupMemberRoleBanned  int8 = 5 // 封禁成员  → proto GROUP_ROLE_BANNED
)

// GroupMemberStatus 圈子成员状态（对应 group_members.status 字段）
// 值与模型定义一致：1=正常 2=已退出 3=封禁/移除 4=禁言 5=待审核
const (
	GroupMemberStatusActive  int8 = 1 // 正常成员
	GroupMemberStatusLeft    int8 = 2 // 已退出（主动退出或被踢出）
	GroupMemberStatusBanned  int8 = 3 // 封禁 / 已移除
	GroupMemberStatusMuted   int8 = 4 // 已禁言
	GroupMemberStatusPending int8 = 5 // 待审核（JOIN_MODE_APPROVAL 场景）
)

// GroupStatus 圈子状态（对应 groups.status 字段）
// 值与 proto GroupStatus 枚举一一对应
const (
	GroupStatusActive    int8 = 1 // 活跃       → proto GROUP_STATUS_ACTIVE
	GroupStatusInactive  int8 = 2 // 非活跃     → proto GROUP_STATUS_INACTIVE
	GroupStatusSuspended int8 = 3 // 暂停       → proto GROUP_STATUS_SUSPENDED
	GroupStatusDeleted   int8 = 4 // 已删除     → proto GROUP_STATUS_DELETED
)

// GroupVisibility 圈子可见性（对应 groups.visibility 字段）
// 值与 proto GroupVisibility 枚举一一对应
const (
	GroupVisibilityPublic  int8 = 1 // 公开       → proto GROUP_VISIBILITY_PUBLIC
	GroupVisibilityPrivate int8 = 2 // 私有       → proto GROUP_VISIBILITY_PRIVATE
	GroupVisibilitySecret  int8 = 3 // 秘密       → proto GROUP_VISIBILITY_SECRET
)

// GroupJoinMode 加入方式（对应 groups.join_mode 字段）
// 值与 proto JoinMode 枚举一一对应
const (
	GroupJoinModeOpen       int8 = 1 // 开放加入   → proto JOIN_MODE_OPEN
	GroupJoinModeApproval   int8 = 2 // 需要审批   → proto JOIN_MODE_APPROVAL
	GroupJoinModeInviteOnly int8 = 3 // 仅邀请     → proto JOIN_MODE_INVITE_ONLY
	GroupJoinModePaidOpen   int8 = 4 // 付费公开   → proto JOIN_MODE_PAID_OPEN
)
