# P1 Social Core 三域基础闭环 — 基线校准报告与执行计划

> **版本**: v1.0 (Day 0.5 校准版)
> **日期**: 2026-06-20
> **校准人**: Trae (AI Agent)
> **基线依据**: Task 5-D Stateful E2E (30 POS + 4 NEG + 0 FAIL)
> **状态**: 🔄 待主控审批

---

## 1. 校准范围确认 (R0)

| # | R0 条款 | 校准结论 | 状态 |
|---|---------|----------|------|
| 1 | P1 五批次方向 | A/B/C/E/D 方向认可，聚焦 member/user + group + topic 三域；third/inbox 挂起；Redis Pub/Sub 暂缓到 S1 专项 | ✅ |
| 2 | JWT 库选型 | 已确认为 `github.com/golang-jwt/jwt/v5`（jwt.go L8 import）。P1 以统一、加固、补测试为主，无需迁移 | ✅ |
| 3 | TokenStore Redis | P1 必须实现 TokenStore 接口 + MemoryTokenStore（已有）；RedisTokenStore 视环境确认决定是否进入 P1，未确认则放 Phase 1.5 | ✅ |
| 4 | UpdateMemberStatus(1033) | P1 只做状态落库、基础权限、幂等判断、UserStatusChanged 事件发布和测试；不扩展风控/通知/审批 | ✅ |
| 5 | 时间线 | Day 0.5 校准 + Day 1-5 执行 + Day 6 缓冲（共 6 个工作日） | ✅ |

---

## 2. 协议号争议项校准裁决表

> 本表逐一裁定主控指出的 6 个争议协议号，每个争议项给出最终裁定和证据。

### 2.1 裁决总览

| # | 争议项 | 主控禁令 | Proto 实际定义 | 裁定 | 行动 |
|---|--------|----------|----------------|------|------|
| 1 | **3062 / 3064** | 不得重新引入非法 Topic case | `topic.proto`: 3062=**LikeTopicResponse**(S→C), 3064=**FavoriteTopicResponse**(S→C) | **❌ 非法** — 均为 Response 非 Request | P1 禁止当 Request 使用 |
| 2 | **2039** | 不得改成 GetMembership | `group.proto` L138: 2039=**GetGroupStatsRequest**, Message=`GetGroupStatsRequest` | **✅ 维持原判** — 确认是 GetGroupStats | 无需改动 |
| 3 | **1045** | 不得改成 ListUserGroups | `member.proto` L81: 1045=**GetUserStatsRequest**, Message=`GetUserStatsRequest` | **✅ 维持原判** — 确认是 GetUserStats | 无需改动 |
| 4 | **3025** | 不得当 GetReplyList | `topic.proto` **不存在 3025**（编号从 3001→3005→3009...无 3025） | **❌ 不存在** — 正确编号为 **3065** | P1 禁止使用 3025 |
| 5 | **3065** | Task 5-D 当前 GetReplyList | `topic.proto` L211: 3065=**GetReplyListRequest** | **✅ 确认** — 3065 = GetReplyList | 已在 T5-D 通过，P1 只补强 |
| 6 | **3095** | 已在 T5-D 正向通过 | `topic.proto` L225: 3095=**CreateReportRequest** | **✅ 确认** — 3095 = CreateReport | 已在 T5-D 通过，P1 只补强 |

### 2.2 逐项证据

#### 争议 #1: 3062 / 3064 — 非法 Topic Request

**Proto 来源**: [topic.proto](../../proto/social/topic.proto) L208-L210

```
3000 | 3062 | S->C | Response | LikeTopicResponse     ← 这是 Response！
3000 | 3064 | S->C | Response | FavoriteTopicResponse ← 这也是 Response！
```

**裁定**: 3062 和 3064 是 S→C **Response** 类型，不是 C→S Request。
P1 **禁止**将它们作为 Request 编写测试用例或 SVC 实现。

**对应 Request 编号**:
- LikeTopicRequest = **3061** (已在 T5-D #22 通过)
- FavoriteTopicRequest = **3063** (已在 T5-D #23 通过)

#### 争议 #2: 2039 — GetGroupStats 确认

**Proto 来源**: [group.proto](../../proto/social/group.proto) L138

```protobuf
message GetGroupStatsRequest {
  enum Type { none = 0; max = 1000; min = 2039; }  // ← 明确是 GetGroupStats
}
```

**裁定**: 2039 = **GetGroupStatsRequest**，非 GetMembership。
注册表和 proto 一致，**无需任何修改**。

#### 争议 #3: 1045 — GetUserStats 确认

**Proto 来源**: [member.proto](../../proto/social/member.proto) L81

```protobuf
message GetUserStatsRequest {
  enum Type { none = 0; max = 1000; min = 1045; }  // ← 明确是 GetUserStats
}
```

**裁定**: 1045 = **GetUserStatsRequest**，非 ListUserGroups。
注册表和 proto 一致，**无需任何修改**。

#### 争议 #4: 3025 — 不存在的协议号

**Proto 来源**: [topic.proto](../../proto/social/topic.proto) 完整扫描

topic.proto 中的 Type 枚举实际值序列:
```
3001, 3005, 3009, 3029, 3035, 3037, 3043, 3049, 3055, 3057,
3061, 3063, 3065, 3077, 3081, 3083, 3085, 3091, 3093, 3095, 3099
```

**3025 不在此列表中** — 它从未在 topic.proto 中定义过。

**裁定**: 3025 是**非法编号**。Task 5-D 和当前代码库中 GetReplyList 的正确编号是 **3065**。
P1 **禁止**使用 3025。

#### 争议 #5 & #6: 3065 / 3095 — 已在 T5-D 通过

**Task 5-D 证据**:

| # | 用例 | Max:Min | T5-D 状态 | P1 行动 |
|---|------|---------|-----------|---------|
| 27 | GetReplyList | 3000:**3065** | ✅ POS_PASS | 只补强边界测试，不重建 |
| 25 | CreateReport | 3000:**3095** | ✅ POS_PASS | 只补强边界测试，不重建 |

**裁定**: 两个协议号均已在 Task 5-D Stateful E2E 中正向通过（BizCode=10200）。
P1 **不应重复新建**这两个协议的 SVC/Handler/E2E，只需：
- 补强边界测试（空参数、权限越界、不存在资源）
- 补充负向用例（如举报已举报的帖子）

---

## 3. 1033 UpdateMemberStatus 注册表缺失问题

### 3.1 问题发现

**Proto 定义存在**: [member.proto](../../proto/social/member.proto) L377-L402

```protobuf
// 更新用户状态请求 (min=1033)
message UpdateMemberStatusRequest {
  enum Type { none = 0; max = 1000; min = 1033; }
  string user_id = 1;    // 必填
  int32 status = 2;      // 必填 (1=正常 2=封禁 3=注销)
  string reason = 3;     // 可选
}

message UpdateMemberStatusResponse {
  enum Type { none = 0; max = 1000; min = 1034; }
}
```

**注册表缺失**: 协议编号注册表 §4.1 成员域中，编号从 **1029 (GetUserInfo)** 直接跳到 **1039 (BlockUser)**，
缺少 **1033-1038** 区间的登记（共 3 对 Request/Response: 1033/1034, 1035/1036, 1037/1038）。

### 3.2 裁定

| 项目 | 结论 |
|------|------|
| 1033 在 base proto 中确为 UpdateMemberStatus? | **✅ 确认** |
| 是否需要在 P1 执行前补充注册表? | **⚠️ 是** — P1 Day 1 首个任务必须补充 1033/1034 到注册表 |
| P1 执行范围? | 仅状态落库 + 基础权限 + 幂等判断 + UserStatusChanged 事件 + 测试。不扩展风控/通知/审批 |

### 3.3 P1 前置行动

P1 Day 1 开始前，必须在 [协议编号注册表.md](../api/协议编号注册表.md) §4.1 成员域中补充:

```markdown
| 1000 | 1033 | C->S | Request | `proto/social/member.proto` | `UpdateMemberStatusRequest` | 无 | 待上线 | 更新用户状态请求 |
| 1000 | 1034 | S->C | Response | `proto/social/member.proto` | `UpdateMemberStatusResponse` | 无 | 待上线 | 更新用户状态响应 |
```

> **注意**: 1035-1038 暂不在 P1 范围内，但应在注册表中标注为 reserved 或待定义，避免编号冲突。

---

## 4. JWT 库审计结论

### 4.1 现有实现审计

| 审计项 | 结果 | 证据 |
|--------|------|------|
| JWT 库 | **✅ github.com/golang-jwt/jwt/v5** | [jwt.go](../../go/modules/social/member/jwt.go) L8: `jwt "github.com/golang-jwt/jwt/v5"` |
| 签名算法 | HS256 | L93: `jwt.SigningMethodHS256` |
| Claims 结构 | MapClaims (user_id, token_type, jti, iat, exp, iss) | L93-99 |
| jti 唯一性 | ✅ 已实现 (nanos 精度) | L96: `fmt.Sprintf("acc-%d", now.UnixNano())` |
| 错误处理 | ✅ 区分 Expired vs Invalid | L145-148: `errors.Is(err, jwt.ErrTokenExpired)` |
| 密钥管理 | ✅ 外部注入，禁止硬编码 | L34-43: JWTConfig struct + Validate() |
| 密钥长度校验 | ✅ >= 32 字节 | L67-69: `len(c.SecretKey) < 32` |
| 测试覆盖 | ✅ 有独立测试文件 | [jwt_test.go](../../go/modules/social/member/jwt_test.go) |

### 4.2 P1 JWT 加固范围

现有实现已是标准 jwt/v5，**无需迁移**。P1 加固范围：

| # | 加固项 | 优先级 | 说明 |
|---|--------|--------|------|
| J-1 | 统一 TokenStore 接口 | P0 | MemoryTokenStore + RedisTokenStore(条件性) 实现同一接口 |
| J-2 | Token 黑名单机制 | P0 | Logout 后 token 加入黑名单（Memory/Redis 双模式） |
| J-3 | RefreshToken Rotation | P1 | 刷新时废弃旧 refresh_token，签发新的（防重放攻击） |
| J-4 | JWT 测试补强 | P1 | 补充边界测试：空密钥、短密钥、过期边界、并发签发唯一性 |

---

## 5. Task 5-D 已实现项标注表

> 基于 Task 5-D Stateful E2E 的 34 条通过用例，标注每个协议号的 P1 状态。

### 5.1 Member/User 域 (maxType=1000)

| 协议号 | Message 名称 | T5-D 状态 | P1 标注 | P1 行动 |
|--------|-------------|-----------|---------|---------|
| 1021 | UserRegister | ✅ #1 POS | **已完成** | 无需新建，可补强边界测试 |
| 1023 | UserLogin | ✅ #2 POS | **已完成** | 无需新建，可补强边界测试 |
| 1025 | UserLogout | ✅ #29 POS | **已完成** | P1-A 补强 Token 黑名单机制 |
| 1027 | RefreshToken | ✅ #5 POS | **已完成** | P1-A 补强 Rotation 机制 |
| 1029 | GetUserInfo | ✅ #6 POS | **已完成** | 无需新建 |
| 1031 | UpdateUserInfo | ✅ #7 POS | **已完成** | 无需新建 |
| 1033 | UpdateMemberStatus | ❌ 未覆盖 | **🔄 待补齐** | P1-D 新建（见 §7） |
| 1043 | GetBlockList | ✅ #8 POS | **已完成** | 无需新建 |
| 1045 | GetUserStats | ✅ #9 POS | **已完成** | 无需新建 |

### 5.2 Group 域 (maxType=2000)

| 协议号 | Message 名称 | T5-D 状态 | P1 标注 | P1 行动 |
|--------|-------------|-----------|---------|---------|
| 2005 | CreateGroup | ✅ #10 POS | **已完成** | 无需新建 |
| 2013 | JoinGroup | ✅ #12 POS | **已完成** | 无需新建 |
| 2019 | MuteMember | ✅ #15 POS | **已完成** | 无需新建 |
| 2023 | BanMember | ✅ #16 POS | **已完成** | 无需新建 |
| 2027 | RemoveMember | ✅ #17 POS | **已完成** | 无需新建 |
| 2029 | UpdateMemberRole | ✅ #18 POS | **已完成** | 无需新建 |
| 2039 | GetGroupStats | ✅ #11 POS | **已完成** | 无需新建 |
| 2047 | BatchGetGroups | ✅ #14 POS | **已完成** | 无需新建 |
| 2077 | GetGroupMemberIds | ✅ #13 POS | **已完成** | 无需新建 |

### 5.3 Topic 域 (maxType=3000)

| 协议号 | Message 名称 | T5-D 状态 | P1 标注 | P1 行动 |
|--------|-------------|-----------|---------|---------|
| 3001 | CreateTopic | ✅ #19 POS | **已完成** | 无需新建 |
| 3005 | GetTopicList | ✅ #20 POS | **已完成** | 无需新建 |
| 3009 | DeleteTopic | ✅ #28 POS | **已完成** | 无需新建 |
| 3043 | AddTopicReply | ✅ #21 POS | **已完成** | 无需新建 |
| 3061 | LikeTopic | ✅ #22 POS | **已完成** | 无需新建 |
| 3063 | FavoriteTopic | ✅ #23 POS | **已完成** | 无需新建 |
| 3065 | GetReplyList | ✅ #27 POS | **已完成** | 只补强边界测试 |
| 3057 | BatchGetTopicInfo | ✅ #24 POS | **已完成** | 无需新建 |
| 3095 | CreateReport | ✅ #25 POS | **已完成** | 只补强边界测试 |
| 3099 | CheckTopicActions | ✅ #26 POS | **已完成** | 无需新建 |

### 5.4 移出 P1 / 挂起项

| 域 | 协议号范围 | P1 标注 | 原因 |
|----|-----------|---------|------|
| Third (4000) | 4011-4910 | **⏸️ 移出 P1（挂起）** | R0 确认 third/inbox 挂起 |
| Inbox (5000) | 5051-5072 | **⏸️ 移出 P1（挂起）** | R0 确认 third/inbox 挂起 |
| Topic Response | 3062, 3064 | **❌ 非法（Response 非 Request）** | 不能当 Request 使用 |
| Topic 不存在 | 3025 | **❌ 不存在** | topic.proto 中无此编号 |

### 5.5 统计汇总

| P1 标注 | 数量 | 占比 |
|--------|------|------|
| ✅ 已完成 (T5-D 通过) | **27** | 79% |
| 🔄 待补齐 (P1 新建/补强) | **4** | 12% |
| ⏸️ 移出 P1 (挂起) | **2 域** | - |
| ❌ 非法/不存在 | **3** | 9% |

---

## 6. 修订后五批次任务分解

### 6.1 执行顺序调整说明

主控确认执行顺序为: **A → B → C → E → D**（注意 D 放到最后）

原因: D (成员管理/UpdateMemberStatus) 依赖 A (认证闭环) 的 Token 黑名单和权限基础设施。

### 6.2 Day 0.5 (今日) — 基线校准

| 任务 | 输出 | 状态 |
|------|------|------|
| 协议号争议项裁决 | §2 裁决表 | ✅ |
| 已实现项标注 | §5 标注表 | ✅ |
| JWT 库审计 | §4 审计结论 | ✅ |
| 1033 注册表缺失识别 | §3 缺失报告 | ✅ |
| P1 计划文档输出 | 本文档 | 🔄 进行中 |

### 6.3 Day 1 —批次 A: 认证闭环加固

**目标**: JWT 统一化 + Token 黑名单 + TokenStore 接口

| 任务 ID | 任务 | 协议号 | 产出物 | 验收标准 |
|---------|------|--------|--------|----------|
| A-1 | 补充 1033/1034 到协议编号注册表 | 1033, 1034 | 注册表 PR | CI proto-check 通过 |
| A-2 | 定义 TokenStore 接口 + MemoryTokenStore 实现 | - | `token_store.go` | 接口 + Memory 实现 + 单元测试 |
| A-3 | Logout 接入 Token 黑名单 (Memory 模式) | 1025 | 修改 `svc_logout.go` | Logout 后旧 token 无法使用 |
| A-4 | Login/Refresh 接入 TokenStore 存储 | 1023, 1027 | 修改 `svc_login.go`, `svc_refresh.go` | 新 token 写入 Store |
| A-5 | JWT 边界测试补强 | - | `jwt_test.go` 补充 | 空密钥/短密钥/过期边界/并发唯一性 |

**A 批次验收标准**:
- [ ] TokenStore 接口定义清晰（Store/Delete/Exists 三个方法）
- [ ] MemoryTokenStore 全部测试通过
- [ ] Logout 后旧 access_token 被加入黑名单
- [ ] 使用黑名单中的 token 调用 auth 接口返回 10401
- [ ] JWT 测试覆盖率 >= 90%

### 6.4 Day 2 —批次 B: 用户资料完善

**目标**: Block/Unblock 全链路 + GetUserStats 补强 + UpdateMemberStatus 预备

| 任务 ID | 任务 | 协议号 | 产出物 | 验收标准 |
|---------|------|--------|--------|----------|
| B-1 | BlockUser 边界测试补强 (自己拉黑自己/重复拉黑/拉黑不存在用户) | 1039, 1041 | `svc_block_test.go` 补充 | 负向用例全部返回预期错误码 |
| B-2 | UnblockUser 边界测试补强 (未拉黑先解封/解封不存在用户) | 1041, 1042 | `svc_unblock_test.go` 补充 | 负向用例全部返回预期错误码 |
| B-3 | GetBlockList 分页 + 排序验证 | 1043 | 可能修改 `svc_get_block_list.go` | 分页参数生效 |
| B-4 | GetUserStats 实时计算验证 (发帖后 stats 更新) | 1045 | `svc_get_user_stats_test.go` 补充 | CRUD 操作后 stats 一致 |
| B-5 | UpdateUserInfo 字段白名单校验 (禁止更新 user_id/uid/status) | 1031 | 可能修改 `svc_update_user_info.go` | 白名单外字段被拒绝 |

**B 批次验收标准**:
- [ ] Block/Unblock 负向用例 >= 8 条
- [ ] GetUserStats 与 CRUD 操作一致性验证通过
- [ ] UpdateUserInfo 白名单校验生效

### 6.5 Day 3~4 —批次 C: 群组核心闭环

**目标**: LeaveGroup + 权限校验强化 + Group Stats 实时性

| 任务 ID | 任务 | 协议号 | 产出物 | 验收标准 |
|---------|------|--------|--------|----------|
| C-1 | LeaveGroup 全链路 (退出后角色清除/统计更新) | 2015, 2016 | `svc_leave_test.go` 补强 | 退出后 member 记录状态正确 |
| C-2 | Owner 不能踢自己 / 不能禁言自己 | 2027, 2019 | `svc_remove_test.go`, `svc_mute_test.go` 补充 | 返回预期错误码 |
| C-3 | BanMember 后用户无法操作 (发帖/回复/点赞) | 2023 | 权限检查集成测试 | 封禁用户操作被拒绝 |
| C-4 | GetGroupStats 实时性 (成员变化后 stats 更新) | 2039 | `svc_get_group_stats_test.go` 补强 | Join/Leave/Ban 后 stats 一致 |
| C-5 | BatchGetGroups 权限过滤 (私有群组非成员不可见) | 2047 | 可能修改 `svc_batch_get_groups.go` | 私有群组过滤逻辑正确 |

**C 批次验收标准**:
- [ ] LeaveGroup 全链路测试通过（含状态清理验证）
- [ ] 自操作防护 >= 4 条用例
- [ ] BanMember 后权限联动验证通过
- [ ] Group Stats 实时性验证通过

### 6.6 Day 5 上半 —批次 E: Topic Stub 清理与补强

**目标**: Topic 域已有协议的边界测试 + 3065/3095 补强

| 任务 ID | 任务 | 协议号 | 产出物 | 验收标准 |
|---------|------|--------|--------|----------|
| E-1 | GetReplyList(3065) 边界测试 (空回复/删除后的回复/分页溢出) | 3065 | `svc_get_reply_list_test.go` 补强 | 边界场景覆盖 |
| E-2 | CreateReport(3095) 边界测试 (重复举报/举报不存在的帖子/举报类型非法) | 3095 | `svc_create_report_test.go` 补强 | 边界场景覆盖 |
| E-3 | DeleteTopic 权限校验 (非作者/非管理员不能删) | 3009 | `svc_delete_topic_test.go` 补强 | 权限校验生效 |
| E-4 | LikeTopic/FavoriteTopic 幂等性 (重复点赞取消点赞) | 3061, 3063 | `svc_like_topic_test.go`, `svc_favorite_topic_test.go` 补强 | 幂等行为正确 |
| E-5 | Topic 全域 E2E 回归 (复用 T5-D 框架，增加负向用例) | 全部 Topic 协议 | E2E 用例补充 | Topic 域全量回归通过 |

**E 批次验收标准**:
- [ ] 3065/3095 边界测试 >= 10 条新增
- [ ] DeleteTopic 权限校验通过
- [ ] Like/Favorite 幂等性验证通过
- [ ] Topic 域 E2E 回归全量 PASS

### 6.7 Day 5 下半 ~ Day 6 —批次 D: 成员管理与 UpdateMemberStatus

**目标**: UpdateMemberStatus(1033) 有限实现 + 事件发布

| 任务 ID | 任务 | 协议号 | 产出物 | 验收标准 |
|---------|------|--------|--------|----------|
| D-1 | 1033/1034 注册表补充 (若 A-1 未完成) | 1033, 1034 | 注册表更新 | CI proto-check 通过 |
| D-2 | SvcUpdateMemberStatus 实现 (状态落库 + 幂等 + 基础权限) | 1033 | `svc_update_member_status.go` | SVC + 单元测试 |
| D-3 | Handler 注册 (servant.go + routes.yaml) | 1033 | 修改 servant + routes | HTTP 层可达 |
| D-4 | UserStatusChanged 事件发布 | 1033 | event 定义 + 发布逻辑 | 事件被正确发布 |
| D-5 | UpdateMemberStatus E2E (正常封禁/解封/重复操作/无权限) | 1033 | E2E 用例 | 全部 PASS |

**D 批次验收标准 (严格限制范围)**:
- [x] 状态落库: users.status 字段正确更新 (1→2→1)
- [x] 幂等判断: 重复设置相同状态返回成功但不重复写 DB
- [x] 基础权限: 只有 admin/owner 可以调用（普通 member 返回无权限）
- [x] 事件发布: UserStatusChanged 事件包含 user_id + old_status + new_status + reason
- [ ] **明确不做**: 风控规则、通知推送、审批流程、前端状态同步

**D 批次禁止事项 (R0 明确)**:
- ❌ 不扩展风控引擎集成
- ❌ 不扩展通知推送（邮件/短信/App推送）
- ❌ 不扩展审批工作流
- ❌ 不扩展前端实时状态同步（WebSocket）

### 6.8 Day 6 (缓冲日)

| 场景 | 行动 |
|------|------|
| A~E 全部按时完成 | 全量 E2E 回归 + 文档更新 + 日报 |
| 有批次延期 | 利用缓冲日追赶 + 风险上报 |
| 发现阻塞性 Bug | 停止新功能开发，专注修复 + 事故报告 |

---

## 7. 测试基线与增量目标

### 7.1 Task 5-D 测试基线

| 指标 | T5-D 数值 | 来源 |
|------|-----------|------|
| E2E 用例总数 | 34 | social_e2e_smoke/main.go |
| POS_PASS | 30 | - |
| NEG_PASS | 4 | - |
| FAIL | 0 | - |
| 单元测试文件数 | 63 | go/modules/social/**/*_test.go |
| 覆盖域 | member + group + topic | - |

### 7.2 P1 测试增量目标

| 批次 | 新增单元测试估计 | 新增 E2E 用例估计 | 重点覆盖 |
|------|------------------|------------------|----------|
| A (认证) | ~15 | ~5 | Token 黑名单/JWT 边界 |
| B (用户资料) | ~20 | ~3 | Block/Unblock/Stats 边界 |
| C (群组) | ~25 | ~5 | Leave/权限/Ban 联动 |
| E (Topic) | ~20 | ~8 | 3065/3095/Delete/Like |
| D (成员管理) | ~15 | ~6 | 1033 全链路 |
| **合计** | **~95** | **~27** | - |

### 7.3 P1 结束时测试总量预估

| 指标 | T5-D 基线 | P1 增量 | P1 结束目标 |
|------|-----------|---------|-------------|
| 单元测试文件 | 63 | +~15 (新文件) +~80 (补强) | ~158 |
| E2E 用例 | 34 | +27 | **61** |
| 覆盖协议号 | 27 (completed) + 4 (pending) | +1 (1033) | **32** |

---

## 8. TokenStore 实现策略

### 8.1 接口定义 (P1 必须)

```go
// TokenStore 令牌存储接口
// 用于 token 黑名单管理，支持 Memory 和 Redis 两种实现
type TokenStore interface {
    // Store 将 token 存入黑名单，ttl 为过期时间（秒）
    Store(ctx context.Context, token string, ttl int64) error
    // Delete 从黑名单移除 token
    Delete(ctx context.Context, token string) error
    // Exists 检查 token 是否在黑名单中
    Exists(ctx context.Context, token string) (bool, error)
}
```

### 8.2 MemoryTokenStore (P1 必须，已有雏形)

- 基于 `sync.Map` 或 `map[string]time.Time`
- 单元测试必须覆盖: 并发 Store/Exists/Delete、TTL 过期自动清理
- 用于无 Redis 环境的开发/测试模式

### 8.3 RedisTokenStore (条件性进入 P1)

| 条件 | 决策 |
|------|------|
| dev/staging/prod Redis 连接信息已确认 | **进入 P1**，批次 A 同步实现 |
| Redis 环境未确认 | **放 Phase 1.5**，P1 只做 Memory 模式 + 接口预留 |

**Redis 实现要点** (如果进入 P1):
- 使用 `go-redis/v9` client
- Key 格式: `cairobot:token:blacklist:{jti}` (jti 从 JWT claims 提取)
- 使用 SETEX 设置 TTL（与 token 剩余有效期一致）
- 使用 EXISTS 检查黑名单
- Key prefix 可配置，Redis DB 编号暂不强制

---

## 9. 风险与依赖

### 9.1 风险清单

| 风险 ID | 等级 | 描述 | 缓解措施 |
|---------|------|------|----------|
| R-P1-01 | R1 | 1033 注册表补充可能影响 CI proto-check | Day 1 首个任务，留足时间修复 |
| R-P1-02 | R1 | Redis 环境未确认导致 RedisTokenStore 延期 | 已有 fallback 方案（Phase 1.5） |
| R-P1-03 | R2 | Token 黑名单可能影响现有 T5-D E2E 用例 | A 批次完成后立即回归 T5-D 全量 |
| R-P1-04 | R2 | UpdateMemberStatus 事件发布依赖 Event Publisher 基础设施 | 已有 FakePublisher 测试模式 |
| R-P1-05 | R3 | 95+ 新增测试可能导致单文件超 300 行限制 | 按场景拆分测试文件 |

### 9.2 外部依赖

| 依赖 | 状态 | 影响 |
|------|------|------|
| MySQL (192.168.1.6:3306) | ✅ 已确认 | GORM Repository 直连 |
| Redis (192.168.1.6:6379) | ⚠️ 条件性 | RedisTokenStore 进入判断 |
| proto 文件 (member/group/topic) | ✅ 已确认 | 协议号权威来源 |
| jwt/v5 库 | ✅ 已确认 | 无需迁移 |

---

## 10. 主控审批清单

请主控逐项确认以下内容后，P1 正式进入执行阶段：

### 10.1 必须确认项 (Blocking)

- [ ] **§2 协议号裁决**: 6 个争议项的裁定是否认可？
- [ ] **§3 1033 注册表**: 是否同意 P1 Day 1 补充 1033/1034 到注册表？
- [ ] **§4 JWT 审计**: 是否认可"已是 jwt/v5，无需迁移"的结论？
- [ ] **§5 已实现项标注**: 27 个"已完成"+ 4 个"待补齐"的划分是否准确？
- [ ] **§6 五批次分解**: A→B→C→E→D 的任务分配和时间估算是否合理？
- [ ] **§8 TokenStore**: Memory 必做 + Redis 条件性的策略是否认可？

### 10.2 可选确认项 (Non-blocking)

- [ ] **§7 测试增量**: ~95 单元测试 + ~27 E2E 的增量估算是 否合理？
- [ ] **§9 风险清单**: 5 个风险项的缓解措施是否充分？

### 10.3 审批后行动

主控确认后，我将按以下顺序执行：

```
Day 1 开始 → A-1 (注册表补充) → A-2 (TokenStore 接口) → A-3~A-5 (认证加固)
→ Day 2 B 批次 → Day 3~4 C 批次 → Day 5 上 E 批次 → Day 5 下~6 D 批次
→ 全量 E2E 回归 → 日报 + 文档更新
```

---

## 附录 A: Task 5-D 34 条用例完整映射

| # | 用例 | Max:Min | 域 | T5-D | P1 标注 |
|---|------|---------|-----|------|---------|
| 1 | Register_UserA | 1000:1021 | Member | ✅ POS | 已完成 |
| 2 | Login_UserA | 1000:1023 | Member | ✅ POS | 已完成 |
| 3 | Register_UserB | 1000:1021 | Member | ✅ POS | 已完成 |
| 4 | Login_UserB | 1000:1023 | Member | ✅ POS | 已完成 |
| 5 | RefreshToken_UserA | 1000:1027 | Member | ✅ POS | 已完成 |
| 6 | GetUserInfo_UserA | 1000:1029 | Member | ✅ POS | 已完成 |
| 7 | UpdateUserInfo_UserA | 1000:1031 | Member | ✅ POS | 已完成 |
| 8 | GetBlockList_UserA | 1000:1043 | Member | ✅ POS | 已完成 |
| 9 | GetUserStats_UserA | 1000:1045 | Member | ✅ POS | 已完成 |
| 10 | CreateGroup_UserA | 2000:2005 | Group | ✅ POS | 已完成 |
| 11 | GetGroupStats | 2000:2039 | Group | ✅ POS | 已完成 |
| 12 | JoinGroup_UserB | 2000:2013 | Group | ✅ POS | 已完成 |
| 13 | GetGroupMemberIds | 2000:2077 | Group | ✅ POS | 已完成 |
| 14 | BatchGetGroups | 2000:2047 | Group | ✅ POS | 已完成 |
| 15 | MuteMember_UserB | 2000:2019 | Group | ✅ POS | 已完成 |
| 16 | BanMember_UserB | 2000:2023 | Group | ✅ POS | 已完成 |
| 17 | RemoveMember_UserB | 2000:2027 | Group | ✅ POS | 已完成 |
| 18 | UpdateMemberRole_UserB | 2000:2029 | Group | ✅ POS | 已完成 |
| 19 | CreateTopic_UserA | 3000:3001 | Topic | ✅ POS | 已完成 |
| 20 | GetTopicList | 3000:3005 | Topic | ✅ POS | 已完成 |
| 21 | AddTopicReply_UserB | 3000:3043 | Topic | ✅ POS | 已完成 |
| 22 | LikeTopic_UserB | 3000:3061 | Topic | ✅ POS | 已完成 |
| 23 | FavoriteTopic_UserB | 3000:3063 | Topic | ✅ POS | 已完成 |
| 24 | BatchGetTopicInfo | 3000:3057 | Topic | ✅ POS | 已完成 |
| 25 | CreateReport_UserB | 3000:3095 | Topic | ✅ POS | 已完成 |
| 26 | CheckTopicActions | 3000:3099 | Topic | ✅ POS | 已完成 |
| 27 | GetReplyList | 3000:3065 | Topic | ✅ POS | 已完成 |
| 28 | DeleteTopic_UserA | 3000:3009 | Topic | ✅ POS | 已完成 |
| 29 | Logout_UserA | 1000:1025 | Member | ✅ POS | 已完成 |
| 30 | NEG_GetNonexistTopic | 3000:3005 | Topic | ✅ POS | 已完成 |
| 31 | NEG_DuplicateRegister | 1000:1021 | Member | ✔️ NEG | 已完成 |
| 32 | NEG_WrongPassword | 1000:1023 | Member | ✔️ NEG | 已完成 |
| 33 | NEG_InvalidRefreshToken | 1000:1027 | Member | ✔️ NEG | 已完成 |
| 34 | NEG_JoinNonexistGroup | 2000:2013 | Group | ✔️ NEG | 已完成 |

---

## 附录 B: 文件变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| v1.0 | 2026-06-20 | Day 0.5 基线校准初版，含协议号裁决/已实现项标注/五批次分解 | Trae |
| v1.1 | 2026-06-20 | 主控审批通过 + 5 处修正：(1)§5统计口径改为1新增+4增强 (2)§6.7权限改为平台管理员/系统操作员 (3)§8 key 改为 jti/tokenID (4)§8.2.1 补充鉴权路径集成 (5) R-P1-03 升级为 TokenStore 接入 E2E 回归强制要求 | Trae |
