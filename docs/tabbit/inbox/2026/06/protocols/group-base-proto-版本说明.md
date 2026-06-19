# group_base.proto 精简版本说明

**版本日期**: 2026-06-15
**变更类型**: refactor
**状态**: 已完成

---

## 1. 变更概述

本次变更将 `group_base.proto` 中废弃或暂未使用的 message / enum 迁移至 `deprecated.proto`，实现主协议文件的精简，降低维护成本。

- **原文件**: `protobuf/protocols/base/group_base.proto`
- **目标文件**: `protobuf/protocols/base/deprecated.proto`
- **go_package**: 保持不变（`github.com/jimiechen/mineplanet/protocols/generated/go/proto/base`）

---

## 2. 迁移清单

### 2.1 已迁移到 deprecated.proto 的 message / enum（共 43 个）

| 序号 | 类型 | 名称 | 原 minType | 备注 |
|------|------|------|-----------|------|
| 1 | message | `GroupApplication` | - | 圈子申请信息 |
| 2 | message | `GroupSetting` | - | 数据类型（已保留在 group_base.proto，见备注） |
| 3 | message | `UpdateGroupSettingRequest` | 2063 | 更新圈子扩展设置项 |
| 4 | message | `UpdateGroupSettingResponse` | 2064 | 更新圈子扩展设置项响应 |
| 5 | message | `GroupMemberView` | - | 成员详情视图（已标记即将废弃） |
| 6 | message | `GetMemberDetailRequest` | 2033 | 获取成员详情 |
| 7 | message | `GetMemberDetailResponse` | 2034 | 获取成员详情响应 |
| 8 | message | `UpdatePaymentCycleRequest` | 2035 | 调整成员付费周期 |
| 9 | message | `UpdatePaymentCycleResponse` | 2036 | 调整成员付费周期响应 |
| 10 | enum | `GroupUserStatusMask` | - | 字段掩码枚举 |
| 11 | message | `GroupUserStatusRequest` | 2085 | 用户状态请求 |
| 12 | message | `GroupRoleUsers` | - | 管理团队成员信息（已保留在 group_base.proto，见备注） |
| 13 | message | `RolePermissionConfig` | - | 角色权限配置项 |
| 14 | message | `CurrentUserPermissions` | - | 当前用户权限信息 |
| 15 | message | `MembershipStatus` | - | 会员状态信息 |
| 16 | message | `MembershipExpiration` | - | 会员到期时间信息 |
| 17 | message | `ExpirationReminder` | - | 到期提醒信息 |
| 18 | message | `GroupUserStatusResponse` | 2086 | 用户状态响应 |
| 19 | message | `GroupMember` | - | 圈子成员信息（已标记即将废弃） |
| 20 | message | `GetGroupMembersRequest` | 2017 | 分页获取圈子成员列表 |
| 21 | message | `GetGroupMembersResponse` | 2018 | 分页获取圈子成员列表响应 |
| 22 | message | `GroupInfoExtra` | - | 圈子扩展信息 |
| 23 | message | `GetGroupListRequest` | 2001 | 分页获取圈子列表 |
| 24 | message | `GetGroupListResponse` | 2002 | 分页获取圈子列表响应 |
| 25 | message | `PermissionConfigDetail` | - | 权限配置详情 |
| 26 | message | `PermissionCondition` | - | 权限条件 |
| 27 | message | `PermissionConstraints` | - | 权限约束 |
| 28 | message | `UpdateGroupPermissionRequest` | 2081 | 更新圈子角色权限配置 |
| 29 | message | `UpdateGroupPermissionResponse` | 2082 | 更新圈子角色权限配置响应 |
| 30 | message | `GroupRealtimeState` | - | 圈子实时状态数据 |
| 31 | message | `RefreshGroupStatResponse` | 2042 | 返回圈子实时状态列表 |
| 32 | message | `CreateGroupRequest` | 2005 | 创建圈子（含可选付费配置） |
| 33 | message | `GetGroupInfoRequest` | 2007 | 获取圈子详情 |
| 34 | message | `GetGroupInfoResponse` | 2008 | 返回圈子详情 |
| 35 | message | `UpdateMemberPermissionsRequest` | 2031 | 更新成员权限 |
| 36 | message | `UpdateMemberPermissionsResponse` | 2032 | 更新成员权限响应 |
| 37 | message | `GetGroupHomePageRequest` | 2061 | 获取圈子主页 |
| 38 | message | `GetGroupHomePageResponse` | 2062 | 获取圈子主页响应 |
| 39 | message | `GetGroupSettingsRequest` | 2083 | 获取圈子扩展设置项 |
| 40 | message | `GetGroupSettingsResponse` | 2084 | 返回圈子扩展设置项 |
| 41 | message | `GroupSettings` | - | 圈子设置 |
| 42 | message | `GroupSetting` | - | 扩展设置项 |
| 43 | message | `RolePermissionsFlat` | - | 角色权限扁平化 |

> **备注**: `GroupRoleUsers` 因被 `group_base.proto` 中活跃使用的 response 引用，为避免循环导入问题，**保留**在了 `group_base.proto` 中，未实际迁移。`GroupSetting` 原被 `GetGroupInfoResponse` 引用，但该 response 已迁移至 `deprecated.proto`，故 `GroupSetting` 一并迁移。

---

## 3. group_base.proto 剩余协议（共 93 个 message / enum）

### 3.1 枚举类型（共 13 个）

| 序号 | 枚举名称 | 说明 |
|------|---------|------|
| 1 | `MemberSortBy` | 成员排序方式 |
| 2 | `GroupStatus` | 圈子状态 |
| 3 | `GroupVisibility` | 圈子可见性 |
| 4 | `JoinMode` | 加入方式 |
| 5 | `PayCycleUnit` | 付费周期单位 |
| 6 | `QuestionLimit` | 提问限制 |
| 7 | `WhoCanPost` | 发帖权限 |
| 8 | `DiscountType` | 折扣类型 |
| 9 | `MemberStatus` | 成员状态 |
| 10 | `InvitationStatus` | 邀请状态 |
| 11 | `ApplicationStatus` | 申请状态 |
| 12 | `JoinStatus` | 加入状态 |
| 13 | `MuteDuration` | 禁言时长 |

### 3.2 数据对象（共 19 个）

| 序号 | 对象名称 | 说明 |
|------|---------|------|
| 1 | `GroupPayConfig` | 圈子付费配置 |
| 2 | `GroupStats` | 圈子统计信息 |
| 3 | `UserMemberInfo` | 用户成员信息 |
| 4 | `GroupInfo` | 圈子信息 |
| 5 | `OwnerProfile` | 圈主信息 |
| 6 | `LatestTopicPreview` | 最新帖子预览 |
| 7 | `MemberInfo` | 成员信息 |
| 8 | `GroupInvitation` | 圈子邀请 |
| 9 | `GroupRoleUsers` | 管理团队成员信息 |
| 10 | `DiscountItem` | 折扣项 |
| 11 | `ContentNavItem` | 内容导航项 |
| 12 | `GroupActivityStats` | 圈子活动统计 |
| 13 | `UserGroupQuota` | 用户圈子配额 |
| 14 | `MembershipStatusLite` | 会员状态精简版 |
| 15 | `GroupStatsLite` | 圈子统计精简版 |
| 16 | `GroupInfoLite` | 圈子信息精简版 |
| 17 | `AdvancedSettingsLite` | 高级设置精简版 |
| 18 | `MemberRoleMapping` | 成员角色映射 |
| 19 | `MemberRoleMappingByMemberId` | 按成员ID角色映射 |

### 3.3 接口协议（共 61 个 Request / Response）

#### 3.3.1 圈子管理

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `GetGroupStatsRequest` | 2065 | 获取圈子统计 |
| 2 | `GetGroupStatsResponse` | 2066 | 获取圈子统计响应 |
| 3 | `BatchGetGroupsRequest` | 2073 | 批量获取圈子 |
| 4 | `BatchGetGroupsResponse` | 2074 | 批量获取圈子响应 |
| 5 | `CheckGroupNameAvailabilityRequest` | 2083 | 检查圈子名称可用性 |
| 6 | `CheckGroupNameAvailabilityResponse` | 2084 | 检查圈子名称可用性响应 |
| 7 | `RefreshGroupStatRequest` | 2041 | 刷新圈子状态 |
| 8 | `CreateGroupResponse` | 2006 | 创建圈子响应 |
| 9 | `UpdateGroupRequest` | 2009 | 更新圈子 |
| 10 | `UpdateGroupResponse` | 2010 | 更新圈子响应 |
| 11 | `DeleteGroupRequest` | 2059 | 删除圈子 |
| 12 | `DeleteGroupResponse` | 2060 | 删除圈子响应 |
| 13 | `CheckGroupSlugAvailabilityRequest` | 2091 | 检查圈子标识符可用性 |
| 14 | `CheckGroupSlugAvailabilityResponse` | 2092 | 检查圈子标识符可用性响应 |

#### 3.3.2 成员管理

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `JoinGroupRequest` | 2011 | 加入圈子 |
| 2 | `JoinGroupResponse` | 2012 | 加入圈子响应 |
| 3 | `LeaveGroupRequest` | 2013 | 离开圈子 |
| 4 | `LeaveGroupResponse` | 2014 | 离开圈子响应 |
| 5 | `GetGroupMemberUserIdsRequest` | 2015 | 获取圈子成员用户ID |
| 6 | `GetGroupMemberUserIdsResponse` | 2016 | 获取圈子成员用户ID响应 |
| 7 | `GetMemberInfosRequest` | 2079 | 获取成员信息 |
| 8 | `GetMemberInfosResponse` | 2080 | 获取成员信息响应 |
| 9 | `SearchGroupMembersRequest` | 2077 | 搜索圈子成员 |
| 10 | `SearchGroupMembersResponse` | 2078 | 搜索圈子成员响应 |
| 11 | `GetMemberRolesRequest` | 2019 | 获取成员角色 |
| 12 | `GetMemberRolesResponse` | 2020 | 获取成员角色响应 |
| 13 | `GetMemberRolesByMemberIdsRequest` | 2021 | 按成员ID获取角色 |
| 14 | `GetMemberRolesByMemberIdsResponse` | 2022 | 按成员ID获取角色响应 |
| 15 | `MuteMemberRequest` | 2043 | 禁言成员 |
| 16 | `MuteMemberResponse` | 2044 | 禁言成员响应 |
| 17 | `UnmuteMemberRequest` | 2045 | 解除禁言 |
| 18 | `UnmuteMemberResponse` | 2046 | 解除禁言响应 |
| 19 | `BanMemberRequest` | 2047 | 封禁成员 |
| 20 | `BanMemberResponse` | 2048 | 封禁成员响应 |
| 21 | `UnbanMemberRequest` | 2049 | 解除封禁 |
| 22 | `UnbanMemberResponse` | 2050 | 解除封禁响应 |
| 23 | `GetBannedMembersRequest` | 2051 | 获取封禁成员列表 |
| 24 | `GetBannedMembersResponse` | 2052 | 获取封禁成员列表响应 |
| 25 | `RemoveMemberRequest` | 2053 | 移除成员 |
| 26 | `RemoveMemberResponse` | 2054 | 移除成员响应 |
| 27 | `UpdateMemberRoleRequest` | 2025 | 更新成员角色 |
| 28 | `UpdateMemberRoleResponse` | 2026 | 更新成员角色响应 |
| 29 | `RenewMemberRequest` | 2023 | 成员续费 |
| 30 | `RenewMemberResponse` | 2024 | 成员续费响应 |
| 31 | `UpdateMemberInfoRequest` | 2039 | 更新成员信息 |
| 32 | `UpdateMemberInfoResponse` | 2040 | 更新成员信息响应 |

#### 3.3.3 权限与配置

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `UpdateGroupDiscountsRequest` | 2029 | 更新圈子折扣 |
| 2 | `UpdateGroupDiscountsResponse` | 2030 | 更新圈子折扣响应 |
| 3 | `GetGroupDiscountsRequest` | 2037 | 获取圈子折扣 |
| 4 | `GetGroupDiscountsResponse` | 2038 | 获取圈子折扣响应 |
| 5 | `CalcPayableAmountRequest` | 2071 | 计算应付金额 |
| 6 | `CalcPayableAmountResponse` | 2072 | 计算应付金额响应 |
| 7 | `GroupUserEnterRequest` | 2067 | 用户进入圈子 |
| 8 | `GroupUserEnterResponse` | 2068 | 用户进入圈子响应 |
| 9 | `GetUserGroupQuotaRequest` | 2093 | 获取用户圈子配额 |
| 10 | `GetUserGroupQuotaResponse` | 2094 | 获取用户圈子配额响应 |
| 11 | `GetGroupPermissionsLiteRequest` | 2089 | 获取圈子权限精简版 |
| 12 | `GetGroupPermissionsLiteResponse` | 2090 | 获取圈子权限精简版响应 |
| 13 | `UpdateRolePermissionsRequest` | 2055 | 更新角色权限 |
| 14 | `UpdateRolePermissionsResponse` | 2056 | 更新角色权限响应 |
| 15 | `QuickUpdateSettingRequest` | 2057 | 快速更新设置 |
| 16 | `QuickUpdateSettingResponse` | 2058 | 快速更新设置响应 |

---

## 4. deprecated.proto 已迁移协议

```mermaid
flowchart LR
    subgraph 已废弃协议
        F1[GroupApplication]
        F2[UpdateGroupSettingRequest]
        F3[UpdateGroupSettingResponse]
        F4[GroupMemberView]
        F5[GetMemberDetailRequest]
        F6[GetMemberDetailResponse]
        F7[UpdatePaymentCycleRequest]
        F8[UpdatePaymentCycleResponse]
        F9[GroupUserStatusMask]
        F10[GroupUserStatusRequest]
        F11[RolePermissionConfig]
        F12[CurrentUserPermissions]
        F13[MembershipStatus]
        F14[MembershipExpiration]
        F15[ExpirationReminder]
        F16[GroupUserStatusResponse]
        F17[GroupMember]
        F18[GetGroupMembersRequest]
        F19[GetGroupMembersResponse]
        F20[GroupInfoExtra]
        F21[GetGroupListRequest]
        F22[GetGroupListResponse]
        F23[PermissionConfigDetail]
        F24[PermissionCondition]
        F25[PermissionConstraints]
        F26[UpdateGroupPermissionRequest]
        F27[UpdateGroupPermissionResponse]
        F28[GroupRealtimeState]
        F29[RefreshGroupStatResponse]
        F30[CreateGroupRequest]
        F31[GetGroupInfoRequest]
        F32[GetGroupInfoResponse]
        F33[UpdateMemberPermissionsRequest]
        F34[UpdateMemberPermissionsResponse]
        F35[GetGroupHomePageRequest]
        F36[GetGroupHomePageResponse]
        F37[GetGroupSettingsRequest]
        F38[GetGroupSettingsResponse]
        F39[GroupSettings]
        F40[GroupSetting]
        F41[RolePermissionsFlat]
    end
```

---

## 5. 编译验证

```bash
cd /Users/mac/StudioProjects/2026/open-citycloud-workspace/protobuf/protocols
protoc --proto_path=. base/*.proto --descriptor_set_out=/dev/null
```

**结果**: ✅ 编译通过，无错误无警告。

---

## 6. 后续事项

- [ ] 重新生成 Go 代码（`make proto`）
- [ ] 业务代码编译检查（`go build ./...`）
- [ ] 运行相关单元测试
- [ ] 提交并推送代码
