# PRD-11：proto-tester 协议测试客户端（v2.0 完整版）

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | PRD-11 |
| 名称 | proto-tester 协议测试客户端（Web UI + CLI 双模式） |
| 状态 | 已确认 |
| 优先级 | P0 |
| 创建日期 | 2026-05-27 |
| 最后更新 | 2026-05-27 |
| 创建人 | 项目团队 |
| 实施方案 | [docs/drafts/proto-tester-impl-v1.1.md](../drafts/proto-tester-impl-v1.1.md) |
| 总工期 | 10 天（T0~T6 共 6 天 Web UI + T7~T9 共 4 天 CLI） |

## 2. 背景

研发 / QA 团队在 MessagePacket + Protobuf 协议联调场景下，使用 Postman/Apifox 存在以下痛点：
- 无法直接处理 Protobuf 二进制序列化（需手动构造 bytes）
- 无法自动填充 MessagePacket.extend 标准字段
- 无法按协议编号注册表快速选择协议
- 缺少 traceId 端到端日志串联能力
- Trae Skill 自动化验收时缺少命令行入口，只能调 curl 或手写 fetch

proto-tester 定位为**研发/QA 内部工具**，提供 Web UI（人手联调）+ CLI（自动化验收）双模式，共享同一套 lib/ 层代码。

## 3. 目标

- 提供基于 google-protobuf 的协议编解码能力，替代 Postman 在 Protobuf 场景下的不足
- 提供 Web UI 模式：协议选择 / extend 参数 / proto schema 驱动表单 / 枚举 / traceId 检索 / Token 注入
- 提供 CLI 模式：send / run / capture / trace 四个子命令，支持 Trae Skill 触发
- 全部元数据来自构建期 gen-protocols.mjs 输出，运行时不接触 .proto 文件
- 安全约束：Token 100% 内存化 / CSP 硬编码 / 内网部署 / blocklist 禁止 prod

## 4. 非目标

- 不部署在公网
- 不接入业务鉴权体系
- 不复用 typescript/web 的业务组件库
- 不直连 TarsGo Servant
- 不做"伪客户端 SDK"
- 不成为生产监控工具
- 不引入 protobufjs 运行时
- 不使用 Monaco Editor

## 5. 使用角色

| 角色 | 使用场景 | 入口模式 |
|---|---|---|
| 研发工程师 | 协议联调 / 接口调试 | Web UI |
| QA 工程师 | 回归测试 / 冒烟测试 | CLI run |
| Trae Skill | 自动化端到端验收 | CLI send/run/capture/trace |
| 项目主控 | 审查报告 / 聚合日志 | CLI trace |

## 6. 技术决策摘要

| 决策项 | 选择 | 理由 |
|---|---|---|
| Protobuf 运行时 | google-protobuf（精确版本锁定） | 与 typescript/web 一致，避免双运行时 |
| JSON 编辑器 | CodeMirror 6（精确版本锁定） | <250KB vs Monaco 3.3MB |
| 状态管理 | zustand | 轻量，含 persist 中间件 |
| 本地存储 | idb（IndexedDB Promise 封装） | 支持复合索引 + 版本迁移 |
| HTTP 客户端 | axios | 成熟稳定 |
| CLI 框架 | commander.js | 社区标准，ESM 兼容 |
| 浏览器自动化 | Playwright（optionalDependencies） | 仅 capture 使用 |
| 包管理 | pnpm workspace | monorepo 统一管理 |

## 7. 任务分解总览

```
Phase 1 — Web UI 模式（T0 ~ T6，6 天，串行）
├── T0.0  方案补丁回灌验证          [0.5d] ──┐
├── T0    脚手架与依赖              [1d]   ──┤
├── T1    协议元数据 + 表单构建      [1.5d] ──┤ 前置：T0
├── T2    MessagePacket + API 客户端  [1d]   ──┤ 前置：T0, codes.go 摘要
├── T3    表单渲染 + 枚举            [1.5d] ──┤ 前置：T1
├── T4    用户切换 + Token 注入      [1d]   ──┤ 前置：T0
├── T5    traceId 日志检索 + 历史    [1d]   ──┤ 前置：T0
└── T6    测试 + 文档 + 部署        [1d]   ──┘ 前置：T1~T5

Phase 2 — CLI 模式（T7 ~ T9，4 天，串行）
├── T7    CLI 基础设施              [2d]   ──┐ 前置：T0~T6 全部通过
├── T8    用例集 + 浏览器自动化      [1.5d] ──┤ 前置：T7
└── T9    Trae Skill 接入 + 文档    [0.5d] ──┘ 前置：T8
```

---

## 8. Phase 1 详细任务（Web UI）

### 8.1 T0.0：方案补丁回灌验证（0.5 天）

**目标**：确认 v1.2 所有修订已落实到代码/配置层面。

**任务清单**：

| # | 任务 | 验证方式 | 对应修订 |
|---|---|---|---|
| T0.0.1 | 确认 package.json 无 protobufjs 依赖 | grep 验证 | 修订 1 |
| T0.0.2 | 确认有 google-protobuf + @proto/* 引用 | grep 验证 | 修订 1 |
| T0.0.3 | 确认 protoFormBuilder.ts 无 protobuf.js import | grep 验证 | 修订 2 |
| T0.0.4 | 确认 EnumSelect 含动态枚举缓存隔离前缀 | 代码审查 | 修订 3 |
| T0.0.5 | 确认 package.json 有 @codemirror/* 无 monaco-editor | grep 验证 | 修订 4 |
| T0.0.6 | 确认 history.ts 使用 idb 库 | 代码审查 | 修订 5 |
| T0.0.7 | 确认目录结构含 scripts/, JsonEditor.tsx, protoMetadata.ts | ls 验证 | 目录变更 |
| T0.0.8 | 确认 .eslintrc.cjs 引用 admin-web 配置（非副本） | diff 验证 | ESLint |
| T0.0.9 | 确认 T2 通过判据分两层写法 | 代码审查 | 修订 15 |
| T0.0.10 | 确认 docker-compose.dev.yml 创建逻辑存在 | 文件检查 | T6.6 |
| T0.0.11 | 确认 google-protobuf 精确版本与 typescript/web 一致 | diff 验证 | v1.2 收紧 1 |
| T0.0.12 | 确认 protocols.json 含 messageSchemas 字段 | JSON 校验 | v1.2 收紧 2 |
| T0.0.13 | 确认 public/config/endpoints.json 存在 | 文件检查 | v1.2 收紧 4 |
| T0.0.14 | 确认 index.html 含 CSP meta 标签 | grep 验证 | v1.2 拍板 1 |

**产出物**：
- T0.0 修订回灌 checklist（全勾选）
- [codes.go 错误码映射摘要](#appendix-b)（供 T2 使用）

**前置条件**：无

**通过判据**：
- [ ] 14 项 checklist 全部勾选
- [ ] codes.go 错误码映射摘要已输出到本文档附录 B

---

### 8.2 T0：脚手架与依赖（1 天）

**目标**：创建 pnpm workspace 工程，配置 Vite/ESLint/Prettier/TypeScript，验证 @proto/* alias 可用。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T0.1 | pnpm + Vite 创建工程，写入精确版本依赖 | package.json, vite.config.ts, tsconfig.json | `pnpm dev` 启动成功 |
| T0.2 | **创建 pnpm-workspace.yaml**（当前不存在），workspace 接入 | pnpm-workspace.yaml, root package.json | web 全量构建+单测 PASS |
| T0.3 | 配置 Vite（alias / proxy / optimizeDeps / server） | vite.config.ts | localhost:3001 可访问 |
| T0.4 | ESLint + Prettier 引用 admin-web 配置 | .eslintrc.cjs, .prettierrc | lint 通过 |
| T0.5 | 编写 README.md（安全规范 + 启动说明） | README.md | 信息完整 |

**关键细节 — T0.1 依赖清单**：

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "antd": "^5.x",
    "google-protobuf": "<精确版本，从 typescript/web/package.json 取>",
    "axios": "^1.x",
    "zustand": "^4.x",
    "react-router-dom": "^6.x",
    "@codemirror/state": "<精确版本>",
    "@codemirror/view": "<精确版本>",
    "@codemirror/lang-json": "<精确版本>",
    "@codemirror/theme-one-dark": "<精确版本>",
    "dayjs": "^11.x",
    "lodash-es": "^4.x",
    "uuid": "^9.x",
    "idb": "^8.x"
  }
}
```

> **操作步骤**：先执行 `grep '"google-protobuf"' typescript/web/package.json` 取当前精确版本，再写入 package.json。

**关键细节 — T0.2 workspace 兼容性验证（必过阻塞点）**：

```bash
# 必须全部 PASS，任一 FAIL 则 T0 阻塞
cd typescript/web && pnpm build   # 必须 PASS
cd typescript/web && pnpm test:run # 必须 PASS
```

**产出物**：
- 可运行的 Vite + React 脚手架（http://127.0.0.1:3001）
- [docs/runbook/pnpm-workspace.md](../runbook/pnpm-workspace.md)

**前置条件**：无

**通过判据**：
- [ ] `pnpm --filter proto-tester dev` 启动 vite 成功
- [ ] 默认页可访问（http://127.0.0.1:3001）
- [ ] `pnpm --filter proto-tester test` 跑通空测试
- [ ] `@proto/*` alias 可正常 import（如 `import { MessagePacket } from '@proto/base/message'`）
- [ ] **typescript/web 全量构建 + 单测 PASS（T0.2.d 必过）**
- [ ] CodeMirror 6 可正常 import 和基础渲染
- [ ] 9 处修订 checklist 全部通过（T0.0 产出）

---

### 8.3 T1：协议元数据 + ProtoFormBuilder（1.5 天）

**目标**：实现构建期协议元数据生成 + 运行时表单 Schema 构建，主验证目标为 app_config.proto（5 层嵌套复杂度）。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T1.1 | 实现 scripts/gen-protocols.mjs | scripts/gen-protocols.mjs | 两次运行结果一致（幂等） |
| T1.2 | 实现 src/lib/protoMetadata.ts | src/lib/protoMetadata.ts | 单测 ≥80% |
| T1.3 | 实现 src/lib/protoFormBuilder.ts（messageSchemas 驱动） | src/lib/protoFormBuilder.ts | 单测 ≥80%，AppConfigsReq 覆盖 |
| T1.4 | 实现 src/store/protocols.ts（Zustand） | src/store/protocols.ts | 列表/收藏/搜索可用 |
| T1.5 | gen-protocols.mjs 单元测试 | tests/unit/gen-protocols.test.ts | 注册表解析正确 |

**关键约束（v1.2 收紧）**：
- protoFormBuilder **唯一数据来源** = protocols.json.messageSchemas（构建期产物）
- **禁止**：运行时 import .proto 文件
- **禁止**：用 .toObject() 推断字段
- **禁止**：用 keysOf(class.prototype) 扫方法名推断字段

**主验证目标**：

| 目标 | 级别 | 要求 |
|---|---|---|
| AppConfigsReq + AppConfigsRsp | **主验证** | 5 层嵌套 + map<string,string> + repeated + FieldDescriptor nested 全覆盖 |
| HelloWorldRequest | smoke | 基本渲染即可 |
| HealthCheckRequest | smoke | 基本 |

超限降级策略：嵌套 > 5 层 → 自动降级 CodeMirror 6 JSON 模式（设计内兜底，不算失败）。

**产出物**：
- src/data/protocols.json（含 protocols[] + messageSchemas{}）
- 可重入的 gen-protocols.mjs 构建脚本

**前置条件**：T0 通过

**通过判据**：
- [ ] gen-protocols.mjs 能从注册表生成正确的 protocols.json
- [ ] protoMetadata.ts 能正确查询协议元数据
- [ ] protoFormBuilder 单测覆盖率 ≥80%
- [ ] **AppConfigsReq/Rsp 完整渲染通过（主验证）**
- [ ] gen-protocols.mjs 可重入（两次运行结果一致）
- [ ] messageSchemas 含 12 条协议引用到的所有 message
- [ ] protoFormBuilder 不 import .proto 文件（grep 自检）

---

### 8.4 T2：MessagePacket 包装 + API 客户端（1 天）

**目标**：实现 Protobuf 二进制编解码 + axios 封装 + 错误码映射，通过判据拆为两层（离线即收工 + 联调延后）。

**前置产出**：codes.go 错误码映射摘要（T0.0.2 产出，见[附录 B](#appendix-b)）

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T2.1 | 实现 src/lib/messagePacket.ts（encodePacket / decodePacket） | src/lib/messagePacket.ts | buffer 一致性比对 |
| T2.2 | 实现 src/lib/apiClient.ts（axios + 拦截器 + 错误码映射） | src/lib/apiClient.ts | msw 拦截 6 条全覆盖 |
| T2.3 | 实现 src/store/session.ts（内存 token + endpoint） | src/store/session.ts | token 不持久化 |
| T2.4 | 单元测试（msw 拦截） | tests/unit/apiClient.test.ts | ≥80% 覆盖率 |

**错误码映射（必须 100% 覆盖）**：

| 常量 | 值 | apiClient 处理 |
|---|---:|---|
| CodeSuccess | 10200 | resolve(data) |
| CodeBadRequest | 10400 | reject(BadRequestError) |
| CodeUnauthorized | 10401 | 触发 Token 过期提示 |
| CodeNotFound | 10404 | reject(NotFoundError) |
| CodeInternalError | 10500 | reject(InternalError + traceId) |
| CodeTarsNotImplemented | 10501 | reject(NotImplementedError) |

**通过判据（两层拆分）**：

**层 1（必过，离线可验收）**：
- [ ] hello 协议 MessagePacket 二进制构建正确（encode → decode → field-by-field 比对）
- [ ] encode 输出与 typescript/web/src/utils/proto-client.ts 的 buildPacket 输出 buffer 一致
- [ ] apiClient.ts 单测 ≥80%（msw 拦截覆盖成功/失败/超时/错误码全路径）
- [ ] 错误码映射表 6 条覆盖率 100%
- [ ] axios 拦截器拒绝向 CSP connect-src 之外域名的请求

**层 2（联调验收，移交 T6.5）**：
- [ ] 真实 Gateway 端到端 hello 协议返回成功
- [ ] traceId 在 Gateway 日志中可检索
- [ ] extend.token 通过鉴权（返回非 10401）
- [ ] 若后端未就绪 → 标注 **P0 阻塞：等待后端环境就绪**

> **完工标准 = 层 1 全部 PASS**。层 2 不卡 T2 完工。

**前置条件**：T0 通过 + codes.go 摘要已产出

---

### 8.5 T3：表单渲染 + 枚举（1.5 天）

**目标**：实现 ProtoFormRenderer（9 种类型全覆盖）+ EnumSelect（静态/动态）+ ExtendParamsPanel + JsonEditor。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T3.1 | 实现 ProtoFormRenderer.tsx（受控表单 + 9 类型） | components/ProtoFormRenderer.tsx | 组件测试 ≥80% |
| T3.2 | 实现 EnumSelect.tsx（静态枚举 + 动态枚举缓存） | components/EnumSelect.tsx | 缓存隔离前缀正确 |
| T3.3 | 实现 ExtendParamsPanel.tsx | components/ExtendParamsPanel.tsx | 全局/协议级覆盖可用 |
| T3.4 | 实现 JsonEditor.tsx（CodeMirror 6） | components/JsonEditor.tsx | <250KB 体积 |
| T3.5 | 单元 + 组件测试 | tests/component/*.test.tsx | ≥80% 覆盖率 |

**9 种类型渲染规则**：

| Proto 类型 | 控件 | 特殊处理 |
|---|---|---|
| string | Input | 最大长度从 comment 解析 |
| int32/int64 | InputNumber | 有符号/无符号区分 |
| bool | Switch | 默认 false |
| bytes | 双模式：文件上传 / hex 文本 | hex 格式校验 |
| enum | Select | 从 enumValues 取值 |
| repeated | 列表（可增删行） | 子项递归渲染 |
| nested message | Collapse 折叠面板 | 深度限制 5 层 |
| oneof | Tab 选择器 | 蓝色"二选一"徽标 |
| map<K,V> | 键值对编辑器 | K=string 时用 Input |

**前置条件**：T1 通过

**通过判据**：
- [ ] 任意协议选中后表单自动渲染正确
- [ ] **AppConfigsReq/Rsp 嵌套 / repeated / oneof / map 全部可编辑（主验证）**
- [ ] 嵌套超限时自动降级 CodeMirror JSON 模式
- [ ] 动态枚举可加载 + 缓存 + 隔离（前缀 `cache:proto-tester:dynamic-enum:*`）
- [ ] 单测 + 组件测试覆盖率 ≥80%

---

### 8.6 T4：用户切换 + Token 注入（1 天）

**目标**：实现测试用户池 + Token 内存化注入 + CSP 硬编码安全约束。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T4.1 | 创建 test_users.json（≥5 个账号） | src/data/test_users.json | 无生产账号 |
| T4.2 | 实现 UserSwitcher.tsx | components/UserSwitcher.tsx | 5 用户一键切换 |
| T4.3 | 实现 TokenInjector.tsx（100% 内存化） | components/TokenInjector.tsx | 无 localStorage 持久化 |
| T4.4 | CSP 硬编码 + 安全约束 | index.html <head> | connect-src 仅 Gateway |
| T4.5 | 单元 + 安全测试 | tests/unit/security.test.ts | ≥80% 覆盖率 |

**CSP 配置（硬编码到 index.html）**：

```html
<meta http-equiv="Content-Security-Policy" content="
  default-src 'self';
  script-src 'self' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data:;
  connect-src 'self' http://localhost:8080 http://127.0.0.1:8080;
  font-src 'self' data:;
  object-src 'none';
  base-uri 'self';
  form-action 'none';
  frame-ancestors 'none';
">
```

**Token 安全铁律**：
- Token 仅存内存（Zustand session store 或闭包变量）
- **不写 localStorage**、**不写 sessionStorage**、**不写 IndexedDB**
- 刷新页面后需用户手动"装载"按钮重新注入
- UI 上保留"复制 hash 用于审计"按钮（纯计算展示，不存储任何地方）

**前置条件**：T0 通过

**通过判据**：
- [ ] 5 个测试账号一键切换成功
- [ ] 越权测试账号（user_005）能复现 401/403
- [ ] Token 永不出现在 console.log
- [ ] CSP 限制有效（DevTools 验证）
- [ ] **localStorage 中无 token 持久化键（vi.spyOn 断言）**
- [ ] 安全测试覆盖率 ≥80%

---

### 8.7 T5：traceId 日志检索 + 历史（1 天）

**目标**：实现 IndexedDB 历史存储（idb 库，复合容量策略）+ traceId 时间线页面 + 日志检索 stub。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T5.1 | 实现 history.ts（IndexedDB via idb，复合容量） | store/history.ts | 7天/1000条限制工作 |
| T5.2 | 实现 history.tsx（列表 + 筛选 + 详情 + 导出） | routes/history.tsx | 导出/清空功能可用 |
| T5.3 | 实现 trace.tsx（traceId 检索 + stub） | routes/trace.tsx | 本地 stub 数据可展示时间线 |
| T5.4 | TraceTimeline.tsx 单元测试 | tests/component/TraceTimeline.test.tsx | 排序/合并/乱序处理 |
| T5.5 | 一键复制 traceId + 服务端日志跳转 | utils/traceId.ts | 功能可用 |

**IndexedDB Schema（完整索引声明）**：

```
数据库：proto-tester / 版本：1 / Store：requestHistory
PRIMARY KEY: id (autoIncrement)
INDEX:
  by-traceId    keyPath: "traceId"
  by-timestamp  keyPath: "timestamp"
  by-protocol   keyPath: ["maxType", "minType"]
容量：7 天 OR 1000 条取交集（先到期者先生效）
清理：每次 add 后异步触发 + 启动时全量清理
```

**前置条件**：T0 通过

**通过判据**：
- [ ] 发送任意请求后历史立刻可见
- [ ] traceId 检索能展示时间线（即使仅本地 stub 数据）
- [ ] IndexedDB 复合容量限制工作正常（7 天 / 1000 条）
- [ ] 导出功能可用（JSON 下载）
- [ ] 清空功能有二次确认

---

### 8.8 T6：测试 + 文档 + 部署（1 天）

**目标**：全量测试 + 文档同步 + ADR 记录 + docker-compose + 端到端联调。

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T6.1 | 单元测试覆盖率达标 | coverage 报告 | ≥80% |
| T6.2 | 端到端测试（可选 Playwright） | tests/e2e/ | hello/health/app_config 三条 |
| T6.3 | README.md 更新 | README.md | 信息完整 |
| T6.4 | 文档同步（CODE-WIKI / LLM-WIKI / ADR-016） | docs/wiki/* | 同步完毕 |
| T6.5 | **端到端联调**（吸纳 T2 层 2） | 联调报告 | hello+AppConfigs 跑通或标注阻塞 |
| T6.6 | docker-compose.dev.yml | docker-compose.dev.yml | 仅监听 127.0.0.1:3001 |
| T6.7 | 文档补全（JSON Schema / 测试缺口 / 风险更新） | 多文件 | 缺口已补充 |
| T6.8 | ADR-016 记录 | docs/wiki/adr/ADR-016-proto-tester.md | Status: Accepted |

**前置条件**：T1~T5 全部通过

**通过判据**：
- [ ] `pnpm --filter proto-tester build` 成功
- [ ] `pnpm --filter proto-tester test` 覆盖率 ≥80%
- [ ] e2e（如启用）通过
- [ ] 文档齐全（README + CODE-WIKI + ADR-016）
- [ ] docker-compose.dev.yml 仅监听 127.0.0.1
- [ ] **T6.5 联调通过或标注阻塞原因**

---

## 9. Phase 2 详细任务（CLI 模式）

### 9.1 T7：CLI 基础设施（2 天）

**目标**：搭建 CLI 入口层（commander.js），实现 send/trace 两个最简命令 + reporter + endpoints 多环境扩展。

**前置条件**：T0~T6 全部通过（Web UI 完整可用）

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T7.1 | CLI 入口 commander.js 注册 4 子命令 | src/cli/index.ts | 仅 4 个子命令（grep 校验） |
| T7.2 | 实现 send 命令（复用 lib/messagePacket + lib/apiClient） | src/cli/commands/send.ts | 4 种退出码全覆盖 |
| T7.3 | 实现 trace 命令（复用 lib/apiClient） | src/cli/commands/trace.ts | 降级退出码 4 |
| T7.4 | 实现 reporter.ts（MD + JUnit XML） | src/cli/reporter.ts | 输出格式可解析 |
| T7.5 | endpoints.json 扩展为多环境 + blocklist | public/config/endpoints.json | prod 被硬拦截 |
| T7.6 | package.json bin 入口 + scripts | package.json | npx 可执行 |
| T7.7 | CLI 单元测试 | tests/unit/cli/*.test.ts | ≥80% 覆盖率 |

**关键约束**：
- **仅暴露 4 个子命令**（send/run/capture/trace），禁止顺手添加调试命令
- **bin 必须指向 dist/cli/index.js**（不能指向 src/）
- **blocklist 硬拦截 prod**（退出码 3，非 WARN）
- send 命令 **必须 import from lib/**（禁止在 cli/ 下重新实现序列化）

**endpoints.json 多环境结构**：

```json
{
  "default": "dev",
  "environments": {
    "dev": { "gateway": "http://localhost:8080", ... },
    "test": { "gateway": "http://test.gateway.internal:8080", ... },
    "staging": { "gateway": "http://staging.gateway.internal:8080", ... }
  },
  "blocklist": ["prod"]
}
```

**CLI 退出码全集**：

| 码 | 含义 | 触发条件 |
|---:|---|---|
| 0 | 成功 | 业务码 10200 |
| 1 | 业务失败 | 业务码非 10200 |
| 2 | 传输失败 | 超时/Gateway不可达 |
| 3 | 参数错误/prod 拦截 | 协议未注册/env 命中 blocklist |
| 4 | API 不可达 | trace 后端未就绪 |
| 5 | Playwright 缺失 | capture 命令但未安装 |

**通过判据**：
- [ ] `pnpm --filter proto-tester cli send --max 2100 --min 2101 --payload '{}' --user user_001` 跑通
- [ ] `pnpm --filter proto-tester cli trace --id 7f3a2b1c` 至少能输出 traceId 摘要
- [ ] `pnpm --filter proto-tester cli send --env prod` 被拒绝（退出码 3）
- [ ] CLI 单测覆盖率 ≥80%
- [ ] `npx proto-tester send ...` 可执行（bin 入口）
- [ ] **CLI send 输出 buffer 与 Web UI 发送 buffer 一致**（G1 测试缺口覆盖）

---

### 9.2 T8：用例集运行 + 浏览器自动化（1.5 天）

**目标**：实现 run 命令（YAML 驱动 + 7 种匹配器）+ capture 命令（Playwright 浏览器自动化）。

**前置条件**：T7 全部通过

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T8.1 | 实现 runner.ts（YAML 解析 + 匹配器 + 并行控制） | src/cli/runner.ts | 7 种匹配器全覆盖 |
| T8.2 | 实现 run 命令 | src/cli/commands/run.ts | summary.md + junit.xml 产出 |
| T8.3 | 实现 browser.ts + capture 命令（Playwright） | src/cli/browser.ts, commands/capture.ts | 视频+截图生成 |
| T8.4 | YAML Schema 校验（ajv） | src/cli/schema.ts | 格式错误退出码 3 |
| T8.5 | 测试用例与场景文件 | tests/suites/*.yaml, tests/scenarios/*.yaml | 12 协议全覆盖 |
| T8.6 | Web UI 组件 data-id 补全 | components/*.tsx | 覆盖率 ≥90% |

**7 种匹配器**：exact / regex / contains / minLength / maxLength / exists / jsonPath

**YAML 用例示例（admin-mvp-protocols.yaml，12 条协议）**：

> ⚠️ 注意：payload 字段名必须与 protocols.json messageSchemas 中定义的真实字段名一致，不得杜撰。

**capture 前提条件**：
- vite dev server 必须已在 http://127.0.0.1:3001 运行
- **未启动时 capture 应优雅失败（退出码 ≥ 3 + 可读错误信息），非 unhandled exception**

**data-id 命名规范（沿用 admin E2E 四段式）**：

| 组件 | data-id |
|---|---|
| ProtoFormRenderer 字段 | pt-form-{fieldName} |
| ResponseViewer | pt-response-json |
| UserSwitcher | pt-user-switcher |
| 发送按钮 | pt-btn-send |

**Playwright 隔离**：
- optionalDependencies，仅 capture 命令 import
- 缺失时提示安装 + 退出码 5
- send/run/trace 不引入 Playwright

**通过判据**：
- [ ] `pnpm cli run --suite tests/suites/admin-mvp-protocols.yaml --parallel 1` 至少跑通 1 个用例
- [ ] `pnpm cli capture --scenario tests/scenarios/hello-smoke.yaml --video on` 生成视频 + 截图
- [ ] JUnit XML 在 CI 中可被标准工具解析
- [ ] data-id 覆盖率 ≥90%（grep `data-id="pt-`）
- [ ] YAML 格式错误时退出码 3 + 可读错误信息
- [ ] **vite dev server 未启动时 capture 优雅失败**（M2 修复项）

---

### 9.3 T9：Trae Skill 接入 + 文档（0.5 天）

**目标**：注册 Trae Skill YAML + .gitignore 更新 + 文档同步 + 端到端验收。

**前置条件**：T8 全部通过

**任务清单**：

| # | 任务 | 产出文件 | 验证方式 |
|---|---|---|---|
| T9.1 | 创建 .trae/skills/proto-tester.yaml | .trae/skills/proto-tester.yaml | Skill 校验通过 |
| T9.2 | .gitignore 更新（排除证据产物） | .gitignore | *.png/*.webm/*.bin 已排除 |
| T9.3 | README.md 新增 CLI 章节 | README.md | 4 命令示例完整 |
| T9.4 | docs/wiki/modules/proto-tester.md 更新 | wiki 页面 | CLI + Skill 小节齐全 |
| T9.5 | ADR-016 增补 CLI 决策 | ADR-016 | 4 项决策记录 |
| T9.6 | 端到端人工验收 | 验收报告 | 4 个 Skill 命令均跑通 |

**Skill 关键约束（必须硬编码）**：

| 约束 | 值 | 检查方式 |
|---|---|---|
| envBlocklist | ["prod"] | grep yaml |
| maxCasesPerSuite | 100 | grep yaml |
| captureMaxParallel | 1 | grep yaml |
| timeoutSeconds | 600 | grep yaml |

**通过判据（可勾选格式）**：
- [ ] `.trae/skills/proto-tester.yaml` 通过 Trae Skill 校验
- [ ] 主控调用 Skill send → 报告文件存在且 Markdown 可解析
- [ ] 主控调用 Skill run-suite → junit.xml 存在且 CI 工具可解析
- [ ] 主控调用 Skill capture → video.webm 文件大小 > 0
- [ ] 主控调用 Skill trace → 输出含 traceId 的 Markdown 报告
- [ ] 录屏 + 截图本地化（.gitignore 已排除，不进 git）
- [ ] README + ADR-016 + CODE-WIKI 更新完毕

---

## 10. 禁止事项全集（30 条）

### Web UI 模式（#1 ~ #24）

| # | 禁止事项 | 来源 |
|---|---|---|
| 1 | 不部署公网，仅内网可达 | §1.2 |
| 2 | 不接入业务鉴权 | §1.2 |
| 3 | 不复用 typescript/web 业务组件 | §1.2 |
| 4 | 不依赖 admin-web 运维账号 | §1.2 |
| 5 | 不做伪客户端 SDK | §1.2 |
| 6 | 不绕过 Gateway 直连 Tars | §1.2 |
| 7 | 不写入业务库 | §1.2 |
| 8 | 不引入 protobufjs 运行时 | 修订 1 |
| 9 | 不使用 Monaco Editor | 修订 4 |
| 10 | 不允许 google-protobuf 版本不精确 | v1.2 收紧 1 |
| 11 | 不允许 protoFormBuilder 运行时 import .proto | v1.2 收紧 2 |
| 12 | 不允许 Token 以任何形式持久化 | v1.2 收紧 3 |
| 13 | 不允许 endpoints.json 放在 src/data/ | v1.2 收紧 4 |
| 14 | 不允许 CSP connect-src 含非 Gateway 域名 | v1.2 拍板 1 |
| 15 | 不允许 IndexedDB 索引缺少 keyPath | v1.2 收紧 4 |
| 16 | 不允许用 hasField/clearField 反推必填 | 修订 16 |
| 17 | 不允许动态枚举缓存用 admin-web 前缀 | 修订 17 |
| 18 | 不允许 T2 用层 2 判据卡完工 | 修订 15 |
| 19 | 不允许手写 LRU（用 idb 库） | 修订 6 |
| 20 | 不允许自维护 ESLint/Prettier 副本 | 修订 14 |
| 21 | 不允许将 proto-tester 接入业务鉴权/公网部署 | 总原则 |
| 22 | 不允许直连 Tars/绕过 Gateway | 总原则 |
| 23 | 不允许复用 typescript/web 业务组件库 | 总原则 |
| 24 | 不允许响应中执行任意 HTML（防 XSS） | 总原则 |

### CLI 模式（#25 ~ #30）

| # | 禁止事项 | 来源 |
|---|---|---|
| 25 | 禁止 CLI 连接 prod 环境（blocklist 硬校验） | v2.0 |
| 26 | 禁止录屏/截图进 git（.gitignore 强制排除） | v2.0 |
| 27 | 禁止 Skill 单次调用 > 100 用例 | v2.0 |
| 28 | 禁止 CLI 复用 typescript/web 业务组件 | v2.0 |
| 29 | 禁止 capture 跳过 vite dev server 直连生产页面 | v2.0 |
| 30 | 禁止 CLI 与 Web UI 各自维护 messagePacket/apiClient 副本 | v2.0 |

---

## 11. 风险登记表

| 编号 | 风险 | 概率 | 影响 | 状态 | 应对措施 |
|---|---|---:|---|---|---|
| R2 | app_config.proto 复杂度超出 ProtoFormRenderer | 中 | 高 | 🟡 降低 | T1 主验证目标；超限降级 CodeMirror JSON |
| R4 | pnpm workspace 影响现有 web 构建 | 低 | 中 | 🟢 监控 | T0.2 必须验证 web 全量构建 PASS |
| R5 | google-protobuf 预生成代码与 Vite ESM 兼容性 | 低 | 高 | 🟢 监控 | T0.1 验证 @proto/* alias |
| R6 | idb 库 API 学习曲线 | 低 | 低 | 🟢 接受 | API 简洁，社区文档完善 |
| R7 | CLI tsconfig 与 Vite tsconfig 目标冲突 | 中 | 高 | 🟡 监控 | T7.1 确认 target/compilerOptions 不冲突 |
| R8 | commander.js ESM 兼容性 | 低 | 中 | 🟢 监控 | 锁定版本前验证 pnpm workspace 下可正常 import |
| R9 | capture 依赖 vite dev server 导致 CI 无法全自动 | 高 | 中 | 🟡 降低 | capture 标注为手动验收命令 |
| R10 | Skill maxCasesPerSuite=100 当前无约束力 | 低 | 低 | 🟢 接受 | 协议增长到 50+ 时重新评估 |

---

## 12. 依赖关系图

```
T0.0 (0.5d) ──────────────────────────────┐
                                           │
T0   (1d)  ────────────────────────────────┤
         ├── T1 (1.5d) ──┬── T3 (1.5d) ───┤
         │               │                │
         ├── T2 (1d) ────┤                │
         │               │                │
         ├── T4 (1d) ────┤                │
         │               │                │
         ├── T5 (1d) ────┤                │
         │               │                │
         └───────────────┴── T6 (1d) ─────┘ Phase 1 完成（6d）

                                          │
T7   (2d)  ──────────────────────────────┤ Phase 2 前提：T0~T6
         ├── T8 (1.5d) ───────────────────┤
         └───────────────┴── T9 (0.5d) ───┘ Phase 2 完成（4d）
                                            总计 10 天
```

---

## 13. 验收总 Checklist

### Phase 1 验收（T0~T6 完成后）

- [ ] Web UI 可在 http://127.0.0.1:3001 正常访问
- [ ] 12 条协议均可选择并渲染表单
- [ ] AppConfigsReq/Rsp（5 层嵌套）完整可编辑
- [ ] send hello 协议返回成功（真实 Gateway 或 msw mock）
- [ ] traceId 可检索并展示时间线
- [ ] Token 100% 内存化，刷新后需重新装载
- [ ] CSP 策略有效（connect-src 仅 Gateway）
- [ ] 单测覆盖率 ≥80%
- [ ] IndexedDB 复合容量策略工作正常
- [ ] typescript/web workspace 构建不受影响

### Phase 2 验收（T7~T9 完成后）

- [ ] `pnpm cli send` 发送成功并生成 Markdown 报告
- [ ] `pnpm cli run` 执行 YAML 用例集并生成 JUnit XML
- [ ] `pnpm cli capture` 生成视频 + 截图（需 vite dev server 运行中）
- [ ] `pnpm cli trace` 输出 traceId 日志摘要
- [ ] `--env prod` 被硬拦截（退出码 3）
- [ ] Skill YAML 4 个命令均可被 Trae 触发
- [ ] 录屏/截图证据不进 git（.gitignore 验证）
- [ ] CLI 单测覆盖率 ≥80%
- [ ] CLI 与 Web UI 共享 lib/ 无代码分裂（grep 审计）

---

## 附录 A：相关文档

| 文档 | 路径 | 说明 |
|---|---|---|
| 实施方案 v2.0 | docs/drafts/proto-tester-impl-v1.1.md | 本 PRD 的技术依据 |
| 协议编号注册表 | docs/api/协议编号注册表.md | 12 条协议源数据 |
| message.proto | proto/base/message.proto | MessagePacket 定义 |
| app_config.proto | proto/base/app_config.proto | 主验证目标 proto |
| ADR-016 | docs/wiki/adr/ADR-016-proto-tester.md | 架构决策记录 |
| CODE-WIKI | docs/wiki/CODE-WIKI.md | 第 9 章 proto-tester 小节 |

## 附录 B：错误码映射表（T2 前置产出）

来源：go/common-lib/codes.go

| 常量 | 值 | HTTP 映射 | 前端处理 |
|---|---|---|---|
| CodeSuccess | 10200 | 200 | resolve(data) |
| CodeBadRequest | 10400 | 400 | reject(new BadRequestError(message)) |
| CodeUnauthorized | 10401 | 401 | 触发 Token 过期 → UserSwitcher 提示重新登录 |
| CodeNotFound | 10404 | 404 | reject(new NotFoundError(message)) |
| CodeInternalError | 10500 | 500 | reject(new InternalError(`服务端错误，traceId: ${traceId}`)) |
| CodeTarsNotImplemented | 10501 | 501 | reject(new NotImplementedError('S1 阶段暂未实现')) |

## 附录 C：测试用户池

| ID | 名称 | 角色 | tokenSource | 用途 |
|---|---|---|---|---|
| user_001 | 管理员张三 | admin | fixed | 全权限测试 |
| user_002 | 运营李四 | operator | fixed | 操作员权限测试 |
| user_003 | 只读王五 | viewer | fixed | 只读权限测试 |
| user_004 | 普户赵六 | user | mock_jwt | 普通用户场景 |
| user_005 | 越权测试 | attacker | fixed(过期) | 401/403 复现 |

---

*本文档由实施方案 v2.0 拆分而来，作为 T0~T9 执行的任务级依据。*
*状态：已确认，可进入开发。*
