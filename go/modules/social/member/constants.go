package member

// UserStatus 用户账号状态（对应 users.status 字段）
// svc 层比较/赋值此字段时必须使用本常量，不得硬编码整数字面量（参见 DevGuide §17）
const (
	UserStatusActive    int8 = 1 // 正常活跃
	UserStatusInactive  int8 = 2 // 未激活（邮箱未验证等）
	UserStatusSuspended int8 = 3 // 封禁 / 暂停
	UserStatusDeleted   int8 = 4 // 已注销
)

// contextKey 是本包专用的 context key 类型，防止跨包 key 碰撞
type contextKey string

// CtxKeyUserID 是认证中间件注入当前登录用户 ID 的 context key。
// 值必须与 Gateway AuthMiddleware.Extend["user_id"] 保持一致（"user_id"）
// 其他包（如 topic、permission）需要读取当前用户 ID 时，应 import member 包并引用此常量。
const CtxKeyUserID contextKey = "user_id"

// Token 类型标识（generateToken / SvcRefresh 使用）
// 禁止在 svc_*.go 中直接使用 "access" / "refresh" 字符串字面量
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)
