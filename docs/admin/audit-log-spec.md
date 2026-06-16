# 运营后台操作审计日志规范

> 本文件负责定义 CaiRobot MVP 运营后台（Admin）的操作审计日志规范，包括表结构、必记操作清单、风险分级、脱敏规则、保留策略及与 go-admin 自带日志的关系。
> 所属模块：运营后台（provider-admin）
> 依赖文档：
> - PRD-10-Admin管理后台MVP.md（运营后台功能范围）
> - ADR-010-admin-boundary-sdk.md（Admin 边界与 SDK 引用规范）
>
> 不处理的职责边界：
> - 不定义审计日志的查询 API 接口（由 PRD-10 后续迭代补充）
> - 不定义前端审计日志展示页面
> - 不处理合规性报告生成（如 GDPR / 等保审计报告）
> - 不涉及用户行为埋点与统计分析日志

---

## 一、审计日志表 DDL（admin_operation_logs）

```sql
-- ============================================================
-- admin_operation_logs：运营后台操作审计日志表
-- 职责：记录管理员在运营后台执行的所有高权限业务操作
-- 引擎：InnoDB，字符集 utf8mb4
-- ============================================================

CREATE TABLE `admin_operation_logs` (
    -- 主键，使用 char32 保证全局唯一且有序
    `id`                    CHAR(32)     NOT NULL DEFAULT '' COMMENT '主键ID（UUID去横杠）',

    -- 操作人信息
    `admin_user_id`         CHAR(32)     NOT NULL DEFAULT '' COMMENT '操作管理员用户ID',
    `admin_username`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作管理员用户名',
    `admin_role`            VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '操作管理员角色（admin/operator/viewer等）',

    -- 操作分类信息
    `module`                VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '所属模块（user/group/topic/comment/payment/config/advert/message/enum等）',
    `action`                VARCHAR(80)  NOT NULL DEFAULT '' COMMENT '操作动作（disable_user/enable_user/reset_password等）',

    -- 操作对象信息
    `target_type`           VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '目标对象类型',
    `target_id`             CHAR(32)     NOT NULL DEFAULT '' COMMENT '目标对象ID',
    `target_name`           VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目标对象名称（用于可读展示）',

    -- 变更数据快照（JSON 格式记录变更前后状态）
    `before_data`           JSON         DEFAULT NULL COMMENT '变更前数据快照',
    `after_data`            JSON         DEFAULT NULL COMMENT '变更后数据快照',
    `changed_fields`        JSON         DEFAULT NULL COMMENT '变更字段列表（field: {old, new}）',

    -- 操作备注与原因
    `reason`                VARCHAR(500) NOT NULL DEFAULT '' COMMENT '操作原因（管理员填写）',
    `remark`                VARCHAR(500) NOT NULL DEFAULT '' COMMENT '操作备注',

    -- 请求上下文信息
    `request_ip`            VARCHAR(45)  NOT NULL DEFAULT '' COMMENT '请求来源IP（支持IPv6）',
    `user_agent`            VARCHAR(500) NOT NULL DEFAULT '' COMMENT '请求User-Agent',
    `request_path`          VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
    `request_method`        VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '请求方法（GET/POST/PUT/DELETE等）',
    `request_params`        JSON         DEFAULT NULL COMMENT '请求参数（已脱敏）',

    -- 响应结果信息
    `response_status`       VARCHAR(30)  NOT NULL DEFAULT '' COMMENT 'HTTP响应状态码或业务状态',
    `error_message`         TEXT         DEFAULT NULL COMMENT '错误信息（操作失败时记录）',

    -- 风险控制字段
    `risk_level`            VARCHAR(20)  NOT NULL DEFAULT 'low' COMMENT '风险等级（low/medium/high/critical）',
    `is_sensitive`          TINYINT(1)   NOT NULL DEFAULT 0  COMMENT '是否为敏感操作（0否/1是）',
    `require_confirm`       TINYINT(1)   NOT NULL DEFAULT 0  COMMENT '是否需要二次确认（0否/1是）',
    `confirmed`             TINYINT(1)   NOT NULL DEFAULT 0  COMMENT '是否已完成二次确认（0未确认/1已确认）',
    `confirmed_by`          CHAR(32)     NOT NULL DEFAULT '' COMMENT '二次确认人ID',
    `confirmed_at`          BIGINT       NOT NULL DEFAULT 0  COMMENT '二次确认时间戳（毫秒）',

    -- 链路追踪字段
    `trace_id`              VARCHAR(100) NOT NULL DEFAULT '' COMMENT '分布式链路追踪ID',
    `request_id`            VARCHAR(100) NOT NULL DEFAULT '' COMMENT '请求唯一标识ID',

    -- 时间字段
    `created_at`            BIGINT       NOT NULL DEFAULT 0  COMMENT '创建时间戳（毫秒）',

    PRIMARY KEY (`id`),

    -- 按管理员+时间索引：用于按操作人查询历史记录
    INDEX `idx_admin_user_time` (`admin_user_id`, `created_at`),
    -- 按模块+动作+时间索引：用于按业务模块筛选操作
    INDEX `idx_module_action_time` (`module`, `action`, `created_at`),
    -- 按目标对象索引：用于追溯某对象的全部操作历史
    INDEX `idx_target` (`target_type`, `target_id`),
    -- 按风险等级+时间索引：用于安全审计和风险排查
    INDEX `idx_risk_level_time` (`risk_level`, `created_at`),
    -- 按链路追踪ID索引：用于关联一次完整请求链路
    INDEX `idx_trace_id` (`trace_id`),
    -- 按请求ID索引：用于精确定位单次请求的审计记录
    INDEX `idx_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='运营后台操作审计日志表';
```

---

## 二、必须记录审计的操作清单（共 29 项）

以下操作必须在 `admin_operation_logs` 中产生审计记录。每条记录必须包含完整的 before_data / after_data 快照。

### 2.1 用户管理类（3 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 1 | 禁用用户 | user | disable_user | user | 冻结用户账号，禁止登录 |
| 2 | 启用用户 | user | enable_user | user | 解冻用户账号，恢复登录权限 |
| 3 | 重置密码 | user | reset_password | user | 强制重置用户登录密码 |

### 2.2 群组管理类（8 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 4 | 禁用群组 | group | disable_group | group | 停用群组，成员无法交互 |
| 5 | 启用群组 | group | enable_group | group | 恢复群组正常使用 |
| 6 | 转让群主 | group | transfer_owner | group | 将群主身份转让给其他成员 |
| 7 | 设置管理员 | group | set_admin | group_member | 设置群组成员为管理员 |
| 8 | 移除管理员 | group | remove_admin | group_member | 取消成员的管理员身份 |
| 9 | 设置嘉宾 | group | set_guest | group_member | 设置群组成员为嘉宾 |
| 10 | 移除嘉宾 | group | remove_guest | group_member | 取消成员的嘉宾身份 |
| 11 | 禁言成员 | group | mute_member | group_member | 对群组成员执行禁言 |

### 2.3 内容管理类（7 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 12 | 删除主题 | topic | delete_topic | topic | 物理删除主题内容 |
| 13 | 下架主题 | topic | offline_topic | topic | 下架主题使其不可见 |
| 14 | 恢复主题 | topic | restore_topic | topic | 恢复已下架/删除的主题 |
| 15 | 删除评论 | comment | delete_comment | comment | 物理删除评论内容 |
| 16 | 隐藏评论 | comment | hide_comment | comment | 隐藏评论使其对普通用户不可见 |
| 17 | 撤回消息 | message | recall_message | message | 撤回已发送的消息 |
| 18 | 删除消息 | message | delete_message | message | 物理删除消息 |

### 2.4 成员管理类（1 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 19 | 移除成员 | group | remove_member | group_member | 将成员从群组中移除 |

### 2.5 支付与资金类（4 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 20 | 确认支付 | payment | confirm_payment | order | 手动确认支付订单状态 |
| 21 | 退款 | payment | refund | order | 对订单执行退款操作 |
| 22 | 确认分账 | payment | confirm_split | settlement | 确认分账结算 |
| 23 | 标记打款 | payment | mark_transfer | withdrawal | 标记打款完成 |

### 2.6 配置与发布类（4 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 24 | 发布App配置 | config | publish_app_config | app_config | 发布App配置到生产环境 |
| 25 | 回滚App配置 | config | rollback_app_config | app_config | 回滚App配置到指定版本 |
| 26 | 发布广告 | advert | publish_advert | advert | 上线广告投放 |
| 27 | 下架广告 | advert | offline_advert | advert | 下线广告投放 |

### 2.7 消息与系统类（2 项）

| 序号 | 操作名称 | module | action | target_type | 说明 |
|---:|---|---|---|---|---|
| 28 | 发布模板消息 | message | publish_template_msg | template_msg | 发布模板消息配置 |
| 29 | 修改全局枚举 | enum | update_global_enum | global_enum | 修改全局枚举值定义 |

---

## 三、高风险操作二次确认规则

以下矩阵规定了各操作的 **风险等级** 和 **是否需要二次确认**。二次确认指操作提交前需另一名同级或更高级别管理员审批通过。

### 3.1 风险等级定义

| 风险等级 | 定义 | 典型场景 |
|---|---|---|
| low | 低风险，常规运维操作，无需额外确认 | 查询类、低影响配置变更 |
| medium | 中等风险，可能影响用户体验，建议确认 | 内容下架/恢复、广告上下架 |
| high | 高风险，直接影响用户资产或账号状态 | 用户禁用/启用、退款、资金操作 |
| critical | 极高风险，不可逆或影响面广，强制双人确认 | 密码重置、数据删除、配置发布到生产 |

### 3.2 操作风险矩阵

| 操作 | risk_level | require_confirm | confirmed 默认值 | 说明 |
|---|---|---:|---:|---|
| **禁用用户** | high | ✅ 是 | 0 | 影响用户正常使用，需确认 |
| **启用用户** | high | ✅ 是 | 0 | 可能误启封禁账号，需确认 |
| **重置密码** | critical | ✅ 是 | 0 | 涉及用户凭证安全，强制双人确认 |
| **禁用群组** | high | ✅ 是 | 0 | 影响群组内所有成员 |
| **启用群组** | high | ✅ 是 | 0 | 可能恢复违规群组 |
| **转让群主** | high | ✅ 是 | 0 | 权限变更，影响群组管理权 |
| **设置管理员** | medium | ❌ 否 | 1 | 群组内部权限调整 |
| **移除管理员** | medium | ❌ 否 | 1 | 群组内部权限调整 |
| **设置嘉宾** | low | ❌ 否 | 1 | 仅展示性质变更 |
| **移除嘉宾** | low | ❌ 否 | 1 | 仅展示性质变更 |
| **禁言成员** | medium | ❌ 否 | 1 | 可逆操作，影响有限 |
| **移除成员** | high | ✅ 是 | 0 | 影响用户社交关系 |
| **删除主题** | critical | ✅ 是 | 0 | 不可逆，数据永久丢失 |
| **下架主题** | medium | ❌ 否 | 1 | 可逆操作（可恢复） |
| **恢复主题** | medium | ❌ 否 | 1 | 可逆操作 |
| **删除评论** | high | ✅ 是 | 0 | 不可逆，数据永久丢失 |
| **隐藏评论** | medium | ❌ 否 | 1 | 可逆操作（可取消隐藏） |
| **确认支付** | high | ✅ 是 | 0 | 涉及资金状态变更 |
| **退款** | critical | ✅ 是 | 0 | 涉及资金流出，强制双人确认 |
| **确认分账** | high | ✅ 是 | 0 | 涉及资金分配 |
| **标记打款** | high | ✅ 是 | 0 | 涉及资金流出 |
| **发布App配置** | critical | ✅ 是 | 0 | 影响全量App用户，不可轻率发布 |
| **回滚App配置** | critical | ✅ 是 | 0 | 同样影响全量用户 |
| **发布广告** | medium | ❌ 否 | 1 | 可逆操作（可下架） |
| **下架广告** | medium | ❌ 否 | 1 | 可逆操作 |
| **发布模板消息** | high | ✅ 是 | 0 | 影响全量用户触达 |
| **修改全局枚举** | high | ✅ 是 | 0 | 影响系统核心逻辑分支 |
| **撤回消息** | medium | ❌ 否 | 1 | 可逆操作，时效窗口有限 |
| **删除消息** | high | ✅ 是 | 0 | 不可逆，数据永久丢失 |

---

## 四、脱敏规则

审计日志中不得明文存储敏感信息。以下规则适用于 `before_data`、`after_data`、`request_params` 中所有字段的写入前脱敏。

### 4.1 脱敏规则总表

| 数据类型 | 脱敏策略 | 示例 | 适用场景 |
|---|---|---|---|
| 密码（password） | **不记录**，字段值为 `null` 或 `"***REDACTED***"` | — | 所有包含密码的字段 |
| 密码盐值（salt） | **不记录** | — | 加密相关字段 |
| Token / JWT / Session ID | **不记录**，仅记录是否存在（`has_token: true/false`） | — | 认证凭据字段 |
| 手机号 | 中间四位掩码 | `138****5678` | 用户手机号字段 |
| 邮箱 | 用户名部分掩码，保留域名 | `j***e@example.com` | 用户邮箱字段 |
| 身份证号 | 仅保留后四位 | `************1234` | 实名认证字段 |
| 银行卡号 | 仅保留后四位 | `************5678` | 支付相关银行卡字段 |
| 私信正文 | 默认记录摘要（前 50 字符 + `...`），拥有「审计全文」权限的管理员可查看原文 | `"这是一条测试消息内容..."` | message 表 body 字段 |
| App 敏感配置 | 仅记录 key 名称 + 变更动作（create/update/delete），不记录 value | `{"key": "api_secret", "action": "update"}` | 含 secret/key/token 的配置项 |
| JSON 配置对象 | **递归脱敏**：遍历 JSON 所有叶子节点，命中上述规则的按对应策略处理 | — | before_data / after_data 整体 |

### 4.2 递归脱敏算法说明

```
函数 desensitizeJson(input: any, rules: DesensitizeRule[]): any

输入：任意 JSON 值 + 脱敏规则列表
输出：脱敏后的 JSON 值

算法步骤：
1. 若 input 为基本类型（string/number/boolean/null），直接返回
2. 若 input 为数组，对每个元素递归调用 desensitizeJson
3. 若 input 为对象（Object/Map）：
   a. 遍历每个 key-value 对
   b. 检查 key 是否匹配敏感字段名模式（如 *password*, *secret*, *token*, *salt* 等）
   c. 若匹配，按规则表中对应策略脱敏
   d. 若不匹配且 value 为嵌套对象/数组，递归调用 desensitizeJson
   e. 若不匹配且 value 为基本类型，保持原值
4. 返回脱敏后的副本（不修改原对象）
```

### 4.3 敏感字段名匹配模式

以下字段名（不区分大小写）触发自动脱敏：

```text
password, passwd, pwd, secret, token, session,
salt, credential, api_key, private_key,
access_token, refresh_token, id_card, idcard,
bank_card, bankcard, card_no, phone, mobile,
email, id_number, ssn, cvv, cvc
```

---

## 五、日志保留与归档策略

### 5.1 在线保留期

| 项目 | 规则 |
|---|---|
| **在线保留时长** | 最近 **90 天** |
| **存储位置** | MySQL 主库 `admin_operation_logs` 表 |
| **查询能力** | 支持全条件组合查询（按管理员/模块/时间/风险等级/目标对象等） |
| **性能要求** | 单次查询响应 < 500ms（合理索引覆盖下） |

### 5.2 归档策略

| 项目 | 规则 |
|---|---|
| **归档触发** | 定时任务每日凌晨检查，将 `created_at` 距今超过 90 天的记录归档 |
| **归档目标** | 冷存储（推荐对象存储 OSS / S3 兼容存储） |
| **归档格式** | 按月分区 Parquet 文件：`audit_logs/YYYY/MM/admin_operation_logs_YYYYMM.parquet` |
| **归档后处理** | 主库中删除已归档记录，释放存储空间 |
| **归档查询** | 提供独立的归档查询接口（不在主库 SQL 层面支持），按 trace_id 或 request_id 精确检索 |

### 5.3 归档生命周期

```
第 0 ~ 90 天：在线存储（MySQL 主库）→ 全条件实时查询
第 91 天 ~ 1 年：冷存储（对象存储）→ 按 trace_id/request_id 精确查询
超过 1 年：进入深度归档 → 仅在合规审计/法律要求时按申请提取
```

---

## 六、与 go-admin 自带 sys_oper_log 的关系说明

### 6.1 两套日志定位对比

| 维度 | go-admin sys_oper_log（自带） | admin_operation_logs（本规范新增） |
|---|---|---|
| **记录目标** | 基础 CRUD 操作日志 | 高权限业务操作审计日志 |
| **典型场景** | 新增/编辑/删除字典、角色、菜单等基础配置 | 禁用用户、退款、发布生产配置等业务操作 |
| **数据粒度** | 操作概要（谁在什么时间做了什么操作） | 完整变更快照（变更前后的完整数据） |
| **风险分级** | 无 | 四级风险分级（low/medium/high/critical） |
| **二次确认** | 不支持 | 支持 require_confirm / confirmed 流程 |
| **脱敏处理** | 不处理 | 强制递归脱敏 |
| **归档策略** | 无（长期留存主库） | 90 天在线 + 冷存储归档 |
| **合规用途** | 运维排查 | 安全审计、合规取证、事故复盘 |

### 6.2 共存原则

1. **两套日志并存，互不替代**
   - `sys_oper_log` 继续记录 go-admin 框架层的所有 CRUD 操作
   - `admin_operation_logs` 记录本规范定义的 29 项高权限业务操作

2. **写入时机不同**
   - `sys_oper_log` 由 go-admin 操作日志中间件自动写入（拦截所有 Controller 方法调用）
   - `admin_operation_logs` 由业务 Service 层显式调用审计写入方法（仅在上述 29 项操作触发时写入）

3. **关联方式**
   - 通过 `trace_id` 和 `request_id` 关联同一请求在两套日志中的记录
   - 一次「禁用用户」操作会同时产生一条 `sys_oper_log`（框架层）和一条 `admin_operation_logs`（业务层）

4. **查询入口分离**
   - 运营日常查看：使用 `sys_oper_log` 页面（go-admin 自带前端）
   - 安全审计调查：使用 `admin_operation_logs` 专用查询界面（后续迭代实现）

---

## 七、附录

### 7.1 审计日志写入伪代码参考

```go
// AuditService.WriteOperationLog 写入操作审计日志
// 输入：操作上下文 ctx + 审计记录 dto
// 输出：写入结果 error
// 前置条件：dto 必须经过 DTO 校验器验证
// 失败行为：写入失败不影响业务操作成功，但需记录告警日志
func (s *AuditService) WriteOperationLog(ctx context.Context, dto *AuditLogDTO) error {
    // 1. 脱敏处理：递归脱敏 before_data / after_data / request_params
    sanitizedBefore := desensitizeJson(dto.BeforeData, sensitiveRules)
    sanitizedAfter  := desensitizeJson(dto.AfterData,  sensitiveRules)
    sanitizedParams := desensitizeJson(dto.RequestParams, sensitiveRules)

    // 2. 补充默认字段
    if dto.RiskLevel == "" {
        dto.RiskLevel = resolveRiskLevel(dto.Module, dto.Action)
    }
    if dto.RequireConfirm == nil {
        requireConf := isRequireConfirm(dto.Module, dto.Action)
        dto.RequireConfirm = &requireConf
    }

    // 3. 异步写入（不阻塞主流程）
    go func() {
        err := s.repo.Insert(&model.AdminOperationLog{
            ID:             generateUUID(),
            AdminUserID:    dto.AdminUserID,
            AdminUsername:  dto.AdminUsername,
            AdminRole:      dto.AdminRole,
            Module:         dto.Module,
            Action:         dto.Action,
            TargetType:     dto.TargetType,
            TargetID:       dto.TargetID,
            TargetName:     dto.TargetName,
            BeforeData:     sanitizedBefore,
            AfterData:      sanitizedAfter,
            ChangedFields:  dto.ChangedFields,
            Reason:         dto.Reason,
            Remark:         dto.Remark,
            RequestIP:      getClientIP(ctx),
            UserAgent:      ctx.Request.UserAgent(),
            RequestPath:    ctx.Request.URL.Path,
            RequestMethod:  ctx.Request.Method,
            RequestParams:  sanitizedParams,
            ResponseStatus: dto.ResponseStatus,
            ErrorMessage:   dto.ErrorMessage,
            RiskLevel:      dto.RiskLevel,
            IsSensitive:    dto.IsSensitive,
            RequireConfirm: *dto.RequireConfirm,
            Confirmed:      false, // 默认待确认
            TraceID:        getTraceID(ctx),
            RequestID:      getRequestID(ctx),
            CreatedAt:      time.Now().UnixMilli(),
        })
        if err != nil {
            // 写入失败仅记录告警，不回滚业务操作
            logger.Error("审计日志写入失败", "error", err, "trace_id", dto.TraceID)
        }
    }()

    return nil
}
```

### 7.2 变更字段计算逻辑

```go
// calculateChangedFields 计算变更前后差异字段
// 输入：变更前数据 before + 变更后数据 after（均为 map[string]interface{}）
// 输出：变更字段 map，格式 {fieldName: {old: oldValue, new: newValue}}
// 只记录发生变化的字段，未变化字段不输出
func calculateChangedFields(before, after map[string]interface{}) map[string]interface{} {
    changed := make(map[string]interface{})
    allKeys := unionKeys(before, after)

    for _, key := range allKeys {
        oldVal, hasOld := before[key]
        newVal, hasAfter := after[key]

        // 新增字段（after 有，before 没有）
        if !hasOld && hasAfter {
            changed[key] = map[string]interface{}{
                "old": nil,
                "new": newVal,
            }
            continue
        }

        // 删除字段（before 有，after 没有）
        if oldVal && !hasAfter {
            changed[key] = map[string]interface{}{
                "old": oldVal,
                "new": nil,
            }
            continue
        }

        // 修改字段（值不相等）
        if !reflect.DeepEqual(oldVal, newVal) {
            changed[key] = map[string]interface{}{
                "old": oldVal,
                "new": newVal,
            }
        }
    }

    return changed
}
```
