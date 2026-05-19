# ADR-0012: 多语言 Monorepo 目录布局

## 状态

已接受

## 背景

CaiRobot MVP 是一个多语言技术栈项目，包含：
- Go（后端服务、网关、TarsGo 服务）
- Python（AI 服务）
- TypeScript/React（前端应用）
- Protobuf（跨语言协议契约）

当前目录结构存在以下问题：
1. 根目录放置 `go.work`，将 Go 的模块管理逻辑暴露到项目根，与 Python/TypeScript 目录平级，边界模糊
2. 根目录放置 `Makefile`，试图用单一构建工具管理多语言项目，导致职责混乱
3. `gateway/`、`tars/`、`services/`、`ai/`、`web/` 等目录在根目录平铺，缺乏按语言分层的聚合
4. CI workflow 路径引用分散，难以维护

## 决策

采用按语言分层的目录布局，根目录只保留工程级公共资产。

### 布局原则

1. **根目录只保留公共资产**：AGENTS.md、README.md、.github/、.trae/、docs/、proto/、configs/、scripts/、deploy/、tests/
2. **按语言聚合**：Go 进 `go/`，Python 进 `python/`，TypeScript 进 `typescript/`
3. **跨语言契约独立**：`proto/` 保持在根目录，不归属于任何单一语言
4. **去中心化构建**：删除根目录 Makefile，各语言使用自己的构建工具（go.mod、pyproject.toml、package.json）
5. **脚本统一入口**：自动化逻辑统一放入 `scripts/ci/`、`scripts/dev/`、`scripts/proto/`、`scripts/docs/`

### 禁止事项

| 禁止项 | 说明 | 替代方案 |
|---|---|---|
| 根目录放 `go.work` | Go 模块管理逻辑不应暴露到项目根 | 放入 `go/go.work` |
| 根目录放 `Makefile` | 不应试图用单一构建工具管理多语言项目 | 各语言使用自己的构建工具；自动化逻辑放入 `scripts/` |
| 根目录新增语言目录 | 新增语言应在根目录新增一级语言目录 | 如新增 Rust，创建 `rust/` |
| 语言目录交叉引用 | Go 代码不应直接引用 `python/` 下的文件 | 通过 `proto/` 或 `scripts/` 协调 |

### 目录结构

```text
根目录/
├── AGENTS.md
├── README.md
├── .gitignore
├── .github/
│   ├── workflows/
│   │   ├── ci.yml
│   │   └── daily-knowledge-distillation.yml
│   ├── ISSUE_TEMPLATE/
│   └── pull_request_template.md
├── .trae/
│   ├── commands/
│   ├── rules/
│   └── skills/
├── docs/
│   ├── adr/
│   ├── api/
│   ├── prd/
│   ├── testing/
│   ├── wiki/
│   └── reports/
├── proto/
│   └── base/
├── configs/
├── scripts/
│   ├── ci/
│   │   ├── check_required_docs.py
│   │   ├── check_proto_registry.py
│   │   ├── check_reports.py
│   │   └── check_go_modules.sh
│   ├── dev/
│   │   └── go-test.sh
│   ├── proto/
│   │   └── generate-go.sh
│   └── docs/
├── deploy/
├── tests/
├── go/                              # Go 语言资产
│   ├── go.work
│   ├── README.md
│   ├── gateway/
│   │   └── proto-gateway/
│   │       ├── go.mod
│   │       ├── cmd/
│   │       ├── internal/
│   │       └── configs/
│   ├── tars/
│   │   ├── system/
│   │   │   ├── go.mod
│   │   │   ├── cmd/
│   │   │   ├── internal/
│   │   │   └── localhandler/
│   │   ├── auth/
│   │   ├── audit/
│   │   └── ...
│   ├── shared/
│   │   ├── audit/
│   │   ├── config/
│   │   ├── result/
│   │   └── protoadapter/
│   └── third_party/
│       └── TarsGo/
├── python/                          # Python 语言资产
│   ├── ai/
│   │   ├── service/
│   │   └── README.md
│   └── tools/
│       └── README.md
└── typescript/                      # TypeScript 语言资产
    ├── web/
    │   ├── src/
    │   └── package.json
    ├── admin-web/
    ├── app-h5/
    ├── packages/
    └── README.md
```

### Module Path 规范

所有 Go module 使用统一 path：

```text
github.com/jimiechen/mineplanet/go/...
```

示例：
- `github.com/jimiechen/mineplanet/go/gateway/proto-gateway`
- `github.com/jimiechen/mineplanet/go/tars/system`

### Go Workspace

`go.work` 从根目录移动到 `go/go.work`，只管理 Go 子模块：

```go
go 1.21

use (
    ./gateway/proto-gateway
    ./tars/system
)
```

## 后果

### 正面

1. 语言边界清晰，开发者可以快速定位自己关心的代码
2. 根目录整洁，工程级公共资产一目了然
3. CI 路径清晰，各语言 job 使用独立的 working-directory
4. 去除了根目录 Makefile 的维护负担
5. 新增语言栈时，只需在根目录新增一个语言目录，不影响现有结构

### 负面

1. Go import 路径需要批量更新（从 `github.com/jimiechen/mineplanet/gateway/...` 改为 `github.com/jimiechen/mineplanet/go/gateway/...`）
2. 文档中的路径引用需要批量更新
3. 开发者需要适应 `cd go/` 后再执行 Go 命令

## 替代方案

### 方案 B：保持现状

保持根目录 `go.work` 和 `Makefile`，只新增 `python/` 和 `typescript/` 目录。

- 优点：改动最小
- 缺点：Go 的模块管理逻辑继续污染根目录，长期维护困难

### 方案 C：按功能域分层

不按语言分层，而是按功能域分层（`gateway/`、`ai/`、`web/` 等都在根目录平铺）。

- 优点：功能聚合度高
- 缺点：多语言项目按功能域分层会导致每个目录内部语言混杂，构建工具难以统一

## 相关文档

- [ADR-0001-总体系统架构.md](ADR-0001-总体系统架构.md)
- [docs/wiki/CODE-WIKI.md](../wiki/CODE-WIKI.md)
- [docs/wiki/LLM-WIKI.md](../wiki/LLM-WIKI.md)
