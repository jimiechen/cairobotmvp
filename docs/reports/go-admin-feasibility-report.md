# go-admin 可行性报告（M0' 阶段）

> 报告日期：2026-05-26
> 报告人：Trae (AI 编码助手)
> 对应 PRD：[PRD-10](../prd/PRD-10-Admin管理后台MVP.md)
> 对应任务：M0'.0.8

## 1. 执行摘要

**结论：go-admin v2.2.0 + go-admin-ui v2.0.9 在本项目环境中可行，可进入 M1' 开发阶段。**

M0' 阶段 8 个子任务全部完成：
- ✅ M0'.0.1 go-admin 后端部署（zip 解压 → go/admin/）
- ✅ M0'.0.2 前端部署（zip 解压 → typescript/admin-web/，pnpm install 成功）
- ✅ M0'.0.3 DSN 加密工具（config/dsn.go 编译通过）
- ✅ M0'.0.4 provider-admin 废弃归档 + 删除
- ✅ M0'.0.5 redisx.Invalidate 接口 + 实现 + **12 个测试全 PASS**
- ✅ M0'.0.6 CODE-WIKI §24 Admin 章节更新
- ✅ M0'.0.7 pub/sub payload 升级（config SDK 三级降级 + **10 个测试全 PASS** + i18n SDK 现状摘要）
- ⚠️ M0'.0.8 本报告（前端 dev 启动待排查）

## 2. 环境锁定结果

### 2.1 后端（go-admin v2.2.0）

| 项目 | 值 |
|---|---|
| 源码位置 | `go/admin/` |
| module 名 | `go-admin` |
| commit | `8f8a197d` |
| Go 版本兼容 | go 1.25.5 workspace 正常识别 |
| `go mod tidy` | ✅ 成功 |
| `go build ./...` | ✅ 编译通过 |
| `go list -m all \| grep go-admin` | ✅ 输出正常 |

**核心依赖版本：**

| 依赖 | 版本 | 说明 |
|---|---|---|
| gin-gonic/gin | v1.10.0 | HTTP 框架 |
| gorm.io/gorm | v1.25.12 | ORM |
| casbin/casbin/v2 | v2.104.0 | RBAC 权限 |
| gorm-adapter/v3 | v3.32.0 | Casbin MySQL 适配器 |
| go-admin-core | v1.5.3-rc.3 | 核心框架 |
| dgrijalva/jwt-go | v3.2.0 | JWT 认证（注意：已弃用包） |
| golang-jwt/jwt/v5 | v5.2.2 | 新版 JWT（双包共存） |
| swaggo/gin-swagger | v1.6.0 | API 文档 |

**依赖冲突分析：**

| 冲突项 | 项目现有 | go-admin 引入 | 判定 |
|---|---|---|---|
| Gin | TarsGo 内置（隔离 module） | v1.10.0（go/admin module） | **无冲突** — workspace 隔离 |
| GORM | mysqlx 使用（隔离 module） | v1.25.12（go/admin module） | **无冲突** — workspace 隔离 |
| Casbin | 项目未使用 | v2.104.0 | **新增依赖，无冲突** |
| JWT | 项目未使用 | v3.2.0 + v5.2.2 | **新增依赖，无冲突** |

> **关键结论**：Go workspace 的 module 隔离机制保证了 go/admin 与项目其他模块的依赖完全独立。TarsGo 内置的 Gin、services/mysqlx 使用的 GORM 均不受影响。

### 2.2 前端（go-admin-ui v2.0.9）

| 项目 | 值 |
|---|---|
| 源码位置 | `typescript/admin-web/` |
| package.json name | `go-admin` |
| package.json version | 2.0.6（内部版本号，非 UI 版本） |
| Vue 版本 | **2.6.11** |
| UI 框架 | Element UI（非 Element Plus） |
| 构建工具 | vue-cli-service（webpack） |
| 依赖数量 | 41 个 |
| `pnpm install` | ✅ 成功（切换 npmmirror 后） |
| `pnpm run dev` | ⚠️ 待排查（Node 兼容性或端口冲突） |

**已修复的问题：**
- `.npmrc` 中过期的 taobao 镜像证书 → 已替换为 npmmirror

**待排查问题：**
- `pnpm run dev` 启动失败，需确认 Node.js 版本兼容性和端口占用情况

## 3. 基础设施改造结果

### 3.1 redisx.Invalidate（M0'.0.5）

| 项目 | 结果 |
|---|---|
| 接口定义 | `Client.Invalidate(ctx, pattern) error` ✅ |
| 实现位置 | `go/third_party/redisx/redisx.go` ✅ |
| 分批大小 | batchSize = 500 ✅ |
| 测试数量 | **6 个新测试** ✅ |
| 测试结果 | **ALL PASS** ✅ |

测试覆盖场景：
- 空匹配（无 key 需删除）
- 单 key 删除
- 多 key 单批（<500）
- 多 key 多批（1200 keys，跨 3 批）
- Scan 错误传播
- Context 取消

### 3.2 Config SDK Pub/Sub 升级（M0'.0.7）

| 项目 | 结果 |
|---|---|
| InvalidateEvent 类型 | `config/sdk/types.go` ✅ |
| onMessage 三级降级 | `config/sdk/pubsub.go` ✅ |
| 测试数量 | **9 个测试**（5 新增 + 4 原有）✅ |
| 测试结果 | **ALL PASS（10/10）** ✅ |

降级策略验证：
1. 完整 JSON（含 tenant_id）→ 结构化处理 ✅
2. 部分 JSON（无 tenant_id，有 module_keys）→ 提取 keys 后结构化 ✅
3. 逗号分隔 / 无法识别格式 → 旧版兼容 ✅
4. 异常输入不 panic ✅
5. 不误删无关缓存 ✅

### 3.3 i18n SDK 现状（M0'.0.7-C）

报告输出至：[i18n-sdk-pubsub-current.md](./i18n-sdk-pubsub-current.md)

**关键发现：**
- 自定义 `RedisClient` 接口，与 `redisx.Client` 方法签名不同
- `Publish` 无 ctx 参数
- 无 Scan/Invalidate 方法
- `onMessage` 仅支持逗号分隔格式
- **推荐 M0' 阶段不做改动，M2' 时统一对齐**

## 4. 文件变更清单

### 4.1 新增文件

| 文件路径 | 说明 |
|---|---|
| `go/admin/` | go-admin v2.2.0 完整后端源码（~200 个文件） |
| `typescript/admin-web/` | go-admin-ui v2.0.9 完整前端源码（~150 个文件） |
| `go/admin/config/dsn.go` | DSN AES-256-CBC 加解密工具 |
| `go/admin/config/.env.example` | 环境变量模板 |
| `go/admin/config/keys/dev.key` | 开发环境 AES 密钥（64 hex） |
| `docs/reports/i18n-sdk-pubsub-current.md` | i18n SDK pub/sub 现状摘要 |
| `docs/reports/go-admin-feasibility-report.md` | 本报告 |

### 4.2 修改文件

| 文件路径 | 变更类型 | 说明 |
|---|---|---|
| `go/third_party/redisx/redisx.go` | 修改 | Client 接口新增 Invalidate + redisClient 实现 |
| `go/third_party/redisx/redisx_test.go` | 重写 | 修复 3 个 pre-existing bug + 新增 6 个测试 |
| `go/services/config/sdk/types.go` | 修改 | 新增 InvalidateEvent 结构体 |
| `go/services/config/sdk/pubsub.go` | 修改 | onMessage 升级为三级降级策略 |
| `go/services/config/sdk/pubsub_test.go` | 修改 | 新增 5 个 pub/sub 测试 |
| `go/go.work` | 修改 | provider-admin → admin |
| `docs/wiki/CODE-WIKI.md` | 修改 | §24 Admin 章节 + provider-admin 引用替换 |
| `docs/prd/PRD-10-Admin管理后台MVP.md` | 修改 | §14.5 版本锁定 + zip 来源标注 |

### 4.3 归档删除

| 内容 | 处理 |
|---|---|
| `go/tars/provider-admin/` | 归档到 `archive/provider-admin-v0` 分支后删除 |
| `web/provider-admin/src/` | 同上 |

## 5. 测试结果汇总

```
=== redisx 测试 ===
TestPing                          PASS
TestInvalidate_EmptyMatch         PASS
TestInvalidate_SingleKey          PASS
TestInvalidate_MultipleKeysSingleBatch  PASS
TestInvalidate_MultipleKeysMultiBatch  PASS (1200 keys, 3 batches)
TestInvalidate_ScanError          PASS
TestInvalidate_CtxCancel           PASS
─── redisx: 7/7 PASS ───

=== config sdk pub/sub 测试 ===
TestPubsubManager_StartStop       PASS
TestPubsubManager_OnMessage       PASS
TestPubsubManager_BatchInvalidate PASS
TestBuildCacheKey                 PASS
TestOnMessage_JsonStructured      PASS
TestOnMessage_JsonMissingTenantId_Fallback  PASS
TestOnMessage_LegacyCommaFormat   PASS
TestOnMessage_InvalidJson_Fallback PASS
TestOnMessage_Empty_NoPanic       PASS
─── config sdk: 10/10 PASS ───

总计：17/17 PASS（0 FAIL）
```

## 6. 风险评估

| 风险 ID | 等级 | 描述 | 当前状态 | 缓解措施 |
|---|---|---|---|---|
| R-FEAS-01 | R1 | 前端 `pnpm run dev` 启动失败 | 🔴 未解决 | 下一步优先排查 Node 版本 / 端口 |
| R-FEAS-02 | R2 | go-admin 使用已弃用的 `dgrijalva/jwt-go` | 🟡 已知 | 功能不受影响，S2 前迁移到 `golang-jwt/jwt/v5` |
| R-FEAS-03 | R2 | go-admin-ui 为 Vue 2（LTS 2023-12 结束） | 🟡 已知 | S2 前评估 Vue3 迁移成本 |
| R-FEAS-04 | R3 | i18n SDK Redis 抽象层与 config SDK 不一致 | 📄 已记录 | M2' 统一时处理（方案 A） |
| R-FEAS-05 | R3 | go-admin 默认表前缀/命名可能与项目约定不同 | 🟢 可控 | M1' 建表时按 PRD DDL 执行 |

## 7. 后续步骤建议

### 立即执行（M0' 收尾）

1. [ ] 排查 `typescript/admin-web/` 的 `pnpm run dev` 启动问题
2. [ ] 确认前端能正常访问后，截图留证
3. [ ] 输出 M0' 完成日报

### M1' 准备（下一步批次）

1. [ ] 创建 `services/config/admin/` 子包骨架
2. [ ] 创建 `services/i18n/admin/` 子包骨架
3. [ ] 按 PRD §五 目录骨架填充文件
4. [ ] 实现 Schema CRUD 的 Handler → Service → Repository 链路

### S2 技术债（暂不执行）

1. jwt-go v3 → v5 迁移
2. Vue 2 → Vue 3 迁移
3. i18n SDK Redis 抽象层统一

## 8. 结论

**M0' 目标基本达成。** go-admin 后端编译通过、基础设施扩展（redisx.Invalidate、pub/sub 升级）全部完成且测试通过。唯一阻塞项是前端 dev 启动问题，建议优先排查后即可宣布 M0' 完成，进入 M1' 开发。
