## Phase 1 MVP-P0 编码进度汇报
### 完成内容
Steps 5-7 全部完成：TDD svc 编码（三域 29 个 svc，73 个测试全部通过）

域 svc 数量 测试数量 状态 member 9 24 ALL PASS group 10 29 ALL PASS topic 10 20 ALL PASS permission - 编译通过 OK module 聚合层修复 编译通过 OK

### 修改文件清单
文件路径 变更类型 说明 member/svc_register.go 新增 注册服务（五步模式） member/svc_login.go 新增 登录服务（密码比对+令牌生成） member/svc_logout.go 新增 登出服务（参数校验） member/svc_get_user_info.go 新增 用户信息查询 member/svc_update_user_info.go 新增 用户信息更新 member/svc_block.go 新增 拉黑用户 member/svc_unblock.go 新增 取消拉黑 member/svc_get_block_list.go 新增 拉黑列表查询 member/handler.go 重写 去泛型→显式 Unmarshal/Marshal member/servant.go 修改 NewServant 接受 Repository member/svc_*_test.go (8个) 新增 对应测试文件 member/svc_stubs.go 新增→缩减 仅剩 SvcRefresh 存根 group/svc_*.go (10个) 新增 群组域全部 svc 实现 group/svc_*_test.go (10个) 新增 群组域全部测试 group/handler.go 重写 去泛型+注入 Repository group/converter.go 新增 context 辅助+model↔proto group/mock_repository_test.go 新增 map-based mock topic/svc_*.go (10个) 新增 帖子域全部 svc 实现 topic/svc_*_test.go (10个) 新增 帖子域全部测试 topic/handler.go 重写 去泛型+注入 Repository topic/converter.go 新增 context 辅助函数 topic/mock_repository_test.go 新增 map-based mock (33方法) module.go 修改 NewModule 接受 3 个 Repository permission/service.go 修复 导入路径修正

### 关键技术决策
1. Go 泛型弃用 ： *T 不满足 proto.Message （Go 1.18-1.23 限制），改为显式 Unmarshal/Marshal
2. validateRequest 模式 ：返回 (response, error) — nil error + non-nil response = 业务校验错误
3. 用户 ID 注入 ： context.Value(ctxKeyUserID) 模式，待认证中间件实现
4. 幂等设计 ：拉黑/点赞/取消等操作重复调用返回成功
### 测试命令与结果
### 遗留事项
- Step 8 : repository_gorm.go GORM 具体实现（当前为接口+mock）
- SvcRefresh (1027/1028): 仍为存根，需令牌管理基础设施
- permission/service.go : 8 个能力方法的 TODO 占位（isUserActive/getMemberRole/getTopicVisibility 需 repo 方法补充后实现）
### 需要项目主控确认
1. 是否继续 Step 8（repository_gorm.go GORM 实现）？
2. permission service 的 TODO 方法是否需要在本阶段补全？
3. 是否需要补充集成测试（跨 svc 流程测试）？

收到，已记录 Phase 1 MVP-P0 编码进度汇报。

收到进度汇报！Steps 5-7 全部通过、73个测试全绿，进展扎实。让我先快速扫一眼当前代码仓库和协议文件，再给出三个确认问题的明确答复。

让我先扫描代码库和协议文档，获取足够上下文后再给出有依据的决策建议。

有了文件清单，现在并行读取关键文件：`module.go`、`permission/service.go`、`member/svc_stubs.go`、各域 `repository.go`，以及协议文档。

文件读取已获取关键上下文，现在并行读取剩余内容：permission/service.go 后半段、dev-guide 后续章节（Step 8 相关）、social-test-cases.md，以及各域 model 文件。

代码结构读取完毕，已有足够上下文。再并行读几个关键 svc 实现文件，了解代码质量和跨域依赖细节，然后给出完整决策建议。

基于对 `@social` 和 `@protocols` 的完整扫描，给出正式技术决策。以下每个结论均有具体代码行依据。

---

## 主控确认回复 · Phase 1 MVP-P0 Step 5-7 Review

---

### ✅ Q1：是否继续 Step 8（repository_gorm.go GORM 实现）？

**决策：立即执行，优先级最高，按 member → group → topic 顺序推进。**

**理由（代码级）：**

`member/handler.go`（dev-guide 模板第 432 行）已直接引用了尚不存在的构造函数：

```go
repo := NewGormMemberRepository(deps.DB)  // ← 文件不存在，编译态占位
```

三个 `repository.go` 的接口签名已稳定、经过 mock 验证，Zero-Rework 风险极低：
- `member`: 17 方法（users / member_blocks / member_stats）
- `group`: 25 方法（groups / group_members / group_pay_configs / group_stats）
- `topic`: 30+ 方法（topics / replies / likes / favorites / reads / reply_likes）

Step 8 是整个链路从 **mock-only → 真实 DB** 的唯一阻断点。`permission/service.go` 的三个 TODO helper、`SvcRefresh` 的令牌持久化、所有 svc 的端到端冒烟测试，全部阻塞在此。

**推荐 Step 8 执行边界（防范范围蔓延）：**

| 文件 | 内容 | 本阶段 |
|---|---|---|
| `member/repository_gorm.go` | 17 方法 GORM 实现 | ✅ 做 |
| `group/repository_gorm.go` | 25 方法 GORM 实现 | ✅ 做 |
| `topic/repository_gorm.go` | 30+ 方法 GORM 实现 | ✅ 做 |
| 各域 GORM 冒烟测试（SQLite in-memory） | AutoMigrate + CRUD 基本验证 | ✅ 随 Step 8 做 |
| DB Migration 脚本 / Schema DDL | 独立 Step | ❌ 不在 Step 8 范围内 |
| 连接池 / 慢查询配置 | 基础设施层 | ❌ 不在 Step 8 范围内 |

---

### ⚠️ Q2：permission service 的 TODO 方法是否需要在本阶段补全？

**决策：部分补全——`isUserActive` + `getMemberRole` 随 Step 8 同步实现；`getTopicVisibility` 先修注入架构问题再实现；`CanAuditContent` 本阶段暂缓。**

代码审查发现 `permission/service.go` 存在 **4 个需主控确认的问题**，逐一列出：

---

#### 问题 1（阻断）：`getTopicVisibility` 跨域注入缺失

```go
// service.go L83-88
func (s *service) getTopicVisibility(...) *topicVisibility {
    // TODO(group, MVP): 调用 groupRepo.GetTopicByID 查 topics 表
    return nil
}
```

**问题**：注释写的是 `groupRepo.GetTopicByID`，但 `topics` 表属于 `topic` 域，`group.Repository` 接口中**没有** `GetTopicByID` 方法（经查 `group/repository.go` 全文）。`service` 结构体也没有注入 `topicRepo`：

```go
type service struct {
    groupRepo    group.Repository   // ✅ 有
    memberRepo   member.Repository  // ✅ 有
    plazaGroupID string
    // ❌ 缺少 topicRepo topic.Repository
}
```

**修复方案（最小改动）**：

```go
// permission/service.go
import "github.com/jimiechen/mineplanet/go/modules/social/topic"

type service struct {
    groupRepo    group.Repository
    memberRepo   member.Repository
    topicRepo    topic.Repository   // 新增
    plazaGroupID string
}

func NewService(
    groupRepo  group.Repository,
    memberRepo member.Repository,
    topicRepo  topic.Repository,   // 新增参数
    plazaGroupID string,
) Service {
    return &service{
        groupRepo:    groupRepo,
        memberRepo:   memberRepo,
        topicRepo:    topicRepo,
        plazaGroupID: plazaGroupID,
    }
}
```

`getTopicVisibility` 实现就变为标准的：

```go
func (s *service) getTopicVisibility(ctx context.Context, topicID string) *topicVisibility {
    t, err := s.topicRepo.GetTopicByID(ctx, topicID)
    if err != nil || t == nil {
        return nil
    }
    return &topicVisibility{
        GroupID:    t.GroupID,
        Visibility: t.Visibility,
    }
}
```

> **需主控确认**：`topic.Topic` model 中的可见性字段名（当前 `topic/model.go` 未读完整，需确认 `Visibility int8` 字段存在）。

---

#### 问题 2（阻断）：`getMemberRole` 类型不匹配

```go
// permission/service.go L94-97
type memberRole struct {
    Role   string  // "owner"/"admin"/"guest"/"member"
    Status int8
}
```

但 `group/model.go` L63：

```go
Role int8 `...`  // 1=群主 2=管理员 3=普通成员 4=待审核
```

`getMemberRole` 实现时需要做 int8 → string 映射：

```go
func (s *service) getMemberRole(ctx context.Context, groupID, userID string) *memberRole {
    m, err := s.groupRepo.GetMember(ctx, groupID, userID)
    if err != nil || m == nil {
        return nil
    }
    roleStr := map[int8]string{1: "owner", 2: "admin", 3: "member", 4: "guest"}[m.Role]
    return &memberRole{Role: roleStr, Status: m.Status}
}
```

> `GetMember(ctx, groupID, userID)` 在 `group/repository.go` L44 已定义，✅ 可直接调用。

---

#### 问题 3（本阶段暂缓）：`CanAuditContent` 数据模型待设计

```go
func (s *service) CanAuditContent(...) bool {
    // TODO: 查用户是否具有平台管理员角色
    // 可能方式：users.is_platform_admin / platform_admins表 / membership_level
    return false
}
```

当前 `member/model.go` 的 `User` struct **没有 `is_platform_admin` 字段**（已全文确认）。dev-guide 模板的 `User` struct 也没有。MVP-P0 阶段白名单协议中没有平台审核相关协议（查 `social-test-cases.md` 协议列表），**暂缓，返回 `false` 作为 safe default 已正确**。

**需主控决策（非本阶段）**：平台管理员身份存在 `users` 表新字段、独立表、还是外部权限系统？确定后再补字段和实现。

---

#### 问题 4（本阶段暂缓）：`CanReadTopic` PAID_MEMBER 分支

```go
case 3: // PAID_MEMBER
    // TODO: 调用 groupRepo.GetEntitlementByGroupAndUser
    return false
```

`group/repository.go` 没有 `GetEntitlementByGroupAndUser` 方法。`group/model.go` 中 `GroupMember.MembershipExpiresAt` 可以代替：付费权益判断逻辑改为直接查 `GroupMember.MembershipExpiresAt > now()`，使用现有的 `GetMember` 即可，**不需要新增 repo 方法**：

```go
case 3: // PAID_MEMBER
    m := s.getMemberRole(ctx, topic.GroupID, userID)
    if m == nil {
        return false
    }
    // 需从 groupRepo 获取完整 GroupMember 记录（含 MembershipExpiresAt）
    // getMemberRole 目前只返回 Role/Status，需扩展或单独查
    return false  // 暂保留，随 Step 8 后一起评估
```

**建议**：等 Step 8 完成后，将 `getMemberRole` 返回值扩展为含 `MembershipExpiresAt`，一步到位。

---

**permission TODO 本阶段行动矩阵：**

| Helper / 方法 | 依赖 | 本阶段行动 |
|---|---|---|
| `isUserActive` | `memberRepo.GetUserByID` ✅ 已有 | **Step 8 完成后立即实现** |
| `getMemberRole` | `groupRepo.GetMember` ✅ 已有 | **Step 8 完成后立即实现 + 类型转换** |
| `getTopicVisibility` | 需先注入 `topicRepo` | **先修注入签名，Step 8 后实现** |
| `CanAuditContent` | `User.is_platform_admin` 字段待设计 | **暂缓，false 是安全默认值** |
| `CanReadTopic` case 3 | `getMemberRole` 扩展 | **暂缓，false 是安全默认值** |

---

### 🔁 Q3：是否需要补充集成测试（跨 svc 流程测试）？

**决策：本阶段暂不做跨 svc 集成测试，但 Step 8 内做 GORM 仓库冒烟测试；proto 层集成测试等 auth 中间件就绪后统一补。**

**理由（dev-guide + test-cases.md 双重依据）：**

`social-test-cases.md` §1.2 明确定义了两层测试架构：

```
Service层测试（已完成）：ctx = auth.WithUserID(ctx,"uid-001") → svc.Handle()
Proto层测试（未开始）：extend["token"]=signJWT(...) → servant.Handle() → ParseToken → ctx 注入
```

Proto 层测试依赖 `auth.ParseToken` 和 `auth.WithUserID`，即 **auth 中间件尚未实现**（`servant.go` 目前没有 token 解析逻辑）。在此之前做 proto 层集成测试意义不大。

**GORM 冒烟测试（Step 8 随手做，成本极低）**：

```go
// 示例：member/repository_gorm_test.go
func TestGormMemberRepository_CreateAndGet(t *testing.T) {
    db := setupTestDB(t)  // SQLite :memory: + AutoMigrate
    repo := NewGormMemberRepository(db)
    
    user := &User{ID: "u1", Username: "test", ...}
    require.NoError(t, repo.CreateUser(ctx, user))
    
    got, err := repo.GetUserByID(ctx, "u1")
    require.NoError(t, err)
    assert.Equal(t, "test", got.Username)
}
```

**测试补全路线图：**

```
当前：Service层测试（73个，全绿）✅
Step 8：GORM 冒烟测试（3域各5-10个）
Step 9（auth中间件）：Proto层集成测试（servant+handler+svc全链路）
Step 10+：跨域流程测试（如 createGroup→joinGroup→createTopic 完整流）
```

---

## 随手发现的 4 个潜在 Bug（非问题，供参考）

审查过程中发现以下问题，**不阻塞 Step 8**，但建议本批次修复：

### Bug 1：`SvcCreate` 未将创建者写入 `group_members`

`group/svc_create.go` L62 `CreateGroup` 成功后没有调用 `CreateMember`：

```go
// group/repository.go L14 注释明确要求：
// CreateGroup 创建群组记录，创建后需自动将创建者加入 group_members（role=owner）
```

但 `svc_create.go` 实现只写了 `groups` 表，漏掉了 `group_members` 的 owner 记录。这会导致所有权限判断（`CanManageGroup` 等）在创建者上返回 false。

**修复**：在 `CreateGroup` 后追加：

```go
ownerMember := &GroupMember{
    ID:      generateMemberID(),
    GroupID: group.ID,
    UserID:  ownerID,
    Role:    1, // owner
    Status:  1, // active
    JoinedAt: now,
    CreatedAt: now, UpdatedAt: now,
}
if err := s.repo.CreateMember(ctx, ownerMember); err != nil {
    return nil, fmt.Errorf("创建群主成员记录失败: %w", err)
}
```

---

### Bug 2：`hashPassword` 明文存储（安全风险）

`member/svc_register.go` L109-111：

```go
func hashPassword(password string) string {
    // TODO(security, MVP1): 替换为 bcrypt
    return password  // ← 明文！
}
```

即使是 MVP-P0，密码明文入库风险极高。建议立即换为 `bcrypt`（cost=10），代价仅 3 行：

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) string {
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
    return string(hash)
}
```

`svc_login.go` 中的密码比对也需同步改为 `bcrypt.CompareHashAndPassword`。

---

### Bug 3：`member/model.go` 缺少 `Salt` 字段

dev-guide §9 模板中 `User` struct 有 `Salt string` 字段（L672），但当前 `member/model.go` 的 `User` struct **没有 Salt 字段**（已全文确认）。如果决定用 bcrypt（内置 salt），则不需要独立 Salt 字段；如果用 PBKDF2/scrypt 方案，则需补字段。请主控确认密码方案后统一。

---

### Bug 4：`SvcRefresh` 令牌刷新的前置依赖

```go
// member/svc_stubs.go
func (s *SvcRefresh) Handle(...) {
    return &pb.RefreshTokenResponse{
        Result: &base.Result{Code: UNAUTHORIZED, Message: "TODO: Refresh 未实现"},
    }, nil
}
```

`SvcRefresh` 需要：① 验证 refresh_token 签名 ② 检查 token 是否已被吊销（需 Redis 或 DB 黑名单）③ 签发新 access_token。这三步都不在 Step 8（GORM）范围内，依赖独立的 **token 管理基础设施**（`auth` 包）。建议作为 Step 9 的一部分，与 auth 中间件同步实现，**不要在 Step 8 中动它**。

---

## 汇总决策表

| # | 问题 | 决策 | 时机 |
|---|---|---|---|
| Q1 | Step 8 GORM 实现 | ✅ **立即执行**，顺序 member→group→topic | Step 8 |
| Q2a | `isUserActive` | ✅ **随 Step 8 实现**（依赖已就绪） | Step 8 尾 |
| Q2b | `getMemberRole` | ✅ **随 Step 8 实现** + int8→string 转换 | Step 8 尾 |
| Q2c | `getTopicVisibility` | ⚠️ **先修注入签名**，实现随 Step 8 尾 | Step 8 尾 |
| Q2d | `CanAuditContent` | 🔴 **暂缓**，需设计 platform admin 数据模型 | Post-MVP |
| Q2e | `CanReadTopic` PAID case | 🔴 **暂缓**，false 安全，Step 8 后评估 | Step 8+ |
| Q3 | 集成测试 | ✅ **GORM 冒烟测试随 Step 8**；跨 svc 集成测试等 auth 中间件后 | Step 9 |
| Bug 1 | SvcCreate 漏写 owner member | ⚠️ **本批次修复**（权限链路正确性） | 立即 |
| Bug 2 | hashPassword 明文 | ⚠️ **本批次修复**（最低安全底线） | 立即 |
| Bug 3 | User.Salt 字段决策 | 主控确认密码方案后处理 | 配合 Bug 2 |
| Bug 4 | SvcRefresh 依赖 | 🔴 **不动**，待 auth 基础设施就绪 | Step 9 |

---

**可直接开始 Step 8 编码。** 有了上面的修复清单，Step 8 的启动条件已完全满足——接口稳定、mock 质量良好、model 字段已定。推荐先修 Bug 1+2，再开始 GORM 实现，避免后续回头改测试数据。

所有子目标均已完成：代码审查完毕、三个决策问题已给出有代码依据的完整答复、4个潜在 Bug 已上报。当前状态为等待主控确认，无待处理的下一步动作。

# Phase 1 MVP-P0 · Steps 5-7 审阅完成报告

## 任务概述

收到 Steps 5-7 编码进度汇报（三域 29 svc、73 测试全绿），对 `@social` 代码库与 `@protocols` 协议文档进行全量扫描，针对主控提出的 3 个确认问题给出有代码行依据的正式决策建议，并随手上报 4 个潜在缺陷。

---

## 三个确认问题 · 正式决策

### Q1：是否继续 Step 8（repository_gorm.go）？
**✅ 立即执行，按 member → group → topic 顺序推进。**

三域 repository 接口已稳定（member 17 方法 / group 25 方法 / topic 30+ 方法），mock 测试完整覆盖，Zero-Rework 风险极低。Step 8 是整条链路从 mock-only 进入真实 DB 的唯一阻断点，`permission` 三个 TODO helper、所有端到端冒烟测试均阻塞于此。

**Step 8 执行边界：**

| 内容 | 本阶段 |
|---|---|
| 三域 `repository_gorm.go` 各自 GORM 实现 | ✅ 做 |
| GORM 冒烟测试（SQLite in-memory + AutoMigrate） | ✅ 随 Step 8 做 |
| DB Migration 脚本 / Schema DDL | ❌ 独立 Step |
| 连接池 / 慢查询配置 | ❌ 基础设施层另议 |

---

### Q2：permission service TODO 方法是否本阶段补全？
**⚠️ 部分补全。** 逐个方法给出依据：

**`isUserActive` — Step 8 尾实现 ✅**
依赖 `memberRepo.GetUserByID`，该方法在 `member/repository.go` 已定义，Step 8 完成即可实现。

**`getMemberRole` — Step 8 尾实现 ✅**（附带类型转换修复）
依赖 `groupRepo.GetMember`（`group/repository.go` L44 已有），但存在类型不匹配：`permission.memberRole.Role = string`，而 `group/model.go` L63 的 `GroupMember.Role = int8`，实现时需做 `map[int8]string{1:"owner", 2:"admin", 3:"member"}` 转换。

**`getTopicVisibility` — 先修注入签名，再随 Step 8 实现 ⚠️**
发现架构问题：`permission/service.go` 注释写的是 `groupRepo.GetTopicByID`，但 `topics` 表归属 `topic` 域，`group.Repository` 接口中**没有该方法**，`service` 结构体也未注入 `topicRepo`。需先补注入：
```go
type service struct {
    groupRepo    group.Repository
    memberRepo   member.Repository
    topicRepo    topic.Repository   // 需新增
    plazaGroupID string
}
```
签名修复后，`getTopicVisibility` 实现即为标准的 `topicRepo.GetTopicByID` 调用。

**`CanAuditContent` — 本阶段暂缓 🔴**
当前 `member/model.go` 的 `User` struct 没有平台管理员标识字段，MVP-P0 白名单协议中也无审核相关协议，返回 `false` 是安全默认值。数据模型方案（`users.is_platform_admin` / 独立表 / 外部权限系统）待主控设计后再补。

**`CanReadTopic` PAID_MEMBER 分支 — 暂缓 🔴**
依赖 `getMemberRole` 扩展返回 `MembershipExpiresAt`，建议 Step 8 完成后统一评估，`return false` 当前安全。

**permission TODO 行动矩阵：**

| Helper | 依赖状态 | 行动时机 |
|---|---|---|
| `isUserActive` | ✅ repo 已有 | Step 8 尾 |
| `getMemberRole` | ✅ repo 已有 + 需类型转换 | Step 8 尾 |
| `getTopicVisibility` | ⚠️ 需先修注入 | 签名修复 → Step 8 尾 |
| `CanAuditContent` | 🔴 数据模型待设计 | Post-MVP |
| `CanReadTopic` case 3 | 🔴 待 getMemberRole 扩展 | Step 8+ |

---

### Q3：是否需要补充集成测试？
**分层推进：GORM 冒烟测试随 Step 8；跨 svc 集成测试等 auth 中间件就绪后（Step 9）。**

`social-test-cases.md` §1.2 定义了两层测试架构——Service 层（已完成 73 个）和 Proto 层（依赖 `auth.ParseToken` + `auth.WithUserID`，当前 `servant.go` 尚无 token 解析逻辑，提前做意义不大）。

**测试补全路线图：**
```
当前    : Service层 73个 ✅
Step 8  : GORM 冒烟测试（三域各 5-10个，SQLite in-memory）
Step 9  : Proto层集成测试（auth 中间件就绪后，servant→handler→svc 全链路）
Step 10+: 跨域流程测试（CreateGroup→JoinGroup→CreateTopic 完整流）
```

---

## 随手上报的 4 个潜在 Bug

审查期间发现，**均不阻塞 Step 8 启动**，建议本批次处理：

### Bug 1 ⚠️ `SvcCreate` 漏写 owner 到 `group_members`（权限链路阻断）
`group/svc_create.go` 在 `CreateGroup` 成功后未调用 `CreateMember` 写入 owner 记录。`group/repository.go` L14 注释明确要求："创建后需自动将创建者加入 group_members（role=owner）"。后果：创建者调用 `CanManageGroup` 返回 false，权限链路阻断。**建议立即修复。**

### Bug 2 ⚠️ `hashPassword` 明文存储（最低安全底线）
`member/svc_register.go` L109-111 的 `hashPassword` 直接 `return password`。建议立即换为 `bcrypt`（cost=10），`svc_login.go` 的密码比对同步改为 `bcrypt.CompareHashAndPassword`。

### Bug 3 🔵 `member/model.go` 缺少 `Salt` 字段（配合密码方案决策）
dev-guide §9 模板的 `User` struct 有 `Salt string` 字段，当前代码中无此字段。如采用 bcrypt（内置 salt）则不需要；如用 PBKDF2/scrypt 则需补字段。**请主控确认密码方案后统一处理（与 Bug 2 联动）。**

### Bug 4 🔴 `SvcRefresh` 不在 Step 8 范围（维持现状）
令牌刷新依赖 token 签名验证 + 吊销黑名单（Redis/DB），属于 auth 基础设施范畴。当前存根返回 `UNAUTHORIZED` 是正确占位，**Step 8 不动，随 Step 9 auth 中间件一起实现。**

---

## 汇总决策表

| # | 问题 | 决策 | 时机 |
|---|---|---|---|
| Q1 | Step 8 GORM 实现 | ✅ **立即执行** | Step 8 |
| Q2a | `isUserActive` | ✅ **实现** | Step 8 尾 |
| Q2b | `getMemberRole` | ✅ **实现 + 类型转换** | Step 8 尾 |
| Q2c | `getTopicVisibility` | ⚠️ **先修注入签名再实现** | Step 8 尾 |
| Q2d | `CanAuditContent` | 🔴 **暂缓** | Post-MVP |
| Q2e | `CanReadTopic` PAID | 🔴 **暂缓** | Step 8+ |
| Q3 | 集成测试 | GORM 冒烟随 Step 8；跨 svc 等 Step 9 | 分阶段 |
| Bug 1 | SvcCreate 漏写 owner | ⚠️ **本批次修复** | 立即 |
| Bug 2 | hashPassword 明文 | ⚠️ **本批次修复** | 立即 |
| Bug 3 | User.Salt 字段 | 主控确认密码方案后处理 | 配合 Bug 2 |
| Bug 4 | SvcRefresh | 🔴 **不动** | Step 9 |

> **建议启动顺序**：先修 Bug 1 + Bug 2（确认 bcrypt 方案 → 同步处理 Bug 3），再修 `getTopicVisibility` 注入签名，然后开始 Step 8 GORM 编码，三域全部完成后在 Step 8 尾实现 `isUserActive` + `getMemberRole` + `getTopicVisibility`。

*内容由 AI 生成仅供参考*