# Makefile 工程入口规范

## 1. 架构决策

**已废弃根目录 Makefile**，采用"scripts 统一入口 + 各语言子 Makefile"的两层结构。

**理由：**
- 多语言 monorepo 不应在根目录维护单一构建工具
- 根目录 Makefile 会随语言增多而膨胀，维护困难
- scripts 更灵活，支持复杂逻辑和跨语言编排
- 符合 ADR-0012 多语言 monorepo 目录布局决策

**历史：**
- 2026-05-19 前：根目录有 Makefile，按语言分 target
- 2026-05-19 后：删除根目录 Makefile，功能迁移到 `scripts/`

## 2. 目录结构

```
project-root/
├── scripts/              # 跨语言自动化入口
│   ├── ci/               # CI 检查脚本
│   ├── dev/              # 开发辅助脚本
│   ├── proto/            # Protobuf 生成脚本
│   ├── docs/             # 文档生成脚本
│   └── coverage/         # 覆盖率脚本
├── go/
│   └── Makefile          # Go 子模块（可选）
├── typescript/
│   └── Makefile          # TypeScript 子模块（可选）
└── python/
    └── Makefile          # Python 子模块（可选）
```

## 3. 根目录禁止事项

| 禁止项 | 说明 |
|---|---|
| 根目录 Makefile | 已由 ADR-0012 明确禁止 |
| 根目录 `make` 调用 | CI workflow 应直接调用 `scripts/`，不依赖 make |

## 4. Scripts 统一入口

### 4.1 scripts/ci/

| 脚本 | 功能 |
|---|---|
| check_required_docs.py | 检查关键文档存在性 |
| check_proto_registry.py | 检查协议编号注册表 |
| check_reports.py | 检查测试报告 |
| check_go_modules.sh | 检查 Go 模块 |
| check_directory_layout.py | 检查目录布局 |
| check_module_paths.py | 检查模块路径一致性 |
| check_chinese_comments.py | 检查中文注释 |
| check_testcase_registry.py | 检查测试用例注册表 |

### 4.2 scripts/dev/

| 脚本 | 功能 |
|---|---|
| go-test.sh | 运行所有 Go 模块测试 |

### 4.3 scripts/proto/

| 脚本 | 功能 |
|---|---|
| generate-go.sh | 生成 Go Protobuf 代码 |
| generate-ts.sh | 生成 TypeScript Protobuf 代码 |
| generate-python.sh | 生成 Python Protobuf 代码 |
| generate-tarsgo.sh | 生成 TarsGo 代码 |

### 4.4 scripts/coverage/

| 脚本 | 功能 |
|---|---|
| go_coverage.sh | Go 覆盖率收集 |
| ts_coverage.sh | TypeScript 覆盖率收集 |
| python_coverage.sh | Python 覆盖率收集 |
| merge_coverage_reports.sh | 合并覆盖率摘要 |

## 5. 子 Makefile 规范（可选）

各语言目录可保留自己的 Makefile，但：
- 只管理本语言构建
- 不跨语言调用
- 根目录不调用子 Makefile

### 5.1 go/Makefile（可选）

如需创建，应支持：
- `make proto`：调用 scripts/proto/generate-go.sh
- `make test`：遍历 go.work 中所有 module 执行 `go test ./...`
- `make coverage`：生成 HTML 覆盖率报告

## 6. CI 集成

`.github/workflows/ci.yml` 直接调用 scripts：

```yaml
- name: Run Go tests
  run: bash scripts/dev/go-test.sh
```

## 7. 相关文档

- [ADR-0012-polyglot-monorepo-directory-layout.md](../adr/ADR-0012-polyglot-monorepo-directory-layout.md)
- .trae/rules/commenting.md: 中文注释规范
