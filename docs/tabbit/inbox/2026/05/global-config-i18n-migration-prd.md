# 全局应用配置和多语言模块 - 现状诊断与迁移准备文档

**文档编号**: DIAG-CONFIG-I18N-001\
**版本**: v1.0\
**日期**: 2026-05-22\
**用途**: 新项目经理接手参考 / 模块升级评估依据

***

## 1. 模块定位与范围

本模块负责 MineplanetGo 客户端启动时的两项核心能力：

| 能力            | 说明                     | 客户端触发时机                       |
| ------------- | ---------------------- | ----------------------------- |
| **全局应用配置**    | 域名、正则、支付方式、OSS地址、会员权限等 | App 启动后首次调用 6001 协议           |
| **多语言(i18n)** | 界面文案的多语言支持，含全量包/增量更新   | App 启动后按需调用 6003/6005/6007 协议 |

***

## 2. 当前数据架构（真实查询结果）

### 2.0 数据来源声明

| 项目       | 说明                                     |
| -------- | -------------------------------------- |
| **数据来源** | ✅ **真实查询**                             |
| 查询时间     | 2026-05-22                             |
| 数据库环境    | production (mineplanet\_community\_db) |
| 数据准确性    | 已核实                                    |

### 2.1 四张核心表总览

| 表名                      | 记录数       | 用途                        | 存储引擎   |
| ----------------------- | --------- | ------------------------- | ------ |
| `sys_config_version`    | **135 条** | 应用配置版本主表（含语言包冗余）          | InnoDB |
| `sys_lang_pack`         | **2 条**   | 语言包主表（zh-CN, en）          | InnoDB |
| `sys_lang_pack_release` | **1 条**   | 语言包离线发布记录                 | InnoDB |
| `sys_lang_string`       | **393 条** | 语言字符串明细（\~196 keys × 2语言） | InnoDB |

> 另有 `sys_config` 表（旧版 KV 配置），本次不展开。

### 2.2 sys\_config\_version — 建表语句 & 当前数据

```sql
-- ============================================
-- 表: sys_config_version（应用配置版本主表）
-- 用途: 存储 AppConfigs(6001)协议返回的所有配置模块
-- ============================================
CREATE TABLE `sys_config_version` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `module_key` varchar(100) DEFAULT NULL COMMENT '模块键 (base_cfg/member_cfg/lang_cfg/...)',
  `env` varchar(50) DEFAULT NULL COMMENT '环境 (dev/test/prod)',
  `version` bigint(20) DEFAULT NULL COMMENT '版本号',
  `config_json` json DEFAULT NULL COMMENT '模块配置JSON(与proto结构一致)',
  `is_published` bigint(20) DEFAULT '1' COMMENT '是否发布',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `published_at` datetime DEFAULT NULL COMMENT '发布时间',
  `updated_at` datetime DEFAULT NULL COMMENT '最后更新时间',
  `create_by` varchar(100) DEFAULT NULL COMMENT '创建人',
  `update_by` varchar(100) DEFAULT NULL COMMENT '更新人',
  PRIMARY KEY (`id`),
  KEY `idx_module_env_version` (`module_key`,`env`,`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**当前已发布的 module\_key 分布（共 11 种模块）**:

| module\_key          | dev 最新版       | test 最新版      | prod 最新版      | 配置内容概要                                                          |
| -------------------- | ------------- | ------------- | ------------- | --------------------------------------------------------------- |
| `base_cfg`           | v1            | v1            | v1            | domain\_root, domain\_wap, sign\_rand, construct\_email         |
| `member_cfg`         | v1766815228⚠️ | v1766815221⚠️ | v1766815221⚠️ | max\_free\_group\_*, max\_paid\_group\_* (6个字段)                 |
| `question_quota_cfg` | v1            | v1            | v1            | question\_quota\_normal/premium/premium\_plus                   |
| `topic_limit_cfg`    | v1            | v1            | v1            | topic\_normal/question/article max/min\_length, summary\_length |
| `wap_cfg`            | v1            | v1            | v1            | user\_agreement\_url, privacy\_policy\_url                      |
| `regex_cfg`          | v1            | v1            | v1            | regex\_email/password/phone/nick/circle\_name                   |
| `pay_cfg`            | v22           | v22           | v22           | circle\_pays\[] (method/display\_name/rang.min.max)             |
| `oss_cfg`            | v1            | v1            | v1            | oss\_host, oss\_domain, cdn\_domain                             |
| `lang_cfg`           | v3            | v1            | v1            | languages\[], lang\_code (语言元数据列表)                              |
| `lang_pack_default`  | v5            | ❌无            | ❌无            | ⚠️ 冗余: 全量语言字符串map                                               |
| `lang_pack_weba`     | v8            | ❌无            | ❌无            | ⚠️ 冗余: 全量语言字符串map                                               |

> ⚠️ **关键发现**: `lang_pack_default` 和 `lang_pack_weba` 是历史遗留的冗余数据，功能已被 `sys_lang_pack` + `sys_lang_string` 二级表替代。

### 2.3 sys\_lang\_pack — 建表语句 & 当前数据

```sql
-- ============================================
-- 表: sys_lang_pack（语言包主表）
-- 用途: 维护每种语言的版本信息和元数据
-- ============================================
CREATE TABLE `sys_lang_pack` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `pack_name` varchar(64) NOT NULL DEFAULT '' COMMENT '语言包名称 (webp)',
  `env` varchar(32) NOT NULL DEFAULT 'dev' COMMENT '环境 (dev/test/prod)',
  `version` int(11) NOT NULL COMMENT '版本号',
  `lang_code` varchar(32) NOT NULL COMMENT '语言代码 (zh-CN/en-US)',
  `description` varchar(255) DEFAULT NULL COMMENT '版本描述',
  `is_published` tinyint(4) DEFAULT '0' COMMENT '是否已发布: 0-未发布 1-已发布',
  `published_at` datetime DEFAULT NULL COMMENT '发布时间',
  `published_by` bigint(20) DEFAULT NULL COMMENT '发布人ID',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_pack_env_lang` (`pack_name`,`env`,`lang_code`),
  KEY `idx_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**当前数据**:

| id | pack\_name | env  | lang\_code | version | is\_published | description |
| -- | ---------- | ---- | ---------- | ------- | ------------- | ----------- |
| 3  | webp       | prod | zh-CN      | **3**   | 1             | 简体中文        |
| 4  | webp       | prod | en         | **3**   | 1             | English     |

### 2.4 sys\_lang\_string — 建表语句 & 当前数据

```sql
-- ============================================
-- 表: sys_lang_string（语言字符串明细表）
-- 用途: 存储每个语言包下的所有 key-value 对，支持增量版本管理
-- ============================================
CREATE TABLE `sys_lang_string` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `pack_id` bigint(20) NOT NULL COMMENT '关联的语言包ID (→sys_lang_pack.id)',
  `string_key` varchar(255) NOT NULL COMMENT '语言key (如: svc_app_me, svc_common_ok)',
  `string_value` text NOT NULL COMMENT '语言值',
  `group` varchar(64) NOT NULL DEFAULT 'common' COMMENT '分组 (common/app/error/group/topic/user)',
  `version` int(11) NOT NULL COMMENT '版本号',
  `operation_type` varchar(16) NOT NULL DEFAULT 'ADD' COMMENT '操作类型: ADD-新增 MOD-修改 DEL-删除',
  `prev_value` text COMMENT '修改前的值 (用于追踪变更)',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_pack_id` (`pack_id`),
  KEY `idx_pack_key` (`pack_id`,`string_key`),
  KEY `idx_operation` (`pack_id`,`operation_type`),
  KEY `idx_string_key` (`string_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**当前数据分布**:

| lang\_code | 有效字符串数  | 最大版本 | 操作类型分布(估算)                   |
| ---------- | ------- | ---- | ---------------------------- |
| zh-CN      | **196** | 3    | ADD \~150, MOD \~40, DEL \~6 |
| en         | **195** | 3    | ADD \~150, MOD \~40, DEL \~5 |

### 2.5 sys\_lang\_pack\_release — 建表语句 & 当前数据

```sql
-- ============================================
-- 表: sys_lang_pack_release（语言包发布记录）
-- 用途: 记录每次离线包发布的信息（OSS地址、MD5等）
-- ============================================
CREATE TABLE `sys_lang_pack_release` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `env` varchar(32) NOT NULL DEFAULT 'dev' COMMENT '环境',
  `lang_code` varchar(32) NOT NULL COMMENT '语言代码',
  `lang_pack` varchar(64) NOT NULL COMMENT '打包类型',
  `version` int(11) NOT NULL COMMENT '版本号',
  `oss_url` varchar(512) DEFAULT NULL COMMENT 'OSS下载链接',
  `file_size` bigint(20) DEFAULT NULL COMMENT '文件大小(字节)',
  `md5_hash` varchar(64) DEFAULT NULL COMMENT 'MD5校验值',
  `strings_count` int(11) DEFAULT '0' COMMENT '字符串数量',
  `status` tinyint(4) DEFAULT '1' COMMENT '状态：1-有效，0-已回滚',
  `publisher` varchar(64) DEFAULT NULL COMMENT '发布人',
  `published_at` bigint(20) DEFAULT NULL COMMENT '发布时间',
  `release_note` varchar(500) DEFAULT NULL COMMENT '发布说明',
  `created_by` bigint(20) DEFAULT NULL,
  `updated_by` bigint(20) DEFAULT NULL,
  `updated_at` bigint(20) DEFAULT NULL,
  `created_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_env_lang` (`env`,`lang_code`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**当前数据**: 仅 1 条发布记录（en 语言包 v1 版本）

***

## 3. 协议定义层（Protobuf）

**文件**: [app\_base.proto](/Users/mac/StudioProjects/2026/open-citycloud-workspace/protobuf/protocols/base/app_base.proto)

### 3.1 配置相关协议

```
客户端 ──[6001 AppConfigsReq]──▶ ThirdServer ──[6001 AppConfigsRsp]──▶ 客户端
                                       │
                                       ▼
                              readonly-query.GetAppConfigs()
                                       │
                          ┌────────────┼────────────┬────────────┐
                          ▼            ▼            ▼            ▼
                     base_cfg     member_cfg   topic_limit_cfg  ... (10个模块)
                     (AppBase)    (AppBase扩展) (AppBase扩展)
```

**AppConfigsRsp 结构** (app\_base.proto:145-160):

```protobuf
message AppConfigsRsp {
  Result result = 1;
  AppBaseConfigs base_cfg = 2;       // 33 个字段
  AppWapUrlConfigs wap_cfg = 3;      // 2 个字段
  AppRegexConfigs regex_cfg = 4;     // 5 个字段
  AppPayConfigs pay_cfg = 5;         // circle_pays[]
  AppOssConfigs oss_cfg = 6;         // 3 个字段
  AppLanguageConfigs lang_cfg = 7;   // languages[] + lang_code
  AppMuteConfigs mute_cfg = 8;       // durations[]
  AppGroupConfigs group_cfg = 9;     // group_config_pay_notice
}
```

### 3.2 多语言相关协议

```
客户端 ──[6003 AppFetchLanguageReq]──▶ 获取语言元数据 (name/native_name/lang_code/...)
客户端 ──[6005 AppFetchLangPackReq]───▶ 获取全量语言包 (map[string]string)
客户端 ──[6007 AppFetchLangDifferenceReq] ▶ 获取增量 (additions/deletions)
```

**三条协议的数据源路径**:

| 协议                     | MinType | 服务端函数                    | 数据源表                                | 状态                    |
| ---------------------- | ------- | ------------------------ | ----------------------------------- | --------------------- |
| AppFetchLanguage       | 6003    | `GetAppLanguage()`       | `sys_app_language`                  | ✅ 正确                  |
| AppFetchLangPack       | 6005    | `GetLangPackConfig()`    | `sys_config_version` ⚠️             | ❌ 应改用 `sys_lang_pack` |
| AppFetchLangDifference | 6007    | `GetAppLangDifference()` | `sys_lang_pack` + `sys_lang_string` | ✅ 正确                  |

***

## 4. 服务端调用链路

### 4.1 完整请求处理流程

```
┌─────────────┐     POST /api/hello      ┌──────────────┐
│   客户端     │ ─────────────────────────▶│    Gateway    │
│ (Flutter/React) │   MessagePacket(PB)     │  (路由转发)   │
└─────────────┘                           └──────┬───────┘
                                                 │ MaxType=6000
                                                 ▼
                                        ┌──────────────────┐
                                        │  ThirdServer      │
                                        │ (ThirdObjImp)     │
                                        │                  │
                                        │ Init() 时注入:    │
                                        │ ├ readonly-query  │
                                        │ │  (QueryService)  │
                                        │ ├ Redis Cache     │
                                        │ └ MySQL DB        │
                                        └────────┬─────────┘
                                                 │
                                    ┌────────────┼────────────────┐
                                    ▼            ▼                ▼
                              GetAppConfigs  GetLangPackConfig  GetAppLangDifference
                                    │            │                │
                                    ▼            ▼                ▼
                            sys_config_   sys_config_      sys_lang_pack +
                            version       version          sys_lang_string
                            (10个模块)    (lang_pack_*)    (二级表 ✅)
```

### 4.2 核心代码文件索引

| 文件                                                                                                                                                                    | 行数         | 职责                   | 状态        |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | -------------------- | --------- |
| [ThirdObj\_imp.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/phanethoodGO/modules/third-service/handlers/ThirdObj_imp.go)           | 372        | Tars 接口入口、依赖注入、服务初始化 | 正常        |
| [config\_query.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/phanethoodGO/modules/readonly-query/query/config_query.go)             | **563** ⚠️ | 所有配置/语言包查询逻辑         | **超限待拆分** |
| [config\_initializer.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/phanethoodGO/modules/readonly-query/query/config_initializer.go) | \~60       | 配置初始化与缓存预热           | 正常        |

***

## 5. 后台运维系统（admin-server + admin-web）

### 5.1 架构关系

```
┌─────────────────────────────────────────────────┐
│                 admin-web (Vue2)                 │
│  ┌─────────────┐ ┌──────────────┐ ┌───────────┐ │
│  │ sys-config  │ │ lang-management│ │lang-pack- │ │
│  │ (配置管理页) │ │ (语言包管理页) │ │server页   │ │
│  └──────┬──────┘ └──────┬───────┘ └─────┬─────┘ │
│         │               │              │         │
│  sys-config.js   lang-pack.js   lang-pack-      │
│  sys-app-lang.js  sys-lang-string.js server.js   │
└─────────┼───────────────┼──────────────┼─────────┘
          │               │              │
          ▼               ▼              ▼
┌─────────────────────────────────────────────────┐
│               admin-server (Go/Gin)             │
│                                                  │
│  ┌─────────────────────────────────────────┐    │
│  │           Controller Layer (apis/)       │    │
│  │  sys_config.go  sys_lang_pack.go        │    │
│  │  sys_lang_pack_release.go  sys_lang_string.go │
│  └────────────────────┬────────────────────┘    │
│                       │                          │
│  ┌────────────────────▼────────────────────┐    │
│  │           Service Layer (service/)       │    │
│  │  sys_config.go (183行) ✅               │    │
│  │  lang_pack_service.go (1267行) ⚠️       │    │
│  │  lang_pack_release_service.go           │    │
│  └────────────────────┬────────────────────┘    │
│                       │                          │
│  ┌────────────────────▼────────────────────┐    │
│  │           Model Layer (models/)         │    │
│  │  SysConfig  SysConfigVersion            │    │
│  │  SysLangPack  SysLangString             │    │
│  │  SysLangPackRelease                    │    │
│  └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
          │               │              │
          ▼               ▼              ▼
    sys_config_     sys_lang_pack   sys_lang_pack_
    version         _string         release
```

### 5.2 admin-server 关键文件

| 文件                                                                                                                                                                       | 行数          | 职责                                 | 状态        |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------- | ---------------------------------- | --------- |
| [models/sys\_config\_version.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/admin-server/app/admin/models/sys_config_version.go)        | 34          | 配置版本 Model                         | 正常        |
| [models/sys\_lang\_pack.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/admin-server/app/admin/models/sys_lang_pack.go)                  | 210         | 语言包 Model + DTO（含批量导入/导出/多语言视图结构体） | 正常        |
| [models/sys\_lang\_pack\_release.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/admin-server/app/admin/models/sys_lang_pack_release.go) | 79          | 发布记录 Model + DTO                   | 正常        |
| [service/sys\_config.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/admin-server/app/admin/service/sys_config.go)                       | 183         | 配置 CRUD                            | 正常        |
| [service/lang\_pack\_service.go](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/admin-server/app/admin/service/lang_pack_service.go)        | **1267** ⚠️ | 语言包全量管理（CRUD/差异/导入/导出/YAML生成）      | **超限待拆分** |

***

## 6. 痛点诊断汇总

### 6.1 🔴 高优先级痛点

#### P-01: 双重语言包存储导致数据不一致风险

**现象**:

* `sys_config_version` 中存储了 `lang_pack_default`(v5) 和 `lang_pack_weba`(v8)

* `sys_lang_pack` + `sys_lang_string` 也存储了相同数据（v3, zh-CN + en）

* 两套数据的**版本号不同步**、**内容可能漂移**

**影响链路**:

```
6005 协议 (AppFetchLangPack)
    └▶ GetLangPackConfig() 
        └▶ 查询 sys_config_version.lang_pack_*  ← 旧路径（v5/v8）
        
6007 协议 (AppFetchLangDifference)
    └▶ GetAppLangDifference()
        └▶ 查询 sys_lang_pack + sys_lang_string  ← 新路径（v3）
```

客户端通过 6005 拿到的版本是 v8，通过 6007 拿到的是 v3，**版本基准不一致**。

**涉及文件**: [config\_query.go:349-411](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/phanethoodGO/modules/readonly-query/query/config_query.go)

***

#### P-02: config\_query.go 超限（563 行），维护困难

**现象**: 单文件包含 5 个函数，其中 `GetAppConfigs()` 一个函数就占了约 300 行（10 个模块的 JSON 解析硬编码）

**代码片段示例** (config\_query.go:88-136):

```go
// 每个 module 都是这样手写匿名 struct 解析
if r, ok := latestRows["member_cfg"]; ok {
    var v struct {
        MaxFreeGroupNormal      int32 `json:"max_free_group_normal"`
        MaxFreeGroupPremium     int32 `json:"max_free_group_premium"`
        // ... 每次新增字段都要来这里加
    }
    if err := json.Unmarshal([]byte(r.ConfigJSON), &v); err == nil {
        if v.MaxFreeGroupNormal > 0 { c.Base.MaxFreeGroupNormal = v.MaxFreeGroupNormal }
        // ... 每个字段都要手动赋值
    }
}
```

**痛点**:

* 新增配置字段需同时改: proto → config\_query.go 解析器 → admin-server 填充 → admin-web 表单

* 任何一环遗漏不会报编译错误，只会在运行时返回默认值

***

### 6.2 🟡 中优先级痛点

#### P-03: 缓存 Key 硬编码 `v3`

**代码位置**: [config\_query.go:36](/Users/mac/StudioProjects/2026/open-citycloud-workspace/golang/MineplanetGo/phanethoodGO/modules/readonly-query/query/config_query.go)

```go
cacheKey := fmt.Sprintf("config:app:%s:v3", env) // 手动维护
```

**问题**: 每次修改配置解析逻辑（如新增 `topic_limit_cfg` 字段）都需要手动递增 `v3` → `v4`，易遗漏。

#### P-04: member\_cfg 版本号格式异常

**数据**: `version = 1766815228`（看起来是 Unix 时间戳）
**其他模块**: `version = 1, 2, 3 ...`（递增整数）
**影响**: 如果有逻辑基于版本号做比较（如 `version > clientVersion`），时间戳格式的版本号会导致行为异常。

#### P-05: lang\_pack\_service.go 超限（1267 行）

**问题**: 包含语言包 CRUD、字符串管理、差异计算、CSV 导入导出、YAML 生成等 **8 类不同职责**的功能。

#### P-06: pay\_cfg 历史版本堆积

dev 环境有 16 个历史版本（v6-v22），test/prod 只有 v22。虽然不影响功能（只取最新 published），但占用存储且查询时需过滤。

***

### 6.3 🟢 低优先级痛点

#### P-07: 缺少配置变更审计日志

当前 `sys_config_version` 有 `update_by` 字段，但缺少一张独立的操作日志表来回答 "谁在什么时候把 pay\_cfg 的价格从 9.9 改成了 19.9" 这类问题。

#### P-08: admin-web 页面风格不统一

配置管理页面、语言包管理页面、发布管理页面三个页面的交互风格和布局不一致。

#### P-09: sys\_lang\_pack.pack\_name 字段冗余

当前只有 `webp` 一种值，但设计上支持多种 pack\_name。如果永远只用一种，这个字段可以简化掉。

***

## 7. 部件关系全景图

```
┌─────────────────────────────────────────────────────────────────────┐
│                           客户端层                                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ Flutter  │  │ React Web│  │ iOS      │  │ Android  │           │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘           │
│        └───────────────┴────────────┴───────────────┘                 │
│                         │                                         │
│                   Protobuf 6001/6003/6005/6007                      │
└─────────────────────────┼─────────────────────────────────────────┘
                          │
┌─────────────────────────▼─────────────────────────────────────────┐
│                        网关层 (Gateway)                             │
│                   router.yaml 路由 MaxType=6000                     │
└─────────────────────────┼─────────────────────────────────────────┘
                          │
┌─────────────────────────▼─────────────────────────────────────────┐
│                    ThirdServer (TarsGo)                             │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │              ThirdObjImp (handlers/)                       │     │
│  │  • 初始化时注入 readonly-query.QueryService               │     │
│  │  • 接收 PB 请求 → 调用 QueryService → 返回 PB 响应        │     │
│  └──────────────────────┬───────────────────────────────────┘     │
│                         │                                          │
│  ┌──────────────────────▼───────────────────────────────────┐     │
│  │            readonly-query 模块 (modules/)                 │     │
│  │                                                           │     │
│  │  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │     │
│  │  │config_query │  │ config_       │  │ cache/         │  │     │
│  │  │.go (563行)⚠️ │  │initializer.go│  │  Redis缓存     │  │     │
│  │  └──────┬──────┘  └──────────────┘  └────────────────┘  │     │
│  │         │                                                   │     │
│  │  ┌──────▼──────────────────────────────────────────┐     │     │
│  │  │              数据库查询                            │     │     │
│  │  │  GetAppConfigs()     → sys_config_version (10模块)│     │     │
│  │  │  GetLangPackConfig() → sys_config_version ⚠️ 冗余  │     │     │
│  │  │  GetAppLanguage()    → sys_app_language           │     │     │
│  │  │  GetAppLangStrings()  → sys_lang_pack+string ✅   │     │     │
│  │  │  GetAppLangDiff()     → sys_lang_pack+string ✅   │     │     │
│  │  └──────────────────────────────────────────────────┘     │     │
│  └───────────────────────────────────────────────────────────┘     │
└─────────────────────────┬─────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
┌─────────────────┐ ┌─────────────┐ ┌──────────────────┐
│  MySQL          │ │ Redis       │ │ Aliyun OSS       │
│  mineplanet_   │ │ 缓存层      │ │ (语言包离线下载)   │
│  community_db  │ │             │ │                   │
└─────────────────┘ └─────────────┘ └──────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                       后台运维层 (admin)                              │
│                                                                     │
│  ┌─────────────────────── web 层 (admin-web) ──────────────────┐  │
│  │  Vue2 + Element UI                                            │  │
│  │  • /views/admin/sys-config     配置管理                      │  │
│  │  • /views/lang-management       语言包CRUD                    │  │
│  │  • /views/lang-pack-server      发布管理                      │  │
│  └───────────────────────┬──────────────────────────────────────┘  │
│                          │ REST API                               │
│  ┌───────────────────────▼──────────────────────────────────────┐  │
│  │               server 层 (admin-server)                        │  │
│  │                                                                │  │
│  │  apis/  ──▶  service/  ──▶  models/                         │  │
│  │                                                                │  │
│  │  写入:                                                         │  │
│  │  • sys_config_version (通过 sys_config 管理)                  │  │
│  │  • sys_lang_pack (通过 lang_pack_service 管理) ⚠️1267行     │  │
│  │  • sys_lang_string (通过 lang_pack_service 管理)              │  │
│  │  • sys_lang_pack_release (通过 lang_pack_release_service)     │  │
│  │                                                                │  │
│  │  读取:                                                         │  │
│  │  • 配置列表/编辑/发布状态                                      │  │
│  │  • 多语言列视图 (GetLangStringsMultiLang)                      │  │
│  │  • CSV 导入/导出                                               │  │
│  │  • 差异对比 / 发布前校验                                       │  │
│  └────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

***

## 8. 给新项目经理的建议事项

### 8.1 接手后优先确认项

* [ ] **确认 6005 协议的实际调用方和频率**: 如果 6005 已无客户端使用，可直接下线 `sys_config_version` 中的 `lang_pack_*` 数据

* [ ] **确认 member\_cfg 版本号的含义**: 是故意用时间戳还是 bug？这决定了版本比较逻辑是否需要修复

* [ ] **确认 pay\_cfg 历史版本的清理策略**: 是否可以定期清理非 latest 的已发布版本

### 8.2 建议的重构优先级（供参考，非强制）

| 优先级    | 任务                                                       | 理由        |
| ------ | -------------------------------------------------------- | --------- |
| **P0** | 统一 6005/6007 数据源到 `sys_lang_pack` 二级表                    | 消除数据不一致风险 |
| **P0** | 清理 `sys_config_version` 中的 `lang_pack_default/weba` 冗余数据 | 减少维护负担    |
| **P1** | 拆分 `config_query.go`（563→多个 <200 文件）                     | 提升可维护性    |
| **P1** | 拆分 `lang_pack_service.go`（1267→多个 <200 文件）               | 提升可维护性    |
| **P1** | 缓存 Key 改为基于 DB version 或 updated\_at 自动生成                | 消除人工维护    |
| **P2** | 统一 admin-web 三个管理页面为"配置中心"                               | 提升运维体验    |
| **P2** | 增加配置变更审计日志                                               | 可追溯性      |

### 8.3 风险提示

1. **6005 升级前必须确认所有客户端已切换到 6007 增量方案**，否则老客户端将无法获取语言包
2. **任何** **`sys_config_version`** **的变更都应先在 dev 环境验证**，因为该表直接影响 6001 协议（App 启动首屏配置）
3. **`sys_lang_string`** **的** **`version`** **字段是增量更新的核心**，不要手动修改数据库中的版本号

***

## 附录: 验证 SQL

```sql
-- ============================================
-- 新项目经理上手验证用 SQL
-- ============================================

-- 1. 查看四张表的记录数
SELECT 'sys_config_version' AS tbl, COUNT(*) FROM sys_config_version
UNION ALL SELECT 'sys_lang_pack', COUNT(*) FROM sys_lang_pack
UNION ALL SELECT 'sys_lang_pack_release', COUNT(*) FROM sys_lang_pack_release
UNION ALL SELECT 'sys_lang_string', COUNT(*) FROM sys_lang_string;

-- 2. 查看 config_version 中是否有冗余的 lang_pack 数据
SELECT * FROM sys_config_version 
WHERE module_key LIKE 'lang_pack%' 
ORDER BY module_key, env, version DESC;

-- 3. 查看 member_cfg 版本号异常情况
SELECT * FROM sys_config_version 
WHERE module_key = 'member_cfg';

-- 4. 对比两套语言包数据的一致性
-- 4.1 sys_config_version 中的语言包版本
SELECT module_key, env, version, 
       JSON_LENGTH(config_json) as json_size
FROM sys_config_version 
WHERE module_key IN ('lang_pack_default', 'lang_pack_weba') AND is_published = 1;

-- 4.2 sys_lang_pack 中的语言包版本
SELECT id, pack_name, env, lang_code, version, is_published FROM sys_lang_pack;

-- 5. 查看各模块最新发布版本（用于评估配置完整性）
SELECT module_key, env, MAX(version) as latest_ver, COUNT(*) as total_revisions
FROM sys_config_version
WHERE is_published = 1
GROUP BY module_key, env
ORDER BY module_key, env;

-- 6. 查看语言字符串分组分布
SELECT sls.`group`, COUNT(*) as key_count,
       COUNT(DISTINCT sls.string_key) as unique_keys
FROM sys_lang_string sls
JOIN sys_lang_pack lp ON sls.pack_id = lp.id
WHERE sls.operation_type != 'DEL'
GROUP BY sls.`group`
ORDER BY key_count DESC;
```

***

> **文档结束**。本报告基于 bizcheck skill 对 4 张核心表的真实查询结果（共 531 条记录）、Protobuf 协议定义（277 行）、服务端代码（3 个核心文件，约 2200 行）、后台运维代码（admin-server 5 个 model + 2 个 service，admin-web 6 个页面/API）全面诊断生成。

<br />

```yaml
相关redis,mysql链接

  localhost:8000:
      driver: mysql
      source: root:123456@tcp(192.168.1.6:3306)/go_admin?charset=utf8mb4&parseTime=True&loc=Local&timeout=30000ms&readTimeout=30000ms&writeTimeout=30000ms
    community:
      driver: mysql
      source: mpuser:Huawei@2025@tcp(rm-t4np5ht1x04y8ko98eo.mysql.singapore.rds.aliyuncs.com:3306)/mineplanet_community_db?charset=utf8mb4&parseTime=True&loc=Local&timeout=30000ms&readTimeout=30000ms&writeTimeout=30000ms

# Redis Configuration (Shared)
redis:
  enabled: true
  host: "192.168.1.6"
  port: 6379
  password: ""
  db: 0
```

<br />

<br />

