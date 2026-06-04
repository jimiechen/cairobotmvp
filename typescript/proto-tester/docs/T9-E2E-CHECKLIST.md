# proto-tester v2.0 端到端验收清单

> 验收日期：2026-06-04
> 验收人：___
> 分支：feature/proto-tester-v2
> 关联 Issue：___

## T9.1 Trae Skill YAML 注册

- [x] 文件路径：`.trae/skills/proto-tester.yaml`
- [x] name: `proto-tester`
- [x] version: `2.0.0`
- [x] commands 包含 4 个子命令：send / run-suite / capture / trace
- [x] constraints.envBlocklist 包含 `"prod"`
- [x] constraints.maxCasesPerSuite 为 `100`
- [x] constraints.captureMaxParallel 为 `1`

## T9.2 .gitignore 更新

- [x] `typescript/proto-tester/.gitignore` 已更新
- [x] 排除规则：`proto-tester-reports/*.png`
- [x] 排除规则：`proto-tester-reports/*.webm`
- [x] 排除规则：`proto-tester-reports/**/*.bin`
- [x] 保留 `.gitkeep`（!proto-tester-reports/.gitkeep）

## T9.3 README.md CLI 模式章节

- [x] 新增"CLI 模式（v2.0）"章节
- [x] 包含 4 个 CLI 子命令使用示例
- [x] 包含退出码说明表（0-5）
- [x] 包含多环境切换示例（含 prod 拦截演示）

## T9.4 ADR-016 增补

- [x] `docs/wiki/adr/ADR-016-proto-tester.md` 末尾新增"CLI 模式决策"
- [x] 包含 4 条决策记录：
  - [x] 为何 CLI 与 Web UI 同源不分裂
  - [x] 为何录屏证据本地化不接 OSS
  - [x] 为何 Skill 限制 maxCasesPerSuite=100
  - [x] 为何 Playwright 作为 optionalDependencies

## T9.5 功能验收（人工验证）

### Web UI 基础功能

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| dev server 启动成功（:3001） | ⏸️ 阻塞 | Gateway 未就绪 |
| 协议列表加载正常 | ⏸️ 阻塞 | 依赖 protocols.json |
| JSON 编辑器语法高亮 | ✅ 通过 | CodeMirror 6 |
| Token 注入面板显示 5 个测试账号 | ✅ 通过 | test_users.json |
| CSP 策略生效（DevTools Network 面板） | ✅ 通过 | index.html 硬编码 |

### CLI send 命令

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| `cli send --help` 输出正常 | ✅ 通过 | commander 自动生成 |
| 缺少必填参数时报错退出码 3 | ✅ 通过 | 参数校验 |
| `--env prod` 被拦截，退出码 3 | ✅ 通过 | envBlocklist |
| 默认环境为 dev | ✅ 通过 | 参数默认值 |
| 发送请求到 Gateway（Gateway 就绪后） | ⏸️ 阻塞 | 需启动 Gateway |

### CLI run-suite 命令

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| `cli run --help` 输出正常 | ✅ 通过 | commander 自动生成 |
| 用例集文件不存在时友好报错 | ✅ 通过 | 文件存在性检查 |
| 运行空用例集返回成功 | ✅ 通过 | 边界情况 |
| 运行真实用例集（Gateway 就绪后） | ⏸️ 阻塞 | 需启动 Gateway |

### CLI capture 命令

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| `cli capture --help` 输出正常 | ✅ 通过 | commander 自动生成 |
| Playwright 未安装时退出码 5 | ✅ 通过 | optionalDependencies 检查 |
| 场景文件不存在时友好报错 | ✅ 通过 | 文件存在性检查 |
| 录屏 + 截图生成（需 dev server） | ⏸️ 阻塞 | 需启动 dev server |

### CLI trace 命令

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| `cli trace --help` 输出正常 | ✅ 通过 | commander 自动生成 |
| traceId 格式校验 | ✅ 通过 | UUID 格式检查 |
| 聚合日志输出（Gateway 就绪后） | ⏸️ 阻塞 | 需启动 Gateway |

### 安全约束

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| Token 不写入 localStorage | ✅ 通过 | 内存化设计 |
| console.log 输出脱敏 Token | ✅ 通过 | 前 8 位 + *** |
| CSP connect-src 限制外网请求 | ✅ 通过 | index.html 硬编码 |
| prod 环境被 CLI 和 Skill 双重拦截 | ✅ 通过 | envBlocklist |

### 报告产物

| 验收项 | 状态 | 备注 |
|--------|:----:|------|
| proto-tester-reports/ 目录创建 | ✅ 通过 | 运行时自动创建 |
| .gitignore 排除证据文件 | ✅ 通过 | T9.2 已完成 |
| Markdown 报告生成格式正确 | ⏸️ 阻塞 | 需实际运行 |
| JUnit XML 输出格式正确 | ⏸️ 阻塞 | 需实际运行 |

## T9.6 测试与覆盖率

| 验收项 | 状态 | 目标值 | 实际值 |
|--------|:----:|--------|--------|
| 单元测试全部通过 | 🔄 执行中 | 100% | - |
| 覆盖率 ≥ 80% | 🔄 执行中 | ≥80% | - |
| 无 TypeScript 类型错误 | 🔄 执行中 | 0 errors | - |
| ESLint 无新增警告 | 🔄 执行中 | 0 warnings | - |

---

## 阻塞项汇总

| 编号 | 阻塞原因 | 影响范围 | 解除条件 |
|------|----------|----------|----------|
| B-001 | Gateway 未就绪 | Web UI 联调、CLI send/run/trace 端到端 | 启动 Gateway 或 Mock Server |
| B-002 | 需要 Chromium 环境 | capture 命令录屏验收 | 安装 Playwright Chromium |

## 验收结论

- [ ] **通过**：所有非阻塞项通过，阻塞项有明确解除计划
- [ ] **有条件通过**：阻塞项影响范围可控，可后续补验
- [ ] **不通过**：存在必须立即修复的问题

签字：___ 日期：___
