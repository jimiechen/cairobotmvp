# proto-tester

> ⚠️ **内网开发工具，禁止公网部署**

## 项目定位

proto-tester 是 CaiRobot MVP 研发/QA 团队的**内部协议测试工具**，用于 Protobuf 协议联调场景下的 MessagePacket 构造、发送、响应解析和链路追踪。

## 技术栈

| 类别 | 技术 | 版本 | 用途 |
|------|------|------|------|
| Protobuf 运行时 | google-protobuf | 3.21.2 | 协议编解码（与 typescript/web 共享） |
| JSON 编辑器 | CodeMirror 6 | ^6.0.x | 请求体 JSON 编辑（<250KB） |
| UI 框架 | antd | ^5.x | 企业级组件库 |
| 状态管理 | zustand | ^4.5.5 | 轻量级状态容器 |
| 本地存储 | idb | ^8.0.0 | IndexedDB 封装（请求历史） |
| HTTP 客户端 | axios | ^1.7.2 | Gateway 通信 |
| 构建工具 | vite | ^5.4.0 | 开发服务器 + 打包 |
| 测试框架 | vitest | ^1.6.0 | 单元测试 + 覆盖率 |
| CLI 框架 | commander | 内置 | 命令行入口（send/trace/run/capture） |

## 快速启动

```bash
# 安装依赖
pnpm --filter proto-tester install

# 启动开发服务器
pnpm --filter proto-tester dev
```

访问 http://127.0.0.1:3001

默认代理 Gateway → http://localhost:8080

## 目录结构

```text
typescript/proto-tester/
├── src/
│   ├── components/          # React 组件
│   │   ├── ProtoFormRenderer.tsx    # 协议表单渲染器（核心）
│   │   ├── EnumSelect.tsx          # 枚举下拉选择器
│   │   ├── JsonEditor.tsx           # CodeMirror 6 JSON 编辑器
│   │   ├── TokenInjector.tsx        # Token 注入面板
│   │   ├── UserSwitcher.tsx         # 测试用户切换器
│   │   └── ExtendParamsPanel.tsx    # extend 参数编辑面板
│   ├── lib/                 # 核心业务逻辑（纯函数）
│   │   ├── apiClient.ts             # HTTP 请求封装
│   │   ├── messagePacket.ts         # MessagePacket 编解码
│   │   ├── protoFormBuilder.ts      # 动态表单 Schema 构建
│   │   ├── protoMetadata.ts         # 协议元数据管理
│   │   └── errors.ts                # 自定义错误类型
│   ├── store/               # Zustand 状态管理
│   │   ├── session.ts               # 会话状态（token/用户）
│   │   ├── protocols.ts             # 协议列表（筛选/收藏）
│   │   └── history.ts               # 请求历史（IndexedDB CRUD）
│   ├── routes/              # 页面路由组件
│   │   ├── history.tsx              # 历史记录页
│   │   └── trace.tsx               # 链路追踪页
│   ├── cli/                 # CLI 命令行工具
│   │   ├── index.ts                  # 入口（commander 注册）
│   │   ├── commands/
│   │   │   ├── send.ts              # send 子命令
│   │   │   └── trace.ts             # trace 子命令
│   │   └── reporter.ts              # 报告生成器
│   ├── data/                # 静态数据
│   │   ├── protocols.json            # 协议元数据（gen-protocols.mjs 生成）
│   │   ├── test_users.json          # 5 个测试账号配置
│   │   └── endpoints.default.json   # 默认端点配置
│   ├── utils/               # 工具函数
│   │   └── traceId.ts               # 链路追踪 ID 生成
│   ├── App.tsx               # 应用入口
│   └── main.tsx              # DOM 挂载
├── scripts/
│   └── gen-protocols.mjs     # 构建期：从 .proto 生成 protocols.json
├── tests/                   # 测试文件
│   ├── unit/                     # 单元测试（9 个文件 / 100 cases）
│   │   ├── T0-smoke.test.ts       # 冒烟测试
│   │   ├── apiClient.test.ts      # HTTP 客户端
│   │   ├── protoFormBuilder.test.ts # 表单构建
│   │   ├── security.test.ts        # 安全约束
│   │   └── cli/                    # CLI 命令测试
│   └── component/                # 组件测试
├── public/
│   ├── config/endpoints.json   # 运行时可覆盖的端点配置
│   └── proto/.gitkeep           # proto 文件占位
├── index.html                # CSP 硬编码入口
├── vite.config.ts            # Vite 配置（含覆盖率阈值）
└── package.json
```

## 测试账号

系统预置 5 个角色账号（见 `src/data/test_users.json`）：

| ID | 角色 | 名称 | tokenSource | 说明 |
|----|------|------|-------------|------|
| user_001 | admin | 管理员张三 | fixed | 全权限固定 Token |
| user_002 | operator | 运营李四 | fixed | 运维操作固定 Token |
| user_003 | viewer | 只读王五 | fixed | 只读查看固定 Token |
| user_004 | user | 普户赵六 | mock_jwt | 需手动输入 Token |
| user_005 | attacker | 越权测试 | fixed | 过期 Token（401 测试用） |

> 所有 Token 仅存内存，刷新页面后需重新装载。

## 安全规范

### Token 内存化
- Token **100% 存储在 Zustand store 内存中**
- 不写入 localStorage / sessionStorage / IndexedDB
- 刷新页面后自动清空，需手动重新装载
- console.log 仅输出前 8 位 + `***`，不输出完整 Token

### CSP 策略（硬编码于 index.html）

```
connect-src: self localhost:8080 127.0.0.1:8080
script-src:  self 'unsafe-inline'
style-src:  self 'unsafe-inline'
object-src: none
form-action: none
frame-ancestors: none
```

### 环境拦截
- CLI 的 `--env prod` 在 blocklist 中被拦截（退出码 3）
- 不存储生产环境账号

### 内网部署要求
- **禁止公网部署**
- 默认绑定 `127.0.0.1:3001`
- CSP connect-src 仅允许内网 Gateway 地址

## 开发命令一览表

```bash
# 开发
pnpm --filter proto-tester dev              # 启动开发服务器（:3001）

# 构建
pnpm --filter proto-tester build            # tsc 类型检查 + vite build
pnpm --filter proto-tester preview          # 预览构建产物

# 测试
pnpm --filter proto-tester test             # vitest watch 模式
pnpm --filter proto-tester test:run         # vitest 一次性运行
npx vitest run --coverage                   # 含覆盖率报告

# 协议元数据生成
node scripts/gen-protocols.mjs              # 从 .proto 生成 protocols.json

# Lint
pnpm --filter proto-tester lint             # ESLint 检查

# CLI
pnpm --filter proto-tester cli -- send --max 2100 --min 2101 --payload '{"name":"test"}'
```

## 端点配置

默认 Gateway: http://localhost:8080

可通过以下方式修改：
1. Settings 页面（UI 操作）
2. `public/config/endpoints.json`（文件修改，重启生效）
3. CLI `--gateway <url>` 参数（单次覆盖）

## CLI 模式（v2.0）

proto-tester 支持 4 个 CLI 子命令，可通过 Trae Skill 触发。

### 安装依赖后使用

```bash
# 单次发送
pnpm --filter proto-tester cli send --max 2100 --min 2101 --payload '{"name":"Trae"}'

# 运行用例集
pnpm --filter proto-tester cli run --suite tests/suites/admin-mvp-protocols.yaml

# 浏览器自动化（需先启动 dev server）
pnpm --filter proto-tester dev &  # 终端 1
pnpm --filter proto-tester cli capture --scenario tests/scenarios/hello-smoke.yaml --video on  # 终端 2

# traceId 查询
pnpm --filter proto-tester cli trace --id 7f3a2b1c...
```

### 退出码

| 码 | 含义 |
|---:|---|
| 0 | 成功 |
| 1 | 业务失败 |
| 2 | 传输失败 |
| 3 | 参数错误 / prod 被拦截 |
| 4 | API 不可达 |
| 5 | Playwright 缺失 |

### 多环境切换

```bash
# 默认 dev 环境
pnpm cli send --max 2100 --min 2101

# 切换到 test 环境
pnpm cli send --max 2100 --min 2101 --env test

# 尝试连接 prod 会被拦截（退出码 3）
pnpm cli send --env prod  # ✗ blocked!
```
