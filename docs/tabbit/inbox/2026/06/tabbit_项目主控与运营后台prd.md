请输出运营后台各个主要模块，相关列表字段，操作列清单，筛选列清单，高级筛选列清单，以及相关的运营后台需求prd，运营管理员可以查看用户，群组，主题，评论，删除不良主题评论，禁用不良群组和用户，查看订单，管理分账，维护多语言，配置app信息，闪屏，广告，banner维护，模板消息，基于go-admin框架设计，维护全局枚举对于常量数据字典，也整理下，补充一点，全局的所有用户都在一个大群组广场，采用虚拟成员关系，不维护group_members表记录，只有广场管理员嘉宾存储在group_members，虚拟成员，默认所有注册的用户都是该群组的普通成员，成员间可以相互block，

### **运营后台应按“用户、群组、主题、评论、订单、分账、消息、多语言、App 配置、闪屏广告 Banner、模板消息、全局枚举字典”十二个核心模块设计；其中“大群组广场”采用虚拟成员关系：所有注册用户默认是广场普通成员，不写入 `group_members`，只有广场管理员和嘉宾写入 `group_members`，成员间 block 关系独立由 `member_blocks` 维护。**

下面这份可以直接作为 `docs/prd/PRD-admin-backend.md` 的主体内容，也可以拆分到 `docs/api/enums-dictionary.md`、`docs/wiki/CODE-WIKI.md` 和 go-admin 菜单权限配置文档中。

---

## **1. 运营后台产品定位**

CaiRobot MVP 运营后台基于 `go-admin v2.2.0` 框架建设，前端使用 Vue 2 + Element UI，后端复用 Go 管理后台能力，并通过社交域服务或数据库管理能力支持运营人员完成用户、群组、主题、评论、订单、分账、消息、多语言、App 配置、广告资源、模板消息和全局枚举字典维护。

运营后台不是面向普通 App 用户的业务入口，而是面向平台运营、客服、审核、财务和系统管理员的内部管理工具。后台所有敏感操作必须记录操作日志，尤其是删除主题、删除评论、禁用用户、禁用群组、确认支付、退款、分账、修改 App 配置、发布广告、修改全局枚举等操作。

后台权限建议按 RBAC 设计，复用 go-admin 的菜单、角色、权限、按钮级权限和操作日志能力。

### **1.1 与 PRD-10（Admin 管理后台 MVP）的关系**

本 PRD 是在 **[PRD-10](../prd/PRD-10-Admin管理后台MVP.md)** 已完成基础上的**社交域运营扩展**，不是替代或重新设计。两者关系如下：

| 能力模块 | PRD-10 状态 | 本 PRD 处理方式 |
|---|---|---|
| go-admin v2.2.0 框架初始化 | **已实现**（含 Gin 引擎、登录鉴权、RBAC 骨架） | 直接复用，不重新定义 |
| Config Schema CRUD（配置字段定义） | **已实现未验收**（`services/config/admin/` + `plugins/config_admin/`） | 标注"已实现待验收"，仅做范围声明 |
| Config Value CRUD（按环境切换配置值） | **已实现未验收**（含发布 + 缓存失效 + pub/sub 广播） | 同上 |
| i18n String CRUD（多语言文案） | **已实现未验收**（`services/i18n/admin/` + `plugins/i18n_admin/`） | 同上 |
| i18n Pack 管理（语言包版本发布） | **已实现未验收**（含 CSV 导入导出） | 同上 |
| admin-web 前端页面（Vue 2 + Element UI） | **已实现未验收**（4 个配置页 + 4 个 i18n 页） | 扩展新增页面，不修改已有页面 |
| **全局枚举字典管理** | **未实现** | 本 PRD 新增定义（§16 / `docs/api/enums-dictionary.md`） |
| **用户治理 / 群组治理 / 内容审核** | **未实现**（依赖社交域） | 本 P0/P1 新增 |
| **订单 / 分账 / 广告 Banner / 模板消息** | **未实现** | 本 P1/P2 新增 |
| **操作审计日志（admin_operation_logs）** | **未实现**（go-admin 自带 sys_oper_log 仅记录基础 CRUD） | 本 PRD 新增业务审计层 |

**关键约束**：
- 本 PRD 的 App 配置（§12）和多语言维护（§11）**不重新设计**已在 PRD-10 中实现的 Schema/String/Pack 三套 CRUD
- 仅对这两模块做**范围确认和状态标注**：已实现的部分标注"已实现待验收"，待扩展的部分（如多语言的引用查看、App 配置的环境复制等）在本 PRD 中补充
- 全局枚举字典（§16）是全新模块，PRD-10 未覆盖
- 社交域运营管理模块（用户/群组/主题/评论/订单/分账/消息）全部为新增，依赖 `go/modules/social` 服务层

---

## **2. 核心架构约束**

运营后台需要遵循当前 CaiRobot MVP 的整体架构：

| 约束项 | 设计要求 |
|---|---|
| 后台框架 | 基于 `go-admin v2.2.0` |
| 前端技术 | Vue 2 + Element UI |
| 后端技术 | Go + GORM + go-admin 插件/模块 |
| 社交域业务 | 复用 `go/modules/social` 或后续拆分服务 |
| 外部 App 协议 | App 仍走 `POST /api/hello` + `MessagePacket` + Protobuf |
| 后台管理接口 | 可使用 go-admin 后台管理路由，但核心业务规则必须复用 service/usecase |
| 数据一致性 | 1级数据强一致写 MySQL，2级数据由 Redis + 事件驱动更新 |
| 审计日志 | 所有高权限操作必须写入后台操作日志和业务审计表 |
| 数据字典 | 枚举和常量通过全局数据字典维护，代码中只保留稳定常量 key |

---

## **3. 大群组广场虚拟成员关系规则**

这是运营后台和社交 App 都必须遵守的关键产品规则。

### **3.1 规则定义**

平台存在一个全局默认大群组，称为“广场群组”或“大群组广场”。所有注册用户默认都是该广场群组的普通成员，但这种普通成员关系是虚拟关系，不写入 `group_members` 表。

也就是说：

```text
所有注册用户 = 广场群组虚拟普通成员
```

后台和业务服务判断用户是否属于广场群组时，不应该去 `group_members` 查普通成员记录，而应该通过以下逻辑判断：

```text
如果 group.type = plaza 或 group.id = system_plaza_group_id：
    只要 users.status = active，即视为普通成员
```

### **3.2 需要写入 `group_members` 的广场成员**

只有以下特殊身份需要写入 `group_members`：

| 身份 | 是否写入 `group_members` | 说明 |
|---|---:|---|
| 广场普通成员 | 否 | 所有注册用户默认拥有虚拟普通成员身份 |
| 广场嘉宾 | 是 | 需要明确邀请、授权、过期时间和权限 |
| 广场管理员 | 是 | 需要后台可审计、可撤销、可授权 |
| 广场所有人 | 是 | 系统初始化或平台超级管理员配置 |
| 被禁用用户 | 不通过普通成员记录表达 | 使用 `users.status` 或风控表控制 |

### **3.3 虚拟成员与 block 关系**

成员之间可以相互 block。block 关系与是否真实写入 `group_members` 无关，应独立存储在 `member_blocks` 表。

```text
member_blocks 控制用户间可见性、私信、评论、互动；
group_members 控制群组内特殊角色和管理权限；
广场普通成员身份由 users + plaza group 规则虚拟推导。
```

### **3.4 权限判断规则**

针对广场群组，权限服务需要特殊处理：

```go
CanViewGroup(userID, plazaGroupID)
CanReadTopic(userID, topicID)
CanCommentTopic(userID, topicID)
CanManageGroup(operatorID, plazaGroupID)
CanManageMember(operatorID, plazaGroupID, targetUserID)
IsBlockedBetween(userA, userB)
```

判断逻辑建议：

```text
1. 如果用户未登录，只能浏览公开内容摘要；
2. 如果用户已注册且 users.status=active，默认是广场普通成员；
3. 如果用户是广场管理员/嘉宾/所有人，从 group_members 获取角色；
4. 如果用户被作者 block，不能查看作者的受限主题完整内容；
5. 如果用户被平台禁用，不能发帖、评论、私信；
6. 管理员治理行为不受普通 block 限制，但必须审计。
```

---

# **运营后台模块清单**

## **4. 用户管理**

### **4.1 模块目标**

用户管理用于查看、检索、编辑和治理平台注册用户。运营管理员可以查看用户基础资料、账号状态、登录信息、群组关系、订单记录、消息记录、block 状态，并对违规用户执行禁用、启用、重置密码、发送站内信等操作。

### **4.2 关联数据表**

| 表 | 用途 |
|---|---|
| `users` | 用户身份主表 |
| `member_sessions` | 登录会话 |
| `member_blocks` | 用户 block 关系 |
| `member_stats` | 用户统计快照 |
| `group_members` | 特殊群组角色，广场仅存管理员/嘉宾 |
| `group_orders` | 用户订单 |
| `messages` | 用户消息 |

### **4.3 列表字段**

| 字段 | 说明 |
|---|---|
| 用户 ID | `users.id` |
| 展示 ID | `users.user_id` 或 `uid` |
| 用户名 | `username` |
| 昵称 | `nickname` |
| 头像 | `avatar` |
| 手机号 | `phone` |
| 邮箱 | `email` |
| 状态 | active / inactive / banned / deleted |
| 会员等级 | 平台会员等级，不等于群组权益 |
| 注册时间 | `created_at` |
| 最后登录时间 | `last_login_at` |
| 最后登录 IP | `last_login_ip` |
| 登录次数 | `login_count` |
| 加入群组数 | 来自 `member_stats.joined_groups_count` |
| 发布主题数 | 来自 `member_stats.topics_count` |
| 未读消息数 | 来自 Redis 或 `member_stats.unread_messages_count` |

### **4.4 操作列清单**

| 操作 | 说明 | 权限建议 |
|---|---|---|
| 查看详情 | 查看用户完整资料、订单、主题、评论、消息、block 关系 | 用户查看权限 |
| 编辑资料 | 修改昵称、头像、手机号、邮箱等非敏感字段 | 用户编辑权限 |
| 禁用用户 | 将用户状态改为 banned | 用户治理权限 |
| 启用用户 | 恢复 active | 用户治理权限 |
| 重置密码 | 管理员重置密码或生成重置链接 | 安全管理权限 |
| 查看群组 | 查看用户加入/管理/嘉宾群组 | 用户查看权限 |
| 查看订单 | 跳转订单管理并按用户筛选 | 订单查看权限 |
| 查看主题 | 跳转主题管理并按作者筛选 | 内容查看权限 |
| 查看评论 | 跳转评论管理并按作者筛选 | 内容查看权限 |
| 查看 block 关系 | 查看拉黑/被拉黑列表 | 风控查看权限 |
| 发送站内信 | 给用户发送系统消息 | 消息管理权限 |
| 清退会话 | 让用户 token/session 失效 | 安全管理权限 |

### **4.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 用户 ID / 展示 ID | 输入框 |
| 用户名 | 输入框 |
| 昵称 | 输入框 |
| 手机号 | 输入框 |
| 邮箱 | 输入框 |
| 用户状态 | 下拉枚举 |
| 会员等级 | 下拉枚举 |
| 注册时间范围 | 日期范围 |
| 最后登录时间范围 | 日期范围 |

### **4.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 登录 IP | 输入框 |
| 登录次数范围 | 数字区间 |
| 是否邮箱验证 | 单选 |
| 是否手机验证 | 单选 |
| 是否 IM 注册 | 单选 |
| 是否存在 block 记录 | 单选 |
| 是否发布过主题 | 单选 |
| 是否有订单 | 单选 |
| 加入群组数范围 | 数字区间 |
| 发布主题数范围 | 数字区间 |
| 未读消息数范围 | 数字区间 |
| 是否广场管理员/嘉宾 | 下拉枚举 |

---

## **5. 群组管理**

### **5.1 模块目标**

群组管理用于运营人员查看和治理平台群组，包括广场群组、免费群组、付费群组、邀请制群组。后台可以审核群组、禁用不良群组、设置推荐、管理群组所有人、查看成员、查看付费方案和订单。

### **5.2 关联数据表**

| 表 | 用途 |
|---|---|
| `groups` | 群组主表 |
| `group_members` | 管理员/嘉宾/普通群成员关系 |
| `group_stats` | 群组统计 |
| `group_plans` | 付费方案 |
| `group_plan_periods` | 付费周期 |
| `group_discounts` | 折扣周期 |
| `group_invites` | 嘉宾邀请 |
| `group_admin_actions` | 群组管理审计 |
| `group_orders` | 群组订单 |

### **5.3 列表字段**

| 字段 | 说明 |
|---|---|
| 群组 ID | `groups.id` |
| 群组名称 | `name` |
| slug | `slug` |
| 群组头像 | `avatar` |
| 群组类型 | free / paid / mixed / invite / plaza |
| 可见性 | public / link / private |
| 加入方式 | free / apply / paid / invite / virtual |
| 状态 | active / auditing / banned / deleted |
| 是否官方 | `is_official` |
| 是否推荐 | `is_featured` |
| 圈主 | `owner_id` |
| 成员数 | `group_stats.members_count` 或 `groups.members_count` |
| 主题数 | `group_stats.topics_count` |
| 付费成员数 | `group_stats.paid_members_count` |
| 嘉宾数 | `group_stats.guest_members_count` |
| 订单数 | `group_stats.orders_count` |
| 创建时间 | `created_at` |
| 更新时间 | `updated_at` |

### **5.4 操作列清单**

| 操作 | 说明 | 权限建议 |
|---|---|---|
| 查看详情 | 查看群组资料、统计、成员、主题、订单 | 群组查看权限 |
| 编辑群组 | 修改名称、简介、封面、规则、联系方式等 | 群组编辑权限 |
| 审核通过 | 将审核中群组设为 active | 群组审核权限 |
| 审核驳回 | 驳回群组创建或修改 | 群组审核权限 |
| 禁用群组 | 禁用不良群组 | 群组治理权限 |
| 启用群组 | 恢复群组 | 群组治理权限 |
| 设置推荐 | 设置 `is_featured` | 运营推荐权限 |
| 取消推荐 | 取消推荐 | 运营推荐权限 |
| 设置官方 | 设置 `is_official` | 超级管理员 |
| 查看成员 | 跳转群组成员页 | 群组查看权限 |
| 查看嘉宾 | 查看 `role=guest` 成员 | 群组查看权限 |
| 查看管理员 | 查看 `role=admin/owner` 成员 | 群组查看权限 |
| 管理付费方案 | 跳转付费方案管理 | 财务/群组配置权限 |
| 查看订单 | 跳转订单列表并按群组筛选 | 订单查看权限 |
| 转让群主 | 修改 `owner_id` 并更新角色 | 高危权限 |
| 查看审计日志 | 查看 `group_admin_actions` | 审计权限 |

### **5.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 群组 ID | 输入框 |
| 群组名称 | 输入框 |
| slug | 输入框 |
| 群组类型 | 下拉枚举 |
| 可见性 | 下拉枚举 |
| 加入方式 | 下拉枚举 |
| 状态 | 下拉枚举 |
| 是否官方 | 单选 |
| 是否推荐 | 单选 |
| 圈主 ID / 昵称 | 输入框 |
| 创建时间范围 | 日期范围 |

### **5.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 成员数范围 | 数字区间 |
| 主题数范围 | 数字区间 |
| 付费成员数范围 | 数字区间 |
| 嘉宾数范围 | 数字区间 |
| 订单数范围 | 数字区间 |
| 订单金额范围 | 金额区间 |
| 是否配置付费方案 | 单选 |
| 是否配置折扣 | 单选 |
| 是否存在举报 | 单选 |
| 最后活跃时间范围 | 日期范围 |
| 是否广场群组 | 单选 |
| 是否虚拟成员群组 | 单选 |

---

## **6. 群组成员与嘉宾管理**

### **6.1 模块目标**

该模块用于管理群组中的真实成员记录，尤其是普通私域群组成员、付费群组成员、嘉宾、管理员、所有人。对于大群组广场，普通成员不展示为真实 `group_members` 记录，仅展示管理员和嘉宾记录，并在页面上明确提示“广场普通成员为虚拟成员”。

### **6.2 关联数据表**

| 表 | 用途 |
|---|---|
| `group_members` | 群组真实成员关系 |
| `group_member_entitlements` | 成员权益 |
| `group_invites` | 嘉宾邀请 |
| `member_blocks` | 成员 block |
| `users` | 用户资料 |
| `group_admin_actions` | 管理审计 |

### **6.3 列表字段**

| 字段 | 说明 |
|---|---|
| 成员记录 ID | `group_members.id` |
| 群组 ID | `group_id` |
| 用户 ID | `user_id` |
| 用户昵称 | 来自 `users.nickname` |
| 角色 | owner / admin / guest / member |
| 状态 | active / pending / muted / banned / left / expired |
| 加入来源 | free / paid / invite / admin / import |
| 邀请人 | `invited_by` |
| 加入时间 | `joined_at` |
| 禁言截止时间 | `muted_until` |
| 权益类型 | free / paid / guest / gift / admin_grant |
| 权益过期时间 | 来自 `group_member_entitlements` |
| 创建时间 | `created_at` |

### **6.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看成员详情 | 查看用户、权益、订单、邀请来源 |
| 设置管理员 | 将成员设置为 admin |
| 取消管理员 | 将 admin 恢复为 member |
| 设置嘉宾 | 将成员设置为 guest |
| 移除嘉宾 | 移除 guest 权限 |
| 禁言 | 设置 `muted_until` |
| 解除禁言 | 清空禁言时间 |
| 移除成员 | 设置状态 removed/left |
| 恢复成员 | 恢复 active |
| 查看权益 | 查看权益记录 |
| 手动开通权益 | 管理员手动 grant |
| 查看邀请 | 查看邀请来源 |
| 查看 block 关系 | 查看该成员 block 状态 |

### **6.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 群组 ID | 输入框 |
| 用户 ID / 昵称 | 输入框 |
| 角色 | 下拉 |
| 状态 | 下拉 |
| 加入来源 | 下拉 |
| 加入时间范围 | 日期范围 |
| 权益状态 | 下拉 |

### **6.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 权益过期时间范围 | 日期范围 |
| 是否即将过期 | 单选 |
| 是否被禁言 | 单选 |
| 是否被邀请 | 单选 |
| 邀请人 ID | 输入框 |
| 是否存在 block 关系 | 单选 |
| 是否广场特殊成员 | 单选 |
| 是否虚拟普通成员 | 单选，只用于广场说明，不查 `group_members` |

---

## **7. 主题管理**

### **7.1 模块目标**

主题管理用于查看、审核、编辑、下架和删除平台主题。运营管理员可以处理不良主题，设置置顶、精选，查看阅读、评论、点赞、分享等数据。

### **7.2 关联数据表**

| 表 | 用途 |
|---|---|
| `topics` | 主题主表 |
| `topic_stats` | 主题统计 |
| `topic_comments` | 评论 |
| `topic_reactions` | 点赞/收藏 |
| `topic_shares` | 分享 |
| `topic_read_records` | 阅读记录 |
| `users` | 作者 |
| `groups` | 所属群组 |

### **7.3 列表字段**

| 字段 | 说明 |
|---|---|
| 主题 ID | `topics.id` |
| 标题 | `title` |
| 内容摘要 | `summary` |
| 作者 | `author_id` + 昵称 |
| 所属群组 | `group_id` + 群组名称 |
| 主题类型 | 普通/公告/问答/付费内容 |
| 内容类型 | 文本/图文/视频/文档 |
| 状态 | draft / pending / published / hidden / deleted |
| 可见性 | public / group_member / paid_member / owner_only |
| 是否匿名 | `is_anonymous` |
| 是否锁定 | `is_locked` |
| 是否置顶 | `is_pinned` |
| 是否精选 | `is_featured` |
| 是否允许评论 | `allow_comments` |
| 阅读数 | `topic_stats.read_count` |
| 评论数 | `topic_stats.comments_count` |
| 点赞数 | `topic_stats.likes_count` |
| 分享数 | `topic_stats.shares_count` |
| 发布时间 | `published_at` |
| 创建时间 | `created_at` |

### **7.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看主题完整内容和媒体附件 |
| 编辑主题 | 修改标题、摘要、标签、可见性 |
| 审核通过 | 发布待审核主题 |
| 审核驳回 | 驳回主题 |
| 下架主题 | 隐藏不良主题 |
| 恢复主题 | 恢复下架主题 |
| 删除主题 | 软删除违规主题 |
| 置顶/取消置顶 | 更新 `is_pinned` |
| 精选/取消精选 | 更新 `is_featured` |
| 锁定/解锁 | 控制评论或互动 |
| 查看作者 | 跳转用户详情 |
| 查看群组 | 跳转群组详情 |
| 查看评论 | 跳转评论列表 |
| 查看分享 | 查看 `topic_shares` |
| 查看阅读记录 | 查看阅读数据 |
| 发送整改通知 | 给作者发送站内信 |

### **7.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 主题 ID | 输入框 |
| 标题关键词 | 输入框 |
| 作者 ID / 昵称 | 输入框 |
| 群组 ID / 名称 | 输入框 |
| 主题类型 | 下拉 |
| 内容类型 | 下拉 |
| 状态 | 下拉 |
| 可见性 | 下拉 |
| 发布时间范围 | 日期范围 |
| 创建时间范围 | 日期范围 |

### **7.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 是否匿名 | 单选 |
| 是否锁定 | 单选 |
| 是否置顶 | 单选 |
| 是否精选 | 单选 |
| 是否允许评论 | 单选 |
| 是否含媒体 | 单选 |
| 是否含文档 | 单选 |
| 阅读数范围 | 数字区间 |
| 评论数范围 | 数字区间 |
| 点赞数范围 | 数字区间 |
| 分享数范围 | 数字区间 |
| 是否被举报 | 单选 |
| 是否命中敏感词 | 单选 |
| 最后活跃时间范围 | 日期范围 |
| 是否来自广场群组 | 单选 |

---

## **8. 评论管理**

### **8.1 模块目标**

评论管理用于审核、隐藏、删除不良评论，处理违规互动。运营人员可以按主题、作者、群组、状态、举报情况检索评论。

### **8.2 关联数据表**

| 表 | 用途 |
|---|---|
| `topic_comments` | 评论主表 |
| `topics` | 所属主题 |
| `groups` | 所属群组 |
| `users` | 评论作者 |
| `messages` | 评论通知消息 |

### **8.3 列表字段**

| 字段 | 说明 |
|---|---|
| 评论 ID | `id` |
| 内容摘要 | 评论内容截断展示 |
| 作者 | `user_id` |
| 主题 | `topic_id` |
| 群组 | `group_id` |
| 父评论 | `parent_id` |
| 回复用户 | `reply_to_user_id` |
| 状态 | normal / pending / hidden / deleted |
| 创建时间 | `created_at` |
| 更新时间 | `updated_at` |
| 举报次数 | 风控或举报统计 |
| 是否命中敏感词 | 风控结果 |

### **8.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看完整评论内容 |
| 编辑评论 | 管理员修正或脱敏 |
| 隐藏评论 | 隐藏不良评论 |
| 恢复评论 | 恢复评论 |
| 删除评论 | 删除违规评论 |
| 查看作者 | 跳转用户详情 |
| 查看主题 | 跳转主题详情 |
| 查看群组 | 跳转群组详情 |
| 禁言作者 | 对作者执行群组禁言 |
| 禁用作者 | 对严重违规用户禁用 |
| 批量删除 | 批量处理违规评论 |
| 发送通知 | 通知评论作者 |

### **8.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 评论 ID | 输入框 |
| 内容关键词 | 输入框 |
| 作者 ID / 昵称 | 输入框 |
| 主题 ID / 标题 | 输入框 |
| 群组 ID / 名称 | 输入框 |
| 状态 | 下拉 |
| 创建时间范围 | 日期范围 |

### **8.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 是否回复评论 | 单选 |
| 是否被举报 | 单选 |
| 举报次数范围 | 数字区间 |
| 是否命中敏感词 | 单选 |
| 是否来自广场群组 | 单选 |
| 作者状态 | 下拉 |
| 主题状态 | 下拉 |
| 评论长度范围 | 数字区间 |

---

## **9. 订单管理**

### **9.1 模块目标**

订单管理用于查看用户付费入群、续费、赠送、管理员开通等订单记录。运营或财务人员可以确认支付、取消订单、退款、查看权益和订单事件。

### **9.2 关联数据表**

| 表 | 用途 |
|---|---|
| `group_orders` | 群组订单 |
| `group_order_events` | 订单事件 |
| `group_plans` | 付费方案 |
| `group_plan_periods` | 付费周期 |
| `group_discounts` | 折扣 |
| `group_member_entitlements` | 权益 |
| `users` | 下单用户 |
| `groups` | 所属群组 |

### **9.3 列表字段**

| 字段 | 说明 |
|---|---|
| 订单 ID | `id` |
| 订单号 | `order_no` |
| 用户 | `user_id` |
| 群组 | `group_id` |
| 方案 | `plan_id` |
| 周期 | `period_id` |
| 折扣 | `discount_id` |
| 订单类型 | join / renew / gift / admin_grant |
| 原价 | `original_amount_cent` |
| 折扣金额 | `discount_amount_cent` |
| 实付金额 | `pay_amount_cent` |
| 币种 | `currency` |
| 状态 | pending / paid / cancelled / refunded / failed / expired |
| 支付渠道 | `pay_channel` |
| 支付时间 | `paid_at` |
| 权益开始时间 | `entitlement_started_at` |
| 权益过期时间 | `entitlement_expired_at` |
| 创建时间 | `created_at` |

### **9.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看订单、用户、群组、权益、事件 |
| 确认支付 | 手动确认支付成功 |
| 取消订单 | 取消 pending 订单 |
| 退款 | 标记退款并处理权益 |
| 查看权益 | 查看该订单生成的权益 |
| 查看订单事件 | 查看状态流转 |
| 查看用户 | 跳转用户详情 |
| 查看群组 | 跳转群组详情 |
| 导出订单 | 导出筛选结果 |
| 发送通知 | 给用户发送订单通知 |

### **9.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 订单号 | 输入框 |
| 用户 ID / 昵称 | 输入框 |
| 群组 ID / 名称 | 输入框 |
| 订单类型 | 下拉 |
| 订单状态 | 下拉 |
| 支付渠道 | 下拉 |
| 创建时间范围 | 日期范围 |
| 支付时间范围 | 日期范围 |

### **9.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 实付金额范围 | 金额区间 |
| 折扣金额范围 | 金额区间 |
| 是否使用折扣 | 单选 |
| 方案 ID | 输入框 |
| 周期 ID | 输入框 |
| 权益过期时间范围 | 日期范围 |
| 是否即将过期 | 单选 |
| 是否续费订单 | 单选 |
| 是否管理员开通 | 单选 |
| 是否已生成分账 | 单选 |

---

## **10. 分账管理**

### **10.1 模块目标**

分账管理用于统计群组付费收入、平台分成、群主收益、退款扣减和结算状态。该模块服务财务运营，首期可以先做手动结算和报表导出。

### **10.2 建议新增数据表**

```text
group_settlements
group_settlement_items
```

### **10.3 列表字段**

| 字段 | 说明 |
|---|---|
| 分账单 ID | settlement_id |
| 分账单号 | settlement_no |
| 群组 | group_id |
| 群主 | owner_id |
| 结算周期 | start_at / end_at |
| 订单数 | orders_count |
| 退款数 | refunds_count |
| 总收入 | gross_amount_cent |
| 退款金额 | refund_amount_cent |
| 平台分成比例 | platform_rate |
| 平台分成金额 | platform_amount_cent |
| 群主收益 | owner_amount_cent |
| 状态 | pending / confirmed / paid / rejected |
| 确认时间 | confirmed_at |
| 打款时间 | paid_at |
| 创建时间 | created_at |

### **10.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看分账单详情 |
| 查看明细 | 查看订单维度明细 |
| 生成分账 | 按周期生成 |
| 确认分账 | 财务确认 |
| 驳回分账 | 标记异常 |
| 标记打款 | 记录打款完成 |
| 导出报表 | 导出 Excel/CSV |
| 查看群组 | 跳转群组 |
| 查看群主 | 跳转用户 |
| 查看相关订单 | 跳转订单列表 |

### **10.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 分账单号 | 输入框 |
| 群组 ID / 名称 | 输入框 |
| 群主 ID / 昵称 | 输入框 |
| 状态 | 下拉 |
| 结算周期 | 日期范围 |
| 创建时间范围 | 日期范围 |

### **10.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 总收入范围 | 金额区间 |
| 群主收益范围 | 金额区间 |
| 平台分成比例范围 | 数字区间 |
| 订单数范围 | 数字区间 |
| 退款金额范围 | 金额区间 |
| 是否已打款 | 单选 |
| 打款时间范围 | 日期范围 |
| 是否存在异常订单 | 单选 |

---

## **11. 多语言维护**

### **11.1 模块目标**

多语言维护用于管理 App 文案、后台文案、模板消息文案和错误提示文案。建议与现有 I18n 服务保持一致，后台维护后发布到配置中心或 i18n 服务。

### **11.2 关联数据表**

| 表 | 用途 |
|---|---|
| `i18n_strings` | 多语言键值 |
| `i18n_packages` | 语言包 |
| `app_configs` | 配置发布 |
| `message_templates` | 模板消息文案 |

### **11.3 列表字段**

| 字段 | 说明 |
|---|---|
| 文案 ID | id |
| Key | i18n key |
| 默认语言内容 | default_value |
| 语言 | locale |
| 翻译内容 | value |
| 所属模块 | module |
| 使用端 | app / admin / message / error |
| 状态 | draft / active / disabled |
| 版本 | version |
| 更新时间 | updated_at |
| 更新人 | updated_by |

### **11.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看所有语言版本 |
| 新增文案 | 新增 key |
| 编辑翻译 | 修改某语言内容 |
| 批量导入 | 导入翻译文件 |
| 批量导出 | 导出翻译文件 |
| 发布 | 发布到运行环境 |
| 回滚 | 回滚上一版本 |
| 查看引用 | 查看哪些模板/页面引用 |
| 禁用 | 禁用文案 key |

### **11.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| Key | 输入框 |
| 文案内容 | 输入框 |
| 语言 | 下拉 |
| 所属模块 | 下拉 |
| 使用端 | 下拉 |
| 状态 | 下拉 |
| 更新时间范围 | 日期范围 |

### **11.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 是否缺失翻译 | 单选 |
| 是否存在空值 | 单选 |
| 是否已发布 | 单选 |
| 版本号 | 输入框 |
| 更新人 | 输入框 |
| 使用次数范围 | 数字区间 |
| 最后使用时间范围 | 日期范围 |

---

## **12. App 配置管理**

### **12.1 模块目标**

App 配置管理用于维护客户端运行参数、功能开关、版本配置、审核开关、支付开关、广场群组配置等。

### **12.2 关联数据表**

| 表 | 用途 |
|---|---|
| `app_configs` | App 配置 |
| `sys_dict_type` / `sys_dict_data` | 配置枚举 |
| `config_history` | 配置历史 |

### **12.3 列表字段**

| 字段 | 说明 |
|---|---|
| 配置 ID | id |
| 配置 Key | config_key |
| 配置名称 | config_name |
| 配置值 | config_value |
| 配置类型 | string / number / bool / json |
| 环境 | dev / test / staging / prod |
| 所属模块 | app / payment / plaza / message / audit |
| 是否敏感 | is_sensitive |
| 状态 | draft / active / disabled |
| 版本 | version |
| 更新时间 | updated_at |
| 更新人 | updated_by |

### **12.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看配置详情 |
| 编辑 | 修改配置 |
| 发布 | 发布到指定环境 |
| 回滚 | 回滚历史版本 |
| 禁用 | 禁用配置 |
| 复制配置 | 复制到其他环境 |
| 查看历史 | 查看变更历史 |
| 校验 JSON | JSON 类型配置校验 |

### **12.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 配置 Key | 输入框 |
| 配置名称 | 输入框 |
| 配置类型 | 下拉 |
| 环境 | 下拉 |
| 所属模块 | 下拉 |
| 状态 | 下拉 |
| 更新时间范围 | 日期范围 |

### **12.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 是否敏感配置 | 单选 |
| 是否系统内置 | 单选 |
| 是否已发布 | 单选 |
| 版本号 | 输入框 |
| 更新人 | 输入框 |
| 是否 JSON 合法 | 单选 |
| 是否影响客户端启动 | 单选 |

---

## **13. 闪屏、广告、Banner 维护**

### **13.1 模块目标**

用于维护 App 闪屏广告、首页 Banner、群组推荐 Banner、活动广告、弹窗广告等资源。运营人员可以配置投放时间、位置、跳转链接、上下架状态和展示统计。

### **13.2 关联数据表**

| 表 | 用途 |
|---|---|
| `app_banners` | Banner 资源 |
| `app_ads` | 广告资源 |
| `app_splash_screens` | 闪屏资源 |
| `media_assets` | 图片/视频资源 |
| `ad_stats` | 展示点击统计 |

### **13.3 列表字段**

| 字段 | 说明 |
|---|---|
| 资源 ID | id |
| 标题 | title |
| 图片/视频 | media_url |
| 类型 | splash / banner / popup / feed |
| 投放位置 | app_home / plaza / group / topic |
| 跳转类型 | none / url / group / topic / app_page |
| 跳转目标 | target |
| 状态 | draft / active / offline / expired |
| 排序 | sort |
| 开始时间 | start_at |
| 结束时间 | end_at |
| 展示次数 | impression_count |
| 点击次数 | click_count |
| 点击率 | ctr |
| 创建时间 | created_at |

### **13.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看资源详情 |
| 新增 | 新增广告/Banner |
| 编辑 | 修改配置 |
| 上架 | 状态改为 active |
| 下架 | 状态改为 offline |
| 预览 | 预览素材和跳转 |
| 复制 | 复制一条资源 |
| 查看数据 | 查看展示、点击 |
| 删除 | 删除草稿或过期资源 |

### **13.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 标题 | 输入框 |
| 类型 | 下拉 |
| 投放位置 | 下拉 |
| 跳转类型 | 下拉 |
| 状态 | 下拉 |
| 投放时间范围 | 日期范围 |

### **13.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 展示次数范围 | 数字区间 |
| 点击次数范围 | 数字区间 |
| 点击率范围 | 数字区间 |
| 是否当前生效 | 单选 |
| 是否已过期 | 单选 |
| 素材类型 | 图片/视频 |
| 创建人 | 输入框 |
| 排序范围 | 数字区间 |

---

## **14. 模板消息管理**

### **14.1 模块目标**

模板消息用于维护系统通知、订单通知、群组通知、评论点赞通知、私信提醒等消息模板。消息中心发送消息时应引用模板 key 和变量。

### **14.2 关联数据表**

| 表 | 用途 |
|---|---|
| `message_templates` | 消息模板 |
| `messages` | 消息记录 |
| `message_delivery_logs` | 投递日志 |
| `i18n_strings` | 多语言模板文案 |

### **14.3 列表字段**

| 字段 | 说明 |
|---|---|
| 模板 ID | id |
| 模板 Key | template_key |
| 模板名称 | template_name |
| 模板类型 | private / group / order / interaction / system |
| 标题模板 | title_template |
| 内容模板 | content_template |
| 变量列表 | variables |
| 语言 | locale |
| 状态 | draft / active / disabled |
| 使用次数 | use_count |
| 更新时间 | updated_at |
| 更新人 | updated_by |

### **14.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看模板和变量 |
| 新增模板 | 新建模板 |
| 编辑模板 | 修改标题/内容/变量 |
| 测试发送 | 输入变量测试发送 |
| 发布 | 模板生效 |
| 停用 | 停用模板 |
| 复制 | 复制模板 |
| 查看发送记录 | 跳转消息日志 |
| 查看多语言 | 查看关联 i18n |

### **14.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 模板 Key | 输入框 |
| 模板名称 | 输入框 |
| 模板类型 | 下拉 |
| 语言 | 下拉 |
| 状态 | 下拉 |
| 更新时间范围 | 日期范围 |

### **14.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 使用次数范围 | 数字区间 |
| 是否缺少变量 | 单选 |
| 是否已发布 | 单选 |
| 是否多语言模板 | 单选 |
| 更新人 | 输入框 |
| 最后使用时间范围 | 日期范围 |

---

## **15. 消息中心管理**

### **15.1 模块目标**

消息中心管理用于查看系统产生的消息、私信、订单通知、群组状态通知、点赞评论通知等。运营人员通常只具备查看和处理异常消息的权限，不应随意查看用户私信全文，除非具备安全审计权限。

### **15.2 关联数据表**

| 表 | 用途 |
|---|---|
| `conversations` | 会话 |
| `messages` | 消息 |
| `message_receipts` | 已读未读 |
| `message_delivery_logs` | 投递日志 |
| `message_templates` | 模板 |

### **15.3 列表字段**

| 字段 | 说明 |
|---|---|
| 消息 ID | id |
| 会话 ID | conversation_id |
| 消息类型 | private_text / group_status / order_status / like / comment / system |
| 发送人 | sender_id |
| 接收人 | receiver_id |
| 群组 | group_id |
| 主题 | topic_id |
| 标题 | title |
| 内容摘要 | content 摘要 |
| 状态 | normal / recalled / deleted |
| 是否已读 | receipt status |
| 创建时间 | created_at |

### **15.4 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看消息详情 |
| 查看投递日志 | 查看发送失败原因 |
| 重试投递 | 对失败消息重试 |
| 删除消息 | 删除违规系统消息 |
| 撤回消息 | 撤回误发消息 |
| 查看发送人 | 跳转用户 |
| 查看接收人 | 跳转用户 |
| 查看关联订单 | 跳转订单 |
| 查看关联主题 | 跳转主题 |

### **15.5 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 消息 ID | 输入框 |
| 发送人 | 输入框 |
| 接收人 | 输入框 |
| 消息类型 | 下拉 |
| 状态 | 下拉 |
| 创建时间范围 | 日期范围 |

### **15.6 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 是否已读 | 单选 |
| 是否投递失败 | 单选 |
| 投递渠道 | 下拉 |
| 群组 ID | 输入框 |
| 主题 ID | 输入框 |
| 订单 ID | 输入框 |
| 是否模板消息 | 单选 |
| 模板 Key | 输入框 |

---

## **16. 全局枚举字典管理**

### **16.1 模块目标**

全局枚举字典用于维护系统中所有稳定枚举、状态值、类型值、下拉选项、展示文案和排序。后台页面的状态、类型、来源、角色等字段应优先从数据字典读取，避免硬编码散落在前后端。

建议复用 go-admin 常见字典结构：

```text
sys_dict_type
sys_dict_data
```

### **16.2 字典类型表：`sys_dict_type`**

| 字段 | 说明 |
|---|---|
| dict_id | 字典类型 ID |
| dict_name | 字典名称 |
| dict_type | 字典类型 key |
| status | 状态 |
| remark | 备注 |
| created_at | 创建时间 |
| updated_at | 更新时间 |

### **16.3 字典数据表：`sys_dict_data`**

| 字段 | 说明 |
|---|---|
| dict_code | 字典数据 ID |
| dict_sort | 排序 |
| dict_label | 展示文案 |
| dict_value | 实际值 |
| dict_type | 所属字典类型 |
| css_class | 样式 class |
| list_class | 列表标签样式 |
| is_default | 是否默认 |
| status | 状态 |
| remark | 备注 |
| created_at | 创建时间 |
| updated_at | 更新时间 |

### **16.4 列表字段**

| 字段 | 说明 |
|---|---|
| 字典类型 | dict_type |
| 字典名称 | dict_name |
| 字典标签 | dict_label |
| 字典值 | dict_value |
| 排序 | dict_sort |
| 是否默认 | is_default |
| 状态 | active / disabled |
| 使用范围 | app / admin / social / payment / message |
| 是否系统内置 | system_builtin |
| 更新时间 | updated_at |

### **16.5 操作列清单**

| 操作 | 说明 |
|---|---|
| 查看详情 | 查看字典类型和字典项 |
| 新增类型 | 新增 dict_type |
| 编辑类型 | 修改名称或备注 |
| 新增字典项 | 新增 dict_data |
| 编辑字典项 | 修改 label/value/sort |
| 启用/停用 | 控制字典项是否可用 |
| 删除 | 仅允许删除非系统内置项 |
| 导入 | 批量导入字典 |
| 导出 | 导出字典 |
| 查看引用 | 查看哪些模块使用该字典 |
| 刷新缓存 | 刷新字典缓存 |

### **16.6 筛选列清单**

| 筛选项 | 类型 |
|---|---|
| 字典类型 | 输入框/下拉 |
| 字典名称 | 输入框 |
| 字典标签 | 输入框 |
| 字典值 | 输入框 |
| 状态 | 下拉 |
| 是否默认 | 单选 |

### **16.7 高级筛选列清单**

| 高级筛选项 | 类型 |
|---|---|
| 使用范围 | 下拉 |
| 是否系统内置 | 单选 |
| 是否允许删除 | 单选 |
| 是否前端可见 | 单选 |
| 是否 App 可见 | 单选 |
| 排序范围 | 数字区间 |
| 更新时间范围 | 日期范围 |
| 更新人 | 输入框 |

---

# **全局枚举字典清单**

## **17. 必备字典类型**

| dict_type | 字典名称 | 示例值 |
|---|---|---|
| `user_status` | 用户状态 | active / inactive / banned / deleted |
| `user_membership_level` | 用户平台会员等级 | normal / vip / svip |
| `group_type` | 群组类型 | plaza / free / paid / mixed / invite |
| `group_visibility` | 群组可见性 | public / link / private |
| `group_join_mode` | 群组加入方式 | virtual / free / apply / paid / invite |
| `group_status` | 群组状态 | active / auditing / rejected / banned / deleted |
| `group_member_role` | 群组成员角色 | owner / admin / guest / member |
| `group_member_status` | 群组成员状态 | active / pending / muted / banned / left / expired |
| `group_join_source` | 入群来源 | virtual / free / paid / invite / admin / import |
| `entitlement_type` | 权益类型 | free / paid / guest / gift / admin_grant |
| `entitlement_status` | 权益状态 | active / expired / revoked |
| `topic_type` | 主题类型 | normal / announcement / qa / paid |
| `topic_status` | 主题状态 | draft / pending / published / hidden / deleted |
| `topic_visibility` | 主题可见性 | public / group_member / paid_member / owner_only |
| `topic_content_type` | 内容类型 | text / image_text / video / doc / link |
| `comment_status` | 评论状态 | normal / pending / hidden / deleted |
| `reaction_type` | 互动类型 | like / favorite / share |
| `order_type` | 订单类型 | join / renew / gift / admin_grant |
| `order_status` | 订单状态 | pending / paid / cancelled / refunded / failed / expired |
| `pay_channel` | 支付渠道 | manual / wechat / alipay / apple / google |
| `settlement_status` | 分账状态 | pending / confirmed / paid / rejected |
| `message_type` | 消息类型 | private_text / group_status / order_status / like / comment / system |
| `message_status` | 消息状态 | normal / recalled / deleted |
| `message_receipt_status` | 消息回执状态 | unread / read / deleted |
| `app_config_type` | 配置类型 | string / number / bool / json |
| `app_config_env` | 配置环境 | dev / test / staging / prod |
| `ad_type` | 广告类型 | splash / banner / popup / feed |
| `ad_position` | 广告位置 | app_home / plaza / group / topic |
| `ad_status` | 广告状态 | draft / active / offline / expired |
| `template_type` | 模板类型 | private / group / order / interaction / system |
| `template_status` | 模板状态 | draft / active / disabled |
| `i18n_status` | 多语言状态 | draft / active / disabled |
| `audit_action_type` | 审计动作类型 | ban_user / ban_group / delete_topic / delete_comment / refund_order |
| `block_status` | block 状态 | active / cancelled |

---

# **运营后台 PRD 草案**

## **18. PRD：CaiRobot MVP 运营后台**

### **18.1 背景**

CaiRobot MVP 是一个运营管理综合平台，当前社交域已规划成员、群组、主题、订单、消息、缓存、事件等核心能力。为了支撑平台运营，需要建设一个基于 go-admin 框架的运营后台，帮助运营、审核、客服、财务、系统管理员完成日常治理和配置工作。

后台需要支持用户治理、群组治理、内容审核、评论删除、订单查看、分账管理、多语言维护、App 配置、闪屏广告 Banner 管理、模板消息管理和全局枚举字典维护。

### **18.2 目标**

本 PRD 的目标是定义运营后台的功能范围、模块结构、列表字段、操作列、筛选列、高级筛选列、权限规则、审计要求和数据字典规范。

核心目标包括：

1. 运营管理员可以查看和治理用户。
2. 运营管理员可以查看、审核、禁用群组。
3. 运营管理员可以查看、下架、删除不良主题。
4. 运营管理员可以查看、隐藏、删除不良评论。
5. 财务运营可以查看订单和管理分账。
6. 配置运营可以维护多语言、App 配置、闪屏、广告、Banner。
7. 消息运营可以维护模板消息和查看消息投递。
8. 系统管理员可以维护全局枚举字典。
9. 系统支持大群组广场虚拟成员关系。
10. 所有高权限操作必须记录审计日志。

### **18.3 非目标**

首期不包含以下内容：

```text
1. 完整 BI 数据分析平台；
2. 自动内容审核 AI；
3. 自动分账打款到第三方支付渠道；
4. 复杂广告投放算法；
5. 用户画像推荐系统；
6. 实时客服系统；
7. 多租户 SaaS 化后台。
```

### **18.4 用户角色**

| 后台角色 | 核心权限 |
|---|---|
| 超级管理员 | 全部权限、角色权限配置、系统字典、敏感配置 |
| 运营管理员 | 用户、群组、主题、评论、广告、消息模板管理 |
| 内容审核员 | 主题审核、评论审核、违规内容处理 |
| 客服人员 | 查看用户、订单、消息、发送通知 |
| 财务人员 | 订单查看、退款处理、分账确认、导出报表 |
| 配置管理员 | App 配置、多语言、模板消息、广告 Banner |
| 只读审计员 | 查看审计日志，不可修改业务数据 |

### **18.5 功能模块**

后台首期包含以下模块：

```text
1. 用户管理
2. 群组管理
3. 群组成员与嘉宾管理
4. 主题管理
5. 评论管理
6. 订单管理
7. 分账管理
8. 多语言维护
9. App 配置管理
10. 闪屏/广告/Banner 维护
11. 模板消息管理
12. 消息中心管理
13. 全局枚举字典管理
14. 操作日志与审计
```

### **18.6 核心业务规则**

#### **用户治理规则**

运营管理员可以禁用违规用户。用户被禁用后：

```text
1. 不能登录；
2. 不能发帖；
3. 不能评论；
4. 不能发送私信；
5. 不能购买或续费；
6. 已发布内容是否下架由管理员另行选择。
```

#### **群组治理规则**

运营管理员可以禁用违规群组。群组被禁用后：

```text
1. 群组不可被普通用户访问；
2. 群组主题不可继续发布；
3. 已有主题可统一隐藏或保留为管理员可见；
4. 群组订单暂停创建；
5. 群主和管理员收到系统通知；
6. 操作写入 group_admin_actions 和后台操作日志。
```

#### **主题治理规则**

运营管理员可以下架或删除不良主题：

```text
1. 下架主题后普通用户不可见；
2. 删除主题使用软删除；
3. 删除或下架后需要清理 topic:detail、group:topics、topic:stats 等缓存；
4. 需要通知作者；
5. 需要记录操作原因。
```

#### **评论治理规则**

运营管理员可以隐藏或删除不良评论：

```text
1. 隐藏评论后普通用户不可见；
2. 删除评论使用软删除或状态删除；
3. 更新评论统计；
4. 通知评论作者；
5. 记录操作日志。
```

#### **广场虚拟成员规则**

```text
1. 广场群组是全局默认大群组；
2. 所有注册且 active 的用户默认是广场普通成员；
3. 广场普通成员不写 group_members；
4. 广场管理员和嘉宾写 group_members；
5. 广场主题权限判断需要支持虚拟成员；
6. 广场成员间 block 使用 member_blocks；
7. block 关系影响私信、主题可见性、评论和互动。
```

### **18.7 操作审计要求**

以下操作必须写操作日志：

```text
禁用用户
启用用户
重置密码
禁用群组
启用群组
转让群主
设置管理员
移除管理员
设置嘉宾
移除嘉宾
禁言成员
移除成员
删除主题
下架主题
恢复主题
删除评论
隐藏评论
确认支付
退款
确认分账
标记打款
发布 App 配置
回滚 App 配置
发布广告
下架广告
发布模板消息
修改全局枚举
```

日志字段建议：

| 字段 | 说明 |
|---|---|
| 操作人 ID | admin_user_id |
| 操作人名称 | admin_username |
| 操作模块 | module |
| 操作类型 | action |
| 目标类型 | target_type |
| 目标 ID | target_id |
| 操

日志字段建议，输出中断了，请继续补充

日志字段建议可以继续补充如下：

| 字段 | 说明 |
|---|---|
| 操作人 ID | `admin_user_id` |
| 操作人名称 | `admin_username` |
| 操作人角色 | `admin_role` |
| 操作模块 | `module`，如 user / group / topic / comment / order / settlement / config |
| 操作类型 | `action`，如 ban_user / delete_topic / refund_order |
| 目标类型 | `target_type`，如 user / group / topic / comment / order |
| 目标 ID | `target_id` |
| 目标展示名称 | `target_name`，如用户昵称、群组名称、主题标题 |
| 操作前数据 | `before_data`，JSON |
| 操作后数据 | `after_data`，JSON |
| 变更字段 | `changed_fields`，JSON 数组 |
| 操作原因 | `reason`，管理员填写 |
| 备注 | `remark` |
| 请求 IP | `request_ip` |
| User-Agent | `user_agent` |
| 请求路径 | `request_path` |
| 请求方法 | `request_method` |
| 请求参数摘要 | `request_params`，敏感字段脱敏 |
| 响应状态 | `response_status` |
| 错误信息 | `error_message`，失败时记录 |
| 风险等级 | `risk_level`，low / medium / high / critical |
| 是否敏感操作 | `is_sensitive` |
| 是否需要二次确认 | `require_confirm` |
| 是否二次确认通过 | `confirmed` |
| 确认人 ID | `confirmed_by` |
| 确认时间 | `confirmed_at` |
| 业务追踪 ID | `trace_id` |
| 请求链路 ID | `request_id` |
| 创建时间 | `created_at` |

### **18.8 建议操作日志表**

```sql
CREATE TABLE `admin_operation_logs` (
  `id` char(32) NOT NULL COMMENT '日志 ID',
  `admin_user_id` char(32) NOT NULL COMMENT '操作人 ID',
  `admin_username` varchar(100) NOT NULL COMMENT '操作人名称',
  `admin_role` varchar(100) DEFAULT NULL COMMENT '操作人角色',
  `module` varchar(50) NOT NULL COMMENT '操作模块',
  `action` varchar(80) NOT NULL COMMENT '操作类型',
  `target_type` varchar(50) NOT NULL COMMENT '目标类型',
  `target_id` char(32) DEFAULT NULL COMMENT '目标 ID',
  `target_name` varchar(255) DEFAULT NULL COMMENT '目标展示名称',
  `before_data` json DEFAULT NULL COMMENT '操作前数据',
  `after_data` json DEFAULT NULL COMMENT '操作后数据',
  `changed_fields` json DEFAULT NULL COMMENT '变更字段',
  `reason` varchar(500) DEFAULT NULL COMMENT '操作原因',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `request_ip` varchar(45) DEFAULT NULL COMMENT '请求 IP',
  `user_agent` varchar(500) DEFAULT NULL COMMENT 'User-Agent',
  `request_path` varchar(255) DEFAULT NULL COMMENT '请求路径',
  `request_method` varchar(20) DEFAULT NULL COMMENT '请求方法',
  `request_params` json DEFAULT NULL COMMENT '请求参数摘要，敏感字段脱敏',
  `response_status` varchar(30) DEFAULT NULL COMMENT '响应状态',
  `error_message` text DEFAULT NULL COMMENT '失败错误信息',
  `risk_level` varchar(20) NOT NULL DEFAULT 'low' COMMENT '风险等级 low/medium/high/critical',
  `is_sensitive` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否敏感操作',
  `require_confirm` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否需要二次确认',
  `confirmed` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否二次确认通过',
  `confirmed_by` char(32) DEFAULT NULL COMMENT '确认人 ID',
  `confirmed_at` bigint(20) DEFAULT NULL COMMENT '确认时间',
  `trace_id` varchar(100) DEFAULT NULL COMMENT '业务追踪 ID',
  `request_id` varchar(100) DEFAULT NULL COMMENT '请求链路 ID',
  `created_at` bigint(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_user_time` (`admin_user_id`, `created_at`),
  KEY `idx_module_action_time` (`module`, `action`, `created_at`),
  KEY `idx_target` (`target_type`, `target_id`),
  KEY `idx_risk_level_time` (`risk_level`, `created_at`),
  KEY `idx_trace_id` (`trace_id`),
  KEY `idx_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运营后台操作日志表';
```

### **18.9 高风险操作二次确认规则**

以下操作建议要求二次确认，必要时要求更高权限角色审批：

| 操作 | 风险等级 | 二次确认 |
|---|---|---|
| 禁用用户 | high | 是 |
| 删除用户 | critical | 是 |
| 重置用户密码 | high | 是 |
| 禁用群组 | high | 是 |
| 转让群主 | critical | 是 |
| 删除主题 | high | 是 |
| 批量删除主题 | critical | 是 |
| 删除评论 | medium | 可选 |
| 批量删除评论 | high | 是 |
| 确认支付 | high | 是 |
| 退款 | critical | 是 |
| 确认分账 | critical | 是 |
| 标记打款 | critical | 是 |
| 发布生产 App 配置 | critical | 是 |
| 回滚生产 App 配置 | high | 是 |
| 发布广告 | medium | 可选 |
| 下架广告 | medium | 可选 |
| 修改系统内置枚举 | critical | 是 |
| 删除字典项 | high | 是 |

### **18.10 日志脱敏规则**

操作日志可以记录请求参数和变更前后数据，但必须做脱敏处理，避免后台日志变成敏感数据泄露源。

| 数据类型 | 脱敏规则 |
|---|---|
| 密码 | 不记录 |
| salt | 不记录 |
| token | 不记录 |
| session | 不记录 |
| 手机号 | 保留前三后四，如 `138****5678` |
| 邮箱 | 保留首尾和域名，如 `j***e@example.com` |
| 身份证/证件号 | 仅保留后四位 |
| 银行卡/支付账号 | 仅保留后四位 |
| 私信正文 | 默认只记录摘要，审计权限才可查看完整内容 |
| App 敏感配置 | 只记录 key 和变更动作，不记录完整 value |
| JSON 配置 | 对敏感字段递归脱敏 |

### **18.11 后台权限设计**

运营后台建议使用“菜单权限 + 按钮权限 + 数据权限 + 审计权限”的组合模型。

| 权限类型 | 说明 |
|---|---|
| 菜单权限 | 控制用户能看到哪些后台菜单 |
| 按钮权限 | 控制新增、编辑、删除、导出、审核、退款等按钮 |
| 数据权限 | 控制用户能查看哪些数据范围 |
| 审计权限 | 控制用户是否能查看操作日志、敏感字段、私信内容 |
| 二次确认权限 | 控制用户是否能执行或审批高风险操作 |

### **18.12 建议权限编码**

| 模块 | 权限编码示例 |
|---|---|
| 用户管理 | `admin:user:view` / `admin:user:edit` / `admin:user:ban` / `admin:user:reset_password` |
| 群组管理 | `admin:group:view` / `admin:group:edit` / `admin:group:audit` / `admin:group:ban` / `admin:group:transfer_owner` |
| 群组成员 | `admin:group_member:view` / `admin:group_member:mute` / `admin:group_member:remove` / `admin:group_member:set_guest` / `admin:group_member:set_admin` |
| 主题管理 | `admin:topic:view` / `admin:topic:audit` / `admin:topic:edit` / `admin:topic:hide` / `admin:topic:delete` |
| 评论管理 | `admin:comment:view` / `admin:comment:hide` / `admin:comment:delete` / `admin:comment:batch_delete` |
| 订单管理 | `admin:order:view` / `admin:order:confirm_pay` / `admin:order:cancel` / `admin:order:refund` / `admin:order:export` |
| 分账管理 | `admin:settlement:view` / `admin:settlement:generate` / `admin:settlement:confirm` / `admin:settlement:mark_paid` / `admin:settlement:export` |
| 多语言维护 | `admin:i18n:view` / `admin:i18n:edit` / `admin:i18n:publish` / `admin:i18n:rollback` |
| App 配置 | `admin:config:view` / `admin:config:edit` / `admin:config:publish` / `admin:config:rollback` |
| 广告 Banner | `admin:ad:view` / `admin:ad:edit` / `admin:ad:publish` / `admin:ad:offline` |
| 模板消息 | `admin:template:view` / `admin:template:edit` / `admin:template:test_send` / `admin:template:publish` |
| 消息中心 | `admin:message:view` / `admin:message:retry` / `admin:message:recall` / `admin:message:delete` |
| 全局枚举 | `admin:dict:view` / `admin:dict:edit` / `admin:dict:import` / `admin:dict:export` / `admin:dict:refresh_cache` |
| 操作日志 | `admin:operation_log:view` / `admin:operation_log:export` / `admin:operation_log:view_sensitive` |

### **18.13 go-admin 菜单建议**

```text
运营后台
├── 用户管理
│   ├── 用户列表
│   ├── 用户详情
│   └── Block 关系
├── 群组管理
│   ├── 群组列表
│   ├── 群组成员
│   ├── 嘉宾管理
│   ├── 付费方案
│   └── 群组审计
├── 内容管理
│   ├── 主题管理
│   ├── 评论管理
│   └── 举报管理
├── 订单财务
│   ├── 订单管理
│   ├── 权益管理
│   └── 分账管理
├── 消息中心
│   ├── 消息列表
│   ├── 会话管理
│   └── 模板消息
├── 运营配置
│   ├── 多语言维护
│   ├── App 配置
│   ├── 闪屏广告
│   ├── Banner 管理
│   └── 全局枚举字典
└── 系统审计
    ├── 操作日志
    ├── 登录日志
    └── 高风险审批
```

### **18.14 与社交域服务的对接方式**

运营后台可以分两类对接：

| 类型 | 说明 |
|---|---|
| 直接管理型 | 对字典、多语言、App 配置、广告 Banner 等后台自有表直接 CRUD |
| 业务服务型 | 对用户禁用、群组禁用、主题下架、订单确认、退款、分账等操作必须走 service/usecase，不能绕过业务规则直接改表 |

对于业务服务型操作，应封装后台 service，例如：

```text
AdminUserService
AdminGroupService
AdminTopicService
AdminCommentService
AdminOrderService
AdminSettlementService
AdminConfigService
AdminAuditService
```

这些 service 内部复用社交域 usecase 或调用社交域协议，确保缓存失效、事件发布、统计更新、消息通知、审计日志都能统一执行。

### **18.15 缓存与事件要求**

后台操作修改业务数据后，必须处理缓存和事件：

| 操作 | 缓存处理 | 事件 |
|---|---|---|
| 禁用用户 | 删除 `user:profile:{id}`、`user:session:{id}` | `UserBanned` |
| 启用用户 | 删除 `user:profile:{id}` | `UserActivated` |
| 禁用群组 | 删除 `group:detail:{id}`、`group:list:*` | `GroupBanned` |
| 修改群组 | 删除 `group:detail:{id}`、`group:list:*` | `GroupUpdated` |
| 设置群组推荐 | 删除 `group:list:*`、`plaza:groups:*` | `GroupFeaturedUpdated` |
| 设置嘉宾 | 删除 `group:member:{gid}:{uid}` | `GroupGuestGranted` |
| 移除嘉宾 | 删除 `group:member:{gid}:{uid}` | `GroupGuestRevoked` |
| 下架主题 | 删除 `topic:detail:{id}`、`group:topics:{gid}:*` | `TopicHidden` |
| 删除主题 | 删除 `topic:detail:{id}`、`group:topics:{gid}:*` | `TopicDeleted` |
| 删除评论 | 删除 `topic:comments:{topic_id}:*`、`topic:stats:{topic_id}` | `CommentDeleted` |
| 确认支付 | 删除 `group:member:{gid}:{uid}`、`group:stats:{gid}` | `GroupOrderPaid` |
| 退款 | 删除 `group:member:{gid}:{uid}`、`group:stats:{gid}` | `GroupOrderRefunded` |
| 发布配置 | 删除 `app:config:*` | `AppConfigPublished` |
| 修改字典 | 删除 `dict:*` | `DictUpdated` |

### **18.16 验收标准**

运营后台首期验收建议按模块和横向能力验收。

| 验收项 | 标准 |
|---|---|
| 用户管理 | 可以检索用户、查看详情、禁用/启用用户、查看订单与内容 |
| 群组管理 | 可以检索群组、审核、禁用/启用、设置推荐、查看成员 |
| 广场虚拟成员 | 广场普通成员不落库，active 用户默认具备普通成员权限 |
| 嘉宾管理 | 广场管理员/嘉宾写入 `group_members`，可新增、移除、审计 |
| 主题管理 | 可以审核、下架、恢复、删除主题 |
| 评论管理 | 可以隐藏、恢复、删除评论 |
| 订单管理 | 可以查看订单、确认支付、取消、退款、查看权益 |
| 分账管理 | 可以生成分账单、确认、标记打款、导出 |
| 多语言维护 | 可以新增、编辑、发布、回滚文案 |
| App 配置 | 可以按环境维护配置，生产发布需要二次确认 |
| 广告 Banner | 可以配置上下架、投放时间、跳转目标 |
| 模板消息 | 可以编辑模板、测试发送、发布停用 |
| 消息中心 | 可以查看消息和投递日志，私信正文默认脱敏 |
| 全局枚举 | 可以维护字典类型和字典项，系统内置枚举不可随意删除 |
| 操作日志 | 高风险操作全部记录 before/after/reason/operator |
| 权限控制 | 不同角色只看到授权菜单和按钮 |
| 缓存事件 | 业务变更后缓存失效和事件发布符合要求 |
| PRD-10 已有模块验收 | Config/i18n 8 页面全部通过补测，SDK 联动延迟 ≤100ms | 见上方补充表格 |
| 权限边界测试 | 各角色仅看到授权菜单和按钮，越权返回 403 | 见 docs/admin/menu-permissions.md |
| 广场虚拟成员测试 | 新建 active 用户无需写 group_members 即可在广场发帖 | 见 ADR-plaza-virtual-membership |
| 审计日志完整性测试 | 高风险操作全部记录 before/after/reason/operator | 见 docs/admin/audit-log-spec.md |
| 脱敏正确性测试 | 日志中手机号/邮箱/银行卡等符合脱敏规则 | 见 docs/admin/audit-log-spec.md §4 |

### **18.17 建议文档落点**

| 文档 | 路径 | 内容 |
|---|---|---|
| 运营后台 PRD | `docs/prd/PRD-admin-backend.md` | 后台功能、字段、操作、权限、验收 |
| 全局枚举字典 | `docs/api/enums-dictionary.md` | 所有 dict_type 和 dict_value |
| 后台菜单权限 | `docs/admin/menu-permissions.md` | go-admin 菜单、按钮权限编码 |
| 后台审计规范 | `docs/admin/audit-log-spec.md` | 操作日志字段、脱敏、二次确认 |
| 广场虚拟成员规则 | `docs/adr/ADR-plaza-virtual-membership.md` | 广场普通成员不落库的架构决策 |
| 社交后台 API | `docs/api/admin-social-openapi.yaml` | 后台管理接口定义 |
| CODE-WIKI 更新 | `docs/wiki/CODE-WIKI.md` | 补充 admin、social、dict、audit 模块说明 |

*内容由 AI 生成仅供参考*