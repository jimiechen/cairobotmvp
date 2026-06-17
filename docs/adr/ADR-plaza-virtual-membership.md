# ADR-0014: 广场虚拟成员机制（Plaza Virtual Membership）

## 状态

已采纳（2026-06-15）

## 背景

CaiRobot Social App 的核心产品形态是**圈子（群组）社交**。产品中存在一个特殊的"广场"概念——一个全局可见的默认浏览入口，所有活跃用户都可以浏览和参与广场中的内容。

### 问题

如果将广场建模为一个普通群组（`groups` 表中的一行 + `group_members` 中的成员关系），会带来以下问题：

1. **数据膨胀**：假设平台有 10 万活跃用户，每个用户都需要一条 `group_members` 记录来表示"我是广场的虚拟成员"
2. **写入放大**：每次用户注册/激活都需要 INSERT 一条广场成员记录
3. **维护成本**：用户封禁/注销时需要同步清理广场成员记录
4. **语义混淆**：广场不是真正的"群组"，它是一个全局浏览上下文，不应与付费群组混为一谈

### 约束条件

- 权限判断铁律：**只能使用 1 级数据做权限决策**，不能依赖 Redis 缓存
- Permission Service 是社交域所有权限判断的唯一入口
- 广场群组的 ID 从配置中心读取（`social.plaza_group_id`），不硬编码

## 决策

采用**广场虚拟成员（Plaza Virtual Membership）**机制：

> **核心思想**：广场群组中，普通用户的"成员身份"是**虚拟推导**的——只要用户 status=active 即视为广场普通成员，不需要 `group_members` 表中的物理记录。管理员/嘉宾角色仍需物理记录。

### 特化规则

```go
// isPlazaGroup 判断是否为广场群组（从配置中心读取）
func (s *permissionService) isPlazaGroup(groupID string) bool {
    return groupID == s.plazaGroupID
}

// 广场特化规则:
//   isPlazaGroup == true:
//     普通成员  →  active 用户即具备，不查 group_members
//     管理员/嘉宾  →  从 group_members 查 role = admin/guest
//     被封禁用户  →  从 group_members 查 role = banned（有记录则拒绝）
//   非广场群组:  所有判断走 group_members 表
```

### 7 个能力方法的广场特化行为

| 能力方法 | 广场普通成员行为 | 广场管理员/嘉宾行为 | 非广场行为 |
|---|---|---|---|
| `CanViewGroup` | active 用户 = 有权 | 同左（管理员也是成员） | 查 group_members |
| `CanJoinGroup` | 返回 false（无需主动加入） | N/A | 查 group_members |
| `CanViewTopicDetail` | PUBLIC 帖子直接放行；GROUP_MEMBER 帖子走虚拟成员逻辑 | 同普通成员 | 查 group_members |
| `CanViewTopicSummary` | visibility >= 1 即可 | 同左 | visibility >= 1 |
| `CanCreateTopic` | active 且未被 ban 则可发帖 | 同左 | 查 group_members status=1 |
| `CanManageGroup` | false（普通用户不能管理广场） | 查 group_members role=admin/owner | 查 group_members |
| `CanManageMember` | false | owner > admin > member 层级判断 | 查 group_members |

## 数据等级影响

| 数据项 | 广场场景 | 数据等级 | 说明 |
|---|---|---|---|
| 用户 status | `users.status == 1` | **1 级** | 必须查 MySQL，禁止用缓存 |
| 广场 ban 记录 | `group_members.role = banned` | **1 级** | ban 是物理记录，需查 MySQL |
| 广场 admin/guest | `group_members.role = admin\|guest` | **1 级** | 物理记录，需查 MySQL |
| 广场成员列表 | 推导：active 用户 - banned 用户 | **2 级** | 可由事件驱动异步计算 |
| 广场在线人数 | Redis Counter | **2 级** | 事件驱动更新 |

## 禁止行为

以下行为在广场场景中被**明确禁止**：

| # | 禁止行为 | 原因 | 后果 |
|---|---|---|---|
| 1 | 向 `group_members` 写入广场普通成员记录 | 违反虚拟成员设计 | 数据膨胀、写入放大 |
| 2 | 用 Redis 缓存的 user_status 判断广场权限 | 违反 1 级数据铁律 | 缓存不一致时越权 |
| 3 | 在 svc_*.go 中硬编码 plaza_group_id | 违反配置化原则 | 无法动态切换广场 |
| 4 | 广场 CanJoinGroup 返回 true | 广场不需要主动加入 | 产生无效 JoinGroup 调用 |

## 与其他模块的关系

### 与 groups 表的关系

广场在 `groups` 表中有物理行记录（`is_plaza=true` 或类似标记），但**不在 `group_members` 中为每个用户创建记录**。

### 与 Permission Service 的关系

广场特化逻辑**内嵌在 Permission Service 的 7 个方法内部**，通过 `isPlazaGroup()` 前置判断分流。调用方（svc_*.go）无需感知广场特殊逻辑。

### 与 Cache Aside 策略的关系

- 广场权限判断**不走缓存**（1 级数据铁律）
- 广场成员列表/统计等 2 级数据仍使用事件驱动 + TTL 回退策略（参见 ADR-social-data-level-and-cache-strategy）

### 与 Gateway/routes.yaml 的关系

广场相关的请求（如 `GroupUserEnter` minType=2087）与其他群组请求走相同的路由路径（maxType=2000 → SocialServant），由 Permission Service 内部区分处理。

## 实现要点

### 配置项

```yaml
# configs/social/social.local.conf
social:
  plaza_group_id: "plaza_global_001"  # 广场群组 ID，从配置中心读取
```

### 测试维度

| 测试场景 | 输入 | 期望输出 |
|---|---|---|
| 广场 active 用户查看广场 | userID=active_user, groupID=plaza | true |
| 广场 inactive 用户查看广场 | userID=inactive, groupID=plaza | false |
| 广场 banned 用户发帖 | userID=banned, groupID=plaza | false（查到 ban 记录） |
| 广场 admin 管理成员 | userID=admin, target=member, groupID=plaza | true |
| 非广场群组正常流程 | userID=user, groupID=normal_group | 走标准 group_members 查询 |
| 广场 JoinGroup | userID=any, groupID=plaza | false（不允许主动加入） |

## 替代方案

### 方案 A：广场作为真实群组（已否决）

所有用户注册时自动 INSERT 一条 `group_members(plaza_id, user_id)` 记录。

**否决原因**：10 万用户 = 10 万条冗余记录；注册/封禁/注销都需要同步维护。

### 方案 B：Redis Set 维护广场成员集合（已否决）

用 `SADD plaza:members {user_id}` 维护成员集合，权限判断走 Redis。

**否决原因**：违反 1 级数据铁律；Redis 故障时无法做权限判断；缓存一致性问题。

### 方案 C：虚拟成员（已采纳）

当前方案。零存储开销，权限判断基于 users.status 的 1 级数据查询。

## 后续演进

1. **P1**：广场支持多级虚拟身份（如"认证作者"需要额外审核记录）
2. **P1**：广场内容推荐算法（基于虚拟成员的阅读行为）
3. **P2**：多广场支持（不同分类的独立广场，各自独立的虚拟成员规则）

---

## 相关文档

- [PRD-social-app-mvp.md](../prd/PRD-social-app-mvp.md)：产品需求文档
- [ADR-social-data-level-and-cache-strategy.md](ADR-social-data-level-and-cache-strategy.md)：数据分级与缓存策略
- [social-service-dev-guide.md](../tabbit/inbox/2026/06/protocols/social-service-dev-guide.md) §10：Permission Service 完整实现
- [协议编号注册表.md](../../api/协议编号注册表.md)：协议编号登记（含 GroupUserEnter 2087/2088）
