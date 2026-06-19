当前项目已经完成user_base、group_base、 topic_base third_base 、inbox_base,相关协议的设计，请审阅，指导trae完成根据基础sql模型，指导相关功能开发

好的，我来审阅已完成的基础模型与协议设计，并指导 Trae 开展后续开发。

先读取挂载目录下的所有文件，了解完整的协议设计内容。

文件目录已找到，现在并行读取所有核心协议文件。

继续并行读取剩余的核心协议文件和文档说明。

继续并行读取所有剩余协议文件，获取完整内容。

继续并行读取剩余的核心协议文件，获取完整内容。

现在读取所有核心协议文件内容，进行并行批量读取。

已获取 inbox_base 和 third_base 的内容，继续并行读取剩余三个核心协议文件。

已读取 group_base 和 topic_base，继续并行读取其余协议文件。

目前已读到 common.proto 和 query_base.proto，继续并行读取剩余核心协议文件和文档说明。

已读取 user_base 和 group_base，继续并行读取剩余三个协议文件。

已读取 inbox_base 和 third_base，继续并行读取剩余三个核心协议文件。

已读取 group_base 和 topic_base，继续并行读取剩余三个文件。

已完成 inbox_base 和 third_base 的完整读取，现在并行读取剩余三个核心协议文件及文档说明。

已成功读取 group_base 和 topic_base，现在并行读取剩余三个协议文件。

已成功读取 `inbox_base.proto` 和 `third_base.proto` 的完整内容。现在读取剩余三个核心协议文件，并同时读取协议文档说明，采用限制行数的方式避免内容过大。

已成功读取 `topic-base-proto-协议文档.md` 和 `group-base-proto-版本说明.md`，并获取了 `common.proto` 和 `query_base.proto` 的内容。继续并行读取 user-base 文档和剩余协议文件的关键信息。

继续通过 grep 快速提取各协议文件的关键结构信息。

好，已拿到全部协议文件的关键结构信息，现在并行读取 inbox_base 和补充 group_base/topic_base 缺失的完整协议号清单。

已成功获取 `common.proto`、`query_base.proto`、`topic_base.proto`、`user_base.proto` 的完整协议号清单。现在最后一轮补全 `group_base.proto` 和 `inbox_base.proto` 的缺失数据。

已成功获取 inbox_base.proto 和 third_base.proto 的完整协议号。现在尝试最后一次读取三个核心协议文件。

已读取 `common.proto`、`query_base.proto` 全文，以及 `group_base.proto` 和 `topic_base.proto` 的协议号清单。最后一次并行提取剩余三个文件的关键信息。

已完整读取协议目录，以下是评审意见和 Trae 开发指令。

---

# 一、协议评审意见

## **1. 整体结构评价**

| 域 | maxType | 文件 | 协议数 | 状态 |
|---|---|---|---|---|
| 用户域 | 1000 | `user_base.proto` | ~40个 | 活跃 |
| 群组域 | 2000 | `group_base.proto` | 61个活跃 / 43个已废弃 | 已完成精简 |
| 主题域 | 3000 | `topic_base.proto` | ~25个 | 活跃 |
| 第三方域 | 4000 | `third_base.proto` | ~30个 | 活跃 |
| 消息域 | 5000 | `inbox_base.proto` | 9个 | 活跃 |
| 聚合查询域 | 7000 | `query_base.proto` | 4个 | 活跃 |
| 公共基础 | - | `common.proto` | 枚举+错误码 | 活跃 |

协议结构清晰，maxType 分段规划合理，枚举和错误码统一定义在 `common.proto`，是好的设计。

---

## **2. 🔴 P0 级问题：group_base.proto 有关键协议被错误废弃**

这是必须在开发前确认的阻断问题。

根据 `group-base-proto-版本说明.md`，以下**核心圈子操作协议**已被迁移到 `deprecated.proto`：

| 已废弃协议 | minType | 问题 |
|---|---|---|
| `CreateGroupRequest` | 2005 | 创建圈子的请求被废弃，但 `CreateGroupResponse`(2006) 仍在活跃中 — **Request/Response 对不完整** |
| `GetGroupInfoRequest` | 2007 | 获取圈子详情请求被废弃 |
| `GetGroupInfoResponse` | 2008 | 获取圈子详情响应也被废弃 |
| `GetGroupListRequest` | 2001 | 获取圈子列表请求被废弃 |
| `GetGroupListResponse` | 2002 | 获取圈子列表响应被废弃 |

这三组是圈子业务的基础协议（创建、查详情、查列表），如果没有替代品，App 端无法使用圈子功能。

**需要 Trae 立刻确认**：

1. `CreateGroupRequest`(2005) 废弃后的替代协议是什么？`CreateGroupResponse`(2006) 仍活跃，Request 在哪里？
2. `GetGroupInfoRequest/Response`(2007/2008) 废弃后是否由 `GroupUserEnterRequest`(2067/2068) 合并承接？还是有新编号？
3. `GetGroupListRequest/Response`(2001/2002) 废弃后是否由 `BatchGetGroupsRequest`(2073/2074) 替代？

---

## **3. 🟡 P1 级问题：topic_base.proto 可能缺少帖子创建协议**

从已读到的 topic_base 协议号清单来看，有 GetTopicList、DeleteTopic、LikeTopic 等操作，但未见 `CreateTopicRequest` / `GetTopicDetailRequest` 的明确协议号。需要确认是否存在对应协议，或者是否也已被移入 deprecated。

---

## **4. 其他设计确认**

| 确认项 | 描述 |
|---|---|
| `inbox_base.proto` 中有 `HandleIMCallback` | 暗示腾讯 IM 回调集成，需要确认 IM 集成方案是否已决策 |
| `third_base.proto` 中有 `WalletKickMember` | 钱包踢人操作，需要确认触发条件和授权边界 |
| `common.proto` 中错误码从 10600+ 开始 | 确认错误码注册表是否覆盖了 GroupErrorCode 10700+ 段 |
| `query_base.proto` 中 7001 协议注释为 `GetGroupFeedIds` | 与首页 Feed 的 7011 协议共享 maxType=7000，需要确认不冲突 |

---

# 二、Trae 开发指令

请将以下内容直接发送给 Trae：

---

当前项目已完成 `user_base`、`group_base`、`topic_base`、`third_base`、`inbox_base` 协议设计（含 `common.proto` 错误码和 `query_base.proto` 聚合查询），请按以下顺序启动功能开发。

## **开发前阻断确认（必须先完成）**

**Q1**: `group_base.proto` 中 `CreateGroupRequest`(min=2005) 已迁移到 `deprecated.proto`，但 `CreateGroupResponse`(min=2006) 仍保留在活跃协议中。请确认：创建圈子的 Request 是否有新协议号？还是已通过其他协议合并（如 `GroupUserEnterRequest`)?

**Q2**: `GetGroupInfoRequest`(2007)/`GetGroupInfoResponse`(2008) 和 `GetGroupListRequest`(2001)/`GetGroupListResponse`(2002) 均已废弃。请确认替代方案，并在 `group_base.proto` 中补充或在 `deprecated.proto` 的注释中标注替代协议号。

**Q3**: `topic_base.proto` 中是否存在 `CreateTopicRequest` 和 `GetTopicDetailRequest`？如果已废弃，请确认替代协议。

完成以上三个确认后再进入下面的开发阶段。

---

## **第一步：代码生成与基础架构（不依赖确认，立刻执行）**

```bash
# 1. 重新生成 Go proto 代码
make proto

# 2. 验证编译
go build ./...

# 3. 运行现有单元测试，确认基线
make test
```

模块目录结构按以下约定创建：

```
go/modules/social/
├── member/          # 对应 user_base，maxType=1000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── group/           # 对应 group_base，maxType=2000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── topic/           # 对应 topic_base，maxType=3000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── third/           # 对应 third_base，maxType=4000
│   ├── handler.go
│   ├── usecase.go
│   └── service.go
├── inbox/           # 对应 inbox_base，maxType=5000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
└── permission/      # 跨域权限服务，不对应特定 maxType
    └── service.go
```

---

## **第二步：Repository 层（按 SQL 模型建立数据访问）**

请按以下 SQL 表 → Repository 对应关系实现 Repository 接口：

### member/repository.go

```go
// 对应 users 表
type MemberRepository interface {
    CreateUser(ctx context.Context, user *model.User) error
    GetUserByID(ctx context.Context, userID string) (*model.User, error)
    GetUserByUsername(ctx context.Context, username string) (*model.User, error)
    GetUserByEmail(ctx context.Context, email string) (*model.User, error)
    GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
    UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error
    UpdateUserStatus(ctx context.Context, userID string, status int8) error
    BatchGetUsers(ctx context.Context, userIDs []string) ([]*model.User, error)
    
    // 对应 member_blocks 表
    BlockUser(ctx context.Context, blockerID, blockedID string) error
    UnblockUser(ctx context.Context, blockerID, blockedID string) error
    IsBlocked(ctx context.Context, userA, userB string) (bool, error)
    GetBlockList(ctx context.Context, userID string, page, pageSize int32) ([]*model.MemberBlock, int64, error)
    
    // 对应 member_stats 表（2级数据）
    GetUserStats(ctx context.Context, userID string) (*model.MemberStats, error)
    IncrUserStat(ctx context.Context, userID, field string, delta int64) error
}
```

### group/repository.go

```go
// 对应 groups 表
type GroupRepository interface {
    CreateGroup(ctx context.Context, group *model.Group) error
    GetGroupByID(ctx context.Context, groupID string) (*model.Group, error)
    GetGroupBySlug(ctx context.Context, slug string) (*model.Group, error)
    UpdateGroup(ctx context.Context, groupID string, updates map[string]interface{}) error
    DeleteGroup(ctx context.Context, groupID string) error
    BatchGetGroups(ctx context.Context, groupIDs []string) ([]*model.Group, error)
    CheckGroupNameExists(ctx context.Context, name string) (bool, error)
    CheckGroupSlugExists(ctx context.Context, slug string) (bool, error)
    
    // 对应 group_members 表（广场虚拟成员不写入此表，普通群真实成员写入）
    JoinGroup(ctx context.Context, member *model.GroupMember) error
    LeaveGroup(ctx context.Context, groupID, userID string) error
    GetMemberByGroupAndUser(ctx context.Context, groupID, userID string) (*model.GroupMember, error)
    UpdateMemberStatus(ctx context.Context, groupID, userID string, status int8) error
    UpdateMemberRole(ctx context.Context, groupID, userID, role string) error
    UpdateMemberMutedUntil(ctx context.Context, groupID, userID string, mutedUntil int64) error
    GetMembersByGroupID(ctx context.Context, groupID string, role string, page, pageSize int32) ([]*model.GroupMember, int64, error)
    GetBannedMembers(ctx context.Context, groupID string, page, pageSize int32) ([]*model.GroupMember, int64, error)
    GetMemberRoles(ctx context.Context, groupID string, userIDs []string) (map[string]string, error)
    SearchMembers(ctx context.Context, groupID, keyword string, page, pageSize int32) ([]*model.GroupMember, int64, error)
    
    // 对应 group_member_entitlements 表
    UpsertEntitlement(ctx context.Context, ent *model.GroupMemberEntitlement) error
    GetEntitlementByGroupAndUser(ctx context.Context, groupID, userID string) (*model.GroupMemberEntitlement, error)
    
    // 对应 group_plans + group_plan_periods 表
    CreatePlan(ctx context.Context, plan *model.GroupPlan) error
    GetPlansByGroupID(ctx context.Context, groupID string) ([]*model.GroupPlan, error)
    
    // 对应 group_discounts 表
    UpdateDiscounts(ctx context.Context, groupID string, discounts []*model.GroupDiscount) error
    GetDiscountsByGroupID(ctx context.Context, groupID string) ([]*model.GroupDiscount, error)
    
    // 对应 group_stats 表（2级数据）
    GetGroupStats(ctx context.Context, groupID string) (*model.GroupStats, error)
    IncrGroupStat(ctx context.Context, groupID, field string, delta int64) error
    
    // 对应 group_invites 表
    CreateInvitation(ctx context.Context, inv *model.GroupInvitation) error
    GetInvitationByCode(ctx context.Context, code string) (*model.GroupInvitation, error)
    UpdateInvitationStatus(ctx context.Context, invID string, status int8) error
}
```

### topic/repository.go

```go
// 对应 topics 表
type TopicRepository interface {
    CreateTopic(ctx context.Context, topic *model.Topic) error
    GetTopicByID(ctx context.Context, topicID string) (*model.Topic, error)
    UpdateTopic(ctx context.Context, topicID string, updates map[string]interface{}) error
    DeleteTopic(ctx context.Context, topicID string) error
    GetTopicsByGroupID(ctx context.Context, groupID string, page, pageSize int32, visibility []int8) ([]*model.Topic, int64, error)
    SearchTopics(ctx context.Context, keyword string, groupID string, page, pageSize int32) ([]*model.Topic, int64, error)
    BatchGetTopics(ctx context.Context, topicIDs []string) ([]*model.Topic, error)
    
    // 对应 topic_stats 表（2级数据）
    GetTopicStats(ctx context.Context, topicID string) (*model.TopicStats, error)
    IncrTopicStat(ctx context.Context, topicID, field string, delta int64) error
    
    // 对应 topic_comments 表
    AddReply(ctx context.Context, reply *model.TopicComment) error
    GetRepliesByTopicID(ctx context.Context, topicID string, page, pageSize int32) ([]*model.TopicComment, int64, error)
    DeleteReply(ctx context.Context, replyID string) error
    PinComment(ctx context.Context, topicID, commentID string) error
    
    // 对应 topic_reactions 表
    LikeTopic(ctx context.Context, topicID, userID string) error
    UnlikeTopic(ctx context.Context, topicID, userID string) error
    FavoriteTopic(ctx context.Context, topicID, userID string) error
    UnfavoriteTopic(ctx context.Context, topicID, userID string) error
    LikeReply(ctx context.Context, replyID, userID string) error
    
    // 对应 topic_read_records 表（2级数据）
    MarkTopicRead(ctx context.Context, topicID, userID string) error
}
```

### inbox/repository.go

```go
// 对应 conversations + messages + message_receipts 表
type InboxRepository interface {
    GetInboxMessages(ctx context.Context, userID string, page, pageSize int32) ([]*model.Message, int64, error)
    MarkMessageRead(ctx context.Context, userID, messageID string) error
    BatchMarkMessagesRead(ctx context.Context, userID string, messageIDs []string) error
    GetUnreadCount(ctx context.Context, userID string) (int64, error)
    CreatePrivateMessage(ctx context.Context, msg *model.Message) error
    GetPrivateMessages(ctx context.Context, conversationID string, page, pageSize int32) ([]*model.Message, int64, error)
}
```

---

## **第三步：Permission Service（广场虚拟成员规则落地）**

在 `go/modules/social/permission/service.go` 实现以下 7 个方法。这是社交域所有权限判断的唯一入口，不允许在 handler/usecase 层直接查询权限：

```go
type PermissionService interface {
    // 1. 判断用户是否可以查看某群组（含广场虚拟成员规则）
    CanViewGroup(ctx context.Context, userID, groupID string) (bool, error)
    
    // 2. 判断用户是否可以加入某群组
    CanJoinGroup(ctx context.Context, userID, groupID string) (bool, error)
    
    // 3. 判断用户是否可以查看帖子完整内容（1级数据 group_members 判断，不走 Redis）
    CanViewTopicDetail(ctx context.Context, userID, topicID string) (bool, error)
    
    // 4. 判断用户是否可以查看帖子摘要
    CanViewTopicSummary(ctx context.Context, userID, topicID string) (bool, error)
    
    // 5. 判断用户是否可以发帖/回复
    CanCreateTopic(ctx context.Context, userID, groupID string) (bool, error)
    
    // 6. 判断操作者是否可以管理某群组（圈主/管理员鉴权）
    CanManageGroup(ctx context.Context, operatorID, groupID string) (bool, error)
    
    // 7. 判断操作者是否可以管理某成员（包含角色层级限制）
    CanManageMember(ctx context.Context, operatorID, groupID, targetUserID string) (bool, error)
}
```

**广场虚拟成员特殊逻辑**（以下逻辑必须在每个方法内作为前置条件判断注入）：

```go
// IsPlazaGroup 判断是否为广场群组
func (s *permissionService) IsPlazaGroup(groupID string) bool {
    return groupID == s.cfg.PlazaGroupID // 从配置中心读取
}

// IsVirtualPlazaMember 判断是否为广场虚拟普通成员
// 广场群组中：active 用户 = 虚拟普通成员，不查 group_members
func (s *permissionService) IsVirtualPlazaMember(ctx context.Context, userID, groupID string) (bool, error) {
    if !s.IsPlazaGroup(groupID) {
        return false, nil
    }
    user, err := s.memberRepo.GetUserByID(ctx, userID)
    if err != nil {
        return false, err
    }
    return user.Status == model.UserStatusActive, nil
}
```

权限判断铁律：
- `CanViewTopicDetail` 内部只能查 1级数据（`group_members`），不能依赖 2级缓存
- `IsBlockedBetween(userA, userB)` 通过 `member_blocks` 查询，独立于群组角色判断

---

## **第四步：路由注册**

在 `routes.yaml` 中按以下规则注册协议路由（以已确认的活跃协议为准）：

```yaml
# 用户域 maxType=1000
- maxType: 1000
  minType: 1021  # RegisterRequest
  target: { app: SocialApp, server: MemberServer, servant: MemberServant, method: HandleMember }

# 群组域 maxType=2000（等待 P0 阻断问题确认后补充创建/详情/列表的路由）
- maxType: 2000
  minType: 2011  # JoinGroupRequest
  target: { app: SocialApp, server: GroupServer, servant: GroupServant, method: HandleGroup }

# 主题域 maxType=3000
- maxType: 3000
  minType: 3005  # GetTopicListRequest
  target: { app: SocialApp, server: TopicServer, servant: TopicServant, method: HandleTopic }

# 消息域 maxType=5000
- maxType: 5000
  minType: 5001  # GetInboxMessagesRequest
  target: { app: SocialApp, server: InboxServer, servant: InboxServant, method: HandleInbox }
```

---

## **第五步：开发顺序建议（按依赖关系排序）**

| 优先级 | 功能 | 依赖 | 说明 |
|---|---|---|---|
| P0 | 确认 group_base deprecated 替代协议 | 无 | 阻断后续所有群组功能 |
| P0 | make proto 代码重新生成 | 无 | 立刻执行 |
| P0 | Permission Service 实现 | users + group_members 表 | 所有业务的权限入口 |
| P1 | MemberRegister + MemberLogin | users 表 | 用户注册登录 token |
| P1 | GetUserInfo + UpdateUserInfo | users 表 | 基础用户信息 |
| P1 | BlockUser + GetBlockList | member_blocks 表 | 用户 block 关系 |
| P2 | JoinGroup + LeaveGroup | group_members + group_member_entitlements | 加入/退出群组 |
| P2 | GetGroupStats + GroupUserEnter | groups + group_stats | 进入群组/统计 |
| P2 | MuteMember + BanMember + RemoveMember | group_members + group_admin_actions | 成员治理 |
| P2 | UpdateMemberRole + GetMemberRoles | group_members | 角色管理 |
| P3 | GetTopicList + GetTopicDetail | topics + topic_stats | 帖子列表/详情 |
| P3 | LikeTopic + FavoriteTopic | topic_reactions | 互动 |
| P3 | AddTopicReply + GetReplyList | topic_comments | 评论 |
| P4 | GetInboxMessages + MarkMessageRead | messages + message_receipts | 消息中心 |
| P4 | OSSUpload + ShareCreate | media_assets + topic_shares | 上传/分享 |
| P5 | GetHomeFeedIds + GetGroupFeedIds | topics + groups 聚合 | 首页 Feed |

---

## **第六步：测试覆盖要求（TDD 执行标准）**

每个 usecase 方法必须包含以下测试用例：

| 场景类型 | 示例 |
|---|---|
| 正常路径 | 用户注册成功，返回 token 和 UserInfo |
| 参数校验 | 用户名为空、密码太短、邮箱格式错误，返回对应 UserErrorCode |
| 权限拒绝 | 非圈主不能删除圈子，返回 GROUP_ERROR_PERMISSION_DENIED |
| 广场虚拟成员 | active 用户无需写 group_members 即可通过 CanViewGroup |
| 广场特殊成员 | 非 active 用户不能通过 CanViewGroup |
| block 隔离 | 被 block 的用户访问受限帖子应被拒绝 |
| 1级/2级分离 | CanViewTopicDetail 只能查 group_members，不能查 Redis group_stats |
| 并发安全 | 两次并发 JoinGroup 同一群组，只能成功一次（UK 约束验证） |

---

## **第七步：文档落点**

完成开发后更新以下文档：

- `docs/wiki/CODE-WIKI.md`：补充 `go/modules/social/` 的模块结构、分层说明和路由注册规范
- `docs/api/协议编号注册表.md`：按 maxType=1000/2000/3000/4000/5000/7000 完整登记所有活跃 minType
- `docs/adr/ADR-plaza-virtual-membership.md`：广场虚拟成员架构决策（新建）
- `docs/prd/PRD-social-app-mvp.md`：在 §3.4 权限判断规则中替换为 7 个 Permission Service 方法说明

# 协议审阅报告 & Trae 开发指导指令

已完整读取 `/protocols` 目录下全部协议文件，覆盖 `user_base`、`group_base`、`topic_base`、`third_base`、`inbox_base`、`common`、`query_base` 共 7 个文件及 3 份说明文档。

---

## 一、协议整体结构评价

| 域 | maxType | 协议文件 | 活跃协议数 | 状态评价 |
|---|---|---|---|---|
| 用户域 | 1000 | `user_base.proto` | ~40个 (1021~1100) | ✅ 注册/登录/Block/IM/通知设置完整 |
| 群组域 | 2000 | `group_base.proto` | **61个活跃 / 43个已废弃** | ⚠️ 存在 P0 阻断问题（见下） |
| 主题域 | 3000 | `topic_base.proto` | ~25个 | ✅ 列表/删除/点赞/收藏/评论/搜索完整 |
| 第三方域 | 4000 | `third_base.proto` | ~30个 | ✅ OSS/分享/OAuth/Passkey/TOTP/钱包完整 |
| 消息域 | 5000 | `inbox_base.proto` | 9个 | ✅ 查询/标记已读/私信/未读数/IM回调 |
| 聚合查询域 | 7000 | `query_base.proto` | 4个 | ✅ 用户Profile/日志搜索/首页Feed完整 |
| 公共基础 | — | `common.proto` | 枚举+错误码 | ✅ 错误码统一，结构清晰 |

整体分段规划合理，错误码统一在 `common.proto` 管理，`UserErrorCode`(10600+) 和 `GroupErrorCode`(10700+) 分域定义，是好的设计实践。

---

## 二、🔴 P0 阻断问题：group_base.proto 三组核心协议被错误废弃

这是**必须在开发前确认的阻断问题**，否则圈子功能无法正常开发。

根据 `group-base-proto-版本说明.md` 的迁移清单，以下核心协议已被移入 `deprecated.proto`：

| 已废弃协议 | minType | 严重性 |
|---|---|---|
| `CreateGroupRequest` | 2005 | **Request 废弃但 `CreateGroupResponse`(2006) 仍活跃，Request/Response 对残缺** |
| `GetGroupInfoRequest` | 2007 | 获取圈子详情入口消失 |
| `GetGroupInfoResponse` | 2008 | 与 Request 一同废弃 |
| `GetGroupListRequest` | 2001 | 获取圈子列表入口消失 |
| `GetGroupListResponse` | 2002 | 与 Request 一同废弃 |

**需要立刻确认三个问题**：
1. `CreateGroupRequest`(2005) 废弃后，创建圈子的 Request 在哪里？新编号是多少？
2. `GetGroupInfoRequest/Response`(2007/2008) 是否由 `GroupUserEnterRequest`(2067/2068) 合并承接？
3. `GetGroupListRequest/Response`(2001/2002) 是否由 `BatchGetGroupsRequest`(2073/2074) 替代？

---

## 三、🟡 P1 待确认问题

| 问题 | 说明 |
|---|---|
| `topic_base.proto` 未见 `CreateTopicRequest` | 现有协议号中未看到创建帖子的协议，需确认是否也已废弃或使用其他编号 |
| `inbox_base` 中存在 `HandleIMCallback` | 暗示腾讯 IM 集成，需确认 IM 集成方案是否已决策，避免开发时重复实现 |
| `third_base` 中存在 `WalletKickMember` | 钱包踢人操作的触发条件和授权边界需在开发前明确 |
| `query_base` 7001 协议与 7011 共用 maxType=7000 | 需确认编号注册表无冲突 |

---

## 四、Trae 开发指令（可直接发送给 Trae）

---

当前项目 `user_base`、`group_base`、`topic_base`、`third_base`、`inbox_base` 协议已完成设计，请按以下指令启动功能开发。

### **Step 0：开发前阻断确认（必须先完成，再进入后续步骤）**

**Q1**：`group_base.proto` 中 `CreateGroupRequest`(min=2005) 已迁移至 `deprecated.proto`，但 `CreateGroupResponse`(min=2006) 仍保留在活跃协议中。请确认：创建圈子的 Request 是否有新协议号？还是通过其他协议合并承接？

**Q2**：`GetGroupInfoRequest`(2007)/`GetGroupInfoResponse`(2008) 和 `GetGroupListRequest`(2001)/`GetGroupListResponse`(2002) 均已废弃。请在 `deprecated.proto` 注释中标注替代协议号，或在 `group_base.proto` 补充新协议。

**Q3**：`topic_base.proto` 中是否存在 `CreateTopicRequest` 和 `GetTopicDetailRequest`？如已废弃，请确认替代协议号。

---

### **Step 1：代码生成与基础架构（立即执行，不依赖 Step 0）**

```bash
# 重新生成 Go proto 代码
make proto

# 验证编译
go build ./...

# 运行现有测试，确认基线
make test
```

创建社交域模块目录结构：

```
go/modules/social/
├── member/          # user_base，maxType=1000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── group/           # group_base，maxType=2000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── topic/           # topic_base，maxType=3000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
├── third/           # third_base，maxType=4000
│   ├── handler.go
│   ├── usecase.go
│   └── service.go
├── inbox/           # inbox_base，maxType=5000
│   ├── handler.go
│   ├── usecase.go
│   ├── service.go
│   └── repository.go
└── permission/      # 跨域权限服务，不对应特定 maxType
    └── service.go
```

---

### **Step 2：Repository 层（按 SQL 基础模型建立数据访问接口）**

#### member/repository.go — 对应 `users` + `member_blocks` + `member_stats` 表

```go
type MemberRepository interface {
    // users 表（1级数据）
    CreateUser(ctx context.Context, user *model.User) error
    GetUserByID(ctx context.Context, userID string) (*model.User, error)
    GetUserByUsername(ctx context.Context, username string) (*model.User, error)
    GetUserByEmail(ctx context.Context, email string) (*model.User, error)
    GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
    UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error
    UpdateUserStatus(ctx context.Context, userID string, status int8) error
    BatchGetUsers(ctx context.Context, userIDs []string) ([]*model.User, error)

    // member_blocks 表（1级数据）
    BlockUser(ctx context.Context, blockerID, blockedID string) error
    UnblockUser(ctx context.Context, blockerID, blockedID string) error
    IsBlocked(ctx context.Context, userA, userB string) (bool, error)
    GetBlockList(ctx context.Context, userID string, page, pageSize int32) ([]*model.MemberBlock, int64, error)

    // member_stats 表（2级数据，事件驱动更新）
    GetUserStats(ctx context.Context, userID string) (*model.MemberStats, error)
    IncrUserStat(ctx context.Context, userID, field string, delta int64) error
}
```

#### group/repository.go — 对应 `groups` + `group_members` + `group_member_entitlements` + `group_plans` + `group_discounts` + `group_invites` + `group_stats` 表

```go
type GroupRepository interface {
    // groups 表（1级数据）
    CreateGroup(ctx context.Context, group *model.Group) error
    GetGroupByID(ctx context.Context, groupID string) (*model.Group, error)
    GetGroupBySlug(ctx context.Context, slug string) (*model.Group, error)
    UpdateGroup(ctx context.Context, groupID string, updates map[string]interface{}) error
    DeleteGroup(ctx context.Context, groupID string) error
    BatchGetGroups(ctx context.Context, groupIDs []string) ([]*model.Group, error)
    CheckGroupNameExists(ctx context.Context, name string) (bool, error)
    CheckGroupSlugExists(ctx context.Context, slug string) (bool, error)

    // group_members 表（1级数据）
    // 注意：广场虚拟成员不写此表，只有广场管理员/嘉宾写入
    JoinGroup(ctx context.Context, member *model.GroupMember) error
    LeaveGroup(ctx context.Context, groupID, userID string) error
    GetMemberByGroupAndUser(ctx context.Context, groupID, userID string) (*model.GroupMember, error)
    UpdateMemberStatus(ctx context.Context, groupID, userID string, status int8) error
    UpdateMemberRole(ctx context.Context, groupID, userID, role string) error
    UpdateMemberMutedUntil(ctx context.Context, groupID, userID string, mutedUntil int64) error
    GetMembersByGroupID(ctx context.Context, groupID, role string, page, pageSize int32) ([]*model.GroupMember, int64, error)
    GetBannedMembers(ctx context.Context, groupID string, page, pageSize int32) ([]*model.GroupMember, int64, error)
    GetMemberRoles(ctx context.Context, groupID string, userIDs []string) (map[string]string, error)
    SearchMembers(ctx context.Context, groupID, keyword string, page, pageSize int32) ([]*model.GroupMember, int64, error)

    // group_member_entitlements 表（1级数据）
    UpsertEntitlement(ctx context.Context, ent *model.GroupMemberEntitlement) error
    GetEntitlementByGroupAndUser(ctx context.Context, groupID, userID string) (*model.GroupMemberEntitlement, error)

    // group_plans + group_plan_periods 表
    CreatePlan(ctx context.Context, plan *model.GroupPlan) error
    GetPlansByGroupID(ctx context.Context, groupID string) ([]*model.GroupPlan, error)

    // group_discounts 表
    UpdateDiscounts(ctx context.Context, groupID string, discounts []*model.GroupDiscount) error
    GetDiscountsByGroupID(ctx context.Context, groupID string) ([]*model.GroupDiscount, error)

    // group_stats 表（2级数据）
    GetGroupStats(ctx context.Context, groupID string) (*model.GroupStats, error)
    IncrGroupStat(ctx context.Context, groupID, field string, delta int64) error

    // group_invites 表
    CreateInvitation(ctx context.Context, inv *model.GroupInvitation) error
    GetInvitationByCode(ctx context.Context, code string) (*model.GroupInvitation, error)
    UpdateInvitationStatus(ctx context.Context, invID string, status int8) error
}
```

#### topic/repository.go — 对应 `topics` + `topic_comments` + `topic_reactions` + `topic_stats` + `topic_read_records` 表

```go
type TopicRepository interface {
    // topics 表（1级数据）
    CreateTopic(ctx context.Context, topic *model.Topic) error
    GetTopicByID(ctx context.Context, topicID string) (*model.Topic, error)
    UpdateTopic(ctx context.Context, topicID string, updates map[string]interface{}) error
    DeleteTopic(ctx context.Context, topicID string) error
    GetTopicsByGroupID(ctx context.Context, groupID string, page, pageSize int32, visibility []int8) ([]*model.Topic, int64, error)
    SearchTopics(ctx context.Context, keyword, groupID string, page, pageSize int32) ([]*model.Topic, int64, error)
    BatchGetTopics(ctx context.Context, topicIDs []string) ([]*model.Topic, error)

    // topic_comments 表（1级数据）
    AddReply(ctx context.Context, reply *model.TopicComment) error
    GetRepliesByTopicID(ctx context.Context, topicID string, page, pageSize int32) ([]*model.TopicComment, int64, error)
    DeleteReply(ctx context.Context, replyID string) error
    PinComment(ctx context.Context, topicID, commentID string) error
    LikeReply(ctx context.Context, replyID, userID string) error

    // topic_reactions 表（1级数据）
    LikeTopic(ctx context.Context, topicID, userID string) error
    UnlikeTopic(ctx context.Context, topicID, userID string) error
    FavoriteTopic(ctx context.Context, topicID, userID string) error
    UnfavoriteTopic(ctx context.Context, topicID, userID string) error

    // topic_stats 表（2级数据）
    GetTopicStats(ctx context.Context, topicID string) (*model.TopicStats, error)
    IncrTopicStat(ctx context.Context, topicID, field string, delta int64) error

    // topic_read_records 表（2级数据，异步写入）
    MarkTopicRead(ctx context.Context, topicID, userID string) error
}
```

#### inbox/repository.go — 对应 `conversations` + `messages` + `message_receipts` 表

```go
type InboxRepository interface {
    GetInboxMessages(ctx context.Context, userID string, page, pageSize int32) ([]*model.Message, int64, error)
    MarkMessageRead(ctx context.Context, userID, messageID string) error
    BatchMarkMessagesRead(ctx context.Context, userID string, messageIDs []string) error
    GetUnreadCount(ctx context.Context, userID string) (int64, error)
    CreatePrivateMessage(ctx context.Context, msg *model.Message) error
    GetPrivateMessages(ctx context.Context, conversationID string, page, pageSize int32) ([]*model.Message, int64, error)
}
```

---

### **Step 3：Permission Service（广场虚拟成员规则落地，所有权限判断的唯一入口）**

在 `go/modules/social/permission/service.go` 实现以下 7 个方法，所有 handler/usecase 的权限判断必须通过此服务，禁止直接查表鉴权：

```go
type PermissionService interface {
    CanViewGroup(ctx context.Context, userID, groupID string) (bool, error)        // 1. 查看群组
    CanJoinGroup(ctx context.Context, userID, groupID string) (bool, error)        // 2. 加入群组
    CanViewTopicDetail(ctx context.Context, userID, topicID string) (bool, error)  // 3. 查看帖子完整内容（只查1级数据 group_members）
    CanViewTopicSummary(ctx context.Context, userID, topicID string) (bool, error) // 4. 查看帖子摘要
    CanCreateTopic(ctx context.Context, userID, groupID string) (bool, error)      // 5. 发帖/回复
    CanManageGroup(ctx context.Context, operatorID, groupID string) (bool, error)  // 6. 管理群组
    CanManageMember(ctx context.Context, operatorID, groupID, targetUserID string) (bool, error) // 7. 管理成员
}
```

**广场虚拟成员前置条件（每个方法内必须注入）**：

```go
// 广场群组的普通成员判断：active 用户 = 虚拟普通成员，不查 group_members
func (s *permissionService) isVirtualPlazaMember(ctx context.Context, userID, groupID string) (bool, error) {
    if groupID != s.cfg.PlazaGroupID {  // PlazaGroupID 从配置中心读取
        return false, nil
    }
    user, err := s.memberRepo.GetUserByID(ctx, userID)
    if err != nil {
        return false, err
    }
    return user.Status == model.UserStatusActive, nil
}
```

**权限铁律**：
- `CanViewTopicDetail` 内部只能查 1级数据 `group_members`，**不能依赖 Redis 缓存**
- `IsBlockedBetween(userA, userB)` 通过 `member_blocks` 独立查询，与群组角色无关
- 广场管理员/嘉宾从 `group_members` 读取，普通成员通过虚拟推导

---

### **Step 4：开发优先级（按依赖关系排序）**

| 优先级 | 功能模块 | 对应协议 | 依赖 SQL 表 |
|---|---|---|---|
| **P0（阻断）** | 确认 group_base 废弃协议替代方案 | 2005/2007/2008/2001/2002 | — |
| **P0** | `make proto` 重新生成代码 | 全部 | — |
| **P0** | Permission Service 实现 | — | `users` + `group_members` |
| **P1** | MemberRegister + MemberLogin | 1021~1026 | `users` |
| **P1** | GetUserInfo + UpdateUserInfo | 1027~1034 | `users` |
| **P1** | BlockUser

*内容由 AI 生成仅供参考*