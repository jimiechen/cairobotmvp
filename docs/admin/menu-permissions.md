# 运营后台菜单权限配置文档

> 本文件定义 CaiRobot MVP 运营后台（go-admin）的菜单树结构、按钮权限编码、角色-权限矩阵及数据权限规则。
>
> **职责范围**：菜单层级与权限点定义、RBAC 角色权限分配、数据权限边界说明。  
> **不负责**：前端路由守卫实现、后端鉴权中间件逻辑、数据库表结构设计（见 PRD-10 与相关 ADR）。  
> **所属模块**：`web/provider-admin/`  
> **关联文档**：PRD-10-运营后台、ADR-0012-多语言 monorepo 目录布局

---

## 1. 菜单树结构

### 1.1 文本树形

```
运营后台
├── 用户管理
│   ├── 用户列表
│   ├── 用户详情
│   └── Block 关系
├── 群组管理
│   ├── 群组列表
│   ├── 群组成员
│   ├── 嘉宾管理
│   ├── 付费方案
│   └── 群组审计
├── 内容管理
│   ├── 主题管理
│   ├── 评论管理
│   └── 举报管理
├── 订单财务
│   ├── 订单管理
│   ├── 权益管理
│   └── 分账管理
├── 消息中心
│   ├── 消息列表
│   ├── 会话管理
│   └── 模板消息
├── 运营配置
│   ├── 多语言维护
│   ├── App 配置
│   ├── 闪屏广告
│   ├── Banner 管理
│   └── 全局枚举字典
└── 系统审计
    ├── 操作日志
    ├── 登录日志
    └── 高风险审批
```

### 1.2 JSON 结构（go-admin 菜单初始化数据）

```json
[
  {
    "parentId": 0,
    "path": "/admin",
    "name": "运营后台",
    "component": "Layout",
    "redirect": "/admin/users",
    "meta": { "title": "运营后台", "icon": "dashboard" }
  },
  {
    "parentId": 1,
    "path": "users",
    "name": "用户管理",
    "component": "admin/user/index",
    "meta": { "title": "用户管理", "icon": "user" }
  },
  {
    "parentId": 2,
    "path": "user-list",
    "name": "用户列表",
    "component": "admin/user/list",
    "meta": { "title": "用户列表" }
  },
  {
    "parentId": 2,
    "path": "user-detail",
    "name": "用户详情",
    "component": "admin/user/detail",
    "meta": { "title": "用户详情" }
  },
  {
    "parentId": 2,
    "path": "block-relation",
    "name": "Block 关系",
    "component": "admin/user/block",
    "meta": { "title": "Block 关系" }
  },
  {
    "parentId": 1,
    "path": "groups",
    "name": "群组管理",
    "component": "admin/group/index",
    "meta": { "title": "群组管理", "icon": "team" }
  },
  {
    "parentId": 6,
    "path": "group-list",
    "name": "群组列表",
    "component": "admin/group/list",
    "meta": { "title": "群组列表" }
  },
  {
    "parentId": 6,
    "path": "group-members",
    "name": "群组成员",
    "component": "admin/group/members",
    "meta": { "title": "群组成员" }
  },
  {
    "parentId": 6,
    "path": "guests",
    "name": "嘉宾管理",
    "component": "admin/group/guests",
    "meta": { "title": "嘉宾管理" }
  },
  {
    "parentId": 6,
    "path": "payment-plans",
    "name": "付费方案",
    "component": "admin/group/plans",
    "meta": { "title": "付费方案" }
  },
  {
    "parentId": 6,
    "path": "group-audit",
    "name": "群组审计",
    "component": "admin/group/audit",
    "meta": { "title": "群组审计" }
  },
  {
    "parentId": 1,
    "path": "content",
    "name": "内容管理",
    "component": "admin/content/index",
    "meta": { "title": "内容管理", "icon": "file-text" }
  },
  {
    "parentId": 12,
    "path": "topics",
    "name": "主题管理",
    "component": "admin/content/topics",
    "meta": { "title": "主题管理" }
  },
  {
    "parentId": 12,
    "path": "comments",
    "name": "评论管理",
    "component": "admin/content/comments",
    "meta": { "title": "评论管理" }
  },
  {
    "parentId": 12,
    "path": "reports",
    "name": "举报管理",
    "component": "admin/content/reports",
    "meta": { "title": "举报管理" }
  },
  {
    "parentId": 1,
    "path": "finance",
    "name": "订单财务",
    "component": "admin/finance/index",
    "meta": { "title": "订单财务", "icon": "dollar" }
  },
  {
    "parentId": 16,
    "path": "orders",
    "name": "订单管理",
    "component": "admin/finance/orders",
    "meta": { "title": "订单管理" }
  },
  {
    "parentId": 16,
    "path": "benefits",
    "name": "权益管理",
    "component": "admin/finance/benefits",
    "meta": { "title": "权益管理" }
  },
  {
    "parentId": 16,
    "path": "settlements",
    "name": "分账管理",
    "component": "admin/finance/settlements",
    "meta": { "title": "分账管理" }
  },
  {
    "parentId": 1,
    "path": "messages",
    "name": "消息中心",
    "component": "admin/message/index",
    "meta": { "title": "消息中心", "icon": "message" }
  },
  {
    "parentId": 20,
    "path": "message-list",
    "name": "消息列表",
    "component": "admin/message/list",
    "meta": { "title": "消息列表" }
  },
  {
    "parentId": 20,
    "path": "conversations",
    "name": "会话管理",
    "component": "admin/message/conversations",
    "meta": { "title": "会话管理" }
  },
  {
    "parentId": 20,
    "path": "template-messages",
    "name": "模板消息",
    "component": "admin/message/templates",
    "meta": { "title": "模板消息" }
  },
  {
    "parentId": 1,
    "path": "config",
    "name": "运营配置",
    "component": "admin/config/index",
    "meta": { "title": "运营配置", "icon": "setting" }
  },
  {
    "parentId": 24,
    "path": "i18n",
    "name": "多语言维护",
    "component": "admin/config/i18n",
    "meta": { "title": "多语言维护" }
  },
  {
    "parentId": 24,
    "path": "app-config",
    "name": "App 配置",
    "component": "admin/config/app",
    "meta": { "title": "App 配置" }
  },
  {
    "parentId": 24,
    "path": "splash-ad",
    "name": "闪屏广告",
    "component": "admin/config/splash",
    "meta": { "title": "闪屏广告" }
  },
  {
    "parentId": 24,
    "path": "banners",
    "name": "Banner 管理",
    "component": "admin/config/banners",
    "meta": { "title": "Banner 管理" }
  },
  {
    "parentId": 24,
    "path": "dict",
    "name": "全局枚举字典",
    "component": "admin/config/dict",
    "meta": { "title": "全局枚举字典" }
  },
  {
    "parentId": 1,
    "path": "audit",
    "name": "系统审计",
    "component": "admin/audit/index",
    "meta": { "title": "系统审计", "icon": "audit" }
  },
  {
    "parentId": 30,
    "path": "operation-log",
    "name": "操作日志",
    "component": "admin/audit/operation-log",
    "meta": { "title": "操作日志" }
  },
  {
    "parentId": 30,
    "path": "login-log",
    "name": "登录日志",
    "component": "admin/audit/login-log",
    "meta": { "title": "登录日志" }
  },
  {
    "parentId": 30,
    "path": "high-risk-approval",
    "name": "高风险审批",
    "component": "admin/audit/high-risk",
    "meta": { "title": "高风险审批" }
  }
]
```

---

## 2. 全量按钮权限编码表

编码风格：`admin:{module}:{action}`，与 PRD-10 已有 `config:schema:*` / `i18n:string:*` / `i18n:pack:*` 保持一致。

### 2.1 用户管理

| 权限编码 | 说明 |
|---|---|
| `admin:user:view` | 查看用户信息 |
| `admin:user:edit` | 编辑用户资料 |
| `admin:user:ban` | 封禁/解封用户 |
| `admin:user:reset_password` | 重置用户密码 |
| `admin:user:send_message` | 向用户发送站内通知 |
| `admin:user:clear_session` | 强制下线（清除会话） |

### 2.2 群组管理

| 权限编码 | 说明 |
|---|---|
| `admin:group:view` | 查看群组信息 |
| `admin:group:edit` | 编辑群组基本信息 |
| `admin:group:audit` | 审核群组创建/变更 |
| `admin:group:ban` | 封禁/解封群组 |
| `admin:group:transfer_owner` | 转让群主 |
| `admin:group:set_featured` | 设置/取消精选群组 |
| `admin:group:set_official` | 设置/取消官方认证 |

### 2.3 群组成员

| 权限编码 | 说明 |
|---|---|
| `admin:group_member:view` | 查看成员列表与详情 |
| `admin:group_member:mute` | 禁言/解除禁言 |
| `admin:group_member:remove` | 移除成员 |
| `admin:group_member:set_guest` | 设置/取消嘉宾身份 |
| `admin:group_member:set_admin` | 设置/撤销管理员 |

### 2.4 主题管理

| 权限编码 | 说明 |
|---|---|
| `admin:topic:view` | 查看主题内容 |
| `admin:topic:audit` | 审核主题（通过/拒绝） |
| `admin:topic:edit` | 编辑主题内容 |
| `admin:topic:hide` | 隐藏/恢复显示 |
| `admin:topic:delete` | 删除主题 |
| `admin:topic:pin` | 置顶/取消置顶 |
| `admin:topic:feature` | 加精/取消加精 |

### 2.5 评论管理

| 权限编码 | 说明 |
|---|---|
| `admin:comment:view` | 查看评论 |
| `admin:comment:hide` | 隐藏/恢复评论 |
| `admin:comment:delete` | 删除单条评论 |
| `admin:comment:batch_delete` | 批量删除评论 |

### 2.6 订单管理

| 权限编码 | 说明 |
|---|---|
| `admin:order:view` | 查看订单详情 |
| `admin:order:confirm_pay` | 手动确认支付 |
| `admin:order:cancel` | 取消订单 |
| `admin:order:refund` | 发起退款 |
| `admin:order:export` | 导出订单数据 |

### 2.7 分账管理

| 权限编码 | 说明 |
|---|---|
| `admin:settlement:view` | 查看分账记录 |
| `admin:settlement:generate` | 生成分账单 |
| `admin:settlement:confirm` | 确认分账金额 |
| `admin:settlement:mark_paid` | 标记已打款 |
| `admin:settlement:export` | 导出分账报表 |

### 2.8 多语言维护

| 权限编码 | 说明 |
|---|---|
| `admin:i18n:view` | 查看翻译条目 |
| `admin:i18n:edit` | 编辑翻译文本 |
| `admin:i18n:publish` | 发布翻译包 |
| `admin:i18n:rollback` | 回滚到历史版本 |

### 2.9 App 配置

| 权限编码 | 说明 |
|---|---|
| `admin:config:view` | 查看配置项 |
| `admin:config:edit` | 编辑配置值 |
| `admin:config:publish` | 发布配置版本 |
| `admin:config:rollback` | 回滚配置 |

### 2.10 广告 & Banner

| 权限编码 | 说明 |
|---|---|
| `admin:ad:view` | 查看广告/Banner |
| `admin:ad:edit` | 创建/编辑广告素材 |
| `admin:ad:publish` | 上线发布 |
| `admin:ad:offline` | 下线撤回 |

### 2.11 模板消息

| 权限编码 | 说明 |
|---|---|
| `admin:template:view` | 查看模板消息 |
| `admin:template:edit` | 编辑模板内容 |
| `admin:template:test_send` | 测试发送 |
| `admin:template:publish` | 发布生效 |

### 2.12 消息中心

| 权限编码 | 说明 |
|---|---|
| `admin:message:view` | 查看消息与会话 |
| `admin:message:retry` | 重发失败消息 |
| `admin:message:recall` | 撤回已发消息 |
| `admin:message:delete` | 删除消息记录 |

### 2.13 全局枚举字典

| 权限编码 | 说明 |
|---|---|
| `admin:dict:view` | 查看枚举字典 |
| `admin:dict:edit` | 编辑枚举值 |
| `admin:dict:import` | 批量导入 |
| `admin:dict:export` | 导出字典数据 |
| `admin:dict:refresh_cache` | 刷新缓存 |

### 2.14 操作日志

| 权限编码 | 说明 |
|---|---|
| `admin:operation_log:view` | 查看操作日志 |
| `admin:operation_log:export` | 导出日志 |
| `admin:operation_log:view_sensitive` | 查看敏感操作日志（需额外授权） |

---

## 3. 角色 - 权限矩阵

### 3.1 角色定义总览

| 后台角色 | 核心权限范围 |
|---|---|
| **超级管理员** | 全部权限（`admin:*`） |
| **运营管理员** | 用户 + 群组 + 主题 + 评论 + 广告 + 消息模板 |
| **内容审核员** | 主题审核 + 评论审核 + 违规处理 |
| **客服人员** | 查看 + 订单 + 消息 + 通知 |
| **财务人员** | 订单查看 + 退款 + 分账确认 + 导出 |
| **配置管理员** | App 配置 + 多语言 + 模板消息 + 广告 Banner |
| **只读审计员** | 仅查看操作日志 |

### 3.2 详细矩阵

| 权限分组 | 超级管理员 | 运营管理员 | 内容审核员 | 客服人员 | 财务人员 | 配置管理员 | 只读审计员 |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **用户管理** (`admin:user:*`) | ✅ 全部 | ✅ view/edit/ban | ❌ | ✅ view | ❌ | ❌ | ❌ |
| **群组管理** (`admin:group:*`) | ✅ 全部 | ✅ view/edit/audit/ban/set_featured | ❌ | ✅ view | ❌ | ❌ | ❌ |
| **群组成员** (`admin:group_member:*`) | ✅ 全部 | ✅ view/mute/remove | ❌ | ❌ | ❌ | ❌ | ❌ |
| **主题管理** (`admin:topic:*`) | ✅ 全部 | ✅ view/edit/hide/pin/feature | ✅ view/audit/hide/delete/pin | ✅ view | ❌ | ❌ | ❌ |
| **评论管理** (`admin:comment:*`) | ✅ 全部 | ✅ view/hide/delete/batch_delete | ✅ view/hide/delete/batch_delete | ✅ view | ❌ | ❌ | ❌ |
| **订单管理** (`admin:order:*`) | ✅ 全部 | ❌ | ❌ | ✅ view/cancel | ✅ view/refund/export | ❌ | ❌ |
| **分账管理** (`admin:settlement:*`) | ✅ 全部 | ❌ | ❌ | ❌ | ✅ view/confirm/mark_paid/export | ❌ | ❌ |
| **多语言维护** (`admin:i18n:*`) | ✅ 全部 | ❌ | ❌ | ❌ | ❌ | ✅ 全部 | ❌ |
| **App 配置** (`admin:config:*`) | ✅ 全部 | ❌ | ❌ | ❌ | ❌ | ✅ 全部 | ❌ |
| **广告 Banner** (`admin:ad:*`) | ✅ 全部 | ✅ 全部 | ❌ | ❌ | ❌ | ✅ 全部 | ❌ |
| **模板消息** (`admin:template:*`) | ✅ 全部 | ✅ 全部 | ❌ | ✅ view/test_send | ❌ | ✅ 全部 | ❌ |
| **消息中心** (`admin:message:*`) | ✅ 全部 | ✅ view/retry/recall | ❌ | ✅ view/retry | ❌ | ❌ | ❌ |
| **全局枚举** (`admin:dict:*`) | ✅ 全部 | ✅ view | ❌ | ❌ | ❌ | ✅ 全部 | ❌ |
| **操作日志** (`admin:operation_log:*`) | ✅ 全部 | ✅ view/export | ❌ | ❌ | ✅ view/export | ✅ view/export | ✅ view |

> 注：✅ 表示拥有该分组下的对应子权限（详见各角色说明），❌ 表示无权访问。

---

## 4. 数据权限说明

| 角色 | 数据可见范围 | 说明 |
|---|---|---|
| **超级管理员** | 全部数据 | 无任何行级过滤 |
| **运营管理员** | 所管辖范围数据 | 按区域/业务线隔离，仅能操作归属自己负责域的数据 |
| **内容审核员** | 待审核池数据 | 默认只看到待审核队列，已处理内容可按条件检索 |
| **客服人员** | 仅关联业务数据 | 仅能看到用户主动发起工单或咨询所涉及的订单/消息上下文 |
| **财务人员** | 仅关联财务数据 | 仅涉及金额相关的订单与分账记录，不含用户聊天内容 |
| **配置管理员** | 全局配置数据 | 配置类资源无行级隔离（所有环境共享同一套配置） |
| **只读审计员** | 全部日志数据 | 可查看全量操作日志用于审计追溯，但不可执行写操作 |

---

## 5. 与 PRD-10 已有权限的关系

### 5.1 已有权限点（PRD-10）

PRD-10 已实现的权限编码：

| 权限编码 | 所属模块 | 说明 |
|---|---|---|
| `config:schema:view` | 配置 Schema | 查看 Schema 定义 |
| `config:schema:edit` | 配置 Schema | 编辑 Schema |
| `i18n:string:view` | 多语言字符串 | 查看翻译条目 |
| `i18n:string:edit` | 多语言字符串 | 编辑翻译文本 |
| `i18n:pack:view` | 翻译包 | 查看翻译包 |
| `i18n:pack:publish` | 翻译包 | 发布翻译包 |

### 5.2 本文档扩展的社交域运营权限

本文档在 PRD-10 基础上新增以下社交域运营权限模块：

- **用户管理**：`admin:user:*`（6 个权限点）
- **群组管理**：`admin:group:*`（7 个权限点）
- **群组成员**：`admin:group_member:*`（5 个权限点）
- **内容管理**：`admin:topic:*`（7 个）+ `admin:comment:*`（4 个）
- **订单财务**：`admin:order:*`（5 个）+ `admin:settlement:*`（5 个）
- **消息中心**：`admin:message:*`（4 个）+ `admin:template:*`（4 个）
- **运营配置**：`admin:ad:*`（4 个）+ `admin:dict:*`（5 个）
- **系统审计**：`admin:operation_log:*`（3 个）

### 5.3 编码风格一致性

- 新增权限统一使用 `admin:{module}:{action}` 格式
- action 动词语义约定：
  - `view` — 查询/列表/详情
  - `edit` — 新增/修改
  - `delete` — 物理删除
  - `hide` — 逻辑隐藏/软删除
  - `audit` — 审核（通过/拒绝）
  - `ban` — 封禁/解封（状态切换）
  - `publish` — 发布上线
  - `rollback` — 回滚版本
  - `export` — 数据导出
  - `refresh_cache` — 缓存刷新

---

## 6. 附录

### A. 权限编码汇总清单（共 59 个）

| # | 编码 | 模块 | 动作 |
|---|---|---|---|
| 1 | `admin:user:view` | 用户管理 | 查看 |
| 2 | `admin:user:edit` | 用户管理 | 编辑 |
| 3 | `admin:user:ban` | 用户管理 | 封禁 |
| 4 | `admin:user:reset_password` | 用户管理 | 重置密码 |
| 5 | `admin:user:send_message` | 用户管理 | 发送通知 |
| 6 | `admin:user:clear_session` | 用户管理 | 清除会话 |
| 7 | `admin:group:view` | 群组管理 | 查看 |
| 8 | `admin:group:edit` | 群组管理 | 编辑 |
| 9 | `admin:group:audit` | 群组管理 | 审核 |
| 10 | `admin:group:ban` | 群组管理 | 封禁 |
| 11 | `admin:group:transfer_owner` | 群组管理 | 转让群主 |
| 12 | `admin:group:set_featured` | 群组管理 | 设置精选 |
| 13 | `admin:group:set_official` | 群组管理 | 设置官方认证 |
| 14 | `admin:group_member:view` | 群组成员 | 查看 |
| 15 | `admin:group_member:mute` | 群组成员 | 禁言 |
| 16 | `admin:group_member:remove` | 群组成员 | 移除 |
| 17 | `admin:group_member:set_guest` | 群组成员 | 设置嘉宾 |
| 18 | `admin:group_member:set_admin` | 群组成员 | 设置管理员 |
| 19 | `admin:topic:view` | 主题管理 | 查看 |
| 20 | `admin:topic:audit` | 主题管理 | 审核 |
| 21 | `admin:topic:edit` | 主题管理 | 编辑 |
| 22 | `admin:topic:hide` | 主题管理 | 隐藏 |
| 23 | `admin:topic:delete` | 主题管理 | 删除 |
| 24 | `admin:topic:pin` | 主题管理 | 置顶 |
| 25 | `admin:topic:feature` | 主题管理 | 加精 |
| 26 | `admin:comment:view` | 评论管理 | 查看 |
| 27 | `admin:comment:hide` | 评论管理 | 隐藏 |
| 28 | `admin:comment:delete` | 评论管理 | 删除 |
| 29 | `admin:comment:batch_delete` | 评论管理 | 批量删除 |
| 30 | `admin:order:view` | 订单管理 | 查看 |
| 31 | `admin:order:confirm_pay` | 订单管理 | 确认支付 |
| 32 | `admin:order:cancel` | 订单管理 | 取消 |
| 33 | `admin:order:refund` | 订单管理 | 退款 |
| 34 | `admin:order:export` | 订单管理 | 导出 |
| 35 | `admin:settlement:view` | 分账管理 | 查看 |
| 36 | `admin:settlement:generate` | 分账管理 | 生成 |
| 37 | `admin:settlement:confirm` | 分账管理 | 确认 |
| 38 | `admin:settlement:mark_paid` | 分账管理 | 标记已付 |
| 39 | `admin:settlement:export` | 分账管理 | 导出 |
| 40 | `admin:i18n:view` | 多语言维护 | 查看 |
| 41 | `admin:i18n:edit` | 多语言维护 | 编辑 |
| 42 | `admin:i18n:publish` | 多语言维护 | 发布 |
| 43 | `admin:i18n:rollback` | 多语言维护 | 回滚 |
| 44 | `admin:config:view` | App 配置 | 查看 |
| 45 | `admin:config:edit` | App 配置 | 编辑 |
| 46 | `admin:config:publish` | App 配置 | 发布 |
| 47 | `admin:config:rollback` | App 配置 | 回滚 |
| 48 | `admin:ad:view` | 广告Banner | 查看 |
| 49 | `admin:ad:edit` | 广告Banner | 编辑 |
| 50 | `admin:ad:publish` | 广告Banner | 发布 |
| 51 | `admin:ad:offline` | 广告Banner | 下线 |
| 52 | `admin:template:view` | 模板消息 | 查看 |
| 53 | `admin:template:edit` | 模板消息 | 编辑 |
| 54 | `admin:template:test_send` | 模板消息 | 测试发送 |
| 55 | `admin:template:publish` | 模板消息 | 发布 |
| 56 | `admin:message:view` | 消息中心 | 查看 |
| 57 | `admin:message:retry` | 消息中心 | 重发 |
| 58 | `admin:message:recall` | 消息中心 | 撤回 |
| 59 | `admin:message:delete` | 消息中心 | 删除 |
| 60 | `admin:dict:view` | 全局枚举 | 查看 |
| 61 | `admin:dict:edit` | 全局枚举 | 编辑 |
| 62 | `admin:dict:import` | 全局枚举 | 导入 |
| 63 | `admin:dict:export` | 全局枚举 | 导出 |
| 64 | `admin:dict:refresh_cache` | 全局枚举 | 刷新缓存 |
| 65 | `admin:operation_log:view` | 操作日志 | 查看 |
| 66 | `admin:operation_log:export` | 操作日志 | 导出 |
| 67 | `admin:operation_log:view_sensitive` | 操作日志 | 敏感查看 |

### B. 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|---|---|---|---|
| v1.0 | 2026-06-15 | 初始版本，定义 7 大模块 67 个权限编码、7 种角色矩阵 | Trae |
