## Phase 1 MVP-P0 主控指令执行完成汇报
### 1. 执行内容
按主控决策表逐项执行， 全部 7 项任务已完成 ：
# 任务 决策 状态 Bug 1 SvcCreate 漏写 owner member 立即修复 ✅ 已在之前实现中包含 Bug 2 hashPassword → bcrypt 立即修复 ✅ 已在之前实现中包含 Q2c permission 注入 topicRepo + helpers Step 8 前 ✅ 全部已实现 Step 8 member GORM + 冒烟测试 立即执行 ✅ 11 测试通过 Step 8 group GORM + 冒烟测试 立即执行 ✅ 14 测试通过（修复 3 个断言类型） Step 8 topic GORM + 冒烟测试 立即执行 ✅ 16 测试通过（修复 upsert+索引） Step 8 尾 permission helpers 实现 随 Step 8 ✅ isUserActive/getMemberRole/getTopicVisibility
### 2. 本轮修复的文件
文件路径 变更类型 说明 group/repository_gorm_test.go 修复 GroupStats 断言 int64→int（3 个字段） topic/repository_gorm.go 修复 UpsertRead ReadDuration 改为替换而非累加 topic/model.go 修复 TopicRead uniqueIndex 改为复合索引 (topic_id+user_id)
### 3. 测试命令与结果
全域 102 个测试全绿，0 失败。
### 4. Bug 与事故
无新增 Bug。本轮修复 3 个既有 GORM 测试缺陷（均为类型/逻辑问题，非业务 Bug）。
### 5. 文档同步情况
无需更新文档（代码实现与 DevGuide / 实现计划一致）。
### 6. 遗留事项（主控已确认暂缓）
项目 状态 计划 CanAuditContent 暂缓返回 false Post-MVP，待 platform admin 数据模型设计 CanReadTopic PAID_MEMBER 分支 暂缓返回 false Step 8+ 后评估 getMemberRole 扩展 SvcRefresh (1027) 存根不动 Step 9 auth 中间件同步实现 跨 svc 集成测试 不做 Step 9 auth 中间件就绪后补 Proto 层测试
### 7. 需要项目主控确认
1. Phase 1 Steps 5-8 全部完成 ，是否继续 Steps 9-12 （module 路由聚合 → routes.yaml → CI 验证）？
2. 是否需要补充 permission service 单元测试 （当前仅编译通过，无测试覆盖）？

### **确认：Phase 1 Steps 5-8 可判定完成，下一步继续 Steps 9-12；同时要求先补 permission service 单元测试，作为进入路由聚合前的质量门禁。**

收到执行完成汇报。基于当前状态：三域 svc、三域 GORM repository、Step 8 尾 permission helpers 均已实现，全域 102 个测试全绿，Phase 1 MVP-P0 的 Steps 5-8 可以进入完成态。

## **主控确认结论**

### **1. 是否继续 Steps 9-12？**

**确认继续。**

执行顺序建议如下：

```text
Step 9  : module 路由聚合 / servant wiring / handler 接入校验
Step 10 : routes.yaml 注册 maxType/minType → Tars Obj 映射
Step 11 : CI 验证 / go test ./... / 静态检查 / 编译检查
Step 12 : Phase 1 P0 收口验收 / 文档同步 / 遗留项冻结
```

其中 Step 9 不建议直接跳到 routes.yaml。应先确认 `module.go`、三域 `servant.go`、三域 `handler.go`、permission service 注入链路都能完整串起来，避免路由配置已经登记但运行时依赖未注入完整。

### **2. 是否需要补充 permission service 单元测试？**

**需要补充，而且建议作为 Step 9 前置任务。**

原因是 permission service 是后续所有高权限操作和帖子阅读权限的安全边界。当前“编译通过”只能证明依赖注入和类型签名正确，不能证明权限决策正确。尤其是本轮刚补了 `isUserActive`、`getMemberRole`、`getTopicVisibility`，这些 helper 直接影响 `CanManageGroup`、`CanManageMember`、`CanPublishTopic`、`CanReadTopic` 等核心能力。

建议新增一个小任务：

```text
Step 8.5 : permission service 单元测试补齐
```

## **Step 8.5 permission service 测试范围**

最低覆盖以下用例即可，不需要扩大成跨 svc 集成测试。

| 能力 | 必测场景 | 期望 |
|---|---|---|
| `isUserActive` | active 用户 | true |
| `isUserActive` | banned/deleted/不存在用户 | false |
| `getMemberRole` | owner/admin/member active | 返回正确 role/status |
| `getMemberRole` | member 不存在 | nil |
| `getTopicVisibility` | topic 存在 | 返回 groupID + visibility |
| `getTopicVisibility` | topic 不存在 | nil |
| `CanManageGroup` | owner/admin | allow |
| `CanManageGroup` | 普通 member/非成员 | deny |
| `CanManageMember` | admin 操作普通 member | allow |
| `CanManageMember` | member 操作他人 | deny |
| `CanManageMember` | 操作 owner | deny |
| `CanPublishTopic` | active member + active group + 未禁言 | allow |
| `CanPublishTopic` | 被禁言/非成员/群关闭 | deny |
| `CanReadTopic` | PUBLIC | allow |
| `CanReadTopic` | GROUP_MEMBER + active member | allow |
| `CanReadTopic` | GROUP_MEMBER + 非成员 | deny |
| `CanReadTopic` | OWNER_ONLY + owner/admin | allow |
| `CanReadTopic` | OWNER_ONLY + 普通 member | deny |
| `CanReadTopic` | PAID_MEMBER | 当前保持 false，作为已知暂缓行为锁定 |

其中 `PAID_MEMBER` 分支当前主控已确认暂缓，因此测试应明确锁定当前行为，避免未来误以为已经支持付费阅读：

```go
func TestCanReadTopic_PaidMember_TemporaryDenied(t *testing.T) {
    // 当前 MVP-P0 暂缓 PAID_MEMBER 权益判断，统一返回 false。
}
```

## **Step 9 开始前新增质量门禁**

建议 Step 9 的进入条件调整为：

```text
1. member/group/topic repository GORM 测试全绿
2. member/group/topic svc 测试全绿
3. permission service 单元测试全绿
4. go test ./go/modules/social/... 全绿
5. 不存在 svc 层硬编码 proto 枚举值
6. 不存在新增魔法字符串散落
```

特别是第 5、6 点，需要延续前面补充的编码规范：svc 层使用 protobuf 枚举常量，不直接写 `1/2/3`；业务字符串使用常量，不散落 `"owner"`、`"admin"`、`"active"`、`"paid"` 等字面量。

## **后续执行指令**

主控确认如下：

```text
结论：Phase 1 Steps 5-8 完成。
下一步：先执行 Step 8.5 permission service 单元测试补齐。
通过后：继续 Step 9-12。
```

建议任务拆分为：

```text
Step 8.5.1 新增 permission/mock_repository_test.go
Step 8.5.2 覆盖 isUserActive/getMemberRole/getTopicVisibility helper
Step 8.5.3 覆盖 CanManageGroup/CanManageMember/CanPublishTopic/CanReadTopic
Step 8.5.4 执行 go test ./go/modules/social/...
Step 9.1 检查 module.go 聚合注入
Step 9.2 检查 servant/handler 路由分发
Step 10 注册 routes.yaml
Step 11 CI 全量验证
Step 12 Phase 1 P0 收口
```

`SvcRefresh`、`CanAuditContent`、`CanReadTopic PAID_MEMBER`、跨 svc 集成测试继续保持暂缓，不进入 Step 8.5 范围，避免范围蔓延。

*内容由 AI 生成仅供参考*