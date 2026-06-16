# ADR: 大群组广场虚拟成员关系

## 状态

Proposed（提议中）

## 背景

CaiRobot MVP 社交域需要一个大群组"广场"，所有注册用户默认是广场的普通成员。如果为每个用户都写入 group_members 表，当用户量达到百万级时，group_members 表将膨胀且大部分记录无实际业务价值（仅表示"是普通成员"这一默认状态）。

**问题本质**：
- 广场成员关系是"默认状态"，而非"特殊状态"
- 为默认状态维护物理记录会造成数据冗余和存储浪费
- 百万级用户的 group_members 记录会严重影响查询性能和维护成本

## 决策

采用**虚拟成员关系**设计：

### 1. 核心设计原则

- **广场群组**是一个全局默认大群组（groups 表中 type=plaza 的记录）
- **所有注册且 status=active 的用户默认是广场普通成员**——这种关系不写入 group_members 表
- **仅特殊身份写入 group_members**：
  - 广场嘉宾（role=guest）：需要明确邀请、授权、过期时间
  - 广场管理员（role=admin）：需要后台可审计、可撤销
  - 广场所有人（role=owner）：系统初始化或平台超级管理员配置
- **block 关系独立存储在 member_blocks 表**，与是否落库 group_members 无关
- **权限判断时**，对广场群组做特殊处理：不查 group_members 判断普通成员身份，而是通过 `users.status == active` 虚拟推导

### 2. 对 Permission Service 的 7 个能力方法的影响

社交域 PermissionService 定义了 7 个方法：

| 方法 | 说明 |
|------|------|
| CanViewGroup(userID, groupID) | 判断用户是否可查看群组 |
| CanJoinGroup(userID, groupID) | 判断用户是否可加入群组 |
| CanReadTopic(userID, topicID) | 判断用户是否可阅读话题 |
| CanManageGroup(operatorID, groupID) | 判断操作者是否可管理群组 |
| CanManageMember(operatorID, groupID, targetUserID) | 判断操作者是否可管理目标成员 |
| CanPublishTopic(userID, groupID) | 判断用户是否可在群组发帖 |
| CanAuditContent(operatorID) | 判断操作者是否可审核内容 |

#### 广场特化补充规则（不是替换）

- **CanViewGroup(plaza)**：如果 users.status==active → ALLOW（不需要查 group_members）
- **CanJoinGroup(plaza)**：注册即加入，无需写 group_members
- **CanReadTopic(plaza topic)**：active 用户视为 GROUP_MEMBER 级别
- **CanManageGroup(plaza)**：只有 owner/admin role 在 group_members 中的人才允许
- **CanPublishTopic(plaza)**：active 用户 + 未被禁言即可发帖
- **CanManageMember(plaza)**：同普通群组逻辑，但目标用户的"普通成员"身份是虚拟的
- **CanAuditContent**：不受影响

## 后果

### 正面影响

1. **存储优化**：group_members 表不会因广场普通成员而膨胀
2. **注册流程简化**：新用户注册无需写 group_members
3. **状态联动**：用户被禁用时，自动失去广场成员身份（只需改 users.status）
4. **查询性能提升**：权限判断时减少一次 group_members 表查询

### 负面影响

1. **逻辑复杂度增加**：权限判断逻辑需要区分广场群组和普通群组
2. **统计逻辑变更**：统计"广场成员数"不能直接 COUNT(group_members)，需 COUNT(users WHERE status=active)
3. **常量维护**：需要在代码中维护一个 plaza_group_id 常量
4. **代码分支增多**：PermissionService 需要增加广场特化分支

### 风险点

1. **扩展性风险**：如果未来需要对广场普通成员做精细化管理（如等级、积分），虚拟模型可能不够用
2. **缓存适配风险**：缓存层需要适配：不能缓存"广场成员关系"，只能缓存 block 关系
3. **一致性风险**：虚拟关系与物理关系混合可能导致边界情况处理不当

## 替代方案

### 方案A（已选）：虚拟成员，不落库

**优点**：
- 存储效率高
- 查询性能好
- 注册流程简单

**缺点**：
- 逻辑复杂度高
- 需要特殊处理广场场景

**适用场景**：百万级用户、广场为主要使用场景

### 方案B：全量落库，所有用户都写 group_members

**优点**：
- 实现简单统一
- 无需特殊逻辑
- 统计方便

**缺点**：
- 数据量大（百万级记录）
- 存储成本高
- 注册流程需额外写入
- 大部分记录无实际业务价值

**适用场景**：用户量较小（<10万）、开发资源有限

### 方案C：异步同步，用户注册后异步写入 group_members

**优点**：
- 最终一致性保证
- 不阻塞注册流程
- 逻辑相对简单

**缺点**：
- 存在延迟窗口（注册后短时间内无法查到成员关系）
- 需要额外的异步任务基础设施
- 数据量大（同方案B）

**适用场景**：对实时性要求不高、已有异步任务框架

## 相关文档

- PRD: 社交域大群组功能需求
- ADR: 社交域权限模型设计
- API 文档: groups 表结构定义
- API 文档: group_members 表结构定义

## 实施建议

1. **第一阶段**：实现基础虚拟成员关系，支持 7 个 PermissionService 方法的广场特化
2. **第二阶段**：优化统计查询，支持高效的广场成员数计算
3. **第三阶段**：根据实际业务需求评估是否需要引入广场成员等级等扩展功能
