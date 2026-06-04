# ADR-016: proto-tester 协议测试客户端技术选型

## 状态: Accepted

## 背景
研发/QA 团队在 Protobuf 协议联调场景下缺少专用工具。现有方案（Postman / grpcurl）存在以下痛点：
1. Postman 不支持 Protobuf 原生编解码，需手动构造二进制或依赖后端提供 JSON 转换接口
2. grpcurl 是命令行工具，无 GUI，不适合非开发角色（QA/产品）
3. 缺少链路追踪（TraceID）和请求历史管理能力
4. 无法在团队内共享协议元数据和测试账号配置

## 决策

### 1. Protobuf 运行时：google-protobuf 3.21.2（不使用 protobufjs 运行时）

**选择理由：**
- typescript/web 子项目已使用 google-protobuf 3.21.2 作为运行时
- 保持 monorepo 内 protobuf 运行时一致性，避免版本冲突
- google-protobuf 是 Google 官方维护的纯 JS 实现，稳定性高

**约束：**
- `gen-protocols.mjs` 是唯一使用 `protobufjs` 的位置（构建时解析 .proto 文件）
- 构建产物 `protocols.json` 包含所有运行时所需的元数据，运行时不接触 .proto 源文件

### 2. JSON 编辑器：CodeMirror 6（不使用 Monaco Editor）

**选择理由：**
| 维度 | CodeMirror 6 | Monaco Editor |
|------|-------------|---------------|
| 体积 | <250KB | ~3.3MB |
| 许可 | MIT | Custom（VS Code 衍生） |
| Tree-shaking | ✅ 完全支持 | ⚠️ 部分支持 |
| JSON Schema 校验 | ✅ @codemirror/lang-json | ✅ 内置 |
| 加载速度 | <100ms | ~500ms |

**约束：**
- 通过 `@codemirror/lang-json` + `@codemirror/lint` 提供语法高亮和校验
- 编辑器高度限制为 300px，避免大 JSON 导致性能问题

### 3. 元数据驱动架构

**设计原则：**
- **构建期**：`scripts/gen-protocols.mjs` 使用 protobufjs 解析 `.proto` → 生成 `protocols.json`
- **运行时**：应用仅读取 `protocols.json`，不接触任何 `.proto` 文件
- **安全性**：运行时零 protobuf 依赖，攻击面最小化

**数据流：**

```text
.proto 文件
    ↓ [构建期，protobufjs]
protocols.json（JSON 元数据）
    ↓ [运行时，google-protobuf]
MessagePacket 编解码
```

### 4. Token 100% 内存化

**决策：**
- Token 仅存储在 Zustand store 内存中
- 不写入 localStorage、sessionStorage、IndexedDB 或任何持久化存储
- 页面刷新后自动清空，需用户重新装载
- console.log 输出脱敏格式（前 8 位 + `***`）

**安全依据：**
- proto-tester 是内网工具，但仍遵循最小权限原则
- 防止 Token 通过浏览器 DevTools Storage 面板泄露
- 防止 Token 被 XSS payload 从持久化存储中窃取

### 5. CSP 硬编码到 index.html

**决策：**
- Content-Security-Policy 直接写在 `<meta>` 标签中
- 不依赖服务器端响应头（因为可能通过 file:// 或简单 HTTP 服务器访问）

**策略内容：**
```
connect-src: self localhost:8080 127.0.0.1:8080
script-src:  self 'unsafe-inline'
style-src:  self 'unsafe-inline'
object-src: none
form-action: none
frame-ancestors: none
```

### 6. pnpm workspace 统一管理

**决策：**
- proto-tester 作为 `typescript/proto-tester` 存在于 pnpm workspace 中
- 共享根目录的 `pnpm-lock.yaml`
- 与 typescript/web 共享 `google-protobuf` 等依赖（通过 workspace 协议）

## 后果

### 正面
- 与 typescript/web 共享同一 protobuf 运行时，无版本冲突风险
- CodeMirror 编辑器体积 <250KB，首屏加载快
- 运行时零 .proto 依赖，构建产物安全性高
- Token 内存化 + CSP 双重防护，满足内部安全审计要求
- 元数据驱动架构使协议扩展只需重新生成 protocols.json

### 负面
- `gen-protocols.mjs` 是唯一使用 protobufjs 的位置，增加了构建工具链复杂度
- Token 刷新后需手动重新装载，用户体验略差（可接受，因为是内网工具）
- CodeMirror 的 JSON Schema 校验能力弱于 Monaco，复杂嵌套场景需额外处理
- CLI 子命令（send/trace/run/capture）需要 Node.js 环境，不能在浏览器中使用

## 替代方案

### 方案 A：使用 protobuf.js 运行时解析 .proto
**否决原因：**
- protobuf.js 运行时与 google-protobuf 在同一进程中存在命名空间冲突
- 运行时动态解析 .proto 增加攻击面（.proto 文件注入风险）
- 体积更大（~200KB vs google-protobuf 已有依赖）

### 方案 B：使用 Monaco Editor
**否决原因：**
- 3.3MB 体积过大，影响加载性能
- VS Code 自定义许可证，商用需确认合规性
- 对于 JSON 编辑场景过度工程化

### 方案 C：Token 持久化到 localStorage
**否决原因：**
- 违反最小权限原则
- 可被 XSS payload 或浏览器扩展窃取
- 内网工具虽风险较低，但不应养成不良实践

## 相关文档
- PRD: [proto-tester 产品需求](../../prd/proto-tester.md)
- 技术栈版本锁定: `package.json`
- 安全规范: README.md § 安全规范

## CLI 模式决策（v2.0 新增）

### 为何 CLI 与 Web UI 同源不分裂
- 共享 lib/ 层代码避免双维护成本
- messagePacket / apiClient / protoMetadata 均为纯函数，天然可复用
- 同源性审计：CLI 不重新实现序列化逻辑

### 为何录屏证据本地化不接 OSS
- 沿用 admin E2E 简化决策
- 证据仅供本地审查，不上传至任何远程存储
- .gitignore 强制排除截图/视频文件

### 为何 Skill 限制 maxCasesPerSuite=100
- 资源保护：防止误触发大量请求对 Gateway 造成压力
- 当前 12 条协议远低于此上限，作为防火墙存在

### 为何 Playwright 作为 optionalDependencies
- send/run/trace 三个命令不需要 Chromium
- 仅 capture 命令需要浏览器环境
- 减少无 capture 用户的安装体积
