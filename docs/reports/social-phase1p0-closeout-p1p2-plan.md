# 社交域 Phase 1 P0 收口报告 & Phase 1 P1 三域基础闭环方案

> 文档版本：**v2.0**（主控评审修订版）
> 日期：2026-06-18
> 状态：**待主控最终确认**
> 评审依据：
> - [社交域 MVP 实施方案 v2](./social-mvp-implementation-plan.md)
> - [编码进度汇报 (3)](../tabbit/inbox/2026/06/tabbit_编码进度汇报 (3).md) — 领域事件评审答复
> - [主控评审 — tabbit_phase 2.md](../tabbit/inbox/2026/06/tabbit_phase%202.md) — 本版修订指令
> - [ADR-social-data-level-and-cache-strategy](../adr/ADR-social-data-level-and-cache-strategy.md)

---

## 一、Phase 1 P0 完成状态总览

### 1.1 交付物清单

| 类别 | 数量 | 详情 |
|---|---:|---|
| Go 源文件 (.go，非测试) | **62** | member(15) + group(17) + topic(17) + permission(2) + event(7) + eventhandler(1) + cache(2) + module.go + converter + constants |
| 测试文件 (_test.go) | **38** | member(9) + group(12) + topic(12) + permission(2) + event(2) + eventhandler(1) |
| 测试用例数 | **215** | 全部通过（6 包全绿） |
| Proto 文件 (.proto) | **3** | member.proto / group.proto / topic.proto（与历史 base/*.proto 协议号 100% 一致） |
| SQL 迁移脚本 | **1** | `004_social_domain_tables.sql` — 16 表 DDL + 回滚 |
| 迁移执行器 | **1** | `social_migrate.go` — 已编译验证通过（7.2MB） |
| 路由注册 | **28** | routes.yaml：Member(9) + Group(10) + Topic(9) |

### 1.2 CI 验证矩阵

| 检查项 | 命令 | 结果 |
|---|---|---|
| 静态分析 | `go vet ./...` | PASS（1 个 unreachable code 警告，非阻塞） |
| 编译 | `go build ./...` | PASS |
| 单元测试 | `go test ./... -count=1` | **215 用例 / 6 包全绿** |
| YAML 语法 | `python3 yaml.safe_load(routes.yaml)` | OK（35 条路由 = 7 原有 + 28 社交域） |
| 协议号对齐 | base/*.proto vs proto/social/*.proto vs routes.yaml | **100% 一致**（28 条已注册协议全部三源匹配） |

### 1.3 测试详情

| 包 | 用例数 | 耗时 | 状态 |
|---|---:|---:|---|
| member | ~45 | 4.364s | ok |
| group | ~50 | 1.757s | ok |
| topic | ~55 | 3.199s | ok |
| permission | ~20 | 2.845s | ok |
| event | ~25 | 1.257s | ok |
| eventhandler | ~20 | 2.443s | ok |
| **合计** | **~215** | **~16.7s** | **ALL PASS** |

### 1.4 svc 实现状态（29 个 Handle 方法）

#### Member 域（11 个）

| # | svc 方法 | minType | 实现状态 | 说明 |
|---:|---|---|---|---|
| 1 | SvcRegister.Handle | 1021 | **真实实现** | 参数校验 + repo.CreateUser + 返回 user_id |
| 2 | SvcLogin.Handle | 1023 | **真实实现** | 参数校验 + repo.FindByPhone + **伪 token** ⚠️ |
| 3 | SvcLogout.Handle | 1025 | **Stub** | 直接返回空 Response；**TODO: Token 黑名单** |
| 4 | SvcRefresh.Handle | 1027 | **Stub** | 返回错误 "TODO: Refresh 未实现" |
| 5 | SvcGetUserInfo.Handle | 1029 | **真实实现** | repo.FindByID + model→proto 转换 |
| 6 | SvcUpdateUserInfo.Handle | 1031 | **真实实现** | 参数校验 + repo.Update |
| 7 | SvcBlock.Handle | 1039 | **真实实现** | repo.CreateBlock + 权限检查 |
| 8 | SvcUnblock.Handle | 1041 | **真实实现** | repo.DeleteBlock |
| 9 | SvcGetBlockList.Handle | 1043 | **真实实现** | repo.ListBlocks + 分页 |

#### Group 域（10 个）— **全部真实实现**

| # | svc 方法 | minType | 实现状态 |
|---:|---|---|---|
| 10 | SvcCreate.Handle | 2005 | **真实实现** |
| 11 | SvcJoin.Handle | 2013 | **真实实现** |
| 12 | SvcLeave.Handle | 2015 | **真实实现** |
| 13 | SvcMuteMember.Handle | 2019 | **真实实现** |
| 14 | SvcBanMember.Handle | 2023 | **真实实现** |
| 15 | SvcRemoveMember.Handle | 2027 | **真实实现** |
| 16 | SvcUpdateMemberRole.Handle | 2029 | **真实实现** |
| 17 | SvcRenewMember.Handle | 2037 | **真实实现** |
| 18 | SvcCalcPayableAmount.Handle | 2073 | **真实实现** |
| 19 | SvcEnter.Handle | 2087 | **真实实现** |

#### Topic 域（含 handler case）

| # | svc / case | minType | 实现状态 | 说明 |
|---:|---|---|---|---|
| 20 | SvcCreateTopic.Handle | 3001 | **真实实现** | |
| 21 | SvcListTopic.Handle | 3005 | **真实实现** | |
| 22 | SvcDeleteTopic.Handle | 3009 | **真实实现** | |
| 23 | SvcReplyTopic.Handle | 3043 | **真实实现** | |
| 24 | SvcLikeTopic.Handle | 3061 | **真实实现** | |
| 25 | SvcUnlikeTopic.Handle | 3062 | **Stub** | 复用 LikeTopic 反向逻辑 ⚠️ |
| 26 | SvcFavoriteTopic.Handle | 3063 | **真实实现** | |
| 27 | SvcGetTopicDetail.Handle | 3057 | **真实实现** | |
| 28 | SvcReadTopic.Handle | 3099 | **真实实现** | |
| 29 | SvcUpdateTopic.Handle | 3095 | **Stub** | 复用 CreateTopicRequest 作为 CreateReport 占位 ⚠️ |

### 1.5 实现率统计

| 维度 | 总数 | 已实现(真实) | Stub/占位 | 实现率 |
|---|---:|---:|---:|---:|
| svc Handle 方法 | 29 | 24 | 5 | **82.8%** |
| Member 域 | 11 | 9 | 2 | 81.8% |
| Group 域 | 10 | 10 | 0 | **100%** |
| Topic 域 | 9+1 | 7 | 3 | 70% |
| Handler dispatch case | 30 | 30 | 0 | **100%** |
| 路由注册 | 28 | 28 | 0 | **100%** |

---

## 二、遗留问题清单（33 项 TODO/FIXME/Stub）

### 2.1 按优先级分类

#### P0 — 必须在 P1 修复（影响核心链路）

| ID | 文件 | 内容 | 影响 |
|---|---|---|---|
| T-01 | `member/svc_login.go:121` | JWT 令牌生成为伪实现 | 登录后 token 无法用于鉴权 |
| T-02 | `member/svc_logout.go:32` | 缺少 Token 黑名单 | 登出后 token 仍有效 |
| T-03 | `member/svc_stubs.go:16-27` | RefreshToken 完全未实现 | 无法刷新访问令牌 |
| T-04 | `topic/handler.go:160-166` | CreateReport(3095) 复用 stub | 举报功能不可用 |
| T-05 | `topic/handler.go:135` | 取消收藏(3062) 复用反向逻辑 | 类型不安全 |

#### P1 — P1 应补全（功能缺失）

| ID | 文件 | 内容 | 影响 |
|---|---|---|---|
| T-07 | `permission/service.go:168` | CanManageGroup 权益查询未接入 | 圈主管理权限判断不完整 |
| T-08 | `permission/service.go:259` | 平台管理员角色查询未接入 | 超管权限判断缺数据源 |

#### P2+ — 延后（不在本轮范围）

| ID | 文件 | 内容 | 归属阶段 |
|---|---|---|---|
| T-06 | `event/redis_pubsub.go:8-50` | Redis Pub/Sub 未实现 | **事件基础设施专项**（P1 后） |
| T-09~T-17 | `eventhandler/handler.go:61-104` | StatsHandler 9 个统计更新 TODO | **事件消费者专项** |
| T-18~T-20 | `eventhandler/handler.go:233-239` | NotifyHandler 3 个通知 TODO | **通知系统专项** |
| T-21 | `eventhandler/handler.go:19-21` | StatsHandler 缺 repository 注入 | **事件消费者专项** |

---

## 三、Proto 未覆盖协议（base 定义但未实现）

### Member 域（proto 21 对 → 已实现 9 对 → P1 新增 3 对）

| minType | 协议名 | P1 计划 | 说明 |
|---:|---|---|---|
| 1033 | UpdateMemberStatus | **P1-E4** | 用户状态变更核心能力 |
| 1045 | ListUserGroups | **P1-E5** | 我的群组列表 |
| 其余 9 对 | — | 冻结 | Phase 2+ 按需启动 |

### Group 域（proto 34 对 → 已实现 10 对 → P1 新增 3 对）

| minType | 协议名 | P1 计划 | 说明 |
|---:|---|---|---|
| 2001 | GetGroupInfo | **P1-E1** | 群组详情（基础） |
| 2017 | GetGroupMembers | **P1-E2** | 成员列表分页 |
| 2039 | GetMembership | **P1-E3** | 会员信息查询 |
| 其余 21 对 | — | 冻结 | Phase 2+ 按需启动 |

### Topic 域（proto 21 对 → 已实现 9 对 → P1 新增 1 对）

| minType | 协议名 | P1 计划 | 说明 |
|---:|---|---|---|
| 3025 | GetReplyList | **P1-E6** | 评论列表分页 |
| 其余 11 对 | — | 冻结 | Phase 2+ 按需启动 |

---

## 四、Phase 1 P1 方案 — Social Core 三域基础闭环

### 4.1 目标（主控修订）

```text
完成 user/member、group、topic 三个基础模块的核心协议、核心 Repository、
核心 Service、Handler Dispatch、Gateway Route、单元测试和基础鉴权闭环。
```

### 4.2 范围定义（主控修订）

#### P1 **不做**：

```text
✗ third 域（4000 段）
✗ inbox 域（5000 段）
✗ 缓存层生产部署
✗ Redis Pub/Sub 生产化
✗ 完整事件消费者（StatsHandler/NotifyHandler/CacheHandler）
✗ 通知系统
✗ UI/前端
✗ 收益/钱包/支付渠道/IM 签名
✗ 复杂 CQRS 读模型
```

#### P1 **专注**：

```text
✓ user/member、group、topic 三个基础模块
✓ 认证闭环（JWT + Token 黑名单 + RefreshToken）
✓ 权限服务补全（CanManageGroup/CanManageMember 数据源接入）
✓ Topic Stub 清理（CreateReport/UnfavoriteTopic 独立 svc）
✓ 核心协议补齐（6 个三域最小闭环协议）
✓ 事件接口注入链路（Noop 默认 + Fake 可测 + 6 个发布点）
✓ 所有 Stub/占位从 handler 主链路清除
```

#### 当前阶段追求：

```text
基础业务能跑通    ← 核心协议端到端可调
协议号对齐        ← base/social/routes 三源一致
Stub 明显减少     ← handler 无复用/占位
认证链路可用      ← Login→JWT→API→Logout→Refresh
权限判断可信      ← 基于 1级 MySQL 数据的正确边界
事件接口不阻塞演进 ← Publisher 注入完成，默认 Noop
```

#### 当前阶段**不追求**：

```text
所有社交增强能力完整落地
完整事件驱动统计体系
完整缓存体系
完整消息通知体系
```

### 4.3 任务清单（主控裁剪后）

#### 批次 A：认证闭环（T-01 ~ T-03）— **保留，必须做**

| 任务ID | 任务 | 涉及文件 | 工作量 | 依赖 |
|---|---|---|---:|---|
| P1-A1 | 实现 JWT 令牌生成（替换伪 token） | 新建 `member/jwt.go` + 改 `svc_login.go` | 0.5d | 无 |
| P1-A2 | Token 黑名单接口抽象 + 实现 | 新建 `member/token_store.go` + 改 `svc_logout.go` | 0.5d | P1-A1 |
| P1-A3 | 实现 RefreshToken | 重构 `svc_stubs.go` → `svc_refresh.go` | 0.5d | P1-A1 |
| P1-A4 | 认证相关测试补充 | `jwt_test.go` + `svc_refresh_test.go` + `svc_logout_test.go` | 0.5d | A1~A3 |

**TokenStore 接口设计（主控建议）**：

```go
// TokenStore 定义 token 黑名单存储抽象
// 生产环境使用 Redis 实现；单测/本地使用 Memory 实现
type TokenStore interface {
    // Blacklist 将 token 加入黑名单，ttl 为剩余有效期
    Blacklist(ctx context.Context, token string, ttl time.Duration) error
    // IsBlacklisted 检查 token 是否在黑名单中
    IsBlacklisted(ctx context.Context, token string) (bool, error)
}
```

实现分层：
- `MemoryTokenStore`：单测/本地开发（map + sync.RWMutex）
- `RedisTokenStore`：生产环境（Redis SET + TTL）

**验收标准**：
- Login 返回有效 JWT access_token + refresh_token
- Logout 后原 access_token 被加入黑名单（Memory 或 Redis），中间件可拦截
- RefreshToken 用 refresh_token 换取新的 access_token + refresh_token 对

#### 批次 B：权限服务补全（T-07, T-08）— **保留，必须做**

| 任务ID | 任务 | 涉及文件 | 工作量 | 依赖 |
|---|---|---|---:|---|
| P1-B1 | CanManageGroup 接入 groupRepo.GetEntitlementByGroupAndUser | `permission/service.go` | 0.25d | 无 |
| P1-B2 | CanManageMember 接入平台管理员角色查询 | `permission/service.go` | 0.25d | 无 |
| P1-B3 | 权限边界测试补充（圈主权益/超管权限/普通用户拒绝） | `permission/service_test.go` | 0.25d | B1,B2 |

**验收标准**：
- CanManageGroup 在付费群中正确区分 owner/admin/member 权限边界
- CanManageMember 正确识别超管角色并放行
- 所有权限判断基于 1级 MySQL 数据（Repository），不读 Redis 缓存或 2级统计

#### 批次 C：Topic Stub 修复（T-04, T-05）— **保留，必须做**

| 任务ID | 任务 | 涉及文件 | 工作量 | 依赖 |
|---|---|---|---:|---|
| P1-C1 | 新建独立 SvcCreateReport.Handle（3095） | 新建 `topic/svc_create_report.go` | 0.5d | 无 |
| P1-C2 | 新建独立 SvcUnfavoriteTopic.Handle（3062） | 新建 `topic/svc_unfavorite_topic.go` | 0.25d | 无 |
| P1-C3 | 更新 topic/handler.go 引用新 svc | `topic/handler.go` | 0.1d | C1,C2 |
| P1-C4 | Report/Unfavorite 测试 | `svc_create_report_test.go` + `svc_unfavorite_topic_test.go` | 0.25d | C1,C2 |

**注意**：CreateReport 当前阶段只做到「举报记录落库 + 返回 report_id」，不扩展到审核流和通知。

**验收标准**：
- CreateReport 创建举报记录并返回 report_id（幂等）
- UnfavoriteTopic 删除收藏记录（幂等，不存在不报错）
- handler 中不再有 "stub"、"复用"、"占位" 类注释

#### 批次 D：事件注入链路 — **降级处理（主控裁定）**

> 原计划包含 Redis Pub/Sub 生产实现（P1-D6）。主控裁定：当前阶段只做接口注入和发布点，
> Redis Pub/Sub 推迟到「事件基础设施专项」，不阻塞三域基础闭环。

| 任务ID | 调整后任务 | 是否 P1 执行 | 工作量 |
|---|---|---|---:|---|
| P1-D1 | ModuleOptions 结构体 + Option 模式 | **执行** | 0.25d |
| P1-D2 | Servant 构造函数注入 event.Publisher | **执行** | 0.25d |
| P1-D3 | Handler 透传 publisher 到 svc | **执行** | 0.25d |
| P1-D4 | 6 个核心事件发布点（事务成功后 Publish） | **执行** | 1d |
| P1-D5 | 事件发布失败不阻塞主业务测试 + FakePublisher 断言 | **执行** | 0.5d |
| ~~P1-D6~~ | ~~Redis Pub/Sub 基础实现~~ | **暂缓 → 事件专项** | — |

**首批接入事件的 6 个操作**：

| 事件常量 | 触发 svc | Payload | 默认行为（Noop） |
|---|---|---|---|
| EventMemberRegistered | SvcRegister | UserRegisteredPayload | 无操作 |
| EventUserStatusChanged | SvcUpdateStatus (P1-E4) | UserStatusChangedPayload | 无操作 |
| EventGroupCreated | SvcCreate | GroupCreatedPayload | 无操作 |
| EventGroupMemberJoined | SvcJoin | GroupMemberChangedPayload | 无操作 |
| EventGroupMemberLeft | SvcLeave | GroupMemberChangedPayload | 无操作 |
| EventTopicCreated | SvcCreateTopic | TopicCreatedPayload | 无操作 |

**验收标准（主控修订）**：

```text
✓ 6 个核心事件可通过 event.Publisher 接口发布
✓ 默认 NoopPublisher 不影响业务响应
✓ 测试中可通过 FakePublisher/MemoryBus 断言事件类型和 payload
✓ 事件发布失败不阻塞主业务（best-effort 模式）
✗ Redis Pub/Sub 不作为当前阶段强制验收项
```

#### 批次 E：核心协议补齐 — **保留但确认 6 个**

| 任务ID | 任务 | minType | 涉及文件 | 工作量 |
|---|---|---|---|---:|---|
| P1-E1 | GetGroupInfo（群组详情） | 2001 | 新建 `group/svc_get_group_info.go` + handler + route | 0.5d |
| P1-E2 | GetGroupMembers（成员列表分页） | 2017 | 新建 `group/svc_get_members.go` + handler + route | 0.5d |
| P1-E3 | GetMembership（会员信息查询） | 2039 | 新建 `group/svc_get_membership.go` + handler + route | 0.5d |
| P1-E4 | UpdateMemberStatus（用户状态变更） | 1033 | 新建 `member/svc_update_status.go` + handler + route | 0.5d |
| P1-E5 | ListUserGroups（我的群组列表） | 1045 | 新建 `member/svc_list_groups.go` + handler + route | 0.5d |
| P1-E6 | GetReplyList（评论列表分页） | 3025 | 新建 `topic/svc_get_reply_list.go` + handler + route | 0.5d |
| P1-E7 | 补充协议的 routes.yaml 注册 + 测试 | `routes.yaml` + 6 个 test 文件 | 0.5d | E1~E6 |

**UpdateMemberStatus 定位（主控明确）**：当前阶段至少做到：
- 状态更新落库
- 发布 UserStatusChanged 事件（默认 Noop）
- 清理鉴权/用户资料缓存的接口预留
- 权限检查最小可用
- 测试覆盖 active/banned/deleted 状态流转

**不扩展到**：完整风控、通知推送、后台操作台。

**验收标准**：
- 6 条新协议可通过 Gateway → Handler → Svc → Repo 全链路调用
- routes.yaml 注册数达到 **34 条**
- 测试用例总数达到 **260+**

### 4.4 P1 时间线（主控压缩版）

```
Day 1:    批次 A（认证闭环）
├─ P1-A1 JWT 令牌生成
├─ P1-A2 TokenStore 接口 + Memory/Redis 双实现
├─ P1-A3 RefreshToken
└─ P1-A4 认证测试

Day 2:    批次 B（权限补全）+ 批次 C（Stub 修复）
├─ P1-B1 CanManageGroup 数据源接入
├─ P1-B2 CanManageMember 超管角色接入
├─ P1-B3 权限边界测试
├─ P1-C1 CreateReport 独立 svc
├─ P1-C2 UnfavoriteTopic 独立 svc
└─ P1-C3/C4 handler 更新 + 测试

Day 3:    批次 E-1~E3（Group 补充协议）
├─ P1-E1 GetGroupInfo (2001)
├─ P1-E2 GetGroupMembers (2017)
├─ P1-E3 GetMembership (2039)
├─ routes.yaml 注册
└─ Group 补充协议测试

Day 4:    批次 E-4~E6（Member/Topic 补充协议）
├─ P1-E4 UpdateMemberStatus (1033)
├─ P1-E5 ListUserGroups (1045)
├─ P1-E6 GetReplyList (3025)
├─ routes.yaml 注册
└─ Member/Topic 补充协议测试

Day 5:    批次 D（事件注入链路）+ 全量收口
├─ P1-D1 ModuleOptions + Option 模式
├─ P1-D2/D3 Servant/Handler/Publisher 注入
├─ P1-D4 6 个核心事件发布点
├─ P1-D5 事件发布测试（FakePublisher 断言）
├─ go build / go test / go vet 全量 CI
└─ 文档同步 + 收口报告
```

**预估工期：5 个工作日**（可扩展至 6-7 天如遇阻塞）

> 不因 Redis Pub/Sub、eventhandler 统计更新、third/inbox 拉长当前阶段。

### 4.5 P1 验收标准（主控修订版）

| # | 检查项 | 通过标准 |
|---:|---|---|
| 1 | 编译 | `go build ./...` PASS |
| 2 | 测试 | `go test ./... -count=1` PASS，>= **260** 用例 |
| 3 | 静态检查 | `go vet ./...` 无新增阻塞项 |
| 4 | 协议对齐 | base proto / social proto / routes.yaml **三源一致** |
| 5 | 认证链路 | Login → JWT → API 调用 → Logout(黑名单) → RefreshToken 流程可测 |
| 6 | 权限链路 | CanManageGroup / CanManageMember / CanReadTopic 核心边界可测 |
| 7 | User 基础协议 | Register/Login/Get/Update/**Status**/ListGroups 全部可用 |
| 8 | Group 基础协议 | Create/Join/Leave/**GetInfo**/**GetMembers**/**GetMembership** 全部可用 |
| 9 | Topic 基础协议 | Create/List/Detail/Reply/**GetReplyList**/Like/Favorite/**Unfavorite**/**Report** 全部可用 |
| 10 | 事件链路 | event.Publisher 注入完成，Noop 默认，Fake/MemoryBus 可测 |
| 11 | Stub 清理 | 三域 handler 主链路不再有 stub/复用/占位 |
| 12 | third/inbox | 不创建、不排期、不纳入验收 |

---

## 五、冻结项池（Third / Inbox 挂起）

> 主控裁定：third 和 inbox 从 Phase 2 主计划中移出，改为冻结项。
> 等待外部依赖确认后单独立项启动。

### 冻结项 F1：Third 域（4000 段）

```text
状态：挂起，不进入当前 P1，不进入当前 Phase 2 主计划
启动条件：等待支付渠道、OSS、OAuth、分享链路产品需求确认
涉及能力：
  - OSS 上传配置/签名
  - 分享链接生成
  - OAuth 第三方登录
  - 钱包支付对接
proto 文件：暂不创建 proto/social/third.proto
```

### 冻结项 F2：Inbox 域（5000 段）

```text
状态：挂起，不进入当前 P1，不进入当前 Phase 2 主计划
启动条件：等待通知系统、站内信模型、IM 签名能力需求确认
涉及能力：
  - 消息查询/分页
  - 已读标记
  - 系统通知
  - IM 签名
proto 文件：暂不创建 proto/social/inbox.proto
```

---

## 六、后续专项（P1 交付后的自然延伸）

以下专项不在 P1 范围内，但作为路线图记录，P1 完成后按优先级依次启动：

### 专项 S1：事件基础设施生产化

| 子任务 | 内容 | 前置条件 |
|---|---|---|
| S1-1 | Redis Pub/Sub Publisher/Subscriber 生产实现 | P1-D 完成（接口已就位） |
| S1-2 | StatsHandler 统计更新实现（T-09~T-17） | S1-1 |
| S1-3 | CacheHandler 缓存失效实现 | S1-1 |
| S1-4 | NotifyHandler 通知推送对接（T-18~T-20） | 通知系统 API 就绪 |
| S1-5 | 事件幂等性保障（CAS rowsAffected） | S1-1 |
| S1-6 | 事件可观测性（结构化日志 + metrics） | 监控系统就绪 |

### 专项 S2：剩余协议补全（三域内）

Member 域：GetMemberProfile(1035)、SearchUser(1037)、CheckUserActions(1047)、GetUserStats(1049) 等 9 对

Group 域：UpdateGroupInfo(2007)、DeleteGroup(2009)、TransferGroupOwner(2011)、KickMember(2025)、GroupPlan CRUD(2031/2033/2035) 等 21 对

Topic 域：UpdateTopic(3011)、PinTopic(3013/3015)、DeleteReply(3031)、PinComment(3033)、GetReactions/Favorites(3071/3075)、AuditLog(3085/3087)、Stats(3091/3093) 等 11 对

### 专项 S3：缓存层生产部署

Cache Aside 策略落地、Redis Key 规范执行、缓存命中率监控。

---

## 七、架构决策记录（本次收口确认）

### ADR-SOCIAL-001：协议号权威源

**决策**：以 `docs/tabbit/inbox/2026/06/protocols/base/*.proto` 为协议号唯一权威来源。`proto/social/*.proto` 和 `configs/gateway/routes.yaml` 必须与其保持 100% 一致。

### ADR-SOCIAL-002：3095 号冲突处置

**决策**：3095/3096 归属 **CreateReport（举报）**。UpdateTopic 需在后续申请新协议号（建议 3011 或新分配）。

### ADR-SOCIAL-003：事件基础设施选型

**决策**：「接口隔离 + Redis 生产默认 + MemoryBus 测试 + Noop 兜底」四层模型。当前 P1 只完成接口注入和发布点，Redis 实现推迟到专项 S1。

### ADR-SOCIAL-004：TokenStore 接口隔离（新增，主控建议）

**决策**：Token 黑名单通过 `TokenStore` 接口抽象，支持 Memory（测试）和 Redis（生产）双实现。避免认证模块强依赖 Redis 运行环境。

**理由**：即使 Redis 环境尚未确认，也不影响 P1 基础模块测试闭环。

---

## 八、风险与缓解措施（主控修订版）

| 风险ID | 等级 | 描述 | 缓解措施 | 处理阶段 |
|---|---|---|---|---|
| R-P1-001 | **R1** | JWT 库选型未确定 | P1-A1 前由主控确认（golang-jwt/jwt / crypto/jwt / 其他） | P1 Day 1 |
| R-P1-002A | **R1** | 认证黑名单依赖 Redis | **TokenStore 接口隔离**：Memory 实现用于测试，本地/生产使用 Redis | P1 批次 A |
| R-P1-002B | **R2** | 事件 Pub/Sub 生产实现暂缓 | 当前只完成 Publisher 接口注入和 Fake 测试；RedisPublisher 进入**专项 S1** | 专项 S1 |
| R-P1-003 | **R2** | 事件注入链路改动面大（module→servant→handler→svc） | Option 模式渐进注入，默认 NoopPublisher 兼容旧调用方 | P1 批次 D |
| R-P1-004 | **R2** | UpdateTopic 无可用协议号 | 专项 S2 中申请新编号；当前通过 Create + Delete 组合模拟编辑 | 专项 S2 |
| R-P1-005 | **R3** | eventhandler unreachable code vet 警告 | 非阻塞，专项 S1 重构时清理 dead code | 专项 S1 |
| R-P1-006 | **R2** | FakePublisher _test.go 跨包不可见 | 如需跨包则迁移至非 test 文件；当前 nil 即满足 P1 需求 | P1 或 S1 |

---

## 九、需要项目主控最终确认的问题

### R0 — 必须确认方可进入 P1 执行

- [ ] **P1 范围确认**：批次 A/B/C/D/E 的任务划分是否合理？（已按主控意见裁剪）
- [ ] **JWT 库选型**：P1-A1 使用哪个 JWT 库？
- [ ] **TokenStore Redis 环境**：dev/staging/prod 是否均已部署 Redis？DB 编号约定？
- [ ] **UpdateMemberStatus(1033) 范围确认**：只做状态落库+事件发布+基础权限，不扩展风控/通知？
- [ ] **时间线确认**：5 个工作日是否合理？是否需要预留缓冲？

### R1 — 已由主控在评审中决策（记录备查）

- [x] **third/inbox 挂起** — 已确认：不创建、不排期、不纳入验收
- [x] **事件系统降级** — 已确认：P1 只做接口注入+Noop+Fake+发布点；Redis Pub/Sub 进专项 S1
- [x] **验收重点调整** — 已确认：从「事件基础设施生产化」调整为「三域基础业务可用」
- [x] **TokenStore 接口隔离** — 已确认：Memory/Redis 双实现

### R2 — 记录即可

- [ ] FakePublisher 跨包可见性（P1 或 S1 处理均可）
- [ ] eventhandler unreachable code 清理时机（S1 处理）

---

## 十、附录

### A. 文件索引

| 文档 | 路径 | 用途 |
|---|---|---|
| **本文档（v2.0）** | `docs/reports/social-phase1p0-closeout-p1p2-plan.md` | P0 收口 + P1 三域闭环方案（主控修订版） |
| MVP 实施方案 v2 | `docs/reports/social-mvp-implementation-plan.md` | 12 步开发计划（权威） |
| 编码进度汇报 (3) | `docs/tabbit/inbox/2026/06/tabbit_编码进度汇报 (3).md` | 领域事件评审答复 |
| **主控评审** | `docs/tabbit/inbox/2026/06/tabbit_phase 2.md` | P1/P2 范围裁剪指令（本版依据） |
| 主控 DevGuide | `docs/tabbit/inbox/.../social-service-dev-guide.md` | 代码模板/规范（权威） |
| ADR 数据分级 | `docs/adr/ADR-social-data-level-and-cache-strategy.md` | 缓存策略 |
| 协议基准 | `docs/tabbit/inbox/2026/06/protocols/base/*.proto` | 协议号权威源 |
| 社交域 Proto | `proto/social/*.proto` | 工程副本（与基准一致） |
| 路由注册 | `configs/gateway/routes.yaml` | Gateway 分发配置 |

### B. P0 交付物速查表

```
go/modules/social/
├── module.go                          ✅ 3 Servant 聚合
├── converter/converter.go             ✅ 枚举转换集中
├── permission/
│   ├── service.go                     ✅ 8 个方法
│   ├── constants.go                   ✅ 域常量
│   ├── mock_repository_test.go        ✅ Mock
│   └── service_test.go                ✅ 28 用例
├── member/                            ✅ 15 源 + 9 测试
│   ├── servant.go / handler.go        ✅ 11 case dispatch
│   ├── repository.go / _gorm.go       ✅ GORM
│   ├── model.go                       ✅ 3 Model
│   ├── constants.go                   ✅ 包级常量
│   └── svc_{register,login,...}       ✅ 9 svc (2 stub → P1 修复)
├── group/                             ✅ 17 源 + 12 测试
│   ├── servant.go / handler.go        ✅ 10 case dispatch
│   ├── repository.go / _gorm.go       ✅ GORM
│   ├── model.go                       ✅ 5 Model
│   └── svc_{create,join,leave,...}    ✅ 10 svc (全真实)
├── topic/                             ✅ 17 源 + 12 测试
│   ├── servant.go / handler.go        ✅ 11 case dispatch
│   ├── repository.go / _gorm.go       ✅ GORM
│   ├── model.go                       ✅ 7 Model
│   └── svc_{create_topic,list,...}    ✅ 9+ svc (3 stub → P1 修复)
├── event/                             ✅ 7 源 + 2 测试
│   ├── constants.go / event.go        ✅ 事件类型定义
│   ├── payload.go                     ✅ 强类型 payload
│   ├── publisher.go / subscriber.go   ✅ 接口定义
│   ├── noop.go                        ✅ Noop 实现
│   ├── memory_bus.go                  ✅ 内存实现（测试用）
│   └── redis_pubsub.go                ⚠️ TODO → 专项 S1
├── eventhandler/handler.go            ✅ 1 源 + 1 测试 (17 TODO → 专项 S1)
└── cache/
    ├── keys.go                        ✅ 缓存 Key 定义
    └── invalidator.go                 ✅ CacheInvalidator 接口

proto/social/
├── member.proto                       ✅ 21 对 Request/Response
├── group.proto                        ✅ 34 对 Request/Response
└── topic.proto                        ✅ 21 对 Request/Response

scripts/migration/
├── 004_social_domain_tables.sql       ✅ 16 表 DDL
├── social_migrate.go                  ✅ 迁移执行器（已编译）

configs/gateway/routes.yaml            ✅ 35 条路由 (7 原有 + 28 社交域)
```

### C. 协议号覆盖矩阵（三域聚焦版）

| 域 | Proto 定义 | P0 已实现 | P1 新增 | P1 后合计 | 覆盖率 |
|---|---:|---:|---:|---:|---:|
| Member (1000) | 21 | 9 | **+3** | 12 | 57% |
| Group (2000) | 34 | 10 | **+3** | 13 | 38% |
| Topic (3000) | 21 | 9 | **+1** | 10 | 48% |
| **三域合计** | **76** | **28** | **+7** | **35** | **46%** |
| Third (4000) | ~28 | 0 | 0 | 0 | 冻结 |
| Inbox (5000) | ~9 | 0 | 0 | 0 | 冻结 |

> 注：覆盖率以「已实现协议数 / proto 定义总数」计算。P1 聚焦三域最小闭环（35/76 = 46%），
> 剩余 41 对协议在专项 S2 中按产品优先级逐步补全。
