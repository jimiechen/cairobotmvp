# 全局配置元模型 + 多语言参数化模板 — 实施方案

**文档编号**: IMP-CONFIG-I18N-001
**版本**: v1.0
**日期**: 2026-05-22
**状态**: 待评审
**基于方案**: v2（配置元模型+i18n模板）、v3（SDK+admin边界）、PRD（现状诊断）

---

## 0. 执行摘要

### 核心结论：这是"从零构建"，而非"迁移改造"

三份输入文档（v2/v3/PRD）描述的背景是 **MineplanetGo 旧项目的迁移场景**——假设已有 `services/config`、`services/i18n`、admin-server、6001-6007 协议等完整运行系统。

但经对 CaiRobot MVP 代码库的实际调研，当前项目处于 **S0 骨架阶段**：

| 能力 | 方案假设 | 实际现状 | 差距 |
|---|---|---|---|
| services/config 服务 | 已存在，需改造 | ❌ 不存在，需从零创建 | **全量新建** |
| services/i18n 服务 | 已存在，需改造 | ❌ 不存在，需从零创建 | **全量新建** |
| proto/app_config.proto | 已存在，需扩展 | ❌ 不存在 | **全量新建** |
| proto/i18n.proto | 已存在，需扩展 | ❌ 不存在 | **全量新建** |
| 6001/6003/6005/6007 协议 | 已在 routes.yaml 运行中 | ❌ 未注册 | **从零注册** |
| admin-server | 已存在，需加 API | ❌ 仅占位 README | **全量新建** |
| configsdk / i18nsdk | 需新增包 | ❌ 不存在 | **全量新建** |
| sys_config_schema 表 | 需新建 | ❌ 不存在 | **DDL 新建** |
| sys_lang_string 扩展字段 | 需 ALTER | ❌ 表不存在 | **随建表一并设计** |

**本方案将三份文档的架构设计思想，适配到 CaiRobot MVP 的 S0 骨架上，以"从零构建"的方式分阶段交付。**

---

## 1. 项目现状基线（调研结论）

### 1.1 已有资产

| 资产 | 路径 | 状态 | 说明 |
|---|---|---|---|
| 网关 Gateway | `go/gateway/proto-gateway/` | ✅ 可运行 | HTTP Server + routes.yaml + TarsClient |
| MessagePacket 协议 | `proto/base/message.proto` | ✅ 已定义 | maxType/minType/extend/platform/data |
| Result 通用响应 | `proto/base/result.proto` | ✅ 已定义 | code/message/PageInfo/ErrorDetail |
| HealthCheck 协议 | `proto/base/health.proto` | ✅ 已启用 | 2100:2097/2098 |
| HelloWorld 协议 | `proto/base/hello.proto` | ✅ 草案 | 2100:2101/2102 |
| System Tars 服务 | `go/tars/system/` | ✅ 可运行 | HealthCheck + HelloWorld 实现 |
| TarsGo 框架 | `go/third_party/TarsGo/` | ✅ 就绪 | 1.4.6 完整框架 |
| go.work 工作区 | `go/go.work` | ✅ 已配置 | 6 个 module |
| 协议编号注册表 | `docs/api/协议编号注册表.md` | ✅ 已建立 | 当前仅 4 条记录 |
| CODE-WIKI | `docs/wiki/CODE-WIKI.md` | ✅ 已建立 | 17 章节骨架 |
| 路由配置 | `configs/gateway/routes.yaml` | ✅ 已配置 | 2 条路由 |

### 1.2 占位资产（仅有 README）

| 资产 | 路径 | 说明 |
|---|---|---|
| ai-bridge | `go/tars/ai-bridge/` | AI 桥接服务占位 |
| audit | `go/tars/audit/` | 审计服务占位 |
| auth | `go/tars/auth/` | 认证服务占位 |
| device-gateway | `go/tars/device-gateway/` | 设备网关占位 |
| open-platform | `go/tars/open-platform/` | 开放平台占位 |
| provider-admin | `go/tars/provider-admin/` | 供应商管理后台占位 |
| user-center | `go/tars/user-center/` | 用户中心占位 |

### 1.3 缺失资产（本次实施范围）

所有与"全局配置 + 多语言"相关的资产均不存在，需要从零创建。

---

## 2. 设计目标（继承自 v2 方案 G1-G6，适配 MVP）

### 2.1 功能目标

| 编号 | 目标 | MVP 优先级 | 说明 |
|---|---|---|---|
| G1 | 后台新增配置字段无需改 Go 代码 | P0 | 核心价值：运营自助 |
| G2 | 客户端通过统一协议拿到任意结构配置 | P0 | dynamic_modules 容器 |
| G3 | 后台新增多语言 key 即时生效 | P0 | 增量协议 6007 |
| G4 | 支持参数化模板（named/icu） | P1 | MVP 先实现 named，icu 后续 |
| G5 | 不破坏已有 2100:x 协议 wire 兼容 | P0 | 6000 段是全新段，天然兼容 |
| G6 | 不引入新进程、不改部署模式 | P0 | 复用现有 monolith 结构 |

### 2.2 架构目标（继承自 v3 方案）

| 编号 | 目标 | MVP 优先级 |
|---|---|---|
| G7 | 统一 SDK（configsdk + i18nsdk）供业务服务引用 | P1（阶段 B.5） |
| G8 | admin-server 边界清晰：仅写路径，不在业务读路径 | P1（阶段 C） |
| G9 | SDK 三层缓存 + Redis pub/sub 热更新 | P2（阶段 B.5 后续） |
| G10 | Python SDK 对等接口（AI 服务用） | P2（远期） |

---

## 3. 总体架构

### 3.1 三条数据流（继承自 v3 §1.1，适配 MVP）

```
┌─────────────────────────────────────────────────────────────┐
│                   写路径（运维侧，低频）                      │
│  admin-web (Vue/React)                                       │
│        │ HTTP                                                │
│        ▼                                                     │
│  admin-server (Go/Gin)         ← 阶段 C 从零构建             │
│        │ 直接读写 MySQL                                      │
│        ▼                                                     │
│  sys_config_schema / sys_config_version                      │
│  sys_lang_pack / sys_lang_string                             │
│        │ 同时主动失效 Redis 缓存                             │
│        ▼                                                     │
│  Redis (config:* / i18n:*)                                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                客户端读路径（高频，走网关）                   │
│  App / Web / 第三方                                          │
│        │ POST /api/hello + MessagePacket                     │
│        ▼                                                     │
│  proto-gateway (已有)                                        │
│        │ TarsGo bytes                                        │
│        ▼                                                     │
│  tars/config (新建) / tars/i18n (新建)                       │
│        │ 调用                                                │
│        ▼                                                     │
│  services/config (新建) / services/i18n (新建)               │
│        │                                                     │
│        ▼                                                     │
│  Redis → MySQL                                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│            服务端业务读路径（阶段 B.5 构建）                  │
│  OpenAPI / 设备网关 / 用户中台 / AI 服务 / Gateway 自身       │
│        │ Go 函数调用                                         │
│        ▼                                                     │
│  configsdk.Client / i18nsdk.Client   ← 阶段 B.5 新建         │
│        │ 内存 LRU + Redis + 远程兜底                         │
│        ▼                                                     │
│  services/config / services/i18n                             │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 与现有骨架的集成点

```
                    现有资产                          本次新建
                    ────                            ────────
proto/base/
  message.proto     ✅ 已有                           （不变）
  result.proto      ✅ 已有                           （不变）
  app_config.proto  ❌ 不存在                        ⭐ 新建
  i18n.proto        ❌ 不存在                        ⭐ 新建

go/gateway/proto-gateway/
  http_server.go    ✅ 已有                           （不变）
  routes.go         ✅ 已有                           （增加 6000 段校验）
  route_table.go    ✅ 已有                           （不变）

configs/gateway/
  routes.yaml       ✅ 已有（2条路由）               ⭐ 增加 6001-6009

go/tars/
  system/           ✅ 已有                           （不变）
  config/           ❌ 仅 README                     ⭐ 新建 Tars Servant
  i18n/             ❌ 仅 README                     ⭐ 新建 Tars Servant

go/services/
  (空目录,仅骨架)    ✅ 有 go.mod                     ⭐ 新建 config + i18n module

go/services/config/
  domain/                                             ⭐ 新建
  repository/                                         ⭐ 新建
  cache/                                              ⭐ 新建
  service/                                            ⭐ 新建
  sdk/                                                ⭐ 阶段 B.5 新建

go/services/i18n/
  domain/                                             ⭐ 新建
  repository/                                         ⭐ 新建
  cache/                                              ⭐ 新建
  service/                                            ⭐ 新建
  sdk/                                                ⭐ 阶段 B.5 新建

docs/api/
  协议编号注册表.md   ✅ 已有（4条）                 ⭐ 登记 6000 段编号
```

---

## 4. 协议设计

### 4.1 协议编号分配（6000 段：App/Web 前端交互协议）

根据 CODE-WIKI §4 的编号范围约定，6000-6999 属于「App、Web、前端交互协议」。

| max | min | 方向 | 报文类型 | Proto 文件 | Message | 状态 | 说明 |
|---:|---:|---|---|---|---|---|---|
| 6000 | 6001 | C->S | Request | `base/app_config.proto` | `AppConfigsReq` | 新增 | 获取全量应用配置 |
| 6000 | 6002 | S->C | Response | `base/app_config.proto` | `AppConfigsRsp` | 新增 | 应用配置响应（含 dynamic_modules） |
| 6000 | 6003 | C->S | Request | `base/i18n.proto` | `AppFetchLanguageReq` | 新增 | 获取语言元数据 |
| 6000 | 6004 | S->C | Response | `base/i18n.proto` | `AppFetchLanguageRsp` | 新增 | 语言元数据响应 |
| 6000 | 6005 | C->S | Request | `base/i18n.proto` | `AppFetchLangPackReq` | 新增 | 获取全量语言包 |
| 6000 | 6006 | S->C | Response | `base/i18n.proto` | `AppFetchLangPackRsp` | 新增 | 全量语言包响应 |
| 6000 | 6007 | C->S | Request | `base/i18n.proto` | `AppFetchLangDifferenceReq` | 新增 | 获取增量语言包 |
| 6000 | 6008 | S->C | Response | `base/i18n.proto` | `AppFetchLangDifferenceRsp` | 新增 | 增量语言包响应 |
| 6000 | 6009 | C->S | Request | `base/app_config.proto` | `AppConfigVersionReq` | 新增 | 配置/语言包版本轮询 |
| 6000 | 6010 | S->C | Response | `base/app_config.proto` | `AppConfigVersionRsp` | 新增 | 版本轮询响应 |

> **注意**: 原 PRD 中 6001-6008 的编号在此重新梳理为 6001-6010（Request/Response 分离），确保每个报文独立编号。

### 4.2 AppConfigsRsp Protobuf 定义（继承 v2 §2.3）

```protobuf
// proto/base/app_config.proto
syntax = "proto3";
package com.mineplanet.pojo;
option go_package = "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base";

import "base/result.proto";

// 请求：获取全量应用配置
message AppConfigsReq {
  // env 目标环境，不传则由网关根据请求来源判断
  string env = 1;
  // client_scope 客户端范围过滤
  string client_scope = 2;
  // client_version 客户端版本号，用于兼容性过滤
  string client_version = 3;
  // requested_modules 请求的模块列表，为空则返回全部
  repeated string requested_modules = 4;
}

// 响应：应用配置（含自描述动态模块）
message AppConfigsRsp {
  Result result = 1;

  // 强类型模块字段（按 schema 自动填充，兼容未来可能的老客户端）
  AppBaseConfigs    base_cfg   = 2;
  AppWapUrlConfigs  wap_cfg    = 3;
  AppRegexConfigs   regex_cfg  = 4;
  AppPayConfigs     pay_cfg    = 5;
  AppOssConfigs     oss_cfg    = 6;
  AppLanguageConfigs lang_cfg  = 7;
  AppMuteConfigs    mute_cfg   = 8;
  AppGroupConfigs   group_cfg  = 9;

  // 自描述动态模块容器（新模块/新字段都走这里）
  repeated DynamicConfigModule dynamic_modules = 100;
}

// 动态配置模块：module_key + version + 字段 map + schema 摘要
message DynamicConfigModule {
  string module_key = 1;
  int64  version    = 2;
  // 字段 → JSON-encoded value，类型由 schema 描述
  map<string, string> fields = 3;
  // 内嵌 schema 摘要，便于客户端调试
  repeated FieldDescriptor descriptors = 4;
}

// 字段描述符
message FieldDescriptor {
  string field_key   = 1;
  string field_type  = 2;  // string/int/bool/float/enum/json/list
  bool   is_required = 3;
  string default_val = 4;
}

// ===== 以下为强类型消息定义（初始 8 个模块） =====
message AppBaseConfigs {
  string domain_root = 1;
  string domain_wap  = 2;
  string sign_rand   = 3;
  string construct_email = 4;
}

message AppWapUrlConfigs {
  string user_agreement_url = 1;
  string privacy_policy_url = 2;
}

message AppRegexConfigs {
  string regex_email  = 1;
  string regex_password = 2;
  string regex_phone = 3;
  string regex_nick  = 4;
  string regex_circle_name = 5;
}

message AppPayConfigs {
  repeated PayMethod circle_pays = 1;
}
message PayMethod {
  string method       = 1;
  string display_name = 2;
  double rang_min     = 3;
  double rang_max     = 4;
}

message AppOssConfigs {
  string oss_host    = 1;
  string oss_domain  = 2;
  string cdn_domain  = 3;
}

message AppLanguageConfigs {
  repeated LanguageMeta languages = 1;
  string lang_code = 2;
}
message LanguageMeta {
  string code        = 1;
  string name        = 2;
  string native_name = 3;
  bool   is_default  = 4;
}

message AppMuteConfigs {
  repeated MuteDuration durations = 1;
}
message MuteDuration {
  string label   = 1;
  int32  seconds = 2;
}

message AppGroupConfigs {
  string group_config_pay_notice = 1;
}

// 版本轮询请求
message AppConfigVersionReq {
  string env = 1;
  // known_versions 客户端已知的各 module 版本
  map<string, int64> known_versions = 2;
  // known_lang_packs 客户端已知的各语言包版本
  map<string, int64> known_lang_packs = 3;
}

// 版本轮询响应
message AppConfigVersionRsp {
  Result result = 1;
  // config_versions 各 module_key → 最新 version
  map<string, int64> config_versions = 2;
  // lang_pack_versions 各 lang_code → 最新 pack version
  map<string, int64> lang_pack_versions = 3;
  // has_changes 是否有任何变更
  bool has_changes = 4;
}
```

### 4.3 i18n Protobuf 定义（继承 v2 §3.4）

```protobuf
// proto/base/i18n.proto
syntax = "proto3";
package com.mineplanet.pojo;
option go_package = "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base";

import "base/result.proto";

// 请求：获取支持的语言列表
message AppFetchLanguageReq {
  string client_version = 1;
}

// 响应：语言元数据列表
message AppFetchLanguageRsp {
  Result result = 1;
  repeated LanguageMeta languages = 2;
}

// 请求：获取全量语言包
message AppFetchLangPackReq {
  string lang_code = 1;     // 语言代码，如 zh-CN
  string client_version = 2; // 客户端版本号
}

// 响应：全量语言包（含模板元信息）
message AppFetchLangPackRsp {
  Result result = 1;
  int64  pack_version = 2;   // 语言包版本号
  repeated LangStringEntry strings = 3;
}

// 请求：获取增量语言包
message AppFetchLangDifferenceReq {
  string lang_code = 1;
  int64  since_version = 2;  // 客户端当前版本号
  string client_version = 3;
}

// 响应：增量语言包
message AppFetchLangDifferenceRsp {
  Result result = 1;
  int64  current_version = 2;  // 服务端最新版本号
  repeated LangStringEntry additions = 3;  // 新增/修改的字符串
  repeated string deletions = 4;          // 删除的 key 列表
}

// 语言字符串条目（含模板类型和参数描述）
message LangStringEntry {
  string key = 1;
  string value = 2;                       // 旧客户端只读这个，按 plain 处理
  string template_type = 3;               // plain/named/icu
  repeated LangParam params = 4;          // 参数描述
  string operation_type = 5;              // ADD/MOD/DEL（增量协议用）
}

// 参数描述
message LangParam {
  string name     = 1;
  string type     = 2;  // string/int/float/date
  bool   required = 3;
  string default_v = 4;
}
```

---

## 5. 数据库设计

### 5.1 sys_config_schema（新建）

```sql
-- 配置字段元数据注册表
-- 用途：定义某个 module_key 下有哪些 field、类型、默认值、校验规则
-- 运营通过此表新增字段，无需改 Go 代码
CREATE TABLE sys_config_schema (
  id            BIGINT PRIMARY KEY AUTO_INCREMENT,
  module_key    VARCHAR(100) NOT NULL COMMENT '模块键 (base_cfg/member_cfg/...)',
  field_key     VARCHAR(100) NOT NULL COMMENT '字段键',
  field_type    VARCHAR(32)  NOT NULL COMMENT '类型: string/int/bool/float/enum/json/list',
  default_value TEXT COMMENT '默认值',
  validator     VARCHAR(255) COMMENT '校验规则 (正则/范围/枚举 JSON)',
  is_required   TINYINT DEFAULT 0 COMMENT '是否必填',
  is_secret     TINYINT DEFAULT 0 COMMENT '是否敏感 (客户端不可见)',
  description   VARCHAR(500) COMMENT '字段说明',
  client_scope  VARCHAR(64) DEFAULT 'all' COMMENT '客户端范围: all/android/ios/web/admin',
  min_app_ver   VARCHAR(32) COMMENT '最低支持的客户端版本',
  sort_order    INT DEFAULT 0 COMMENT '排序序号',
  is_enabled    TINYINT DEFAULT 1 COMMENT '是否启用',
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_module_field (module_key, field_key),
  KEY idx_module_key (module_key),
  KEY idx_client_scope (client_scope)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置字段元数据注册表';
```

### 5.2 sys_config_version（新建，替代旧项目同名表的职责）

```sql
-- 应用配置版本主表
-- 存储 AppConfigs(6001)协议返回的所有配置模块的 JSON 值
CREATE TABLE sys_config_version (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  module_key  VARCHAR(100) NOT NULL COMMENT '模块键',
  env         VARCHAR(50)  NOT NULL DEFAULT 'dev' COMMENT '环境 (dev/test/prod)',
  version     BIGINT       NOT NULL DEFAULT 1 COMMENT '版本号 (单调递增)',
  config_json JSON         NOT NULL COMMENT '模块配置 JSON',
  is_published TINYINT     NOT NULL DEFAULT 0 COMMENT '是否发布: 0-未发布 1-已发布',
  published_at DATETIME DEFAULT NULL COMMENT '发布时间',
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  create_by    VARCHAR(100) DEFAULT NULL,
  update_by    VARCHAR(100) DEFAULT NULL,
  UNIQUE KEY uk_module_env_version (module_key, env, version),
  KEY idx_module_env_latest (module_key, env, is_published, version DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用配置版本主表';
```

### 5.3 sys_lang_pack（新建）

```sql
-- 语言包主表
-- 维护每种语言的版本信息和元数据
CREATE TABLE sys_lang_pack (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  pack_name   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '语言包名称 (webp)',
  env         VARCHAR(32)  NOT NULL DEFAULT 'dev' COMMENT '环境',
  version     INT          NOT NULL DEFAULT 1 COMMENT '版本号 (单调递增)',
  lang_code   VARCHAR(32)  NOT NULL COMMENT '语言代码 (zh-CN/en-US)',
  description VARCHAR(255) DEFAULT NULL COMMENT '版本描述',
  is_published TINYINT     NOT NULL DEFAULT 0 COMMENT '是否已发布',
  published_at DATETIME DEFAULT NULL,
  published_by BIGINT DEFAULT NULL,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_pack_env_lang (pack_name, env, lang_code),
  KEY idx_published (is_published)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语言包主表';
```

### 5.4 sys_lang_string（新建，含模板扩展字段）

```sql
-- 语言字符串明细表
-- 存储每个语言包下的所有 key-value 对，支持增量版本管理和参数化模板
CREATE TABLE sys_lang_string (
  id             BIGINT PRIMARY KEY AUTO_INCREMENT,
  pack_id        BIGINT NOT NULL COMMENT '关联语言包 ID (→sys_lang_pack.id)',
  string_key     VARCHAR(255) NOT NULL COMMENT '语言 key',
  string_value   TEXT NOT NULL COMMENT '语言值 (可含 {placeholder})',
  group_name     VARCHAR(64) NOT NULL DEFAULT 'common' COMMENT '分组 (common/app/error/group/topic/user)',
  version        INT NOT NULL DEFAULT 1 COMMENT '版本号',
  operation_type VARCHAR(16) NOT NULL DEFAULT 'ADD' COMMENT '操作类型: ADD/MOD/DEL',
  prev_value     TEXT COMMENT '修改前的值',

  -- 模板扩展字段（v2 §3.2 新增能力）
  template_type  VARCHAR(16) DEFAULT 'plain' COMMENT '模板类型: plain/named/icu',
  params_schema JSON DEFAULT NULL COMMENT '参数描述: [{name,type,required,default}]',
  preview_sample JSON DEFAULT NULL COMMENT '预览示例值',

  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY uk_pack_key (pack_id, string_key),
  KEY idx_pack_id (pack_id),
  KEY idx_operation (pack_id, operation_type),
  KEY idx_string_key (string_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='语言字符串明细表（含参数化模板支持）';
```

### 5.5 ER 关系图

```
sys_config_schema (1) ──┬── (N) sys_config_version
  │  module_key+field_key     module_key → 外键语义关联
  │  定义"有哪些字段"          存储"字段的实际值"
  │
  │
sys_lang_pack (1) ─────────── (N) sys_lang_string
  │  pack_name+env+lang_code  pack_id → 外键
  │  定义"有哪些语言包"        存储"具体翻译文本+模板"
```

---

## 6. 服务层设计

### 6.1 目录结构总览

```
go/services/config/
├── go.mod
├── domain/
│   ├── module_key.go          # ModuleKey 枚举/常量
│   ├── version.go             # ConfigVersion 实体
│   ├── schema.go              # FieldSchema / ModuleSchema 实体
│   └── value.go               # TypedValue 类型安全包装
├── repository/
│   ├── interface.go           # Repository 接口定义
│   ├── mysql_repo.go          # MySQL 实现
│   ├── sqlite_repo.go         # SQLite 实现（单测用）
│   └── schema_repo.go         # sys_config_schema CRUD
├── cache/
│   ├── interface.go           # Cache 接口
│   └── redis_cache.go         # Redis 缓存实现
├── service/
│   ├── interface.go           # Service 接口
│   ├── fetch.go               # GetAppConfigs 主流程
│   ├── parse.go               # 按 schema 解析 config_json
│   ├── validate.go            # 按 schema.validator 校验
│   ├── compose.go             # 合并强类型字段 + dynamic_modules ★核心文件★
│   └── schema_service.go      # Schema 管理（运营接口）
└── sdk/                       # 阶段 B.5 新建
    ├── client.go              # Client 接口 + Default()
    ├── get.go                 # GetString/GetInt/GetBool...
    ├── module.go              # GetModule/Bind
    ├── watch.go               # Watch 变更订阅
    ├── cache_lru.go           # L1 进程内 LRU
    ├── cache_redis.go         # L2 Redis
    ├── remote.go              # L3 远程兜底
    └── pubsub.go             # Redis pub/sub 订阅

go/services/i18n/
├── go.mod
├── domain/
│   ├── lang_pack.go           # LangPack 实体
│   ├── lang_string.go         # LangString 实体（含 TemplateType/ParamsSchema）
│   ├── operation.go           # OperationType 枚举
│   ├── template.go            # TemplateType 枚举与校验规则
│   └── version.go            # PackVersion 实体
├── repository/
│   ├── interface.go
│   ├── mysql_repo.go
│   ├── sqlite_repo.go
├── cache/
│   ├── interface.go
│   └── redis_cache.go
├── service/
│   ├── interface.go
│   ├── pack.go                # GetLangPack 全量
│   ├── diff.go                # GetLangDifference 增量
│   ├── language.go            # GetLanguages 元数据
│   ├── publish.go             # 发布管理
│   ├── template_validator.go  # 模板校验（占位符一致性等）★质量门★
│   └── compat_filter.go       # 按 client_version 过滤 named/icu
└── sdk/                       # 阶段 B.5 新建
    ├── client.go              # Client 接口 + Default()
    ├── translate.go           # T() 渲染 + Raw() 原始模板
    ├── batch.go               # BatchT 批量
    ├── watch.go               # Watch 订阅
    ├── cache_lru.go
    ├── cache_redis.go
    ├── remote.go
    └── pubsub.go
```

### 6.2 核心文件职责约束

| 文件 | 职责 | 行数上限 | 禁止事项 |
|---|---|---|---|
| `compose.py` | 合并强类型字段 + dynamic_modules | 150 行 | 禁止包含 HTTP/Proto 序列化逻辑 |
| `template_validator.go` | 占位符与 params_schema 一致性校验 | 120 行 | 禁止做渲染，只做校验 |
| `compat_filter.go` | 按 client_version 过滤模板类型 | 80 行 | 禁止修改原始数据 |
| `schema_service.go` | Schema CRUD + 默认值同步 | 150 行 | 禁止直接操作 Redis |

### 6.3 Tars Servant 接入

新建两个 Tars Servant 作为 Gateway 到 Service 的桥梁：

```
go/tars/config/
├── cmd/main.go               # Config Tars 服务入口
├── adapter/
│   └── config_adapter.go     # Tars bytes → service 调用
├── internal/
│   └── handler/
│       └── config_handler.go # 6001/6002/6009/6010 处理
├── go.mod
└── go.sum

go/tars/i18n/
├── cmd/main.go
├── adapter/
│   └── i18n_adapter.go
├── internal/
│   └── handler/
│       └── i18n_handler.go   # 6003-6008 处理
├── go.mod
└── go.sum
```

---

## 7. SDK 设计（阶段 B.5，继承 v3 §2）

### 7.1 configsdk 对外 API

```go
package configsdk

type Client interface {
    GetString(ctx context.Context, moduleKey, fieldKey string) (string, error)
    GetInt(ctx context.Context, moduleKey, fieldKey string) (int64, error)
    GetBool(ctx context.Context, moduleKey, fieldKey string) (bool, error)
    GetFloat(ctx context.Context, moduleKey, fieldKey string) (float64, error)
    GetJSON(ctx context.Context, moduleKey, fieldKey string, out any) error
    GetModule(ctx context.Context, moduleKey string) (Module, error)
    Bind(ctx context.Context, moduleKey string, out any) error
    Watch(moduleKey string, handler func(Module)) (cancel func())
    Ping(ctx context.Context) error
}
```

### 7.2 i18nsdk 对外 API

```go
package i18nsdk

type Client interface {
    T(ctx context.Context, langCode, key string, params map[string]any) (string, error)
    Raw(ctx context.Context, langCode, key string) (Template, error)
    BatchT(ctx context.Context, langCode string, keys []string, params map[string]any) (map[string]string, error)
    Watch(langCode string, handler func(packVersion int64)) (cancel func())
    Ping(ctx context.Context) error
}
```

### 7.3 双模式初始化

```go
// 单体模式（MVP 初期）：同进程函数调用
configsdk.Init(configsdk.Options{
    Mode:    configsdk.ModeInProcess,
    Service: configSvc,
    Redis:   rdb,
})

// 微服务模式（后续）：Tars 远程调用
configsdk.Init(configsdk.Options{
    Mode:        configsdk.ModeRemote,
    TarsServant: "cairobot.config.AppConfigsObj",
    Redis:       rdb,
})
```

---

## 8. Admin 侧设计（阶段 C）

### 8.1 admin-server API 清单

```
配置 Schema 管理：
  POST   /api/config/schema         新增字段定义
  PUT    /api/config/schema/:id     修改字段定义
  DELETE /api/config/schema/:id     软删除字段
  GET    /api/config/schema         列表查询

配置值管理：
  GET    /api/config/value/:env     按 env 获取所有 module_key 值
  PUT    /api/config/value          更新某 module_key 的值
  POST   /api/config/value/publish  发布配置版本

多语言管理：
  POST   /api/i18n/pack             创建/更新语言包
  GET    /api/i18n/pack/:lang_code 获取语言包信息
  POST   /api/i18n/string           新增多语言 key（含模板类型）
  PUT    /api/i18n/string/:id       修改多语言 key
  DELETE /api/i18n/string/:id       删除多语言 key
  GET    /api/i18n/diff             获取增量变更
  POST   /api/i18n/publish          发布语言包版本
  POST   /api/i18n/preview          预览模板渲染效果
```

### 8.2 admin-web 最小页面（MVP 先做表格态）

| 页面 | 路由 | 功能 |
|---|---|---|
| 配置 Schema 管理 | `/admin/config/schema` | 字段元数据 CRUD 表格 |
| 配置值管理 | `/admin/config/value` | 按 module 编辑配置值 |
| 多语言 Key 管理 | `/admin/i18n/strings` | 多语言 key CRUD（含模板类型选择） |
| 语言包发布 | `/admin/i18n/packs` | 发布管理与版本查看 |

---

## 9. 缓存策略

### 9.1 缓存 Key 命名空间

| 缓存 Key | TTL | 说明 |
|---|---|---|
| `config:schema:{module_key}` | 5 分钟 | 单个模块的 schema |
| `config:schema:all` | 5 分钟 | 全量 schema |
| `config:value:{env}:{module_key}:{version}` | 10 分钟 | 单个模块配置值 |
| `config:compose:{env}:{client_scope}:{ver_hash}` | 10 分钟 | 合成结果（6001 响应） |
| `i18n:pack:{env}:{lang_code}:{version}` | 1 小时 | 全量语言包 |
| `i18n:diff:{env}:{lang_code}:{since}` | 10 分钟 | 增量 diff |
| `i18n:meta:languages` | 30 分钟 | 语言元数据列表（6003 响应） |

### 9.2 ver_hash 计算

```text
ver_hash = fnv64(
  所有请求 module_key 的当前 version +
  所有涉及 module_key 的 schema version +
  client_scope +
  env
)
```

schema 或 config 任一变更，ver_hash 必然变化，缓存自然失效。

### 9.3 Redis pub/sub 失效通道（阶段 B.5）

| Channel | Payload | 发布者 | 订阅者 |
|---|---|---|---|
| `cairobot.config.invalidate` | `{"module","version","env"}` | admin-server | configsdk |
| `cairobot.i18n.invalidate` | `{"lang_code","pack_version","env"}` | admin-server | i18nsdk |

---

## 10. 兼容性策略

### 10.1 客户端版本兼容

```text
旧客户端（无 dynamic_modules / template_type 感知能力）：
  → 6001 响应：dynamic_modules 为空数组，只用强类型字段
  → 6007 响应：过滤掉 template_type != "plain" 的 key
  → 行为等同于"没有新功能"，不会崩溃

新客户端（识别 dynamic_modules + template_type）：
  → 6001 响应：读取 dynamic_modules 获取新字段
  → 6007 响应：按 template_type 选择渲染分支
  → 完整功能可用
```

### 10.2 过滤依据

通过 `MessagePacket.extend` 中的 `client_version` 字段（已在 [message.proto](proto/base/message.proto) 中标准化）或 `AppConfigsReq.client_version` 做过滤。

---

## 11. 阶段化实施计划

### 阶段 A：协议与数据基础（预估 2 天）

**目标**：建立协议身份和数据存储基础，不影响任何现有功能。

| 任务编号 | 任务 | 产出物 | 验证标准 |
|---|---|---|---|
| A-1 | 创建 `proto/base/app_config.proto` | Proto 文件 | protoc 编译通过 |
| A-2 | 创建 `proto/base/i18n.proto` | Proto 文件 | protoc 编译通过 |
| A-3 | 登记协议编号 6001-6010 到注册表 | 更新 `协议编号注册表.md` | max+min 唯一性检查通过 |
| A-4 | 在 routes.yaml 增加 6001/6009 路由 | 更新 `routes.yaml` | 路由加载校验通过 |
| A-5 | 在 routes.yaml 增加 6003-6008 路由 | 更新 `routes.yaml` | 路由加载校验通过 |
| A-6 | 创建 DDL 迁移脚本（4 张表） | `scripts/migration/001_config_i18n.sql` | SQLite 可执行 |
| A-7 | 更新 go.work 加入新 module | 更新 `go/work` | `go build ./...` 通过 |

**风险**：无。纯新增，不动已有代码。

### 阶段 B：services 层核心实现（预估 3 天）

**目标**：实现 config 和 i18n 的完整领域逻辑，SQLite 单测全绿。

| 任务编号 | 任务 | 产出物 | 验证标准 |
|---|---|---|---|
| B-1 | 创建 `go/services/config/domain/` 全部实体 | 4 个 Go 文件 | 单测覆盖 |
| B-2 | 创建 `go/services/config/repository/` 接口 + SQLite 实现 | 4 个 Go 文件 | 单测覆盖 |
| B-3 | 创建 `go/services/config/cache/` 接口 + mock | 2 个 Go 文件 | 单测覆盖 |
| B-4 | 创建 `go/services/config/service/fetch.go` | 1 个 Go 文件 | 单测覆盖 |
| B-5 | 创建 `go/services/config/service/parse.go` | 1 个 Go 文件 | 单测覆盖 |
| B-6 | 创建 `go/services/config/service/compose.go` | 1 个 Go 文件 | 单测覆盖 |
| B-7 | 创建 `go/services/config/service/schema_service.go` | 1 个 Go 文件 | 单测覆盖 |
| B-8 | 创建 `go/services/config/service/validate.go` | 1 个 Go 文件 | 单测覆盖 |
| B-9 | 创建 `go/services/i18n/domain/` 全部实体 | 5 个 Go 文件 | 单测覆盖 |
| B-10 | 创建 `go/services/i18n/repository/` 接口 + SQLite 实现 | 3 个 Go 文件 | 单测覆盖 |
| B-11 | 创建 `go/services/i18n/cache/` 接口 + mock | 2 个 Go 文件 | 单测覆盖 |
| B-12 | 创建 `go/services/i18n/service/pack.go` | 1 个 Go 文件 | 单测覆盖 |
| B-13 | 创建 `go/services/i18n/service/diff.go` | 1 个 Go 文件 | 单测覆盖 |
| B-14 | 创建 `go/services/i18n/service/language.go` | 1 个 Go 文件 | 单测覆盖 |
| B-15 | 创建 `go/services/i18n/service/template_validator.go` | 1 个 Go 文件 | 单测覆盖 |
| B-16 | 创建 `go/services/i18n/service/compat_filter.go` | 1 个 Go 文件 | 单测覆盖 |
| B-17 | 全量单元测试 + 覆盖率报告 | 测试输出 | 覆盖率 ≥ 80% |

**准入条件**：阶段 A 全部完成。
**完成标准**：`go test ./go/services/config/... ./go/services/i18n/... -v` 全绿。

### 阶段 B.5：SDK 与热更新机制（预估 2 天）

**目标**：实现 configsdk + i18nsdk，业务服务可通过 SDK 引用配置和多语言。

| 任务编号 | 任务 | 产出物 | 验证标准 |
|---|---|---|---|
| B5-1 | 创建 `go/services/config/sdk/client.go` 接口 | 1 个 Go 文件 | 接口编译通过 |
| B5-2 | 创建 configsdk Get*/GetModule/Bind 实现 | 3 个 Go 文件 | 单测覆盖 |
| B5-3 | 创建 configsdk Watch + LRU 缓存 | 2 个 Go 文件 | 单测覆盖 |
| B5-4 | 创建 configsdk InProcess/Remote 双模式 | 2 个 Go 文件 | 单测覆盖 |
| B5-5 | 创建 `go/services/i18n/sdk/client.go` 接口 | 1 个 Go 文件 | 接口编译通过 |
| B5-6 | 创建 i18nsdk T/Raw/BatchT 实现 | 3 个 Go 文件 | 单测覆盖（named 模板渲染） |
| B5-7 | 创建 i18nsdk Watch + LRU 缓存 | 2 个 Go 文件 | 单测覆盖 |
| B5-8 | 创建 i18nsdk InProcess/Remote 双模式 | 2 个 Go 文件 | 单测覆盖 |
| B5-9 | admin-server Redis pub/sub 发布逻辑 | 1 个 Go 文件 | miniredis 集成测试 |
| B5-10 | SDK Redis pub/sub 订阅逻辑 | 2 个 Go 文件 | miniredis 集成测试 |
| B5-11 | SDK 全量测试 | 测试输出 | 覆盖率 ≥ 80% |

**准入条件**：阶段 B 全部完成。
**完成标准**：SDK 可在同进程模式下正确读取配置和多语言。

### 阶段 C：Admin 最小可用（预估 3 天）

**目标**：admin-server 提供 CRUD API，admin-web 提供最小管理页面。

| 任务编号 | 任务 | 产出物 | 验证标准 |
|---|---|---|---|
| C-1 | 创建 `go/tars/provider-admin/` 基础 Gin 服务 | 3 个 Go 文件 | 服务可启动 |
| C-2 | 实现 `/api/config/schema` CRUD | 2 个 Go 文件 | API 测试通过 |
| C-3 | 实现 `/api/config/value` 读写 + 发布 | 2 个 Go 文件 | API 测试通过 |
| C-4 | 实现 `/api/i18n/*` 全部 API | 4 个 Go 文件 | API 测试通过 |
| C-5 | 接入 template_validator 质量门 | 改造 C-4 | 含非法占位符的保存被拒绝 |
| C-6 | 接入 Redis 缓存失效 + pub/sub | 2 个 Go 文件 | 写入后缓存即时失效 |
| C-7 | admin-web 配置 Schema 管理页 | Vue 组件 | 可 CRUD |
| C-8 | admin-web 配置值管理页 | Vue 组件 | 可编辑/发布 |
| C-9 | admin-web 多语言 Key 管理页 | Vue 组件 | 可编辑含模板类型的 key |
| C-10 | admin-web 预览功能 | Vue 组件 | preview_sample 可渲染 |

**准入条件**：阶段 B 完成（B.5 可并行）。

### 阶段 D：Gateway 端到端集成（预估 2 天）

**目标**：全链路打通，单体模式端到端测试通过。

| 任务编号 | 任务 | 产出物 | 验证标准 |
|---|---|---|---|
| D-1 | 创建 `go/tars/config/` Tars Servant | 4 个 Go 文件 | 注册到 TarsRegistry |
| D-2 | 创建 `go/tars/i18n/` Tars Servant | 4 个 Go 文件 | 注册到 TarsRegistry |
| D-3 | 接入 repository MySQL 实现 | 2 个 Go 文件 | 可读写真实 MySQL |
| D-4 | 接入 cache Redis 实现 | 2 个 Go 文件 | 缓存命中/穿透正常 |
| D-5 | 6001 全链路测试：客户端→Gateway→Tars→Service→DB | e2e 测试 | 返回正确的 dynamic_modules |
| D-6 | 6007 全链路测试：增量拉取 | e2e 测试 | 增量数据正确 |
| D-7 | 6009 全链路测试：版本轮询触发增量 | e2e 测试 | 变更检测 ≤1s |
| D-8 | "新增字段不改代码"端到端验证 | 场景测试 | admin 加字段 → 6001 自动返回 |
| D-9 | "新增 named key"端到端验证 | 场景测试 | admin 加 key → 6007 自动返回 |
| D-10 | 兼容性测试：模拟老客户端请求 | 场景测试 | 老客户端不崩溃 |

**准入条件**：阶段 B + C 完成。

### 阶段 E：文档与归档（预估 半天）

| 任务编号 | 任务 | 产出物 |
|---|---|---|
| E-1 | CODE-WIKI 更新（§3/§4/§5/§9/§17 共 5 处） | 更新 `CODE-WIKI.md` |
| E-2 | ADR-009：Config Schema Registry 与 i18n 模板架构决策 | 新建 `docs/adr/ADR-009-config-i18n-schema-template.md` |
| E-3 | ADR-010：admin 边界与 SDK 引用规范 | 新建 `docs/adr/ADR-010-admin-boundary-sdk.md` |
| E-4 | config service 模块文档 | 更新建 `docs/wiki/modules/config-service.md` |
| E-5 | i18n service 模块文档 | 新建 `docs/wiki/modules/i18n-service.md` |
| E-6 | SDK 使用规范文档 | 新建 `docs/wiki/modules/config-i18n-sdk-guide.md` |
| E-7 | 测试报告 | 新建 `docs/reports/testing/config-i18n-test-report.md` |
| E-8 | 测试用例注册表登记 | 更新测试用例注册表 |

---

## 11. 风险评估

### 11.1 技术风险

| 风险 ID | 描述 | 概率 | 影响 | 应对措施 |
|---|---|---|---|---|
| R-01 | Protobuf map<string,string> 在某些客户端 SDK 上序列化行为不一致 | 低 | 中 | dynamic_modules 的 fields 使用 map 但提供 fallback JSON string |
| R-02 | ICU MessageFormat 在 Go 标准库无原生支持 | 中 | 低 | MVP 阶段先只实现 named 模板（简单 `{key}` 替换），icu 留到 P1 |
| R-03 | Redis pub/sub 消息丢失导致 SDK 缓存不一致 | 低 | 中 | TTL 兜底（L1 30s + L2 10min），pub/sub 只加速失效 |
| R-04 | SQLite 单测与 MySQL 行为差异（JSON 函数） | 中 | 低 | 抽象 Repository 接口，单测用 SQLite，集成测用 MySQL |
| R-05 | go.work 多 module 协调开发复杂度 | 中 | 低 | 每个 stage 独立 commit，避免大 PR |

### 11.2 范围风险

| 风险 ID | 描述 | 概率 | 影响 | 应对措施 |
|---|---|---|---|---|
| R-06 | 阶段 C admin-web 开发工作量超预期 | 中 | 中 | MVP 先做表格态 CRUD，复杂联动留 S2 |
| R-07 | 阶段 B.5 SDK 三层缓存调试复杂度高 | 中 | 中 | 先实现 InProcess 单层直通，缓存分层逐步叠加 |
| R-08 | 6000 段协议编号与未来其他 App 协议冲突 | 低 | 低 | 本方案已预留 6001-6010，后续 App 协议从 6011 开始 |

### 11.3 依赖风险

| 风险 ID | 描述 | 概率 | 影响 | 应对措施 |
|---|---|---|---|---|
| R-09 | MySQL / Redis 基础设施在 MVP 环境不可用 | 中 | 高 | 阶段 B 全部基于 SQLite + mock cache 完成，阶段 D 才引入真实依赖 |
| R-10 | TarsGo 框架与 Protobuf 生成的代码集成问题 | 中 | 中 | 已有 health/hello 作为集成范例可参考 |

---

## 12. 禁止事项（继承自方案铁律）

1. **不允许**破坏已有的 2100:2097-2102 协议 wire 兼容
2. **不允许**服务端做多语言模板渲染（渲染由客户端完成，SDK.T() 除外）
3. **不允许**跳过 schema 校验直接改 JSON 字段
4. **不允许**单文件超过 200 行、单函数超过 50 行
5. **不允许**新增字段缺少 admin-web 可视化管理入口
6. **不允许**业务服务直接 import services/config 或 services/i18n（必须走 SDK）
7. **不允许**业务服务调用 admin-server 任何 API
8. **不允许** git push（仅本地 commit）

---

## 13. 成功标准

### 13.1 阶段 A 完成标准

- [ ] `proto/base/app_config.proto` 和 `proto/base/i18n.proto` 存在且 protoc 编译通过
- [ ] 协议编号注册表中 6001-6010 全部登记，max+min 唯一性检查通过
- [ ] `routes.yaml` 包含 6001/6003/6005/6007/6009 路由，网关启动加载正常
- [ ] DDL 脚本可在 SQLite 和 MySQL 上执行成功

### 13.2 阶段 B 完成标准

- [ ] `go/services/config` 和 `go/services/i18n` 全部单元测试通过
- [ ] 测试覆盖率 ≥ 80%
- [ ] `compose.go` 正确合并强类型字段和 dynamic_modules
- [ ] `template_validator.go` 正确拦截不一致的占位符
- [ ] `compat_filter.go` 正确按 client_version 过滤

### 13.3 阶段 B.5 完成标准

- [ ] configsdk.InProcess 模式下 GetString/GetInt/GetModule 正确返回
- [ ] i18nsdk.InProcess 模式下 T() 正确渲染 named 模板
- [ ] Watch 回调在值变更时正确触发
- [ ] miniredis 集成测试验证 pub/sub 失效链路

### 13.4 阶段 D 完成标准（最终验收）

- [ ] 6001 端到端：客户端请求 → 返回含 dynamic_modules 的完整配置
- [ ] 6007 端到端：客户端请求 → 返回正确的增量 diff
- [ ] 6009 端到端：版本轮询 → 正确标识 has_changes
- [ ] **"运营加字段 → 零代码改动 → 客户端拿到"** 全链路验证通过
- [ ] **"运营加多语言 key → 零代码改动 → 客户端拿到"** 全链路验证通过
- [ ] 老客户端模拟请求不崩溃

---

## 14. 与三份输入文档的映射关系

| 本方案章节 | 来源文档 | 适配说明 |
|---|---|---|
| §3 总体架构 | v3 §1 | 完全继承三条数据流，标注了 MVP 新建 vs 已有资产 |
| §4 协议设计 | v2 §2.3 + §3.4 | 从零设计 6001-6010，非扩展现有协议 |
| §5 数据库设计 | v2 §2.2 + §3.2 + PRD §2 | 合并 v2 表设计与 PRD 现状诊断，从零建表 |
| §6 服务层设计 | v2 §2.4 + §3.6 | 目录结构完全继承，标注"从零创建" |
| §7 SDK 设计 | v3 §2 | 完全继承接口定义 |
| §8 Admin 设计 | v2 §2.5 + v3 §3 | 合并两版 admin 规范 |
| §9 缓存策略 | v2 §2.6 + v3 §2.4 | 完全继承 |
| §10 兼容性 | v2 §3.7 | 完全继承 |
| §11-14 | 新增 | 针对 MVP 从零构建的风险、标准、禁止项 |

---

## 附录 A：关键差异清单（方案 vs 现实）

| 维度 | 方案文档假设 | CaiRobot MVP 实际 | 本方案应对 |
|---|---|---|---|
| 起始状态 | 已有成熟运行的 config/i18n 服务 | 仅 S0 骨架（gateway + hello） | **从零构建**，复用架构思想 |
| 协议基础 | 6001-6007 已在线运行 | 2100 段仅 2 条路由 | 6000 段全新开辟，无兼容负担 |
| 数据库 | 4 张表已有数据（135/2/1/393 条） | 无数据库连接 | DDL 从零建，种子数据后续导入 |
| 代码规模 | config_query.go 563行需拆分 | 无此文件 | 按正确分层直接创建 <200 行文件 |
| admin-server | 已有 Gin 服务 + 1267行超限文件 | 仅 README 占位 | 按正确分层从零创建 |
| 迁移风险 | 旧数据兼容、旧客户端兼容 | 无历史包袱 | **无迁移风险**，但需注意未来兼容设计 |

---

> **本文档待评审通过后，按阶段 A → B → B.5 → C → D → E 顺序执行。**
> **每阶段完成后输出阶段汇报，评审通过后进入下一阶段。**
