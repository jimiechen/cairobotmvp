# Protobuf Go 生成代码

本目录是 **单一 Go module**，统一管理所有 Protobuf 生成的 Go 代码。

## 目录结构

```
proto/generated/go/
├── go.mod              ← 唯一的 module 声明（禁止在子目录创建 go.mod）
├── go.sum              ← 依赖校验和
├── base/               ← package base（基础协议）
│   ├── health.pb.go
│   ├── hello.pb.go
│   ├── message.pb.go
│   └── result.pb.go
├── ai/                 ← package ai（AI 协议，MVP2 阶段）
└── tars/               ← package tars（TARS 协议，MVP2 阶段）
```

## 约束规则

### 1. 单一 Module 原则

| 规则 | 说明 | 违规后果 |
|---|---|---|
| **go.mod 只能有一个** | 位于 `proto/generated/go/go.mod` | CI 报错、依赖解析混乱 |
| **禁止子目录 go.mod** | `base/`、`ai/`、`tars/` 下不允许有 go.mod | go.work 路径冲突 |
| **go.work 引用顶层** | `go/go.work` 中写 `../proto/generated/go` | 否则无法传递依赖 |

### 2. Import 路径规范

module 声明为 `github.com/jimiechen/mineplanet/protocols/generated/go`，子包 import 路径：

| 子包 | Import Path | 示例 |
|---|---|---|
| base | `{module}/base` | `pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"` |
| ai | `{module}/ai` | `pb "github.com/jimiechen/mineplanet/protocols/generated/go/ai"` |
| tars | `{module}/tars` | `pb "github.com/jimiechen/mineplanet/protocols/generated/go/tars"` |

### 3. 新增 Proto 包流程

1. 在 `proto/` 目录下新建 `.proto` 文件（如 `proto/ai/chat.proto`）
2. 执行 `make proto` 生成代码
3. 生成脚本自动输出到 `proto/generated/go/ai/*.pb.go`
4. 在消费方使用对应 import path
5. 提交 `proto/generated/go/go.sum` 更新

**不需要**：新建 go.mod、修改 go.work、修改 generate-go.sh。

### 4. 生成配置

- **脚本**：`scripts/proto/generate-go.sh`
- **Module 前缀**：`GO_MODULE_PREFIX="github.com/jimiechen/mineplanet/protocols/generated/go"`
- **路径模式**：`paths=source_relative`（保持 proto 目录结构与 Go 包一致）
- **protoc-gen-go 版本**：v1.36.11（与 go.mod require 一致）

### 5. CI 校验

CI 通过 `make proto-check` 校验（不安装 protoc）：

1. **文件存在性**：检查所有预期 `.pb.go` 文件存在
2. **协议编号唯一性**：检查 max+min 无重复
3. **注册表一致性**：检查编号已登记到 `docs/api/协议编号注册表.md`

### 6. 与其他语言的关系

```
proto/
├── base/              ← .proto 源文件（各语言共用）
└── generated/
    ├── go/            ← 本目录（Go module）
    │   ├── base/      ← 对应 proto/base/*.proto
    │   └── ai/        ← 对应 proto/ai/*.proto（未来）
    ├── ts/            ← TypeScript（独立目录）
    └── python/        ← Python（独立目录）
```

## 依赖关系

```
gateway / tars (消费者)
  └─ import "{module}/base"
       │
       ▼ go.work use ../proto/generated/go
  proto/generated/go/ (本 module)
  ├── go.mod: require google.golang.org/protobuf v1.36.11
  ├── base/package: *.pb.go (protoc 自动生成)
  └── go.sum: 依赖锁定
```

## 注意事项

- **不要手动编辑 `.pb.go` 文件**：它们由 protoc-gen-go 自动生成
- **不要在子目录执行 `go mod init`**：会破坏单一 module 结构
- **`make proto` 后必须提交 `go.sum`**：依赖可能变化
- **新增 .proto 文件后重新 `make proto`**：确保生成代码同步
