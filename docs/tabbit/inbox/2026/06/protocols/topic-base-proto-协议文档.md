# topic_base.proto 协议文档

**整理日期**: 2026-06-15
**文件路径**: `protobuf/protocols/base/topic_base.proto`
**go_package**: `github.com/jimiechen/mineplanet/protocols/generated/go/proto/base`

---

## 1. 协议概述

本文档整理了 `topic_base.proto` 中定义的全部协议，按**枚举类型**、**数据对象**、**接口协议（Request/Response）** 分类展示。

---

## 2. 枚举类型（共 12 个）

| 序号 | 枚举名称 | 说明 |
|------|---------|------|
| 1 | `TopicStatus` | 话题状态 |
| 2 | `TopicType` | 话题类型 |
| 3 | `TopicNavType` | 话题导航类型 |
| 4 | `ReplyStatus` | 回复状态 |
| 5 | `MediaType` | 媒体类型 |
| 6 | `MediaStatus` | 媒体状态 |
| 7 | `ContentType` | 内容类型 |
| 8 | `ContentFormat` | 内容格式 |
| 9 | `Visibility` | 可见性 |
| 10 | `MimeType` | MIME 类型 |
| 11 | `TopicLifecycleStatus` | 话题生命周期状态 |
| 12 | `InteractionType` | 互动类型 |

---

## 3. 数据对象（共 13 个）

| 序号 | 对象名称 | 说明 |
|------|---------|------|
| 1 | `TopicInfo` | 话题信息 |
| 2 | `ExternalLink` | 外部链接 |
| 3 | `MediaItem` | 媒体项 |
| 4 | `TopicInfoExtra` | 话题扩展信息 |
| 5 | `TopicStat` | 话题统计 |
| 6 | `ReferReplyInfo` | 引用回复信息 |
| 7 | `ReplyInfo` | 回复信息 |
| 8 | `TopicAuthorInfo` | 话题作者信息 |
| 9 | `LinkInfo` | 链接信息 |
| 10 | `TopicUserStatus` | 话题用户状态 |
| 11 | `TopicActionState` | 话题行为状态 |
| 12 | `ReportTypeOption` | 举报类型选项 |

---

## 4. 接口协议（共 30 个 Request / Response）

### 4.1 话题管理

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `CreateTopicRequest` | 3001 | 创建话题 |
| 2 | `CreateTopicResponse` | 3002 | 创建话题响应 |
| 3 | `GetTopicListRequest` | 3005 | 获取话题列表 |
| 4 | `GetTopicListResponse` | 3006 | 获取话题列表响应 |
| 5 | `DeleteTopicRequest` | 3009 | 删除话题 |
| 6 | `DeleteTopicResponse` | 3010 | 删除话题响应 |
| 7 | `PinTopicRequest` | 3029 | 置顶话题 |
| 8 | `PinTopicResponse` | 3030 | 置顶话题响应 |
| 9 | `LikeTopicRequest` | 3061 | 点赞话题 |
| 10 | `LikeTopicResponse` | 3062 | 点赞话题响应 |
| 11 | `FavoriteTopicRequest` | 3063 | 收藏话题 |
| 12 | `FavoriteTopicResponse` | 3064 | 收藏话题响应 |
| 13 | `SearchTopicsRequest` | 3049 | 搜索话题 |
| 14 | `SearchTopicsResponse` | 3050 | 搜索话题响应 |
| 15 | `FeatureTopicRequest` | 3091 | 精华话题 |
| 16 | `FeatureTopicResponse` | 3092 | 精华话题响应 |

### 4.2 回复管理

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `AddTopicReplyRequest` | 3043 | 添加话题回复 |
| 2 | `AddTopicReplyResponse` | 3044 | 添加话题回复响应 |
| 3 | `DeleteTopicReplyRequest` | 3055 | 删除话题回复 |
| 4 | `DeleteTopicReplyResponse` | 3056 | 删除话题回复响应 |
| 5 | `GetReplyListRequest` | 3065 | 获取回复列表 |
| 6 | `GetReplyListResponse` | 3066 | 获取回复列表响应 |
| 7 | `LikeReplyRequest` | 3077 | 点赞回复 |
| 8 | `LikeReplyResponse` | 3078 | 点赞回复响应 |
| 9 | `PinCommentRequest` | 3081 | 置顶评论 |
| 10 | `PinCommentResponse` | 3082 | 置顶评论响应 |

### 4.3 统计与批量

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `GetTopicStatRequest` | 3035 | 获取话题统计 |
| 2 | `GetTopicStatResponse` | 3036 | 获取话题统计响应 |
| 3 | `RefreshTopicStatRequest` | 3037 | 刷新话题统计 |
| 4 | `RefreshTopicStatResponse` | 3038 | 刷新话题统计响应 |
| 5 | `BatchGetTopicInfoRequest` | 3057 | 批量获取话题信息 |
| 6 | `BatchGetTopicInfoResponse` | 3058 | 批量获取话题信息响应 |

### 4.4 用户状态

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `CheckTopicsUserStatusRequest` | 3083 | 检查话题用户状态 |
| 2 | `CheckTopicsUserStatusResponse` | 3084 | 检查话题用户状态响应 |
| 3 | `GetTopicIdsRequest` | 3085 | 获取话题ID列表 |
| 4 | `GetTopicIdsResponse` | 3086 | 获取话题ID列表响应 |

### 4.5 举报

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `GetReportTypesRequest` | 3093 | 获取举报类型 |
| 2 | `GetReportTypesResponse` | 3094 | 获取举报类型响应 |
| 3 | `CreateReportRequest` | 3095 | 创建举报 |
| 4 | `CreateReportResponse` | 3096 | 创建举报响应 |

### 4.6 行为检查

| 序号 | 协议名称 | minType | 说明 |
|------|---------|---------|------|
| 1 | `CheckTopicActionsRequest` | 3099 | 检查话题行为 |
| 2 | `CheckTopicActionsResponse` | 3100 | 检查话题行为响应 |

---

## 5. 协议编号范围

| 范围 | 用途 |
|------|------|
| 3001-3002, 3005-3010 | 话题基础操作（创建、列表、删除） |
| 3029-3030 | 置顶话题 |
| 3035-3038 | 统计与批量获取 |
| 3043-3044 | 添加话题回复 |
| 3049-3050 | 搜索 |
| 3055-3058 | 回复删除与批量获取 |
| 3061-3066 | 点赞与收藏、回复列表 |
| 3077-3078 | 点赞回复 |
| 3081-3086 | 评论置顶与用户状态、话题ID |
| 3091-3096 | 精华、举报 |
| 3099-3100 | 行为检查 |

---

## 6. 编译验证

```bash
cd /Users/mac/StudioProjects/2026/open-citycloud-workspace/protobuf/protocols
protoc --proto_path=. base/topic_base.proto --descriptor_set_out=/dev/null
```

**结果**: ✅ 编译通过，无错误无警告。
