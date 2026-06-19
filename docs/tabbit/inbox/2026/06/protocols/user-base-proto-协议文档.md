# user_base.proto 协议文档

**整理日期**: 2026-06-15
**文件路径**: `protobuf/protocols/base/user_base.proto`
**go_package**: `github.com/jimiechen/mineplanet/protocols/generated/go/proto/base`

---

## 1. 协议概述

本文档整理了 `user_base.proto` 中定义的全部协议，按**枚举类型**、**数据对象**、**接口协议（Request/Response）** 分类展示。

---

## 2. 枚举类型（共 6 个）

| 序号 | 枚举名称 | 说明 |
|------|---------|------|
| 1 | `GroupMemberRole` | 群组成员角色 |
| 2 | `UserConfigScope` | 用户配置作用域 |
| 3 | `UserConfigKey` | 用户配置键名 |
| 4 | `UserConfigValueType` | 用户配置值类型 |
| 5 | `NotificationDeliveryMode` | 通知投递模式 |
| 6 | `NotificationCategory` | 通知类别 |

---

## 3. 数据对象（共 6 个）

| 序号 | 对象名称 | 说明 |
|------|---------|------|
| 1 | `UserBlock` | 用户黑名单信息（极简版） |
| 2 | `GroupUserInfo` | 群组用户信息 |
| 3 | `UserStats` | 用户统计信息 |
| 4 | `UserConfigItem` | 用户配置项（精简版） |
| 5 | `NotificationSettingsPayload` | 通知设置载荷 |
| 6 | `NotificationCategoryToggle` | 通知类别开关项 |

---

## 4. 接口协议（共 42 个 Request / Response）

### 4.1 登录认证

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `UserRegisterRequest` | 1021 | 用户注册 |
| 2 | `UserRegisterResponse` | 1022 | 用户注册响应 |
| 3 | `UserLoginRequest` | 1023 | 用户登录 |
| 4 | `UserLoginResponse` | 1024 | 用户登录响应 |
| 5 | `UserLogoutRequest` | 1025 | 用户登出 |
| 6 | `UserLogoutResponse` | 1026 | 用户登出响应 |
| 7 | `RefreshTokenRequest` | 1027 | 刷新令牌 |
| 8 | `RefreshTokenResponse` | 1028 | 刷新令牌响应 |
| 9 | `GetUserInfoRequest` | 1029 | 获取当前用户信息 |
| 10 | `GetUserInfoResponse` | 1030 | 获取当前用户信息响应 |

### 4.2 用户管理

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `UpdateUserInfoRequest` | 1031 | 更新用户信息 |
| 2 | `UpdateUserInfoResponse` | 1032 | 更新用户信息响应 |
| 3 | `BlockUserRequest` | 1039 | 拉黑用户 |
| 4 | `BlockUserResponse` | 1040 | 拉黑用户响应 |
| 5 | `UnblockUserRequest` | 1041 | 解除拉黑 |
| 6 | `UnblockUserResponse` | 1042 | 解除拉黑响应 |
| 7 | `GetBlockListRequest` | 1043 | 获取黑名单列表 |
| 8 | `GetBlockListResponse` | 1044 | 获取黑名单列表响应 |
| 9 | `GetBlockCountRequest` | 1047 | 获取拉黑数量 |
| 10 | `GetBlockCountResponse` | 1048 | 获取拉黑数量响应 |

### 4.3 用户查询与统计

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `BatchGetUserInfoRequest` | 1049 | 批量获取用户信息 |
| 2 | `BatchGetUserInfoResponse` | 1050 | 批量获取用户信息响应 |
| 3 | `UpgradeMembershipRequest` | 1051 | 会员升级 |
| 4 | `UpgradeMembershipResponse` | 1052 | 会员升级响应 |
| 5 | `GetNotificationSettingsRequest` | 1053 | 获取用户通知设置 |
| 6 | `GetNotificationSettingsResponse` | 1054 | 获取用户通知设置响应 |
| 7 | `UpdateNotificationSettingsRequest` | 1055 | 更新用户通知设置 |
| 8 | `UpdateNotificationSettingsResponse` | 1056 | 更新用户通知设置响应 |
| 9 | `GetUserStatsRequest` | 1045 | 获取用户统计 |
| 10 | `GetUserStatsResponse` | 1046 | 获取用户统计响应 |

### 4.4 IM 服务

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `GetIMUserSigRequest` | 1074 | 获取腾讯云 IM 登录签名 |
| 2 | `GetIMUserSigResponse` | 1075 | 获取腾讯云 IM 登录签名响应 |

### 4.5 用户配置

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `GetUserConfigRequest` | 1091 | 获取用户配置 |
| 2 | `GetUserConfigResponse` | 1092 | 获取用户配置响应 |
| 3 | `UpdateUserConfigRequest` | 1093 | 更新用户配置（单条） |
| 4 | `UpdateUserConfigResponse` | 1094 | 更新用户配置响应 |
| 5 | `BatchUpdateUserConfigRequest` | 1095 | 批量更新用户配置 |
| 6 | `BatchUpdateUserConfigResponse` | 1096 | 批量更新用户配置响应 |
| 7 | `DeleteUserConfigRequest` | 1097 | 删除用户配置 |
| 8 | `DeleteUserConfigResponse` | 1098 | 删除用户配置响应 |
| 9 | `DismissPopupRequest` | 1099 | 关闭弹框 |
| 10 | `DismissPopupResponse` | 1100 | 关闭弹框响应 |

---

## 5. 协议编号范围

| 范围 | 用途 |
|------|------|
| 1021-1032 | 登录认证与用户基础管理 |
| 1039-1044 | 黑名单管理 |
| 1045-1048 | 用户统计与拉黑数量 |
| 1049-1056 | 批量查询、会员升级、通知设置 |
| 1074-1075 | IM 服务签名 |
| 1091-1100 | 用户配置与弹框管理 |

---

## 6. 编译验证

```bash
cd /Users/mac/StudioProjects/2026/open-citycloud-workspace/protobuf/protocols
protoc --proto_path=. base/user_base.proto --descriptor_set_out=/dev/null
```

**结果**: ✅ 编译通过，无错误无警告。
