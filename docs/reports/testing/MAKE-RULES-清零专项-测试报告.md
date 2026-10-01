# MAKE-RULES 清零专项测试报告

## 1. 基本信息

| 字段 | 值 |
|---|---|
| 测试报告 ID | REPORT-RULES-001 |
| 测试日期 | 2026-10-01 |
| 测试人 | Trae |
| 当前分支 | feature/fa-orchestrator-skeleton |
| 相关 Issue | make rules 308 项历史遗留失败清零（用户专项确认） |
| 相关 PRD | 无（工程规范专项，不涉及 PRD 范围） |
| 相关 ADR | ADR-0013（Makefile 工程入口与规范强制执行） |
| 相关测试用例 | TC-CI-0001~0005（六项检查）、TC-GW/TC-TARS/TC-WEB 批量补登记 141 项 |

## 2. 测试环境

| 字段 | 值 |
|---|---|
| 操作系统 | macOS（darwin/amd64） |
| 运行环境 | Go 1.25.5 / Python 3.11 / Node 22.20 |
| 依赖版本 | protoc 34.1 / protoc-gen-go 1.36.11 |
| 执行命令 | `make rules` / `make proto-check` / `make unit` |

## 3. 测试对象

验证 `make rules` 六项工程规范检查从「308+ 项历史遗留失败」清零：

1. check_tools.sh 工具链检查
2. check_directory_layout.py 目录布局检查
3. check_make_targets.sh Make target 完整性检查
4. check_module_paths.py 模块路径一致性检查
5. check_chinese_comments.py 中文注释规范性检查
6. check_testcase_registry.py 测试用例注册表一致性检查

## 4. 失败归因与修复清单

### 4.1 检查器缺陷修复（scripts/ci/，共 5 处）

| 文件 | 缺陷 | 修复 | 消除误报数 |
|---|---|---|---|
| check_testcase_registry.py | rglob 未排除 node_modules，第三方依赖自带测试文件被要求登记 | 新增 EXCLUDED_DIR_PARTS / EXCLUDED_PATH_PREFIXES 排除规则 | 118 |
| check_testcase_registry.py | 未排除 vendored 第三方（TarsGo 1.4.6 / go-admin 框架） | 同上（go/third_party/TarsGo/ 与 go/admin/ 前缀排除） | 25 |
| check_testcase_registry.py | 表头过滤条件 `"前缀" in line` 误伤测试函数名含「前缀」的行（如 TestMockCache_Invalidate_按前缀清除） | 收紧为 `line.startswith("| 前缀")` | 1 |
| check_module_paths.py | go.work 解析正则 `\./...` 将 `../proto/generated/go` 误当 `./` 处理，报「模块不存在」 | 区分 `../`（相对仓库根）与 `./`（相对 go/）两类条目解析 | 1 |
| check_chinese_comments.py | 未豁免 vendored go-admin 框架的导出符号注释 | 新增 EXCLUDED_PATH_PREFIXES（go/admin/） | 4 |

**豁免依据**：go/admin 为开源 go-admin 框架（module go-admin，自带 LICENSE.md）；go/third_party/TarsGo 为 vendored 的 TarsGo 1.4.6 官方源码；node_modules 为 npm/pnpm 第三方依赖。三者均非本项目自研测试资产。

### 4.2 注册表数据修复（docs/testing/测试用例注册表.md）

| 修复类型 | 数量 | 说明 |
|---|---|---|
| 路径漂移修正 | 7 个文件 | 模块化重构后文件移动未同步注册表：internal/tarsclient/ → tarsclient/（2 文件 4 处）、router_test.go → route_table_test.go、tars/system/localhandler/ → tars/system/adapter/deprecated/、tars/system/modules/{hello,health} → go/modules/{hello,health}、web/src/pages/hello/ → web/tests/unit/hello/ |
| 存量自研测试补登记 | 129 行（Go）+ 12 行（TS） | 按注册规则第 1 条批量补登记；TC-ID 沿用既有前缀续号（TC-GW-0010~、TC-TARS-0005~、TC-WEB-0002~）；测试函数从源码自动提取；状态统一 ⏸️ Pending（未经逐项验证，诚实标注） |

### 4.3 proto 注释补齐（真实违规修复，17 个 message）

| 文件 | 补注释 message |
|---|---|
| proto/base/common.proto | UserInfo、LoginData |
| proto/social/member.proto | GroupUserInfo |
| proto/social/group.proto | OwnerProfile、LatestTopicPreview |
| proto/social/topic.proto | ExternalLink、MediaItem、TopicInfoExtra、TopicStat、ReplyInfo、AddTopicReplyRequest、AddTopicReplyResponse、GetReplyListRequest、GetReplyListResponse、LikeReplyRequest、LikeReplyResponse、LinkInfo |

注释风格沿用文件既有「// 实体说明：」「// 接口用途：」惯例，接口类注明协议号。仅补注释未改协议结构，无需重新生成代码。

## 5. 测试用例执行结果

| 用例 ID | 检查项 | 结果 | 说明 |
|---|---|---|---|
| TC-CI-0001 | check_directory_layout | 通过 | 0 错误 0 警告 |
| TC-CI-0002 | check_make_targets | 通过 | Root/Go/TS/Python targets 全部存在 |
| TC-CI-0003 | check_module_paths | 通过 | go.work 16 个模块全部存在 |
| TC-CI-0004 | check_chinese_comments | 通过 | 0 错误；余 field 级 warning 不阻断（检查器设计如此） |
| TC-CI-0005 | check_testcase_registry | 通过 | Go 139/142、Python 30/31、TS 13/13，登记 0 缺失 0 幽灵 |
| （工具链） | check_tools.sh | 通过 | 必要工具全就绪（3 个可选工具未装仅提示） |

## 6. 验证证据

```text
修复前：make rules 输出 811 行，622 个 ❌（含汇总标记去重后约 308 项独立失败）
修复后：make rules 输出 196 行，六项检查全部 ✅，0 错误
```

## 7. 结论

make rules 六项检查全绿。PRD-12 §14 验收第 7 条「CI 检查通过」的遗留阻塞因素已消除。

## 8. 风险与遗留问题

- [P2/R1] **go 存量技术债（与本次修改无关，建议开独立任务）**：`make unit` 中 2 个 go 失败——
  1. `go/tars/system/cmd [build failed]`：system_adapter.go:16-24 引用旧 API（hello.HelloService / hello.NewService()），而 hello/health 模块已在 8147079 重构为 Handler-UseCase-Checker 三层架构（New(deps) / Service.SayHello），重构时 adapter 未同步；
  2. `tarsclient.TestRegisterModuleHandlers/HelloWorld` FAIL（期望 10200 实际 10400）：模块化注册语义与重构后 hello 模块不匹配，同源。
  判定依据：本次专项仅修改 Python 检查器脚本 / 注册表 md / proto 注释（不涉及 Go 代码与生成物），不可能影响 Go 编译与运行；adapter 旧引用来自历史提交 23ed5e2 与 8147079 的接口漂移。
- [R2] 141 项补登记测试状态为 ⏸️ Pending，未逐项人工验证；后续按模块补齐验证后转 ✅ 有效。
- [R3] proto field 级注释 warning（app_config.proto 等 10 处）为非阻断项，可在后续协议维护任务中顺带补齐。
- [R3] python 登记数（31）略大于实际文件数（30），存在重复登记条目，不影响检查结论，可在下次注册表维护时清理。
- go/admin 与 third_party/TarsGo 的排除规则已写入检查器注释，后续新增 vendored 目录需同步维护 EXCLUDED 常量。
