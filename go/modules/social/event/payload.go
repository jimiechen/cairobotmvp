package event

// ===== 成员域事件 Payload =====

// MemberRegisteredPayload 用户注册事件载荷
// 触发点：UserRegister 写入 users 表成功后
type MemberRegisteredPayload struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username,omitempty"`
	Nickname        string `json:"nickname,omitempty"`
	MembershipLevel string `json:"membership_level,omitempty"`
}

// UserStatusChangedPayload 用户状态变更事件载荷
// 触发点：UpdateMemberStatus / AdminUpdateMemberStatus 事务提交后
type UserStatusChangedPayload struct {
	UserID     string `json:"user_id"`
	OldStatus  int32  `json:"old_status"`
	NewStatus  int32  `json:"new_status"`
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// ===== 群组域事件 Payload =====

// GroupCreatedPayload 群组创建事件载荷
// 触发点：CreateGroup 写入 groups + owner member 成功后
type GroupCreatedPayload struct {
	GroupID    string `json:"group_id"`
	OwnerID    string `json:"owner_id"`
	Type       string `json:"type"`
	Visibility int8   `json:"visibility"`
	JoinMode   int8   `json:"join_mode"`
}

// GroupJoinedPayload 加入群组事件载荷
// 触发点：JoinGroup 写入 group_members 成功后
type GroupJoinedPayload struct {
	GroupID   string `json:"group_id"`
	UserID    string `json:"user_id"`
	MemberID  string `json:"member_id"`
	Status    int8   `json:"status"`
	JoinSource string `json:"join_source,omitempty"`
	ExpiredAt int64  `json:"expired_at,omitempty"`
}

// GroupLeftPayload 退出群组事件载荷
// 触发点：LeaveGroup 更新 member.status=left 且状态确实变化后
type GroupLeftPayload struct {
	GroupID  string `json:"group_id"`
	UserID   string `json:"user_id"`
	MemberID string `json:"member_id,omitempty"`
}

// GroupMemberChangedPayload 群组成员变更事件载荷（Ban/Remove/Mute/Recover 共享）
// 不同事件的必填字段不同，详见各事件定义：
//   - GroupMemberMuted: 必须包含 MutedUntil
//   - GroupMemberRemoved: 必须包含 OldStatus/NewStatus
//   - GroupMemberBanned: 必须包含 Reason/OperatorID
//   - GroupMemberRecovered: 必须包含 OldStatus/NewStatus
type GroupMemberChangedPayload struct {
	GroupID      string `json:"group_id"`
	OperatorID   string `json:"operator_id"`
	TargetUserID string `json:"target_user_id"`
	Action       string `json:"action"` // 使用 constants.go 中 ActionXxx 常量
	OldStatus    int32  `json:"old_status,omitempty"`
	NewStatus    int32  `json:"new_status,omitempty"`
	MutedUntil   int64  `json:"muted_until,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// GroupPlanCreatedPayload 付费方案创建事件载荷
type GroupPlanCreatedPayload struct {
	GroupID      string `json:"group_id"`
	PlanID       string `json:"plan_id"`
	PlanType     string `json:"plan_type"`
	PriceCent    int64  `json:"price_cent"`
	DurationDays int32  `json:"duration_days"`
}

// GroupOrderPaidPayload 订单支付确认事件载荷
type GroupOrderPaidPayload struct {
	OrderID     string `json:"order_id"`
	OrderNo     string `json:"order_no"`
	GroupID     string `json:"group_id"`
	UserID      string `json:"user_id"`
	PlanID      string `json:"plan_id,omitempty"`
	AmountCent  int64  `json:"amount_cent"`
	PaidAt      int64  `json:"paid_at"`
	ExpiredAt   int64  `json:"expired_at"`
}

// ===== 主题域事件 Payload =====

// TopicCreatedPayload 帖子创建事件载荷
type TopicCreatedPayload struct {
	TopicID     string `json:"topic_id"`
	GroupID     string `json:"group_id"`
	AuthorID    string `json:"author_id"`
	Status      int8   `json:"status"`
	Visibility  int8   `json:"visibility"`
	PublishedAt int64  `json:"published_at,omitempty"`
}

// TopicDeletedPayload 帖子删除事件载荷
type TopicDeletedPayload struct {
	TopicID  string `json:"topic_id"`
	GroupID  string `json:"group_id"`
	AuthorID string `json:"author_id"`
}

// TopicCommentCreatedPayload 评论/回复创建事件载荷
type TopicCommentCreatedPayload struct {
	CommentID string `json:"comment_id"`
	TopicID   string `json:"topic_id"`
	GroupID   string `json:"group_id"`
	UserID    string `json:"user_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Status    int8   `json:"status"`
}

// TopicReactedPayload 帖子互动事件载荷（点赞/收藏/分享）
type TopicReactedPayload struct {
	TopicID      string `json:"topic_id"`
	UserID       string `json:"user_id"`
	ReactionType string `json:"reaction_type"` // 使用 constants.go 中 ReactionTypeXxx 常量
	Status       int8   `json:"status"`         // active / cancelled
}

// TopicAuditPayload 帖子审核事件载荷（Approved/Rejected/Banned 共享）
type TopicAuditPayload struct {
	TopicID    string `json:"topic_id"`
	GroupID    string `json:"group_id"`
	AuthorID   string `json:"author_id"`
	OperatorID string `json:"operator_id"`
	Reason     string `json:"reason,omitempty"`
	AuditedAt  int64  `json:"audited_at"`
}
