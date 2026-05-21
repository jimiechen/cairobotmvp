# Go Monorepo 模块化目录重构设计文档

**日期：** 2026-05-21
**状态：** 待审批
**关联 ADR：** [ADR-0012-polyglot-monorepo-directory-layout.md](../../adr/ADR-0012-polyglot-monorepo-directory-layout.md), [ADR-0014-message-packet-data-format-protobuf-bytes.md](../../adr/ADR-0014-message-packet-data-format-protobuf-bytes.md)
**决策人：** 项目主控

---

## 1. 背景与动机

### 1.1 当前问题

当前 Go monorepo 的目录结构存在**分层混乱**问题：

```
go/tars/system/
├── internal/service/      ← 旧版单体服务（应废弃）
├── localhandler/          ← Tars 调用适配层
└── modules/               ← 业务模块（嵌套在 Tars 层内）
    ├── hello/
    └── health/
```

**核心矛盾：**
- 业务模块 (`modules/hello`, `modules/health`) 被错误地放置在 `tars/system/` 内部
- 违反分层原则：业务逻辑不应依赖 Tars 框架层
- Gateway 层必须通过 `tars/system/modules/` 路径导入业务模块，造成概念混淆
- 缺少 `common-lib` 公共库层，错误码和类型定义散落各处

### 1.2 目标结构

用户指定的目标目录结构：

```
go/
├── common-lib/            # 公共库（错误码、类型、接口）
├── modules/               # 业务模块层（独立于框架）
│   ├── hello/
│   ├── health/
│   ├── users/
│   ├── auth/
│   ├── groups/
│   ├── topics/
│   └── readonly/
│       ├── users/
│       ├── groups/
│       └── topics/
├── gateway/
│   └── proto-gateway/
└── tars/
    └── system/            # 只放 Tars 调用层
```

### 1.3 设计原则

1. **分层清晰**：业务模块与 Tars 框架完全解耦
2. **强隔离**：每个业务模块独立 go.mod，可独立版本化
3. **向后兼容**：保留旧代码标记 @deprecated，给过渡期
4. **TDD 优先**：每步迁移必须先测试通过
5. **零回归**：全量测试套件必须在迁移后全部通过

---

## 2. 架构设计

### 2.1 分层架构图

```
┌─────────────────────────────────────────────────────┐
│                   Client Layer                      │
│              (HTTP / gRPC / Tars)                    │
└─────────────────────┬───────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│                Gateway Layer                        │
│         gateway/proto-gateway/                      │
│  ┌──────────────────────────────────────────┐       │
│  │ invoker.go: LocalInvoker / TarsGoInvoker │       │
│  │  → import: modules/hello, modules/health │       │
│  └──────────────────────────────────────────┘       │
└─────────────────────┬───────────────────────────────┘
                      │
          ┌───────────┴───────────┐
          ▼                       ▼
┌─────────────────┐    ┌─────────────────────┐
│  Modules Layer  │    │   Tars Adapter Layer│
│                 │    │                     │
│ ┌─────────────┐ │    │  tars/system/        │
│ │ hello       │ │    │  ┌───────────────┐  │
│ │ (独立go.mod)│ │    │  │ adapter/      │  │
│ ├─────────────┤ │    │  │ system_adapter│  │
│ │ health      │ │    │  └───────────────┘  │
│ │ (独立go.mod)│ │    │  ┌───────────────┐  │
│ ├─────────────┤ │    │  │ @deprecated:  │  │
│ │ future...   │ │    │  │ SystemService │  │
│ └─────────────┘ │    │  └───────────────┘  │
└────────┬────────┘    └─────────────────────┘
         │
         ▼
┌─────────────────┐
│  Common Lib     │
│                 │
│ ┌─────────────┐ │
│ │ codes.go    │ │  ← 错误码常量
│ │ types.go    │ │  ← 公共接口定义
│ └─────────────┘ │
└─────────────────┘
```

### 2.2 模块依赖方向

```
common-lib ← modules/* ← gateway/proto-gateway
                  ↑
            tars/system (adapter, 可选依赖)
```

**严格禁止的依赖：**
- ❌ modules → tars/system（业务模块不能依赖 Tars 层）
- ❌ gateway → tars/system/modules（已移除）
- ❌ common-lib → modules（公共库不能依赖业务模块）

### 2.3 go.mod 设计

#### 多 go.mod 策略（用户确认）

每个业务模块独立 go.module，实现强隔离：

| 模块 | go.mod path | 职责 |
|------|------------|------|
| common-lib | `github.com/jimiechen/mineplanet/go/common-lib` | 错误码、公共类型 |
| modules/hello | `github.com/jimiechen/mineplanet/go/modules/hello` | Hello 业务逻辑 |
| modules/health | `github.com/jimiechen/mineplanet/go/modules/health` | Health 业务逻辑 |
| gateway/proto-gateway | `github.com/jimiechen/mineplanet/go/gateway/proto-gateway` | HTTP 网关入口 |
| tars/system | `github.com/jimiechen/mineplanet/go/tars/system` | Tars 调用适配层 |

#### go.work 配置（变更后）

```go
go 1.23

use (
    ../proto/generated/go
    ./common-lib
    ./modules/hello
    ./modules/health
    ./gateway/proto-gateway
    ./tars/system
)
```

---

## 3. 详细设计

### 3.1 common-lib 模块

**路径：** `go/common-lib/`
**职责：** 定义跨模块共享的错误码和接口类型

#### 3.1.1 文件结构

```
go/common-lib/
├── go.mod
├── codes.go              # 统一错误码
├── types.go              # 公共接口定义
└── codes_test.go         # 错误码测试
```

#### 3.1.2 codes.go 设计

```go
package commonlib

// 统一业务错误码
const (
    CodeSuccess           = 10200 // 操作成功
    CodeBadRequest        = 10400 // 请求参数错误
    CodeUnauthorized      = 10401 // 未授权
    CodeNotFound          = 10404 // 资源未找到
    CodeInternalError     = 10500 // 内部错误
    CodeTarsNotImplemented = 10501 // Tars 远程调用未实现
)
```

#### 3.1.3 types.go 设计

```go
package commonlib

import "context"

// ModuleInvokeFunc 模块服务调用函数签名
// 业务模块统一使用 Protobuf bytes 作为输入输出
type ModuleInvokeFunc func(ctx context.Context, request []byte) ([]byte, error)

// ModuleHandler 模块处理器接口
// 所有业务模块必须实现此接口或通过 Adapter 适配
type ModuleHandler interface {
    Invoke(ctx context.Context, request []byte) ([]byte, error)
}
```

### 3.2 modules/hello 模块

**路径：** `go/modules/hello/`
**来源：** 从 `go/tars/system/modules/hello/` 迁移

#### 3.2.1 变更点

| 项目 | 变更前 | 变更后 |
|------|--------|--------|
| go.mod | 无（嵌入 tars/system） | `github.com/jimiechen/mineplanet/go/modules/hello` |
| import path | N/A | 直接引用 common-lib 和 protobuf |
| 代码逻辑 | 不变 | 不变 |

#### 3.2.2 go.mod 示例

```go
module github.com/jimiechen/mineplanet/go/modules/hello

go 1.23

require (
    github.com/jimiechen/mineplanet/go/common-lib v0.0.0
    google.golang.org/protobuf v1.36.11
)

replace github.com/jimiechen/mineplanet/protocols/generated/go/base => ../../../proto/generated/go/proto/base
replace github.com/jimiechen/mineplanet/go/common-lib => ../common-lib
```

### 3.3 modules/health 模块

**路径：** `go/modules/health/`
**来源：** 从 `go/tars/system/modules/health/` 迁移

设计与 hello 模块对称，不再赘述。

### 3.4 tars/system 精简

**路径：** `go/tars/system/`
**变更：** 从"业务+Tars混合层"精简为"纯 Tars 调用适配层"

#### 3.4.1 目录变更

| 操作 | 原路径 | 新路径/状态 |
|------|--------|------------|
| 重命名 | `localhandler/` | `adapter/` |
| 标记废弃 | `internal/service/system_service.go` | 保留 + `@deprecated` 注释 |
| 删除 | `modules/` | 已迁移至 `go/modules/` |

#### 3.4.2 adapter 设计

```go
package adapter

import (
    "context"
    "fmt"

    "github.com/jimiechen/mineplanet/go/common-lib"
    "github.com/jimiechen/mineplanet/go/modules/hello"
    "github.com/jimiechen/mineplanet/go/modules/health"
)

// SystemAdapter Tars 调用适配器
// 将 Tars servant 接口适配到模块化业务服务
type SystemAdapter struct {
    helloSvc  hello.HelloService
    healthSvc health.HealthService
}

func NewSystemAdapter() *SystemAdapter {
    return &SystemAdapter{
        helloSvc:  hello.NewService(),
        healthSvc: health.NewService(),
    }
}

// Invoke 执行 Tars 调用分发
func (a *SystemAdapter) Invoke(ctx context.Context, method string, request []byte) (int, []byte, error) {
    switch method {
    case "HealthCheck":
        resp, err := a.healthSvc.Check(ctx, request)
        if err != nil {
            return commonlib.CodeInternalError, nil, err
        }
        return commonlib.CodeSuccess, resp, nil

    case "HelloWorld":
        resp, err := a.helloSvc.SayHello(ctx, request)
        if err != nil {
            return commonlib.CodeInternalError, nil, err
        }
        return commonlib.CodeSuccess, resp, nil

    default:
        return commonlib.CodeNotFound, nil, fmt.Errorf("unknown method: %s", method)
    }
}
```

### 3.5 gateway/proto-gateway 更新

**路径：** `go/gateway/proto-gateway/internal/tarsclient/invoker.go`

#### 3.5.1 import 路径变更

```go
// 变更前
import (
    "github.com/jimiechen/mineplanet/go/tars/system/localhandler"
    "github.com/jimiechen/mineplanet/go/tars/system/modules/hello"
    "github.com/jimiechen/mineplanet/go/tars/system/modules/health"
)

// 变更后
import (
    "github.com/jimiechen/mineplanet/go/common-lib"
    "github.com/jimiechen/mineplanet/go/modules/hello"
    "github.com/jimiechen/mineplanet/go/modules/health"
)
```

---

## 4. 迁移计划

### 4.1 步骤总览（共 7 步）

| 步骤 | 操作 | 风险等级 | 预计影响文件数 |
|:---:|------|:-------:|:-----------:|
| 1 | 创建 common-lib | 🟢 低 | 3 (新增) |
| 2 | 创建 modules/hello 并迁移 | 🟡 中 | 4 (移动+修改) |
| 3 | 创建 modules/health 并迁移 | 🟡 中 | 4 (移动+修改) |
| 4 | 精简 tars/system | 🟡 中 | 5 (重命名+标记) |
| 5 | 更新 gateway import | 🟡 中 | 2 (修改) |
| 6 | 更新 go.work | 🟢 低 | 1 (修改) |
| 7 | 全量回归测试 | 🔴 高 | 验证所有 |

### 4.2 详细步骤

#### Step 1: 创建 common-lib（预计 10 分钟）

**操作清单：**
- [ ] 创建 `go/common-lib/` 目录
- [ ] 初始化 `go.mod`：`module github.com/jimiechen/mineplanet/go/common-lib`
- [ ] 创建 `codes.go`：定义错误码常量
- [ ] 创建 `types.go`：定义 ModuleInvokeFunc, ModuleHandler
- [ ] 创建 `codes_test.go`：验证错误码值
- [ ] 运行 `go build` 验证编译
- [ ] 运行 `go test` 验证测试

**验收标准：**
- ✅ `go build ./...` 通过
- ✅ `go test -v ./...` 通过
- ✅ 错误码值与现有代码一致 (10200/10404/10500)

---

#### Step 2: 迁移 modules/hello（预计 15 分钟）

**操作清单：**
- [ ] 创建 `go/modules/hello/` 目录
- [ ] 初始化 `go.mod`：`module github.com/jimiechen/mineplanet/go/modules/hello`
- [ ] 从 `go/tars/system/modules/hello/` 移动 `service.go`, `service_test.go`
- [ ] 更新 `go.mod`：添加 common-lib 依赖
- [ ] 更新 replace 指向 `../common-lib` 和 protobuf
- [ ] 可选：引入 commonlib.CodeSuccess 替代硬编码 10200
- [ ] 运行 `go test -v` 验证 3 个测试通过

**验收标准：**
- ✅ `go test -v ./...` 全部通过 (3/3)
- ✅ import 路径不包含 `tars/system`
- ✅ 可被 gateway 层正确导入

---

#### Step 3: 迁移 modules/health（预计 15 分钟）

**操作清单：**
- [ ] 创建 `go/modules/health/` 目录
- [ ] 初始化 `go.mod`：`module github.com/jimiechen/mineplanet/go/modules/health`
- [ ] 从 `go/tars/system/modules/health/` 移动文件
- [ ] 配置依赖和 replace
- [ ] 运行 `go test -v` 验证 4 个测试通过

**验收标准：**
- ✅ `go test -v ./...` 全部通过 (4/4)
- ✅ 与 hello 模块结构对称

---

#### Step 4: 精简 tars/system（预计 20 分钟）

**操作清单：**
- [ ] 重命名 `localhandler/` → `adapter/`
- [ ] 更新 `adapter/system_adapter.go`：
  - 修改 package 名为 `adapter`
  - 更新 import 路径指向 `modules/hello`, `modules/health`
  - 引入 `commonlib` 错误码
  - 重构为 `SystemAdapter` 结构体（替代旧 Handler）
- [ ] 标记 `internal/service/system_service.go` 为 `@deprecated`：
  - 添加注释说明废弃原因和替代方案
  - 保留代码不动，确保旧测试仍可运行
- [ ] 删除空目录 `modules/`（已迁移）
- [ ] 更新 `go.mod`：添加 common-lib, modules/hello, modules/health 依赖
- [ ] 运行完整测试验证

**验收标准：**
- ✅ `go test -v ./...` 通过（含旧 SystemService 测试）
- ✅ 新 adapter 测试通过
- ✅ `@deprecated` 标记清晰可见

---

#### Step 5: 更新 gateway/proto-gateway（预计 10 分钟）

**操作清单：**
- [ ] 修改 `internal/tarsclient/invoker.go`：
  - 更新 import：`tars/system/modules/*` → `modules/*`
  - 添加 `common-lib` import（可选）
  - 更新 `RegisterModuleHandlers()` 使用新路径
- [ ] 可选：更新 `RegisterSystemHandlers()` 使用新 adapter
- [ ] 运行 `go test -v ./...` 验证 Gateway 46 个测试通过

**验收标准：**
- ✅ Gateway 46/46 测试通过
- ✅ 零编译错误
- ✅ import 路径不含 `tars/system/modules`

---

#### Step 6: 更新 go.work（预计 5 分钟）

**操作清单：**
- [ ] 编辑 `go/go.work`：
  - 添加 `./common-lib`
  - 添加 `./modules/hello`
  - 添加 `./modules/health`
  - 保持现有条目不变
- [ ] 运行 `go work sync` 验证

**验收标准：**
- ✅ `go work sync` 无错误
- ✅ 所有模块可互相解析

---

#### Step 7: 全量回归测试（预计 15 分钟）

**操作清单：**
- [ ] 运行 `cd go/modules/hello && go test -v -count=1 ./...`
- [ ] 运行 `cd go/modules/health && go test -v -count=1 ./...`
- [ ] 运行 `cd go/tars/system && go test -v -count=1 ./...`
- [ ] 运行 `cd go/gateway/proto-gateway && go test -v -count=1 ./...`
- [ ] 运行 `cd go/common-lib && go test -v -count=1 ./...`
- [ ] 统计总测试数并确认零回归
- [ ] 更新 `docs/testing/测试用例注册表.md`

**验收标准：**
- ✅ 所有测试通过（预计 60+ 子测试）
- ✅ 零回归（原 60 个测试全部 PASS）
- ✅ 测试注册表同步更新

---

## 5. 风险评估

### 5.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|:----:|:----:|----------|
| go.mod 循环依赖 | 中 | 高 | 严格单向依赖：common-lib ← modules ← gateway |
| replace 路径错误 | 中 | 中 | 使用相对路径，逐步验证 |
| 旧测试 import 失效 | 低 | 中 | 保留 @deprecated 代码，延迟清理 |
| IDE 缓存未刷新 | 高 | 低 | 重启 IDE，执行 `go clean -cache` |

### 5.2 业务风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|:----:|:----:|----------|
| 迁移期间功能回归 | 低 | 高 | 每步必跑全量测试 |
| 团队理解成本 | 中 | 低 | 本文档 + 代码注释 |
| 未来模块模板不一致 | 中 | 中 | 建立 modules 模板脚手架 |

---

## 6. 成功标准

### 6.1 必须达成（P0）

- [ ] 目录结构符合用户指定布局
- [ ] 每个 modules/* 独立 go.mod
- [ ] common-lib 存在且可被所有模块引用
- [ ] tars/system 精简为纯 adapter 层
- [ ] 全量测试通过（60+ 子测试），零回归
- [ ] go.work 正确配置

### 6.2 应当达成（P1）

- [ ] 旧 SystemService 标记 @deprecated 且有替代说明
- [ ] 代码注释清晰说明各层职责边界
- [ ] 测试注册表同步更新
- [ ] 无 TODO/FIXME 遗留（除明确的未来规划）

### 6.3 可以达成（P2）

- [ ] 建立 modules 模板脚手架（供未来 users/auth 等模块复用）
- [ ] 补充 common-lib 的基准测试
- [ ] 编写迁移回滚脚本

---

## 7. 后续规划

### 7.1 本次范围

仅迁移 **hello** 和 **health** 两个已完成模块。

### 7.2 未来扩展

按需创建以下模块（遵循相同模式）：

| 模块 | 优先级 | 触发条件 |
|------|:-----:|---------|
| modules/users | P1 | 用户中台需求启动 |
| modules/auth | P1 | 认证需求启动 |
| modules/groups | P2 | 群组功能需求 |
| modules/topics | P2 | 话题功能需求 |
| modules/readonly/* | P2 | 只读查询需求 |

---

## 8. 审批记录

| 日期 | 审批人 | 结论 | 备注 |
|------|--------|:----:|------|
| 2026-05-21 | 待审批 | ⏳ 待定 | 初稿完成 |

---

**文档版本：** v1.0
**最后更新：** 2026-05-21
**作者：** Trae AI Assistant
