# Phase 1：Social MVP-P0 核心链路实施方案

> 文档版本：v1.0
> 编写日期：2026-06-18
> 状态：**待主控确认**
> 基础文档：
> - [social-mvp-implementation-plan.md](../reports/social-mvp-implementation-plan.md) — Phase 0 实施方案（v2）
> - [PRD-social-app-mvp.md](../prd/PRD-social-app-mvp.md)
> - [ADR-social-data-level-and-cache-strategy.md](../adr/ADR-social-data-level-and-cache-strategy.md)
> - [ADR-plaza-virtual-membership.md](../adr/ADR-plaza-virtual-membership.md)

---

## 1. 当前状态审计（Phase 0 产出物清单）

### 1.1 已完成项（Phase 0 全部验收通过）

| # | 产出物 | 状态 | 路径 |
|---|--------|------|------|
| P0-01 | Proto 定义（3 文件） | ✅ 完成 | `proto/social/{member,group,topic}.proto` + `proto/base/common.proto` |
| P0-02 | Proto Go 生成代码 | ✅ 完成 | `proto/generated/go/{base,social}/` |
| P0-03 | Go 项目骨架（按域分子包） | ✅ 完成 | `go/modules/social/` |
| P1-01 | SQL 迁移脚本 | ✅ 完成 | `scripts/sql/` + `scripts/migration/004_social_domain_tables.sql` |
| P1-02 | 环境配置模板 | ✅ 完成 | `configs/social/social.{local,staging,prod}.conf` |
| P1-03 | Gateway 路由表 | ✅ 完成 | `configs/gateway/routes.yaml`（25 条社交域路由） |
| P1-04 | Gateway CORS 中间件 | ✅ 完成 | `go/gateway/proto-gateway/internal/middleware/cors.go` |
| P1-05 | Gateway Auth JWT 模块 | ✅ 完成 | `go/tars/auth/jwt.go` |

### 1.2 Social 模块代码完成度详查

#### Member 域（maxType=1000）— **11 个 svc，全部已实现**

| svc 文件 | minType | 协议对 | 状态 | 行数 | 五步模式 |
|----------|---------|--------|------|------|----------|
| `svc_register.go` | 1021 | UserRegister | ✅ 完整实现 | ~145 | ✅ 校验→唯一性检查→创建用户→事件→响应 |
| `svc_login.go` | 1023 | UserLogin | ✅ 完整实现 | ~128 | ✅ 校验→查用户→密码比对→JWT生成→响应 |
| `svc_logout.go` | 1025 | UserLogout | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_refresh.go` | 1027 | RefreshToken | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_get_user_info.go` | 1029 | GetUserInfo | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_update_user_info.go` | 1031 | UpdateUserInfo | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_block.go` | 1039 | BlockUser | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_unblock.go` | 1041 | UnblockUser | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_get_block_list.go` | 1043 | GetBlockList | ✅ 已实现 | - | 需验证五步完整性 |
| `svc_update_member_status.go` | 1033 | UpdateMemberStatus | ✅ 已实现 | - | P1-E 补齐 |
| `svc_get_user_stats.go` | 1045 | GetUserStats | ✅ 已实现 | - | P1-E 补齐 |

**Member 基础设施：**
- `model.go` — User / MemberBlock / MemberStats ✅
- `repository.go` — 17 个方法接口定义 ✅
- `repository_gorm.go` — GORM 实现 ✅
- `handler.go` — 12 个 case 分发 ✅
- `servant.go` — TarsGo Servant ✅
- 测试文件 — 每个 svc 有对应 `_test.go` ✅

#### Group 域（maxType=2000）— **12 个 svc，全部已实现**

| svc 文件 | minType | 协议对 | 状态 | 备注 |
|----------|---------|--------|------|------|
| `svc_create.go` | 2005 | CreateGroup | ✅ 完整实现 | 含自动创建群主成员记录 |
| `svc_join.go` | 2013 | JoinGroup | ✅ 完整实现 | 含幂等检查 |
| `svc_leave.go` | 2015 | LeaveGroup | ✅ 已实现 | - |
| `svc_enter.go` | 2087 | GroupUserEnter | ✅ 已实现 | **修正后正确编号** |
| `svc_renew.go` | 2037 | RenewMember | ✅ 已实现 | - |
| `svc_calc_payable.go` | 2073 | CalcPayableAmount | ✅ 已实现 | - |
| `svc_mute.go` | 2019 | MuteMember | ✅ 已实现 | - |
| `svc_ban.go` | 2023 | BanMember | ✅ 已实现 | - |
| `svc_remove.go` | 2027 | RemoveMember | ✅ 已实现 | - |
| `svc_update_role.go` | 2029 | UpdateMemberRole | ✅ 已实现 | - |
| `svc_batch_get_groups.go` | 2047 | BatchGetGroups | ✅ 已实现 | P1-E 补齐 |
| `svc_get_group_stats.go` | 2039 | GetGroupStats | ✅ 已实现 | P1-E 补齐 |
| `svc_get_group_member_user_ids.go` | 2077 | GetGroupMemberUserIds | ✅ 已实现 | P1-E 补齐 |

**Group 基础设施：**
- `model.go` — Group / GroupMember / GroupPayConfig ✅
- `repository.go` — 接口定义 ✅
- `repository_gorm.go` — GORM 实现 ✅
- `handler.go` — 13 个 case 分发 ✅
- `servant.go` — TarsGo Servant ✅
- `converter.go` — proto↔model 转换函数 ✅
- 测试文件 — 每个 svc 有对应 `_test.go` ✅

#### Topic 域（maxType=3000）— **14 个 svc，全部已实现**

| svc 文件 | minType | 协议对 | 状态 | 备注 |
|----------|---------|--------|------|------|
| `svc_create_topic.go` | 3001 | CreateTopic | ✅ 完整实现 | - |
| `svc_list_topic.go` | 3005 | GetTopicList | ✅ 已实现 | - |
| `svc_delete_topic.go` | 3009 | DeleteTopic | ✅ 已实现 | - |
| `svc_get_topic_detail.go` | 3003 | GetTopicDetail | ✅ 已实现 | - |
| `svc_update_topic.go` | 3007 | UpdateTopic | ✅ 已实现 | - |
| `svc_read_topic.go` | 3011 | ReadTopic | ✅ 已实现 | - |
| `svc_reply_topic.go` | 3043 | AddTopicReply | ✅ 已实现 | - |
| `svc_like_topic.go` | 3061 | LikeTopic | ✅ 已实现 | - |
| `svc_unlike_topic.go` | 3063 | UnlikeTopic | ✅ 已实现 | - |
| `svc_favorite_topic.go` | 3065 | FavoriteTopic | ✅ 已实现 | - |
| `svc_unfavorite_topic.go` | 3067 | UnfavoriteTopic | ✅ 已实现 | - |
| `svc_get_reply_list.go` | 3065 | GetReplyList | ✅ 已实现 | - |
| `svc_create_report.go` | 3095 | CreateReport | ✅ 已实现 | - |

**Topic 基础设施：**
- `model.go` — Topic / TopicReply / TopicLike / TopicFavorite / TopicRead / Report ✅
- `repository.go` — 接口定义 ✅
- `repository_gorm.go` — GORM 实现 ✅
- `handler.go` — 分发器 ✅
- `servant.go` — TarsGo Servant ✅
- 测试文件 — 每个 svc 有对应 `_test.go` ✅

#### 跨域基础设施 — **全部已完成**

| 组件 | 状态 | 说明 |
|------|------|------|
| `module.go` | ✅ | 聚合 3 域 Servant，支持 Publisher 注入 |
| `permission/service.go` | ✅ | **8 个权限方法全部实现**，含广场虚拟成员特化 |
| `permission/service_test.go` | ✅ | Mock 测试覆盖 |
| `permission/mock_repository_test.go` | ✅ | Mock Repository 测试数据 |
| `event/` (8 文件) | ✅ | 事件类型、payload、publisher（memory_bus/redis_pubsub/noop） |
| `eventhandler/handler.go` | ✅ | 事件处理器注册 |
| `cache/keys.go` | ✅ | 缓存 Key 定义（member:/group:/topic: 前缀） |
| `cache/invalidator.go` | ✅ | 缓存失效器 |

### 1.3 关键发现

**发现 1：Phase 0 的产出远超预期。** 原计划 Phase 0 只做"骨架+模板"，实际已完成：
- 37 个 svc 业务逻辑文件（非空骨架）
- 37+ 个测试文件
- Permission Service 8 方法完整实现含广场特化
- Event 系统（3 种 publisher 实现）
- Cache 层（Key 定义 + 失效器）
- Gateway 路由表 25 条完整配置

**发现 2：Phase 1 的核心工作从"编写 svc"转变为"集成联调+测试验证"。**

---

## 2. Phase 1 目标与范围

### 2.1 目标

将 Social 模块从"代码编写完成"推进到"可端到端运行"，确保：

1. **编译通过**：`go build ./...` 零错误
2. **单元测试通过**：`go test ./go/modules/social/...` 全绿
3. **Gateway ↔ Social 集成**：LocalInvoker 正确路由到 Social Servant
4. **Proto Tester 可发送请求**：前端 → Gateway → Social → 返回正确 Response
5. **E2E 验证**：MVP-P0 白名单中 10 对核心协议可通过 Proto Tester 发送并收到正确响应

### 2.2 范围边界

| 包含 | 不包含 |
|------|--------|
| 编译错误修复 | 新增协议（超出 MVP-P0 白名单） |
| 单元测试补全与修复 | Redis 真实连接测试（需基础设施） |
| Gateway ↔ Social 集成 | MySQL 真实连接测试（需基础设施） |
| Proto Tester E2E 验证 | third(4000)/inbox(5000) 域 |
| 文档同步更新 | Admin 后台管理界面 |

---

## 3. 缺口分析（Gap Analysis）

### 3.1 P0 缺口（阻塞 E2E）

| # | 缺口描述 | 影响范围 | 修复方式 |
|---|----------|----------|----------|
| G-001 | **Gateway LocalInvoker 未注册 Social Servant** | 所有社交协议返回 `unsupported route` | 在 invoker_assemble.go 中注册 member/group/topic 三个 Handler |
| G-002 | **go.mod import 路径不匹配** | 编译失败 | social 模块使用 `github.com/jimiechen/mineplanet/protocols/generated/go/social` 但 go.mod 可能未声明此路径 |
| G-003 | **context 用户身份传递缺失** | join/create_topic 等 svc 需要 userID | 实现 context 传递机制（JWT 解析 → 注入 context） |
| G-004 | **repository_gorm.go 可能引用了未实现的 DB 方法** | 运行时 panic | 逐个验证 GORM 实现 vs 接口定义一致性 |

### 3.2 P1 缺口（影响测试质量）

| # | 缺口描述 | 影响 | 修复方式 |
|---|----------|------|----------|
| G-005 | 部分 svc 缺少 Step 2 权限校验调用 permission.Service | 越权操作未被拦截 | 补充 CanXxx 调用 |
| G-006 | 事件发布使用 fmt.Printf 而非结构化日志 | 生产环境不可观测 | 引入 log 接口或保留 NoopPublisher 行为 |
| G-007 | generateUserID() 使用时间戳而非 ULID | ID 可预测/冲突 | MVP-P0 可接受，P1 改为 ULID |
| G-008 | 密码哈希 cost=10 可能不够安全 | 安全审计项 | MVP-P0 可接受 |

### 3.3 P2 缺口（优化项）

| # | 缺口描述 | 影响 | 修复方式 |
|---|----------|------|----------|
| G-009 | handler.Dispatch 大量重复 Unmarshal/Marshal 代码 | 可维护性 | 抽取泛型 dispatch 辅助函数（DevGuide 提到的 dispatchProto） |
| G-010 | converter.go 转换逻辑分散在各 svc 中 | 一致性风险 | 统一到 converter 包 |

---

## 4. 实施任务分解

### 4.1 Task 1：编译修复与依赖整理（预估 0.5 天）

**目标**：`go build ./go/modules/social/...` 零错误

| 步骤 | 操作 | 验证标准 |
|------|------|----------|
| 1.1 | 检查 `go/modules/social/go.mod` 和 `go/go.work` 的模块路径声明 | 无循环依赖 |
| 1.2 | 检查所有 import 路径是否与 go.mod module name 匹配 | `go build` 通过 |
| 1.3 | 检查 `proto/generated/go/social/` 的 go.mod 是否正确 | 无 vendor 冲突 |
| 1.4 | 运行 `go vet ./go/modules/social/...` | 无 vet 警告 |
| 1.5 | 运行 `gofmt -l ./go/modules/social/...` | 无格式问题 |

**涉及文件**：
- `go/modules/social/go.mod`
- `go/go.work`
- `go/go.work.sum`
- `proto/generated/go/social/go.mod`

---

### 4.2 Task 2：Gateway ↔ Social 集成（预估 1 天）

**目标**：Gateway LocalInvoker 能正确路由社交协议到 Social Servant

#### 4.2.1 注册 Social Servant 到 LocalInvoker

当前 `invoker_assemble.go` 仅注册 Hello/Health handler。需要扩展：

```go
// go/gateway/proto-gateway/tarsclient/invoker_assemble.go

// 新增：创建 Social Module 并注册到 LocalInvoker
func RegisterSocialHandlers(invoker *LocalInvoker, socialModule *social.Module) {
    // Member 域 (maxType=1000)
    invoker.RegisterHandler(
        TargetKey{App: "CaiRobotSocialApp", Server: "SocialServer", Servant: "SocialObj", Method: "HandleMember"},
        func(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
            return socialModule.MemberServant.Handle(ctx, req, extend)
        },
    )

    // Group 域 (maxType=2000)
    invoker.RegisterHandler(
        TargetKey{App: "CaiRobotSocialApp", Server: "SocialServer", Servant: "SocialObj", Method: "HandleGroup"},
        func(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
            return socialModule.GroupServant.Handle(ctx, req, extend)
        },
    )

    // Topic 域 (maxType=3000)
    invoker.RegisterHandler(
        TargetKey{App: "CaiRobotSocialApp", Server: "SocialServer", Servant: "SocialObj", Method: "HandleTopic"},
        func(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
            return socialModule.TopicServant.Handle(ctx, req, extend)
        },
    )
}
```

**涉及文件**：
- `go/gateway/proto-gateway/tarsclient/invoker_assemble.go` — 修改
- `go/gateway/proto-gateway/cmd/server/main.go` — 修改（初始化 Social Module 并注册）

#### 4.2.2 routes.yaml 与 Invoker 路由键对齐

**关键映射关系**（必须严格一致）：

| routes.yaml tars_method | Invoker TargetKey.Method | Social Servant |
|------------------------|---------------------------|----------------|
| HandleMember | HandleMember | member.Servant.Handle |
| HandleGroup | HandleGroup | group.Servant.Handle |
| HandleTopic | HandleTopic | topic.Servant.Handle |

**当前状态**：routes.yaml 已正确填写上述映射 ✅

#### 4.2.3 main.go 启动链路修改

```go
// cmd/server/main.go 新增内容

// 1. 初始化 Mock Repository（MVP-P0 用 Memory Repo，后续替换为 GORM）
memberRepo := member.NewMemoryRepository()
groupRepo := group.NewMemoryRepository()
topicRepo := topic.NewMemoryRepository()

// 2. 创建 Permission Service
permSvc := permission.NewService(groupRepo, memberRepo, topicRepo, "plaza_global_001")

// 3. 创建 Social Module
socialModule := social.NewModule(memberRepo, groupRepo, topicRepo)

// 4. 注册到 Gateway Invoker
RegisterSocialHandlers(invoker, socialModule)
```

**注意**：MVP-P0 阶段使用 Memory Repository（无需真实数据库），确保 E2E 可跑通。MySQL/GORM 连接在后续 Task 中接入。

---

### 4.3 Task 3：单元测试修复与补全（预估 1-2 天）

**目标**：`go test ./go/modules/social/...` 全绿

#### 4.3.1 测试策略

| 测试层级 | 方式 | 覆盖范围 |
|----------|------|----------|
| 单元测试 | Mock Repository | 每个 svc 的正常/异常/边界路径 |
| 集成测试 | Memory Repository | 跨 svc 场景（register → login → create_group → join） |
| Handler 分发测试 | 直接调用 Dispatch | Unmarshal/Marshal 路径覆盖 |

#### 4.3.2 必须覆盖的测试维度（DevGuide §14）

每个 svc 至少覆盖以下场景：

| 维度 | member 示例 | group 示例 | topic 示例 |
|------|-------------|-------------|------------|
| 正常路径 | 注册新用户成功 | 创建圈子成功 | 发帖成功 |
| 参数校验失败 | 用户名为空 | 名称为空 | 标题为空 |
| 唯一性冲突 | 用户名已占用 | slug 已占用 | - |
| 权限拒绝 | - | 非管理员禁言成员 | - |
| DB 失败 | CreateUser 返回 error | CreateGroup 返回 error | CreateTopic 返回 error |
| 事件发布 | 注册成功触发 MemberRegistered | 创建成功触发 GroupCreated | 发帖成功触发 TopicCreated |

#### 4.3.3 Mock Repository 实现

需要确保每个域的 `mock_repository_test.go` 提供了完整的 Mock 实现：

```go
// 以 member 为例
type mockRepository struct {
    users      map[string]*User
    blocks     map[string][]*MemberBlock
    stats      map[string]*MemberStats
    // ...
}

func NewMockRepository() *mockRepository { ... }
// 实现 Repository 接口的全部 17 个方法
```

**当前状态**：已有 `mock_repository_test.go` 文件 ✅，需验证接口完整性。

---

### 4.4 Task 4：Proto Tester E2E 验证（预估 1 天）

**目标**：通过 Proto Tester 前端发送 10 对核心协议，收到正确 Response

#### 4.4.1 E2E 验证矩阵（MVP-P0 核心白名单）

| # | 协议对 | minType | 域 | 优先级 | 验证场景 |
|---|--------|---------|-----|--------|----------|
| E2E-01 | UserRegister | 1021 | member | P0 | 正常注册 + 用户名重复 + 参数为空 |
| E2E-02 | UserLogin | 1023 | member | P0 | 正常登录 + 密码错误 + 用户不存在 |
| E2E-03 | GetUserInfo | 1029 | member | P0 | 查询自己信息 + 查询他人信息 |
| E2E-04 | CreateGroup | 2005 | group | P0 | 正常创建 + slug 重复 + 名称为空 |
| E2E-05 | JoinGroup | 2013 | group | P0 | 正常加入 + 重复加入 + 群组不存在 |
| E2E-06 | GroupUserEnter | 2087 | group | P0 | 进入广场 + 进入普通群组 |
| E2E-07 | RenewMember | 2037 | group | P0 | 正常续费 + 非成员续费 |
| E2E-08 | CalcPayableAmount | 2073 | group | P0 | 免费群计算 + 付费群计算 |
| E2E-09 | CreateTopic | 3001 | topic | P0 | 正常发帖 + 标题为空 |
| E2E-10 | GetTopicList | 3005 | topic | P0 | 空列表 + 有数据列表 |

#### 4.4.2 E2E 测试步骤

```
1. 启动 Gateway（make gateway-restart）
2. 打开 Proto Tester（make proto-tester-dev 或手动启动 Vite dev server）
3. 选择协议 → 填写参数 → 发送请求
4. 验证响应：
   a. HTTP Status = 200
   b. _meta.code = 0（成功）或对应 ErrorCode
   c. _data._raw 可反序列化为对应 Proto Response
   d. 业务字段值符合预期
```

#### 4.4.3 E2E 测试产物

| 产物 | 路径 | 格式 |
|------|------|------|
| 测试报告 | `docs/reports/testing/e2e-social-phase1.md` | Markdown |
| 截图证据 | `docs/reports/evidence/screenshots/` | PNG |
| 请求/响应日志 | `docs/reports/evidence/logs/` | JSON |

---

### 4.5 Task 5：权限校验补充（预估 0.5 天）

**目标**：所有需要权限的 svc 都调用了 permission.Service

#### 4.5.1 需要补充权限校验的 svc

| svc | 当前状态 | 需要的权限方法 |
|-----|----------|---------------|
| `svc_ban.go` | 可能缺少 | CanManageMember |
| `svc_remove.go` | 可能缺少 | CanManageMember |
| `svc_mute.go` | 可能缺少 | CanManageMember |
| `svc_update_role.go` | 可能缺少 | CanManageMember |
| `svc_create_topic.go` | 公开操作 | CanPublishTopic（可选） |
| `svc_reply_topic.go` | 可能缺少 | CanReadTopic + CanPublishTopic |
| `svc_like_topic.go` | 可能缺少 | CanReadTopic |
| `svc_favorite_topic.go` | 可能缺少 | CanReadTopic |
| `svc_delete_topic.go` | 可能缺少 | CanManageGroup |
| `svc_leave.go` | 公开操作 | 无需额外权限 |

#### 4.5.2 权限注入方式

Permission Service 需要注入到各 svc 构造函数中：

```go
// 修改前
func NewSvcBan(repo Repository, publisher event.Publisher) *SvcBanMember { ... }

// 修改后
func NewSvcBan(repo Repository, perm permission.Service, publisher event.Publisher) *SvcBanMember { ... }
```

**影响范围**：
- 各 svc 的 `New*()` 构造函数签名变更
- Handler 中 `New*()` 调用处同步修改
- `module.go` 中构造参数传递

---

### 4.6 Task 6：文档同步与收尾（预估 0.5 天）

| 文档 | 变更内容 |
|------|----------|
| `CODE-WIKI.md` | 更新 Phase 1 完成状态 |
| `social-mvp-implementation-plan.md` | 更新 Phase 1 执行结果 |
| 协议编号注册表 | 如有新增编号则更新 |
| E2E 测试报告 | 新建 `e2e-social-phase1.md` |

---

## 5. 执行顺序与依赖关系

```
Task 1: 编译修复 ─────────────────────────────┐
                                            │
Task 2: Gateway集成 ◄─── 依赖 Task 1 编译通过 ─┤
         │                                     │
         ├─► Task 2.1: Invoker 注册             │
         ├─► Task 2.2: main.py 启动链路          │
         │                                     │
Task 3: 单元测试 ◄─── 依赖 Task 1+2 ───────────┤
         │                                     │
Task 5: 权限校验 ◄─── 可与 Task 3 并行 ─────────┤
         │                                     │
Task 4: E2E验证 ◄─── 依赖 Task 2+3+5 ──────────┘
         │
Task 6: 文档同步 ◄─── 依赖 Task 4
```

**建议执行顺序**：Task 1 → Task 2 → (Task 3 || Task 5) → Task 4 → Task 6

---

## 6. 验收标准

### 6.1 必须通过（P0）

| # | 检查项 | 检查方式 | 通过标准 |
|---|--------|----------|----------|
| A-01 | Go 编译 | `go build ./go/modules/social/...` | 零错误零警告 |
| A-02 | Go 单元测试 | `go test ./go/modules/social/... -v` | 全部 PASS |
| A-03 | Gateway 启动 | `make gateway-restart` | :8080 LISTEN |
| A-04 | CORS OPTIONS | `curl -X OPTIONS http://localhost:8080/api/hello` | 204 + CORS 头 |
| A-05 | E2E-01 UserRegister | Proto Tester 发送 | HTTP 200 + Result.Code=0 |
| A-06 | E2E-02 UserLogin | Proto Tester 发送 | HTTP 200 + AccessToken 非空 |
| A-07 | E2E-04 CreateGroup | Proto Tester 发送 | HTTP 200 + GroupId 非空 |
| A-08 | E2E-09 CreateTopic | Proto Tester 发送 | HTTP 200 + TopicId 非空 |

### 6.2 建议通过（P1）

| # | 检查项 | 检查方式 | 通过标准 |
|---|--------|----------|----------|
| B-01 | 10 对 E2E 全部通过 | Proto Tester | 见 4.4.1 矩阵 |
| B-02 | 权限校验覆盖 | Code Review | ban/remove/mute/update_role 调用了 CanManageMember |
| B-03 | 测试覆盖率 | `go test -cover` | >= 70%（MVP-P0 最低线） |

### 6.3 记录即可（P2）

| # | 检查项 |
|---|--------|
| C-01 | fmt.Printf 替换为日志（G-006） |
| C-02 | generateUserID 改为 ULID（G-007） |
| C-03 | Dispatch 泛型抽取（G-009） |

---

## 7. 风险评估

| 风险ID | 等级 | 描述 | 缓解措施 |
|--------|------|------|----------|
| R-P1-001 | **R0** | go.mod 循环依赖导致编译失败 | Task 1 第一步即排查；必要时拆分子模块 |
| R-P1-002 | **R0** | proto generated code 与手写代码 import 路径不一致 | 统一使用 `go.work` 管理 workspace |
| R-P1-003 | **R1** | Memory Repository 实现不完整导致测试失败 | Task 3 同步补全 Mock 实现 |
| R-P1-004 | **R1** | Gateway 路由解析与 Invoker TargetKey 不匹配 | Task 2 严格对照 routes.yaml 映射表 |
| R-P1-005 | **R2** | Permission Service 注入导致大量签名变更 | Task 5 集中处理，一次性改完 |
| R-P1-006 | **R2** | Proto Tester 前端编解码与 Gateway 不兼容 | 已有 CORS 修复；需验证 Protobuf 序列化一致性 |

---

## 8. 工作量估算汇总

| Task | 内容 | 预估时间 | 依赖 |
|------|------|----------|------|
| Task 1 | 编译修复与依赖整理 | 0.5 天 | 无 |
| Task 2 | Gateway ↔ Social 集成 | 1 天 | Task 1 |
| Task 3 | 单元测试修复与补全 | 1-2 天 | Task 1+2 |
| Task 4 | Proto Tester E2E 验证 | 1 天 | Task 2+3+5 |
| Task 5 | 权限校验补充 | 0.5 天 | 可与 Task 3 并行 |
| Task 6 | 文档同步与收尾 | 0.5 天 | Task 4 |
| **合计** | | **4.5-6 天** | |

---

## 9. 待主控确认事项

### R0 — 必须确认方可实施

- [ ] **Phase 1 范围确认**：以"集成联调+测试验证"为主，不新增业务协议？
- [ ] **MVP-P0 白名单确认**：10 对核心协议（§4.4.1）是否准确？有无遗漏/多余？
- [ ] **Memory Repository 策略**：MVP-P0 先用 Memory Repo 跑通 E2E，MySQL/GORM 在 Phase 1.5 接入？
- [ ] **Permission Service 注入时机**：在 Task 5 集中注入还是随各 Task 逐步加入？

### R1 — 建议讨论

- [ ] Task 3 测试覆盖率底线：70% 是否合理？是否提高到 80%？
- [ ] E2E 验证工具：仅用 Proto Tester 手动验证，还是同时编写自动化 E2E 测试？
- [ ] generateUserID 时间戳方案在 MVP-P0 是否可接受？

### R2 — 记录即可

- [ ] fmt.Printf 日志方式在 MVP-P0 保留
- [ ] dispatchProto 泛型抽取推迟到 Phase 1.5 重构
