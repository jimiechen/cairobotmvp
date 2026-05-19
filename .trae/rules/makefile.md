# Makefile 工程入口规范

## 1. 架构决策

采用 **"根目录 Makefile 编排器 + 各语言子 Makefile + scripts 脚本"** 的三层结构。

**理由：**
- 根目录 Makefile 只做编排（16 个 target），不包含复杂逻辑
- 各语言子 Makefile 管理本语言构建细节
- scripts/ 处理需要编程逻辑的检查和生成任务
- CI 通过 `make ci` 一键触发全量检查

**历史：**
- 2026-05-19 前：根目录有 Makefile，按语言分 target
- 2026-05-19 中：短暂废弃根目录 Makefile，迁移到 scripts/
- 2026-05-19 后：恢复三层架构，根目录 Makefile 作为编排器

## 2. 目录结构

```
project-root/
├── Makefile              # 根目录编排器（16 个 target）
├── scripts/              # 跨语言自动化脚本
│   ├── ci/               # CI 检查脚本
│   ├── dev/              # 开发辅助脚本
│   ├── proto/            # Protobuf 生成脚本
│   ├── docs/             # 文档生成脚本
│   └── coverage/         # 覆盖率脚本
├── go/
│   └── Makefile          # Go 子模块
├── typescript/
│   └── Makefile          # TypeScript 子模块
└── python/
    └── Makefile          # Python 子模块
```

## 3. 根目录 Makefile Target 列表

| Target | 功能 | 说明 |
|---|---|---|
| `help` | 显示帮助信息 | `make help` |
| `bootstrap` | 初始化开发环境 | 检查工具链 + 安装依赖 |
| `proto` | 生成 Protobuf 代码 | 需要 protoc + 插件（本地开发用） |
| `proto-check` | 校验 Protobuf 代码 | 不需要 protoc（CI 用） |
| `lint` | 运行 Lint 检查 | 调用各语言子 Makefile |
| `test` | 运行所有测试 | 单元 + 集成 |
| `unit` | 运行单元测试 | 调用各语言子 Makefile |
| `integration` | 运行集成测试 | 调用各语言子 Makefile |
| `coverage` | 生成覆盖率报告 | 输出到 docs/reports/coverage/ |
| `build` | 构建可执行文件 | 调用各语言子 Makefile |
| `package` | 打包发布产物 | 调用各语言子 Makefile |
| `docs` | 检查文档完整性 | 调用 check_required_docs.py |
| `rules` | 执行工程规范检查 | 6 项检查串联 |
| `testcase-check` | 检查测试用例注册表 | 调用 check_testcase_registry.py |
| `comment-check` | 检查中文注释 | 调用 check_chinese_comments.py |
| `ci` | 完整 CI 检查 | 编排上述所有检查 |
| `clean` | 清理构建产物 | 调用各语言子 Makefile |

## 4. Proto 生成 vs 校验分离

**核心设计原则：CI 不安装 protoc，只做校验。**

| 阶段 | 命令 | 动作 | 需要 protoc？ |
|---|---|---|---|
| 本地开发 | `make proto` | 生成 Go/TS/Python/TarsGo 代码 | ✅ 需要 |
| CI 检查 | `make proto-check` | 校验生成文件存在性 + 协议编号一致性 | ❌ 不需要 |

### 4.1 proto-check 校验内容

1. **生成文件存在性**：检查 16 个预期文件是否都存在（Go 4 + TS 4 + Python 8）
2. **协议编号唯一性**：扫描 .proto 文件，检查 max+min 无重复
3. **协议编号登记**：检查所有编号已在 `docs/api/协议编号注册表.md` 登记

### 4.2 CI Workflow 精简

CI 只需安装运行时依赖：
- make, python3, bash（基础）
- Go 1.21（go test）
- Node.js 20（未来 TS 测试）
- Python 3.11（pytest + 脚本）

**不需要安装：**
- protobuf-compiler
- protoc-gen-go / protoc-gen-go-grpc
- protoc-gen-ts
- grpcio-tools

## 5. Scripts 统一入口

### 5.1 scripts/ci/

| 脚本 | 功能 | 调用方式 |
|---|---|---|
| check_required_docs.py | 检查关键文档存在性 | `make docs` |
| check_proto_registry.py | Protobuf 校验（文件存在性 + 编号） | `make proto-check` |
| check_reports.py | 检查测试报告 | CI job |
| check_directory_layout.py | 检查目录布局 | `make rules` |
| check_module_paths.py | 检查模块路径一致性 | `make rules` |
| check_chinese_comments.py | 检查中文注释 | `make rules` |
| check_testcase_registry.py | 检查测试用例注册表 | `make rules` |
| check_tools.sh | 检查工具链 | `make rules` / `make bootstrap` |
| check_make_targets.sh | 检查 Make target 完整性 | `make rules` |

### 5.2 scripts/proto/

| 脚本 | 功能 | 调用方式 |
|---|---|---|
| generate-go.sh | 生成 Go Protobuf 代码 | `make proto` |
| generate-ts.sh | 生成 TS Protobuf 代码 | `make proto` |
| generate-python.sh | 生成 Python Protobuf 代码 | `make proto` |
| generate-tarsgo.sh | 生成 TarsGo 代码 | `make proto` |
| check-generated-clean.sh | 检查生成代码是否已提交 | CI job |

### 5.3 scripts/coverage/

| 脚本 | 功能 | 调用方式 |
|---|---|---|
| go_coverage.sh | Go 覆盖率收集 | `make coverage` |
| ts_coverage.sh | TS 覆盖率收集 | `make coverage` |
| python_coverage.sh | Python 覆盖率收集 | `make coverage` |
| merge_coverage_reports.sh | 合并覆盖率摘要 | `make coverage` |

## 6. 子 Makefile 规范

各语言子 Makefile 只管理本语言构建，不跨语言调用。

### 6.1 go/Makefile

支持 target：
- `make install`：go mod download / go mod tidy
- `make unit`：遍历 go.work 中 module 执行 go test
- `make lint`：golangci-lint 或 vet
- `make build`：go build
- `make coverage`：go test -coverprofile
- `make clean`：清理产物

### 6.2 typescript/Makefile

支持 target：
- `make install`：pnpm install
- `make unit`：pnpm test
- `make lint`：eslint
- `make build`：tsc / webpack / vite
- `make clean`：清理 node_modules 和 dist

### 6.3 python/Makefile

支持 target：
- `make install`：pip install -r requirements.txt
- `make unit`：pytest
- `make lint`：ruff
- `make coverage`：pytest --cov
- `make clean`：清理 __pycache__ 和 dist

## 7. CI 集成

`.github/workflows/ci.yml` 通过 `make ci` 触发完整检查：

```yaml
- name: 执行完整 CI 检查
  run: make ci
```

CI 环境只需安装运行时依赖（见 §4.2），不需要 protoc。

## 8. 相关文档

- [ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md](../adr/ADR-0013-makefile-engineering-entrypoint-and-rule-enforcement.md)
- [ADR-0012-polyglot-monorepo-directory-layout.md](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- .trae/rules/commenting.md: 中文注释规范
