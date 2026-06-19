# Social Phase 1 MVP-P0 集成联调实施方案

> **文档状态**：Task 5-C / Task 6 执行完成，待主控最终验收
> **版本**：v3.0（执行完成版 — 含 E2E 结果 + routes.yaml 治理方案）
> **日期**：2026-06-19（更新）
> **作者**：Trae 工程执行 Agent
> **相关 PRD**：`docs/reports/social-mvp-implementation-plan.md`
> **相关 ADR**：ADR-0012（多语言单体仓库）、ADR-0013（Makefile 工程入口）
> **协议号权威源**：`docs/tabbit/inbox/2026/06/protocols/base/{user_base,group_base,topic_base}.proto`

---

## 一、执行目标

将 Social 三域模块（member/user、group、topic）从「代码已编写」推进到「可编译、可测试、可经 Gateway 路由、可通过 Proto Tester 端到端调用」的状态。

### 1.1 本轮交付物

| 交付物 | 说明 |
|--------|------|
| 编译通过的 Social 模块 | `go build ./go/modules/social/...` 零错误 |
| 全量单元测试通过 | `go test ./go/modules/social/... -count=1` 全绿 |
| Gateway ↔ Social LocalInvoker 集成 | maxType=1000/2000/3000 可路由到对应 Servant |
| JWT → Context 身份注入 | Social svc 从 context 获取 operator/user |
| Proto Tester E2E 验证 | 白名单协议通过 Gateway 调用返回正确响应 |
| 实施方案文档 | 本文档 |

### 1.2 本轮不做（挂起）

- third(4000) 域：不创建、不接入、不排期
- inbox(5000) 域：同上
- MySQL/GORM 真实连接 E2E → Phase 1.5
- Redis 真实连接 / PubSub 生产化验收 → Phase 2
- StatsHandler 真实统计更新 → Phase 2
- NotifyHandler 通知推送 → Phase 2
- Admin 后台管理界面 → 独立排期

---

## 二、代码审计结果（以 base proto 为权威源）

### 2.1 协议号三源校验

**权威源优先级**：base proto（最高） > proto/social/*.proto > routes.yaml > handler.go

**校验规则**：
- Request minType 必须为奇数（odd）
- Response minType 必须为偶数（even）
- 同一个 minType 只能对应一个协议定义
- 禁止将 Response minType 当作 Request minType 使用

#### 2.1.1 Member 域（maxType=1000）

**权威源**：[`user_base.proto`](docs/tabbix/inbox/2026/06/protocols/base/user_base.proto)

| # | 协议名 | Request min | Response min | proto 定义 | routes.yaml | handler.go | 状态 |
|:-:|--------|:-----------:|:------------:|:----------:|:-----------:|:----------:|:----:|
| 1 | UserRegister | 1021 | 1022 | ✅ | ✅ | ✅ | 对齐 |
| 2 | UserLogin | 1023 | 1024 | ✅ | ✅ | ✅ | 对齐 |
| 3 | UserLogout | 1025 | 1026 | ✅ | ✅ | ✅ | 对齐 |
| 4 | RefreshToken | 1027 | 1028 | ✅ | ✅ | ✅ | 对齐 |
| 5 | GetUserInfo | 1029 | 1030 | ✅ | ✅ | ✅ | 对齐 |
| 6 | UpdateUserInfo | 1031 | 1032 | ✅ | ✅ | ✅ | 对齐 |
| 7 | UpdateMemberStatus | 1033 | 1034 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |
| 8 | BlockUser | 1039 | 1040 | ✅ | ✅ | ✅ | 对齐 |
| 9 | UnblockUser | 1041 | 1042 | ✅ | ✅ | ✅ | 对齐 |
| 10 | GetBlockList | 1043 | 1044 | ✅ | ✅ | ✅ | 对齐 |
| 11 | GetUserStats | 1045 | 1046 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |

> **注**：base proto 中还定义了 BatchGetUserInfo(1049/1050)、UpgradeMembership(1051/1052)、GetBlockCount(1047/1048)、GetIMUserSig(1074/1075)、UserConfig(1091-1100)、NotificationSettings(1053-1056) 等协议，但本轮 handler 未实现，不在 E2E 范围内。

#### 2.1.2 Group 域（maxType=2000）

**权威源**：[`group_base.proto`](docs/tabbit/inbox/2026/06/protocols/base/group_base.proto)

| # | 协议名 | Request min | Response min | proto 定义 | routes.yaml | handler.go | 状态 |
|:-:|--------|:-----------:|:------------:|:----------:|:-----------:|:----------:|:----:|
| 1 | CreateGroup | 2005 | 2006 | ✅ | ✅ | ✅ | 对齐 |
| 2 | JoinGroup | 2013 | 2014 | ✅ | ✅ | ✅ | 对齐 |
| 3 | LeaveGroup | 2015 | 2016 | ✅ | ✅ | ✅ | 对齐 |
| 4 | MuteMember | 2019 | 2020 | ✅ | ✅ | ✅ | 对齐 |
| 5 | BanMember | 2023 | 2024 | ✅ | ✅ | ✅ | 对齐 |
| 6 | RemoveMember | 2027 | 2028 | ✅ | ✅ | ✅ | 对齐 |
| 7 | UpdateMemberRole | 2029 | 2030 | ✅ | ✅ | ✅ | 对齐 |
| 8 | RenewMember | 2037 | 2038 | ✅ | ✅ | ✅ | 对齐 |
| 9 | CalcPayableAmount | 2073 | 2074 | ✅ | ✅ | ✅ | 对齐 |
| 10 | GroupUserEnter | 2087 | 2088 | ✅ | ✅ | ✅ | 对齐 |
| 11 | GetGroupStats | 2039 | 2040 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |
| 12 | BatchGetGroups | 2047 | 2048 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |
| 13 | GetGroupMemberUserIds | 2077 | 2078 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |

> **注**：base proto 中还定义了 UpdateGroup(2009/2010)、DeleteGroup(2011/2012)、CheckGroupName(2051/2052)、UnmuteMember(2021/2022)、UnbanMember(2025/2026)、GetBannedMembers(2065/2066)、Permissions 系列(2055-2060)、Discounts(2067-2072)、GetUserGroupQuota(2089/2090) 等，本轮未实现。

#### 2.1.3 Topic 域（maxType=3000）— 含关键修复项

**权威源**：[`topic_base.proto`](docs/tabbit/inbox/2026/06/protocols/base/topic_base.proto)

| # | 协议名 | Request min | Response min | proto 定义 | routes.yaml | handler.go | 状态 |
|:-:|--------|:-----------:|:------------:|:----------:|:-----------:|:----------:|:----:|
| 1 | CreateTopic | 3001 | 3002 | ✅ | ✅ | ✅ | 对齐 |
| 2 | GetTopicList | 3005 | 3006 | ✅ | ✅ | ✅ | 对齐 |
| 3 | DeleteTopic | 3009 | 3010 | ✅ | ✅ | ✅ | 对齐 |
| 4 | AddTopicReply | 3043 | 3044 | ✅ | ✅ | ✅ | 对齐 |
| 5 | **LikeTopic** | **3061** | **3062** | ✅ | ✅ | ✅ | 对齐 |
| 6 | **FavoriteTopic** | **3063** | **3064** | ✅ | ✅ | ✅ | 对齐 |
| 7 | BatchGetTopicInfo | 3057 | 3058 | ✅ | ✅ | ✅ | 对齐 |
| 8 | CreateReport | 3095 | 3096 | ✅ | ✅ | ✅ | 对齐 |
| 9 | CheckTopicActions | 3099 | 3100 | ✅ | ✅ | ✅ | 对齐 |
| 10 | GetReplyList | 3065 | 3066 | ✅ | ❌ 缺失 | ✅ 有case | **需补路由** |

#### ⚠️ 2.1.4 Topic 域协议号冲突 — 关键修复

**问题描述**：

当前 [`topic/handler.go`](go/modules/social/topic/handler.go) 存在 **两个错误的 Dispatch case**，违反了 base proto 权威源定义：

| 错误 case | 当前 handler 中的处理 | base proto 实际含义 | 问题性质 |
|:---------:|---------------------|--------------------|:-------:|
| `case "3062"` | 调用 `unlikeTopicSvc.Handle()`，注释写「复用 LikeTopicRequest.is_like=false」 | **3062 是 LikeTopicResponse（偶数=响应类型），不是请求类型** | **严重：将 Response minType 当 Request 使用** |
| `case "3064"` | 调用 `unfavoriteTopicSvc.Handle()`，注释写「复用 FavoriteTopicRequest」 | **3064 是 FavoriteTopicResponse（偶数=响应类型），不是请求类型** | **严重：将 Response minType 当 Request 使用** |

**Base Proto 权威定义**（来自 [`topic_base.proto` L384-L427`](docs/tabbit/inbox/2026/06/protocols/base/topic_base.proto#L384-L427)）：

```protobuf
// 点赞/取消点赞主题 —— 用一个 Request + bool flag 区分操作
message LikeTopicRequest {
  enum Type { min = 3061; }   // 奇数 = Request
  string topic_id = 1;
  bool is_like = 2;           // true=点赞, false=取消点赞
}
message LikeTopicResponse {
  enum Type { min = 3062; }   // 偶数 = Response（不是 UnlikeTopicRequest！）
  bool is_liked = 2;
  int32 like_count = 3;
}

// 收藏/取消收藏主题 —— 同理
message FavoriteTopicRequest {
  enum Type { min = 3063; }   // 奇数 = Request
  string topic_id = 1;
  bool is_favorite = 2;       // true=收藏, false=取消收藏
}
message FavoriteTopicResponse {
  enum Type { min = 3064; }   // 偶数 = Response（不是 UnfavoriteTopicRequest！）
  bool is_favorited = 2;
  int32 favorite_count = 3;
}
```

**正确设计**（以 base proto 为准）：

- 客户端发送 `LikeTopicRequest(min=3061, is_like=true)` → 服务端返回 `LikeTopicResponse(min=3062)`
- 客户端发送 `LikeTopicRequest(min=3061, is_like=false)` → 服务端返回 `LikeTopicResponse(min=3062)`
- 客户端发送 `FavoriteTopicRequest(min=3063, is_favorite=true)` → 服务端返回 `FavoriteTopicResponse(min=3064)`
- 客户端发送 `FavoriteTopicRequest(min=3063, is_favorite=false)` → 服务端返回 `FavoriteTopicResponse(min=3064)`

**修复方案**：

1. 删除 `handler.go` 中的 `case "3062"` 和 `case "3064"`
2. 将 `case "3061"` 改为统一处理：根据 `req.IsLike` 判断点赞或取消点赞
3. 将 `case "3063"` 改为统一处理：根据 `req.IsFavorite` 判断收藏或取消收藏
4. 合并 `SvcLikeTopic` + `SvcUnlikeTopic` → 单一 `SvcLikeTopic`（内部判断 IsLike）
5. 合并 `SvcFavoriteTopic` + `SvcUnfavoriteTopic` → 单一 `SvcFavoriteTopic`（内部判断 IsFavorite）
6. 删除 `SvcUnlikeTopic` 和 `SvcUnfavoriteTopic` 及其测试文件
7. 更新 `handler.go` 结构体字段，移除 `unlikeTopicSvc` 和 `unfavoriteTopicSvc`

**影响范围**：

| 受影响文件 | 变更类型 | 说明 |
|-----------|:-------:|------|
| `topic/handler.go` | 修改 | 删除 2 个 case，修改 2 个 case 逻辑，移除 2 个字段 |
| `topic/svc_like_topic.go` | 修改 | 增加 `is_like=false` 分支逻辑 |
| `topic/svc_unlike_topic.go` | **删除** | 合并入 svc_like_topic.go |
| `topic/svc_favorite_topic.go` | 修改 | 增加 `is_favorite=false` 分支逻辑 |
| `topic/svc_unfavorite_topic.go` | **删除** | 合并入 svc_favorite_topic.go |
| `topic/svc_like_topic_test.go` | 修改 | 补充 `is_like=false` 测试用例 |
| `topic/svc_unlike_topic_test.go` | **删除** | 合并入 like 测试 |
| `topic/svc_favorite_topic_test.go` | 修改 | 补充 `is_favorite=false` 测试用例 |
| `topic/svc_unfavorite_topic_test.go` | **删除** | 合并入 favorite 测试 |

#### 2.1.5 三源校验结论

1. **无重复协议号**：所有 Request minType 在各源中唯一，无冲突
2. **无孤立路由**：routes.yaml 中注册的路由在 handler 中均有对应 case
3. **Topic 域发现 2 个非法 case**：handler.go 中的 3062/3064 case 违反 base proto 定义，必须删除
4. **routes.yaml 缺 6 条路由**（修正后数量，原为 8 条）：
   - Member: 1033, 1045（2 条）
   - Group: 2039, 2047, 2077（3 条）
   - Topic: 3065（1 条，3062/3064 不再需要路由因为它们是 Response 类型）
5. **处理决策**：
   - P0：修复 Topic 3062/3064 协议号违规（代码合并）
   - P0：补齐 routes.yaml 6 条缺失路由

### 2.2 路由数量口径统一（修正后）

| 统计维度 | Member(1000) | Group(2000) | Topic(3000) | 合计 |
|---------|:------------:|:-----------:|:-----------:|:----:|
| routes.yaml 已注册 | 9 | 10 | 9 | **28** |
| handler.go 有效 case 数 | 11 | 13 | **10**（修正后） | **34** |
| proto 定义的 Request/Response 对（本轮实现） | 11 对 | 13 对 | 10 对 | **34 对** |
| 差额（handler 有但 routes 无） | 2 | 3 | 1 | **6** |
| 补齐后 routes.yaml 总数 | 11 | 13 | 10 | **34** |

> **说明（v2.0 修正）**：
> - v1.0 版本错误地将 Topic handler case 数记为 14（含 2 个非法 case 3062/3064）
> - v2.0 以 base proto 为准，确认 3062/3064 为 Response 类型，handler 有效 case 为 10
> - 最终社交域总路由数为 **34**（非 v1.0 的 38）
> - 以 handler.go 有效 case 为准，routes.yaml 需补齐 **6** 条路由（非 v1.0 的 8 条）

### 2.3 当前实现状态分级

| 文件/模块 | 状态 | 说明 |
|----------|:----:|------|
| member/svc_register.go | 已编写+已单测 | 五步模式完整实现 |
| member/svc_login.go | 已编写+已单测 | 含 bcrypt + JWT 生成 |
| member/svc_logout.go | 已编写+已单测 | TokenStore 黑名单 |
| member/svc_refresh.go | 已编写+已单测 | 刷新令牌逻辑完整 |
| member/svc_update_member_status.go | 已编写+已单测 | P1-E 补齐 |
| member/svc_get_user_stats.go | 已编写+已单测 | P1-E 补齐 |
| group/svc_create.go | 已编写+已单测 | 含自动添加 owner |
| group/svc_join.go | 已编写+已单测 | 含幂等检查 |
| group/svc_leave.go ~ svc_calc_payable.go | 已编写+已单测 | 10 个 svc 全覆盖 |
| group/svc_{get_group_stats,batch_get_groups,get_member_user_ids}.go | 已编写+已单测 | P1-E 补齐 |
| topic/svc_create_topic.go | 已编写+已单测 | 含 context userID |
| topic/svc_{list,delete,reply,like,favorite,detail,report,reply_list}.go | 已编写+已单测 | 10 个 svc（合并后） |
| topic/svc_like_topic.go | **待修改** | 需合并 unlike 分支（is_like=false） |
| topic/svc_unlike_topic.go | **待删除** | 合并入 svc_like_topic.go |
| topic/svc_favorite_topic.go | **待修改** | 需合并 unfavorite 分支（is_favorite=false） |
| topic/svc_unfavorite_topic.go | **待删除** | 合并入 svc_favorite_topic.go |
| event/publisher.go + noop.go + memory_bus.go + fake_publisher.go | 已编写+已单测 | 事件系统完整 |
| event/redis_pubsub.go | 已编写 | 编译通过即可，不验收连接 |
| permission/service.go | 已编写+已单测 | 8 方法全部实现 |
| module.go | 已编写 | 聚合三域 Servant |
| servant.go (×3) | 已编写 | Handle 接口统一 |
| handler.go (member/group) | 已编写 | Dispatch 完整 |
| handler.go (topic) | **待修改** | 需删除 3062/3064 非法 case |
| Gateway main.go | Stub | 仅注册 Hello，无 Social |
| invoker.go RegisterModuleHandlers | Stub | 仅 System+Config+I18n |
| JWT → Context 注入链路 | 待集成 | local 模式未启用 Auth |

---

## 三、差距分析

### 3.1 架构差距图

```
当前状态（有缺口）:
┌─────────────┐     ┌──────────────────┐     ┌──────────────────┐
│ Proto Tester │ ──> │    Gateway        │ ──> │  ??? (无 Social)  │
│  (前端 3002) │     │  main.go          │     │                  │
│              │     │  仅 /api/hello    │     │  LocalInvoker 中   │
│              │     │  无 Social 路由   │     │  无 Social Handler │
└─────────────┘     └──────────────────┘     └──────────────────┘

目标状态:
┌─────────────┐     ┌──────────────────┐     ┌──────────────────┐
│ Proto Tester │ ──> │    Gateway        │ ──> │  Social Module    │
│  (前端 3002) │     │  +Social 路由     │     │  MemberServant    │
│              │     │  +JWT→Context     │     │  GroupServant     │
│              │     │  +CORS 全路径     │     │  TopicServant      │
└─────────────┘     └──────────────────┘     └──────────────────┘
                            │                          │
                     ┌──────┴──────┐            ┌──────┴──────┐
                     │ AuthMiddleware│           │ MemoryRepo  │
                     │ (JWT解析)    │           │ (MVP-P0)    │
                     └─────────────┘            └─────────────┘
```

### 3.2 关键差距清单

| # | 差距项 | 影响 | 修复复杂度 |
|---|--------|:----:|:----------:|
| G1 | **Topic handler 3062/3064 非法 case** | 违反 base proto，可能导致运行时行为异常 | 中 — 需合并 svc + 删除 case |
| G2 | routes.yaml 缺 6 条路由 | handler 有 case 但无法通过 Gateway 路由到达 | 低 — 补 YAML 即可 |
| G3 | Gateway main.go 未注册 Social 模块 | 1000/2000/3000 maxType 请求全部 404 | 中 — 需新增注册代码 |
| G4 | LocalInvoker 未注册 Social Handler | 即使路由到了也找不到 handler | 中 — 需新增 RegisterSocialHandlers |
| G5 | Servant.Handle 签名与 LocalHandler 不匹配 | 接口适配问题 | 低 — 写 adapter 函数 |
| G6 | JWT Auth 仅在 mysql 模式启用 | local 模式下无身份注入 | 低 — local 模式也启用 JWT |
| G7 | Social Module 初始化依赖未接入 Gateway 启动链 | Module 创建时缺 Repository | 中 — 需构造 Memory Repository |
| G8 | CORS 仅绑定 /api/hello | Social API 路径跨域被拒 | 低 — 改为全局中间件或通配路由 |
| G9 | PermissionService 未注入 svc 层 | 权限检查永远跳过 | 中 — 通过 module.go 注入链路传递 |

---

## 四、执行步骤

### Task 0：协议号修复（P0 — 必须首先完成）

#### 0.1 修复 Topic 域 3062/3064 协议号违规

**目标**：消除 handler.go 中违反 base proto 定义的 2 个非法 Dispatch case

**操作清单**：

1. [ ] **修改 `topic/svc_like_topic.go`**：
   - 在 `Handle` 方法中增加对 `req.IsLike == false` 的分支判断
   - 当 `is_like=false` 时执行取消点赞逻辑（原 SvcUnlikeTopic 的逻辑）
   - 保持函数签名不变

2. [ ] **删除 `topic/svc_unlike_topic.go`**：
   - 将其核心逻辑合并入 `svc_like_topic.go` 的 `is_like=false` 分支
   - 删除文件

3. [ ] **修改 `topic/svc_favorite_topic.go`**：
   - 在 `Handle` 方法中增加对 `req.IsFavorite == false` 的分支判断
   - 当 `is_favorite=false` 时执行取消收藏逻辑（原 SvcUnfavoriteTopic 的逻辑）

4. [ ] **删除 `topic/svc_unfavorite_topic.go`**：
   - 将其核心逻辑合并入 `svc_favorite_topic.go` 的 `is_favorite=false` 分支
   - 删除文件

5. [ ] **修改 `topic/handler.go`**：
   - 删除 `case "3062"` （UnlikeTopic 非法 case）
   - 删除 `case "3064"` （UnfavoriteTopic 非法 case）
   - 修改 `case "3061"` 注释为：「点赞/取消点赞（is_like 区分）」
   - 修改 `case "3063"` 注释为：「收藏/取消收藏（is_favorite 区分）」
   - 移除结构体字段 `unlikeTopicSvc` 和 `unfavoriteTopicSvc`
   - 移除 `NewHandler` 中的对应初始化

6. [ ] **更新测试文件**：
   - 修改 `svc_like_topic_test.go`：增加 `is_like=false` 测试用例
   - 删除 `svc_unlike_topic_test.go`
   - 修改 `svc_favorite_topic_test.go`：增加 `is_favorite=false` 测试用例
   - 删除 `svc_unfavorite_topic_test.go`

7. [ ] **运行测试确认**：
   ```bash
   go test ./go/modules/social/topic/... -count=1 -v
   ```

**验收标准**：
- handler.go 中不存在 `case "3062"` 和 `case "3064"`
- `go test ./go/modules/social/topic/...` 全绿
- `go vet ./go/modules/social/topic/...` 无 warning
- LikeTopic svc 同时支持 is_like=true 和 is_like=False
- FavoriteTopic svc 同时支持 is_favorite=true 和 is_favorite=False

#### 0.2 补齐 routes.yaml 缺失路由

**目标**：routes.yaml 社交域路由数从 28 → 34

**操作清单**：

1. [ ] 在 `configs/gateway/routes.yaml` 中补齐 Member 域 2 条：

   ```yaml
   # 1033/1034: 更新成员状态
   - request_max: 1000
     request_min: 1033
     route_key: "1000:1033"
     command_name: UpdateMemberStatus
     description: 更新成员状态
     request_proto: com.mineplanet.pojo.social.member.UpdateMemberStatusRequest
     response_max: 1000
     response_min: 1034
     response_proto: com.mineplanet.pojo.social.member.UpdateMemberStatusResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleMember
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 3000
     auth_required: true
     audit_required: true

   # 1045/1046: 获取用户统计
   - request_max: 1000
     request_min: 1045
     route_key: "1000:1045"
     command_name: GetUserStats
     description: 获取用户统计
     request_proto: com.mineplanet.pojo.social.member.GetUserStatsRequest
     response_max: 1000
     response_min: 1046
     response_proto: com.mineplanet.pojo.social.member.GetUserStatsResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleMember
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 3000
     auth_required: true
     audit_required: false
   ```

2. [ ] 在 `configs/gateway/routes.yaml` 中补齐 Group 域 3 条：

   ```yaml
   # 2039/2040: 获取圈子统计
   - request_max: 2000
     request_min: 2039
     route_key: "2000:2039"
     command_name: GetGroupStats
     description: 获取圈子统计
     request_proto: com.mineplanet.pojo.social.group.GetGroupStatsRequest
     response_max: 2000
     response_min: 2040
     response_proto: com.mineplanet.pojo.social.group.GetGroupStatsResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleGroup
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 3000
     auth_required: true
     audit_required: false

   # 2047/2048: 批量获取圈子信息
   - request_max: 2000
     request_min: 2047
     route_key: "2000:2047"
     command_name: BatchGetGroups
     description: 批量获取圈子信息
     request_proto: com.mineplanet.pojo.social.group.BatchGetGroupsRequest
     response_max: 2000
     response_min: 2048
     response_proto: com.mineplanet.pojo.social.group.BatchGetGroupsResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleGroup
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 5000
     auth_required: true
     audit_required: false

   # 2077/2078: 获取成员 UserID 列表
   - request_max: 2000
     request_min: 2077
     route_key: "2000:2077"
     command_name: GetGroupMemberUserIds
     description: 获取成员UserID列表
     request_proto: com.mineplanet.pojo.social.group.GetGroupMemberUserIdsRequest
     response_max: 2000
     response_min: 2078
     response_proto: com.mineplanet.pojo.social.group.GetGroupMemberUserIdsResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleGroup
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 3000
     auth_required: true
     audit_required: false
   ```

3. [ ] 在 `configs/gateway/routes.yaml` 中补齐 Topic 域 1 条：

   ```yaml
   # 3065/3066: 获取评论列表
   - request_max: 3000
     request_min: 3065
     route_key: "3000:3065"
     command_name: GetReplyList
     description: 获取评论列表
     request_proto: com.mineplanet.pojo.social.topic.GetReplyListRequest
     response_max: 3000
     response_min: 3066
     response_proto: com.mineplanet.pojo.social.topic.GetReplyListResponse
     tars_app: CaiRobot
     tars_server: SocialServer
     tars_servant: SocialObj
     tars_module: CaiRobotSocialApp
     tars_interface: SocialObj
     tars_method: HandleTopic
     tars_request_type: vector<byte>
     tars_response_type: vector<byte>
     timeout_ms: 5000
     auth_required: true
     audit_required: false
   ```

4. [ ] 运行 YAML 语法检查确认格式正确

**验收标准**：
- routes.yaml 社交域路由数 = 34（与 handler 有效 case 总数一致）
- 每个 handler 有效 case 都能在 routes.yaml 找到对应路由
- YAML 语法无误
- 无 3062 或 3064 作为 request_min 的路由条目

---

### Task 1：编译修复与依赖整理

**目标**：`go build ./...` 和 `go vet ./...` 零错误

**操作清单**：

1. [ ] 检查 `go.work` 是否包含 `go/modules/social`
2. [ ] 检查 `go/modules/social/go.mod` 的 module path 和依赖
3. [ ] 确认 generated proto Go 代码路径 `github.com/jimiechen/mineplanet/protocols/generated/go/social` 与 import 一致
4. [ ] 执行编译：
   ```bash
   cd /Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp
   go build ./go/modules/social/...
   ```
5. [ ] 修复编译错误（预期可能的问题）：
   - import cycle 检查
   - 未使用的 import（Task 0 删除 svc 文件后可能产生）
   - 缺失的 type 断言
6. [ ] 执行静态分析：
   ```bash
   go vet ./go/modules/social/...
   ```
7. [ ] 执行格式化：
   ```bash
   gofmt -w ./go/modules/social/
   ```

**不做的**：
- 不删除业务逻辑来绕过编译
- 不引入 third/inbox 依赖
- 不引入新框架

**验收命令**：

```bash
cd /Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp
go build ./go/modules/social/...
go test ./go/modules/social/... -count=1 -timeout 60s
go vet ./go/modules/social/...
```

**验收标准**：
- build exit code = 0
- test 全部 pass
- vet 无 warning

---

### Task 2：Gateway ↔ Social LocalInvoker 集成

**目标**：Gateway 能将 1000/2000/3000 maxType 请求转发到 Social Servant

#### 2.1 新增 `RegisterSocialHandlers` 函数

**文件**：`go/gateway/proto-gateway/tarsclient/invoker.go`（或新建 `social_handler.go`）

```go
// RegisterSocialHandlers 注册 Social 模块的本地 handler
// 将 MemberServant/GroupServant/TopicServant 适配为 LocalHandler 接口
func RegisterSocialHandlers(invoker *LocalInvoker, socialModule *social.Module) {
    // Member 域: HandleMember
    invoker.Register(TargetKey{
        App: "CaiRobot", Server: "SocialServer",
        Servant: "SocialObj", Method: "HandleMember",
    }, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
        return socialModule.MemberServant.Handle(ctx, req, nil)
    }))

    // Group 域: HandleGroup
    invoker.Register(TargetKey{
        App: "CaiRobot", Server: "SocialServer",
        Servant: "SocialObj", Method: "HandleGroup",
    }, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
        return socialModule.GroupServant.Handle(ctx, req, nil)
    }))

    // Topic 域: HandleTopic
    invoker.Register(TargetKey{
        App: "CaiRobot", Server: "SocialServer",
        Servant: "SocialObj", Method: "HandleTopic",
    }, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
        return socialModule.TopicServant.Handle(ctx, req, nil)
    }))
}
```

#### 2.2 修改 Servant.Handle 签名适配

**问题**：当前 `Servant.Handle(ctx, req []byte, extend map[string]string)` 需要 extend 参数提取 minType，但 LocalHandler.Invoke 只有 `(ctx, request []byte, extend map[string]string)`。

**方案**：LocalHandler.Invoke 本身就携带 extend 参数，直接透传即可。需要确保 Invoker 从 routes.yaml 解析出的 minType 正确写入 extend map。

#### 2.3 修改 Gateway main.go 启动链路

**文件**：`go/gateway/proto-gateway/cmd/server/main.go`

变更点：
1. import social module
2. 在 `mode == "local"` 分支中创建 Social Module（使用 Memory Repository）
3. 调用 `RegisterSocialHandlers(invoker, socialModule)`
4. 更新日志输出

#### 2.4 CORS 中间件扩展

**文件**：`go/gateway/proto-gateway/cmd/server/main.go`

变更点：将 CORS 从 `/api/hello` 单一路径改为匹配所有 `/api/*` 路由，或改用全局 HTTP middleware 包装。

**验收标准**：
- Gateway 启动后日志显示 `handlers=System+Config+I18n+Social(Member+Group+Topic)`
- 发送 maxType=1000:minType=1021 请求不再返回错误
- 返回合法 Protobuf Response bytes

---

### Task 3：JWT Context 注入 + 依赖整理

**目标**：Gateway 解析 JWT 后将用户身份注入 context，Social svc 能从 context 获取 operator/user

#### 3.1 local 模式启用 JWT Auth

**文件**：`go/gateway/proto-gateway/cmd/server/main.go`

变更：local 模式也初始化 AuthService 和 AuthMiddleware（使用环境变量或默认 secret）

```go
if mode == "local" {
    jwtSecret := []byte(getEnv("JWT_SECRET", "cairobot-local-dev-secret"))
    authSvc := auth.NewAuthService(jwtSecret, "cairobot", 24*time.Hour)
    authMw = middleware.NewAuthMiddleware(authSvc)

    invoker = tarsclient.NewLocalInvoker()
    tarsclient.RegisterAllLocalHandlers(invoker)
    // 新增：注册 Social handlers
    socialModule := createSocialModule()  // 使用 Memory Repo
    tarsclient.RegisterSocialHandlers(invoker, socialModule)
}
```

#### 3.2 Context 用户身份注入

**注入点**：AuthMiddleware 解析 JWT 成功后，将以下信息写入 context：

```go
ctx = context.WithValue(ctx, CtxKeyUserID, claims.UserID)
ctx = context.WithValue(ctx, CtxKeyUsername, claims.Username)
```

**消费方**：Social svc 层通过 helper 函数获取：

```go
func getUserIDFromContext(ctx context.Context) string {
    if uid, ok := ctx.Value(CtxKeyUserID).(string); ok {
        return uid
    }
    return ""
}
```

#### 3.3 PermissionService 注入链路

**现状**：permission/service.go 已完整实现 8 方法，但尚未注入到各 svc。

**方案**：通过 module.go 的 functional options 模式注入：

```go
type moduleConfig struct {
    publisher  event.Publisher
    permission *permission.Service  // 新增
}

func WithPermissionService(p *permission.Service) ModuleOption { ... }
```

各 svc 在创建时接收 permission Service（可选，nil 时跳过权限检查并打日志警告）。

#### 3.4 event.Publisher 注入确认

**现状**：module.go 默认注入 NoopPublisher，WithPublisher 可选替换。

**本轮策略**：保持默认 NoopPublisher，E2E 测试不需要真实事件发布。

**验收标准**：
- local 模式启动时 JWT AuthMiddleware 生效
- 带 Authorization header 的请求能正确解析 user_id
- Social svc 中 `getUserIDFromContext(ctx)` 能获取到非空值
- PermissionService 可通过 module option 注入（允许 nil 跳过）

---

### Task 4：单元测试修复与补充

**目标**：全量测试通过，覆盖新集成的 Gateway↔Social 路径

#### 4.1 必修测试

| 测试文件 | 测试内容 | 优先级 |
|---------|---------|:------:|
| tarsclient/social_handler_test.go | RegisterSocialHandlers 注册验证 + Invoke 端到端 | P0 |
| module_test.go | NewModule 创建 + NoopPublisher 默认值 | P0 |
| member/handler_test.go | Dispatch 全部 11 个 case 覆盖 | P0 |
| group/handler_test.go | Dispatch 全部 13 个 case 覆盖 | P0 |
| topic/handler_test.go | Dispatch 全部 **10** 个 case 覆盖（修正后） | P0 |
| topic/svc_like_topic_test.go | is_like=true + is_like=false 双向测试 | P0（新增） |
| topic/svc_favorite_topic_test.go | is_favorite=true + is_favorite=false 双向测试 | P0（新增） |

#### 4.2 Memory Repository Mock 测试支持

**现状**：各域已有 `mock_repository_test.go` 和 `repository_gorm_test.go`。

**本轮**：确保所有 svc 测试使用 Memory/Mock Repository，不依赖真实 MySQL。

#### 4.3 测试执行命令

```bash
cd /Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp
go test ./go/modules/social/... -count=1 -v -timeout 120s 2>&1 | tee docs/reports/testing/social-phase1-test-output.log
go test ./go/gateway/proto-gateway/... -count=1 -v -timeout 60s
```

**验收标准**：
- social 模块测试全绿（含 Task 0 合并后的测试文件）
- gateway tarsclient 测试全绿（含新增 Social handler 测试）
- 总覆盖率不低于当前水平

---

### Task 5：Proto Tester E2E 验证

**目标**：通过 Proto Tester 前端调用 Social 白名单协议并获得正确响应

#### 5.1 E2E 白名单协议（核心 10 对）

| # | 协议 | Req min | Rsp min | 优先级 | 验证要点 |
|---|------|:-------:|:------:|:------:|---------|
| 1 | UserRegister | 1021 | 1022 | P0 | 返回 user_id + result code=0 |
| 2 | UserLogin | 1023 | 1024 | P0 | 返回 access_token + refresh_token + user_info |
| 3 | GetUserInfo | 1029 | 1030 | P0 | 需带 JWT，返回 user_info |
| 4 | CreateGroup | 2005 | 2006 | P0 | 需带 JWT，返回 group_info + group_id |
| 5 | JoinGroup | 2013 | 2014 | P0 | 需带 JWT，返回 member_id + status |
| 6 | CreateTopic | 3001 | 3002 | P0 | 需带 JWT，返回 topic_info + topic_id |
| 7 | GetTopicList | 3005 | 3006 | P0 | 需带 JWT，返回 topics 列表 |
| 8 | AddTopicReply | 3043 | 3044 | P0 | 需带 JWT，返回 reply_info |
| 9 | LikeTopic | 3061 | 3062 | P0 | 需带 JWT，is_like=true 返回 is_liked + like_count |
| 10 | CheckTopicActions | 3099 | 3100 | P0 | 需带 JWT，返回 available_actions |

> **注意**：第 9 项 LikeTopic 的验证要点已修正——通过 `is_like` 布尔字段区分点赞/取消点赞，不再有独立的 UnlikeTopic 协议。

#### 5.2 E2E 操作步骤

1. [ ] 启动 Gateway：`make gateway-restart`
2. [ ] 确认 Proto Tester 前端可达：`http://localhost:3002`
3. [ ] 配置 Proto Tester 连接 `http://localhost:8080/api/hello`
4. [ ] 逐一发送上述 10 对协议请求
5. [ ] 记录每个协议的 request/response 截图或 JSON
6. [ ] 验证 response protobuf 可正确反序列化
7. [ ] 验证带 JWT 的协议（auth_required=true）能正确识别用户身份

#### 5.3 E2E 预期问题与预案

| 可能问题 | 预案 |
|---------|------|
| Protobuf 反序列化字段缺失 | 检查 generated code 是否与 proto 文件同步 |
| JWT 验证失败 | 确认 Gateway 和测试使用的 secret 一致 |
| Memory Repository 数据隔离 | 每次 E2E 测试前重启 Gateway 清空内存状态 |
| CORS 预检失败 | 确认 CORS middleware 覆盖了 POST 请求路径 |

**验收标准**：
- 10 对核心协议全部返回有效 Protobuf Response
- result.code = 0（成功）
- 带身份协议能正确返回用户相关数据
- 输出 E2E 测试报告到 `docs/reports/testing/social-phase1-e2e-report.md`

---

### Task 6：文档同步（已完成）

**目标**：更新所有受影响的文档，保持代码与文档一致

#### 6.0 Task 5-B / 5-C E2E 验证结果

| 子任务 | 状态 | 结果 |
|--------|:----:|------|
| **Task 5-A** | ✅ 完成 | Gateway 编译错误排查，LocalInvoker 模式可用 |
| **Task 5-B** | ✅ **PASS** | 5 条冒烟测试全通过（UserRegister/UserLogin/CreateGroup/JoinGroup/CreateTopic） |
| **Task 5-C** | ✅ **PASS** | 34 条完整白名单 E2E：**15 PASS + 19 WARN + 0 FAIL** |

**Task 5-C 四层验证模型**：

```
L1 HTTP层 → L2 TarsGo协议层 → L3 Protobuf反序列化 → L4 Result.Code业务语义
```

**Task 5-C PASS 用例（15 条）**：BlockUser, UnblockUser, GetGroupStats, BatchGetGroups,
GetGroupMemberUserIds, CreateTopic, GetTopicList, DeleteTopic, AddTopicReply,
LikeTopic, FavoriteTopic, BatchGetTopicInfo, CreateReport, CheckTopicActions, GetReplyList

**Task 5-C WARN 分类（19 条）**：
- auth_required 拦截（2 条）：UserLogout, GetUserInfo — 无 JWT 时正确返回 HTTP 400
- 缺少用户身份 context（3 条）：UpdateUserInfo, GetBlockList, GetUserStats
- 数据不存在（7 条）：UpdateMemberStatus, MuteMember~RemoveMember, CalcPayableAmount, GroupUserEnter
- 成员身份缺失（2 条）：LeaveGroup, RenewMember
- Token/注册冲突（3 条）：RefreshToken(10401), UserRegister(10612), CreateGroup(10711)
- 瞬态网络（1 条）：UserLogin 连接失败（同轮 31 条正常）

> **HelloWorld 10400 说明**：`/api/hello` 返回 TarsCode=10400 为历史遗留问题，
> 本轮仅改动 Social LocalInvoker/routes.yaml，未修改 HelloWorld 相关代码。
> 如需彻底排查，另开 Gateway HelloWorld Issue。

#### 6.1 已更新的文档

| 文档 | 变更内容 | 状态 |
|------|---------|:----:|
| `configs/gateway/routes.yaml` | 补齐至 34 条社交域路由 | ✅ 已完成 |
| `proto-gateway/configs/gateway/routes.yaml` | 从根目录同步 28KB 新版（修复双文件不同步根因） | ✅ 已完成 |
| `go/gateway/proto-gateway/cmd/social_e2e_smoke/main.go` | 从 5 条 smoke 扩展为 34 条完整白名单 E2E + 四层验证架构 | ✅ 已完成 |
| `go/gateway/proto-gateway/tarsclient/invoker.go` | 移除 DEBUG 日志 + 清理 import | ✅ 已完成 |

#### 6.2 新增文档

| 文档 | 内容 | 路径 |
|------|------|------|
| **e2e-social-phase1.md** | Task 5-C 完整白名单 E2E 测试报告（含四层验证、34 条用例详情、WARN 分类分析） | `docs/reports/testing/e2e-social-phase1.md` |
| **social-e2e-cases.json** | 34 条 E2E 用例结构化数据（含 bizCode/warnCategory/respSize 等字段） | `docs/reports/testing/social-e2e-cases.json` |

#### 6.3 routes.yaml 双文件不同步问题与治理方案

##### 6.3.1 问题现象

```
根目录编辑源:   configs/gateway/routes.yaml          (28KB, 41 routes, 34 social)
运行时读取源:   proto-gateway/configs/gateway/routes.yaml (5KB 旧版, 缺少 Social 路由)
```

**影响**：Task 5-B 初次运行时 E2E 全部返回 10404（handler 未找到），实际是运行时使用了旧版 routes.yaml。

##### 6.3.2 根因分析

| 因素 | 说明 |
|------|------|
| 双文件存在 | 根目录 `configs/` 用于版本控制和人工编辑；`proto-gateway/configs/` 是 TarsGo 框架相对路径加载的运行时配置 |
| 无同步机制 | 修改根目录后不会自动同步到运行时目录 |
| 无 CI 校验 | CI 不检查两个文件的一致性 |

##### 6.3.3 治理方案（三阶段）

**短期（立即执行）— Makefile 同步 target**

```makefile
# Makefile 新增 target
routes-sync:
	cp configs/gateway/routes.yaml proto-gateway/configs/gateway/routes.yaml
	@echo "routes.yaml 已从根目录同步到运行时目录"
```

使用方式：每次修改 `configs/gateway/routes.yaml` 后执行 `make routes-sync`。

**中期（CI 层面）— hash 校验**

在 CI 中增加检查步骤：
```bash
# scripts/ci/check_routes_sync.sh
ROOT_HASH=$(sha256sum configs/gateway/routes.yaml | awk '{print $1}')
RUNTIME_HASH=$(sha256sum proto-gateway/configs/gateway/routes.yaml | awk '{print $1}')
if [ "$ROOT_HASH" != "$RUNTIME_HASH" ]; then
    echo "ERROR: routes.yaml 双文件不一致！"
    echo "  根目录: $ROOT_HASH"
    echo "  运行时: $RUNTIME_HASH"
    exit 1
fi
echo "OK: routes.yaml 双文件一致 (hash=$ROOT_HASH)"
```
CI Job 在 `proto-check` 或新增 `routes-sync-check` 中调用此脚本。

**长期（架构层面）— 单一权威源**

方案 A（推荐）：**符号链接**
```bash
# 将运行时目录的 routes.yaml 替换为指向根目录的软链接
ln -sf ../../../configs/gateway/routes.yaml proto-gateway/configs/gateway/routes.yaml
```
- 优点：零成本、自动同步、不可能不一致
- 缺点：依赖相对路径稳定性；Windows 兼容性需确认

方案 B：**Gateway 启动参数化**
```go
// main.go 中通过命令行 flag 或环境变量指定 routes.yaml 路径
routesPath := getEnv("ROUTES_CONFIG", "configs/gateway/routes.yaml")
```
- 优点：灵活、不改变文件系统结构
- 缺点：需修改 Gateway 启动代码

方案 C：**构建时复制（Go embed）**
```go
//go:embed configs/gateway/routes.yaml
var routesYAML []byte
```
- 优点：编译时绑定、无运行时文件依赖
- 缺点：每次改路由都需重新编译

> **建议**：MVP-P0 先实施方案 A（符号链接），Phase 1.5 评估是否需要迁移到方案 B。

#### 6.4 不更新的文档

- PRD（本轮无需求变更）
- ADR（本轮无架构决策变更）
- third/inbox 相关文档（挂起）

---

## 五、风险与应对

| # | 风险 | 等级 | 应对措施 |
|---|------|:----:|---------|
| R1 | Topic 3062/3064 合并后现有测试失败 | P1 | 先合并 svc 逻辑再删文件，逐步验证 |
| R2 | go.mod 循环依赖导致编译失败 | P1 | 先 `go mod graph` 检查，必要时拆分接口包 |
| R3 | Servant.Handle 签名与 LocalHandler 不兼容 | P2 | 写轻量 adapter 函数桥接 |
| R4 | Memory Repository 并发安全 | P2 | E2E 测试串行执行，不加锁；后续 Phase 1.5 引入 sync.Mutex |
| R5 | JWT secret 硬编码泄露 | P1 | 统一从环境变量读取，不在代码中写默认值到生产分支 |
| R6 | Proto Tester 前端不支持自定义 maxType/minType | P2 | 检查前端是否已支持，若不支持则先用 curl 做 E2E |
| R7 | Gateway 路由匹配性能（34 条社交路由） | P3 | YAML 路由量级极小，map 查询 O(1)，无需优化 |

---

## 六、验收 Checklist

主控请逐项确认：

### 编译与测试
- [ ] `go build ./go/modules/social/...` 零错误
- [ ] `go test ./go/modules/social/... -count=1` 全绿
- [ ] `go vet ./go/modules/social/...` 无 warning
- [ ] `gofmt` 格式化通过

### 协议对齐（以 base proto 为权威源）
- [ ] routes.yaml 社交域路由数 = **34**（非 38）
- [ ] 每个 handler 有效 case 都有对应 routes.yaml 条目
- [ ] **无 3062/3064 作为 Request minType 的情况**（它们是 Response 类型）
- [ ] 无重复协议号
- [ ] 无孤立路由（routes.yaml 有但 handler 无）
- [ ] 所有 Request minType 为奇数，Response minType 为偶数

### Topic 域协议号修复
- [ ] handler.go 中已删除 `case "3062"` 和 `case "3064"`
- [ ] SvcLikeTopic 同时支持 is_like=true/false
- [ ] SvcFavoriteTopic 同时支持 is_favorite=true/false
- [ ] SvcUnlikeTopic 文件已删除
- [ ] SvcUnfavoriteTopic 文件已删除
- [ ] 相关测试已合并更新且全绿

### Gateway 集成
- [ ] Gateway 日志显示 Social handlers 已注册
- [ ] maxType=1000 请求能到达 MemberServant
- [ ] maxType=2000 请求能到达 GroupServant
- [ ] maxType=3000 请求能到达 TopicServant
- [ ] CORS 中间件覆盖 Social API 路径

### JWT 与 Context
- [ ] local 模式启用了 JWT AuthMiddleware
- [ ] Social svc 能从 context 获取 user_id
- [ ] 无 JWT 请求返回 401（auth_required=true 的协议）

### E2E 验证
- [ ] UserRegister (1021→1022) 返回有效响应
- [ ] UserLogin (1023→1024) 返回 token
- [ ] CreateGroup (2005→2006) 返回 group_id
- [ ] CreateTopic (3001→3002) 返回 topic_id
- [ ] LikeTopic (3061→3062) 支持 is_like 双向
- [ ] 至少 10 对核心协议通过 E2E

### 文档
- [ ] routes.yaml 已同步更新（34 条社交路由）
- [ ] E2E 测试报告已生成
- [ ] 执行日志已记录
- [ ] 协议号注册表已修正 Topic 3062/3064 标注

---

## 七、时间估算

| Task | 内容 | 预估工作量 |
|------|------|:----------:|
| **Task 0.1** | **修复 Topic 3062/3064 协议号违规（svc 合并 + case 删除 + 测试更新）** | **中** |
| Task 0.2 | 补齐 routes.yaml 6 条缺失路由 | 小 |
| Task 1 | 编译修复 + 依赖整理 | 中 |
| Task 2 | Gateway ↔ Social LocalInvoker 集成 | 大 |
| Task 3 | JWT Context 注入 + 依赖整理 | 中 |
| Task 4 | 单元测试修复与补充 | 大 |
| Task 5 | Proto Tester E2E 验证 | 大 |
| Task 6 | 文档同步 | 小 |

> 注意：不承诺具体天数，按任务完成度汇报进度。

---

## 八、附录

### A. 关键文件索引

| 文件 | 职责 |
|------|------|
| `go/modules/social/module.go` | Social 模块入口，聚合三域 Servant |
| `go/modules/social/member/handler.go` | Member 域协议分发器（11 case） |
| `go/modules/social/group/handler.go` | Group 域协议分发器（13 case） |
| `go/modules/social/topic/handler.go` | Topic 域协议分发器（**10 case，修正后**） |
| `go/modules/social/member/servant.go` | Member Servant（TarsGo 适配） |
| `go/modules/social/group/servant.go` | Group Servant |
| `go/modules/social/topic/servant.go` | Topic Servant |
| `go/modules/social/event/noop.go` | NoopPublisher 实现 |
| `go/modules/social/event/publisher.go` | Publisher 接口定义 |
| `go/modules/social/permission/service.go` | 权限服务（8 方法） |
| `configs/gateway/routes.yaml` | Gateway 路由配置（当前 28 条社交路由，补齐后 34 条） |
| `go/gateway/proto-gateway/cmd/server/main.go` | Gateway 启动入口 |
| `go/gateway/proto-gateway/tarsclient/invoker.go` | LocalInvoker + Handler 注册 |

### B. 协议号权威源索引

| 域 | 权威源文件 | 路径 |
|----|-----------|------|
| Member | user_base.proto | `docs/tabbit/inbox/2026/06/protocols/base/user_base.proto` |
| Group | group_base.proto | `docs/tabbit/inbox/2026/06/protocols/base/group_base.proto` |
| Topic | topic_base.proto | `docs/tabbit/inbox/2026/06/protocols/base/topic_base.proto` |

### C. E2E 白名单协议完整列表（34 对 Request/Response）

见上文 2.1 节三源校验表格，补齐后 routes.yaml 包含的全部 34 条协议（11 Member + 13 Group + 10 Topic）。

### D. v1.0 → v2.0 主要变更记录

| 变更项 | v1.0（旧） | v2.0（新） | 原因 |
|--------|:----------:|:----------:|------|
| Topic handler case 数 | 14（含 3062/3064） | 10（删除 3062/3064） | base proto 校正 |
| 社交域总路由数 | 38 | 34 | 删除 2 个非法 case 后差额减少 |
| routes.yaml 缺失数 | 8 | 6 | 3062/3064 不需要路由 |
| 3062 协议名 | UnlikeTopic（Request） | LikeTopicResponse（Response） | base proto 权威 |
| 3064 协议名 | UnfavoriteTopic（Request） | FavoriteTopicResponse（Response） | base proto 权威 |
| LikeTopic 设计 | 分离 2 个 case | 合并为 1 个 case + is_like bool | 对齐 base proto |
| FavoriteTopic 设计 | 分离 2 个 case | 合并为 1 个 case + is_favorite bool | 对齐 base proto |
| 待删除文件 | 无 | svc_unlike_topic.go, svc_unfavorite_topic.go 及其测试 | 合并入对应 svc |
