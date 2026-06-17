# 社交域 MVP 实施方案 v2

> 文档版本：v2.0（整合主控 DevGuide DEV-GUIDE-SOCIAL-001）
> 更新日期：2026-06-16
> 状态：**待评审**
> 基础文档：
> - [PRD-social-app-mvp.md](../prd/PRD-social-app-mvp.md)
> - [ADR-social-data-level-and-cache-strategy.md](../adr/ADR-social-data-level-and-cache-strategy.md)
> - [ADR-plaza-virtual-membership.md](../adr/ADR-plaza-virtual-membership.md)
> - [social-service-dev-guide.md](../tabbit/inbox/2026/06/protocols/social-service-dev-guide.md) — 主控下发开发规范
> - [协议编号注册表.md](../../api/协议编号注册表.md) §4

---

## 1. 审阅意见总览

### 1.1 对主控 DevGuide (DEV-GUIDE-SOCIAL-001) 的审阅结论：**WARN（有条件通过）**

DevGuide 在架构设计、代码模板、开发规范方面**质量极高**（1208 行覆盖 7 层调用链、5 个域完整目录、6 套代码模板、12 步开发顺序、17 条禁止行为），但存在以下必须修正的问题：

| # | 级别 | 问题 | 正确值 | 影响 |
|---|---|---|---|---|
| 1 | **P0** | §3 L143 `svc_group_user_enter.go` 标注 `min=2067` | **min=2087**（2067 是 UpdateGroupDiscounts） | 运行时路由分发到错误的 svc |
| 2 | **P0** | §6 L410 / §7 L518 proto import 路径含 `/proto/` 多余层 | `github.com/jimiechen/mineplanet/protocols/generated/go/base`（去掉 `/proto/`） | 编译失败 |
| 3 | **P1** | Permission Service 缺少 `CanAuditContent` 方法（PRD §10.2 第 8 个能力） | 补充第 8 个方法 | 审核帖子无统一权限入口 |
| 4 | **P1** | 权限方法命名与 PRD 不一致（CanViewTopicDetail vs CanReadTopic 等） | 统一为 PRD §10.2 命名 | 代码审查混乱 |
| 5 | **P1** | 缓存 Key 前缀用 `user:` 而 PRD 用 `member:` | 统一为 PRD §11.4 的 `member:` 前缀 | 缓存 key 不一致 |
| 6 | **P2** | 未标注 MVP 首期范围（5 个域全部列出，易误解需一次实现） | 标注 member/group/topic 为首期，third/inbox 推迟 Phase 2 | 范围蔓延风险 |

### 1.2 对 Trae Implementation Plan v1 的审阅结论：**需要重写**

v1 方案在实施条件评估和工程准备任务方面有价值，但存在以下致命缺陷：

| # | 问题 | 说明 |
|---|---|---|
| 1 | 目录结构按层分子包（internal/handler/usecase/...） | 与 DevGuide 按域分子包及 PRD §2.2 约定冲突 |
| 2 | Proto import 路径 `proto/gen/go/social/v1` 不存在于项目 | 实际路径是 `protocols/generated/go/base` |
| 3 | 缺少全部代码模板（servant/handler/svc/repo/model/permission） | 开发者无法开始编码 |
| 4 | 缺少开发规范（三条铁律、禁止行为、数据等级表） | 违规无法预防 |

### 1.3 合并策略

> **以 DevGuide 为骨架（代码模板、规范、约束），以 v1 的工程准备和项目管理内容为补充。**

---

## 2. 三条不可违反规则（来自 DevGuide §1）

### Rule 1：一协议一 svc 文件
每个 Request/Response 协议对对应一个独立的 `svc_{action}.go` 文件。
Trae **每次只允许创建或修改一个** `svc_*.go` 文件，禁止在一个文件里实现多个协议。

### Rule 2：proto 文件冻结
客户端已接入。`*Request` / `*Response` 的字段名、字段编号、package 名全部不可变。
Service 文件内部使用内部 `model.*` 类型，通过转换函数与 proto 类型解耦。

### Rule 3：数据库 model 层独立于 proto
`model.go` 中的结构体字段按 `basemodel.md` + PRD 设计，与 proto 字段无关。
proto 字段 → model 字段的映射在 `svc_*.go` 中的转换函数内完成。

---

## 3. 架构总览（7 层调用链）

```
Client App
   │  POST /api/hello
   │  Body: MessagePacket{ maxType, minType, extend, platform, data(proto bytes) }
   ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Gateway  go/gateway/proto-gateway/                  │
│  ① Decode MessagePacket → {maxType, minType, extend, data}  │
│  ② 查 routes.yaml → Target{App,Server,Servant,Method}       │
│  ③ extend["minType"] = "1021"  (注入 minType 字符串)         │
│  ④ TarsInvoker.Invoke(ctx, target, data, extend)             │
└──────────────────────────┬──────────────────────────────────┘
                           │ TarsInvoker.Invoke()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 2: TarsInvoker  tarsclient/invoker.go                  │
│  MVP S1 单体: LocalInvoker                                   │
│    handlers map[TargetKey]LocalHandler                       │
│    按 TargetKey{App,Server,Servant,Method} 查 map            │
│    → handler(ctx, reqBytes, extend)                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ LocalHandler(ctx, reqBytes, extend)
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 3: Servant  {domain}/servant.go                       │
│  Handle(ctx, req []byte, extend map[string]string)           │
│  → (retCode int, respBytes []byte, err error)               │
│  职责：注册自身 + 提取 minType + 转发给 Handler              │
│  ⚠️ 不包含业务逻辑；不解析 proto bytes                        │
└──────────────────────────┬──────────────────────────────────┘
                           │ Handler.Dispatch()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 4: Handler  {domain}/handler.go                      │
│  switch minType → 找到对应 svc                               │
│  proto.Unmarshal → svc.Handle → proto.Marshal                │
│  ⚠️ switch case 以外禁止业务逻辑                              │
└──────────────────────────┬──────────────────────────────────┘
                           │ svc.Handle()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 5: svc_*.go  (一协议一文件)                            │
│  ① 参数校验 → UserErrorCode/GroupErrorCode                   │
│  ② 权限校验 → permission.Service                             │
│  ③ 1级数据读写 → MySQL 事务，Repository 接口                 │
│  ④ 发布领域事件 → 2级数据异步更新 Redis                       │
│  ⑤ 返回 proto Response                                     │
└──────────────────────────┬──────────────────────────────────┘
                           │ repo.XXX()
┌──────────────────────────▼──────────────────────────────────┐
│ Layer 6: Repository + Model                                  │
│  repository.go        ← DB 操作接口定义                       │
│  repository_gorm.go   ← GORM 实现                            │
│  model.go             ← 内部 DB 模型（非 proto 类型）          │
│       → MySQL   (1级强一致数据)                               │
│       → Redis   (2级最终一致缓存)                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. 目录结构（以 DevGuide §3 为准，按域分子包）

> **MVP 首期仅实现 member(1000) / group(2000) / topic(3000) 三域**
> third(4000) 和 inbox(5000) 推迟到 Phase 2

```
go/modules/social/
├── module.go                           # 模块注册入口，聚合所有域 Servant
├── permission/
│   └── service.go                      # 跨域权限服务（8个方法，唯一权限入口）
│
├── member/                             # maxType=1000 ← MVP 首期
│   ├── servant.go                      # [L3] TarsGo Servant 注册+转发
│   ├── handler.go                      # [L4] minType switch dispatch
│   ├── repository.go                   # [L6] DB 接口定义
│   ├── repository_gorm.go              # [L6] GORM 实现
│   ├── model.go                        # User / MemberBlock / MemberStats
│   ├── svc_register.go                 # min=1021  UserRegisterRequest
│   ├── svc_login.go                    # min=1023  UserLoginRequest
│   ├── svc_logout.go                   # min=1025  UserLogoutRequest
│   ├── svc_refresh_token.go            # min=1027
│   ├── svc_get_user_info.go            # min=1029
│   ├── svc_update_user_info.go         # min=1031
│   ├── svc_block_user.go               # min=1039
│   ├── svc_unblock_user.go             # min=1041
│   ├── svc_get_block_list.go           # min=1043
│   └── ...（其余 member 协议各一个文件）
│
├── group/                              # maxType=2000 ← MVP 首期
│   ├── servant.go / handler.go / repository.go
│   ├── repository_gorm.go / model.go
│   ├── svc_create_group.go             # min=2005
│   ├── svc_update_group.go             # min=2009
│   ├── svc_delete_group.go             # min=2011
│   ├── svc_join_group.go               # min=2013
│   ├── svc_leave_group.go              # min=2015
│   ├── svc_group_user_enter.go         # min=2087  ⚠️ 原 DevGuide 误标为 2067
│   ├── svc_renew_member.go             # min=2037
│   ├── svc_calc_payable.go             # min=2073
│   ├── svc_mute_member.go              # min=2019
│   ├── svc_ban_member.go               # min=2023
│   ├── svc_remove_member.go            # min=2027
│   ├── svc_update_member_role.go       # min=2029
│   └── ...（其余 group 协议）
│
├── topic/                              # maxType=3000 ← MVP 首期
│   ├── servant.go / handler.go / repository.go
│   ├── repository_gorm.go / model.go
│   ├── svc_create_topic.go             # min=3001
│   ├── svc_get_topic_list.go           # min=3005
│   ├── svc_delete_topic.go             # min=3009
│   ├── svc_add_reply.go                # min=3043
│   ├── svc_like_topic.go               # min=3061
│   ├── svc_favorite_topic.go           # min=3063
│   ├── svc_pin_comment.go              # min=3081
│   ├── svc_create_report.go            # min=3095
│   └── ...（其余 topic 协议）
│
├── third/                              # maxType=4000 ← Phase 2
│   └── （目录结构同上，28 个 svc 文件）
│
└── inbox/                              # maxType=5000 ← Phase 2
    └── （目录结构同上，9 个 svc 文件）
```

### 与 hello/health 简单模块的差异说明

hello/health 是单域单协议模块，平铺文件即可。social 是**多域多协议复杂模块**（5 个域 × 30+ 个 svc 文件），按域分子包可确保：
- 每个域的 servant/handler/repo/model 内聚
- 跨域依赖仅通过 `permission/service.go` 的轻量接口
- 新增域时不影响已有域的代码

---

## 5. routes.yaml 配置（追加到现有 routes.yaml）

每个 maxType 域使用一个 Servant。minType 通过 Gateway 注入 `extend["minType"]` 传递，**不在 routes.yaml 中按 minType 拆分 Method**。

```yaml
# 追加到现有 routes.yaml

routes:
  # 用户成员域 (MVP 首期)
  - maxType: 1000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialMemberServer
      servant: SocialMemberServant
      method:  Handle
    desc: "社交域-用户成员协议组 (register/login/block...)"

  # 群组域 (MVP 首期)
  - maxType: 2000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialGroupServer
      servant: SocialGroupServant
      method:  Handle
    desc: "社交域-群组协议组 (create/join/mute/ban...)"

  # 主题域 (MVP 首期)
  - maxType: 3000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialTopicServer
      servant: SocialTopicServant
      method:  Handle
    desc: "社交域-主题协议组 (create/list/like/reply...)"

  # 第三方服务域 (Phase 2)
  - maxType: 4000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialThirdServer
      servant: SocialThirdServant
      method:  Handle
    desc: "社交域-第三方 (oss/share/oauth/wallet...)"

  # 收件箱消息域 (Phase 2)
  - maxType: 5000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialInboxServer
      servant: SocialInboxServant
      method:  Handle
    desc: "社交域-收件箱 (query/mark-read/send...)"
```

---

## 6. Phase 0：工程准备（预估 1-2 天）

> **前置条件**：本阶段所有任务完成后方可进入 Phase 1 编码。

### 6.1 任务清单

| 任务ID | 任务描述 | 对应缺口 | 产出物 | 验收标准 |
|---|---|---|---|---|
| T0.1 | 创建 `proto/social/member.proto` | P0-01 | proto 文件 | 含 21 对 Request/Response + Type(max=1000) 枚举 |
| T0.2 | 创建 `proto/social/group.proto` | P0-01 | proto 文件 | 含 32 对 Request/Response + Type(max=2000) 枚举 |
| T0.3 | 创建 `proto/social/topic.proto` | P0-01 | proto 文件 | 含 21 对 Request/Response + Type(max=3000) 枚举 |
| T0.4 | 统一 Proto 路径引用 | P0-02 | PRD/OpenAPI/注册表/CODE-WIKI | 全部指向实际生成路径 |
| T0.5 | 创建 Go 项目骨架（按域分子包） | P1-01 | `go/modules/social/` 目录 | `go test ./...` 可通过 |
| T0.6 | 创建 SQL 迁移脚本 | P1-02 | `scripts/sql/social-migration-001-initial.sql` + rollback |
| T0.7 | 创建环境配置模板 | P1-03 | `configs/social/social.{local,staging,prod}.conf` |
| T0.8 | 修正 CODE-WIKI §26.4 user_follows 残留 | P1-04 | CODE-WIKI 更新 |
| T0.9 | 清理 OpenAPI 残留 + 补全成员域 | WARN | social-openapi.yaml 更新 |
| T0.10 | **修正 DevGuide P0/P1 问题** | 审阅发现 | DevGuide 更新版（见§6.1 修正清单）|

### 6.2 必须修正的 DevGuide 问题清单

| # | 位置 | 错误值 | 正确值 | 原因 |
|---|---|---|---|---|
| DG-Fix-1 | §3 L143 | `svc_group_user_enter.go # min=2067` | **# min=2087** | 2067=UpdateGroupDiscounts, 2087=GroupUserEnter |
| DG-Fix-2 | §6 L410 | `protocols/generated/go/proto/base` | **`protocols/generated/go/base`**（去掉 `/proto/`） | 与 hello/go.mod 实际路径一致 |
| DG-Fix-3 | §7 L518 | 同上 | 同上 | 同上 |
| DG-Fix-4 | §10 Permission Service | 7 个方法 | **8 个方法**（补充 `CanAuditContent`） | PRD §10.2 定义了第 8 个能力 |
| DG-Fix-5 | §10 方法命名 | CanViewTopicDetail / CanCreateTopic | **CanReadTopic / CanPublishTopic** | 与 PRD §10.2 统一命名 |
| DG-Fix-6 | §12 缓存 Key | `user:profile:{userID}` | **`member:profile:{userID}`** | 与 PRD §11.4 统一前缀 |

### 6.3 Proto 创建详细规格

#### 数据源

从 tabbit inbox 提取 message 定义（不做语义修改，仅调整 package/import）：

| 目标文件 | 数据源 | 提取规则 |
|---|---|---|
| `proto/social/member.proto` | `docs/tabbit/inbox/.../base/user_base.proto` | 全部 Request/Response + Type 枚举 + 辅助类型 |
| `proto/social/group.proto` | `docs/tabbit/inbox/.../base/group_base.proto` | 全部 Request/Response + Type 枚举 + 辅助类型 |
| `proto/social/topic.proto` | `docs/tabbit/inbox/.../base/topic_base.proto` | 全部 Request/Response + Type 枚举 + 辅助类型 |

#### Package 命名（使用项目已验证的路径）

```protobuf
syntax = "proto3";
package cairobot.social.v1;
option go_package = "github.com/jimiechen/mineplanet/protocols/generated/go/base";  // 与 hello/gateway 一致
```

> **注意**：不创建新的 `proto/gen/go/social/v1` 路径。social domain 的 proto 生成后复用已有的 `protocols/generated/go/base` 包路径，通过 package name `cairobot.social.v1` 区分消息归属。

#### Type 枚举要求

每个 Request/Response message 必须包含内嵌 Type 枚举：

```protobuf
message UserRegisterRequest {
    enum Type {
        option allow_alias = true;
        none = 0;
        max = 1000;
        min = 1021;   // 实际编号，非连续
    }
    // ... fields
}
```

### 6.4 SQL 迁移脚本执行顺序

```sql
-- 第 1 批：无外键依赖的基础表
CREATE TABLE IF NOT EXISTS groups ( ... );
CREATE TABLE IF NOT EXISTS topics ( ... );
CREATE TABLE IF NOT EXISTS topic_replies ( ... );

-- 第 2 批：依赖 groups/users 的表
CREATE TABLE IF NOT EXISTS group_members ( ... );       -- FK → groups, users
CREATE TABLE IF NOT EXISTS group_pay_configs ( ... );   -- FK → groups

-- 第 3 批：依赖 group_members/topics 的表
CREATE TABLE IF NOT EXISTS topic_reads ( ... );          -- FK → topics, users
CREATE TABLE IF NOT EXISTS topic_likes ( ... );          -- FK → topics, users
CREATE TABLE IF NOT EXISTS topic_favorites ( ... );      -- FK → topics, users
CREATE TABLE IF NOT EXISTS reply_likes ( ... );          -- FK → topic_replies, users

-- 第 4 批：审计/统计表（可延迟）
CREATE TABLE IF NOT EXISTS audit_logs ( ... );
```

回滚顺序：**第 4 批 → 第 1 批**（反向依赖）。

### 6.5 环境配置模板关键项

```yaml
# configs/social/social.local.conf
social:
  db:
    dsn: "user:pass@tcp(127.0.0.1:3306)/cairobot_social?charset=utf8mb4&parseTime=True"
    max_open_conns: 50
    max_idle_conns: 10
  redis:
    addr: "127.0.0.1:6379"
    db: 2                          # 社交域使用 DB 2
  plaza_group_id: "plaza_global_001"  # 广场虚拟成员群组 ID
  jwt:
    secret: "${SOCIAL_JWT_SECRET}"
    access_token_ttl: 7200
    refresh_token_ttl: 604800
  cache:
    default_ttl: 3600
    stats_ttl: 300
    permissions_ttl: 600
  membership:
    free_group_max_members: 500
    paid_group_max_members: 2000
```

---

## 7. Phase 1：MVP-P0 核心链路（预估 2-3 周）

> **进入条件**：Phase 0 全部验收通过（含 DevGuide 6 项修正）。
> **退出标准**：P0 功能点全部实现并通过测试。

### 7.1 MVP 首期协议白名单（20 对）

> 其余协议标记为"后续迭代"，不在本阶段实现。

| # | 协议对 | minType | 域 | 优先级 | 依赖 |
|---|---|---|---|---|---|
| 1 | UserRegister / Response | 1021/1022 | member | P0 | 无 |
| 2 | UserLogin / Response | 1023/1024 | member | P0 | #1 |
| 3 | GetUserInfo / Response | 1029/1030 | member | P0 | #2 |
| 4 | UpdateUserInfo / Response | 1031/1032 | member | P1 | #2 |
| 5 | CreateGroup / Response | 2005/2006 | group | P0 | #2 |
| 6 | JoinGroup / Response | 2013/2014 | group | P0 | #5 |
| 7 | LeaveGroup / Response | 2015/2016 | group | P1 | #6 |
| 8 | RenewMember / Response | 2037/2038 | group | P0 | #6 |
| 9 | CalcPayableAmount / Response | 2073/2074 | group | P0 | #8 |
| 10 | GroupUserEnter / Response | **2087/2088** | group | P0 | #6 |
| 11 | CreateTopic / Response | 3001/3002 | topic | P0 | #6,#10 |
| 12 | GetTopicList / Response | 3005/3006 | topic | P0 | #11 |
| 13 | DeleteTopic / Response | 3009/3010 | topic | P1 | #11 |
| 14 | AddTopicReply / Response | 3043/3044 | topic | P1 | #12 |
| 15 | LikeTopic / Response | 3061/3062 | topic | P1 | #12 |
| 16 | FavoriteTopic / Response | 3063/3064 | topic | P1 | #12 |
| 17 | GetReplyList / Response | 3065/3066 | topic | P1 | #14 |
| 18 | PinComment / Response | 3081/3082 | topic | P2 | #14 |
| 19 | CreateReport / Response | 3095/3096 | topic | P2 | #12 |
| 20 | CheckTopicActions / Response | 3099/3100 | topic | P0 | #12 |

### 7.2 开发顺序（DevGuide 12 步 + 时间线）

#### 固定的 12 步基础流程（DevGuide §13，不可乱序）

| 步骤 | 操作 | 说明 |
|---|---|---|
| **1** | `make proto` | 重新生成 Go proto 代码（不修改 .proto 文件） |
| **2** | 各域 `model.go` | 按 basemodel.md + PRD 字段设计，不引入 proto 包 |
| **3** | 各域 `repository.go` | DB 操作接口定义 |
| **4** | `permission/service.go` | 8 个权限方法 + 广场虚拟成员规则 |
| **5** | 各域 `servant.go` | TarsGo Servant 注册 + extend minType 提取 |
| **6** | 各域 `handler.go` | minType switch dispatch（含 dispatchProto 泛型辅助函数） |
| **7** | **逐个** `svc_*.go` | 一次只写一个文件，写完跑测试再继续 |
| **8** | 各域 `repository_gorm.go` | GORM 实现 repository 接口 |
| **9** | `module.go` | 聚合注册所有域 Servant |
| **10** | `routes.yaml` | 追加社交域路由 |
| **11** | 各域 `svc_*_test.go` | Mock Repository 隔离数据库 |
| **12** | `make test` | 全量测试 + 报告 |

#### Step 7 的 svc 执行顺序（10 批次，按依赖升序）

| 批次 | 域 | 先做的协议 | 原因 |
|---|---|---|---|
| 1 | member | register → login → logout | 基础认证，无跨域依赖 |
| 2 | member | get_user_info → update_user_info | 用户信息读写 |
| 3 | member | block → unblock → get_block_list | 依赖 member_blocks 表 |
| 4 | group | create_group → group_user_enter | 群组基础创建和进入 |
| 5 | group | join_group → leave_group | 成员关系核心 |
| 6 | group | mute → ban → remove → update_role | 成员治理（依赖 permission） |
| 7 | topic | create_topic → get_topic_list | 内容基础 |
| 8 | topic | add_reply → like → favorite | 互动 |
| 9 | *(Phase 2)* | inbox | query_messages → mark_read | 消息查询 |
| 10 | *(Phase 2)* | third | get_upload_config → create_share | OSS + 分享 |

#### Week 时间线

```
Week 1: 基础设施 + 用户（步骤 1-6 + 批次 1-3）
├─ make proto + model + repo + permission + servant + handler
├─ svc_register/login/logout/get_info/update
└─ svc_block/unblock/get_block_list

Week 2: 群组核心（批次 4-6）
├─ svc_create_group → enter → join → leave
├─ svc_mute/ban/remove/update_role
├─ svc_renew → calc_payable
└─ Cache Aside 层就位

Week 3: 主题 + 集成（批次 7-8 + 步骤 8-12）
├─ svc_create_topic → list → delete
├─ svc_reply → like → favorite → get_reply_list
├─ svc_check_actions（权限查询）
├─ 全量测试 + CI 通过
└─ 文档同步更新
```

### 7.3 svc_*.go 五步固定模式（DevGuide §7）

每个 svc 文件严格遵循以下五步：

```text
Step 1: 参数校验 → common.proto UserErrorCode/GroupErrorCode
Step 2: 权限校验 → permission.Service（不直接查表）
Step 3: 1级数据读写 → MySQL 事务，通过 Repository 接口
Step 4: 发布领域事件 → 2级数据异步更新 Redis/stats
Step 5: 返回 proto Response（字段不可改，错误通过 Result 表达）
```

**规则**: 永远不返回 nil Response；业务错误通过 Result.Code 表达；系统错误才返回 error。

### 7.4 数据等级强制约束表（DevGuide §12）

| 操作 | 数据等级 | 写法规范 | 禁止行为 |
|---|---|---|---|
| 写 user/成员/订单/权益 | **1级** | `repo.Create*()` / `repo.Update*()` in MySQL 事务 | 禁止异步写或跳过事务 |
| 权限判断 | **1级** | `permission.Service.CanXxx()` 查 MySQL | 禁止查 Redis stats 代替 |
| `CanReadTopic` | **1级** | 只查 MySQL group_members | 禁止读 Redis |
| 更新统计数字 | **2级** | 发布领域事件 → 异步更新 | 禁止请求链路同步写 stats |
| 缓存失效 | **2级** | 写 1级成功后 `DEL cache key` | 禁止先删缓存再写 DB |
| 广场普通成员 | 虚拟 | 不写 group_members；由 user.status=active 推导 | 禁止写入 group_members |
| 广场 admin/guest | **1级** | 写 group_members，role=admin/guest | 禁止跳过落库 |

### 7.5 缓存 Key 规范（以 PRD §11.4 为准，统一前缀 `member:` / `group:` / `topic:`）

| 业务操作 | 需要 DEL 的缓存 Key |
|---|---|
| 更新用户信息 | `member:profile:{userID}` |
| 禁用/启用用户 | `member:profile:{userID}` `member:session:{userID}` |
| 加入群组 | `group:member:{groupID}:{userID}` `group:stats:{groupID}` |
| 退出群组 | `group:member:{groupID}:{userID}` `group:stats:{groupID}` |
| 禁言/解禁成员 | `group:member:{groupID}:{userID}` |
| 移除成员 | `group:member:{groupID}:{userID}` `group:stats:{groupID}` |
| 更新群组信息 | `group:detail:{groupID}` `group:list:*` |
| 下架/删除帖子 | `topic:detail:{topicID}` `group:topics:{groupID}:*` |
| 删除评论 | `topic:comments:{topicID}:*` `topic:stats:{topicID}` |
| 确认支付/续费 | `group:member:{groupID}:{userID}` `group:stats:{groupID}` |
| 点赞/取消点赞帖子 | `topic:detail:{topicID}` `topic:stats:{topicID}` |
| 收藏/取消收藏 | `topic:detail:{topicID}` `user:favorites:{userID}` |

### 7.6 Permission Service（8 个方法，统一 PRD 命名）

> 广场特化规则：每个方法内通过 isPlazaGroup() 前置分支处理。

| # | 方法名（PRD 命名） | 广场普通成员行为 | 广场 admin 行为 |
|---|---|---|---|
| 1 | CanViewGroup | active 用户即具备 | 同左 |
| 2 | CanJoinGroup | **false**（不需主动加入） | N/A |
| 3 | **CanReadTopic** | PUBLIC 直接放行；GROUP_MEMBER 走虚拟成员逻辑 | N/A |
| 4 | CanViewTopicSummary | visibility >= 1 即可 | 同左 |
| 5 | **CanPublishTopic** | active 且未被 ban 则可发帖 | N/A |
| 6 | CanManageGroup | false（普通用户不能管理广场） | 查 role=admin/owner |
| 7 | CanManageMember | false | owner > admin > member 层级 |
| 8 | **CanAuditContent**（新增） | false（平台管理员专属） | 查 role=admin/owner |

### 7.7 测试规范（DevGuide §14）

- 每个 `svc_*.go` 对应一个 `svc_*_test.go`
- 使用 Mock Repository 隔离数据库
- 命名格式：`Test{ServiceName}_{场景描述}`（中文场景描述）
- 必须覆盖维度：正常路径 / 参数校验 / 重复性检查 / 权限拒绝 / 广场虚拟成员 / 广场 ban 成员 / block 隔离 / 1级2级分离 / DB 失败 / 并发安全

### 7.8 验收检查矩阵

| 检查项 | 通过标准 | 检查方式 |
|---|---|---|
| Proto 编译 | `make proto` 无错误 | CI auto |
| 单元测试 | `go test ./...` 全部通过 + 覆盖率 >= 80% | CI auto |
| 协议一致性 | 注册表 max+min 唯一性检查通过 | `make proto-check` |
| 缓存铁律 | 权限判断代码审查：无 Redis 读操作 | Code Review |
| 数据等级标注 | 所有 repository 方法注释含 1级/2级 | Code Review |
| 广场特化 | Permission Service 8 方法均含 isPlazaGroup 分支 | 单元测试 |
| SQL 迁移 | migration-001 在空库上幂等执行成功 | 手动验证 |
| OpenAPI | yaml 语法合法 + minType 与注册表一致 | `make rules` |

---

## 8. 禁止行为清单（DevGuide §15，17 条）

| # | 禁止行为 | 违反规则 | 后果 |
|---|---|---|---|
| 1 | 一个 svc 处理多个 minType | Rule 1 | vibecoding，难以 review 和测试 |
| 2 | handler/servant 中写业务逻辑 | 分层约束 | 职责混乱，无法单独测试 |
| 3 | 修改 .proto 字段名/编号/package | Rule 2 | 破坏客户端兼容 |
| 4 | svc 中直接查表做权限判断 | 权限服务约束 | 权限逻辑散落 |
| 5 | CanReadTopic 中读 Redis | 数据等级规范 | 付费权限不准 |
| 6 | 请求链路中同步更新 stats | 数据等级规范 | 延迟增加，事务扩大 |
| 7 | 广场普通成员写入 group_members | 广场虚拟成员规则 | 数据污染 |
| 8 | 广场 admin/guest 不写 group_members | 广场虚拟成员规则 | 无法审计 |
| 9 | model.go 中引入 proto 包 | Rule 3 | model 层依赖 proto |
| 10 | 先删缓存再写 DB | Cache Aside 规范 | 缓存穿透 |
| 11 | mock 跳过核心业务逻辑 | TDD 规范 | 测试无意义 |
| 12 | 没有 make test 通过就宣称完成 | TDD 规范 | 不满足 CI 绿 |
| 13 | main.go 直接实例化 svc | 模块化约束 | 绕过 Handler/Servant |
| 14 | 跳过 make proto 引用旧代码 | proto 生成规范 | 运行时 panic |
| 15 | 用 2 级数据做权限决策 | 数据等级铁律 | 越权访问 |
| 16 | 在广场 CanJoinGroup 返回 true | 广场虚拟成员规则 | 产生无效 JoinGroup 调用 |
| 17 | 缺少 CanAuditContent 就做审核操作 | 权限完整性 | 审核无统一入口 |

---

## 9. Phase 2：MVP-P1 增强功能（预估 1-2 周）

圈主管理成员（禁言/移除/恢复）、评论/点赞/收藏、阅读记录、平台审核、通知设置、IM 签名。
同时创建 `proto/social/third.proto` 和 `proto/social/inbox.proto`。

---

## 10. 风险与缓解措施

| 风险ID | 等级 | 描述 | 缓解措施 |
|---|---|---|---|
| R-001 | **R0** | Proto 文件缺失（P0-01） | Phase 0 T0.1-T0.3 最先完成 |
| R-002 | **R1** | 120 对协议 vs MVP 仅 20 对，范围蔓延 | 维护白名单（§7.1），超出拒绝合入 |
| R-003 | **R1** | DevGuide minType=2067 错误（DG-Fix-1） | 已识别，Phase 0 T0.10 强制修正 |
| R-004 | **R2** | 广场特化增加 PermissionService 复杂度 | Phase 1 Week 1 同步实现 |
| R-005 | **R2** | 1级/2级治理学习成本高 | CODE-WIKI 补充开发 Checklist |
| R-006 | **R2** | group_base.proto 40+ 辅助类型 | Phase 0 创建时按职责拆分 |

---

## 11. 评审要点（请项目主控重点确认）

### R0 — 必须确认

- [ ] **Proto 迁移策略**：tabbit inbox base/*.proto → proto/social/*.proto，package 复用 `protocols/generated/go/base`？
- [ ] **目录结构确认**：采用 DevGuide 的按域分子包（member/group/topic/third/inbox）？
- [ ] **MVP 白名单 20 对**（§7.1）：是否有遗漏或多余？
- [ ] **DevGuide 6 项修正**（§6.2）：是否同意全部修正？

### R1 — 建议讨论

- [ ] Phase 0 工作量 1-2 天是否合理？
- [ ] third/inbox 两域 proto 是否纳入 Phase 0？
- [ ] SQL 迁移工具选择（原生 SQL vs golang-migrate）？
- [ ] Permission Service 命名统一为 PRD 的 CanReadTopic/CanPublishTopic/CanAuditContent？

### R2 — 记录即可

- [ ] CODE-WIKI §26 user_follows 修正
- [ ] OpenAPI 成员域补全
- [ ] PRD §15 验收量化指标补充

---

## 12. 附录

### A. 相关文档索引

| 文档 | 路径 | 用途 |
|---|---|---|
| 主控 DevGuide | `docs/tabbit/inbox/.../social-service-dev-guide.md` | 代码模板/规范/约束（权威源） |
| PRD | `docs/prd/PRD-social-app-mvp.md` | 产品需求 |
| ADR-数据分级 | `docs/adr/ADR-social-data-level-and-cache-strategy.md` | 缓存策略决策 |
| ADR-广场虚拟成员 | `docs/adr/ADR-plaza-virtual-membership.md` | 权限架构决策 |
| OpenAPI | `docs/api/social-openapi.yaml` | 协议元数据 |
| 协议注册表 | `docs/api/协议编号注册表.md` §4 | 编号权威源 |
| CODE-WIKI | `docs/wiki/CODE-WIKI.md` §26 | 开发规范 |
| Proto 数据源 | `docs/tabbit/inbox/2026/06/protocols/base/*.proto` | 原始定义 |
| 评审-数据模型 | `docs/reports/review-social-data-model-20260615.md` | 关注模型评审 |
| 评审-DevGuide+Proto | `docs/reports/review-social-devguide-proto-20260615.md` | Proto 一致性评审 |

### B. minType 速查表（DevGuide 修正后的正确编号）

| 常用协议 | 正确 minType(Request) | 易混淆的错误值 |
|---|---|---|
| UserRegister | **1021** | 1001（旧理想编号） |
| UserLogin | **1023** | 1003 |
| CreateGroup | **2005** | 2001 |
| JoinGroup | **2013** | 2009 或 2011 |
| GroupUserEnter | **2087** | ~~2067~~（DevGuide 原文错误） |
| UpdateGroupDiscounts | **2067** | — |
| RenewMember | **2037** | 2031 或 2030 |
| CalcPayableAmount | **2073** | 2071 |
| CreateTopic | **3001** | 3001（一致） |
| GetTopicList | **3005** | 3003 |
| LikeTopic | **3061** | 3019 |
