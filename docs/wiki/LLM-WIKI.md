# LLM Wiki（Index 层入口）

> **本文件是 LLM Wiki 三层结构的 Index 层（目录层），只做导航、摘要、索引和状态入口。**
> **不承载完整日报、完整审计、完整对话或完整蒸馏正文。**

## 1. 项目定位

CaiRobot MVP 是一个多系统、多技术栈的微服务架构项目，采用单网关 + MessagePacket + TarsCloud/TarsGo 的统一架构。

## 2. 当前阶段

**S0：规则与目录骨架** 阶段。

| 阶段 | 说明 |
|---|---|
| S0 | 规则建设、目录骨架（当前） |
| S1 | PRD + ADR 落地 |
| S2 | 核心功能开发 |
| S3 | 稳定运营 |

## 3. 当前最高优先级

1. 完善 PRD-00 到 PRD-06
2. 完善 ADR-0001 到 ADR-0007
3. 建立 Protobuf 协议定义

## 4. 后续 AI 必读顺序

```
1. AGENTS.md                          （顶层工程协作规范）
2. docs/wiki/LLM-WIKI.md              （本文件，Wiki 总索引）
3. docs/wiki/CODE-WIKI.md             （代码架构知识）
4. docs/wiki/工程规范索引.md           （工程规范摘要）
5. 明确任务范围后再开始
```

## 5. 三层知识库结构

LLM Wiki 采用 **Raw → Distillation → Index** 三层结构：

| 层级 | 职责 | 目录示例 |
|---|---|---|
| **Raw 原始层** | 保存原始事实、导出文档、命令输出、测试结果、审计报告。不做过度总结，不删除、不覆盖。 | `docs/tabbit/inbox/`、`docs/reports/daily/`、`docs/reports/testing/`、`docs/reports/audit/`、`docs/reports/coverage/` |
| **Distillation 蒸馏层** | 从 Raw 材料中提取可复用的结构化知识。区分事实与判断，不把计划写成完成。 | `docs/wiki/daily/`、`docs/wiki/tasks/`、`docs/wiki/decisions/`、`docs/wiki/modules/` |
| **Index 目录层** | 只做导航、摘要、索引、状态入口。不承载完整正文。 | `docs/wiki/LLM-WIKI.md`（本文件）、`docs/wiki/CODE-WIKI.md`、`每日蒸馏索引.md`、`任务索引.md` |

架构决策详见：[decisions/llm-wiki-three-layer-architecture.md](./decisions/llm-wiki-three-layer-architecture.md)

### 5.1 核心原则

- **Raw 层保真**：原始材料不可篡改、不可覆盖、不可删除。
- **Distillation 层压缩**：从 Raw 中提炼稳定知识，过滤噪声。
- **Index 层导航**：只保留摘要和链接，不做知识正文池。
- **LLM-WIKI.md 不是正文仓库**：本文件只做入口索引。

## 6. Raw 层入口

| 内容 | 路径 | 说明 |
|---|---|---|
| Tabbit / TabAI 导出收件箱 | `docs/tabbit/inbox/` | 原始会话导出 |
| TRAE 执行导出收件箱 | `docs/trae-export/inbox/` | TRAE 执行记录 |
| 每日日报（Raw） | `docs/reports/daily/` | 每日原始日报 |
| 测试报告 | `docs/reports/testing/` | TDD 红绿报告 |
| 审计报告 | `docs/reports/audit/` | 审计记录 |
| 覆盖率报告 | `docs/reports/coverage/` | 测试覆盖率 |

## 7. Distillation 层入口

| 内容 | 路径 | 说明 |
|---|---|---|
| 每日蒸馏知识 | `docs/wiki/daily/` | 从日报蒸馏的稳定知识 |
| 任务归档 | `docs/wiki/tasks/` | Tabbit / TRAE 归档任务（ADR-0009 体系） |
| 架构决策 | `docs/wiki/decisions/` | 经确认的技术决策（ADR 补充） |
| 模块知识 | `docs/wiki/modules/` | 各模块的稳定知识页 |

## 8. Index 层入口

| 文件 | 说明 |
|---|---|
| **LLM-WIKI.md**（本文件） | 项目级 Wiki 主入口索引 |
| [CODE-WIKI.md](./CODE-WIKI.md) | 代码架构知识（单网关、Protobuf、Tars 等） |
| [工程规范索引.md](./工程规范索引.md) | 工程规范汇总 |
| [ADR索引.md](./ADR索引.md) | 架构决策记录索引 |
| [PRD索引.md](./PRD索引.md) | 产品需求文档索引 |
| [测试索引.md](./测试索引.md) | 测试相关索引 |
| [Bug索引.md](./Bug索引.md) | Bug 追踪索引 |
| [决策记录.md](./决策记录.md) | 重要决策记录 |
| [README.md](./README.md) | Wiki 目录说明 |
| [每日蒸馏索引.md](./每日蒸馏索引.md) | 每日蒸馏产物索引 |
| [任务索引.md](./任务索引.md) | 任务级索引 |

## 9. 系统组成概要

| # | 系统 | 技术栈 | 详情 |
|---|---|---|---|
| 1 | 服务商后台系统 | Golang + ReactJS | 设备、服务、工单、租户管理 |
| 2 | 终端用户中台系统 | Golang + ReactJS | 家庭空间、孩子档案、设备绑定 |
| 3 | 开放平台 API | Golang + Protobuf | 认证、授权、设备状态、Webhook |
| 4 | AI 服务系统 | Python | 意图分类、提示词改写、审核、OCR |
| 5 | Golang 后端服务 | Golang + TarsGo | 服务商后台、用户中台、设备网关 |
| 6 | ReactJS 前端系统 | ReactJS + TypeScript | 服务商前端、用户中台前端、App/H5 |
| 7 | Protobuf 协议层 | Protobuf | 服务间通信统一接口定义 |

完整技术架构见：[CODE-WIKI.md](./CODE-WIKI.md)

## 10. 决策索引

详见 [ADR索引.md](./ADR索引.md) 和 [decisions/](./decisions/)。

核心已接受决策：
- ADR-0008: TarsCloud 内部 RPC 层
- ADR-009: Tabbit 任务归档流程
- ADR-0012: 多语言 Monorepo 目录布局
- ADR-0013: Makefile 工程入口与规则强制执行
- **LLM Wiki 三层架构**（2026-05-19 新增）

## 11. 每日蒸馏索引入口

详见 [每日蒸馏索引.md](./每日蒸馏索引.md)。

## 12. 任务索引入口

详见 [任务索引.md](./任务索引.md)。包含 Tabbit / TRAE / SOLO Web 归档任务（ADR-0009 Canonical Task ID 体系）。

## 13. 当前进度摘要

### 已完成（S0 阶段）

工程规范框架、PRD/ADR/TDD/测试/CI 体系、HelloWorld 验收、Gateway 双模式骨架、三层 Makefile、多语言 monorepo 布局、中文注释规范、Tabbit 归档流程（v3）、**LLM Wiki 三层结构**。

完整清单详见各子索引文件。

### 进行中 / 待完成

- PRD-00 到 PRD-06 详细内容完善
- ADR-0001 到 ADR-0007 详细内容完善
- 具体 Protobuf 协议定义
- 各服务测试框架建立
- 功能开发启动

## 14. SOLO Web 自动化能力边界

SOLO Web 已验证可承担以下职责：
- 定时触发
- 调用大模型生成 Markdown 文件
- 生成 Raw、日报、蒸馏和主控汇报产物
- 遵守"不 git commit、不 git push"限制

**SOLO Web 不承担**：结构性重构、Skill/Command 制作、长期知识库结构调整。

详见：[modules/solo-web-automation.md](./modules/solo-web-automation.md)

## 15. 长期工程规则摘要

| 规则 | 来源 |
|---|---|
| PRD 驱动开发 | AGENTS.md §3 |
| Issue 驱动任务拆分 | AGENTS.md §3 |
| TDD 红-绿-重构 | [.trae/rules/tdd.md](../../.trae/rules/tdd.md) |
| 小步提交 | [.trae/rules/git.md](../../.trae/rules/git.md) |
| 每日 22:00 前日报 | [.trae/rules/reporting.md](../../.trae/rules/reporting.md) |
| Markdown 蒸馏 → LLM Wiki | AGENTS.md §15 |
| 双主控评审 | [.trae/rules/reporting.md](../../.trae/rules/reporting.md) |
| 协议编号 max+min 唯一性 | CODE-WIKI.md §4 |
| 单网关 MessagePacket 入口 | CODE-WIKI.md §3 |

## 16. 禁止事项

1. 不允许删除已有日报、蒸馏、审计、任务归档原始材料
2. 不允许把原始材料直接塞进 LLM-WIKI.md
3. 不允许把未确认内容写成长期确定规则
4. 不允许把计划写成完成、设计写成实现
5. 不允许修改业务代码（本次任务范围）
6. 不允许未经主控确认 git push
7. 不允许把 mock、TODO、空实现写成主链路完成
