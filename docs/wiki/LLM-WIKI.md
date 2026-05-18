# LLM Wiki

这是 CaiRobot MVP 项目的 LLM Wiki，用于帮助后续 AI 助手快速理解项目。

## 1. 项目定位

（待补充）

## 2. 系统组成

CaiRobot MVP 当前包含以下系统：

1. **服务商后台系统**
   - 面向服务商、渠道、售后和运营角色
   - 用于管理设备、服务、工单、租户和运营数据

2. **终端用户中台系统**
   - 面向家庭用户、家长和学生账户体系
   - 负责家庭空间、孩子档案、设备绑定、学习记录和用户设置

3. **开放平台 API**
   - 面向第三方系统和生态伙伴
   - 提供认证、授权、设备状态、用户授权、学习数据和 Webhook 能力

4. **AI 服务系统**
   - 使用 Python 实现
   - 负责意图分类、提示词改写、回答审核、OCR 结果理解、模型网关和安全策略执行

5. **Golang 后端服务**
   - 承担服务商后台、用户中台、开放平台、设备网关、认证、审计等核心后端能力

6. **ReactJS 前端系统**
   - 承担服务商后台前端、用户中台前端和终端 App/H5 前端

7. **Protobuf 协议层**
   - 作为服务间通信、开放平台契约、AI 服务契约和设备协议的统一接口定义来源

## 3. 当前阶段

当前项目处于：**多系统工程规范升级阶段**

## 4. 当前最高优先级

1. 完善 PRD-00 到 PRD-06
2. 完善 ADR-0001 到 ADR-0007
3. 建立 Protobuf 协议定义

## 5. 已完成事项

- [x] 搭建工程规范框架
- [x] 建立 PRD 规范和模板
- [x] 建立 TDD 和测试规范
- [x] 建立日报、测试报告、事故上报机制
- [x] 建立 LLM Wiki 框架
- [x] 建立 Standalone HTML 报告机制
- [x] 建立 Markdown 蒸馏机制
- [x] 升级 PRD 体系为多系统结构
- [x] 建立 ADR 技术决策文档
- [x] 建立 Protobuf 协议规范和目录
- [x] 建立 Golang 后端目录规范
- [x] 建立 Python AI 服务目录规范
- [x] 建立 ReactJS 前端目录规范
- [x] 建立协议编号注册表
- [x] 建立 OpenAPI 与 Protobuf 映射规范
- [x] 更新工程规范文件加入协议变更要求
- [x] 更新 PR 模板和 Issue 模板加入协议变更检查
- [x] 建立工程级 TDD 规范
- [x] 建立 GitHub Actions CI Workflow
- [x] 建立 CI 检查脚本骨架
- [x] 更新 PR 模板和 Issue 模板加入 CI 检查
- [x] 建立 HelloWorld + HealthCheck 验收规范
- [x] 实现 HelloWorld 最小验收工程（Golang + Python + React）
- [x] 新增 proto/base/hello.proto，登记协议编号 2100:2101 和 2100:2102
- [x] 完成 TDD 红绿循环（test(hello) + feat(hello) 两次独立提交）
- [x] 更新 Skill 文档缺陷改进（添加测试证据和 CI URL 硬校验）

## 6. 未完成事项

- [ ] 完善 PRD-00 到 PRD-06 的详细内容
- [ ] 完善 ADR-0001 到 ADR-0007 的详细内容
- [ ] 编写具体的 Protobuf 协议定义
- [ ] 建立各服务测试框架
- [ ] 实现 HelloWorld + HealthCheck 验收工程
- [ ] 开始功能开发

## 7. 工程级 TDD 与 CI 规则

### 7.1 核心原则

**用户任务不完整规则**：用户任务描述不一定完整。Trae 必须根据 AGENTS.md、PRD、ADR、TDD、testing.md、review.md、reporting.md 主动补齐工程闭环。

**完成标准**：没有 CI 通过或本地等价测试结果，不得宣称任务完成。

### 7.2 工程级 TDD 流程

1. **需求红**：先明确 PRD 验收标准
2. **协议红**：先定义 Protobuf / OpenAPI 契约
3. **测试红**：先写失败测试
4. **实现绿**：写最小实现
5. **报告绿**：输出测试报告和证据
6. **CI 绿**：GitHub Actions 或本地等价命令通过
7. **重构**：测试与 CI 通过后再优化
8. **沉淀**：日报、Markdown 蒸馏、LLM Wiki 更新

### 7.3 CI 检查项

| Job | 作用 | 要求 |
|---|---|---|
| `docs-check` | 检查关键文档是否存在 | 必须通过 |
| `proto-check` | 检查协议编号唯一性和注册表同步 | 必须通过 |
| `go-test` | 运行 Golang 测试 | 通过或说明跳过原因 |
| `python-test` | 运行 Python 测试 | 通过或说明跳过原因 |
| `web-test` | 运行 ReactJS App 测试 | 通过或说明跳过原因 |
| `admin-web-test` | 运行 AdminWeb 测试 | 通过或说明跳过原因 |
| `report-check` | 检查测试报告、日报、蒸馏、LLM Wiki | 必须通过 |

### 7.4 CI 文件位置

- Workflow 定义：`.github/workflows/ci.yml`
- 检查脚本：`scripts/ci/`
  - `check_required_docs.py`
  - `check_proto_registry.py`
  - `check_reports.py`

## 8. Protobuf 协议核心规则

**最高级协议规则**：在 CaiRobot MVP 中，Protobuf 协议编号 `max + min` 是接口报文的唯一身份。任何新增、修改、删除接口报文，都必须同步更新协议编号注册表、OpenAPI 映射、测试用例、测试报告和 LLM Wiki。

### 7.1 协议编号注册表

位置：[docs/api/协议编号注册表.md](../api/协议编号注册表.md)

当前已登记的编号：
| max | min | Message | 说明 |
|---:|---:|---|---|
| 2100 | 2097 | ServiceHealthCheckRequest | 服务健康检查请求 |
| 2100 | 2098 | ServiceHealthCheckResponse | 服务健康检查响应 |
| 2100 | 2101 | HelloWorldRequest | HelloWorld 请求 |
| 2100 | 2102 | HelloWorldResponse | HelloWorld 响应 |

### 7.2 协议文件

- 网关统一入口：[proto/base/message.proto](../../proto/base/message.proto)
- 通用返回：[proto/base/result.proto](../../proto/base/result.proto)
- 健康检查示例：[proto/base/health.proto](../../proto/base/health.proto)

### 7.3 协议规范文档

- [protobuf规范.md](../api/protobuf规范.md)
- [openapi-protobuf映射规范.md](../api/openapi-protobuf映射规范.md)

## 8. 关键目录说明

| 目录 | 说明 |
|---|---|
| docs/prd/ | PRD 文档 |
| docs/adr/ | ADR 文档 |
| docs/api/ | API 文档 |
| docs/testing/ | 测试规范 |
| docs/reports/ | 报告目录（日报、测试报告、事故报告等） |
| docs/wiki/ | LLM Wiki（本文件） |
| .trae/rules/ | Trae 规则 |
| .github/ | GitHub 模板 |
| proto/ | Protobuf 协议定义 |
| services/ | Golang 后端服务 |
| ai/ | AI 模块代码 |
| web/ | ReactJS 前端代码 |
| app/ | App 代码 |
| hardware/ | 硬件相关 |
| tests/ | 测试代码 |

## 8. 工程规则索引

详见 [工程规范索引.md](工程规范索引.md)

关键规则摘要：

- **PRD 优先**：开发前必须有 PRD
- **TDD**：先写测试，再写代码
- **每日 22:00 前日报**：每日必须提交日报
- **Markdown 蒸馏**：日报后必须蒸馏知识
- **双主控评审**：重大事项需要项目主控确认

## 9. PRD 索引

详见 [PRD索引.md](PRD索引.md)

| PRD | 状态 | 说明 |
|---|---|---|
| PRD-00/MVP 总纲 | 草稿 | |
| PRD-01/服务商后台系统 | 草稿 | |
| PRD-02/终端用户中台系统 | 草稿 | |
| PRD-03/开放平台 API | 草稿 | |
| PRD-04/AI 服务系统 | 草稿 | |
| PRD-05/App 前端系统 | 草稿 | |
| PRD-06/设备通信与协议 | 草稿 | |

## 10. ADR 索引

详见 [ADR索引.md](ADR索引.md)

| ADR | 状态 | 说明 |
|---|---|---|
| ADR-0001/总体系统架构 | 草稿 | |
| ADR-0002/后端使用 Golang | 草稿 | |
| ADR-0003/服务协议使用 Protobuf | 草稿 | |
| ADR-0004/AI 服务使用 Python | 草稿 | |
| ADR-0005/App 前端使用 ReactJS | 草稿 | |
| ADR-0006/开放平台 API 边界 | 草稿 | |
| ADR-0007/服务商后台与用户中台边界 | 草稿 | |

## 11. 测试索引

详见 [测试索引.md](测试索引.md)

（暂无内容）

## 12. Bug 索引

详见 [Bug索引.md](Bug索引.md)

（暂无内容）

## 13. 事故索引

（暂无内容）

## 14. 每日蒸馏索引

（暂无内容）

## 15. Tabbit / TRAE 任务归档索引

从 `docs/tabbit/inbox/`（Tabbit/TabAI 导出）和 `docs/trae-export/inbox/`（TRAE 导出）通过 `/tabbit-task` 归档的结构化任务文档。

**核心原则：一个任务，一个 Canonical Task ID，一组关联文件。ID 可外部传入，但必须由 `/tabbit-task` 保证存在。**

决策依据：[ADR-0009 v3](../adr/ADR-0009-tabbit-task-archive-flow.md)

归档命令：`/tabbit-task`（v3：自动建档版，支持四种输入模式）

蒸馏 Skill：`tabbit-task-distillation`

Task ID 格式：`TB-{YYYYMMDD}-{HHMMSS}-{topic-slug}`

| Canonical Task ID | 任务 | 时间 | 来源 | 蒸馏状态 | 摘要 |
|---|---|---|---|---|---|
| `TB-20260518-181000` | [Tabbit 任务归档流程设计与落地](./tasks/2026/05/TB-20260518-181000-tabbit-task-archive-flow-design.archive.md) | 18:10 | Tabbit + TRAE | ⏳ pending | 设计并落地了从 Tabbit/TRAE 原始导出到 docs/wiki/tasks 的归档流程，含 ADR-0009 和 /tabbit-task Command |
| `TB-20260518-182000` | [TRAE 评审 TabAI 会话导出方案](./tasks/2026/05/TB-20260518-182000-trae-tabbage-archive-review.archive.md) | 18:20 | TRAE | ⏳ pending | 对原始方案七维工程评审，发现 4 个必须修改项，核心结论是不应新建 docs/llm-wiki/ |
| `TB-20260518-190000` | [Tabbit 归档方案升级为 Task ID 驱动架构](./tasks/2026/05/TB-20260518-190000-tabbage-task-id-upgrade.archive.md) | 19:00 | Tabbit + 架构师 + TRAE | ⏳ pending | v1→v2：语义化文件名升级为 Task ID 驱动资产链，新增 manifest 和蒸馏 Skill |
| `TB-20260518-194000` | [Tabbit 归档方案 v3：Canonical Task ID 自动生成](./tasks/2026/05/TB-20260518-194000-tabbage-canonical-auto-id.archive.md) | 19:40 | Tabbit + 架构师 + TRAE | ⏳ pending | v2→v3：解除对上游 Tabbit 的依赖，Canonical Task ID 由 /tabbit-task 兜底自动生成，支持四种输入模式 |
| `TB-20260518-200000` | [CaiRobot MVP 单网关架构设计](./tasks/2026/05/TB-20260518-200000-cairobot-single-gateway.archive.md) | 20:00 | Tabbit | ⏳ pending | CaiRobot 单网关 MessagePacket + TarsCloud/TarsGo 统一架构方案（初稿 + 评审修订版） |
| `TB-20260518-201000` | [手动占位空文件归档](./tasks/2026/05/TB-20260518-201000-manual-placeholder-note.archive.md) | 20:01 | 手动创建 | ⏳ ⚠️ 需人工确认 | 早期空占位文件 1.md 的归档记录，无可蒸馏内容 |

### Manifest 蒸馏队列

以下任务的 manifest 蒸馏状态为 `pending`，可由 `tabbit-task-distillation` Skill 处理：

| Canonical Task ID | manifest 路径 | 蒸馏优先级 |
|---|---|---|
| TB-20260518-181000 | [manifest](./tasks/2026/05/TB-20260518-181000-tabbage-task-archive-flow-design.manifest.md) | normal |
| TB-20260518-182000 | [manifest](./tasks/2026/05/TB-20260518-182000-trae-tabbage-archive-review.manifest.md) | normal |
| TB-20260518-190000 | [manifest](./tasks/2026/05/TB-20260518-190000-tabbage-task-id-upgrade.manifest.md) | high |
| TB-20260518-194000 | [manifest](./tasks/2026/05/TB-20260518-194000-tabbage-canonical-auto-id.manifest.md) | high |
| TB-20260518-200000 | [manifest](./tasks/2026/05/TB-20260518-200000-cairobot-single-gateway.manifest.md) | normal |
| TB-20260518-201000 | [manifest](./tasks/2026/05/TB-20260518-201000-manual-placeholder-note.manifest.md) | low (⚠️ needs_human_review) |

## 16. 后续 AI 助手必须遵守的规则

AI 助手开始工作前，必须：

1. 阅读 [AGENTS.md](../../AGENTS.md)
2. 阅读 [.trae/rules/project.md](../../.trae/rules/project.md)
3. 阅读本 LLM-WIKI.md
4. 明确任务范围后再开始
5. 严格遵循 TDD
6. 每日 22:00 前提交日报
7. 遇到不确定事项及时上报
8. 不得随意扩大任务范围
