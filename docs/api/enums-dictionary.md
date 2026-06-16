# 全局枚举字典定义

> **文档编号**: ENUMS-DICTIONARY
> **版本**: v1.0
> **创建日期**: 2026-06-15
> **状态**: 正式
> **相关 PRD**: [PRD-social-app-mvp.md](../prd/PRD-social-app-mvp.md) (PRD-20)
> **相关 API**: [social-openapi.yaml](./social-openapi.yaml)
> **相关注册表**: [协议编号注册表](./协议编号注册表.md)

***

## 1. 文档职责与边界

### 1.1 职责

本文档负责定义 CaiRobot MVP 系统中所有**稳定枚举、状态值、类型值**的统一数据字典。

采用 go-admin 风格的 `sys_dict_type` + `sys_dict_data` 双表结构：

- **sys_dict_type**（字典类型表）：定义每个枚举分类的元信息（dict_type key、名称、使用范围、系统内置标识）
- **sys_dict_data**（字典数据表）：定义每个枚举分类下的所有可选值（dict_value、dict_label、排序、默认值）

### 1.2 使用范围

本字典供以下模块共同引用：

| 模块 | 使用方式 |
|------|----------|
| App 端（iOS/Android） | 状态展示、选项渲染、表单校验 |
| Admin 运营后台 | 下拉选择、筛选条件、状态标签 |
| 社交域后端 | 业务逻辑判断、数据校验、状态流转 |
| 支付域后端 | 订单状态、支付渠道、分账状态 |
| 消息域后端 | 消息类型、回执状态、模板管理 |

**原则：后台页面的状态/类型/来源/角色等字段优先从数据字典读取，避免硬编码。**

### 1.3 不负责的内容

- 不负责业务流程的状态机定义（由各域 PRD 和 ADR 定义）
- 不负责协议字段的 Protobuf 定义（由 .proto 文件和 OpenAPI 定义）
- 不负责数据库表的 DDL 定义（由 basemodel 或 migration 定义）
- 不负责前端组件的 UI 样式实现

### 1.4 与其他文档的关系

本文档基于以下文档整理，必须保持枚举值一致性：

1. **PRD-social-app-mvp.md (PRD-20)**：社交域需求规范中的枚举定义
2. **social-openapi.yaml**：社交域 OpenAPI 协议中的枚举约束
3. **协议编号注册表**：全局协议编号的唯一性保证

***

## 2. 枚举字典清单（33 个 dict_type）

### 2.1 用户相关（2 个）

#### 2.1.1 user_status — 用户状态

| 字段 | 说明 |
|------|------|
| **dict_type** | user_status |
| **字典名称** | 用户状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| active | 正常 | 1 | true |
| inactive | 未激活 | 2 | false |
| banned | 已禁用 | 3 | false |
| deleted | 已删除 | 4 | false |

---

#### 2.1.2 user_membership_level — 平台会员等级

| 字段 | 说明 |
|------|------|
| **dict_type** | user_membership_level |
| **字典名称** | 平台会员等级 |
| **使用范围** | app, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| normal | 普通 | 1 | true |
| vip | VIP | 2 | false |
| svip | SVIP | 3 | false |

---

### 2.2 群组相关（7 个）

#### 2.2.1 group_type — 群组类型

| 字段 | 说明 |
|------|------|
| **dict_type** | group_type |
| **字典名称** | 群组类型 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| plaza | 广场 | 1 | true |
| free | 免费 | 2 | false |
| paid | 付费 | 3 | false |
| mixed | 混合 | 4 | false |
| invite | 邀请制 | 5 | false |

---

#### 2.2.2 group_visibility — 群组可见性

| 字段 | 说明 |
|------|------|
| **dict_type** | group_visibility |
| **字典名称** | 群组可见性 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| public | 公开 | 1 | true |
| link | 链接可见 | 2 | false |
| private | 私密 | 3 | false |

---

#### 2.2.3 group_join_mode — 加入方式

| 字段 | 说明 |
|------|------|
| **dict_type** | group_join_mode |
| **字典名称** | 加入方式 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| virtual | 虚拟 | 1 | true |
| free | 免费 | 2 | false |
| apply | 申请 | 3 | false |
| paid | 付费 | 4 | false |
| invite | 邀请 | 5 | false |

---

#### 2.2.4 group_status — 群组状态

| 字段 | 说明 |
|------|------|
| **dict_type** | group_status |
| **字典名称** | 群组状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| active | 正常 | 1 | true |
| auditing | 审核中 | 2 | false |
| rejected | 已驳回 | 3 | false |
| banned | 已禁用 | 4 | false |
| deleted | 已删除 | 5 | false |

---

#### 2.2.5 group_member_role — 成员角色

| 字段 | 说明 |
|------|------|
| **dict_type** | group_member_role |
| **字典名称** | 成员角色 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| owner | 所有人 | 1 | false |
| admin | 管理员 | 2 | false |
| guest | 嘉宾 | 3 | false |
| member | 普通成员 | 4 | true |

---

#### 2.2.6 group_member_status — 成员状态

| 字段 | 说明 |
|------|------|
| **dict_type** | group_member_status |
| **字典名称** | 成员状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| active | 正常 | 1 | true |
| pending | 待审核 | 2 | false |
| muted | 已禁言 | 3 | false |
| banned | 已移除 | 4 | false |
| left | 已退出 | 5 | false |
| expired | 已过期 | 6 | false |

---

#### 2.2.7 group_join_source — 入群来源

| 字段 | 说明 |
|------|------|
| **dict_type** | group_join_source |
| **字典名称** | 入群来源 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| virtual | 虚拟 | 1 | true |
| free | 免费 | 2 | false |
| paid | 付费 | 3 | false |
| invite | 邀请 | 4 | false |
| admin | 管理员添加 | 5 | false |
| import | 导入 | 6 | false |

---

### 2.3 权益与支付相关（6 个）

#### 2.3.1 entitlement_type — 权益类型

| 字段 | 说明 |
|------|------|
| **dict_type** | entitlement_type |
| **字典名称** | 权益类型 |
| **使用范围** | payment, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| free | 免费 | 1 | true |
| paid | 付费 | 2 | false |
| guest | 嘉宾 | 3 | false |
| gift | 赠送 | 4 | false |
| admin_grant | 管理员开通 | 5 | false |

---

#### 2.3.2 entitlement_status — 权益状态

| 字段 | 说明 |
|------|------|
| **dict_type** | entitlement_status |
| **字典名称** | 权益状态 |
| **使用范围** | payment, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| active | 有效 | 1 | true |
| expired | 已过期 | 2 | false |
| revoked | 已撤销 | 3 | false |

---

#### 2.3.3 order_type — 订单类型

| 字段 | 说明 |
|------|------|
| **dict_type** | order_type |
| **字典名称** | 订单类型 |
| **使用范围** | payment, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| join | 入群 | 1 | true |
| renew | 续费 | 2 | false |
| gift | 赠送 | 3 | false |
| admin_grant | 管理员开通 | 4 | false |

---

#### 2.3.4 order_status — 订单状态

| 字段 | 说明 |
|------|------|
| **dict_type** | order_status |
| **字典名称** | 订单状态 |
| **使用范围** | payment, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| pending | 待支付 | 1 | true |
| paid | 已支付 | 2 | false |
| cancelled | 已取消 | 3 | false |
| refunded | 已退款 | 4 | false |
| failed | 失败 | 5 | false |
| expired | 已过期 | 6 | false |

---

#### 2.3.5 pay_channel — 支付渠道

| 字段 | 说明 |
|------|------|
| **dict_type** | pay_channel |
| **字典名称** | 支付渠道 |
| **使用范围** | payment |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| manual | 手动 | 1 | true |
| wechat | 微信 | 2 | false |
| alipay | 支付宝 | 3 | false |
| apple | Apple Pay | 4 | false |
| google | Google Pay | 5 | false |

---

#### 2.3.6 settlement_status — 分账状态

| 字段 | 说明 |
|------|------|
| **dict_type** | settlement_status |
| **字典名称** | 分账状态 |
| **使用范围** | payment, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| pending | 待结算 | 1 | true |
| confirmed | 已确认 | 2 | false |
| paid | 已打款 | 3 | false |
| rejected | 已驳回 | 4 | false |

---

### 2.4 主题/帖子相关（6 个）

#### 2.4.1 topic_type — 主题类型

| 字段 | 说明 |
|------|------|
| **dict_type** | topic_type |
| **字典名称** | 主题类型 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| normal | 普通 | 1 | true |
| announcement | 公告 | 2 | false |
| qa | 问答 | 3 | false |
| paid | 付费内容 | 4 | false |

---

#### 2.4.2 topic_status — 主题状态

| 字段 | 说明 |
|------|------|
| **dict_type** | topic_status |
| **字典名称** | 主题状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| draft | 草稿 | 1 | true |
| pending | 待审核 | 2 | false |
| published | 已发布 | 3 | false |
| hidden | 已隐藏 | 4 | false |
| deleted | 已删除 | 5 | false |

---

#### 2.4.3 topic_visibility — 主题可见性

| 字段 | 说明 |
|------|------|
| **dict_type** | topic_visibility |
| **字典名称** | 主题可见性 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| public | 公开 | 1 | true |
| group_member | 群组成员可见 | 2 | false |
| paid_member | 付费成员可见 | 3 | false |
| owner_only | 仅圈主可见 | 4 | false |

---

#### 2.4.4 topic_content_type — 内容类型

| 字段 | 说明 |
|------|------|
| **dict_type** | topic_content_type |
| **字典名称** | 内容类型 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| text | 文本 | 1 | true |
| image_text | 图文 | 2 | false |
| video | 视频 | 3 | false |
| doc | 文档 | 4 | false |
| link | 链接 | 5 | false |

---

#### 2.4.5 comment_status — 评论状态

| 字段 | 说明 |
|------|------|
| **dict_type** | comment_status |
| **字典名称** | 评论状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| normal | 正常 | 1 | true |
| pending | 待审核 | 2 | false |
| hidden | 已隐藏 | 3 | false |
| deleted | 已删除 | 4 | false |

---

#### 2.4.6 reaction_type — 互动类型

| 字段 | 说明 |
|------|------|
| **dict_type** | reaction_type |
| **字典名称** | 互动类型 |
| **使用范围** | app, social |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| like | 点赞 | 1 | true |
| favorite | 收藏 | 2 | false |
| share | 分享 | 3 | false |

---

### 2.5 消息相关（5 个）

#### 2.5.1 message_type — 消息类型

| 字段 | 说明 |
|------|------|
| **dict_type** | message_type |
| **字典名称** | 消息类型 |
| **使用范围** | message, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| private_text | 私信文本 | 1 | true |
| group_status | 群组状态通知 | 2 | false |
| order_status | 订单通知 | 3 | false |
| like | 点赞通知 | 4 | false |
| comment | 评论通知 | 5 | false |
| system | 系统通知 | 6 | false |

---

#### 2.5.2 message_status — 消息状态

| 字段 | 说明 |
|------|------|
| **dict_type** | message_status |
| **字典名称** | 消息状态 |
| **使用范围** | message, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| normal | 正常 | 1 | true |
| recalled | 已撤回 | 2 | false |
| deleted | 已删除 | 3 | false |

---

#### 2.5.3 message_receipt_status — 回执状态

| 字段 | 说明 |
|------|------|
| **dict_type** | message_receipt_status |
| **字典名称** | 回执状态 |
| **使用范围** | message |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| unread | 未读 | 1 | true |
| read | 已读 | 2 | false |
| deleted | 已删除 | 3 | false |

---

#### 2.5.4 template_type — 模板类型

| 字段 | 说明 |
|------|------|
| **dict_type** | template_type |
| **字典名称** | 模板类型 |
| **使用范围** | message, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| private | 私信 | 1 | true |
| group | 群组通知 | 2 | false |
| order | 订单通知 | 3 | false |
| interaction | 互动通知 | 4 | false |
| system | 系统通知 | 5 | false |

---

#### 2.5.5 template_status — 模板状态

| 字段 | 说明 |
|------|------|
| **dict_type** | template_status |
| **字典名称** | 模板状态 |
| **使用范围** | message, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| draft | 草稿 | 1 | true |
| active | 已生效 | 2 | false |
| disabled | 已停用 | 3 | false |

---

### 2.6 配置与广告相关（5 个）

#### 2.6.1 app_config_type — 配置类型

| 字段 | 说明 |
|------|------|
| **dict_type** | app_config_type |
| **字典名称** | 配置类型 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| string | 字符串 | 1 | true |
| number | 数字 | 2 | false |
| bool | 布尔 | 3 | false |
| json | JSON | 4 | false |

---

#### 2.6.2 app_config_env — 配置环境

| 字段 | 说明 |
|------|------|
| **dict_type** | app_config_env |
| **字典名称** | 配置环境 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| dev | 开发 | 1 | true |
| test | 测试 | 2 | false |
| staging | 预发布 | 3 | false |
| prod | 生产 | 4 | false |

---

#### 2.6.3 ad_type — 广告类型

| 字段 | 说明 |
|------|------|
| **dict_type** | ad_type |
| **字典名称** | 广告类型 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| splash | 闪屏 | 1 | true |
| banner | Banner | 2 | false |
| popup | 弹窗 | 3 | false |
| feed | 信息流 | 4 | false |

---

#### 2.6.4 ad_position — 广告位置

| 字段 | 说明 |
|------|------|
| **dict_type** | ad_position |
| **字典名称** | 广告位置 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| app_home | 首页 | 1 | true |
| plaza | 广场 | 2 | false |
| group | 群组 | 3 | false |
| topic | 帖子详情 | 4 | false |

---

#### 2.6.5 ad_status — 广告状态

| 字段 | 说明 |
|------|------|
| **dict_type** | ad_status |
| **字典名称** | 广告状态 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| draft | 草稿 | 1 | true |
| active | 已上架 | 2 | false |
| offline | 已下架 | 3 | false |
| expired | 已过期 | 4 | false |

---

### 2.7 审计与风控相关（2 个）

#### 2.7.1 audit_action_type — 审计动作

| 字段 | 说明 |
|------|------|
| **dict_type** | audit_action_type |
| **字典名称** | 审计动作 |
| **使用范围** | admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| ban_user | 禁用用户 | 1 | true |
| ban_group | 禁用群组 | 2 | false |
| delete_topic | 删除主题 | 3 | false |
| delete_comment | 删除评论 | 4 | false |
| refund_order | 退款订单 | 5 | false |

---

#### 2.7.2 block_status — Block 状态

| 字段 | 说明 |
|------|------|
| **dict_type** | block_status |
| **字典名称** | Block 状态 |
| **使用范围** | social, admin |
| **system_builtin** | true |

| dict_value | dict_label | dict_sort | is_default |
|------------|------------|-----------|------------|
| active | 生效中 | 1 | true |
| cancelled | 已取消 | 2 | false |

---

## 3. 数据字典统计汇总

| 分类 | 数量 | dict_type 列表 |
|------|------|----------------|
| 用户相关 | 2 | user_status, user_membership_level |
| 群组相关 | 7 | group_type, group_visibility, group_join_mode, group_status, group_member_role, group_member_status, group_join_source |
| 权益与支付 | 6 | entitlement_type, entitlement_status, order_type, order_status, pay_channel, settlement_status |
| 主题/帖子 | 6 | topic_type, topic_status, topic_visibility, topic_content_type, comment_status, reaction_type |
| 消息相关 | 5 | message_type, message_status, message_receipt_status, template_type, template_status |
| 配置与广告 | 5 | app_config_type, app_config_env, ad_type, ad_position, ad_status |
| 审计与风控 | 2 | audit_action_type, block_status |
| **合计** | **33** | - |

## 4. 变更流程与一致性要求

### 4.1 变更触发条件

以下情况必须更新本文档：

1. 新增业务功能需要新的状态/类型/枚举值
2. 现有枚举需要新增或废弃 dict_value
3. 枚举的使用范围发生变化
4. PRD 或 ADR 中对枚举定义进行了修订

### 4.2 强制变更顺序

```
修改本文档 → 同步更新 PRD-social → 同步更新 OpenAPI (social-openapi.yaml) → 通知前端和 App 团队
```

**禁止跳过任何环节**，确保所有引用方保持一致。

### 4.3 一致性校验检查项

每次变更后必须验证：

- [ ] 本文档中的 dict_value 与 PRD-20 中的枚举定义一致
- [ ] 本文档中的 dict_value 与 social-openapi.yaml 中的 enum 约束一致
- [ ] 新增的 dict_type 已在 sys_dict_type 表中注册
- [ ] 新增的 dict_value 已在 sys_dict_data 表中注册
- [ ] 前端/App 端的下拉选项已同步更新
- [ ] 后端的校验逻辑已同步更新

### 4.4 版本记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0 | 2026-06-15 | 初始版本，定义 33 个 dict_type | - |

***
