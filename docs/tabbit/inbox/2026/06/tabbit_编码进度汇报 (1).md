在对话模式下执行。

收到，在对话模式下补充，不执行文件修改。

建议将这条作为 **Phase 1 MVP-P0 后续编码规范补充项** 加入主控决策/代码规范中：

### **补充编码规范：禁止硬编码枚举值与魔法字符串**

从 Step 8 起，svc 层、permission 层、repository 层新增以下强制规范：

#### **1. Protobuf 协议枚举必须直接使用 pb 枚举常量**

svc 层凡是涉及协议中已定义的枚举值，不允许写硬编码整数，例如 `1`、`2`、`3`。必须使用 protobuf 生成代码中的枚举常量，保证业务语义和协议定义保持一致。

错误示例：

```go
member.Role = 1
group.Status = 1
topic.Status = 1
topic.Visibility = 2
```

正确示例：

```go
member.Role = int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_OWNER)
group.Status = int8(pb.GroupStatus_GROUP_STATUS_ACTIVE)
topic.Status = int8(pb.TopicStatus_TOPIC_STATUS_PUBLISHED)
topic.Visibility = int8(pb.TopicVisibility_TOPIC_VISIBILITY_PUBLIC)
```

如果 model 层字段当前使用 `int8`、`int32` 等基础类型，而 proto 枚举是独立类型，svc 层允许在赋值边界做显式类型转换，但枚举来源必须是 pb 枚举常量，不允许直接写数字。

#### **2. 业务字符串必须定义为常量，禁止散落魔法字符串**

svc、permission、converter、repository 中出现的业务字符串，例如角色名、上下文 key、错误消息片段、状态字符串、默认配置名等，不允许直接散落在代码中。应统一定义为包级常量。

错误示例：

```go
if role == "owner" || role == "admin" {
    return true
}

ctx.Value("user_id")
```

正确示例：

```go
const (
    RoleOwner  = "owner"
    RoleAdmin  = "admin"
    RoleMember = "member"

    CtxKeyUserID = "user_id"
)

if role == RoleOwner || role == RoleAdmin {
    return true
}

ctx.Value(CtxKeyUserID)
```

如果字符串来自 protobuf 枚举语义，优先使用 proto 枚举，而不是再定义字符串常量。只有在当前模型确实需要字符串表达时，才允许用本地常量承接。

#### **3. svc 层禁止出现裸字面量状态值**

以下类型的裸值均视为不合规：

```go
Status: 1
Role: 2
Visibility: 3
Type: 1
Mode: "public"
Role: "admin"
```

应替换为：

```go
Status: int8(pb.GroupStatus_GROUP_STATUS_ACTIVE)
Role: int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_ADMIN)
Visibility: int8(pb.TopicVisibility_TOPIC_VISIBILITY_PUBLIC)
```

或：

```go
const (
    DefaultGroupModePublic = "public"
    PermissionRoleAdmin    = "admin"
)
```

#### **4. converter 层负责枚举转换，不允许各 svc 重复映射**

如果 model 和 proto 的枚举类型不一致，例如 model 使用 `int8`，proto 使用 `pb.GroupMemberRole`，应优先在 `converter.go` 中集中处理转换逻辑。

推荐模式：

```go
func toProtoGroupMemberRole(role int8) pb.GroupMemberRole {
    switch role {
    case int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_OWNER):
        return pb.GroupMemberRole_GROUP_MEMBER_ROLE_OWNER
    case int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_ADMIN):
        return pb.GroupMemberRole_GROUP_MEMBER_ROLE_ADMIN
    case int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_MEMBER):
        return pb.GroupMemberRole_GROUP_MEMBER_ROLE_MEMBER
    default:
        return pb.GroupMemberRole_GROUP_MEMBER_ROLE_UNSPECIFIED
    }
}
```

svc 层只负责业务编排，不应在多个 svc 文件中重复维护 `1 = owner`、`2 = admin` 这类映射关系。

#### **5. 测试代码同样适用**

测试中也不应直接写业务枚举数字。允许在断言中使用 pb 枚举常量，避免协议调整后测试仍然“错误地通过”。

错误示例：

```go
assert.Equal(t, int8(1), member.Role)
```

正确示例：

```go
assert.Equal(t, int8(pb.GroupMemberRole_GROUP_MEMBER_ROLE_OWNER), member.Role)
```

#### **6. 本轮遗留代码需纳入整改清单**

结合 Steps 5-7 当前实现，需要补充一条整改任务：

> 扫描 member/group/topic/permission 四个包，替换 svc 层、converter 层、测试中的硬编码枚举值和魔法字符串。所有协议枚举值统一改为 protobuf 生成枚举常量；所有业务字符串统一抽取为包级常量或集中常量文件。

建议加入 Step 8 前置修复项，和之前提到的 Bug 1、Bug 2 一起处理，避免 GORM 实现阶段继续复制硬编码模式。

可以补进汇总决策表：

| 项目 | 问题 | 决策 | 时机 |
|---|---|---|---|
| Code Style 1 | svc 层硬编码 protobuf 枚举值 | 禁止，统一使用 pb 枚举常量 | Step 8 前置整改 |
| Code Style 2 | 魔法字符串散落 | 禁止，统一抽取常量 | Step 8 前置整改 |
| Code Style 3 | model/proto 枚举转换分散 | 禁止，统一放 converter 层 | Step 8 前置整改 |
| Code Style 4 | 测试中硬编码枚举数字 | 禁止，测试也使用 pb 枚举常量 | Step 8 前置整改 |

最终建议把 Step 8 启动前的修复顺序调整为：

```text
1. 修复 SvcCreate 漏写 owner member
2. 修复密码明文存储问题
3. 替换 svc/test 中 protobuf 枚举硬编码
4. 抽取业务字符串常量
5. 修复 permission getTopicVisibility 的 topicRepo 注入
6. 开始 Step 8 repository_gorm.go
```

*内容由 AI 生成仅供参考*