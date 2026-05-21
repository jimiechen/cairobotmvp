# Trae Task Raw Archive

## 元信息

| 字段 | 值 |
|---|---|
| **任务日期** | 2026-05-20 |
| **任务类型** | 架构口径修正 + TarsGo HTTP 模块改造 |
| **执行者** | Trae |
| **分支** | main |
| **关联 PRD** | PRD-00 (MVP 总纲) |
| **关联 ADR** | ADR-0008-use-tarscloud-routing-layer.md |
| **源 ID** | trae-20260520-tarsgo-http-module-refactor |

## 主控裁决摘要

### 核心修正点

1. **local 不是"不使用 Tars"**：local/单体模式是 TarsGo 单体部署模式，使用 TarsGo 框架运行（TarsHttpMux / AddHttpServant / Run），只是不连接远程 TarsCloud 注册中心
2. **LocalInvoker 重新定义**：从"本地调用实现"修正为"单体部署模式下的本进程 TarsGo servant adapter"，不绕过 Tars 框架
3. **TarsGoInvoker 重新定义**：从"TarsGo 调用实现（未实现）"修正为"微服务部署模式下的远程 TarsGo client invoker"
4. **proto-gateway 身份**：从"单网关入口实现"修正为"TarsGo HTTP Servant"

## 执行内容

### 1. go.mod 修改
- 新增 `github.com/TarsCloud/TarsGo v1.4.6` 依赖
- 新增 replace 指向 `../../third_party/TarsGo/TarsGo-1.4.6`

### 2. main.go 改造
- 移除 `context`、`os/signal`、`syscall` 等原始 HTTP server 启动逻辑
- 引入 `github.com/TarsCloud/TarsGo/tars`
- 使用 `tars.TarsHttpMux` 替代原始 `http.ServeMux`
- 使用 `tars.AddHttpServant(mux, objName)` 注册 HTTP servant
- 使用 `tars.Run()` 启动 TarsGo 应用（替代原始 `srv.ListenAndServe()`）
- 通过 `TARS_CONFIG` 环境变量或默认路径加载 Tars 配置

### 3. 新增文件
- `go/gateway/proto-gateway/configs/gateway/gateway.local.conf` - TarsGo 单体部署本地配置（locator 为空）

### 4. invoker.go 注释修正
- LocalInvoker 注释：明确为"本进程 TarsGo servant adapter"，强调不绕过 Tars 框架
- TarsGoInvoker 注释：明确为"远程 TarsGo client invoker"，S1 未实现
- RegisterSystemHandlers 注释：更新为"本地 TarsGo servant handler"

### 5. README.md 重写
- 职责描述：新增"TarsGo HTTP Servant"身份
- 新增"技术基线"章节（TarsCloud/TarsGo v1.4.6）
- 运行模式重写：从"运行模式"改为"部署拓扑"
- 单体部署模式：强调使用 TarsGo 框架运行
- 微服务部署模式：强调相同技术基线 + 远程 TarsGo client
- 新增"两种模式的共同点"章节
- 目录结构更新：标注各文件的 TarsGo 相关职责

### 6. CODE-WIKI.md 更新
- §16 Gateway 运行模式：全面重写为部署拓扑描述
- §17 → §16.3 TarsInvoker 接口：LocalInvoker/TarsGoInvoker 新定义
- §13.2 Go 目录结构：新增 gateway.local.conf、third_party/TarsGo 条目
- 变更日志：新增 2026-05-20 条目

### 7. tars规范.md 更新
- §16 Gateway 运行模式：全面重写
- §16.3 TarsInvoker 接口：新定义 + 共同契约说明
- §15.2 单体模式流程：标题和描述修正
- §18.1 测试要求：TarsGoInvoker 描述更新

### 8. 测试用例注册表更新
- TC-GW-0005 备注更新：明确 TarsGoInvoker 返回 10500
- 关键验证点更新：TarsGoInvoker 未实现说明补充 S1 阶段

## 变更文件清单

| 文件路径 | 变更类型 | 说明 |
|---|---|---|
| `go/gateway/proto-gateway/go.mod` | 修改 | 新增 TarsCloud/TarsGo v1.4.6 依赖 + replace |
| `go/gateway/proto-gateway/cmd/server/main.go` | 修改 | 改造为 TarsGo HTTP 入口（TarsHttpMux/AddHttpServant/Run） |
| `go/gateway/proto-gateway/configs/gateway/gateway.local.conf` | **新增** | TarsGo 单体部署本地配置 |
| `go/gateway/proto-gateway/internal/tarsclient/invoker.go` | 修改 | LocalInvoker/TarsGoInvoker 注释口径修正 |
| `go/gateway/proto-gateway/README.md` | 修改 | 全面重写架构描述 |
| `docs/wiki/CODE-WIKI.md` | 修改 | §16 运行模式 + §13.2 目录结构 + 变更日志 |
| `docs/api/tars规范.md` | 修改 | §16 + §15.2 + §18.1 口径修正 |
| `docs/testing/测试用例注册表.md` | 修改 | TC-GW-0005 + 验证点更新 |

## 验收命令结果

```bash
pwd:           /Volumes/MacintoshHD/Users/jemoo/git/traeMVP/cairobotmvp
branch:        main
git status:    7 files modified, 2 untracked (configs/gateway/, third_party/)
git diff --stat: 7 files, +137 -61
```

**注意**: 当前环境未安装 Go 工具链，以下命令无法执行：
- `cd go/gateway/proto-gateway && go list -m all \| grep tars`
- `cd go/gateway/proto-gateway && go test ./... -v -count=1`
- `cd go/tars/system && go test ./... -v -count=1`
- `make -C go unit`

需要在有 Go 工具链的环境下补充验证。

## 主控约束遵守情况

| # | 要求 | 状态 |
|---:|---|---:|
| 1 | local 模式仍使用 TarsGo | ✅ main.go 使用 TarsHttpMux/AddHttpServant/Run |
| 2 | local 是 TarsGo 单体部署模式 | ✅ README/CODE-WIKI/tars规范 全部修正 |
| 3 | microservice 是 TarsGo 微服务部署模式 | ✅ 文档已同步 |
| 4 | 两种模式都遵守 Tars bytes 契约 | ✅ 共同点章节已添加 |
| 5 | 区别仅是部署拓扑 | ✅ 明确写入驻同进程 vs 跨进程 |
| 6 | LocalInvoker ≠ 绕过 Tars 的普通 Go 调用 | ✅ 注释全面修正 |
| 7 | LocalInvoker = 本进程 TarsGo adapter | ✅ invoker.go 注释已更新 |
| 8 | TarsGoInvoker = 远程 TarsGo client | ✅ invoker.go 注释已更新 |
| 9 | 不实现真实远程 TarsGoInvoker | ✅ 仅修正文档和注释，未实现调用 |
| 10 | 不新增 JSON payload 路线 | ✅ 无变更 |
| 11 | 不新增 REST 业务 path | ✅ 仅 /api/hello |
| 12 | 不扩大到 users/auth/groups/topics | ✅ 无变更 |
| 13 | 不 git push | ✅ 仅本地操作 |
| 14 | 不把依赖放到 docs/wiki/raw | ✅ 放在 go/third_party/ |
| 15 | 不把依赖放到 go/ 根目录 | ✅ 放在 go/third_party/TarsGo/ |

## 风险与遗留事项

1. **Go 工具链缺失**：当前环境无 `go` 命令，无法运行 `go test` 和 `go list` 验收命令。需在有 Go 的环境补充验证。
2. **TarsGo 配置文件格式**：gateway.local.conf 使用 TarsGo 原生 XML 格式，需确认与项目其他配置风格一致。
3. **go.sum 未生成**：因缺少 Go 工具链，`go mod tidy` 未执行，go.sum 可能需要更新。

## 后续动作

- [ ] 在有 Go 工具链的环境运行 `cd go/gateway/proto-gateway && go mod tidy` 生成 go.sum
- [ ] 运行 `cd go/gateway/proto-gateway && go test ./... -v -count=1` 确认测试通过
- [ ] 运行 `cd go/tars/system && go test ./... -v -count=1` 确认 System 模块测试不受影响
- [ ] 确认后提交 PR
