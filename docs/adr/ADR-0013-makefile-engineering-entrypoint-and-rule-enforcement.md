# ADR-0013: Makefile 工程入口与规则强制执行

## 状态

已接受

## 背景

CaiRobot MVP 采用多语言 Monorepo 架构（Go + Python + TypeScript），需要统一的工程入口来协调：
1. 多语言构建、测试、Lint
2. Protobuf 代码生成（Go/TS/Python/TarsGo）
3. 测试用例管理
4. 覆盖率收集
5. 工程规范检查（中文注释、目录布局等）

此前删除根目录 Makefile 是为了避免"大而全 Makefile"，但导致开发者缺少统一入口。

## 决策

采用**三层结构**：

```
根目录 Makefile（总控编排）
    ├── go/Makefile（Go 语言细节）
    ├── typescript/Makefile（TypeScript 语言细节）
    ├── python/Makefile（Python 语言细节）
    └── scripts/（承载复杂逻辑）
        ├── ci/*（工程规范检查脚本）
        ├── proto/*（Protobuf 生成脚本）
        └── coverage/*（覆盖率收集脚本）
```

## 详细设计

### 1. 根目录 Makefile

只做工程级任务编排，不写复杂语言细节。提供以下 target：

| Target | 用途 |
|---|---|
| help | 显示帮助 |
| bootstrap | 初始化开发环境 |
| proto | 生成所有语言 Protobuf 代码 |
| lint | 运行 Lint |
| test | 全部测试 |
| unit | 单元测试 |
| integration | 集成测试 |
| coverage | 覆盖率报告 |
| build | 构建 |
| package | 打包 |
| docs | 文档检查 |
| rules | 规范检查 |
| testcase-check | 测试用例检查 |
| comment-check | 中文注释检查 |
| ci | 完整 CI |
| clean | 清理 |

### 2. 子 Makefile

每个语言目录有独立 Makefile，处理该语言的特定逻辑。

### 3. 脚本层

复杂逻辑由 shell/python 脚本承载，Makefile 只做调用。

### 4. 规则强制执行

通过 `make rules` 和 `make ci` 执行以下检查：
- 目录布局一致性（check_directory_layout.py）
- Make target 完整性（check_make_targets.sh）
- 模块路径一致性（check_module_paths.py）
- 中文注释规范性（check_chinese_comments.py）
- 测试用例注册表一致性（check_testcase_registry.py）

### 5. 测试用例管理

建立测试用例 ID 规范和注册表，确保测试文件可追溯。

### 6. 覆盖率管理

统一收集各语言覆盖率，输出到 `docs/reports/coverage/`。

## 理由

1. **可维护性**：三层结构避免单点臃肿
2. **可扩展性**：新增语言只需新增子目录和子 Makefile
3. **本地等价性**：`make ci` 可在本地复现完整 CI 流程
4. **规则落地**：通过脚本将规范从文档变为可执行的检查

## 后果

### 正面
- 开发者有统一的工程入口
- CI 通过 make 命令编排，逻辑清晰
- 规范检查自动化，减少人工遗漏

### 负面
- 需要维护多个文件（Makefile × 4 + 脚本 × 12+）
- 新成员需要了解三层结构

## 替代方案

1. **单一 Makefile**：被否决，会导致维护噩梦
2. **纯 Taskfile/Just**：被否决，团队更熟悉 Make
3. **npm scripts only**：被否决，不支持 Go/Python

## 约束

- 根目录 Makefile 单个 recipe 不超过 10 行
- 所有脚本必须有错误处理（set -euo pipefail）
- CI 必须优先调用 make 命令

## 相关文档

- .trae/rules/makefile.md
- .trae/rules/commenting.md
- .trae/rules/monorepo.md
- docs/testing/测试用例注册表.md
- ADR-0012: 多语言 Monorepo 目录布局

## 后续待确认事项

- 是否需要在 CI 中集成覆盖率阈值门禁？
- 是否需要自动创建 Issue 当测试未登记？
