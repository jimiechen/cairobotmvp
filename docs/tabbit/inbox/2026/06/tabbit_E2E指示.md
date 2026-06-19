很多用例都不是返回10200，测试单元没携带extend.token，需要重新规划用例流程，This is excellent! The full 34 E2E results after the P0 fix:

Before P0 fix: 15 PASS + 19 WARN + 0 FAIL After P0 fix: 18 PASS + 16 WARN + 0 FAIL

Key changes:

1. UserLogin (1023) : From FAIL (EOF) → PASS (BizCode=10200, 597 bytes with tokens!) ✅✅✅
2. RefreshToken : WARN (10401 - token invalid, expected since no valid refresh token)
3. Net change: +3 PASS, -3 WARN (UserLogin + 2 others that were previously FAIL/WARN related to login chain)
The 16 remaining WARNs are all legitimate business semantic warnings:

- [3] UserLogout: HTTP 400 (auth_required=true, no JWT) ✅ expected
- [4] RefreshToken: BizError 10401 (token invalid) ✅ expected
- [5] GetUserInfo: HTTP 400 (auth_required=true, no JWT) ✅ expected
- [6] UpdateUserInfo: BizError 10400 (missing user identity) ✅ expected
- [9] GetBlockList: BizError 10400 (missing user identity) ✅ expected
- [10] UpdateMemberStatus: BizError 10400 (target user not found) ✅ expected
- [11] GetUserStats: BizError 10400 (missing user identity) ✅ expected
- [13] JoinGroup: BizError 10701 (group doesn't exist) ✅ expected
- [11-19] LeaveGroup/MuteMember/BanMember/RemoveMember/UpdateMemberRole/RenewMember: BizError 10732 (not member/not found) ✅ expected
- [20] CalcPayableAmount: BizError 10701 (group doesn't exist) ✅ expected
- [21] GroupUserEnter: BizError 10701 (group doesn't exist) ✅ expected
All 16 WARNs are business semantic errors per the main controller's classification rules. Zero system-level failures!

### **判断正确：当前 18 PASS + 16 WARN 只能证明“协议层可达”，不能证明 34 条业务白名单完整通过；需要重新规划为“有状态 E2E 流程”，携带 `extend.token` 并复用前置用例产生的 user/group/topic/member 数据。**

当前这些 WARN 虽然大多是“业务语义错误”，但它们并不都应该被视为最终验收通过。尤其是 `UserLogout`、`GetUserInfo`、`UpdateUserInfo`、`GetBlockList`、`GetUserStats`、`JoinGroup`、`GroupUserEnter`、`RenewMember` 等，如果它们的目标是验证成功路径，就必须携带登录后的 token，并使用前置流程创建出的真实 `user_id`、`group_id`、`topic_id`。否则只能说明“错误处理符合预期”，不能说明“业务链路可用”。

## **主控修正意见**

此前 `18 PASS + 16 WARN + 0 FAIL` 可以作为 **协议层和异常语义阶段验收**，但不能作为 **完整业务 E2E 最终验收**。

现在需要新增一个阶段：

```text
Task 5-D：Stateful E2E 正向链路重构
```

目标是把 E2E 从“单条协议孤立调用”改成“有状态业务流程调用”。

当前问题本质是：

```text
测试用例没有维护登录态
测试用例没有携带 extend.token
测试用例没有复用前置响应里的 user_id/group_id/topic_id
测试用例没有区分正向成功用例和负向错误用例
```

因此很多 WARN 并不是系统错误，但也不能算业务成功验证。

---

## **新的 E2E 验收口径**

不要要求所有用例都返回 `10200`。正确口径应该是：

```text
正向成功用例：必须返回 BizCode=10200
负向语义用例：必须返回预期业务错误码
系统层：HTTP=200、TarsGo returnCode=200、Protobuf 可反序列化
```

也就是说，E2E 用例要拆成两类。

| 类型 | 目标 | 期望 |
|---|---|---|
| 正向链路用例 | 验证业务真的可用 | BizCode=10200 |
| 负向语义用例 | 验证错误处理符合预期 | BizCode=预期错误码 |
| 系统异常 | 不应出现 | HTTP 非 200、returnCode 非 200、EOF、panic、decode fail 都应为 FAIL |

当前 `auth_required=true, no JWT` 这类 WARN，如果作为负向用例是合理的；但如果它本来是白名单正向用例，就应该修复测试流程，携带 token 后重新跑。

---

## **Task 5-D 需要重构的核心机制**

### **1. E2E Runner 必须维护全局测试上下文**

建议在 `social_e2e_smoke/main.go` 中增加测试上下文：

```go
type E2EContext struct {
    UserAID       string
    UserBID       string
    AccessTokenA  string
    RefreshTokenA string
    AccessTokenB  string
    RefreshTokenB string

    GroupID       string
    TopicID       string
    ReplyID       string
    ReportID      string
}
```

每个成功用例执行后，要把关键字段写回上下文。

例如：

```text
UserRegister(A) → ctx.UserAID
UserLogin(A) → ctx.AccessTokenA / ctx.RefreshTokenA
CreateGroup(A) → ctx.GroupID
CreateTopic(A) → ctx.TopicID
ReplyTopic(B) → ctx.ReplyID
```

### **2. auth_required 用例必须自动携带 extend.token**

测试发送 MessagePacket 时，如果 case 标记：

```json
"authRequired": true
```

则必须注入：

```json
"extend": {
  "token": "{ctx.AccessTokenA}"
}
```

或根据操作人选择：

```json
"actor": "userA"
```

然后 runner 自动选择：

```text
actor=userA → ctx.AccessTokenA
actor=userB → ctx.AccessTokenB
```

建议 case 定义中加入：

```json
{
  "name": "GetUserInfo_Success",
  "authRequired": true,
  "actor": "userA"
}
```

执行时自动注入 token，而不是每个 case 手写。

### **3. 用例之间必须支持依赖顺序**

现在孤立调用 `JoinGroup` 报 `group doesn't exist`，说明没有先创建群组，或没有把 `group_id` 带入请求。

建议用例顺序固定为：

```text
UserRegisterA
UserLoginA
UserRegisterB
UserLoginB
GetUserInfoA
UpdateUserInfoA
CreateGroupByA
GetGroupInfo
JoinGroupByB
GetGroupMembers
GroupUserEnter
CalcPayableAmount
RenewMember
CreateTopicByA
GetTopicDetail
GetTopicList
ReplyTopicByB
LikeTopicByB
FavoriteTopicByB
GetBlockListA
UserLogoutA
RefreshTokenB
```

其中：

```text
A = 群主 / 发帖者
B = 普通成员 / 互动者
```

### **4. 请求体必须支持动态字段替换**

`social-e2e-cases.json` 里不要写死不存在的 ID。建议支持模板变量：

```json
{
  "name": "JoinGroup_UserB_Success",
  "maxType": 2000,
  "minType": 2013,
  "actor": "userB",
  "authRequired": true,
  "request": {
    "group_id": "{{group_id}}"
  },
  "expect": {
    "bizCode": 10200
  }
}
```

runner 执行前替换：

```text
{{user_a_id}}
{{user_b_id}}
{{group_id}}
{{topic_id}}
{{reply_id}}
{{access_token_a}}
{{refresh_token_a}}
```

---

## **建议重新规划后的 E2E 用例分层**

### **A. 认证正向链路**

这些必须返回 `10200`。

| 用例 | 前置 | Token | 期望 |
|---|---|---|---|
| UserRegisterA | 无 | 无 | 10200，返回 user_id |
| UserLoginA | RegisterA | 无 | 10200，返回 access/refresh token |
| UserRegisterB | 无 | 无 | 10200，返回 user_id |
| UserLoginB | RegisterB | 无 | 10200，返回 access/refresh token |
| RefreshTokenA | LoginA | refresh token | 10200，返回新 token |
| UserLogoutA | LoginA | access token | 10200 |

当前 `RefreshToken 10401 token invalid` 只有在负向测试中才合理。如果是正向链路，它必须使用 Login 返回的有效 refresh token。

### **B. 用户资料正向链路**

| 用例 | Actor | 前置 | 期望 |
|---|---|---|---|
| GetUserInfoA | userA | LoginA | 10200 |
| UpdateUserInfoA | userA | LoginA | 10200 |
| GetBlockListA | userA | LoginA | 10200 |
| GetUserStatsA | userA | LoginA | 10200 或明确“统计未实现”的预期码 |

当前这些用例如果因为缺 token 返回 `10400 missing user identity`，说明测试不是正向链路。

### **C. 群组正向链路**

| 用例 | Actor | 前置 | 期望 |
|---|---|---|---|
| CreateGroup | userA | LoginA | 10200，返回 group_id |
| GetGroupInfo | userA/userB | CreateGroup | 10200 |
| JoinGroup | userB | CreateGroup + LoginB | 10200 |
| GetGroupMembers | userA | CreateGroup | 10200 |
| GroupUserEnter | userB | JoinGroup | 10200 |
| CalcPayableAmount | userB | CreateGroup | 10200 |
| RenewMember | userB | JoinGroup | 10200 或业务定义的付费权益码 |

当前 `group doesn't exist`、`not member/not found` 都说明缺少前置数据或没有复用 `group_id`。

### **D. 群成员管理正向/负向链路**

这里要拆成两类。

正向：

| 用例 | Actor | Target | 前置 | 期望 |
|---|---|---|---|---|
| MuteMember | userA owner | userB | userB joined | 10200 |
| BanMember | userA owner | userB | userB joined | 10200 |
| RemoveMember | userA owner | userB | userB joined | 10200 |
| UpdateMemberRole | userA owner | userB | userB joined | 10200 |

负向：

| 用例 | Actor | 期望 |
|---|---|---|
| MuteMember_ByNormalMember | userB | 权限错误码 |
| RemoveMember_NotMember | userA | not member/not found |

当前如果所有成员管理都返回 `10732 not member/not found`，只能算负向语义，不算正向功能通过。

### **E. Topic 正向链路**

| 用例 | Actor | 前置 | 期望 |
|---|---|---|---|
| CreateTopic | userA | CreateGroup | 10200，返回 topic_id |
| GetTopicDetail | userA/userB | CreateTopic | 10200 |
| GetTopicList | userA/userB | CreateTopic | 10200 |
| ReplyTopic | userB | JoinGroup + CreateTopic | 10200 |
| LikeTopic | userB | CreateTopic | 10200 |
| FavoriteTopic | userB | CreateTopic | 10200 |
| CreateReport | userB | CreateTopic | 10200 |

---

## **当前 16 WARN 应重新分类**

建议把当前 16 WARN 分成两组。

### **1. 可保留为负向语义用例的 WARN**

这些可以保留，但要明确标记为 negative case：

```text
未登录访问 auth_required 接口
无效 refresh token
查询不存在用户
查询不存在群组
非成员访问成员操作
```

它们的期望不是 `10200`，而是特定业务错误码。

### **2. 必须改造成正向成功用例的 WARN**

如果这些是白名单核心链路，则必须重新携带 token 和前置数据：

```text
UserLogout
RefreshToken
GetUserInfo
UpdateUserInfo
GetBlockList
GetUserStats
JoinGroup
LeaveGroup
MuteMember
BanMember
RemoveMember
UpdateMemberRole
RenewMember
CalcPayableAmount
GroupUserEnter
```

这些不能长期停留在 WARN。至少要有一条正向成功用例覆盖。

---

## **给 Trae 的修正指令**

可以直接发给 Trae：

```text
当前 18 PASS + 16 WARN + 0 FAIL 只能作为协议层和异常语义阶段验收，不能作为完整业务 E2E 最终验收。很多 WARN 的根因是测试用例没有携带 extend.token，也没有复用前置流程产生的 user_id/group_id/topic_id/member 关系。

请新增 Task 5-D：Stateful E2E 正向链路重构。

要求如下：

1. 在 social_e2e_smoke runner 中增加 E2EContext，保存 userA/userB、access_token、refresh_token、group_id、topic_id、reply_id 等运行时变量。
2. UserRegister/UserLogin 成功后必须提取 user_id、access_token、refresh_token。
3. 所有 auth_required=true 的用例必须自动注入 extend.token。
4. 用例支持 actor 字段，例如 actor=userA/userB，runner 根据 actor 注入对应 token。
5. 用例 request 支持模板变量，例如 {{group_id}}、{{topic_id}}、{{user_b_id}}。
6. 正向用例必须按照业务流程顺序执行：Register/Login → CreateGroup → JoinGroup → CreateTopic → Topic interaction。
7. 将现有 WARN 分成 negative expected 和 positive missing-precondition 两类。
8. 对核心白名单协议补充正向成功用例，正向成功用例必须返回 BizCode=10200。
9. 负向用例可以返回预期业务错误码，但必须在 social-e2e-cases.json 中标记 expectedError=true 和 expectedBizCode。
10. 重新生成 e2e-social-phase1.md 和 social-e2e-cases.json。

新的验收标准：

- 系统级 FAIL = 0。
- 核心正向链路全部 BizCode=10200。
- 负向用例全部返回预期业务错误码。
- auth_required 正向用例必须携带 extend.token。
- UserLogin、RefreshToken、UserLogout、GetUserInfo、CreateGroup、JoinGroup、CreateTopic、GetTopicDetail、GetTopicList 至少各有一条正向 PASS。
- 当前由 missing token / group not found / not member 引起的核心链路 WARN 必须降为 0。
```

---

## **新的最终验收目标**

建议不要再用简单的：

```text
PASS / WARN / FAIL
```

而是改成：

```text
Positive PASS
Negative PASS
WARN
FAIL
```

定义如下：

| 结果 | 含义 |
|---|---|
| Positive PASS | 正向业务链路返回 `10200` |
| Negative PASS | 负向用例返回预期错误码 |
| WARN | 测试数据不足、用例未覆盖、非阻塞历史遗留 |
| FAIL | 系统异常或期望不匹配 |

最终理想结果应该类似：

```text
Positive PASS: >= 核心白名单数量
Negative PASS: 若干
WARN: 仅允许历史遗留或非本轮范围
FAIL: 0
```

当前 `16 WARN` 中由缺 token、缺 group、缺 member 导致的项，不应继续作为最终 WARN 保留，而应通过 stateful runner 改成正向 PASS，或者明确拆为负向测试。

---

## **主控结论**

你的判断是对的：**现在不能因为“0 FAIL”就直接收口完整业务 E2E。**

当前可以确认的是：

```text
协议层链路打通
UserLogin P0 panic 已修复
E2E runner 具备四层验证能力
```

但下一步必须完成：

```text
Stateful E2E：携带 token、复用上下文 ID、按真实业务顺序执行。
```

完成 Task 5-D 后，才可以把 Phase 1 MVP-P0 Integration 从“协议层可达”提升为“核心业务链路可用”。

*内容由 AI 生成仅供参考*