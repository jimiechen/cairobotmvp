package permission

// 权限角色名字符串常量（permission service 内部使用）
// 对应 group.GroupMemberRole 的字符串表示，禁止在代码中使用裸字符串字面量
const (
	RoleStrOwner  = "owner"  // 群主
	RoleStrAdmin  = "admin"  // 管理员
	RoleStrMember = "member" // 普通成员
	RoleStrGuest  = "guest"  // 嘉宾/待审核
)
